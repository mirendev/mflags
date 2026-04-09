package mflags

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

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
	values     *[]string
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
