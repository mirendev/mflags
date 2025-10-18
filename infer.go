package mflags

import (
	"fmt"
	"reflect"
)

// inferredCommand is a Command implementation that uses reflection to infer flags from a function signature
type inferredCommand struct {
	fn           reflect.Value
	configType   reflect.Type
	configValue  reflect.Value
	flags        *FlagSet
	usage        string
	outputFormat OutputFormat
}

// Infer creates a Command from a function using reflection.
// The function must have the signature: func(*ConfigStruct) error
// where ConfigStruct is a struct type with mflags struct tags.
//
// Example:
//
//	type DeployConfig struct {
//	    Environment string `position:"0" usage:"Target environment"`
//	    DryRun      bool   `long:"dry-run" usage:"Simulate deployment"`
//	}
//
//	func deploy(config *DeployConfig) error {
//	    fmt.Printf("Deploying to %s\n", config.Environment)
//	    return nil
//	}
//
//	cmd := mflags.Infer(deploy, mflags.WithUsage("Deploy the application"))
//	dispatcher.Dispatch("deploy", cmd)
func Infer(fn interface{}, opts ...CommandOption) Command {
	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	// Validate function signature
	if fnType.Kind() != reflect.Func {
		panic(fmt.Sprintf("Infer: argument must be a function, got %v", fnType.Kind()))
	}

	if fnType.NumIn() != 1 {
		panic(fmt.Sprintf("Infer: function must have exactly 1 parameter, got %d", fnType.NumIn()))
	}

	if fnType.NumOut() != 1 {
		panic(fmt.Sprintf("Infer: function must return exactly 1 value (error), got %d", fnType.NumOut()))
	}

	// Check that return type is error
	errorInterface := reflect.TypeOf((*error)(nil)).Elem()
	if !fnType.Out(0).Implements(errorInterface) {
		panic(fmt.Sprintf("Infer: function must return error, got %v", fnType.Out(0)))
	}

	// Check that first parameter is a pointer to a struct
	paramType := fnType.In(0)
	if paramType.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("Infer: function parameter must be a pointer to a struct, got %v", paramType.Kind()))
	}

	structType := paramType.Elem()
	if structType.Kind() != reflect.Struct {
		panic(fmt.Sprintf("Infer: function parameter must be a pointer to a struct, got pointer to %v", structType.Kind()))
	}

	// Create an instance of the config struct
	configValue := reflect.New(structType)

	// Create a FlagSet and populate it from the struct
	flags := NewFlagSet("")
	if err := flags.FromStruct(configValue.Interface()); err != nil {
		panic(fmt.Sprintf("Infer: error creating flags from struct: %v", err))
	}

	cmd := &inferredCommand{
		fn:           fnValue,
		configType:   structType,
		configValue:  configValue,
		flags:        flags,
		usage:        "",
		outputFormat: OutputFormatRaw,
	}

	// Apply options
	for _, opt := range opts {
		// Use the funcCommand option application
		fc := &funcCommand{usage: cmd.usage, outputFormat: cmd.outputFormat}
		opt(fc)
		cmd.usage = fc.usage
		cmd.outputFormat = fc.outputFormat
	}

	return cmd
}

// FlagSet returns the flagset for this command
func (c *inferredCommand) FlagSet() *FlagSet {
	return c.flags
}

// Run executes the command by calling the inferred function with the parsed config
func (c *inferredCommand) Run(fs *FlagSet, args []string) error {
	// Call the function with the config struct
	results := c.fn.Call([]reflect.Value{c.configValue})

	// Extract the error return value
	errValue := results[0].Interface()
	if errValue != nil {
		return errValue.(error)
	}
	return nil
}

// Usage returns the usage description for this command
func (c *inferredCommand) Usage() string {
	return c.usage
}

// OutputFormat returns the output format for this command
func (c *inferredCommand) OutputFormat() OutputFormat {
	return c.outputFormat
}
