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

export interface ReadToVariableParams {
  selector: ElementSelector;
  variableName: string;
  attribute?: 'text' | 'contentDesc' | 'resourceId' | 'className' | 'bounds';
  regex?: string;
  defaultValue?: string;
}

export interface SessionParams {
  // Session name (for start_session)
  sessionName?: string;

  // Logcat config
  logcatEnabled?: boolean;
  logcatPackageName?: string;
  logcatPreFilter?: string;
  logcatExcludeFilter?: string;

  // Recording config
  recordingEnabled?: boolean;
  recordingQuality?: 'low' | 'medium' | 'high';

  // Proxy config
  proxyEnabled?: boolean;
  proxyPort?: number;
  proxyMitmEnabled?: boolean;

  // Monitor config
  monitorEnabled?: boolean;

  // For end_session: session end status
  status?: 'completed' | 'error' | 'cancelled';
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
  | 'read_to_variable'
  | 'adb'
  | 'run_workflow'
  | 'key_back'
  | 'key_home'
  | 'key_recent'
  | 'key_power'
  | 'key_volume_up'
  | 'key_volume_down'
  | 'screen_on'
  | 'screen_off'
  | 'start_session'
  | 'end_session';

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
  readToVariable?: ReadToVariableParams;
  session?: SessionParams;

  // UI Layout
  layout?: StepLayout;
}

// ============== Workflow ==============

export interface Workflow {
  id: string;
  name: string;
  description?: string;
  version?: number;  // Schema version (2 = current V2 format)
  steps: WorkflowStep[];
  variables?: Record<string, string>;
  createdAt?: string;
  updatedAt?: string;
}

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


