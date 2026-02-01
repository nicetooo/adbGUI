package main

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/reporter"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// protoreflectMessageDescriptor is an alias for convenience.
type protoreflectMessageDescriptor = protoreflect.MessageDescriptor

// ProtoCompiler compiles .proto files at runtime and provides message descriptors.
type ProtoCompiler struct {
	mu          sync.RWMutex
	descriptors map[string]protoreflect.MessageDescriptor // fullName -> descriptor
}

// NewProtoCompiler creates a new compiler.
func NewProtoCompiler() *ProtoCompiler {
	return &ProtoCompiler{
		descriptors: make(map[string]protoreflect.MessageDescriptor),
	}
}

// Compile compiles a set of .proto files (name→content) and indexes all message types.
func (c *ProtoCompiler) Compile(files map[string]string) error {
	if len(files) == 0 {
		c.mu.Lock()
		c.descriptors = make(map[string]protoreflect.MessageDescriptor)
		c.mu.Unlock()
		return nil
	}

	// Build an in-memory resolver for the source files
	accessor := protocompile.SourceAccessorFromMap(files)

	// Collect file names
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}

	// Create a silent reporter to avoid printing to stderr
	silentReporter := reporter.NewReporter(
		func(err reporter.ErrorWithPos) error { return err },
		func(err reporter.ErrorWithPos) {},
	)

	// Use WithStandardImports to make well-known types (google/protobuf/*)
	// available for import resolution without needing to provide their source.
	sourceResolver := &protocompile.SourceResolver{
		Accessor: accessor,
	}
	compiler := protocompile.Compiler{
		Resolver: protocompile.WithStandardImports(sourceResolver),
		Reporter: silentReporter,
	}

	compiled, err := compiler.Compile(context.Background(), names...)
	if err != nil {
		return fmt.Errorf("proto compile error: %w", err)
	}

	// Index all message types
	descs := make(map[string]protoreflect.MessageDescriptor)
	for _, f := range compiled {
		indexMessages(f.Messages(), descs)
	}

	c.mu.Lock()
	c.descriptors = descs
	c.mu.Unlock()

	return nil
}

// indexMessages recursively indexes all message types in a file.
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

// Ensure SourceAccessorFromMap signature matches - it returns a func(string) (io.ReadCloser, error)
var _ func(string) (io.ReadCloser, error) = protocompile.SourceAccessorFromMap(nil)
