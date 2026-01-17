import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

export interface SortConfig {
  key: string;
  order: 'ascend' | 'descend';
  sorter: (a: any, b: any) => number;
}

interface VirtualTableState {
  // Sort configs keyed by table ID
  sortConfigs: Record<string, SortConfig | null>;

  // Actions
  setSortConfig: (tableId: string, config: SortConfig | null) => void;
  clearSortConfig: (tableId: string) => void;
  clearAllSortConfigs: () => void;
}

export const useVirtualTableStore = create<VirtualTableState>()(
  immer((set) => ({
    sortConfigs: {},

    setSortConfig: (tableId, config) => set((state) => {
      state.sortConfigs[tableId] = config;
    }),

    clearSortConfig: (tableId) => set((state) => {
      delete state.sortConfigs[tableId];
    }),

    clearAllSortConfigs: () => set((state) => {
      state.sortConfigs = {};
    }),
  }))
);
