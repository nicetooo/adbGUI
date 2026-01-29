package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ========================================
// Assertion Execution API Methods
// ========================================

// ExecuteAssertion executes an assertion immediately
func (a *App) ExecuteAssertion(assertion Assertion) (*AssertionResult, error) {
	if a.assertionEngine == nil {
		return nil, fmt.Errorf("assertion engine not initialized")
	}
	return a.assertionEngine.ExecuteAssertion(&assertion)
}

// ExecuteAssertionJSON executes an assertion from JSON (for frontend)
func (a *App) ExecuteAssertionJSON(assertionJSON string) (*AssertionResult, error) {
	if a.assertionEngine == nil {
		return nil, fmt.Errorf("assertion engine not initialized")
	}

	var assertion Assertion
	if err := json.Unmarshal([]byte(assertionJSON), &assertion); err != nil {
		return nil, fmt.Errorf("invalid assertion JSON: %v", err)
	}

	return a.assertionEngine.ExecuteAssertion(&assertion)
}

// ListAssertionResults lists assertion results for a session
func (a *App) ListAssertionResults(sessionID string, limit int) []*AssertionResult {
	if a.assertionEngine == nil {
		return nil
	}
	return a.assertionEngine.ListResults(sessionID, limit)
}

// ========================================
// Quick Assertion Methods
// ========================================

// QuickAssertExists creates and executes a quick "exists" assertion
func (a *App) QuickAssertExists(sessionID, deviceID, eventType, titleMatch string) (*AssertionResult, error) {
	if a.assertionEngine == nil {
		return nil, fmt.Errorf("assertion engine not initialized")
	}

	assertion := Assertion{
		ID:        fmt.Sprintf("quick_%d", time.Now().UnixNano()),
		Name:      fmt.Sprintf("Quick check: %s exists", eventType),
		Type:      AssertExists,
		SessionID: sessionID,
		DeviceID:  deviceID,
		Criteria: EventCriteria{
			Types:      []string{eventType},
			TitleMatch: titleMatch,
		},
		Expected: AssertionExpected{
			Exists: true,
		},
		CreatedAt: time.Now().UnixMilli(),
	}

	return a.assertionEngine.ExecuteAssertion(&assertion)
}

// QuickAssertCount creates and executes a quick "count" assertion
func (a *App) QuickAssertCount(sessionID, deviceID, eventType string, minCount, maxCount int) (*AssertionResult, error) {
	if a.assertionEngine == nil {
		return nil, fmt.Errorf("assertion engine not initialized")
	}

	assertion := Assertion{
		ID:        fmt.Sprintf("quick_%d", time.Now().UnixNano()),
		Name:      fmt.Sprintf("Quick count: %s", eventType),
		Type:      AssertCount,
		SessionID: sessionID,
		DeviceID:  deviceID,
		Criteria: EventCriteria{
			Types: []string{eventType},
		},
		Expected: AssertionExpected{
			MinCount: &minCount,
			MaxCount: &maxCount,
		},
		CreatedAt: time.Now().UnixMilli(),
	}

	return a.assertionEngine.ExecuteAssertion(&assertion)
}

// QuickAssertNoErrors creates and executes a quick "no errors" assertion
func (a *App) QuickAssertNoErrors(sessionID, deviceID string) (*AssertionResult, error) {
	if a.assertionEngine == nil {
		return nil, fmt.Errorf("assertion engine not initialized")
	}

	assertion := Assertion{
		ID:        fmt.Sprintf("quick_%d", time.Now().UnixNano()),
		Name:      "Quick check: no errors",
		Type:      AssertNotExists,
		SessionID: sessionID,
		DeviceID:  deviceID,
		Criteria: EventCriteria{
			Levels: []EventLevel{LevelError, LevelFatal},
		},
		Expected: AssertionExpected{
			Exists: false,
		},
		CreatedAt: time.Now().UnixMilli(),
	}

	return a.assertionEngine.ExecuteAssertion(&assertion)
}

// QuickAssertNoCrashes creates and executes a quick "no crashes" assertion
func (a *App) QuickAssertNoCrashes(sessionID, deviceID string) (*AssertionResult, error) {
	if a.assertionEngine == nil {
		return nil, fmt.Errorf("assertion engine not initialized")
	}

	assertion := Assertion{
		ID:        fmt.Sprintf("quick_%d", time.Now().UnixNano()),
		Name:      "Quick check: no crashes/ANR",
		Type:      AssertNotExists,
		SessionID: sessionID,
		DeviceID:  deviceID,
		Criteria: EventCriteria{
			Types: []string{"app_crash", "app_anr"},
		},
		Expected: AssertionExpected{
			Exists: false,
		},
		CreatedAt: time.Now().UnixMilli(),
	}

	return a.assertionEngine.ExecuteAssertion(&assertion)
}

// QuickAssertSequence creates and executes a quick "sequence" assertion
func (a *App) QuickAssertSequence(sessionID, deviceID string, eventTypes []string) (*AssertionResult, error) {
	if a.assertionEngine == nil {
		return nil, fmt.Errorf("assertion engine not initialized")
	}

	sequence := make([]EventCriteria, len(eventTypes))
	for i, eventType := range eventTypes {
		sequence[i] = EventCriteria{Types: []string{eventType}}
	}

	assertion := Assertion{
		ID:        fmt.Sprintf("quick_%d", time.Now().UnixNano()),
		Name:      fmt.Sprintf("Quick sequence: %s", strings.Join(eventTypes, " -> ")),
		Type:      AssertSequence,
		SessionID: sessionID,
		DeviceID:  deviceID,
		Criteria:  EventCriteria{},
		Expected: AssertionExpected{
			Sequence: sequence,
			Ordered:  true,
		},
		CreatedAt: time.Now().UnixMilli(),
	}

	return a.assertionEngine.ExecuteAssertion(&assertion)
}

// ========================================
// Assertion Management API Methods
// ========================================

// CreateStoredAssertion creates and persists a new assertion
func (a *App) CreateStoredAssertion(assertion Assertion, saveAsTemplate bool) error {
	if a.assertionEngine == nil {
		return fmt.Errorf("assertion engine not initialized")
	}
	return a.assertionEngine.CreateAssertion(&assertion, saveAsTemplate)
}

// CreateStoredAssertionJSON creates and persists a new assertion from JSON
func (a *App) CreateStoredAssertionJSON(assertionJSON string, saveAsTemplate bool) error {
	if a.assertionEngine == nil {
		return fmt.Errorf("assertion engine not initialized")
	}

	var assertion Assertion
	if err := json.Unmarshal([]byte(assertionJSON), &assertion); err != nil {
		return fmt.Errorf("invalid assertion JSON: %v", err)
	}

	return a.assertionEngine.CreateAssertion(&assertion, saveAsTemplate)
}

// GetStoredAssertion retrieves a stored assertion by ID
func (a *App) GetStoredAssertion(assertionID string) (*StoredAssertion, error) {
	if a.assertionEngine == nil {
		return nil, fmt.Errorf("assertion engine not initialized")
	}
	return a.assertionEngine.GetStoredAssertion(assertionID)
}

// UpdateStoredAssertionJSON updates an existing stored assertion
func (a *App) UpdateStoredAssertionJSON(assertionID string, assertionJSON string) error {
	if a.assertionEngine == nil {
		return fmt.Errorf("assertion engine not initialized")
	}

	var assertion Assertion
	if err := json.Unmarshal([]byte(assertionJSON), &assertion); err != nil {
		return fmt.Errorf("invalid assertion JSON: %v", err)
	}

	// Ensure ID matches
	assertion.ID = assertionID

	return a.assertionEngine.UpdateAssertion(&assertion)
}

// ListStoredAssertions lists stored assertions
func (a *App) ListStoredAssertions(sessionID, deviceID string, templatesOnly bool, limit int) ([]StoredAssertion, error) {
	if a.assertionEngine == nil {
		return nil, fmt.Errorf("assertion engine not initialized")
	}
	return a.assertionEngine.ListStoredAssertions(sessionID, deviceID, templatesOnly, limit)
}

// DeleteStoredAssertion deletes a stored assertion
func (a *App) DeleteStoredAssertion(assertionID string) error {
	if a.assertionEngine == nil {
		return fmt.Errorf("assertion engine not initialized")
	}
	return a.assertionEngine.DeleteStoredAssertion(assertionID)
}

// ExecuteStoredAssertionInSession executes a stored assertion in a specific session context
func (a *App) ExecuteStoredAssertionInSession(assertionID, sessionID, deviceID string) (*AssertionResult, error) {
	if a.assertionEngine == nil {
		return nil, fmt.Errorf("assertion engine not initialized")
	}

	stored, err := a.assertionEngine.GetStoredAssertion(assertionID)
	if err != nil {
		return nil, err
	}
	if stored == nil {
		return nil, fmt.Errorf("assertion not found: %s", assertionID)
	}

	// Convert StoredAssertion to Assertion
	var criteria EventCriteria
	if err := json.Unmarshal(stored.Criteria, &criteria); err != nil {
		return nil, fmt.Errorf("invalid criteria: %v", err)
	}
	var expected AssertionExpected
	if err := json.Unmarshal(stored.Expected, &expected); err != nil {
		return nil, fmt.Errorf("invalid expected: %v", err)
	}
	var metadata map[string]interface{}
	if len(stored.Metadata) > 0 {
		if err := json.Unmarshal(stored.Metadata, &metadata); err != nil {
			LogWarn("assertion").Err(err).Str("assertionId", stored.ID).Msg("Failed to unmarshal assertion metadata")
		}
	}

	// Use provided sessionID/deviceID to override stored values (for global assertions)
	effectiveSessionID := sessionID
	if effectiveSessionID == "" {
		effectiveSessionID = stored.SessionID
	}
	effectiveDeviceID := deviceID
	if effectiveDeviceID == "" {
		effectiveDeviceID = stored.DeviceID
	}

	assertion := Assertion{
		ID:          stored.ID,
		Name:        stored.Name,
		Description: stored.Description,
		Type:        AssertionType(stored.Type),
		SessionID:   effectiveSessionID,
		DeviceID:    effectiveDeviceID,
		Criteria:    criteria,
		Expected:    expected,
		Timeout:     stored.Timeout,
		Tags:        stored.Tags,
		Metadata:    metadata,
		CreatedAt:   stored.CreatedAt,
	}

	if stored.TimeRange != nil {
		assertion.TimeRange = &TimeRange{
			Start: stored.TimeRange.Start,
			End:   stored.TimeRange.End,
		}
	}

	return a.assertionEngine.ExecuteAssertion(&assertion)
}
