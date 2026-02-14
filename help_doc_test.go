package mflags

import (
	"encoding/json"
	"testing"
)

func TestFlagSetHelpDoc_BasicFlags(t *testing.T) {
	fs := NewFlagSet("test")
	fs.String("name", 'n', "default-name", "Set the name")
	fs.Bool("verbose", 'v', false, "Enable verbose output")
	fs.Int("count", 'c', 5, "Set count")

	doc := fs.HelpDoc()

	if doc.Name != "test" {
		t.Errorf("expected name %q, got %q", "test", doc.Name)
	}

	if len(doc.Flags) != 3 {
		t.Fatalf("expected 3 flags, got %d", len(doc.Flags))
	}

	// Flags are sorted by name via VisitAll
	// count, name, verbose
	assertFlag(t, doc.Flags[0], "count", "c", "int", "5", "Set count", false)
	assertFlag(t, doc.Flags[1], "name", "n", "string", "default-name", "Set the name", false)
	assertFlag(t, doc.Flags[2], "verbose", "v", "bool", "false", "Enable verbose output", true)
}

func TestFlagSetHelpDoc_ChoiceFlags(t *testing.T) {
	fs := NewFlagSet("test")
	fs.Choice("env", 'e', "staging", []string{"dev", "staging", "prod"}, "Target environment")

	doc := fs.HelpDoc()

	if len(doc.Flags) != 1 {
		t.Fatalf("expected 1 flag, got %d", len(doc.Flags))
	}

	f := doc.Flags[0]
	if len(f.Choices) != 3 {
		t.Fatalf("expected 3 choices, got %d", len(f.Choices))
	}
	if f.Choices[0] != "dev" || f.Choices[1] != "staging" || f.Choices[2] != "prod" {
		t.Errorf("unexpected choices: %v", f.Choices)
	}
}

func TestFlagSetHelpDoc_PositionalArgs(t *testing.T) {
	fs := NewFlagSet("test")
	fs.StringPos("target", 0, "", "Deployment target")
	fs.IntPos("port", 1, 8080, "Port number")

	doc := fs.HelpDoc()

	if len(doc.PositionalArgs) != 2 {
		t.Fatalf("expected 2 positional args, got %d", len(doc.PositionalArgs))
	}

	if doc.PositionalArgs[0].Name != "target" {
		t.Errorf("expected positional name %q, got %q", "target", doc.PositionalArgs[0].Name)
	}
	if doc.PositionalArgs[0].Usage != "Deployment target" {
		t.Errorf("expected positional usage %q, got %q", "Deployment target", doc.PositionalArgs[0].Usage)
	}
	if doc.PositionalArgs[0].Type != "string" {
		t.Errorf("expected positional type %q, got %q", "string", doc.PositionalArgs[0].Type)
	}
	if doc.PositionalArgs[1].Name != "port" {
		t.Errorf("expected positional name %q, got %q", "port", doc.PositionalArgs[1].Name)
	}
	if doc.PositionalArgs[1].Type != "int" {
		t.Errorf("expected positional type %q, got %q", "int", doc.PositionalArgs[1].Type)
	}
}

func TestFlagSetHelpDoc_RestArgs(t *testing.T) {
	fs := NewFlagSet("test")
	var rest []string
	fs.Rest(&rest, "remaining arguments")

	doc := fs.HelpDoc()

	if !doc.AcceptsRestArgs {
		t.Error("expected AcceptsRestArgs to be true")
	}
}

func TestFlagSetHelpDoc_NoRestArgs(t *testing.T) {
	fs := NewFlagSet("test")

	doc := fs.HelpDoc()

	if doc.AcceptsRestArgs {
		t.Error("expected AcceptsRestArgs to be false")
	}
}

func TestFlagSetHelpDoc_FlagGroups(t *testing.T) {
	fs := NewFlagSet("test")
	fs.Group("Advanced")
	fs.String("config", 0, "", "Config file path")
	fs.Group("Debug")
	fs.Bool("trace", 0, false, "Enable tracing")
	fs.Group("")
	fs.Bool("verbose", 'v', false, "Verbose output")

	doc := fs.HelpDoc()

	if len(doc.FlagGroups) != 2 {
		t.Fatalf("expected 2 flag groups, got %d", len(doc.FlagGroups))
	}
	if doc.FlagGroups[0] != "Advanced" {
		t.Errorf("expected group %q, got %q", "Advanced", doc.FlagGroups[0])
	}
	if doc.FlagGroups[1] != "Debug" {
		t.Errorf("expected group %q, got %q", "Debug", doc.FlagGroups[1])
	}

	// Verify flags have correct groups
	for _, f := range doc.Flags {
		switch f.Name {
		case "config":
			if f.Group != "Advanced" {
				t.Errorf("expected config group %q, got %q", "Advanced", f.Group)
			}
		case "trace":
			if f.Group != "Debug" {
				t.Errorf("expected trace group %q, got %q", "Debug", f.Group)
			}
		case "verbose":
			if f.Group != "" {
				t.Errorf("expected verbose group %q, got %q", "", f.Group)
			}
		}
	}
}

func TestFlagSetHelpDoc_EmptyFlagSet(t *testing.T) {
	fs := NewFlagSet("empty")

	doc := fs.HelpDoc()

	if doc.Name != "empty" {
		t.Errorf("expected name %q, got %q", "empty", doc.Name)
	}
	if len(doc.Flags) != 0 {
		t.Errorf("expected 0 flags, got %d", len(doc.Flags))
	}
	if len(doc.FlagGroups) != 0 {
		t.Errorf("expected 0 flag groups, got %d", len(doc.FlagGroups))
	}
	if len(doc.PositionalArgs) != 0 {
		t.Errorf("expected 0 positional args, got %d", len(doc.PositionalArgs))
	}
	if doc.AcceptsRestArgs {
		t.Error("expected AcceptsRestArgs to be false")
	}
}

func TestDispatcherHelpDoc_BasicCommand(t *testing.T) {
	d := NewDispatcher("myapp")

	fs := NewFlagSet("deploy")
	fs.String("target", 't', "", "Deploy target")

	d.Dispatch("deploy", NewCommand(fs, nil, WithUsage("Deploy the application")))

	doc := d.HelpDoc()

	if doc.Name != "myapp" {
		t.Errorf("expected name %q, got %q", "myapp", doc.Name)
	}
	if len(doc.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(doc.Commands))
	}

	cmd := doc.Commands[0]
	if cmd.Path != "deploy" {
		t.Errorf("expected path %q, got %q", "deploy", cmd.Path)
	}
	if cmd.Usage != "Deploy the application" {
		t.Errorf("expected usage %q, got %q", "Deploy the application", cmd.Usage)
	}
	if len(cmd.Flags) != 1 {
		t.Fatalf("expected 1 flag, got %d", len(cmd.Flags))
	}
	if cmd.Flags[0].Name != "target" {
		t.Errorf("expected flag name %q, got %q", "target", cmd.Flags[0].Name)
	}
}

func TestDispatcherHelpDoc_NestedCommands(t *testing.T) {
	d := NewDispatcher("myapp")

	fs1 := NewFlagSet("deploy")
	d.Dispatch("deploy", NewCommand(fs1, nil, WithUsage("Deploy")))

	fs2 := NewFlagSet("deploy prod")
	fs2.Bool("force", 'f', false, "Force deploy")
	d.Dispatch("deploy prod", NewCommand(fs2, nil, WithUsage("Deploy to production")))

	fs3 := NewFlagSet("deploy staging")
	d.Dispatch("deploy staging", NewCommand(fs3, nil, WithUsage("Deploy to staging")))

	doc := d.HelpDoc()

	if len(doc.Commands) != 1 {
		t.Fatalf("expected 1 top-level command, got %d", len(doc.Commands))
	}

	deploy := doc.Commands[0]
	if deploy.Path != "deploy" {
		t.Errorf("expected path %q, got %q", "deploy", deploy.Path)
	}
	if len(deploy.Subcommands) != 2 {
		t.Fatalf("expected 2 subcommands, got %d", len(deploy.Subcommands))
	}

	// Subcommands are sorted by name
	prod := deploy.Subcommands[0]
	staging := deploy.Subcommands[1]

	if prod.Path != "deploy prod" {
		t.Errorf("expected path %q, got %q", "deploy prod", prod.Path)
	}
	if prod.Usage != "Deploy to production" {
		t.Errorf("expected usage %q, got %q", "Deploy to production", prod.Usage)
	}
	if len(prod.Flags) != 1 {
		t.Fatalf("expected 1 flag on prod, got %d", len(prod.Flags))
	}
	if prod.Flags[0].Name != "force" {
		t.Errorf("expected flag %q, got %q", "force", prod.Flags[0].Name)
	}

	if staging.Path != "deploy staging" {
		t.Errorf("expected path %q, got %q", "deploy staging", staging.Path)
	}
}

func TestDispatcherHelpDoc_ImplicitNamespace(t *testing.T) {
	d := NewDispatcher("myapp")

	// Register "debug entity list" and "debug entity show" without registering "debug" or "debug entity"
	fs1 := NewFlagSet("list")
	d.Dispatch("debug entity list", NewCommand(fs1, nil, WithUsage("List entities")))

	fs2 := NewFlagSet("show")
	d.Dispatch("debug entity show", NewCommand(fs2, nil, WithUsage("Show entity details")))

	doc := d.HelpDoc()

	if len(doc.Commands) != 1 {
		t.Fatalf("expected 1 top-level command, got %d", len(doc.Commands))
	}

	debug := doc.Commands[0]
	if debug.Path != "debug" {
		t.Errorf("expected path %q, got %q", "debug", debug.Path)
	}
	if debug.Usage != "" {
		t.Errorf("expected empty usage for namespace, got %q", debug.Usage)
	}
	// Namespace should have no flags
	if len(debug.Flags) != 0 {
		t.Errorf("expected 0 flags for namespace, got %d", len(debug.Flags))
	}

	// Should have "entity" as subcommand (also a namespace)
	if len(debug.Subcommands) != 1 {
		t.Fatalf("expected 1 subcommand under debug, got %d", len(debug.Subcommands))
	}

	entity := debug.Subcommands[0]
	if entity.Path != "debug entity" {
		t.Errorf("expected path %q, got %q", "debug entity", entity.Path)
	}
	if len(entity.Subcommands) != 2 {
		t.Fatalf("expected 2 subcommands under entity, got %d", len(entity.Subcommands))
	}

	if entity.Subcommands[0].Path != "debug entity list" {
		t.Errorf("expected path %q, got %q", "debug entity list", entity.Subcommands[0].Path)
	}
	if entity.Subcommands[1].Path != "debug entity show" {
		t.Errorf("expected path %q, got %q", "debug entity show", entity.Subcommands[1].Path)
	}
}

func TestDispatcherHelpDoc_EmptyDispatcher(t *testing.T) {
	d := NewDispatcher("myapp")

	doc := d.HelpDoc()

	if doc.Name != "myapp" {
		t.Errorf("expected name %q, got %q", "myapp", doc.Name)
	}
	if doc.Commands == nil {
		t.Fatal("expected non-nil commands slice")
	}
	if len(doc.Commands) != 0 {
		t.Errorf("expected 0 commands, got %d", len(doc.Commands))
	}
}

func TestDispatcherHelpJSON_Roundtrip(t *testing.T) {
	d := NewDispatcher("myapp")

	fs := NewFlagSet("greet")
	fs.String("name", 'n', "world", "Name to greet")
	fs.Bool("loud", 'l', false, "Shout the greeting")
	fs.Choice("lang", 0, "en", []string{"en", "es", "fr"}, "Language")
	fs.StringPos("target", 0, "", "Greeting target")
	var rest []string
	fs.Rest(&rest, "extra args")
	fs.Group("Output")
	fs.String("format", 'f', "text", "Output format")

	d.Dispatch("greet", NewCommand(fs, nil, WithUsage("Greet someone")))

	data, err := d.HelpJSON()
	if err != nil {
		t.Fatalf("HelpJSON failed: %v", err)
	}

	// Verify it's valid JSON by unmarshalling
	var roundtrip HelpDocument
	if err := json.Unmarshal(data, &roundtrip); err != nil {
		t.Fatalf("JSON roundtrip unmarshal failed: %v", err)
	}

	if roundtrip.Name != "myapp" {
		t.Errorf("roundtrip name: expected %q, got %q", "myapp", roundtrip.Name)
	}
	if len(roundtrip.Commands) != 1 {
		t.Fatalf("roundtrip: expected 1 command, got %d", len(roundtrip.Commands))
	}

	cmd := roundtrip.Commands[0]
	if cmd.Path != "greet" {
		t.Errorf("roundtrip path: expected %q, got %q", "greet", cmd.Path)
	}
	if !cmd.AcceptsRestArgs {
		t.Error("roundtrip: expected AcceptsRestArgs true")
	}
	if len(cmd.PositionalArgs) != 1 {
		t.Fatalf("roundtrip: expected 1 positional arg, got %d", len(cmd.PositionalArgs))
	}
	if cmd.PositionalArgs[0].Name != "target" {
		t.Errorf("roundtrip positional: expected %q, got %q", "target", cmd.PositionalArgs[0].Name)
	}

	// Verify choices survived roundtrip
	var langFlag *FlagDoc
	for i := range cmd.Flags {
		if cmd.Flags[i].Name == "lang" {
			langFlag = &cmd.Flags[i]
			break
		}
	}
	if langFlag == nil {
		t.Fatal("roundtrip: lang flag not found")
	}
	if len(langFlag.Choices) != 3 {
		t.Errorf("roundtrip: expected 3 choices, got %d", len(langFlag.Choices))
	}
}

func TestFlagSetHelpJSON(t *testing.T) {
	fs := NewFlagSet("standalone")
	fs.String("output", 'o', "", "Output file")

	data, err := fs.HelpJSON()
	if err != nil {
		t.Fatalf("HelpJSON failed: %v", err)
	}

	var doc FlagSetDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if doc.Name != "standalone" {
		t.Errorf("expected name %q, got %q", "standalone", doc.Name)
	}
	if len(doc.Flags) != 1 {
		t.Fatalf("expected 1 flag, got %d", len(doc.Flags))
	}
	if doc.Flags[0].Name != "output" {
		t.Errorf("expected flag name %q, got %q", "output", doc.Flags[0].Name)
	}
	if doc.Flags[0].Short != "o" {
		t.Errorf("expected short %q, got %q", "o", doc.Flags[0].Short)
	}
}

func TestDispatcherHelpDoc_CommandWithRestArgs(t *testing.T) {
	d := NewDispatcher("myapp")

	fs := NewFlagSet("exec")
	var rest []string
	fs.Rest(&rest, "command and arguments")

	d.Dispatch("exec", NewCommand(fs, nil, WithUsage("Execute a command")))

	doc := d.HelpDoc()

	if len(doc.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(doc.Commands))
	}
	if !doc.Commands[0].AcceptsRestArgs {
		t.Error("expected AcceptsRestArgs to be true")
	}
}

func TestFlagSetHelpDoc_EmptyChoices(t *testing.T) {
	fs := NewFlagSet("test")
	fs.String("name", 0, "", "A name")

	doc := fs.HelpDoc()

	if len(doc.Flags) != 1 {
		t.Fatalf("expected 1 flag, got %d", len(doc.Flags))
	}
	// Non-choice flags should have empty (not nil) choices slice
	if doc.Flags[0].Choices == nil {
		t.Error("expected non-nil choices slice")
	}
	if len(doc.Flags[0].Choices) != 0 {
		t.Errorf("expected 0 choices, got %d", len(doc.Flags[0].Choices))
	}
}

func TestDispatcherHelpDoc_EmptySubarrays(t *testing.T) {
	d := NewDispatcher("myapp")

	fs := NewFlagSet("simple")
	d.Dispatch("simple", NewCommand(fs, nil, WithUsage("A simple command")))

	doc := d.HelpDoc()
	cmd := doc.Commands[0]

	// Flags and FlagGroups should be non-nil (always present in JSON)
	if cmd.Flags == nil {
		t.Error("expected non-nil flags")
	}
	if cmd.FlagGroups == nil {
		t.Error("expected non-nil flag groups")
	}

	// Verify JSON output: flags/flagGroups always present, omitempty fields omitted
	data, err := d.HelpJSON()
	if err != nil {
		t.Fatalf("HelpJSON failed: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	cmds := raw["commands"].([]interface{})
	cmdMap := cmds[0].(map[string]interface{})

	// flags and flagGroups should always be present as empty arrays
	for _, field := range []string{"flags", "flagGroups"} {
		if cmdMap[field] == nil {
			t.Errorf("expected %q to be an empty array in JSON, got null", field)
		}
	}

	// omitempty fields should be omitted when empty
	for _, field := range []string{"positionalArgs", "subcommands", "examples"} {
		if _, exists := cmdMap[field]; exists {
			t.Errorf("expected %q to be omitted from JSON when empty", field)
		}
	}
}

func TestDispatcherHelpDoc_FromStruct(t *testing.T) {
	type DeployFlags struct {
		Environment string `long:"environment" short:"e" default:"staging" usage:"Target environment" choice:"dev" choice:"staging" choice:"prod"`
		Force       bool   `long:"force" short:"f" usage:"Force deploy"`
	}

	d := NewDispatcher("myapp")

	fs := NewFlagSet("deploy")
	flags := &DeployFlags{}
	if err := fs.FromStruct(flags); err != nil {
		t.Fatalf("FromStruct failed: %v", err)
	}

	d.Dispatch("deploy", NewCommand(fs, nil, WithUsage("Deploy the app")))

	doc := d.HelpDoc()
	cmd := doc.Commands[0]

	if len(cmd.Flags) != 2 {
		t.Fatalf("expected 2 flags, got %d", len(cmd.Flags))
	}

	// Flags sorted: environment, force
	envFlag := cmd.Flags[0]
	if envFlag.Name != "environment" {
		t.Errorf("expected %q, got %q", "environment", envFlag.Name)
	}
	if envFlag.Short != "e" {
		t.Errorf("expected short %q, got %q", "e", envFlag.Short)
	}
	if envFlag.Default != "staging" {
		t.Errorf("expected default %q, got %q", "staging", envFlag.Default)
	}
	if len(envFlag.Choices) != 3 {
		t.Fatalf("expected 3 choices, got %d", len(envFlag.Choices))
	}
}

func TestDispatcherHelpDoc_Examples(t *testing.T) {
	d := NewDispatcher("myapp")

	fs := NewFlagSet("deploy")
	fs.String("env", 'e', "staging", "Target environment")

	d.Dispatch("deploy", NewCommand(fs, nil,
		WithUsage("Deploy the application"),
		WithExamples(
			Example{Name: "Deploy to staging", Body: "myapp deploy --env staging"},
			Example{Name: "Deploy to production", Body: "myapp deploy --env prod"},
		),
	))

	doc := d.HelpDoc()

	if len(doc.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(doc.Commands))
	}

	cmd := doc.Commands[0]
	if len(cmd.Examples) != 2 {
		t.Fatalf("expected 2 examples, got %d", len(cmd.Examples))
	}

	if cmd.Examples[0].Name != "Deploy to staging" {
		t.Errorf("expected example name %q, got %q", "Deploy to staging", cmd.Examples[0].Name)
	}
	if cmd.Examples[0].Body != "myapp deploy --env staging" {
		t.Errorf("expected example body %q, got %q", "myapp deploy --env staging", cmd.Examples[0].Body)
	}
	if cmd.Examples[1].Name != "Deploy to production" {
		t.Errorf("expected example name %q, got %q", "Deploy to production", cmd.Examples[1].Name)
	}
	if cmd.Examples[1].Body != "myapp deploy --env prod" {
		t.Errorf("expected example body %q, got %q", "myapp deploy --env prod", cmd.Examples[1].Body)
	}
}

func TestDispatcherHelpDoc_NoExamples(t *testing.T) {
	d := NewDispatcher("myapp")

	fs := NewFlagSet("simple")
	d.Dispatch("simple", NewCommand(fs, nil, WithUsage("Simple command")))

	doc := d.HelpDoc()
	cmd := doc.Commands[0]

	if cmd.Examples != nil {
		t.Errorf("expected nil examples for command without examples, got %v", cmd.Examples)
	}

	// Verify it's omitted from JSON (omitempty)
	data, err := d.HelpJSON()
	if err != nil {
		t.Fatalf("HelpJSON failed: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	cmds := raw["commands"].([]interface{})
	cmdMap := cmds[0].(map[string]interface{})

	if _, exists := cmdMap["examples"]; exists {
		t.Error("expected examples to be omitted from JSON when empty")
	}
}

func TestDispatcherHelpDoc_ExamplesJSONRoundtrip(t *testing.T) {
	d := NewDispatcher("myapp")

	fs := NewFlagSet("run")
	d.Dispatch("run", NewCommand(fs, nil,
		WithUsage("Run a task"),
		WithExamples(
			Example{Name: "Run all tests", Body: "myapp run --all"},
		),
	))

	data, err := d.HelpJSON()
	if err != nil {
		t.Fatalf("HelpJSON failed: %v", err)
	}

	var roundtrip HelpDocument
	if err := json.Unmarshal(data, &roundtrip); err != nil {
		t.Fatalf("JSON roundtrip unmarshal failed: %v", err)
	}

	cmd := roundtrip.Commands[0]
	if len(cmd.Examples) != 1 {
		t.Fatalf("roundtrip: expected 1 example, got %d", len(cmd.Examples))
	}
	if cmd.Examples[0].Name != "Run all tests" {
		t.Errorf("roundtrip: expected example name %q, got %q", "Run all tests", cmd.Examples[0].Name)
	}
	if cmd.Examples[0].Body != "myapp run --all" {
		t.Errorf("roundtrip: expected example body %q, got %q", "myapp run --all", cmd.Examples[0].Body)
	}
}

// assertFlag is a test helper for common flag assertions.
func assertFlag(t *testing.T, f FlagDoc, name, short, typ, def, usage string, isBool bool) {
	t.Helper()
	if f.Name != name {
		t.Errorf("expected name %q, got %q", name, f.Name)
	}
	if f.Short != short {
		t.Errorf("flag %q: expected short %q, got %q", name, short, f.Short)
	}
	if f.Type != typ {
		t.Errorf("flag %q: expected type %q, got %q", name, typ, f.Type)
	}
	if f.Default != def {
		t.Errorf("flag %q: expected default %q, got %q", name, def, f.Default)
	}
	if f.Usage != usage {
		t.Errorf("flag %q: expected usage %q, got %q", name, usage, f.Usage)
	}
	if f.IsBool != isBool {
		t.Errorf("flag %q: expected isBool %v, got %v", name, isBool, f.IsBool)
	}
}
