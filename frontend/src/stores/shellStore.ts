import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

interface ShellState {
  shellCmd: string;
  shellOutput: string;
  history: string[];
  historyIndex: number;
  
  setShellCmd: (cmd: string) => void;
  setShellOutput: (output: string) => void;
  addToHistory: (cmd: string) => void;
  setHistoryIndex: (index: number) => void;
  navigateHistory: (direction: 'up' | 'down') => void;
  clearOutput: () => void;
}

export const useShellStore = create<ShellState>()(
  immer((set, get) => ({
    shellCmd: "",
    shellOutput: "",
    history: [],
    historyIndex: -1,
    
    setShellCmd: (cmd) => set((state: ShellState) => {
      state.shellCmd = cmd;
    }),
    
    setShellOutput: (output) => set((state: ShellState) => {
      state.shellOutput = output;
    }),
    
    addToHistory: (cmd) => set((state: ShellState) => {
      state.history = [cmd, ...state.history.filter(c => c !== cmd)].slice(0, 50);
      state.historyIndex = -1;
    }),
    
    setHistoryIndex: (index) => set((state: ShellState) => {
      state.historyIndex = index;
    }),
    
    navigateHistory: (direction) => {
      const { history, historyIndex } = get();
      if (direction === 'up') {
        if (historyIndex < history.length - 1) {
          const newIndex = historyIndex + 1;
          set((state: ShellState) => {
            state.historyIndex = newIndex;
            state.shellCmd = state.history[newIndex];
          });
        }
      } else if (direction === 'down') {
        if (historyIndex > 0) {
          const newIndex = historyIndex - 1;
          set((state: ShellState) => {
            state.historyIndex = newIndex;
            state.shellCmd = state.history[newIndex];
          });
        } else if (historyIndex === 0) {
          set((state: ShellState) => {
            state.historyIndex = -1;
            state.shellCmd = "";
          });
        }
      }
    },
    
    clearOutput: () => set((state: ShellState) => {
      state.shellOutput = "";
    }),
  }))
);
