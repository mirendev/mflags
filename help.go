package mflags

import (
	"fmt"
	"strings"
)

// formatFlagLine formats a single flag for help output.
func formatFlagLine(flag *Flag) string {
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

	// Format with usage
	if flag.Usage != "" {
		line := fmt.Sprintf("%-30s %s", flagStr, flag.Usage)
		if flag.DefValue != "" && flag.DefValue != "false" && flag.DefValue != "0" {
			line += fmt.Sprintf(" (default: %s)", flag.DefValue)
		}
		if flag.EnvVar != "" {
			line += fmt.Sprintf(" (env: %s)", flag.EnvVar)
		}
		return line
	}
	return flagStr
}

// WriteFlagHelp writes group-aware flag help output to stdout.
// Named groups are printed first in insertion order, followed by the default
// (unnamed) group under "Options:".
func (f *FlagSet) WriteFlagHelp() {
	// Collect flags into groups, preserving VisitAll sort order
	type groupFlags struct {
		name  string
		flags []*Flag
	}

	groupMap := make(map[string]*groupFlags)
	var defaultFlags []*Flag

	f.VisitAll(func(flag *Flag) {
		if flag.Hidden {
			return
		}
		if flag.Group == "" {
			defaultFlags = append(defaultFlags, flag)
		} else {
			gf, ok := groupMap[flag.Group]
			if !ok {
				gf = &groupFlags{name: flag.Group}
				groupMap[flag.Group] = gf
			}
			gf.flags = append(gf.flags, flag)
		}
	})

	// Print named groups in insertion order
	for _, groupName := range f.groupOrder {
		gf, ok := groupMap[groupName]
		if !ok || len(gf.flags) == 0 {
			continue
		}
		fmt.Printf("\n%s:\n", gf.name)
		for _, flag := range gf.flags {
			fmt.Println(formatFlagLine(flag))
		}
	}

	// Print default group last
	if len(defaultFlags) > 0 {
		fmt.Println("\nOptions:")
		for _, flag := range defaultFlags {
			fmt.Println(formatFlagLine(flag))
		}
	}
}

// ShowHelp displays help information for the flag set, including all defined flags
// and their usage information.
func (f *FlagSet) ShowHelp() {
	if f.name != "" {
		fmt.Printf("Usage: %s [options]", f.name)
		// Show positional arguments by name
		if len(f.posFields) > 0 {
			// Find max position to iterate in order
			maxPos := -1
			for pos := range f.posFields {
				if pos > maxPos {
					maxPos = pos
				}
			}
			// Print each positional argument name
			for i := 0; i <= maxPos; i++ {
				if field, ok := f.posFields[i]; ok {
					fmt.Printf(" <%s>", strings.ToLower(field.Name))
				}
			}
		}
		if f.restField != nil {
			fmt.Print(" [arguments...]")
		}
		fmt.Println()
	}

	// Show positional arguments with descriptions if any have usage text
	if len(f.posFields) > 0 {
		// Find max position to iterate in order
		maxPos := -1
		hasUsage := false
		for pos, field := range f.posFields {
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
				if field, ok := f.posFields[i]; ok {
					argStr := fmt.Sprintf("  <%s>", strings.ToLower(field.Name))
					if field.Usage != "" {
						line := fmt.Sprintf("%-30s %s", argStr, field.Usage)
						if field.EnvVar != "" {
							line += fmt.Sprintf(" (env: %s)", field.EnvVar)
						}
						fmt.Println(line)
					} else {
						fmt.Println(argStr)
					}
				}
			}
		}
	}

	f.WriteFlagHelp()
}
