import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';
import { Device, HistoryDevice, BatchOperation, BatchResult, BatchOperationResult } from './types';
import {
  GetDevices,
  GetDeviceInfo,
  AdbPair,
  AdbConnect,
  AdbDisconnect,
  SwitchToWireless,
  GetHistoryDevices,
  RemoveHistoryDevice,
  OpenSettings,
  TogglePinDevice,
  RestartAdbServer,
  ExecuteBatchOperation,
  SelectAPKForBatch,
  SelectFileForBatch,
} from '../../wailsjs/go/main/App';
// @ts-ignore
import { main } from '../types/wails-models';
// @ts-ignore
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';

interface DeviceState {
  // State
  devices: Device[];
  historyDevices: HistoryDevice[];
  selectedDevice: string;
  loading: boolean;
  busyDevices: Set<string>;
  isFetching: boolean;

  // Multi-select state for batch operations
  selectedDevices: Set<string>;
  batchModalVisible: boolean;
  batchOperationInProgress: boolean;
  batchResults: BatchResult[];

  // Device info modal state
  deviceInfoVisible: boolean;
  deviceInfoLoading: boolean;
  selectedDeviceInfo: main.DeviceInfo | null;

  // Actions
  fetchDevices: (silent?: boolean) => Promise<void>;
  setSelectedDevice: (id: string) => void;
  handleFetchDeviceInfo: (deviceId: string) => Promise<void>;
  closeDeviceInfo: () => void;
  handleAdbConnect: (address: string) => Promise<void>;
  handleAdbPair: (address: string, code: string) => Promise<void>;
  handleSwitchToWireless: (deviceId: string) => Promise<void>;
  handleAdbDisconnect: (deviceId: string) => Promise<void>;
  handleRemoveHistoryDevice: (deviceId: string) => Promise<void>;
  handleOpenSettings: (deviceId: string, action?: string, data?: string) => Promise<void>;
  handleTogglePin: (serial: string) => Promise<void>;
  handleRestartAdbServer: () => Promise<void>;
  subscribeToDeviceEvents: () => () => void;

  // Batch operation actions
  toggleDeviceSelection: (deviceId: string) => void;
  selectAllDevices: () => void;
  clearSelection: () => void;
  openBatchModal: () => void;
  closeBatchModal: () => void;
  executeBatchOperation: (op: BatchOperation) => Promise<BatchOperationResult>;
  selectAPKForBatch: () => Promise<string>;
  selectFileForBatch: () => Promise<string>;
  subscribeToBatchEvents: () => () => void;
}

export const useDeviceStore = create<DeviceState>()(
  immer((set, get) => ({
    // Initial state
    devices: [],
    historyDevices: [],
    selectedDevice: '',
    loading: false,
    busyDevices: new Set(),
    isFetching: false,

    // Multi-select state
    selectedDevices: new Set(),
    batchModalVisible: false,
    batchOperationInProgress: false,
    batchResults: [],

    deviceInfoVisible: false,
    deviceInfoLoading: false,
    selectedDeviceInfo: null,

    // Actions
    fetchDevices: async (silent = false) => {
      const { isFetching, selectedDevice } = get();
      if (isFetching) return;

      set((state: DeviceState) => {
        state.isFetching = true;
        if (!silent) {
          state.loading = true;
        }
      });

      try {
        const res = await GetDevices(!silent);
        const devices = res || [];
        set((state: DeviceState) => {
          state.devices = devices;
        });

        try {
          const history = await GetHistoryDevices();
          set((state: DeviceState) => {
            state.historyDevices = history || [];
          });
        } catch {
          set((state: DeviceState) => {
            state.historyDevices = [];
          });
        }

        // Auto-select first device if none selected
        if (devices.length > 0 && !selectedDevice) {
          set((state: DeviceState) => {
            state.selectedDevice = devices[0].id;
          });
        }
      } catch (err) {
        if (!silent) {
          throw err; // Let the caller handle the error for UI feedback
        }
      } finally {
        set((state: DeviceState) => {
          state.loading = false;
          state.isFetching = false;
        });
      }
    },

    setSelectedDevice: (id: string) => {
      set({ selectedDevice: id });
    },

    handleFetchDeviceInfo: async (deviceId: string) => {
      set((state: DeviceState) => {
        state.deviceInfoVisible = true;
        state.deviceInfoLoading = true;
      });
      try {
        const res = await GetDeviceInfo(deviceId);
        set((state: DeviceState) => {
          state.selectedDeviceInfo = res;
        });
      } catch (err) {
        throw err;
      } finally {
        set((state: DeviceState) => {
          state.deviceInfoLoading = false;
        });
      }
    },

    closeDeviceInfo: () => {
      set((state: DeviceState) => {
        state.deviceInfoVisible = false;
        state.selectedDeviceInfo = null;
      });
    },

    handleAdbConnect: async (address: string) => {
      const res = await AdbConnect(address);
      if (res.includes('connected to')) {
        await get().fetchDevices();
      } else {
        throw new Error(res);
      }
    },

    handleAdbPair: async (address: string, code: string) => {
      const res = await AdbPair(address, code);
      if (res.includes('Successfully paired')) {
        await get().fetchDevices();
      } else {
        throw new Error(res);
      }
    },

    handleSwitchToWireless: async (deviceId: string) => {
      set((state: DeviceState) => {
        state.busyDevices.add(deviceId);
      });

      try {
        const res = await SwitchToWireless(deviceId);
        if (res.includes('connected to')) {
          await get().fetchDevices(true);
        } else {
          throw new Error(res);
        }
      } finally {
        set((state: DeviceState) => {
          state.busyDevices.delete(deviceId);
        });
      }
    },

    handleAdbDisconnect: async (deviceId: string) => {
      await AdbDisconnect(deviceId);
      await get().fetchDevices();
    },

    handleRemoveHistoryDevice: async (deviceId: string) => {
      await RemoveHistoryDevice(deviceId);
      await get().fetchDevices();
    },

    handleOpenSettings: async (deviceId: string, action = '', data = '') => {
      await OpenSettings(deviceId, action, data);
    },

    handleTogglePin: async (serial: string) => {
      await TogglePinDevice(serial);
      await get().fetchDevices(true);
    },

    handleRestartAdbServer: async () => {
      await RestartAdbServer();
      await get().fetchDevices();
    },

    subscribeToDeviceEvents: () => {
      const handler = async (devices: Device[]) => {
        const { selectedDevice } = get();
        set((state: DeviceState) => {
          state.devices = devices || [];
        });

        // Also fetch history devices
        try {
          const history = await GetHistoryDevices();
          set((state: DeviceState) => {
            state.historyDevices = history || [];
          });
        } catch {
          // Ignore history fetch errors
        }

        // Auto-select first device if none selected
        if (devices && devices.length > 0 && !selectedDevice) {
          set((state: DeviceState) => {
            state.selectedDevice = devices[0].id;
          });
        }
      };

      EventsOn('devices-changed', handler);
      return () => {
        EventsOff('devices-changed');
      };
    },

    // Batch operation actions
    toggleDeviceSelection: (deviceId: string) => {
      set((state: DeviceState) => {
        if (state.selectedDevices.has(deviceId)) {
          state.selectedDevices.delete(deviceId);
        } else {
          state.selectedDevices.add(deviceId);
        }
      });
    },

    selectAllDevices: () => {
      const { devices } = get();
      const onlineDevices = devices.filter(d => d.state === 'device');
      set((state: DeviceState) => {
        state.selectedDevices = new Set(onlineDevices.map(d => d.id));
      });
    },

    clearSelection: () => {
      set((state: DeviceState) => {
        state.selectedDevices.clear();
        state.batchResults = [];
      });
    },

    openBatchModal: () => {
      set((state: DeviceState) => {
        state.batchModalVisible = true;
        state.batchResults = [];
      });
    },

    closeBatchModal: () => {
      set((state: DeviceState) => {
        state.batchModalVisible = false;
        state.batchOperationInProgress = false;
      });
    },

    executeBatchOperation: async (op: BatchOperation): Promise<BatchOperationResult> => {
      set((state: DeviceState) => {
        state.batchOperationInProgress = true;
        state.batchResults = [];
      });
      try {
        const result = await ExecuteBatchOperation(op);
        set((state: DeviceState) => {
          state.batchResults = result.results || [];
        });
        return result;
      } finally {
        set((state: DeviceState) => {
          state.batchOperationInProgress = false;
        });
      }
    },

    selectAPKForBatch: async (): Promise<string> => {
      return await SelectAPKForBatch();
    },

    selectFileForBatch: async (): Promise<string> => {
      return await SelectFileForBatch();
    },

    subscribeToBatchEvents: () => {
      const handler = (result: BatchResult) => {
        set((state: DeviceState) => {
          state.batchResults.push(result);
        });
      };

      EventsOn('batch-progress', handler);
      return () => {
        EventsOff('batch-progress');
      };
    },
  }))
);
