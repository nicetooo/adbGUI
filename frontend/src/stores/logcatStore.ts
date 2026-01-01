import { create } from 'zustand';
import { StartLogcat, StopLogcat } from '../../wailsjs/go/main/App';

// @ts-ignore
const EventsOn = (window as any).runtime.EventsOn;
// @ts-ignore
const EventsOff = (window as any).runtime.EventsOff;

interface LogcatState {
  // State
  logs: string[];
  isLogging: boolean;
  selectedPackage: string;

  // Filter states
  logFilter: string;
  useRegex: boolean;
  preFilter: string;
  preUseRegex: boolean;
  excludeFilter: string;
  excludeUseRegex: boolean;

  // Internal buffer (replaces useRef)
  logBuffer: string[];
  flushTimerId: number | null;

  // Actions
  setLogs: (logs: string[] | ((prev: string[]) => string[])) => void;
  clearLogs: () => void;
  setSelectedPackage: (pkg: string) => void;
  setLogFilter: (filter: string) => void;
  setUseRegex: (useRegex: boolean) => void;
  setPreFilter: (filter: string) => void;
  setPreUseRegex: (useRegex: boolean) => void;
  setExcludeFilter: (filter: string) => void;
  setExcludeUseRegex: (useRegex: boolean) => void;

  // Logcat control
  toggleLogcat: (deviceId: string, pkg: string) => Promise<void>;
  stopLogcat: () => void;

  // Reset for device change
  reset: () => void;
}

const MAX_LOGS = 200000;

export const useLogcatStore = create<LogcatState>((set, get) => ({
  // Initial state
  logs: [],
  isLogging: false,
  selectedPackage: '',

  logFilter: '',
  useRegex: false,
  preFilter: '',
  preUseRegex: false,
  excludeFilter: '',
  excludeUseRegex: false,

  logBuffer: [],
  flushTimerId: null,

  // Actions
  setLogs: (logsOrUpdater) => {
    set(state => {
      const newLogs = typeof logsOrUpdater === 'function'
        ? logsOrUpdater(state.logs)
        : logsOrUpdater;
      return { logs: newLogs.length > MAX_LOGS ? newLogs.slice(-MAX_LOGS) : newLogs };
    });
  },

  clearLogs: () => set({ logs: [] }),

  setSelectedPackage: (pkg) => set({ selectedPackage: pkg }),

  setLogFilter: (filter) => set({ logFilter: filter }),
  setUseRegex: (useRegex) => set({ useRegex }),
  setPreFilter: (filter) => set({ preFilter: filter }),
  setPreUseRegex: (useRegex) => set({ preUseRegex: useRegex }),
  setExcludeFilter: (filter) => set({ excludeFilter: filter }),
  setExcludeUseRegex: (useRegex) => set({ excludeUseRegex: useRegex }),

  toggleLogcat: async (deviceId: string, pkg: string) => {
    const { isLogging, preFilter, preUseRegex, excludeFilter, excludeUseRegex } = get();

    if (isLogging) {
      get().stopLogcat();
    } else {
      set({ logs: [], isLogging: true, logBuffer: [] });

      // Flush logs every 100ms
      const timerId = window.setInterval(() => {
        const { logBuffer } = get();
        if (logBuffer.length > 0) {
          const chunk = [...logBuffer];
          set({ logBuffer: [] });

          set(state => {
            const next = [...state.logs, ...chunk];
            return { logs: next.length > MAX_LOGS ? next.slice(-MAX_LOGS) : next };
          });
        }
      }, 100);

      set({ flushTimerId: timerId });

      // Subscribe to logcat events
      EventsOn('logcat-data', (data: string | string[]) => {
        const lines = Array.isArray(data) ? data : [data];

        // Access current filter state using get()
        const { preFilter, preUseRegex, excludeFilter, excludeUseRegex, logBuffer } = get();

        const filteredLines: string[] = [];

        for (const line of lines) {
          let shouldKeep = true;

          // 1. Pre-filter (Include)
          if (preFilter.trim()) {
            try {
              if (preUseRegex) {
                const regex = new RegExp(preFilter, 'i');
                if (!regex.test(line)) shouldKeep = false;
              } else {
                if (!line.toLowerCase().includes(preFilter.toLowerCase())) shouldKeep = false;
              }
            } catch {
              if (!line.toLowerCase().includes(preFilter.toLowerCase())) shouldKeep = false;
            }
          }

          // 2. Exclude filter (Negative)
          if (shouldKeep && excludeFilter.trim()) {
            try {
              if (excludeUseRegex) {
                const regex = new RegExp(excludeFilter, 'i');
                if (regex.test(line)) shouldKeep = false;
              } else {
                if (line.toLowerCase().includes(excludeFilter.toLowerCase())) shouldKeep = false;
              }
            } catch {
              if (line.toLowerCase().includes(excludeFilter.toLowerCase())) shouldKeep = false;
            }
          }

          if (shouldKeep) {
            filteredLines.push(line);
          }
        }

        if (filteredLines.length > 0) {
          set({ logBuffer: [...logBuffer, ...filteredLines] });
        }
      });

      try {
        await StartLogcat(deviceId, pkg, preFilter, preUseRegex, excludeFilter, excludeUseRegex);
      } catch (err) {
        get().stopLogcat();
        throw err;
      }
    }
  },

  stopLogcat: () => {
    const { flushTimerId } = get();

    StopLogcat();
    EventsOff('logcat-data');

    if (flushTimerId) {
      clearInterval(flushTimerId);
    }

    set({ isLogging: false, logBuffer: [], flushTimerId: null });
  },

  reset: () => {
    const { isLogging } = get();
    if (isLogging) {
      get().stopLogcat();
    }
    set({
      logs: [],
      selectedPackage: '',
      logFilter: '',
      useRegex: false,
      preFilter: '',
      preUseRegex: false,
      excludeFilter: '',
      excludeUseRegex: false,
    });
  },
}));
