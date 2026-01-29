import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

const EventsOn = (window as any).runtime.EventsOn;
const EventsOff = (window as any).runtime.EventsOff;

// ========================================
// Unified Element Store
// Shared state for UI elements across Recording, Workflow, and UI Inspector
// ========================================

// Re-export UINode type for convenience (also defined in automationStore)
export interface UINode {
  text: string;
  resourceId: string;
  class: string;
  package: string;
  contentDesc: string;
  bounds: string;
  checkable: string;
  checked: string;
  clickable: string;
  enabled: string;
  focusable: string;
  focused: string;
  scrollable: string;
  longClickable: string;
  password: string;
  selected: string;
  nodes: UINode[];
}

export interface ElementSelector {
  type: 'text' | 'id' | 'desc' | 'class' | 'xpath' | 'bounds' | 'contains' | 'coordinates' | 'advanced';
  value: string;
  index?: number;
}

export interface SelectorSuggestion {
  type: string;
  value: string;
  priority: number;
  description: string;
}

interface ElementState {
  // Shared UI hierarchy
  hierarchy: UINode | null;
  rawXml: string | null;
  isLoading: boolean;
  lastFetchTime: number | null;
  lastFetchDeviceId: string | null;

  // Selected element state (for element picker)
  selectedNode: UINode | null;
  highlightedNode: UINode | null;

  // Element picker modal state
  isPickerOpen: boolean;
  pickerCallback: ((selector: ElementSelector | null) => void) | null;

  // Actions
  fetchHierarchy: (deviceId: string, force?: boolean) => Promise<UINode | null>;
  clearHierarchy: () => void;
  setSelectedNode: (node: UINode | null) => void;
  setHighlightedNode: (node: UINode | null) => void;
  openElementPicker: (callback: (selector: ElementSelector | null) => void) => void;
  closeElementPicker: () => void;
  confirmElementPicker: (selector: ElementSelector) => void;

  // Backend element operations
  clickElement: (deviceId: string, selector: ElementSelector) => Promise<void>;
  inputText: (deviceId: string, selector: ElementSelector, text: string) => Promise<void>;
  waitForElement: (deviceId: string, selector: ElementSelector, timeout?: number) => Promise<void>;
  getElementProperties: (deviceId: string, selector: ElementSelector) => Promise<Record<string, any>>;

  // Event subscription
  subscribeToEvents: () => () => void;
}

export const useElementStore = create<ElementState>()(
  immer((set, get) => ({
    // Initial state
    hierarchy: null,
    rawXml: null,
    isLoading: false,
    lastFetchTime: null,
    lastFetchDeviceId: null,
    selectedNode: null,
    highlightedNode: null,
    isPickerOpen: false,
    pickerCallback: null,

    // Actions
    fetchHierarchy: async (deviceId: string, force = false) => {
      const { lastFetchTime, lastFetchDeviceId, rawXml } = get();

      // Cache check: skip if same device and fetched within 2 seconds
      const now = Date.now();
      if (!force && lastFetchDeviceId === deviceId && lastFetchTime && now - lastFetchTime < 2000) {
        return get().hierarchy;
      }

      set((state: ElementState) => {
        state.isLoading = true;
      });
      try {
        const result = await (window as any).go.main.App.GetUIHierarchy(deviceId);

        // Only update if content changed
        if (result.rawXml !== rawXml) {
          set((state: ElementState) => {
            state.hierarchy = result.root;
            state.rawXml = result.rawXml;
            state.lastFetchTime = now;
            state.lastFetchDeviceId = deviceId;
            state.isLoading = false;
          });
        } else {
          set((state: ElementState) => {
            state.isLoading = false;
            state.lastFetchTime = now;
          });
        }
        return result.root;
      } catch (err) {
        console.error('Failed to fetch UI hierarchy:', err);
        set((state: ElementState) => {
          state.isLoading = false;
        });
        throw err;
      }
    },

    clearHierarchy: () => {
      set((state: ElementState) => {
        state.hierarchy = null;
        state.rawXml = null;
        state.selectedNode = null;
        state.highlightedNode = null;
        state.lastFetchTime = null;
        state.lastFetchDeviceId = null;
      });
    },

    setSelectedNode: (node: UINode | null) => {
      set((state: ElementState) => {
        state.selectedNode = node;
      });
    },

    setHighlightedNode: (node: UINode | null) => {
      set((state: ElementState) => {
        state.highlightedNode = node;
      });
    },

    openElementPicker: (callback: (selector: ElementSelector | null) => void) => {
      set((state: ElementState) => {
        state.isPickerOpen = true;
        state.pickerCallback = callback;
        state.selectedNode = null;
      });
    },

    closeElementPicker: () => {
      const { pickerCallback } = get();
      if (pickerCallback) {
        pickerCallback(null);
      }
      set((state: ElementState) => {
        state.isPickerOpen = false;
        state.pickerCallback = null;
        state.selectedNode = null;
      });
    },

    confirmElementPicker: (selector: ElementSelector) => {
      const { pickerCallback } = get();
      if (pickerCallback) {
        pickerCallback(selector);
      }
      set((state: ElementState) => {
        state.isPickerOpen = false;
        state.pickerCallback = null;
        state.selectedNode = null;
      });
    },

    // Backend element operations
    clickElement: async (deviceId: string, selector: ElementSelector) => {
      await (window as any).go.main.App.ClickElement(
        { Cancelled: false } as any, // context placeholder
        deviceId,
        selector,
        null // use default config
      );
    },

    inputText: async (deviceId: string, selector: ElementSelector, text: string) => {
      await (window as any).go.main.App.InputTextToElement(
        { Cancelled: false } as any,
        deviceId,
        selector,
        text,
        false, // clearFirst
        null
      );
    },

    waitForElement: async (deviceId: string, selector: ElementSelector, timeout = 10000) => {
      await (window as any).go.main.App.WaitForElement(
        { Cancelled: false } as any,
        deviceId,
        selector,
        timeout
      );
    },

    getElementProperties: async (deviceId: string, selector: ElementSelector) => {
      return await (window as any).go.main.App.GetElementProperties(deviceId, selector);
    },

    // Event subscription
    subscribeToEvents: () => {
      // Listen for UI hierarchy updates from other sources
      const handleHierarchyUpdate = (data: any) => {
        if (data.root) {
          set((state: ElementState) => {
            state.hierarchy = data.root;
            state.rawXml = data.rawXml;
            state.lastFetchTime = Date.now();
            state.lastFetchDeviceId = data.deviceId;
          });
        }
      };

      EventsOn('ui-hierarchy-updated', handleHierarchyUpdate);

      return () => {
        EventsOff('ui-hierarchy-updated');
      };
    },
  }))
);
