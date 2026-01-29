import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';
import { MirrorStatus, RecordStatus } from './types';
import { main } from '../types/wails-models';

// @ts-ignore
const EventsOn = (window as any).runtime.EventsOn;
// @ts-ignore
const EventsOff = (window as any).runtime.EventsOff;

interface MirrorState {
  // State
  mirrorStatuses: Record<string, MirrorStatus>;
  recordStatuses: Record<string, RecordStatus>;
  recordPathRefs: Record<string, string>;
  
  // New: MirrorView specific states
  deviceConfigs: Record<string, main.ScrcpyConfig>;
  deviceShouldRecord: Record<string, boolean>;
  availableCameras: string[];
  availableDisplays: string[];
  deviceAndroidVer: number;

  // Actions
  setMirrorStatus: (deviceId: string, status: Partial<MirrorStatus>) => void;
  setRecordStatus: (deviceId: string, status: Partial<RecordStatus>) => void;
  
  // New: MirrorView specific actions
  setDeviceConfig: (deviceId: string, config: main.ScrcpyConfig) => void;
  setDeviceShouldRecord: (deviceId: string, should: boolean) => void;
  setAvailableCameras: (cameras: string[]) => void;
  setAvailableDisplays: (displays: string[]) => void;
  setDeviceAndroidVer: (ver: number) => void;

  // Duration update (called by interval)
  updateDurations: () => void;

  // Event subscription
  subscribeToEvents: (onRecordSaved: (deviceId: string, path: string) => void) => () => void;
}

export const useMirrorStore = create<MirrorState>()(
  immer((set, get) => ({
    // Initial state
    mirrorStatuses: {},
    recordStatuses: {},
    recordPathRefs: {},
    
    // New: MirrorView specific initial states
    deviceConfigs: {},
    deviceShouldRecord: {},
    availableCameras: [],
    availableDisplays: [],
    deviceAndroidVer: 0,

    // Actions
    setMirrorStatus: (deviceId: string, status: Partial<MirrorStatus>) => {
      set((state: MirrorState) => {
        if (!state.mirrorStatuses[deviceId]) {
          state.mirrorStatuses[deviceId] = {
            isMirroring: false,
            startTime: null,
            duration: 0,
          };
        }
        Object.assign(state.mirrorStatuses[deviceId], status);
      });
    },

    setRecordStatus: (deviceId: string, status: Partial<RecordStatus>) => {
      set((state: MirrorState) => {
        if (!state.recordStatuses[deviceId]) {
          state.recordStatuses[deviceId] = {
            isRecording: false,
            startTime: null,
            duration: 0,
          };
        }
        Object.assign(state.recordStatuses[deviceId], status);
      });
    },
    
    // New: MirrorView specific actions
    setDeviceConfig: (deviceId: string, config: main.ScrcpyConfig) => {
      set((state: MirrorState) => {
        state.deviceConfigs[deviceId] = config;
      });
    },
    
    setDeviceShouldRecord: (deviceId: string, should: boolean) => {
      set((state: MirrorState) => {
        state.deviceShouldRecord[deviceId] = should;
      });
    },
    
    setAvailableCameras: (cameras: string[]) => {
      set((state: MirrorState) => {
        state.availableCameras = cameras;
      });
    },
    
    setAvailableDisplays: (displays: string[]) => {
      set((state: MirrorState) => {
        state.availableDisplays = displays;
      });
    },
    
    setDeviceAndroidVer: (ver: number) => {
      set((state: MirrorState) => {
        state.deviceAndroidVer = ver;
      });
    },

    updateDurations: () => {
      const now = Math.floor(Date.now() / 1000);

      set((state: MirrorState) => {
        Object.keys(state.mirrorStatuses).forEach(id => {
          const status = state.mirrorStatuses[id];
          if (status.isMirroring && status.startTime) {
            status.duration = now - status.startTime;
          }
        });

        Object.keys(state.recordStatuses).forEach(id => {
          const status = state.recordStatuses[id];
          if (status.isRecording && status.startTime) {
            status.duration = now - status.startTime;
          }
        });
      });
    },

    subscribeToEvents: (onRecordSaved) => {
      EventsOn('scrcpy-started', (data: any) => {
        const deviceId = data.deviceId;
        set((state: MirrorState) => {
          state.mirrorStatuses[deviceId] = {
            isMirroring: true,
            startTime: data.startTime,
            duration: 0,
          };
        });
      });

      EventsOn('scrcpy-stopped', (deviceId: string) => {
        set((state: MirrorState) => {
          if (state.mirrorStatuses[deviceId]) {
            state.mirrorStatuses[deviceId].isMirroring = false;
            state.mirrorStatuses[deviceId].startTime = null;
          }
        });
      });

      EventsOn('scrcpy-record-started', (data: any) => {
        const deviceId = data.deviceId;
        set((state: MirrorState) => {
          state.recordStatuses[deviceId] = {
            isRecording: true,
            startTime: data.startTime,
            duration: 0,
            recordPath: data.recordPath,
          };
          state.recordPathRefs[deviceId] = data.recordPath;
        });
      });

      EventsOn('scrcpy-record-stopped', (deviceId: string) => {
        const { recordPathRefs } = get();
        const path = recordPathRefs[deviceId];

        set((state: MirrorState) => {
          if (state.recordStatuses[deviceId]) {
            state.recordStatuses[deviceId].isRecording = false;
            state.recordStatuses[deviceId].startTime = null;
          }
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
  }))
);
