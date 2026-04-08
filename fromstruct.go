package mflags

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

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

// knownTags is the set of struct tag keys that FromStruct knows how to handle.
// Keep this in sync with the Tag.Get() and getTagValues() calls in FromStruct.
var knownTags = map[string]bool{
	"long":        true,
	"short":       true,
	"default":     true,
	"env":         true,
	"required":    true,
	"usage":       true,
	"description": true,
	"choice":      true,
	"position":    true,
	"rest":        true,
	"unknown":     true,
	"group":       true,
}

// validateStructTags checks that every struct tag on exported fields is one
// that FromStruct actually reads. It returns an error listing all unrecognized
// tags so the caller can fix them all in one pass.
func validateStructTags(rt reflect.Type) error {
	var errs []string
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)

		// Skip the blank identifier used for group declarations
		if field.Name == "_" {
			continue
		}

		if !field.IsExported() {
			continue
		}

		// Recurse into embedded structs
		ft := field.Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}
		if field.Anonymous && ft.Kind() == reflect.Struct {
			if err := validateStructTags(ft); err != nil {
				errs = append(errs, err.Error())
			}
			continue
		}

		// Parse the raw tag string into key:"value" pairs and check each key
		tagStr := string(field.Tag)
		for tagStr != "" {
			// Skip leading spaces
			tagStr = strings.TrimLeft(tagStr, " ")
			if tagStr == "" {
				break
			}

			// Find the key (everything before the colon)
			colon := strings.Index(tagStr, ":")
			if colon < 0 {
				break
			}
			key := tagStr[:colon]
			tagStr = tagStr[colon+1:]

			// Skip past the quoted value
			if len(tagStr) == 0 || tagStr[0] != '"' {
				break
			}
			end := 1
			for end < len(tagStr) && tagStr[end] != '"' {
				if tagStr[end] == '\\' && end+1 < len(tagStr) {
					end += 2
				} else {
					end++
				}
			}
			if end >= len(tagStr) {
				break
			}
			tagStr = tagStr[end+1:]

			if !knownTags[key] {
				errs = append(errs, fmt.Sprintf("unknown struct tag %q on field %s in %s", key, field.Name, rt.Name()))
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("invalid struct tags:\n  %s", strings.Join(errs, "\n  "))
	}
	return nil
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

// FromStructOption configures how FromStruct processes a struct.
type FromStructOption func(*fromStructConfig)

type fromStructConfig struct {
	group string
}

// InGroup sets the group name for all flags created by FromStruct.
func InGroup(name string) FromStructOption {
	return func(c *fromStructConfig) { c.group = name }
}

// FromStruct creates flag definitions from a struct's fields using struct tags.
// The argument must be a pointer to a struct. Struct tags control how fields are parsed:
//   - `long:"name"` - long flag name (defaults to lowercase field name)
//   - `short:"x"` - short flag name (single character)
//   - `default:"value"` - default value for the flag
//   - `env:"VAR_NAME"` - populate default from an environment variable (overrides default, overridden by CLI)
//   - `required:"true"` - return a parse error if the flag/positional wasn't provided
//   - `usage:"description"` - usage description
//   - `description:"description"` - alternate usage description
//   - `choice:"value"` - constrain string field to specific values (can be repeated for multiple choices)
//   - `position:"0"` - positional argument at index 0
//   - `rest:"true"` - capture all remaining arguments in a []string field
//   - `unknown:"true"` - capture unknown flags in a []string field (automatically enables AllowUnknownFlags)
//   - `group:"name"` - on a `_ struct{}` field, declares the group for all flags in the struct
//   - `group:"name"` - on an embedded struct field, overrides the embedded struct's self-declared group
//
// Supports bool, string, int, []string, and time.Duration field types.
// Anonymous embedded structs are recursively processed.
func (f *FlagSet) FromStruct(v any, opts ...FromStructOption) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("FromStruct requires a non-nil pointer to a struct")
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("FromStruct requires a pointer to a struct")
	}

	// Apply options
	var cfg fromStructConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	// Save and restore currentGroup
	prevGroup := f.currentGroup
	defer func() { f.currentGroup = prevGroup }()

	if cfg.group != "" {
		f.currentGroup = cfg.group
	}

	rt := rv.Type()

	// Validate that all struct tags are ones we know how to handle
	if err := validateStructTags(rt); err != nil {
		return err
	}

	// First pass: check for self-declared group via `_ struct{} \`group:"..."\``
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if field.Name == "_" && field.Type == reflect.TypeOf(struct{}{}) {
			if groupTag := field.Tag.Get("group"); groupTag != "" && f.currentGroup == "" {
				f.currentGroup = groupTag
			}
			break
		}
	}

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)

		// Skip the `_` group declaration field
		if field.Name == "_" {
			continue
		}

		if !field.IsExported() {
			continue
		}

		fieldValue := rv.Field(i)
		if !fieldValue.CanAddr() {
			continue
		}

		// Check for anonymous/embedded struct fields and descend into them
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			// Check for group tag on the embedding site
			if groupTag := field.Tag.Get("group"); groupTag != "" {
				if err := f.FromStruct(fieldValue.Addr().Interface(), InGroup(groupTag)); err != nil {
					return err
				}
			} else {
				if err := f.FromStruct(fieldValue.Addr().Interface()); err != nil {
					return err
				}
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
				posEnvVar := field.Tag.Get("env")
				posRequired := field.Tag.Get("required") == "true"
				posHasValue := false

				// Environment variable sets the positional default
				if posEnvVar != "" {
					if envVal, ok := os.LookupEnv(posEnvVar); ok {
						if err := setFieldValue(fieldValue, envVal); err != nil {
							return fmt.Errorf("invalid value for env var %s: %w", posEnvVar, err)
						}
						posHasValue = true
					}
				}

				f.posFields[pos] = &PositionalField{
					Name:     field.Name,
					Usage:    posUsage,
					Value:    fieldValue,
					Type:     field.Type,
					EnvVar:   posEnvVar,
					Required: posRequired,
					HasValue: posHasValue,
				}

				if posRequired {
					f.requiredPos = append(f.requiredPos, pos)
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
		envVar := field.Tag.Get("env")
		required := field.Tag.Get("required") == "true"

		usage := field.Tag.Get("usage")
		if usage == "" {
			usage = field.Tag.Get("description")
			if usage == "" {
				usage = fmt.Sprintf("%s value", field.Name)
			}
		}

		if required {
			f.requiredFlags = append(f.requiredFlags, longName)
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

		// Set env/required metadata on the registered flag, and apply
		// env var value through the flag's Value.Set path so it gets
		// validated and works for all types including pointers and slices.
		if flag, ok := f.flags[longName]; ok {
			flag.EnvVar = envVar
			flag.Required = required
			if envVar != "" {
				if envVal, ok := os.LookupEnv(envVar); ok {
					if err := flag.Value.Set(envVal); err != nil {
						return fmt.Errorf("invalid value for env var %s: %w", envVar, err)
					}
					flag.HasValue = true
				}
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
