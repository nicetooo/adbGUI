import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';
import type { Node, Edge } from '@xyflow/react';

// Import workflow types
import {
  Workflow,
  WorkflowStep,
  StepType,
  StepConnections,
  StepCommon,
  getDefaultStepCommon,
  getDefaultConnections,
  createEmptyWorkflow,
} from '../types/workflow';

// Re-export types for external use
export type { Workflow, WorkflowStep, StepType, StepConnections, StepCommon };

// Legacy type aliases for backward compatibility
// @deprecated Use Workflow instead
// @deprecated Use WorkflowStep instead
// These are kept for backward compatibility and will be removed in a future version

interface StepStatus {
  stepId: string;
  isWaiting: boolean;
  waitPhase?: 'pre' | 'post';
}

// Runtime context for workflow execution observation
export interface WorkflowRuntimeContext {
  deviceId: string;
  workflowId: string;
  workflowName: string;
  status: string;
  currentStepId: string;
  currentStepName: string;
  currentStepType: string;
  stepsExecuted: number;
  stepsTotal: number;
  isPaused: boolean;
  variables: Record<string, string>;
}

export type WorkflowViewMode = 'canvas' | 'list';

interface WorkflowState {
  // Core data
  workflows: Workflow[];
  selectedWorkflowId: string | null;
  nodes: Node[];
  edges: Edge[];
  
  // View mode
  viewMode: WorkflowViewMode;
  
  // Execution state
  isRunning: boolean;
  runningWorkflowIds: string[];
  isPaused: boolean;
  currentStepId: string | null;
  waitingStepId: string | null;
  waitingPhase: 'pre' | 'post' | null;
  executionLogs: string[];
  workflowStepMap: Record<string, StepStatus>;
  runtimeContext: WorkflowRuntimeContext | null;
  
  // UI state
  workflowModalVisible: boolean;
  drawerVisible: boolean;
  elementPickerVisible: boolean;
  variablesModalVisible: boolean;
  editingNodeId: string | null;
  editingWorkflowId: string | null;
  editingWorkflowName: string;
  
  // Temporary data
  tempVariables: { key: string; value: string }[];
  packages: any[];
  appsLoading: boolean;
  needsAutoSave: boolean;
  
  // Workflow operations
  setWorkflows: (workflows: Workflow[] | ((prev: Workflow[]) => Workflow[])) => void;
  setWorkflowsWithMigration: (workflows: unknown[]) => void;
  addWorkflow: (workflow: Workflow) => void;
  createWorkflow: (id: string, name: string) => Workflow;
  updateWorkflow: (id: string, updates: Partial<Workflow>) => void;
  deleteWorkflow: (id: string) => void;
  selectWorkflow: (id: string | null) => void;
  getSelectedWorkflow: () => Workflow | null;
  getWorkflowById: (id: string) => Workflow | undefined;
  
  // Step operations
  addStep: (workflowId: string, step: WorkflowStep) => void;
  updateStep: (workflowId: string, stepId: string, updates: Partial<WorkflowStep>) => void;
  deleteStep: (workflowId: string, stepId: string) => void;
  getStep: (workflowId: string, stepId: string) => WorkflowStep | undefined;
  updateStepConnections: (workflowId: string, stepId: string, connections: Partial<StepConnections>) => void;
  
  // Graph operations
  setNodes: (nodes: Node[]) => void;
  setEdges: (edges: Edge[]) => void;
  updateGraph: (nodes: Node[], edges: Edge[]) => void;
  
  // Execution state operations
  setIsRunning: (running: boolean) => void;
  setRunningWorkflowIds: (ids: string[] | ((prev: string[]) => string[])) => void;
  addRunningWorkflowId: (id: string) => void;
  removeRunningWorkflowId: (id: string) => void;
  setIsPaused: (paused: boolean) => void;
  setCurrentStepId: (stepId: string | null) => void;
  setWaitingStep: (stepId: string | null, phase: 'pre' | 'post' | null) => void;
  // Legacy compatibility methods
  setWaitingStepId: (stepId: string | null) => void;
  setWaitingPhase: (phase: 'pre' | 'post' | null) => void;
  setExecutionLogs: (logs: string[] | ((prev: string[]) => string[])) => void;
  addExecutionLog: (log: string) => void;
  clearExecutionLogs: () => void;
  setWorkflowStepMap: (map: Record<string, StepStatus> | ((prev: Record<string, StepStatus>) => Record<string, StepStatus>)) => void;
  updateWorkflowStepStatus: (workflowId: string, status: StepStatus) => void;
  clearWorkflowStepStatus: (workflowId: string) => void;
  setRuntimeContext: (context: WorkflowRuntimeContext | null) => void;
  clearRuntimeContext: () => void;
  
  // UI state operations
  setWorkflowModalVisible: (visible: boolean) => void;
  setDrawerVisible: (visible: boolean) => void;
  setElementPickerVisible: (visible: boolean) => void;
  setVariablesModalVisible: (visible: boolean) => void;
  setEditingNode: (nodeId: string | null) => void;
  // Legacy compatibility methods
  setEditingNodeId: (nodeId: string | null) => void;
  setEditingWorkflow: (id: string | null, name: string) => void;
  // Legacy compatibility methods
  setEditingWorkflowId: (id: string | null) => void;
  setEditingWorkflowName: (name: string) => void;
  setSelectedWorkflow: (workflow: Workflow | null) => void;
  
  // Temporary data operations
  setTempVariables: (vars: { key: string; value: string }[] | ((prev: { key: string; value: string }[]) => { key: string; value: string }[])) => void;
  setPackages: (packages: any[]) => void;
  setAppsLoading: (loading: boolean) => void;
  setNeedsAutoSave: (needs: boolean) => void;
  
  // View mode
  setViewMode: (mode: WorkflowViewMode) => void;
  
  // Export utilities
  exportWorkflow: (id: string) => Workflow | undefined;
}

export const useWorkflowStore = create<WorkflowState>()(
  immer((set, get) => ({
    // Initial state
    workflows: [],
    selectedWorkflowId: null,
    nodes: [],
    edges: [],
    
    viewMode: 'canvas' as WorkflowViewMode,
    
    isRunning: false,
    runningWorkflowIds: [],
    isPaused: false,
    currentStepId: null,
    waitingStepId: null,
    waitingPhase: null,
    executionLogs: [],
    workflowStepMap: {},
    runtimeContext: null,
    
    workflowModalVisible: false,
    drawerVisible: false,
    elementPickerVisible: false,
    variablesModalVisible: false,
    editingNodeId: null,
    editingWorkflowId: null,
    editingWorkflowName: "",
    
    tempVariables: [],
    packages: [],
    appsLoading: false,
    needsAutoSave: false,
    
    // Workflow operations
    setWorkflows: (workflows) => {
      if (typeof workflows === 'function') {
        set((state: WorkflowState) => {
          state.workflows = workflows(state.workflows);
        });
      } else {
        set({ workflows });
      }
    },
    
    setWorkflowsWithMigration: (workflows) => {
      // No migration needed - all workflows are now in the unified format
      set({ workflows: workflows as Workflow[] });
    },
    
    addWorkflow: (workflow) => set((state: WorkflowState) => {
      state.workflows.push(workflow);
    }),
    
    createWorkflow: (id, name) => {
      const workflow = createEmptyWorkflow(id, name);
      set((state: WorkflowState) => {
        state.workflows.push(workflow);
      });
      return workflow;
    },
    
    updateWorkflow: (id, updates) => set((state: WorkflowState) => {
      const workflow = state.workflows.find(w => w.id === id);
      if (workflow) {
        Object.assign(workflow, updates, { updatedAt: new Date().toISOString() });
      }
    }),
    
    deleteWorkflow: (id) => set((state: WorkflowState) => {
      state.workflows = state.workflows.filter(w => w.id !== id);
      if (state.selectedWorkflowId === id) {
        state.selectedWorkflowId = null;
      }
    }),
    
    selectWorkflow: (id) => set({ selectedWorkflowId: id }),
    
    getSelectedWorkflow: () => {
      const state = get();
      return state.workflows.find(w => w.id === state.selectedWorkflowId) || null;
    },
    
    getWorkflowById: (id) => {
      return get().workflows.find(w => w.id === id);
    },
    
    // Step operations
    addStep: (workflowId, step) => set((state: WorkflowState) => {
      const workflow = state.workflows.find(w => w.id === workflowId);
      if (workflow) {
        // Ensure step has required fields
        const fullStep: WorkflowStep = {
          ...step,
          common: step.common || getDefaultStepCommon(),
          connections: step.connections || getDefaultConnections(),
        };
        workflow.steps.push(fullStep);
        workflow.updatedAt = new Date().toISOString();
      }
    }),
    
    updateStep: (workflowId, stepId, updates) => set((state: WorkflowState) => {
      const workflow = state.workflows.find(w => w.id === workflowId);
      if (workflow) {
        const stepIndex = workflow.steps.findIndex(s => s.id === stepId);
        if (stepIndex !== -1) {
          workflow.steps[stepIndex] = { ...workflow.steps[stepIndex], ...updates };
          workflow.updatedAt = new Date().toISOString();
        }
      }
    }),
    
    deleteStep: (workflowId, stepId) => set((state: WorkflowState) => {
      const workflow = state.workflows.find(w => w.id === workflowId);
      if (workflow) {
        // Don't delete the start node
        if (stepId === 'start') return;
        
        workflow.steps = workflow.steps.filter(s => s.id !== stepId);
        
        // Clean up connections pointing to deleted step
        workflow.steps.forEach(step => {
          if (step.connections) {
            if (step.connections.successStepId === stepId) {
              step.connections.successStepId = undefined;
            }
            if (step.connections.errorStepId === stepId) {
              step.connections.errorStepId = undefined;
            }
            if (step.connections.trueStepId === stepId) {
              step.connections.trueStepId = undefined;
            }
            if (step.connections.falseStepId === stepId) {
              step.connections.falseStepId = undefined;
            }
          }
        });
        
        workflow.updatedAt = new Date().toISOString();
      }
    }),
    
    getStep: (workflowId, stepId) => {
      const workflow = get().workflows.find(w => w.id === workflowId);
      return workflow?.steps.find(s => s.id === stepId);
    },
    
    updateStepConnections: (workflowId, stepId, connections) => set((state: WorkflowState) => {
      const workflow = state.workflows.find(w => w.id === workflowId);
      if (workflow) {
        const step = workflow.steps.find(s => s.id === stepId);
        if (step) {
          step.connections = { ...step.connections, ...connections };
          workflow.updatedAt = new Date().toISOString();
        }
      }
    }),
    
    // Graph operations
    setNodes: (nodes) => set({ nodes }),
    
    setEdges: (edges) => set({ edges }),
    
    updateGraph: (nodes, edges) => set({ nodes, edges }),
    
    // Execution state operations
    setIsRunning: (running) => set({ isRunning: running }),
    
    setRunningWorkflowIds: (ids) => {
      if (typeof ids === 'function') {
        set((state: WorkflowState) => {
          state.runningWorkflowIds = ids(state.runningWorkflowIds);
        });
      } else {
        set({ runningWorkflowIds: ids });
      }
    },
    
    addRunningWorkflowId: (id) => set((state: WorkflowState) => {
      if (!state.runningWorkflowIds.includes(id)) {
        state.runningWorkflowIds.push(id);
      }
    }),
    
    removeRunningWorkflowId: (id) => set((state: WorkflowState) => {
      state.runningWorkflowIds = state.runningWorkflowIds.filter(wid => wid !== id);
    }),
    
    setIsPaused: (paused) => set({ isPaused: paused }),
    
    setCurrentStepId: (stepId) => set({ currentStepId: stepId }),
    
    setWaitingStep: (stepId, phase) => set({ 
      waitingStepId: stepId, 
      waitingPhase: phase 
    }),
    
    // Legacy compatibility methods
    setWaitingStepId: (stepId) => set({ waitingStepId: stepId }),
    setWaitingPhase: (phase) => set({ waitingPhase: phase }),
    
    setExecutionLogs: (logs) => {
      if (typeof logs === 'function') {
        set((state: WorkflowState) => {
          state.executionLogs = logs(state.executionLogs);
        });
      } else {
        set({ executionLogs: logs });
      }
    },
    
    addExecutionLog: (log) => set((state: WorkflowState) => {
      state.executionLogs.push(log);
    }),
    
    clearExecutionLogs: () => set({ executionLogs: [] }),
    
    setWorkflowStepMap: (map) => {
      if (typeof map === 'function') {
        const currentMap = get().workflowStepMap;
        const newMap = map(currentMap);
        set({ workflowStepMap: { ...newMap } });
      } else {
        set({ workflowStepMap: { ...map } });
      }
    },
    
    updateWorkflowStepStatus: (workflowId, status) => set((state: WorkflowState) => {
      state.workflowStepMap[workflowId] = status;
    }),
    
    clearWorkflowStepStatus: (workflowId) => set((state: WorkflowState) => {
      delete state.workflowStepMap[workflowId];
    }),
    
    setRuntimeContext: (context) => set({ runtimeContext: context }),
    
    clearRuntimeContext: () => set({ runtimeContext: null }),
    
    // UI state operations
    setWorkflowModalVisible: (visible) => set({ workflowModalVisible: visible }),
    
    setDrawerVisible: (visible) => set({ drawerVisible: visible }),
    
    setElementPickerVisible: (visible) => set({ elementPickerVisible: visible }),
    
    setVariablesModalVisible: (visible) => set({ variablesModalVisible: visible }),
    
    setEditingNode: (nodeId) => set({ editingNodeId: nodeId }),
    
    // Legacy compatibility methods
    setEditingNodeId: (nodeId) => set({ editingNodeId: nodeId }),
    
    setEditingWorkflow: (id, name) => set({ 
      editingWorkflowId: id, 
      editingWorkflowName: name 
    }),
    
    // Legacy compatibility methods
    setEditingWorkflowId: (id) => set({ editingWorkflowId: id }),
    setEditingWorkflowName: (name) => set((state: WorkflowState) => { 
      state.editingWorkflowName = name;
    }),
    setSelectedWorkflow: (workflow) => {
      if (workflow === null) {
        set({ selectedWorkflowId: null });
      } else {
        set({ selectedWorkflowId: workflow.id });
      }
    },
    
    // Temporary data operations
    setTempVariables: (vars) => {
      if (typeof vars === 'function') {
        set((state: WorkflowState) => {
          state.tempVariables = vars(state.tempVariables);
        });
      } else {
        set({ tempVariables: vars });
      }
    },
    
    setPackages: (packages) => set({ packages }),
    
    setAppsLoading: (loading) => set({ appsLoading: loading }),
    
    setNeedsAutoSave: (needs) => set({ needsAutoSave: needs }),
    
    // View mode
    setViewMode: (mode) => set({ viewMode: mode }),
    
    // Export utilities
    exportWorkflow: (id) => {
      return get().workflows.find(w => w.id === id);
    },
  }))
);

// Helper functions for creating steps
export function createStep(
  id: string,
  type: StepType,
  name?: string,
  overrides?: Partial<WorkflowStep>
): WorkflowStep {
  return {
    id,
    type,
    name: name || type,
    common: getDefaultStepCommon(),
    connections: getDefaultConnections(),
    layout: { posX: 0, posY: 0 },
    ...overrides,
  };
}

// Re-export utility functions
export { createEmptyWorkflow, getDefaultStepCommon, getDefaultConnections };
