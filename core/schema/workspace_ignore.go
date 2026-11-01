package schema

import (
	"context"
	"strings"

	"github.com/dagger/dagger/core"
)

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

// withWorkspaceIgnore appends dagger.toml ignore patterns to a copy filter.
// A bare path also excludes its descendants so include globs cannot resurrect them.
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
