package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// ========================================
// AI Configuration - AI 配置管理
// ========================================

// LLMSource represents the preferred LLM source
type LLMSource string

const (
	LLMSourceAuto   LLMSource = "auto"   // Auto-detect best available
	LLMSourceLocal  LLMSource = "local"  // Prefer local services
	LLMSourceOnline LLMSource = "online" // Prefer online APIs
)

// AIConfig represents the AI service configuration
type AIConfig struct {
	// General settings
	Enabled         bool      `json:"enabled"`
	PreferredSource LLMSource `json:"preferredSource"` // "auto", "local", "online"

	// Local service settings
	LocalServices *LocalServicesConfig `json:"localServices,omitempty"`

	// Online provider settings
	OnlineProviders *OnlineProvidersConfig `json:"onlineProviders,omitempty"`

	// Active provider (runtime state)
	ActiveProvider *ActiveProviderConfig `json:"activeProvider,omitempty"`

	// Feature flags
	Features *AIFeaturesConfig `json:"features,omitempty"`
}

// LocalServicesConfig configures local LLM service discovery
type LocalServicesConfig struct {
	AutoDetect      bool             `json:"autoDetect"`
	CustomEndpoints []CustomEndpoint `json:"customEndpoints,omitempty"`
}

// OnlineProvidersConfig configures online LLM providers
type OnlineProvidersConfig struct {
	OpenAI ProviderConfig `json:"openai"`
	Claude ProviderConfig `json:"claude"`
	Custom ProviderConfig `json:"custom"`
}

// ProviderConfig configures a single provider
type ProviderConfig struct {
	Enabled  bool   `json:"enabled"`
	APIKey   string `json:"apiKey,omitempty"`
	Endpoint string `json:"endpoint,omitempty"` // For custom providers
	Model    string `json:"model,omitempty"`
}

// ActiveProviderConfig represents the currently active provider
type ActiveProviderConfig struct {
	Type     string `json:"type"`     // "local", "online"
	Name     string `json:"name"`     // "Ollama", "OpenAI", etc.
	Endpoint string `json:"endpoint"` // API endpoint
	Model    string `json:"model"`    // Model name
}

// AIFeaturesConfig configures AI features
type AIFeaturesConfig struct {
	LogAnalysis        bool `json:"logAnalysis"`        // Smart log analysis
	NaturalSearch      bool `json:"naturalSearch"`      // Natural language search
	WorkflowGeneration bool `json:"workflowGeneration"` // Workflow auto-generation
	WorkflowAI         bool `json:"workflowAI"`         // AI-assisted workflow execution
	CrashAnalysis      bool `json:"crashAnalysis"`      // Crash root cause analysis
	AssertionGen       bool `json:"assertionGen"`       // Auto assertion generation
	VideoAnalysis      bool `json:"videoAnalysis"`      // Video frame analysis
}

// AIConfigManager manages AI configuration persistence
type AIConfigManager struct {
	config   *AIConfig
	configMu sync.RWMutex
	filePath string
}

// NewAIConfigManager creates a new config manager
func NewAIConfigManager(dataDir string) *AIConfigManager {
	return &AIConfigManager{
		filePath: filepath.Join(dataDir, "ai-config.json"),
		config:   DefaultAIConfig(),
	}
}

// DefaultAIConfig returns the default AI configuration
func DefaultAIConfig() *AIConfig {
	return &AIConfig{
		Enabled:         false, // Disabled by default
		PreferredSource: LLMSourceAuto,
		LocalServices: &LocalServicesConfig{
			AutoDetect:      true,
			CustomEndpoints: []CustomEndpoint{},
		},
		OnlineProviders: &OnlineProvidersConfig{
			OpenAI: ProviderConfig{
				Enabled: false,
				Model:   "gpt-4o-mini",
			},
			Claude: ProviderConfig{
				Enabled: false,
				Model:   "claude-3-haiku-20240307",
			},
			Custom: ProviderConfig{
				Enabled: false,
			},
		},
		Features: &AIFeaturesConfig{
			LogAnalysis:        true,
			NaturalSearch:      true,
			WorkflowGeneration: true,
			WorkflowAI:         true,
			CrashAnalysis:      true,
			AssertionGen:       true,
			VideoAnalysis:      true,
		},
	}
}

// Load loads the configuration from disk
func (m *AIConfigManager) Load() error {
	m.configMu.Lock()
	defer m.configMu.Unlock()

	data, err := os.ReadFile(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Use default config
			m.config = DefaultAIConfig()
			return nil
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config AIConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Merge with defaults to ensure all fields exist
	m.config = mergeWithDefaults(&config)
	return nil
}

// Save saves the configuration to disk
func (m *AIConfigManager) Save() error {
	m.configMu.RLock()
	config := m.config
	m.configMu.RUnlock()

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(m.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(m.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfig returns a copy of the current configuration
func (m *AIConfigManager) GetConfig() *AIConfig {
	m.configMu.RLock()
	defer m.configMu.RUnlock()

	// Return a copy
	configCopy := *m.config
	return &configCopy
}

// UpdateConfig updates the configuration
func (m *AIConfigManager) UpdateConfig(updater func(*AIConfig)) error {
	m.configMu.Lock()
	defer m.configMu.Unlock()

	updater(m.config)
	return nil
}

// SetEnabled enables or disables AI features
func (m *AIConfigManager) SetEnabled(enabled bool) error {
	m.configMu.Lock()
	m.config.Enabled = enabled
	m.configMu.Unlock()
	return m.Save()
}

// SetPreferredSource sets the preferred LLM source
func (m *AIConfigManager) SetPreferredSource(source LLMSource) error {
	m.configMu.Lock()
	m.config.PreferredSource = source
	m.configMu.Unlock()
	return m.Save()
}

// SetOpenAIConfig sets OpenAI configuration
func (m *AIConfigManager) SetOpenAIConfig(apiKey, model string, enabled bool) error {
	m.configMu.Lock()
	if m.config.OnlineProviders == nil {
		m.config.OnlineProviders = &OnlineProvidersConfig{}
	}
	m.config.OnlineProviders.OpenAI = ProviderConfig{
		Enabled: enabled,
		APIKey:  apiKey,
		Model:   model,
	}
	m.configMu.Unlock()
	return m.Save()
}

// SetClaudeConfig sets Claude configuration
func (m *AIConfigManager) SetClaudeConfig(apiKey, model string, enabled bool) error {
	m.configMu.Lock()
	if m.config.OnlineProviders == nil {
		m.config.OnlineProviders = &OnlineProvidersConfig{}
	}
	m.config.OnlineProviders.Claude = ProviderConfig{
		Enabled: enabled,
		APIKey:  apiKey,
		Model:   model,
	}
	m.configMu.Unlock()
	return m.Save()
}

// SetCustomProviderConfig sets custom provider configuration
func (m *AIConfigManager) SetCustomProviderConfig(endpoint, apiKey, model string, enabled bool) error {
	m.configMu.Lock()
	if m.config.OnlineProviders == nil {
		m.config.OnlineProviders = &OnlineProvidersConfig{}
	}
	m.config.OnlineProviders.Custom = ProviderConfig{
		Enabled:  enabled,
		Endpoint: endpoint,
		APIKey:   apiKey,
		Model:    model,
	}
	m.configMu.Unlock()
	return m.Save()
}

// AddCustomEndpoint adds a custom local endpoint
func (m *AIConfigManager) AddCustomEndpoint(endpoint CustomEndpoint) error {
	m.configMu.Lock()
	if m.config.LocalServices == nil {
		m.config.LocalServices = &LocalServicesConfig{
			AutoDetect: true,
		}
	}
	m.config.LocalServices.CustomEndpoints = append(m.config.LocalServices.CustomEndpoints, endpoint)
	m.configMu.Unlock()
	return m.Save()
}

// RemoveCustomEndpoint removes a custom endpoint by name
func (m *AIConfigManager) RemoveCustomEndpoint(name string) error {
	m.configMu.Lock()
	if m.config.LocalServices != nil {
		var filtered []CustomEndpoint
		for _, e := range m.config.LocalServices.CustomEndpoints {
			if e.Name != name {
				filtered = append(filtered, e)
			}
		}
		m.config.LocalServices.CustomEndpoints = filtered
	}
	m.configMu.Unlock()
	return m.Save()
}

// SetActiveProvider sets the active provider
func (m *AIConfigManager) SetActiveProvider(providerType, name, endpoint, model string) error {
	m.configMu.Lock()
	m.config.ActiveProvider = &ActiveProviderConfig{
		Type:     providerType,
		Name:     name,
		Endpoint: endpoint,
		Model:    model,
	}
	m.configMu.Unlock()
	return m.Save()
}

// SetFeature enables or disables a specific feature
func (m *AIConfigManager) SetFeature(feature string, enabled bool) error {
	m.configMu.Lock()
	if m.config.Features == nil {
		m.config.Features = &AIFeaturesConfig{}
	}

	switch feature {
	case "logAnalysis":
		m.config.Features.LogAnalysis = enabled
	case "naturalSearch":
		m.config.Features.NaturalSearch = enabled
	case "workflowGeneration":
		m.config.Features.WorkflowGeneration = enabled
	case "workflowAI":
		m.config.Features.WorkflowAI = enabled
	case "crashAnalysis":
		m.config.Features.CrashAnalysis = enabled
	case "assertionGen":
		m.config.Features.AssertionGen = enabled
	case "videoAnalysis":
		m.config.Features.VideoAnalysis = enabled
	default:
		m.configMu.Unlock()
		return fmt.Errorf("unknown feature: %s", feature)
	}

	m.configMu.Unlock()
	return m.Save()
}

// mergeWithDefaults merges loaded config with defaults
func mergeWithDefaults(loaded *AIConfig) *AIConfig {
	defaults := DefaultAIConfig()

	if loaded.LocalServices == nil {
		loaded.LocalServices = defaults.LocalServices
	}
	if loaded.OnlineProviders == nil {
		loaded.OnlineProviders = defaults.OnlineProviders
	}
	if loaded.Features == nil {
		loaded.Features = defaults.Features
	}

	return loaded
}

// MaskAPIKey returns a masked version of an API key for display
func MaskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "****"
	}
	return apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
}

// ValidateAPIKey performs basic validation on an API key
func ValidateAPIKey(apiKey string) bool {
	return len(apiKey) >= 20
}
