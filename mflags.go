package mflags

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

var (
	ErrUnknownFlag   = errors.New("unknown flag")
	ErrMissingValue  = errors.New("flag needs an argument")
	ErrInvalidValue  = errors.New("invalid flag value")
	ErrInvalidChoice = errors.New("invalid choice")
	ErrHelp          = errors.New("help requested")
	ErrShowHelp      = errors.New("show help") // Return from Command.Run to trigger help display
	ErrRequired      = errors.New("required flag not provided")
)

// PositionalField represents a positional argument field
type PositionalField struct {
	Name     string        // Field name (e.g., "Command", "Target")
	Usage    string        // Usage description for help output
	Value    reflect.Value // The reflect.Value of the field
	Type     reflect.Type  // The type of the field
	EnvVar   string        // environment variable name (from env:"..." struct tag)
	Required bool          // whether this positional must be provided
	HasValue bool          // true if value was set by env var or CLI arg
}

type FlagSet struct {
	name              string
	flags             map[string]*Flag
	shortMap          map[rune]*Flag
	allFlags          []*Flag // All registered flags (for iteration)
	args              []string
	parsed            bool
	restField         *[]string                // Pointer to field marked with "rest" tag
	posFields         map[int]*PositionalField // Map of position to positional field info
	allowUnknownFlags bool                     // If true, accumulate unknown flags instead of erroring
	unknownFlags      []string                 // Accumulated unknown flags when allowUnknownFlags is true
	unknownField      *[]string                // Pointer to field marked with "unknown" tag
	disableAutoHelp   bool                     // If true, don't automatically handle -h/--help in Parse
	currentGroup      string                   // ambient group name set by FromStruct options or Group() calls
	groupOrder        []string                 // ordered list of distinct group names (insertion order)
	requiredFlags     []string                 // flag names marked required:"true"
	requiredPos       []int                    // positional indices marked required:"true"
}

type Flag struct {
	Name     string
	Short    rune
	Usage    string
	Value    Value
	DefValue string
	Group    string // group name for help rendering; empty = default "Options:"
	EnvVar   string // environment variable name (from env:"..." struct tag)
	Required bool   // whether this flag must be provided
	HasValue bool   // true if value was set by env var or CLI arg
}

type Value interface {
	String() string
	Set(string) error
	IsBool() bool
	Type() string
}

// NewFlagSet returns a new, empty flag set with the specified name.
// The name is used for error messages and help output.
func NewFlagSet(name string) *FlagSet {
	return &FlagSet{
		name:      name,
		flags:     make(map[string]*Flag),
		shortMap:  make(map[rune]*Flag),
		posFields: make(map[int]*PositionalField),
	}
}

// BoolVar defines a bool flag with the specified name, short form, default value, and usage string.
// The argument p points to a bool variable in which to store the value of the flag.
func (f *FlagSet) BoolVar(p *bool, name string, short rune, value bool, usage string) {
	*p = value
	f.Var((*boolValue)(p), name, short, usage)
}

// Bool defines a bool flag with the specified name, short form, default value, and usage string.
// The return value is the address of a bool variable that stores the value of the flag.
func (f *FlagSet) Bool(name string, short rune, value bool, usage string) *bool {
	p := new(bool)
	f.BoolVar(p, name, short, value, usage)
	return p
}

// StringVar defines a string flag with the specified name, short form, default value, and usage string.
// The argument p points to a string variable in which to store the value of the flag.
func (f *FlagSet) StringVar(p *string, name string, short rune, value string, usage string) {
	*p = value
	f.Var((*stringValue)(p), name, short, usage)
}

// String defines a string flag with the specified name, short form, default value, and usage string.
// The return value is the address of a string variable that stores the value of the flag.
func (f *FlagSet) String(name string, short rune, value string, usage string) *string {
	p := new(string)
	f.StringVar(p, name, short, value, usage)
	return p
}

// IntVar defines an int flag with the specified name, short form, default value, and usage string.
// The argument p points to an int variable in which to store the value of the flag.
func (f *FlagSet) IntVar(p *int, name string, short rune, value int, usage string) {
	*p = value
	f.Var((*intValue)(p), name, short, usage)
}

// Int defines an int flag with the specified name, short form, default value, and usage string.
// The return value is the address of an int variable that stores the value of the flag.
func (f *FlagSet) Int(name string, short rune, value int, usage string) *int {
	p := new(int)
	f.IntVar(p, name, short, value, usage)
	return p
}

// Int64Var defines an int64 flag with the specified name, short form, default value, and usage string.
// The argument p points to an int64 variable in which to store the value of the flag.
func (f *FlagSet) Int64Var(p *int64, name string, short rune, value int64, usage string) {
	*p = value
	f.Var((*int64Value)(p), name, short, usage)
}

// Int64 defines an int64 flag with the specified name, short form, default value, and usage string.
// The return value is the address of an int64 variable that stores the value of the flag.
func (f *FlagSet) Int64(name string, short rune, value int64, usage string) *int64 {
	p := new(int64)
	f.Int64Var(p, name, short, value, usage)
	return p
}

// Int8Var defines an int8 flag with the specified name, short form, default value, and usage string.
// The argument p points to an int8 variable in which to store the value of the flag.
func (f *FlagSet) Int8Var(p *int8, name string, short rune, value int8, usage string) {
	*p = value
	f.Var((*int8Value)(p), name, short, usage)
}

// Int8 defines an int8 flag with the specified name, short form, default value, and usage string.
// The return value is the address of an int8 variable that stores the value of the flag.
func (f *FlagSet) Int8(name string, short rune, value int8, usage string) *int8 {
	p := new(int8)
	f.Int8Var(p, name, short, value, usage)
	return p
}

// Int16Var defines an int16 flag with the specified name, short form, default value, and usage string.
// The argument p points to an int16 variable in which to store the value of the flag.
func (f *FlagSet) Int16Var(p *int16, name string, short rune, value int16, usage string) {
	*p = value
	f.Var((*int16Value)(p), name, short, usage)
}

// Int16 defines an int16 flag with the specified name, short form, default value, and usage string.
// The return value is the address of an int16 variable that stores the value of the flag.
func (f *FlagSet) Int16(name string, short rune, value int16, usage string) *int16 {
	p := new(int16)
	f.Int16Var(p, name, short, value, usage)
	return p
}

// Int32Var defines an int32 flag with the specified name, short form, default value, and usage string.
// The argument p points to an int32 variable in which to store the value of the flag.
func (f *FlagSet) Int32Var(p *int32, name string, short rune, value int32, usage string) {
	*p = value
	f.Var((*int32Value)(p), name, short, usage)
}

// Int32 defines an int32 flag with the specified name, short form, default value, and usage string.
// The return value is the address of an int32 variable that stores the value of the flag.
func (f *FlagSet) Int32(name string, short rune, value int32, usage string) *int32 {
	p := new(int32)
	f.Int32Var(p, name, short, value, usage)
	return p
}

// UintVar defines a uint flag with the specified name, short form, default value, and usage string.
// The argument p points to a uint variable in which to store the value of the flag.
func (f *FlagSet) UintVar(p *uint, name string, short rune, value uint, usage string) {
	*p = value
	f.Var((*uintValue)(p), name, short, usage)
}

// Uint defines a uint flag with the specified name, short form, default value, and usage string.
// The return value is the address of a uint variable that stores the value of the flag.
func (f *FlagSet) Uint(name string, short rune, value uint, usage string) *uint {
	p := new(uint)
	f.UintVar(p, name, short, value, usage)
	return p
}

// Uint8Var defines a uint8 flag with the specified name, short form, default value, and usage string.
// The argument p points to a uint8 variable in which to store the value of the flag.
func (f *FlagSet) Uint8Var(p *uint8, name string, short rune, value uint8, usage string) {
	*p = value
	f.Var((*uint8Value)(p), name, short, usage)
}

// Uint8 defines a uint8 flag with the specified name, short form, default value, and usage string.
// The return value is the address of a uint8 variable that stores the value of the flag.
func (f *FlagSet) Uint8(name string, short rune, value uint8, usage string) *uint8 {
	p := new(uint8)
	f.Uint8Var(p, name, short, value, usage)
	return p
}

// Uint16Var defines a uint16 flag with the specified name, short form, default value, and usage string.
// The argument p points to a uint16 variable in which to store the value of the flag.
func (f *FlagSet) Uint16Var(p *uint16, name string, short rune, value uint16, usage string) {
	*p = value
	f.Var((*uint16Value)(p), name, short, usage)
}

// Uint16 defines a uint16 flag with the specified name, short form, default value, and usage string.
// The return value is the address of a uint16 variable that stores the value of the flag.
func (f *FlagSet) Uint16(name string, short rune, value uint16, usage string) *uint16 {
	p := new(uint16)
	f.Uint16Var(p, name, short, value, usage)
	return p
}

// Uint32Var defines a uint32 flag with the specified name, short form, default value, and usage string.
// The argument p points to a uint32 variable in which to store the value of the flag.
func (f *FlagSet) Uint32Var(p *uint32, name string, short rune, value uint32, usage string) {
	*p = value
	f.Var((*uint32Value)(p), name, short, usage)
}

// Uint32 defines a uint32 flag with the specified name, short form, default value, and usage string.
// The return value is the address of a uint32 variable that stores the value of the flag.
func (f *FlagSet) Uint32(name string, short rune, value uint32, usage string) *uint32 {
	p := new(uint32)
	f.Uint32Var(p, name, short, value, usage)
	return p
}

// Uint64Var defines a uint64 flag with the specified name, short form, default value, and usage string.
// The argument p points to a uint64 variable in which to store the value of the flag.
func (f *FlagSet) Uint64Var(p *uint64, name string, short rune, value uint64, usage string) {
	*p = value
	f.Var((*uint64Value)(p), name, short, usage)
}

// Uint64 defines a uint64 flag with the specified name, short form, default value, and usage string.
// The return value is the address of a uint64 variable that stores the value of the flag.
func (f *FlagSet) Uint64(name string, short rune, value uint64, usage string) *uint64 {
	p := new(uint64)
	f.Uint64Var(p, name, short, value, usage)
	return p
}

// StringArrayVar defines a string array flag with the specified name, short form, default value, and usage string.
// The argument p points to a []string variable in which to store the value of the flag.
// The flag value is expected to be a comma-separated list of strings.
func (f *FlagSet) StringArrayVar(p *[]string, name string, short rune, value []string, usage string) {
	if value != nil {
		*p = value
	} else {
		*p = []string{}
	}
	f.Var(&stringArrayValue{values: p}, name, short, usage)
}

// StringArray defines a string array flag with the specified name, short form, default value, and usage string.
// The return value is the address of a []string variable that stores the value of the flag.
// The flag value is expected to be a comma-separated list of strings.
func (f *FlagSet) StringArray(name string, short rune, value []string, usage string) *[]string {
	p := new([]string)
	f.StringArrayVar(p, name, short, value, usage)
	return p
}

// BoolArrayVar defines a bool array flag with the specified name, short form, and usage string.
// The argument p points to a []bool variable in which to store the values.
// Each occurrence of the flag appends true to the slice, allowing patterns like "-v -v -v"
// to count verbosity levels.
func (f *FlagSet) BoolArrayVar(p *[]bool, name string, short rune, usage string) {
	if *p == nil {
		*p = []bool{}
	}
	f.Var((*boolArrayValue)(p), name, short, usage)
}

// BoolArray defines a bool array flag with the specified name, short form, and usage string.
// The return value is the address of a []bool variable that stores the values.
// Each occurrence of the flag appends true to the slice, allowing patterns like "-v -v -v"
// to count verbosity levels.
func (f *FlagSet) BoolArray(name string, short rune, usage string) *[]bool {
	p := new([]bool)
	f.BoolArrayVar(p, name, short, usage)
	return p
}

// IntArrayVar defines an int array flag with the specified name, short form, and usage string.
// The argument p points to a []int variable in which to store the values.
// Each occurrence of the flag appends to the slice, allowing patterns like "-n 1 -n 2 -n 3".
func (f *FlagSet) IntArrayVar(p *[]int, name string, short rune, usage string) {
	if *p == nil {
		*p = []int{}
	}
	f.Var((*intArrayValue)(p), name, short, usage)
}

// IntArray defines an int array flag with the specified name, short form, and usage string.
// The return value is the address of a []int variable that stores the values.
// Each occurrence of the flag appends to the slice, allowing patterns like "-n 1 -n 2 -n 3".
func (f *FlagSet) IntArray(name string, short rune, usage string) *[]int {
	p := new([]int)
	f.IntArrayVar(p, name, short, usage)
	return p
}

// DurationVar defines a time.Duration flag with the specified name, short form, default value, and usage string.
// The argument p points to a time.Duration variable in which to store the value of the flag.
// The flag accepts values parseable by time.ParseDuration.
func (f *FlagSet) DurationVar(p *time.Duration, name string, short rune, value time.Duration, usage string) {
	*p = value
	f.Var((*durationValue)(p), name, short, usage)
}

// Duration defines a time.Duration flag with the specified name, short form, default value, and usage string.
// The return value is the address of a time.Duration variable that stores the value of the flag.
// The flag accepts values parseable by time.ParseDuration.
func (f *FlagSet) Duration(name string, short rune, value time.Duration, usage string) *time.Duration {
	p := new(time.Duration)
	f.DurationVar(p, name, short, value, usage)
	return p
}

// ChoiceVar defines a string flag that only accepts values from a predefined set of choices.
// The argument p points to a string variable in which to store the value of the flag.
// The value argument is the default value, which must be one of the provided choices (or empty).
// If an invalid choice is provided during parsing, ErrInvalidChoice is returned.
func (f *FlagSet) ChoiceVar(p *string, name string, short rune, value string, choices []string, usage string) {
	*p = value
	f.Var(&choiceValue{value: p, choices: choices}, name, short, usage)
}

// Choice defines a string flag that only accepts values from a predefined set of choices.
// The return value is the address of a string variable that stores the value of the flag.
// The value argument is the default value, which must be one of the provided choices (or empty).
// If an invalid choice is provided during parsing, ErrInvalidChoice is returned.
func (f *FlagSet) Choice(name string, short rune, value string, choices []string, usage string) *string {
	p := new(string)
	f.ChoiceVar(p, name, short, value, choices, usage)
	return p
}

// BoolPosVar defines a bool positional argument at the specified position with a default value and usage string.
// The argument p points to a bool variable in which to store the value of the positional argument.
// Position 0 is the first non-flag argument, position 1 is the second, etc.
func (f *FlagSet) BoolPosVar(p *bool, name string, position int, value bool, usage string) {
	*p = value
	f.posFields[position] = &PositionalField{
		Name:  name,
		Usage: usage,
		Value: reflect.ValueOf(p).Elem(),
		Type:  reflect.TypeOf(*p),
	}
}

// BoolPos defines a bool positional argument at the specified position with a default value and usage string.
// The return value is the address of a bool variable that stores the value of the positional argument.
// Position 0 is the first non-flag argument, position 1 is the second, etc.
func (f *FlagSet) BoolPos(name string, position int, value bool, usage string) *bool {
	p := new(bool)
	f.BoolPosVar(p, name, position, value, usage)
	return p
}

// StringPosVar defines a string positional argument at the specified position with a default value and usage string.
// The argument p points to a string variable in which to store the value of the positional argument.
// Position 0 is the first non-flag argument, position 1 is the second, etc.
func (f *FlagSet) StringPosVar(p *string, name string, position int, value string, usage string) {
	*p = value
	f.posFields[position] = &PositionalField{
		Name:  name,
		Usage: usage,
		Value: reflect.ValueOf(p).Elem(),
		Type:  reflect.TypeOf(*p),
	}
}

// StringPos defines a string positional argument at the specified position with a default value and usage string.
// The return value is the address of a string variable that stores the value of the positional argument.
// Position 0 is the first non-flag argument, position 1 is the second, etc.
func (f *FlagSet) StringPos(name string, position int, value string, usage string) *string {
	p := new(string)
	f.StringPosVar(p, name, position, value, usage)
	return p
}

// IntPosVar defines an int positional argument at the specified position with a default value and usage string.
// The argument p points to an int variable in which to store the value of the positional argument.
// Position 0 is the first non-flag argument, position 1 is the second, etc.
func (f *FlagSet) IntPosVar(p *int, name string, position int, value int, usage string) {
	*p = value
	f.posFields[position] = &PositionalField{
		Name:  name,
		Usage: usage,
		Value: reflect.ValueOf(p).Elem(),
		Type:  reflect.TypeOf(*p),
	}
}

// IntPos defines an int positional argument at the specified position with a default value and usage string.
// The return value is the address of an int variable that stores the value of the positional argument.
// Position 0 is the first non-flag argument, position 1 is the second, etc.
func (f *FlagSet) IntPos(name string, position int, value int, usage string) *int {
	p := new(int)
	f.IntPosVar(p, name, position, value, usage)
	return p
}

// DurationPosVar defines a time.Duration positional argument at the specified position with a default value and usage string.
// The argument p points to a time.Duration variable in which to store the value of the positional argument.
// Position 0 is the first non-flag argument, position 1 is the second, etc.
func (f *FlagSet) DurationPosVar(p *time.Duration, name string, position int, value time.Duration, usage string) {
	*p = value
	f.posFields[position] = &PositionalField{
		Name:  name,
		Usage: usage,
		Value: reflect.ValueOf(p).Elem(),
		Type:  reflect.TypeOf(*p),
	}
}

// DurationPos defines a time.Duration positional argument at the specified position with a default value and usage string.
// The return value is the address of a time.Duration variable that stores the value of the positional argument.
// Position 0 is the first non-flag argument, position 1 is the second, etc.
func (f *FlagSet) DurationPos(name string, position int, value time.Duration, usage string) *time.Duration {
	p := new(time.Duration)
	f.DurationPosVar(p, name, position, value, usage)
	return p
}

// Rest defines a slice to capture all remaining non-flag arguments.
// The argument p points to a []string variable that will be populated with all non-flag arguments.
// This is useful for commands that accept variable-length argument lists.
func (f *FlagSet) Rest(p *[]string, usage string) {
	if p == nil {
		panic("Rest: pointer cannot be nil")
	}
	*p = []string{}
	f.restField = p
}

// Var defines a flag with the specified name, short form, and usage string.
// The type and value of the flag are represented by the first argument, of type Value,
// which typically holds a user-defined implementation of Value.
func (f *FlagSet) Var(value Value, name string, short rune, usage string) {
	flag := &Flag{
		Name:     name,
		Short:    short,
		Usage:    usage,
		Value:    value,
		DefValue: value.String(),
		Group:    f.currentGroup,
	}

	if name != "" {
		if existing, ok := f.flags[name]; ok {
			panic(fmt.Sprintf("flag %q already registered as --%s", name, existing.Name))
		}
		f.flags[name] = flag
	}
	if short != 0 {
		if existing, ok := f.shortMap[short]; ok {
			panic(fmt.Sprintf("short flag '%c' already registered for --%s", short, existing.Name))
		}
		f.shortMap[short] = flag
	}

	// Track group insertion order
	if f.currentGroup != "" {
		found := false
		for _, g := range f.groupOrder {
			if g == f.currentGroup {
				found = true
				break
			}
		}
		if !found {
			f.groupOrder = append(f.groupOrder, f.currentGroup)
		}
	}

	// Add to the list of all flags for iteration
	f.allFlags = append(f.allFlags, flag)
}

// Group sets the current group name for subsequently registered flags.
func (f *FlagSet) Group(name string) { f.currentGroup = name }

// Lookup returns the Flag with the given name, or nil if not found
func (f *FlagSet) Lookup(name string) *Flag {
	return f.flags[name]
}

// HasPositionalArgs returns true if the FlagSet has positional arguments defined
func (f *FlagSet) HasPositionalArgs() bool {
	return len(f.posFields) > 0
}

// HasRestArgs returns true if the FlagSet accepts remaining arguments
func (f *FlagSet) HasRestArgs() bool {
	return f.restField != nil
}

// PositionalCount returns the number of positional arguments defined
func (f *FlagSet) PositionalCount() int {
	if len(f.posFields) == 0 {
		return 0
	}
	maxPos := -1
	for pos := range f.posFields {
		if pos > maxPos {
			maxPos = pos
		}
	}
	return maxPos + 1
}

// GetPositionalFields returns the positional fields in order
func (f *FlagSet) GetPositionalFields() []*PositionalField {
	if len(f.posFields) == 0 {
		return nil
	}

	// Find max position
	maxPos := -1
	for pos := range f.posFields {
		if pos > maxPos {
			maxPos = pos
		}
	}

	// Build ordered slice
	result := make([]*PositionalField, 0, maxPos+1)
	for i := 0; i <= maxPos; i++ {
		if field, ok := f.posFields[i]; ok {
			result = append(result, field)
		}
	}
	return result
}

// Parse parses flag and positional argument definitions from the argument list,
// which should not include the command name. Must be called after all flags are defined
// and before flags are accessed by the program.
// The return value will be ErrHelp if -help or -h were set but not defined.
func (f *FlagSet) Parse(arguments []string) error {
	f.parsed = true
	f.args = nil
	f.unknownFlags = nil

	// Check for help flags (-h or --help) before parsing, stop at --
	// If allowUnknownFlags is true, only show help if there are no other arguments
	// Skip automatic help if disableAutoHelp is set (e.g., when used through Dispatcher)
	if !f.disableAutoHelp {
		hasHelpFlag := false
		hasOtherArgs := false

		for _, arg := range arguments {
			if arg == "--" {
				break
			}
			if arg == "-h" || arg == "--help" {
				// Check if these flags are already defined
				_, hDefined := f.shortMap['h']
				_, helpDefined := f.flags["help"]

				// If help flags are not defined, mark that we found a help flag
				if (arg == "-h" && !hDefined) || (arg == "--help" && !helpDefined) {
					hasHelpFlag = true
				}
			} else if !strings.HasPrefix(arg, "-") {
				// Found a non-flag argument
				hasOtherArgs = true
			}
		}

		// Show help if we found a help flag and either:
		// 1. allowUnknownFlags is false (always show help), OR
		// 2. allowUnknownFlags is true but there are no other arguments
		if hasHelpFlag && (!f.allowUnknownFlags || !hasOtherArgs) {
			f.ShowHelp()
			return ErrHelp
		}
	}

	for i := 0; i < len(arguments); i++ {
		arg := arguments[i]

		if arg == "--" {
			f.args = append(f.args, arguments[i+1:]...)
			break
		}

		if strings.HasPrefix(arg, "--") {
			consumed, err := f.parseLongFlag(arg[2:], arguments, &i)
			if err != nil {
				return err
			}
			if consumed {
				continue
			}
			continue
		}

		if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			err := f.parseShortFlags(arg[1:], arguments, &i)
			if err != nil {
				return err
			}
			continue
		}

		f.args = append(f.args, arg)
	}

	// Process positional arguments
	for pos, field := range f.posFields {
		if pos < len(f.args) {
			if err := setFieldValue(field.Value, f.args[pos]); err != nil {
				return fmt.Errorf("invalid value for position %d: %v", pos, err)
			}
			field.HasValue = true
		}
	}

	// Check for unexpected extra arguments when no rest field is defined
	// Skip validation if allowUnknownFlags is enabled (pass-through mode)
	if f.restField == nil && !f.allowUnknownFlags && len(f.args) > f.PositionalCount() {
		extra := f.args[f.PositionalCount():]
		return fmt.Errorf("unexpected arguments: %v", extra)
	}

	// If we have a rest field, populate it with remaining args after positional ones
	if f.restField != nil {
		pc := f.PositionalCount()
		if pc > 0 && pc <= len(f.args) {
			*f.restField = f.args[pc:]
		} else if pc == 0 {
			*f.restField = f.args
		}
	}

	// If we have an unknown field, populate it with unknown flags
	if f.unknownField != nil {
		*f.unknownField = f.unknownFlags
	}

	if err := f.validateRequired(); err != nil {
		return err
	}

	return nil
}

// validateRequired checks that all flags and positionals marked required:"true"
// have been provided a value (via CLI arg or env var).
func (f *FlagSet) validateRequired() error {
	var missing []string

	for _, name := range f.requiredFlags {
		if flag, ok := f.flags[name]; ok {
			if !flag.HasValue {
				missing = append(missing, "--"+name)
			}
		}
	}

	for _, pos := range f.requiredPos {
		if field, ok := f.posFields[pos]; ok {
			if !field.HasValue {
				missing = append(missing, field.Name)
			}
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("%w: %s", ErrRequired, strings.Join(missing, ", "))
	}

	return nil
}

func (f *FlagSet) parseLongFlag(name string, args []string, index *int) (bool, error) {
	var value string
	hasValue := false

	if strings.Contains(name, "=") {
		parts := strings.SplitN(name, "=", 2)
		name = parts[0]
		value = parts[1]
		hasValue = true
	}

	flag, ok := f.flags[name]
	if !ok {
		if f.allowUnknownFlags {
			// Unknown flag encountered - accumulate this and all remaining args
			f.unknownFlags = append(f.unknownFlags, args[*index:]...)
			*index = len(args) - 1 // Skip to end
			return true, nil
		}
		return false, fmt.Errorf("%w: --%s", ErrUnknownFlag, name)
	}

	if flag.Value.IsBool() {
		if !hasValue {
			value = "true"
		}
	} else {
		if !hasValue {
			if *index+1 >= len(args) {
				return false, fmt.Errorf("%w: --%s", ErrMissingValue, name)
			}
			value = args[*index+1]
			*index++
		}
	}

	if err := flag.Value.Set(value); err != nil {
		return false, fmt.Errorf("%w: --%s: %v", ErrInvalidValue, name, err)
	}

	flag.HasValue = true


	return true, nil
}

func (f *FlagSet) parseShortFlags(shortFlags string, args []string, index *int) error {
	runes := []rune(shortFlags)

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		flag, ok := f.shortMap[r]
		if !ok {
			if f.allowUnknownFlags {
				// Unknown flag encountered - accumulate this and all remaining args
				f.unknownFlags = append(f.unknownFlags, args[*index:]...)
				*index = len(args) - 1 // Skip to end
				return nil
			}
			return fmt.Errorf("%w: -%c", ErrUnknownFlag, r)
		}

		if flag.Value.IsBool() {
			if err := flag.Value.Set("true"); err != nil {
				return fmt.Errorf("%w: -%c: %v", ErrInvalidValue, r, err)
			}
			flag.HasValue = true

		} else {
			// Check if there are more characters after this flag
			if i < len(runes)-1 {
				// Check if the next character is also a flag that needs an argument
				nextRune := runes[i+1]
				if nextFlag, exists := f.shortMap[nextRune]; exists && !nextFlag.Value.IsBool() {
					// Both flags need arguments, this is an error
					return fmt.Errorf("%w: -%c", ErrMissingValue, r)
				}
				// Otherwise use the rest as the value
				value := string(runes[i+1:])
				if err := flag.Value.Set(value); err != nil {
					return fmt.Errorf("%w: -%c: %v", ErrInvalidValue, r, err)
				}
				flag.HasValue = true

				break
			} else if *index+1 < len(args) {
				value := args[*index+1]
				*index++
				if err := flag.Value.Set(value); err != nil {
					return fmt.Errorf("%w: -%c: %v", ErrInvalidValue, r, err)
				}
				flag.HasValue = true

			} else {
				return fmt.Errorf("%w: -%c", ErrMissingValue, r)
			}
			break
		}
	}

	return nil
}

// Args returns the non-flag arguments.
func (f *FlagSet) Args() []string {
	return f.args
}

// Parsed reports whether f.Parse has been called.
func (f *FlagSet) Parsed() bool {
	return f.parsed
}

// AllowUnknownFlags enables or disables accumulation of unknown flags.
// When enabled, unknown flags will not cause an error but will be accumulated
// and can be retrieved via UnknownFlags().
func (f *FlagSet) AllowUnknownFlags(allow bool) {
	f.allowUnknownFlags = allow
}

// UnknownFlags returns the list of unknown flags encountered during parsing.
// This is only populated when AllowUnknownFlags(true) has been called.
// Each entry includes the flag as it appeared (e.g., "--unknown" or "-u").
func (f *FlagSet) UnknownFlags() []string {
	return f.unknownFlags
}
