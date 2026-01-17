import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';
import { GetBackendLogs } from '../../wailsjs/go/main/App';

declare global {
  interface Window {
    runtimeLogs: string[];
  }
}

interface FeedbackState {
  // Logs
  backendLogs: string[];
  frontendLogs: string[];
  
  // Selected log indices
  selectedBackendLogs: Set<number>;
  selectedFrontendLogs: Set<number>;
  
  // Description
  description: string;
  
  // Actions
  setBackendLogs: (logs: string[]) => void;
  setFrontendLogs: (logs: string[]) => void;
  setSelectedBackendLogs: (selected: Set<number>) => void;
  setSelectedFrontendLogs: (selected: Set<number>) => void;
  setDescription: (description: string) => void;
  
  // Bulk actions
  toggleAllBackend: (checked: boolean) => void;
  toggleAllFrontend: (checked: boolean) => void;
  toggleBackendLog: (index: number, checked: boolean) => void;
  toggleFrontendLog: (index: number, checked: boolean) => void;
  
  // Fetch logs
  fetchLogs: () => Promise<void>;
  
  // Reset
  reset: () => void;
}

export const useFeedbackStore = create<FeedbackState>()(
  immer((set, get) => ({
    backendLogs: [],
    frontendLogs: [],
    selectedBackendLogs: new Set(),
    selectedFrontendLogs: new Set(),
    description: '',

    setBackendLogs: (logs) => set((state) => {
      state.backendLogs = logs;
    }),

    setFrontendLogs: (logs) => set((state) => {
      state.frontendLogs = logs;
    }),

    setSelectedBackendLogs: (selected) => set((state) => {
      state.selectedBackendLogs = selected;
    }),

    setSelectedFrontendLogs: (selected) => set((state) => {
      state.selectedFrontendLogs = selected;
    }),

    setDescription: (description) => set((state) => {
      state.description = description;
    }),

    toggleAllBackend: (checked) => set((state) => {
      if (checked) {
        state.selectedBackendLogs = new Set(state.backendLogs.map((_, i) => i));
      } else {
        state.selectedBackendLogs = new Set();
      }
    }),

    toggleAllFrontend: (checked) => set((state) => {
      if (checked) {
        state.selectedFrontendLogs = new Set(state.frontendLogs.map((_, i) => i));
      } else {
        state.selectedFrontendLogs = new Set();
      }
    }),

    toggleBackendLog: (index, checked) => set((state) => {
      const next = new Set(state.selectedBackendLogs);
      if (checked) {
        next.add(index);
      } else {
        next.delete(index);
      }
      state.selectedBackendLogs = next;
    }),

    toggleFrontendLog: (index, checked) => set((state) => {
      const next = new Set(state.selectedFrontendLogs);
      if (checked) {
        next.add(index);
      } else {
        next.delete(index);
      }
      state.selectedFrontendLogs = next;
    }),

    fetchLogs: async () => {
      try {
        const bLogs = await GetBackendLogs();
        set((state) => {
          state.backendLogs = bLogs || [];
        });

        const fLogs = window.runtimeLogs || [];
        set((state) => {
          state.frontendLogs = fLogs;
        });
      } catch (err) {
        console.error('Failed to fetch logs:', err);
      }
    },

    reset: () => set((state) => {
      state.backendLogs = [];
      state.frontendLogs = [];
      state.selectedBackendLogs = new Set();
      state.selectedFrontendLogs = new Set();
      state.description = '';
    }),
  }))
);
