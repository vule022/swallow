package fsutil

import (
	"path/filepath"
	"strings"
)

// Filter controls which files are included during directory walks.
type Filter struct {
	IgnoreDirs     []string
	IgnoreExts     []string
	IgnorePatterns []string
	MaxSizeBytes   int64
}

// DefaultFilter returns sensible defaults for source code projects.
func DefaultFilter() Filter {
	return Filter{
		IgnoreDirs: []string{
			".git", "node_modules", "vendor", ".venv", "venv", "__pycache__",
			".terraform", "dist", "build", "target", ".next", "tmp", "coverage",
			".nyc_output", ".cache", "out",
		},
		IgnoreExts: []string{
			".jpg", ".jpeg", ".png", ".gif", ".svg", ".ico", ".webp",
			".woff", ".woff2", ".ttf", ".eot", ".otf",
			".pdf", ".zip", ".tar", ".gz", ".bz2", ".xz", ".rar",
			".exe", ".bin", ".so", ".dylib", ".dll", ".a", ".o",
			".pyc", ".class", ".jar",
			".mp3", ".mp4", ".avi", ".mov", ".mkv",
			".db", ".sqlite", ".sqlite3",
		},
		IgnorePatterns: []string{
			"*.min.js", "*.min.css", "*.map",
			"package-lock.json", "yarn.lock", "pnpm-lock.yaml",
			"go.sum", "Cargo.lock", "poetry.lock", "Gemfile.lock",
			"*.pb.go", "*.pb.gw.go",
		},
		MaxSizeBytes: 1 * 1024 * 1024, // 1 MB
	}
}

// ShouldIncludeDir returns true if the directory should be descended into.
func (f Filter) ShouldIncludeDir(name string) bool {
	lower := strings.ToLower(name)
	for _, ignore := range f.IgnoreDirs {
		if lower == strings.ToLower(ignore) {
			return false
		}
	}
	// Skip most hidden directories (starting with '.') by default.
	// Exception: the ignore list already covers .git explicitly.
	// Other hidden dirs like .vscode, .idea, etc. are usually noise.
	if strings.HasPrefix(name, ".") && name != "." {
		return false
	}
	return true
}

// ShouldIncludeFile returns true if the file should be ingested.
func (f Filter) ShouldIncludeFile(path string, size int64) bool {
	if f.MaxSizeBytes > 0 && size > f.MaxSizeBytes {
		return false
	}

	ext := strings.ToLower(filepath.Ext(path))
	for _, ignore := range f.IgnoreExts {
		if ext == ignore {
			return false
		}
	}

	base := filepath.Base(path)
	for _, pattern := range f.IgnorePatterns {
		if matched, _ := filepath.Match(pattern, base); matched {
			return false
		}
	}

	return true
}

// DetectKind classifies a file path into a kind string.
func DetectKind(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	base := strings.ToLower(filepath.Base(path))

	switch ext {
	case ".go", ".py", ".ts", ".tsx", ".js", ".jsx", ".rs", ".java",
		".kt", ".swift", ".c", ".cpp", ".h", ".hpp", ".rb", ".php",
		".cs", ".scala", ".clj", ".ex", ".exs", ".hs":
		return "source"
	case ".md", ".mdx", ".rst":
		return "markdown"
	case ".txt":
		return "text"
	case ".json", ".yaml", ".yml", ".toml", ".ini", ".env":
		return "config"
	}

	switch base {
	case "makefile", "dockerfile", "jenkinsfile", "procfile", "rakefile":
		return "config"
	case "readme", "readme.md", "readme.txt", "readme.rst":
		return "markdown"
	}

	if strings.HasSuffix(base, ".env.example") || strings.HasSuffix(base, ".env.sample") {
		return "config"
	}

	return "other"
}
