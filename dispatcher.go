package mflags

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"golang.org/x/term"
)

var stdoutIsTerminal = sync.OnceValue(func() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
})

// faint returns the string wrapped in ANSI faint styling if stdout is a TTY
// and the NO_COLOR environment variable is not set.
func faint(s string) string {
	if os.Getenv("NO_COLOR") != "" {
		return s
	}
	if !stdoutIsTerminal() {
		return s
	}
	return "\033[2m" + s + "\033[0m"
}

func subCommandsLabel(n int) string {
	if n == 1 {
		return "(1 sub-command)"
	}
	return fmt.Sprintf("(%d sub-commands)", n)
}

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

	// Check for help flags anywhere in the arguments, but stop at --
	hasHelp := false
	for _, arg := range args {
		if arg == "--" {
			// Stop processing flags after --
			break
		}
		if arg == "-h" || arg == "--help" || arg == "help" {
			hasHelp = true
			break
		}
	}

	// Try to find the longest matching command, handling interspersed flags
	entry, allArgs := d.findCommandWithInterspersedFlags(args)

	// Check for non-flag arguments in the args AFTER the command
	// (to determine if help should be shown when allowUnknownFlags is true)
	hasOtherArgs := false
	for _, arg := range allArgs {
		if arg == "--" {
			break
		}
		if !strings.HasPrefix(arg, "-") {
			hasOtherArgs = true
			break
		}
	}

	if entry == nil {
		// No command found — check if the args form a namespace prefix
		var nonFlagArgs []string
		for _, arg := range args {
			if arg == "--" {
				break
			}
			if !strings.HasPrefix(arg, "-") && arg != "help" {
				nonFlagArgs = append(nonFlagArgs, arg)
			}
		}

		if len(nonFlagArgs) > 0 {
			namespacePath := normalizeCommandPath(strings.Join(nonFlagArgs, " "))
			if d.isNamespace(namespacePath) {
				return d.showNamespaceHelp(namespacePath)
			}
		}

		if hasHelp {
			return d.showHelp()
		}
		return fmt.Errorf("unknown command: %s", strings.Join(args, " "))
	}

	// Check if remaining non-flag args form a namespace under the matched command.
	// For example, if "debug" matched but the user typed "debug entity" and
	// "debug entity" is a namespace (has subcommands), show namespace help.
	var remainingNonFlags []string
	for _, arg := range allArgs {
		if arg == "--" {
			break
		}
		if !strings.HasPrefix(arg, "-") && arg != "help" {
			remainingNonFlags = append(remainingNonFlags, arg)
		}
	}

	if len(remainingNonFlags) > 0 {
		// Try progressively longer namespace paths from the remaining args
		for i := len(remainingNonFlags); i > 0; i-- {
			namespacePath := entry.Path + " " + strings.Join(remainingNonFlags[:i], " ")
			if d.isNamespace(namespacePath) {
				// Check if this namespace also has a registered command
				if cmdEntry, ok := d.commands[namespacePath]; ok {
					return d.showCommandHelp(cmdEntry)
				}
				return d.showNamespaceHelp(namespacePath)
			}
		}
	}

	// If help is requested, show command-specific help
	// BUT if the command allows unknown flags and there are other args, don't show help
	fs := entry.Command.FlagSet()
	shouldShowHelp := hasHelp
	if fs != nil && fs.allowUnknownFlags && hasOtherArgs {
		shouldShowHelp = false
	}

	if shouldShowHelp {
		return d.showCommandHelp(entry)
	}

	// Parse flags for this command
	// Disable automatic help in the FlagSet since the Dispatcher already handled it
	if fs != nil {
		fs.disableAutoHelp = true
	}
	if err := fs.Parse(allArgs); err != nil {
		return fmt.Errorf("error parsing flags: %w", err)
	}

	// Execute the command with the parsed flagset and remaining args
	err := entry.Command.Run(fs, fs.Args())

	// Check if command requested help to be shown
	if errors.Is(err, ErrShowHelp) {
		return d.showCommandHelp(entry)
	}

	return err
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

	children := d.getDirectChildren("")

	// Find max length for alignment
	maxLen := 0
	for _, child := range children {
		if len(child.Name) > maxLen {
			maxLen = len(child.Name)
		}
	}

	for _, child := range children {
		grandchildren := d.getDirectChildren(child.Path)
		suffix := ""
		if len(grandchildren) > 0 {
			suffix = " " + faint(subCommandsLabel(len(grandchildren)))
		}
		if child.Usage != "" {
			fmt.Printf("  %-*s  %s%s\n", maxLen+2, child.Name, child.Usage, suffix)
		} else if suffix != "" {
			fmt.Printf("  %-*s %s\n", maxLen+2, child.Name, suffix)
		} else {
			fmt.Printf("  %s\n", child.Name)
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
		// Show positional arguments by name
		if len(fs.posFields) > 0 {
			// Find max position to iterate in order
			maxPos := -1
			for pos := range fs.posFields {
				if pos > maxPos {
					maxPos = pos
				}
			}
			// Print each positional argument name
			for i := 0; i <= maxPos; i++ {
				if field, ok := fs.posFields[i]; ok {
					fmt.Printf(" <%s>", strings.ToLower(field.Name))
				}
			}
		}
		if fs.restField != nil {
			fmt.Print(" [arguments...]")
		}
	}
	fmt.Println()

	if entry.Usage != "" {
		fmt.Printf("\n%s\n", entry.Usage)
	}

	// Show positional arguments with descriptions if any have usage text
	if fs != nil && len(fs.posFields) > 0 {
		// Find max position to iterate in order
		maxPos := -1
		hasUsage := false
		for pos, field := range fs.posFields {
			if pos > maxPos {
				maxPos = pos
			}
			if field.Usage != "" {
				hasUsage = true
			}
		}

		if hasUsage {
			fmt.Println("\nArguments:")
			for i := 0; i <= maxPos; i++ {
				if field, ok := fs.posFields[i]; ok {
					argStr := fmt.Sprintf("  <%s>", strings.ToLower(field.Name))
					if field.Usage != "" {
						fmt.Printf("%-30s %s\n", argStr, field.Usage)
					} else {
						fmt.Println(argStr)
					}
				}
			}
		}
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
				flagStr += fmt.Sprintf(" <%s>", flag.Value.Type())
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

	// Show sub-commands if any exist (including implicit namespaces)
	children := d.getDirectChildren(entry.Path)
	if len(children) > 0 {
		fmt.Println("\nSub-commands:")

		// Find the maximum length for alignment
		maxLen := 0
		for _, child := range children {
			if len(child.Name) > maxLen {
				maxLen = len(child.Name)
			}
		}

		// Print sub-commands with usage
		for _, child := range children {
			grandchildren := d.getDirectChildren(child.Path)
			suffix := ""
			if len(grandchildren) > 0 {
				suffix = " " + faint(subCommandsLabel(len(grandchildren)))
			}
			if child.Usage != "" {
				fmt.Printf("  %-*s  %s%s\n", maxLen+2, child.Name, child.Usage, suffix)
			} else if suffix != "" {
				fmt.Printf("  %-*s %s\n", maxLen+2, child.Name, suffix)
			} else {
				fmt.Printf("  %s\n", child.Name)
			}
		}
	}

	return nil
}

// getSubCommands returns all direct sub-commands of a command
func (d *Dispatcher) getSubCommands(parentPath string) []*CommandEntry {
	var subCommands []*CommandEntry
	prefix := parentPath + " "

	// Find all commands that start with the parent path + space
	for path, entry := range d.commands {
		if strings.HasPrefix(path, prefix) {
			// Only include direct sub-commands (no further nesting)
			remainder := strings.TrimPrefix(path, prefix)
			if !strings.Contains(remainder, " ") {
				subCommands = append(subCommands, entry)
			}
		}
	}

	// Sort sub-commands by path
	sort.Slice(subCommands, func(i, j int) bool {
		return subCommands[i].Path < subCommands[j].Path
	})

	return subCommands
}

// ChildEntry represents a direct child of a command path, which may be either
// a registered command or an implicit namespace (prefix of deeper commands).
type ChildEntry struct {
	Name    string // The child name (single word)
	Path    string // The full path (parentPath + " " + Name, or just Name for top-level)
	Usage   string // Usage text (from registered command, or empty for namespaces)
	IsEntry bool   // True if this is a registered command, false if just a namespace
}

// getDirectChildren returns the direct children of a path, including both
// registered commands and implicit namespaces. If parentPath is empty, returns
// top-level entries.
func (d *Dispatcher) getDirectChildren(parentPath string) []ChildEntry {
	children := make(map[string]*ChildEntry)

	for path, entry := range d.commands {
		var remainder string
		if parentPath == "" {
			remainder = path
		} else {
			prefix := parentPath + " "
			if !strings.HasPrefix(path, prefix) {
				continue
			}
			remainder = strings.TrimPrefix(path, prefix)
		}

		parts := strings.Fields(remainder)
		if len(parts) == 0 {
			continue
		}

		childName := parts[0]
		childPath := childName
		if parentPath != "" {
			childPath = parentPath + " " + childName
		}

		if len(parts) == 1 {
			// Direct child command
			children[childName] = &ChildEntry{
				Name:    childName,
				Path:    childPath,
				Usage:   entry.Usage,
				IsEntry: true,
			}
		} else {
			// Deeper command — childName is a namespace (unless already registered)
			if _, ok := children[childName]; !ok {
				children[childName] = &ChildEntry{
					Name:    childName,
					Path:    childPath,
					Usage:   "",
					IsEntry: false,
				}
			}
		}
	}

	// Sort by name
	var result []ChildEntry
	for _, child := range children {
		result = append(result, *child)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

// isNamespace returns true if the given path is a prefix of any registered command
// (i.e., there are commands under this path even if the path itself isn't a command).
func (d *Dispatcher) isNamespace(path string) bool {
	prefix := path + " "
	for cmdPath := range d.commands {
		if strings.HasPrefix(cmdPath, prefix) {
			return true
		}
	}
	return false
}

// showNamespaceHelp displays help for an implicit namespace (a command prefix
// that isn't itself a registered command but has subcommands under it).
func (d *Dispatcher) showNamespaceHelp(namespacePath string) error {
	fmt.Printf("Usage: %s %s <command> [arguments]\n", d.name, namespacePath)

	children := d.getDirectChildren(namespacePath)

	if len(children) > 0 {
		fmt.Println("\nAvailable commands:")

		maxLen := 0
		for _, child := range children {
			if len(child.Name) > maxLen {
				maxLen = len(child.Name)
			}
		}

		for _, child := range children {
			grandchildren := d.getDirectChildren(child.Path)
			suffix := ""
			if len(grandchildren) > 0 {
				suffix = " " + faint(subCommandsLabel(len(grandchildren)))
			}
			if child.Usage != "" {
				fmt.Printf("  %-*s  %s%s\n", maxLen+2, child.Name, child.Usage, suffix)
			} else if suffix != "" {
				fmt.Printf("  %-*s %s\n", maxLen+2, child.Name, suffix)
			} else {
				fmt.Printf("  %s\n", child.Name)
			}
		}
	}

	fmt.Printf("\nUse '%s %s <command> --help' for more information about a command.\n", d.name, namespacePath)
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
