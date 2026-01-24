package main

import (
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// WorkflowWatcher monitors the workflows directory for changes from external processes (e.g., MCP)
type WorkflowWatcher struct {
	app     *App
	watcher *fsnotify.Watcher
	stopCh  chan struct{}
	mu      sync.Mutex
}

// NewWorkflowWatcher creates a new workflow directory watcher
func NewWorkflowWatcher(app *App) *WorkflowWatcher {
	return &WorkflowWatcher{
		app:    app,
		stopCh: make(chan struct{}),
	}
}

// Start begins watching the workflows directory
func (w *WorkflowWatcher) Start() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Don't start watcher in MCP mode (no GUI to notify)
	if w.app.mcpMode {
		return nil
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	w.watcher = watcher

	workflowsPath := w.app.getWorkflowsPath()
	if err := watcher.Add(workflowsPath); err != nil {
		watcher.Close()
		return err
	}

	LogInfo("workflow_watcher").Str("path", workflowsPath).Msg("Started watching workflows directory")

	go w.watch()
	return nil
}

// Stop stops watching the workflows directory
func (w *WorkflowWatcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.watcher != nil {
		close(w.stopCh)
		w.watcher.Close()
		w.watcher = nil
		LogInfo("workflow_watcher").Msg("Stopped watching workflows directory")
	}
}

// watch is the main watch loop
func (w *WorkflowWatcher) watch() {
	// Debounce: wait for events to settle before notifying
	var debounceTimer *time.Timer
	var lastEvent time.Time
	debounceDelay := 300 * time.Millisecond

	notifyChange := func(action, workflowId string) {
		// Skip if too soon after the last internal save (prevent self-triggering)
		if time.Since(lastEvent) < 100*time.Millisecond {
			return
		}

		if !w.app.mcpMode && w.app.ctx != nil {
			wailsRuntime.EventsEmit(w.app.ctx, "workflow-list-changed", map[string]interface{}{
				"action":     action,
				"workflowId": workflowId,
				"external":   true, // Mark as external change (from MCP or file system)
			})
			LogDebug("workflow_watcher").
				Str("action", action).
				Str("workflowId", workflowId).
				Msg("Emitted workflow-list-changed event")
		}
	}

	for {
		select {
		case <-w.stopCh:
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			return

		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			// Only care about JSON files
			if !strings.HasSuffix(event.Name, ".json") {
				continue
			}

			// Extract workflow ID from filename
			workflowId := strings.TrimSuffix(filepath.Base(event.Name), ".json")

			// Debounce: reset timer on each event
			if debounceTimer != nil {
				debounceTimer.Stop()
			}

			action := ""
			switch {
			case event.Op&fsnotify.Create == fsnotify.Create:
				action = "create"
			case event.Op&fsnotify.Write == fsnotify.Write:
				action = "save"
			case event.Op&fsnotify.Remove == fsnotify.Remove:
				action = "delete"
			case event.Op&fsnotify.Rename == fsnotify.Rename:
				action = "delete" // Rename is often used for atomic writes
			}

			if action != "" {
				lastEvent = time.Now()
				debounceTimer = time.AfterFunc(debounceDelay, func() {
					notifyChange(action, workflowId)
				})
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			LogError("workflow_watcher").Err(err).Msg("Watcher error")
		}
	}
}
