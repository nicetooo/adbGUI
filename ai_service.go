package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ========================================
// AI Service - AI 服务主框架
// ========================================

// AIService provides AI capabilities for the application
type AIService struct {
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex

	// Configuration
	config *AIConfig

	// LLM Provider (OpenAI-compatible)
	llmProvider LLMProvider

	// Service discovery
	discovery *LLMDiscovery

	// Status
	status    AIServiceStatus
	statusMu  sync.RWMutex
	lastError error
}

// AIServiceStatus represents the current status of the AI service
type AIServiceStatus string

const (
	AIStatusInitializing AIServiceStatus = "initializing"
	AIStatusReady        AIServiceStatus = "ready"
	AIStatusNoProvider   AIServiceStatus = "no_provider"
	AIStatusError        AIServiceStatus = "error"
	AIStatusDisabled     AIServiceStatus = "disabled"
)

// LLMProvider defines the interface for LLM providers
type LLMProvider interface {
	// Name returns the provider name
	Name() string

	// Model returns the current model name
	Model() string

	// Endpoint returns the API endpoint
	Endpoint() string

	// IsAvailable checks if the provider is available
	IsAvailable(ctx context.Context) bool

	// SupportsVision returns whether the current model supports vision/image input
	SupportsVision() bool

	// Complete sends a completion request
	Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)

	// CompleteStream sends a streaming completion request
	CompleteStream(ctx context.Context, req *CompletionRequest) (<-chan StreamChunk, error)

	// ListModels returns available models
	ListModels(ctx context.Context) ([]string, error)
}

// CompletionRequest represents a completion request
type CompletionRequest struct {
	Messages    []ChatMessage `json:"messages"`
	Model       string        `json:"model,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	TopP        float64       `json:"top_p,omitempty"`
	Stop        []string      `json:"stop,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
	Timeout     time.Duration `json:"-"` // Request-specific timeout (0 = use default)
}

// ChatMessage represents a chat message
type ChatMessage struct {
	Role    string         `json:"role"`              // "system", "user", "assistant"
	Content string         `json:"content"`           // Text content
	Images  []ImageContent `json:"images,omitempty"`  // Image content for vision models
}

// ImageContent represents an image in a message
type ImageContent struct {
	Type   string `json:"type"`   // "base64" or "url"
	Data   string `json:"data"`   // Base64 data (without data:image/xxx;base64, prefix) or URL
	Format string `json:"format"` // "jpeg", "png", "gif", "webp"
}

// CompletionResponse represents a completion response
type CompletionResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Content string `json:"content"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	FinishReason string `json:"finish_reason"`
}

// StreamChunk represents a streaming response chunk
type StreamChunk struct {
	Content string `json:"content"`
	Done    bool   `json:"done"`
	Error   error  `json:"error,omitempty"`
}

// NewAIService creates a new AI service instance
func NewAIService(ctx context.Context, config *AIConfig) (*AIService, error) {
	ctx, cancel := context.WithCancel(ctx)

	service := &AIService{
		ctx:    ctx,
		cancel: cancel,
		config: config,
		status: AIStatusInitializing,
	}

	// Initialize service discovery
	service.discovery = NewLLMDiscovery()

	// Initialize provider based on config
	if err := service.initializeProvider(); err != nil {
		service.setStatus(AIStatusError)
		service.lastError = err
		LogError("ai_service").Err(err).Msg("Failed to initialize AI provider")
		// Don't return error - service can work without provider
	}

	return service, nil
}

// initializeProvider initializes the LLM provider based on configuration
func (s *AIService) initializeProvider() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.config == nil || !s.config.Enabled {
		s.setStatusLocked(AIStatusDisabled)
		return nil
	}

	// First, try to restore active provider from config if available
	if s.config.ActiveProvider != nil && s.config.ActiveProvider.Endpoint != "" {
		if err := s.tryRestoreActiveProvider(); err == nil {
			return nil
		}
		// If restore failed, continue with normal initialization
		LogInfo("ai_service").Msg("Failed to restore active provider, trying discovery")
	}

	// Try to find a working provider based on preference
	var err error
	switch s.config.PreferredSource {
	case LLMSourceAuto:
		return s.autoSelectProvider()
	case LLMSourceLocal:
		err = s.initLocalProvider()
		if err != nil {
			// Fallback to online if local fails
			if onlineErr := s.initOnlineProvider(); onlineErr == nil {
				return nil
			}
		}
		return err
	case LLMSourceOnline:
		err = s.initOnlineProvider()
		if err != nil {
			// Fallback to local if online fails
			if localErr := s.initLocalProvider(); localErr == nil {
				return nil
			}
		}
		return err
	default:
		return s.autoSelectProvider()
	}
}

// tryRestoreActiveProvider tries to restore the previously active provider
func (s *AIService) tryRestoreActiveProvider() error {
	if s.config.ActiveProvider == nil || s.config.ActiveProvider.Endpoint == "" {
		return fmt.Errorf("no active provider configured")
	}

	ap := s.config.ActiveProvider
	var provider LLMProvider

	// Check if it's a local service type
	name := ap.Name
	if name == "Ollama" || name == "ollama" || name == "LM Studio" || name == "lmstudio" || name == "LocalAI" || name == "localai" {
		provider = NewOpenAICompatProvider(OpenAICompatConfig{
			Name:     ap.Name,
			Endpoint: ap.Endpoint,
			APIKey:   "", // Local services typically don't need API key
			Model:    ap.Model,
			Timeout:  30 * time.Second,
		})
	} else if name == "Claude" || name == "claude" {
		// Claude needs API key from config
		if s.config.OnlineProviders != nil && s.config.OnlineProviders.Claude.APIKey != "" {
			provider = NewClaudeProvider(ClaudeProviderConfig{
				APIKey: s.config.OnlineProviders.Claude.APIKey,
				Model:  ap.Model,
			})
		} else {
			return fmt.Errorf("Claude API key not configured")
		}
	} else {
		// OpenAI or custom
		apiKey := ""
		if s.config.OnlineProviders != nil {
			if name == "OpenAI" || name == "openai" {
				apiKey = s.config.OnlineProviders.OpenAI.APIKey
			} else if s.config.OnlineProviders.Custom.Enabled {
				apiKey = s.config.OnlineProviders.Custom.APIKey
			}
		}
		provider = NewOpenAICompatProvider(OpenAICompatConfig{
			Name:     ap.Name,
			Endpoint: ap.Endpoint,
			APIKey:   apiKey,
			Model:    ap.Model,
			Timeout:  30 * time.Second,
		})
	}

	if provider.IsAvailable(s.ctx) {
		s.llmProvider = provider
		s.setStatusLocked(AIStatusReady)
		LogInfo("ai_service").
			Str("provider", ap.Name).
			Str("endpoint", ap.Endpoint).
			Str("model", ap.Model).
			Msg("Restored active provider")
		return nil
	}

	return fmt.Errorf("active provider not available")
}

// autoSelectProvider automatically selects the best available provider
func (s *AIService) autoSelectProvider() error {
	// 1. Try local services first (Ollama, LM Studio, LocalAI)
	if err := s.initLocalProvider(); err == nil {
		return nil
	}

	// 2. Try online providers
	if err := s.initOnlineProvider(); err == nil {
		return nil
	}

	s.setStatusLocked(AIStatusNoProvider)
	return fmt.Errorf("no LLM provider available")
}

// initLocalProvider initializes a local LLM provider
func (s *AIService) initLocalProvider() error {
	// Discover local services
	services := s.discovery.DiscoverAll(s.ctx)

	for _, svc := range services {
		if svc.Status == LLMServiceRunning {
			// Try to connect
			provider := NewOpenAICompatProvider(OpenAICompatConfig{
				Name:     svc.Name,
				Endpoint: svc.OpenAIEndpoint(),
				APIKey:   "", // Local services typically don't need API key
				Model:    svc.GetDefaultModel(),
				Timeout:  30 * time.Second,
			})

			if provider.IsAvailable(s.ctx) {
				s.llmProvider = provider
				s.setStatusLocked(AIStatusReady)
				LogInfo("ai_service").
					Str("provider", svc.Name).
					Str("endpoint", svc.Endpoint).
					Str("model", provider.Model()).
					Msg("Connected to local LLM service")
				return nil
			}
		}
	}

	return fmt.Errorf("no local LLM service available")
}

// initOnlineProvider initializes an online LLM provider
func (s *AIService) initOnlineProvider() error {
	if s.config.OnlineProviders == nil {
		return fmt.Errorf("no online providers configured")
	}

	// Try OpenAI
	if s.config.OnlineProviders.OpenAI.Enabled && s.config.OnlineProviders.OpenAI.APIKey != "" {
		provider := NewOpenAICompatProvider(OpenAICompatConfig{
			Name:     "OpenAI",
			Endpoint: "https://api.openai.com/v1",
			APIKey:   s.config.OnlineProviders.OpenAI.APIKey,
			Model:    s.config.OnlineProviders.OpenAI.Model,
			Timeout:  60 * time.Second,
		})

		if provider.IsAvailable(s.ctx) {
			s.llmProvider = provider
			s.setStatusLocked(AIStatusReady)
			LogInfo("ai_service").
				Str("provider", "OpenAI").
				Str("model", provider.Model()).
				Msg("Connected to OpenAI")
			return nil
		}
	}

	// Try Claude (via OpenAI-compatible endpoint if available)
	if s.config.OnlineProviders.Claude.Enabled && s.config.OnlineProviders.Claude.APIKey != "" {
		provider := NewClaudeProvider(ClaudeProviderConfig{
			APIKey: s.config.OnlineProviders.Claude.APIKey,
			Model:  s.config.OnlineProviders.Claude.Model,
		})

		if provider.IsAvailable(s.ctx) {
			s.llmProvider = provider
			s.setStatusLocked(AIStatusReady)
			LogInfo("ai_service").
				Str("provider", "Claude").
				Str("model", provider.Model()).
				Msg("Connected to Claude")
			return nil
		}
	}

	// Try custom OpenAI-compatible provider
	if s.config.OnlineProviders.Custom.Enabled && s.config.OnlineProviders.Custom.Endpoint != "" {
		provider := NewOpenAICompatProvider(OpenAICompatConfig{
			Name:     "Custom",
			Endpoint: s.config.OnlineProviders.Custom.Endpoint,
			APIKey:   s.config.OnlineProviders.Custom.APIKey,
			Model:    s.config.OnlineProviders.Custom.Model,
			Timeout:  60 * time.Second,
		})

		if provider.IsAvailable(s.ctx) {
			s.llmProvider = provider
			s.setStatusLocked(AIStatusReady)
			LogInfo("ai_service").
				Str("provider", "Custom").
				Str("endpoint", s.config.OnlineProviders.Custom.Endpoint).
				Str("model", provider.Model()).
				Msg("Connected to custom provider")
			return nil
		}
	}

	return fmt.Errorf("no online provider available")
}

// setStatus sets the service status (thread-safe)
func (s *AIService) setStatus(status AIServiceStatus) {
	s.statusMu.Lock()
	defer s.statusMu.Unlock()
	s.status = status
}

// setStatusLocked sets the service status (caller must hold lock)
func (s *AIService) setStatusLocked(status AIServiceStatus) {
	s.statusMu.Lock()
	defer s.statusMu.Unlock()
	s.status = status
}

// GetStatus returns the current service status
func (s *AIService) GetStatus() AIServiceStatus {
	s.statusMu.RLock()
	defer s.statusMu.RUnlock()
	return s.status
}

// IsReady returns true if the service is ready to handle requests
func (s *AIService) IsReady() bool {
	return s.GetStatus() == AIStatusReady
}

// SupportsVision returns whether the current provider/model supports vision input
func (s *AIService) SupportsVision() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.llmProvider == nil {
		return false
	}
	return s.llmProvider.SupportsVision()
}

// GetProvider returns the current LLM provider
func (s *AIService) GetProvider() LLMProvider {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.llmProvider
}

// Complete sends a completion request to the current provider
func (s *AIService) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	if !s.IsReady() {
		return nil, fmt.Errorf("AI service not ready: %s", s.GetStatus())
	}

	provider := s.GetProvider()
	if provider == nil {
		return nil, fmt.Errorf("no LLM provider available")
	}

	return provider.Complete(ctx, req)
}

// CompleteStream sends a streaming completion request
func (s *AIService) CompleteStream(ctx context.Context, req *CompletionRequest) (<-chan StreamChunk, error) {
	if !s.IsReady() {
		return nil, fmt.Errorf("AI service not ready: %s", s.GetStatus())
	}

	provider := s.GetProvider()
	if provider == nil {
		return nil, fmt.Errorf("no LLM provider available")
	}

	return provider.CompleteStream(ctx, req)
}

// DiscoverServices discovers available LLM services
func (s *AIService) DiscoverServices(ctx context.Context) []LocalLLMService {
	if s.discovery == nil {
		return nil
	}
	return s.discovery.DiscoverAll(ctx)
}

// SwitchProvider switches to a different provider
func (s *AIService) SwitchProvider(providerType string, config map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var provider LLMProvider
	var err error

	switch providerType {
	case "ollama", "lmstudio", "localai":
		endpoint := config["endpoint"]
		if endpoint == "" {
			return fmt.Errorf("endpoint required for local provider")
		}
		provider = NewOpenAICompatProvider(OpenAICompatConfig{
			Name:     providerType,
			Endpoint: endpoint,
			Model:    config["model"],
			Timeout:  30 * time.Second,
		})

	case "openai":
		apiKey := config["api_key"]
		if apiKey == "" {
			return fmt.Errorf("API key required for OpenAI")
		}
		provider = NewOpenAICompatProvider(OpenAICompatConfig{
			Name:     "OpenAI",
			Endpoint: "https://api.openai.com/v1",
			APIKey:   apiKey,
			Model:    config["model"],
			Timeout:  60 * time.Second,
		})

	case "claude":
		apiKey := config["api_key"]
		if apiKey == "" {
			return fmt.Errorf("API key required for Claude")
		}
		provider = NewClaudeProvider(ClaudeProviderConfig{
			APIKey: apiKey,
			Model:  config["model"],
		})

	case "custom":
		endpoint := config["endpoint"]
		if endpoint == "" {
			return fmt.Errorf("endpoint required for custom provider")
		}
		provider = NewOpenAICompatProvider(OpenAICompatConfig{
			Name:     "Custom",
			Endpoint: endpoint,
			APIKey:   config["api_key"],
			Model:    config["model"],
			Timeout:  60 * time.Second,
		})

	default:
		return fmt.Errorf("unknown provider type: %s", providerType)
	}

	if !provider.IsAvailable(s.ctx) {
		return fmt.Errorf("provider not available")
	}

	s.llmProvider = provider
	s.setStatusLocked(AIStatusReady)

	LogInfo("ai_service").
		Str("provider", providerType).
		Str("model", provider.Model()).
		Msg("Switched LLM provider")

	return err
}

// RefreshProviders re-discovers available providers
func (s *AIService) RefreshProviders(ctx context.Context) error {
	// Re-discover services (no lock needed for this)
	s.discovery = NewLLMDiscovery()

	// Re-initialize provider (will acquire its own lock internally)
	return s.initializeProvider()
}

// Shutdown shuts down the AI service
func (s *AIService) Shutdown() {
	s.cancel()
	LogInfo("ai_service").Msg("AI service shutdown")
}

// GetServiceInfo returns information about the current AI service state
func (s *AIService) GetServiceInfo() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info := map[string]interface{}{
		"status":  string(s.status),
		"enabled": s.config != nil && s.config.Enabled,
	}

	if s.llmProvider != nil {
		info["provider"] = map[string]interface{}{
			"name":     s.llmProvider.Name(),
			"model":    s.llmProvider.Model(),
			"endpoint": s.llmProvider.Endpoint(),
		}
	}

	if s.lastError != nil {
		info["lastError"] = s.lastError.Error()
	}

	return info
}
