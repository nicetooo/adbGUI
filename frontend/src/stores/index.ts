// Barrel file for all stores
export * from './types';
export { useDeviceStore } from './deviceStore';
export { useMirrorStore, openRecordPath } from './mirrorStore';
export { useLogcatStore } from './logcatStore';
export type { FilterPreset } from './logcatStore';
export { useUIStore } from './uiStore';
export { useAutomationStore } from './automationStore';
export type { TouchEvent, TouchScript, ScriptTask, TaskStep } from './automationStore';
export { useElementStore } from './elementStore';
export type { ElementSelector, UINode, SelectorSuggestion, ElementInfo } from './elementStore';
export { useSessionStore, formatEventTime, formatDuration, getEventIcon, getEventColor, categoryColors, levelStyles } from './sessionStore';
export type { Session, SessionEvent, SessionFilter } from './sessionStore';

// New stores
export { useProxyStore } from './proxyStore';
export type { RequestLog, NetworkStats } from './proxyStore';
export { useWorkflowStore } from './workflowStore';
export type { WorkflowStep, Workflow } from './workflowStore';
export { useUIInspectorStore } from './uiInspectorStore';
export { useAppsStore } from './appsStore';
export { useFilesStore } from './filesStore';
export { useShellStore } from './shellStore';
export { useRecordingStore } from './recordingStore';
export { useTimelineStore } from './timelineStore';
