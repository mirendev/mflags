package mflags

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// Command is an interface for executable commands
type Command interface {
	// FlagSet returns the flagset for this command
	FlagSet() *FlagSet

	// Run executes the command with parsed flags and remaining arguments
	Run(fs *FlagSet, args []string) error

	// Usage returns the usage description for this command
	Usage() string
}

// OutputFormatter is an interface for commands that can specify their output format
type OutputFormatter interface {
	// OutputFormat returns the output format for this command
	OutputFormat() OutputFormat
}

// OutputFormat defines how a command formats its output
type OutputFormat string

const (
	OutputFormatRaw  OutputFormat = "raw"
	OutputFormatJSON OutputFormat = "json"
)

// funcCommand is a basic implementation of Command interface
type funcCommand struct {
	flags        *FlagSet
	handler      func(fs *FlagSet, args []string) error
	usage        string
	outputFormat OutputFormat
}

// CommandOption is a functional option for configuring a command
type CommandOption func(*funcCommand)

// WithUsage sets the usage description for the command
func WithUsage(usage string) CommandOption {
	return func(c *funcCommand) {
		c.usage = usage
	}
}

// WithOutputFormat sets the output format for the command
func WithOutputFormat(format OutputFormat) CommandOption {
	return func(c *funcCommand) {
		c.outputFormat = format
	}
}

// NewCommand creates a new command with the given options
func NewCommand(fs *FlagSet, handler func(fs *FlagSet, args []string) error, opts ...CommandOption) Command {
	c := &funcCommand{
		flags:        fs,
		handler:      handler,
		usage:        "",
		outputFormat: OutputFormatRaw, // Default to raw
	}

	// Apply options
	for _, opt := range opts {
		opt(c)
	}

	return c
}

// FlagSet returns the flagset for this command
func (c *funcCommand) FlagSet() *FlagSet {
	return c.flags
}

// Run executes the command
func (c *funcCommand) Run(fs *FlagSet, args []string) error {
	if c.handler != nil {
		return c.handler(fs, args)
	}
	return nil
}

// Usage returns the usage description for this command
func (c *funcCommand) Usage() string {
	return c.usage
}

// OutputFormat returns the output format for this command
func (c *funcCommand) OutputFormat() OutputFormat {
	return c.outputFormat
}

// SetOutputFormat sets the output format for this command
func (c *funcCommand) SetOutputFormat(format OutputFormat) {
	c.outputFormat = format
}

// CommandEntry represents a registered command entry
type CommandEntry struct {
	Path    string  // The command path (e.g., "foo bar")
	Command Command // The command implementation
	Usage   string  // Optional usage description
}

// Dispatcher manages command routing and execution
type Dispatcher struct {
	commands map[string]*CommandEntry
	name     string
}

// NewDispatcher creates a new command dispatcher
func NewDispatcher(name string) *Dispatcher {
	return &Dispatcher{
		commands: make(map[string]*CommandEntry),
		name:     name,
	}
}

// Dispatch registers a command
func (d *Dispatcher) Dispatch(path string, cmd Command) {
	// Normalize the path by trimming spaces and collapsing multiple spaces
	normalizedPath := normalizeCommandPath(path)

	d.commands[normalizedPath] = &CommandEntry{
		Path:    normalizedPath,
		Command: cmd,
		Usage:   cmd.Usage(),
	}
}

// Execute runs the dispatcher with the given arguments
func (d *Dispatcher) Execute(args []string) error {
	// Check for completion requests first
	if d.HandleCompletion(args) {
		return nil
	}

	if len(args) == 0 {
		return d.showHelp()
	}

	// Check for help flags anywhere in the arguments
	hasHelp := false
	for _, arg := range args {
		if arg == "-h" || arg == "--help" || arg == "help" {
			hasHelp = true
			break
		}
	}

	// Try to find the longest matching command, handling interspersed flags
	entry, allArgs := d.findCommandWithInterspersedFlags(args)

	if entry == nil {
		// No command found, check for help flags
		if hasHelp {
			return d.showHelp()
		}
		return fmt.Errorf("unknown command: %s", strings.Join(args, " "))
	}

	// If help is requested, show command-specific help
	if hasHelp {
		return d.showCommandHelp(entry)
	}

	// Parse flags for this command
	fs := entry.Command.FlagSet()
	if err := fs.Parse(allArgs); err != nil {
		return fmt.Errorf("error parsing flags: %w", err)
	}

	// Execute the command with the parsed flagset and remaining args
	return entry.Command.Run(fs, fs.Args())
}

// Run is an alias for Execute
func (d *Dispatcher) Run(args []string) error {
	return d.Execute(args)
}

// findCommand finds the best matching command for the given arguments
func (d *Dispatcher) findCommand(args []string) (*CommandEntry, []string) {
	// Try progressively shorter command paths
	for i := len(args); i > 0; i-- {
		path := normalizeCommandPath(strings.Join(args[:i], " "))
		if entry, ok := d.commands[path]; ok {
			return entry, args[i:]
		}
	}

	return nil, args
}

// findCommandWithInterspersedFlags finds a command while handling interspersed flags
func (d *Dispatcher) findCommandWithInterspersedFlags(args []string) (*CommandEntry, []string) {
	type flagInfo struct {
		flag     string
		value    string
		hasValue bool
		index    int // Track original position
	}

	// Track command parts and flags we encounter
	var commandParts []string
	var commandPartIndices []int // Track indices of command parts
	var skippedItems []flagInfo

	i := 0
	for i < len(args) {
		arg := args[i]

		// If this looks like a flag
		if strings.HasPrefix(arg, "-") {
			// Try to continue searching for command parts after this flag
			// We'll tentatively assume this flag might take an argument

			// Look ahead to see if there's a potential command part
			nextIsCommand := false
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				// Check if adding the next non-flag argument would form a valid command
				testParts := append(commandParts, args[i+1])
				testPath := normalizeCommandPath(strings.Join(testParts, " "))

				// Check if this forms a valid command or subcommand
				for cmdPath := range d.commands {
					if strings.HasPrefix(cmdPath, testPath) {
						nextIsCommand = true
						break
					}
				}
			}

			if nextIsCommand {
				// The next argument is part of the command, so this flag has no argument
				skippedItems = append(skippedItems, flagInfo{flag: arg, hasValue: false, index: i})
				i++
			} else if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				// The next argument might be a flag value
				skippedItems = append(skippedItems, flagInfo{
					flag:     arg,
					value:    args[i+1],
					hasValue: true,
					index:    i,
				})
				i += 2
			} else {
				// Flag with no argument (at end or followed by another flag)
				skippedItems = append(skippedItems, flagInfo{flag: arg, hasValue: false, index: i})
				i++
			}
		} else {
			// This is a potential command part
			commandParts = append(commandParts, arg)
			commandPartIndices = append(commandPartIndices, i)
			i++
		}
	}

	// Now try to find the longest matching command from the command parts we collected
	for j := len(commandParts); j > 0; j-- {
		testPath := normalizeCommandPath(strings.Join(commandParts[:j], " "))
		if entry, ok := d.commands[testPath]; ok {
			// We found a command! Now build the args for it
			fs := entry.Command.FlagSet()

			// Figure out where the command ends in the original args
			lastCommandIndex := -1
			if j > 0 && j <= len(commandPartIndices) {
				lastCommandIndex = commandPartIndices[j-1]
			}

			// Build the full argument list for this command
			// Include interspersed flags that came before the command and everything after
			var fullArgs []string

			// Add flags that were interspersed before/during the command
			for _, fi := range skippedItems {
				if fi.index <= lastCommandIndex {
					fullArgs = append(fullArgs, fi.flag)
					if fi.hasValue {
						fullArgs = append(fullArgs, fi.value)
					}
				}
			}

			// Add everything after the command
			if lastCommandIndex >= 0 && lastCommandIndex+1 < len(args) {
				fullArgs = append(fullArgs, args[lastCommandIndex+1:]...)
			}

			// Try to validate our flag assumptions were correct
			// Only validate flags that came before the command
			valid := true
			for _, fi := range skippedItems {
				// Only check flags that came before the end of the command
				if fi.index > lastCommandIndex {
					continue // This flag is after the command, will be handled by Parse
				}

				flagName := strings.TrimPrefix(fi.flag, "--")
				flagName = strings.TrimPrefix(flagName, "-")

				// Check if this flag exists in the command's flagset
				flagFound := false
				fs.VisitAll(func(f *Flag) {
					if (len(flagName) == 1 && f.Short == rune(flagName[0])) || f.Name == flagName {
						flagFound = true
						// Check if our assumption about the flag taking a value was correct
						if fi.hasValue && f.Value.IsBool() {
							valid = false // Bool flags don't take values
						}
					}
				})

				if !flagFound && !isHelpFlag(fi.flag) {
					// Unknown flag (unless it's a help flag which is always valid)
					valid = false
				}
			}

			if valid {
				return entry, fullArgs
			}
		}
	}

	// No valid command found
	return nil, args
}

// isHelpFlag checks if a flag is a help flag
func isHelpFlag(flag string) bool {
	return flag == "-h" || flag == "--help"
}

// normalizeCommandPath normalizes a command path for consistent lookup
func normalizeCommandPath(path string) string {
	// Split by spaces, filter empty strings, and rejoin
	parts := strings.Fields(path)
	return strings.Join(parts, " ")
}

// showHelp displays available commands
func (d *Dispatcher) showHelp() error {
	fmt.Printf("Usage: %s <command> [arguments]\n\n", d.name)
	fmt.Println("Available commands:")

	// Collect and sort command paths
	var paths []string
	maxLen := 0
	for path := range d.commands {
		paths = append(paths, path)
		if len(path) > maxLen {
			maxLen = len(path)
		}
	}

	// Sort paths for consistent output
	sortedPaths := make([]string, len(paths))
	copy(sortedPaths, paths)
	for i := 0; i < len(sortedPaths); i++ {
		for j := i + 1; j < len(sortedPaths); j++ {
			if sortedPaths[i] > sortedPaths[j] {
				sortedPaths[i], sortedPaths[j] = sortedPaths[j], sortedPaths[i]
			}
		}
	}

	// Print commands with usage
	for _, path := range sortedPaths {
		entry := d.commands[path]
		if entry.Usage != "" {
			fmt.Printf("  %-*s  %s\n", maxLen+2, path, entry.Usage)
		} else {
			fmt.Printf("  %s\n", path)
		}
	}

	fmt.Println("\nUse '<command> --help' for more information about a command.")
	return nil
}

// showCommandHelp displays help for a specific command
func (d *Dispatcher) showCommandHelp(entry *CommandEntry) error {
	fmt.Printf("Usage: %s %s [options]", d.name, entry.Path)
	fs := entry.Command.FlagSet()
	if fs != nil {
		// Check if there are positional arguments expected
		hasPositional := false
		if len(fs.posFields) > 0 {
			hasPositional = true
		}
		if fs.restField != nil {
			hasPositional = true
		}
		if hasPositional {
			fmt.Print(" [arguments]")
		}
	}
	fmt.Println()

	if entry.Usage != "" {
		fmt.Printf("\n%s\n", entry.Usage)
	}

	// Show flags if any are defined
	if fs != nil {
		hasFlags := false
		fs.VisitAll(func(flag *Flag) {
			if !hasFlags {
				fmt.Println("\nOptions:")
				hasFlags = true
			}

			// Format flag display
			var flagStr string
			if flag.Short != 0 && flag.Name != "" {
				flagStr = fmt.Sprintf("  -%c, --%s", flag.Short, flag.Name)
			} else if flag.Short != 0 {
				flagStr = fmt.Sprintf("  -%c", flag.Short)
			} else {
				flagStr = fmt.Sprintf("      --%s", flag.Name)
			}

			// Add value placeholder for non-boolean flags
			if !flag.Value.IsBool() {
				flagStr += " <value>"
			}

			// Print flag with usage
			if flag.Usage != "" {
				fmt.Printf("%-30s %s", flagStr, flag.Usage)
				if flag.DefValue != "" && flag.DefValue != "false" && flag.DefValue != "0" {
					fmt.Printf(" (default: %s)", flag.DefValue)
				}
				fmt.Println()
			} else {
				fmt.Println(flagStr)
			}
		})
	}

	return nil
}

// GetCommand returns the command for a given path, or nil if not found
func (d *Dispatcher) GetCommand(path string) Command {
	normalizedPath := normalizeCommandPath(path)
	if entry, ok := d.commands[normalizedPath]; ok {
		return entry.Command
	}
	return nil
}

// GetCommandEntry returns the command entry for a given path, or nil if not found
func (d *Dispatcher) GetCommandEntry(path string) *CommandEntry {
	normalizedPath := normalizeCommandPath(path)
	return d.commands[normalizedPath]
}

// GetCommands returns all registered commands
func (d *Dispatcher) GetCommands() map[string]Command {
	// Return a copy to prevent external modification
	result := make(map[string]Command)
	for k, v := range d.commands {
		result[k] = v.Command
	}
	return result
}

// HasCommand checks if a command is registered
func (d *Dispatcher) HasCommand(path string) bool {
	normalizedPath := normalizeCommandPath(path)
	_, exists := d.commands[normalizedPath]
	return exists
}

// Remove unregisters a command
func (d *Dispatcher) Remove(path string) {
	normalizedPath := normalizeCommandPath(path)
	delete(d.commands, normalizedPath)
}

// GetCommandCompletions returns completions for commands based on the prefix
func (d *Dispatcher) GetCommandCompletions(prefix string) []Completion {
	var completions []Completion

	// Normalize the prefix
	normalizedPrefix := normalizeCommandPath(prefix)

	for path, entry := range d.commands {
		// Check if the command path starts with the prefix
		if strings.HasPrefix(path, normalizedPrefix) {
			completions = append(completions, Completion{
				Value:       path,
				Description: entry.Usage,
				IsBool:      false, // Commands are not boolean flags
			})
		}
	}

	// Sort completions
	sort.Slice(completions, func(i, j int) bool {
		return completions[i].Value < completions[j].Value
	})

	return completions
}

// HandleCompletion handles shell completion requests for the dispatcher
// Returns true if a completion request was handled
func (d *Dispatcher) HandleCompletion(args []string) bool {
	// Check for bash completion mode
	if os.Getenv("COMP_LINE") != "" {
		// We're being called by bash completion
		d.PrintBashCompletions(args)
		return true
	}

	// Check for explicit completion flags
	if len(args) > 0 {
		switch args[0] {
		case "--complete-bash":
			d.PrintBashCompletions(args[1:])
			return true
		case "--complete-zsh":
			d.PrintZshCompletions(args[1:])
			return true
		case "--generate-bash-completion":
			fmt.Print(d.GenerateBashCompletion())
			return true
		case "--generate-zsh-completion":
			fmt.Print(d.GenerateZshCompletion())
			return true
		}
	}

	return false
}

// PrintBashCompletions outputs completions in bash format
func (d *Dispatcher) PrintBashCompletions(args []string) {
	// Determine what we're completing
	if len(args) == 0 {
		// Complete commands
		completions := d.GetCommandCompletions("")
		for _, comp := range completions {
			fmt.Println(comp.Value)
		}
		return
	}

	// Try to find the command being completed
	currentWord := ""
	if len(args) > 0 {
		currentWord = args[len(args)-1]
	}

	// First, check if we're completing a partial command
	entry, remainingArgs := d.findCommand(args)

	if entry == nil {
		// No exact command match, show command completions
		prefix := strings.Join(args, " ")
		completions := d.GetCommandCompletions(prefix)
		for _, comp := range completions {
			fmt.Println(comp.Value)
		}
	} else {
		// We have a command, complete its flags
		fs := entry.Command.FlagSet()
		if fs != nil {
			// Check if we need to complete a flag value
			if len(remainingArgs) >= 2 {
				prevArg := remainingArgs[len(remainingArgs)-2]
				if strings.HasPrefix(prevArg, "-") {
					// Check if previous arg was a flag that needs a value
					flagName := strings.TrimLeft(prevArg, "-")

					// Check long flags
					if flag, ok := fs.flags[flagName]; ok && !flag.Value.IsBool() {
						// We're completing a value for this flag
						return
					}

					// Check short flags
					if len(prevArg) == 2 {
						if flag, ok := fs.shortMap[rune(prevArg[1])]; ok && !flag.Value.IsBool() {
							// We're completing a value for this flag
							return
						}
					}
				}
			}

			// Get flag completions
			completions := fs.GetFlagCompletions(currentWord)
			for _, comp := range completions {
				fmt.Println(comp.Value)
			}
		}
	}
}

// PrintZshCompletions outputs completions in zsh format
func (d *Dispatcher) PrintZshCompletions(args []string) {
	// Get all command completions
	commandCompletions := d.GetCommandCompletions("")

	// Print command completions
	for _, comp := range commandCompletions {
		if comp.Description != "" {
			fmt.Printf("%s:%s\n", comp.Value, comp.Description)
		} else {
			fmt.Println(comp.Value)
		}
	}

	// If we have a specific command, also show its flags
	if len(args) > 0 {
		entry, _ := d.findCommand(args)
		if entry != nil {
			fs := entry.Command.FlagSet()
			if fs != nil {
				flagCompletions := fs.GetFlagCompletions("")
				for _, comp := range flagCompletions {
					if comp.Description != "" {
						fmt.Printf("%s:%s\n", comp.Value, comp.Description)
					} else {
						fmt.Println(comp.Value)
					}
				}
			}
		}
	}
}

// GenerateBashCompletion generates a bash completion script for the dispatcher
func (d *Dispatcher) GenerateBashCompletion() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Bash completion for %s\n", d.name))
	sb.WriteString(fmt.Sprintf("_%s_completion() {\n", d.name))
	sb.WriteString("    local cur prev words cword\n")
	sb.WriteString("    _init_completion || return\n\n")
	sb.WriteString("    # Get completions from the program\n")
	sb.WriteString(fmt.Sprintf("    local completions=$(%s --complete-bash \"${COMP_WORDS[@]:1:$COMP_CWORD}\")\n", d.name))
	sb.WriteString("    COMPREPLY=( $(compgen -W \"$completions\" -- \"$cur\") )\n")
	sb.WriteString("}\n\n")
	sb.WriteString(fmt.Sprintf("complete -F _%s_completion %s\n", d.name, d.name))

	return sb.String()
}

// GenerateZshCompletion generates a zsh completion script for the dispatcher
func (d *Dispatcher) GenerateZshCompletion() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("#compdef %s\n\n", d.name))
	sb.WriteString(fmt.Sprintf("_%s() {\n", d.name))
	sb.WriteString("    local -a commands\n")
	sb.WriteString("    commands=(\n")

	// Add all commands with descriptions
	for path, entry := range d.commands {
		desc := strings.ReplaceAll(entry.Usage, "'", "'\"'\"'")
		if desc != "" {
			sb.WriteString(fmt.Sprintf("        '%s[%s]'\n", path, desc))
		} else {
			sb.WriteString(fmt.Sprintf("        '%s'\n", path))
		}
	}

	sb.WriteString("    )\n\n")
	sb.WriteString("    # If we have a command, complete its flags\n")
	sb.WriteString("    if (( CURRENT > 2 )); then\n")
	sb.WriteString("        # Get the command\n")
	sb.WriteString("        local cmd=\"${words[2]}\"\n")
	sb.WriteString("        # TODO: Add flag completion for specific commands\n")
	sb.WriteString("    fi\n\n")
	sb.WriteString("    _describe 'command' commands\n")
	sb.WriteString("}\n\n")
	sb.WriteString(fmt.Sprintf("_%s\n", d.name))

	return sb.String()
}
