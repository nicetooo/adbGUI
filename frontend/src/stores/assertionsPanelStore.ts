import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

interface AssertionResultDisplay {
  id: string;
  name: string;
  passed: boolean;
  message: string;
  executedAt: number;
  duration: number;
  matchedCount: number;
}

interface StoredAssertionDisplay {
  id: string;
  name: string;
  type: string;
  createdAt: number;
}

interface AssertionsPanelState {
  loading: boolean;
  results: AssertionResultDisplay[];
  customModalOpen: boolean;
  availableEventTypes: string[];
  loadingEventTypes: boolean;
  previewCount: number | null;
  previewLoading: boolean;
  storedAssertions: StoredAssertionDisplay[];
  loadingStored: boolean;
  editModalOpen: boolean;
  editingAssertion: any;

  // Actions
  setLoading: (loading: boolean) => void;
  setResults: (results: AssertionResultDisplay[]) => void;
  addResult: (result: AssertionResultDisplay) => void;
  removeResult: (id: string) => void;
  clearAllResults: () => void;
  setCustomModalOpen: (open: boolean) => void;
  setAvailableEventTypes: (types: string[]) => void;
  setLoadingEventTypes: (loading: boolean) => void;
  setPreviewCount: (count: number | null) => void;
  setPreviewLoading: (loading: boolean) => void;
  setStoredAssertions: (assertions: StoredAssertionDisplay[]) => void;
  removeStoredAssertion: (id: string) => void;
  setLoadingStored: (loading: boolean) => void;
  setEditModalOpen: (open: boolean) => void;
  setEditingAssertion: (assertion: any) => void;
  reset: () => void;
}

export const useAssertionsPanelStore = create<AssertionsPanelState>()(
  immer((set) => ({
    loading: false,
    results: [],
    customModalOpen: false,
    availableEventTypes: [],
    loadingEventTypes: false,
    previewCount: null,
    previewLoading: false,
    storedAssertions: [],
    loadingStored: false,
    editModalOpen: false,
    editingAssertion: null,

    setLoading: (loading) => set((state) => {
      state.loading = loading;
    }),

    setResults: (results) => set((state) => {
      state.results = results;
    }),

    addResult: (result) => set((state) => {
      state.results = [result, ...state.results];
    }),

    removeResult: (id) => set((state) => {
      state.results = state.results.filter(r => r.id !== id);
    }),

    clearAllResults: () => set((state) => {
      state.results = [];
    }),

    setCustomModalOpen: (open) => set((state) => {
      state.customModalOpen = open;
    }),

    setAvailableEventTypes: (types) => set((state) => {
      state.availableEventTypes = types;
    }),

    setLoadingEventTypes: (loading) => set((state) => {
      state.loadingEventTypes = loading;
    }),

    setPreviewCount: (count) => set((state) => {
      state.previewCount = count;
    }),

    setPreviewLoading: (loading) => set((state) => {
      state.previewLoading = loading;
    }),

    setStoredAssertions: (assertions) => set((state) => {
      state.storedAssertions = assertions;
    }),

    removeStoredAssertion: (id) => set((state) => {
      state.storedAssertions = state.storedAssertions.filter(a => a.id !== id);
    }),

    setLoadingStored: (loading) => set((state) => {
      state.loadingStored = loading;
    }),

    setEditModalOpen: (open) => set((state) => {
      state.editModalOpen = open;
    }),

    setEditingAssertion: (assertion) => set((state) => {
      state.editingAssertion = assertion;
    }),

    reset: () => set((state) => {
      state.loading = false;
      state.results = [];
      state.customModalOpen = false;
      state.availableEventTypes = [];
      state.loadingEventTypes = false;
      state.previewCount = null;
      state.previewLoading = false;
      state.storedAssertions = [];
      state.loadingStored = false;
      state.editModalOpen = false;
      state.editingAssertion = null;
    }),
  }))
);
