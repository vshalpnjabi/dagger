package schema

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWorkspaceIgnoreForPath(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		ignore []string
		read   string
		want   []string
	}{
		{
			name:   "root keeps configured patterns",
			ignore: []string{".delta", "pkg/vendor", "*.log", "**/.generated"},
			read:   ".",
			want:   []string{".delta", "pkg/vendor", "*.log", "**/.generated"},
		},
		{
			name:   "root glob is not applied under a subdirectory",
			ignore: []string{"*.log"},
			read:   "pkg",
		},
		{
			name:   "one-segment glob strips against the read path",
			ignore: []string{"*/vendor"},
			read:   "pkg",
			want:   []string{"vendor"},
		},
		{
			name:   "one-segment glob does not hide a deeper vendor",
			ignore: []string{"*/vendor"},
			read:   "pkg/x",
		},
		{
			name:   "read inside a matching glob dir is fully excluded",
			ignore: []string{".gen*"},
			read:   ".generated",
			want:   []string{"*"},
		},
		{
			name:   "read under a starstar ignore is fully excluded",
			ignore: []string{"**/.generated"},
			read:   ".generated/sub",
			want:   []string{"*", "**/.generated"},
		},
		{
			name:   "starstar patterns stay as-is under an unrelated path",
			ignore: []string{"**/.generated"},
			read:   "pkg",
			want:   []string{"**/.generated"},
		},
		{
			name:   "sibling of an ignored dir is not excluded",
			ignore: []string{".delta"},
			read:   "pkg",
		},
		{
			name:   "prefix under the read path is stripped",
			ignore: []string{"pkg/vendor"},
			read:   "pkg",
			want:   []string{"vendor"},
		},
		{
			name:   "exact ignored dir excludes everything",
			ignore: []string{".delta"},
			read:   ".delta",
			want:   []string{"*"},
		},
		{
			name:   "child of an ignored dir excludes everything",
			ignore: []string{".delta"},
			read:   ".delta/worktrees",
			want:   []string{"*"},
		},
		{
			name:   "negation keeps bang through a covering match",
			ignore: []string{".delta", "!.delta/keep"},
			read:   ".delta",
			want:   []string{"*", "!keep"},
		},
		{
			name:   "negation can unignore the read path",
			ignore: []string{".delta", "!.delta"},
			read:   ".delta",
			want:   []string{"*", "!*"},
		},
		{
			name:   "negated root glob is dropped under a subdirectory",
			ignore: []string{"*.log", "!keep.log"},
			read:   "pkg",
		},
		{
			name:   "nested glob after an exact prefix",
			ignore: []string{"pkg/*/tmp"},
			read:   "pkg/foo",
			want:   []string{"tmp"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, workspaceIgnoreForPath(tc.ignore, tc.read))
		})
	}
}

func TestWorkspaceIgnoreFromConfigBytes(t *testing.T) {
	t.Parallel()

	t.Run("not found is no patterns", func(t *testing.T) {
		t.Parallel()
		got, err := workspaceIgnoreFromConfigBytes(nil, "dagger.toml", os.ErrNotExist)
		require.NoError(t, err)
		require.Nil(t, got)
	})

	t.Run("context cancel is propagated", func(t *testing.T) {
		t.Parallel()
		got, err := workspaceIgnoreFromConfigBytes(nil, "dagger.toml", context.Canceled)
		require.ErrorIs(t, err, context.Canceled)
		require.Nil(t, got)
	})

	t.Run("parse error is no patterns", func(t *testing.T) {
		t.Parallel()
		got, err := workspaceIgnoreFromConfigBytes([]byte("ignore = [\n"), "dagger.toml", nil)
		require.NoError(t, err)
		require.Nil(t, got)
	})

	t.Run("leading slash and dot", func(t *testing.T) {
		t.Parallel()
		got, err := workspaceIgnoreFromConfigBytes([]byte("ignore = [\"/dist\", \".\"]\n"), "dagger.toml", nil)
		require.NoError(t, err)
		require.Equal(t, []string{"dist"}, got)
	})
}
