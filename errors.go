package mflags

import (
	"fmt"
	"strings"
)

// UnknownCommandError reports a command word that matches nothing registered.
// It carries the surrounding context — which program, which parent command, and
// what the user might have meant — so callers can render it themselves rather
// than parsing the message back out of a string.
type UnknownCommandError struct {
	// Program is the binary name, e.g. "miren".
	Program string
	// ParentPath is the command the unknown word appeared under, e.g. "app".
	// Empty when the word was typed at the top level.
	ParentPath string
	// Name is the single word that matched nothing.
	Name string
	// Suggestions holds close matches, nearest first. May be empty.
	Suggestions []string
}

func (e *UnknownCommandError) Error() string {
	helpTarget := strings.TrimSpace(e.Program + " " + e.ParentPath)

	var sections []string

	if e.ParentPath != "" {
		sections = append(sections, fmt.Sprintf("unknown command %q for %q", e.Name, helpTarget))
	} else {
		sections = append(sections, fmt.Sprintf("unknown command %q", e.Name))
	}

	if len(e.Suggestions) > 0 {
		var b strings.Builder
		b.WriteString("Did you mean?")
		for _, s := range e.Suggestions {
			fmt.Fprintf(&b, "\n  %s", s)
		}
		sections = append(sections, b.String())
	}

	if helpTarget != "" {
		sections = append(sections, fmt.Sprintf("Run '%s --help' to see available commands.", helpTarget))
	}

	return strings.Join(sections, "\n\n")
}

// UnexpectedArgsError reports positional arguments a command has nowhere to put.
type UnexpectedArgsError struct {
	Args []string
}

func (e *UnexpectedArgsError) Error() string {
	if len(e.Args) == 1 {
		return fmt.Sprintf("unexpected argument %q", e.Args[0])
	}

	quoted := make([]string, 0, len(e.Args))
	for _, a := range e.Args {
		quoted = append(quoted, fmt.Sprintf("%q", a))
	}
	return "unexpected arguments: " + strings.Join(quoted, " ")
}
