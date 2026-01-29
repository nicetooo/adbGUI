import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';
import { main } from '../types/wails-models';
import type { UINode } from './elementStore';
import {
  StartTouchRecording,
  StopTouchRecording,
  PlayTouchScript,
  StopTouchPlayback,
  LoadTouchScripts,
  SaveTouchScript,
  DeleteTouchScript,
  RenameTouchScript,
  SaveScriptTask,
  LoadScriptTasks,
  DeleteScriptTask,
  RunScriptTask,
  PauseTask,
  ResumeTask,
  StopTask,
} from '../../wailsjs/go/main/App';

const EventsOn = (window as any).runtime.EventsOn;
const EventsOff = (window as any).runtime.EventsOff;

// Re-export types from Wails models for convenience
export type TouchEvent = main.TouchEvent;
export type TouchScript = main.TouchScript;

export interface TaskStep {
  type: string;
  value: string;
  loop: number;
  postDelay?: number;
  checkType?: string;
  checkValue?: string;
  waitTimeout?: number;
  onFailure?: string;
}

export interface ScriptTask {
  name: string;
  steps: TaskStep[];
  createdAt: string;
}

// Re-export UINode from elementStore for backward compatibility
export type { UINode };

interface AutomationState {
  // State
  isRecording: boolean;
  isPlaying: boolean;
  recordingDeviceId: string | null;
  playingDeviceId: string | null;
  recordingStartTime: number | null;
  recordingDuration: number;
  currentScript: main.TouchScript | null;
  scripts: main.TouchScript[];
  tasks: ScriptTask[];
  playbackProgress: { current: number; total: number } | null;
  taskProgress: { 
    stepIndex: number; 
    totalSteps: number; 
    currentLoop: number; 
    totalLoops: number; 
    currentAction?: string 
  } | null;
  isTaskRunning: boolean;
  isPaused: boolean;
  runningTaskName: string | null;

  recordedActionCount: number;
  uiHierarchy: UINode | null;
  rawXml: string | null;
  isFetchingHierarchy: boolean;

  // Recording mode state
  recordingMode: 'fast' | 'precise';
  isWaitingForSelector: boolean;
  pendingSelectorData: any | null;
  isPreCapturing: boolean;
  isAnalyzing: boolean;
  
  // RecordingView UI state
  selectedScriptNames: string[];
  saveModalVisible: boolean;
  scriptName: string;
  renameModalVisible: boolean;
  editingScriptName: string;
  newScriptName: string;
  selectedScript: main.TouchScript | null;

  // Actions
  startRecording: (deviceId: string, mode?: 'fast' | 'precise') => Promise<void>;
  stopRecording: () => Promise<main.TouchScript | null>;
  playScript: (deviceId: string, script: main.TouchScript) => Promise<void>;
  stopPlayback: () => void;
  loadScripts: () => Promise<void>;
  saveScript: (script: main.TouchScript) => Promise<void>;
  deleteScript: (name: string) => Promise<void>;
  deleteScripts: (names: string[]) => Promise<void>;
  renameScript: (oldName: string, newName: string) => Promise<void>;
  setCurrentScript: (script: main.TouchScript | null) => void;
  updateRecordingDuration: () => void;

  // Selector choice actions
  submitSelectorChoice: (selectorType: string, selectorValue: string) => Promise<void>;
  setRecordingMode: (mode: 'fast' | 'precise') => void;

  // Task Actions
  loadTasks: () => Promise<void>;
  saveTask: (task: ScriptTask) => Promise<void>;
  deleteTask: (name: string) => Promise<void>;
  deleteTasks: (names: string[]) => Promise<void>;
  runTask: (deviceId: string, task: ScriptTask) => Promise<void>;
  pauseTask: () => Promise<void>;
  resumeTask: () => Promise<void>;
  stopTask: () => Promise<void>;
  fetchUIHierarchy: (deviceId: string) => Promise<void>;
  checkAndRefreshUIHierarchy: (deviceId: string) => Promise<boolean>;

  // Event subscription
  subscribeToEvents: () => () => void;
  
  // RecordingView UI actions
  setSelectedScriptNames: (names: string[]) => void;
  toggleScriptSelection: (name: string) => void;
  setSaveModalVisible: (visible: boolean) => void;
  setScriptName: (name: string) => void;
  setRenameModalVisible: (visible: boolean) => void;
  setEditingScriptName: (name: string) => void;
  setNewScriptName: (name: string) => void;
  setSelectedScript: (script: main.TouchScript | null) => void;
  openRenameModal: (scriptName: string) => void;
  closeRenameModal: () => void;
  closeSaveModal: () => void;
}

export const useAutomationStore = create<AutomationState>()(
  immer((set, get) => ({
    // Initial state
    isRecording: false,
    isPlaying: false,
    recordingDeviceId: null,
    playingDeviceId: null,
    recordingStartTime: null,
    recordingDuration: 0,
    recordedActionCount: 0,
    currentScript: null,
    scripts: [],
    tasks: [],
    playbackProgress: null,
    taskProgress: null,
    isTaskRunning: false,
    isPaused: false,
    runningTaskName: null,
    uiHierarchy: null,
    rawXml: null,
    isFetchingHierarchy: false,
    
    // Recording mode state
    recordingMode: 'fast',
    isWaitingForSelector: false,
    pendingSelectorData: null,
    isPreCapturing: false,
    isAnalyzing: false,
    
    // RecordingView UI state
    selectedScriptNames: [],
    saveModalVisible: false,
    scriptName: '',
    renameModalVisible: false,
    editingScriptName: '',
    newScriptName: '',
    selectedScript: null,

    // Actions
    startRecording: async (deviceId: string, mode: 'fast' | 'precise' = 'fast') => {
      try {
        await StartTouchRecording(deviceId, mode);
        set((state: AutomationState) => {
          state.isRecording = true;
          state.recordingDeviceId = deviceId;
          state.recordingStartTime = Date.now();
          state.recordingDuration = 0;
          state.recordedActionCount = 0;
          state.currentScript = null;
          state.recordingMode = mode;
          state.isWaitingForSelector = false;
          state.pendingSelectorData = null;
        });
      } catch (err) {
        console.error('Failed to start recording:', err);
        throw err;
      }
    },

    stopRecording: async () => {
      const { recordingDeviceId } = get();
      if (!recordingDeviceId) return null;

      try {
        const script = await StopTouchRecording(recordingDeviceId);
        set((state: AutomationState) => {
          state.isRecording = false;
          state.recordingDeviceId = null;
          state.recordingStartTime = null;
          state.recordingDuration = 0;
          state.recordedActionCount = 0;
          state.currentScript = script;
        });
        return script;
      } catch (err) {
        console.error('Failed to stop recording:', err);
        set((state: AutomationState) => {
          state.isRecording = false;
          state.recordingDeviceId = null;
          state.recordingStartTime = null;
          state.recordingDuration = 0;
          state.recordedActionCount = 0;
        });
        throw err;
      }
    },

    playScript: async (deviceId: string, script: main.TouchScript) => {
      try {
        await PlayTouchScript(deviceId, script);
        set((state: AutomationState) => {
          state.isPlaying = true;
          state.playingDeviceId = deviceId;
          state.playbackProgress = { current: 0, total: script.events?.length || 0 };
        });
      } catch (err) {
        console.error('Failed to play script:', err);
        throw err;
      }
    },

    stopPlayback: () => {
      const { playingDeviceId } = get();
      if (playingDeviceId) {
        StopTouchPlayback(playingDeviceId);
      }
      set((state: AutomationState) => {
        state.isPlaying = false;
        state.playingDeviceId = null;
        state.playbackProgress = null;
      });
    },

    loadScripts: async () => {
      try {
        const scripts = await LoadTouchScripts();
        set((state: AutomationState) => {
          state.scripts = scripts || [];
        });
      } catch (err) {
        console.error('Failed to load scripts:', err);
        set((state: AutomationState) => {
          state.scripts = [];
        });
      }
    },

    saveScript: async (script: main.TouchScript) => {
      try {
        await SaveTouchScript(script);
        await get().loadScripts();
      } catch (err) {
        console.error('Failed to save script:', err);
        throw err;
      }
    },

    deleteScript: async (name: string) => {
      try {
        await DeleteTouchScript(name);
        await get().loadScripts();
      } catch (err) {
        console.error('Failed to delete script:', err);
        throw err;
      }
    },

    deleteScripts: async (names: string[]) => {
      try {
        for (const name of names) {
          await DeleteTouchScript(name);
        }
        await get().loadScripts();
      } catch (err) {
        console.error('Failed to delete scripts:', err);
        throw err;
      }
    },

    renameScript: async (oldName: string, newName: string) => {
      try {
        // Cast to any because RenameTouchScript might be missing in older bindings
        await (RenameTouchScript as any)(oldName, newName);
        await get().loadScripts();
      } catch (err) {
        console.error('Failed to rename script:', err);
        throw err;
      }
    },

    loadTasks: async () => {
      try {
        const tasks = await LoadScriptTasks();
        set((state: AutomationState) => {
          state.tasks = (tasks as unknown as ScriptTask[]) || [];
        });
      } catch (err) {
        console.error('Failed to load tasks:', err);
        set((state: AutomationState) => {
          state.tasks = [];
        });
      }
    },

    saveTask: async (task: ScriptTask) => {
      try {
        await SaveScriptTask(task as any);
        await get().loadTasks();
      } catch (err) {
        console.error('Failed to save task:', err);
        throw err;
      }
    },

    deleteTask: async (name: string) => {
      try {
        await DeleteScriptTask(name);
        await get().loadTasks();
      } catch (err) {
        console.error('Failed to delete task:', err);
        throw err;
      }
    },

    deleteTasks: async (names: string[]) => {
      try {
        for (const name of names) {
          await DeleteScriptTask(name);
        }
        await get().loadTasks();
      } catch (err) {
        console.error('Failed to delete tasks:', err);
        throw err;
      }
    },

    runTask: async (deviceId: string, task: ScriptTask) => {
      try {
        await RunScriptTask(deviceId, task as any);
        set((state: AutomationState) => {
          state.isTaskRunning = true;
          state.isPaused = false;
          state.runningTaskName = task.name;
        });
      } catch (err) {
        console.error('Failed to run task:', err);
        throw err;
      }
    },

    pauseTask: async () => {
      const deviceId = get().playingDeviceId;
      if (deviceId) {
        await (PauseTask as any)(deviceId);
        set((state: AutomationState) => {
          state.isPaused = true;
        });
      }
    },

    resumeTask: async () => {
      const deviceId = get().playingDeviceId;
      if (deviceId) {
        await (ResumeTask as any)(deviceId);
        set((state: AutomationState) => {
          state.isPaused = false;
        });
      }
    },

    stopTask: async () => {
      const deviceId = get().playingDeviceId;
      if (deviceId) {
        await (StopTask as any)(deviceId);
        set((state: AutomationState) => {
          state.isTaskRunning = false;
          state.isPaused = false;
          state.runningTaskName = null;
          state.taskProgress = null;
        });
      }
    },

    fetchUIHierarchy: async (deviceId: string) => {
      if (!deviceId) return;
      set((state: AutomationState) => {
        state.isFetchingHierarchy = true;
      });
      try {
        // Delegate to elementStore for shared caching
        const { useElementStore } = await import('./elementStore');
        const elementStore = useElementStore.getState();
        const hierarchy = await elementStore.fetchHierarchy(deviceId, true);
        // Sync state for backward compatibility
        set((state: AutomationState) => {
          state.uiHierarchy = hierarchy as unknown as UINode;
          state.rawXml = elementStore.rawXml;
          state.isFetchingHierarchy = false;
        });
      } catch (err) {
        console.error('Failed to fetch UI hierarchy:', err);
        set((state: AutomationState) => {
          state.isFetchingHierarchy = false;
        });
        throw err;
      }
    },

    checkAndRefreshUIHierarchy: async (deviceId: string) => {
      if (!deviceId) return false;
      try {
        // Delegate to elementStore for shared caching
        const { useElementStore } = await import('./elementStore');
        const elementStore = useElementStore.getState();
        const oldRawXml = get().rawXml;
        await elementStore.fetchHierarchy(deviceId, false);
        if (elementStore.rawXml !== oldRawXml) {
          set((state: AutomationState) => {
            state.uiHierarchy = elementStore.hierarchy as unknown as UINode;
            state.rawXml = elementStore.rawXml;
          });
          return true;
        }
        return false;
      } catch (err) {
        console.error('Background UI refresh failed:', err);
        return false;
      }
    },

    setCurrentScript: (script: main.TouchScript | null) => {
      set((state: AutomationState) => {
        state.currentScript = script;
      });
    },

    updateRecordingDuration: () => {
      const { isRecording, recordingStartTime } = get();
      if (isRecording && recordingStartTime) {
        const duration = Math.floor((Date.now() - recordingStartTime) / 1000);
        set((state: AutomationState) => {
          state.recordingDuration = duration;
        });
      }
    },

    submitSelectorChoice: async (selectorType: string, selectorValue: string) => {
      const { recordingDeviceId } = get();
      if (!recordingDeviceId) {
        throw new Error('No active recording');
      }
      
      try {
        await (window as any).go.main.App.SubmitSelectorChoice(
          recordingDeviceId,
          selectorType,
          selectorValue
        );
        set((state: AutomationState) => {
          state.isWaitingForSelector = false;
          state.pendingSelectorData = null;
        });
      } catch (err) {
        console.error('Failed to submit selector choice:', err);
        throw err;
      }
    },

    setRecordingMode: (mode: 'fast' | 'precise') => {
      set((state: AutomationState) => {
        state.recordingMode = mode;
      });
    },
    
    // RecordingView UI actions
    setSelectedScriptNames: (names: string[]) => {
      set((state: AutomationState) => {
        state.selectedScriptNames = names;
      });
    },
    
    toggleScriptSelection: (name: string) => {
      set((state: AutomationState) => {
        const index = state.selectedScriptNames.indexOf(name);
        if (index > -1) {
          state.selectedScriptNames.splice(index, 1);
        } else {
          state.selectedScriptNames.push(name);
        }
      });
    },
    
    setSaveModalVisible: (visible: boolean) => {
      set((state: AutomationState) => {
        state.saveModalVisible = visible;
      });
    },
    
    setScriptName: (name: string) => {
      set((state: AutomationState) => {
        state.scriptName = name;
      });
    },
    
    setRenameModalVisible: (visible: boolean) => {
      set((state: AutomationState) => {
        state.renameModalVisible = visible;
      });
    },
    
    setEditingScriptName: (name: string) => {
      set((state: AutomationState) => {
        state.editingScriptName = name;
      });
    },
    
    setNewScriptName: (name: string) => {
      set((state: AutomationState) => {
        state.newScriptName = name;
      });
    },
    
    setSelectedScript: (script: main.TouchScript | null) => {
      set((state: AutomationState) => {
        state.selectedScript = script;
      });
    },
    
    openRenameModal: (scriptName: string) => {
      set((state: AutomationState) => {
        state.editingScriptName = scriptName;
        state.newScriptName = scriptName;
        state.renameModalVisible = true;
      });
    },
    
    closeRenameModal: () => {
      set((state: AutomationState) => {
        state.renameModalVisible = false;
        state.editingScriptName = '';
        state.newScriptName = '';
      });
    },
    
    closeSaveModal: () => {
      set((state: AutomationState) => {
        state.saveModalVisible = false;
        state.scriptName = '';
      });
    },

    // Event subscription
    subscribeToEvents: () => {
      const handleRecordStarted = (data: any) => {
        set((state: AutomationState) => {
          state.isRecording = true;
          state.recordingDeviceId = data.deviceId;
          state.recordingStartTime = data.startTime * 1000;
          state.recordedActionCount = 0;
        });
      };

      const handleRecordStopped = (data: any) => {
        set((state: AutomationState) => {
          state.isRecording = false;
          state.recordingDeviceId = null;
          state.recordingStartTime = null;
          state.recordingDuration = 0;
          state.recordedActionCount = 0;
        });
      };

      const handleTouchActionRecorded = (data: any) => {
        set((state: AutomationState) => {
          state.recordedActionCount += 1;
        });
      };

      const handlePlaybackStarted = (data: any) => {
        set((state: AutomationState) => {
          state.isPlaying = true;
          state.playingDeviceId = data.deviceId;
          state.playbackProgress = { current: 0, total: data.total };
        });
      };

      const handlePlaybackProgress = (data: any) => {
        set((state: AutomationState) => {
          if (state.playbackProgress) {
            state.playbackProgress.current = data.current;
            state.playbackProgress.total = data.total;
          }
        });
      };

      const handlePlaybackCompleted = (data: any) => {
        set((state: AutomationState) => {
          state.isPlaying = false;
          state.playingDeviceId = null;
          state.playbackProgress = null;
        });
      };

      const handleTaskStarted = (data: any) => {
        set((state: AutomationState) => {
          state.isTaskRunning = true;
          state.isPaused = false;
          state.runningTaskName = data.taskName;
          state.playingDeviceId = data.deviceId;
        });
      };

      const handleTaskCompleted = (data: any) => {
        set((state: AutomationState) => {
          state.isTaskRunning = false;
          state.isPaused = false;
          state.runningTaskName = null;
          state.taskProgress = null;
          state.playingDeviceId = null;
        });
      };

      const handleTaskPaused = (data: any) => set((state: AutomationState) => {
        state.isPaused = true;
      });
      const handleTaskResumed = (data: any) => set((state: AutomationState) => {
        state.isPaused = false;
      });

      const handleTaskStepRunning = (data: any) => {
        set((state: AutomationState) => {
          state.taskProgress = {
            stepIndex: data.stepIndex,
            totalSteps: data.totalSteps,
            currentLoop: data.currentLoop,
            totalLoops: data.totalLoops,
            currentAction: data.currentAction || (data.type === 'wait' ? `Wait ${data.value}ms` : `Running: ${data.value}`),
          };
        });
      };

      const handleRecordingPausedForSelector = (data: any) => {
        set((state: AutomationState) => {
          state.isWaitingForSelector = true;
          state.pendingSelectorData = data;
          state.isAnalyzing = false;
        });
      };

      const handleRecordingResumed = (data: any) => {
        set((state: AutomationState) => {
          state.isWaitingForSelector = false;
          state.pendingSelectorData = null;
          state.isAnalyzing = false;
        });
      };

      const offPreCaptureStarted = EventsOn('recording-pre-capture-started', (data: any) => {
        set((state: AutomationState) => {
          state.isPreCapturing = true;
        });
      });

      const offPreCaptureFinished = EventsOn('recording-pre-capture-finished', (data: any) => {
        set((state: AutomationState) => {
          state.isPreCapturing = false;
        });
      });

      const offAnalysisStarted = EventsOn('recording-analysis-started', (data: any) => {
        set((state: AutomationState) => {
          state.isAnalyzing = true;
        });
      });

      EventsOn('touch-record-started', handleRecordStarted);
      EventsOn('touch-record-stopped', handleRecordStopped);
      EventsOn('touch-action-recorded', handleTouchActionRecorded);
      EventsOn('touch-playback-started', handlePlaybackStarted);
      EventsOn('touch-playback-progress', handlePlaybackProgress);
      EventsOn('touch-playback-completed', handlePlaybackCompleted);
      EventsOn('task-started', handleTaskStarted);
      EventsOn('task-completed', handleTaskCompleted);
      EventsOn('task-step-running', handleTaskStepRunning);
      EventsOn('task-paused', handleTaskPaused);
      EventsOn('task-resumed', handleTaskResumed);
      EventsOn('recording-paused-for-selector', handleRecordingPausedForSelector);
      EventsOn('recording-resumed', handleRecordingResumed);

      return () => {
        EventsOff('touch-record-started');
        EventsOff('touch-record-stopped');
        EventsOff('touch-action-recorded');
        EventsOff('touch-playback-started');
        EventsOff('touch-playback-progress');
        EventsOff('touch-playback-completed');
        EventsOff('task-started');
        EventsOff('task-completed');
        EventsOff('task-step-running');
        EventsOff('recording-paused-for-selector');
        EventsOff('recording-resumed');
      };
    },
  }))
);
