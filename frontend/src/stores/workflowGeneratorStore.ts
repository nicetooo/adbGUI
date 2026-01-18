/**
 * Workflow Generator Store - WorkflowGenerator 组件的状态管理
 */

import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

// GeneratedStep is a simplified step structure used during workflow generation
// It differs from the full WorkflowStep type as it's a lightweight representation
// for AI-generated workflows before conversion to the full format
interface GeneratedStep {
  id: string;
  type: string;
  name: string;
  selector?: { type: string; value: string };
  value?: string;
  timeout?: number;
}

interface LoopPattern {
  startIndex: number;
  endIndex: number;
  iterations: number;
  confidence: number;
}

interface BranchPattern {
  condition: string;
  thenSteps: number[];
  elseSteps: number[];
  confidence: number;
}

interface GeneratedWorkflow {
  sessionId: string;
  name: string;
  description: string;
  steps: GeneratedStep[];
  suggestions: string[];
  confidence: number;
  loops: LoopPattern[];
  branches: BranchPattern[];
}

interface AILogEntry {
  stage: string;
  message: string;
  timestamp: number;
  frameBase64?: string;
  frameIndex?: number;
  frameTimeMs?: number;
  sceneType?: string;
  description?: string;
  ocrText?: string[];
}

interface WorkflowGeneratorConfig {
  includeWaits: boolean;
  optimizeSelectors: boolean;
  detectLoops: boolean;
  detectBranches: boolean;
  generateAssertions: boolean;
  minConfidence: number;
  useVideoAnalysis: boolean;
}

interface WorkflowGeneratorState {
  // 生成状态
  isGenerating: boolean;
  progress: number;
  progressMessage: string;
  progressStage: string;
  error: string | null;
  
  // 生成结果
  generatedWorkflow: GeneratedWorkflow | null;
  
  // 编辑弹窗
  editModalVisible: boolean;
  editingStep: GeneratedStep | null;
  editingIndex: number;
  
  // 配置弹窗
  configModalVisible: boolean;
  
  // AI 日志
  aiLogs: AILogEntry[];
  
  // 配置
  config: WorkflowGeneratorConfig;
  
  // Actions
  setIsGenerating: (generating: boolean) => void;
  setProgress: (progress: number) => void;
  setProgressMessage: (message: string) => void;
  setProgressStage: (stage: string) => void;
  setError: (error: string | null) => void;
  setGeneratedWorkflow: (workflow: GeneratedWorkflow | null) => void;
  
  // 编辑弹窗
  openEditModal: (step: GeneratedStep, index: number) => void;
  closeEditModal: () => void;
  setEditingStep: (step: GeneratedStep | null) => void;
  updateEditingStep: (updates: Partial<GeneratedStep>) => void;
  
  // 配置弹窗
  openConfigModal: () => void;
  closeConfigModal: () => void;
  
  // AI 日志
  addAiLog: (log: AILogEntry) => void;
  clearAiLogs: () => void;
  
  // 配置
  setConfig: (config: Partial<WorkflowGeneratorConfig>) => void;
  
  // 重置
  resetGeneration: () => void;
  reset: () => void;
}

const defaultConfig: WorkflowGeneratorConfig = {
  includeWaits: true,
  optimizeSelectors: true,
  detectLoops: true,
  detectBranches: true,
  generateAssertions: false,
  minConfidence: 0.7,
  useVideoAnalysis: true,
};

export const useWorkflowGeneratorStore = create<WorkflowGeneratorState>()(
  immer((set) => ({
    // 初始状态
    isGenerating: false,
    progress: 0,
    progressMessage: '',
    progressStage: '',
    error: null,
    generatedWorkflow: null,
    editModalVisible: false,
    editingStep: null,
    editingIndex: -1,
    configModalVisible: false,
    aiLogs: [],
    config: defaultConfig,
    
    // Actions
    setIsGenerating: (generating) => set((state) => {
      state.isGenerating = generating;
    }),
    
    setProgress: (progress) => set((state) => {
      state.progress = progress;
    }),
    
    setProgressMessage: (message) => set((state) => {
      state.progressMessage = message;
    }),
    
    setProgressStage: (stage) => set((state) => {
      state.progressStage = stage;
    }),
    
    setError: (error) => set((state) => {
      state.error = error;
    }),
    
    setGeneratedWorkflow: (workflow) => set((state) => {
      state.generatedWorkflow = workflow;
    }),
    
    // 编辑弹窗
    openEditModal: (step, index) => set((state) => {
      state.editingStep = step;
      state.editingIndex = index;
      state.editModalVisible = true;
    }),
    
    closeEditModal: () => set((state) => {
      state.editModalVisible = false;
      state.editingStep = null;
      state.editingIndex = -1;
    }),
    
    setEditingStep: (step) => set((state) => {
      state.editingStep = step;
    }),
    
    updateEditingStep: (updates) => set((state) => {
      if (state.editingStep) {
        state.editingStep = { ...state.editingStep, ...updates };
      }
    }),
    
    // 配置弹窗
    openConfigModal: () => set((state) => {
      state.configModalVisible = true;
    }),
    
    closeConfigModal: () => set((state) => {
      state.configModalVisible = false;
    }),
    
    // AI 日志
    addAiLog: (log) => set((state) => {
      state.aiLogs.push(log);
    }),
    
    clearAiLogs: () => set((state) => {
      state.aiLogs = [];
    }),
    
    // 配置
    setConfig: (configUpdate) => set((state) => {
      state.config = { ...state.config, ...configUpdate };
    }),
    
    // 重置生成状态
    resetGeneration: () => set((state) => {
      state.isGenerating = false;
      state.progress = 0;
      state.progressMessage = '';
      state.progressStage = '';
      state.error = null;
      state.aiLogs = [];
    }),
    
    // 完全重置
    reset: () => set((state) => {
      state.isGenerating = false;
      state.progress = 0;
      state.progressMessage = '';
      state.progressStage = '';
      state.error = null;
      state.generatedWorkflow = null;
      state.editModalVisible = false;
      state.editingStep = null;
      state.editingIndex = -1;
      state.configModalVisible = false;
      state.aiLogs = [];
      state.config = defaultConfig;
    }),
  }))
);
