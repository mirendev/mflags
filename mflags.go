package mflags

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	ErrUnknownFlag  = errors.New("unknown flag")
	ErrMissingValue = errors.New("flag needs an argument")
	ErrInvalidValue = errors.New("invalid flag value")
	ErrHelp         = errors.New("help requested")
)

// PositionalField represents a positional argument field
type PositionalField struct {
	Name  string        // Field name (e.g., "Command", "Target")
	Value reflect.Value // The reflect.Value of the field
	Type  reflect.Type  // The type of the field
}

type FlagSet struct {
	name              string
	flags             map[string]*Flag
	shortMap          map[rune]*Flag
	args              []string
	parsed            bool
	restField         *[]string                // Pointer to field marked with "rest" tag
	posFields         map[int]*PositionalField // Map of position to positional field info
	allowUnknownFlags bool                     // If true, accumulate unknown flags instead of erroring
	unknownFlags      []string                 // Accumulated unknown flags when allowUnknownFlags is true
	unknownField      *[]string                // Pointer to field marked with "unknown" tag
}

type Flag struct {
	Name     string
	Short    rune
	Usage    string
	Value    Value
	DefValue string
}

type Value interface {
	String() string
	Set(string) error
	IsBool() bool
	Type() string
}

type boolValue bool

func (b *boolValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	*b = boolValue(v)
	return nil
}

func (b *boolValue) String() string {
	return strconv.FormatBool(bool(*b))
}

func (b *boolValue) IsBool() bool {
	return true
}

func (b *boolValue) Type() string {
	return "bool"
}

type stringValue string

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) String() string {
	return string(*s)
}

func (s *stringValue) IsBool() bool {
	return false
}

func (s *stringValue) Type() string {
	return "string"
}

type intValue int

func (i *intValue) Set(s string) error {
	v, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*i = intValue(v)
	return nil
}

func (i *intValue) String() string {
	return strconv.Itoa(int(*i))
}

func (i *intValue) IsBool() bool {
	return false
}

func (i *intValue) Type() string {
	return "int"
}

type stringArrayValue []string

func (s *stringArrayValue) Set(val string) error {
	*s = strings.Split(val, ",")
	return nil
}

func (s *stringArrayValue) String() string {
	return strings.Join(*s, ",")
}

func (s *stringArrayValue) IsBool() bool {
	return false
}

func (s *stringArrayValue) Type() string {
	return "value,..."
}

type durationValue time.Duration

func (d *durationValue) Set(s string) error {
	v, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = durationValue(v)
	return nil
}

func (d *durationValue) String() string {
	return time.Duration(*d).String()
}

func (d *durationValue) IsBool() bool {
	return false
}

func (d *durationValue) Type() string {
	return "duration"
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
	f.Var((*boolValue)(p), name, short, usage)
	*p = value
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
	f.Var((*stringValue)(p), name, short, usage)
	*p = value
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
	f.Var((*intValue)(p), name, short, usage)
	*p = value
}

// Int defines an int flag with the specified name, short form, default value, and usage string.
// The return value is the address of an int variable that stores the value of the flag.
func (f *FlagSet) Int(name string, short rune, value int, usage string) *int {
	p := new(int)
	f.IntVar(p, name, short, value, usage)
	return p
}

// StringArrayVar defines a string array flag with the specified name, short form, default value, and usage string.
// The argument p points to a []string variable in which to store the value of the flag.
// The flag value is expected to be a comma-separated list of strings.
func (f *FlagSet) StringArrayVar(p *[]string, name string, short rune, value []string, usage string) {
	f.Var((*stringArrayValue)(p), name, short, usage)
	if value != nil {
		*p = value
	} else {
		*p = []string{}
	}
}

// StringArray defines a string array flag with the specified name, short form, default value, and usage string.
// The return value is the address of a []string variable that stores the value of the flag.
// The flag value is expected to be a comma-separated list of strings.
func (f *FlagSet) StringArray(name string, short rune, value []string, usage string) *[]string {
	p := new([]string)
	f.StringArrayVar(p, name, short, value, usage)
	return p
}

// DurationVar defines a time.Duration flag with the specified name, short form, default value, and usage string.
// The argument p points to a time.Duration variable in which to store the value of the flag.
// The flag accepts values parseable by time.ParseDuration.
func (f *FlagSet) DurationVar(p *time.Duration, name string, short rune, value time.Duration, usage string) {
	f.Var((*durationValue)(p), name, short, usage)
	*p = value
}

// Duration defines a time.Duration flag with the specified name, short form, default value, and usage string.
// The return value is the address of a time.Duration variable that stores the value of the flag.
// The flag accepts values parseable by time.ParseDuration.
func (f *FlagSet) Duration(name string, short rune, value time.Duration, usage string) *time.Duration {
	p := new(time.Duration)
	f.DurationVar(p, name, short, value, usage)
	return p
}

// BoolPosVar defines a bool positional argument at the specified position with a default value and usage string.
// The argument p points to a bool variable in which to store the value of the positional argument.
// Position 0 is the first non-flag argument, position 1 is the second, etc.
func (f *FlagSet) BoolPosVar(p *bool, name string, position int, value bool, usage string) {
	*p = value
	f.posFields[position] = &PositionalField{
		Name:  name,
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
	}

	if name != "" {
		f.flags[name] = flag
	}
	if short != 0 {
		f.shortMap[short] = flag
	}
}

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
		}
	}

	// If we have a rest field, populate it with remaining args
	if f.restField != nil {
		*f.restField = f.args
	}

	// If we have an unknown field, populate it with unknown flags
	if f.unknownField != nil {
		*f.unknownField = f.unknownFlags
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
				break
			} else if *index+1 < len(args) {
				value := args[*index+1]
				*index++
				if err := flag.Value.Set(value); err != nil {
					return fmt.Errorf("%w: -%c: %v", ErrInvalidValue, r, err)
				}
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

// setFieldValue sets a string value to a reflect.Value based on its type
func setFieldValue(fieldValue reflect.Value, value string) error {
	switch fieldValue.Kind() {
	case reflect.String:
		fieldValue.SetString(value)
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		fieldValue.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if fieldValue.Type() == reflect.TypeOf(time.Duration(0)) {
			d, err := time.ParseDuration(value)
			if err != nil {
				return err
			}
			fieldValue.SetInt(int64(d))
		} else {
			i, err := strconv.ParseInt(value, 10, fieldValue.Type().Bits())
			if err != nil {
				return err
			}
			fieldValue.SetInt(i)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(value, 10, fieldValue.Type().Bits())
		if err != nil {
			return err
		}
		fieldValue.SetUint(u)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(value, fieldValue.Type().Bits())
		if err != nil {
			return err
		}
		fieldValue.SetFloat(f)
	default:
		return fmt.Errorf("unsupported type: %v", fieldValue.Type())
	}
	return nil
}

// FromStruct creates flag definitions from a struct's fields using struct tags.
// The argument must be a pointer to a struct. Struct tags control how fields are parsed:
//   - `long:"name"` - long flag name (defaults to lowercase field name)
//   - `short:"x"` - short flag name (single character)
//   - `default:"value"` - default value for the flag
//   - `usage:"description"` - usage description
//   - `position:"0"` - positional argument at index 0
//   - `rest:"true"` - capture all remaining arguments in a []string field
//   - `unknown:"true"` - capture unknown flags in a []string field (automatically enables AllowUnknownFlags)
//
// Supports bool, string, int, []string, and time.Duration field types.
// Anonymous embedded structs are recursively processed.
func (f *FlagSet) FromStruct(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("FromStruct requires a non-nil pointer to a struct")
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("FromStruct requires a pointer to a struct")
	}

	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if !field.IsExported() {
			continue
		}

		fieldValue := rv.Field(i)
		if !fieldValue.CanAddr() {
			continue
		}

		// Check for anonymous/embedded struct fields and descend into them
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			if err := f.FromStruct(fieldValue.Addr().Interface()); err != nil {
				return err
			}
			continue
		}

		// Check for "position" tag - capture positional argument
		if posStr := field.Tag.Get("position"); posStr != "" {
			pos, err := strconv.Atoi(posStr)
			if err == nil && pos >= 0 {
				f.posFields[pos] = &PositionalField{
					Name:  field.Name,
					Value: fieldValue,
					Type:  field.Type,
				}
			}
			continue // Don't process position field as a flag
		}

		// Check for "rest" tag - special handling for remaining arguments
		if field.Tag.Get("rest") != "" {
			if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.String {
				f.restField = fieldValue.Addr().Interface().(*[]string)
			}
			continue // Don't process rest field as a flag
		}

		// Check for "unknown" tag - special handling for unknown flags
		if field.Tag.Get("unknown") != "" {
			if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.String {
				f.unknownField = fieldValue.Addr().Interface().(*[]string)
				f.allowUnknownFlags = true // Automatically enable unknown flag handling
			}
			continue // Don't process unknown field as a flag
		}

		// Parse struct tags
		longName := field.Tag.Get("long")
		if longName == "" {
			longName = strings.ToLower(field.Name)
		}

		shortName := field.Tag.Get("short")
		var short rune
		if shortName != "" && len(shortName) == 1 {
			short = rune(shortName[0])
		}

		if longName == "" && short == 0 {
			continue // No flag name provided
		}

		defaultValue := field.Tag.Get("default")
		usage := field.Tag.Get("usage")
		if usage == "" {
			usage = fmt.Sprintf("%s value", field.Name)
		}

		// Register the flag based on field type
		switch field.Type.Kind() {
		case reflect.Bool:
			var defVal bool
			if defaultValue != "" {
				defVal, _ = strconv.ParseBool(defaultValue)
			}
			f.BoolVar(fieldValue.Addr().Interface().(*bool), longName, short, defVal, usage)

		case reflect.String:
			f.StringVar(fieldValue.Addr().Interface().(*string), longName, short, defaultValue, usage)

		case reflect.Int:
			var defVal int
			if defaultValue != "" {
				defVal, _ = strconv.Atoi(defaultValue)
			}
			f.IntVar(fieldValue.Addr().Interface().(*int), longName, short, defVal, usage)

		case reflect.Slice:
			if field.Type.Elem().Kind() == reflect.String {
				var defVal []string
				if defaultValue != "" {
					defVal = strings.Split(defaultValue, ",")
				}
				f.StringArrayVar(fieldValue.Addr().Interface().(*[]string), longName, short, defVal, usage)
			}

		case reflect.Int64:
			// Check if it's a time.Duration
			if field.Type == reflect.TypeOf(time.Duration(0)) {
				var defVal time.Duration
				if defaultValue != "" {
					defVal, _ = time.ParseDuration(defaultValue)
				}
				f.DurationVar(fieldValue.Addr().Interface().(*time.Duration), longName, short, defVal, usage)
			}
		}
	}

	return nil
}

// ParseStruct parses command line arguments and updates the struct fields.
// This is a convenience function that creates a FlagSet, calls FromStruct, and parses the arguments.
// See FromStruct for documentation on supported struct tags and field types.
func ParseStruct(v any, arguments []string) error {
	fs := NewFlagSet("")
	if err := fs.FromStruct(v); err != nil {
		return err
	}
	return fs.Parse(arguments)
}
