package mflags

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// unknownCommandDispatcher builds a dispatcher shaped like a real CLI: a plain
// top-level command, a leaf command that also has children, a section-style
// command that tolerates unknown flags, and a command that takes free-form
// arguments.
func unknownCommandDispatcher() *Dispatcher {
	d := NewDispatcher("myapp")

	noArgs := func() *FlagSet { return NewFlagSet("x") }
	run := func(fs *FlagSet, args []string) error { return nil }

	d.Dispatch("deploy", NewCommand(noArgs(), run, WithUsage("Deploy an application")))
	d.Dispatch("upgrade", NewCommand(noArgs(), run, WithUsage("Upgrade")))

	// "app" is a real command that also has sub-commands under it.
	d.Dispatch("app", NewCommand(noArgs(), run, WithUsage("Manage applications")))
	d.Dispatch("app list", NewCommand(noArgs(), run, WithUsage("List applications")))
	d.Dispatch("app destroy", NewCommand(noArgs(), run, WithUsage("Destroy an application")))

	// "runner" stands in for a Section: no flags of its own, but it tolerates
	// unknown ones so that global flags may precede it.
	section := NewFlagSet("runner")
	section.AllowUnknownFlags(true)
	d.Dispatch("runner", NewCommand(section, run, WithUsage("Manage runners")))
	d.Dispatch("runner upgrade", NewCommand(noArgs(), run, WithUsage("Upgrade a runner")))

	// "logs" takes a positional argument, so a stray word is a value, not a typo.
	logs := NewFlagSet("logs")
	logs.StringPos("app", 0, "", "Application name")
	d.Dispatch("logs", NewCommand(logs, run, WithUsage("View logs")))
	d.Dispatch("logs build", NewCommand(noArgs(), run, WithUsage("View build logs")))

	// "exec" passes everything through.
	type execOpts struct {
		Args []string `rest:"true"`
	}
	exec := NewFlagSet("exec")
	exec.FromStruct(&execOpts{})
	d.Dispatch("exec", NewCommand(exec, run, WithUsage("Run a command")))
	d.Dispatch("exec shell", NewCommand(noArgs(), run, WithUsage("Open a shell")))

	return d
}

func TestUnknownCommandSuggestions(t *testing.T) {
	t.Run("top-level typo suggests a top-level command", func(t *testing.T) {
		err := unknownCommandDispatcher().Execute([]string{"depoy"})
		require.Error(t, err)

		var uce *UnknownCommandError
		require.ErrorAs(t, err, &uce)
		assert.Equal(t, "depoy", uce.Name)
		assert.Equal(t, "", uce.ParentPath)
		assert.Equal(t, []string{"deploy"}, uce.Suggestions)

		assert.Equal(t, `unknown command "depoy"

Did you mean?
  deploy

Run 'myapp --help' to see available commands.`, err.Error())
	})

	t.Run("typo under a command with sub-commands", func(t *testing.T) {
		err := unknownCommandDispatcher().Execute([]string{"app", "destory"})
		require.Error(t, err)

		var uce *UnknownCommandError
		require.ErrorAs(t, err, &uce)
		assert.Equal(t, "destory", uce.Name)
		assert.Equal(t, "app", uce.ParentPath)

		assert.Equal(t, `unknown command "destory" for "myapp app"

Did you mean?
  destroy

Run 'myapp app --help' to see available commands.`, err.Error())
	})

	t.Run("typo under a section errors instead of silently showing help", func(t *testing.T) {
		err := unknownCommandDispatcher().Execute([]string{"runner", "upgrde"})
		require.Error(t, err)

		var uce *UnknownCommandError
		require.ErrorAs(t, err, &uce)
		assert.Equal(t, "upgrde", uce.Name)
		assert.Equal(t, "runner", uce.ParentPath)
		assert.Equal(t, []string{"upgrade"}, uce.Suggestions)
	})

	t.Run("typo under an unregistered namespace blames the right word", func(t *testing.T) {
		d := NewDispatcher("myapp")
		run := func(fs *FlagSet, args []string) error { return nil }
		// "debug" itself is never registered; it exists only as a prefix.
		d.Dispatch("debug entity", NewCommand(NewFlagSet("x"), run, WithUsage("Entities")))
		d.Dispatch("debug etcd", NewCommand(NewFlagSet("x"), run, WithUsage("Etcd")))

		err := d.Execute([]string{"debug", "entty"})
		require.Error(t, err)

		var uce *UnknownCommandError
		require.ErrorAs(t, err, &uce)
		assert.Equal(t, "entty", uce.Name, "should blame the bad word, not the valid namespace")
		assert.Equal(t, "debug", uce.ParentPath)
		assert.Equal(t, []string{"entity"}, uce.Suggestions)
	})

	t.Run("no close match still names the word", func(t *testing.T) {
		err := unknownCommandDispatcher().Execute([]string{"zzzzzzzz"})
		require.Error(t, err)

		var uce *UnknownCommandError
		require.ErrorAs(t, err, &uce)
		assert.Empty(t, uce.Suggestions)

		assert.Equal(t, `unknown command "zzzzzzzz"

Run 'myapp --help' to see available commands.`, err.Error())
		assert.NotContains(t, err.Error(), "Did you mean")
	})

	t.Run("only the first word is blamed", func(t *testing.T) {
		err := unknownCommandDispatcher().Execute([]string{"depoy", "myapp"})
		require.Error(t, err)

		var uce *UnknownCommandError
		require.ErrorAs(t, err, &uce)
		assert.Equal(t, "depoy", uce.Name)
	})
}

func TestUnknownCommandDoesNotFireOnValidInput(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"command with a positional accepts it", []string{"logs", "myapp"}},
		{"rest field accepts arbitrary words", []string{"exec", "echo", "hello"}},
		{"real sub-command still runs", []string{"app", "list"}},
		{"real nested sub-command still runs", []string{"runner", "upgrade"}},
		{"bare command with children shows help", []string{"app"}},
		// The dispatcher handles "help" itself; it is never a mistyped
		// sub-command, even under a command that tolerates unknown flags.
		{"trailing help keyword", []string{"runner", "help"}},
		{"leading help keyword", []string{"help", "runner"}},
		{"help keyword before a real sub-command", []string{"runner", "help", "upgrade"}},
		// An unknown flag's value is claimed by Parse, not blamed as a word.
		{"unknown value flag before a section", []string{"runner", "-C", "prod"}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.NoError(t, unknownCommandDispatcher().Execute(c.args))
		})
	}
}

func TestUnexpectedArgsWithoutSubCommands(t *testing.T) {
	// "version" has no children, so there is no sub-command to suggest. The
	// message should still avoid both the Go slice syntax and the misleading
	// "error parsing flags" prefix.
	d := NewDispatcher("myapp")
	d.Dispatch("version", NewCommand(NewFlagSet("version"),
		func(fs *FlagSet, args []string) error { return nil }, WithUsage("Print the version")))

	err := d.Execute([]string{"version", "foo"})
	require.Error(t, err)

	var uae *UnexpectedArgsError
	require.ErrorAs(t, err, &uae)
	assert.Equal(t, `unexpected argument "foo"`, err.Error())
	assert.NotContains(t, err.Error(), "error parsing flags")
}

func TestUnknownFlagSuggestions(t *testing.T) {
	newDispatcher := func() *Dispatcher {
		d := NewDispatcher("myapp")
		fs := NewFlagSet("deploy")
		fs.Bool("verbose", 'v', false, "Verbose output")
		fs.String("name", 'n', "", "Application name")
		fs.String("namespace", 0, "", "Namespace")
		d.Dispatch("deploy", NewCommand(fs,
			func(fs *FlagSet, args []string) error { return nil }, WithUsage("Deploy")))
		return d
	}

	t.Run("suggests a close long flag", func(t *testing.T) {
		err := newDispatcher().Execute([]string{"deploy", "--naem", "x"})
		require.Error(t, err)

		var ufe *UnknownFlagError
		require.ErrorAs(t, err, &ufe)
		assert.Equal(t, "--naem", ufe.Flag)
		assert.Equal(t, []string{"--name"}, ufe.Suggestions)

		assert.Equal(t, `unknown flag: --naem

Did you mean?
  --name`, err.Error())
	})

	t.Run("still unwraps to ErrUnknownFlag", func(t *testing.T) {
		err := newDispatcher().Execute([]string{"deploy", "--naem", "x"})
		assert.ErrorIs(t, err, ErrUnknownFlag)
	})

	t.Run("drops the redundant flag-parsing prefix", func(t *testing.T) {
		err := newDispatcher().Execute([]string{"deploy", "--naem", "x"})
		require.Error(t, err)
		assert.NotContains(t, err.Error(), "error parsing flags")
	})

	t.Run("nothing close offers no guess", func(t *testing.T) {
		err := newDispatcher().Execute([]string{"deploy", "--zzzzzzzz"})
		require.Error(t, err)
		assert.Equal(t, "unknown flag: --zzzzzzzz", err.Error())
	})

	t.Run("short flags carry no suggestions", func(t *testing.T) {
		err := newDispatcher().Execute([]string{"deploy", "-q"})
		require.Error(t, err)

		var ufe *UnknownFlagError
		require.ErrorAs(t, err, &ufe)
		assert.Equal(t, "-q", ufe.Flag)
		assert.Empty(t, ufe.Suggestions)
		assert.Equal(t, "unknown flag: -q", err.Error())
	})

	t.Run("a value flag keeps the parse-error prefix", func(t *testing.T) {
		d := NewDispatcher("myapp")
		fs := NewFlagSet("deploy")
		fs.Int("port", 'p', 0, "Port")
		d.Dispatch("deploy", NewCommand(fs,
			func(fs *FlagSet, args []string) error { return nil }, WithUsage("Deploy")))

		err := d.Execute([]string{"deploy", "--port", "notanumber"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error parsing flags")
	})
}
