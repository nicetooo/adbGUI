package main

import (
	"context"
	"fmt"
	"time"
)

// ========================================
// AI Service APIs (exposed to frontend)
// ========================================

// AIServiceInfo represents the AI service status info
type AIServiceInfo struct {
	Status   string                 `json:"status"`
	Enabled  bool                   `json:"enabled"`
	Provider *AIProviderInfo        `json:"provider,omitempty"`
	Error    string                 `json:"error,omitempty"`
	Features *AIFeaturesConfig      `json:"features,omitempty"`
}

// AIProviderInfo represents the current provider info
type AIProviderInfo struct {
	Name     string `json:"name"`
	Model    string `json:"model"`
	Endpoint string `json:"endpoint"`
	Type     string `json:"type"` // "local" or "online"
}

// DiscoveredService represents a discovered LLM service for frontend
type DiscoveredService struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Endpoint string   `json:"endpoint"`
	Status   string   `json:"status"`
	Models   []string `json:"models"`
}

// GetAIServiceInfo returns the current AI service status
func (a *App) GetAIServiceInfo() AIServiceInfo {
	a.aiServiceMu.RLock()
	defer a.aiServiceMu.RUnlock()

	info := AIServiceInfo{
		Status:  "disabled",
		Enabled: false,
	}

	if a.aiConfigMgr != nil {
		config := a.aiConfigMgr.GetConfig()
		info.Enabled = config.Enabled
		info.Features = config.Features
	}

	if a.aiService == nil {
		return info
	}

	info.Status = string(a.aiService.GetStatus())

	provider := a.aiService.GetProvider()
	if provider != nil {
		providerType := "online"
		// Check if it's a local service
		name := provider.Name()
		if name == "Ollama" || name == "LM Studio" || name == "LocalAI" {
			providerType = "local"
		}
		info.Provider = &AIProviderInfo{
			Name:     provider.Name(),
			Model:    provider.Model(),
			Endpoint: provider.Endpoint(),
			Type:     providerType,
		}
	}

	return info
}

// GetAIConfig returns the current AI configuration
func (a *App) GetAIConfig() *AIConfig {
	a.aiServiceMu.RLock()
	defer a.aiServiceMu.RUnlock()

	if a.aiConfigMgr == nil {
		return DefaultAIConfig()
	}
	return a.aiConfigMgr.GetConfig()
}

// SetAIEnabled enables or disables AI features
func (a *App) SetAIEnabled(enabled bool) error {
	a.aiServiceMu.Lock()
	defer a.aiServiceMu.Unlock()

	if a.aiConfigMgr == nil {
		return fmt.Errorf("AI config manager not initialized")
	}

	if err := a.aiConfigMgr.SetEnabled(enabled); err != nil {
		return err
	}

	// Reinitialize AI service if enabling
	if enabled && a.aiService != nil {
		config := a.aiConfigMgr.GetConfig()
		if err := a.aiService.RefreshProviders(context.Background()); err != nil {
			a.Log("Failed to refresh AI providers: %v", err)
		}
		_ = config // Config is already loaded
	}

	return nil
}

// SetAIPreferredSource sets the preferred LLM source
func (a *App) SetAIPreferredSource(source string) error {
	a.aiServiceMu.Lock()
	defer a.aiServiceMu.Unlock()

	if a.aiConfigMgr == nil {
		return fmt.Errorf("AI config manager not initialized")
	}

	return a.aiConfigMgr.SetPreferredSource(LLMSource(source))
}

// DiscoverLLMServices discovers available LLM services
func (a *App) DiscoverLLMServices() []DiscoveredService {
	a.aiServiceMu.RLock()
	aiService := a.aiService
	a.aiServiceMu.RUnlock()

	if aiService == nil {
		// Create a temporary discovery
		discovery := NewLLMDiscovery()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		services := discovery.DiscoverAll(ctx)
		return convertToDiscoveredServices(services)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	services := aiService.DiscoverServices(ctx)
	return convertToDiscoveredServices(services)
}

// convertToDiscoveredServices converts LocalLLMService to DiscoveredService
func convertToDiscoveredServices(services []LocalLLMService) []DiscoveredService {
	result := make([]DiscoveredService, len(services))
	for i, s := range services {
		result[i] = DiscoveredService{
			Name:     s.Name,
			Type:     string(s.Type),
			Endpoint: s.Endpoint,
			Status:   string(s.Status),
			Models:   s.Models,
		}
	}
	return result
}

// SetOpenAIConfig configures the OpenAI provider
func (a *App) SetOpenAIConfig(apiKey, model string, enabled bool) error {
	a.aiServiceMu.Lock()
	defer a.aiServiceMu.Unlock()

	if a.aiConfigMgr == nil {
		return fmt.Errorf("AI config manager not initialized")
	}

	return a.aiConfigMgr.SetOpenAIConfig(apiKey, model, enabled)
}

// SetClaudeConfig configures the Claude provider
func (a *App) SetClaudeConfig(apiKey, model string, enabled bool) error {
	a.aiServiceMu.Lock()
	defer a.aiServiceMu.Unlock()

	if a.aiConfigMgr == nil {
		return fmt.Errorf("AI config manager not initialized")
	}

	return a.aiConfigMgr.SetClaudeConfig(apiKey, model, enabled)
}

// SetCustomProviderConfig configures a custom OpenAI-compatible provider
func (a *App) SetCustomProviderConfig(endpoint, apiKey, model string, enabled bool) error {
	a.aiServiceMu.Lock()
	defer a.aiServiceMu.Unlock()

	if a.aiConfigMgr == nil {
		return fmt.Errorf("AI config manager not initialized")
	}

	return a.aiConfigMgr.SetCustomProviderConfig(endpoint, apiKey, model, enabled)
}

// SwitchAIProvider switches to a different AI provider
func (a *App) SwitchAIProvider(providerType string, config map[string]string) error {
	a.aiServiceMu.Lock()
	defer a.aiServiceMu.Unlock()

	if a.aiService == nil {
		return fmt.Errorf("AI service not initialized")
	}

	err := a.aiService.SwitchProvider(providerType, config)
	if err != nil {
		return err
	}

	// Update config with active provider
	provider := a.aiService.GetProvider()
	if provider != nil && a.aiConfigMgr != nil {
		provType := "online"
		name := provider.Name()
		if name == "Ollama" || name == "LM Studio" || name == "LocalAI" {
			provType = "local"
		}
		_ = a.aiConfigMgr.SetActiveProvider(provType, name, provider.Endpoint(), provider.Model())
	}

	return nil
}

// TestAIProvider tests connectivity to a provider
func (a *App) TestAIProvider(providerType string, config map[string]string) (bool, string) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var provider LLMProvider

	switch providerType {
	case "ollama", "lmstudio", "localai", "custom":
		endpoint := config["endpoint"]
		if endpoint == "" {
			return false, "Endpoint is required"
		}
		provider = NewOpenAICompatProvider(OpenAICompatConfig{
			Name:     providerType,
			Endpoint: endpoint,
			APIKey:   config["api_key"],
			Model:    config["model"],
			Timeout:  15 * time.Second,
		})

	case "openai":
		apiKey := config["api_key"]
		if apiKey == "" {
			return false, "API key is required"
		}
		provider = NewOpenAICompatProvider(OpenAICompatConfig{
			Name:     "OpenAI",
			Endpoint: "https://api.openai.com/v1",
			APIKey:   apiKey,
			Model:    config["model"],
			Timeout:  15 * time.Second,
		})

	case "claude":
		apiKey := config["api_key"]
		if apiKey == "" {
			return false, "API key is required"
		}
		provider = NewClaudeProvider(ClaudeProviderConfig{
			APIKey:  apiKey,
			Model:   config["model"],
			Timeout: 15 * time.Second,
		})

	default:
		return false, fmt.Sprintf("Unknown provider type: %s", providerType)
	}

	if provider.IsAvailable(ctx) {
		return true, "Connection successful"
	}
	return false, "Failed to connect to provider"
}

// GetAvailableModels returns available models for a provider
func (a *App) GetAvailableModels(providerType string, config map[string]string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var provider LLMProvider

	switch providerType {
	case "ollama", "lmstudio", "localai", "custom":
		endpoint := config["endpoint"]
		if endpoint == "" {
			return nil, fmt.Errorf("endpoint is required")
		}
		provider = NewOpenAICompatProvider(OpenAICompatConfig{
			Name:     providerType,
			Endpoint: endpoint,
			APIKey:   config["api_key"],
			Timeout:  15 * time.Second,
		})

	case "openai":
		apiKey := config["api_key"]
		if apiKey == "" {
			return nil, fmt.Errorf("API key is required")
		}
		provider = NewOpenAICompatProvider(OpenAICompatConfig{
			Name:     "OpenAI",
			Endpoint: "https://api.openai.com/v1",
			APIKey:   apiKey,
			Timeout:  15 * time.Second,
		})

	case "claude":
		provider = NewClaudeProvider(ClaudeProviderConfig{
			APIKey: config["api_key"],
		})

	default:
		return nil, fmt.Errorf("unknown provider type: %s", providerType)
	}

	return provider.ListModels(ctx)
}

// SetAIFeature enables or disables a specific AI feature
func (a *App) SetAIFeature(feature string, enabled bool) error {
	a.aiServiceMu.Lock()
	defer a.aiServiceMu.Unlock()

	if a.aiConfigMgr == nil {
		return fmt.Errorf("AI config manager not initialized")
	}

	return a.aiConfigMgr.SetFeature(feature, enabled)
}

// AddCustomLLMEndpoint adds a custom LLM endpoint
func (a *App) AddCustomLLMEndpoint(name, endpoint, serviceType string) error {
	a.aiServiceMu.Lock()
	defer a.aiServiceMu.Unlock()

	if a.aiConfigMgr == nil {
		return fmt.Errorf("AI config manager not initialized")
	}

	return a.aiConfigMgr.AddCustomEndpoint(CustomEndpoint{
		Name:     name,
		Endpoint: endpoint,
		Type:     LLMServiceType(serviceType),
	})
}

// RemoveCustomLLMEndpoint removes a custom LLM endpoint
func (a *App) RemoveCustomLLMEndpoint(name string) error {
	a.aiServiceMu.Lock()
	defer a.aiServiceMu.Unlock()

	if a.aiConfigMgr == nil {
		return fmt.Errorf("AI config manager not initialized")
	}

	return a.aiConfigMgr.RemoveCustomEndpoint(name)
}

// RefreshAIProviders re-discovers available providers
func (a *App) RefreshAIProviders() error {
	a.aiServiceMu.Lock()
	defer a.aiServiceMu.Unlock()

	if a.aiService == nil {
		return fmt.Errorf("AI service not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	return a.aiService.RefreshProviders(ctx)
}

// ========================================
// AI Analysis APIs
// ========================================

// AIGenerateWorkflow generates a workflow from session events
func (a *App) AIGenerateWorkflow(sessionID string, config *WorkflowGeneratorConfig) (*GeneratedWorkflow, error) {
	return a.GenerateWorkflowFromSession(sessionID, config)
}

// AIParseNaturalQuery parses a natural language query into structured query
func (a *App) AIParseNaturalQuery(query string, sessionID string) (*NLQueryResult, error) {
	return a.ParseNaturalQuery(query, sessionID)
}

// AIAnalyzeLog analyzes a log entry
func (a *App) AIAnalyzeLog(tag, message, level string) (*LogAnalysisResult, error) {
	return a.AnalyzeLog(tag, message, level)
}

// AIAnalyzeCrash analyzes a crash event and its preceding events
func (a *App) AIAnalyzeCrash(crashEventID, sessionID string) (*RootCauseAnalysis, error) {
	return a.AnalyzeCrashRootCause(crashEventID, sessionID)
}

// AISuggestAssertions suggests assertions based on session events
func (a *App) AISuggestAssertions(sessionID string) ([]AssertionSuggestion, error) {
	return a.SuggestAssertions(sessionID)
}

// AISummarizeSession generates an AI-powered session summary
func (a *App) AISummarizeSession(sessionID string) (*SessionSummary, error) {
	return a.SummarizeSession(sessionID)
}

// AIComplete sends a completion request (for testing/advanced use)
func (a *App) AIComplete(messages []map[string]string, options map[string]interface{}) (string, error) {
	a.aiServiceMu.RLock()
	aiService := a.aiService
	a.aiServiceMu.RUnlock()

	if aiService == nil {
		return "", fmt.Errorf("AI service not initialized")
	}

	if !aiService.IsReady() {
		return "", fmt.Errorf("AI service not ready: %s", aiService.GetStatus())
	}

	// Convert messages
	chatMessages := make([]ChatMessage, len(messages))
	for i, m := range messages {
		chatMessages[i] = ChatMessage{
			Role:    m["role"],
			Content: m["content"],
		}
	}

	// Build request
	req := &CompletionRequest{
		Messages: chatMessages,
	}

	if model, ok := options["model"].(string); ok {
		req.Model = model
	}
	if maxTokens, ok := options["max_tokens"].(float64); ok {
		req.MaxTokens = int(maxTokens)
	}
	if temp, ok := options["temperature"].(float64); ok {
		req.Temperature = temp
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := aiService.Complete(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}
