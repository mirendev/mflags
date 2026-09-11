package mflags

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLevenshtein(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"", "abc", 3},
		{"abc", "", 3},
		{"deploy", "deploy", 0},
		{"depoy", "deploy", 1},   // deletion
		{"deployy", "deploy", 1}, // insertion
		{"deploj", "deploy", 1},  // substitution
		{"depoly", "deploy", 2},  // transposition costs two here
		{"kitten", "sitting", 3},
		{"café", "cafe", 1}, // one rune, not two bytes
	}

	for _, c := range cases {
		assert.Equal(t, c.want, levenshtein(c.a, c.b), "levenshtein(%q, %q)", c.a, c.b)
		assert.Equal(t, c.want, levenshtein(c.b, c.a), "levenshtein(%q, %q)", c.b, c.a)
	}
}

func TestSuggestNames(t *testing.T) {
	commands := []string{"app", "deploy", "logs", "rollback", "runner", "server", "upgrade", "version"}

	t.Run("one edit away", func(t *testing.T) {
		assert.Equal(t, []string{"deploy"}, suggestNames("depoy", commands))
	})

	t.Run("transposition", func(t *testing.T) {
		assert.Equal(t, []string{"deploy"}, suggestNames("depoly", commands))
	})

	t.Run("prefix beats distance", func(t *testing.T) {
		assert.Equal(t, []string{"deploy"}, suggestNames("depl", commands))
	})

	t.Run("prefix matches rank ahead of edit-distance matches", func(t *testing.T) {
		got := suggestNames("serve", []string{"server", "serve-it", "sever"})
		assert.Equal(t, []string{"server", "serve-it", "sever"}, got)
	})

	t.Run("shorter prefix completion wins", func(t *testing.T) {
		assert.Equal(t, []string{"apply", "applesauce"}, suggestNames("appl", []string{"apply", "applesauce"}))
	})

	t.Run("nothing close returns nil", func(t *testing.T) {
		assert.Nil(t, suggestNames("zzzzzz", commands))
	})

	t.Run("short words are not guessed at on distance alone", func(t *testing.T) {
		// "vp" is one edit from "up"; at two characters that is too weak a
		// signal to guess on.
		assert.Nil(t, suggestNames("vp", []string{"up", "down"}))
	})

	t.Run("but a short prefix still counts", func(t *testing.T) {
		// "ap" is only two characters, yet it can only be reaching for "app".
		assert.Equal(t, []string{"app"}, suggestNames("ap", commands))
	})

	t.Run("empty input", func(t *testing.T) {
		assert.Nil(t, suggestNames("", commands))
		assert.Nil(t, suggestNames("deploy", nil))
	})

	t.Run("capped at three", func(t *testing.T) {
		got := suggestNames("aaa", []string{"aab", "aac", "aad", "aae", "aaf"})
		assert.Len(t, got, maxSuggestions)
	})

	t.Run("ties are alphabetical", func(t *testing.T) {
		got := suggestNames("aaa", []string{"aaf", "aab", "aad"})
		assert.Equal(t, []string{"aab", "aad", "aaf"}, got)
	})
}
