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

PLUGIN API SPECIFICATION:

Your plugin code must export a 'plugin' object with the following structure:

  const plugin = {
    metadata: {
      name: "string",        // Display name (optional, overridden by 'name' parameter)
      version: "string",     // Semantic version (optional)
      description: "string"  // What this plugin does (optional)
    },
    
    onInit: function(context) {
      // Optional: Called once when plugin is loaded
      // Use context.state to initialize persistent state across events
      // Example: context.state.counter = 0;
    },
    
    onEvent: function(event, context) {
      // Required: Called for each matching event
      // Return a PluginResult object or null/undefined
      // Example:
      //   return {
      //     derivedEvents: [{ source: "plugin", type: "alert", level: "warn", title: "Issue found" }]
      //   };
    },
    
    onDestroy: function(context) {
      // Optional: Called when plugin is unloaded
      // Use for cleanup if needed
    }
  };

EVENT OBJECT (passed to onEvent):
  {
    id: "string",               // Unique event ID
    deviceId: "string",         // Device ID
    sessionId: "string",        // Session ID
    timestamp: number,          // Unix milliseconds
    relativeTime: number,       // Milliseconds from session start
    source: "string",           // Event source: "network", "logcat", "app", "device", etc.
    category: "string",         // Category: "network", "log", "state", etc.
    type: "string",             // Event type: "http_request", "logcat", "app_crash", etc.
    level: "string",            // Level: "info", "warn", "error", "debug", "verbose", "fatal"
    title: "string",            // Short description
    content?: "string",         // Optional detailed content
    data?: any                  // Event-specific data as JSON object
  }

CONTEXT OBJECT (passed to onEvent and onInit):
  {
    config: any,                // User configuration from config_json parameter
    state: any,                 // Persistent state object (survives across events)
    
    // Helper functions:
    log(message: string): void,              // Write debug log
    emit(type, title, data): void,           // Emit a derived event (shortcut)
    jsonPath(json, path): any,               // Extract JSON value by path (e.g., "user.name")
    matchURL(url, pattern): boolean          // URL wildcard matching (* supported)
  }

PLUGIN RESULT (return value from onEvent):
  {
    derivedEvents?: [           // Array of new events to generate
      {
        source: "string",       // Event source (typically "plugin")
        type: "string",         // Event type (e.g., "validation_error", "alert")
        level: "string",        // "info", "warn", or "error"
        title: "string",        // Short description
        content?: "string",     // Optional detailed content
        data?: any              // Optional data object
      }
    ],
    tags?: ["string"],          // Tags to add to the original event
    metadata?: { key: value }   // Metadata to attach to the original event
  }

COMPLETE EXAMPLE - API Tracking Validator:

  const plugin = {
    metadata: {
      name: "API Tracking Validator",
      version: "1.0.0",
      description: "Validates tracking API calls have required parameters"
    },
    
    onEvent: function(event, context) {
      // Extract URL from event data
      const data = event.data || {};
      const url = data.url || "";
      
      if (!url) return null;
      
      try {
        const urlObj = new URL(url);
        const params = urlObj.searchParams;
        
        // Get required parameters from config
        const required = context.config.requiredParams || ["user_id", "event_name"];
        const missing = [];
        
        for (const param of required) {
          if (!params.has(param)) {
            missing.push(param);
          }
        }
        
        // If parameters are missing, emit a warning event
        if (missing.length > 0) {
          return {
            derivedEvents: [{
              source: "plugin",
              type: "validation_error",
              level: "warn",
              title: "Missing tracking parameters: " + missing.join(", "),
              data: {
                url: url,
                missingParams: missing,
                foundParams: Array.from(params.keys())
              }
            }]
          };
        }
      } catch (err) {
        context.log("Failed to parse URL: " + err.message);
      }
      
      return null;
    }
  };

PARAMETERS:
  id: Unique plugin ID (lowercase, numbers, hyphens only, e.g., "tracking-validator")
  name: Display name (e.g., "Tracking Parameter Validator")
  version: Semantic version (default: "1.0.0")
  author: Plugin author name (optional)
  description: Plugin description (optional)
  source_code: Plugin source code (JavaScript or TypeScript)
  language: "javascript" or "typescript" (default: "javascript")
  compiled_code: Compiled JavaScript code (for TypeScript, compile to JS first)
  
FILTERS (JSON string):
  All filters use AND logic (event must match all specified filters).
  Empty arrays match all values for that field. Multiple values use OR logic within the field.
  
  {
    "sources": ["network", "logcat"],   // Event sources (OR logic)
    "types": ["http_request"],          // Event types (OR logic)
    "levels": ["error", "warn"],        // Event levels (OR logic)
    "urlPattern": "*/api/track*",       // URL wildcard (only for network events, * = wildcard)
    "titleMatch": "ActivityManager.*"   // Title regex pattern (optional)
  }
  
CONFIG (JSON string):
  User-defined configuration passed to context.config in plugin code.
  Example: {"requiredParams": ["user_id", "event_name"], "timeout": 5000}

USAGE EXAMPLE:
  # 1. Create the plugin
  plugin_create \
    id="tracking-validator" \
    name="API Tracking Validator" \
    version="1.0.0" \
    source_code='const plugin = { ... (see complete example above) ... }' \
    language="javascript" \
    compiled_code='const plugin = { ... }' \
    filters_json='{"sources": ["network"], "urlPattern": "*/api/track*"}' \
    config_json='{"requiredParams": ["user_id", "event_name"]}'

  # 2. Test against a real event before deploying
  plugin_test \
    script='const plugin = { ... }' \
    event_id="abc123-network-event-id"

  # 3. Enable the plugin
  plugin_toggle plugin_id="tracking-validator" enabled=true

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

This is the primary debugging tool for plugin development. It executes your plugin
code against a real event from your session and shows exactly what derived events
would be generated.

DEBUGGING WORKFLOW:

1. Find a matching event:
   session_events session_id="<id>" sources="network" limit=10
   
2. Copy an event ID from the results

3. Write your plugin code (see plugin_create for API specification)

4. Test it:
   plugin_test script='const plugin = { ... }' event_id="<event-id>"

5. Check the output:
   - If empty array: plugin didn't generate any derived events (check your logic)
   - If has events: review the derived event structure
   - If error: fix the JavaScript error and test again

6. Iterate until working correctly

7. Create the plugin:
   plugin_create id="my-plugin" name="My Plugin" source_code='...' compiled_code='...'

ERROR HANDLING:

If your plugin code throws an error, you'll see:
  Error: <JavaScript error message with line number>

Common errors:
- "undefined is not a function": Check function names (onEvent, not onevent)
- "Cannot read property 'x' of undefined": Add null checks for event.data
- "URL is not a constructor": Wrap URL parsing in try-catch

DEBUGGING TIPS:

1. Use context.log() to output debug messages:
   context.log("URL: " + event.data.url);
   
2. Check event.data structure first:
   context.log("Event data: " + JSON.stringify(event.data));
   
3. Always validate data before processing:
   if (!event.data || !event.data.url) return null;

4. Test with multiple events to ensure robustness:
   - Find events with different data structures
   - Test edge cases (missing fields, null values, etc.)

EXAMPLE:
  # Get an event ID from a session
  session_events session_id="session-abc" sources="network" limit=1
  
  # Test plugin against that event
  plugin_test \
    script='const plugin = {
      onEvent: function(event, context) {
        context.log("Processing event: " + event.type);
        if (event.source === "network") {
          return {
            derivedEvents: [{
              source: "plugin",
              type: "network_detected",
              level: "info",
              title: "Network event processed"
            }]
          };
        }
        return null;
      }
    }' \
    event_id="abc123-event-id"

RETURNS: 
  On success: JSON array of derived events that would be generated
  On error: Error message with details
  
NOTES:
  - The event must exist in the database (use session_events to find IDs)
  - context.state is available but not persisted (testing is stateless)
  - context.log() output is not returned (only visible in server logs)`),
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
