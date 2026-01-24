// Barrel file for all stores
export * from './types';
export { useDeviceStore } from './deviceStore';
export { useMirrorStore, openRecordPath } from './mirrorStore';
export { useLogcatStore } from './logcatStore';
export type { FilterPreset, ParsedLog } from './logcatStore';
export { useUIStore } from './uiStore';
export { useAutomationStore } from './automationStore';
export type { TouchEvent, TouchScript, ScriptTask, TaskStep } from './automationStore';
export { useElementStore } from './elementStore';
export type { ElementSelector, UINode, SelectorSuggestion, ElementInfo } from './elementStore';
export { useSessionStore, formatEventTime, formatDuration, getEventIcon, getEventColor, categoryColors, levelStyles } from './sessionStore';
export type { Session, SessionEvent, SessionFilter } from './sessionStore';

// New unified event system
export * from './eventTypes';
export { useEventStore, useCurrentBookmarks } from './eventStore';

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

// Video Store
export { useVideoStore } from './videoStore';
export type {
  VideoServiceInfo,
  VideoMetadata,
  VideoThumbnail,
  SessionVideoInfo,
  Recording,
} from './videoStore';

// Session Manager Store
export { useSessionManagerStore } from './sessionManagerStore';

// Event Timeline Store
export { useEventTimelineStore } from './eventTimelineStore';

// Wireless Connect Store
export { useWirelessConnectStore } from './wirelessConnectStore';

// Session Config Store
export { useSessionConfigStore } from './sessionConfigStore';

// Virtual Table Store
export { useVirtualTableStore } from './virtualTableStore';
export type { SortConfig } from './virtualTableStore';

// Feedback Store
export { useFeedbackStore } from './feedbackStore';

// Element Picker Store
export { useElementPickerStore } from './elementPickerStore';

// Device Info Modal Store
export { useDeviceInfoModalStore } from './deviceInfoModalStore';

// Batch Operation Store
export { useBatchOperationStore } from './batchOperationStore';

// Assertions Panel Store
export { useAssertionsPanelStore } from './assertionsPanelStore';

// Automation View Store
export { useAutomationViewStore } from './automationViewStore';

// Thumbnail Store (for FilesView)
export { useThumbnailStore } from './thumbnailStore';
