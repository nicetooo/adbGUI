import { describe, it, expect, beforeEach } from 'vitest';
import { useWorkflowStore, createStep } from './workflowStore';
import type { Workflow, WorkflowStep } from '../types/workflow';

// Reset store state before each test
beforeEach(() => {
  useWorkflowStore.setState({
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
  });
});

// ============== Workflow CRUD Tests ==============

describe('Workflow CRUD Operations', () => {
  it('should create a new workflow with proper structure', () => {
    const store = useWorkflowStore.getState();
    const workflow = store.createWorkflow('test-1', 'Test Workflow');
    
    expect(workflow.id).toBe('test-1');
    expect(workflow.name).toBe('Test Workflow');
    expect(workflow.steps).toHaveLength(1);
    expect(workflow.steps[0].type).toBe('start');
    expect(workflow.steps[0].connections).toBeDefined();
    
    const stored = useWorkflowStore.getState().workflows;
    expect(stored).toHaveLength(1);
  });

  it('should add a workflow directly', () => {
    const store = useWorkflowStore.getState();
    const workflow: Workflow = {
      id: 'direct-add',
      name: 'Direct Add',
      steps: [],
    };
    
    store.addWorkflow(workflow);
    
    expect(useWorkflowStore.getState().workflows).toHaveLength(1);
    expect(useWorkflowStore.getState().workflows[0].id).toBe('direct-add');
  });

  it('should update workflow properties', () => {
    const store = useWorkflowStore.getState();
    store.createWorkflow('update-test', 'Original Name');
    
    store.updateWorkflow('update-test', { name: 'Updated Name', description: 'New description' });
    
    const updated = useWorkflowStore.getState().workflows.find(w => w.id === 'update-test');
    expect(updated?.name).toBe('Updated Name');
    expect(updated?.description).toBe('New description');
    expect(updated?.updatedAt).toBeDefined();
  });

  it('should delete workflow and clear selection', () => {
    const store = useWorkflowStore.getState();
    store.createWorkflow('to-delete', 'To Delete');
    store.selectWorkflow('to-delete');
    
    expect(useWorkflowStore.getState().selectedWorkflowId).toBe('to-delete');
    
    store.deleteWorkflow('to-delete');
    
    expect(useWorkflowStore.getState().workflows).toHaveLength(0);
    expect(useWorkflowStore.getState().selectedWorkflowId).toBeNull();
  });

  it('should get selected workflow', () => {
    const store = useWorkflowStore.getState();
    store.createWorkflow('get-test', 'Get Test');
    store.selectWorkflow('get-test');
    
    const selected = store.getSelectedWorkflow();
    expect(selected?.id).toBe('get-test');
  });

  it('should get workflow by ID', () => {
    const store = useWorkflowStore.getState();
    store.createWorkflow('by-id', 'By ID');
    
    const workflow = store.getWorkflowById('by-id');
    expect(workflow?.name).toBe('By ID');
    
    const notFound = store.getWorkflowById('nonexistent');
    expect(notFound).toBeUndefined();
  });
});

// ============== Step Operations Tests ==============

describe('Step Operations', () => {
  beforeEach(() => {
    const store = useWorkflowStore.getState();
    store.createWorkflow('step-test', 'Step Test');
  });

  it('should add a step to workflow', () => {
    const store = useWorkflowStore.getState();
    const newStep: WorkflowStep = {
      id: 'tap-1',
      type: 'tap',
      tap: { x: 100, y: 200 },
    };
    
    store.addStep('step-test', newStep);
    
    const workflow = store.getWorkflowById('step-test');
    expect(workflow?.steps).toHaveLength(2); // start + tap-1
    expect(workflow?.steps[1].id).toBe('tap-1');
    expect(workflow?.steps[1].common).toBeDefined(); // Should have defaults
    expect(workflow?.steps[1].connections).toBeDefined();
  });

  it('should update a step', () => {
    const store = useWorkflowStore.getState();
    store.addStep('step-test', { id: 'tap-1', type: 'tap', tap: { x: 100, y: 200 } });
    
    store.updateStep('step-test', 'tap-1', { 
      tap: { x: 300, y: 400 },
      name: 'Updated Tap',
    });
    
    const step = store.getStep('step-test', 'tap-1');
    expect(step?.tap?.x).toBe(300);
    expect(step?.tap?.y).toBe(400);
    expect(step?.name).toBe('Updated Tap');
  });

  it('should delete a step and clean up connections', () => {
    const store = useWorkflowStore.getState();
    store.addStep('step-test', { 
      id: 'tap-1', 
      type: 'tap', 
      tap: { x: 100, y: 200 },
    });
    store.addStep('step-test', { 
      id: 'tap-2', 
      type: 'tap', 
      tap: { x: 300, y: 400 },
    });
    
    // Connect start -> tap-1 -> tap-2
    store.updateStepConnections('step-test', 'start', { successStepId: 'tap-1' });
    store.updateStepConnections('step-test', 'tap-1', { successStepId: 'tap-2' });
    
    // Delete tap-1
    store.deleteStep('step-test', 'tap-1');
    
    const workflow = store.getWorkflowById('step-test');
    expect(workflow?.steps).toHaveLength(2); // start + tap-2
    
    // Connection from start should be cleared
    const startStep = workflow?.steps.find(s => s.id === 'start');
    expect(startStep?.connections?.successStepId).toBeUndefined();
  });

  it('should not delete start node', () => {
    const store = useWorkflowStore.getState();
    
    store.deleteStep('step-test', 'start');
    
    const workflow = store.getWorkflowById('step-test');
    expect(workflow?.steps.find(s => s.id === 'start')).toBeDefined();
  });

  it('should update step connections', () => {
    const store = useWorkflowStore.getState();
    store.addStep('step-test', { id: 'tap-1', type: 'tap', tap: { x: 100, y: 200 } });
    store.addStep('step-test', { id: 'tap-2', type: 'tap', tap: { x: 300, y: 400 } });
    
    store.updateStepConnections('step-test', 'tap-1', { 
      successStepId: 'tap-2',
      errorStepId: 'start',
    });
    
    const step = store.getStep('step-test', 'tap-1');
    expect(step?.connections?.successStepId).toBe('tap-2');
    expect(step?.connections?.errorStepId).toBe('start');
  });
});

// ============== createStep Helper Tests ==============

describe('createStep Helper', () => {
  it('should create a basic step with defaults', () => {
    const step = createStep('test-step', 'tap', 'Test Tap');
    
    expect(step.id).toBe('test-step');
    expect(step.type).toBe('tap');
    expect(step.name).toBe('Test Tap');
    expect(step.common).toBeDefined();
    expect(step.common?.timeout).toBe(5000);
    expect(step.common?.onError).toBe('stop');
    expect(step.connections).toBeDefined();
    expect(step.layout).toBeDefined();
  });

  it('should allow overriding defaults', () => {
    const step = createStep('test-step', 'tap', 'Test Tap', {
      tap: { x: 100, y: 200 },
      common: { timeout: 10000, onError: 'continue' },
    });
    
    expect(step.tap?.x).toBe(100);
    expect(step.common?.timeout).toBe(10000);
    expect(step.common?.onError).toBe('continue');
  });

  it('should use type as name if not provided', () => {
    const step = createStep('test-step', 'key_back');
    expect(step.name).toBe('key_back');
  });
});

// ============== Execution State Tests ==============

describe('Execution State', () => {
  it('should manage running state', () => {
    const store = useWorkflowStore.getState();
    
    store.setIsRunning(true);
    expect(useWorkflowStore.getState().isRunning).toBe(true);
    
    store.setIsRunning(false);
    expect(useWorkflowStore.getState().isRunning).toBe(false);
  });

  it('should manage running workflow IDs', () => {
    const store = useWorkflowStore.getState();
    
    store.addRunningWorkflowId('wf-1');
    store.addRunningWorkflowId('wf-2');
    
    expect(useWorkflowStore.getState().runningWorkflowIds).toEqual(['wf-1', 'wf-2']);
    
    // Should not add duplicates
    store.addRunningWorkflowId('wf-1');
    expect(useWorkflowStore.getState().runningWorkflowIds).toEqual(['wf-1', 'wf-2']);
    
    store.removeRunningWorkflowId('wf-1');
    expect(useWorkflowStore.getState().runningWorkflowIds).toEqual(['wf-2']);
  });

  it('should manage waiting step state', () => {
    const store = useWorkflowStore.getState();
    
    store.setWaitingStep('step-1', 'pre');
    
    const state = useWorkflowStore.getState();
    expect(state.waitingStepId).toBe('step-1');
    expect(state.waitingPhase).toBe('pre');
    
    store.setWaitingStep(null, null);
    expect(useWorkflowStore.getState().waitingStepId).toBeNull();
  });

  it('should manage execution logs', () => {
    const store = useWorkflowStore.getState();
    
    store.addExecutionLog('Step 1 executed');
    store.addExecutionLog('Step 2 executed');
    
    expect(useWorkflowStore.getState().executionLogs).toEqual([
      'Step 1 executed',
      'Step 2 executed',
    ]);
    
    store.clearExecutionLogs();
    expect(useWorkflowStore.getState().executionLogs).toEqual([]);
  });
});

// ============== UI State Tests ==============

describe('UI State', () => {
  it('should manage modal visibility', () => {
    const store = useWorkflowStore.getState();
    
    store.setWorkflowModalVisible(true);
    expect(useWorkflowStore.getState().workflowModalVisible).toBe(true);
    
    store.setDrawerVisible(true);
    expect(useWorkflowStore.getState().drawerVisible).toBe(true);
    
    store.setVariablesModalVisible(true);
    expect(useWorkflowStore.getState().variablesModalVisible).toBe(true);
  });

  it('should manage editing state', () => {
    const store = useWorkflowStore.getState();
    
    store.setEditingNode('node-1');
    expect(useWorkflowStore.getState().editingNodeId).toBe('node-1');
    
    store.setEditingWorkflow('wf-1', 'Workflow Name');
    const state = useWorkflowStore.getState();
    expect(state.editingWorkflowId).toBe('wf-1');
    expect(state.editingWorkflowName).toBe('Workflow Name');
  });
});

// ============== Export Tests ==============

describe('Export Utilities', () => {
  beforeEach(() => {
    const store = useWorkflowStore.getState();
    const workflow: Workflow = {
      id: 'export-test',
      name: 'Export Test',
      steps: [
        {
          id: 'start',
          type: 'start',
          connections: { successStepId: 'tap-1' },
          layout: { posX: 100, posY: 100 },
        },
        {
          id: 'tap-1',
          type: 'tap',
          tap: { x: 500, y: 300 },
          connections: { successStepId: '' },
          layout: { posX: 300, posY: 100 },
        },
      ],
    };
    store.addWorkflow(workflow);
  });

  it('should export workflow', () => {
    const store = useWorkflowStore.getState();
    const exported = store.exportWorkflow('export-test');
    
    expect(exported?.steps[1].tap?.x).toBe(500);
  });

  it('should return undefined for non-existent workflow', () => {
    const store = useWorkflowStore.getState();
    
    expect(store.exportWorkflow('nonexistent')).toBeUndefined();
  });
});

// ============== Graph State Tests ==============

describe('Graph State', () => {
  it('should manage nodes and edges', () => {
    const store = useWorkflowStore.getState();
    
    const nodes = [
      { id: 'node-1', position: { x: 0, y: 0 }, data: {} },
      { id: 'node-2', position: { x: 100, y: 100 }, data: {} },
    ];
    const edges = [
      { id: 'edge-1', source: 'node-1', target: 'node-2' },
    ];
    
    store.setNodes(nodes);
    store.setEdges(edges);
    
    expect(useWorkflowStore.getState().nodes).toEqual(nodes);
    expect(useWorkflowStore.getState().edges).toEqual(edges);
  });

  it('should update graph atomically', () => {
    const store = useWorkflowStore.getState();
    
    const nodes = [{ id: 'node-1', position: { x: 0, y: 0 }, data: {} }];
    const edges = [{ id: 'edge-1', source: 'node-1', target: 'node-2' }];
    
    store.updateGraph(nodes, edges);
    
    const state = useWorkflowStore.getState();
    expect(state.nodes).toEqual(nodes);
    expect(state.edges).toEqual(edges);
  });
});

// ============== Functional Updates Tests ==============

describe('Functional Updates', () => {
  it('should support functional updates for workflows', () => {
    const store = useWorkflowStore.getState();
    store.createWorkflow('func-1', 'Func 1');
    store.createWorkflow('func-2', 'Func 2');
    
    store.setWorkflows(prev => prev.filter(w => w.id !== 'func-1'));
    
    expect(useWorkflowStore.getState().workflows).toHaveLength(1);
    expect(useWorkflowStore.getState().workflows[0].id).toBe('func-2');
  });

  it('should support functional updates for execution logs', () => {
    const store = useWorkflowStore.getState();
    store.setExecutionLogs(['Log 1']);
    
    store.setExecutionLogs(prev => [...prev, 'Log 2']);
    
    expect(useWorkflowStore.getState().executionLogs).toEqual(['Log 1', 'Log 2']);
  });

  it('should support functional updates for running workflow IDs', () => {
    const store = useWorkflowStore.getState();
    store.setRunningWorkflowIds(['wf-1']);
    
    store.setRunningWorkflowIds(prev => [...prev, 'wf-2']);
    
    expect(useWorkflowStore.getState().runningWorkflowIds).toEqual(['wf-1', 'wf-2']);
  });
});
