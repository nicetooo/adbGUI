package main

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

// TestAutoMatchDecode tests the auto-match scoring logic
func TestAutoMatchDecode(t *testing.T) {
	reg := NewProtoRegistry()
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
	reg := NewProtoRegistry()

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
	reg := NewProtoRegistry()
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

// TestLenientCompiler_DuplicateEnumValues tests that the lenient compiler
// handles proto files with duplicate enum value names across enums (C++ scoping conflict)
func TestLenientCompiler_DuplicateEnumValues(t *testing.T) {
	compiler := NewProtoCompiler()

	// This proto has multiple enums with the same value name "UNKNOWN"
	// which violates proto3 C++ scoping rules but is common in production protos
	files := map[string]string{
		"im_api.proto": `syntax = "proto3";
package im.api;

message ChatMessage {
  int64 msg_id = 1;
  string content = 2;
  MessageType type = 3;
  MessageStatus status = 4;

  enum MessageType {
    UNKNOWN = 0;
    TEXT = 1;
    IMAGE = 2;
  }

  enum MessageStatus {
    UNKNOWN = 0;
    SENDING = 1;
    SENT = 2;
  }
}

message GroupInfo {
  int64 group_id = 1;

  enum GroupType {
    UNKNOWN = 0;
    NORMAL = 1;
  }
}`,
	}

	err := compiler.Compile(files)
	if err != nil {
		t.Fatalf("Lenient compiler should succeed with duplicate enum values, got: %v", err)
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

// TestLenientCompiler_RealErrors tests that actual compile errors still propagate
func TestLenientCompiler_RealErrors(t *testing.T) {
	compiler := NewProtoCompiler()

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

// TestEnumPrefix tests the enum name prefix generation
func TestEnumPrefix(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"MessageType", "MT"},
		{"MessageStatus", "MS"},
		{"GroupType", "GT"},
		{"GroupStatus", "GS"},
		{"NotificationType", "NT"},
		{"Status", "S"},
		{"ALLCAPS", "ALLCAPS"}, // all upper = include all
		{"type", "t"},          // all lower = first char
	}

	for _, tt := range tests {
		got := enumPrefix(tt.input)
		if got != tt.want {
			t.Errorf("enumPrefix(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// TestFixDuplicateEnumValues tests the enum value deduplication
func TestFixDuplicateEnumValues(t *testing.T) {
	input := `syntax = "proto3";
package test;

message Msg {
  enum TypeA {
    UNKNOWN = 0;
    FOO = 1;
  }
  enum TypeB {
    UNKNOWN = 0;
    BAR = 1;
  }
}
`
	result := fixDuplicateEnumValues(input)

	// UNKNOWN should be renamed in both enums
	if strings.Contains(result, "\n    UNKNOWN = 0;") {
		// At least one should be renamed
		count := strings.Count(result, "UNKNOWN = 0")
		if count > 1 {
			t.Errorf("Expected at most 1 bare UNKNOWN remaining, got %d.\nResult:\n%s", count, result)
		}
	}

	// Should still have the enum structure
	if !strings.Contains(result, "enum TypeA") || !strings.Contains(result, "enum TypeB") {
		t.Error("Enum blocks should be preserved")
	}

	// Non-duplicate values should be unchanged
	if !strings.Contains(result, "FOO = 1") {
		t.Error("Non-duplicate value FOO should be unchanged")
	}
	if !strings.Contains(result, "BAR = 1") {
		t.Error("Non-duplicate value BAR should be unchanged")
	}
}

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
			name:   "map type with nested message untouched",
			input:  `map<string, SomeMessage> msgs = 6;`,
			expect: `map<string, SomeMessage> msgs = 6;`,
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
			name:   "multiple options on separate lines",
			input:  "  optional string a = 1 [(opt1) = <x: 1>];\n  optional string b = 2 [(opt2) = <y: 2>];",
			expect: "  optional string a = 1 [(opt1) = {x: 1}];\n  optional string b = 2 [(opt2) = {y: 2}];",
		},
		{
			name:   "no angle brackets - passthrough",
			input:  `syntax = "proto3"; message Foo { string bar = 1; }`,
			expect: `syntax = "proto3"; message Foo { string bar = 1; }`,
		},
		{
			name:   "string inside angle bracket option",
			input:  `optional string f = 1 [(opt) = <name: "has <inner> angle">];`,
			expect: `optional string f = 1 [(opt) = {name: "has <inner> angle"}];`,
		},
		{
			name:   "block comment with angle brackets untouched",
			input:  "/* <angle> brackets */\nstring name = 1;",
			expect: "/* <angle> brackets */\nstring name = 1;",
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

// TestCompiler_AngleBracketOptions tests that the compiler auto-fixes angle bracket syntax
func TestCompiler_AngleBracketOptions(t *testing.T) {
	compiler := NewProtoCompiler()

	// Proto2 file with angle bracket option syntax - this is what fails without the fix
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
		t.Fatalf("Compiler should handle angle bracket options, got: %v", err)
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
