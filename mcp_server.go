package mflags

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"
)

// MCP Protocol Version
const MCPProtocolVersion = "2025-06-18"

// MCPRequest represents a JSON-RPC request in the MCP protocol
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

// MCPResponse represents a JSON-RPC response in the MCP protocol
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      any         `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError represents a JSON-RPC error
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// InitializeRequest represents the initialize request parameters
type InitializeRequest struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      Implementation     `json:"clientInfo"`
}

// InitializeResult represents the initialize response
type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      Implementation     `json:"serverInfo"`
	Instructions    string             `json:"instructions,omitempty"`
}

// Implementation represents client or server information
type Implementation struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ClientCapabilities represents client capabilities
type ClientCapabilities struct {
	Experimental map[string]interface{} `json:"experimental,omitempty"`
	Sampling     *SamplingCapability    `json:"sampling,omitempty"`
	Roots        *RootsCapability       `json:"roots,omitempty"`
}

// ServerCapabilities represents server capabilities
type ServerCapabilities struct {
	Experimental map[string]interface{} `json:"experimental,omitempty"`
	Logging      *LoggingCapability     `json:"logging,omitempty"`
	Prompts      *PromptsCapability     `json:"prompts,omitempty"`
	Resources    *ResourcesCapability   `json:"resources,omitempty"`
	Tools        *ToolsCapability       `json:"tools,omitempty"`
}

// Capability types
type LoggingCapability struct{}
type SamplingCapability struct{}
type RootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}
type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}
type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// Tool represents an MCP tool
type Tool struct {
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	InputSchema *InputSchema `json:"inputSchema"`
}

// InputSchema represents the JSON schema for tool input
type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
}

// Property represents a JSON schema property
type Property struct {
	Type        string      `json:"type"`
	Description string      `json:"description,omitempty"`
	Items       *Property   `json:"items,omitempty"`
	Default     interface{} `json:"default,omitempty"`
}

// ToolsListRequest represents the tools/list request parameters
type ToolsListRequest struct {
	Cursor string `json:"cursor,omitempty"`
}

// ToolsListResult represents the tools/list response
type ToolsListResult struct {
	Tools      []Tool `json:"tools"`
	NextCursor string `json:"nextCursor,omitempty"`
}

// ToolCallRequest represents the tools/call request parameters
type ToolCallRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// ToolCallResult represents the tools/call response
type ToolCallResult struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

// Content represents tool output content
type Content struct {
	Type     string          `json:"type"`
	Text     string          `json:"text,omitempty"`
	Data     json.RawMessage `json:"data,omitempty"`
	MimeType string          `json:"mimeType,omitempty"`
}

// ResourcesListResult represents the resources/list response
type ResourcesListResult struct {
	Resources  []Resource `json:"resources"`
	NextCursor string     `json:"nextCursor,omitempty"`
}

// Resource represents an MCP resource
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// PromptsListResult represents the prompts/list response
type PromptsListResult struct {
	Prompts    []Prompt `json:"prompts"`
	NextCursor string   `json:"nextCursor,omitempty"`
}

// Prompt represents an MCP prompt
type Prompt struct {
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Arguments   []Argument `json:"arguments,omitempty"`
}

// Argument represents a prompt argument
type Argument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// MCPServer handles MCP protocol communication
type MCPServer struct {
	dispatcher  *Dispatcher
	input       io.Reader
	output      io.Writer
	errorOutput io.Writer
	mu          sync.Mutex
	initialized bool
	serverInfo  Implementation
}

// NewMCPServer creates a new MCP server
func NewMCPServer(dispatcher *Dispatcher) *MCPServer {
	return &MCPServer{
		dispatcher:  dispatcher,
		input:       os.Stdin,
		output:      os.Stdout,
		errorOutput: os.Stderr,
		serverInfo: Implementation{
			Name:    "mflags-mcp-server",
			Version: "1.0.0",
		},
	}
}

// SetInput sets the input reader
func (s *MCPServer) SetInput(r io.Reader) {
	s.input = r
}

// SetOutput sets the output writer
func (s *MCPServer) SetOutput(w io.Writer) {
	s.output = w
}

// SetErrorOutput sets the error output writer
func (s *MCPServer) SetErrorOutput(w io.Writer) {
	s.errorOutput = w
}

// Run starts the MCP server and processes requests
func (s *MCPServer) Run() error {
	scanner := bufio.NewScanner(s.input)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Parse JSON-RPC request
		var request MCPRequest
		if err := json.Unmarshal([]byte(line), &request); err != nil {
			s.sendErrorResponse(nil, -32700, "Parse error", err.Error())
			continue
		}

		// Handle the request
		s.handleRequest(request)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	return nil
}

// handleRequest processes a single MCP request
func (s *MCPServer) handleRequest(request MCPRequest) {
	// Validate JSON-RPC version
	if request.JSONRPC != "2.0" {
		s.sendErrorResponse(request.ID, -32600, "Invalid Request", "JSON-RPC version must be 2.0")
		return
	}

	// Handle different methods
	switch request.Method {
	case "initialize":
		s.handleInitialize(request)
	case "notifications/initialized":
		s.handleInitialized(request)
	case "tools/list":
		s.handleToolsList(request)
	case "tools/call":
		s.handleToolCall(request)
	case "resources/list":
		s.handleResourcesList(request)
	case "resources/read":
		s.handleResourceRead(request)
	case "prompts/list":
		s.handlePromptsList(request)
	case "prompts/get":
		s.handlePromptGet(request)
	default:
		s.sendErrorResponse(request.ID, -32601, "Method not found", fmt.Sprintf("Unknown method: %s", request.Method))
	}
}

// handleInitialize handles the initialize request
func (s *MCPServer) handleInitialize(request MCPRequest) {
	var params InitializeRequest
	if request.Params != nil {
		if err := json.Unmarshal(request.Params, &params); err != nil {
			s.sendErrorResponse(request.ID, -32602, "Invalid params", err.Error())
			return
		}
	}

	// Check protocol version compatibility
	if params.ProtocolVersion != MCPProtocolVersion {
		// For now, we only support one version
		s.sendErrorResponse(request.ID, -32602, "Unsupported protocol version",
			map[string]string{
				"supported": MCPProtocolVersion,
				"requested": params.ProtocolVersion,
			})
		return
	}

	// Build server capabilities
	capabilities := ServerCapabilities{
		Tools: &ToolsCapability{
			ListChanged: false,
		},
		// We support empty resources and prompts
		Resources: &ResourcesCapability{
			Subscribe:   false,
			ListChanged: false,
		},
		Prompts: &PromptsCapability{
			ListChanged: false,
		},
	}

	result := InitializeResult{
		ProtocolVersion: MCPProtocolVersion,
		Capabilities:    capabilities,
		ServerInfo:      s.serverInfo,
		Instructions:    "This MCP server exposes command-line tools from the mflags dispatcher.",
	}

	s.sendResponse(request.ID, result)
	s.initialized = true
}

// handleInitialized handles the initialized notification
func (s *MCPServer) handleInitialized(request MCPRequest) {
	// This is a notification, no response needed
	// Just mark that we're ready for normal operations
}

// handleToolsList handles the tools/list request
func (s *MCPServer) handleToolsList(request MCPRequest) {
	if !s.initialized {
		s.sendErrorResponse(request.ID, -32002, "Server not initialized", nil)
		return
	}

	// Convert dispatcher commands to MCP tools
	var tools []Tool
	commands := s.dispatcher.GetCommands()

	for name, cmd := range commands {
		tool := Tool{
			Name:        name,
			Description: cmd.Usage(),
			InputSchema: s.buildToolSchema(cmd),
		}
		tools = append(tools, tool)
	}

	result := ToolsListResult{
		Tools: tools,
	}

	s.sendResponse(request.ID, result)
}

// buildToolSchema builds a JSON schema from a command's FlagSet
func (s *MCPServer) buildToolSchema(cmd Command) *InputSchema {
	schema := &InputSchema{
		Type:       "object",
		Properties: make(map[string]Property),
		Required:   []string{},
	}

	fs := cmd.FlagSet()
	if fs == nil {
		return schema
	}

	// Add properties for each flag
	fs.VisitAll(func(flag *Flag) {
		prop := Property{
			Type:        s.getJSONType(flag.Value),
			Description: flag.Usage,
		}

		// Set default value if available
		if flag.DefValue != "" && flag.DefValue != "false" && flag.DefValue != "0" && flag.DefValue != "[]" {
			prop.Default = flag.DefValue
		}

		// Use the long name if available, otherwise use string of short flag
		propName := flag.Name
		if propName == "" && flag.Short != 0 {
			propName = string(flag.Short)
		}

		if propName != "" {
			schema.Properties[propName] = prop
		}
	})

	// Add positional arguments as named parameters
	positionalFields := fs.GetPositionalFields()
	for _, field := range positionalFields {
		// Convert field name to lowercase for consistency
		paramName := strings.ToLower(field.Name)

		// Determine JSON type based on field type
		jsonType := s.getTypeForReflectType(field.Type)

		prop := Property{
			Type:        jsonType,
			Description: fmt.Sprintf("Positional argument %s", field.Name),
		}

		schema.Properties[paramName] = prop
		// Positional arguments are required
		schema.Required = append(schema.Required, paramName)
	}

	// Check if there are rest arguments
	if fs.HasRestArgs() {
		// Add rest arguments as an array property
		schema.Properties["arguments"] = Property{
			Type:        "array",
			Description: "Additional command arguments",
			Items: &Property{
				Type: "string",
			},
		}
	}

	return schema
}

// getJSONType returns the JSON schema type for a flag value
func (s *MCPServer) getJSONType(v Value) string {
	if v == nil {
		return "string"
	}

	// Check the underlying type
	switch v.(type) {
	case *boolValue:
		return "boolean"
	case *intValue:
		return "integer"
	case *durationValue:
		return "string" // Duration is represented as string
	case *stringArrayValue:
		return "array"
	default:
		// For custom types, try to infer from the value
		val := reflect.ValueOf(v)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		switch val.Kind() {
		case reflect.Bool:
			return "boolean"
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return "integer"
		case reflect.Float32, reflect.Float64:
			return "number"
		case reflect.Slice, reflect.Array:
			return "array"
		default:
			return "string"
		}
	}
}

// getTypeForReflectType returns the JSON schema type for a reflect.Type
func (s *MCPServer) getTypeForReflectType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Check if it's a time.Duration
		if t == reflect.TypeOf(time.Duration(0)) {
			return "string"
		}
		return "integer"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.String:
		return "string"
	case reflect.Slice:
		return "array"
	default:
		return "string"
	}
}

// handleToolCall handles the tools/call request
func (s *MCPServer) handleToolCall(request MCPRequest) {
	if !s.initialized {
		s.sendErrorResponse(request.ID, -32002, "Server not initialized", nil)
		return
	}

	var params ToolCallRequest
	if err := json.Unmarshal(request.Params, &params); err != nil {
		s.sendErrorResponse(request.ID, -32602, "Invalid params", err.Error())
		return
	}

	// Check if the command exists
	cmd := s.dispatcher.GetCommand(params.Name)
	if cmd == nil {
		s.sendErrorResponse(request.ID, -32602, "Tool not found",
			fmt.Sprintf("No tool named '%s'", params.Name))
		return
	}

	// Build command arguments from the tool call parameters
	var args []string

	// Process the arguments map to build command-line flags
	if params.Arguments != nil {
		fs := cmd.FlagSet()
		if fs != nil {
			// Get positional field names
			positionalFields := fs.GetPositionalFields()
			positionalNames := make(map[string]bool)
			for _, field := range positionalFields {
				positionalNames[strings.ToLower(field.Name)] = true
			}

			// Separate positional arguments from flags
			var positionalArgs []string
			flagArgs := make(map[string]interface{})

			for key, value := range params.Arguments {
				if key == "arguments" {
					// These are rest arguments, handle separately
					continue
				}

				// Check if this is a positional argument
				if positionalNames[key] {
					// Store positional arguments in order
					positionalArgs = append(positionalArgs, fmt.Sprintf("%v", value))
				} else {
					// Store as flag
					flagArgs[key] = value
				}
			}

			// Add flags first
			for key, value := range flagArgs {
				// Convert the argument to command-line flag format
				if boolVal, ok := value.(bool); ok && boolVal {
					// For boolean flags that are true, just add the flag
					if len(key) == 1 {
						args = append(args, "-"+key)
					} else {
						args = append(args, "--"+key)
					}
				} else if !ok || !boolVal {
					// For non-boolean or false boolean, add flag with value
					if boolVal, ok := value.(bool); ok && !boolVal {
						// Skip false boolean flags (they're off by default)
						continue
					}

					// Add the flag and its value
					if len(key) == 1 {
						args = append(args, "-"+key)
					} else {
						args = append(args, "--"+key)
					}
					args = append(args, fmt.Sprintf("%v", value))
				}
			}

			// Add positional arguments in the correct order
			// We need to order them based on their position indices
			if len(positionalFields) > 0 {
				orderedPositional := make([]string, len(positionalFields))
				for i, field := range positionalFields {
					paramName := strings.ToLower(field.Name)
					if val, ok := params.Arguments[paramName]; ok {
						orderedPositional[i] = fmt.Sprintf("%v", val)
					}
				}
				args = append(args, orderedPositional...)
			}

			// Add rest arguments at the end
			if posArgs, ok := params.Arguments["arguments"].([]interface{}); ok {
				for _, arg := range posArgs {
					args = append(args, fmt.Sprintf("%v", arg))
				}
			}
		}
	}

	// Capture output by replacing stdout temporarily
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	// Create buffers to capture output
	var stdoutBuf, stderrBuf bytes.Buffer

	// Create fake file descriptors
	stdoutR, stdoutW, _ := os.Pipe()
	stderrR, stderrW, _ := os.Pipe()

	os.Stdout = stdoutW
	os.Stderr = stderrW

	// Start goroutines to read from pipes
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(&stdoutBuf, stdoutR)
	}()

	go func() {
		defer wg.Done()
		io.Copy(&stderrBuf, stderrR)
	}()

	// Execute the command (dispatcher expects command name and then args)
	err := s.dispatcher.Execute(append([]string{params.Name}, args...))

	// Close write ends of pipes
	stdoutW.Close()
	stderrW.Close()

	// Wait for readers to finish
	wg.Wait()

	// Restore original stdout/stderr
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Prepare the response
	var contents []Content
	isError := err != nil

	// Determine output format
	outputFormat := OutputFormatRaw
	if formatter, ok := cmd.(OutputFormatter); ok {
		outputFormat = formatter.OutputFormat()
	}

	// Combine output
	output := stdoutBuf.String()
	if output == "" && stderrBuf.Len() > 0 {
		output = stderrBuf.String()
	} else if stderrBuf.Len() > 0 {
		// If we have both stdout and stderr, append stderr
		output = output + "\n" + stderrBuf.String()
	}

	if err != nil {
		// Include error message in output
		if output != "" {
			output = output + "\n" + err.Error()
		} else {
			output = err.Error()
		}
	}

	// Create content based on output format
	if outputFormat == OutputFormatJSON && json.Valid([]byte(output)) {
		// For valid JSON output, include it as data
		contents = append(contents, Content{
			Type:     "text",
			Text:     output,
			Data:     json.RawMessage(output),
			MimeType: "application/json",
		})
	} else {
		// For text output
		contents = append(contents, Content{
			Type: "text",
			Text: output,
		})
	}

	result := ToolCallResult{
		Content: contents,
		IsError: isError,
	}

	s.sendResponse(request.ID, result)
}

// handleResourcesList handles the resources/list request
func (s *MCPServer) handleResourcesList(request MCPRequest) {
	if !s.initialized {
		s.sendErrorResponse(request.ID, -32002, "Server not initialized", nil)
		return
	}

	// Return empty resources list
	result := ResourcesListResult{
		Resources: []Resource{},
	}

	s.sendResponse(request.ID, result)
}

// handleResourceRead handles the resources/read request
func (s *MCPServer) handleResourceRead(request MCPRequest) {
	if !s.initialized {
		s.sendErrorResponse(request.ID, -32002, "Server not initialized", nil)
		return
	}

	// Resources not implemented
	s.sendErrorResponse(request.ID, -32601, "Method not implemented",
		"Resource reading is not supported by this server")
}

// handlePromptsList handles the prompts/list request
func (s *MCPServer) handlePromptsList(request MCPRequest) {
	if !s.initialized {
		s.sendErrorResponse(request.ID, -32002, "Server not initialized", nil)
		return
	}

	// Return empty prompts list
	result := PromptsListResult{
		Prompts: []Prompt{},
	}

	s.sendResponse(request.ID, result)
}

// handlePromptGet handles the prompts/get request
func (s *MCPServer) handlePromptGet(request MCPRequest) {
	if !s.initialized {
		s.sendErrorResponse(request.ID, -32002, "Server not initialized", nil)
		return
	}

	// Prompts not implemented
	s.sendErrorResponse(request.ID, -32601, "Method not implemented",
		"Prompt retrieval is not supported by this server")
}

// sendResponse sends a successful JSON-RPC response
func (s *MCPServer) sendResponse(id any, result interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}

	data, err := json.Marshal(response)
	if err != nil {
		fmt.Fprintf(s.errorOutput, "Error marshaling response: %v\n", err)
		return
	}

	fmt.Fprintln(s.output, string(data))
}

// sendErrorResponse sends an error JSON-RPC response
func (s *MCPServer) sendErrorResponse(id any, code int, message string, data any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		fmt.Fprintf(s.errorOutput, "Error marshaling error response: %v\n", err)
		return
	}

	fmt.Fprintln(s.output, string(responseData))
}

// MCPServerCommand creates a command that runs the dispatcher as an MCP server
type MCPServerCommand struct {
	dispatcher *Dispatcher
	flags      *FlagSet
}

// NewMCPServerCommand creates a new MCP server command
func NewMCPServerCommand(dispatcher *Dispatcher) *MCPServerCommand {
	fs := NewFlagSet("mcp-server")

	return &MCPServerCommand{
		dispatcher: dispatcher,
		flags:      fs,
	}
}

// FlagSet returns the flagset for this command
func (c *MCPServerCommand) FlagSet() *FlagSet {
	return c.flags
}

// Run executes the MCP server
func (c *MCPServerCommand) Run(fs *FlagSet, args []string) error {
	server := NewMCPServer(c.dispatcher)
	return server.Run()
}

// Usage returns the usage description for this command
func (c *MCPServerCommand) Usage() string {
	return "Run as an MCP server for remote command execution"
}

// OutputFormat returns the output format for the MCP server command itself
func (c *MCPServerCommand) OutputFormat() OutputFormat {
	return OutputFormatJSON
}
