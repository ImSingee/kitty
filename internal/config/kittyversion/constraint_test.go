package kittyversion

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ysmood/gson"
)

func TestParseRequired(t *testing.T) {
	t.Run("reads string shorthand", func(t *testing.T) {
		config := map[string]gson.JSON{
			"kitty": gson.New("0.2.2"),
		}

		assert.Equal(t, "0.2.2", ParseRequired(config))
	})

	t.Run("reads version from object", func(t *testing.T) {
		config := map[string]gson.JSON{
			"kitty": gson.New(map[string]any{
				"version": ">=0.2.2",
			}),
		}

		assert.Equal(t, ">=0.2.2", ParseRequired(config))
	})

	t.Run("ignores unsupported shapes", func(t *testing.T) {
		config := map[string]gson.JSON{
			"kitty": gson.New([]any{"0.2.2"}),
		}

		assert.Equal(t, "", ParseRequired(config))
	})
}

func TestSatisfies(t *testing.T) {
	t.Run("plain version follows semver constraint semantics", func(t *testing.T) {
		ok, err := Satisfies("0.2.2", "0.2.2")
		require.NoError(t, err)
		assert.True(t, ok)

		ok, err = Satisfies("0.2.2", "0.2.3")
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("greater than is strict", func(t *testing.T) {
		ok, err := Satisfies("0.2.2", ">0.2.2")
		require.NoError(t, err)
		assert.False(t, ok)

		ok, err = Satisfies("0.2.3", ">0.2.2")
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("dev supports everything", func(t *testing.T) {
		ok, err := Satisfies("DEV", ">=999.0.0")
		require.NoError(t, err)
		assert.True(t, ok)
	})
}
