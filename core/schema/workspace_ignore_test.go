package schema

import (
	"testing"

	"github.com/dagger/dagger/core"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceRelPathIgnored(t *testing.T) {
	require.True(t, workspaceRelPathIgnored(".delta/worktrees/nested", []string{".delta"}))
	require.True(t, workspaceRelPathIgnored(".delta", []string{".delta"}))
	require.True(t, workspaceRelPathIgnored("foo/.delta/bar", []string{".delta"}))
	require.False(t, workspaceRelPathIgnored("pkg", []string{".delta"}))
	require.False(t, workspaceRelPathIgnored(".", []string{".delta"}))
	require.False(t, workspaceRelPathIgnored("pkg", nil))
	require.True(t, workspaceRelPathIgnored("vendor/lib", []string{"vendor"}))
}

func TestWithWorkspaceIgnore(t *testing.T) {
	filter := withWorkspaceIgnore(core.CopyFilter{
		Include: []string{"**/go.mod"},
		Exclude: []string{"**/node_modules/**"},
	}, []string{".delta"})
	require.Equal(t, []string{"**/go.mod"}, filter.Include)
	require.Contains(t, filter.Exclude, "**/node_modules/**")
	require.Contains(t, filter.Exclude, ".delta")
	require.Contains(t, filter.Exclude, ".delta/**")
}

func TestFilterIgnoredWorkspaceDirs(t *testing.T) {
	dirs, err := filterIgnoredWorkspaceDirs(
		[]string{".", "pkg", ".delta/worktrees/nested"},
		".",
		[]string{".delta"},
	)
	require.NoError(t, err)
	require.Equal(t, []string{".", "pkg"}, dirs)

	dirs, err = filterIgnoredWorkspaceDirs(
		[]string{"..", ".", "dir/sub"},
		"sub",
		[]string{".delta"},
	)
	require.NoError(t, err)
	require.Equal(t, []string{"..", ".", "dir/sub"}, dirs)
}
