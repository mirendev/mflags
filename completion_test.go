package mflags

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLongFlags(t *testing.T) {
	fs := NewFlagSet("test")
	fs.Bool("verbose", 'v', false, "verbose output")
	fs.String("output", 'o', "stdout", "output file")
	fs.Int("count", 'c', 1, "count value")

	flags := fs.GetLongFlags()
	assert.Equal(t, []string{"--count", "--output", "--verbose"}, flags)
}

func TestGetShortFlags(t *testing.T) {
	fs := NewFlagSet("test")
	fs.Bool("verbose", 'v', false, "verbose output")
	fs.String("output", 'o', "stdout", "output file")
	fs.Int("count", 'c', 1, "count value")

	flags := fs.GetShortFlags()
	assert.Equal(t, []string{"-c", "-o", "-v"}, flags)
}

func TestVisitAll(t *testing.T) {
	fs := NewFlagSet("test")
	fs.Bool("verbose", 'v', false, "verbose output")
	fs.String("output", 'o', "stdout", "output file")
	fs.Int("count", 'c', 1, "count value")

	var names []string
	fs.VisitAll(func(flag *Flag) {
		names = append(names, flag.Name)
	})

	assert.Equal(t, []string{"count", "output", "verbose"}, names)
}

func TestGetFlagCompletions(t *testing.T) {
	fs := NewFlagSet("test")
	fs.Bool("verbose", 'v', false, "verbose output")
	fs.Bool("version", 0, false, "show version")
	fs.String("output", 'o', "stdout", "output file")
	fs.Int("count", 'c', 1, "count value")

	tests := []struct {
		name     string
		prefix   string
		expected []string
	}{
		{
			name:   "no prefix shows all",
			prefix: "",
			expected: []string{
				"--count", "--output", "--verbose", "--version",
				"-c", "-o", "-v",
			},
		},
		{
			name:     "long flag prefix",
			prefix:   "--ver",
			expected: []string{"--verbose", "--version"},
		},
		{
			name:     "exact long flag",
			prefix:   "--verbose",
			expected: []string{"--verbose"},
		},
		{
			name:     "short flag prefix",
			prefix:   "-",
			expected: []string{"-c", "-o", "-v"},
		},
		{
			name:     "specific short flag",
			prefix:   "-v",
			expected: []string{"-v"},
		},
		{
			name:     "no matches",
			prefix:   "--xyz",
			expected: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			completions := fs.GetFlagCompletions(test.prefix)
			var values []string
			for _, c := range completions {
				values = append(values, c.Value)
			}
			if len(test.expected) == 0 && len(values) == 0 {
				// Both are empty, consider them equal
				return
			}
			assert.Equal(t, test.expected, values)
		})
	}
}

func TestCompletionDescriptions(t *testing.T) {
	fs := NewFlagSet("test")
	fs.Bool("verbose", 'v', false, "enable verbose output")
	fs.String("output", 'o', "stdout", "specify output file")

	completions := fs.GetFlagCompletions("--")

	assert.Len(t, completions, 2)

	// Find the verbose completion
	var verboseComp *Completion
	for i := range completions {
		if completions[i].Value == "--verbose" {
			verboseComp = &completions[i]
			break
		}
	}

	assert.NotNil(t, verboseComp)
	assert.Equal(t, "--verbose", verboseComp.Value)
	assert.Equal(t, "enable verbose output", verboseComp.Description)
	assert.True(t, verboseComp.IsBool)

	// Find the output completion
	var outputComp *Completion
	for i := range completions {
		if completions[i].Value == "--output" {
			outputComp = &completions[i]
			break
		}
	}

	assert.NotNil(t, outputComp)
	assert.Equal(t, "--output", outputComp.Value)
	assert.Equal(t, "specify output file", outputComp.Description)
	assert.False(t, outputComp.IsBool)
}

func TestPrintBashCompletions(t *testing.T) {
	fs := NewFlagSet("test")
	fs.Bool("verbose", 'v', false, "verbose output")
	fs.String("output", 'o', "stdout", "output file")

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test completing flags
	fs.PrintBashCompletions([]string{"--ver"})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.Contains(t, output, "--verbose")
	assert.NotContains(t, output, "--output")
}

func TestHandleCompletion(t *testing.T) {
	fs := NewFlagSet("test")
	fs.Bool("verbose", 'v', false, "verbose output")

	tests := []struct {
		name     string
		args     []string
		env      map[string]string
		expected bool
	}{
		{
			name:     "bash completion flag",
			args:     []string{"--complete-bash", "--ver"},
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
			name:     "normal args",
			args:     []string{"--verbose", "file.txt"},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range test.env {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			// Capture stdout to prevent test output pollution
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			handled := fs.HandleCompletion(test.args)

			w.Close()
			os.Stdout = old

			// Drain the pipe
			var buf bytes.Buffer
			io.Copy(&buf, r)

			assert.Equal(t, test.expected, handled)
		})
	}
}

func TestGenerateBashCompletion(t *testing.T) {
	fs := NewFlagSet("myapp")
	fs.Bool("verbose", 'v', false, "verbose output")
	fs.String("output", 'o', "stdout", "output file")

	script := fs.GenerateBashCompletion("myapp")

	assert.Contains(t, script, "myapp_completion")
	assert.Contains(t, script, "complete -F _myapp_completion myapp")
	assert.Contains(t, script, "COMPREPLY")
}

func TestGenerateZshCompletion(t *testing.T) {
	fs := NewFlagSet("myapp")
	fs.Bool("verbose", 'v', false, "verbose output")
	fs.String("output", 'o', "stdout", "output file")

	script := fs.GenerateZshCompletion("myapp")

	assert.Contains(t, script, "#compdef myapp")
	assert.Contains(t, script, "_myapp()")
	assert.Contains(t, script, "--verbose[verbose output]")
	assert.Contains(t, script, "--output=[output file]")
	assert.Contains(t, script, "_arguments")
}

func TestCompletionWithStruct(t *testing.T) {
	type Config struct {
		Verbose bool   `long:"verbose" short:"v" usage:"Enable verbose mode"`
		Output  string `long:"output" short:"o" default:"stdout" usage:"Output file"`
		Count   int    `long:"count" short:"c" default:"1" usage:"Number of items"`
	}

	config := &Config{}
	fs := NewFlagSet("test")
	err := fs.FromStruct(config)
	assert.NoError(t, err)

	// Test that struct-based flags appear in completions
	completions := fs.GetFlagCompletions("--")

	var names []string
	for _, c := range completions {
		names = append(names, c.Value)
	}

	assert.Contains(t, names, "--verbose")
	assert.Contains(t, names, "--output")
	assert.Contains(t, names, "--count")

	// Test descriptions from struct tags
	for _, c := range completions {
		switch c.Value {
		case "--verbose":
			assert.Equal(t, "Enable verbose mode", c.Description)
		case "--output":
			assert.Equal(t, "Output file", c.Description)
		case "--count":
			assert.Equal(t, "Number of items", c.Description)
		}
	}
}

func TestCompletionEdgeCases(t *testing.T) {
	fs := NewFlagSet("test")

	// Test with no flags
	completions := fs.GetFlagCompletions("")
	assert.Empty(t, completions)

	// Add a flag with only long name
	fs.Bool("verbose", 0, false, "verbose")
	completions = fs.GetFlagCompletions("")
	assert.Len(t, completions, 1)
	assert.Equal(t, "--verbose", completions[0].Value)

	// Test with special characters in descriptions
	fs2 := NewFlagSet("test2")
	fs2.Bool("test", 't', false, "This has 'quotes' and \"double quotes\"")

	script := fs2.GenerateZshCompletion("test")
	// Should escape quotes properly
	assert.Contains(t, script, "'\"'\"'")
}

func TestCompletionSorting(t *testing.T) {
	fs := NewFlagSet("test")
	fs.Bool("zebra", 'z', false, "last alphabetically")
	fs.Bool("alpha", 'a', false, "first alphabetically")
	fs.Bool("middle", 'm', false, "middle alphabetically")

	completions := fs.GetFlagCompletions("")

	// Long flags should be sorted
	longFlags := []string{}
	for _, c := range completions {
		if strings.HasPrefix(c.Value, "--") {
			longFlags = append(longFlags, c.Value)
		}
	}

	assert.Equal(t, []string{"--alpha", "--middle", "--zebra"}, longFlags[:3])
}
