package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"sync"

	"github.com/bufbuild/protocompile"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// recoverablePatterns maps error substrings to the fix functions that should be applied.
// When compilation fails, the error message is checked against these patterns.
// If a match is found, the corresponding fix is applied to all files and compilation is retried.
var recoverablePatterns = []struct {
	pattern string
	fixName string
}{
	{"already defined at", "enum_dedup"},         // duplicate enum value names (C++ scoping)
	{"C++ scoping rules for enum", "enum_dedup"}, // related
	{"previously defined at", "enum_dedup"},      // duplicate symbol
	{"unexpected '<'", "angle_brackets"},         // text-format angle bracket syntax
	{"syntax error: unexpected '<'", "angle_brackets"},
	{"unknown '<'", "angle_brackets"},
}

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
// It uses a two-pass strategy: first tries strict compilation, then if a recoverable
// error is detected, pre-processes files to fix the issue and retries.
//
// Recoverable errors include:
//   - Duplicate enum value names across enums (proto3 C++ scoping rules)
//   - Angle bracket `<>` syntax in option values (text-format style, not supported by protocompile)
func (c *ProtoCompiler) Compile(files map[string]string) error {
	if len(files) == 0 {
		c.mu.Lock()
		c.descriptors = make(map[string]protoreflect.MessageDescriptor)
		c.mu.Unlock()
		return nil
	}

	// First pass: try strict compilation
	compiled, err := c.compileFiles(files)
	if err == nil {
		c.indexCompiled(compiled)
		return nil
	}

	// Check if the error matches any recoverable pattern
	errMsg := err.Error()
	fixes := make(map[string]bool) // deduplicated set of fixes to apply
	for _, rp := range recoverablePatterns {
		if strings.Contains(errMsg, rp.pattern) {
			fixes[rp.fixName] = true
		}
	}

	if len(fixes) == 0 {
		return err // not recoverable
	}

	// Second pass: apply all matched fixes and retry
	fixedFiles := make(map[string]string, len(files))
	for name, content := range files {
		fixedFiles[name] = content
	}

	if fixes["angle_brackets"] {
		log.Printf("[Proto] Detected angle bracket syntax, converting <> to {}...")
		for name, content := range fixedFiles {
			fixedFiles[name] = fixAngleBracketOptions(content)
		}
	}
	if fixes["enum_dedup"] {
		log.Printf("[Proto] Detected enum scoping conflict, deduplicating enum values...")
		for name, content := range fixedFiles {
			fixedFiles[name] = fixDuplicateEnumValues(content)
		}
	}

	compiled, err = c.compileFiles(fixedFiles)
	if err != nil {
		return fmt.Errorf("proto compile error (even after auto-fix): %w", err)
	}

	log.Printf("[Proto] Compilation succeeded after automatic source fixes")
	c.indexCompiled(compiled)
	return nil
}

// compileFiles does the actual protocompile compilation.
func (c *ProtoCompiler) compileFiles(files map[string]string) ([]protoreflect.FileDescriptor, error) {
	accessor := protocompile.SourceAccessorFromMap(files)

	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}

	sourceResolver := &protocompile.SourceResolver{
		Accessor: accessor,
	}
	compiler := protocompile.Compiler{
		Resolver: protocompile.WithStandardImports(sourceResolver),
	}

	compiled, err := compiler.Compile(context.Background(), names...)
	if err != nil {
		return nil, fmt.Errorf("proto compile error: %w", err)
	}

	// Convert to []protoreflect.FileDescriptor
	result := make([]protoreflect.FileDescriptor, len(compiled))
	for i, f := range compiled {
		result[i] = f
	}
	return result, nil
}

// indexCompiled indexes all message types from compiled file descriptors.
func (c *ProtoCompiler) indexCompiled(compiled []protoreflect.FileDescriptor) {
	descs := make(map[string]protoreflect.MessageDescriptor)
	for _, f := range compiled {
		if f == nil {
			continue
		}
		indexMessages(f.Messages(), descs)
	}

	c.mu.Lock()
	c.descriptors = descs
	c.mu.Unlock()
}

// fixDuplicateEnumValues processes proto content to make duplicate enum value names
// unique within their enclosing scope. Proto3 uses C++ scoping rules where enum values
// are scoped to the enclosing message, not the enum itself. So if two enums in the same
// message both have "UNKNOWN = 0", they conflict.
//
// This function detects such conflicts and prefixes duplicate values with the enum name.
// For example, if both MessageType.UNKNOWN and MessageStatus.UNKNOWN exist in ChatMessage,
// MessageStatus.UNKNOWN becomes MessageStatus.MS_UNKNOWN (using initials of the enum name).
//
// This transformation doesn't affect protobuf wire format decoding since only field numbers
// matter, not enum value names.
func fixDuplicateEnumValues(content string) string {
	// Parse all top-level and nested message blocks, fix enums within each scope
	return fixEnumsInScope(content)
}

// enumValueRegex matches enum value declarations like "UNKNOWN = 0;"
var enumValueRegex = regexp.MustCompile(`(?m)^\s*([A-Z][A-Z0-9_]*)\s*=\s*(-?\d+)\s*;`)

// enumBlockRegex matches enum blocks: "enum Name { ... }"
var enumBlockRegex = regexp.MustCompile(`(?ms)enum\s+(\w+)\s*\{([^}]*)\}`)

// fixEnumsInScope finds all enum blocks in the content and renames duplicate values.
func fixEnumsInScope(content string) string {
	// Strategy: find all enum blocks, collect value names, detect duplicates,
	// then rename duplicates by prefixing with a short form of the enum name.

	// We process the content in a scope-aware manner. For simplicity, we process
	// the entire file as one scope (which matches proto3 C++ scoping behavior where
	// enum values in nested enums still conflict with values in sibling enums
	// within the same message).

	// Step 1: Find all enum blocks and their value names
	type enumInfo struct {
		name       string
		matchStart int
		matchEnd   int
		body       string
		values     []string // value names in order
	}

	matches := enumBlockRegex.FindAllStringSubmatchIndex(content, -1)
	if len(matches) == 0 {
		return content
	}

	var enums []enumInfo
	for _, m := range matches {
		enumName := content[m[2]:m[3]]
		body := content[m[4]:m[5]]

		valueMatches := enumValueRegex.FindAllStringSubmatch(body, -1)
		var values []string
		for _, vm := range valueMatches {
			values = append(values, vm[1])
		}

		enums = append(enums, enumInfo{
			name:       enumName,
			matchStart: m[0],
			matchEnd:   m[1],
			body:       body,
			values:     values,
		})
	}

	// Step 2: Detect duplicate value names across enums
	// Count how many enums define each value name
	valueCounts := make(map[string]int)
	for _, e := range enums {
		for _, v := range e.values {
			valueCounts[v]++
		}
	}

	// Find which values are duplicated
	hasDuplicates := false
	for _, count := range valueCounts {
		if count > 1 {
			hasDuplicates = true
			break
		}
	}

	if !hasDuplicates {
		return content
	}

	// Step 3: Rename duplicates by prefixing with enum name abbreviation
	// Process in reverse order to preserve positions
	result := content
	for i := len(enums) - 1; i >= 0; i-- {
		e := enums[i]
		prefix := enumPrefix(e.name)

		needsFix := false
		for _, v := range e.values {
			if valueCounts[v] > 1 {
				needsFix = true
				break
			}
		}

		if !needsFix {
			continue
		}

		// Replace duplicate value names in this enum's body
		newBody := e.body
		for _, v := range e.values {
			if valueCounts[v] > 1 {
				newName := prefix + "_" + v
				// Replace the value declaration (careful to match whole word)
				re := regexp.MustCompile(`(?m)(^\s*)` + regexp.QuoteMeta(v) + `(\s*=)`)
				newBody = re.ReplaceAllString(newBody, "${1}"+newName+"${2}")
			}
		}

		if newBody != e.body {
			// Reconstruct the enum block
			oldBlock := content[e.matchStart:e.matchEnd]
			newBlock := strings.Replace(oldBlock, e.body, newBody, 1)
			result = result[:e.matchStart] + newBlock + result[e.matchEnd:]
		}
	}

	return result
}

// enumPrefix generates a short prefix from an enum name using uppercase letters.
// Examples: "MessageType" -> "MT", "GroupStatus" -> "GS", "Status" -> "S"
func enumPrefix(enumName string) string {
	var prefix strings.Builder
	for i, ch := range enumName {
		if i == 0 || (ch >= 'A' && ch <= 'Z') {
			prefix.WriteRune(ch)
		}
	}
	result := prefix.String()
	if result == "" {
		return strings.ToUpper(enumName[:1])
	}
	return result
}

// fixAngleBracketOptions converts text-format angle bracket syntax `<...>` to `{...}`
// in protobuf option values. The `<>` message literal syntax is valid in protobuf
// text-format (used extensively in Google's internal protos and many proto2 files),
// but the bufbuild/protocompile parser only accepts `{}`.
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
	// We use a character-by-character scanner that tracks context:
	// - inside string literal (skip)
	// - inside comment (skip)
	// - inside option value assignment after '=' (replace <> with {})

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
			// This is map<K,V> syntax; pass through until matching >
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

		// Check for option value context: `= <...>`
		// We look for '<' and check backwards (ignoring whitespace) for '='
		if ch == '<' && isOptionAngleBracket(runes, result) {
			// Replace < with { and find the matching > (handling nesting)
			result = append(result, '{')
			i++
			depth := 1
			for i < n && depth > 0 {
				c := runes[i]

				// Even inside option values, respect strings
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
// It looks backwards from position i for the keyword "map" (with optional whitespace).
func isMapType(runes []rune, i int) bool {
	// Walk backwards from i, skipping whitespace
	j := i - 1
	for j >= 0 && (runes[j] == ' ' || runes[j] == '\t') {
		j--
	}
	// Check for "map" (case-sensitive, as proto keywords are lowercase)
	if j >= 2 && runes[j-2] == 'm' && runes[j-1] == 'a' && runes[j] == 'p' {
		// Make sure 'map' is a whole word (not part of a longer identifier)
		if j-3 < 0 || !isIdentRune(runes[j-3]) {
			return true
		}
	}
	return false
}

// isOptionAngleBracket checks if a '<' (about to be appended) is in an option value context.
// It looks backwards through the already-built result for '=' (skipping whitespace).
// This catches patterns like:
//
//	= <...>
//	= < ...>
//	: <...>  (nested message field in text-format)
func isOptionAngleBracket(runes []rune, result []rune) bool {
	// Walk backwards through result, skipping whitespace
	j := len(result) - 1
	for j >= 0 && (result[j] == ' ' || result[j] == '\t' || result[j] == '\n' || result[j] == '\r') {
		j--
	}
	if j < 0 {
		return false
	}
	// '=' is the option value assignment operator
	// ':' is used in text-format for field values (nested messages)
	return result[j] == '=' || result[j] == ':'
}

// isIdentRune returns true if r is a valid identifier character.
func isIdentRune(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
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
