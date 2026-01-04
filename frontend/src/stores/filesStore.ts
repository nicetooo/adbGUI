import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

interface FilesState {
  currentPath: string;
  fileList: any[];
  filesLoading: boolean;
  clipboard: { path: string; type: "copy" | "cut" } | null;
  showHiddenFiles: boolean;
  
  setCurrentPath: (path: string) => void;
  setFileList: (files: any[]) => void;
  setFilesLoading: (loading: boolean) => void;
  setClipboard: (clipboard: { path: string; type: "copy" | "cut" } | null) => void;
  setShowHiddenFiles: (show: boolean) => void;
  clearClipboard: () => void;
}

export const useFilesStore = create<FilesState>()(
  immer((set) => ({
    currentPath: "/sdcard",
    fileList: [],
    filesLoading: false,
    clipboard: null,
    showHiddenFiles: false,
    
    setCurrentPath: (path) => set((state: FilesState) => {
      state.currentPath = path;
    }),
    setFileList: (files) => set((state: FilesState) => {
      state.fileList = files;
    }),
    setFilesLoading: (loading) => set((state: FilesState) => {
      state.filesLoading = loading;
    }),
    setClipboard: (clipboard) => set((state: FilesState) => {
      state.clipboard = clipboard;
    }),
    setShowHiddenFiles: (show) => set((state: FilesState) => {
      state.showHiddenFiles = show;
    }),
    clearClipboard: () => set((state: FilesState) => {
      state.clipboard = null;
    }),
  }))
);
