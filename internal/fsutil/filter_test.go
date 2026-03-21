package fsutil

import (
	"testing"
)

func TestShouldIncludeDir(t *testing.T) {
	f := DefaultFilter()

	tests := []struct {
		name string
		want bool
	}{
		{"src", true},
		{"internal", true},
		{"node_modules", false},
		{".git", false},
		{"vendor", false},
		{"dist", false},
		{".vscode", false},
		{".idea", false},
		{"build", false},
	}

	for _, tt := range tests {
		got := f.ShouldIncludeDir(tt.name)
		if got != tt.want {
			t.Errorf("ShouldIncludeDir(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestShouldIncludeFile(t *testing.T) {
	f := DefaultFilter()

	tests := []struct {
		path string
		size int64
		want bool
	}{
		{"main.go", 1000, true},
		{"README.md", 500, true},
		{"config.yaml", 200, true},
		{"image.png", 5000, false},
		{"bundle.min.js", 1000, false},
		{"package-lock.json", 10000, false},
		{"go.sum", 5000, false},
		{"main.go", 2 * 1024 * 1024, false}, // over size limit
		{"binary.exe", 100, false},
	}

	for _, tt := range tests {
		got := f.ShouldIncludeFile(tt.path, tt.size)
		if got != tt.want {
			t.Errorf("ShouldIncludeFile(%q, %d) = %v, want %v", tt.path, tt.size, got, tt.want)
		}
	}
}

func TestDetectKind(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"main.go", "source"},
		{"app.ts", "source"},
		{"utils.py", "source"},
		{"README.md", "markdown"},
		{"notes.txt", "text"},
		{"config.yaml", "config"},
		{"docker-compose.yml", "config"},
		{"Makefile", "config"},
		{"image.png", "other"},
	}

	for _, tt := range tests {
		got := DetectKind(tt.path)
		if got != tt.want {
			t.Errorf("DetectKind(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}
