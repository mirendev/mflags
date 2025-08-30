package mflags

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// Completion represents a single completion suggestion
type Completion struct {
	Value       string // The completion value (e.g., "--verbose" or "-v")
	Description string // Optional description for the completion
	IsBool      bool   // Whether this flag takes no argument
}

// VisitAll calls fn for each flag in lexicographical order
func (f *FlagSet) VisitAll(fn func(*Flag)) {
	// Collect all flags
	var flags []*Flag
	seen := make(map[*Flag]bool)

	for _, flag := range f.flags {
		if !seen[flag] {
			flags = append(flags, flag)
			seen[flag] = true
		}
	}

	// Sort by name
	sort.Slice(flags, func(i, j int) bool {
		return flags[i].Name < flags[j].Name
	})

	// Call function for each flag
	for _, flag := range flags {
		fn(flag)
	}
}

// GetLongFlags returns all long flag names with "--" prefix
func (f *FlagSet) GetLongFlags() []string {
	var flags []string
	for name := range f.flags {
		if name != "" {
			flags = append(flags, "--"+name)
		}
	}
	sort.Strings(flags)
	return flags
}

// GetShortFlags returns all short flag names with "-" prefix
func (f *FlagSet) GetShortFlags() []string {
	var flags []string
	seen := make(map[rune]bool)

	for r := range f.shortMap {
		if r != 0 && !seen[r] {
			flags = append(flags, fmt.Sprintf("-%c", r))
			seen[r] = true
		}
	}
	sort.Strings(flags)
	return flags
}

// GetFlagCompletions returns completions for the current context
func (f *FlagSet) GetFlagCompletions(prefix string) []Completion {
	var completions []Completion

	// Handle different prefix types
	if strings.HasPrefix(prefix, "--") {
		// Long flag completion
		search := prefix[2:]
		for name, flag := range f.flags {
			if name != "" && strings.HasPrefix(name, search) {
				completions = append(completions, Completion{
					Value:       "--" + name,
					Description: flag.Usage,
					IsBool:      flag.Value.IsBool(),
				})
			}
		}
	} else if strings.HasPrefix(prefix, "-") && len(prefix) <= 2 {
		// Short flag completion
		if len(prefix) == 1 {
			// Show all short flags
			for r, flag := range f.shortMap {
				completions = append(completions, Completion{
					Value:       fmt.Sprintf("-%c", r),
					Description: flag.Usage,
					IsBool:      flag.Value.IsBool(),
				})
			}
		} else {
			// Filter by the character after -
			search := rune(prefix[1])
			if flag, ok := f.shortMap[search]; ok {
				completions = append(completions, Completion{
					Value:       prefix,
					Description: flag.Usage,
					IsBool:      flag.Value.IsBool(),
				})
			}
		}
	} else if prefix == "" {
		// No prefix, show all flags
		for name, flag := range f.flags {
			if name != "" {
				completions = append(completions, Completion{
					Value:       "--" + name,
					Description: flag.Usage,
					IsBool:      flag.Value.IsBool(),
				})
			}
		}
		for r, flag := range f.shortMap {
			completions = append(completions, Completion{
				Value:       fmt.Sprintf("-%c", r),
				Description: flag.Usage,
				IsBool:      flag.Value.IsBool(),
			})
		}
	}

	// Sort completions
	sort.Slice(completions, func(i, j int) bool {
		return completions[i].Value < completions[j].Value
	})

	return completions
}

// PrintBashCompletions outputs completions in bash format
func (f *FlagSet) PrintBashCompletions(args []string) {
	// Determine what we're completing
	if len(args) == 0 {
		return
	}

	// Get the current word being completed
	currentWord := ""
	if len(args) > 0 {
		currentWord = args[len(args)-1]
	}

	// Check if we're completing a flag value
	if len(args) >= 2 {
		prevArg := args[len(args)-2]
		if strings.HasPrefix(prevArg, "-") {
			// Check if previous arg was a flag that needs a value
			flagName := strings.TrimLeft(prevArg, "-")

			// Check long flags
			if flag, ok := f.flags[flagName]; ok && !flag.Value.IsBool() {
				// We're completing a value for this flag
				// For now, we don't provide value completions
				return
			}

			// Check short flags
			if len(prevArg) == 2 {
				if flag, ok := f.shortMap[rune(prevArg[1])]; ok && !flag.Value.IsBool() {
					// We're completing a value for this flag
					return
				}
			}
		}
	}

	// Get completions for flags
	completions := f.GetFlagCompletions(currentWord)

	// Print completions (one per line for bash)
	for _, comp := range completions {
		fmt.Println(comp.Value)
	}
}

// PrintZshCompletions outputs completions in zsh format
func (f *FlagSet) PrintZshCompletions(args []string) {
	// Get all completions
	completions := f.GetFlagCompletions("")

	// Print in zsh format with descriptions
	for _, comp := range completions {
		if comp.Description != "" {
			fmt.Printf("%s:%s\n", comp.Value, comp.Description)
		} else {
			fmt.Println(comp.Value)
		}
	}
}

// GenerateBashCompletion generates a bash completion script
func (f *FlagSet) GenerateBashCompletion(programName string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Bash completion for %s\n", programName))
	sb.WriteString(fmt.Sprintf("_%s_completion() {\n", programName))
	sb.WriteString("    local cur prev words cword\n")
	sb.WriteString("    _init_completion || return\n\n")
	sb.WriteString("    # Get flag completions from the program\n")
	sb.WriteString(fmt.Sprintf("    local completions=$(%s --complete-bash \"${COMP_WORDS[@]:1:$COMP_CWORD}\")\n", programName))
	sb.WriteString("    COMPREPLY=( $(compgen -W \"$completions\" -- \"$cur\") )\n")
	sb.WriteString("}\n\n")
	sb.WriteString(fmt.Sprintf("complete -F _%s_completion %s\n", programName, programName))

	return sb.String()
}

// GenerateZshCompletion generates a zsh completion script
func (f *FlagSet) GenerateZshCompletion(programName string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("#compdef %s\n\n", programName))
	sb.WriteString(fmt.Sprintf("_%s() {\n", programName))
	sb.WriteString("    local -a flags\n")
	sb.WriteString("    flags=(\n")

	// Add all flags with descriptions
	f.VisitAll(func(flag *Flag) {
		if flag.Name != "" {
			desc := strings.ReplaceAll(flag.Usage, "'", "'\"'\"'")
			if flag.Value.IsBool() {
				sb.WriteString(fmt.Sprintf("        '--%s[%s]'\n", flag.Name, desc))
			} else {
				sb.WriteString(fmt.Sprintf("        '--%s=[%s]:value'\n", flag.Name, desc))
			}
		}
		if flag.Short != 0 {
			desc := strings.ReplaceAll(flag.Usage, "'", "'\"'\"'")
			if flag.Value.IsBool() {
				sb.WriteString(fmt.Sprintf("        '-%c[%s]'\n", flag.Short, desc))
			} else {
				sb.WriteString(fmt.Sprintf("        '-%c[%s]:value'\n", flag.Short, desc))
			}
		}
	})

	sb.WriteString("    )\n")
	sb.WriteString("    _arguments -s $flags\n")
	sb.WriteString("}\n\n")
	sb.WriteString(fmt.Sprintf("_%s\n", programName))

	return sb.String()
}

// HandleCompletion checks for completion requests and handles them
// Returns true if a completion request was handled
func (f *FlagSet) HandleCompletion(args []string) bool {
	// Check for bash completion mode
	if os.Getenv("COMP_LINE") != "" {
		// We're being called by bash completion
		f.PrintBashCompletions(args)
		return true
	}

	// Check for explicit completion flags
	if len(args) > 0 {
		switch args[0] {
		case "--complete-bash":
			f.PrintBashCompletions(args[1:])
			return true
		case "--complete-zsh":
			f.PrintZshCompletions(args[1:])
			return true
		case "--generate-bash-completion":
			programName := "program"
			if f.name != "" {
				programName = f.name
			}
			fmt.Print(f.GenerateBashCompletion(programName))
			return true
		case "--generate-zsh-completion":
			programName := "program"
			if f.name != "" {
				programName = f.name
			}
			fmt.Print(f.GenerateZshCompletion(programName))
			return true
		}
	}

	return false
}
