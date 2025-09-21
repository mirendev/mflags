package mflags

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringPos(t *testing.T) {
	fs := NewFlagSet("test")

	// Define positional arguments
	cmd := fs.StringPos("command", 0, "build", "The command to run")
	target := fs.StringPos("target", 1, "", "The target file")

	// Parse arguments
	err := fs.Parse([]string{"test", "main.go"})
	require.NoError(t, err)

	assert.Equal(t, "test", *cmd)
	assert.Equal(t, "main.go", *target)
}

func TestStringPosVar(t *testing.T) {
	fs := NewFlagSet("test")

	var cmd, target string
	fs.StringPosVar(&cmd, "command", 0, "build", "The command to run")
	fs.StringPosVar(&target, "target", 1, "", "The target file")

	err := fs.Parse([]string{"compile", "app.go"})
	require.NoError(t, err)

	assert.Equal(t, "compile", cmd)
	assert.Equal(t, "app.go", target)
}

func TestIntPos(t *testing.T) {
	fs := NewFlagSet("test")

	count := fs.IntPos("count", 0, 1, "Number of items")
	port := fs.IntPos("port", 1, 8080, "Port number")

	err := fs.Parse([]string{"5", "3000"})
	require.NoError(t, err)

	assert.Equal(t, 5, *count)
	assert.Equal(t, 3000, *port)
}

func TestIntPosVar(t *testing.T) {
	fs := NewFlagSet("test")

	var count, port int
	fs.IntPosVar(&count, "count", 0, 1, "Number of items")
	fs.IntPosVar(&port, "port", 1, 8080, "Port number")

	err := fs.Parse([]string{"10", "9000"})
	require.NoError(t, err)

	assert.Equal(t, 10, count)
	assert.Equal(t, 9000, port)
}

func TestBoolPos(t *testing.T) {
	fs := NewFlagSet("test")

	enabled := fs.BoolPos("enabled", 0, false, "Enable feature")
	verbose := fs.BoolPos("verbose", 1, false, "Verbose output")

	err := fs.Parse([]string{"true", "false"})
	require.NoError(t, err)

	assert.True(t, *enabled)
	assert.False(t, *verbose)
}

func TestBoolPosVar(t *testing.T) {
	fs := NewFlagSet("test")

	var enabled, verbose bool
	fs.BoolPosVar(&enabled, "enabled", 0, false, "Enable feature")
	fs.BoolPosVar(&verbose, "verbose", 1, true, "Verbose output")

	err := fs.Parse([]string{"false", "true"})
	require.NoError(t, err)

	assert.False(t, enabled)
	assert.True(t, verbose)
}

func TestDurationPos(t *testing.T) {
	fs := NewFlagSet("test")

	timeout := fs.DurationPos("timeout", 0, 30*time.Second, "Timeout duration")
	interval := fs.DurationPos("interval", 1, 1*time.Second, "Check interval")

	err := fs.Parse([]string{"1m", "500ms"})
	require.NoError(t, err)

	assert.Equal(t, 1*time.Minute, *timeout)
	assert.Equal(t, 500*time.Millisecond, *interval)
}

func TestDurationPosVar(t *testing.T) {
	fs := NewFlagSet("test")

	var timeout, interval time.Duration
	fs.DurationPosVar(&timeout, "timeout", 0, 30*time.Second, "Timeout duration")
	fs.DurationPosVar(&interval, "interval", 1, 1*time.Second, "Check interval")

	err := fs.Parse([]string{"2h", "5s"})
	require.NoError(t, err)

	assert.Equal(t, 2*time.Hour, timeout)
	assert.Equal(t, 5*time.Second, interval)
}

func TestRestAPI(t *testing.T) {
	fs := NewFlagSet("test")

	var files []string
	fs.Rest(&files, "Files to process")

	err := fs.Parse([]string{"file1.txt", "file2.txt", "file3.txt"})
	require.NoError(t, err)

	assert.Equal(t, []string{"file1.txt", "file2.txt", "file3.txt"}, files)
}

func TestPositionalWithRest(t *testing.T) {
	fs := NewFlagSet("test")

	cmd := fs.StringPos("command", 0, "", "Command to run")
	var args []string
	fs.Rest(&args, "Command arguments")

	err := fs.Parse([]string{"echo", "hello", "world", "!"})
	require.NoError(t, err)

	assert.Equal(t, "echo", *cmd)
	// Rest includes all args, including positional ones
	assert.Equal(t, []string{"echo", "hello", "world", "!"}, args)
}

func TestPositionalWithFlags(t *testing.T) {
	fs := NewFlagSet("test")

	// Mix of positional arguments and flags
	source := fs.StringPos("source", 0, "", "Source file")
	dest := fs.StringPos("dest", 1, "", "Destination file")
	verbose := fs.Bool("verbose", 'v', false, "Verbose output")
	force := fs.Bool("force", 'f', false, "Force overwrite")

	err := fs.Parse([]string{"input.txt", "output.txt", "--verbose", "-f"})
	require.NoError(t, err)

	assert.Equal(t, "input.txt", *source)
	assert.Equal(t, "output.txt", *dest)
	assert.True(t, *verbose)
	assert.True(t, *force)
}

func TestPositionalDefaultValues(t *testing.T) {
	fs := NewFlagSet("test")

	// Define positional arguments with defaults
	cmd := fs.StringPos("command", 0, "help", "Command to run")
	count := fs.IntPos("count", 1, 10, "Item count")

	// Parse with no arguments - defaults should be used
	err := fs.Parse([]string{})
	require.NoError(t, err)

	assert.Equal(t, "help", *cmd)
	assert.Equal(t, 10, *count)
}

func TestPositionalPartialArgs(t *testing.T) {
	fs := NewFlagSet("test")

	// Define 3 positional arguments
	first := fs.StringPos("first", 0, "default1", "First arg")
	second := fs.StringPos("second", 1, "default2", "Second arg")
	third := fs.StringPos("third", 2, "default3", "Third arg")

	// Only provide 2 arguments
	err := fs.Parse([]string{"arg1", "arg2"})
	require.NoError(t, err)

	assert.Equal(t, "arg1", *first)
	assert.Equal(t, "arg2", *second)
	assert.Equal(t, "default3", *third) // Should keep default
}

func TestPositionalGaps(t *testing.T) {
	fs := NewFlagSet("test")

	// Define positional arguments with gaps (0, 2, 4)
	first := fs.StringPos("first", 0, "", "First arg")
	third := fs.StringPos("third", 2, "", "Third arg")
	fifth := fs.StringPos("fifth", 4, "", "Fifth arg")

	err := fs.Parse([]string{"a", "b", "c", "d", "e"})
	require.NoError(t, err)

	assert.Equal(t, "a", *first)
	assert.Equal(t, "c", *third)
	assert.Equal(t, "e", *fifth)
}

func TestPositionalInvalidValue(t *testing.T) {
	fs := NewFlagSet("test")

	// Define an integer positional argument
	count := fs.IntPos("count", 0, 0, "Count value")

	// Try to parse a non-integer value
	err := fs.Parse([]string{"not-a-number"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid value for position 0")

	// Default should still be there
	assert.Equal(t, 0, *count)
}

func TestRestWithDoubleHyphen(t *testing.T) {
	fs := NewFlagSet("test")

	cmd := fs.StringPos("command", 0, "", "Command")
	var rest []string
	fs.Rest(&rest, "Arguments")
	verbose := fs.Bool("verbose", 'v', false, "Verbose")

	err := fs.Parse([]string{"build", "-v", "--", "--not-a-flag", "-x"})
	require.NoError(t, err)

	assert.Equal(t, "build", *cmd)
	assert.True(t, *verbose)
	// Rest includes positional arg and everything after --
	assert.Equal(t, []string{"build", "--not-a-flag", "-x"}, rest)
}

func TestRestEmpty(t *testing.T) {
	fs := NewFlagSet("test")

	var files []string
	fs.Rest(&files, "Files")

	err := fs.Parse([]string{})
	require.NoError(t, err)

	assert.Empty(t, files)
}

func TestMixedPositionalTypes(t *testing.T) {
	fs := NewFlagSet("test")

	// Mix different types of positional arguments
	cmd := fs.StringPos("command", 0, "", "Command")
	count := fs.IntPos("count", 1, 1, "Count")
	enabled := fs.BoolPos("enabled", 2, false, "Enabled")
	timeout := fs.DurationPos("timeout", 3, 1*time.Second, "Timeout")

	err := fs.Parse([]string{"test", "42", "true", "30s"})
	require.NoError(t, err)

	assert.Equal(t, "test", *cmd)
	assert.Equal(t, 42, *count)
	assert.True(t, *enabled)
	assert.Equal(t, 30*time.Second, *timeout)
}

func TestGetPositionalFieldsAPI(t *testing.T) {
	fs := NewFlagSet("test")

	// Add positional arguments via API
	fs.StringPos("source", 0, "", "Source file")
	fs.StringPos("dest", 1, "", "Destination")
	fs.IntPos("count", 2, 1, "Count")

	// Get positional fields
	fields := fs.GetPositionalFields()
	require.Len(t, fields, 3)

	assert.Equal(t, "source", fields[0].Name)
	assert.Equal(t, "dest", fields[1].Name)
	assert.Equal(t, "count", fields[2].Name)
}

func TestHasPositionalArgsAPI(t *testing.T) {
	fs1 := NewFlagSet("test1")
	assert.False(t, fs1.HasPositionalArgs())

	fs1.StringPos("arg", 0, "", "An argument")
	assert.True(t, fs1.HasPositionalArgs())

	fs2 := NewFlagSet("test2")
	fs2.String("flag", 'f', "", "A flag")
	assert.False(t, fs2.HasPositionalArgs())
}

func TestHasRestArgsAPI(t *testing.T) {
	fs1 := NewFlagSet("test1")
	assert.False(t, fs1.HasRestArgs())

	var rest []string
	fs1.Rest(&rest, "Rest arguments")
	assert.True(t, fs1.HasRestArgs())

	fs2 := NewFlagSet("test2")
	fs2.String("flag", 'f', "", "A flag")
	assert.False(t, fs2.HasRestArgs())
}

func TestPositionalCountAPI(t *testing.T) {
	fs := NewFlagSet("test")
	assert.Equal(t, 0, fs.PositionalCount())

	fs.StringPos("first", 0, "", "First")
	assert.Equal(t, 1, fs.PositionalCount())

	fs.StringPos("third", 2, "", "Third")
	assert.Equal(t, 3, fs.PositionalCount()) // Count includes gaps

	fs.StringPos("second", 1, "", "Second")
	assert.Equal(t, 3, fs.PositionalCount()) // Still 3
}

func TestPositionalBetweenOptions(t *testing.T) {
	fs := NewFlagSet("test")

	// Define flags and positional arguments
	optimize := fs.Bool("optimize", 'o', false, "Enable optimization")
	force := fs.Bool("force", 'f', false, "Force execution")
	verbose := fs.Bool("verbose", 'v', false, "Verbose output")
	command := fs.StringPos("command", 0, "", "Command to run")
	target := fs.StringPos("target", 1, "", "Target file")

	// Test case 1: positional between short flags
	err := fs.Parse([]string{"-o", "run", "-f"})
	require.NoError(t, err)

	assert.True(t, *optimize)
	assert.True(t, *force)
	assert.False(t, *verbose)
	assert.Equal(t, "run", *command)
	assert.Equal(t, "", *target) // Should keep default

	// Reset for next test
	fs = NewFlagSet("test")
	optimize = fs.Bool("optimize", 'o', false, "Enable optimization")
	force = fs.Bool("force", 'f', false, "Force execution")
	verbose = fs.Bool("verbose", 'v', false, "Verbose output")
	command = fs.StringPos("command", 0, "", "Command to run")
	target = fs.StringPos("target", 1, "", "Target file")

	// Test case 2: multiple positionals between flags
	err = fs.Parse([]string{"-o", "build", "main.go", "--verbose", "-f"})
	require.NoError(t, err)

	assert.True(t, *optimize)
	assert.True(t, *force)
	assert.True(t, *verbose)
	assert.Equal(t, "build", *command)
	assert.Equal(t, "main.go", *target)

	// Reset for next test with non-boolean flag
	fs = NewFlagSet("test")
	output := fs.String("output", 'o', "", "Output file")
	force = fs.Bool("force", 'f', false, "Force execution")
	command = fs.StringPos("command", 0, "", "Command to run")
	target = fs.StringPos("target", 1, "", "Target file")

	// Test case 3: positional after non-boolean flag
	err = fs.Parse([]string{"-o", "out.txt", "compile", "-f", "src.go"})
	require.NoError(t, err)

	assert.Equal(t, "out.txt", *output)
	assert.True(t, *force)
	assert.Equal(t, "compile", *command)
	assert.Equal(t, "src.go", *target)

	// Reset for mixed long and short flags
	fs = NewFlagSet("test")
	optimize = fs.Bool("optimize", 'o', false, "Enable optimization")
	force = fs.Bool("force", 'f', false, "Force execution")
	command = fs.StringPos("command", 0, "", "Command to run")
	target = fs.StringPos("target", 1, "", "Target file")

	// Test case 4: positional between long and short flags
	err = fs.Parse([]string{"--optimize", "test", "--force", "file.go"})
	require.NoError(t, err)

	assert.True(t, *optimize)
	assert.True(t, *force)
	assert.Equal(t, "test", *command)
	assert.Equal(t, "file.go", *target)
}

func TestPositionalWithRestBetweenOptions(t *testing.T) {
	fs := NewFlagSet("test")

	// Define flags, positional, and rest arguments
	optimize := fs.Bool("optimize", 'o', false, "Enable optimization")
	force := fs.Bool("force", 'f', false, "Force execution")
	command := fs.StringPos("command", 0, "", "Command to run")
	var rest []string
	fs.Rest(&rest, "Additional arguments")

	err := fs.Parse([]string{"-o", "build", "file1.go", "-f", "file2.go", "file3.go"})
	require.NoError(t, err)

	assert.True(t, *optimize)
	assert.True(t, *force)
	assert.Equal(t, "build", *command)
	// Rest should include all non-flag arguments
	assert.Equal(t, []string{"build", "file1.go", "file2.go", "file3.go"}, rest)
}

func TestDoubleHyphenEndsFlags(t *testing.T) {
	fs := NewFlagSet("test")

	// Define flags and positional arguments
	verbose := fs.Bool("verbose", 'v', false, "Verbose output")
	force := fs.Bool("force", 'f', false, "Force execution")
	command := fs.StringPos("command", 0, "", "Command to run")
	arg1 := fs.StringPos("arg1", 1, "", "First argument")
	arg2 := fs.StringPos("arg2", 2, "", "Second argument")
	arg3 := fs.StringPos("arg3", 3, "", "Third argument")

	// Test case 1: -- makes everything after it positional, even if it looks like flags
	err := fs.Parse([]string{"-v", "--", "-f", "--force", "--verbose", "test"})
	require.NoError(t, err)

	assert.True(t, *verbose)            // -v before -- is parsed as flag
	assert.False(t, *force)             // -f after -- is NOT parsed as flag
	assert.Equal(t, "-f", *command)     // -f becomes positional arg 0
	assert.Equal(t, "--force", *arg1)   // --force becomes positional arg 1
	assert.Equal(t, "--verbose", *arg2) // --verbose becomes positional arg 2
	assert.Equal(t, "test", *arg3)      // test becomes positional arg 3

	// Reset for test case 2
	fs = NewFlagSet("test")
	verbose = fs.Bool("verbose", 'v', false, "Verbose output")
	output := fs.String("output", 'o', "", "Output file")
	command = fs.StringPos("command", 0, "", "Command to run")
	target := fs.StringPos("target", 1, "", "Target")

	// Test case 2: -- with non-boolean flags
	err = fs.Parse([]string{"-v", "--", "-o", "output.txt"})
	require.NoError(t, err)

	assert.True(t, *verbose)
	assert.Equal(t, "", *output) // -o after -- is not parsed as flag
	assert.Equal(t, "-o", *command)
	assert.Equal(t, "output.txt", *target)

	// Reset for test case 3 with rest arguments
	fs = NewFlagSet("test")
	verbose = fs.Bool("verbose", 'v', false, "Verbose output")
	command = fs.StringPos("command", 0, "", "Command to run")
	var rest []string
	fs.Rest(&rest, "Rest arguments")

	// Test case 3: -- with rest arguments
	err = fs.Parse([]string{"-v", "run", "--", "-f", "--test", "arg1", "arg2"})
	require.NoError(t, err)

	assert.True(t, *verbose)
	assert.Equal(t, "run", *command)
	// Rest should include all args after --, treating them as literals
	assert.Equal(t, []string{"run", "-f", "--test", "arg1", "arg2"}, rest)

	// Reset for test case 4: only --
	fs = NewFlagSet("test")
	verbose = fs.Bool("verbose", 'v', false, "Verbose output")

	// Test case 4: just -- with nothing after
	err = fs.Parse([]string{"-v", "--"})
	require.NoError(t, err)

	assert.True(t, *verbose)

	// Reset for test case 5: -- at the beginning
	fs = NewFlagSet("test")
	verbose = fs.Bool("verbose", 'v', false, "Verbose output")
	force = fs.Bool("force", 'f', false, "Force")
	command = fs.StringPos("command", 0, "", "Command to run")
	arg1 = fs.StringPos("arg1", 1, "", "First argument")

	// Test case 5: -- at the beginning treats all as positional
	err = fs.Parse([]string{"--", "-v", "-f"})
	require.NoError(t, err)

	assert.False(t, *verbose) // -v after -- is not parsed as flag
	assert.False(t, *force)   // -f after -- is not parsed as flag
	assert.Equal(t, "-v", *command)
	assert.Equal(t, "-f", *arg1)
}
