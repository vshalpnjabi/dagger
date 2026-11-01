package schema

import (
	"context"
	"path"
	"path/filepath"
	"strings"

	"github.com/dagger/dagger/core"
)

// workspaceConfigIgnore returns the workspace's dagger.toml ignore patterns.
func workspaceConfigIgnore(ctx context.Context, ws *core.Workspace) ([]string, error) {
	if ws == nil {
		return nil, nil
	}
	cfg, err := workspaceEffectiveConfig(ctx, ws)
	if err != nil {
		return nil, err
	}
	if cfg == nil || len(cfg.Ignore) == 0 {
		return nil, nil
	}
	return append([]string(nil), cfg.Ignore...), nil
}

// withWorkspaceIgnore appends dagger.toml ignore patterns to a copy filter
// so they apply as Directory exclude globs. A bare path also excludes its
// descendants, so include globs such as **/go.mod cannot resurrect them.
func withWorkspaceIgnore(filter core.CopyFilter, ignore []string) core.CopyFilter {
	if len(ignore) == 0 {
		return filter
	}
	filter.Exclude = append(append([]string{}, filter.Exclude...), workspaceIgnoreExcludePatterns(ignore)...)
	return filter
}

func workspaceIgnoreExcludePatterns(ignore []string) []string {
	out := make([]string, 0, len(ignore)*2)
	for _, pat := range ignore {
		pat = strings.TrimSpace(pat)
		if pat == "" {
			continue
		}
		out = append(out, pat)
		if strings.ContainsAny(pat, "*?[") || strings.HasSuffix(pat, "/**") {
			continue
		}
		out = append(out, strings.TrimSuffix(pat, "/")+"/**")
	}
	return out
}

// workspaceRelPathIgnored reports whether a workspace-root-relative path is
// covered by dagger.toml ignore patterns. A pattern without slashes or globs
// matches that name as a path prefix or as any path segment.
func workspaceRelPathIgnored(rel string, ignore []string) bool {
	rel = path.Clean(filepath.ToSlash(strings.TrimPrefix(rel, "/")))
	if rel == "." || rel == "" || len(ignore) == 0 {
		return false
	}
	for _, pat := range ignore {
		pat = strings.TrimSpace(filepath.ToSlash(pat))
		if pat == "" {
			continue
		}
		trimmed := strings.TrimSuffix(pat, "/")
		trimmed = strings.TrimSuffix(trimmed, "/**")
		if !strings.ContainsAny(trimmed, "*?[") {
			if rel == trimmed || strings.HasPrefix(rel, trimmed+"/") {
				return true
			}
			if !strings.Contains(trimmed, "/") {
				for _, seg := range strings.Split(rel, "/") {
					if seg == trimmed {
						return true
					}
				}
			}
			continue
		}
		if matched, err := path.Match(pat, rel); err == nil && matched {
			return true
		}
	}
	return false
}

// filterIgnoredWorkspaceDirs drops cwd-relative findRoots results that sit
// under a dagger.toml ignore pattern. Conversion to a workspace-root path
// keeps the match independent of cwd.
func filterIgnoredWorkspaceDirs(dirs []string, cwd string, ignore []string) ([]string, error) {
	if len(ignore) == 0 {
		return dirs, nil
	}
	out := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		rootRel, err := resolveWorkspacePath(dir, cwd)
		if err != nil {
			return nil, err
		}
		if workspaceRelPathIgnored(rootRel, ignore) {
			continue
		}
		out = append(out, dir)
	}
	return out, nil
}
