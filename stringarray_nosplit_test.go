package mflags

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringArrayNoSplitVar(t *testing.T) {
	t.Run("keeps commas in a single value", func(t *testing.T) {
		fs := NewFlagSet("test")
		var env []string
		fs.StringArrayNoSplitVar(&env, "env", 'e', nil, "env vars")

		require.NoError(t, fs.Parse([]string{"--env", "KEY=a,b,c"}))
		assert.Equal(t, []string{"KEY=a,b,c"}, env)
	})

	t.Run("one element per occurrence", func(t *testing.T) {
		fs := NewFlagSet("test")
		var env []string
		fs.StringArrayNoSplitVar(&env, "env", 'e', nil, "env vars")

		require.NoError(t, fs.Parse([]string{"-e", "A=1", "-eB=2,3", "--env=C=4,5"}))
		assert.Equal(t, []string{"A=1", "B=2,3", "C=4,5"}, env)
	})

	t.Run("default is replaced on first set", func(t *testing.T) {
		fs := NewFlagSet("test")
		var env []string
		fs.StringArrayNoSplitVar(&env, "env", 'e', []string{"D=x,y"}, "env vars")
		assert.Equal(t, []string{"D=x,y"}, env)

		require.NoError(t, fs.Parse([]string{"-e", "A=1,2"}))
		assert.Equal(t, []string{"A=1,2"}, env)
	})

	t.Run("nil default yields an empty slice", func(t *testing.T) {
		fs := NewFlagSet("test")
		var env []string
		fs.StringArrayNoSplitVar(&env, "env", 'e', nil, "env vars")
		assert.Equal(t, []string{}, env)
	})
}

func TestFromStructSplitTag(t *testing.T) {
	t.Run("split false keeps values whole", func(t *testing.T) {
		type Config struct {
			Env  []string `long:"env" short:"e" split:"false"`
			Tags []string `long:"tags" short:"t"`
		}
		var cfg Config
		fs := NewFlagSet("test")
		require.NoError(t, fs.FromStruct(&cfg))

		require.NoError(t, fs.Parse([]string{"-e", "A=1,2", "-t", "x,y", "-e", "B=3"}))
		assert.Equal(t, []string{"A=1,2", "B=3"}, cfg.Env)
		assert.Equal(t, []string{"x", "y"}, cfg.Tags, "flags without the tag still split")
	})

	t.Run("split true is the default behavior", func(t *testing.T) {
		type Config struct {
			Tags []string `long:"tags" split:"true"`
		}
		var cfg Config
		fs := NewFlagSet("test")
		require.NoError(t, fs.FromStruct(&cfg))

		require.NoError(t, fs.Parse([]string{"--tags", "x,y"}))
		assert.Equal(t, []string{"x", "y"}, cfg.Tags)
	})

	t.Run("default value is kept whole", func(t *testing.T) {
		type Config struct {
			Env []string `long:"env" split:"false" default:"A=1,2"`
		}
		var cfg Config
		fs := NewFlagSet("test")
		require.NoError(t, fs.FromStruct(&cfg))
		assert.Equal(t, []string{"A=1,2"}, cfg.Env)
	})

	t.Run("other values are rejected", func(t *testing.T) {
		type Config struct {
			Env []string `long:"env" split:"maybe"`
		}
		var cfg Config
		fs := NewFlagSet("test")
		err := fs.FromStruct(&cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "split")
	})
}
