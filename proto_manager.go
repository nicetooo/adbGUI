package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
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
		log.Printf("[Proto] Compile warning: %v", err)
		// Don't return error — partial compilation is acceptable
	}

	getProtoDecoder().ClearAutoCache()
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

	getProtoDecoder().ClearAutoCache()
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

	getProtoDecoder().ClearAutoCache()
	return saveProtoFiles(reg)
}

// RemoveProtoFile removes a .proto file.
func (a *App) RemoveProtoFile(id string) error {
	reg := getProtoRegistry()
	if err := reg.RemoveFile(id); err != nil {
		return err
	}
	getProtoDecoder().ClearAutoCache()
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

	getProtoDecoder().ClearAutoCache()
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
	getProtoDecoder().ClearAutoCache()
	return saveProtoMappings(reg)
}

// RemoveProtoMapping removes a URL mapping.
func (a *App) RemoveProtoMapping(id string) error {
	reg := getProtoRegistry()
	reg.RemoveMapping(id)
	getProtoDecoder().ClearAutoCache()
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

// LoadProtoFromURL fetches a .proto file from a remote URL and automatically
// resolves and downloads all `import` dependencies recursively.
// Well-known imports (google/protobuf/*) are skipped (handled by protocompile).
// Returns (ids, error) for all added files.
func (a *App) LoadProtoFromURL(rawURL string) ([]string, error) {
	if rawURL == "" {
		return nil, fmt.Errorf("URL is required")
	}

	reg := getProtoRegistry()
	existingNames := make(map[string]bool)
	reg.mu.RLock()
	for _, f := range reg.files {
		existingNames[f.Name] = true
	}
	reg.mu.RUnlock()

	// Collected files: importPath -> content
	collected := make(map[string]string)
	visited := make(map[string]bool) // URLs already fetched

	// Resolve the base directory URL for relative import resolution
	baseDir := rawURL
	if idx := strings.LastIndex(baseDir, "/"); idx != -1 {
		baseDir = baseDir[:idx]
	}

	var resolveErr error
	var resolve func(fetchURL, importPath string)
	resolve = func(fetchURL, importPath string) {
		if resolveErr != nil || visited[fetchURL] {
			return
		}
		visited[fetchURL] = true

		content, err := fetchProtoURL(fetchURL)
		if err != nil {
			resolveErr = fmt.Errorf("failed to fetch %s: %w", importPath, err)
			return
		}

		collected[importPath] = content

		// Parse import statements and resolve dependencies
		for _, imp := range parseProtoImports(content) {
			// Skip well-known types (protocompile has built-in support)
			if strings.HasPrefix(imp, "google/protobuf/") {
				continue
			}
			// Skip if we already have this file in registry or collected
			if existingNames[imp] || collected[imp] != "" {
				continue
			}
			// Resolve relative to the base directory of the original URL
			impURL := baseDir + "/" + imp
			resolve(impURL, imp)
		}
	}

	// Start with the root file
	rootName := extractProtoFileName(rawURL)
	resolve(rawURL, rootName)

	if resolveErr != nil {
		return nil, resolveErr
	}

	// Add all collected files
	var ids []string
	for name, content := range collected {
		id, err := a.AddProtoFile(name, content)
		if err != nil {
			// Non-fatal: log and continue (partial success)
			log.Printf("[Proto] Warning: failed to add %s: %v", name, err)
			continue
		}
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return nil, fmt.Errorf("no files were added")
	}
	return ids, nil
}

// fetchProtoURL downloads content from a URL.
func fetchProtoURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	content := string(body)
	if len(strings.TrimSpace(content)) == 0 {
		return "", fmt.Errorf("empty content")
	}
	return content, nil
}

// parseProtoImports extracts import paths from .proto file content.
// Matches: import "path/to/file.proto"; and import public "path/to/file.proto";
func parseProtoImports(content string) []string {
	var imports []string
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		// Match: import "..." or import public "..."
		if !strings.HasPrefix(line, "import ") {
			continue
		}
		// Extract the quoted path
		q1 := strings.IndexByte(line, '"')
		if q1 < 0 {
			continue
		}
		q2 := strings.IndexByte(line[q1+1:], '"')
		if q2 < 0 {
			continue
		}
		imp := line[q1+1 : q1+1+q2]
		if imp != "" {
			imports = append(imports, imp)
		}
	}
	return imports
}

// extractProtoFileName extracts a clean filename from a URL.
func extractProtoFileName(rawURL string) string {
	name := filepath.Base(rawURL)
	if idx := strings.IndexByte(name, '?'); idx != -1 {
		name = name[:idx]
	}
	if !strings.HasSuffix(name, ".proto") {
		name += ".proto"
	}
	return name
}

// LoadProtoFromDisk opens a file dialog to select local .proto files and adds them.
// Automatically resolves `import` dependencies from the same directory tree.
// Returns ([]id, error) for all successfully added files.
func (a *App) LoadProtoFromDisk() ([]string, error) {
	paths, err := wailsRuntime.OpenMultipleFilesDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select Proto Files",
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "Proto Files (*.proto)", Pattern: "*.proto"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, nil // user cancelled
	}

	reg := getProtoRegistry()
	existingNames := make(map[string]bool)
	reg.mu.RLock()
	for _, f := range reg.files {
		existingNames[f.Name] = true
	}
	reg.mu.RUnlock()

	// Collect all files: importPath -> content
	collected := make(map[string]string)
	visited := make(map[string]bool) // absolute paths already read

	// Determine a common base directory from selected files for resolving imports
	// Use the directory of the first selected file as base
	baseDir := filepath.Dir(paths[0])

	var resolveLocal func(absPath, importPath string)
	resolveLocal = func(absPath, importPath string) {
		if visited[absPath] {
			return
		}
		visited[absPath] = true

		data, err := os.ReadFile(absPath)
		if err != nil {
			log.Printf("[Proto] Warning: cannot read %s: %v", importPath, err)
			return
		}
		content := string(data)
		if len(strings.TrimSpace(content)) == 0 {
			return
		}
		collected[importPath] = content

		// Resolve imports relative to the base directory
		dir := filepath.Dir(absPath)
		for _, imp := range parseProtoImports(content) {
			if strings.HasPrefix(imp, "google/protobuf/") {
				continue
			}
			if existingNames[imp] || collected[imp] != "" {
				continue
			}
			// Try resolving relative to current file's directory first, then base dir
			impAbs := filepath.Join(dir, imp)
			if _, err := os.Stat(impAbs); err != nil {
				impAbs = filepath.Join(baseDir, imp)
				if _, err := os.Stat(impAbs); err != nil {
					log.Printf("[Proto] Warning: import %q not found locally", imp)
					continue
				}
			}
			resolveLocal(impAbs, imp)
		}
	}

	// Process each selected file and resolve its dependencies
	for _, p := range paths {
		name := filepath.Base(p)
		resolveLocal(p, name)
	}

	// Add all collected files
	var ids []string
	var addErrors []string
	for name, content := range collected {
		id, err := a.AddProtoFile(name, content)
		if err != nil {
			addErrors = append(addErrors, fmt.Sprintf("%s: %v", name, err))
			continue
		}
		ids = append(ids, id)
	}

	if len(addErrors) > 0 && len(ids) == 0 {
		return nil, fmt.Errorf("all files failed: %s", strings.Join(addErrors, "; "))
	}
	return ids, nil
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
