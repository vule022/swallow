package prompt

import "fmt"

// DetailLevel controls how much context is included and which sections are required.
type DetailLevel string

const (
	DetailCompact  DetailLevel = "compact"
	DetailStandard DetailLevel = "standard"
	DetailDetailed DetailLevel = "detailed"
)

// Parse parses a string into a DetailLevel.
func Parse(s string) (DetailLevel, error) {
	switch s {
	case "compact":
		return DetailCompact, nil
	case "standard", "":
		return DetailStandard, nil
	case "detailed", "full":
		return DetailDetailed, nil
	default:
		return DetailStandard, fmt.Errorf("unknown detail level %q: use compact, standard, or detailed", s)
	}
}

// ContextLimits returns how many outputs, sessions, and doc summaries to include.
func (d DetailLevel) ContextLimits() (maxOutputs, maxSessions, maxDocs int) {
	switch d {
	case DetailCompact:
		return 2, 2, 3
	case DetailDetailed:
		return 10, 8, 15
	default: // standard
		return 5, 5, 8
	}
}
