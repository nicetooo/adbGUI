package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// registerPluginTools registers plugin management tools
func (s *MCPServer) registerPluginTools() {
	// plugin_list - List all plugins
	s.server.AddTool(
		mcp.NewTool("plugin_list",
			mcp.WithDescription(`List all saved plugins.

Returns an array of plugin metadata including ID, name, version, author, description, 
enabled status, filters, and timestamps.

EXAMPLE:
  # List all plugins
  plugin_list

RETURNS: Array of plugin metadata objects`),
		),
		s.handlePluginList,
	)

	// plugin_get - Get plugin details
	s.server.AddTool(
		mcp.NewTool("plugin_get",
			mcp.WithDescription(`Get detailed information about a specific plugin including source code.

Returns complete plugin data:
- Metadata (name, version, filters, config)
- Source code (TypeScript or JavaScript)
- Compiled code (JavaScript for execution)
- Language type

EXAMPLE:
  plugin_get plugin_id="tracking-validator"

RETURNS: Full plugin object with source code`),
			mcp.WithString("plugin_id",
				mcp.Required(),
				mcp.Description("Plugin ID to retrieve"),
			),
		),
		s.handlePluginGet,
	)

	// plugin_create - Create a new plugin
	s.server.AddTool(
		mcp.NewTool("plugin_create",
			mcp.WithDescription(`Create a new plugin from source code.

PARAMETERS:
  id: Unique plugin ID (e.g., "tracking-validator")
  name: Display name (e.g., "Tracking Parameter Validator")
  version: Semantic version (e.g., "1.0.0")
  source_code: Plugin source code (JavaScript or TypeScript)
  language: "javascript" or "typescript"
  compiled_code: Compiled JavaScript code (for TypeScript, pass TS→JS compiled output)
  
FILTERS (JSON string):
  sources: Array of event sources to match (e.g., ["network", "logcat"])
  types: Array of event types to match (e.g., ["http_request"])
  levels: Array of event levels to match (e.g., ["error", "warn"])
  urlPattern: URL wildcard pattern (e.g., "*/api/track*")
  
CONFIG (JSON string):
  Any user-defined configuration object (e.g., {"requiredParams": ["user_id"]})

EXAMPLE:
  plugin_create \
    id="tracking-validator" \
    name="Tracking Parameter Validator" \
    version="1.0.0" \
    source_code='const plugin = { ... }' \
    language="javascript" \
    compiled_code='const plugin = { ... }' \
    filters_json='{"sources": ["network"], "urlPattern": "*/api/track*"}' \
    config_json='{"requiredParams": ["user_id", "event_name"]}'

RETURNS: Success message with plugin ID`),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("Unique plugin ID (e.g., 'tracking-validator')"),
			),
			mcp.WithString("name",
				mcp.Required(),
				mcp.Description("Plugin display name"),
			),
			mcp.WithString("version",
				mcp.Description("Plugin version (default: '1.0.0')"),
			),
			mcp.WithString("author",
				mcp.Description("Plugin author name (optional)"),
			),
			mcp.WithString("description",
				mcp.Description("Plugin description (optional)"),
			),
			mcp.WithString("source_code",
				mcp.Required(),
				mcp.Description("Plugin source code (JavaScript or TypeScript)"),
			),
			mcp.WithString("language",
				mcp.Description("Source language: 'javascript' or 'typescript' (default: 'javascript')"),
			),
			mcp.WithString("compiled_code",
				mcp.Required(),
				mcp.Description("Compiled JavaScript code for execution"),
			),
			mcp.WithString("filters_json",
				mcp.Description("JSON string of event filters (e.g., '{\"sources\": [\"network\"]}')"),
			),
			mcp.WithString("config_json",
				mcp.Description("JSON string of plugin configuration (e.g., '{\"timeout\": 5000}')"),
			),
		),
		s.handlePluginCreate,
	)

	// plugin_update - Update an existing plugin
	s.server.AddTool(
		mcp.NewTool("plugin_update",
			mcp.WithDescription(`Update an existing plugin. All fields are replaced with new values.

Same parameters as plugin_create. Use plugin_get first to retrieve current values,
then modify and update.

EXAMPLE:
  plugin_update \
    id="tracking-validator" \
    name="Tracking Validator v2" \
    version="2.0.0" \
    source_code='const plugin = { ... }' \
    compiled_code='const plugin = { ... }'

RETURNS: Success message`),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("Plugin ID to update"),
			),
			mcp.WithString("name",
				mcp.Required(),
				mcp.Description("New plugin name"),
			),
			mcp.WithString("version",
				mcp.Description("New version"),
			),
			mcp.WithString("author",
				mcp.Description("New author"),
			),
			mcp.WithString("description",
				mcp.Description("New description"),
			),
			mcp.WithString("source_code",
				mcp.Required(),
				mcp.Description("New source code"),
			),
			mcp.WithString("language",
				mcp.Description("Source language"),
			),
			mcp.WithString("compiled_code",
				mcp.Required(),
				mcp.Description("New compiled code"),
			),
			mcp.WithString("filters_json",
				mcp.Description("New filters JSON"),
			),
			mcp.WithString("config_json",
				mcp.Description("New config JSON"),
			),
		),
		s.handlePluginUpdate,
	)

	// plugin_delete - Delete a plugin
	s.server.AddTool(
		mcp.NewTool("plugin_delete",
			mcp.WithDescription(`Delete a plugin permanently.

The plugin will be unloaded from memory and removed from the database.
This action cannot be undone.

EXAMPLE:
  plugin_delete plugin_id="old-plugin"

RETURNS: Success message`),
			mcp.WithString("plugin_id",
				mcp.Required(),
				mcp.Description("Plugin ID to delete"),
			),
		),
		s.handlePluginDelete,
	)

	// plugin_toggle - Enable/disable a plugin
	s.server.AddTool(
		mcp.NewTool("plugin_toggle",
			mcp.WithDescription(`Enable or disable a plugin without deleting it.

Disabled plugins will not process any events but remain in the database.

EXAMPLE:
  plugin_toggle plugin_id="tracking-validator" enabled=true
  plugin_toggle plugin_id="old-plugin" enabled=false

RETURNS: Success message`),
			mcp.WithString("plugin_id",
				mcp.Required(),
				mcp.Description("Plugin ID to toggle"),
			),
			mcp.WithBoolean("enabled",
				mcp.Required(),
				mcp.Description("Enable (true) or disable (false) the plugin"),
			),
		),
		s.handlePluginToggle,
	)

	// plugin_test - Test a plugin against a specific event
	s.server.AddTool(
		mcp.NewTool("plugin_test",
			mcp.WithDescription(`Test plugin code against a specific event without saving to database.

This is useful for debugging and validating plugin logic before deployment.

WORKFLOW:
1. Find an event ID using session_events
2. Write your plugin code
3. Test it with plugin_test
4. If results are correct, create the plugin with plugin_create

EXAMPLE:
  plugin_test \
    script='const plugin = { metadata: {...}, onEvent: function(event, context) { ... } }' \
    event_id="abc123-event-id"

RETURNS: Array of derived events that the plugin would generate`),
			mcp.WithString("script",
				mcp.Required(),
				mcp.Description("Plugin JavaScript code to test"),
			),
			mcp.WithString("event_id",
				mcp.Required(),
				mcp.Description("Event ID to test against (from session_events)"),
			),
		),
		s.handlePluginTest,
	)
}

// ========== Handler functions ==========

func (s *MCPServer) handlePluginList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	plugins, err := s.app.ListPlugins()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Error: %v", err))},
			IsError: true,
		}, nil
	}

	jsonData, _ := json.MarshalIndent(plugins, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent(string(jsonData))},
	}, nil
}

func (s *MCPServer) handlePluginGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	pluginID, _ := args["plugin_id"].(string)

	if pluginID == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent("Error: plugin_id is required")},
			IsError: true,
		}, nil
	}

	plugin, err := s.app.GetPlugin(pluginID)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Error: %v", err))},
			IsError: true,
		}, nil
	}

	jsonData, _ := json.MarshalIndent(plugin, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent(string(jsonData))},
	}, nil
}

func (s *MCPServer) handlePluginCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handlePluginSave(ctx, request, false)
}

func (s *MCPServer) handlePluginUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handlePluginSave(ctx, request, true)
}

func (s *MCPServer) handlePluginSave(ctx context.Context, request mcp.CallToolRequest, isUpdate bool) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	id, _ := args["id"].(string)
	name, _ := args["name"].(string)
	version, _ := args["version"].(string)
	author, _ := args["author"].(string)
	description, _ := args["description"].(string)
	sourceCode, _ := args["source_code"].(string)
	language, _ := args["language"].(string)
	compiledCode, _ := args["compiled_code"].(string)
	filtersJSON, _ := args["filters_json"].(string)
	configJSON, _ := args["config_json"].(string)

	// Validate required fields
	if id == "" || name == "" || sourceCode == "" || compiledCode == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent("Error: id, name, source_code, and compiled_code are required")},
			IsError: true,
		}, nil
	}

	// Set defaults
	if version == "" {
		version = "1.0.0"
	}
	if language == "" {
		language = "javascript"
	}

	// Parse filters
	var filters struct {
		Sources    []string `json:"sources"`
		Types      []string `json:"types"`
		Levels     []string `json:"levels"`
		URLPattern string   `json:"urlPattern"`
		TitleMatch string   `json:"titleMatch"`
	}
	if filtersJSON != "" {
		if err := json.Unmarshal([]byte(filtersJSON), &filters); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Error parsing filters_json: %v", err))},
				IsError: true,
			}, nil
		}
	}

	// Parse config
	var config map[string]interface{}
	if configJSON != "" {
		if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Error parsing config_json: %v", err))},
				IsError: true,
			}, nil
		}
	} else {
		config = make(map[string]interface{})
	}

	// Import PluginSaveRequest and PluginFilters from main package
	// These types need to be accessible via the bridge
	req := struct {
		ID           string                 `json:"id"`
		Name         string                 `json:"name"`
		Version      string                 `json:"version"`
		Author       string                 `json:"author"`
		Description  string                 `json:"description"`
		SourceCode   string                 `json:"sourceCode"`
		Language     string                 `json:"language"`
		CompiledCode string                 `json:"compiledCode"`
		Filters      interface{}            `json:"filters"`
		Config       map[string]interface{} `json:"config"`
	}{
		ID:           id,
		Name:         name,
		Version:      version,
		Author:       author,
		Description:  description,
		SourceCode:   sourceCode,
		Language:     language,
		CompiledCode: compiledCode,
		Filters: map[string]interface{}{
			"sources":    filters.Sources,
			"types":      filters.Types,
			"levels":     filters.Levels,
			"urlPattern": filters.URLPattern,
			"titleMatch": filters.TitleMatch,
		},
		Config: config,
	}

	// Convert to JSON and back to get proper type
	reqJSON, _ := json.Marshal(req)
	var saveReq interface{}
	json.Unmarshal(reqJSON, &saveReq)

	// Call bridge method
	err := s.app.SavePlugin(saveReq)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Error: %v", err))},
			IsError: true,
		}, nil
	}

	action := "created"
	if isUpdate {
		action = "updated"
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Plugin %s: %s", action, id))},
	}, nil
}

func (s *MCPServer) handlePluginDelete(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	pluginID, _ := args["plugin_id"].(string)

	if pluginID == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent("Error: plugin_id is required")},
			IsError: true,
		}, nil
	}

	err := s.app.DeletePlugin(pluginID)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Error: %v", err))},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Plugin deleted: %s", pluginID))},
	}, nil
}

func (s *MCPServer) handlePluginToggle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	pluginID, _ := args["plugin_id"].(string)
	enabled, enabledOk := args["enabled"].(bool)

	if pluginID == "" || !enabledOk {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent("Error: plugin_id and enabled are required")},
			IsError: true,
		}, nil
	}

	err := s.app.TogglePlugin(pluginID, enabled)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Error: %v", err))},
			IsError: true,
		}, nil
	}

	action := "enabled"
	if !enabled {
		action = "disabled"
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Plugin %s: %s", action, pluginID))},
	}, nil
}

func (s *MCPServer) handlePluginTest(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	script, _ := args["script"].(string)
	eventID, _ := args["event_id"].(string)

	if script == "" || eventID == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent("Error: script and event_id are required")},
			IsError: true,
		}, nil
	}

	derivedEvents, err := s.app.TestPlugin(script, eventID)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Error: %v", err))},
			IsError: true,
		}, nil
	}

	jsonData, _ := json.MarshalIndent(derivedEvents, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent(string(jsonData))},
	}, nil
}
