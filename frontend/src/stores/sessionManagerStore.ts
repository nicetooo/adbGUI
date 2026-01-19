/**
 * Session Manager Store - SessionManager 组件的状态管理
 */

import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';
import type { DeviceSession } from './eventTypes';

interface SessionManagerState {
  // 数据
  sessions: DeviceSession[];
  loading: boolean;

  // 选中状态
  selectedSession: DeviceSession | null;
  selectedRowKeys: React.Key[];

  // 搜索和过滤
  searchText: string;
  statusFilter: string | null;

  // 详情弹窗
  detailModalOpen: boolean;

  // 重命名弹窗
  renameModalOpen: boolean;
  renameSession: DeviceSession | null;
  newName: string;

  // Actions
  setSessions: (sessions: DeviceSession[]) => void;
  setLoading: (loading: boolean) => void;
  setSelectedSession: (session: DeviceSession | null) => void;
  setSelectedRowKeys: (keys: React.Key[]) => void;
  setSearchText: (text: string) => void;
  setStatusFilter: (status: string | null) => void;

  // 详情弹窗
  openDetailModal: (session: DeviceSession) => void;
  closeDetailModal: () => void;

  // 重命名弹窗
  openRenameModal: (session: DeviceSession) => void;
  closeRenameModal: () => void;
  setNewName: (name: string) => void;

  // 重置
  reset: () => void;
}

export const useSessionManagerStore = create<SessionManagerState>()(
  immer((set) => ({
    // 初始状态
    sessions: [],
    loading: false,
    selectedSession: null,
    selectedRowKeys: [],
    searchText: '',
    statusFilter: null,
    detailModalOpen: false,
    renameModalOpen: false,
    renameSession: null,
    newName: '',
    // Actions
    setSessions: (sessions) => set((state) => {
      state.sessions = sessions;
    }),

    setLoading: (loading) => set((state) => {
      state.loading = loading;
    }),

    setSelectedSession: (session) => set((state) => {
      state.selectedSession = session;
    }),

    setSelectedRowKeys: (keys) => set((state) => {
      state.selectedRowKeys = keys;
    }),

    setSearchText: (text) => set((state) => {
      state.searchText = text;
    }),

    setStatusFilter: (status) => set((state) => {
      state.statusFilter = status;
    }),

    // 详情弹窗
    openDetailModal: (session) => set((state) => {
      state.selectedSession = session;
      state.detailModalOpen = true;
    }),

    closeDetailModal: () => set((state) => {
      state.detailModalOpen = false;
    }),

    // 重命名弹窗
    openRenameModal: (session) => set((state) => {
      state.renameSession = session;
      state.newName = session.name || '';
      state.renameModalOpen = true;
    }),

    closeRenameModal: () => set((state) => {
      state.renameModalOpen = false;
      state.renameSession = null;
      state.newName = '';
    }),

    setNewName: (name) => set((state) => {
      state.newName = name;
    }),

    // 重置
    reset: () => set((state) => {
      state.sessions = [];
      state.loading = false;
      state.selectedSession = null;
      state.selectedRowKeys = [];
      state.searchText = '';
      state.statusFilter = null;
      state.detailModalOpen = false;
      state.renameModalOpen = false;
      state.renameSession = null;
      state.newName = '';
    }),
  }))
);
