package mflags

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"strings"
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
	var rest []string
	fs.Rest(&rest, "args")

	err := fs.Parse([]string{"-v", "arg1", "--name", "test", "arg2", "arg3"})
	assert.NoError(t, err)
	assert.True(t, *verbose)
	assert.Equal(t, "test", *name)
	assert.Equal(t, []string{"arg1", "arg2", "arg3"}, fs.Args())
}

func TestFlagsAfterDoubleHyphen(t *testing.T) {
	fs := NewFlagSet("test")
	verbose := fs.Bool("verbose", 'v', false, "verbose output")
	var rest []string
	fs.Rest(&rest, "args")

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
	var rest []string
	fs.Rest(&rest, "files")

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
	var rest []string
	fs.Rest(&rest, "args")

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
	var rest []string
	fs.Rest(&rest, "args")

	err := fs.Parse([]string{"-v", "--tags", "a,b,c", "--name", "test", "arg1"})
	assert.NoError(t, err)
	assert.True(t, *verbose)
	assert.Equal(t, []string{"a", "b", "c"}, *tags)
	assert.Equal(t, "test", *name)
	assert.Equal(t, []string{"arg1"}, fs.Args())
}

func TestStringArrayRepeatedShortFlag(t *testing.T) {
	fs := NewFlagSet("test")
	tags := fs.StringArray("tags", 't', nil, "tags to apply")

	err := fs.Parse([]string{"-t", "foo", "-t", "bar", "-t", "baz"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"foo", "bar", "baz"}, *tags)
}

func TestStringArrayRepeatedLongFlag(t *testing.T) {
	fs := NewFlagSet("test")
	tags := fs.StringArray("tags", 't', nil, "tags to apply")

	err := fs.Parse([]string{"--tags", "foo", "--tags", "bar", "--tags", "baz"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"foo", "bar", "baz"}, *tags)
}

func TestStringArrayRepeatedMixedWithCommas(t *testing.T) {
	fs := NewFlagSet("test")
	tags := fs.StringArray("tags", 't', nil, "tags to apply")

	// Mix of repeated flags and comma-separated values
	err := fs.Parse([]string{"-t", "foo,bar", "-t", "baz"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"foo", "bar", "baz"}, *tags)
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
	var rest []string
	fs.Rest(&rest, "args")

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

type Int64Config struct {
	Size     int64  `long:"size" short:"s" default:"1000" usage:"Size in bytes"`
	Offset   int64  `long:"offset" short:"o" usage:"Offset position"`
	OptSize  *int64 `long:"opt-size" short:"O" usage:"Optional size"`
}

func TestFromStructInt64(t *testing.T) {
	config := &Int64Config{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{"--size", "9223372036854775807", "--offset", "12345"})
	assert.NoError(t, err)

	assert.Equal(t, int64(9223372036854775807), config.Size)
	assert.Equal(t, int64(12345), config.Offset)
	assert.Nil(t, config.OptSize)
}

func TestFromStructInt64Defaults(t *testing.T) {
	config := &Int64Config{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{})
	assert.NoError(t, err)

	assert.Equal(t, int64(1000), config.Size)
	assert.Equal(t, int64(0), config.Offset)
	assert.Nil(t, config.OptSize)
}

func TestFromStructInt64Pointer(t *testing.T) {
	config := &Int64Config{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{"--opt-size", "42"})
	assert.NoError(t, err)

	assert.NotNil(t, config.OptSize)
	assert.Equal(t, int64(42), *config.OptSize)
}

type AllIntTypesConfig struct {
	Int8Val    int8    `long:"int8" default:"10" usage:"int8 value"`
	Int16Val   int16   `long:"int16" default:"1000" usage:"int16 value"`
	Int32Val   int32   `long:"int32" default:"100000" usage:"int32 value"`
	UintVal    uint    `long:"uint" default:"50" usage:"uint value"`
	Uint8Val   uint8   `long:"uint8" default:"200" usage:"uint8 value"`
	Uint16Val  uint16  `long:"uint16" default:"60000" usage:"uint16 value"`
	Uint32Val  uint32  `long:"uint32" default:"4000000000" usage:"uint32 value"`
	Uint64Val  uint64  `long:"uint64" default:"10000000000" usage:"uint64 value"`
	OptInt8    *int8   `long:"opt-int8" usage:"optional int8"`
	OptUint64  *uint64 `long:"opt-uint64" usage:"optional uint64"`
}

func TestFromStructAllIntTypes(t *testing.T) {
	config := &AllIntTypesConfig{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{
		"--int8", "127",
		"--int16", "32767",
		"--int32", "2147483647",
		"--uint", "100",
		"--uint8", "255",
		"--uint16", "65535",
		"--uint32", "4294967295",
		"--uint64", "18446744073709551615",
	})
	assert.NoError(t, err)

	assert.Equal(t, int8(127), config.Int8Val)
	assert.Equal(t, int16(32767), config.Int16Val)
	assert.Equal(t, int32(2147483647), config.Int32Val)
	assert.Equal(t, uint(100), config.UintVal)
	assert.Equal(t, uint8(255), config.Uint8Val)
	assert.Equal(t, uint16(65535), config.Uint16Val)
	assert.Equal(t, uint32(4294967295), config.Uint32Val)
	assert.Equal(t, uint64(18446744073709551615), config.Uint64Val)
	assert.Nil(t, config.OptInt8)
	assert.Nil(t, config.OptUint64)
}

func TestFromStructAllIntTypesDefaults(t *testing.T) {
	config := &AllIntTypesConfig{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{})
	assert.NoError(t, err)

	assert.Equal(t, int8(10), config.Int8Val)
	assert.Equal(t, int16(1000), config.Int16Val)
	assert.Equal(t, int32(100000), config.Int32Val)
	assert.Equal(t, uint(50), config.UintVal)
	assert.Equal(t, uint8(200), config.Uint8Val)
	assert.Equal(t, uint16(60000), config.Uint16Val)
	assert.Equal(t, uint32(4000000000), config.Uint32Val)
	assert.Equal(t, uint64(10000000000), config.Uint64Val)
}

func TestFromStructIntPointerTypes(t *testing.T) {
	config := &AllIntTypesConfig{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{"--opt-int8", "-128", "--opt-uint64", "999"})
	assert.NoError(t, err)

	assert.NotNil(t, config.OptInt8)
	assert.Equal(t, int8(-128), *config.OptInt8)
	assert.NotNil(t, config.OptUint64)
	assert.Equal(t, uint64(999), *config.OptUint64)
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

func TestFromStructRejectsUnknownTags(t *testing.T) {
	t.Run("single unknown tag", func(t *testing.T) {
		type Opts struct {
			Args struct{} `positional-args:"yes"`
		}
		fs := NewFlagSet("test")
		err := fs.FromStruct(&Opts{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), `unknown struct tag "positional-args"`)
		assert.Contains(t, err.Error(), "Args")
	})

	t.Run("multiple unknown tags across fields", func(t *testing.T) {
		type Opts struct {
			Name string `long:"name" bogus:"yes"`
			Port int    `long:"port" nope:"true"`
		}
		fs := NewFlagSet("test")
		err := fs.FromStruct(&Opts{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), `"bogus"`)
		assert.Contains(t, err.Error(), `"nope"`)
	})

	t.Run("env and required are recognized tags", func(t *testing.T) {
		type Opts struct {
			Name string `long:"name" env:"MY_NAME"`
			Port int    `long:"port" required:"true"`
		}
		fs := NewFlagSet("test")
		err := fs.FromStruct(&Opts{})
		assert.NoError(t, err)
	})

	t.Run("non-mflags tag like json", func(t *testing.T) {
		type Opts struct {
			Name string `json:"name"`
		}
		fs := NewFlagSet("test")
		err := fs.FromStruct(&Opts{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), `"json"`)
	})

	t.Run("embedded struct with unknown tags", func(t *testing.T) {
		type Inner struct {
			Addr string `long:"addr" yaml:"addr"`
		}
		type Outer struct {
			Inner
			Verbose bool `long:"verbose"`
		}
		fs := NewFlagSet("test")
		err := fs.FromStruct(&Outer{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), `"yaml"`)
		assert.Contains(t, err.Error(), "Addr")
	})

	t.Run("unexported fields with arbitrary tags are ignored", func(t *testing.T) {
		type Opts struct {
			Verbose  bool   `long:"verbose"`
			internal string `json:"internal" xml:"internal"`
		}
		fs := NewFlagSet("test")
		err := fs.FromStruct(&Opts{})
		assert.NoError(t, err)
	})

	t.Run("valid tags pass", func(t *testing.T) {
		type Opts struct {
			Name    string   `long:"name" short:"n" default:"world" description:"Your name"`
			Verbose bool     `long:"verbose" usage:"Enable verbose"`
			Env     string   `long:"env" choice:"dev" choice:"prod"`
			File    string   `position:"0"`
			Rest    []string `rest:"true"`
			Unknown []string `unknown:"true"`
		}
		fs := NewFlagSet("test")
		err := fs.FromStruct(&Opts{})
		assert.NoError(t, err)
	})

	t.Run("group tag on blank field is allowed", func(t *testing.T) {
		type Opts struct {
			_       struct{} `group:"Server Options"`
			Verbose bool     `long:"verbose"`
		}
		fs := NewFlagSet("test")
		err := fs.FromStruct(&Opts{})
		assert.NoError(t, err)
	})
}

type CombinedUsageConfig struct {
	Verbose bool          `long:"verbose" short:"v"`
	Files   []string      `long:"files" short:"f"`
	Timeout time.Duration `long:"timeout" short:"t" default:"1m"`
	Rest    []string      `rest:"true"`
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

	// Since the rest field is invalid and ignored, extra args should be rejected
	err = fs.Parse([]string{"arg1", "arg2"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected arguments")

	// The rest field should be ignored since it's not []string
	assert.Equal(t, "", config.RestField)
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
	// Rest should include only non-flag args after positional ones
	assert.Equal(t, []string{"file1.go", "file2.go", "file3.go"}, config.Files)
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

	// With 3 positional args defined (0, 1, 2), any extra args are rejected
	err = fs.Parse([]string{"--", "cmd", "target", "3"})
	assert.NoError(t, err)

	assert.Equal(t, "cmd", config.Command)
	assert.Equal(t, "target", config.Target)
	assert.Equal(t, 3, config.Count)
	assert.False(t, config.Verbose) // Verbose flag not set since no flags parsed
}

type ConfigInvalidPosition struct {
	Item string `position:"invalid"`
}

func TestInvalidPositionTag(t *testing.T) {
	config := &ConfigInvalidPosition{}
	fs := NewFlagSet("test")

	err := fs.FromStruct(config)
	assert.NoError(t, err) // Should not error, just ignore invalid position

	// Since the position tag is invalid, no positional args are defined,
	// so any args should be rejected
	err = fs.Parse([]string{"value"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected arguments")

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

	// Since the position tag is invalid (negative), no positional args are defined,
	// so any args should be rejected
	err = fs.Parse([]string{"value"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected arguments")

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

func TestAutomaticHelpFlag(t *testing.T) {
	// Test -h shows help when not defined
	t.Run("automatic -h", func(t *testing.T) {
		fs := NewFlagSet("myapp")
		fs.String("output", 'o', "a.out", "output file")
		fs.Bool("verbose", 'v', false, "verbose output")

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := fs.Parse([]string{"-h"})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		assert.Equal(t, ErrHelp, err)
		assert.Contains(t, output, "Usage: myapp")
		assert.Contains(t, output, "-o, --output")
		assert.Contains(t, output, "-v, --verbose")
	})

	// Test --help shows help when not defined
	t.Run("automatic --help", func(t *testing.T) {
		fs := NewFlagSet("myapp")
		fs.Int("jobs", 'j', 1, "number of jobs")

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := fs.Parse([]string{"--help"})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		assert.Equal(t, ErrHelp, err)
		assert.Contains(t, output, "Usage: myapp")
		assert.Contains(t, output, "-j, --jobs")
	})

	// Test -h is not treated as help when already defined
	t.Run("-h defined as flag", func(t *testing.T) {
		fs := NewFlagSet("myapp")
		host := fs.String("host", 'h', "localhost", "hostname")

		err := fs.Parse([]string{"-h", "example.com"})

		assert.NoError(t, err)
		assert.Equal(t, "example.com", *host)
	})

	// Test --help is not treated as help when already defined
	t.Run("--help defined as flag", func(t *testing.T) {
		fs := NewFlagSet("myapp")
		helpText := fs.String("help", 0, "", "help text to display")

		err := fs.Parse([]string{"--help", "custom help message"})

		assert.NoError(t, err)
		assert.Equal(t, "custom help message", *helpText)
	})

	// Test help flags after -- are not treated as help
	t.Run("help after --", func(t *testing.T) {
		fs := NewFlagSet("myapp")
		fs.String("output", 'o', "a.out", "output file")
		var rest []string
		fs.Rest(&rest, "args")

		err := fs.Parse([]string{"--", "-h", "--help"})

		assert.NoError(t, err, "Should not show help when -h/--help appear after --")
		assert.Equal(t, []string{"-h", "--help"}, fs.Args())
	})

	// Test specifically that -- stops help detection
	t.Run("-- stops help detection", func(t *testing.T) {
		fs := NewFlagSet("myapp")
		var rest []string
		fs.Rest(&rest, "args")

		// Capture stdout to ensure help is NOT shown
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := fs.Parse([]string{"--", "-h"})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		assert.NoError(t, err, "Should not return ErrHelp when -h appears after --")
		assert.Equal(t, []string{"-h"}, fs.Args())
		assert.NotContains(t, output, "Usage:", "Should not print help text")
	})

	// Test help with mixed flags
	t.Run("help with other flags", func(t *testing.T) {
		fs := NewFlagSet("myapp")
		fs.Bool("verbose", 'v', false, "verbose output")

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := fs.Parse([]string{"-v", "-h"})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)

		assert.Equal(t, ErrHelp, err)
		assert.Contains(t, buf.String(), "Usage:")
	})
}

func TestAutomaticHelpWithUnknownFlags(t *testing.T) {
	// Test that -h alone shows help even with allowUnknownFlags
	t.Run("-h alone with allowUnknownFlags", func(t *testing.T) {
		fs := NewFlagSet("myapp")
		fs.AllowUnknownFlags(true)
		fs.String("output", 'o', "a.out", "output file")

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := fs.Parse([]string{"-h"})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)

		assert.Equal(t, ErrHelp, err)
		assert.Contains(t, buf.String(), "Usage:")
	})

	// Test that -h with other args does NOT show help with allowUnknownFlags
	t.Run("-h with args and allowUnknownFlags", func(t *testing.T) {
		fs := NewFlagSet("myapp")
		fs.AllowUnknownFlags(true)
		fs.String("output", 'o', "a.out", "output file")

		err := fs.Parse([]string{"blah", "-h"})

		assert.NoError(t, err, "Should not show help when other args are present")
		assert.Equal(t, []string{"blah"}, fs.Args())
		assert.Equal(t, []string{"-h"}, fs.UnknownFlags())
	})

	// Test that --help with other args does NOT show help with allowUnknownFlags
	t.Run("--help with args and allowUnknownFlags", func(t *testing.T) {
		fs := NewFlagSet("myapp")
		fs.AllowUnknownFlags(true)

		err := fs.Parse([]string{"command", "--help"})

		assert.NoError(t, err, "Should not show help when other args are present")
		assert.Equal(t, []string{"command"}, fs.Args())
		assert.Equal(t, []string{"--help"}, fs.UnknownFlags())
	})

	// Test the user's specific case: myapp run blah -h
	t.Run("myapp run blah -h", func(t *testing.T) {
		fs := NewFlagSet("myapp")
		fs.AllowUnknownFlags(true)

		err := fs.Parse([]string{"run", "blah", "-h"})

		assert.NoError(t, err, "Should not show help when other args are present")
		assert.Equal(t, []string{"run", "blah"}, fs.Args())
		assert.Equal(t, []string{"-h"}, fs.UnknownFlags())
	})

	// Test that without allowUnknownFlags, -h still shows help
	t.Run("-h with args but no allowUnknownFlags", func(t *testing.T) {
		fs := NewFlagSet("myapp")
		fs.String("output", 'o', "a.out", "output file")

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := fs.Parse([]string{"blah", "-h"})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)

		assert.Equal(t, ErrHelp, err, "Should show help even with other args when allowUnknownFlags is false")
		assert.Contains(t, buf.String(), "Usage:")
	})
}

func TestShowHelp(t *testing.T) {
	fs := NewFlagSet("testapp")
	fs.String("output", 'o', "a.out", "output file")
	fs.Int("jobs", 'j', 4, "number of parallel jobs")
	fs.Bool("verbose", 'v', false, "verbose output")
	fs.Duration("timeout", 't', 30*time.Second, "operation timeout")

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fs.ShowHelp()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.Contains(t, output, "Usage: testapp [options]")
	assert.Contains(t, output, "Options:")
	assert.Contains(t, output, "-o, --output <string>")
	assert.Contains(t, output, "output file")
	assert.Contains(t, output, "(default: a.out)")
	assert.Contains(t, output, "-j, --jobs <int>")
	assert.Contains(t, output, "number of parallel jobs")
	assert.Contains(t, output, "(default: 4)")
	assert.Contains(t, output, "-v, --verbose")
	assert.Contains(t, output, "verbose output")
	assert.Contains(t, output, "-t, --timeout <duration>")
	assert.Contains(t, output, "operation timeout")
}

func TestShowHelpWithPositionalArguments(t *testing.T) {
	t.Run("positional arguments shown by name with descriptions", func(t *testing.T) {
		type Config struct {
			Verbose     bool   `long:"verbose" short:"v" description:"Enable verbose output"`
			Environment string `position:"0" usage:"Target environment (dev, staging, prod)"`
			Version     string `position:"1" usage:"Version to deploy"`
		}

		fs := NewFlagSet("deploy")
		var cfg Config
		err := fs.FromStruct(&cfg)
		assert.NoError(t, err)

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		fs.ShowHelp()

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		assert.Contains(t, output, "Usage: deploy [options] <environment> <version>")
		assert.NotContains(t, output, "[arguments]")
		assert.Contains(t, output, "Arguments:")
		assert.Contains(t, output, "<environment>")
		assert.Contains(t, output, "Target environment (dev, staging, prod)")
		assert.Contains(t, output, "<version>")
		assert.Contains(t, output, "Version to deploy")
	})

	t.Run("positional arguments with rest field", func(t *testing.T) {
		type Config struct {
			Verbose bool     `long:"verbose" short:"v" description:"Enable verbose output"`
			Command string   `position:"0" usage:"Command to execute"`
			Args    []string `rest:"true"`
		}

		fs := NewFlagSet("runner")
		var cfg Config
		err := fs.FromStruct(&cfg)
		assert.NoError(t, err)

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		fs.ShowHelp()

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		assert.Contains(t, output, "Usage: runner [options] <command> [arguments...]")
		assert.Contains(t, output, "Arguments:")
		assert.Contains(t, output, "Command to execute")
	})

	t.Run("only rest field shows arguments", func(t *testing.T) {
		type Config struct {
			Verbose bool     `long:"verbose" short:"v" description:"Enable verbose output"`
			Args    []string `rest:"true"`
		}

		fs := NewFlagSet("echo")
		var cfg Config
		err := fs.FromStruct(&cfg)
		assert.NoError(t, err)

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		fs.ShowHelp()

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		assert.Contains(t, output, "Usage: echo [options] [arguments...]")
		// No Arguments: section when no positional args have usage
		assert.NotContains(t, output, "Arguments:")
	})

	t.Run("non-contiguous positional arguments", func(t *testing.T) {
		type Config struct {
			First  string `position:"0"`
			Third  string `position:"2"`
			Second string `position:"1"`
		}

		fs := NewFlagSet("app")
		var cfg Config
		err := fs.FromStruct(&cfg)
		assert.NoError(t, err)

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		fs.ShowHelp()

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		// Should show in position order regardless of struct field order
		assert.Contains(t, output, "Usage: app [options] <first> <second> <third>")
		// No Arguments: section when no positional args have usage
		assert.NotContains(t, output, "Arguments:")
	})

	t.Run("positional arguments without usage don't show Arguments section", func(t *testing.T) {
		type Config struct {
			Verbose bool   `long:"verbose" short:"v" description:"Enable verbose output"`
			Name    string `position:"0"`
		}

		fs := NewFlagSet("greet")
		var cfg Config
		err := fs.FromStruct(&cfg)
		assert.NoError(t, err)

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		fs.ShowHelp()

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		assert.Contains(t, output, "Usage: greet [options] <name>")
		// No Arguments: section when no positional args have usage
		assert.NotContains(t, output, "Arguments:")
	})
}

// Tests for description tag alias for usage

func TestDescriptionTagAsUsageAlias(t *testing.T) {
	// Test that description tag works as an alias for usage
	t.Run("description tag is used for help output", func(t *testing.T) {
		type Config struct {
			Verbose bool   `long:"verbose" short:"v" description:"Enable verbose output"`
			Name    string `long:"name" short:"n" description:"Name to use"`
		}

		config := &Config{}
		fs := NewFlagSet("myapp")

		err := fs.FromStruct(config)
		assert.NoError(t, err)

		// Check that the flags have the correct usage text
		verboseFlag := fs.Lookup("verbose")
		assert.NotNil(t, verboseFlag)
		assert.Equal(t, "Enable verbose output", verboseFlag.Usage)

		nameFlag := fs.Lookup("name")
		assert.NotNil(t, nameFlag)
		assert.Equal(t, "Name to use", nameFlag.Usage)
	})

	// Test that usage tag takes precedence over description
	t.Run("usage tag takes precedence over description", func(t *testing.T) {
		type Config struct {
			Verbose bool `long:"verbose" usage:"Usage text" description:"Description text"`
		}

		config := &Config{}
		fs := NewFlagSet("myapp")

		err := fs.FromStruct(config)
		assert.NoError(t, err)

		verboseFlag := fs.Lookup("verbose")
		assert.NotNil(t, verboseFlag)
		assert.Equal(t, "Usage text", verboseFlag.Usage)
	})

	// Test that description appears in help output
	t.Run("description appears in help output", func(t *testing.T) {
		type Config struct {
			Output  string `long:"output" short:"o" description:"Output file path"`
			Verbose bool   `long:"verbose" short:"v" description:"Enable verbose logging"`
		}

		config := &Config{}
		fs := NewFlagSet("myapp")

		err := fs.FromStruct(config)
		assert.NoError(t, err)

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		fs.ShowHelp()

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		assert.Contains(t, output, "Output file path")
		assert.Contains(t, output, "Enable verbose logging")
	})

	// Test description with various field types
	t.Run("description works with all field types", func(t *testing.T) {
		type Config struct {
			BoolField     bool          `long:"bool" description:"Bool description"`
			StringField   string        `long:"string" description:"String description"`
			IntField      int           `long:"int" description:"Int description"`
			DurationField time.Duration `long:"duration" description:"Duration description"`
			ArrayField    []string      `long:"array" description:"Array description"`
		}

		config := &Config{}
		fs := NewFlagSet("myapp")

		err := fs.FromStruct(config)
		assert.NoError(t, err)

		assert.Equal(t, "Bool description", fs.Lookup("bool").Usage)
		assert.Equal(t, "String description", fs.Lookup("string").Usage)
		assert.Equal(t, "Int description", fs.Lookup("int").Usage)
		assert.Equal(t, "Duration description", fs.Lookup("duration").Usage)
		assert.Equal(t, "Array description", fs.Lookup("array").Usage)
	})

	// Test that default usage is generated when neither tag is present
	t.Run("default usage when no tags", func(t *testing.T) {
		type Config struct {
			SomeField string `long:"somefield"`
		}

		config := &Config{}
		fs := NewFlagSet("myapp")

		err := fs.FromStruct(config)
		assert.NoError(t, err)

		flag := fs.Lookup("somefield")
		assert.NotNil(t, flag)
		assert.Equal(t, "SomeField value", flag.Usage)
	})

	// Test description with embedded structs
	t.Run("description works with embedded structs", func(t *testing.T) {
		type BaseConfig struct {
			Debug bool `long:"debug" description:"Enable debug mode"`
		}

		type AppConfig struct {
			BaseConfig
			Name string `long:"name" description:"Application name"`
		}

		config := &AppConfig{}
		fs := NewFlagSet("myapp")

		err := fs.FromStruct(config)
		assert.NoError(t, err)

		debugFlag := fs.Lookup("debug")
		assert.NotNil(t, debugFlag)
		assert.Equal(t, "Enable debug mode", debugFlag.Usage)

		nameFlag := fs.Lookup("name")
		assert.NotNil(t, nameFlag)
		assert.Equal(t, "Application name", nameFlag.Usage)
	})
}

// Tests for []bool (repeatable bool flags)

func TestBoolArrayFlag(t *testing.T) {
	// Test basic repeated flag usage
	t.Run("repeated long flag", func(t *testing.T) {
		fs := NewFlagSet("test")
		verbose := fs.BoolArray("verbose", 'v', "verbosity level")

		err := fs.Parse([]string{"--verbose", "--verbose", "--verbose"})
		assert.NoError(t, err)
		assert.Equal(t, []bool{true, true, true}, *verbose)
	})

	// Test repeated short flag
	t.Run("repeated short flag", func(t *testing.T) {
		fs := NewFlagSet("test")
		verbose := fs.BoolArray("verbose", 'v', "verbosity level")

		err := fs.Parse([]string{"-v", "-v", "-v"})
		assert.NoError(t, err)
		assert.Equal(t, []bool{true, true, true}, *verbose)
	})

	// Test combined short flags
	t.Run("combined short flags", func(t *testing.T) {
		fs := NewFlagSet("test")
		verbose := fs.BoolArray("verbose", 'v', "verbosity level")

		err := fs.Parse([]string{"-vvv"})
		assert.NoError(t, err)
		assert.Equal(t, []bool{true, true, true}, *verbose)
	})

	// Test mixed long and short
	t.Run("mixed long and short", func(t *testing.T) {
		fs := NewFlagSet("test")
		verbose := fs.BoolArray("verbose", 'v', "verbosity level")

		err := fs.Parse([]string{"-v", "--verbose", "-v"})
		assert.NoError(t, err)
		assert.Equal(t, []bool{true, true, true}, *verbose)
	})

	// Test no occurrences
	t.Run("no occurrences", func(t *testing.T) {
		fs := NewFlagSet("test")
		verbose := fs.BoolArray("verbose", 'v', "verbosity level")

		err := fs.Parse([]string{})
		assert.NoError(t, err)
		assert.Empty(t, *verbose)
	})

	// Test single occurrence
	t.Run("single occurrence", func(t *testing.T) {
		fs := NewFlagSet("test")
		verbose := fs.BoolArray("verbose", 'v', "verbosity level")

		err := fs.Parse([]string{"-v"})
		assert.NoError(t, err)
		assert.Equal(t, []bool{true}, *verbose)
	})

	// Test with explicit value
	t.Run("explicit true value", func(t *testing.T) {
		fs := NewFlagSet("test")
		verbose := fs.BoolArray("verbose", 'v', "verbosity level")

		err := fs.Parse([]string{"--verbose=true", "--verbose=true"})
		assert.NoError(t, err)
		assert.Equal(t, []bool{true, true}, *verbose)
	})

	// Test with false value
	t.Run("explicit false value", func(t *testing.T) {
		fs := NewFlagSet("test")
		verbose := fs.BoolArray("verbose", 'v', "verbosity level")

		err := fs.Parse([]string{"--verbose=false", "--verbose=true"})
		assert.NoError(t, err)
		assert.Equal(t, []bool{false, true}, *verbose)
	})

	// Test BoolArrayVar
	t.Run("BoolArrayVar", func(t *testing.T) {
		fs := NewFlagSet("test")
		var verbose []bool
		fs.BoolArrayVar(&verbose, "verbose", 'v', "verbosity level")

		err := fs.Parse([]string{"-v", "-v"})
		assert.NoError(t, err)
		assert.Equal(t, []bool{true, true}, verbose)
	})

	// Test with other flags
	t.Run("with other flags", func(t *testing.T) {
		fs := NewFlagSet("test")
		verbose := fs.BoolArray("verbose", 'v', "verbosity level")
		name := fs.String("name", 'n', "default", "name to use")
		debug := fs.Bool("debug", 'd', false, "debug mode")

		err := fs.Parse([]string{"-v", "--name", "test", "-v", "-d", "-v"})
		assert.NoError(t, err)
		assert.Equal(t, []bool{true, true, true}, *verbose)
		assert.Equal(t, "test", *name)
		assert.True(t, *debug)
	})
}

func TestBoolArrayFromStruct(t *testing.T) {
	// Test basic struct usage
	t.Run("basic struct usage", func(t *testing.T) {
		type Config struct {
			Verbose []bool `long:"verbose" short:"v" description:"verbosity level"`
		}

		config := &Config{}
		fs := NewFlagSet("test")

		err := fs.FromStruct(config)
		assert.NoError(t, err)

		err = fs.Parse([]string{"-v", "-v", "-v"})
		assert.NoError(t, err)
		assert.Equal(t, []bool{true, true, true}, config.Verbose)
	})

	// Test combined short flags with struct
	t.Run("combined short flags with struct", func(t *testing.T) {
		type Config struct {
			Verbose []bool `long:"verbose" short:"v" description:"verbosity level"`
		}

		config := &Config{}
		err := ParseStruct(config, []string{"-vvvv"})
		assert.NoError(t, err)
		assert.Equal(t, []bool{true, true, true, true}, config.Verbose)
	})

	// Test with other fields
	t.Run("with other fields", func(t *testing.T) {
		type Config struct {
			Verbose []bool `long:"verbose" short:"v" description:"verbosity level"`
			Name    string `long:"name" short:"n" description:"name to use"`
			Debug   bool   `long:"debug" short:"d" description:"debug mode"`
		}

		config := &Config{}
		err := ParseStruct(config, []string{"-v", "--name", "test", "-vv", "-d"})
		assert.NoError(t, err)
		assert.Equal(t, []bool{true, true, true}, config.Verbose)
		assert.Equal(t, "test", config.Name)
		assert.True(t, config.Debug)
	})

	// Test no occurrences
	t.Run("no occurrences", func(t *testing.T) {
		type Config struct {
			Verbose []bool `long:"verbose" short:"v" description:"verbosity level"`
			Name    string `long:"name" short:"n" default:"default" description:"name to use"`
		}

		config := &Config{}
		err := ParseStruct(config, []string{"--name", "test"})
		assert.NoError(t, err)
		assert.Empty(t, config.Verbose)
		assert.Equal(t, "test", config.Name)
	})

	// Test counting pattern (common use case)
	t.Run("counting pattern", func(t *testing.T) {
		type Config struct {
			Verbose []bool `long:"verbose" short:"v" description:"increase verbosity"`
		}

		config := &Config{}
		err := ParseStruct(config, []string{"-vvvvv"})
		assert.NoError(t, err)
		// Can use len() to get verbosity level
		assert.Equal(t, 5, len(config.Verbose))
	})

	// Test in help output
	t.Run("appears in help output", func(t *testing.T) {
		type Config struct {
			Verbose []bool `long:"verbose" short:"v" description:"increase verbosity level"`
		}

		config := &Config{}
		fs := NewFlagSet("myapp")
		err := fs.FromStruct(config)
		assert.NoError(t, err)

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		fs.ShowHelp()

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		assert.Contains(t, output, "-v, --verbose")
		assert.Contains(t, output, "increase verbosity level")
	})
}

// Tests for []int (repeatable int flags)

func TestIntArrayFlag(t *testing.T) {
	// Test basic repeated flag usage
	t.Run("repeated long flag", func(t *testing.T) {
		fs := NewFlagSet("test")
		nums := fs.IntArray("num", 'n', "numbers to use")

		err := fs.Parse([]string{"--num", "1", "--num", "2", "--num", "3"})
		assert.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3}, *nums)
	})

	// Test repeated short flag
	t.Run("repeated short flag", func(t *testing.T) {
		fs := NewFlagSet("test")
		nums := fs.IntArray("num", 'n', "numbers to use")

		err := fs.Parse([]string{"-n", "10", "-n", "20", "-n", "30"})
		assert.NoError(t, err)
		assert.Equal(t, []int{10, 20, 30}, *nums)
	})

	// Test with equals syntax
	t.Run("equals syntax", func(t *testing.T) {
		fs := NewFlagSet("test")
		nums := fs.IntArray("num", 'n', "numbers to use")

		err := fs.Parse([]string{"--num=100", "--num=200"})
		assert.NoError(t, err)
		assert.Equal(t, []int{100, 200}, *nums)
	})

	// Test no occurrences
	t.Run("no occurrences", func(t *testing.T) {
		fs := NewFlagSet("test")
		nums := fs.IntArray("num", 'n', "numbers to use")

		err := fs.Parse([]string{})
		assert.NoError(t, err)
		assert.Empty(t, *nums)
	})

	// Test single occurrence
	t.Run("single occurrence", func(t *testing.T) {
		fs := NewFlagSet("test")
		nums := fs.IntArray("num", 'n', "numbers to use")

		err := fs.Parse([]string{"-n", "42"})
		assert.NoError(t, err)
		assert.Equal(t, []int{42}, *nums)
	})

	// Test with negative numbers
	t.Run("negative numbers", func(t *testing.T) {
		fs := NewFlagSet("test")
		nums := fs.IntArray("num", 'n', "numbers to use")

		err := fs.Parse([]string{"--num=-5", "--num", "10", "--num=-15"})
		assert.NoError(t, err)
		assert.Equal(t, []int{-5, 10, -15}, *nums)
	})

	// Test IntArrayVar
	t.Run("IntArrayVar", func(t *testing.T) {
		fs := NewFlagSet("test")
		var nums []int
		fs.IntArrayVar(&nums, "num", 'n', "numbers to use")

		err := fs.Parse([]string{"-n", "1", "-n", "2"})
		assert.NoError(t, err)
		assert.Equal(t, []int{1, 2}, nums)
	})

	// Test with other flags
	t.Run("with other flags", func(t *testing.T) {
		fs := NewFlagSet("test")
		nums := fs.IntArray("num", 'n', "numbers to use")
		name := fs.String("name", 0, "default", "name to use")
		verbose := fs.Bool("verbose", 'v', false, "verbose mode")

		err := fs.Parse([]string{"-n", "1", "--name", "test", "-n", "2", "-v", "-n", "3"})
		assert.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3}, *nums)
		assert.Equal(t, "test", *name)
		assert.True(t, *verbose)
	})

	// Test invalid value
	t.Run("invalid value", func(t *testing.T) {
		fs := NewFlagSet("test")
		fs.IntArray("num", 'n', "numbers to use")

		err := fs.Parse([]string{"-n", "notanumber"})
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidValue)
	})

	// Test short flag with immediate value
	t.Run("short flag with immediate value", func(t *testing.T) {
		fs := NewFlagSet("test")
		nums := fs.IntArray("num", 'n', "numbers to use")

		err := fs.Parse([]string{"-n42", "-n99"})
		assert.NoError(t, err)
		assert.Equal(t, []int{42, 99}, *nums)
	})
}

func TestIntArrayFromStruct(t *testing.T) {
	// Test basic struct usage
	t.Run("basic struct usage", func(t *testing.T) {
		type Config struct {
			Numbers []int `long:"num" short:"n" description:"numbers to process"`
		}

		config := &Config{}
		fs := NewFlagSet("test")

		err := fs.FromStruct(config)
		assert.NoError(t, err)

		err = fs.Parse([]string{"-n", "1", "-n", "2", "-n", "3"})
		assert.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3}, config.Numbers)
	})

	// Test with ParseStruct
	t.Run("with ParseStruct", func(t *testing.T) {
		type Config struct {
			Numbers []int `long:"num" short:"n" description:"numbers to process"`
		}

		config := &Config{}
		err := ParseStruct(config, []string{"--num", "10", "--num", "20", "--num", "30"})
		assert.NoError(t, err)
		assert.Equal(t, []int{10, 20, 30}, config.Numbers)
	})

	// Test with other fields
	t.Run("with other fields", func(t *testing.T) {
		type Config struct {
			Numbers []int  `long:"num" short:"n" description:"numbers to process"`
			Name    string `long:"name" description:"name to use"`
			Debug   bool   `long:"debug" short:"d" description:"debug mode"`
		}

		config := &Config{}
		err := ParseStruct(config, []string{"-n", "1", "--name", "test", "-n", "2", "-d", "-n", "3"})
		assert.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3}, config.Numbers)
		assert.Equal(t, "test", config.Name)
		assert.True(t, config.Debug)
	})

	// Test no occurrences
	t.Run("no occurrences", func(t *testing.T) {
		type Config struct {
			Numbers []int  `long:"num" short:"n" description:"numbers to process"`
			Name    string `long:"name" default:"default" description:"name to use"`
		}

		config := &Config{}
		err := ParseStruct(config, []string{"--name", "test"})
		assert.NoError(t, err)
		assert.Empty(t, config.Numbers)
		assert.Equal(t, "test", config.Name)
	})

	// Test with []bool together
	t.Run("with []bool together", func(t *testing.T) {
		type Config struct {
			Verbose []bool `long:"verbose" short:"v" description:"increase verbosity"`
			Numbers []int  `long:"num" short:"n" description:"numbers to process"`
		}

		config := &Config{}
		err := ParseStruct(config, []string{"-vv", "-n", "1", "-v", "-n", "2"})
		assert.NoError(t, err)
		assert.Equal(t, []bool{true, true, true}, config.Verbose)
		assert.Equal(t, []int{1, 2}, config.Numbers)
	})

	// Test in help output
	t.Run("appears in help output", func(t *testing.T) {
		type Config struct {
			Numbers []int `long:"num" short:"n" description:"numbers to process"`
		}

		config := &Config{}
		fs := NewFlagSet("myapp")
		err := fs.FromStruct(config)
		assert.NoError(t, err)

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		fs.ShowHelp()

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		assert.Contains(t, output, "-n, --num")
		assert.Contains(t, output, "numbers to process")
	})

	// Test computing sum (practical use case)
	t.Run("computing sum", func(t *testing.T) {
		type Config struct {
			Numbers []int `long:"num" short:"n" description:"numbers to add"`
		}

		config := &Config{}
		err := ParseStruct(config, []string{"-n", "10", "-n", "20", "-n", "30"})
		assert.NoError(t, err)

		sum := 0
		for _, n := range config.Numbers {
			sum += n
		}
		assert.Equal(t, 60, sum)
	})
}

// Tests for pointer types in FromStruct

func TestPointerTypesFromStruct(t *testing.T) {
	// Test *bool pointer - nil when not set
	t.Run("*bool nil when not set", func(t *testing.T) {
		type Config struct {
			Verbose *bool `long:"verbose" short:"v" description:"verbose mode"`
		}

		config := &Config{}
		err := ParseStruct(config, []string{})
		assert.NoError(t, err)
		assert.Nil(t, config.Verbose)
	})

	// Test *bool pointer - set to true
	t.Run("*bool set to true", func(t *testing.T) {
		type Config struct {
			Verbose *bool `long:"verbose" short:"v" description:"verbose mode"`
		}

		config := &Config{}
		err := ParseStruct(config, []string{"-v"})
		assert.NoError(t, err)
		assert.NotNil(t, config.Verbose)
		assert.True(t, *config.Verbose)
	})

	// Test *bool pointer - set to false explicitly
	t.Run("*bool set to false explicitly", func(t *testing.T) {
		type Config struct {
			Verbose *bool `long:"verbose" short:"v" description:"verbose mode"`
		}

		config := &Config{}
		err := ParseStruct(config, []string{"--verbose=false"})
		assert.NoError(t, err)
		assert.NotNil(t, config.Verbose)
		assert.False(t, *config.Verbose)
	})

	// Test *string pointer - nil when not set
	t.Run("*string nil when not set", func(t *testing.T) {
		type Config struct {
			Name *string `long:"name" short:"n" description:"name to use"`
		}

		config := &Config{}
		err := ParseStruct(config, []string{})
		assert.NoError(t, err)
		assert.Nil(t, config.Name)
	})

	// Test *string pointer - set to value
	t.Run("*string set to value", func(t *testing.T) {
		type Config struct {
			Name *string `long:"name" short:"n" description:"name to use"`
		}

		config := &Config{}
		err := ParseStruct(config, []string{"--name", "Alice"})
		assert.NoError(t, err)
		assert.NotNil(t, config.Name)
		assert.Equal(t, "Alice", *config.Name)
	})

	// Test *string pointer - set to empty string (distinguishable from nil)
	t.Run("*string set to empty string", func(t *testing.T) {
		type Config struct {
			Name *string `long:"name" short:"n" description:"name to use"`
		}

		config := &Config{}
		err := ParseStruct(config, []string{"--name", ""})
		assert.NoError(t, err)
		assert.NotNil(t, config.Name)
		assert.Equal(t, "", *config.Name)
	})

	// Test *int pointer - nil when not set
	t.Run("*int nil when not set", func(t *testing.T) {
		type Config struct {
			Count *int `long:"count" short:"c" description:"count value"`
		}

		config := &Config{}
		err := ParseStruct(config, []string{})
		assert.NoError(t, err)
		assert.Nil(t, config.Count)
	})

	// Test *int pointer - set to value
	t.Run("*int set to value", func(t *testing.T) {
		type Config struct {
			Count *int `long:"count" short:"c" description:"count value"`
		}

		config := &Config{}
		err := ParseStruct(config, []string{"--count", "42"})
		assert.NoError(t, err)
		assert.NotNil(t, config.Count)
		assert.Equal(t, 42, *config.Count)
	})

	// Test *int pointer - set to zero (distinguishable from nil)
	t.Run("*int set to zero", func(t *testing.T) {
		type Config struct {
			Count *int `long:"count" short:"c" description:"count value"`
		}

		config := &Config{}
		err := ParseStruct(config, []string{"--count", "0"})
		assert.NoError(t, err)
		assert.NotNil(t, config.Count)
		assert.Equal(t, 0, *config.Count)
	})

	// Test *time.Duration pointer - nil when not set
	t.Run("*time.Duration nil when not set", func(t *testing.T) {
		type Config struct {
			Timeout *time.Duration `long:"timeout" short:"t" description:"timeout duration"`
		}

		config := &Config{}
		err := ParseStruct(config, []string{})
		assert.NoError(t, err)
		assert.Nil(t, config.Timeout)
	})

	// Test *time.Duration pointer - set to value
	t.Run("*time.Duration set to value", func(t *testing.T) {
		type Config struct {
			Timeout *time.Duration `long:"timeout" short:"t" description:"timeout duration"`
		}

		config := &Config{}
		err := ParseStruct(config, []string{"--timeout", "5s"})
		assert.NoError(t, err)
		assert.NotNil(t, config.Timeout)
		assert.Equal(t, 5*time.Second, *config.Timeout)
	})

	// Test *time.Duration pointer - set to zero (distinguishable from nil)
	t.Run("*time.Duration set to zero", func(t *testing.T) {
		type Config struct {
			Timeout *time.Duration `long:"timeout" short:"t" description:"timeout duration"`
		}

		config := &Config{}
		err := ParseStruct(config, []string{"--timeout", "0s"})
		assert.NoError(t, err)
		assert.NotNil(t, config.Timeout)
		assert.Equal(t, time.Duration(0), *config.Timeout)
	})

	// Test mixed pointer and non-pointer types
	t.Run("mixed pointer and non-pointer types", func(t *testing.T) {
		type Config struct {
			Name     string  `long:"name" description:"required name"`
			Verbose  bool    `long:"verbose" short:"v" description:"verbose mode"`
			OptCount *int    `long:"count" short:"c" description:"optional count"`
			OptName  *string `long:"opt-name" description:"optional name"`
		}

		config := &Config{}
		err := ParseStruct(config, []string{"--name", "test", "-v", "--count", "10"})
		assert.NoError(t, err)
		assert.Equal(t, "test", config.Name)
		assert.True(t, config.Verbose)
		assert.NotNil(t, config.OptCount)
		assert.Equal(t, 10, *config.OptCount)
		assert.Nil(t, config.OptName) // Not set
	})

	// Test practical use case - conditional logic based on nil
	t.Run("practical use case - conditional logic", func(t *testing.T) {
		type Config struct {
			Port    *int           `long:"port" short:"p" description:"port number"`
			Host    *string        `long:"host" short:"h" description:"hostname"`
			Timeout *time.Duration `long:"timeout" description:"connection timeout"`
		}

		config := &Config{}
		err := ParseStruct(config, []string{"--port", "8080"})
		assert.NoError(t, err)

		// Simulate real-world usage
		actualPort := 3000 // default
		if config.Port != nil {
			actualPort = *config.Port
		}
		assert.Equal(t, 8080, actualPort)

		actualHost := "localhost" // default
		if config.Host != nil {
			actualHost = *config.Host
		}
		assert.Equal(t, "localhost", actualHost) // Host was nil, use default

		actualTimeout := 30 * time.Second // default
		if config.Timeout != nil {
			actualTimeout = *config.Timeout
		}
		assert.Equal(t, 30*time.Second, actualTimeout) // Timeout was nil, use default
	})

	// Test all pointer types with short flags
	t.Run("all pointer types with short flags", func(t *testing.T) {
		type Config struct {
			BoolPtr     *bool          `short:"b" description:"bool pointer"`
			StringPtr   *string        `short:"s" description:"string pointer"`
			IntPtr      *int           `short:"i" description:"int pointer"`
			DurationPtr *time.Duration `short:"d" description:"duration pointer"`
		}

		config := &Config{}
		err := ParseStruct(config, []string{"-b", "-s", "hello", "-i", "99", "-d", "1h"})
		assert.NoError(t, err)
		assert.NotNil(t, config.BoolPtr)
		assert.True(t, *config.BoolPtr)
		assert.NotNil(t, config.StringPtr)
		assert.Equal(t, "hello", *config.StringPtr)
		assert.NotNil(t, config.IntPtr)
		assert.Equal(t, 99, *config.IntPtr)
		assert.NotNil(t, config.DurationPtr)
		assert.Equal(t, time.Hour, *config.DurationPtr)
	})

	// Test pointer types appear in help output
	t.Run("pointer types appear in help output", func(t *testing.T) {
		type Config struct {
			Name    *string        `long:"name" short:"n" description:"optional name"`
			Count   *int           `long:"count" short:"c" description:"optional count"`
			Verbose *bool          `long:"verbose" short:"v" description:"verbose mode"`
			Timeout *time.Duration `long:"timeout" short:"t" description:"timeout"`
		}

		config := &Config{}
		fs := NewFlagSet("myapp")
		err := fs.FromStruct(config)
		assert.NoError(t, err)

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		fs.ShowHelp()

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		assert.Contains(t, output, "-n, --name")
		assert.Contains(t, output, "optional name")
		assert.Contains(t, output, "-c, --count")
		assert.Contains(t, output, "optional count")
		assert.Contains(t, output, "-v, --verbose")
		assert.Contains(t, output, "-t, --timeout")
	})

	// Test invalid values for pointer types
	t.Run("invalid int value", func(t *testing.T) {
		type Config struct {
			Count *int `long:"count" description:"count"`
		}

		config := &Config{}
		err := ParseStruct(config, []string{"--count", "notanumber"})
		assert.Error(t, err)
	})

	t.Run("invalid duration value", func(t *testing.T) {
		type Config struct {
			Timeout *time.Duration `long:"timeout" description:"timeout"`
		}

		config := &Config{}
		err := ParseStruct(config, []string{"--timeout", "notaduration"})
		assert.Error(t, err)
	})

	t.Run("invalid bool value", func(t *testing.T) {
		type Config struct {
			Verbose *bool `long:"verbose" description:"verbose"`
		}

		config := &Config{}
		err := ParseStruct(config, []string{"--verbose=notabool"})
		assert.Error(t, err)
	})
}

func TestShortOnlyFlags(t *testing.T) {
	// Test that short-only flags work in parsing
	t.Run("parse short-only flag", func(t *testing.T) {
		fs := NewFlagSet("myapp")
		verbose := fs.Bool("", 'v', false, "verbose output")
		debug := fs.Bool("", 'd', false, "debug mode")

		err := fs.Parse([]string{"-v", "-d"})

		assert.NoError(t, err)
		assert.True(t, *verbose)
		assert.True(t, *debug)
	})

	// Test that short-only flags appear in help
	t.Run("short-only flags in help", func(t *testing.T) {
		fs := NewFlagSet("myapp")
		fs.Bool("", 'v', false, "verbose output")
		fs.String("", 's', "default", "short string flag")

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		fs.ShowHelp()

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		assert.Contains(t, output, "-v")
		assert.Contains(t, output, "verbose output")
		assert.Contains(t, output, "-s <string>")
		assert.Contains(t, output, "short string flag")
	})

	// Test mixed long-only, short-only, and both
	t.Run("mixed flag types in help", func(t *testing.T) {
		fs := NewFlagSet("myapp")
		fs.Bool("verbose", 'v', false, "verbose (both)")
		fs.Bool("quiet", 0, false, "quiet (long-only)")
		fs.Bool("", 'd', false, "debug (short-only)")

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		fs.ShowHelp()

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		// Check all three types appear
		assert.Contains(t, output, "-v, --verbose")
		assert.Contains(t, output, "verbose (both)")
		assert.Contains(t, output, "--quiet")
		assert.Contains(t, output, "quiet (long-only)")
		assert.Contains(t, output, "-d")
		assert.Contains(t, output, "debug (short-only)")
	})

	// Test parsing mixed flag types
	t.Run("parse mixed flag types", func(t *testing.T) {
		fs := NewFlagSet("myapp")
		verbose := fs.Bool("verbose", 'v', false, "verbose (both)")
		quiet := fs.Bool("quiet", 0, false, "quiet (long-only)")
		debug := fs.Bool("", 'd', false, "debug (short-only)")

		err := fs.Parse([]string{"-v", "--quiet", "-d"})

		assert.NoError(t, err)
		assert.True(t, *verbose)
		assert.True(t, *quiet)
		assert.True(t, *debug)
	})
}

// Tests for Choice/enum flags
func TestChoiceFlag(t *testing.T) {
	fs := NewFlagSet("test")
	level := fs.Choice("level", 'l', "", []string{"debug", "info", "warn", "error"}, "log level")

	err := fs.Parse([]string{"--level", "info"})
	assert.NoError(t, err)
	assert.Equal(t, "info", *level)
}

func TestChoiceFlagShort(t *testing.T) {
	fs := NewFlagSet("test")
	level := fs.Choice("level", 'l', "", []string{"debug", "info", "warn", "error"}, "log level")

	err := fs.Parse([]string{"-l", "warn"})
	assert.NoError(t, err)
	assert.Equal(t, "warn", *level)
}

func TestChoiceFlagWithEquals(t *testing.T) {
	fs := NewFlagSet("test")
	level := fs.Choice("level", 'l', "", []string{"debug", "info", "warn", "error"}, "log level")

	err := fs.Parse([]string{"--level=error"})
	assert.NoError(t, err)
	assert.Equal(t, "error", *level)
}

func TestChoiceFlagDefault(t *testing.T) {
	fs := NewFlagSet("test")
	level := fs.Choice("level", 'l', "info", []string{"debug", "info", "warn", "error"}, "log level")

	err := fs.Parse([]string{})
	assert.NoError(t, err)
	assert.Equal(t, "info", *level)
}

func TestChoiceFlagInvalidValue(t *testing.T) {
	fs := NewFlagSet("test")
	_ = fs.Choice("level", 'l', "", []string{"debug", "info", "warn", "error"}, "log level")

	err := fs.Parse([]string{"--level", "trace"})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidValue)
	assert.Contains(t, err.Error(), "invalid choice")
	assert.Contains(t, err.Error(), "trace")
	assert.Contains(t, err.Error(), "debug, info, warn, error")
}

func TestChoiceFlagInvalidValueShort(t *testing.T) {
	fs := NewFlagSet("test")
	_ = fs.Choice("level", 'l', "", []string{"debug", "info", "warn", "error"}, "log level")

	err := fs.Parse([]string{"-l", "invalid"})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidValue)
	assert.Contains(t, err.Error(), "invalid choice")
}

func TestChoiceVar(t *testing.T) {
	fs := NewFlagSet("test")
	var level string
	fs.ChoiceVar(&level, "level", 'l', "warn", []string{"debug", "info", "warn", "error"}, "log level")

	err := fs.Parse([]string{"--level", "debug"})
	assert.NoError(t, err)
	assert.Equal(t, "debug", level)
}

func TestChoiceVarDefault(t *testing.T) {
	fs := NewFlagSet("test")
	var level string
	fs.ChoiceVar(&level, "level", 'l', "warn", []string{"debug", "info", "warn", "error"}, "log level")

	err := fs.Parse([]string{})
	assert.NoError(t, err)
	assert.Equal(t, "warn", level)
}

func TestChoiceFromStruct(t *testing.T) {
	type Config struct {
		Level string `long:"level" short:"l" choice:"debug" choice:"info" choice:"warn" choice:"error" usage:"log level"`
	}

	config := &Config{}
	fs := NewFlagSet("test")
	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{"--level", "info"})
	assert.NoError(t, err)
	assert.Equal(t, "info", config.Level)
}

func TestChoiceFromStructShort(t *testing.T) {
	type Config struct {
		Level string `long:"level" short:"l" choice:"debug" choice:"info" choice:"warn" choice:"error"`
	}

	config := &Config{}
	fs := NewFlagSet("test")
	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{"-l", "warn"})
	assert.NoError(t, err)
	assert.Equal(t, "warn", config.Level)
}

func TestChoiceFromStructWithDefault(t *testing.T) {
	type Config struct {
		Level string `long:"level" short:"l" default:"info" choice:"debug" choice:"info" choice:"warn" choice:"error"`
	}

	config := &Config{}
	fs := NewFlagSet("test")
	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{})
	assert.NoError(t, err)
	assert.Equal(t, "info", config.Level)
}

func TestChoiceFromStructInvalid(t *testing.T) {
	type Config struct {
		Level string `long:"level" short:"l" choice:"debug" choice:"info" choice:"warn" choice:"error"`
	}

	config := &Config{}
	fs := NewFlagSet("test")
	err := fs.FromStruct(config)
	assert.NoError(t, err)

	err = fs.Parse([]string{"--level", "trace"})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidValue)
	assert.Contains(t, err.Error(), "invalid choice")
}

func TestChoiceHelpOutput(t *testing.T) {
	fs := NewFlagSet("myapp")
	_ = fs.Choice("level", 'l', "info", []string{"debug", "info", "warn", "error"}, "log level")

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fs.ShowHelp()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check that choices are displayed in the type placeholder
	assert.Contains(t, output, "debug|info|warn|error")
	assert.Contains(t, output, "-l, --level")
	assert.Contains(t, output, "log level")
	assert.Contains(t, output, "(default: info)")
}

func TestChoiceFromStructHelpOutput(t *testing.T) {
	type Config struct {
		Level string `long:"level" short:"l" default:"warn" choice:"debug" choice:"info" choice:"warn" choice:"error" usage:"log level"`
	}

	config := &Config{}
	fs := NewFlagSet("myapp")
	err := fs.FromStruct(config)
	assert.NoError(t, err)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fs.ShowHelp()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check that choices are displayed
	assert.Contains(t, output, "debug|info|warn|error")
	assert.Contains(t, output, "-l, --level")
	assert.Contains(t, output, "log level")
	assert.Contains(t, output, "(default: warn)")
}

func TestChoiceValueType(t *testing.T) {
	cv := &choiceValue{
		value:   new(string),
		choices: []string{"a", "b", "c"},
	}
	assert.Equal(t, "a|b|c", cv.Type())
	assert.False(t, cv.IsBool())
}

func TestChoiceValueChoices(t *testing.T) {
	cv := &choiceValue{
		value:   new(string),
		choices: []string{"x", "y", "z"},
	}
	assert.Equal(t, []string{"x", "y", "z"}, cv.Choices())
}

func TestGetTagValues(t *testing.T) {
	type Test struct {
		Field string `choice:"a" choice:"b" choice:"c"`
	}

	rt := reflect.TypeOf(Test{})
	field, _ := rt.FieldByName("Field")
	choices := getTagValues(field.Tag, "choice")

	assert.Equal(t, []string{"a", "b", "c"}, choices)
}

func TestGetTagValuesEmpty(t *testing.T) {
	type Test struct {
		Field string `long:"field"`
	}

	rt := reflect.TypeOf(Test{})
	field, _ := rt.FieldByName("Field")
	choices := getTagValues(field.Tag, "choice")

	assert.Empty(t, choices)
}

func TestGetTagValuesMixed(t *testing.T) {
	type Test struct {
		Field string `long:"level" short:"l" choice:"debug" default:"info" choice:"info" usage:"log level" choice:"warn"`
	}

	rt := reflect.TypeOf(Test{})
	field, _ := rt.FieldByName("Field")
	choices := getTagValues(field.Tag, "choice")

	assert.Equal(t, []string{"debug", "info", "warn"}, choices)
}

func TestDuplicateLongFlagPanics(t *testing.T) {
	fs := NewFlagSet("test")
	fs.Bool("verbose", 'v', false, "verbose output")

	assert.PanicsWithValue(t, `flag "verbose" already registered as --verbose`, func() {
		fs.Bool("verbose", 0, false, "duplicate verbose")
	})
}

func TestDuplicateShortFlagPanics(t *testing.T) {
	fs := NewFlagSet("test")
	fs.Bool("verbose", 'v', false, "verbose output")

	assert.PanicsWithValue(t, `short flag 'v' already registered for --verbose`, func() {
		fs.String("version", 'v', "", "show version")
	})
}

// captureStdout captures stdout output from the given function.
func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestFromStructSelfDeclaredGroup(t *testing.T) {
	type GlobalOptions struct {
		_       struct{} `group:"Global Options"`
		Verbose bool     `long:"verbose" short:"v" usage:"Enable verbose output"`
		Debug   bool     `long:"debug" short:"d" usage:"Enable debug mode"`
	}

	fs := NewFlagSet("test")
	var opts GlobalOptions
	err := fs.FromStruct(&opts)
	assert.NoError(t, err)

	// Verify all flags got the group
	fs.VisitAll(func(flag *Flag) {
		assert.Equal(t, "Global Options", flag.Group)
	})
}

func TestFromStructWithGroup(t *testing.T) {
	type Config struct {
		Verbose bool   `long:"verbose" short:"v" usage:"Enable verbose output"`
		Output  string `long:"output" short:"o" usage:"Output file"`
	}

	fs := NewFlagSet("test")
	var cfg Config
	err := fs.FromStruct(&cfg, InGroup("Build Options"))
	assert.NoError(t, err)

	// Verify all flags got the group from InGroup
	fs.VisitAll(func(flag *Flag) {
		assert.Equal(t, "Build Options", flag.Group)
	})
}

func TestFromStructEmbeddedGroupTag(t *testing.T) {
	type GlobalOptions struct {
		Verbose bool `long:"verbose" short:"v" usage:"Enable verbose output"`
	}

	type Config struct {
		GlobalOptions `group:"Global Options"`
		Output        string `long:"output" short:"o" usage:"Output file"`
	}

	fs := NewFlagSet("test")
	var cfg Config
	err := fs.FromStruct(&cfg)
	assert.NoError(t, err)

	verboseFlag := fs.Lookup("verbose")
	assert.NotNil(t, verboseFlag)
	assert.Equal(t, "Global Options", verboseFlag.Group)

	outputFlag := fs.Lookup("output")
	assert.NotNil(t, outputFlag)
	assert.Equal(t, "", outputFlag.Group)
}

func TestGroupPrecedence(t *testing.T) {
	t.Run("InGroup overrides embedded tag", func(t *testing.T) {
		type GlobalOptions struct {
			Verbose bool `long:"verbose" short:"v" usage:"Enable verbose output"`
		}

		type Config struct {
			GlobalOptions `group:"Embedded Name"`
			Output        string `long:"output" short:"o" usage:"Output file"`
		}

		fs := NewFlagSet("test")
		var cfg Config
		err := fs.FromStruct(&cfg, InGroup("Caller Name"))
		assert.NoError(t, err)

		// InGroup should override the embedded tag for the parent struct's flags
		outputFlag := fs.Lookup("output")
		assert.Equal(t, "Caller Name", outputFlag.Group)

		// The embedded struct gets its own group from the embedding site tag
		verboseFlag := fs.Lookup("verbose")
		assert.Equal(t, "Embedded Name", verboseFlag.Group)
	})

	t.Run("embedded tag overrides self-declaration", func(t *testing.T) {
		type GlobalOptions struct {
			_       struct{} `group:"Self Declared"`
			Verbose bool     `long:"verbose" short:"v" usage:"Enable verbose output"`
		}

		type Config struct {
			GlobalOptions `group:"Embedded Override"`
			Output        string `long:"output" short:"o" usage:"Output file"`
		}

		fs := NewFlagSet("test")
		var cfg Config
		err := fs.FromStruct(&cfg)
		assert.NoError(t, err)

		verboseFlag := fs.Lookup("verbose")
		assert.Equal(t, "Embedded Override", verboseFlag.Group)

		outputFlag := fs.Lookup("output")
		assert.Equal(t, "", outputFlag.Group)
	})

	t.Run("InGroup overrides self-declaration", func(t *testing.T) {
		type GlobalOptions struct {
			_       struct{} `group:"Self Declared"`
			Verbose bool     `long:"verbose" short:"v" usage:"Enable verbose output"`
		}

		fs := NewFlagSet("test")
		var opts GlobalOptions
		err := fs.FromStruct(&opts, InGroup("Caller Override"))
		assert.NoError(t, err)

		verboseFlag := fs.Lookup("verbose")
		assert.Equal(t, "Caller Override", verboseFlag.Group)
	})
}

func TestGroupHelpOutput(t *testing.T) {
	type GlobalOptions struct {
		_       struct{} `group:"Global Options"`
		Verbose bool     `long:"verbose" short:"v" usage:"Enable verbose output"`
	}

	type Config struct {
		GlobalOptions
		Output string `long:"output" short:"o" usage:"Output file"`
	}

	fs := NewFlagSet("test")
	var cfg Config
	err := fs.FromStruct(&cfg)
	assert.NoError(t, err)

	output := captureStdout(func() {
		fs.WriteFlagHelp()
	})

	// Named groups come first
	globalIdx := strings.Index(output, "Global Options:")
	// Find standalone "Options:" (preceded by newline, not part of "Global Options:")
	optionsIdx := strings.Index(output, "\nOptions:")
	assert.Greater(t, globalIdx, -1, "should contain Global Options heading")
	assert.Greater(t, optionsIdx, -1, "should contain Options heading")
	assert.Less(t, globalIdx, optionsIdx, "Global Options should appear before Options")

	// Verify flag placement
	verboseIdx := strings.Index(output, "--verbose")
	outputFlagIdx := strings.Index(output, "--output")
	assert.Greater(t, verboseIdx, globalIdx, "verbose should be under Global Options")
	assert.Less(t, verboseIdx, optionsIdx, "verbose should be before Options heading")
	assert.Greater(t, outputFlagIdx, optionsIdx, "output should be under Options")
}

func TestGroupBackwardCompatibility(t *testing.T) {
	// No groups = single "Options:" heading
	fs := NewFlagSet("test")
	fs.Bool("verbose", 'v', false, "verbose output")
	fs.String("output", 'o', "", "output file")

	output := captureStdout(func() {
		fs.WriteFlagHelp()
	})

	assert.Contains(t, output, "Options:")
	assert.Equal(t, 1, strings.Count(output, "Options:"), "should have exactly one Options heading")
	assert.Contains(t, output, "--verbose")
	assert.Contains(t, output, "--output")
}

func TestGroupMethodProgrammatic(t *testing.T) {
	fs := NewFlagSet("test")

	fs.Group("Network")
	fs.String("host", 'H', "localhost", "server host")
	fs.Int("port", 'p', 8080, "server port")

	fs.Group("")
	fs.Bool("verbose", 'v', false, "verbose output")

	hostFlag := fs.Lookup("host")
	assert.Equal(t, "Network", hostFlag.Group)

	portFlag := fs.Lookup("port")
	assert.Equal(t, "Network", portFlag.Group)

	verboseFlag := fs.Lookup("verbose")
	assert.Equal(t, "", verboseFlag.Group)

	output := captureStdout(func() {
		fs.WriteFlagHelp()
	})

	networkIdx := strings.Index(output, "Network:")
	optionsIdx := strings.Index(output, "Options:")
	assert.Greater(t, networkIdx, -1)
	assert.Greater(t, optionsIdx, -1)
	assert.Less(t, networkIdx, optionsIdx)
}

func TestGroupOrderPreserved(t *testing.T) {
	fs := NewFlagSet("test")

	fs.Group("Zebra")
	fs.Bool("z-flag", 'z', false, "z flag")

	fs.Group("Alpha")
	fs.Bool("a-flag", 'a', false, "a flag")

	fs.Group("")
	fs.Bool("verbose", 'v', false, "verbose")

	output := captureStdout(func() {
		fs.WriteFlagHelp()
	})

	zebraIdx := strings.Index(output, "Zebra:")
	alphaIdx := strings.Index(output, "Alpha:")
	optionsIdx := strings.Index(output, "Options:")

	// Insertion order: Zebra first, then Alpha, then default Options last
	assert.Less(t, zebraIdx, alphaIdx, "Zebra should appear before Alpha (insertion order)")
	assert.Less(t, alphaIdx, optionsIdx, "Alpha should appear before Options")
}

func TestShowHelpWithGroups(t *testing.T) {
	type GlobalOptions struct {
		_       struct{} `group:"Global Options"`
		Verbose bool     `long:"verbose" short:"v" usage:"Enable verbose output"`
	}

	type Config struct {
		GlobalOptions
		Output string `long:"output" short:"o" usage:"Output file"`
	}

	fs := NewFlagSet("myapp")
	var cfg Config
	err := fs.FromStruct(&cfg)
	assert.NoError(t, err)

	output := captureStdout(func() {
		fs.ShowHelp()
	})

	assert.Contains(t, output, "Usage: myapp [options]")
	assert.Contains(t, output, "Global Options:")
	assert.Contains(t, output, "Options:")
	assert.Contains(t, output, "-v, --verbose")
	assert.Contains(t, output, "-o, --output")
}

// --- env tag tests ---

func TestFromStructEnvFlag(t *testing.T) {
	type Config struct {
		Name string `long:"name" env:"TEST_MFLAGS_NAME"`
	}

	t.Run("picks up env var", func(t *testing.T) {
		t.Setenv("TEST_MFLAGS_NAME", "from-env")
		config := &Config{}
		fs := NewFlagSet("test")
		err := fs.FromStruct(config)
		assert.NoError(t, err)
		err = fs.Parse([]string{})
		assert.NoError(t, err)
		assert.Equal(t, "from-env", config.Name)
	})

	t.Run("env overrides default", func(t *testing.T) {
		type ConfigWithDefault struct {
			Name string `long:"name" default:"hardcoded" env:"TEST_MFLAGS_NAME"`
		}
		t.Setenv("TEST_MFLAGS_NAME", "from-env")
		config := &ConfigWithDefault{}
		fs := NewFlagSet("test")
		err := fs.FromStruct(config)
		assert.NoError(t, err)
		err = fs.Parse([]string{})
		assert.NoError(t, err)
		assert.Equal(t, "from-env", config.Name)
	})

	t.Run("CLI arg overrides env", func(t *testing.T) {
		t.Setenv("TEST_MFLAGS_NAME", "from-env")
		config := &Config{}
		fs := NewFlagSet("test")
		err := fs.FromStruct(config)
		assert.NoError(t, err)
		err = fs.Parse([]string{"--name", "from-cli"})
		assert.NoError(t, err)
		assert.Equal(t, "from-cli", config.Name)
	})

	t.Run("unset env uses default", func(t *testing.T) {
		type ConfigWithDefault struct {
			Name string `long:"name" default:"hardcoded" env:"TEST_MFLAGS_NAME_UNSET"`
		}
		config := &ConfigWithDefault{}
		fs := NewFlagSet("test")
		err := fs.FromStruct(config)
		assert.NoError(t, err)
		err = fs.Parse([]string{})
		assert.NoError(t, err)
		assert.Equal(t, "hardcoded", config.Name)
	})

	t.Run("env with int flag", func(t *testing.T) {
		type ConfigInt struct {
			Count int `long:"count" env:"TEST_MFLAGS_COUNT"`
		}
		t.Setenv("TEST_MFLAGS_COUNT", "42")
		config := &ConfigInt{}
		fs := NewFlagSet("test")
		err := fs.FromStruct(config)
		assert.NoError(t, err)
		err = fs.Parse([]string{})
		assert.NoError(t, err)
		assert.Equal(t, 42, config.Count)
	})

	t.Run("env with invalid int errors", func(t *testing.T) {
		type ConfigInt struct {
			Count int `long:"count" env:"TEST_MFLAGS_COUNT"`
		}
		t.Setenv("TEST_MFLAGS_COUNT", "abc")
		config := &ConfigInt{}
		fs := NewFlagSet("test")
		err := fs.FromStruct(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "TEST_MFLAGS_COUNT")
	})

	t.Run("env with pointer type", func(t *testing.T) {
		type ConfigPtr struct {
			Name *string `long:"name" env:"TEST_MFLAGS_PTR_NAME"`
		}
		t.Setenv("TEST_MFLAGS_PTR_NAME", "from-env")
		config := &ConfigPtr{}
		fs := NewFlagSet("test")
		err := fs.FromStruct(config)
		assert.NoError(t, err)
		err = fs.Parse([]string{})
		assert.NoError(t, err)
		assert.NotNil(t, config.Name)
		assert.Equal(t, "from-env", *config.Name)
	})

	t.Run("env with bool flag", func(t *testing.T) {
		type ConfigBool struct {
			Verbose bool `long:"verbose" env:"TEST_MFLAGS_VERBOSE"`
		}
		t.Setenv("TEST_MFLAGS_VERBOSE", "true")
		config := &ConfigBool{}
		fs := NewFlagSet("test")
		err := fs.FromStruct(config)
		assert.NoError(t, err)
		err = fs.Parse([]string{})
		assert.NoError(t, err)
		assert.True(t, config.Verbose)
	})
}

func TestFromStructEnvPositional(t *testing.T) {
	type Config struct {
		Target string `position:"0" env:"TEST_MFLAGS_TARGET"`
	}

	t.Run("picks up env var", func(t *testing.T) {
		t.Setenv("TEST_MFLAGS_TARGET", "from-env")
		config := &Config{}
		fs := NewFlagSet("test")
		err := fs.FromStruct(config)
		assert.NoError(t, err)
		err = fs.Parse([]string{})
		assert.NoError(t, err)
		assert.Equal(t, "from-env", config.Target)
	})

	t.Run("CLI arg overrides env", func(t *testing.T) {
		t.Setenv("TEST_MFLAGS_TARGET", "from-env")
		config := &Config{}
		fs := NewFlagSet("test")
		err := fs.FromStruct(config)
		assert.NoError(t, err)
		err = fs.Parse([]string{"from-cli"})
		assert.NoError(t, err)
		assert.Equal(t, "from-cli", config.Target)
	})
}

func TestEnvHelpText(t *testing.T) {
	type Config struct {
		Name string `long:"name" env:"MY_APP_NAME" usage:"Application name"`
	}

	config := &Config{}
	fs := NewFlagSet("test")
	err := fs.FromStruct(config)
	assert.NoError(t, err)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fs.WriteFlagHelp()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.Contains(t, output, "(env: MY_APP_NAME)")
}

// --- required tag tests ---

func TestFromStructRequiredFlag(t *testing.T) {
	type Config struct {
		Name string `long:"name" required:"true"`
	}

	t.Run("errors when not provided", func(t *testing.T) {
		config := &Config{}
		fs := NewFlagSet("test")
		err := fs.FromStruct(config)
		assert.NoError(t, err)
		err = fs.Parse([]string{})
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrRequired)
		assert.Contains(t, err.Error(), "--name")
	})

	t.Run("succeeds when provided via CLI", func(t *testing.T) {
		config := &Config{}
		fs := NewFlagSet("test")
		err := fs.FromStruct(config)
		assert.NoError(t, err)
		err = fs.Parse([]string{"--name", "hello"})
		assert.NoError(t, err)
		assert.Equal(t, "hello", config.Name)
	})

	t.Run("succeeds when provided via env", func(t *testing.T) {
		type ConfigEnv struct {
			Name string `long:"name" required:"true" env:"TEST_MFLAGS_REQ_NAME"`
		}
		t.Setenv("TEST_MFLAGS_REQ_NAME", "from-env")
		config := &ConfigEnv{}
		fs := NewFlagSet("test")
		err := fs.FromStruct(config)
		assert.NoError(t, err)
		err = fs.Parse([]string{})
		assert.NoError(t, err)
		assert.Equal(t, "from-env", config.Name)
	})

	t.Run("multiple missing reports all", func(t *testing.T) {
		type ConfigMulti struct {
			Name string `long:"name" required:"true"`
			Host string `long:"host" required:"true"`
		}
		config := &ConfigMulti{}
		fs := NewFlagSet("test")
		err := fs.FromStruct(config)
		assert.NoError(t, err)
		err = fs.Parse([]string{})
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrRequired)
		assert.Contains(t, err.Error(), "--name")
		assert.Contains(t, err.Error(), "--host")
	})
}

func TestFromStructRequiredPositional(t *testing.T) {
	type Config struct {
		Command string `position:"0" required:"true"`
	}

	t.Run("errors when not provided", func(t *testing.T) {
		config := &Config{}
		fs := NewFlagSet("test")
		err := fs.FromStruct(config)
		assert.NoError(t, err)
		err = fs.Parse([]string{})
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrRequired)
		assert.Contains(t, err.Error(), "Command")
	})

	t.Run("succeeds when provided", func(t *testing.T) {
		config := &Config{}
		fs := NewFlagSet("test")
		err := fs.FromStruct(config)
		assert.NoError(t, err)
		err = fs.Parse([]string{"deploy"})
		assert.NoError(t, err)
		assert.Equal(t, "deploy", config.Command)
	})

	t.Run("succeeds when provided via env", func(t *testing.T) {
		type ConfigEnv struct {
			Command string `position:"0" required:"true" env:"TEST_MFLAGS_CMD"`
		}
		t.Setenv("TEST_MFLAGS_CMD", "from-env")
		config := &ConfigEnv{}
		fs := NewFlagSet("test")
		err := fs.FromStruct(config)
		assert.NoError(t, err)
		err = fs.Parse([]string{})
		assert.NoError(t, err)
		assert.Equal(t, "from-env", config.Command)
	})
}

func TestRequiredBoolFlag(t *testing.T) {
	type Config struct {
		Accept bool `long:"accept" required:"true"`
	}

	t.Run("errors when not provided", func(t *testing.T) {
		config := &Config{}
		fs := NewFlagSet("test")
		err := fs.FromStruct(config)
		assert.NoError(t, err)
		err = fs.Parse([]string{})
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrRequired)
	})

	t.Run("succeeds when provided", func(t *testing.T) {
		config := &Config{}
		fs := NewFlagSet("test")
		err := fs.FromStruct(config)
		assert.NoError(t, err)
		err = fs.Parse([]string{"--accept"})
		assert.NoError(t, err)
		assert.True(t, config.Accept)
	})
}
