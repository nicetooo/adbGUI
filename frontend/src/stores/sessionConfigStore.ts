/**
 * Session Config Store - SessionConfigModal state management
 */

import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';
import { SessionConfig, defaultSessionConfig } from './eventTypes';

interface SessionConfigState {
  // State
  config: SessionConfig;
  packages: string[];
  loadingPackages: boolean;
  
  // Actions
  setConfig: (config: SessionConfig) => void;
  updateConfig: (path: string[], value: any) => void;
  setPackages: (packages: string[]) => void;
  setLoadingPackages: (loading: boolean) => void;
  resetConfig: () => void;
}

export const useSessionConfigStore = create<SessionConfigState>()(
  immer((set) => ({
    // Initial state
    config: defaultSessionConfig,
    packages: [],
    loadingPackages: false,
    
    // Actions
    setConfig: (config) => set((state) => {
      state.config = config;
    }),
    
    updateConfig: (path, value) => set((state) => {
      let current: any = state.config;
      for (let i = 0; i < path.length - 1; i++) {
        current = current[path[i]];
      }
      current[path[path.length - 1]] = value;
    }),
    
    setPackages: (packages) => set((state) => {
      state.packages = packages;
    }),
    
    setLoadingPackages: (loading) => set((state) => {
      state.loadingPackages = loading;
    }),
    
    resetConfig: () => set((state) => {
      state.config = defaultSessionConfig;
    }),
  }))
);
