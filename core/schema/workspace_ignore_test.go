package schema

import (
	"testing"

	"github.com/dagger/dagger/core"
	"github.com/stretchr/testify/require"
)

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
