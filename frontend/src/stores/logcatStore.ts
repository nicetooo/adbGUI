import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';
import { StartLogcat, StopLogcat } from '../../wailsjs/go/main/App';
// @ts-ignore
import { main } from '../types/wails-models';

export interface FilterPreset {
  id: string;
  name: string;
  pattern: string;
  isRegex: boolean;
}

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

  // Persistent Filter Presets Selection
  selectedPresetIds: string[];
  selectedExcludePresetIds: string[];

  // UI State
  packages: main.AppPackage[];
  appRunningStatus: boolean;
  autoScroll: boolean;
  levelFilter: string[];
  matchCase: boolean;
  matchWholeWord: boolean;
  savedFilters: FilterPreset[];

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

  setSelectedPresetIds: (ids: string[] | ((prev: string[]) => string[])) => void;
  setSelectedExcludePresetIds: (ids: string[] | ((prev: string[]) => string[])) => void;
  
  setPackages: (packages: main.AppPackage[]) => void;
  setAppRunningStatus: (status: boolean) => void;
  setAutoScroll: (autoScroll: boolean) => void;
  setLevelFilter: (filter: string[]) => void;
  setMatchCase: (match: boolean) => void;
  setMatchWholeWord: (match: boolean) => void;
  setSavedFilters: (filters: FilterPreset[] | ((prev: FilterPreset[]) => FilterPreset[])) => void;

  // Logcat control
  toggleLogcat: (deviceId: string, pkg: string) => Promise<void>;
  stopLogcat: () => void;

  // Reset for device change
  reset: () => void;
}

// Local buffer to throttle updates (prevents 1000s of state updates per second)
let logBuffer: string[] = [];
let flushTimerId: number | null = null;

const MAX_LOGS = 50000; // Reduced from 200k to 50k for performance

export const useLogcatStore = create<LogcatState>()(
  immer((set, get) => ({
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

    selectedPresetIds: [],
    selectedExcludePresetIds: [],

    // UI State Init
    packages: [],
    appRunningStatus: false,
    autoScroll: true,
    levelFilter: [],
    matchCase: false,
    matchWholeWord: false,
    savedFilters: (() => {
      try {
        const saved = localStorage.getItem("adbGUI_logcat_filters");
        return saved ? JSON.parse(saved) : [];
      } catch {
        return [];
      }
    })(),

    // Actions
    setLogs: (logsOrUpdater) => {
      set((state: LogcatState) => {
        const nextLogs = typeof logsOrUpdater === 'function'
          ? logsOrUpdater(state.logs)
          : logsOrUpdater;
        
        if (nextLogs.length > MAX_LOGS) {
          state.logs = nextLogs.slice(-MAX_LOGS);
        } else {
          state.logs = nextLogs;
        }
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

    setSelectedPresetIds: (idsOrUpdater) => set((state: LogcatState) => {
       state.selectedPresetIds = typeof idsOrUpdater === 'function' 
         ? idsOrUpdater(state.selectedPresetIds)
         : idsOrUpdater;
    }),
    
    setSelectedExcludePresetIds: (idsOrUpdater) => set((state: LogcatState) => {
       state.selectedExcludePresetIds = typeof idsOrUpdater === 'function' 
         ? idsOrUpdater(state.selectedExcludePresetIds)
         : idsOrUpdater;
    }),

    setPackages: (packages) => set({ packages }),
    setAppRunningStatus: (status) => set({ appRunningStatus: status }),
    setAutoScroll: (autoScroll) => set({ autoScroll }),
    setLevelFilter: (filter) => set({ levelFilter: filter }),
    setMatchCase: (match) => set({ matchCase: match }),
    setMatchWholeWord: (match) => set({ matchWholeWord: match }),

    setSavedFilters: (filtersOrUpdater) => set((state: LogcatState) => {
      const nextFilters = typeof filtersOrUpdater === 'function'
        ? filtersOrUpdater(state.savedFilters)
        : filtersOrUpdater;
      state.savedFilters = nextFilters;
      localStorage.setItem("adbGUI_logcat_filters", JSON.stringify(nextFilters));
    }),

    toggleLogcat: async (deviceId: string, pkg: string) => {
      const { isLogging } = get();

      if (isLogging) {
        get().stopLogcat();
      } else {
        set((state: LogcatState) => {
          state.logs = [];
          state.isLogging = true;
        });
        
        // Reset buffer
        logBuffer = [];

        // Flush logs every 100ms
        flushTimerId = window.setInterval(() => {
          if (logBuffer.length > 0) {
            const chunk = [...logBuffer];
            logBuffer = []; // Clear local buffer
            
            set((state: LogcatState) => {
              if (chunk.length > MAX_LOGS) {
                 // XOR swap or just replace if chunk is huge
                 state.logs = chunk.slice(-MAX_LOGS);
              } else {
                state.logs.push(...chunk);
                if (state.logs.length > MAX_LOGS) {
                  state.logs = state.logs.slice(-MAX_LOGS);
                }
              }
            });
          }
        }, 100);

        // Subscribe to session events batch (unified event source)
        EventsOn('session-events-batch', (events: any[]) => {
          // Filter for log events only
          const logEvents = events.filter((e: any) => e.category === 'log');
          if (logEvents.length === 0) return;

          // Access current filter state using get()
          const { preFilter, preUseRegex, excludeFilter, excludeUseRegex } = get();

          const filteredLines: string[] = [];

          for (const event of logEvents) {
            let linesToProcess: string[] = [];
            
            if (Array.isArray(event.detail)) {
                // Aggregated event from backend
                linesToProcess = event.detail.map((d: any) => d.raw).filter(Boolean);
            } else if (event.detail && event.detail.raw) {
                // Single event (legacy)
                linesToProcess = [event.detail.raw];
            } else {
                // Fallback for title-only events
                linesToProcess = [event.title];
            }

            for (const line of linesToProcess) {
                if (!line) continue;

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
          }

          if (filteredLines.length > 0) {
             // Push to local buffer, NO state update here
             logBuffer.push(...filteredLines);
          }
        });

        try {
          const { preFilter, preUseRegex, excludeFilter, excludeUseRegex } = get();
          await StartLogcat(deviceId, pkg, preFilter, preUseRegex, excludeFilter, excludeUseRegex);
        } catch (err) {
          get().stopLogcat();
          throw err;
        }
      }
    },

    stopLogcat: () => {
      StopLogcat();
      EventsOff('session-events-batch');

      if (flushTimerId) {
        clearInterval(flushTimerId);
        flushTimerId = null;
      }
      logBuffer = [];

      set((state: LogcatState) => {
        state.isLogging = false;
      });
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
        selectedPresetIds: [],
        selectedExcludePresetIds: [],
        packages: [],
        appRunningStatus: false,
        autoScroll: true,
        levelFilter: [],
        matchCase: false,
        matchWholeWord: false,
      });
    },
  }))
);
