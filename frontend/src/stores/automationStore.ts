import { create } from 'zustand';
import { main } from '../../wailsjs/go/models';
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
}

export interface ScriptTask {
  name: string;
  steps: TaskStep[];
  createdAt: string;
}

export type ScriptTaskModel = main.ScriptTask;

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

  // Actions
  startRecording: (deviceId: string) => Promise<void>;
  stopRecording: () => Promise<main.TouchScript | null>;
  playScript: (deviceId: string, script: main.TouchScript) => Promise<void>;
  stopPlayback: () => void;
  loadScripts: () => Promise<void>;
  saveScript: (script: main.TouchScript) => Promise<void>;
  deleteScript: (name: string) => Promise<void>;
  renameScript: (oldName: string, newName: string) => Promise<void>;
  setCurrentScript: (script: main.TouchScript | null) => void;
  updateRecordingDuration: () => void;

  // Task Actions
  loadTasks: () => Promise<void>;
  saveTask: (task: ScriptTask) => Promise<void>;
  deleteTask: (name: string) => Promise<void>;
  runTask: (deviceId: string, task: ScriptTask) => Promise<void>;
  pauseTask: () => Promise<void>;
  resumeTask: () => Promise<void>;
  stopTask: () => Promise<void>;

  // Event subscription
  subscribeToEvents: () => () => void;
}

export const useAutomationStore = create<AutomationState>((set, get) => ({
  // Initial state
  isRecording: false,
  isPlaying: false,
  recordingDeviceId: null,
  playingDeviceId: null,
  recordingStartTime: null,
  recordingDuration: 0,
  currentScript: null,
  scripts: [],
  tasks: [],
  playbackProgress: null,
  taskProgress: null,
  isTaskRunning: false,
  isPaused: false,
  runningTaskName: null,

  // Actions
  startRecording: async (deviceId: string) => {
    try {
      await StartTouchRecording(deviceId);
      set({
        isRecording: true,
        recordingDeviceId: deviceId,
        recordingStartTime: Date.now(),
        recordingDuration: 0,
        currentScript: null,
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
      set({
        isRecording: false,
        recordingDeviceId: null,
        recordingStartTime: null,
        recordingDuration: 0,
        currentScript: script,
      });
      return script;
    } catch (err) {
      console.error('Failed to stop recording:', err);
      set({
        isRecording: false,
        recordingDeviceId: null,
        recordingStartTime: null,
        recordingDuration: 0,
      });
      throw err;
    }
  },

  playScript: async (deviceId: string, script: main.TouchScript) => {
    try {
      await PlayTouchScript(deviceId, script);
      set({
        isPlaying: true,
        playingDeviceId: deviceId,
        playbackProgress: { current: 0, total: script.events?.length || 0 },
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
    set({
      isPlaying: false,
      playingDeviceId: null,
      playbackProgress: null,
    });
  },

  loadScripts: async () => {
    try {
      const scripts = await LoadTouchScripts();
      set({ scripts: scripts || [] });
    } catch (err) {
      console.error('Failed to load scripts:', err);
      set({ scripts: [] });
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
      // Manual cast if needed or simple assignment
      set({ tasks: (tasks as unknown as ScriptTask[]) || [] });
    } catch (err) {
      console.error('Failed to load tasks:', err);
      set({ tasks: [] });
    }
  },

  saveTask: async (task: ScriptTask) => {
    try {
      // Cast to any to bypass strict type check if Wails models aren't updated yet
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

  runTask: async (deviceId: string, task: ScriptTask) => {
    try {
      await RunScriptTask(deviceId, task as any);
      set({
        isTaskRunning: true,
        isPaused: false,
        runningTaskName: task.name,
      });
    } catch (err) {
      console.error('Failed to run task:', err);
      throw err;
    }
  },

  pauseTask: async () => {
    const deviceId = get().playingDeviceId;
    if (deviceId) {
      // Cast to any because PauseTask might not be in Wails definitions yet
      await (PauseTask as any)(deviceId);
      set({ isPaused: true });
    }
  },

  resumeTask: async () => {
    const deviceId = get().playingDeviceId;
    if (deviceId) {
      await (ResumeTask as any)(deviceId);
      set({ isPaused: false });
    }
  },

  stopTask: async () => {
    const deviceId = get().playingDeviceId;
    if (deviceId) {
      await (StopTask as any)(deviceId);
      set({ isTaskRunning: false, isPaused: false, runningTaskName: null, taskProgress: null });
    }
  },

  setCurrentScript: (script: main.TouchScript | null) => {
    set({ currentScript: script });
  },

  updateRecordingDuration: () => {
    const { isRecording, recordingStartTime } = get();
    if (isRecording && recordingStartTime) {
      const duration = Math.floor((Date.now() - recordingStartTime) / 1000);
      set({ recordingDuration: duration });
    }
  },

  // Event subscription
  subscribeToEvents: () => {
    const handleRecordStarted = (data: any) => {
      set({
        isRecording: true,
        recordingDeviceId: data.deviceId,
        recordingStartTime: data.startTime * 1000,
      });
    };

    const handleRecordStopped = (data: any) => {
      set({
        isRecording: false,
        recordingDeviceId: null,
        recordingStartTime: null,
        recordingDuration: 0,
      });
    };

    const handlePlaybackStarted = (data: any) => {
      set({
        isPlaying: true,
        playingDeviceId: data.deviceId,
        playbackProgress: { current: 0, total: data.total },
      });
    };

    const handlePlaybackProgress = (data: any) => {
      set({
        playbackProgress: { current: data.current, total: data.total },
      });
    };

    const handlePlaybackCompleted = (data: any) => {
      set({
        isPlaying: false,
        playingDeviceId: null,
        playbackProgress: null,
      });
    };

    const handleTaskStarted = (data: any) => {
      set({
        isTaskRunning: true,
        isPaused: false,
        runningTaskName: data.taskName,
        playingDeviceId: data.deviceId,
      });
    };

    const handleTaskCompleted = (data: any) => {
      set({
        isTaskRunning: false,
        isPaused: false,
        runningTaskName: null,
        taskProgress: null,
        playingDeviceId: null,
      });
    };

    const handleTaskPaused = (data: any) => set({ isPaused: true });
    const handleTaskResumed = (data: any) => set({ isPaused: false });

    const handleTaskStepRunning = (data: any) => {
      set({
        taskProgress: {
          stepIndex: data.stepIndex,
          totalSteps: data.totalSteps,
          currentLoop: data.currentLoop,
          totalLoops: data.totalLoops,
          currentAction: data.currentAction || (data.type === 'wait' ? `Wait ${data.value}ms` : `Running: ${data.value}`),
        },
      });
    };

    EventsOn('touch-record-started', handleRecordStarted);
    EventsOn('touch-record-stopped', handleRecordStopped);
    EventsOn('touch-playback-started', handlePlaybackStarted);
    EventsOn('touch-playback-progress', handlePlaybackProgress);
    EventsOn('touch-playback-completed', handlePlaybackCompleted);
    EventsOn('task-started', handleTaskStarted);
    EventsOn('task-completed', handleTaskCompleted);
    EventsOn('task-step-running', handleTaskStepRunning);
    EventsOn('task-paused', handleTaskPaused);
    EventsOn('task-resumed', handleTaskResumed);

    return () => {
      EventsOff('touch-record-started');
      EventsOff('touch-record-stopped');
      EventsOff('touch-playback-started');
      EventsOff('touch-playback-progress');
      EventsOff('touch-playback-completed');
      EventsOff('task-started');
      EventsOff('task-completed');
      EventsOff('task-step-running');
    };
  },
}));
