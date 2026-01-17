import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';
import type { TouchScript, ScriptTask } from './automationStore';

interface AutomationViewState {
  // Selection
  selectedScriptNames: string[];
  selectedTaskNames: string[];
  
  // Save modal
  saveModalVisible: boolean;
  scriptName: string;
  
  // Rename modal
  renameModalVisible: boolean;
  editingScriptName: string;
  newScriptName: string;
  
  // Selected script preview
  selectedScript: TouchScript | null;
  
  // Task modal
  taskModalVisible: boolean;
  editingTask: ScriptTask | null;
  
  // Active tab
  activeTab: string;
  
  // Actions
  setSelectedScriptNames: (names: string[]) => void;
  toggleSelectedScriptName: (name: string, selected: boolean) => void;
  setSelectedTaskNames: (names: string[]) => void;
  toggleSelectedTaskName: (name: string, selected: boolean) => void;
  setSaveModalVisible: (visible: boolean) => void;
  setScriptName: (name: string) => void;
  setRenameModalVisible: (visible: boolean) => void;
  setEditingScriptName: (name: string) => void;
  setNewScriptName: (name: string) => void;
  setSelectedScript: (script: TouchScript | null) => void;
  setTaskModalVisible: (visible: boolean) => void;
  setEditingTask: (task: ScriptTask | null) => void;
  setActiveTab: (tab: string) => void;
  reset: () => void;
}

export const useAutomationViewStore = create<AutomationViewState>()(
  immer((set) => ({
    selectedScriptNames: [],
    selectedTaskNames: [],
    saveModalVisible: false,
    scriptName: '',
    renameModalVisible: false,
    editingScriptName: '',
    newScriptName: '',
    selectedScript: null,
    taskModalVisible: false,
    editingTask: null,
    activeTab: 'scripts',

    setSelectedScriptNames: (names) => set((state) => {
      state.selectedScriptNames = names;
    }),

    toggleSelectedScriptName: (name, selected) => set((state) => {
      if (selected) {
        state.selectedScriptNames = [...state.selectedScriptNames, name];
      } else {
        state.selectedScriptNames = state.selectedScriptNames.filter(n => n !== name);
      }
    }),

    setSelectedTaskNames: (names) => set((state) => {
      state.selectedTaskNames = names;
    }),

    toggleSelectedTaskName: (name, selected) => set((state) => {
      if (selected) {
        state.selectedTaskNames = [...state.selectedTaskNames, name];
      } else {
        state.selectedTaskNames = state.selectedTaskNames.filter(n => n !== name);
      }
    }),

    setSaveModalVisible: (visible) => set((state) => {
      state.saveModalVisible = visible;
    }),

    setScriptName: (name) => set((state) => {
      state.scriptName = name;
    }),

    setRenameModalVisible: (visible) => set((state) => {
      state.renameModalVisible = visible;
    }),

    setEditingScriptName: (name) => set((state) => {
      state.editingScriptName = name;
    }),

    setNewScriptName: (name) => set((state) => {
      state.newScriptName = name;
    }),

    setSelectedScript: (script) => set((state) => {
      state.selectedScript = script;
    }),

    setTaskModalVisible: (visible) => set((state) => {
      state.taskModalVisible = visible;
    }),

    setEditingTask: (task) => set((state) => {
      state.editingTask = task;
    }),

    setActiveTab: (tab) => set((state) => {
      state.activeTab = tab;
    }),

    reset: () => set((state) => {
      state.selectedScriptNames = [];
      state.selectedTaskNames = [];
      state.saveModalVisible = false;
      state.scriptName = '';
      state.renameModalVisible = false;
      state.editingScriptName = '';
      state.newScriptName = '';
      state.selectedScript = null;
      state.taskModalVisible = false;
      state.editingTask = null;
      state.activeTab = 'scripts';
    }),
  }))
);
