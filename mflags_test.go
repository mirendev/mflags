package mflags

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBasicBoolFlag(t *testing.T) {
	fs := NewFlagSet("test")
	verbose := fs.Bool("verbose", 'v', false, "verbose output")

	err := fs.Parse([]string{"--verbose"})
	assert.NoError(t, err)
	assert.True(t, *verbose)
	assert.Empty(t, fs.Args())
}

func TestShortBoolFlag(t *testing.T) {
	fs := NewFlagSet("test")
	verbose := fs.Bool("verbose", 'v', false, "verbose output")

	err := fs.Parse([]string{"-v"})
	assert.NoError(t, err)
	assert.True(t, *verbose)
	assert.Empty(t, fs.Args())
}

func TestCombinedShortFlags(t *testing.T) {
	fs := NewFlagSet("test")
	verbose := fs.Bool("verbose", 'v', false, "verbose output")
	list := fs.Bool("list", 'l', false, "list mode")
	all := fs.Bool("all", 'a', false, "show all")

	err := fs.Parse([]string{"-vla"})
	assert.NoError(t, err)
	assert.True(t, *verbose)
	assert.True(t, *list)
	assert.True(t, *all)
	assert.Empty(t, fs.Args())
}

func TestStringFlag(t *testing.T) {
	fs := NewFlagSet("test")
	name := fs.String("name", 'n', "default", "name to use")

	err := fs.Parse([]string{"--name", "test-value"})
	assert.NoError(t, err)
	assert.Equal(t, "test-value", *name)
	assert.Empty(t, fs.Args())
}

func TestStringFlagWithEquals(t *testing.T) {
	fs := NewFlagSet("test")
	name := fs.String("name", 'n', "default", "name to use")

	err := fs.Parse([]string{"--name=test-value"})
	assert.NoError(t, err)
	assert.Equal(t, "test-value", *name)
	assert.Empty(t, fs.Args())
}

func TestShortStringFlag(t *testing.T) {
	fs := NewFlagSet("test")
	name := fs.String("name", 'n', "default", "name to use")

	err := fs.Parse([]string{"-n", "test-value"})
	assert.NoError(t, err)
	assert.Equal(t, "test-value", *name)
	assert.Empty(t, fs.Args())
}

func TestIntFlag(t *testing.T) {
	fs := NewFlagSet("test")
	count := fs.Int("count", 'c', 0, "count value")

	err := fs.Parse([]string{"--count", "42"})
	assert.NoError(t, err)
	assert.Equal(t, 42, *count)
	assert.Empty(t, fs.Args())
}

func TestMixedFlagsAndArgs(t *testing.T) {
	fs := NewFlagSet("test")
	verbose := fs.Bool("verbose", 'v', false, "verbose output")
	name := fs.String("name", 'n', "default", "name to use")

	err := fs.Parse([]string{"-v", "arg1", "--name", "test", "arg2", "arg3"})
	assert.NoError(t, err)
	assert.True(t, *verbose)
	assert.Equal(t, "test", *name)
	assert.Equal(t, []string{"arg1", "arg2", "arg3"}, fs.Args())
}

func TestFlagsAfterDoubleHyphen(t *testing.T) {
	fs := NewFlagSet("test")
	verbose := fs.Bool("verbose", 'v', false, "verbose output")

	err := fs.Parse([]string{"-v", "--", "-v", "--verbose"})
	assert.NoError(t, err)
	assert.True(t, *verbose)
	assert.Equal(t, []string{"-v", "--verbose"}, fs.Args())
}

func TestUnknownFlag(t *testing.T) {
	fs := NewFlagSet("test")

	err := fs.Parse([]string{"--unknown"})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUnknownFlag)
}

func TestUnknownShortFlag(t *testing.T) {
	fs := NewFlagSet("test")

	err := fs.Parse([]string{"-x"})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUnknownFlag)
}

func TestMissingFlagValue(t *testing.T) {
	fs := NewFlagSet("test")
	fs.String("name", 'n', "default", "name to use")

	err := fs.Parse([]string{"--name"})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrMissingValue)
}

func TestMissingShortFlagValue(t *testing.T) {
	fs := NewFlagSet("test")
	fs.String("name", 'n', "default", "name to use")

	err := fs.Parse([]string{"-n"})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrMissingValue)
}

func TestCombinedShortFlagsWithValue(t *testing.T) {
	fs := NewFlagSet("test")
	verbose := fs.Bool("verbose", 'v', false, "verbose output")
	name := fs.String("name", 'n', "default", "name to use")

	err := fs.Parse([]string{"-vn", "test-value"})
	assert.NoError(t, err)
	assert.True(t, *verbose)
	assert.Equal(t, "test-value", *name)
	assert.Empty(t, fs.Args())
}

func TestBoolFlagWithExplicitValue(t *testing.T) {
	fs := NewFlagSet("test")
	verbose := fs.Bool("verbose", 'v', true, "verbose output")

	err := fs.Parse([]string{"--verbose=false"})
	assert.NoError(t, err)
	assert.False(t, *verbose)
	assert.Empty(t, fs.Args())
}

func TestDefaultValues(t *testing.T) {
	fs := NewFlagSet("test")
	verbose := fs.Bool("verbose", 'v', true, "verbose output")
	name := fs.String("name", 'n', "default-name", "name to use")
	count := fs.Int("count", 'c', 10, "count value")

	err := fs.Parse([]string{})
	assert.NoError(t, err)
	assert.True(t, *verbose)
	assert.Equal(t, "default-name", *name)
	assert.Equal(t, 10, *count)
	assert.Empty(t, fs.Args())
}

func TestVarMethod(t *testing.T) {
	fs := NewFlagSet("test")

	var boolVal bool
	fs.BoolVar(&boolVal, "bool", 'b', false, "bool flag")

	var stringVal string
	fs.StringVar(&stringVal, "string", 's', "default", "string flag")

	var intVal int
	fs.IntVar(&intVal, "int", 'i', 0, "int flag")

	err := fs.Parse([]string{"-b", "--string", "test", "-i", "42"})
	assert.NoError(t, err)
	assert.True(t, boolVal)
	assert.Equal(t, "test", stringVal)
	assert.Equal(t, 42, intVal)
}

func TestParsedFlag(t *testing.T) {
	fs := NewFlagSet("test")

	assert.False(t, fs.Parsed())

	err := fs.Parse([]string{})
	assert.NoError(t, err)
	assert.True(t, fs.Parsed())
}

func TestComplexScenario(t *testing.T) {
	fs := NewFlagSet("test")
	verbose := fs.Bool("verbose", 'v', false, "verbose output")
	quiet := fs.Bool("quiet", 'q', false, "quiet mode")
	output := fs.String("output", 'o', "stdout", "output file")
	level := fs.Int("level", 'l', 1, "level")

	args := []string{
		"cmd",
		"-vq",
		"file1.txt",
		"--output=result.txt",
		"-l", "3",
		"file2.txt",
		"--",
		"-v",
		"--output=other.txt",
	}

	err := fs.Parse(args)
	assert.NoError(t, err)
	assert.True(t, *verbose)
	assert.True(t, *quiet)
	assert.Equal(t, "result.txt", *output)
	assert.Equal(t, 3, *level)
	assert.Equal(t, []string{"cmd", "file1.txt", "file2.txt", "-v", "--output=other.txt"}, fs.Args())
}

func TestInvalidIntValue(t *testing.T) {
	fs := NewFlagSet("test")
	fs.Int("count", 'c', 0, "count value")

	err := fs.Parse([]string{"--count", "not-a-number"})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidValue)
}

func TestInvalidBoolValue(t *testing.T) {
	fs := NewFlagSet("test")
	fs.Bool("verbose", 'v', false, "verbose output")

	err := fs.Parse([]string{"--verbose=maybe"})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidValue)
}

func TestShortFlagWithImmediateValue(t *testing.T) {
	fs := NewFlagSet("test")
	name := fs.String("name", 'n', "default", "name to use")

	err := fs.Parse([]string{"-nvalue"})
	assert.NoError(t, err)
	assert.Equal(t, "value", *name)
	assert.Empty(t, fs.Args())
}

func TestOnlyDoubleHyphen(t *testing.T) {
	fs := NewFlagSet("test")
	verbose := fs.Bool("verbose", 'v', false, "verbose output")

	err := fs.Parse([]string{"--", "arg1", "arg2"})
	assert.NoError(t, err)
	assert.False(t, *verbose)
	assert.Equal(t, []string{"arg1", "arg2"}, fs.Args())
}

func TestCombinedShortFlagsWithMultipleArguments(t *testing.T) {
	fs := NewFlagSet("test")
	fs.String("aaa", 'a', "default", "first string")
	fs.String("bbb", 'b', "default", "second string")

	err := fs.Parse([]string{"-ab", "value"})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrMissingValue)
}

func TestCombinedBoolAndStringFlag(t *testing.T) {
	fs := NewFlagSet("test")
	verbose := fs.Bool("verbose", 'v', false, "verbose output")
	name := fs.String("name", 'n', "default", "name to use")

	// This should work: bool flag followed by string flag
	err := fs.Parse([]string{"-vn", "test-value"})
	assert.NoError(t, err)
	assert.True(t, *verbose)
	assert.Equal(t, "test-value", *name)
}

func TestStringFlagFollowedByBool(t *testing.T) {
	fs := NewFlagSet("test")
	name := fs.String("name", 'n', "default", "name to use")
	verbose := fs.Bool("verbose", 'v', false, "verbose output")

	// This should work: string flag takes 'v' as its value
	err := fs.Parse([]string{"-nv"})
	assert.NoError(t, err)
	assert.Equal(t, "v", *name)
	assert.False(t, *verbose) // v was consumed as value, not parsed as flag
}

func TestStringArrayFlag(t *testing.T) {
	fs := NewFlagSet("test")
	tags := fs.StringArray("tags", 't', nil, "tags to apply")

	err := fs.Parse([]string{"--tags", "foo,bar,baz"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"foo", "bar", "baz"}, *tags)
	assert.Empty(t, fs.Args())
}

func TestStringArrayFlagWithEquals(t *testing.T) {
	fs := NewFlagSet("test")
	tags := fs.StringArray("tags", 't', nil, "tags to apply")

	err := fs.Parse([]string{"--tags=alpha,beta,gamma"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"alpha", "beta", "gamma"}, *tags)
	assert.Empty(t, fs.Args())
}

func TestShortStringArrayFlag(t *testing.T) {
	fs := NewFlagSet("test")
	tags := fs.StringArray("tags", 't', nil, "tags to apply")

	err := fs.Parse([]string{"-t", "one,two,three"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"one", "two", "three"}, *tags)
	assert.Empty(t, fs.Args())
}

func TestStringArrayWithSingleValue(t *testing.T) {
	fs := NewFlagSet("test")
	tags := fs.StringArray("tags", 't', nil, "tags to apply")

	err := fs.Parse([]string{"--tags", "single"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"single"}, *tags)
	assert.Empty(t, fs.Args())
}

func TestStringArrayWithEmptyValue(t *testing.T) {
	fs := NewFlagSet("test")
	tags := fs.StringArray("tags", 't', nil, "tags to apply")

	err := fs.Parse([]string{"--tags", ""})
	assert.NoError(t, err)
	assert.Equal(t, []string{""}, *tags)
	assert.Empty(t, fs.Args())
}

func TestStringArrayDefaultValue(t *testing.T) {
	fs := NewFlagSet("test")
	tags := fs.StringArray("tags", 't', []string{"default", "values"}, "tags to apply")

	err := fs.Parse([]string{})
	assert.NoError(t, err)
	assert.Equal(t, []string{"default", "values"}, *tags)
	assert.Empty(t, fs.Args())
}

func TestStringArrayVarMethod(t *testing.T) {
	fs := NewFlagSet("test")

	var tags []string
	fs.StringArrayVar(&tags, "tags", 't', []string{"initial"}, "tags to apply")

	err := fs.Parse([]string{"--tags", "new,values"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"new", "values"}, tags)
}

func TestStringArrayWithSpaces(t *testing.T) {
	fs := NewFlagSet("test")
	tags := fs.StringArray("tags", 't', nil, "tags to apply")

	err := fs.Parse([]string{"--tags", "foo bar,baz qux,test"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"foo bar", "baz qux", "test"}, *tags)
	assert.Empty(t, fs.Args())
}

func TestStringArrayMixedWithOtherFlags(t *testing.T) {
	fs := NewFlagSet("test")
	verbose := fs.Bool("verbose", 'v', false, "verbose output")
	tags := fs.StringArray("tags", 't', nil, "tags to apply")
	name := fs.String("name", 'n', "default", "name to use")

	err := fs.Parse([]string{"-v", "--tags", "a,b,c", "--name", "test", "arg1"})
	assert.NoError(t, err)
	assert.True(t, *verbose)
	assert.Equal(t, []string{"a", "b", "c"}, *tags)
	assert.Equal(t, "test", *name)
	assert.Equal(t, []string{"arg1"}, fs.Args())
}

func TestDurationFlag(t *testing.T) {
	fs := NewFlagSet("test")
	timeout := fs.Duration("timeout", 't', 0, "timeout duration")

	err := fs.Parse([]string{"--timeout", "5s"})
	assert.NoError(t, err)
	assert.Equal(t, 5*time.Second, *timeout)
	assert.Empty(t, fs.Args())
}

func TestDurationFlagWithEquals(t *testing.T) {
	fs := NewFlagSet("test")
	timeout := fs.Duration("timeout", 't', 0, "timeout duration")

	err := fs.Parse([]string{"--timeout=1m30s"})
	assert.NoError(t, err)
	assert.Equal(t, 90*time.Second, *timeout)
	assert.Empty(t, fs.Args())
}

func TestShortDurationFlag(t *testing.T) {
	fs := NewFlagSet("test")
	timeout := fs.Duration("timeout", 't', 0, "timeout duration")

	err := fs.Parse([]string{"-t", "100ms"})
	assert.NoError(t, err)
	assert.Equal(t, 100*time.Millisecond, *timeout)
	assert.Empty(t, fs.Args())
}

func TestDurationWithVariousFormats(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"1h", time.Hour},
		{"2h30m", 2*time.Hour + 30*time.Minute},
		{"1m", time.Minute},
		{"45s", 45 * time.Second},
		{"500ms", 500 * time.Millisecond},
		{"1us", time.Microsecond},
		{"100ns", 100 * time.Nanosecond},
		{"1h30m45s", time.Hour + 30*time.Minute + 45*time.Second},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			fs := NewFlagSet("test")
			timeout := fs.Duration("timeout", 't', 0, "timeout duration")

			err := fs.Parse([]string{"--timeout", test.input})
			assert.NoError(t, err)
			assert.Equal(t, test.expected, *timeout)
		})
	}
}

func TestDurationDefaultValue(t *testing.T) {
	fs := NewFlagSet("test")
	timeout := fs.Duration("timeout", 't', 5*time.Minute, "timeout duration")

	err := fs.Parse([]string{})
	assert.NoError(t, err)
	assert.Equal(t, 5*time.Minute, *timeout)
	assert.Empty(t, fs.Args())
}

func TestDurationVarMethod(t *testing.T) {
	fs := NewFlagSet("test")

	var timeout time.Duration
	fs.DurationVar(&timeout, "timeout", 't', 10*time.Second, "timeout duration")

	err := fs.Parse([]string{"--timeout", "2m"})
	assert.NoError(t, err)
	assert.Equal(t, 2*time.Minute, timeout)
}

func TestInvalidDurationValue(t *testing.T) {
	fs := NewFlagSet("test")
	fs.Duration("timeout", 't', 0, "timeout duration")

	err := fs.Parse([]string{"--timeout", "invalid"})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidValue)
}

func TestDurationMixedWithOtherFlags(t *testing.T) {
	fs := NewFlagSet("test")
	verbose := fs.Bool("verbose", 'v', false, "verbose output")
	timeout := fs.Duration("timeout", 't', 0, "timeout duration")
	retries := fs.Int("retries", 'r', 3, "number of retries")

	err := fs.Parse([]string{"-v", "--timeout", "30s", "-r", "5", "arg1"})
	assert.NoError(t, err)
	assert.True(t, *verbose)
	assert.Equal(t, 30*time.Second, *timeout)
	assert.Equal(t, 5, *retries)
	assert.Equal(t, []string{"arg1"}, fs.Args())
}

func TestDurationWithNegativeValue(t *testing.T) {
	fs := NewFlagSet("test")
	timeout := fs.Duration("timeout", 't', 0, "timeout duration")

	err := fs.Parse([]string{"--timeout", "-5s"})
	assert.NoError(t, err)
	assert.Equal(t, -5*time.Second, *timeout)
	assert.Empty(t, fs.Args())
}

// Tests for struct-based flag parsing

type SimpleConfig struct {
	Verbose bool   `long:"verbose" short:"v" default:"false" usage:"Enable verbose output"`
	Name    string `long:"name" short:"n" default:"test" usage:"Name to use"`
	Count   int    `long:"count" short:"c" default:"10" usage:"Number of items"`
}

func TestFromStructSimple(t *testing.T) {
	config := &SimpleConfig{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{"--verbose", "--name", "myname", "--count", "42"})
	assert.NoError(t, err)

	assert.True(t, config.Verbose)
	assert.Equal(t, "myname", config.Name)
	assert.Equal(t, 42, config.Count)
}

func TestFromStructWithShortFlags(t *testing.T) {
	config := &SimpleConfig{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{"-v", "-n", "shortname", "-c", "5"})
	assert.NoError(t, err)

	assert.True(t, config.Verbose)
	assert.Equal(t, "shortname", config.Name)
	assert.Equal(t, 5, config.Count)
}

func TestFromStructDefaults(t *testing.T) {
	config := &SimpleConfig{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{})
	assert.NoError(t, err)

	assert.False(t, config.Verbose)
	assert.Equal(t, "test", config.Name)
	assert.Equal(t, 10, config.Count)
}

type AdvancedConfig struct {
	Timeout  time.Duration `long:"timeout" short:"t" default:"30s" usage:"Request timeout"`
	Tags     []string      `long:"tags" short:"T" default:"tag1,tag2" usage:"Tags to apply"`
	Enabled  bool          `long:"enabled" short:"e" usage:"Enable feature"`
	LogLevel string        `long:"log-level" short:"l" default:"info" usage:"Log level"`
}

func TestFromStructAdvanced(t *testing.T) {
	config := &AdvancedConfig{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{
		"--timeout", "1m",
		"--tags", "foo,bar,baz",
		"--enabled",
		"--log-level", "debug",
	})
	assert.NoError(t, err)

	assert.Equal(t, time.Minute, config.Timeout)
	assert.Equal(t, []string{"foo", "bar", "baz"}, config.Tags)
	assert.True(t, config.Enabled)
	assert.Equal(t, "debug", config.LogLevel)
}

func TestFromStructAdvancedDefaults(t *testing.T) {
	config := &AdvancedConfig{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{})
	assert.NoError(t, err)

	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.Equal(t, []string{"tag1", "tag2"}, config.Tags)
	assert.False(t, config.Enabled)
	assert.Equal(t, "info", config.LogLevel)
}

type NoTagsConfig struct {
	Verbose bool
	Name    string
	Count   int
}

func TestFromStructNoTags(t *testing.T) {
	config := &NoTagsConfig{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	// Field names are automatically lowercased for long names
	err = fs.Parse([]string{"--verbose", "--name", "auto", "--count", "7"})
	assert.NoError(t, err)

	assert.True(t, config.Verbose)
	assert.Equal(t, "auto", config.Name)
	assert.Equal(t, 7, config.Count)
}

type MixedConfig struct {
	unexported string        // Should be ignored
	Public     string        `long:"public" short:"p"`
	NoTag      int           // Uses field name as flag
	Duration   time.Duration `long:"duration" short:"d" default:"5s"`
}

func TestFromStructMixed(t *testing.T) {
	config := &MixedConfig{
		unexported: "should-not-change",
	}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{"--public", "test", "--notag", "99", "--duration", "10s"})
	assert.NoError(t, err)

	assert.Equal(t, "should-not-change", config.unexported) // Unexported field unchanged
	assert.Equal(t, "test", config.Public)
	assert.Equal(t, 99, config.NoTag)
	assert.Equal(t, 10*time.Second, config.Duration)
}

func TestParseStruct(t *testing.T) {
	config := &SimpleConfig{}

	err := ParseStruct(config, []string{"--verbose", "--name", "quick", "--count", "3"})
	assert.NoError(t, err)

	assert.True(t, config.Verbose)
	assert.Equal(t, "quick", config.Name)
	assert.Equal(t, 3, config.Count)
}

func TestFromStructErrors(t *testing.T) {
	fs := NewFlagSet("test")

	// Test with nil pointer
	err := fs.FromStruct(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-nil pointer")

	// Test with non-pointer
	config := SimpleConfig{}
	err = fs.FromStruct(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-nil pointer")

	// Test with pointer to non-struct
	str := "not a struct"
	err = fs.FromStruct(&str)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pointer to a struct")
}

type CombinedUsageConfig struct {
	Verbose bool          `long:"verbose" short:"v"`
	Files   []string      `long:"files" short:"f"`
	Timeout time.Duration `long:"timeout" short:"t" default:"1m"`
}

func TestFromStructWithArgs(t *testing.T) {
	config := &CombinedUsageConfig{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{"-v", "arg1", "--files", "a.txt,b.txt", "arg2", "--timeout", "5s", "arg3"})
	assert.NoError(t, err)

	assert.True(t, config.Verbose)
	assert.Equal(t, []string{"a.txt", "b.txt"}, config.Files)
	assert.Equal(t, 5*time.Second, config.Timeout)
	assert.Equal(t, []string{"arg1", "arg2", "arg3"}, fs.Args())
}

type ConfigWithRest struct {
	Verbose bool     `long:"verbose" short:"v" usage:"Enable verbose mode"`
	Output  string   `long:"output" short:"o" default:"stdout" usage:"Output file"`
	Files   []string `rest:"true" usage:"Input files"`
}

func TestRestTag(t *testing.T) {
	config := &ConfigWithRest{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{"--verbose", "file1.txt", "--output", "result.txt", "file2.txt", "file3.txt"})
	assert.NoError(t, err)

	assert.True(t, config.Verbose)
	assert.Equal(t, "result.txt", config.Output)
	assert.Equal(t, []string{"file1.txt", "file2.txt", "file3.txt"}, config.Files)
}

func TestRestTagWithDoubleHyphen(t *testing.T) {
	config := &ConfigWithRest{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{"-v", "file1.txt", "--", "--output", "file2.txt", "-v"})
	assert.NoError(t, err)

	assert.True(t, config.Verbose)
	assert.Equal(t, "stdout", config.Output) // Default value, since --output is after --
	assert.Equal(t, []string{"file1.txt", "--output", "file2.txt", "-v"}, config.Files)
}

func TestRestTagEmpty(t *testing.T) {
	config := &ConfigWithRest{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{"--verbose", "--output", "out.txt"})
	assert.NoError(t, err)

	assert.True(t, config.Verbose)
	assert.Equal(t, "out.txt", config.Output)
	assert.Empty(t, config.Files)
}

type ConfigWithMultipleFields struct {
	Name  string   `long:"name" short:"n"`
	Count int      `long:"count" short:"c" default:"1"`
	Tags  []string `long:"tags" short:"t"`
	Rest  []string `rest:"true"`
}

func TestRestWithOtherArrayField(t *testing.T) {
	config := &ConfigWithMultipleFields{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{"--name", "test", "arg1", "--tags", "a,b", "arg2", "--count", "5", "arg3"})
	assert.NoError(t, err)

	assert.Equal(t, "test", config.Name)
	assert.Equal(t, 5, config.Count)
	assert.Equal(t, []string{"a", "b"}, config.Tags)
	assert.Equal(t, []string{"arg1", "arg2", "arg3"}, config.Rest)
}

type ConfigOnlyRest struct {
	Arguments []string `rest:"true"`
}

func TestOnlyRestField(t *testing.T) {
	config := &ConfigOnlyRest{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{"arg1", "arg2", "arg3"})
	assert.NoError(t, err)

	assert.Equal(t, []string{"arg1", "arg2", "arg3"}, config.Arguments)
}

func TestParseStructWithRest(t *testing.T) {
	config := &ConfigWithRest{}

	err := ParseStruct(config, []string{"-v", "file1.txt", "-o", "output.txt", "file2.txt"})
	assert.NoError(t, err)

	assert.True(t, config.Verbose)
	assert.Equal(t, "output.txt", config.Output)
	assert.Equal(t, []string{"file1.txt", "file2.txt"}, config.Files)
}

type InvalidRestConfig struct {
	RestField string `rest:"true"` // Invalid: not a []string
}

func TestInvalidRestFieldType(t *testing.T) {
	config := &InvalidRestConfig{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err) // Should not error, just ignore the invalid rest field

	err = fs.Parse([]string{"arg1", "arg2"})
	assert.NoError(t, err)

	// The rest field should be ignored since it's not []string
	assert.Equal(t, "", config.RestField)
	assert.Equal(t, []string{"arg1", "arg2"}, fs.Args())
}

// Tests for position tag

type ConfigWithPosition struct {
	Command string `position:"0"`
	Target  string `position:"1"`
	Count   int    `position:"2"`
	Verbose bool   `long:"verbose" short:"v"`
}

func TestPositionTag(t *testing.T) {
	config := &ConfigWithPosition{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{"build", "main.go", "5", "--verbose"})
	assert.NoError(t, err)

	assert.Equal(t, "build", config.Command)
	assert.Equal(t, "main.go", config.Target)
	assert.Equal(t, 5, config.Count)
	assert.True(t, config.Verbose)
}

func TestPositionWithFlags(t *testing.T) {
	config := &ConfigWithPosition{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{"-v", "test", "file.txt", "10"})
	assert.NoError(t, err)

	assert.Equal(t, "test", config.Command)
	assert.Equal(t, "file.txt", config.Target)
	assert.Equal(t, 10, config.Count)
	assert.True(t, config.Verbose)
}

func TestPositionMissing(t *testing.T) {
	config := &ConfigWithPosition{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	// Only provide first position
	err = fs.Parse([]string{"run", "--verbose"})
	assert.NoError(t, err)

	assert.Equal(t, "run", config.Command)
	assert.Equal(t, "", config.Target) // Missing position gets zero value
	assert.Equal(t, 0, config.Count)   // Missing position gets zero value
	assert.True(t, config.Verbose)
}

type ConfigWithGaps struct {
	First  string `position:"0"`
	Third  string `position:"2"`
	Second string `position:"1"`
}

func TestPositionWithGaps(t *testing.T) {
	config := &ConfigWithGaps{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{"one", "two", "three"})
	assert.NoError(t, err)

	assert.Equal(t, "one", config.First)
	assert.Equal(t, "two", config.Second)
	assert.Equal(t, "three", config.Third)
}

type ConfigWithRestAndPosition struct {
	Command string   `position:"0"`
	Output  string   `long:"output" short:"o"`
	Files   []string `rest:"true"`
}

func TestPositionWithRest(t *testing.T) {
	config := &ConfigWithRestAndPosition{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{"compile", "file1.go", "file2.go", "--output", "out.bin", "file3.go"})
	assert.NoError(t, err)

	assert.Equal(t, "compile", config.Command)
	assert.Equal(t, "out.bin", config.Output)
	// Rest should include all non-flag args including the one at position 0
	assert.Equal(t, []string{"compile", "file1.go", "file2.go", "file3.go"}, config.Files)
}

type ConfigWithTypes struct {
	Name     string        `position:"0"`
	Count    int           `position:"1"`
	Ratio    float64       `position:"2"`
	Enabled  bool          `position:"3"`
	Duration time.Duration `position:"4"`
}

func TestPositionWithVariousTypes(t *testing.T) {
	config := &ConfigWithTypes{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{"test", "42", "3.14", "true", "5s"})
	assert.NoError(t, err)

	assert.Equal(t, "test", config.Name)
	assert.Equal(t, 42, config.Count)
	assert.Equal(t, 3.14, config.Ratio)
	assert.True(t, config.Enabled)
	assert.Equal(t, 5*time.Second, config.Duration)
}

func TestPositionInvalidValue(t *testing.T) {
	config := &ConfigWithPosition{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	// "notanumber" cannot be parsed as int for position 2
	err = fs.Parse([]string{"build", "main.go", "notanumber"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid value for position 2")
}

type ConfigWithHighPosition struct {
	Item string `position:"10"`
}

func TestPositionHigherThanArgs(t *testing.T) {
	config := &ConfigWithHighPosition{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{"one", "two", "three"})
	assert.NoError(t, err)

	// Position 10 doesn't exist, so field remains at zero value
	assert.Equal(t, "", config.Item)
}

func TestPositionAfterDoubleHyphen(t *testing.T) {
	config := &ConfigWithPosition{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{"--", "cmd", "target", "3", "--verbose"})
	assert.NoError(t, err)

	assert.Equal(t, "cmd", config.Command)
	assert.Equal(t, "target", config.Target)
	assert.Equal(t, 3, config.Count)
	assert.False(t, config.Verbose) // --verbose is after --, so not parsed as flag
}

type ConfigInvalidPosition struct {
	Item string `position:"invalid"`
}

func TestInvalidPositionTag(t *testing.T) {
	config := &ConfigInvalidPosition{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err) // Should not error, just ignore invalid position

	err = fs.Parse([]string{"value"})
	assert.NoError(t, err)

	assert.Equal(t, "", config.Item) // Field is ignored due to invalid position
}

type ConfigNegativePosition struct {
	Item string `position:"-1"`
}

func TestNegativePositionTag(t *testing.T) {
	config := &ConfigNegativePosition{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err) // Should not error, just ignore negative position

	err = fs.Parse([]string{"value"})
	assert.NoError(t, err)

	assert.Equal(t, "", config.Item) // Field is ignored due to negative position
}

type DatabaseConfig struct {
	Host string `long:"db-host" default:"localhost" usage:"Database host"`
	Port int    `long:"db-port" default:"5432" usage:"Database port"`
}

type ServerConfig struct {
	Port    int  `long:"server-port" short:"p" default:"8080" usage:"Server port"`
	Verbose bool `long:"verbose" short:"v" usage:"Enable verbose logging"`
}

type EmbeddedConfig struct {
	DatabaseConfig
	ServerConfig
	AppName string `long:"app-name" short:"a" default:"myapp" usage:"Application name"`
}

func TestFromStructEmbedded(t *testing.T) {
	config := &EmbeddedConfig{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{
		"--db-host", "db.example.com",
		"--db-port", "3306",
		"--server-port", "9000",
		"--verbose",
		"--app-name", "testapp",
	})
	assert.NoError(t, err)

	assert.Equal(t, "db.example.com", config.Host)
	assert.Equal(t, 3306, config.DatabaseConfig.Port)
	assert.Equal(t, 9000, config.ServerConfig.Port)
	assert.True(t, config.Verbose)
	assert.Equal(t, "testapp", config.AppName)
}

func TestFromStructEmbeddedDefaults(t *testing.T) {
	config := &EmbeddedConfig{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{})
	assert.NoError(t, err)

	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 5432, config.DatabaseConfig.Port)
	assert.Equal(t, 8080, config.ServerConfig.Port)
	assert.False(t, config.Verbose)
	assert.Equal(t, "myapp", config.AppName)
}

func TestFromStructEmbeddedShortFlags(t *testing.T) {
	config := &EmbeddedConfig{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{"-p", "7000", "-v", "-a", "shortapp"})
	assert.NoError(t, err)

	assert.Equal(t, 7000, config.ServerConfig.Port)
	assert.True(t, config.Verbose)
	assert.Equal(t, "shortapp", config.AppName)
}

// Tests for AllowUnknownFlags feature

func TestAllowUnknownFlagsLong(t *testing.T) {
	fs := NewFlagSet("test")
	verbose := fs.Bool("verbose", 'v', false, "verbose output")

	fs.AllowUnknownFlags(true)

	err := fs.Parse([]string{"--verbose", "--unknown", "--another"})
	assert.NoError(t, err)
	assert.True(t, *verbose)
	assert.Equal(t, []string{"--unknown", "--another"}, fs.UnknownFlags())
	assert.Empty(t, fs.Args())
}

func TestAllowUnknownFlagsLongWithValue(t *testing.T) {
	fs := NewFlagSet("test")
	name := fs.String("name", 'n', "default", "name to use")

	fs.AllowUnknownFlags(true)

	err := fs.Parse([]string{"--name", "test", "--unknown", "value"})
	assert.NoError(t, err)
	assert.Equal(t, "test", *name)
	assert.Equal(t, []string{"--unknown", "value"}, fs.UnknownFlags())
	assert.Empty(t, fs.Args())
}

func TestAllowUnknownFlagsLongWithEquals(t *testing.T) {
	fs := NewFlagSet("test")
	verbose := fs.Bool("verbose", 'v', false, "verbose output")

	fs.AllowUnknownFlags(true)

	err := fs.Parse([]string{"--verbose", "--unknown=value", "--another=test"})
	assert.NoError(t, err)
	assert.True(t, *verbose)
	assert.Equal(t, []string{"--unknown=value", "--another=test"}, fs.UnknownFlags())
	assert.Empty(t, fs.Args())
}

func TestAllowUnknownFlagsShort(t *testing.T) {
	fs := NewFlagSet("test")
	verbose := fs.Bool("verbose", 'v', false, "verbose output")

	fs.AllowUnknownFlags(true)

	err := fs.Parse([]string{"-v", "-x", "-y"})
	assert.NoError(t, err)
	assert.True(t, *verbose)
	assert.Equal(t, []string{"-x", "-y"}, fs.UnknownFlags())
	assert.Empty(t, fs.Args())
}

func TestAllowUnknownFlagsShortWithValue(t *testing.T) {
	fs := NewFlagSet("test")
	verbose := fs.Bool("verbose", 'v', false, "verbose output")

	fs.AllowUnknownFlags(true)

	err := fs.Parse([]string{"-v", "-x", "value1", "-y", "value2"})
	assert.NoError(t, err)
	assert.True(t, *verbose)
	assert.Equal(t, []string{"-x", "value1", "-y", "value2"}, fs.UnknownFlags())
	assert.Empty(t, fs.Args())
}

func TestAllowUnknownFlagsShortWithImmediateValue(t *testing.T) {
	fs := NewFlagSet("test")
	verbose := fs.Bool("verbose", 'v', false, "verbose output")

	fs.AllowUnknownFlags(true)

	err := fs.Parse([]string{"-v", "-xvalue"})
	assert.NoError(t, err)
	assert.True(t, *verbose)
	assert.Equal(t, []string{"-xvalue"}, fs.UnknownFlags())
	assert.Empty(t, fs.Args())
}

func TestAllowUnknownFlagsMixed(t *testing.T) {
	fs := NewFlagSet("test")
	verbose := fs.Bool("verbose", 'v', false, "verbose output")
	name := fs.String("name", 'n', "default", "name to use")

	fs.AllowUnknownFlags(true)

	err := fs.Parse([]string{"-v", "--name", "test", "--unknown1", "arg1", "-x", "--unknown2=val", "arg2"})
	assert.NoError(t, err)
	assert.True(t, *verbose)
	assert.Equal(t, "test", *name)
	// Once unknown flag is encountered, everything after is accumulated
	assert.Equal(t, []string{"--unknown1", "arg1", "-x", "--unknown2=val", "arg2"}, fs.UnknownFlags())
	assert.Empty(t, fs.Args())
}

func TestAllowUnknownFlagsDisabled(t *testing.T) {
	fs := NewFlagSet("test")
	fs.Bool("verbose", 'v', false, "verbose output")

	// Default behavior - should error on unknown flags
	fs.AllowUnknownFlags(false)

	err := fs.Parse([]string{"--unknown"})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUnknownFlag)
	assert.Empty(t, fs.UnknownFlags())
}

func TestAllowUnknownFlagsWithArgs(t *testing.T) {
	fs := NewFlagSet("test")
	verbose := fs.Bool("verbose", 'v', false, "verbose output")

	fs.AllowUnknownFlags(true)

	err := fs.Parse([]string{"--verbose", "--unknown", "arg1", "arg2"})
	assert.NoError(t, err)
	assert.True(t, *verbose)
	// Once unknown flag is encountered, everything after is accumulated
	assert.Equal(t, []string{"--unknown", "arg1", "arg2"}, fs.UnknownFlags())
	assert.Empty(t, fs.Args())
}

func TestAllowUnknownFlagsAfterDoubleHyphen(t *testing.T) {
	fs := NewFlagSet("test")
	verbose := fs.Bool("verbose", 'v', false, "verbose output")

	fs.AllowUnknownFlags(true)

	err := fs.Parse([]string{"--verbose", "--unknown", "--", "--another-unknown"})
	assert.NoError(t, err)
	assert.True(t, *verbose)
	// Once unknown flag is encountered, everything after is accumulated (including --)
	assert.Equal(t, []string{"--unknown", "--", "--another-unknown"}, fs.UnknownFlags())
	assert.Empty(t, fs.Args())
}

func TestAllowUnknownFlagsMultipleParseCalls(t *testing.T) {
	fs := NewFlagSet("test")
	verbose := fs.Bool("verbose", 'v', false, "verbose output")

	fs.AllowUnknownFlags(true)

	// First parse
	err := fs.Parse([]string{"--verbose", "--unknown1"})
	assert.NoError(t, err)
	assert.True(t, *verbose)
	assert.Equal(t, []string{"--unknown1"}, fs.UnknownFlags())

	// Second parse - unknownFlags should be reset
	err = fs.Parse([]string{"--unknown2", "--unknown3"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"--unknown2", "--unknown3"}, fs.UnknownFlags())
}

func TestAllowUnknownFlagsEmpty(t *testing.T) {
	fs := NewFlagSet("test")
	verbose := fs.Bool("verbose", 'v', false, "verbose output")

	fs.AllowUnknownFlags(true)

	err := fs.Parse([]string{"--verbose"})
	assert.NoError(t, err)
	assert.True(t, *verbose)
	assert.Empty(t, fs.UnknownFlags())
}

func TestAllowUnknownFlagsWithPositional(t *testing.T) {
	type Config struct {
		Command string `position:"0"`
		Verbose bool   `long:"verbose" short:"v"`
	}

	config := &Config{}
	fs := NewFlagSet("test")
	fs.AllowUnknownFlags(true)

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{"run", "--verbose", "--unknown", "value"})
	assert.NoError(t, err)
	assert.Equal(t, "run", config.Command)
	assert.True(t, config.Verbose)
	assert.Equal(t, []string{"--unknown", "value"}, fs.UnknownFlags())
}

func TestAllowUnknownFlagsWithRest(t *testing.T) {
	type Config struct {
		Verbose bool     `long:"verbose" short:"v"`
		Files   []string `rest:"true"`
	}

	config := &Config{}
	fs := NewFlagSet("test")
	fs.AllowUnknownFlags(true)

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{"--verbose", "--unknown", "file1.txt", "file2.txt"})
	assert.NoError(t, err)
	assert.True(t, config.Verbose)
	// Once unknown flag is encountered, everything after is accumulated
	assert.Equal(t, []string{"--unknown", "file1.txt", "file2.txt"}, fs.UnknownFlags())
	assert.Empty(t, config.Files)
}

// Tests for struct-based unknown flag handling

func TestStructUnknownTag(t *testing.T) {
	type Config struct {
		Verbose      bool     `long:"verbose" short:"v"`
		Name         string   `long:"name" short:"n"`
		UnknownFlags []string `unknown:"true"`
	}

	config := &Config{}
	err := ParseStruct(config, []string{"--verbose", "--name", "test", "--unknown1", "value", "-x"})
	assert.NoError(t, err)
	assert.True(t, config.Verbose)
	assert.Equal(t, "test", config.Name)
	assert.Equal(t, []string{"--unknown1", "value", "-x"}, config.UnknownFlags)
}

func TestStructUnknownTagEmpty(t *testing.T) {
	type Config struct {
		Verbose      bool     `long:"verbose" short:"v"`
		UnknownFlags []string `unknown:"true"`
	}

	config := &Config{}
	err := ParseStruct(config, []string{"--verbose"})
	assert.NoError(t, err)
	assert.True(t, config.Verbose)
	assert.Empty(t, config.UnknownFlags)
}

func TestStructUnknownTagWithPositional(t *testing.T) {
	type Config struct {
		Command      string   `position:"0"`
		Verbose      bool     `long:"verbose" short:"v"`
		UnknownFlags []string `unknown:"true"`
	}

	config := &Config{}
	err := ParseStruct(config, []string{"run", "--verbose", "--unknown", "value"})
	assert.NoError(t, err)
	assert.Equal(t, "run", config.Command)
	assert.True(t, config.Verbose)
	assert.Equal(t, []string{"--unknown", "value"}, config.UnknownFlags)
}

func TestStructUnknownTagWithRest(t *testing.T) {
	type Config struct {
		Verbose      bool     `long:"verbose" short:"v"`
		Files        []string `rest:"true"`
		UnknownFlags []string `unknown:"true"`
	}

	config := &Config{}
	err := ParseStruct(config, []string{"--verbose", "--unknown", "file1.txt", "file2.txt"})
	assert.NoError(t, err)
	assert.True(t, config.Verbose)
	assert.Equal(t, []string{"--unknown", "file1.txt", "file2.txt"}, config.UnknownFlags)
	// Rest field is empty because everything after --unknown goes to unknown flags
	assert.Empty(t, config.Files)
}

func TestStructUnknownTagMultipleUnknownFlags(t *testing.T) {
	type Config struct {
		Debug        bool     `long:"debug" short:"d"`
		UnknownFlags []string `unknown:"true"`
	}

	config := &Config{}
	err := ParseStruct(config, []string{"--debug", "--unknown1", "--unknown2=val", "-x", "arg"})
	assert.NoError(t, err)
	assert.True(t, config.Debug)
	assert.Equal(t, []string{"--unknown1", "--unknown2=val", "-x", "arg"}, config.UnknownFlags)
}

func TestStructUnknownTagOnlyUnknownFlags(t *testing.T) {
	type Config struct {
		UnknownFlags []string `unknown:"true"`
	}

	config := &Config{}
	err := ParseStruct(config, []string{"--unknown1", "value", "--unknown2"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"--unknown1", "value", "--unknown2"}, config.UnknownFlags)
}

func TestStructUnknownTagWithDoubleHyphen(t *testing.T) {
	type Config struct {
		Verbose      bool     `long:"verbose" short:"v"`
		UnknownFlags []string `unknown:"true"`
	}

	config := &Config{}
	err := ParseStruct(config, []string{"--verbose", "--unknown", "--", "arg1", "arg2"})
	assert.NoError(t, err)
	assert.True(t, config.Verbose)
	assert.Equal(t, []string{"--unknown", "--", "arg1", "arg2"}, config.UnknownFlags)
}

func TestStructInvalidUnknownFieldType(t *testing.T) {
	type Config struct {
		Verbose      bool   `long:"verbose" short:"v"`
		UnknownFlags string `unknown:"true"` // Invalid: not []string
	}

	config := &Config{}
	fs := NewFlagSet("test")
	err := fs.FromStruct(config)
	assert.NoError(t, err) // Should not error, just ignore the invalid unknown field

	// allowUnknownFlags should NOT be set because the field type is wrong
	err = fs.Parse([]string{"--verbose", "--unknown"})
	assert.Error(t, err) // Should error because unknown flag handling is not enabled
	assert.ErrorIs(t, err, ErrUnknownFlag)
}

func TestStructUnknownTagWithEmbedded(t *testing.T) {
	type BaseConfig struct {
		Verbose bool `long:"verbose" short:"v"`
	}

	type ExtendedConfig struct {
		BaseConfig
		Name         string   `long:"name" short:"n"`
		UnknownFlags []string `unknown:"true"`
	}

	config := &ExtendedConfig{}
	err := ParseStruct(config, []string{"-v", "--name", "test", "--unknown", "value"})
	assert.NoError(t, err)
	assert.True(t, config.Verbose)
	assert.Equal(t, "test", config.Name)
	assert.Equal(t, []string{"--unknown", "value"}, config.UnknownFlags)
}

func TestStructUnknownTagBeforeKnownFlags(t *testing.T) {
	type Config struct {
		Name         string   `long:"name" short:"n"`
		UnknownFlags []string `unknown:"true"`
	}

	config := &Config{}
	err := ParseStruct(config, []string{"--unknown", "value", "--name", "test"})
	assert.NoError(t, err)
	assert.Equal(t, "", config.Name) // name flag is after unknown, so not processed
	assert.Equal(t, []string{"--unknown", "value", "--name", "test"}, config.UnknownFlags)
}
