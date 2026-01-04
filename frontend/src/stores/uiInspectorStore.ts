import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

interface UIInspectorState {
  selectedNode: any | null;
  searchText: string;
  expandedKeys: React.Key[];
  autoExpandParent: boolean;
  selectedAction: string;
  inputText: string;
  searchMode: "auto" | "xpath" | "advanced";
  
  setSelectedNode: (node: any | null) => void;
  setSearchText: (text: string) => void;
  setExpandedKeys: (keys: React.Key[]) => void;
  setAutoExpandParent: (auto: boolean) => void;
  setSelectedAction: (action: string) => void;
  setInputText: (text: string) => void;
  setSearchMode: (mode: "auto" | "xpath" | "advanced") => void;
}

export const useUIInspectorStore = create<UIInspectorState>()(
  immer((set) => ({
    selectedNode: null,
    searchText: "",
    expandedKeys: [],
    autoExpandParent: true,
    selectedAction: "click",
    inputText: "",
    searchMode: "auto",
    
    setSelectedNode: (node) => set((state: UIInspectorState) => {
      state.selectedNode = node;
    }),
    setSearchText: (text) => set((state: UIInspectorState) => {
      state.searchText = text;
    }),
    setExpandedKeys: (keys) => set((state: UIInspectorState) => {
      state.expandedKeys = keys;
    }),
    setAutoExpandParent: (auto) => set((state: UIInspectorState) => {
      state.autoExpandParent = auto;
    }),
    setSelectedAction: (action) => set((state: UIInspectorState) => {
      state.selectedAction = action;
    }),
    setInputText: (text) => set((state: UIInspectorState) => {
      state.inputText = text;
    }),
    setSearchMode: (mode) => set((state: UIInspectorState) => {
      state.searchMode = mode;
    }),
  }))
);
