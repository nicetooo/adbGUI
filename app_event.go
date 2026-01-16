package main

import (
	"fmt"
	"log"
	"runtime"
	"time"
)

// ========================================
// Event Query API Methods
// ========================================

// QuerySessionEvents queries events from a session
func (a *App) QuerySessionEvents(query EventQuery) (*EventQueryResult, error) {
	log.Printf("[QuerySessionEvents] Called with sessionId=%s, startTime=%d, endTime=%d, limit=%d, sources=%v, categories=%v",
		query.SessionID, query.StartTime, query.EndTime, query.Limit, query.Sources, query.Categories)
	if a.eventStore == nil {
		log.Printf("[QuerySessionEvents] ERROR: eventStore is nil!")
		return &EventQueryResult{Events: []UnifiedEvent{}}, nil
	}
	result, err := a.eventStore.QueryEvents(query)
	if err != nil {
		log.Printf("[QuerySessionEvents] ERROR: %v", err)
	} else {
		log.Printf("[QuerySessionEvents] Returned %d events, total=%d", len(result.Events), result.Total)
	}
	return result, err
}

// GetStoredEvent gets a single event by ID
func (a *App) GetStoredEvent(eventID string) (*UnifiedEvent, error) {
	if a.eventStore == nil {
		return nil, fmt.Errorf("event store not initialized")
	}
	return a.eventStore.GetEvent(eventID)
}

// GetRecentSessionEvents gets recent events from memory for a session
func (a *App) GetRecentSessionEvents(sessionID string, count int) []UnifiedEvent {
	if a.eventPipeline == nil {
		return nil
	}
	return a.eventPipeline.GetRecentEvents(sessionID, count)
}

// ========================================
// Session Storage API Methods
// ========================================

// GetStoredSession gets a session by ID
func (a *App) GetStoredSession(sessionID string) (*DeviceSession, error) {
	log.Printf("[GetStoredSession] Called with sessionID=%s", sessionID)
	if a.eventStore == nil {
		log.Printf("[GetStoredSession] ERROR: eventStore is nil!")
		return nil, fmt.Errorf("event store not initialized")
	}
	session, err := a.eventStore.GetSession(sessionID)
	log.Printf("[GetStoredSession] Result: session=%+v, err=%v", session, err)
	return session, err
}

// ListStoredSessions lists sessions from storage
func (a *App) ListStoredSessions(deviceID string, limit int) ([]DeviceSession, error) {
	if a.eventStore == nil {
		return []DeviceSession{}, nil
	}
	return a.eventStore.ListSessions(deviceID, limit)
}

// DeleteStoredSession deletes a session and its events
func (a *App) DeleteStoredSession(sessionID string) error {
	if a.eventStore == nil {
		return nil
	}
	return a.eventStore.DeleteSession(sessionID)
}

// RenameStoredSession renames a session
func (a *App) RenameStoredSession(sessionID, newName string) error {
	if a.eventStore == nil {
		return nil
	}
	return a.eventStore.RenameSession(sessionID, newName)
}

// ========================================
// Session Metadata API Methods
// ========================================

// GetSessionTimeIndex gets the time index for a session
func (a *App) GetSessionTimeIndex(sessionID string) ([]TimeIndexEntry, error) {
	if a.eventStore == nil {
		return []TimeIndexEntry{}, nil
	}
	return a.eventStore.GetTimeIndex(sessionID)
}

// GetSessionStats gets statistics for a session
func (a *App) GetSessionStats(sessionID string) (map[string]interface{}, error) {
	if a.eventStore == nil {
		return nil, fmt.Errorf("event store not initialized")
	}
	return a.eventStore.GetSessionStats(sessionID)
}

// GetSessionEventTypes gets distinct event types in a session
func (a *App) GetSessionEventTypes(sessionID string) ([]string, error) {
	if a.eventStore == nil {
		return []string{}, nil
	}
	return a.eventStore.GetEventTypes(sessionID)
}

// GetSessionEventSources gets distinct event sources in a session
func (a *App) GetSessionEventSources(sessionID string) ([]string, error) {
	if a.eventStore == nil {
		return []string{}, nil
	}
	return a.eventStore.GetEventSources(sessionID)
}

// GetSessionEventLevels gets distinct event levels in a session
func (a *App) GetSessionEventLevels(sessionID string) ([]string, error) {
	if a.eventStore == nil {
		return []string{}, nil
	}
	return a.eventStore.GetEventLevels(sessionID)
}

// PreviewAssertionMatch previews the count of events matching assertion criteria
func (a *App) PreviewAssertionMatch(sessionID string, types []string, titleMatch string) (int, error) {
	if a.eventStore == nil {
		return 0, nil
	}
	return a.eventStore.PreviewAssertionMatch(sessionID, types, titleMatch)
}

// ========================================
// Bookmark API Methods
// ========================================

// CreateSessionBookmark creates a bookmark in a session
func (a *App) CreateSessionBookmark(sessionID string, relativeTime int64, label, color, bookmarkType string) error {
	if a.eventStore == nil {
		return nil
	}
	bookmark := &Bookmark{
		ID:           fmt.Sprintf("bm_%d", time.Now().UnixNano()),
		SessionID:    sessionID,
		RelativeTime: relativeTime,
		Label:        label,
		Color:        color,
		Type:         bookmarkType,
		CreatedAt:    time.Now().UnixMilli(),
	}
	return a.eventStore.CreateBookmark(bookmark)
}

// GetSessionBookmarks gets bookmarks for a session
func (a *App) GetSessionBookmarks(sessionID string) ([]Bookmark, error) {
	if a.eventStore == nil {
		return []Bookmark{}, nil
	}
	return a.eventStore.GetBookmarks(sessionID)
}

// DeleteSessionBookmark deletes a bookmark
func (a *App) DeleteSessionBookmark(bookmarkID string) error {
	if a.eventStore == nil {
		return nil
	}
	return a.eventStore.DeleteBookmark(bookmarkID)
}

// ========================================
// Event System Management API Methods
// ========================================

// CleanupOldSessionData cleans up old session data
func (a *App) CleanupOldSessionData(maxAgeDays int) (int, error) {
	if a.eventStore == nil {
		return 0, nil
	}
	return a.eventStore.CleanupOldSessions(time.Duration(maxAgeDays) * 24 * time.Hour)
}

// GetEventSystemStats returns statistics about the event system
func (a *App) GetEventSystemStats() map[string]interface{} {
	stats := make(map[string]interface{})

	if a.eventPipeline != nil {
		stats["backpressure"] = a.eventPipeline.GetBackpressureStats()
		stats["pipeline"] = a.eventPipeline.GetPipelineStats()
	}

	if a.eventStore != nil {
		stats["dataDir"] = a.dataDir
		stats["store"] = a.eventStore.GetStoreStats()
	}

	// Runtime stats
	stats["runtime"] = GetRuntimeStats()

	return stats
}

// GetRuntimeStats returns Go runtime statistics
func GetRuntimeStats() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"goroutines":   runtime.NumGoroutine(),
		"heapAlloc":    m.HeapAlloc,     // bytes allocated and in use
		"heapSys":      m.HeapSys,       // bytes obtained from system
		"heapObjects":  m.HeapObjects,   // total number of allocated objects
		"gcCycles":     m.NumGC,         // number of completed GC cycles
		"gcPauseTotal": m.PauseTotalNs,  // total GC pause time in nanoseconds
		"cpus":         runtime.NumCPU(),
	}
}
