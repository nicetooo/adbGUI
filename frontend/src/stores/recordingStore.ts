import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';
import { TouchScript } from './automationStore';

interface RecordingState {
  selectedScriptNames: string[];
  saveModalVisible: boolean;
  scriptName: string;
  renameModalVisible: boolean;
  editingScriptName: string;
  newScriptName: string;
  selectedScript: TouchScript | null;
  
  setSelectedScriptNames: (names: string[]) => void;
  toggleScriptSelection: (name: string) => void;
  selectAllScripts: (allNames: string[]) => void;
  deselectAllScripts: () => void;
  setSaveModalVisible: (visible: boolean) => void;
  setScriptName: (name: string) => void;
  setRenameModalVisible: (visible: boolean) => void;
  setEditingScriptName: (name: string) => void;
  setNewScriptName: (name: string) => void;
  setSelectedScript: (script: TouchScript | null) => void;
  resetSaveModal: () => void;
  resetRenameModal: () => void;
}

export const useRecordingStore = create<RecordingState>()(
  immer((set) => ({
    selectedScriptNames: [],
    saveModalVisible: false,
    scriptName: "",
    renameModalVisible: false,
    editingScriptName: "",
    newScriptName: "",
    selectedScript: null,
    
    setSelectedScriptNames: (names) => set((state: RecordingState) => {
      state.selectedScriptNames = names;
    }),
    
    toggleScriptSelection: (name) => set((state: RecordingState) => {
      const idx = state.selectedScriptNames.indexOf(name);
      if (idx !== -1) {
        state.selectedScriptNames.splice(idx, 1);
      } else {
        state.selectedScriptNames.push(name);
      }
    }),
    
    selectAllScripts: (allNames) => set((state: RecordingState) => {
      state.selectedScriptNames = allNames;
    }),
    
    deselectAllScripts: () => set((state: RecordingState) => {
      state.selectedScriptNames = [];
    }),
    
    setSaveModalVisible: (visible) => set((state: RecordingState) => {
      state.saveModalVisible = visible;
    }),
    
    setScriptName: (name) => set((state: RecordingState) => {
      state.scriptName = name;
    }),
    
    setRenameModalVisible: (visible) => set((state: RecordingState) => {
      state.renameModalVisible = visible;
    }),
    
    setEditingScriptName: (name) => set((state: RecordingState) => {
      state.editingScriptName = name;
    }),
    
    setNewScriptName: (name) => set((state: RecordingState) => {
      state.newScriptName = name;
    }),
    
    setSelectedScript: (script) => set((state: RecordingState) => {
      state.selectedScript = script;
    }),
    
    resetSaveModal: () => set((state: RecordingState) => {
      state.saveModalVisible = false;
      state.scriptName = "";
    }),
    
    resetRenameModal: () => set((state: RecordingState) => {
      state.renameModalVisible = false;
      state.editingScriptName = "";
      state.newScriptName = "";
    }),
  }))
);
