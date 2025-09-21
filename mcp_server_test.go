package mflags

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCPServerInitialization(t *testing.T) {
	// Create a dispatcher with a test command
	d := NewDispatcher("testapp")

	fs := NewFlagSet("echo")
	cmd := NewSimpleCommand(fs, func(flags *FlagSet, args []string) error {
		fmt.Print("Hello from echo command")
		return nil
	})

	d.Dispatch("echo", cmd)

	// Create MCP server
	server := NewMCPServer(d)

	// Create input/output buffers
	input := bytes.NewBufferString("")
	output := bytes.NewBuffer(nil)

	server.SetInput(input)
	server.SetOutput(output)

	// Create initialize request
	initRequest := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  json.RawMessage(`{"protocolVersion": "2025-06-18", "capabilities": {}, "clientInfo": {"name": "test-client", "version": "1.0"}}`),
	}

	requestBytes, _ := json.Marshal(initRequest)
	input.WriteString(string(requestBytes) + "\n")

	// Send initialized notification
	initializedNotif := MCPRequest{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}
	notifBytes, _ := json.Marshal(initializedNotif)
	input.WriteString(string(notifBytes) + "\n")

	// Run server (will process requests and exit when input ends)
	err := server.Run()
	assert.NoError(t, err)

	// Parse initialize response
	outputStr := output.String()
	lines := strings.Split(outputStr, "\n")
	require.Greater(t, len(lines), 0)

	var response MCPResponse
	err = json.Unmarshal([]byte(lines[0]), &response)
	require.NoError(t, err)

	assert.Equal(t, "2.0", response.JSONRPC)
	assert.Equal(t, float64(1), response.ID)
	assert.Nil(t, response.Error)

	// Check initialize result
	var result InitializeResult
	resultBytes, _ := json.Marshal(response.Result)
	err = json.Unmarshal(resultBytes, &result)
	require.NoError(t, err)

	assert.Equal(t, MCPProtocolVersion, result.ProtocolVersion)
	assert.NotNil(t, result.Capabilities.Tools)
	assert.NotNil(t, result.Capabilities.Resources)
	assert.NotNil(t, result.Capabilities.Prompts)
	assert.Equal(t, "mflags-mcp-server", result.ServerInfo.Name)
	assert.Equal(t, "1.0.0", result.ServerInfo.Version)
}

func TestMCPServerToolsList(t *testing.T) {
	// Create a dispatcher with multiple commands
	d := NewDispatcher("testapp")

	// Add a command with flags
	fs1 := NewFlagSet("build")
	fs1.String("output", 'o', "a.out", "output file")
	fs1.Bool("verbose", 'v', false, "verbose output")
	cmd1 := NewSimpleCommandWithUsage(fs1,
		func(flags *FlagSet, args []string) error { return nil },
		"Build the project")
	d.Dispatch("build", cmd1)

	// Add a simple command
	fs2 := NewFlagSet("test")
	cmd2 := NewSimpleCommandWithUsage(fs2,
		func(flags *FlagSet, args []string) error { return nil },
		"Run tests")
	d.Dispatch("test", cmd2)

	// Create MCP server
	server := NewMCPServer(d)

	// Create input/output buffers
	input := bytes.NewBufferString("")
	output := bytes.NewBuffer(nil)

	server.SetInput(input)
	server.SetOutput(output)

	// Initialize first
	initRequest := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  json.RawMessage(`{"protocolVersion": "2025-06-18", "capabilities": {}, "clientInfo": {"name": "test", "version": "1.0"}}`),
	}
	requestBytes, _ := json.Marshal(initRequest)
	input.WriteString(string(requestBytes) + "\n")

	// Send tools/list request
	toolsListRequest := MCPRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}
	requestBytes, _ = json.Marshal(toolsListRequest)
	input.WriteString(string(requestBytes) + "\n")

	// Run server
	err := server.Run()
	assert.NoError(t, err)

	// Parse responses
	outputStr := output.String()
	lines := strings.Split(outputStr, "\n")
	require.GreaterOrEqual(t, len(lines), 2)

	// Skip initialize response and get tools/list response
	var toolsResponse MCPResponse
	err = json.Unmarshal([]byte(lines[1]), &toolsResponse)
	require.NoError(t, err)

	assert.Equal(t, "2.0", toolsResponse.JSONRPC)
	assert.Equal(t, float64(2), toolsResponse.ID)
	assert.Nil(t, toolsResponse.Error)

	// Check tools list result
	var result ToolsListResult
	resultBytes, _ := json.Marshal(toolsResponse.Result)
	err = json.Unmarshal(resultBytes, &result)
	require.NoError(t, err)

	assert.Len(t, result.Tools, 2)

	// Find each tool
	toolMap := make(map[string]Tool)
	for _, tool := range result.Tools {
		toolMap[tool.Name] = tool
	}

	buildTool, ok := toolMap["build"]
	assert.True(t, ok)
	assert.Equal(t, "Build the project", buildTool.Description)
	assert.NotNil(t, buildTool.InputSchema)
	assert.Contains(t, buildTool.InputSchema.Properties, "output")
	assert.Contains(t, buildTool.InputSchema.Properties, "verbose")
	// No positional or rest arguments, so no "arguments" property

	testTool, ok := toolMap["test"]
	assert.True(t, ok)
	assert.Equal(t, "Run tests", testTool.Description)
	assert.NotNil(t, testTool.InputSchema)
}

func TestMCPServerToolCall(t *testing.T) {
	// Create a dispatcher with a test command
	d := NewDispatcher("testapp")

	fs := NewFlagSet("echo")
	fs.String("prefix", 'p', "", "prefix for output")

	var capturedPrefix string
	var capturedArgs []string

	cmd := NewSimpleCommand(fs, func(flags *FlagSet, args []string) error {
		prefix := flags.Lookup("prefix")
		if prefix != nil {
			capturedPrefix = prefix.Value.String()
		}
		capturedArgs = args

		if capturedPrefix != "" {
			fmt.Printf("%s: ", capturedPrefix)
		}
		fmt.Printf("Hello %s", strings.Join(args, " "))
		return nil
	})

	d.Dispatch("echo", cmd)

	// Create MCP server
	server := NewMCPServer(d)

	// Create input/output buffers
	input := bytes.NewBufferString("")
	output := bytes.NewBuffer(nil)

	server.SetInput(input)
	server.SetOutput(output)

	// Initialize first
	initRequest := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  json.RawMessage(`{"protocolVersion": "2025-06-18", "capabilities": {}, "clientInfo": {"name": "test", "version": "1.0"}}`),
	}
	requestBytes, _ := json.Marshal(initRequest)
	input.WriteString(string(requestBytes) + "\n")

	// Send tools/call request
	toolCallRequest := MCPRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/call",
		Params:  json.RawMessage(`{"name": "echo", "arguments": {"prefix": "TEST", "arguments": ["world", "from", "MCP"]}}`),
	}
	requestBytes, _ = json.Marshal(toolCallRequest)
	input.WriteString(string(requestBytes) + "\n")

	// Run server
	err := server.Run()
	assert.NoError(t, err)

	// Parse responses
	outputStr := output.String()
	lines := strings.Split(outputStr, "\n")
	require.GreaterOrEqual(t, len(lines), 2)

	// Skip initialize response and get tools/call response
	var callResponse MCPResponse
	err = json.Unmarshal([]byte(lines[1]), &callResponse)
	require.NoError(t, err)

	assert.Equal(t, "2.0", callResponse.JSONRPC)
	assert.Equal(t, float64(2), callResponse.ID)
	assert.Nil(t, callResponse.Error)

	// Check tool call result
	var result ToolCallResult
	resultBytes, _ := json.Marshal(callResponse.Result)
	err = json.Unmarshal(resultBytes, &result)
	require.NoError(t, err)

	assert.False(t, result.IsError)
	assert.Len(t, result.Content, 1)
	assert.Equal(t, "text", result.Content[0].Type)
	assert.Equal(t, "TEST: Hello world from MCP", result.Content[0].Text)

	// Verify captured values
	assert.Equal(t, "TEST", capturedPrefix)
	assert.Equal(t, []string{"world", "from", "MCP"}, capturedArgs)
}

func TestMCPServerToolCallWithJSONOutput(t *testing.T) {
	// Create a dispatcher with a JSON-outputting command
	d := NewDispatcher("testapp")

	fs := NewFlagSet("status")

	cmd := NewSimpleCommandWithFormat(fs, func(flags *FlagSet, args []string) error {
		output := map[string]interface{}{
			"status":  "ok",
			"version": "1.0.0",
			"uptime":  3600,
		}
		data, _ := json.Marshal(output)
		fmt.Print(string(data))
		return nil
	}, OutputFormatJSON)

	d.Dispatch("status", cmd)

	// Create MCP server
	server := NewMCPServer(d)

	// Create input/output buffers
	input := bytes.NewBufferString("")
	output := bytes.NewBuffer(nil)

	server.SetInput(input)
	server.SetOutput(output)

	// Initialize first
	initRequest := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  json.RawMessage(`{"protocolVersion": "2025-06-18", "capabilities": {}, "clientInfo": {"name": "test", "version": "1.0"}}`),
	}
	requestBytes, _ := json.Marshal(initRequest)
	input.WriteString(string(requestBytes) + "\n")

	// Send tools/call request
	toolCallRequest := MCPRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/call",
		Params:  json.RawMessage(`{"name": "status"}`),
	}
	requestBytes, _ = json.Marshal(toolCallRequest)
	input.WriteString(string(requestBytes) + "\n")

	// Run server
	err := server.Run()
	assert.NoError(t, err)

	// Parse responses
	outputStr := output.String()
	lines := strings.Split(outputStr, "\n")
	require.GreaterOrEqual(t, len(lines), 2)

	// Skip initialize response and get tools/call response
	var callResponse MCPResponse
	err = json.Unmarshal([]byte(lines[1]), &callResponse)
	require.NoError(t, err)

	assert.Nil(t, callResponse.Error)

	// Check tool call result
	var result ToolCallResult
	resultBytes, _ := json.Marshal(callResponse.Result)
	err = json.Unmarshal(resultBytes, &result)
	require.NoError(t, err)

	assert.False(t, result.IsError)
	assert.Len(t, result.Content, 1)
	assert.Equal(t, "text", result.Content[0].Type)
	assert.Equal(t, "application/json", result.Content[0].MimeType)

	// Verify the JSON data is included
	var statusData map[string]interface{}
	err = json.Unmarshal(result.Content[0].Data, &statusData)
	require.NoError(t, err)
	assert.Equal(t, "ok", statusData["status"])
	assert.Equal(t, "1.0.0", statusData["version"])
	assert.Equal(t, float64(3600), statusData["uptime"])
}

func TestMCPServerErrorHandling(t *testing.T) {
	// Create a dispatcher with a failing command
	d := NewDispatcher("testapp")

	fs := NewFlagSet("fail")
	cmd := NewSimpleCommand(fs, func(flags *FlagSet, args []string) error {
		return fmt.Errorf("command failed with error")
	})

	d.Dispatch("fail", cmd)

	// Create MCP server
	server := NewMCPServer(d)

	// Create input/output buffers
	input := bytes.NewBufferString("")
	output := bytes.NewBuffer(nil)

	server.SetInput(input)
	server.SetOutput(output)

	// Initialize first
	initRequest := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  json.RawMessage(`{"protocolVersion": "2025-06-18", "capabilities": {}, "clientInfo": {"name": "test", "version": "1.0"}}`),
	}
	requestBytes, _ := json.Marshal(initRequest)
	input.WriteString(string(requestBytes) + "\n")

	// Send tools/call request
	toolCallRequest := MCPRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/call",
		Params:  json.RawMessage(`{"name": "fail"}`),
	}
	requestBytes, _ = json.Marshal(toolCallRequest)
	input.WriteString(string(requestBytes) + "\n")

	// Run server
	err := server.Run()
	assert.NoError(t, err)

	// Parse responses
	outputStr := output.String()
	lines := strings.Split(outputStr, "\n")
	require.GreaterOrEqual(t, len(lines), 2)

	// Skip initialize response and get tools/call response
	var callResponse MCPResponse
	err = json.Unmarshal([]byte(lines[1]), &callResponse)
	require.NoError(t, err)

	assert.Nil(t, callResponse.Error) // No JSON-RPC error

	// Check tool call result
	var result ToolCallResult
	resultBytes, _ := json.Marshal(callResponse.Result)
	err = json.Unmarshal(resultBytes, &result)
	require.NoError(t, err)

	assert.True(t, result.IsError)
	assert.Len(t, result.Content, 1)
	assert.Contains(t, result.Content[0].Text, "command failed with error")
}

func TestMCPServerResourcesAndPrompts(t *testing.T) {
	d := NewDispatcher("testapp")
	server := NewMCPServer(d)

	input := bytes.NewBufferString("")
	output := bytes.NewBuffer(nil)

	server.SetInput(input)
	server.SetOutput(output)

	// Initialize first
	initRequest := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  json.RawMessage(`{"protocolVersion": "2025-06-18", "capabilities": {}, "clientInfo": {"name": "test", "version": "1.0"}}`),
	}
	requestBytes, _ := json.Marshal(initRequest)
	input.WriteString(string(requestBytes) + "\n")

	// Test resources/list
	resourcesListRequest := MCPRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "resources/list",
	}
	requestBytes, _ = json.Marshal(resourcesListRequest)
	input.WriteString(string(requestBytes) + "\n")

	// Test prompts/list
	promptsListRequest := MCPRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "prompts/list",
	}
	requestBytes, _ = json.Marshal(promptsListRequest)
	input.WriteString(string(requestBytes) + "\n")

	// Run server
	err := server.Run()
	assert.NoError(t, err)

	// Parse responses
	outputStr := output.String()
	lines := strings.Split(outputStr, "\n")
	require.GreaterOrEqual(t, len(lines), 3)

	// Check resources/list response
	var resourcesResponse MCPResponse
	err = json.Unmarshal([]byte(lines[1]), &resourcesResponse)
	require.NoError(t, err)
	assert.Nil(t, resourcesResponse.Error)

	var resourcesResult ResourcesListResult
	resultBytes, _ := json.Marshal(resourcesResponse.Result)
	err = json.Unmarshal(resultBytes, &resourcesResult)
	require.NoError(t, err)
	assert.Empty(t, resourcesResult.Resources)

	// Check prompts/list response
	var promptsResponse MCPResponse
	err = json.Unmarshal([]byte(lines[2]), &promptsResponse)
	require.NoError(t, err)
	assert.Nil(t, promptsResponse.Error)

	var promptsResult PromptsListResult
	resultBytes, _ = json.Marshal(promptsResponse.Result)
	err = json.Unmarshal(resultBytes, &promptsResult)
	require.NoError(t, err)
	assert.Empty(t, promptsResult.Prompts)
}

func TestMCPServerInvalidRequests(t *testing.T) {
	d := NewDispatcher("testapp")
	server := NewMCPServer(d)

	tests := []struct {
		name          string
		request       string
		expectedError string
		expectedCode  int
	}{
		{
			name:          "invalid JSON",
			request:       `not valid json`,
			expectedError: "Parse error",
			expectedCode:  -32700,
		},
		{
			name: "wrong JSON-RPC version",
			request: func() string {
				req := MCPRequest{
					JSONRPC: "1.0",
					ID:      1,
					Method:  "initialize",
				}
				data, _ := json.Marshal(req)
				return string(data)
			}(),
			expectedError: "Invalid Request",
			expectedCode:  -32600,
		},
		{
			name: "unknown method",
			request: func() string {
				req := MCPRequest{
					JSONRPC: "2.0",
					ID:      1,
					Method:  "unknown",
				}
				data, _ := json.Marshal(req)
				return string(data)
			}(),
			expectedError: "Method not found",
			expectedCode:  -32601,
		},
		{
			name: "wrong protocol version",
			request: func() string {
				req := MCPRequest{
					JSONRPC: "2.0",
					ID:      1,
					Method:  "initialize",
					Params:  json.RawMessage(`{"protocolVersion": "1.0.0"}`),
				}
				data, _ := json.Marshal(req)
				return string(data)
			}(),
			expectedError: "Unsupported protocol version",
			expectedCode:  -32602,
		},
		{
			name: "tools/list before initialization",
			request: func() string {
				req := MCPRequest{
					JSONRPC: "2.0",
					ID:      1,
					Method:  "tools/list",
				}
				data, _ := json.Marshal(req)
				return string(data)
			}(),
			expectedError: "Server not initialized",
			expectedCode:  -32002,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := bytes.NewBufferString(test.request + "\n")
			output := bytes.NewBuffer(nil)

			server.SetInput(input)
			server.SetOutput(output)
			server.initialized = false // Reset for each test

			err := server.Run()
			assert.NoError(t, err)

			// Parse response
			var response MCPResponse
			outputStr := output.String()
			lines := strings.Split(outputStr, "\n")

			err = json.Unmarshal([]byte(lines[0]), &response)
			require.NoError(t, err)

			assert.NotNil(t, response.Error)
			assert.Equal(t, test.expectedCode, response.Error.Code)
			assert.Equal(t, test.expectedError, response.Error.Message)
		})
	}
}

func TestMCPServerCommand(t *testing.T) {
	d := NewDispatcher("testapp")

	// Add a test command
	d.Dispatch("test", NewSimpleCommand(NewFlagSet("test"),
		func(fs *FlagSet, args []string) error {
			fmt.Print("test output")
			return nil
		}))

	// Create MCP server command
	mcpCmd := NewMCPServerCommand(d)

	assert.NotNil(t, mcpCmd.FlagSet())
	assert.Equal(t, "Run as an MCP server for remote command execution", mcpCmd.Usage())
	assert.Equal(t, OutputFormatJSON, mcpCmd.OutputFormat())
}

func TestSimpleCommandWithFormat(t *testing.T) {
	fs := NewFlagSet("test")
	handler := func(fs *FlagSet, args []string) error { return nil }

	// Test NewSimpleCommandWithFormat
	cmd1 := NewSimpleCommandWithFormat(fs, handler, OutputFormatJSON)
	assert.Equal(t, OutputFormatJSON, cmd1.OutputFormat())
	assert.Equal(t, "", cmd1.Usage())

	// Test NewSimpleCommandFull
	cmd2 := NewSimpleCommandFull(fs, handler, "Test command", OutputFormatRaw)
	assert.Equal(t, OutputFormatRaw, cmd2.OutputFormat())
	assert.Equal(t, "Test command", cmd2.Usage())

	// Test default format
	cmd3 := NewSimpleCommand(fs, handler)
	assert.Equal(t, OutputFormatRaw, cmd3.OutputFormat())

	// Test SetOutputFormat
	cmd3.SetOutputFormat(OutputFormatJSON)
	assert.Equal(t, OutputFormatJSON, cmd3.OutputFormat())
}

func TestMCPServerPositionalArgsSchema(t *testing.T) {
	// Create a dispatcher with commands that have positional arguments
	d := NewDispatcher("testapp")

	// Command with positional arguments via struct
	type CopyCmd struct {
		Verbose bool   `long:"verbose" short:"v"`
		Source  string `position:"0"`
		Dest    string `position:"1"`
	}

	copyConfig := &CopyCmd{}
	fs1 := NewFlagSet("copy")
	err := fs1.FromStruct(copyConfig)
	assert.NoError(t, err)

	cmd1 := NewSimpleCommandWithUsage(fs1,
		func(flags *FlagSet, args []string) error { return nil },
		"Copy files from source to destination")
	d.Dispatch("copy", cmd1)

	// Command with rest arguments via struct
	type CatCmd struct {
		Number bool     `long:"number" short:"n"`
		Files  []string `rest:"true"`
	}

	catConfig := &CatCmd{}
	fs2 := NewFlagSet("cat")
	err = fs2.FromStruct(catConfig)
	assert.NoError(t, err)

	cmd2 := NewSimpleCommandWithUsage(fs2,
		func(flags *FlagSet, args []string) error { return nil },
		"Concatenate files")
	d.Dispatch("cat", cmd2)

	// Command with both positional and rest arguments
	type ExecCmd struct {
		Dir     string   `long:"dir" short:"d"`
		Command string   `position:"0"`
		Args    []string `rest:"true"`
	}

	execConfig := &ExecCmd{}
	fs3 := NewFlagSet("exec")
	err = fs3.FromStruct(execConfig)
	assert.NoError(t, err)

	cmd3 := NewSimpleCommandWithUsage(fs3,
		func(flags *FlagSet, args []string) error { return nil },
		"Execute a command")
	d.Dispatch("exec", cmd3)

	// Create MCP server
	server := NewMCPServer(d)

	// Create input/output buffers
	input := bytes.NewBufferString("")
	output := bytes.NewBuffer(nil)

	server.SetInput(input)
	server.SetOutput(output)

	// Initialize first
	initRequest := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  json.RawMessage(`{"protocolVersion": "2025-06-18", "capabilities": {}, "clientInfo": {"name": "test", "version": "1.0"}}`),
	}
	requestBytes, _ := json.Marshal(initRequest)
	input.WriteString(string(requestBytes) + "\n")

	// Send tools/list request
	toolsListRequest := MCPRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}
	requestBytes, _ = json.Marshal(toolsListRequest)
	input.WriteString(string(requestBytes) + "\n")

	// Run server
	err = server.Run()
	assert.NoError(t, err)

	// Parse responses
	outputStr := output.String()
	lines := strings.Split(outputStr, "\n")
	require.GreaterOrEqual(t, len(lines), 2)

	// Skip initialize response and get tools/list response
	var toolsResponse MCPResponse
	err = json.Unmarshal([]byte(lines[1]), &toolsResponse)
	require.NoError(t, err)

	// Check tools list result
	var result ToolsListResult
	resultBytes, _ := json.Marshal(toolsResponse.Result)
	err = json.Unmarshal(resultBytes, &result)
	require.NoError(t, err)

	// Find each tool and verify schema
	toolMap := make(map[string]Tool)
	for _, tool := range result.Tools {
		toolMap[tool.Name] = tool
	}

	// Check copy command schema (2 positional args, no rest)
	copyTool, ok := toolMap["copy"]
	assert.True(t, ok)
	assert.NotNil(t, copyTool.InputSchema)
	// Should have named positional parameters
	assert.Contains(t, copyTool.InputSchema.Properties, "source")
	assert.Contains(t, copyTool.InputSchema.Properties, "dest")
	assert.Contains(t, copyTool.InputSchema.Required, "source")
	assert.Contains(t, copyTool.InputSchema.Required, "dest")
	// Should not have arguments array since there's no rest field
	assert.NotContains(t, copyTool.InputSchema.Properties, "arguments")

	// Check cat command schema (rest args only)
	catTool, ok := toolMap["cat"]
	assert.True(t, ok)
	assert.NotNil(t, catTool.InputSchema)
	assert.Contains(t, catTool.InputSchema.Properties, "arguments")
	// No required positional args
	catArgsProp := catTool.InputSchema.Properties["arguments"]
	assert.Equal(t, "array", catArgsProp.Type)
	assert.Equal(t, "Additional command arguments", catArgsProp.Description)

	// Check exec command schema (1 positional + rest)
	execTool, ok := toolMap["exec"]
	assert.True(t, ok)
	assert.NotNil(t, execTool.InputSchema)
	// Should have named positional parameter
	assert.Contains(t, execTool.InputSchema.Properties, "command")
	assert.Contains(t, execTool.InputSchema.Required, "command")
	// Should also have arguments array for rest args
	assert.Contains(t, execTool.InputSchema.Properties, "arguments")
	execArgsProp := execTool.InputSchema.Properties["arguments"]
	assert.Equal(t, "array", execArgsProp.Type)
	assert.Equal(t, "Additional command arguments", execArgsProp.Description)
}
