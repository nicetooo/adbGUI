/**
 * Event Timeline Store - EventTimeline 组件的状态管理
 */

import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';
import type { UnifiedEvent, DeviceSession, EventSource, EventCategory, EventLevel } from './eventTypes';

interface EventTimelineState {
  // 选中事件
  selectedEventId: string | null;
  selectedEventFull: UnifiedEvent | null;
  detailOpen: boolean;

  // 搜索和筛选
  searchText: string;
  filterSources: EventSource[];
  filterCategories: EventCategory[];
  filterLevels: EventLevel[];

  // 时间控制
  timeRange: { start: number; end: number } | null;
  currentTime: number;

  // Session 列表
  sessionList: DeviceSession[];

  // 面板状态
  assertionsPanelOpen: boolean;
  configModalOpen: boolean;
  bookmarkModalOpen: boolean;
  bookmarkLabel: string;

  // TimeRuler 拖拽状态
  isDragging: boolean;
  dragStart: number | null;
  dragEnd: number | null;

  // NetworkEventDetail Tab 状态
  networkDetailActiveTab: string;

  // Actions
  setSelectedEventId: (id: string | null) => void;
  setSelectedEventFull: (event: UnifiedEvent | null) => void;
  setDetailOpen: (open: boolean) => void;
  openEventDetail: (eventId: string, event?: UnifiedEvent) => void;
  closeEventDetail: () => void;

  setSearchText: (text: string) => void;
  setFilterSources: (sources: EventSource[]) => void;
  setFilterCategories: (categories: EventCategory[]) => void;
  setFilterLevels: (levels: EventLevel[]) => void;
  clearFilters: () => void;

  setTimeRange: (range: { start: number; end: number } | null) => void;
  setCurrentTime: (time: number) => void;

  setSessionList: (list: DeviceSession[]) => void;

  setAssertionsPanelOpen: (open: boolean) => void;
  toggleAssertionsPanel: () => void;
  setConfigModalOpen: (open: boolean) => void;
  setBookmarkModalOpen: (open: boolean) => void;
  setBookmarkLabel: (label: string) => void;
  openBookmarkModal: (defaultLabel: string) => void;
  closeBookmarkModal: () => void;

  // TimeRuler 拖拽
  setIsDragging: (dragging: boolean) => void;
  setDragStart: (start: number | null) => void;
  setDragEnd: (end: number | null) => void;
  startDrag: (time: number) => void;
  updateDrag: (time: number) => void;
  endDrag: () => { start: number; end: number } | null;

  // NetworkEventDetail
  setNetworkDetailActiveTab: (tab: string) => void;

  // 重置
  reset: () => void;
}

const initialState = {
  selectedEventId: null,
  selectedEventFull: null,
  detailOpen: false,
  searchText: '',
  filterSources: [] as EventSource[],
  filterCategories: [] as EventCategory[],
  filterLevels: [] as EventLevel[],
  timeRange: null,
  currentTime: 0,
  sessionList: [] as DeviceSession[],
  assertionsPanelOpen: false,
  configModalOpen: false,
  bookmarkModalOpen: false,
  bookmarkLabel: '',
  isDragging: false,
  dragStart: null,
  dragEnd: null,
  networkDetailActiveTab: 'overview',
};

export const useEventTimelineStore = create<EventTimelineState>()(
  immer((set, get) => ({
    ...initialState,

    // 选中事件
    setSelectedEventId: (id) => set((state) => {
      state.selectedEventId = id;
    }),

    setSelectedEventFull: (event) => set((state) => {
      state.selectedEventFull = event;
    }),

    setDetailOpen: (open) => set((state) => {
      state.detailOpen = open;
    }),

    openEventDetail: (eventId, event) => set((state) => {
      state.selectedEventId = eventId;
      state.detailOpen = true;
      if (event) {
        state.selectedEventFull = event;
      }
    }),

    closeEventDetail: () => set((state) => {
      state.detailOpen = false;
    }),

    // 搜索和筛选
    setSearchText: (text) => set((state) => {
      state.searchText = text;
    }),

    setFilterSources: (sources) => set((state) => {
      state.filterSources = sources;
    }),

    setFilterCategories: (categories) => set((state) => {
      state.filterCategories = categories;
    }),

    setFilterLevels: (levels) => set((state) => {
      state.filterLevels = levels;
    }),

    clearFilters: () => set((state) => {
      state.searchText = '';
      state.filterSources = [];
      state.filterCategories = [];
      state.filterLevels = [];
      state.timeRange = null;
    }),

    // 时间控制
    setTimeRange: (range) => set((state) => {
      state.timeRange = range;
    }),

    setCurrentTime: (time) => set((state) => {
      state.currentTime = time;
    }),

    // Session 列表
    setSessionList: (list) => set((state) => {
      state.sessionList = list;
    }),

    // 面板状态
    setAssertionsPanelOpen: (open) => set((state) => {
      state.assertionsPanelOpen = open;
    }),

    toggleAssertionsPanel: () => set((state) => {
      state.assertionsPanelOpen = !state.assertionsPanelOpen;
    }),

    setConfigModalOpen: (open) => set((state) => {
      state.configModalOpen = open;
    }),

    setBookmarkModalOpen: (open) => set((state) => {
      state.bookmarkModalOpen = open;
    }),

    setBookmarkLabel: (label) => set((state) => {
      state.bookmarkLabel = label;
    }),

    openBookmarkModal: (defaultLabel) => set((state) => {
      state.bookmarkLabel = defaultLabel;
      state.bookmarkModalOpen = true;
    }),

    closeBookmarkModal: () => set((state) => {
      state.bookmarkModalOpen = false;
      state.bookmarkLabel = '';
    }),

    // TimeRuler 拖拽
    setIsDragging: (dragging) => set((state) => {
      state.isDragging = dragging;
    }),

    setDragStart: (start) => set((state) => {
      state.dragStart = start;
    }),

    setDragEnd: (end) => set((state) => {
      state.dragEnd = end;
    }),

    startDrag: (time) => set((state) => {
      state.isDragging = true;
      state.dragStart = time;
      state.dragEnd = time;
    }),

    updateDrag: (time) => set((state) => {
      if (state.isDragging) {
        state.dragEnd = time;
      }
    }),

    endDrag: () => {
      const { isDragging, dragStart, dragEnd } = get();
      if (!isDragging || dragStart === null || dragEnd === null) {
        set((state) => {
          state.isDragging = false;
        });
        return null;
      }

      const start = Math.min(dragStart, dragEnd);
      const end = Math.max(dragStart, dragEnd);

      set((state) => {
        state.isDragging = false;
        state.dragStart = null;
        state.dragEnd = null;
      });

      return { start, end };
    },

    // NetworkEventDetail
    setNetworkDetailActiveTab: (tab) => set((state) => {
      state.networkDetailActiveTab = tab;
    }),

    // 重置
    reset: () => set(() => ({ ...initialState })),
  }))
);
