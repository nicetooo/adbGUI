import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

const EventsOn = (window as any).runtime.EventsOn;
const EventsOff = (window as any).runtime.EventsOff;

// ========================================
// Unified Session Store
// Cross-module event correlation and timeline tracking
// ========================================

export interface SessionEvent {
  id: string;
  sessionId: string;
  deviceId: string;
  timestamp: number; // Unix milliseconds
  type: string;      // e.g., "workflow_step_start", "logcat", "network_request"
  category: string;  // "workflow", "log", "network", "automation", "system"
  level: string;     // "info", "warn", "error", "debug", "verbose"
  title: string;
  detail?: any;
  stepId?: string;
  duration?: number;
  success?: boolean;
}

export interface Session {
  id: string;
  deviceId: string;
  type: string;      // "workflow", "recording", "debug", "manual"
  name: string;
  startTime: number;
  endTime: number;   // 0 if active
  status: string;    // "active", "completed", "failed", "cancelled"
  eventCount: number;
  metadata?: Record<string, any>;
}

export interface SessionFilter {
  categories?: string[];
  types?: string[];
  levels?: string[];
  stepId?: string;
  startTime?: number;
  endTime?: number;
  searchText?: string;
}

// Category colors for timeline display
export const categoryColors: Record<string, string> = {
  workflow: '#1890ff',   // Blue
  log: '#52c41a',        // Green
  network: '#722ed1',    // Purple
  automation: '#fa8c16', // Orange
  system: '#8c8c8c',     // Gray
};

// Level icons/colors
export const levelStyles: Record<string, { color: string; icon: string }> = {
  error: { color: '#ff4d4f', icon: '❌' },
  warn: { color: '#faad14', icon: '⚠️' },
  info: { color: '#1890ff', icon: 'ℹ️' },
  debug: { color: '#8c8c8c', icon: '🔧' },
  verbose: { color: '#bfbfbf', icon: '📝' },
};

// Event type icons
export const eventTypeIcons: Record<string, string> = {
  session_start: '🚀',
  session_end: '🏁',
  workflow_start: '▶️',
  workflow_step_start: '🔷',
  workflow_step_end: '✅',
  workflow_step_error: '❌',
  workflow_step_cancelled: '⏹️',
  workflow_completed: '🏁',
  workflow_error: '❌',
  logcat: '📝',
  network_request: '🌐',
  network_response: '📥',
  ui_action: '👆',
  screenshot: '📸',
};

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
      
      // Deduplicate timeline (Network events may have updates)
      const dedupedTimeline: SessionEvent[] = [];
      const idMap = new Map<string, number>(); // RequestId -> Index

      (timeline || []).forEach((event: SessionEvent) => {
        const reqId = event.category === 'network' ? (event.detail as any)?.id : undefined;
        if (reqId) {
          if (idMap.has(reqId)) {
             // Replace existing event with newer version
             dedupedTimeline[idMap.get(reqId)!] = event;
          } else {
             dedupedTimeline.push(event);
             idMap.set(reqId, dedupedTimeline.length - 1);
          }
        } else {
          dedupedTimeline.push(event);
        }
      });

      set((state: SessionState) => {
        state.timeline = dedupedTimeline;
        state.activeSessionId = sessionId;
      });
    },

    // Clear timeline
    clearTimeline: () => {
      set((state: SessionState) => {
        state.timeline = [];
        state.activeSessionId = null;
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
                      // Update existing event
                      state.timeline[existingIndex] = newEvent;
                  } else {
                      state.timeline.push(newEvent);
                  }
               } else {
                   // Not a network event or no ID, just push
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
