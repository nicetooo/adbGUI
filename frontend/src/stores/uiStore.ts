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
  llmConfigVisible: boolean;

  // App info
  appVersion: string;

  // Component measurements
  syncTimelineContainerWidth: number;

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
  showLLMConfig: () => void;
  hideLLMConfig: () => void;

  // Component measurements
  setSyncTimelineContainerWidth: (width: number) => void;

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
    llmConfigVisible: false,

    appVersion: '',
    syncTimelineContainerWidth: 0,

    // Actions
    setSelectedKey: (key) => set((state: UIState) => {
      state.selectedKey = key as ViewKey;
    }),

    navigateToView: (viewName) => {
      const viewKey = VIEW_NAME_MAP[viewName];
      if (viewKey) {
        set((state: UIState) => {
          state.selectedKey = viewKey;
        });
      }
    },

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
    showLLMConfig: () => set((state: UIState) => {
      state.llmConfigVisible = true;
    }),
    hideLLMConfig: () => set((state: UIState) => {
      state.llmConfigVisible = false;
    }),

    // Component measurements
    setSyncTimelineContainerWidth: (width) => set((state: UIState) => {
      state.syncTimelineContainerWidth = width;
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
