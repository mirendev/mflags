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
	ErrUnknownFlag   = errors.New("unknown flag")
	ErrMissingValue  = errors.New("flag needs an argument")
	ErrInvalidValue  = errors.New("invalid flag value")
	ErrInvalidChoice = errors.New("invalid choice")
	ErrHelp          = errors.New("help requested")
	ErrShowHelp      = errors.New("show help") // Return from Command.Run to trigger help display
)

// PositionalField represents a positional argument field
type PositionalField struct {
	Name  string        // Field name (e.g., "Command", "Target")
	Usage string        // Usage description for help output
	Value reflect.Value // The reflect.Value of the field
	Type  reflect.Type  // The type of the field
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

type boolArrayValue []bool

func (b *boolArrayValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	*b = append(*b, v)
	return nil
}

func (b *boolArrayValue) String() string {
	if len(*b) == 0 {
		return ""
	}
	strs := make([]string, len(*b))
	for i, v := range *b {
		strs[i] = strconv.FormatBool(v)
	}
	return strings.Join(strs, ",")
}

func (b *boolArrayValue) IsBool() bool {
	return true
}

func (b *boolArrayValue) Type() string {
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

type int64Value int64

func (i *int64Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	*i = int64Value(v)
	return nil
}

func (i *int64Value) String() string {
	return strconv.FormatInt(int64(*i), 10)
}

func (i *int64Value) IsBool() bool {
	return false
}

func (i *int64Value) Type() string {
	return "int"
}

type int8Value int8

func (i *int8Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 10, 8)
	if err != nil {
		return err
	}
	*i = int8Value(v)
	return nil
}

func (i *int8Value) String() string {
	return strconv.FormatInt(int64(*i), 10)
}

func (i *int8Value) IsBool() bool {
	return false
}

func (i *int8Value) Type() string {
	return "int"
}

type int16Value int16

func (i *int16Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 10, 16)
	if err != nil {
		return err
	}
	*i = int16Value(v)
	return nil
}

func (i *int16Value) String() string {
	return strconv.FormatInt(int64(*i), 10)
}

func (i *int16Value) IsBool() bool {
	return false
}

func (i *int16Value) Type() string {
	return "int"
}

type int32Value int32

func (i *int32Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return err
	}
	*i = int32Value(v)
	return nil
}

func (i *int32Value) String() string {
	return strconv.FormatInt(int64(*i), 10)
}

func (i *int32Value) IsBool() bool {
	return false
}

func (i *int32Value) Type() string {
	return "int"
}

type uintValue uint

func (i *uintValue) Set(s string) error {
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return err
	}
	*i = uintValue(v)
	return nil
}

func (i *uintValue) String() string {
	return strconv.FormatUint(uint64(*i), 10)
}

func (i *uintValue) IsBool() bool {
	return false
}

func (i *uintValue) Type() string {
	return "uint"
}

type uint8Value uint8

func (i *uint8Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 10, 8)
	if err != nil {
		return err
	}
	*i = uint8Value(v)
	return nil
}

func (i *uint8Value) String() string {
	return strconv.FormatUint(uint64(*i), 10)
}

func (i *uint8Value) IsBool() bool {
	return false
}

func (i *uint8Value) Type() string {
	return "uint"
}

type uint16Value uint16

func (i *uint16Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 10, 16)
	if err != nil {
		return err
	}
	*i = uint16Value(v)
	return nil
}

func (i *uint16Value) String() string {
	return strconv.FormatUint(uint64(*i), 10)
}

func (i *uint16Value) IsBool() bool {
	return false
}

func (i *uint16Value) Type() string {
	return "uint"
}

type uint32Value uint32

func (i *uint32Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return err
	}
	*i = uint32Value(v)
	return nil
}

func (i *uint32Value) String() string {
	return strconv.FormatUint(uint64(*i), 10)
}

func (i *uint32Value) IsBool() bool {
	return false
}

func (i *uint32Value) Type() string {
	return "uint"
}

type uint64Value uint64

func (i *uint64Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return err
	}
	*i = uint64Value(v)
	return nil
}

func (i *uint64Value) String() string {
	return strconv.FormatUint(uint64(*i), 10)
}

func (i *uint64Value) IsBool() bool {
	return false
}

func (i *uint64Value) Type() string {
	return "uint"
}

type intArrayValue []int

func (i *intArrayValue) Set(s string) error {
	v, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*i = append(*i, v)
	return nil
}

func (i *intArrayValue) String() string {
	if len(*i) == 0 {
		return ""
	}
	strs := make([]string, len(*i))
	for idx, v := range *i {
		strs[idx] = strconv.Itoa(v)
	}
	return strings.Join(strs, ",")
}

func (i *intArrayValue) IsBool() bool {
	return false
}

func (i *intArrayValue) Type() string {
	return "int"
}

type stringArrayValue struct {
	values  *[]string
	hasBeenSet bool
}

func (s *stringArrayValue) Set(val string) error {
	// On first Set call, clear any default values
	if !s.hasBeenSet {
		*s.values = nil
		s.hasBeenSet = true
	}
	*s.values = append(*s.values, strings.Split(val, ",")...)
	return nil
}

func (s *stringArrayValue) String() string {
	if s.values == nil {
		return ""
	}
	return strings.Join(*s.values, ",")
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

// Pointer value types - these allocate the pointed-to value on first Set,
// allowing code to distinguish between "not set" (nil) and "set to zero value"

type boolPtrValue struct {
	p **bool
}

func (b *boolPtrValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	*b.p = new(bool)
	**b.p = v
	return nil
}

func (b *boolPtrValue) String() string {
	if *b.p == nil {
		return ""
	}
	return strconv.FormatBool(**b.p)
}

func (b *boolPtrValue) IsBool() bool {
	return true
}

func (b *boolPtrValue) Type() string {
	return "bool"
}

type stringPtrValue struct {
	p **string
}

func (s *stringPtrValue) Set(val string) error {
	*s.p = new(string)
	**s.p = val
	return nil
}

func (s *stringPtrValue) String() string {
	if *s.p == nil {
		return ""
	}
	return **s.p
}

func (s *stringPtrValue) IsBool() bool {
	return false
}

func (s *stringPtrValue) Type() string {
	return "string"
}

type intPtrValue struct {
	p **int
}

func (i *intPtrValue) Set(s string) error {
	v, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*i.p = new(int)
	**i.p = v
	return nil
}

func (i *intPtrValue) String() string {
	if *i.p == nil {
		return ""
	}
	return strconv.Itoa(**i.p)
}

func (i *intPtrValue) IsBool() bool {
	return false
}

func (i *intPtrValue) Type() string {
	return "int"
}

type int64PtrValue struct {
	p **int64
}

func (i *int64PtrValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	*i.p = new(int64)
	**i.p = v
	return nil
}

func (i *int64PtrValue) String() string {
	if *i.p == nil {
		return ""
	}
	return strconv.FormatInt(**i.p, 10)
}

func (i *int64PtrValue) IsBool() bool {
	return false
}

func (i *int64PtrValue) Type() string {
	return "int"
}

type int8PtrValue struct {
	p **int8
}

func (i *int8PtrValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 10, 8)
	if err != nil {
		return err
	}
	*i.p = new(int8)
	**i.p = int8(v)
	return nil
}

func (i *int8PtrValue) String() string {
	if *i.p == nil {
		return ""
	}
	return strconv.FormatInt(int64(**i.p), 10)
}

func (i *int8PtrValue) IsBool() bool {
	return false
}

func (i *int8PtrValue) Type() string {
	return "int"
}

type int16PtrValue struct {
	p **int16
}

func (i *int16PtrValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 10, 16)
	if err != nil {
		return err
	}
	*i.p = new(int16)
	**i.p = int16(v)
	return nil
}

func (i *int16PtrValue) String() string {
	if *i.p == nil {
		return ""
	}
	return strconv.FormatInt(int64(**i.p), 10)
}

func (i *int16PtrValue) IsBool() bool {
	return false
}

func (i *int16PtrValue) Type() string {
	return "int"
}

type int32PtrValue struct {
	p **int32
}

func (i *int32PtrValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return err
	}
	*i.p = new(int32)
	**i.p = int32(v)
	return nil
}

func (i *int32PtrValue) String() string {
	if *i.p == nil {
		return ""
	}
	return strconv.FormatInt(int64(**i.p), 10)
}

func (i *int32PtrValue) IsBool() bool {
	return false
}

func (i *int32PtrValue) Type() string {
	return "int"
}

type uintPtrValue struct {
	p **uint
}

func (i *uintPtrValue) Set(s string) error {
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return err
	}
	*i.p = new(uint)
	**i.p = uint(v)
	return nil
}

func (i *uintPtrValue) String() string {
	if *i.p == nil {
		return ""
	}
	return strconv.FormatUint(uint64(**i.p), 10)
}

func (i *uintPtrValue) IsBool() bool {
	return false
}

func (i *uintPtrValue) Type() string {
	return "uint"
}

type uint8PtrValue struct {
	p **uint8
}

func (i *uint8PtrValue) Set(s string) error {
	v, err := strconv.ParseUint(s, 10, 8)
	if err != nil {
		return err
	}
	*i.p = new(uint8)
	**i.p = uint8(v)
	return nil
}

func (i *uint8PtrValue) String() string {
	if *i.p == nil {
		return ""
	}
	return strconv.FormatUint(uint64(**i.p), 10)
}

func (i *uint8PtrValue) IsBool() bool {
	return false
}

func (i *uint8PtrValue) Type() string {
	return "uint"
}

type uint16PtrValue struct {
	p **uint16
}

func (i *uint16PtrValue) Set(s string) error {
	v, err := strconv.ParseUint(s, 10, 16)
	if err != nil {
		return err
	}
	*i.p = new(uint16)
	**i.p = uint16(v)
	return nil
}

func (i *uint16PtrValue) String() string {
	if *i.p == nil {
		return ""
	}
	return strconv.FormatUint(uint64(**i.p), 10)
}

func (i *uint16PtrValue) IsBool() bool {
	return false
}

func (i *uint16PtrValue) Type() string {
	return "uint"
}

type uint32PtrValue struct {
	p **uint32
}

func (i *uint32PtrValue) Set(s string) error {
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return err
	}
	*i.p = new(uint32)
	**i.p = uint32(v)
	return nil
}

func (i *uint32PtrValue) String() string {
	if *i.p == nil {
		return ""
	}
	return strconv.FormatUint(uint64(**i.p), 10)
}

func (i *uint32PtrValue) IsBool() bool {
	return false
}

func (i *uint32PtrValue) Type() string {
	return "uint"
}

type uint64PtrValue struct {
	p **uint64
}

func (i *uint64PtrValue) Set(s string) error {
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return err
	}
	*i.p = new(uint64)
	**i.p = v
	return nil
}

func (i *uint64PtrValue) String() string {
	if *i.p == nil {
		return ""
	}
	return strconv.FormatUint(**i.p, 10)
}

func (i *uint64PtrValue) IsBool() bool {
	return false
}

func (i *uint64PtrValue) Type() string {
	return "uint"
}

type durationPtrValue struct {
	p **time.Duration
}

func (d *durationPtrValue) Set(s string) error {
	v, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d.p = new(time.Duration)
	**d.p = v
	return nil
}

func (d *durationPtrValue) String() string {
	if *d.p == nil {
		return ""
	}
	return (**d.p).String()
}

func (d *durationPtrValue) IsBool() bool {
	return false
}

func (d *durationPtrValue) Type() string {
	return "duration"
}

// choiceValue represents a string flag that only accepts specific values.
// It validates inputs against a predefined set of choices.
type choiceValue struct {
	value   *string
	choices []string
}

func (c *choiceValue) Set(s string) error {
	for _, choice := range c.choices {
		if s == choice {
			*c.value = s
			return nil
		}
	}
	return fmt.Errorf("%w: %q (valid: %s)", ErrInvalidChoice, s, strings.Join(c.choices, ", "))
}

func (c *choiceValue) String() string {
	if c.value == nil {
		return ""
	}
	return *c.value
}

func (c *choiceValue) IsBool() bool {
	return false
}

func (c *choiceValue) Type() string {
	return strings.Join(c.choices, "|")
}

// Choices returns the valid choices for this value
func (c *choiceValue) Choices() []string {
	return c.choices
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
	}

	if name != "" {
		f.flags[name] = flag
	}
	if short != 0 {
		f.shortMap[short] = flag
	}

	// Add to the list of all flags for iteration
	f.allFlags = append(f.allFlags, flag)
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

// getTagValues extracts all values for a given key from a struct tag.
// This is needed because Go's tag.Get() only returns the first value,
// but we need to support multiple values (e.g., multiple choice tags).
func getTagValues(tag reflect.StructTag, key string) []string {
	var values []string
	tagStr := string(tag)
	searchKey := key + `:`

	for {
		idx := strings.Index(tagStr, searchKey)
		if idx < 0 {
			break
		}

		// Move past the key and colon
		tagStr = tagStr[idx+len(searchKey):]

		// Find the quoted value
		if len(tagStr) == 0 || tagStr[0] != '"' {
			break
		}

		// Find the closing quote
		endIdx := 1
		for endIdx < len(tagStr) && tagStr[endIdx] != '"' {
			if tagStr[endIdx] == '\\' && endIdx+1 < len(tagStr) {
				endIdx += 2 // Skip escaped character
			} else {
				endIdx++
			}
		}

		if endIdx < len(tagStr) {
			value := tagStr[1:endIdx]
			values = append(values, value)
			tagStr = tagStr[endIdx+1:]
		} else {
			break
		}
	}

	return values
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
//   - `description:"description"` - alternate usage description
//   - `choice:"value"` - constrain string field to specific values (can be repeated for multiple choices)
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
				// Get usage from either "usage" or "description" tag
				posUsage := field.Tag.Get("usage")
				if posUsage == "" {
					posUsage = field.Tag.Get("description")
				}
				f.posFields[pos] = &PositionalField{
					Name:  field.Name,
					Usage: posUsage,
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
			usage = field.Tag.Get("description")
			if usage == "" {
				usage = fmt.Sprintf("%s value", field.Name)
			}
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
			// Check for choice tags - if present, use ChoiceVar
			choices := getTagValues(field.Tag, "choice")
			if len(choices) > 0 {
				f.ChoiceVar(fieldValue.Addr().Interface().(*string), longName, short, defaultValue, choices, usage)
			} else {
				f.StringVar(fieldValue.Addr().Interface().(*string), longName, short, defaultValue, usage)
			}

		case reflect.Int:
			var defVal int
			if defaultValue != "" {
				defVal, _ = strconv.Atoi(defaultValue)
			}
			f.IntVar(fieldValue.Addr().Interface().(*int), longName, short, defVal, usage)

		case reflect.Slice:
			switch field.Type.Elem().Kind() {
			case reflect.String:
				var defVal []string
				if defaultValue != "" {
					defVal = strings.Split(defaultValue, ",")
				}
				f.StringArrayVar(fieldValue.Addr().Interface().(*[]string), longName, short, defVal, usage)
			case reflect.Bool:
				f.BoolArrayVar(fieldValue.Addr().Interface().(*[]bool), longName, short, usage)
			case reflect.Int:
				f.IntArrayVar(fieldValue.Addr().Interface().(*[]int), longName, short, usage)
			}

		case reflect.Int64:
			// Check if it's a time.Duration
			if field.Type == reflect.TypeOf(time.Duration(0)) {
				var defVal time.Duration
				if defaultValue != "" {
					defVal, _ = time.ParseDuration(defaultValue)
				}
				f.DurationVar(fieldValue.Addr().Interface().(*time.Duration), longName, short, defVal, usage)
			} else {
				var defVal int64
				if defaultValue != "" {
					defVal, _ = strconv.ParseInt(defaultValue, 10, 64)
				}
				f.Int64Var(fieldValue.Addr().Interface().(*int64), longName, short, defVal, usage)
			}

		case reflect.Int8:
			var defVal int8
			if defaultValue != "" {
				v, _ := strconv.ParseInt(defaultValue, 10, 8)
				defVal = int8(v)
			}
			f.Int8Var(fieldValue.Addr().Interface().(*int8), longName, short, defVal, usage)

		case reflect.Int16:
			var defVal int16
			if defaultValue != "" {
				v, _ := strconv.ParseInt(defaultValue, 10, 16)
				defVal = int16(v)
			}
			f.Int16Var(fieldValue.Addr().Interface().(*int16), longName, short, defVal, usage)

		case reflect.Int32:
			var defVal int32
			if defaultValue != "" {
				v, _ := strconv.ParseInt(defaultValue, 10, 32)
				defVal = int32(v)
			}
			f.Int32Var(fieldValue.Addr().Interface().(*int32), longName, short, defVal, usage)

		case reflect.Uint:
			var defVal uint
			if defaultValue != "" {
				v, _ := strconv.ParseUint(defaultValue, 10, 64)
				defVal = uint(v)
			}
			f.UintVar(fieldValue.Addr().Interface().(*uint), longName, short, defVal, usage)

		case reflect.Uint8:
			var defVal uint8
			if defaultValue != "" {
				v, _ := strconv.ParseUint(defaultValue, 10, 8)
				defVal = uint8(v)
			}
			f.Uint8Var(fieldValue.Addr().Interface().(*uint8), longName, short, defVal, usage)

		case reflect.Uint16:
			var defVal uint16
			if defaultValue != "" {
				v, _ := strconv.ParseUint(defaultValue, 10, 16)
				defVal = uint16(v)
			}
			f.Uint16Var(fieldValue.Addr().Interface().(*uint16), longName, short, defVal, usage)

		case reflect.Uint32:
			var defVal uint32
			if defaultValue != "" {
				v, _ := strconv.ParseUint(defaultValue, 10, 32)
				defVal = uint32(v)
			}
			f.Uint32Var(fieldValue.Addr().Interface().(*uint32), longName, short, defVal, usage)

		case reflect.Uint64:
			var defVal uint64
			if defaultValue != "" {
				defVal, _ = strconv.ParseUint(defaultValue, 10, 64)
			}
			f.Uint64Var(fieldValue.Addr().Interface().(*uint64), longName, short, defVal, usage)

		case reflect.Ptr:
			// Handle pointer types - allows distinguishing "not set" from "zero value"
			elemKind := field.Type.Elem().Kind()
			switch elemKind {
			case reflect.Bool:
				p := fieldValue.Addr().Interface().(**bool)
				f.Var(&boolPtrValue{p: p}, longName, short, usage)
			case reflect.String:
				p := fieldValue.Addr().Interface().(**string)
				f.Var(&stringPtrValue{p: p}, longName, short, usage)
			case reflect.Int:
				p := fieldValue.Addr().Interface().(**int)
				f.Var(&intPtrValue{p: p}, longName, short, usage)
			case reflect.Int64:
				// Check if it's a *time.Duration
				if field.Type.Elem() == reflect.TypeOf(time.Duration(0)) {
					p := fieldValue.Addr().Interface().(**time.Duration)
					f.Var(&durationPtrValue{p: p}, longName, short, usage)
				} else {
					p := fieldValue.Addr().Interface().(**int64)
					f.Var(&int64PtrValue{p: p}, longName, short, usage)
				}
			case reflect.Int8:
				p := fieldValue.Addr().Interface().(**int8)
				f.Var(&int8PtrValue{p: p}, longName, short, usage)
			case reflect.Int16:
				p := fieldValue.Addr().Interface().(**int16)
				f.Var(&int16PtrValue{p: p}, longName, short, usage)
			case reflect.Int32:
				p := fieldValue.Addr().Interface().(**int32)
				f.Var(&int32PtrValue{p: p}, longName, short, usage)
			case reflect.Uint:
				p := fieldValue.Addr().Interface().(**uint)
				f.Var(&uintPtrValue{p: p}, longName, short, usage)
			case reflect.Uint8:
				p := fieldValue.Addr().Interface().(**uint8)
				f.Var(&uint8PtrValue{p: p}, longName, short, usage)
			case reflect.Uint16:
				p := fieldValue.Addr().Interface().(**uint16)
				f.Var(&uint16PtrValue{p: p}, longName, short, usage)
			case reflect.Uint32:
				p := fieldValue.Addr().Interface().(**uint32)
				f.Var(&uint32PtrValue{p: p}, longName, short, usage)
			case reflect.Uint64:
				p := fieldValue.Addr().Interface().(**uint64)
				f.Var(&uint64PtrValue{p: p}, longName, short, usage)
			}
		}
	}

	return nil
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
						fmt.Printf("%-30s %s\n", argStr, field.Usage)
					} else {
						fmt.Println(argStr)
					}
				}
			}
		}
	}

	// Show flags if any are defined
	hasFlags := false
	f.VisitAll(func(flag *Flag) {
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
