import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';
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
  mcpVisible: boolean;

  // App info
  appVersion: string;

  // Actions
  setSelectedKey: (key: ViewKey | string) => void;

  // Modal toggles
  showAbout: () => void;
  hideAbout: () => void;
  showWirelessConnect: () => void;
  hideWirelessConnect: () => void;
  showFeedback: () => void;
  hideFeedback: () => void;
  showMCP: () => void;
  hideMCP: () => void;

  // Initialization
  init: () => Promise<void>;

  // Event subscription (for tray navigation)
  subscribeToEvents: (onDeviceSelect: (deviceId: string) => void) => () => void;
}

export const useUIStore = create<UIState>()(
  immer((set) => ({
    // Initial state
    selectedKey: VIEW_KEYS.DEVICES,

    aboutVisible: false,
    wirelessConnectVisible: false,
    feedbackVisible: false,
    mcpVisible: false,

    appVersion: '',

    // Actions
    setSelectedKey: (key) => set((state: UIState) => {
      state.selectedKey = key as ViewKey;
    }),

    // Modal toggles
    showAbout: () => set((state: UIState) => {
      state.aboutVisible = true;
    }),
    hideAbout: () => set((state: UIState) => {
      state.aboutVisible = false;
    }),
    showWirelessConnect: () => set((state: UIState) => {
      state.wirelessConnectVisible = true;
    }),
    hideWirelessConnect: () => set((state: UIState) => {
      state.wirelessConnectVisible = false;
    }),
    showFeedback: () => set((state: UIState) => {
      state.feedbackVisible = true;
    }),
    hideFeedback: () => set((state: UIState) => {
      state.feedbackVisible = false;
    }),
    showMCP: () => set((state: UIState) => {
      state.mcpVisible = true;
    }),
    hideMCP: () => set((state: UIState) => {
      state.mcpVisible = false;
    }),

    // Initialization
    init: async () => {
      try {
        const version = await GetAppVersion();
        set((state: UIState) => {
          state.appVersion = version;
        });
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
            set((state: UIState) => {
              state.selectedKey = viewKey;
            });
          }
        }
      });

      return () => {
        EventsOff('tray:navigate');
      };
    },
  }))
);
