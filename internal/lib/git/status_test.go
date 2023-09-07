package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseStatus(t *testing.T) {
	t.Run("parse empty", func(t *testing.T) {
		all, err := parseStatus("\n")
		require.NoError(t, err)
		assert.Empty(t, all)

		all, err = parseStatus("")
		require.NoError(t, err)
		assert.Empty(t, all)
	})

	t.Run("parse simple", func(t *testing.T) {
		all, err := parseStatus(`
A  a.txt
 A b.txt
R  c.txt -> d.txt
`)
		require.NoError(t, err)
		assert.Equal(t, []FileStatus{
			{Name: "a.txt", IndexStatus: 'A', WorkingTreeStatus: ' '},
			{Name: "b.txt", IndexStatus: ' ', WorkingTreeStatus: 'A'},
			{Name: "d.txt", IndexStatus: 'R', WorkingTreeStatus: ' ', RenameFrom: "c.txt"},
		}, all)
	})
}
