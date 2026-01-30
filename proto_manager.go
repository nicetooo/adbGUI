package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Global proto registry and decoder, shared across the app.
var (
	protoRegistry *ProtoRegistry
	protoDecoder  *ProtobufDecoder
	protoOnce     sync.Once
)

func getProtoRegistry() *ProtoRegistry {
	protoOnce.Do(func() {
		protoRegistry = NewProtoRegistry()
		protoDecoder = NewProtobufDecoder(protoRegistry)
	})
	return protoRegistry
}

func getProtoDecoder() *ProtobufDecoder {
	getProtoRegistry() // ensure initialized
	return protoDecoder
}

// --- Persistence paths ---

func getProtoDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".adbGUI", "proto")
}

func getProtoFilesPath() string {
	return filepath.Join(getProtoDir(), "files.json")
}

func getProtoMappingsPath() string {
	return filepath.Join(getProtoDir(), "mappings.json")
}

// --- App methods (Wails bindings) ---

// LoadProtoConfig loads proto files and mappings from disk. Called on app startup.
func (a *App) LoadProtoConfig() error {
	reg := getProtoRegistry()

	dir := getProtoDir()
	_ = os.MkdirAll(dir, 0755)

	// Load files
	if data, err := os.ReadFile(getProtoFilesPath()); err == nil {
		var files []*ProtoFileEntry
		if err := json.Unmarshal(data, &files); err == nil {
			reg.mu.Lock()
			for _, f := range files {
				reg.files[f.ID] = f
			}
			reg.mu.Unlock()
		}
	}

	// Load mappings
	if data, err := os.ReadFile(getProtoMappingsPath()); err == nil {
		var mappings []*ProtoMapping
		if err := json.Unmarshal(data, &mappings); err == nil {
			reg.mu.Lock()
			for _, m := range mappings {
				reg.mappings[m.ID] = m
			}
			reg.mu.Unlock()
		}
	}

	// Compile all loaded files
	if err := reg.recompile(); err != nil {
		fmt.Fprintf(os.Stderr, "[Proto] Compile warning: %v\n", err)
		// Don't return error — partial compilation is acceptable
	}

	return nil
}

// AddProtoFile adds a .proto file definition. Content is the raw .proto text.
func (a *App) AddProtoFile(name, content string) (string, error) {
	if name == "" || content == "" {
		return "", fmt.Errorf("name and content are required")
	}

	entry := &ProtoFileEntry{
		ID:       uuid.New().String(),
		Name:     name,
		Content:  content,
		LoadedAt: time.Now().UnixMilli(),
	}

	reg := getProtoRegistry()
	if err := reg.AddFile(entry); err != nil {
		return "", fmt.Errorf("compile error: %w", err)
	}

	_ = saveProtoFiles(reg)
	return entry.ID, nil
}

// UpdateProtoFile updates an existing .proto file's content.
func (a *App) UpdateProtoFile(id, name, content string) error {
	reg := getProtoRegistry()

	reg.mu.RLock()
	existing, ok := reg.files[id]
	reg.mu.RUnlock()
	if !ok {
		return fmt.Errorf("proto file not found: %s", id)
	}

	entry := &ProtoFileEntry{
		ID:       id,
		Name:     name,
		Content:  content,
		LoadedAt: existing.LoadedAt,
	}

	if err := reg.AddFile(entry); err != nil {
		return fmt.Errorf("compile error: %w", err)
	}

	return saveProtoFiles(reg)
}

// RemoveProtoFile removes a .proto file.
func (a *App) RemoveProtoFile(id string) error {
	reg := getProtoRegistry()
	if err := reg.RemoveFile(id); err != nil {
		return err
	}
	return saveProtoFiles(reg)
}

// GetProtoFiles returns all loaded .proto files.
func (a *App) GetProtoFiles() []*ProtoFileEntry {
	reg := getProtoRegistry()
	files := reg.GetFiles()
	sort.Slice(files, func(i, j int) bool {
		return files[i].LoadedAt < files[j].LoadedAt
	})
	return files
}

// AddProtoMapping adds a URL→message type mapping.
func (a *App) AddProtoMapping(urlPattern, messageType, direction, description string) (string, error) {
	if urlPattern == "" || messageType == "" {
		return "", fmt.Errorf("urlPattern and messageType are required")
	}
	if direction == "" {
		direction = "response"
	}

	m := &ProtoMapping{
		ID:          uuid.New().String(),
		URLPattern:  urlPattern,
		MessageType: messageType,
		Direction:   direction,
		Description: description,
	}

	reg := getProtoRegistry()
	reg.AddMapping(m)

	_ = saveProtoMappings(reg)
	return m.ID, nil
}

// UpdateProtoMapping updates an existing mapping.
func (a *App) UpdateProtoMapping(id, urlPattern, messageType, direction, description string) error {
	reg := getProtoRegistry()

	reg.mu.RLock()
	_, ok := reg.mappings[id]
	reg.mu.RUnlock()
	if !ok {
		return fmt.Errorf("mapping not found: %s", id)
	}

	m := &ProtoMapping{
		ID:          id,
		URLPattern:  urlPattern,
		MessageType: messageType,
		Direction:   direction,
		Description: description,
	}

	reg.AddMapping(m)
	return saveProtoMappings(reg)
}

// RemoveProtoMapping removes a URL mapping.
func (a *App) RemoveProtoMapping(id string) error {
	reg := getProtoRegistry()
	reg.RemoveMapping(id)
	return saveProtoMappings(reg)
}

// GetProtoMappings returns all URL→message mappings.
func (a *App) GetProtoMappings() []*ProtoMapping {
	reg := getProtoRegistry()
	return reg.GetMappings()
}

// GetProtoMessageTypes returns all available message type names from compiled protos.
func (a *App) GetProtoMessageTypes() []string {
	reg := getProtoRegistry()
	types := reg.GetAvailableMessageTypes()
	sort.Strings(types)
	return types
}

// --- Persistence helpers ---

func saveProtoFiles(reg *ProtoRegistry) error {
	files := reg.GetFiles()
	sort.Slice(files, func(i, j int) bool {
		return files[i].LoadedAt < files[j].LoadedAt
	})

	data, err := json.MarshalIndent(files, "", "  ")
	if err != nil {
		return err
	}

	dir := getProtoDir()
	_ = os.MkdirAll(dir, 0755)
	return os.WriteFile(getProtoFilesPath(), data, 0644)
}

func saveProtoMappings(reg *ProtoRegistry) error {
	mappings := reg.GetMappings()

	data, err := json.MarshalIndent(mappings, "", "  ")
	if err != nil {
		return err
	}

	dir := getProtoDir()
	_ = os.MkdirAll(dir, 0755)
	return os.WriteFile(getProtoMappingsPath(), data, 0644)
}
