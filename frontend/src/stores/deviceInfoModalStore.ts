import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

interface DeviceInfoModalState {
  searchText: string;
  setSearchText: (text: string) => void;
  reset: () => void;
}

export const useDeviceInfoModalStore = create<DeviceInfoModalState>()(
  immer((set) => ({
    searchText: '',

    setSearchText: (text) => set((state) => {
      state.searchText = text;
    }),

    reset: () => set((state) => {
      state.searchText = '';
    }),
  }))
);
