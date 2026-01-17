import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';
import type { Node, Edge } from '@xyflow/react';

export interface WorkflowStep {
  id: string;
  type: string;
  label?: string;
  name?: string;
  selector?: any;
  value?: string;
  loop?: number;
  timeout?: number;
  postDelay?: number;
  preWait?: number;
  swipeDistance?: number;
  swipeDuration?: number;
  conditionType?: string;
  onError?: 'stop' | 'continue';
  nextStepId?: string;
  trueStepId?: string;
  falseStepId?: string;
  posX?: number;
  posY?: number;
  // Coordinate fields for tap/swipe
  x?: number;
  y?: number;
  x2?: number;
  y2?: number;
  nextSource?: string;
  trueSource?: string;
  falseSource?: string;
  nextTarget?: string;
  trueTarget?: string;
  falseTarget?: string;
  children?: WorkflowStep[];
  scriptId?: string;
  workflowId?: string;
  [key: string]: any;
}

export interface Workflow {
  id: string;
  name: string;
  description?: string;
  steps: WorkflowStep[];
  variables?: Record<string, string>;
  createdAt?: string;
  updatedAt?: string;
}

interface StepStatus {
  stepId: string;
  isWaiting: boolean;
  waitPhase?: 'pre' | 'post';
}

interface WorkflowState {
  // 核心数据
  workflows: Workflow[];
  selectedWorkflowId: string | null;
  nodes: Node[];
  edges: Edge[];
  
  // 执行状态
  isRunning: boolean;
  runningWorkflowIds: string[];
  isPaused: boolean;
  currentStepId: string | null;
  waitingStepId: string | null;
  waitingPhase: 'pre' | 'post' | null;
  executionLogs: string[];
  workflowStepMap: Record<string, StepStatus>;
  
  // UI 状态
  workflowModalVisible: boolean;
  drawerVisible: boolean;
  elementPickerVisible: boolean;
  variablesModalVisible: boolean;
  editingNodeId: string | null;
  editingWorkflowId: string | null;
  editingWorkflowName: string;
  
  // 临时数据
  tempVariables: { key: string; value: string }[];
  packages: any[];
  appsLoading: boolean;
  needsAutoSave: boolean;
  
  // 操作方法
  setWorkflows: (workflows: Workflow[] | ((prev: Workflow[]) => Workflow[])) => void;
  addWorkflow: (workflow: Workflow) => void;
  updateWorkflow: (id: string, updates: Partial<Workflow>) => void;
  deleteWorkflow: (id: string) => void;
  selectWorkflow: (id: string | null) => void;
  getSelectedWorkflow: () => Workflow | null;
  
  setNodes: (nodes: Node[]) => void;
  setEdges: (edges: Edge[]) => void;
  updateGraph: (nodes: Node[], edges: Edge[]) => void;
  
  setIsRunning: (running: boolean) => void;
  setRunningWorkflowIds: (ids: string[] | ((prev: string[]) => string[])) => void;
  addRunningWorkflowId: (id: string) => void;
  removeRunningWorkflowId: (id: string) => void;
  setIsPaused: (paused: boolean) => void;
  setCurrentStepId: (stepId: string | null) => void;
  setWaitingStep: (stepId: string | null, phase: 'pre' | 'post' | null) => void;
  // 兼容方法
  setWaitingStepId: (stepId: string | null) => void;
  setWaitingPhase: (phase: 'pre' | 'post' | null) => void;
  setExecutionLogs: (logs: string[] | ((prev: string[]) => string[])) => void;
  addExecutionLog: (log: string) => void;
  clearExecutionLogs: () => void;
  setWorkflowStepMap: (map: Record<string, StepStatus> | ((prev: Record<string, StepStatus>) => Record<string, StepStatus>)) => void;
  updateWorkflowStepStatus: (workflowId: string, status: StepStatus) => void;
  clearWorkflowStepStatus: (workflowId: string) => void;
  
  setWorkflowModalVisible: (visible: boolean) => void;
  setDrawerVisible: (visible: boolean) => void;
  setElementPickerVisible: (visible: boolean) => void;
  setVariablesModalVisible: (visible: boolean) => void;
  setEditingNode: (nodeId: string | null) => void;
  // 兼容方法
  setEditingNodeId: (nodeId: string | null) => void;
  setEditingWorkflow: (id: string | null, name: string) => void;
  // 兼容方法
  setEditingWorkflowId: (id: string | null) => void;
  setEditingWorkflowName: (name: string) => void;
  setSelectedWorkflow: (workflow: Workflow | null) => void;
  
  setTempVariables: (vars: { key: string; value: string }[] | ((prev: { key: string; value: string }[]) => { key: string; value: string }[])) => void;
  setPackages: (packages: any[]) => void;
  setAppsLoading: (loading: boolean) => void;
  setNeedsAutoSave: (needs: boolean) => void;
}

export const useWorkflowStore = create<WorkflowState>()(
  immer((set, get) => ({
    // 初始状态
    workflows: [],
    selectedWorkflowId: null,
    nodes: [],
    edges: [],
    
    isRunning: false,
    runningWorkflowIds: [],
    isPaused: false,
    currentStepId: null,
    waitingStepId: null,
    waitingPhase: null,
    executionLogs: [],
    workflowStepMap: {},
    
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
    
    // 操作方法
    setWorkflows: (workflows) => {
      if (typeof workflows === 'function') {
        set((state: WorkflowState) => {
          state.workflows = workflows(state.workflows);
        });
      } else {
        set({ workflows });
      }
    },
    
    addWorkflow: (workflow) => set((state: WorkflowState) => {
      state.workflows.push(workflow);
    }),
    
    updateWorkflow: (id, updates) => set((state: WorkflowState) => {
      const workflow = state.workflows.find(w => w.id === id);
      if (workflow) {
        Object.assign(workflow, updates);
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
    
    setNodes: (nodes) => set({ nodes }),
    
    setEdges: (edges) => set({ edges }),
    
    updateGraph: (nodes, edges) => set({ nodes, edges }),
    
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
    
    // 兼容方法
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
        set((state: WorkflowState) => {
          state.workflowStepMap = map(state.workflowStepMap);
        });
      } else {
        set({ workflowStepMap: map });
      }
    },
    
    updateWorkflowStepStatus: (workflowId, status) => set((state: WorkflowState) => {
      state.workflowStepMap[workflowId] = status;
    }),
    
    clearWorkflowStepStatus: (workflowId) => set((state: WorkflowState) => {
      delete state.workflowStepMap[workflowId];
    }),
    
    setWorkflowModalVisible: (visible) => set({ workflowModalVisible: visible }),
    
    setDrawerVisible: (visible) => set({ drawerVisible: visible }),
    
    setElementPickerVisible: (visible) => set({ elementPickerVisible: visible }),
    
    setVariablesModalVisible: (visible) => set({ variablesModalVisible: visible }),
    
    setEditingNode: (nodeId) => set({ editingNodeId: nodeId }),
    
    // 兼容方法
    setEditingNodeId: (nodeId) => set({ editingNodeId: nodeId }),
    
    setEditingWorkflow: (id, name) => set({ 
      editingWorkflowId: id, 
      editingWorkflowName: name 
    }),
    
    // 兼容方法
    setEditingWorkflowId: (id) => set({ editingWorkflowId:id }),
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
  }))
);
