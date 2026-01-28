import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';
import { main } from '../types/wails-models';

interface AppsState {
  tableHeight: number;
  packages: main.AppPackage[];
  appsLoading: boolean;
  packageFilter: string;
  typeFilter: string;
  infoModalVisible: boolean;
  infoLoading: boolean;
  selectedAppInfo: main.AppPackage | null;
  permissionSearch: string;
  activitySearch: string;
  // Drag and drop state
  isDraggingOver: boolean;
  isInstalling: boolean;
  installingFileName: string;
  
  setTableHeight: (height: number) => void;
  setPackages: (packages: main.AppPackage[]) => void;
  setAppsLoading: (loading: boolean) => void;
  setPackageFilter: (filter: string) => void;
  setTypeFilter: (filter: string) => void;
  setInfoModalVisible: (visible: boolean) => void;
  setInfoLoading: (loading: boolean) => void;
  setSelectedAppInfo: (info: main.AppPackage | null) => void;
  setPermissionSearch: (search: string) => void;
  setActivitySearch: (search: string) => void;
  setIsDraggingOver: (isDragging: boolean) => void;
  setIsInstalling: (isInstalling: boolean, fileName?: string) => void;
}

export const useAppsStore = create<AppsState>()(
  immer((set) => ({
    tableHeight: 400,
    packages: [],
    appsLoading: false,
    packageFilter: "",
    typeFilter: "user",
    infoModalVisible: false,
    infoLoading: false,
    selectedAppInfo: null,
    permissionSearch: "",
    activitySearch: "",
    // Drag and drop state
    isDraggingOver: false,
    isInstalling: false,
    installingFileName: "",
    
    setTableHeight: (height) => set((state: AppsState) => {
      state.tableHeight = height;
    }),
    setPackages: (packages) => set((state: AppsState) => {
      state.packages = packages;
    }),
    setAppsLoading: (loading) => set((state: AppsState) => {
      state.appsLoading = loading;
    }),
    setPackageFilter: (filter) => set((state: AppsState) => {
      state.packageFilter = filter;
    }),
    setTypeFilter: (filter) => set((state: AppsState) => {
      state.typeFilter = filter;
    }),
    setInfoModalVisible: (visible) => set((state: AppsState) => {
      state.infoModalVisible = visible;
    }),
    setInfoLoading: (loading) => set((state: AppsState) => {
      state.infoLoading = loading;
    }),
    setSelectedAppInfo: (info) => set((state: AppsState) => {
      state.selectedAppInfo = info;
    }),
    setPermissionSearch: (search) => set((state: AppsState) => {
      state.permissionSearch = search;
    }),
    setActivitySearch: (search) => set((state: AppsState) => {
      state.activitySearch = search;
    }),
    setIsDraggingOver: (isDragging) => set((state: AppsState) => {
      state.isDraggingOver = isDragging;
    }),
    setIsInstalling: (isInstalling, fileName = "") => set((state: AppsState) => {
      state.isInstalling = isInstalling;
      state.installingFileName = fileName;
    }),
  }))
);
