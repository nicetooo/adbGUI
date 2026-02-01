package main

import (
	"strings"
	"testing"
)

func TestParseProtoPackage(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "google.api package",
			content: "syntax = \"proto3\";\npackage google.api;\nimport \"google/api/http.proto\";\n",
			want:    "google.api",
		},
		{
			name:    "grpc.health.v1 package",
			content: "syntax = \"proto3\";\n\npackage grpc.health.v1;\n",
			want:    "grpc.health.v1",
		},
		{
			name:    "no package",
			content: "syntax = \"proto3\";\nmessage Foo {}\n",
			want:    "",
		},
		{
			name:    "package with extra spaces",
			content: "package   my.package  ;\n",
			want:    "my.package",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseProtoPackage(tt.content)
			if got != tt.want {
				t.Errorf("parseProtoPackage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInferProtoRoot(t *testing.T) {
	tests := []struct {
		name    string
		fileURL string
		content string
		want    string
	}{
		{
			name:    "googleapis annotations",
			fileURL: "https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto",
			content: "syntax = \"proto3\";\npackage google.api;\n",
			want:    "https://raw.githubusercontent.com/googleapis/googleapis/master",
		},
		{
			name:    "grpc health",
			fileURL: "https://raw.githubusercontent.com/grpc/grpc-proto/master/grpc/health/v1/health.proto",
			content: "syntax = \"proto3\";\npackage grpc.health.v1;\n",
			want:    "https://raw.githubusercontent.com/grpc/grpc-proto/master",
		},
		{
			name:    "no package declaration",
			fileURL: "https://example.com/proto/service.proto",
			content: "syntax = \"proto3\";\nmessage Foo {}\n",
			want:    "",
		},
		{
			name:    "package doesn't match URL path",
			fileURL: "https://example.com/proto/my_service.proto",
			content: "syntax = \"proto3\";\npackage some.other.pkg;\n",
			want:    "", // can't infer because URL path doesn't end with some/other/pkg/my_service.proto
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inferProtoRoot(tt.fileURL, tt.content)
			if got != tt.want {
				t.Errorf("inferProtoRoot() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildImportURLCandidates(t *testing.T) {
	tests := []struct {
		name          string
		protoRoot     string
		parentFileURL string
		importPath    string
		wantFirst     string // The first (highest priority) candidate should be correct
		wantContains  string // At minimum, this URL must be in candidates
	}{
		{
			name:          "with proto root - same directory import",
			protoRoot:     "https://raw.githubusercontent.com/googleapis/googleapis/master",
			parentFileURL: "https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto",
			importPath:    "google/api/http.proto",
			wantFirst:     "https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto",
		},
		{
			name:          "with proto root - cross directory import",
			protoRoot:     "https://raw.githubusercontent.com/googleapis/googleapis/master",
			parentFileURL: "https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto",
			importPath:    "google/type/date.proto",
			wantFirst:     "https://raw.githubusercontent.com/googleapis/googleapis/master/google/type/date.proto",
		},
		{
			name:          "without proto root - overlap detection",
			protoRoot:     "",
			parentFileURL: "https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto",
			importPath:    "google/api/http.proto",
			wantContains:  "https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto",
		},
		{
			name:          "without proto root - cross directory overlap",
			protoRoot:     "",
			parentFileURL: "https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto",
			importPath:    "google/type/date.proto",
			wantContains:  "https://raw.githubusercontent.com/googleapis/googleapis/master/google/type/date.proto",
		},
		{
			name:          "without proto root - no overlap (subdirectory import)",
			protoRoot:     "",
			parentFileURL: "https://example.com/proto/service.proto",
			importPath:    "common/types.proto",
			wantFirst:     "https://example.com/proto/common/types.proto",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidates := buildImportURLCandidates(tt.protoRoot, tt.parentFileURL, tt.importPath)

			if len(candidates) == 0 {
				t.Fatal("buildImportURLCandidates() returned no candidates")
			}

			if tt.wantFirst != "" && candidates[0] != tt.wantFirst {
				t.Errorf("First candidate = %q, want %q\nAll candidates: %v", candidates[0], tt.wantFirst, candidates)
			}

			if tt.wantContains != "" {
				found := false
				for _, c := range candidates {
					if c == tt.wantContains {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Candidates don't contain %q\nAll candidates: %v", tt.wantContains, candidates)
				}
			}

			// Verify no candidate has doubled paths (the original bug)
			for _, c := range candidates {
				if strings.Contains(c, "/google/api/google/api/") ||
					strings.Contains(c, "/google/type/google/type/") {
					t.Errorf("Candidate has doubled path: %s", c)
				}
			}
		})
	}
}

func TestBuildImportURLCandidates_NeverDoublesPath(t *testing.T) {
	// This is the exact scenario that was broken before the fix
	protoRoot := "https://raw.githubusercontent.com/googleapis/googleapis/master"
	parentURL := "https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto"

	imports := []string{
		"google/api/http.proto",
		"google/protobuf/descriptor.proto", // this would be skipped in practice, but test anyway
		"google/type/date.proto",
		"google/rpc/status.proto",
	}

	for _, imp := range imports {
		candidates := buildImportURLCandidates(protoRoot, parentURL, imp)
		for _, c := range candidates {
			// The import path should appear exactly once in the URL
			importDir := imp[:strings.LastIndex(imp, "/")]
			count := strings.Count(c, "/"+importDir+"/")
			if count > 1 {
				t.Errorf("Import %q: path doubled in candidate URL %q", imp, c)
			}
		}
	}
}

func TestFetchProtoURL_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Test downloading a small, stable proto file from GitHub
	content, err := fetchProtoURL("https://raw.githubusercontent.com/grpc/grpc-proto/master/grpc/health/v1/health.proto")
	if err != nil {
		t.Fatalf("fetchProtoURL() error: %v", err)
	}

	if !strings.Contains(content, "package grpc.health.v1") {
		t.Error("Downloaded content doesn't contain expected package declaration")
	}
	if !strings.Contains(content, "HealthCheckRequest") {
		t.Error("Downloaded content doesn't contain expected message")
	}
}

func TestExtractProtoFileName(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://example.com/path/service.proto", "service.proto"},
		{"https://example.com/path/service.proto?token=abc", "service.proto"},
		{"https://example.com/path/service", "service.proto"},
		{"https://raw.githubusercontent.com/user/repo/main/google/api/annotations.proto", "annotations.proto"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got := extractProtoFileName(tt.url)
			if got != tt.want {
				t.Errorf("extractProtoFileName(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}
