package schema

import (
	"context"
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/dagger/dagger/core"
	"github.com/dagger/dagger/core/workspace"
	"github.com/dagger/dagger/dagql"
	"github.com/moby/patternmatcher"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// loadWorkspaceIgnorePatterns reads ignore from the selected dagger.toml in
// root. A missing file or unparseable config yields no extra excludes.
// Other read errors (context cancel, I/O) are returned to the caller.
func loadWorkspaceIgnorePatterns(ctx context.Context, root dagql.ObjectResult[*core.Directory], configFile string) ([]string, error) {
	if configFile == "" {
		return nil, nil
	}
	data, err := core.DirectoryReadFile(ctx, root, filepath.ToSlash(configFile))
	return workspaceIgnoreFromConfigBytes(data, configFile, err)
}

func workspaceIgnoreFromConfigBytes(data []byte, configFile string, err error) ([]string, error) {
	if err != nil {
		if isAbsentWorkspaceConfig(err) {
			return nil, nil
		}
		return nil, err
	}
	cfg, err := workspace.ParseConfig(data)
	if err != nil {
		// A broken dagger.toml adds no excludes; do not fail workspace load.
		return nil, nil //nolint:nilerr
	}
	return workspace.IgnorePatterns(configFile, cfg.Ignore), nil
}

func isAbsentWorkspaceConfig(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, os.ErrNotExist) {
		return true
	}
	if status.Code(err) == codes.NotFound {
		return true
	}
	return strings.Contains(err.Error(), "no such file or directory")
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

	covered := workspaceReadPathIgnored(ignore, resolvedPath)
	patterns := make([]string, 0, len(ignore)+1)
	if covered {
		patterns = append(patterns, "*")
	}
	for _, pattern := range ignore {
		rel, ok := workspaceIgnorePatternForPath(pattern, resolvedPath)
		if !ok {
			continue
		}
		if rel == "*" && covered {
			continue
		}
		patterns = append(patterns, rel)
	}
	if len(patterns) == 0 {
		return nil
	}
	return patterns
}

func workspaceReadPathIgnored(ignore []string, resolvedPath string) bool {
	pm, err := patternmatcher.New(ignore)
	if err != nil {
		return false
	}
	ok, err := pm.MatchesOrParentMatches(resolvedPath)
	return err == nil && ok
}

func workspaceIgnorePatternForPath(pattern, resolvedPath string) (string, bool) {
	pattern, negative := strings.CutPrefix(pattern, "!")
	if pattern == "" {
		return "", false
	}
	clean := path.Clean(filepath.ToSlash(pattern))
	if clean == "." {
		return "", false
	}

	var rel string
	switch {
	case strings.HasPrefix(clean, "**/"):
		rel = clean
	case clean == resolvedPath:
		rel = "*"
	case strings.HasPrefix(clean, resolvedPath+"/"):
		rel = strings.TrimPrefix(clean, resolvedPath+"/")
	default:
		stripped, ok := workspaceIgnoreStripGlobPrefix(clean, resolvedPath)
		if !ok {
			return "", false
		}
		rel = stripped
	}
	if rel == "" {
		rel = "*"
	}
	if negative {
		rel = "!" + rel
	}
	return rel, true
}

func workspaceIgnoreStripGlobPrefix(pattern, resolvedPath string) (string, bool) {
	pSegs := strings.Split(pattern, "/")
	rSegs := strings.Split(resolvedPath, "/")
	i := 0
	for i < len(pSegs) && i < len(rSegs) {
		ok, err := path.Match(pSegs[i], rSegs[i])
		if err != nil || !ok {
			break
		}
		i++
	}
	if i == 0 || i != len(rSegs) {
		return "", false
	}
	if i == len(pSegs) {
		return "*", true
	}
	return strings.Join(pSegs[i:], "/"), true
}
