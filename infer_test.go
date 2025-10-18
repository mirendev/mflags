package mflags

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

// TestInferBasic tests basic inference with a simple struct
func TestInferBasic(t *testing.T) {
	type Config struct {
		Verbose bool   `long:"verbose" short:"v" usage:"Enable verbose output"`
		Output  string `long:"output" short:"o" default:"stdout" usage:"Output file"`
	}

	var capturedConfig *Config
	fn := func(config *Config) error {
		capturedConfig = config
		return nil
	}

	cmd := Infer(fn)

	// Parse flags
	if err := cmd.FlagSet().Parse([]string{"--verbose", "--output=file.txt"}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Run the command
	if err := cmd.Run(cmd.FlagSet(), []string{}); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Verify the config was populated correctly
	if capturedConfig == nil {
		t.Fatal("Config was not captured")
	}
	if !capturedConfig.Verbose {
		t.Error("Expected Verbose to be true")
	}
	if capturedConfig.Output != "file.txt" {
		t.Errorf("Expected Output='file.txt', got '%s'", capturedConfig.Output)
	}
}

// TestInferWithShortFlags tests inference with short flags
func TestInferWithShortFlags(t *testing.T) {
	type Config struct {
		Verbose bool   `long:"verbose" short:"v" usage:"Enable verbose output"`
		Output  string `long:"output" short:"o" default:"stdout" usage:"Output file"`
	}

	var capturedConfig *Config
	fn := func(config *Config) error {
		capturedConfig = config
		return nil
	}

	cmd := Infer(fn)

	// Parse with short flags
	if err := cmd.FlagSet().Parse([]string{"-v", "-o", "file.txt"}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Run the command
	if err := cmd.Run(cmd.FlagSet(), []string{}); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Verify
	if capturedConfig == nil {
		t.Fatal("Config was not captured")
	}
	if !capturedConfig.Verbose {
		t.Error("Expected Verbose to be true")
	}
	if capturedConfig.Output != "file.txt" {
		t.Errorf("Expected Output='file.txt', got '%s'", capturedConfig.Output)
	}
}

// TestInferMultipleTypes tests inference with different data types
func TestInferMultipleTypes(t *testing.T) {
	type Config struct {
		Verbose bool          `long:"verbose" short:"v" usage:"Verbose output"`
		Count   int           `long:"count" short:"c" default:"1" usage:"Number of items"`
		Output  string        `long:"output" short:"o" default:"stdout" usage:"Output file"`
		Timeout time.Duration `long:"timeout" short:"t" default:"30s" usage:"Timeout duration"`
		Tags    []string      `long:"tags" usage:"Tags"`
	}

	var capturedConfig *Config
	fn := func(config *Config) error {
		capturedConfig = config
		return nil
	}

	cmd := Infer(fn)

	// Parse flags
	if err := cmd.FlagSet().Parse([]string{
		"--verbose",
		"--count=5",
		"--output=file.txt",
		"--timeout=1m",
		"--tags=foo,bar,baz",
	}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Run the command
	if err := cmd.Run(cmd.FlagSet(), []string{}); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Verify
	if capturedConfig == nil {
		t.Fatal("Config was not captured")
	}
	if !capturedConfig.Verbose {
		t.Error("Expected Verbose to be true")
	}
	if capturedConfig.Count != 5 {
		t.Errorf("Expected Count=5, got %d", capturedConfig.Count)
	}
	if capturedConfig.Output != "file.txt" {
		t.Errorf("Expected Output='file.txt', got '%s'", capturedConfig.Output)
	}
	if capturedConfig.Timeout != time.Minute {
		t.Errorf("Expected Timeout=1m, got %v", capturedConfig.Timeout)
	}
	if len(capturedConfig.Tags) != 3 || capturedConfig.Tags[0] != "foo" || capturedConfig.Tags[1] != "bar" || capturedConfig.Tags[2] != "baz" {
		t.Errorf("Expected Tags=[foo bar baz], got %v", capturedConfig.Tags)
	}
}

// TestInferWithPositional tests inference with positional arguments
func TestInferWithPositional(t *testing.T) {
	type Config struct {
		Environment string `position:"0" usage:"Target environment"`
		Version     string `position:"1" usage:"Version to deploy"`
		DryRun      bool   `long:"dry-run" usage:"Dry run mode"`
	}

	var capturedConfig *Config
	fn := func(config *Config) error {
		capturedConfig = config
		return nil
	}

	cmd := Infer(fn)

	// Parse flags
	if err := cmd.FlagSet().Parse([]string{"production", "v1.2.3", "--dry-run"}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Run the command
	if err := cmd.Run(cmd.FlagSet(), []string{}); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Verify
	if capturedConfig == nil {
		t.Fatal("Config was not captured")
	}
	if capturedConfig.Environment != "production" {
		t.Errorf("Expected Environment='production', got '%s'", capturedConfig.Environment)
	}
	if capturedConfig.Version != "v1.2.3" {
		t.Errorf("Expected Version='v1.2.3', got '%s'", capturedConfig.Version)
	}
	if !capturedConfig.DryRun {
		t.Error("Expected DryRun to be true")
	}
}

// TestInferWithRest tests inference with rest arguments
func TestInferWithRest(t *testing.T) {
	type Config struct {
		Verbose bool     `long:"verbose" short:"v" usage:"Verbose output"`
		Files   []string `rest:"true" usage:"Files to process"`
	}

	var capturedConfig *Config
	fn := func(config *Config) error {
		capturedConfig = config
		return nil
	}

	cmd := Infer(fn)

	// Parse flags
	if err := cmd.FlagSet().Parse([]string{"--verbose", "file1.txt", "file2.txt", "file3.txt"}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Run the command
	if err := cmd.Run(cmd.FlagSet(), []string{}); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Verify
	if capturedConfig == nil {
		t.Fatal("Config was not captured")
	}
	if !capturedConfig.Verbose {
		t.Error("Expected Verbose to be true")
	}
	if len(capturedConfig.Files) != 3 {
		t.Errorf("Expected 3 files, got %d", len(capturedConfig.Files))
	}
	if capturedConfig.Files[0] != "file1.txt" || capturedConfig.Files[1] != "file2.txt" || capturedConfig.Files[2] != "file3.txt" {
		t.Errorf("Expected Files=[file1.txt file2.txt file3.txt], got %v", capturedConfig.Files)
	}
}

// TestInferWithUnknown tests inference with unknown flag handling
func TestInferWithUnknown(t *testing.T) {
	type Config struct {
		Port         int      `long:"port" short:"p" default:"8080" usage:"Port"`
		UnknownFlags []string `unknown:"true" usage:"Unknown flags"`
	}

	var capturedConfig *Config
	fn := func(config *Config) error {
		capturedConfig = config
		return nil
	}

	cmd := Infer(fn)

	// Parse flags
	if err := cmd.FlagSet().Parse([]string{"--port=9000", "--unknown", "value", "-x"}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Run the command
	if err := cmd.Run(cmd.FlagSet(), []string{}); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Verify
	if capturedConfig == nil {
		t.Fatal("Config was not captured")
	}
	if capturedConfig.Port != 9000 {
		t.Errorf("Expected Port=9000, got %d", capturedConfig.Port)
	}
	if len(capturedConfig.UnknownFlags) != 3 {
		t.Errorf("Expected 3 unknown flags, got %d: %v", len(capturedConfig.UnknownFlags), capturedConfig.UnknownFlags)
	}
}

// TestInferWithUsage tests that WithUsage option works
func TestInferWithUsage(t *testing.T) {
	type Config struct {
		Verbose bool `long:"verbose" usage:"Verbose output"`
	}

	fn := func(config *Config) error {
		return nil
	}

	cmd := Infer(fn, WithUsage("Deploy the application"))

	if cmd.Usage() != "Deploy the application" {
		t.Errorf("Expected usage='Deploy the application', got '%s'", cmd.Usage())
	}
}

// TestInferWithOutputFormat tests that WithOutputFormat option works
func TestInferWithOutputFormat(t *testing.T) {
	type Config struct {
		Verbose bool `long:"verbose" usage:"Verbose output"`
	}

	fn := func(config *Config) error {
		return nil
	}

	cmd := Infer(fn, WithOutputFormat(OutputFormatJSON))

	// Type assert to OutputFormatter to access OutputFormat method
	formatter, ok := cmd.(OutputFormatter)
	if !ok {
		t.Fatal("Expected command to implement OutputFormatter")
	}

	if formatter.OutputFormat() != OutputFormatJSON {
		t.Errorf("Expected OutputFormat=JSON, got %v", formatter.OutputFormat())
	}
}

// TestInferReturnsError tests that function errors are propagated
func TestInferReturnsError(t *testing.T) {
	type Config struct {
		Fail bool `long:"fail" usage:"Return an error"`
	}

	expectedErr := errors.New("test error")
	fn := func(config *Config) error {
		if config.Fail {
			return expectedErr
		}
		return nil
	}

	cmd := Infer(fn)

	// Parse flags
	if err := cmd.FlagSet().Parse([]string{"--fail"}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Run the command - should return the error
	err := cmd.Run(cmd.FlagSet(), []string{})
	if err != expectedErr {
		t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
	}
}

// TestInferWithDispatcher tests using Infer with a Dispatcher
func TestInferWithDispatcher(t *testing.T) {
	type DeployConfig struct {
		Environment string `position:"0" usage:"Target environment"`
		DryRun      bool   `long:"dry-run" usage:"Dry run mode"`
	}

	var capturedConfig *DeployConfig
	deployFn := func(config *DeployConfig) error {
		capturedConfig = config
		return nil
	}

	dispatcher := NewDispatcher("testapp")
	dispatcher.Dispatch("deploy", Infer(deployFn, WithUsage("Deploy the application")))

	// Run the dispatcher
	if err := dispatcher.Run([]string{"deploy", "production", "--dry-run"}); err != nil {
		t.Fatalf("Dispatcher.Run failed: %v", err)
	}

	// Verify
	if capturedConfig == nil {
		t.Fatal("Config was not captured")
	}
	if capturedConfig.Environment != "production" {
		t.Errorf("Expected Environment='production', got '%s'", capturedConfig.Environment)
	}
	if !capturedConfig.DryRun {
		t.Error("Expected DryRun to be true")
	}
}

// TestInferPanicNotFunction tests that Infer panics if not given a function
func TestInferPanicNotFunction(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic, but didn't get one")
		} else {
			msg := fmt.Sprintf("%v", r)
			if msg != "Infer: argument must be a function, got string" {
				t.Errorf("Expected panic message about not a function, got: %s", msg)
			}
		}
	}()

	Infer("not a function")
}

// TestInferPanicWrongParamCount tests that Infer panics with wrong parameter count
func TestInferPanicWrongParamCount(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic, but didn't get one")
		} else {
			msg := fmt.Sprintf("%v", r)
			if msg != "Infer: function must have exactly 1 parameter, got 0" {
				t.Errorf("Expected panic message about parameter count, got: %s", msg)
			}
		}
	}()

	fn := func() error {
		return nil
	}
	Infer(fn)
}

// TestInferPanicWrongReturnCount tests that Infer panics with wrong return count
func TestInferPanicWrongReturnCount(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic, but didn't get one")
		} else {
			msg := fmt.Sprintf("%v", r)
			if msg != "Infer: function must return exactly 1 value (error), got 0" {
				t.Errorf("Expected panic message about return count, got: %s", msg)
			}
		}
	}()

	type Config struct {
		Value string `long:"value"`
	}
	fn := func(config *Config) {
	}
	Infer(fn)
}

// TestInferPanicWrongReturnType tests that Infer panics with wrong return type
func TestInferPanicWrongReturnType(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic, but didn't get one")
		} else {
			msg := fmt.Sprintf("%v", r)
			if msg != "Infer: function must return error, got string" {
				t.Errorf("Expected panic message about return type, got: %s", msg)
			}
		}
	}()

	type Config struct {
		Value string `long:"value"`
	}
	fn := func(config *Config) string {
		return "not an error"
	}
	Infer(fn)
}

// TestInferPanicParamNotPointer tests that Infer panics if parameter is not a pointer
func TestInferPanicParamNotPointer(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic, but didn't get one")
		} else {
			msg := fmt.Sprintf("%v", r)
			if msg != "Infer: function parameter must be a pointer to a struct, got struct" {
				t.Errorf("Expected panic message about not a pointer, got: %s", msg)
			}
		}
	}()

	type Config struct {
		Value string `long:"value"`
	}
	fn := func(config Config) error {
		return nil
	}
	Infer(fn)
}

// TestInferPanicParamNotStructPointer tests that Infer panics if parameter is not a pointer to struct
func TestInferPanicParamNotStructPointer(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic, but didn't get one")
		} else {
			msg := fmt.Sprintf("%v", r)
			if msg != "Infer: function parameter must be a pointer to a struct, got pointer to string" {
				t.Errorf("Expected panic message about not pointer to struct, got: %s", msg)
			}
		}
	}()

	fn := func(config *string) error {
		return nil
	}
	Infer(fn)
}

// TestInferDefaultValues tests that default values from struct tags work
func TestInferDefaultValues(t *testing.T) {
	type Config struct {
		Port   int    `long:"port" default:"8080" usage:"Port"`
		Host   string `long:"host" default:"localhost" usage:"Host"`
		Debug  bool   `long:"debug" usage:"Debug mode"`
	}

	var capturedConfig *Config
	fn := func(config *Config) error {
		capturedConfig = config
		return nil
	}

	cmd := Infer(fn)

	// Parse with no flags - should use defaults
	if err := cmd.FlagSet().Parse([]string{}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Run the command
	if err := cmd.Run(cmd.FlagSet(), []string{}); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Verify defaults
	if capturedConfig == nil {
		t.Fatal("Config was not captured")
	}
	if capturedConfig.Port != 8080 {
		t.Errorf("Expected Port=8080, got %d", capturedConfig.Port)
	}
	if capturedConfig.Host != "localhost" {
		t.Errorf("Expected Host='localhost', got '%s'", capturedConfig.Host)
	}
	if capturedConfig.Debug {
		t.Error("Expected Debug to be false")
	}
}
