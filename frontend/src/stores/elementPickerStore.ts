import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';
import React from 'react';

type SearchMode = 'auto' | 'xpath' | 'advanced';

interface ElementPickerState {
  // Selected node
  selectedNode: any;
  
  // Search
  searchText: string;
  searchMode: SearchMode;
  
  // Tree expansion
  expandedKeys: React.Key[];
  autoExpandParent: boolean;
  
  // Selector selection
  selectedSelectorIndex: number;
  
  // Point picker
  isPickingPoint: boolean;
  
  // Actions
  setSelectedNode: (node: any) => void;
  setSearchText: (text: string) => void;
  setSearchMode: (mode: SearchMode) => void;
  setExpandedKeys: (keys: React.Key[]) => void;
  setAutoExpandParent: (value: boolean) => void;
  setSelectedSelectorIndex: (index: number) => void;
  setIsPickingPoint: (picking: boolean) => void;
  
  // Reset
  reset: () => void;
}

export const useElementPickerStore = create<ElementPickerState>()(
  immer((set) => ({
    selectedNode: null,
    searchText: '',
    searchMode: 'auto',
    expandedKeys: [],
    autoExpandParent: true,
    selectedSelectorIndex: 0,
    isPickingPoint: false,

    setSelectedNode: (node) => set((state) => {
      state.selectedNode = node;
    }),

    setSearchText: (text) => set((state) => {
      state.searchText = text;
    }),

    setSearchMode: (mode) => set((state) => {
      state.searchMode = mode;
    }),

    setExpandedKeys: (keys) => set((state) => {
      state.expandedKeys = keys;
    }),

    setAutoExpandParent: (value) => set((state) => {
      state.autoExpandParent = value;
    }),

    setSelectedSelectorIndex: (index) => set((state) => {
      state.selectedSelectorIndex = index;
    }),

    setIsPickingPoint: (picking) => set((state) => {
      state.isPickingPoint = picking;
    }),

    reset: () => set((state) => {
      state.selectedNode = null;
      state.searchText = '';
      state.searchMode = 'auto';
      state.expandedKeys = [];
      state.autoExpandParent = true;
      state.selectedSelectorIndex = 0;
      state.isPickingPoint = false;
    }),
  }))
);
