package mflags

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDispatcherBasic(t *testing.T) {
	d := NewDispatcher("myapp")

	// Register a simple command
	fs := NewFlagSet("test")
	verbose := fs.Bool("verbose", 'v', false, "verbose output")

	var executed bool
	var capturedArgs []string

	cmd := NewCommand(fs, func(flags *FlagSet, args []string) error {
		executed = true
		capturedArgs = args
		return nil
	})

	d.Dispatch("test", cmd)

	// Execute the command
	err := d.Execute([]string{"test", "--verbose", "arg1", "arg2"})
	assert.NoError(t, err)
	assert.True(t, executed)
	assert.True(t, *verbose)
	assert.Equal(t, []string{"arg1", "arg2"}, capturedArgs)
}

func TestDispatcherNestedCommands(t *testing.T) {
	d := NewDispatcher("myapp")

	// Track which command was executed
	var executedCommand string

	// Register nested commands
	d.Dispatch("foo", NewCommand(NewFlagSet("foo"), func(fs *FlagSet, args []string) error {
		executedCommand = "foo"
		return nil
	}))

	d.Dispatch("foo bar", NewCommand(NewFlagSet("foo bar"), func(fs *FlagSet, args []string) error {
		executedCommand = "foo bar"
		return nil
	}))

	d.Dispatch("foo bar baz", NewCommand(NewFlagSet("foo bar baz"), func(fs *FlagSet, args []string) error {
		executedCommand = "foo bar baz"
		return nil
	}))

	// Test longest match
	err := d.Execute([]string{"foo", "bar", "baz", "arg1"})
	assert.NoError(t, err)
	assert.Equal(t, "foo bar baz", executedCommand)

	// Test partial match
	err = d.Execute([]string{"foo", "bar", "arg1"})
	assert.NoError(t, err)
	assert.Equal(t, "foo bar", executedCommand)

	// Test single command
	err = d.Execute([]string{"foo", "arg1"})
	assert.NoError(t, err)
	assert.Equal(t, "foo", executedCommand)
}

func TestDispatcherWithFlags(t *testing.T) {
	d := NewDispatcher("myapp")

	// Create flagset with various types
	fs := NewFlagSet("build")
	output := fs.String("output", 'o', "a.out", "output file")
	optimize := fs.Bool("optimize", 'O', false, "enable optimization")
	jobs := fs.Int("jobs", 'j', 1, "number of parallel jobs")

	var capturedFlags struct {
		output   string
		optimize bool
		jobs     int
		args     []string
	}

	cmd := NewCommand(fs, func(flags *FlagSet, args []string) error {
		capturedFlags.output = *output
		capturedFlags.optimize = *optimize
		capturedFlags.jobs = *jobs
		capturedFlags.args = args
		return nil
	})

	d.Dispatch("build", cmd)

	// Execute with flags
	err := d.Execute([]string{"build", "-O", "--output", "program", "-j", "4", "main.go"})
	assert.NoError(t, err)
	assert.Equal(t, "program", capturedFlags.output)
	assert.True(t, capturedFlags.optimize)
	assert.Equal(t, 4, capturedFlags.jobs)
	assert.Equal(t, []string{"main.go"}, capturedFlags.args)
}

func TestDispatcherUnknownCommand(t *testing.T) {
	d := NewDispatcher("myapp")

	d.Dispatch("known", NewCommand(NewFlagSet("known"), func(fs *FlagSet, args []string) error {
		return nil
	}))

	err := d.Execute([]string{"unknown", "command"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown command")
}

func TestDispatcherHelp(t *testing.T) {
	d := NewDispatcher("myapp")

	// Register some commands with usage
	d.Dispatch("build", NewCommand(NewFlagSet("build"),
		func(fs *FlagSet, args []string) error { return nil },
		WithUsage("Build the project")))

	d.Dispatch("test", NewCommand(NewFlagSet("test"),
		func(fs *FlagSet, args []string) error { return nil },
		WithUsage("Run tests")))

	d.Dispatch("clean", NewCommand(NewFlagSet("clean"),
		func(fs *FlagSet, args []string) error { return nil }))

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Show general help
	err := d.Execute([]string{"help"})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "Available commands:")
	assert.Contains(t, output, "build")
	assert.Contains(t, output, "Build the project")
	assert.Contains(t, output, "test")
	assert.Contains(t, output, "Run tests")
	assert.Contains(t, output, "clean")
}

func TestDispatcherCommandHelp(t *testing.T) {
	d := NewDispatcher("myapp")

	// Create a command with flags
	fs := NewFlagSet("build")
	fs.String("output", 'o', "a.out", "output file")
	fs.Bool("verbose", 'v', false, "verbose output")

	d.Dispatch("build", NewCommand(fs,
		func(flags *FlagSet, args []string) error { return nil },
		WithUsage("Build the project with the specified options")))

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Show command-specific help
	err := d.Execute([]string{"build", "--help"})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "Usage: myapp build")
	assert.Contains(t, output, "Build the project")
	assert.Contains(t, output, "-o, --output")
	assert.Contains(t, output, "output file")
	assert.Contains(t, output, "-v, --verbose")
	assert.Contains(t, output, "verbose output")
}

func TestDispatcherErrorHandling(t *testing.T) {
	d := NewDispatcher("myapp")

	// Register a command that returns an error
	d.Dispatch("fail", NewCommand(NewFlagSet("fail"), func(fs *FlagSet, args []string) error {
		return fmt.Errorf("command failed")
	}))

	err := d.Execute([]string{"fail"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "command failed")
}

func TestDispatcherFlagParsingError(t *testing.T) {
	d := NewDispatcher("myapp")

	fs := NewFlagSet("test")
	fs.Int("count", 'c', 0, "count value")

	d.Dispatch("test", NewCommand(fs, func(flags *FlagSet, args []string) error {
		return nil
	}))

	// Invalid int value
	err := d.Execute([]string{"test", "--count", "not-a-number"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing flags")
}

func TestDispatcherNormalizeCommandPath(t *testing.T) {
	d := NewDispatcher("myapp")

	var executed bool
	handler := func(fs *FlagSet, args []string) error {
		executed = true
		return nil
	}

	// Register with extra spaces
	d.Dispatch("  foo   bar  ", NewCommand(NewFlagSet("test"), handler))

	// Should work with normalized path
	executed = false
	err := d.Execute([]string{"foo", "bar"})
	assert.NoError(t, err)
	assert.True(t, executed)

	// Should work with different spacing
	executed = false
	err = d.Execute([]string{"foo", "bar"})
	assert.NoError(t, err)
	assert.True(t, executed)
}

func TestDispatcherGetCommand(t *testing.T) {
	d := NewDispatcher("myapp")

	fs := NewFlagSet("test")
	handler := func(fs *FlagSet, args []string) error { return nil }

	d.Dispatch("foo bar", NewCommand(fs, handler, WithUsage("test command")))

	// Get existing command
	cmd := d.GetCommand("foo bar")
	assert.NotNil(t, cmd)
	assert.Equal(t, fs, cmd.FlagSet())

	// Get existing command entry
	entry := d.GetCommandEntry("foo bar")
	assert.NotNil(t, entry)
	assert.Equal(t, "foo bar", entry.Path)
	assert.Equal(t, "test command", entry.Usage)

	// Get non-existing command
	cmd = d.GetCommand("baz")
	assert.Nil(t, cmd)
}

func TestDispatcherHasCommand(t *testing.T) {
	d := NewDispatcher("myapp")

	d.Dispatch("exists", NewCommand(NewFlagSet("test"), func(fs *FlagSet, args []string) error {
		return nil
	}))

	assert.True(t, d.HasCommand("exists"))
	assert.False(t, d.HasCommand("not-exists"))
}

func TestDispatcherRemove(t *testing.T) {
	d := NewDispatcher("myapp")

	d.Dispatch("temp", NewCommand(NewFlagSet("test"), func(fs *FlagSet, args []string) error {
		return nil
	}))

	assert.True(t, d.HasCommand("temp"))

	d.Remove("temp")

	assert.False(t, d.HasCommand("temp"))

	// Should error when trying to execute removed command
	err := d.Execute([]string{"temp"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown command")
}

func TestDispatcherGetCommands(t *testing.T) {
	d := NewDispatcher("myapp")

	d.Dispatch("cmd1", NewCommand(NewFlagSet("test1"), func(fs *FlagSet, args []string) error { return nil }))
	d.Dispatch("cmd2", NewCommand(NewFlagSet("test2"), func(fs *FlagSet, args []string) error { return nil }))
	d.Dispatch("cmd3", NewCommand(NewFlagSet("test3"), func(fs *FlagSet, args []string) error { return nil }))

	commands := d.GetCommands()
	assert.Len(t, commands, 3)
	assert.Contains(t, commands, "cmd1")
	assert.Contains(t, commands, "cmd2")
	assert.Contains(t, commands, "cmd3")
}

func TestDispatcherRunAlias(t *testing.T) {
	d := NewDispatcher("myapp")

	var executed bool
	d.Dispatch("test", NewCommand(NewFlagSet("test"), func(fs *FlagSet, args []string) error {
		executed = true
		return nil
	}))

	// Test Run method (alias for Execute)
	err := d.Run([]string{"test"})
	assert.NoError(t, err)
	assert.True(t, executed)
}

func TestDispatcherEmptyArgs(t *testing.T) {
	d := NewDispatcher("myapp")

	d.Dispatch("test", NewCommand(NewFlagSet("test"), func(fs *FlagSet, args []string) error {
		return nil
	}))

	// Capture stdout for help output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := d.Execute([]string{})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Available commands:")
}

func TestDispatcherWithStructFlags(t *testing.T) {
	d := NewDispatcher("myapp")

	type Config struct {
		Verbose bool     `long:"verbose" short:"v"`
		Output  string   `long:"output" short:"o" default:"out.txt"`
		Files   []string `rest:"true"`
	}

	config := &Config{}
	fs := NewFlagSet("process")
	err := fs.FromStruct(config)
	assert.NoError(t, err)

	d.Dispatch("process", NewCommand(fs, func(flags *FlagSet, args []string) error {
		// Handler can access config directly since it's been parsed
		return nil
	}))

	err = d.Execute([]string{"process", "-v", "--output", "result.txt", "file1.txt", "file2.txt"})
	assert.NoError(t, err)
	assert.True(t, config.Verbose)
	assert.Equal(t, "result.txt", config.Output)
	assert.Equal(t, []string{"file1.txt", "file2.txt"}, config.Files)
}

func TestDispatcherMultiWordCommandWithArgs(t *testing.T) {
	d := NewDispatcher("myapp")

	var capturedArgs []string
	d.Dispatch("foo bar baz", NewCommand(NewFlagSet("test"), func(fs *FlagSet, args []string) error {
		capturedArgs = args
		return nil
	}))

	err := d.Execute([]string{"foo", "bar", "baz", "arg1", "arg2", "arg3"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"arg1", "arg2", "arg3"}, capturedArgs)
}

func TestDispatcherGetCommandCompletions(t *testing.T) {
	d := NewDispatcher("myapp")

	d.Dispatch("build", NewCommand(NewFlagSet("build"),
		func(fs *FlagSet, args []string) error { return nil },
		WithUsage("Build the project")))
	d.Dispatch("test", NewCommand(NewFlagSet("test"),
		func(fs *FlagSet, args []string) error { return nil },
		WithUsage("Run tests")))
	d.Dispatch("test unit", NewCommand(NewFlagSet("test unit"),
		func(fs *FlagSet, args []string) error { return nil },
		WithUsage("Run unit tests")))
	d.Dispatch("test integration", NewCommand(NewFlagSet("test integration"),
		func(fs *FlagSet, args []string) error { return nil },
		WithUsage("Run integration tests")))

	tests := []struct {
		name     string
		prefix   string
		expected []string
	}{
		{
			name:     "no prefix shows all",
			prefix:   "",
			expected: []string{"build", "test", "test integration", "test unit"},
		},
		{
			name:     "partial match",
			prefix:   "te",
			expected: []string{"test", "test integration", "test unit"},
		},
		{
			name:     "exact match",
			prefix:   "test",
			expected: []string{"test", "test integration", "test unit"},
		},
		{
			name:     "nested command prefix",
			prefix:   "test int",
			expected: []string{"test integration"},
		},
		{
			name:     "no matches",
			prefix:   "unknown",
			expected: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			completions := d.GetCommandCompletions(test.prefix)
			var values []string
			for _, c := range completions {
				values = append(values, c.Value)
			}
			// Handle empty/nil slice comparison
			if len(test.expected) == 0 && len(values) == 0 {
				// Both are empty, consider them equal
				return
			}
			assert.Equal(t, test.expected, values)
		})
	}
}

func TestDispatcherHandleCompletion(t *testing.T) {
	d := NewDispatcher("myapp")

	d.Dispatch("build", NewCommand(NewFlagSet("build"),
		func(fs *FlagSet, args []string) error { return nil }))
	d.Dispatch("test", NewCommand(NewFlagSet("test"),
		func(fs *FlagSet, args []string) error { return nil }))

	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		{
			name:     "bash completion flag",
			args:     []string{"--complete-bash", "te"},
			expected: true,
		},
		{
			name:     "zsh completion flag",
			args:     []string{"--complete-zsh"},
			expected: true,
		},
		{
			name:     "generate bash script",
			args:     []string{"--generate-bash-completion"},
			expected: true,
		},
		{
			name:     "generate zsh script",
			args:     []string{"--generate-zsh-completion"},
			expected: true,
		},
		{
			name:     "normal command",
			args:     []string{"build"},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Capture stdout to prevent test output pollution
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			handled := d.HandleCompletion(test.args)

			w.Close()
			os.Stdout = old

			// Drain the pipe
			var buf bytes.Buffer
			io.Copy(&buf, r)

			assert.Equal(t, test.expected, handled)
		})
	}
}

func TestDispatcherBashCompletions(t *testing.T) {
	d := NewDispatcher("myapp")

	fs := NewFlagSet("build")
	fs.Bool("verbose", 'v', false, "verbose output")
	fs.String("output", 'o', "a.out", "output file")

	d.Dispatch("build", NewCommand(fs,
		func(flags *FlagSet, args []string) error { return nil },
		WithUsage("Build the project")))
	d.Dispatch("test", NewCommand(NewFlagSet("test"),
		func(fs *FlagSet, args []string) error { return nil }))

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test command completion
	d.PrintBashCompletions([]string{"bu"})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Should suggest "build" command
	assert.Contains(t, output, "build")
	assert.NotContains(t, output, "test")
}

func TestDispatcherGenerateCompletionScripts(t *testing.T) {
	d := NewDispatcher("myapp")

	d.Dispatch("build", NewCommand(NewFlagSet("build"),
		func(fs *FlagSet, args []string) error { return nil },
		WithUsage("Build the project")))

	// Test bash completion script generation
	bashScript := d.GenerateBashCompletion()
	assert.Contains(t, bashScript, "_myapp_completion")
	assert.Contains(t, bashScript, "complete -F _myapp_completion myapp")
	assert.Contains(t, bashScript, "--complete-bash")

	// Test zsh completion script generation
	zshScript := d.GenerateZshCompletion()
	assert.Contains(t, zshScript, "#compdef myapp")
	assert.Contains(t, zshScript, "_myapp()")
	assert.Contains(t, zshScript, "build[Build the project]")
}

func TestDispatcherHelpWithInterspersedFlags(t *testing.T) {
	d := NewDispatcher("myapp")

	// Create nested command "foo bar" with its own flags
	barFs := NewFlagSet("bar")
	barVerbose := barFs.Bool("verbose", 'v', false, "verbose output")
	config := barFs.String("config", 'C', "", "config file path")

	d.Dispatch("foo bar", NewCommand(barFs,
		func(fs *FlagSet, args []string) error { return nil },
		WithUsage("Execute the bar subcommand")))

	// Test various patterns of interspersed flags with help
	tests := []struct {
		name        string
		args        []string
		shouldHelp  bool
		description string
	}{
		{
			name:        "help after flag with arg",
			args:        []string{"foo", "-C", "local", "bar", "-h"},
			shouldHelp:  true,
			description: "Should show help even when -C flag with argument comes before command",
		},
		{
			name:        "help after flag with arg using --help",
			args:        []string{"foo", "-C", "local", "bar", "--help"},
			shouldHelp:  true,
			description: "Should show help with --help after flag with argument",
		},
		{
			name:        "help in middle of command path",
			args:        []string{"foo", "-h", "bar"},
			shouldHelp:  true,
			description: "Should show help when -h appears in middle of command path",
		},
		{
			name:        "help after multiple flags",
			args:        []string{"foo", "-C", "config.yml", "bar", "-v", "-h"},
			shouldHelp:  true,
			description: "Should show help even with multiple flags before it",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := d.Execute(test.args)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if test.shouldHelp {
				assert.NoError(t, err, test.description)
				// Check that help output was shown
				assert.Contains(t, output, "Usage:", test.description)
				// For the "foo bar" command specifically
				assert.Contains(t, output, "Execute the bar subcommand", test.description)
			}
		})
	}

	// Also test that without help flag, the command executes normally
	t.Run("normal execution without help", func(t *testing.T) {
		var executed bool
		d.Dispatch("foo bar", NewCommand(barFs,
			func(fs *FlagSet, args []string) error {
				executed = true
				return nil
			},
			WithUsage("Execute the bar subcommand")))

		err := d.Execute([]string{"foo", "-C", "local", "bar", "-v"})
		assert.NoError(t, err)
		assert.True(t, executed)
		assert.True(t, *barVerbose)
		assert.Equal(t, "local", *config)
	})
}

func TestDispatcherFlagsAfterPositionalArgs(t *testing.T) {
	d := NewDispatcher("myapp")

	// Create command "foo bar" with flags and positional arguments
	fs := NewFlagSet("foo bar")
	verbose := fs.Bool("verbose", 'v', false, "verbose output")
	output := fs.String("output", 'o', "default.txt", "output file")

	var capturedArgs []string
	var executed bool

	d.Dispatch("foo bar", NewCommand(fs,
		func(flags *FlagSet, args []string) error {
			executed = true
			capturedArgs = args
			return nil
		},
		WithUsage("Process files with options")))

	tests := []struct {
		name          string
		args          []string
		expectHelp    bool
		expectArgs    []string
		expectVerbose bool
		expectOutput  string
		description   string
	}{
		{
			name:        "help after positional arg",
			args:        []string{"foo", "bar", "baz", "-h"},
			expectHelp:  true,
			description: "Should show help when -h comes after positional arg 'baz'",
		},
		{
			name:        "help after multiple positional args",
			args:        []string{"foo", "bar", "file1", "file2", "--help"},
			expectHelp:  true,
			description: "Should show help when --help comes after multiple positional args",
		},
		{
			name:          "flags after positional arg",
			args:          []string{"foo", "bar", "myfile", "-v", "--output", "result.txt"},
			expectArgs:    []string{"myfile"},
			expectVerbose: true,
			expectOutput:  "result.txt",
			description:   "Should parse flags that come after positional arguments",
		},
		{
			name:          "mixed positional and flags",
			args:          []string{"foo", "bar", "file1", "-v", "file2", "--output", "out.txt", "file3"},
			expectArgs:    []string{"file1", "file2", "file3"},
			expectVerbose: true,
			expectOutput:  "out.txt",
			description:   "Should handle mixed positional args and flags",
		},
		{
			name:          "positional arg named like subcommand",
			args:          []string{"foo", "bar", "baz", "-v"},
			expectArgs:    []string{"baz"},
			expectVerbose: true,
			expectOutput:  "default.txt",
			description:   "Should treat 'baz' as positional arg, not a subcommand",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Reset state
			executed = false
			capturedArgs = nil
			*verbose = false
			*output = "default.txt"

			if test.expectHelp {
				// Capture stdout for help output
				old := os.Stdout
				r, w, _ := os.Pipe()
				os.Stdout = w

				err := d.Execute(test.args)

				w.Close()
				os.Stdout = old

				var buf bytes.Buffer
				io.Copy(&buf, r)
				output := buf.String()

				assert.NoError(t, err, test.description)
				assert.Contains(t, output, "Usage:", test.description)
				assert.Contains(t, output, "Process files with options", test.description)
				assert.False(t, executed, "Command should not execute when help is shown")
			} else {
				err := d.Execute(test.args)
				assert.NoError(t, err, test.description)
				assert.True(t, executed, "Command should execute")
				assert.Equal(t, test.expectArgs, capturedArgs, test.description)
				assert.Equal(t, test.expectVerbose, *verbose, test.description)
				assert.Equal(t, test.expectOutput, *output, test.description)
			}
		})
	}
}
