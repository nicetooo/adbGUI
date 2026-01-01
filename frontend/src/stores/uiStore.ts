import { create } from 'zustand';
import { ViewKey, VIEW_KEYS, VIEW_NAME_MAP } from './types';
import { GetAppVersion } from '../../wailsjs/go/main/App';

// @ts-ignore
const EventsOn = (window as any).runtime.EventsOn;
// @ts-ignore
const EventsOff = (window as any).runtime.EventsOff;

interface UIState {
  // Navigation
  selectedKey: ViewKey;

  // Modals
  aboutVisible: boolean;
  wirelessConnectVisible: boolean;
  feedbackVisible: boolean;

  // App info
  appVersion: string;

  // Actions
  setSelectedKey: (key: ViewKey | string) => void;
  navigateToView: (viewName: string) => void;

  // Modal toggles
  showAbout: () => void;
  hideAbout: () => void;
  showWirelessConnect: () => void;
  hideWirelessConnect: () => void;
  showFeedback: () => void;
  hideFeedback: () => void;

  // Initialization
  init: () => Promise<void>;

  // Event subscription (for tray navigation)
  subscribeToEvents: (onDeviceSelect: (deviceId: string) => void) => () => void;
}

export const useUIStore = create<UIState>((set) => ({
  // Initial state
  selectedKey: VIEW_KEYS.DEVICES,

  aboutVisible: false,
  wirelessConnectVisible: false,
  feedbackVisible: false,

  appVersion: '',

  // Actions
  setSelectedKey: (key) => set({ selectedKey: key as ViewKey }),

  navigateToView: (viewName) => {
    const viewKey = VIEW_NAME_MAP[viewName];
    if (viewKey) {
      set({ selectedKey: viewKey });
    }
  },

  // Modal toggles
  showAbout: () => set({ aboutVisible: true }),
  hideAbout: () => set({ aboutVisible: false }),
  showWirelessConnect: () => set({ wirelessConnectVisible: true }),
  hideWirelessConnect: () => set({ wirelessConnectVisible: false }),
  showFeedback: () => set({ feedbackVisible: true }),
  hideFeedback: () => set({ feedbackVisible: false }),

  // Initialization
  init: async () => {
    try {
      const version = await GetAppVersion();
      set({ appVersion: version });
    } catch {
      // Ignore version fetch errors
    }
  },

  // Event subscription
  subscribeToEvents: (onDeviceSelect) => {
    EventsOn('tray:navigate', (data: any) => {
      if (data.deviceId && onDeviceSelect) {
        onDeviceSelect(data.deviceId);
      }
      if (data.view) {
        const viewKey = VIEW_NAME_MAP[data.view];
        if (viewKey) {
          set({ selectedKey: viewKey });
        }
      }
    });

    return () => {
      EventsOff('tray:navigate');
    };
  },
}));
