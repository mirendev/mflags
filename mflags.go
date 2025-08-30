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

type FlagSet struct {
	name      string
	flags     map[string]*Flag
	shortMap  map[rune]*Flag
	args      []string
	parsed    bool
	restField *[]string             // Pointer to field marked with "rest" tag
	posFields map[int]reflect.Value // Map of position to field value
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

func NewFlagSet(name string) *FlagSet {
	return &FlagSet{
		name:      name,
		flags:     make(map[string]*Flag),
		shortMap:  make(map[rune]*Flag),
		posFields: make(map[int]reflect.Value),
	}
}

func (f *FlagSet) BoolVar(p *bool, name string, short rune, value bool, usage string) {
	f.Var((*boolValue)(p), name, short, usage)
	*p = value
}

func (f *FlagSet) Bool(name string, short rune, value bool, usage string) *bool {
	p := new(bool)
	f.BoolVar(p, name, short, value, usage)
	return p
}

func (f *FlagSet) StringVar(p *string, name string, short rune, value string, usage string) {
	f.Var((*stringValue)(p), name, short, usage)
	*p = value
}

func (f *FlagSet) String(name string, short rune, value string, usage string) *string {
	p := new(string)
	f.StringVar(p, name, short, value, usage)
	return p
}

func (f *FlagSet) IntVar(p *int, name string, short rune, value int, usage string) {
	f.Var((*intValue)(p), name, short, usage)
	*p = value
}

func (f *FlagSet) Int(name string, short rune, value int, usage string) *int {
	p := new(int)
	f.IntVar(p, name, short, value, usage)
	return p
}

func (f *FlagSet) StringArrayVar(p *[]string, name string, short rune, value []string, usage string) {
	f.Var((*stringArrayValue)(p), name, short, usage)
	if value != nil {
		*p = value
	} else {
		*p = []string{}
	}
}

func (f *FlagSet) StringArray(name string, short rune, value []string, usage string) *[]string {
	p := new([]string)
	f.StringArrayVar(p, name, short, value, usage)
	return p
}

func (f *FlagSet) DurationVar(p *time.Duration, name string, short rune, value time.Duration, usage string) {
	f.Var((*durationValue)(p), name, short, usage)
	*p = value
}

func (f *FlagSet) Duration(name string, short rune, value time.Duration, usage string) *time.Duration {
	p := new(time.Duration)
	f.DurationVar(p, name, short, value, usage)
	return p
}

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

func (f *FlagSet) Parse(arguments []string) error {
	f.parsed = true
	f.args = nil

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
	for pos, fieldValue := range f.posFields {
		if pos < len(f.args) {
			if err := setFieldValue(fieldValue, f.args[pos]); err != nil {
				return fmt.Errorf("invalid value for position %d: %v", pos, err)
			}
		}
	}

	// If we have a rest field, populate it with remaining args
	if f.restField != nil {
		*f.restField = f.args
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

func (f *FlagSet) Args() []string {
	return f.args
}

func (f *FlagSet) Parsed() bool {
	return f.parsed
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

// FromStruct creates flag definitions from a struct's fields using struct tags
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

		// Check for "position" tag - capture positional argument
		if posStr := field.Tag.Get("position"); posStr != "" {
			pos, err := strconv.Atoi(posStr)
			if err == nil && pos >= 0 {
				f.posFields[pos] = fieldValue
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

// ParseStruct parses command line arguments and updates the struct fields
func ParseStruct(v any, arguments []string) error {
	fs := NewFlagSet("")
	if err := fs.FromStruct(v); err != nil {
		return err
	}
	return fs.Parse(arguments)
}
