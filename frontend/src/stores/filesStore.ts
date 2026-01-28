import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

interface FilesState {
  currentPath: string;
  fileList: any[];
  filesLoading: boolean;
  clipboard: { path: string; type: "copy" | "cut" } | null;
  showHiddenFiles: boolean;
  
  // Drag & drop upload state
  isDraggingOver: boolean;
  isUploading: boolean;
  uploadingFileName: string;
  uploadProgress: { current: number; total: number } | null;
  
  setCurrentPath: (path: string) => void;
  setFileList: (files: any[]) => void;
  setFilesLoading: (loading: boolean) => void;
  setClipboard: (clipboard: { path: string; type: "copy" | "cut" } | null) => void;
  setShowHiddenFiles: (show: boolean) => void;
  clearClipboard: () => void;
  
  // Drag & drop upload actions
  setIsDraggingOver: (isDragging: boolean) => void;
  setIsUploading: (isUploading: boolean) => void;
  setUploadingFileName: (fileName: string) => void;
  setUploadProgress: (progress: { current: number; total: number } | null) => void;
}

export const useFilesStore = create<FilesState>()(
  immer((set) => ({
    currentPath: "/sdcard",
    fileList: [],
    filesLoading: false,
    clipboard: null,
    showHiddenFiles: false,
    
    // Drag & drop upload state
    isDraggingOver: false,
    isUploading: false,
    uploadingFileName: "",
    uploadProgress: null,
    
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
    
    // Drag & drop upload actions
    setIsDraggingOver: (isDragging) => set((state: FilesState) => {
      state.isDraggingOver = isDragging;
    }),
    setIsUploading: (isUploading) => set((state: FilesState) => {
      state.isUploading = isUploading;
    }),
    setUploadingFileName: (fileName) => set((state: FilesState) => {
      state.uploadingFileName = fileName;
    }),
    setUploadProgress: (progress) => set((state: FilesState) => {
      state.uploadProgress = progress;
    }),
  }))
);
