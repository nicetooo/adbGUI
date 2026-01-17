/**
 * Smart Search Store - AI-powered search state management
 */

import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

// Parsed query result from AI
export interface NLParsedQuery {
  types?: string[];
  sources?: string[];
  levels?: string[];
  keywords?: string[];
  timeRange?: {
    startMs?: number;
    endMs?: number;
    last?: string;
  };
  context?: string;
}

export interface NLQueryResult {
  query: NLParsedQuery;
  explanation: string;
  confidence: number;
  suggestions?: string[];
}

interface SmartSearchState {
  // Parsing state
  isParsing: boolean;
  aiFilterActive: boolean;
  lastParsedResult: NLQueryResult | null;
  
  // Actions
  setIsParsing: (parsing: boolean) => void;
  setAiFilterActive: (active: boolean) => void;
  setLastParsedResult: (result: NLQueryResult | null) => void;
  clearAiFilter: () => void;
  reset: () => void;
}

export const useSmartSearchStore = create<SmartSearchState>()(
  immer((set) => ({
    // Initial state
    isParsing: false,
    aiFilterActive: false,
    lastParsedResult: null,
    
    // Actions
    setIsParsing: (parsing) => set((state) => {
      state.isParsing = parsing;
    }),
    
    setAiFilterActive: (active) => set((state) => {
      state.aiFilterActive = active;
    }),
    
    setLastParsedResult: (result) => set((state) => {
      state.lastParsedResult = result;
    }),
    
    clearAiFilter: () => set((state) => {
      state.aiFilterActive = false;
      state.lastParsedResult = null;
    }),
    
    reset: () => set((state) => {
      state.isParsing = false;
      state.aiFilterActive = false;
      state.lastParsedResult = null;
    }),
  }))
);
