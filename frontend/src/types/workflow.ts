/**
 * Workflow Types
 * Matching Go backend types for type safety
 */

// ============== Step Connections ==============

export interface StepConnections {
  successStepId?: string;  // Next step on success
  errorStepId?: string;    // Next step on error
  trueStepId?: string;     // Branch condition true
  falseStepId?: string;    // Branch condition false
}

// ============== Step Common Config ==============

export interface StepCommon {
  timeout?: number;      // Timeout in milliseconds
  onError?: 'stop' | 'continue';
  loop?: number;         // Number of times to repeat
  postDelay?: number;    // Delay after execution (ms)
  preWait?: number;      // Wait before execution (ms)
}

// ============== Step Layout ==============

export interface HandleInfo {
  sourceHandle?: string;
  targetHandle?: string;
}

export interface StepLayout {
  posX?: number;
  posY?: number;
  handles?: Record<string, HandleInfo>;
}

// ============== Element Selector ==============

export interface ElementSelector {
  type: 'text' | 'id' | 'xpath' | 'contentDesc' | 'className' | 'advanced';
  value: string;
  index?: number;
}

// ============== Type-Specific Parameters ==============

export interface TapParams {
  x: number;
  y: number;
}

export interface SwipeParams {
  // Coordinate mode
  x?: number;
  y?: number;
  x2?: number;
  y2?: number;
  // Direction mode
  direction?: 'up' | 'down' | 'left' | 'right';
  distance?: number;
  // Common
  duration?: number;
}

export interface ElementParams {
  selector: ElementSelector;
  action: 'click' | 'long_click' | 'input' | 'swipe' | 'wait' | 'wait_gone' | 'assert';
  // Action-specific parameters
  inputText?: string;
  swipeDir?: 'up' | 'down' | 'left' | 'right';
  swipeDistance?: number;
  swipeDuration?: number;
}

export interface AppParams {
  packageName: string;
  action: 'launch' | 'stop' | 'clear' | 'settings';
}

export interface BranchParams {
  condition: 'exists' | 'not_exists' | 'text_equals' | 'text_contains' | 'variable_equals';
  selector?: ElementSelector;
  expectedValue?: string;
  variableName?: string;
}

export interface WaitParams {
  durationMs: number;
}

export interface ScriptParams {
  scriptName: string;
}

export interface VariableParams {
  name: string;
  value: string;
}

export interface ADBParams {
  command: string;
}

export interface SubWorkflowParams {
  workflowId: string;
}

// ============== Step Types ==============

export type StepType =
  | 'start'
  | 'tap'
  | 'swipe'
  | 'click_element'
  | 'long_click_element'
  | 'input_text'
  | 'swipe_element'
  | 'wait_element'
  | 'wait_gone'
  | 'assert_element'
  | 'launch_app'
  | 'stop_app'
  | 'clear_app'
  | 'open_settings'
  | 'branch'
  | 'wait'
  | 'script'
  | 'set_variable'
  | 'adb'
  | 'run_workflow'
  | 'key_back'
  | 'key_home'
  | 'key_recent'
  | 'key_power'
  | 'key_volume_up'
  | 'key_volume_down'
  | 'screen_on'
  | 'screen_off';

// ============== WorkflowStep ==============

export interface WorkflowStep {
  id: string;
  type: StepType;
  name?: string;

  // Common configuration
  common?: StepCommon;

  // Flow connections
  connections?: StepConnections;

  // Type-specific parameters (only one should be set based on type)
  tap?: TapParams;
  swipe?: SwipeParams;
  element?: ElementParams;
  app?: AppParams;
  branch?: BranchParams;
  wait?: WaitParams;
  script?: ScriptParams;
  variable?: VariableParams;
  adb?: ADBParams;
  workflow?: SubWorkflowParams;

  // UI Layout
  layout?: StepLayout;
}

// ============== Workflow ==============

export interface Workflow {
  id: string;
  name: string;
  description?: string;
  steps: WorkflowStep[];
  variables?: Record<string, string>;
  createdAt?: string;
  updatedAt?: string;
}

// ============== Type Aliases for Backward Compatibility ==============
// These will be removed in future versions

/** @deprecated Use WorkflowStep instead */
export type WorkflowStepV2 = WorkflowStep;

/** @deprecated Use Workflow instead */
export type WorkflowV2 = Workflow;

/** @deprecated Use ElementParams instead */
export type ElementParamsV2 = ElementParams;

// ============== Helper Functions ==============

export function getDefaultStepCommon(): StepCommon {
  return {
    timeout: 5000,
    onError: 'stop',
    loop: 1,
    postDelay: 0,
    preWait: 0,
  };
}

export function getDefaultConnections(): StepConnections {
  return {
    successStepId: '',
    errorStepId: '',
  };
}

export function createEmptyWorkflow(id: string, name: string): Workflow {
  return {
    id,
    name,
    steps: [
      {
        id: 'start',
        type: 'start',
        name: 'Start',
        connections: { successStepId: '' },
        layout: { posX: 100, posY: 100 },
      },
    ],
    variables: {},
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  };
}

/** @deprecated Use createEmptyWorkflow instead */
export const createEmptyWorkflowV2 = createEmptyWorkflow;

// ============== Step Helpers ==============

export function shouldStopOnError(step: WorkflowStep): boolean {
  // If error path is connected, don't stop
  if (step.connections?.errorStepId) {
    return false;
  }
  // Otherwise, check OnError setting
  return step.common?.onError !== 'continue';
}

export function getNextStepId(step: WorkflowStep, success: boolean, isBranchResult: boolean): string | undefined {
  if (step.type === 'branch' && isBranchResult) {
    return success ? step.connections?.trueStepId : step.connections?.falseStepId;
  }

  if (success) {
    return step.connections?.successStepId;
  }

  return step.connections?.errorStepId;
}

export function getFallbackStepId(step: WorkflowStep): string | undefined {
  if (step.connections?.errorStepId) {
    return step.connections.errorStepId;
  }
  if (step.common?.onError === 'continue') {
    return step.connections?.successStepId;
  }
  return undefined;
}

// ============== Node Type Categories ==============

export const STEP_CATEGORIES = {
  COORDINATE_ACTIONS: ['tap', 'swipe'] as StepType[],
  ELEMENT_ACTIONS: ['click_element', 'long_click_element', 'input_text', 'swipe_element'] as StepType[],
  WAIT_CONDITIONS: ['wait_element', 'wait_gone', 'wait'] as StepType[],
  FLOW_CONTROL: ['start', 'branch', 'set_variable'] as StepType[],
  SCRIPT_ACTIONS: ['script', 'adb'] as StepType[],
  SYSTEM_ACTIONS: ['key_back', 'key_home', 'key_recent', 'key_power', 'key_volume_up', 'key_volume_down', 'screen_on', 'screen_off'] as StepType[],
  APP_ACTIONS: ['launch_app', 'stop_app', 'clear_app', 'open_settings'] as StepType[],
  ASSERTIONS: ['assert_element'] as StepType[],
  NESTED: ['run_workflow'] as StepType[],
};

export function getStepCategory(type: StepType): string {
  for (const [category, types] of Object.entries(STEP_CATEGORIES)) {
    if ((types as StepType[]).includes(type)) {
      return category;
    }
  }
  return 'UNKNOWN';
}

// ============== Handle Types for React Flow ==============

export type HandleType = 'success' | 'error' | 'true' | 'false' | 'source' | 'target';

export function getStepHandles(type: StepType): { outputs: HandleType[]; inputs: HandleType[] } {
  if (type === 'start') {
    return { outputs: ['success'], inputs: [] };
  }

  if (type === 'branch') {
    return { outputs: ['true', 'false', 'error'], inputs: ['target'] };
  }

  // All other nodes have success and error outputs
  return { outputs: ['success', 'error'], inputs: ['target'] };
}
