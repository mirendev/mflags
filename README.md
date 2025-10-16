# mflags

A powerful, feature-rich command-line flag parsing library for Go that extends the standard library's `flag` package with modern CLI capabilities.

## Features

mflags works like Go's standard `flag` package but adds significantly more functionality:

- **Short and long flags** - Support for both `-v` and `--verbose` style flags
- **Positional arguments** - Define and parse positional arguments by position
- **Rest arguments** - Capture all remaining arguments in a slice
- **Unknown flag handling** - Optionally accumulate unknown flags instead of erroring
- **Struct-based parsing** - Define flags using struct tags (similar to JSON marshaling)
- **Command dispatcher** - Build multi-level command hierarchies (like git, kubectl)
- **Interspersed flags** - Allow flags anywhere in the command sequence
- **MCP server mode** - Expose commands as Model Context Protocol tools
- **Shell completion** - Generate bash and zsh completion scripts
- **Embedded structs** - Compose flag definitions from embedded structs

## Installation

```bash
go get miren.dev/mflags
```

## Quick Start

### Basic Flag Parsing

```go
package main

import (
    "fmt"
    "miren.dev/mflags"
)

func main() {
    fs := mflags.NewFlagSet("myapp")

    verbose := fs.Bool("verbose", 'v', false, "enable verbose output")
    output := fs.String("output", 'o', "stdout", "output file")
    count := fs.Int("count", 'c', 1, "number of iterations")

    if err := fs.Parse(os.Args[1:]); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("Verbose: %v, Output: %s, Count: %d\n", *verbose, *output, *count)
    fmt.Printf("Remaining args: %v\n", fs.Args())
}
```

```bash
$ myapp -v --output=file.txt -c 5 arg1 arg2
Verbose: true, Output: file.txt, Count: 5
Remaining args: [arg1 arg2]
```

### Struct-Based Parsing

Define flags using struct tags for a more declarative approach:

```go
type Config struct {
    Verbose  bool          `long:"verbose" short:"v" usage:"Enable verbose output"`
    Output   string        `long:"output" short:"o" default:"stdout" usage:"Output file"`
    Count    int           `long:"count" short:"c" default:"1" usage:"Number of iterations"`
    Timeout  time.Duration `long:"timeout" short:"t" default:"30s" usage:"Request timeout"`
    Tags     []string      `long:"tags" usage:"Comma-separated tags"`
}

func main() {
    config := &Config{}
    if err := mflags.ParseStruct(config, os.Args[1:]); err != nil {
        log.Fatal(err)
    }

    fmt.Printf("%+v\n", config)
}
```

## Advanced Features

### Positional Arguments

Define required positional arguments by position:

```go
type DeployConfig struct {
    Environment string `position:"0" usage:"Deployment environment (dev, staging, prod)"`
    Version     string `position:"1" usage:"Version to deploy"`
    Verbose     bool   `long:"verbose" short:"v" usage:"Verbose output"`
}

config := &DeployConfig{}
mflags.ParseStruct(config, []string{"production", "v1.2.3", "--verbose"})
// Environment: "production", Version: "v1.2.3", Verbose: true
```

### Rest Arguments

Capture all remaining arguments:

```go
type RunConfig struct {
    Verbose bool     `long:"verbose" short:"v" usage:"Verbose output"`
    Files   []string `rest:"true" usage:"Files to process"`
}

config := &RunConfig{}
mflags.ParseStruct(config, []string{"-v", "file1.txt", "file2.txt", "file3.txt"})
// Verbose: true, Files: ["file1.txt", "file2.txt", "file3.txt"]
```

### Unknown Flag Handling

Accumulate unknown flags for pass-through to other commands:

```go
type ProxyConfig struct {
    Port         int      `long:"port" short:"p" default:"8080" usage:"Proxy port"`
    Target       string   `long:"target" short:"t" usage:"Target URL"`
    UnknownFlags []string `unknown:"true" usage:"Flags to pass to target"`
}

config := &ProxyConfig{}
mflags.ParseStruct(config, []string{
    "--port", "9000",
    "--target", "http://localhost:3000",
    "--unknown-flag", "value",
    "-x", "arg",
})
// Port: 9000, Target: "http://localhost:3000"
// UnknownFlags: ["--unknown-flag", "value", "-x", "arg"]
```

## Command Dispatcher

Build sophisticated multi-level command hierarchies like `git`, `docker`, or `kubectl`:

```go
type DeployFlags struct {
    Environment string `position:"0" usage:"Environment to deploy to"`
    DryRun      bool   `long:"dry-run" usage:"Perform a dry run"`
    Verbose     bool   `long:"verbose" short:"v" usage:"Verbose output"`
}

func main() {
    dispatcher := mflags.NewDispatcher("myapp")

    // Add a simple command
    versionFS := mflags.NewFlagSet("version")
    dispatcher.Dispatch("version", mflags.NewCommand(
        versionFS,
        func(fs *mflags.FlagSet, args []string) error {
            fmt.Println("myapp v1.0.0")
            return nil
        },
        mflags.WithUsage("Show version information"),
    ))

    // Add a command with struct-based flags
    deployFS := mflags.NewFlagSet("deploy")
    deployFlags := &DeployFlags{}
    deployFS.FromStruct(deployFlags)

    dispatcher.Dispatch("deploy", mflags.NewCommand(
        deployFS,
        func(fs *mflags.FlagSet, args []string) error {
            fmt.Printf("Deploying to %s (dry-run: %v)\n",
                deployFlags.Environment, deployFlags.DryRun)
            return nil
        },
        mflags.WithUsage("Deploy the application"),
    ))

    // Add nested commands
    configGetFS := mflags.NewFlagSet("config get")
    dispatcher.Dispatch("config get", mflags.NewCommand(
        configGetFS,
        func(fs *mflags.FlagSet, args []string) error {
            if len(args) == 0 {
                return fmt.Errorf("key required")
            }
            fmt.Printf("database.host=localhost\n")
            return nil
        },
        mflags.WithUsage("Get configuration value"),
    ))

    if err := dispatcher.Run(os.Args[1:]); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

```bash
$ myapp version
myapp v1.0.0

$ myapp deploy production --dry-run
Deploying to production (dry-run: true)

$ myapp config get database.host
database.host=localhost
```

### Interspersed Flags

The dispatcher automatically supports flags at any position in the command sequence:

```go
type ServerFlags struct {
    Port    int    `long:"port" short:"p" default:"8080"`
    Host    string `long:"host" default:"localhost"`
    Debug   bool   `long:"debug" short:"d"`
}

dispatcher := mflags.NewDispatcher("myapp")

serverFS := mflags.NewFlagSet("server start")
serverFlags := &ServerFlags{}
serverFS.FromStruct(serverFlags)

dispatcher.Dispatch("server start", mflags.NewCommand(
    serverFS,
    func(fs *mflags.FlagSet, args []string) error {
        fmt.Printf("Starting server on %s:%d (debug: %v)\n",
            serverFlags.Host, serverFlags.Port, serverFlags.Debug)
        return nil
    },
    mflags.WithUsage("Start the server"),
))
```

All of these work:

```bash
$ myapp --debug server --port 9000 start
$ myapp server --port 9000 start --debug
$ myapp server start --port 9000 --debug
```

## MCP Server Mode

Expose your CLI commands as Model Context Protocol (MCP) tools for AI assistants:

```go
type GreetFlags struct {
    Name   string `long:"name" short:"n" default:"World" usage:"Name to greet"`
    Formal bool   `long:"formal" short:"f" usage:"Use formal greeting"`
}

func main() {
    dispatcher := mflags.NewDispatcher("myapp")

    greetFS := mflags.NewFlagSet("greet")
    greetFlags := &GreetFlags{}
    greetFS.FromStruct(greetFlags)

    dispatcher.Dispatch("greet", mflags.NewCommand(
        greetFS,
        func(fs *mflags.FlagSet, args []string) error {
            if greetFlags.Formal {
                fmt.Printf("Good day, %s!\n", greetFlags.Name)
            } else {
                fmt.Printf("Hello, %s!\n", greetFlags.Name)
            }
            return nil
        },
        mflags.WithUsage("Greet someone"),
    ))

    // Create and run MCP server
    server := mflags.NewMCPServer(dispatcher)
    if err := server.Run(); err != nil {
        log.Fatal(err)
    }
}
```

When your application is configured as an MCP server in Claude Desktop, it automatically exposes all commands as tools that Claude can call with structured arguments.

**Configuration in Claude Desktop (`claude_desktop_config.json`):**

```json
{
  "mcpServers": {
    "myapp": {
      "command": "/path/to/myapp"
    }
  }
}
```

Claude can then invoke:
```
Use myapp to greet Alice formally
```

And the MCP server translates this to executing:
```bash
$ myapp greet --name Alice --formal
```

### MCP Features

- **Automatic tool generation** - All commands become MCP tools
- **Type-safe arguments** - Struct tags define parameter types and validation
- **Positional arguments** - Mapped to required MCP parameters
- **Optional flags** - Mapped to optional MCP parameters
- **Help integration** - Usage strings become tool descriptions

## Shell Completion

Generate completion scripts for bash and zsh:

```go
fs := mflags.NewFlagSet("myapp")
fs.Bool("verbose", 'v', false, "verbose output")
fs.String("output", 'o', "stdout", "output file")

// Generate bash completion
bashScript := fs.GenerateBashCompletion("myapp")
fmt.Println(bashScript)

// Generate zsh completion
zshScript := fs.GenerateZshCompletion("myapp")
fmt.Println(zshScript)
```

Or use the dispatcher:

```go
dispatcher := mflags.NewDispatcher("myapp")
// ... add commands ...

bashFS := mflags.NewFlagSet("completion bash")
dispatcher.Dispatch("completion bash", mflags.NewCommand(
    bashFS,
    func(fs *mflags.FlagSet, args []string) error {
        fmt.Println(dispatcher.GenerateBashCompletion())
        return nil
    },
    mflags.WithUsage("Generate bash completion script"),
))

zshFS := mflags.NewFlagSet("completion zsh")
dispatcher.Dispatch("completion zsh", mflags.NewCommand(
    zshFS,
    func(fs *mflags.FlagSet, args []string) error {
        fmt.Println(dispatcher.GenerateZshCompletion())
        return nil
    },
    mflags.WithUsage("Generate zsh completion script"),
))
```

Install completions:

```bash
# Bash
myapp completion bash > /etc/bash_completion.d/myapp

# Zsh
myapp completion zsh > /usr/local/share/zsh/site-functions/_myapp
```

## Supported Types

### Flag Types

- `bool` - Boolean flags
- `string` - String values
- `int` - Integer values
- `[]string` - Comma-separated string arrays
- `time.Duration` - Duration values (parsed by `time.ParseDuration`)

### Struct Tags

| Tag | Description | Example |
|-----|-------------|---------|
| `long` | Long flag name | `long:"verbose"` |
| `short` | Short flag name (single char) | `short:"v"` |
| `default` | Default value | `default:"true"` |
| `usage` | Help text | `usage:"Enable verbose mode"` |
| `position` | Positional argument index | `position:"0"` |
| `rest` | Capture remaining args | `rest:"true"` |
| `unknown` | Capture unknown flags | `unknown:"true"` |

## Embedded Structs

Compose flag definitions from multiple structs:

```go
type CommonFlags struct {
    Verbose bool `long:"verbose" short:"v" usage:"Verbose output"`
    DryRun  bool `long:"dry-run" usage:"Perform dry run"`
}

type DatabaseFlags struct {
    Host string `long:"db-host" default:"localhost" usage:"Database host"`
    Port int    `long:"db-port" default:"5432" usage:"Database port"`
}

type DeployConfig struct {
    CommonFlags
    DatabaseFlags
    Environment string `position:"0" usage:"Deployment environment"`
}
```

## Error Handling

mflags provides descriptive errors:

```go
var (
    ErrUnknownFlag  = errors.New("unknown flag")
    ErrMissingValue = errors.New("flag needs an argument")
    ErrInvalidValue = errors.New("invalid flag value")
    ErrHelp         = errors.New("help requested")
)
```

## Comparison with Standard Library

| Feature | `flag` | `mflags` |
|---------|--------|----------|
| Long flags | ✓ | ✓ |
| Short flags | ✗ | ✓ |
| Combined short flags (`-abc`) | ✗ | ✓ |
| Struct-based parsing | ✗ | ✓ |
| Positional arguments | ✗ | ✓ |
| Rest arguments | ✗ | ✓ |
| Unknown flag handling | ✗ | ✓ |
| Command dispatcher | ✗ | ✓ |
| Interspersed flags | ✗ | ✓ |
| MCP server support | ✗ | ✓ |
| Shell completion | ✗ | ✓ |
| Embedded structs | ✗ | ✓ |

## Examples

### Real-World CLI Application

```go
package main

import (
    "fmt"
    "log"
    "os"
    "miren.dev/mflags"
)

type ServerConfig struct {
    Verbose bool   `long:"verbose" short:"v" usage:"Enable verbose logging"`
    Port    int    `long:"port" short:"p" default:"8080" usage:"Server port"`
    Host    string `long:"host" default:"localhost" usage:"Server host"`
}

type DeployConfig struct {
    Verbose     bool     `long:"verbose" short:"v" usage:"Enable verbose logging"`
    Environment string   `position:"0" usage:"Target environment"`
    Version     string   `position:"1" usage:"Version to deploy"`
    Tags        []string `long:"tags" short:"t" usage:"Deployment tags"`
    DryRun      bool     `long:"dry-run" usage:"Simulate deployment"`
}

func main() {
    dispatcher := mflags.NewDispatcher("myapp")

    // Server start command
    serverFS := mflags.NewFlagSet("server start")
    serverConfig := &ServerConfig{}
    serverFS.FromStruct(serverConfig)

    dispatcher.Dispatch("server start", mflags.NewCommand(
        serverFS,
        func(fs *mflags.FlagSet, args []string) error {
            fmt.Printf("Starting server on %s:%d (verbose: %v)\n",
                serverConfig.Host, serverConfig.Port, serverConfig.Verbose)
            // ... server implementation ...
            return nil
        },
        mflags.WithUsage("Start the application server"),
    ))

    // Deploy command
    deployFS := mflags.NewFlagSet("deploy")
    deployConfig := &DeployConfig{}
    deployFS.FromStruct(deployConfig)

    dispatcher.Dispatch("deploy", mflags.NewCommand(
        deployFS,
        func(fs *mflags.FlagSet, args []string) error {
            fmt.Printf("Deploying %s to %s (dry-run: %v)\n",
                deployConfig.Version, deployConfig.Environment, deployConfig.DryRun)
            if len(deployConfig.Tags) > 0 {
                fmt.Printf("Tags: %v\n", deployConfig.Tags)
            }
            // ... deployment implementation ...
            return nil
        },
        mflags.WithUsage("Deploy application to environment"),
    ))

    // Version command
    versionFS := mflags.NewFlagSet("version")
    dispatcher.Dispatch("version", mflags.NewCommand(
        versionFS,
        func(fs *mflags.FlagSet, args []string) error {
            fmt.Println("myapp version 1.0.0")
            return nil
        },
        mflags.WithUsage("Show version information"),
    ))

    if err := dispatcher.Run(os.Args[1:]); err != nil {
        log.Fatal(err)
    }
}
```

Usage:

```bash
$ myapp server start --port 9000
$ myapp server start -p 9000 -v
$ myapp deploy production v1.2.3 --tags prod,release
$ myapp deploy staging v1.2.3-rc1 --dry-run
$ myapp version
```

## License

[Your License Here]

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
