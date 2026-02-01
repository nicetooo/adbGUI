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

// Assertion Set types
export interface AssertionSetDisplay {
  id: string;
  name: string;
  description: string;
  assertions: string[];
  createdAt: number;
  updatedAt: number;
}

export interface AssertionSetSummaryDisplay {
  total: number;
  passed: number;
  failed: number;
  error: number;
  passRate: number;
}

export interface AssertionSetResultItemDisplay {
  id: string;
  assertionId: string;
  assertionName: string;
  sessionId: string;
  passed: boolean;
  message: string;
  actualValue?: any;
  expectedValue?: any;
  executedAt: number;
  duration: number;
}

export interface AssertionSetResultDisplay {
  id: string;
  setId: string;
  setName: string;
  sessionId: string;
  deviceId: string;
  executionId: string;
  startTime: number;
  endTime: number;
  duration: number;
  status: string;
  summary: AssertionSetSummaryDisplay;
  results: AssertionSetResultItemDisplay[];
  executedAt: number;
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

  // Assertion Sets state
  assertionSets: AssertionSetDisplay[];
  loadingSets: boolean;
  selectedSet: AssertionSetDisplay | null;
  setExecutionResults: AssertionSetResultDisplay[];
  loadingExecution: boolean;
  setCreateModalOpen: boolean;
  setDetailModalOpen: boolean;
  editingSet: AssertionSetDisplay | null;

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

  // Assertion Sets actions
  setAssertionSets: (sets: AssertionSetDisplay[]) => void;
  addAssertionSet: (set: AssertionSetDisplay) => void;
  updateAssertionSet: (id: string, set: Partial<AssertionSetDisplay>) => void;
  removeAssertionSet: (id: string) => void;
  setLoadingSets: (loading: boolean) => void;
  setSelectedSet: (set: AssertionSetDisplay | null) => void;
  setSetExecutionResults: (results: AssertionSetResultDisplay[]) => void;
  addSetExecutionResult: (result: AssertionSetResultDisplay) => void;
  setLoadingExecution: (loading: boolean) => void;
  setSetCreateModalOpen: (open: boolean) => void;
  setSetDetailModalOpen: (open: boolean) => void;
  setEditingSet: (set: AssertionSetDisplay | null) => void;

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

    // Assertion Sets initial state
    assertionSets: [],
    loadingSets: false,
    selectedSet: null,
    setExecutionResults: [],
    loadingExecution: false,
    setCreateModalOpen: false,
    setDetailModalOpen: false,
    editingSet: null,

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

    // Assertion Sets actions
    setAssertionSets: (sets) => set((state) => {
      state.assertionSets = sets;
    }),

    addAssertionSet: (newSet) => set((state) => {
      state.assertionSets = [newSet, ...state.assertionSets];
    }),

    updateAssertionSet: (id, updates) => set((state) => {
      const idx = state.assertionSets.findIndex(s => s.id === id);
      if (idx !== -1) {
        state.assertionSets[idx] = { ...state.assertionSets[idx], ...updates };
      }
    }),

    removeAssertionSet: (id) => set((state) => {
      state.assertionSets = state.assertionSets.filter(s => s.id !== id);
    }),

    setLoadingSets: (loading) => set((state) => {
      state.loadingSets = loading;
    }),

    setSelectedSet: (selectedSet) => set((state) => {
      state.selectedSet = selectedSet;
    }),

    setSetExecutionResults: (results) => set((state) => {
      state.setExecutionResults = results;
    }),

    addSetExecutionResult: (result) => set((state) => {
      state.setExecutionResults = [result, ...state.setExecutionResults];
    }),

    setLoadingExecution: (loading) => set((state) => {
      state.loadingExecution = loading;
    }),

    setSetCreateModalOpen: (open) => set((state) => {
      state.setCreateModalOpen = open;
    }),

    setSetDetailModalOpen: (open) => set((state) => {
      state.setDetailModalOpen = open;
    }),

    setEditingSet: (editingSet) => set((state) => {
      state.editingSet = editingSet;
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
      state.assertionSets = [];
      state.loadingSets = false;
      state.selectedSet = null;
      state.setExecutionResults = [];
      state.loadingExecution = false;
      state.setCreateModalOpen = false;
      state.setDetailModalOpen = false;
      state.editingSet = null;
    }),
  }))
);
