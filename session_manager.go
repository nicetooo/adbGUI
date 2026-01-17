package main

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// ========================================
// Unified Session Manager
// Provides cross-module event correlation and timeline tracking
// ========================================

// SessionEvent represents a unified event from any module
type SessionEvent struct {
	ID        string `json:"id"`
	SessionID string `json:"sessionId"`
	DeviceID  string `json:"deviceId"`
	Timestamp int64  `json:"timestamp"`          // Unix milliseconds
	Type      string `json:"type"`               // e.g., "workflow_step_start", "logcat", "network_request"
	Category  string `json:"category"`           // "workflow", "log", "network", "automation", "system"
	Level     string `json:"level"`              // "info", "warn", "error", "debug", "verbose"
	Title     string `json:"title"`              // Short description
	Detail    any    `json:"detail"`             // Type-specific payload
	StepID    string `json:"stepId,omitempty"`   // Associated workflow step ID
	Duration  int64  `json:"duration,omitempty"` // Duration in ms (for completed events)
	Success   *bool  `json:"success,omitempty"`  // Success status (for completed events)
}

// Session represents an active or completed session
type Session struct {
	ID         string         `json:"id"`
	DeviceID   string         `json:"deviceId"`
	Type       string         `json:"type"`       // "workflow", "recording", "debug", "manual"
	Name       string         `json:"name"`       // Human-readable name
	StartTime  int64          `json:"startTime"`  // Unix milliseconds
	EndTime    int64          `json:"endTime"`    // 0 if still active
	Status     string         `json:"status"`     // "active", "completed", "failed", "cancelled"
	EventCount int            `json:"eventCount"` // Total events in this session
	Metadata   map[string]any `json:"metadata"`   // Additional session data
}

// SessionFilter for querying events
type SessionFilter struct {
	SessionID  string   `json:"sessionId,omitempty"`
	DeviceID   string   `json:"deviceId,omitempty"`
	Categories []string `json:"categories,omitempty"` // Filter by category
	Types      []string `json:"types,omitempty"`      // Filter by type
	Levels     []string `json:"levels,omitempty"`     // Filter by level
	StepID     string   `json:"stepId,omitempty"`     // Filter by step
	StartTime  int64    `json:"startTime,omitempty"`  // Events after this time
	EndTime    int64    `json:"endTime,omitempty"`    // Events before this time
	Limit      int      `json:"limit,omitempty"`      // Max events to return
	Offset     int      `json:"offset,omitempty"`     // Skip first N events
	SearchText string   `json:"searchText,omitempty"` // Filter by text in title/detail
}

// Session manager state
var (
	sessions            = make(map[string]*Session)
	sessionEvents       = make(map[string][]SessionEvent) // sessionId -> events
	activeSession       = make(map[string]string)         // deviceId -> active sessionId
	sessionMu           sync.RWMutex
	maxEventsPerSession = 50000 // Increased for complete data storage

	// Legacy batch sync (deprecated - now using EventPipeline.frontendEmitter)
	batchSyncTicker *time.Ticker
	batchSyncStop   chan struct{}
)

// ========================================
// Session Lifecycle
// ========================================

// CreateSession starts a new session for a device
func (a *App) CreateSession(deviceId, sessionType, name string) string {
	sessionMu.Lock()
	defer sessionMu.Unlock()

	sessionId := uuid.New().String()
	now := time.Now().UnixMilli()

	session := &Session{
		ID:         sessionId,
		DeviceID:   deviceId,
		Type:       sessionType,
		Name:       name,
		StartTime:  now,
		EndTime:    0,
		Status:     "active",
		EventCount: 0,
		Metadata:   make(map[string]any),
	}

	sessions[sessionId] = session
	// Note: sessionEvents memory storage removed - events stored via EventPipeline
	activeSession[deviceId] = sessionId

	// Emit session started event
	wailsRuntime.EventsEmit(a.ctx, "session-started", session)

	// Also emit as first session event
	a.emitEventInternal(SessionEvent{
		ID:        uuid.New().String(),
		SessionID: sessionId,
		DeviceID:  deviceId,
		Timestamp: now,
		Type:      "session_start",
		Category:  "system",
		Level:     "info",
		Title:     "Session started: " + name,
	})

	return sessionId
}

// EndSession ends an active session
func (a *App) EndSession(sessionId string, status string) error {
	sessionMu.Lock()
	defer sessionMu.Unlock()

	session, exists := sessions[sessionId]
	if !exists {
		return nil // Session doesn't exist, ignore
	}

	now := time.Now().UnixMilli()
	session.EndTime = now
	session.Status = status

	// Remove from active sessions
	if activeSession[session.DeviceID] == sessionId {
		delete(activeSession, session.DeviceID)
	}

	// Emit session end event
	a.emitEventInternal(SessionEvent{
		ID:        uuid.New().String(),
		SessionID: sessionId,
		DeviceID:  session.DeviceID,
		Timestamp: now,
		Type:      "session_end",
		Category:  "system",
		Level:     "info",
		Title:     "Session ended: " + session.Name,
		Duration:  now - session.StartTime,
	})

	wailsRuntime.EventsEmit(a.ctx, "session-ended", session)

	return nil
}

// GetActiveSession returns the active session ID for a device
func (a *App) GetActiveSession(deviceId string) string {
	sessionMu.RLock()
	defer sessionMu.RUnlock()
	return activeSession[deviceId]
}

// EnsureActiveSession ensures there is an active session for the device
func (a *App) EnsureActiveSession(deviceId string) string {
	sessionMu.Lock()
	defer sessionMu.Unlock()

	if id, ok := activeSession[deviceId]; ok {
		return id
	}

	// Create new auto session (bypass CreateSession to hold lock)
	sessionId := uuid.New().String()
	now := time.Now().UnixMilli()

	session := &Session{
		ID:         sessionId,
		DeviceID:   deviceId,
		Type:       "system",
		Name:       "Auto Session " + time.Now().Format("15:04:05"),
		StartTime:  now,
		EndTime:    0,
		Status:     "active",
		EventCount: 0,
		Metadata:   make(map[string]any),
	}

	sessions[sessionId] = session
	// Note: sessionEvents memory storage removed - events stored via EventPipeline
	activeSession[deviceId] = sessionId

	// Emit session started event
	wailsRuntime.EventsEmit(a.ctx, "session-started", session)

	// Add start event via unified pipeline
	event := SessionEvent{
		ID:        uuid.New().String(),
		SessionID: sessionId,
		DeviceID:  deviceId,
		Timestamp: now,
		Type:      "session_start",
		Category:  "system",
		Level:     "info",
		Title:     "Auto Session started",
	}

	// Forward to EventPipeline (unified storage)
	a.bridgeToNewPipeline(event)
	session.EventCount++

	// Broadcast event to frontend
	wailsRuntime.EventsEmit(a.ctx, "session-event", event)

	return sessionId
}

// GetSession returns session details
func (a *App) GetSession(sessionId string) *Session {
	sessionMu.RLock()
	defer sessionMu.RUnlock()
	if session, exists := sessions[sessionId]; exists {
		return session
	}
	return nil
}

// GetSessions returns all sessions, optionally filtered by device
func (a *App) GetSessions(deviceId string, limit int) []Session {
	sessionMu.RLock()
	defer sessionMu.RUnlock()

	result := make([]Session, 0)
	for _, session := range sessions {
		if deviceId == "" || session.DeviceID == deviceId {
			result = append(result, *session)
		}
	}

	// Sort by start time descending (newest first)
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].StartTime > result[i].StartTime {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}

	return result
}

// ========================================
// Event Emission
// ========================================

// EmitSessionEvent emits an event to the active session for a device
// If no active session, the event is still emitted but not stored
func (a *App) EmitSessionEvent(deviceId string, eventType, category, level, title string, detail any) {
	sessionMu.Lock()
	defer sessionMu.Unlock()

	sessionId := activeSession[deviceId]
	event := SessionEvent{
		ID:        uuid.New().String(),
		SessionID: sessionId,
		DeviceID:  deviceId,
		Timestamp: time.Now().UnixMilli(),
		Type:      eventType,
		Category:  category,
		Level:     level,
		Title:     title,
		Detail:    detail,
	}

	a.emitEventInternal(event)
}

// EmitSessionEventWithStep emits an event associated with a workflow step
func (a *App) EmitSessionEventWithStep(deviceId, stepId, eventType, category, level, title string, detail any) {
	sessionMu.Lock()
	defer sessionMu.Unlock()

	sessionId := activeSession[deviceId]
	event := SessionEvent{
		ID:        uuid.New().String(),
		SessionID: sessionId,
		DeviceID:  deviceId,
		Timestamp: time.Now().UnixMilli(),
		Type:      eventType,
		Category:  category,
		Level:     level,
		Title:     title,
		Detail:    detail,
		StepID:    stepId,
	}

	a.emitEventInternal(event)
}

// EmitSessionEventFull emits a fully constructed event
func (a *App) EmitSessionEventFull(event SessionEvent) {
	sessionMu.Lock()
	defer sessionMu.Unlock()

	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}
	if event.SessionID == "" {
		event.SessionID = activeSession[event.DeviceID]
	}

	a.emitEventInternal(event)
}

// emitEventInternal forwards event to EventPipeline (unified storage)
// Note: All events now go through EventPipeline -> EventStore -> frontend
func (a *App) emitEventInternal(event SessionEvent) {
	// Update session event count in memory (for quick access)
	if event.SessionID != "" {
		if session, exists := sessions[event.SessionID]; exists {
			session.EventCount++
		}
	}

	// Forward to unified event pipeline (single path to frontend)
	a.eventSystemMu.RLock()
	pipeline := a.eventPipeline
	a.eventSystemMu.RUnlock()

	if pipeline != nil {
		a.bridgeToNewPipeline(event)
	}
}

// bridgeToNewPipeline converts old SessionEvent to UnifiedEvent and forwards to new pipeline
func (a *App) bridgeToNewPipeline(event SessionEvent) {
	// Map old category to new source
	var source EventSource
	switch event.Category {
	case "log":
		source = SourceLogcat
	case "network":
		source = SourceNetwork
	case "workflow", "automation":
		source = SourceWorkflow
	case "system":
		source = SourceSystem
	default:
		source = SourceSystem
	}

	// Map old category to new category
	var category EventCategory
	switch event.Category {
	case "log":
		category = CategoryLog
	case "network":
		category = CategoryNetwork
	case "workflow", "automation":
		category = CategoryAutomation
	case "system":
		category = CategoryDiagnostic
	default:
		category = CategoryDiagnostic
	}

	// Map level
	var level EventLevel
	switch event.Level {
	case "error":
		level = LevelError
	case "warn":
		level = LevelWarn
	case "info":
		level = LevelInfo
	case "debug":
		level = LevelDebug
	case "verbose":
		level = LevelVerbose
	default:
		level = LevelInfo
	}

	// Convert detail to JSON
	var data json.RawMessage
	if event.Detail != nil {
		var err error
		data, err = json.Marshal(event.Detail)
		if err != nil {
			LogWarn("session_manager").Err(err).Str("eventType", event.Type).Msg("Failed to marshal event detail")
			data = []byte("{}")
		}
	}

	// Create unified event
	unifiedEvent := UnifiedEvent{
		ID:        event.ID,
		SessionID: event.SessionID,
		DeviceID:  event.DeviceID,
		Timestamp: event.Timestamp,
		Source:    source,
		Category:  category,
		Type:      event.Type,
		Level:     level,
		Title:     event.Title,
		Data:      data,
		StepID:    event.StepID,
		Duration:  event.Duration,
	}

	// Forward to new pipeline (non-blocking)
	go a.eventPipeline.Emit(unifiedEvent)
}

// StartBatchSync starts the batch sync ticker (call on app startup)
func (a *App) StartBatchSync() {
	if batchSyncTicker != nil {
		return // Already running
	}

	batchSyncTicker = time.NewTicker(1 * time.Second)
	batchSyncStop = make(chan struct{})

	go func() {
		for {
			select {
			case <-batchSyncTicker.C:
				a.flushEventBuffer()
			case <-batchSyncStop:
				return
			}
		}
	}()
}

// StopBatchSync stops the batch sync ticker
func (a *App) StopBatchSync() {
	if batchSyncTicker != nil {
		batchSyncTicker.Stop()
		close(batchSyncStop)
		batchSyncTicker = nil
	}
}

// flushEventBuffer is deprecated - events now flow through EventPipeline.frontendEmitter
// Kept for compatibility but does nothing
func (a *App) flushEventBuffer() {
	// No-op: EventPipeline handles frontend sync via "session-events-batch"
}

// ========================================
// Event Querying
// ========================================

// GetSessionTimeline returns events for a session with optional filtering
// Now reads from EventStore instead of memory for unified event storage
func (a *App) GetSessionTimeline(sessionId string, filter *SessionFilter) []SessionEvent {
	// Use EventStore for queries (unified storage)
	a.eventSystemMu.RLock()
	pipeline := a.eventPipeline
	a.eventSystemMu.RUnlock()

	if pipeline == nil || pipeline.store == nil {
		return []SessionEvent{}
	}

	// Build query from filter
	query := EventQuery{
		SessionID: sessionId,
		OrderDesc: false, // Oldest first for timeline
	}

	if filter != nil {
		if len(filter.Categories) > 0 {
			for _, cat := range filter.Categories {
				query.Categories = append(query.Categories, EventCategory(cat))
			}
		}
		if len(filter.Types) > 0 {
			query.Types = filter.Types
		}
		if len(filter.Levels) > 0 {
			for _, lvl := range filter.Levels {
				query.Levels = append(query.Levels, EventLevel(lvl))
			}
		}
		if filter.StepID != "" {
			query.StepID = filter.StepID
		}
		if filter.SearchText != "" {
			query.SearchText = filter.SearchText
		}
		if filter.Limit > 0 {
			query.Limit = filter.Limit
		}
		if filter.Offset > 0 {
			query.Offset = filter.Offset
		}
	}

	result, err := pipeline.store.QueryEvents(query)
	if err != nil {
		LogWarn("session_manager").Err(err).Str("sessionId", sessionId).Msg("Failed to query events from store")
		return []SessionEvent{}
	}

	// Convert UnifiedEvents to SessionEvents for API compatibility
	return convertUnifiedToSessionEvents(result.Events)
}

// GetRecentEvents returns recent events across all sessions for a device
// Now reads from EventStore instead of memory for unified event storage
func (a *App) GetRecentEvents(deviceId string, limit int, categories []string) []SessionEvent {
	// Use EventStore for queries (unified storage)
	a.eventSystemMu.RLock()
	pipeline := a.eventPipeline
	a.eventSystemMu.RUnlock()

	if pipeline == nil || pipeline.store == nil {
		return []SessionEvent{}
	}

	// Build query
	query := EventQuery{
		DeviceID:  deviceId,
		OrderDesc: true, // Newest first for recent events
	}

	if len(categories) > 0 {
		for _, cat := range categories {
			query.Categories = append(query.Categories, EventCategory(cat))
		}
	}

	if limit > 0 {
		query.Limit = limit
	} else {
		query.Limit = 100 // Default limit
	}

	result, err := pipeline.store.QueryEvents(query)
	if err != nil {
		LogWarn("session_manager").Err(err).Str("deviceId", deviceId).Msg("Failed to query recent events from store")
		return []SessionEvent{}
	}

	// Convert UnifiedEvents to SessionEvents for API compatibility
	return convertUnifiedToSessionEvents(result.Events)
}

// ========================================
// Session Metadata
// ========================================

// SetSessionMetadata sets a metadata value on a session
func (a *App) SetSessionMetadata(sessionId, key string, value any) {
	sessionMu.Lock()
	defer sessionMu.Unlock()

	if session, exists := sessions[sessionId]; exists {
		session.Metadata[key] = value
	}
}

// GetSessionMetadata gets a metadata value from a session
func (a *App) GetSessionMetadata(sessionId, key string) any {
	sessionMu.RLock()
	defer sessionMu.RUnlock()

	if session, exists := sessions[sessionId]; exists {
		return session.Metadata[key]
	}
	return nil
}

// ========================================
// Cleanup
// ========================================

// CleanupOldSessions removes sessions older than the specified duration
func (a *App) CleanupOldSessions(maxAge time.Duration) int {
	sessionMu.Lock()
	defer sessionMu.Unlock()

	cutoff := time.Now().Add(-maxAge).UnixMilli()
	removed := 0

	for id, session := range sessions {
		if session.EndTime > 0 && session.EndTime < cutoff {
			delete(sessions, id)
			delete(sessionEvents, id)
			removed++
		}
	}

	return removed
}

// ClearSession removes a specific session and its events
func (a *App) ClearSession(sessionId string) {
	sessionMu.Lock()
	defer sessionMu.Unlock()

	if session, exists := sessions[sessionId]; exists {
		if activeSession[session.DeviceID] == sessionId {
			delete(activeSession, session.DeviceID)
		}
	}
	delete(sessions, sessionId)
	delete(sessionEvents, sessionId)
}

// ========================================
// Helper Functions
// ========================================

// convertUnifiedToSessionEvents converts UnifiedEvents to SessionEvents for API compatibility
func convertUnifiedToSessionEvents(events []UnifiedEvent) []SessionEvent {
	result := make([]SessionEvent, len(events))
	for i, e := range events {
		// Convert category back to string
		categoryStr := string(e.Category)
		// Map new category back to old category names for compatibility
		switch e.Category {
		case CategoryLog:
			categoryStr = "log"
		case CategoryNetwork:
			categoryStr = "network"
		case CategoryAutomation:
			categoryStr = "workflow"
		case CategoryState:
			categoryStr = "system"
		case CategoryDiagnostic:
			categoryStr = "system"
		}

		// Parse detail from Data JSON
		var detail interface{}
		if len(e.Data) > 0 {
			if err := json.Unmarshal(e.Data, &detail); err != nil {
				LogWarn("session_manager").Err(err).Str("eventId", e.ID).Msg("Failed to unmarshal event data")
			}
		}

		result[i] = SessionEvent{
			ID:        e.ID,
			SessionID: e.SessionID,
			DeviceID:  e.DeviceID,
			Timestamp: e.Timestamp,
			Type:      e.Type,
			Category:  categoryStr,
			Level:     string(e.Level),
			Title:     e.Title,
			Detail:    detail,
			StepID:    e.StepID,
			Duration:  e.Duration,
		}
	}
	return result
}

func matchesFilter(event SessionEvent, filter *SessionFilter) bool {
	if filter.SessionID != "" && event.SessionID != filter.SessionID {
		return false
	}
	if filter.DeviceID != "" && event.DeviceID != filter.DeviceID {
		return false
	}
	if filter.StepID != "" && event.StepID != filter.StepID {
		return false
	}
	if filter.StartTime > 0 && event.Timestamp < filter.StartTime {
		return false
	}
	if filter.EndTime > 0 && event.Timestamp > filter.EndTime {
		return false
	}
	if len(filter.Categories) > 0 && !containsString(filter.Categories, event.Category) {
		return false
	}
	if len(filter.Types) > 0 && !containsString(filter.Types, event.Type) {
		return false
	}
	if len(filter.Levels) > 0 && !containsString(filter.Levels, event.Level) {
		return false
	}
	if filter.SearchText != "" {
		text := strings.ToLower(filter.SearchText)
		if !strings.Contains(strings.ToLower(event.Title), text) {
			// Check detail
			detailBytes, err := json.Marshal(event.Detail)
			if err != nil {
				return false // 无法序列化，跳过搜索匹配
			}
			if !strings.Contains(strings.ToLower(string(detailBytes)), text) {
				return false
			}
		}
	}
	return true
}

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
