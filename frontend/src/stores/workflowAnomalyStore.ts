import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

interface WorkflowAnomalyState {
  selectedAction: string;
  rememberChoice: boolean;
  updateWorkflow: boolean;

  // Actions
  setSelectedAction: (action: string) => void;
  setRememberChoice: (remember: boolean) => void;
  setUpdateWorkflow: (update: boolean) => void;
  reset: () => void;
}

export const useWorkflowAnomalyStore = create<WorkflowAnomalyState>()(
  immer((set) => ({
    selectedAction: '',
    rememberChoice: false,
    updateWorkflow: false,

    setSelectedAction: (action) => set((state) => {
      state.selectedAction = action;
    }),

    setRememberChoice: (remember) => set((state) => {
      state.rememberChoice = remember;
    }),

    setUpdateWorkflow: (update) => set((state) => {
      state.updateWorkflow = update;
    }),

    reset: () => set((state) => {
      state.selectedAction = '';
      state.rememberChoice = false;
      state.updateWorkflow = false;
    }),
  }))
);
