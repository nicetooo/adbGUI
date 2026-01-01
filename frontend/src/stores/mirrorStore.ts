import { create } from 'zustand';
import { MirrorStatus, RecordStatus } from './types';
import { OpenPath } from '../../wailsjs/go/main/App';

// @ts-ignore
const EventsOn = (window as any).runtime.EventsOn;
// @ts-ignore
const EventsOff = (window as any).runtime.EventsOff;

interface MirrorState {
  // State
  mirrorStatuses: Record<string, MirrorStatus>;
  recordStatuses: Record<string, RecordStatus>;
  recordPathRefs: Record<string, string>;

  // Actions
  setMirrorStatus: (deviceId: string, status: Partial<MirrorStatus>) => void;
  setRecordStatus: (deviceId: string, status: Partial<RecordStatus>) => void;

  // Duration update (called by interval)
  updateDurations: () => void;

  // Event subscription
  subscribeToEvents: (onRecordSaved: (deviceId: string, path: string) => void) => () => void;
}

export const useMirrorStore = create<MirrorState>((set, get) => ({
  // Initial state
  mirrorStatuses: {},
  recordStatuses: {},
  recordPathRefs: {},

  // Actions
  setMirrorStatus: (deviceId: string, status: Partial<MirrorStatus>) => {
    set(state => ({
      mirrorStatuses: {
        ...state.mirrorStatuses,
        [deviceId]: {
          ...state.mirrorStatuses[deviceId],
          ...status,
        } as MirrorStatus,
      },
    }));
  },

  setRecordStatus: (deviceId: string, status: Partial<RecordStatus>) => {
    set(state => ({
      recordStatuses: {
        ...state.recordStatuses,
        [deviceId]: {
          ...state.recordStatuses[deviceId],
          ...status,
        } as RecordStatus,
      },
    }));
  },

  updateDurations: () => {
    const now = Math.floor(Date.now() / 1000);

    set(state => {
      let mirrorChanged = false;
      let recordChanged = false;
      const nextMirror = { ...state.mirrorStatuses };
      const nextRecord = { ...state.recordStatuses };

      Object.keys(nextMirror).forEach(id => {
        if (nextMirror[id].isMirroring && nextMirror[id].startTime) {
          nextMirror[id] = {
            ...nextMirror[id],
            duration: now - nextMirror[id].startTime!,
          };
          mirrorChanged = true;
        }
      });

      Object.keys(nextRecord).forEach(id => {
        if (nextRecord[id].isRecording && nextRecord[id].startTime) {
          nextRecord[id] = {
            ...nextRecord[id],
            duration: now - nextRecord[id].startTime!,
          };
          recordChanged = true;
        }
      });

      if (!mirrorChanged && !recordChanged) {
        return state;
      }

      return {
        mirrorStatuses: mirrorChanged ? nextMirror : state.mirrorStatuses,
        recordStatuses: recordChanged ? nextRecord : state.recordStatuses,
      };
    });
  },

  subscribeToEvents: (onRecordSaved) => {
    EventsOn('scrcpy-started', (data: any) => {
      const deviceId = data.deviceId;
      set(state => ({
        mirrorStatuses: {
          ...state.mirrorStatuses,
          [deviceId]: {
            isMirroring: true,
            startTime: data.startTime,
            duration: 0,
          },
        },
      }));
    });

    EventsOn('scrcpy-stopped', (deviceId: string) => {
      set(state => {
        const next = { ...state.mirrorStatuses };
        if (next[deviceId]) {
          next[deviceId] = { ...next[deviceId], isMirroring: false, startTime: null };
        }
        return { mirrorStatuses: next };
      });
    });

    EventsOn('scrcpy-record-started', (data: any) => {
      const deviceId = data.deviceId;
      set(state => ({
        recordStatuses: {
          ...state.recordStatuses,
          [deviceId]: {
            isRecording: true,
            startTime: data.startTime,
            duration: 0,
            recordPath: data.recordPath,
          },
        },
        recordPathRefs: {
          ...state.recordPathRefs,
          [deviceId]: data.recordPath,
        },
      }));
    });

    EventsOn('scrcpy-record-stopped', (deviceId: string) => {
      const { recordPathRefs } = get();
      const path = recordPathRefs[deviceId];

      set(state => {
        const next = { ...state.recordStatuses };
        if (next[deviceId]) {
          next[deviceId] = { ...next[deviceId], isRecording: false, startTime: null };
        }
        return { recordStatuses: next };
      });

      if (path && onRecordSaved) {
        onRecordSaved(deviceId, path);
      }
    });

    // Return cleanup function
    return () => {
      EventsOff('scrcpy-started');
      EventsOff('scrcpy-stopped');
      EventsOff('scrcpy-record-started');
      EventsOff('scrcpy-record-stopped');
    };
  },
}));

// Helper to open recorded file path
export const openRecordPath = (path: string) => {
  OpenPath(path);
};
