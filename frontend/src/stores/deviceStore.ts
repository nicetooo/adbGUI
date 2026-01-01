import { create } from 'zustand';
import { Device, HistoryDevice } from './types';
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
} from '../../wailsjs/go/main/App';
// @ts-ignore
import { main } from '../../wailsjs/go/models';

interface DeviceState {
  // State
  devices: Device[];
  historyDevices: HistoryDevice[];
  selectedDevice: string;
  loading: boolean;
  busyDevices: Set<string>;
  isFetching: boolean;

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
}

export const useDeviceStore = create<DeviceState>((set, get) => ({
  // Initial state
  devices: [],
  historyDevices: [],
  selectedDevice: '',
  loading: false,
  busyDevices: new Set(),
  isFetching: false,

  deviceInfoVisible: false,
  deviceInfoLoading: false,
  selectedDeviceInfo: null,

  // Actions
  fetchDevices: async (silent = false) => {
    const { isFetching, selectedDevice } = get();
    if (isFetching) return;

    set({ isFetching: true });
    if (!silent) {
      set({ loading: true });
    }

    try {
      const res = await GetDevices(!silent);
      const devices = res || [];
      set({ devices });

      try {
        const history = await GetHistoryDevices();
        set({ historyDevices: history || [] });
      } catch {
        set({ historyDevices: [] });
      }

      // Auto-select first device if none selected
      if (devices.length > 0 && !selectedDevice) {
        set({ selectedDevice: devices[0].id });
      }
    } catch (err) {
      if (!silent) {
        throw err; // Let the caller handle the error for UI feedback
      }
    } finally {
      set({ loading: false, isFetching: false });
    }
  },

  setSelectedDevice: (id: string) => {
    set({ selectedDevice: id });
  },

  handleFetchDeviceInfo: async (deviceId: string) => {
    set({ deviceInfoVisible: true, deviceInfoLoading: true });
    try {
      const res = await GetDeviceInfo(deviceId);
      set({ selectedDeviceInfo: res });
    } catch (err) {
      throw err;
    } finally {
      set({ deviceInfoLoading: false });
    }
  },

  closeDeviceInfo: () => {
    set({ deviceInfoVisible: false, selectedDeviceInfo: null });
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
    set(state => ({
      busyDevices: new Set(state.busyDevices).add(deviceId)
    }));

    try {
      const res = await SwitchToWireless(deviceId);
      if (res.includes('connected to')) {
        await get().fetchDevices(true);
      } else {
        throw new Error(res);
      }
    } finally {
      set(state => {
        const next = new Set(state.busyDevices);
        next.delete(deviceId);
        return { busyDevices: next };
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
}));
