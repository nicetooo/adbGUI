/**
 * Event Store - 新事件系统的前端状态管理
 *
 * 特点:
 * - 分层存储: 实时事件 (RingBuffer) + 分页缓存 (LRU) + 按需加载
 * - 与后端 SQLite 存储对接
 * - 支持虚拟滚动的数据加载
 */

import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';
import {
  UnifiedEvent,
  DeviceSession,
  EventQuery,
  EventQueryResult,
  TimeIndexEntry,
  Bookmark,
  EventSource,
  EventCategory,
  EventLevel,
} from './eventTypes';

const EventsOn = (window as any).runtime?.EventsOn;
const EventsOff = (window as any).runtime?.EventsOff;

// ========================================
// LRU Cache 实现
// ========================================

class LRUCache<K, V> {
  private capacity: number;
  private cache: Map<K, V>;

  constructor(capacity: number) {
    this.capacity = capacity;
    this.cache = new Map();
  }

  get(key: K): V | undefined {
    if (!this.cache.has(key)) return undefined;
    const value = this.cache.get(key)!;
    this.cache.delete(key);
    this.cache.set(key, value);
    return value;
  }

  set(key: K, value: V): void {
    if (this.cache.has(key)) {
      this.cache.delete(key);
    } else if (this.cache.size >= this.capacity) {
      const firstKey = this.cache.keys().next().value;
      if (firstKey !== undefined) {
        this.cache.delete(firstKey);
      }
    }
    this.cache.set(key, value);
  }

  has(key: K): boolean {
    return this.cache.has(key);
  }

  clear(): void {
    this.cache.clear();
  }

  get size(): number {
    return this.cache.size;
  }
}

// ========================================
// Ring Buffer 实现
// ========================================

class RingBuffer<T> {
  private buffer: T[];
  private head: number = 0;
  private count: number = 0;

  constructor(private capacity: number) {
    this.buffer = new Array(capacity);
  }

  push(item: T): void {
    this.buffer[this.head] = item;
    this.head = (this.head + 1) % this.capacity;
    if (this.count < this.capacity) this.count++;
  }

  pushMany(items: T[]): void {
    items.forEach(item => this.push(item));
  }

  getAll(): T[] {
    if (this.count === 0) return [];
    const result: T[] = [];
    const start = this.count < this.capacity ? 0 : this.head;
    for (let i = 0; i < this.count; i++) {
      result.push(this.buffer[(start + i) % this.capacity]);
    }
    return result;
  }

  getRecent(n: number): T[] {
    if (n <= 0 || this.count === 0) return [];
    n = Math.min(n, this.count);
    const result: T[] = [];
    const start = (this.head - n + this.capacity) % this.capacity;
    for (let i = 0; i < n; i++) {
      result.push(this.buffer[(start + i) % this.capacity]);
    }
    return result;
  }

  clear(): void {
    this.head = 0;
    this.count = 0;
  }

  get size(): number {
    return this.count;
  }
}

// ========================================
// Constants
// ========================================

const LIVE_BUFFER_SIZE = 2000;   // 实时事件缓冲大小
const PAGE_CACHE_SIZE = 50;      // 分页缓存大小
const DEFAULT_PAGE_SIZE = 200;   // 默认每页事件数

// ========================================
// Store State & Actions
// ========================================

interface EventStoreState {
  // Sessions
  sessions: Map<string, DeviceSession>;
  sessionsVersion: number;  // Force re-render when sessions change
  activeSessionId: string | null;
  activeDeviceId: string | null;

  // Events - 分层存储
  liveEvents: RingBuffer<UnifiedEvent>;
  pageCache: LRUCache<string, UnifiedEvent[]>;

  // 时间索引
  timeIndex: Map<string, TimeIndexEntry[]>;

  // 书签
  bookmarks: Map<string, Bookmark[]>;

  // 当前视图
  visibleRange: { start: number; end: number };
  visibleEvents: UnifiedEvent[];

  // 筛选条件
  filter: EventQuery;

  // 加载状态
  isLoading: boolean;
  totalEventCount: number;    // session 的总事件数
  filteredEventCount: number; // 匹配过滤条件的事件数

  // UI 状态
  selectedEventId: string | null;
  isTimelineOpen: boolean;
  autoScroll: boolean;
}

interface EventStoreActions {
  // Session 管理
  setActiveSession: (sessionId: string | null) => void;
  setActiveDevice: (deviceId: string) => void;
  loadSession: (sessionId: string) => Promise<void>;
  startSession: (deviceId: string, type: string, name: string) => Promise<string>;
  endSession: (sessionId: string, status: string) => Promise<void>;
  deleteSession: (sessionId: string) => Promise<void>;
  loadSessions: (deviceId?: string, limit?: number) => Promise<DeviceSession[]>;

  // 实时事件
  receiveLiveEvents: (events: UnifiedEvent[]) => void;

  // 按需加载
  loadEventsInRange: (startTime: number, endTime: number) => Promise<void>;
  loadPage: (page: number, pageSize?: number) => Promise<EventQueryResult>;
  searchEvents: (query: string) => Promise<UnifiedEvent[]>;
  loadEvent: (eventId: string) => Promise<UnifiedEvent | null>;

  // 时间跳转
  jumpToTime: (relativeTime: number) => Promise<void>;
  jumpToEvent: (eventId: string) => Promise<void>;

  // 筛选
  setFilter: (filter: Partial<EventQuery>) => void;
  clearFilter: () => void;
  applyFilter: () => Promise<void>;

  // 可视区域
  setVisibleRange: (start: number, end: number) => void;

  // 时间索引
  loadTimeIndex: (sessionId: string) => Promise<void>;

  // 书签
  loadBookmarks: (sessionId: string) => Promise<void>;
  createBookmark: (relativeTime: number, label: string, color?: string, type?: string) => Promise<void>;
  deleteBookmark: (bookmarkId: string) => Promise<void>;

  // UI
  selectEvent: (eventId: string | null) => void;
  openTimeline: () => void;
  closeTimeline: () => void;
  setAutoScroll: (enabled: boolean) => void;

  // 统计
  getSessionStats: (sessionId: string) => Promise<Record<string, any>>;
  getSystemStats: () => Promise<Record<string, any>>;

  // 清理
  clearSession: () => void;
  cleanup: (maxAgeDays: number) => Promise<number>;

  // 事件订阅
  subscribeToEvents: () => () => void;
}

// ========================================
// Store 实现
// ========================================

export const useEventStore = create<EventStoreState & EventStoreActions>()(
  immer((set, get) => ({
    // 初始状态
    sessions: new Map(),
    sessionsVersion: 0,
    activeSessionId: null,
    activeDeviceId: null,
    liveEvents: new RingBuffer<UnifiedEvent>(LIVE_BUFFER_SIZE),
    pageCache: new LRUCache<string, UnifiedEvent[]>(PAGE_CACHE_SIZE),
    timeIndex: new Map(),
    bookmarks: new Map(),
    visibleRange: { start: 0, end: 60000 }, // 默认显示前 60 秒
    visibleEvents: [],
    filter: {},
    isLoading: false,
    totalEventCount: 0,
    filteredEventCount: 0,
    selectedEventId: null,
    isTimelineOpen: false,
    autoScroll: true,

    // ========================================
    // Session 管理
    // ========================================

    setActiveSession: (sessionId) => {
      set(state => {
        state.activeSessionId = sessionId;
        if (!sessionId) {
          state.liveEvents.clear();
          state.visibleEvents = [];
          state.totalEventCount = 0;
        }
      });
    },

    setActiveDevice: (deviceId) => {
      set(state => {
        state.activeDeviceId = deviceId;
      });
    },

    loadSession: async (sessionId) => {
      console.log('[eventStore] loadSession called:', sessionId);

      // Reset state for new session
      const currentState = get();
      set(state => {
        state.isLoading = true;
        state.activeSessionId = sessionId;
        state.liveEvents = new RingBuffer<UnifiedEvent>(LIVE_BUFFER_SIZE);
        state.pageCache.clear();
        state.visibleEvents = [];
        state.visibleRange = { start: 0, end: 60000 };
        state.sessionsVersion = currentState.sessionsVersion + 1;
      });

      try {
        // 加载 session 信息
        console.log('[eventStore] loadSession calling GetStoredSession...');
        const session = await (window as any).go.main.App.GetStoredSession(sessionId);
        console.log('[eventStore] loadSession got session:', session);

        if (session) {
          // Create new Map to avoid immer issues
          const newSessions = new Map(get().sessions);
          newSessions.set(sessionId, session);

          set(state => {
            state.sessions = newSessions;
            state.activeDeviceId = session.deviceId;
            state.totalEventCount = session.eventCount || 0;
          });
        }

        // 加载时间索引和书签（不阻塞事件加载）
        Promise.all([
          (window as any).go.main.App.GetSessionTimeIndex(sessionId).catch(() => []),
          (window as any).go.main.App.GetSessionBookmarks(sessionId).catch(() => []),
        ]).then(([index, bookmarks]) => {
          const currentTimeIndex = new Map(get().timeIndex);
          const currentBookmarks = new Map(get().bookmarks);
          currentTimeIndex.set(sessionId, index || []);
          currentBookmarks.set(sessionId, bookmarks || []);

          set(state => {
            state.timeIndex = currentTimeIndex;
            state.bookmarks = currentBookmarks;
          });
        }).catch(err => {
          console.error('[eventStore] loadSession index/bookmarks error:', err);
        });

        // 查询事件数据 - 这是主要的加载
        console.log('[eventStore] loadSession calling QuerySessionEvents...');
        const result = await (window as any).go.main.App.QuerySessionEvents({
          sessionId: sessionId,
          startTime: 0,
          endTime: 300000, // 5 minutes range
          limit: 1000,
        });
        console.log('[eventStore] loadSession got events:', result?.events?.length || 0, 'total:', result?.total);

        set(state => {
          state.visibleEvents = result?.events || [];
          state.totalEventCount = result?.total || 0;
          state.filteredEventCount = result?.total || 0; // 初始无过滤时，filteredCount = totalCount
          state.isLoading = false;
        });

        console.log('[eventStore] loadSession complete');

      } catch (err) {
        console.error('[eventStore] loadSession error:', err);
        set(state => {
          state.isLoading = false;
        });
      }
    },

    startSession: async (deviceId, type, name) => {
      console.log('[eventStore] startSession called:', { deviceId, type, name });
      try {
        const sessionId = await (window as any).go.main.App.StartNewSession(deviceId, type, name);
        console.log('[eventStore] StartNewSession returned:', sessionId);

        // Get session from backend
        const session = await (window as any).go.main.App.GetStoredSession(sessionId);
        console.log('[eventStore] GetStoredSession returned:', session);

        // Get current state
        const currentState = get();
        console.log('[eventStore] Current state before update:', {
          sessionsSize: currentState.sessions.size,
          sessionsVersion: currentState.sessionsVersion
        });

        // Create new sessions map
        const newSessions = new Map(currentState.sessions);
        if (session) {
          newSessions.set(sessionId, session);
        }

        // Update state without immer - use functional update
        set((state) => {
          state.sessions = newSessions;
          state.activeSessionId = sessionId;
          state.activeDeviceId = deviceId;
          state.visibleEvents = [];
          state.totalEventCount = 0;
          state.sessionsVersion = currentState.sessionsVersion + 1;
          state.liveEvents = new RingBuffer<UnifiedEvent>(LIVE_BUFFER_SIZE);
        });

        // Verify update
        const newState = get();
        console.log('[eventStore] State after update:', {
          activeSessionId: newState.activeSessionId,
          sessionsSize: newState.sessions.size,
          sessionsVersion: newState.sessionsVersion,
          sessionInMap: newState.sessions.has(sessionId)
        });

        return sessionId;
      } catch (err) {
        console.error('[eventStore] startSession error:', err);
        throw err;
      }
    },

    endSession: async (sessionId, status) => {
      await (window as any).go.main.App.EndActiveSession(sessionId, status);

      const { activeSessionId } = get();
      if (activeSessionId === sessionId) {
        set(state => {
          state.activeSessionId = null;
        });
      }
    },

    deleteSession: async (sessionId) => {
      await (window as any).go.main.App.DeleteStoredSession(sessionId);

      set(state => {
        state.sessions.delete(sessionId);
        state.timeIndex.delete(sessionId);
        state.bookmarks.delete(sessionId);

        if (state.activeSessionId === sessionId) {
          state.activeSessionId = null;
          state.liveEvents.clear();
          state.visibleEvents = [];
        }
      });
    },

    loadSessions: async (deviceId, limit = 50) => {
      const sessions = await (window as any).go.main.App.ListStoredSessions(deviceId || '', limit);

      set(state => {
        (sessions || []).forEach((session: DeviceSession) => {
          state.sessions.set(session.id, session);
        });
      });

      return sessions || [];
    },

    // ========================================
    // 实时事件
    // ========================================

    receiveLiveEvents: (events) => {
      const { activeSessionId, filter, autoScroll, visibleRange } = get();

      if (!activeSessionId || events.length === 0) return;

      // 只处理当前 session 的事件
      const relevantEvents = events.filter(e => e.sessionId === activeSessionId);
      if (relevantEvents.length === 0) return;

      set(state => {
        // 添加到实时缓冲
        state.liveEvents.pushMany(relevantEvents);
        state.totalEventCount += relevantEvents.length;

        // 过滤新事件
        const newFiltered = filterEvents(relevantEvents, filter);
        // 更新匹配过滤条件的事件数
        state.filteredEventCount += newFiltered.length;

        // 追加所有匹配过滤条件的新事件到列表（不再限制 visibleRange）
        if (newFiltered.length > 0) {
          // 如果自动滚动，更新可视范围到最新
          if (autoScroll) {
            const lastEvent = relevantEvents[relevantEvents.length - 1];
            const newEnd = lastEvent.relativeTime + 5000;
            const windowSize = visibleRange.end - visibleRange.start;
            state.visibleRange = {
              start: Math.max(0, newEnd - windowSize),
              end: newEnd,
            };
          }

          // 直接追加所有匹配的新事件
          state.visibleEvents = [...state.visibleEvents, ...newFiltered];

          // 限制列表数量，防止内存溢出（保留最新的 2000 条）
          if (state.visibleEvents.length > 2000) {
            state.visibleEvents = state.visibleEvents.slice(-2000);
          }
        }
      });
    },

    // ========================================
    // 按需加载
    // ========================================

    loadEventsInRange: async (startTime, endTime) => {
      const { activeSessionId, filter, pageCache } = get();
      // 确保时间是整数（后端需要 int64）
      const start = Math.round(startTime);
      const end = Math.round(endTime);
      console.log('[eventStore] loadEventsInRange:', { activeSessionId, start, end });
      if (!activeSessionId) {
        console.log('[eventStore] loadEventsInRange: no activeSessionId');
        return;
      }

      const cacheKey = `${activeSessionId}:${start}:${end}:${JSON.stringify(filter)}`;

      // 检查缓存
      if (pageCache.has(cacheKey)) {
        const cached = pageCache.get(cacheKey)!;
        console.log('[eventStore] loadEventsInRange: using cache, events:', cached.length);
        set(state => {
          state.visibleEvents = cached;
        });
        return;
      }

      set(state => { state.isLoading = true; });

      try {
        const query: EventQuery = {
          sessionId: activeSessionId,
          startTime: start,
          endTime: end,
          ...filter,
          limit: 1000,
        };

        console.log('[eventStore] loadEventsInRange: querying with', query);
        const result = await (window as any).go.main.App.QuerySessionEvents(query);
        console.log('[eventStore] loadEventsInRange: got result', result);
        const events = result?.events || [];

        // 缓存结果
        set(state => {
          state.pageCache.set(cacheKey, events);
          state.visibleEvents = events;
          state.totalEventCount = result?.total || 0;
        });
        console.log('[eventStore] loadEventsInRange: set visibleEvents:', events.length);

      } finally {
        set(state => { state.isLoading = false; });
      }
    },

    loadPage: async (page, pageSize = DEFAULT_PAGE_SIZE) => {
      const { activeSessionId, filter, pageCache } = get();

      if (!activeSessionId) {
        return { events: [], total: 0, hasMore: false };
      }

      const cacheKey = `${activeSessionId}:page:${page}:${pageSize}:${JSON.stringify(filter)}`;

      // 检查缓存
      const cached = pageCache.get(cacheKey);
      if (cached) {
        return {
          events: cached,
          total: get().totalEventCount,
          hasMore: (page + 1) * pageSize < get().totalEventCount,
        };
      }

      const query: EventQuery = {
        sessionId: activeSessionId,
        ...filter,
        limit: pageSize,
        offset: page * pageSize,
      };

      const result = await (window as any).go.main.App.QuerySessionEvents(query);
      const events = result?.events || [];

      // 缓存
      set(state => {
        state.pageCache.set(cacheKey, events);
        state.totalEventCount = result?.total || 0;
      });

      return {
        events,
        total: result?.total || 0,
        hasMore: result?.hasMore || false,
      };
    },

    searchEvents: async (query) => {
      const { activeSessionId } = get();
      if (!activeSessionId) return [];

      const result = await (window as any).go.main.App.QuerySessionEvents({
        sessionId: activeSessionId,
        searchText: query,
        limit: 100,
      });

      return result?.events || [];
    },

    loadEvent: async (eventId) => {
      const event = await (window as any).go.main.App.GetStoredEvent(eventId);
      return event || null;
    },

    // ========================================
    // 时间跳转
    // ========================================

    jumpToTime: async (relativeTime) => {
      const { activeSessionId, loadEventsInRange } = get();
      if (!activeSessionId) return;

      const windowSize = 30000; // 显示 30 秒窗口
      const start = Math.max(0, relativeTime - windowSize / 2);
      const end = relativeTime + windowSize / 2;

      set(state => {
        state.visibleRange = { start, end };
        state.autoScroll = false; // 手动跳转时关闭自动滚动
      });

      await loadEventsInRange(start, end);
    },

    jumpToEvent: async (eventId) => {
      const { loadEvent, jumpToTime } = get();
      const event = await loadEvent(eventId);

      if (event) {
        await jumpToTime(event.relativeTime);
        set(state => {
          state.selectedEventId = eventId;
        });
      }
    },

    // ========================================
    // 筛选
    // ========================================

    setFilter: (newFilter) => {
      set(state => {
        state.filter = { ...state.filter, ...newFilter };
      });
    },

    clearFilter: () => {
      set(state => {
        state.filter = {};
        state.pageCache.clear();
      });
      get().applyFilter();
    },

    applyFilter: async () => {
      const { activeSessionId, filter, liveEvents, sessions } = get();

      if (!activeSessionId) return;

      // 清除缓存
      set(state => {
        state.pageCache.clear();
        state.isLoading = true;
      });

      try {
        // 获取 session 总事件数（不带过滤条件）
        const totalResult = await (window as any).go.main.App.QuerySessionEvents({
          sessionId: activeSessionId,
          limit: 1, // 只需要 total 数量
        });
        const sessionTotalCount = totalResult?.total || 0;

        // 数据库是 source of truth，查询符合过滤条件的事件
        const query = {
          sessionId: activeSessionId,
          ...filter,
          limit: 2000, // 显示列表最多 2000 条
        };

        const result = await (window as any).go.main.App.QuerySessionEvents(query);
        const dbEvents = result?.events || [];

        // 获取当前 session 状态
        const currentSession = sessions.get(activeSessionId);
        const isLiveSession = currentSession?.status === 'active';

        // 过滤后匹配的总数（从数据库查询结果）
        const filteredTotal = result?.total || 0;

        if (isLiveSession) {
          // Live session: 合并数据库事件 + 内存中尚未持久化的新事件
          const liveEventsArray = liveEvents.getAll();
          const filteredLive = filterEvents(liveEventsArray, filter);

          // 找出内存中有但数据库中还没有的事件（通过 ID 去重）
          const dbEventIds = new Set(dbEvents.map((e: UnifiedEvent) => e.id));
          const newLiveEvents = filteredLive.filter((e: UnifiedEvent) => !dbEventIds.has(e.id));

          // 合并并按时间排序
          const mergedEvents = [...dbEvents, ...newLiveEvents]
            .sort((a, b) => a.relativeTime - b.relativeTime);

          // 限制显示数量（保留最新的 2000 条）
          const displayEvents = mergedEvents.length > 2000
            ? mergedEvents.slice(-2000)
            : mergedEvents;

          // 计算准确的数量
          const newEventsNotInDb = liveEventsArray.filter(e => !dbEventIds.has(e.id)).length;
          const newFilteredNotInDb = newLiveEvents.length;

          set(state => {
            state.visibleEvents = displayEvents;
            state.totalEventCount = sessionTotalCount + newEventsNotInDb;
            state.filteredEventCount = filteredTotal + newFilteredNotInDb;
            state.isLoading = false;
          });
        } else {
          // 非 live session：直接使用数据库结果
          set(state => {
            state.visibleEvents = dbEvents;
            state.totalEventCount = sessionTotalCount;
            state.filteredEventCount = filteredTotal;
            state.isLoading = false;
          });
        }
      } catch (err) {
        console.error('[eventStore] applyFilter error:', err);
        set(state => {
          state.isLoading = false;
        });
      }
    },

    // ========================================
    // 可视区域
    // ========================================

    setVisibleRange: (start, end) => {
      const { loadEventsInRange, visibleRange } = get();

      // 判断是否需要加载更多数据
      const needsLoad = start < visibleRange.start - 10000 || end > visibleRange.end + 10000;

      set(state => {
        state.visibleRange = { start, end };
      });

      if (needsLoad) {
        loadEventsInRange(start - 30000, end + 30000);
      }
    },

    // ========================================
    // 时间索引
    // ========================================

    loadTimeIndex: async (sessionId) => {
      const index = await (window as any).go.main.App.GetSessionTimeIndex(sessionId);

      set(state => {
        state.timeIndex.set(sessionId, index || []);
      });
    },

    // ========================================
    // 书签
    // ========================================

    loadBookmarks: async (sessionId) => {
      const bookmarks = await (window as any).go.main.App.GetSessionBookmarks(sessionId);

      set(state => {
        state.bookmarks.set(sessionId, bookmarks || []);
      });
    },

    createBookmark: async (relativeTime, label, color, type = 'user') => {
      const { activeSessionId, loadBookmarks } = get();
      if (!activeSessionId) return;

      await (window as any).go.main.App.CreateSessionBookmark(
        activeSessionId,
        relativeTime,
        label,
        color || '',
        type
      );

      await loadBookmarks(activeSessionId);
    },

    deleteBookmark: async (bookmarkId) => {
      const { activeSessionId, loadBookmarks } = get();

      await (window as any).go.main.App.DeleteSessionBookmark(bookmarkId);

      if (activeSessionId) {
        await loadBookmarks(activeSessionId);
      }
    },

    // ========================================
    // UI
    // ========================================

    selectEvent: (eventId) => {
      set(state => {
        state.selectedEventId = eventId;
      });
    },

    openTimeline: () => {
      set(state => {
        state.isTimelineOpen = true;
      });
    },

    closeTimeline: () => {
      set(state => {
        state.isTimelineOpen = false;
      });
    },

    setAutoScroll: (enabled) => {
      set(state => {
        state.autoScroll = enabled;
      });
    },

    // ========================================
    // 统计
    // ========================================

    getSessionStats: async (sessionId) => {
      return await (window as any).go.main.App.GetSessionStats(sessionId) || {};
    },

    getSystemStats: async () => {
      return await (window as any).go.main.App.GetEventSystemStats() || {};
    },

    // ========================================
    // 清理
    // ========================================

    clearSession: () => {
      set(state => {
        state.activeSessionId = null;
        state.liveEvents.clear();
        state.pageCache.clear();
        state.visibleEvents = [];
        state.totalEventCount = 0;
        state.selectedEventId = null;
      });
    },

    cleanup: async (maxAgeDays) => {
      const removed = await (window as any).go.main.App.CleanupOldSessionData(maxAgeDays);
      return removed || 0;
    },

    // ========================================
    // 事件订阅
    // ========================================

    subscribeToEvents: () => {
      if (!EventsOn) {
        console.warn('EventsOn not available');
        return () => {};
      }

      const { receiveLiveEvents, loadSessions } = get();

      // 处理新事件批次
      const handleEventsBatch = (events: UnifiedEvent[]) => {
        receiveLiveEvents(events);
      };

      // 处理 session 启动
      const handleSessionStarted = (session: DeviceSession) => {
        set(state => {
          state.sessions.set(session.id, session);
          // 如果是当前设备的 session，自动切换
          if (state.activeDeviceId === session.deviceId) {
            state.activeSessionId = session.id;
            state.liveEvents.clear();
            state.visibleEvents = [];
            state.totalEventCount = 0;
          }
        });
      };

      // 处理 session 结束
      const handleSessionEnded = (session: DeviceSession) => {
        set(state => {
          state.sessions.set(session.id, session);
        });
      };

      // 订阅事件
      EventsOn('events-batch', handleEventsBatch);
      EventsOn('session-started', handleSessionStarted);
      EventsOn('session-ended', handleSessionEnded);

      // 返回取消订阅函数
      return () => {
        if (EventsOff) {
          EventsOff('events-batch');
          EventsOff('session-started');
          EventsOff('session-ended');
        }
      };
    },
  }))
);

// ========================================
// Helper Functions
// ========================================

function filterEvents(events: UnifiedEvent[], filter: EventQuery): UnifiedEvent[] {
  return events.filter(event => {
    if (filter.sources?.length && !filter.sources.includes(event.source)) {
      return false;
    }
    if (filter.categories?.length && !filter.categories.includes(event.category)) {
      return false;
    }
    if (filter.levels?.length && !filter.levels.includes(event.level)) {
      return false;
    }
    if (filter.types?.length && !filter.types.includes(event.type)) {
      return false;
    }
    if (filter.startTime && event.relativeTime < filter.startTime) {
      return false;
    }
    if (filter.endTime && event.relativeTime > filter.endTime) {
      return false;
    }
    if (filter.stepId && event.stepId !== filter.stepId) {
      return false;
    }
    if (filter.searchText) {
      const search = filter.searchText.toLowerCase();
      if (!event.title.toLowerCase().includes(search) &&
          !event.summary?.toLowerCase().includes(search)) {
        return false;
      }
    }
    return true;
  });
}

// ========================================
// Hooks
// ========================================

/**
 * 获取当前 session 的书签
 */
export function useCurrentBookmarks(): Bookmark[] {
  const activeSessionId = useEventStore(state => state.activeSessionId);
  const bookmarks = useEventStore(state => state.bookmarks);

  if (!activeSessionId) return [];
  return bookmarks.get(activeSessionId) || [];
}

/**
 * 获取当前 session 的时间索引
 */
export function useCurrentTimeIndex(): TimeIndexEntry[] {
  const activeSessionId = useEventStore(state => state.activeSessionId);
  const timeIndex = useEventStore(state => state.timeIndex);

  if (!activeSessionId) return [];
  return timeIndex.get(activeSessionId) || [];
}

/**
 * 获取当前 session 信息
 */
export function useCurrentSession(): DeviceSession | null {
  const activeSessionId = useEventStore(state => state.activeSessionId);
  const sessions = useEventStore(state => state.sessions);

  if (!activeSessionId) return null;
  return sessions.get(activeSessionId) || null;
}
