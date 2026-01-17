/**
 * Wireless Connect Store - WirelessConnectModal state management
 */

import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

interface WirelessConnectState {
  // State
  loading: boolean;
  activeTab: string;
  connectUrl: string;
  
  // Actions
  setLoading: (loading: boolean) => void;
  setActiveTab: (tab: string) => void;
  setConnectUrl: (url: string) => void;
  reset: () => void;
}

export const useWirelessConnectStore = create<WirelessConnectState>()(
  immer((set) => ({
    // Initial state
    loading: false,
    activeTab: 'connect',
    connectUrl: '',
    
    // Actions
    setLoading: (loading) => set((state) => {
      state.loading = loading;
    }),
    
    setActiveTab: (tab) => set((state) => {
      state.activeTab = tab;
    }),
    
    setConnectUrl: (url) => set((state) => {
      state.connectUrl = url;
    }),
    
    reset: () => set((state) => {
      state.loading = false;
      state.activeTab = 'connect';
      state.connectUrl = '';
    }),
  }))
);
