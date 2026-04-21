package lintstaged

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptionsSelectionMode(t *testing.T) {
	t.Run("default uses staged selection", func(t *testing.T) {
		options := &Options{}

		require.NoError(t, options.ValidateSelectionMode())
		assert.Equal(t, SelectionModeStaged, options.SelectionMode())
		assert.True(t, options.UsesIndex())
		assert.Equal(t, "staged files", options.SelectedFilesLabel())
	})

	t.Run("custom status uses working tree only", func(t *testing.T) {
		options := &Options{Status: string(SelectionModeTracked)}

		require.NoError(t, options.ValidateSelectionMode())
		assert.Equal(t, SelectionModeTracked, options.SelectionMode())
		assert.False(t, options.UsesIndex())
		assert.Equal(t, "`--status=tracked` was used", options.SelectionReason())
		assert.Equal(t, "tracked changed files", options.SelectedFilesLabel())
	})

	t.Run("diff uses working tree only", func(t *testing.T) {
		options := &Options{Diff: "HEAD"}

		require.NoError(t, options.ValidateSelectionMode())
		assert.False(t, options.UsesIndex())
		assert.Equal(t, "`--diff` was used", options.SelectionReason())
		assert.Equal(t, "selected files", options.SelectedFilesLabel())
	})

	t.Run("invalid status rejected", func(t *testing.T) {
		options := &Options{Status: "weird"}

		require.Error(t, options.ValidateSelectionMode())
	})

	t.Run("diff cannot combine with non default status", func(t *testing.T) {
		options := &Options{
			Diff:   "HEAD",
			Status: string(SelectionModeAll),
		}

		require.Error(t, options.ValidateSelectionMode())
	})
}
