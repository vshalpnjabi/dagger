package workspace

import (
	"path/filepath"
	"strings"
)

// IgnorePatterns resolves dagger.toml ignore values relative to the directory
// holding configFile. Patterns that escape the workspace root are dropped.
// Missing or empty input yields nil.
func IgnorePatterns(configFile string, ignore []string) []string {
	if len(ignore) == 0 {
		return nil
	}
	configDir := filepath.Dir(configFile)
	if configDir == "" {
		configDir = "."
	}
	out := make([]string, 0, len(ignore))
	for _, pattern := range ignore {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		clean, negative := strings.CutPrefix(pattern, "!")
		clean = strings.TrimSpace(clean)
		if clean == "" {
			continue
		}
		fromSlash := filepath.FromSlash(clean)
		if !filepath.IsLocal(fromSlash) {
			continue
		}
		resolved := filepath.ToSlash(filepath.Clean(filepath.Join(configDir, fromSlash)))
		if negative {
			resolved = "!" + resolved
		}
		out = append(out, resolved)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
