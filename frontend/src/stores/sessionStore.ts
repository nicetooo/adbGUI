import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

// Import types from eventTypes (new unified types)
import {
  UnifiedEvent,
  DeviceSession,
  SessionFilter,
  categoryConfig,
  levelConfig,
  eventTypeConfig,
  formatTimestamp,
  formatDuration as formatDurationUtil,
  getEventIcon as getEventIconUtil,
  getEventColor as getEventColorUtil,
} from './eventTypes';

// Re-export types for backward compatibility
export type { UnifiedEvent, DeviceSession, SessionFilter };

// Legacy type aliases
export type SessionEvent = UnifiedEvent;
export type Session = DeviceSession;

const EventsOn = (window as any).runtime?.EventsOn;
const EventsOff = (window as any).runtime?.EventsOff;

// ========================================
// Unified Session Store
// Cross-module event correlation and timeline tracking
// (Legacy store - maintained for backward compatibility)
// For new code, prefer using eventStore.ts
// ========================================

// Category colors for timeline display (legacy - use categoryConfig from eventTypes)
export const categoryColors: Record<string, string> = {
  workflow: categoryConfig.automation?.color || '#1890ff',
  log: categoryConfig.log?.color || '#52c41a',
  network: categoryConfig.network?.color || '#722ed1',
  automation: categoryConfig.automation?.color || '#fa8c16',
  system: categoryConfig.diagnostic?.color || '#8c8c8c',
  state: categoryConfig.state?.color || '#1890ff',
  interaction: categoryConfig.interaction?.color || '#fa8c16',
  diagnostic: categoryConfig.diagnostic?.color || '#8c8c8c',
};

// Level icons/colors (legacy - use levelConfig from eventTypes)
export const levelStyles: Record<string, { color: string; icon: string }> = {
  error: { color: levelConfig.error.color, icon: levelConfig.error.icon },
  warn: { color: levelConfig.warn.color, icon: levelConfig.warn.icon },
  info: { color: levelConfig.info.color, icon: levelConfig.info.icon },
  debug: { color: levelConfig.debug.color, icon: levelConfig.debug.icon },
  verbose: { color: levelConfig.verbose.color, icon: levelConfig.verbose.icon },
  fatal: { color: levelConfig.fatal.color, icon: levelConfig.fatal.icon },
};

// Event type icons (legacy - use eventTypeConfig from eventTypes)
export const eventTypeIcons: Record<string, string> = Object.fromEntries(
  Object.entries(eventTypeConfig).map(([key, value]) => [key, value.icon])
);

interface SessionState {
  // Active sessions
  sessions: Map<string, Session>;
  activeSessionId: string | null;

  // Current session timeline
  timeline: SessionEvent[];

  // Filter state
  filter: SessionFilter;

  // UI state
  isTimelineOpen: boolean;
  selectedEventId: string | null;

  // Actions
  startSession: (deviceId: string, type: string, name: string) => Promise<string>;
  endSession: (sessionId: string, status?: string) => Promise<void>;

  // Timeline management
  loadTimeline: (sessionId: string, filter?: SessionFilter) => Promise<void>;
  clearTimeline: () => void;

  // Filter
  setFilter: (filter: Partial<SessionFilter>) => void;
  getFilteredTimeline: () => SessionEvent[];

  // UI
  openTimeline: () => void;
  closeTimeline: () => void;
  selectEvent: (eventId: string | null) => void;

  // Session queries
  getSessions: (deviceId?: string, limit?: number) => Promise<Session[]>;
  getActiveSession: (deviceId: string) => Promise<string | null>;

  // Event subscription
  subscribeToEvents: () => () => void;
}

export const useSessionStore = create<SessionState>()(
  immer((set, get) => ({
    // Initial state
    sessions: new Map(),
    activeSessionId: null,
    timeline: [],
    filter: {},
    isTimelineOpen: false,
    selectedEventId: null,

    // Start a new session
    startSession: async (deviceId: string, type: string, name: string) => {
      const sessionId = await (window as any).go.main.App.CreateSession(deviceId, type, name);
      set((state: SessionState) => {
        state.activeSessionId = sessionId;
      });
      return sessionId;
    },

    // End a session
    endSession: async (sessionId: string, status = 'completed') => {
      await (window as any).go.main.App.EndSession(sessionId, status);
      const { activeSessionId } = get();
      if (activeSessionId === sessionId) {
        set((state: SessionState) => {
          state.activeSessionId = null;
        });
      }
    },

    // Load timeline for a session
    loadTimeline: async (sessionId: string, filter?: SessionFilter) => {
      const timeline = await (window as any).go.main.App.GetSessionTimeline(sessionId, filter || null);
      
      set((state: SessionState) => {
        state.timeline = timeline || [];
        state.activeSessionId = sessionId;
      });
    },

    // Clear timeline
    clearTimeline: () => {
      set((state: SessionState) => {
        state.timeline = [];
      });
    },

    // Set filter
    setFilter: (newFilter: Partial<SessionFilter>) => {
      set((state: SessionState) => {
        state.filter = { ...state.filter, ...newFilter };
      });
    },

    // Get filtered timeline
    getFilteredTimeline: () => {
      const { timeline, filter } = get();

      return timeline.filter((event) => {
        if (filter.categories?.length && !filter.categories.includes(event.category)) {
          return false;
        }
        if (filter.types?.length && !filter.types.includes(event.type)) {
          return false;
        }
        if (filter.levels?.length && !filter.levels.includes(event.level)) {
          return false;
        }
        if (filter.stepId && event.stepId !== filter.stepId) {
          return false;
        }
        if (filter.startTime && event.timestamp < filter.startTime) {
          return false;
        }
        if (filter.endTime && event.timestamp > filter.endTime) {
          return false;
        }
        if (filter.searchText) {
           const searchLower = filter.searchText.toLowerCase();
           return event.title.toLowerCase().includes(searchLower) || 
                  (event.detail && JSON.stringify(event.detail).toLowerCase().includes(searchLower));
        }
        return true;
      });
    },

    // UI actions
    openTimeline: () => set({ isTimelineOpen: true }),
    closeTimeline: () => set({ isTimelineOpen: false }),
    selectEvent: (eventId: string | null) => set({ selectedEventId: eventId }),

    // Get sessions
    getSessions: async (deviceId?: string, limit?: number) => {
      const sessions = await (window as any).go.main.App.GetSessions(deviceId || '', limit || 0);
      return sessions || [];
    },

    // Get active session for device
    getActiveSession: async (deviceId: string) => {
      return await (window as any).go.main.App.GetActiveSession(deviceId);
    },

    // Subscribe to real-time events
    subscribeToEvents: () => {
      // Handle batch of session events (unified event source)
      const handleSessionEventsBatch = (events: SessionEvent[]) => {
        const { activeSessionId } = get();
        
        if (events.length > 0) {
            // Check for mismatch
            // const firstSessionId = events[0].sessionId;
            // if (activeSessionId && firstSessionId !== activeSessionId) {
            //     console.warn(`[SessionStore] Event mismatch! Active: ${activeSessionId}, Event: ${firstSessionId}`);
            // }
        }

        // Only add to timeline if it's for the active session
        const matchingEvents = events.filter(e => e.sessionId === activeSessionId);
        
        if (matchingEvents.length > 0) {
          set((state: SessionState) => {
            matchingEvents.forEach(newEvent => {
               // Check if it's a network event update
               const reqId = newEvent.category === 'network' ? (newEvent.detail as any)?.id : undefined;
               
               if (reqId) {
                  // Find existing event index
                  const existingIndex = state.timeline.findIndex(e => 
                    e.category === 'network' && (e.detail as any)?.id === reqId
                  );
                  
                  if (existingIndex !== -1) {
                      state.timeline[existingIndex] = newEvent;
                  } else {
                      state.timeline.push(newEvent);
                  }
               } else {
                   // Default push
                   state.timeline.push(newEvent);
               }
            });
          });
        }
      };

      // Handle session started
      const handleSessionStarted = (session: Session) => {
        set((state: SessionState) => {
          state.sessions.set(session.id, session);
          state.activeSessionId = session.id;
        });
      };

      // Handle session ended
      const handleSessionEnded = (session: Session) => {
        set((state: SessionState) => {
          state.sessions.set(session.id, session);
          if (state.activeSessionId === session.id) {
            state.activeSessionId = null;
          }
        });
      };

      EventsOn('session-events-batch', handleSessionEventsBatch);
      EventsOn('session-started', handleSessionStarted);
      EventsOn('session-ended', handleSessionEnded);

      return () => {
        EventsOff('session-events-batch');
        EventsOff('session-started');
        EventsOff('session-ended');
      };
    },
  }))
);

export function formatEventTime(timestamp: number): string {
  const date = new Date(timestamp);
  return date.toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    fractionalSecondDigits: 3,
  });
}

export function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
  const mins = Math.floor(ms / 60000);
  const secs = Math.floor((ms % 60000) / 1000);
  return `${mins}m ${secs}s`;
}

export function getEventIcon(event: SessionEvent): string {
  return eventTypeIcons[event.type] || levelStyles[event.level]?.icon || '•';
}

export function getEventColor(event: SessionEvent): string {
  // Priority: level color > category color
  if (event.level === 'error') return levelStyles.error.color;
  if (event.level === 'warn') return levelStyles.warn.color;
  return categoryColors[event.category] || '#8c8c8c';
}
