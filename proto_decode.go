package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strings"
	"sync"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

// ProtobufDecoder handles decoding protobuf binary data with optional schema support.
// When no explicit URL→message mapping exists, it automatically tries all registered
// message types and picks the best match using a scoring heuristic.
type ProtobufDecoder struct {
	registry    *ProtoRegistry
	autoCache   map[string]string // cacheKey (direction:urlPath) -> matched message type ("" = negative cache)
	autoCacheMu sync.RWMutex
}

// NewProtobufDecoder creates a decoder linked to the given registry.
func NewProtobufDecoder(registry *ProtoRegistry) *ProtobufDecoder {
	return &ProtobufDecoder{
		registry:  registry,
		autoCache: make(map[string]string),
	}
}

// ClearAutoCache resets the auto-match cache. Call after proto files are added/removed/recompiled.
func (d *ProtobufDecoder) ClearAutoCache() {
	d.autoCacheMu.Lock()
	d.autoCache = make(map[string]string)
	d.autoCacheMu.Unlock()
}

// DecodeBody attempts to decode protobuf binary data.
// direction should be "request" or "response" to enable direction-aware mapping.
// It tries: (1) explicit URL mapping → (2) auto-match cache → (3) auto-match all types → (4) raw decode.
// Returns a JSON string representation.
func (d *ProtobufDecoder) DecodeBody(data []byte, contentType, url, direction string) string {
	if len(data) == 0 {
		return ""
	}

	// Strip gRPC frame header (5 bytes: 1 byte compressed flag + 4 bytes length)
	raw := data
	if isGRPCContentType(contentType) && len(data) >= 5 {
		raw = data[5:]
	}

	// 1. Try explicit URL mapping (direction-aware)
	if d.registry != nil {
		msgType := d.registry.FindMessageForURL(url, direction)
		if msgType != "" {
			if result := d.decodeWithType(raw, msgType); result != "" {
				return result
			}
		}
	}

	// 2. Try auto-match (cached or fresh)
	if result := d.tryAutoMatch(raw, url, direction); result != "" {
		return result
	}

	// 3. Fallback: raw decode (no schema)
	result := rawDecodeProtobuf(raw)
	if result == nil {
		return fmt.Sprintf("[Protobuf: %d bytes, decode failed]", len(data))
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Sprintf("[Protobuf: %d bytes]", len(data))
	}
	return string(jsonBytes)
}

// decodeWithType decodes raw protobuf bytes using a specific message type name.
func (d *ProtobufDecoder) decodeWithType(raw []byte, msgType string) string {
	desc := d.registry.GetMessageDescriptor(msgType)
	if desc == nil {
		return ""
	}
	msg := dynamicpb.NewMessage(desc)
	if err := proto.Unmarshal(raw, msg); err != nil {
		return ""
	}
	opts := protojson.MarshalOptions{
		Multiline:       true,
		Indent:          "  ",
		UseProtoNames:   true,
		EmitUnpopulated: true,
	}
	jsonBytes, err := opts.Marshal(msg)
	if err != nil {
		return ""
	}
	return string(jsonBytes)
}

// autoCacheKey builds the cache key from direction + URL path (without query string).
func autoCacheKey(url, direction string) string {
	path := url
	if idx := strings.IndexByte(path, '?'); idx != -1 {
		path = path[:idx]
	}
	return direction + ":" + path
}

// tryAutoMatch attempts auto-matching using cache or brute-force scoring.
func (d *ProtobufDecoder) tryAutoMatch(raw []byte, url, direction string) string {
	if d.registry == nil || d.registry.compiler == nil {
		return ""
	}

	key := autoCacheKey(url, direction)

	// Check cache first
	d.autoCacheMu.RLock()
	cachedType, cached := d.autoCache[key]
	d.autoCacheMu.RUnlock()

	if cached {
		if cachedType == "" {
			return "" // negative cache — previously tried and no match
		}
		return d.decodeWithType(raw, cachedType)
	}

	// Try all known message types and score
	result, matchedType := d.autoMatchDecode(raw)

	// Cache the result (positive or negative)
	d.autoCacheMu.Lock()
	if len(d.autoCache) > 2000 {
		// Prevent unbounded growth: clear on overflow
		d.autoCache = make(map[string]string)
	}
	d.autoCache[key] = matchedType // "" if no match (negative cache)
	d.autoCacheMu.Unlock()

	if result != "" {
		log.Printf("[Proto] Auto-matched %s → %s", key, matchedType)
	}
	return result
}

// autoMatchDecode tries all compiled message types against the raw data and returns
// the best-scoring decode result. Returns ("", "") if no good match found.
func (d *ProtobufDecoder) autoMatchDecode(raw []byte) (string, string) {
	if len(raw) == 0 {
		return "", ""
	}

	allDescs := d.registry.compiler.GetAllDescriptors()
	if len(allDescs) == 0 {
		return "", ""
	}

	type candidate struct {
		json  string
		name  string
		score float64
	}
	var best candidate

	for name, desc := range allDescs {
		msg := dynamicpb.NewMessage(desc)
		if err := proto.Unmarshal(raw, msg); err != nil {
			continue
		}

		// Score: penalize unknown bytes, reward populated known fields
		unknownLen := len(msg.GetUnknown())
		if unknownLen > 0 && float64(unknownLen)/float64(len(raw)) > 0.5 {
			continue // More than half is unknown fields — likely wrong type
		}

		// Count populated known fields
		knownFields := 0
		msg.Range(func(_ protoreflect.FieldDescriptor, _ protoreflect.Value) bool {
			knownFields++
			return true
		})
		if knownFields == 0 {
			continue // No known fields populated — useless match
		}

		// Scoring: 70% byte coverage (fewer unknown = better), 30% field coverage
		unknownRatio := float64(unknownLen) / float64(len(raw))
		byteScore := 1.0 - unknownRatio
		totalFields := desc.Fields().Len()
		fieldScore := 0.0
		if totalFields > 0 {
			fieldScore = float64(knownFields) / float64(totalFields)
		}
		score := byteScore*0.7 + fieldScore*0.3

		if score > best.score {
			opts := protojson.MarshalOptions{
				Multiline:       true,
				Indent:          "  ",
				UseProtoNames:   true,
				EmitUnpopulated: true,
			}
			if jsonBytes, err := opts.Marshal(msg); err == nil {
				best = candidate{json: string(jsonBytes), name: name, score: score}
			}
		}
	}

	// Minimum quality threshold
	if best.score < 0.3 {
		return "", ""
	}

	return best.json, best.name
}

// isProtobufContentType checks if the content type indicates protobuf.
func isProtobufContentType(ct string) bool {
	ct = strings.ToLower(ct)
	return strings.Contains(ct, "application/x-protobuf") ||
		strings.Contains(ct, "application/protobuf") ||
		strings.Contains(ct, "application/grpc") ||
		strings.Contains(ct, "application/grpc+proto") ||
		strings.Contains(ct, "application/grpc-web+proto") ||
		strings.Contains(ct, "application/vnd.google.protobuf")
}

// isGRPCContentType checks if the content type is gRPC (has frame header).
func isGRPCContentType(ct string) bool {
	ct = strings.ToLower(ct)
	return strings.Contains(ct, "application/grpc")
}

// --- Raw protobuf decode (no schema) ---

// rawDecodeProtobuf decodes protobuf wire format without schema.
// Returns a map with field numbers as keys.
func rawDecodeProtobuf(data []byte) map[string]interface{} {
	if len(data) == 0 {
		return nil
	}

	fields := make(map[string][]interface{})
	b := data

	for len(b) > 0 {
		num, wtype, n := protowire.ConsumeTag(b)
		if n < 0 {
			return nil // invalid tag
		}
		b = b[n:]

		key := fmt.Sprintf("%d", num)
		var val interface{}

		switch wtype {
		case protowire.VarintType:
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return nil
			}
			b = b[n:]
			val = decodeVarint(v)

		case protowire.Fixed64Type:
			v, n := protowire.ConsumeFixed64(b)
			if n < 0 {
				return nil
			}
			b = b[n:]
			// Could be double, fixed64, sfixed64. Try double first.
			f := math.Float64frombits(v)
			if !math.IsNaN(f) && !math.IsInf(f, 0) && f != 0 && (math.Abs(f) > 1e-10 && math.Abs(f) < 1e15) {
				val = f
			} else {
				val = v
			}

		case protowire.Fixed32Type:
			v, n := protowire.ConsumeFixed32(b)
			if n < 0 {
				return nil
			}
			b = b[n:]
			f := math.Float32frombits(v)
			if !math.IsNaN(float64(f)) && !math.IsInf(float64(f), 0) && f != 0 && (math.Abs(float64(f)) > 1e-6 && math.Abs(float64(f)) < 1e10) {
				val = f
			} else {
				val = v
			}

		case protowire.BytesType:
			v, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return nil
			}
			b = b[n:]
			val = decodeBytesField(v)

		case protowire.StartGroupType:
			// Groups are deprecated, skip to end group
			_, n := protowire.ConsumeGroup(num, b)
			if n < 0 {
				return nil
			}
			b = b[n:]
			continue

		default:
			return nil // unknown wire type
		}

		fields[key] = append(fields[key], val)
	}

	// Convert to result: use single value if only one, array if multiple (repeated field)
	result := make(map[string]interface{}, len(fields))
	for k, vs := range fields {
		if len(vs) == 1 {
			result[k] = vs[0]
		} else {
			result[k] = vs
		}
	}

	return result
}

// decodeVarint interprets a varint value.
// Returns int64 for small values, uint64 for large.
func decodeVarint(v uint64) interface{} {
	// If it fits in int64 range and looks like a reasonable signed value
	if v <= math.MaxInt64 {
		return int64(v)
	}
	// Could be a negative zigzag-encoded value
	signed := protowire.DecodeZigZag(v)
	if signed < 0 && signed > -1000000000 {
		return signed
	}
	return v
}

// decodeBytesField tries to interpret a bytes field as: sub-message > UTF-8 string > base64-hint.
func decodeBytesField(data []byte) interface{} {
	if len(data) == 0 {
		return ""
	}

	// Try as sub-message first
	sub := rawDecodeProtobuf(data)
	if sub != nil && len(sub) > 0 {
		// Verify it consumed enough of the data to be plausible
		return sub
	}

	// Try as UTF-8 string
	if isValidUTF8String(data) {
		return string(data)
	}

	// Binary data
	return fmt.Sprintf("[bytes: %d]", len(data))
}

// isValidUTF8String checks if data is likely a printable UTF-8 string.
func isValidUTF8String(data []byte) bool {
	if len(data) == 0 {
		return true
	}
	nullCount := 0
	for _, b := range data {
		if b == 0 {
			nullCount++
		}
	}
	// Too many null bytes means likely binary
	if nullCount > 0 {
		return false
	}
	// Check it's valid UTF-8 by looking for control chars (except common ones)
	for _, b := range data {
		if b < 0x20 && b != '\n' && b != '\r' && b != '\t' {
			return false
		}
	}
	return true
}

// --- Proto Registry (schema management) ---

// ProtoFileEntry represents a loaded .proto file.
type ProtoFileEntry struct {
	ID       string `json:"id"`
	Name     string `json:"name"`     // filename
	Content  string `json:"content"`  // .proto file text
	LoadedAt int64  `json:"loadedAt"` // unix ms
}

// ProtoMapping maps a URL pattern to a protobuf message type.
type ProtoMapping struct {
	ID          string `json:"id"`
	URLPattern  string `json:"urlPattern"`  // wildcard pattern
	MessageType string `json:"messageType"` // e.g. "user.UserResponse"
	Direction   string `json:"direction"`   // "request", "response", "both"
	Description string `json:"description"`
}

// ProtoRegistry manages .proto files and URL→message mappings.
type ProtoRegistry struct {
	mu       sync.RWMutex
	files    map[string]*ProtoFileEntry
	mappings map[string]*ProtoMapping
	// compiled descriptors are managed by proto_compile.go
	compiler *ProtoCompiler
}

// NewProtoRegistry creates a new empty registry.
func NewProtoRegistry() *ProtoRegistry {
	return &ProtoRegistry{
		files:    make(map[string]*ProtoFileEntry),
		mappings: make(map[string]*ProtoMapping),
		compiler: NewProtoCompiler(),
	}
}

// AddFile adds or replaces a .proto file and recompiles.
func (r *ProtoRegistry) AddFile(entry *ProtoFileEntry) error {
	r.mu.Lock()
	r.files[entry.ID] = entry
	r.mu.Unlock()

	return r.recompile()
}

// RemoveFile removes a .proto file and recompiles.
func (r *ProtoRegistry) RemoveFile(id string) error {
	r.mu.Lock()
	delete(r.files, id)
	r.mu.Unlock()

	return r.recompile()
}

// GetFiles returns all loaded proto files.
func (r *ProtoRegistry) GetFiles() []*ProtoFileEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*ProtoFileEntry, 0, len(r.files))
	for _, f := range r.files {
		result = append(result, f)
	}
	return result
}

// AddMapping adds a URL→message mapping.
func (r *ProtoRegistry) AddMapping(m *ProtoMapping) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.mappings[m.ID] = m
}

// RemoveMapping removes a mapping.
func (r *ProtoRegistry) RemoveMapping(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.mappings, id)
}

// GetMappings returns all mappings.
func (r *ProtoRegistry) GetMappings() []*ProtoMapping {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*ProtoMapping, 0, len(r.mappings))
	for _, m := range r.mappings {
		result = append(result, m)
	}
	return result
}

// FindMessageForURL looks up the message type for a given URL and direction.
// direction should be "request" or "response". Mappings with Direction "both" always match.
func (r *ProtoRegistry) FindMessageForURL(url, direction string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, m := range r.mappings {
		if matchWildcard(m.URLPattern, url) {
			// Check direction: "both" matches everything, otherwise must match exactly
			if m.Direction == "both" || m.Direction == direction || m.Direction == "" {
				return m.MessageType
			}
		}
	}
	return ""
}

// GetMessageDescriptor returns the compiled message descriptor for a fully-qualified type name.
func (r *ProtoRegistry) GetMessageDescriptor(fullName string) protoreflectMessageDescriptor {
	if r.compiler == nil {
		return nil
	}
	return r.compiler.GetMessageDescriptor(fullName)
}

// recompile recompiles all .proto files in the registry.
func (r *ProtoRegistry) recompile() error {
	r.mu.RLock()
	fileMap := make(map[string]string, len(r.files))
	for _, f := range r.files {
		fileMap[f.Name] = f.Content
	}
	r.mu.RUnlock()

	return r.compiler.Compile(fileMap)
}

// GetAvailableMessageTypes returns all message type names from compiled protos.
func (r *ProtoRegistry) GetAvailableMessageTypes() []string {
	if r.compiler == nil {
		return nil
	}
	return r.compiler.GetAllMessageTypes()
}

// matchWildcard performs simple wildcard matching using proxy.MatchPattern.
func matchWildcard(pattern, url string) bool {
	// Simple wildcard: * matches any sequence of characters
	return wildcardMatch(pattern, url)
}

// wildcardMatch is a simple wildcard matcher supporting *.
func wildcardMatch(pattern, s string) bool {
	if pattern == "*" {
		return true
	}
	// Split pattern by * and match sequentially
	parts := strings.Split(pattern, "*")
	if len(parts) == 1 {
		return pattern == s
	}

	pos := 0
	for i, part := range parts {
		if part == "" {
			continue
		}
		idx := strings.Index(s[pos:], part)
		if idx < 0 {
			return false
		}
		if i == 0 && idx != 0 {
			// First part must match at start if pattern doesn't start with *
			return false
		}
		pos += idx + len(part)
	}
	// If pattern doesn't end with *, the string must end at pos
	if !strings.HasSuffix(pattern, "*") && pos != len(s) {
		return false
	}
	return true
}
