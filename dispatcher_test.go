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
	var restArgs []string
	fs.Rest(&restArgs, "extra arguments")

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

	// Register nested commands - each with a rest field to accept extra args
	fooFs := NewFlagSet("foo")
	var fooRest []string
	fooFs.Rest(&fooRest, "extra arguments")
	d.Dispatch("foo", NewCommand(fooFs, func(fs *FlagSet, args []string) error {
		executedCommand = "foo"
		return nil
	}))

	fooBarFs := NewFlagSet("foo bar")
	var fooBarRest []string
	fooBarFs.Rest(&fooBarRest, "extra arguments")
	d.Dispatch("foo bar", NewCommand(fooBarFs, func(fs *FlagSet, args []string) error {
		executedCommand = "foo bar"
		return nil
	}))

	fooBarBazFs := NewFlagSet("foo bar baz")
	var fooBarBazRest []string
	fooBarBazFs.Rest(&fooBarBazRest, "extra arguments")
	d.Dispatch("foo bar baz", NewCommand(fooBarBazFs, func(fs *FlagSet, args []string) error {
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
	var restArgs []string
	fs.Rest(&restArgs, "source files")

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

func TestDispatcherCommandHelpWithPositionalArguments(t *testing.T) {
	t.Run("positional arguments shown by name with descriptions", func(t *testing.T) {
		d := NewDispatcher("myapp")

		type Config struct {
			Verbose     bool   `long:"verbose" short:"v" description:"Enable verbose output"`
			Environment string `position:"0" usage:"Target environment (dev, staging, prod)"`
			Version     string `position:"1" usage:"Version to deploy"`
		}

		fs := NewFlagSet("deploy")
		var cfg Config
		fs.FromStruct(&cfg)

		d.Dispatch("deploy", NewCommand(fs,
			func(flags *FlagSet, args []string) error { return nil },
			WithUsage("Deploy to an environment")))

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := d.Execute([]string{"deploy", "--help"})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		assert.NoError(t, err)
		assert.Contains(t, output, "Usage: myapp deploy [options] <environment> <version>")
		assert.NotContains(t, output, "[arguments]")
		assert.Contains(t, output, "Arguments:")
		assert.Contains(t, output, "<environment>")
		assert.Contains(t, output, "Target environment (dev, staging, prod)")
		assert.Contains(t, output, "<version>")
		assert.Contains(t, output, "Version to deploy")
	})

	t.Run("positional arguments with rest field", func(t *testing.T) {
		d := NewDispatcher("myapp")

		type Config struct {
			Verbose bool     `long:"verbose" short:"v" description:"Enable verbose output"`
			Command string   `position:"0" usage:"Command to execute"`
			Args    []string `rest:"true"`
		}

		fs := NewFlagSet("exec")
		var cfg Config
		fs.FromStruct(&cfg)

		d.Dispatch("exec", NewCommand(fs,
			func(flags *FlagSet, args []string) error { return nil },
			WithUsage("Execute a command")))

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := d.Execute([]string{"exec", "--help"})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		assert.NoError(t, err)
		assert.Contains(t, output, "Usage: myapp exec [options] <command> [arguments...]")
		assert.Contains(t, output, "Arguments:")
		assert.Contains(t, output, "Command to execute")
	})

	t.Run("only rest field shows arguments", func(t *testing.T) {
		d := NewDispatcher("myapp")

		type Config struct {
			Verbose bool     `long:"verbose" short:"v" description:"Enable verbose output"`
			Args    []string `rest:"true"`
		}

		fs := NewFlagSet("echo")
		var cfg Config
		fs.FromStruct(&cfg)

		d.Dispatch("echo", NewCommand(fs,
			func(flags *FlagSet, args []string) error { return nil },
			WithUsage("Echo arguments")))

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := d.Execute([]string{"echo", "--help"})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		assert.NoError(t, err)
		assert.Contains(t, output, "Usage: myapp echo [options] [arguments...]")
		// No Arguments: section when no positional args have usage
		assert.NotContains(t, output, "Arguments:")
	})
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

	testFs := NewFlagSet("test")
	var restArgs []string
	testFs.Rest(&restArgs, "extra arguments")

	var capturedArgs []string
	d.Dispatch("foo bar baz", NewCommand(testFs, func(fs *FlagSet, args []string) error {
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

	// Create command "foo bar" with flags and a rest field for extra arguments
	fs := NewFlagSet("foo bar")
	verbose := fs.Bool("verbose", 'v', false, "verbose output")
	output := fs.String("output", 'o', "default.txt", "output file")
	var restArgs []string
	fs.Rest(&restArgs, "files to process")

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

// TestDispatcherSubCommandHelp tests that sub-commands are displayed in help output
func TestDispatcherSubCommandHelp(t *testing.T) {
	d := NewDispatcher("myapp")

	// Register parent command
	parentFs := NewFlagSet("git")
	parentFs.Bool("verbose", 'v', false, "verbose output")
	d.Dispatch("git", NewCommand(parentFs,
		func(fs *FlagSet, args []string) error { return nil },
		WithUsage("Git version control system")))

	// Register direct sub-commands
	d.Dispatch("git clone", NewCommand(NewFlagSet("git clone"),
		func(fs *FlagSet, args []string) error { return nil },
		WithUsage("Clone a repository")))

	d.Dispatch("git commit", NewCommand(NewFlagSet("git commit"),
		func(fs *FlagSet, args []string) error { return nil },
		WithUsage("Record changes to the repository")))

	d.Dispatch("git push", NewCommand(NewFlagSet("git push"),
		func(fs *FlagSet, args []string) error { return nil },
		WithUsage("Update remote refs")))

	// Register nested sub-command (should not show in git's help)
	d.Dispatch("git remote add", NewCommand(NewFlagSet("git remote add"),
		func(fs *FlagSet, args []string) error { return nil },
		WithUsage("Add a new remote")))

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Show help for the git command
	err := d.Execute([]string{"git", "--help"})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "Usage: myapp git")
	assert.Contains(t, output, "Git version control system")
	assert.Contains(t, output, "Options:")
	assert.Contains(t, output, "-v, --verbose")
	assert.Contains(t, output, "Sub-commands:")
	assert.Contains(t, output, "clone")
	assert.Contains(t, output, "Clone a repository")
	assert.Contains(t, output, "commit")
	assert.Contains(t, output, "Record changes to the repository")
	assert.Contains(t, output, "push")
	assert.Contains(t, output, "Update remote refs")
	// Should not show nested sub-command "remote add", only direct children
	assert.NotContains(t, output, "remote add")
}

func TestDispatcherNamespaceDiscovery(t *testing.T) {
	t.Run("top-level help shows namespace from subcommand-only registration", func(t *testing.T) {
		d := NewDispatcher("myapp")

		// Register a top-level command and a subcommand without its parent
		d.Dispatch("build", NewCommand(NewFlagSet("build"),
			func(fs *FlagSet, args []string) error { return nil },
			WithUsage("Build the project")))

		d.Dispatch("config get", NewCommand(NewFlagSet("config get"),
			func(fs *FlagSet, args []string) error { return nil },
			WithUsage("Get a config value")))

		d.Dispatch("config set", NewCommand(NewFlagSet("config set"),
			func(fs *FlagSet, args []string) error { return nil },
			WithUsage("Set a config value")))

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := d.Execute([]string{"help"})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		assert.NoError(t, err)
		assert.Contains(t, output, "Available commands:")
		assert.Contains(t, output, "build")
		assert.Contains(t, output, "config")
		// Should NOT show the full subcommand paths at the top level
		assert.NotContains(t, output, "config get")
		assert.NotContains(t, output, "config set")
	})

	t.Run("typing a namespace shows its subcommands", func(t *testing.T) {
		d := NewDispatcher("myapp")

		d.Dispatch("config get", NewCommand(NewFlagSet("config get"),
			func(fs *FlagSet, args []string) error { return nil },
			WithUsage("Get a config value")))

		d.Dispatch("config set", NewCommand(NewFlagSet("config set"),
			func(fs *FlagSet, args []string) error { return nil },
			WithUsage("Set a config value")))

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := d.Execute([]string{"config"})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		assert.NoError(t, err)
		assert.Contains(t, output, "Usage: myapp config <command>")
		assert.Contains(t, output, "Available commands:")
		assert.Contains(t, output, "get")
		assert.Contains(t, output, "Get a config value")
		assert.Contains(t, output, "set")
		assert.Contains(t, output, "Set a config value")
	})

	t.Run("namespace --help shows subcommands", func(t *testing.T) {
		d := NewDispatcher("myapp")

		d.Dispatch("config get", NewCommand(NewFlagSet("config get"),
			func(fs *FlagSet, args []string) error { return nil },
			WithUsage("Get a config value")))

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := d.Execute([]string{"config", "--help"})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		assert.NoError(t, err)
		assert.Contains(t, output, "Usage: myapp config <command>")
		assert.Contains(t, output, "get")
	})

	t.Run("deep namespace surfaces intermediate namespaces", func(t *testing.T) {
		d := NewDispatcher("myapp")

		d.Dispatch("cloud compute instances list", NewCommand(NewFlagSet("list"),
			func(fs *FlagSet, args []string) error { return nil },
			WithUsage("List compute instances")))

		d.Dispatch("cloud storage buckets list", NewCommand(NewFlagSet("list"),
			func(fs *FlagSet, args []string) error { return nil },
			WithUsage("List storage buckets")))

		// Top-level help should show "cloud"
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := d.Execute([]string{"help"})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		assert.NoError(t, err)
		assert.Contains(t, output, "cloud")
		assert.NotContains(t, output, "compute")

		// "cloud" namespace should show "compute" and "storage"
		r, w, _ = os.Pipe()
		os.Stdout = w

		err = d.Execute([]string{"cloud"})

		w.Close()
		os.Stdout = old

		buf.Reset()
		io.Copy(&buf, r)
		output = buf.String()

		assert.NoError(t, err)
		assert.Contains(t, output, "compute")
		assert.Contains(t, output, "storage")

		// "cloud compute" namespace should show "instances"
		r, w, _ = os.Pipe()
		os.Stdout = w

		err = d.Execute([]string{"cloud", "compute"})

		w.Close()
		os.Stdout = old

		buf.Reset()
		io.Copy(&buf, r)
		output = buf.String()

		assert.NoError(t, err)
		assert.Contains(t, output, "instances")
	})

	t.Run("typing namespace under registered parent shows namespace help", func(t *testing.T) {
		d := NewDispatcher("myapp")

		// Register "debug" as a command
		debugFs := NewFlagSet("debug")
		d.Dispatch("debug", NewCommand(debugFs,
			func(fs *FlagSet, args []string) error { return nil },
			WithUsage("Debug commands")))

		// Register subcommands under "debug entity" without registering "debug entity" itself
		d.Dispatch("debug entity list", NewCommand(NewFlagSet("list"),
			func(fs *FlagSet, args []string) error { return nil },
			WithUsage("List entities")))

		d.Dispatch("debug entity show", NewCommand(NewFlagSet("show"),
			func(fs *FlagSet, args []string) error { return nil },
			WithUsage("Show entity details")))

		// "debug entity" should show namespace help, not "unexpected arguments"
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := d.Execute([]string{"debug", "entity"})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		assert.NoError(t, err)
		assert.Contains(t, output, "Usage: myapp debug entity <command>")
		assert.Contains(t, output, "list")
		assert.Contains(t, output, "List entities")
		assert.Contains(t, output, "show")
		assert.Contains(t, output, "Show entity details")
	})

	t.Run("typing namespace with -h under registered parent shows namespace help", func(t *testing.T) {
		d := NewDispatcher("myapp")

		debugFs := NewFlagSet("debug")
		d.Dispatch("debug", NewCommand(debugFs,
			func(fs *FlagSet, args []string) error { return nil },
			WithUsage("Debug commands")))

		d.Dispatch("debug entity list", NewCommand(NewFlagSet("list"),
			func(fs *FlagSet, args []string) error { return nil },
			WithUsage("List entities")))

		// "debug entity -h" should show namespace help for entity, not debug's help
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := d.Execute([]string{"debug", "entity", "-h"})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		assert.NoError(t, err)
		assert.Contains(t, output, "Usage: myapp debug entity <command>")
		assert.Contains(t, output, "list")
	})

	t.Run("registered parent with subcommand-only deeper children shows namespace in sub-commands", func(t *testing.T) {
		d := NewDispatcher("myapp")

		// Register "git" as a command
		d.Dispatch("git", NewCommand(NewFlagSet("git"),
			func(fs *FlagSet, args []string) error { return nil },
			WithUsage("Git version control")))

		// Register "git remote add" without "git remote"
		d.Dispatch("git remote add", NewCommand(NewFlagSet("add"),
			func(fs *FlagSet, args []string) error { return nil },
			WithUsage("Add a new remote")))

		d.Dispatch("git clone", NewCommand(NewFlagSet("clone"),
			func(fs *FlagSet, args []string) error { return nil },
			WithUsage("Clone a repository")))

		// "git --help" should show "remote" as a sub-command
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := d.Execute([]string{"git", "--help"})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		assert.NoError(t, err)
		assert.Contains(t, output, "Sub-commands:")
		assert.Contains(t, output, "clone")
		assert.Contains(t, output, "remote")
	})
}

// TestDispatcherHelpShowsTypes tests that help output shows specific types for flags
func TestDispatcherHelpShowsTypes(t *testing.T) {
	d := NewDispatcher("myapp")

	// Create a command with flags of different types
	fs := NewFlagSet("build")
	fs.String("output", 'o', "a.out", "output file")
	fs.Int("jobs", 'j', 1, "number of parallel jobs")
	fs.Duration("timeout", 't', 0, "build timeout")
	fs.StringArray("tags", 'T', []string{}, "build tags")
	fs.Bool("verbose", 'v', false, "verbose output")

	d.Dispatch("build", NewCommand(fs,
		func(flags *FlagSet, args []string) error { return nil },
		WithUsage("Build the project")))

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
	assert.Contains(t, output, "-o, --output <string>")
	assert.Contains(t, output, "-j, --jobs <int>")
	assert.Contains(t, output, "-t, --timeout <duration>")
	assert.Contains(t, output, "-T, --tags <value,...>")
	assert.Contains(t, output, "-v, --verbose")
}

// TestDispatcherHelpAfterDoubleHyphen tests that help flags after -- are not processed
func TestDispatcherHelpAfterDoubleHyphen(t *testing.T) {
	d := NewDispatcher("myapp")

	fs := NewFlagSet("process")
	fs.Bool("verbose", 'v', false, "verbose output")
	var restArgs []string
	fs.Rest(&restArgs, "files to process")

	var executed bool
	var capturedArgs []string

	cmd := NewCommand(fs, func(flags *FlagSet, args []string) error {
		executed = true
		capturedArgs = args
		return nil
	}, WithUsage("Process files"))

	d.Dispatch("process", cmd)

	// Test that -h after -- is treated as an argument, not a help flag
	executed = false
	capturedArgs = nil
	err := d.Execute([]string{"process", "--", "file.txt", "-h"})

	assert.NoError(t, err, "Should not error when -h comes after --")
	assert.True(t, executed, "Command should execute, not show help")
	assert.Equal(t, []string{"file.txt", "-h"}, capturedArgs, "-h after -- should be treated as argument")

	// Test that --help after -- is treated as an argument, not a help flag
	executed = false
	capturedArgs = nil
	err = d.Execute([]string{"process", "--", "file.txt", "--help"})

	assert.NoError(t, err, "Should not error when --help comes after --")
	assert.True(t, executed, "Command should execute, not show help")
	assert.Equal(t, []string{"file.txt", "--help"}, capturedArgs, "--help after -- should be treated as argument")

	// Test that help after -- is treated as an argument, not a help flag
	executed = false
	capturedArgs = nil
	err = d.Execute([]string{"process", "--", "help"})

	assert.NoError(t, err, "Should not error when help comes after --")
	assert.True(t, executed, "Command should execute, not show help")
	assert.Equal(t, []string{"help"}, capturedArgs, "help after -- should be treated as argument")

	// Ensure -h before -- still works as help
	executed = false
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = d.Execute([]string{"process", "-h"})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.NoError(t, err, "Should not error on help flag")
	assert.Contains(t, output, "Usage:", "Should show help")
	assert.False(t, executed, "Command should not execute when help is requested")
}

func TestDispatcherHelpWithAllowUnknownFlags(t *testing.T) {
	d := NewDispatcher("myapp")

	// Create a command that allows unknown flags
	fs := NewFlagSet("run")
	fs.AllowUnknownFlags(true)
	fs.String("config", 'c', "", "config file")

	var executed bool
	var capturedArgs []string
	var capturedUnknown []string

	cmd := NewCommand(fs, func(flags *FlagSet, args []string) error {
		executed = true
		capturedArgs = args
		capturedUnknown = flags.UnknownFlags()
		return nil
	}, WithUsage("Run a command"))

	d.Dispatch("run", cmd)

	// Test: -h alone should show help even with allowUnknownFlags
	t.Run("-h alone shows help", func(t *testing.T) {
		executed = false
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := d.Execute([]string{"run", "-h"})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Usage:")
		assert.False(t, executed, "Should show help, not execute")
	})

	// Test: run blah -h should NOT show help (user's case)
	t.Run("run blah -h does not show help", func(t *testing.T) {
		executed = false
		capturedArgs = nil
		capturedUnknown = nil

		err := d.Execute([]string{"run", "blah", "-h"})

		assert.NoError(t, err)
		assert.True(t, executed, "Should execute command")
		assert.Equal(t, []string{"blah"}, capturedArgs)
		assert.Equal(t, []string{"-h"}, capturedUnknown)
	})

	// Test: run --help with other args
	t.Run("run command --help does not show help", func(t *testing.T) {
		executed = false
		capturedArgs = nil
		capturedUnknown = nil

		err := d.Execute([]string{"run", "command", "--help"})

		assert.NoError(t, err)
		assert.True(t, executed, "Should execute command")
		assert.Equal(t, []string{"command"}, capturedArgs)
		assert.Equal(t, []string{"--help"}, capturedUnknown)
	})

	// Test: without allowUnknownFlags, -h with args still shows help
	t.Run("without allowUnknownFlags, -h shows help", func(t *testing.T) {
		d2 := NewDispatcher("myapp")
		fs2 := NewFlagSet("test")
		fs2.String("output", 'o', "", "output")

		cmd2 := NewCommand(fs2, func(flags *FlagSet, args []string) error {
			return nil
		}, WithUsage("Test command"))

		d2.Dispatch("test", cmd2)

		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := d2.Execute([]string{"test", "blah", "-h"})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Usage:")
	})
}

func TestDispatcherRejectsExtraArgs(t *testing.T) {
	t.Run("command with no positional args rejects extra", func(t *testing.T) {
		d := NewDispatcher("myapp")

		fs := NewFlagSet("app")
		fs.Bool("verbose", 'v', false, "verbose output")

		d.Dispatch("app", NewCommand(fs,
			func(flags *FlagSet, args []string) error { return nil },
			WithUsage("Application command")))

		err := d.Execute([]string{"app", "services"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected arguments: [services]")
	})

	t.Run("nested command with no positional args rejects extra", func(t *testing.T) {
		d := NewDispatcher("myapp")

		fs := NewFlagSet("app services")
		fs.Bool("verbose", 'v', false, "verbose output")

		d.Dispatch("app services", NewCommand(fs,
			func(flags *FlagSet, args []string) error { return nil },
			WithUsage("List services")))

		err := d.Execute([]string{"app", "services", "extra"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected arguments: [extra]")
	})

	t.Run("command with positional args accepts correct count", func(t *testing.T) {
		d := NewDispatcher("myapp")

		fs := NewFlagSet("deploy")
		fs.StringPos("environment", 0, "", "Target environment")

		var executed bool
		d.Dispatch("deploy", NewCommand(fs,
			func(flags *FlagSet, args []string) error {
				executed = true
				return nil
			},
			WithUsage("Deploy to environment")))

		err := d.Execute([]string{"deploy", "production"})
		assert.NoError(t, err)
		assert.True(t, executed)
	})

	t.Run("command with positional args rejects extra", func(t *testing.T) {
		d := NewDispatcher("myapp")

		fs := NewFlagSet("deploy")
		fs.StringPos("environment", 0, "", "Target environment")

		d.Dispatch("deploy", NewCommand(fs,
			func(flags *FlagSet, args []string) error { return nil },
			WithUsage("Deploy to environment")))

		err := d.Execute([]string{"deploy", "production", "extra"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected arguments: [extra]")
	})

	t.Run("command with rest field accepts all args", func(t *testing.T) {
		d := NewDispatcher("myapp")

		type Config struct {
			Verbose bool     `long:"verbose" short:"v"`
			Files   []string `rest:"true"`
		}

		config := &Config{}
		fs := NewFlagSet("process")
		fs.FromStruct(config)

		d.Dispatch("process", NewCommand(fs,
			func(flags *FlagSet, args []string) error { return nil },
			WithUsage("Process files")))

		err := d.Execute([]string{"process", "file1", "file2", "file3"})
		assert.NoError(t, err)
		assert.Equal(t, []string{"file1", "file2", "file3"}, config.Files)
	})
}

func TestDispatcherErrShowHelp(t *testing.T) {
	// Test that returning ErrShowHelp from a command triggers help display
	t.Run("ErrShowHelp triggers help", func(t *testing.T) {
		d := NewDispatcher("myapp")

		fs := NewFlagSet("test")
		fs.String("output", 'o', "default", "output file")
		fs.Bool("verbose", 'v', false, "verbose mode")

		d.Dispatch("test", NewCommand(fs, func(flags *FlagSet, args []string) error {
			return ErrShowHelp
		}, WithUsage("Test command")))

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := d.Execute([]string{"test"})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		assert.NoError(t, err)
		assert.Contains(t, output, "Usage: myapp test")
		assert.Contains(t, output, "-o, --output")
		assert.Contains(t, output, "-v, --verbose")
		assert.Contains(t, output, "Test command")
	})

	// Test that ErrShowHelp works with arguments
	t.Run("ErrShowHelp with args", func(t *testing.T) {
		d := NewDispatcher("myapp")

		fs := NewFlagSet("greet")
		name := fs.String("name", 'n', "", "name to greet")

		d.Dispatch("greet", NewCommand(fs, func(flags *FlagSet, args []string) error {
			// Show help if name is empty
			if *name == "" {
				return ErrShowHelp
			}
			return nil
		}, WithUsage("Greet someone")))

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := d.Execute([]string{"greet"})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		assert.NoError(t, err)
		assert.Contains(t, output, "Usage: myapp greet")
		assert.Contains(t, output, "-n, --name")
	})

	// Test that normal execution still works
	t.Run("normal execution without ErrShowHelp", func(t *testing.T) {
		d := NewDispatcher("myapp")

		fs := NewFlagSet("greet")
		name := fs.String("name", 'n', "", "name to greet")

		var executed bool
		d.Dispatch("greet", NewCommand(fs, func(flags *FlagSet, args []string) error {
			if *name == "" {
				return ErrShowHelp
			}
			executed = true
			return nil
		}, WithUsage("Greet someone")))

		err := d.Execute([]string{"greet", "--name", "World"})

		assert.NoError(t, err)
		assert.True(t, executed)
	})

	// Test that other errors are still propagated
	t.Run("other errors still propagate", func(t *testing.T) {
		d := NewDispatcher("myapp")

		d.Dispatch("fail", NewCommand(NewFlagSet("fail"), func(flags *FlagSet, args []string) error {
			return fmt.Errorf("something went wrong")
		}))

		err := d.Execute([]string{"fail"})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "something went wrong")
	})

	// Test ErrShowHelp with nested commands
	t.Run("ErrShowHelp with nested command", func(t *testing.T) {
		d := NewDispatcher("myapp")

		fs := NewFlagSet("config get")
		key := fs.String("key", 'k', "", "config key")

		d.Dispatch("config get", NewCommand(fs, func(flags *FlagSet, args []string) error {
			if *key == "" {
				return ErrShowHelp
			}
			return nil
		}, WithUsage("Get a config value")))

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := d.Execute([]string{"config", "get"})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		assert.NoError(t, err)
		assert.Contains(t, output, "Usage: myapp config get")
		assert.Contains(t, output, "-k, --key")
	})

	// Test wrapped ErrShowHelp is detected
	t.Run("wrapped ErrShowHelp", func(t *testing.T) {
		d := NewDispatcher("myapp")

		fs := NewFlagSet("test")
		fs.String("required", 'r', "", "required field")

		d.Dispatch("test", NewCommand(fs, func(flags *FlagSet, args []string) error {
			return fmt.Errorf("missing required field: %w", ErrShowHelp)
		}, WithUsage("Test command")))

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := d.Execute([]string{"test"})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		assert.NoError(t, err)
		assert.Contains(t, output, "Usage: myapp test")
	})
}
