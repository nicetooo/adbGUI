import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

// Types
interface CauseCandidate {
  eventId: string;
  eventType: string;
  eventTitle: string;
  timestamp: number;
  explanation: string;
  probability: number;
  category: string;
}

interface RelatedEvent {
  id: string;
  type: string;
  title: string;
  timestamp: number;
  source: string;
  level: string;
}

interface RootCauseAnalysis {
  crashEventId: string;
  crashType: string;
  crashMessage: string;
  crashTime: number;
  probableCauses: CauseCandidate[];
  relatedEvents: RelatedEvent[];
  summary: string;
  confidence: number;
  recommendations: string[];
}

interface CrashEvent {
  id: string;
  type: string;
  title: string;
  timestamp: number;
  data: {
    exception?: string;
    stackTrace?: string;
    packageName?: string;
    processName?: string;
  };
}

interface CrashAnalysisState {
  selectedCrash: CrashEvent | null;
  analysis: RootCauseAnalysis | null;
  isAnalyzing: boolean;
  error: string | null;

  // Actions
  setSelectedCrash: (crash: CrashEvent | null) => void;
  setAnalysis: (analysis: RootCauseAnalysis | null) => void;
  setIsAnalyzing: (analyzing: boolean) => void;
  setError: (error: string | null) => void;
  reset: () => void;
}

export const useCrashAnalysisStore = create<CrashAnalysisState>()(
  immer((set) => ({
    selectedCrash: null,
    analysis: null,
    isAnalyzing: false,
    error: null,

    setSelectedCrash: (crash) => set((state) => {
      state.selectedCrash = crash;
    }),

    setAnalysis: (analysis) => set((state) => {
      state.analysis = analysis;
    }),

    setIsAnalyzing: (analyzing) => set((state) => {
      state.isAnalyzing = analyzing;
    }),

    setError: (error) => set((state) => {
      state.error = error;
    }),

    reset: () => set((state) => {
      state.selectedCrash = null;
      state.analysis = null;
      state.isAnalyzing = false;
      state.error = null;
    }),
  }))
);
