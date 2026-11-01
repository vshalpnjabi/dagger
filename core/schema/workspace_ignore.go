package schema

import (
	"context"
	"path"
	"path/filepath"
	"strings"

	"github.com/dagger/dagger/core"
	"github.com/dagger/dagger/core/workspace"
	"github.com/dagger/dagger/dagql"
)

// loadWorkspaceIgnorePatterns reads ignore from the selected dagger.toml in
// root. Missing or unreadable config yields no extra excludes.
func loadWorkspaceIgnorePatterns(ctx context.Context, root dagql.ObjectResult[*core.Directory], configFile string) []string {
	if configFile == "" {
		return nil
	}
	data, err := core.DirectoryReadFile(ctx, root, filepath.ToSlash(configFile))
	if err != nil {
		return nil
	}
	cfg, err := workspace.ParseConfig(data)
	if err != nil {
		return nil
	}
	return workspace.IgnorePatterns(configFile, cfg.Ignore)
}

func workspaceFilterWithIgnore(ws *core.Workspace, resolvedPath string, filter core.CopyFilter) core.CopyFilter {
	if ws == nil {
		return filter
	}
	ignore := workspaceIgnoreForPath(ws.IgnorePatterns(), resolvedPath)
	if len(ignore) == 0 {
		return filter
	}
	// User ignore comes first so a dagger.toml `!pattern` cannot re-include a
	// path the engine already excluded.
	filter.Exclude = append(append([]string(nil), ignore...), filter.Exclude...)
	return filter
}

func workspaceIgnoreForPath(ignore []string, resolvedPath string) []string {
	if len(ignore) == 0 {
		return nil
	}
	resolvedPath = path.Clean(filepath.ToSlash(resolvedPath))
	if resolvedPath == "" || resolvedPath == "/" {
		resolvedPath = "."
	}
	if resolvedPath == "." {
		return append([]string(nil), ignore...)
	}
	patterns := make([]string, 0, len(ignore))
	for _, pattern := range ignore {
		if rel, ok := workspaceIgnorePatternForPath(pattern, resolvedPath); ok {
			patterns = append(patterns, rel)
		}
	}
	if len(patterns) == 0 {
		return nil
	}
	return patterns
}

func workspaceIgnorePatternForPath(pattern, resolvedPath string) (string, bool) {
	pattern, negative := strings.CutPrefix(pattern, "!")
	if pattern == "" {
		return "", false
	}
	clean := path.Clean(filepath.ToSlash(pattern))
	clean = strings.TrimPrefix(clean, "/")
	if clean == "." {
		clean = "*"
	}

	var rel string
	switch {
	case clean == resolvedPath:
		rel = "*"
	case strings.HasPrefix(clean, resolvedPath+"/"):
		rel = strings.TrimPrefix(clean, resolvedPath+"/")
	case workspaceIgnorePatternCoversPath(clean, resolvedPath):
		rel = "*"
	case workspaceIgnorePatternMatchesAnywhere(clean):
		rel = clean
	default:
		return "", false
	}
	if negative {
		rel = "!" + rel
	}
	return rel, true
}

func workspaceIgnorePatternCoversPath(pattern, resolvedPath string) bool {
	if strings.ContainsAny(pattern, "*?[") {
		return false
	}
	return strings.HasPrefix(resolvedPath, pattern+"/")
}

func workspaceIgnorePatternMatchesAnywhere(pattern string) bool {
	if strings.Contains(pattern, "**") {
		return true
	}
	first, _, _ := strings.Cut(pattern, "/")
	return strings.ContainsAny(first, "*?[")
}
