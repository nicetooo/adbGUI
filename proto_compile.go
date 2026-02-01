package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// protoreflectMessageDescriptor is an alias for convenience.
type protoreflectMessageDescriptor = protoreflect.MessageDescriptor

// ProtoCompiler compiles .proto files using Google's official protoc binary
// and provides message descriptors for runtime protobuf decoding.
type ProtoCompiler struct {
	mu          sync.RWMutex
	descriptors map[string]protoreflect.MessageDescriptor // fullName -> descriptor

	// Paths to protoc binary and well-known type includes.
	// These are set once at startup and never change.
	protocPath   string
	protoInclude string // path to directory with google/protobuf/*.proto
}

// NewProtoCompiler creates a new compiler.
// protocPath and protoInclude will be set later via SetPaths().
func NewProtoCompiler() *ProtoCompiler {
	return &ProtoCompiler{
		descriptors: make(map[string]protoreflect.MessageDescriptor),
	}
}

// SetPaths configures the protoc binary and include paths.
// Must be called before Compile().
func (c *ProtoCompiler) SetPaths(protocPath, protoInclude string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.protocPath = protocPath
	c.protoInclude = protoInclude
}

// Compile compiles a set of .proto files (name→content) using protoc and indexes all message types.
// It writes files to a temp directory, runs protoc --descriptor_set_out, parses the
// FileDescriptorSet output, and builds protoreflect descriptors.
func (c *ProtoCompiler) Compile(files map[string]string) error {
	if len(files) == 0 {
		c.mu.Lock()
		c.descriptors = make(map[string]protoreflect.MessageDescriptor)
		c.mu.Unlock()
		return nil
	}

	c.mu.RLock()
	protocPath := c.protocPath
	protoInclude := c.protoInclude
	c.mu.RUnlock()

	if protocPath == "" {
		return fmt.Errorf("protoc binary path not configured")
	}

	// Create temp directory for proto files
	tmpDir, err := os.MkdirTemp("", "gaze-proto-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Pre-process: convert angle bracket `<>` option syntax to `{}`.
	// Some proto files (especially from Google and proto2) use `<>` message literal syntax
	// in option values, which is valid text-format but not accepted by protoc in source files.
	preprocessed := make(map[string]string, len(files))
	for name, content := range files {
		preprocessed[name] = fixAngleBracketOptions(content)
	}

	// Write all proto files to temp directory
	var protoNames []string
	for name, content := range preprocessed {
		filePath := filepath.Join(tmpDir, name)
		// Ensure parent directories exist (e.g. for "subdir/file.proto")
		if dir := filepath.Dir(filePath); dir != tmpDir {
			_ = os.MkdirAll(dir, 0755)
		}
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write proto file %s: %w", name, err)
		}
		protoNames = append(protoNames, name)
	}

	// Output descriptor set file
	descSetPath := filepath.Join(tmpDir, "descriptor_set.pb")

	// Build protoc command
	args := []string{
		"--descriptor_set_out=" + descSetPath,
		"--include_imports",
		"--proto_path=" + tmpDir,
	}

	// Add well-known types include path
	if protoInclude != "" {
		args = append(args, "--proto_path="+protoInclude)
	}

	// Add all proto file names
	args = append(args, protoNames...)

	log.Printf("[Proto] Running protoc with %d files", len(protoNames))

	cmd := exec.Command(protocPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		errMsg := strings.TrimSpace(string(output))
		if errMsg == "" {
			errMsg = err.Error()
		}
		return fmt.Errorf("protoc compile error: %s", errMsg)
	}

	// Read and parse the descriptor set
	descSetBytes, err := os.ReadFile(descSetPath)
	if err != nil {
		return fmt.Errorf("failed to read descriptor set: %w", err)
	}

	fdSet := &descriptorpb.FileDescriptorSet{}
	if err := proto.Unmarshal(descSetBytes, fdSet); err != nil {
		return fmt.Errorf("failed to parse descriptor set: %w", err)
	}

	// Build protoreflect file descriptors from the FileDescriptorSet
	descs, err := buildDescriptors(fdSet, files)
	if err != nil {
		return fmt.Errorf("failed to build descriptors: %w", err)
	}

	c.mu.Lock()
	c.descriptors = descs
	c.mu.Unlock()

	log.Printf("[Proto] Compiled %d files, found %d message types", len(protoNames), len(descs))
	return nil
}

// buildDescriptors converts a FileDescriptorSet into a map of message descriptors.
// Only indexes messages from user-provided files (not well-known imports).
func buildDescriptors(fdSet *descriptorpb.FileDescriptorSet, userFiles map[string]string) (map[string]protoreflect.MessageDescriptor, error) {
	// Build a registry of FileDescriptors so cross-file references resolve
	fileDescs := make(map[string]protoreflect.FileDescriptor, len(fdSet.File))

	for _, fdProto := range fdSet.File {
		// Resolve dependencies first
		deps := make([]protodesc.Resolver, 0)
		for _, dep := range fdProto.Dependency {
			if fd, ok := fileDescs[dep]; ok {
				deps = append(deps, resolverFromFile(fd))
			}
		}

		// Build a combined resolver from dependencies
		var resolver protodesc.Resolver
		if len(deps) == 0 {
			resolver = nil
		} else if len(deps) == 1 {
			resolver = deps[0]
		} else {
			resolver = &multiResolver{resolvers: deps}
		}

		opts := protodesc.FileOptions{
			AllowUnresolvable: true, // Lenient: allow missing deps
		}
		if resolver != nil {
			opts.AllowUnresolvable = false
		}

		fd, err := protodesc.NewFile(fdProto, &resolverWrapper{
			fileDescs: fileDescs,
			inner:     resolver,
		})
		if err != nil {
			// Try again with lenient mode
			fd, err = protodesc.NewFile(fdProto, &lenientResolver{fileDescs: fileDescs})
			if err != nil {
				log.Printf("[Proto] Warning: skipping file %s: %v", fdProto.GetName(), err)
				continue
			}
		}

		fileDescs[fdProto.GetName()] = fd
	}

	// Index message types only from user-provided files
	descs := make(map[string]protoreflect.MessageDescriptor)
	for name, fd := range fileDescs {
		if _, isUser := userFiles[name]; isUser {
			indexMessages(fd.Messages(), descs)
		}
	}

	return descs, nil
}

// resolverWrapper implements protodesc.Resolver by looking up files from a map.
type resolverWrapper struct {
	fileDescs map[string]protoreflect.FileDescriptor
	inner     protodesc.Resolver
}

func (r *resolverWrapper) FindFileByPath(path string) (protoreflect.FileDescriptor, error) {
	if fd, ok := r.fileDescs[path]; ok {
		return fd, nil
	}
	if r.inner != nil {
		return r.inner.FindFileByPath(path)
	}
	return nil, fmt.Errorf("file not found: %s", path)
}

func (r *resolverWrapper) FindDescriptorByName(name protoreflect.FullName) (protoreflect.Descriptor, error) {
	// Search through known files
	for _, fd := range r.fileDescs {
		if d := findInFile(fd, name); d != nil {
			return d, nil
		}
	}
	if r.inner != nil {
		return r.inner.FindDescriptorByName(name)
	}
	return nil, fmt.Errorf("descriptor not found: %s", name)
}

// lenientResolver tries to resolve dependencies but never fails.
type lenientResolver struct {
	fileDescs map[string]protoreflect.FileDescriptor
}

func (r *lenientResolver) FindFileByPath(path string) (protoreflect.FileDescriptor, error) {
	if fd, ok := r.fileDescs[path]; ok {
		return fd, nil
	}
	return nil, fmt.Errorf("file not found: %s", path)
}

func (r *lenientResolver) FindDescriptorByName(name protoreflect.FullName) (protoreflect.Descriptor, error) {
	for _, fd := range r.fileDescs {
		if d := findInFile(fd, name); d != nil {
			return d, nil
		}
	}
	return nil, fmt.Errorf("descriptor not found: %s", name)
}

// findInFile searches a file descriptor for a named descriptor.
func findInFile(fd protoreflect.FileDescriptor, name protoreflect.FullName) protoreflect.Descriptor {
	// Check messages
	if d := findMessageByName(fd.Messages(), name); d != nil {
		return d
	}
	// Check enums
	enums := fd.Enums()
	for i := 0; i < enums.Len(); i++ {
		e := enums.Get(i)
		if e.FullName() == name {
			return e
		}
	}
	// Check extensions
	exts := fd.Extensions()
	for i := 0; i < exts.Len(); i++ {
		ext := exts.Get(i)
		if ext.FullName() == name {
			return ext
		}
	}
	return nil
}

func findMessageByName(msgs protoreflect.MessageDescriptors, name protoreflect.FullName) protoreflect.Descriptor {
	for i := 0; i < msgs.Len(); i++ {
		msg := msgs.Get(i)
		if msg.FullName() == name {
			return msg
		}
		// Check nested messages
		if d := findMessageByName(msg.Messages(), name); d != nil {
			return d
		}
		// Check nested enums
		enums := msg.Enums()
		for j := 0; j < enums.Len(); j++ {
			e := enums.Get(j)
			if e.FullName() == name {
				return e
			}
		}
	}
	return nil
}

// multiResolver chains multiple resolvers.
type multiResolver struct {
	resolvers []protodesc.Resolver
}

func (m *multiResolver) FindFileByPath(path string) (protoreflect.FileDescriptor, error) {
	for _, r := range m.resolvers {
		if fd, err := r.FindFileByPath(path); err == nil {
			return fd, nil
		}
	}
	return nil, fmt.Errorf("file not found: %s", path)
}

func (m *multiResolver) FindDescriptorByName(name protoreflect.FullName) (protoreflect.Descriptor, error) {
	for _, r := range m.resolvers {
		if d, err := r.FindDescriptorByName(name); err == nil {
			return d, nil
		}
	}
	return nil, fmt.Errorf("descriptor not found: %s", name)
}

// resolverFromFile creates a Resolver that serves descriptors from a single FileDescriptor.
func resolverFromFile(fd protoreflect.FileDescriptor) protodesc.Resolver {
	return &singleFileResolver{fd: fd}
}

type singleFileResolver struct {
	fd protoreflect.FileDescriptor
}

func (r *singleFileResolver) FindFileByPath(path string) (protoreflect.FileDescriptor, error) {
	if string(r.fd.Path()) == path {
		return r.fd, nil
	}
	return nil, fmt.Errorf("file not found: %s", path)
}

func (r *singleFileResolver) FindDescriptorByName(name protoreflect.FullName) (protoreflect.Descriptor, error) {
	if d := findInFile(r.fd, name); d != nil {
		return d, nil
	}
	return nil, fmt.Errorf("descriptor not found: %s", name)
}

// indexMessages recursively indexes all message types from message descriptors.
func indexMessages(msgs protoreflect.MessageDescriptors, out map[string]protoreflect.MessageDescriptor) {
	for i := 0; i < msgs.Len(); i++ {
		msg := msgs.Get(i)
		fullName := string(msg.FullName())
		out[fullName] = msg
		// Recurse into nested messages
		indexMessages(msg.Messages(), out)
	}
}

// GetMessageDescriptor returns the descriptor for a fully-qualified message name.
func (c *ProtoCompiler) GetMessageDescriptor(fullName string) protoreflect.MessageDescriptor {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Try exact match first
	if d, ok := c.descriptors[fullName]; ok {
		return d
	}

	// Try case-insensitive / partial match (last component)
	lowerName := strings.ToLower(fullName)
	for k, d := range c.descriptors {
		if strings.ToLower(k) == lowerName {
			return d
		}
		// Match by short name (e.g. "UserResponse" matches "user.UserResponse")
		parts := strings.Split(k, ".")
		if len(parts) > 0 && parts[len(parts)-1] == fullName {
			return d
		}
	}
	return nil
}

// GetAllDescriptors returns a copy of all compiled message descriptors.
func (c *ProtoCompiler) GetAllDescriptors() map[string]protoreflect.MessageDescriptor {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]protoreflect.MessageDescriptor, len(c.descriptors))
	for k, v := range c.descriptors {
		result[k] = v
	}
	return result
}

// GetAllMessageTypes returns all known fully-qualified message type names.
func (c *ProtoCompiler) GetAllMessageTypes() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]string, 0, len(c.descriptors))
	for name := range c.descriptors {
		result = append(result, name)
	}
	return result
}

// --- Preprocessor: angle bracket option syntax ---

// fixAngleBracketOptions converts text-format angle bracket syntax `<...>` to `{...}`
// in protobuf option values. The `<>` message literal syntax is valid in protobuf
// text-format (used extensively in Google's internal protos and many proto2 files),
// but protoc only accepts `{}` in .proto source files.
//
// This handles:
//   - Field options:   [(my_opt) = <name: "test", level: 3>]
//   - Message options:  option (my_opt) = <name: "test">;
//   - Nested brackets: <foo: <bar: 1>>
//
// This does NOT touch:
//   - map<K,V> type declarations (different syntax, preceded by "map")
//   - String literals (inside quotes)
//   - Comments (// and /* */)
func fixAngleBracketOptions(content string) string {
	runes := []rune(content)
	result := make([]rune, 0, len(runes))
	i := 0
	n := len(runes)

	for i < n {
		ch := runes[i]

		// Skip // line comments
		if ch == '/' && i+1 < n && runes[i+1] == '/' {
			start := i
			for i < n && runes[i] != '\n' {
				i++
			}
			result = append(result, runes[start:i]...)
			continue
		}

		// Skip /* block comments */
		if ch == '/' && i+1 < n && runes[i+1] == '*' {
			start := i
			i += 2
			for i+1 < n && !(runes[i] == '*' && runes[i+1] == '/') {
				i++
			}
			if i+1 < n {
				i += 2 // skip */
			}
			result = append(result, runes[start:i]...)
			continue
		}

		// Skip string literals (both single and double quoted)
		if ch == '"' || ch == '\'' {
			quote := ch
			result = append(result, ch)
			i++
			for i < n {
				if runes[i] == '\\' && i+1 < n {
					result = append(result, runes[i], runes[i+1])
					i += 2
					continue
				}
				if runes[i] == quote {
					result = append(result, runes[i])
					i++
					break
				}
				result = append(result, runes[i])
				i++
			}
			continue
		}

		// Check for map<K,V> — skip the <> in map type declarations
		if ch == '<' && isMapType(runes, i) {
			result = append(result, ch)
			i++
			depth := 1
			for i < n && depth > 0 {
				if runes[i] == '<' {
					depth++
				} else if runes[i] == '>' {
					depth--
				}
				result = append(result, runes[i])
				i++
			}
			continue
		}

		// Check for option value context: `= <...>` or `: <...>`
		if ch == '<' && isOptionAngleBracket(runes, result) {
			result = append(result, '{')
			i++
			depth := 1
			for i < n && depth > 0 {
				c := runes[i]

				// Respect strings inside option values
				if c == '"' || c == '\'' {
					quote := c
					result = append(result, c)
					i++
					for i < n {
						if runes[i] == '\\' && i+1 < n {
							result = append(result, runes[i], runes[i+1])
							i += 2
							continue
						}
						if runes[i] == quote {
							result = append(result, runes[i])
							i++
							break
						}
						result = append(result, runes[i])
						i++
					}
					continue
				}

				if c == '<' {
					result = append(result, '{')
					depth++
				} else if c == '>' {
					result = append(result, '}')
					depth--
				} else {
					result = append(result, c)
				}
				i++
			}
			continue
		}

		result = append(result, ch)
		i++
	}

	return string(result)
}

// isMapType checks if the '<' at position i is part of a `map<K,V>` declaration.
func isMapType(runes []rune, i int) bool {
	j := i - 1
	for j >= 0 && (runes[j] == ' ' || runes[j] == '\t') {
		j--
	}
	if j >= 2 && runes[j-2] == 'm' && runes[j-1] == 'a' && runes[j] == 'p' {
		if j-3 < 0 || !isIdentRune(runes[j-3]) {
			return true
		}
	}
	return false
}

// isOptionAngleBracket checks if a '<' is in an option value context.
// Looks backwards through the already-built result for '=' or ':' (skipping whitespace).
func isOptionAngleBracket(runes []rune, result []rune) bool {
	j := len(result) - 1
	for j >= 0 && (result[j] == ' ' || result[j] == '\t' || result[j] == '\n' || result[j] == '\r') {
		j--
	}
	if j < 0 {
		return false
	}
	return result[j] == '=' || result[j] == ':'
}

// isIdentRune returns true if r is a valid identifier character.
func isIdentRune(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}
