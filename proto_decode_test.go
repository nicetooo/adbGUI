package main

import (
	"encoding/binary"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

// testProtocPath returns the path to the protoc binary in bin/<platform>/protoc
// relative to the test file's directory.
func testProtocPath(t *testing.T) string {
	t.Helper()
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Dir(filename)

	var binName string
	switch runtime.GOOS {
	case "darwin":
		binName = "bin/darwin/protoc"
	case "linux":
		binName = "bin/linux/protoc"
	case "windows":
		binName = "bin/windows/protoc.exe"
	default:
		t.Skipf("unsupported OS: %s", runtime.GOOS)
	}

	protocPath := filepath.Join(projectRoot, binName)
	if _, err := os.Stat(protocPath); err != nil {
		t.Skipf("protoc binary not found at %s (run tests from project root)", protocPath)
	}
	return protocPath
}

// testProtoIncludePath returns the path to the protoc include directory.
func testProtoIncludePath(t *testing.T) string {
	t.Helper()
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Dir(filename)
	includePath := filepath.Join(projectRoot, "bin", "protoc-include")
	if _, err := os.Stat(includePath); err != nil {
		t.Skipf("protoc include directory not found at %s", includePath)
	}
	return includePath
}

// newTestCompiler creates a ProtoCompiler configured with the project's protoc binary.
func newTestCompiler(t *testing.T) *ProtoCompiler {
	t.Helper()
	c := NewProtoCompiler()
	c.SetPaths(testProtocPath(t), testProtoIncludePath(t))
	return c
}

// newTestRegistry creates a ProtoRegistry configured with the project's protoc binary.
func newTestRegistry(t *testing.T) *ProtoRegistry {
	t.Helper()
	reg := NewProtoRegistry()
	reg.compiler.SetPaths(testProtocPath(t), testProtoIncludePath(t))
	return reg
}

// TestAutoMatchDecode tests the auto-match scoring logic
func TestAutoMatchDecode(t *testing.T) {
	reg := newTestRegistry(t)
	decoder := NewProtobufDecoder(reg)

	// Add a simple proto file
	protoContent := `syntax = "proto3";
package test;
message UserResponse {
  int32 id = 1;
  string name = 2;
  string email = 3;
}
message Empty {}
`
	entry := &ProtoFileEntry{ID: "test-1", Name: "test.proto", Content: protoContent, LoadedAt: 1}
	if err := reg.AddFile(entry); err != nil {
		t.Fatalf("AddFile failed: %v", err)
	}

	// Create a protobuf-encoded UserResponse
	desc := reg.GetMessageDescriptor("test.UserResponse")
	if desc == nil {
		t.Fatal("UserResponse descriptor not found")
	}
	msg := dynamicpb.NewMessage(desc)
	msg.Set(desc.Fields().ByName("id"), protoreflect_valueOf(int32(42)))
	msg.Set(desc.Fields().ByName("name"), protoreflect_valueOf("Alice"))
	msg.Set(desc.Fields().ByName("email"), protoreflect_valueOf("alice@example.com"))

	raw, err := proto.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Auto-match should find UserResponse (not Empty)
	result, matchedType := decoder.autoMatchDecode(raw)
	if matchedType != "test.UserResponse" {
		t.Errorf("Expected auto-match to test.UserResponse, got %q", matchedType)
	}
	if result == "" {
		t.Error("Expected non-empty decode result")
	}

	t.Logf("Matched type: %s", matchedType)
	t.Logf("Decoded: %s", result)
}

// TestDirectionAwareMapping tests that FindMessageForURL respects direction
func TestDirectionAwareMapping(t *testing.T) {
	reg := newTestRegistry(t)

	reg.AddMapping(&ProtoMapping{
		ID:          "m1",
		URLPattern:  "*/api/users*",
		MessageType: "test.UserResponse",
		Direction:   "response",
	})
	reg.AddMapping(&ProtoMapping{
		ID:          "m2",
		URLPattern:  "*/api/users*",
		MessageType: "test.UserRequest",
		Direction:   "request",
	})
	reg.AddMapping(&ProtoMapping{
		ID:          "m3",
		URLPattern:  "*/api/both*",
		MessageType: "test.BothMessage",
		Direction:   "both",
	})

	tests := []struct {
		url       string
		direction string
		expected  string
	}{
		{"https://api.example.com/api/users/1", "response", "test.UserResponse"},
		{"https://api.example.com/api/users/1", "request", "test.UserRequest"},
		{"https://api.example.com/api/both/data", "response", "test.BothMessage"},
		{"https://api.example.com/api/both/data", "request", "test.BothMessage"},
		{"https://api.example.com/api/other", "response", ""},
	}

	for _, tt := range tests {
		result := reg.FindMessageForURL(tt.url, tt.direction)
		if result != tt.expected {
			t.Errorf("FindMessageForURL(%q, %q) = %q, want %q", tt.url, tt.direction, result, tt.expected)
		}
	}
}

// TestAutoCacheKey tests the cache key generation
func TestAutoCacheKey(t *testing.T) {
	tests := []struct {
		url       string
		direction string
		expected  string
	}{
		{"https://api.example.com/api/users?page=1", "response", "response:https://api.example.com/api/users"},
		{"https://api.example.com/api/users", "request", "request:https://api.example.com/api/users"},
		{"/simple/path", "response", "response:/simple/path"},
	}

	for _, tt := range tests {
		result := autoCacheKey(tt.url, tt.direction)
		if result != tt.expected {
			t.Errorf("autoCacheKey(%q, %q) = %q, want %q", tt.url, tt.direction, result, tt.expected)
		}
	}
}

// TestAutoMatchCaching tests that auto-match results are cached
func TestAutoMatchCaching(t *testing.T) {
	reg := newTestRegistry(t)
	decoder := NewProtobufDecoder(reg)

	protoContent := `syntax = "proto3";
package test;
message SimpleMsg {
  string value = 1;
}
`
	entry := &ProtoFileEntry{ID: "test-1", Name: "test.proto", Content: protoContent, LoadedAt: 1}
	if err := reg.AddFile(entry); err != nil {
		t.Fatalf("AddFile failed: %v", err)
	}

	desc := reg.GetMessageDescriptor("test.SimpleMsg")
	if desc == nil {
		t.Fatal("SimpleMsg descriptor not found")
	}
	msg := dynamicpb.NewMessage(desc)
	msg.Set(desc.Fields().ByName("value"), protoreflect_valueOf("hello"))

	raw, err := proto.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// First call: should auto-match and cache
	result1 := decoder.tryAutoMatch(raw, "https://example.com/api/test?q=1", "response")
	if result1 == "" {
		t.Error("Expected auto-match to succeed")
	}

	// Check cache populated
	decoder.autoCacheMu.RLock()
	cachedType, ok := decoder.autoCache["response:https://example.com/api/test"]
	decoder.autoCacheMu.RUnlock()
	if !ok {
		t.Error("Expected cache entry")
	}
	if cachedType != "test.SimpleMsg" {
		t.Errorf("Expected cached type test.SimpleMsg, got %q", cachedType)
	}

	// Second call with different query params: should use cache
	result2 := decoder.tryAutoMatch(raw, "https://example.com/api/test?q=2", "response")
	if result2 == "" {
		t.Error("Expected cached result")
	}

	// Clear cache
	decoder.ClearAutoCache()
	decoder.autoCacheMu.RLock()
	afterClear := len(decoder.autoCache)
	decoder.autoCacheMu.RUnlock()
	if afterClear != 0 {
		t.Errorf("Expected empty cache after clear, got %d entries", afterClear)
	}
}

// TestProtoc_DuplicateEnumValues tests that protoc handles duplicate enum value names
// (C++ scoping). The old protocompile library needed a workaround for this, but protoc
// handles it natively.
func TestProtoc_DuplicateEnumValues(t *testing.T) {
	compiler := newTestCompiler(t)

	// This proto has multiple enums with the same value name "UNKNOWN"
	// which follows proto3 C++ scoping rules
	files := map[string]string{
		"im_api.proto": `syntax = "proto3";
package im.api;

message ChatMessage {
  int64 msg_id = 1;
  string content = 2;
  MessageType type = 3;
  MessageStatus status = 4;

  enum MessageType {
    MT_UNKNOWN = 0;
    TEXT = 1;
    IMAGE = 2;
  }

  enum MessageStatus {
    MS_UNKNOWN = 0;
    SENDING = 1;
    SENT = 2;
  }
}

message GroupInfo {
  int64 group_id = 1;

  enum GroupType {
    GT_UNKNOWN = 0;
    NORMAL = 1;
  }
}`,
	}

	err := compiler.Compile(files)
	if err != nil {
		t.Fatalf("protoc should compile proto with properly-prefixed enum values, got: %v", err)
	}

	// Verify message types were indexed
	types := compiler.GetAllMessageTypes()
	typeMap := make(map[string]bool)
	for _, tp := range types {
		typeMap[tp] = true
	}

	if !typeMap["im.api.ChatMessage"] {
		t.Error("Expected im.api.ChatMessage to be compiled")
	}
	if !typeMap["im.api.GroupInfo"] {
		t.Error("Expected im.api.GroupInfo to be compiled")
	}

	// Verify descriptors work
	desc := compiler.GetMessageDescriptor("im.api.ChatMessage")
	if desc == nil {
		t.Fatal("ChatMessage descriptor not found")
	}
	if desc.Fields().Len() != 4 {
		t.Errorf("ChatMessage should have 4 fields, got %d", desc.Fields().Len())
	}
}

// TestProtoc_DuplicateEnumValues_Preprocessor tests that the fixDuplicateEnumValues
// preprocessor automatically renames colliding enum values so protoc can compile them.
// This matches real-world proto files (e.g. im_api.proto) where different enums in the
// same package have values like "UNKNOWN", "NotUsed", "UNAVAILABLE".
func TestProtoc_DuplicateEnumValues_Preprocessor(t *testing.T) {
	compiler := newTestCompiler(t)

	// Real-world scenario: multiple enums with colliding value names
	files := map[string]string{
		"im_api.proto": `syntax = "proto3";
package im_proto;

enum MessageType {
  UNKNOWN = 0;
  TEXT = 1;
  NotUsed = 99;
}

enum BizStatusMessageType {
  UNKNOWN = 0;
  NotUsed = 99;
  STATUS_UPDATE = 1;
}

enum FallbackStatus {
  UNKNOWN = 0;
  UNAVAILABLE = 1;
}

enum ConnectionStatus {
  UNAVAILABLE = 0;
  CONNECTED = 1;
}

message ChatMessage {
  int64 msg_id = 1;
  string content = 2;
  MessageType type = 3;
}`,
	}

	err := compiler.Compile(files)
	if err != nil {
		t.Fatalf("Preprocessor should fix duplicate enum values, got: %v", err)
	}

	desc := compiler.GetMessageDescriptor("im_proto.ChatMessage")
	if desc == nil {
		t.Fatal("ChatMessage descriptor not found")
	}
	if desc.Fields().Len() != 3 {
		t.Errorf("ChatMessage should have 3 fields, got %d", desc.Fields().Len())
	}
}

// TestFixDuplicateEnumValues tests the preprocessor function directly.
func TestFixDuplicateEnumValues(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantDups bool // should contain renamed values
	}{
		{
			name: "no duplicates - unchanged",
			input: `enum Foo {
  A = 0;
  B = 1;
}
enum Bar {
  C = 0;
  D = 1;
}`,
			wantDups: false,
		},
		{
			name: "duplicates renamed",
			input: `enum MessageType {
  UNKNOWN = 0;
  TEXT = 1;
}
enum Status {
  UNKNOWN = 0;
  ACTIVE = 1;
}`,
			wantDups: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fixDuplicateEnumValues(tt.input)
			if tt.wantDups {
				// UNKNOWN should be renamed in at least one enum
				if result == tt.input {
					t.Error("Expected duplicates to be renamed, but content unchanged")
				}
				// Should contain prefixed versions like MT_UNKNOWN or S_UNKNOWN
				if !strings.Contains(result, "_UNKNOWN") {
					t.Errorf("Expected prefixed UNKNOWN, got:\n%s", result)
				}
			} else {
				if result != tt.input {
					t.Errorf("Expected unchanged content, got:\n%s", result)
				}
			}
		})
	}
}

// TestProtoc_RealErrors tests that actual compile errors still propagate
func TestProtoc_RealErrors(t *testing.T) {
	compiler := newTestCompiler(t)

	files := map[string]string{
		"bad.proto": `syntax = "proto3";
package test;

message Broken {
  string name = 1;
  int32 name = 2;  // duplicate field name - real error
}`,
	}

	err := compiler.Compile(files)
	if err == nil {
		t.Error("Expected real compile error for duplicate field name")
	}
}

// TestProtoc_AngleBracketOptions tests that protoc handles angle bracket syntax natively.
// The old protocompile library needed fixAngleBracketOptions(), but protoc handles this.
func TestProtoc_AngleBracketOptions(t *testing.T) {
	compiler := newTestCompiler(t)

	// Proto2 file with angle bracket option syntax
	files := map[string]string{
		"test_options.proto": `syntax = "proto2";
package test.options;

import "google/protobuf/descriptor.proto";

extend google.protobuf.FieldOptions {
  optional MyFieldOption my_field_opt = 50001;
}

message MyFieldOption {
  optional string name = 1;
  optional int32 level = 2;
}

message TestMessage {
  optional string plain_field = 1;
  optional string annotated_field = 2 [(my_field_opt) = <name: "test", level: 3>];
  optional string nested_field = 3 [(my_field_opt) = <name: "nested">];
}
`,
	}

	err := compiler.Compile(files)
	if err != nil {
		t.Fatalf("protoc should handle angle bracket options natively, got: %v", err)
	}

	// Verify message types were indexed
	desc := compiler.GetMessageDescriptor("test.options.TestMessage")
	if desc == nil {
		t.Fatal("TestMessage descriptor not found")
	}
	if desc.Fields().Len() != 3 {
		t.Errorf("TestMessage should have 3 fields, got %d", desc.Fields().Len())
	}

	desc2 := compiler.GetMessageDescriptor("test.options.MyFieldOption")
	if desc2 == nil {
		t.Fatal("MyFieldOption descriptor not found")
	}
}

// TestProtoc_WellKnownTypes tests that well-known types (google/protobuf/*) are properly resolved.
func TestProtoc_WellKnownTypes(t *testing.T) {
	compiler := newTestCompiler(t)

	files := map[string]string{
		"with_timestamp.proto": `syntax = "proto3";
package test;

import "google/protobuf/timestamp.proto";
import "google/protobuf/any.proto";

message Event {
  string name = 1;
  google.protobuf.Timestamp created_at = 2;
  google.protobuf.Any payload = 3;
}
`,
	}

	err := compiler.Compile(files)
	if err != nil {
		t.Fatalf("protoc should resolve well-known types, got: %v", err)
	}

	desc := compiler.GetMessageDescriptor("test.Event")
	if desc == nil {
		t.Fatal("Event descriptor not found")
	}
	if desc.Fields().Len() != 3 {
		t.Errorf("Event should have 3 fields, got %d", desc.Fields().Len())
	}

	// Verify the timestamp field type
	tsField := desc.Fields().ByName("created_at")
	if tsField == nil {
		t.Fatal("created_at field not found")
	}
	if tsField.Message() == nil {
		t.Fatal("created_at should be a message field")
	}
	if string(tsField.Message().FullName()) != "google.protobuf.Timestamp" {
		t.Errorf("created_at type = %s, want google.protobuf.Timestamp", tsField.Message().FullName())
	}
}

// TestProtoc_MultipleFiles tests compiling multiple interdependent proto files.
func TestProtoc_MultipleFiles(t *testing.T) {
	compiler := newTestCompiler(t)

	files := map[string]string{
		"base.proto": `syntax = "proto3";
package test;

message BaseMessage {
  string id = 1;
  string name = 2;
}
`,
		"derived.proto": `syntax = "proto3";
package test;

import "base.proto";

message DerivedMessage {
  BaseMessage base = 1;
  int32 extra_field = 2;
}
`,
	}

	err := compiler.Compile(files)
	if err != nil {
		t.Fatalf("protoc should compile multiple interdependent files, got: %v", err)
	}

	// Verify both types exist
	base := compiler.GetMessageDescriptor("test.BaseMessage")
	if base == nil {
		t.Fatal("BaseMessage not found")
	}
	derived := compiler.GetMessageDescriptor("test.DerivedMessage")
	if derived == nil {
		t.Fatal("DerivedMessage not found")
	}
	if derived.Fields().Len() != 2 {
		t.Errorf("DerivedMessage should have 2 fields, got %d", derived.Fields().Len())
	}
}

// TestProtoc_EmptyFiles tests compiling with no files
func TestProtoc_EmptyFiles(t *testing.T) {
	compiler := newTestCompiler(t)

	err := compiler.Compile(map[string]string{})
	if err != nil {
		t.Fatalf("Compiling empty files should succeed, got: %v", err)
	}

	types := compiler.GetAllMessageTypes()
	if len(types) != 0 {
		t.Errorf("Expected no types, got %d", len(types))
	}
}

// TestProtoc_PartialMatch tests case-insensitive and short name matching
func TestProtoc_PartialMatch(t *testing.T) {
	compiler := newTestCompiler(t)

	files := map[string]string{
		"api.proto": `syntax = "proto3";
package myapp.v1;

message UserResponse {
  int32 id = 1;
  string name = 2;
}
`,
	}

	err := compiler.Compile(files)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// Exact match
	desc := compiler.GetMessageDescriptor("myapp.v1.UserResponse")
	if desc == nil {
		t.Error("Exact match should work")
	}

	// Short name match
	desc = compiler.GetMessageDescriptor("UserResponse")
	if desc == nil {
		t.Error("Short name match should work")
	}

	// Case-insensitive match
	desc = compiler.GetMessageDescriptor("myapp.v1.userresponse")
	if desc == nil {
		t.Error("Case-insensitive match should work")
	}
}

// TestProtoc_ErrorMessageFormatted tests that protoc errors are readable
func TestProtoc_ErrorMessageFormatted(t *testing.T) {
	compiler := newTestCompiler(t)

	files := map[string]string{
		"invalid.proto": `this is not valid proto syntax at all`,
	}

	err := compiler.Compile(files)
	if err == nil {
		t.Fatal("Expected compile error")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "protoc compile error") {
		t.Errorf("Error should contain 'protoc compile error', got: %s", errMsg)
	}
	t.Logf("Protoc error: %s", errMsg)
}

// --- Preprocessor tests ---

// TestFixAngleBracketOptions tests the <> to {} conversion in option values
func TestFixAngleBracketOptions(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "simple field option",
			input:  `optional string name = 1 [(my_opt) = <name: "test">];`,
			expect: `optional string name = 1 [(my_opt) = {name: "test"}];`,
		},
		{
			name:   "nested angle brackets",
			input:  `optional string name = 1 [(my_opt) = <foo: <bar: 1>>];`,
			expect: `optional string name = 1 [(my_opt) = {foo: {bar: 1}}];`,
		},
		{
			name:   "message option",
			input:  `option (my_opt) = <name: "hello", level: 3>;`,
			expect: `option (my_opt) = {name: "hello", level: 3};`,
		},
		{
			name:   "map type untouched",
			input:  `map<string, int32> tags = 5;`,
			expect: `map<string, int32> tags = 5;`,
		},
		{
			name:   "string literal with angle brackets untouched",
			input:  `string desc = 1 [default = "<hello>"];`,
			expect: `string desc = 1 [default = "<hello>"];`,
		},
		{
			name:   "comment with angle brackets untouched",
			input:  "// This uses <angle> brackets\nstring name = 1;",
			expect: "// This uses <angle> brackets\nstring name = 1;",
		},
		{
			name:   "colon-prefixed nested message",
			input:  `option (my_opt) = <field1: <nested_field: "val">>;`,
			expect: `option (my_opt) = {field1: {nested_field: "val"}};`,
		},
		{
			name:   "no angle brackets - passthrough",
			input:  `syntax = "proto3"; message Foo { string bar = 1; }`,
			expect: `syntax = "proto3"; message Foo { string bar = 1; }`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fixAngleBracketOptions(tt.input)
			if got != tt.expect {
				t.Errorf("fixAngleBracketOptions():\n  input:  %q\n  got:    %q\n  expect: %q", tt.input, got, tt.expect)
			}
		})
	}
}

// TestIsMapType tests the map<K,V> detection
func TestIsMapType(t *testing.T) {
	tests := []struct {
		input    string
		pos      int // position of '<'
		expected bool
	}{
		{"map<string, int32>", 3, true},
		{"map <string>", 4, true},
		{" map<K>", 4, true},
		{"something = <val>", 12, false},
		{"bitmap<x>", 6, false}, // 'map' is part of 'bitmap'
	}

	for _, tt := range tests {
		runes := []rune(tt.input)
		got := isMapType(runes, tt.pos)
		if got != tt.expected {
			t.Errorf("isMapType(%q, %d) = %v, want %v", tt.input, tt.pos, got, tt.expected)
		}
	}
}

// TestLooksLikeGRPCFrame tests the heuristic gRPC frame header detection.
func TestLooksLikeGRPCFrame(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		expect bool
	}{
		{
			name:   "valid uncompressed gRPC frame",
			data:   append([]byte{0x00, 0x00, 0x00, 0x00, 0x05}, []byte("hello")...),
			expect: true,
		},
		{
			name:   "valid compressed gRPC frame (flag=1)",
			data:   append([]byte{0x01, 0x00, 0x00, 0x00, 0x03}, []byte("abc")...),
			expect: true,
		},
		{
			name:   "invalid flag byte (2)",
			data:   append([]byte{0x02, 0x00, 0x00, 0x00, 0x03}, []byte("abc")...),
			expect: false,
		},
		{
			name:   "length mismatch (too short)",
			data:   append([]byte{0x00, 0x00, 0x00, 0x00, 0x0A}, []byte("short")...),
			expect: false,
		},
		{
			name:   "length mismatch (too long)",
			data:   append([]byte{0x00, 0x00, 0x00, 0x00, 0x01}, []byte("toolong")...),
			expect: false,
		},
		{
			name:   "too short (less than 6 bytes)",
			data:   []byte{0x00, 0x00, 0x00, 0x00, 0x00},
			expect: false,
		},
		{
			name:   "valid protobuf start (field 1 varint)",
			data:   []byte{0x08, 0x2A, 0x12, 0x05, 0x41, 0x6C, 0x69, 0x63, 0x65},
			expect: false, // byte[0]=0x08 > 1, not a gRPC frame
		},
		{
			name:   "empty data",
			data:   nil,
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := looksLikeGRPCFrame(tt.data)
			if got != tt.expect {
				t.Errorf("looksLikeGRPCFrame() = %v, want %v", got, tt.expect)
			}
		})
	}
}

// TestDecodeBody_GRPCAutoDetect tests that DecodeBody auto-detects gRPC framing
// when contentType is empty (e.g., WebSocket binary frames).
func TestDecodeBody_GRPCAutoDetect(t *testing.T) {
	// Create a simple protobuf schema
	compiler := newTestCompiler(t)
	registry := NewProtoRegistry()
	registry.compiler = compiler

	protoContent := `syntax = "proto3";
package test;
message Inner {
    string name = 1;
    int32 value = 2;
}
message Outer {
    int32 id = 1;
    Inner nested = 2;
}`
	err := compiler.Compile(map[string]string{"test.proto": protoContent})
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// Create an Outer message with nested Inner
	outerDesc := compiler.GetMessageDescriptor("test.Outer")
	if outerDesc == nil {
		t.Fatal("Could not find test.Outer descriptor")
	}
	innerDesc := compiler.GetMessageDescriptor("test.Inner")
	if innerDesc == nil {
		t.Fatal("Could not find test.Inner descriptor")
	}

	innerMsg := dynamicpb.NewMessage(innerDesc)
	innerMsg.Set(innerDesc.Fields().ByName("name"), protoreflect.ValueOfString("Alice"))
	innerMsg.Set(innerDesc.Fields().ByName("value"), protoreflect.ValueOfInt32(42))

	outerMsg := dynamicpb.NewMessage(outerDesc)
	outerMsg.Set(outerDesc.Fields().ByName("id"), protoreflect.ValueOfInt32(1))
	outerMsg.Set(outerDesc.Fields().ByName("nested"), protoreflect.ValueOfMessage(innerMsg))

	pbBytes, err := proto.Marshal(outerMsg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Wrap in gRPC frame header: [0x00][4-byte length][protobuf data]
	grpcFrame := make([]byte, 5+len(pbBytes))
	grpcFrame[0] = 0x00 // uncompressed
	binary.BigEndian.PutUint32(grpcFrame[1:5], uint32(len(pbBytes)))
	copy(grpcFrame[5:], pbBytes)

	// Add URL mapping
	registry.AddMapping(&ProtoMapping{
		ID:          "test-mapping",
		URLPattern:  "*/api/test*",
		MessageType: "test.Outer",
		Direction:   "both",
	})

	decoder := NewProtobufDecoder(registry)

	// Test 1: HTTP path (with content type) should decode correctly
	httpResult := decoder.DecodeBody(grpcFrame, "application/grpc", "https://example.com/api/test", "response")
	if httpResult == "" {
		t.Fatal("HTTP decode returned empty result")
	}
	var httpParsed map[string]interface{}
	if err := json.Unmarshal([]byte(httpResult), &httpParsed); err != nil {
		t.Fatalf("HTTP result is not valid JSON: %v\nResult: %s", err, httpResult)
	}
	// Verify nested field is decoded
	if nested, ok := httpParsed["nested"].(map[string]interface{}); !ok {
		t.Errorf("HTTP: 'nested' field not decoded as object. Full result: %s", httpResult)
	} else {
		if name, _ := nested["name"].(string); name != "Alice" {
			t.Errorf("HTTP: nested.name = %q, want 'Alice'", name)
		}
	}

	// Test 2: WS path (empty content type) should also decode correctly via auto-detection
	wsResult := decoder.DecodeBody(grpcFrame, "", "https://example.com/api/test", "response")
	if wsResult == "" {
		t.Fatal("WS decode returned empty result")
	}
	var wsParsed map[string]interface{}
	if err := json.Unmarshal([]byte(wsResult), &wsParsed); err != nil {
		t.Fatalf("WS result is not valid JSON: %v\nResult: %s", err, wsResult)
	}
	// Verify nested field is decoded identically
	if nested, ok := wsParsed["nested"].(map[string]interface{}); !ok {
		t.Errorf("WS: 'nested' field not decoded as object. Full result: %s", wsResult)
	} else {
		if name, _ := nested["name"].(string); name != "Alice" {
			t.Errorf("WS: nested.name = %q, want 'Alice'", name)
		}
	}

	// Test 3: Plain protobuf without gRPC frame (should still work for both)
	httpPlain := decoder.DecodeBody(pbBytes, "application/x-protobuf", "https://example.com/api/test", "response")
	wsPlain := decoder.DecodeBody(pbBytes, "", "https://example.com/api/test", "response")
	if httpPlain == "" {
		t.Error("HTTP plain protobuf decode returned empty")
	}
	if wsPlain == "" {
		t.Error("WS plain protobuf decode returned empty")
	}

	t.Logf("HTTP (gRPC frame): %s", httpResult)
	t.Logf("WS   (gRPC frame): %s", wsResult)
	t.Logf("HTTP (plain pb):   %s", httpPlain)
	t.Logf("WS   (plain pb):   %s", wsPlain)
}

// helper to create protoreflect.Value from Go types
func protoreflect_valueOf(v interface{}) protoreflect.Value {
	switch val := v.(type) {
	case int32:
		return protoreflect.ValueOfInt32(val)
	case string:
		return protoreflect.ValueOfString(val)
	default:
		panic("unsupported type")
	}
}
