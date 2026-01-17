package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ========================================
// OpenAI Compatible Provider - OpenAI 兼容客户端
// ========================================

// OpenAICompatConfig configures the OpenAI-compatible provider
type OpenAICompatConfig struct {
	Name     string        // Provider name (e.g., "Ollama", "LM Studio", "OpenAI")
	Endpoint string        // API endpoint (e.g., "http://localhost:11434/v1")
	APIKey   string        // API key (optional for local services)
	Model    string        // Model to use
	Timeout  time.Duration // Request timeout
}

// OpenAICompatProvider implements LLMProvider for OpenAI-compatible APIs
type OpenAICompatProvider struct {
	config OpenAICompatConfig
	client *http.Client
}

// NewOpenAICompatProvider creates a new OpenAI-compatible provider
func NewOpenAICompatProvider(config OpenAICompatConfig) *OpenAICompatProvider {
	if config.Timeout == 0 {
		// Default 180 seconds for vision models, 60 for text-only
		config.Timeout = 180 * time.Second
	}
	if config.Endpoint == "" {
		config.Endpoint = "http://localhost:11434/v1"
	}

	return &OpenAICompatProvider{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Name returns the provider name
func (p *OpenAICompatProvider) Name() string {
	return p.config.Name
}

// Model returns the current model name
func (p *OpenAICompatProvider) Model() string {
	return p.config.Model
}

// Endpoint returns the API endpoint
func (p *OpenAICompatProvider) Endpoint() string {
	return p.config.Endpoint
}

// SupportsVision returns whether the current model supports vision/image input
// This checks for known vision-capable models
func (p *OpenAICompatProvider) SupportsVision() bool {
	model := strings.ToLower(p.config.Model)

	// Known vision-capable models
	visionModels := []string{
		// OpenAI
		"gpt-4-vision", "gpt-4o", "gpt-4-turbo",
		// Google
		"gemini-pro-vision", "gemini-1.5", "gemini-2",
		// Anthropic (via compatible API)
		"claude-3", "claude-3.5",
		// Ollama vision models
		"llava", "bakllava", "llava-llama3", "llava-phi3",
		"moondream", "cogvlm", "minicpm-v",
		// Other vision models
		"qwen-vl", "qwen2-vl", "internvl",
	}

	for _, vm := range visionModels {
		if strings.Contains(model, vm) {
			return true
		}
	}

	return false
}

// IsAvailable checks if the provider is available
func (p *OpenAICompatProvider) IsAvailable(ctx context.Context) bool {
	modelsURL := strings.TrimSuffix(p.config.Endpoint, "/") + "/models"
	req, err := http.NewRequestWithContext(ctx, "GET", modelsURL, nil)
	if err != nil {
		return false
	}

	if p.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// OpenAI API request/response types
type openAIChatRequest struct {
	Model       string              `json:"model"`
	Messages    []openAIChatMessage `json:"messages"`
	MaxTokens   int                 `json:"max_tokens,omitempty"`
	Temperature float64             `json:"temperature,omitempty"`
	TopP        float64             `json:"top_p,omitempty"`
	Stop        []string            `json:"stop,omitempty"`
	Stream      bool                `json:"stream,omitempty"`
}

type openAIChatMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // Can be string or []openAIContentPart for vision
}

// openAIContentPart represents a part of multimodal content
type openAIContentPart struct {
	Type     string              `json:"type"` // "text" or "image_url"
	Text     string              `json:"text,omitempty"`
	ImageURL *openAIImageURL     `json:"image_url,omitempty"`
}

type openAIImageURL struct {
	URL    string `json:"url"`              // data:image/jpeg;base64,... or https://...
	Detail string `json:"detail,omitempty"` // "low", "high", "auto"
}

type openAIChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type openAIStreamChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason,omitempty"`
	} `json:"choices"`
}

// Complete sends a completion request
func (p *OpenAICompatProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	// Determine the HTTP client to use
	// For requests with custom timeout or vision content, use a dedicated client
	client := p.client

	// Check if this request has vision content (images)
	hasVision := false
	for _, m := range req.Messages {
		if len(m.Images) > 0 {
			hasVision = true
			break
		}
	}

	// For vision requests or custom timeout, create a client with appropriate timeout
	if req.Timeout > 0 || hasVision {
		timeout := req.Timeout
		if timeout == 0 {
			timeout = 180 * time.Second // Default vision timeout
		}
		client = &http.Client{
			Timeout: timeout,
		}
		// Also apply to context for other timeout mechanisms
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	// Build OpenAI request with vision support
	messages := make([]openAIChatMessage, len(req.Messages))
	for i, m := range req.Messages {
		// Check if message has images
		if len(m.Images) > 0 {
			// Build multimodal content
			var parts []openAIContentPart

			// Add text content first
			if m.Content != "" {
				parts = append(parts, openAIContentPart{
					Type: "text",
					Text: m.Content,
				})
			}

			// Add images
			for _, img := range m.Images {
				var imageURL string
				if img.Type == "url" {
					imageURL = img.Data
				} else {
					// Base64 format - ensure it has the data URL prefix
					if strings.HasPrefix(img.Data, "data:") {
						imageURL = img.Data
					} else {
						format := img.Format
						if format == "" {
							format = "jpeg"
						}
						imageURL = fmt.Sprintf("data:image/%s;base64,%s", format, img.Data)
					}
				}

				parts = append(parts, openAIContentPart{
					Type: "image_url",
					ImageURL: &openAIImageURL{
						URL:    imageURL,
						Detail: "auto",
					},
				})
			}

			messages[i] = openAIChatMessage{
				Role:    m.Role,
				Content: parts,
			}
		} else {
			// Plain text message
			messages[i] = openAIChatMessage{
				Role:    m.Role,
				Content: m.Content,
			}
		}
	}

	model := req.Model
	if model == "" {
		model = p.config.Model
	}

	openAIReq := openAIChatRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stop:        req.Stop,
		Stream:      false,
	}

	body, err := json.Marshal(openAIReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	chatURL := strings.TrimSuffix(p.config.Endpoint, "/") + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", chatURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if p.config.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var openAIResp openAIChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &CompletionResponse{
		ID:      openAIResp.ID,
		Model:   openAIResp.Model,
		Content: openAIResp.Choices[0].Message.Content,
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     openAIResp.Usage.PromptTokens,
			CompletionTokens: openAIResp.Usage.CompletionTokens,
			TotalTokens:      openAIResp.Usage.TotalTokens,
		},
		FinishReason: openAIResp.Choices[0].FinishReason,
	}, nil
}

// CompleteStream sends a streaming completion request
func (p *OpenAICompatProvider) CompleteStream(ctx context.Context, req *CompletionRequest) (<-chan StreamChunk, error) {
	// Build OpenAI request
	messages := make([]openAIChatMessage, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = openAIChatMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	model := req.Model
	if model == "" {
		model = p.config.Model
	}

	openAIReq := openAIChatRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stop:        req.Stop,
		Stream:      true,
	}

	body, err := json.Marshal(openAIReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	chatURL := strings.TrimSuffix(p.config.Endpoint, "/") + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", chatURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	if p.config.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	// Create channel for streaming response
	ch := make(chan StreamChunk, 100)

	go func() {
		defer close(ch)
		defer resp.Body.Close()

		reader := bufio.NewReader(resp.Body)
		for {
			select {
			case <-ctx.Done():
				ch <- StreamChunk{Error: ctx.Err(), Done: true}
				return
			default:
			}

			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					ch <- StreamChunk{Error: err, Done: true}
				} else {
					ch <- StreamChunk{Done: true}
				}
				return
			}

			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// SSE format: data: {...}
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				ch <- StreamChunk{Done: true}
				return
			}

			var chunk openAIStreamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue // Skip malformed chunks
			}

			if len(chunk.Choices) > 0 {
				content := chunk.Choices[0].Delta.Content
				done := chunk.Choices[0].FinishReason != ""
				ch <- StreamChunk{
					Content: content,
					Done:    done,
				}
				if done {
					return
				}
			}
		}
	}()

	return ch, nil
}

// ListModels returns available models
func (p *OpenAICompatProvider) ListModels(ctx context.Context) ([]string, error) {
	modelsURL := strings.TrimSuffix(p.config.Endpoint, "/") + "/models"
	req, err := http.NewRequestWithContext(ctx, "GET", modelsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if p.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	models := make([]string, len(result.Data))
	for i, m := range result.Data {
		models[i] = m.ID
	}

	return models, nil
}

// ========================================
// Claude Provider - Anthropic Claude API
// ========================================

// ClaudeProviderConfig configures the Claude provider
type ClaudeProviderConfig struct {
	APIKey  string
	Model   string
	Timeout time.Duration
}

// ClaudeProvider implements LLMProvider for Anthropic Claude API
type ClaudeProvider struct {
	config ClaudeProviderConfig
	client *http.Client
}

// NewClaudeProvider creates a new Claude provider
func NewClaudeProvider(config ClaudeProviderConfig) *ClaudeProvider {
	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second
	}
	if config.Model == "" {
		config.Model = "claude-3-haiku-20240307"
	}

	return &ClaudeProvider{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Name returns the provider name
func (p *ClaudeProvider) Name() string {
	return "Claude"
}

// Model returns the current model name
func (p *ClaudeProvider) Model() string {
	return p.config.Model
}

// Endpoint returns the API endpoint
func (p *ClaudeProvider) Endpoint() string {
	return "https://api.anthropic.com/v1"
}

// IsAvailable checks if the provider is available
func (p *ClaudeProvider) IsAvailable(ctx context.Context) bool {
	// Just check if API key is configured
	return p.config.APIKey != ""
}

// SupportsVision returns whether the current model supports vision/image input
// Claude 3 and later models support vision
func (p *ClaudeProvider) SupportsVision() bool {
	model := strings.ToLower(p.config.Model)
	// Claude 3 and later models support vision
	return strings.Contains(model, "claude-3") || strings.Contains(model, "claude-3.5")
}

// Claude API request/response types
type claudeRequest struct {
	Model       string           `json:"model"`
	Messages    []claudeMessage  `json:"messages"`
	MaxTokens   int              `json:"max_tokens"`
	System      string           `json:"system,omitempty"`
	Temperature float64          `json:"temperature,omitempty"`
	TopP        float64          `json:"top_p,omitempty"`
	Stop        []string         `json:"stop_sequences,omitempty"`
	Stream      bool             `json:"stream,omitempty"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeResponse struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Role         string `json:"role"`
	Content      []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence,omitempty"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// Complete sends a completion request
func (p *ClaudeProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	var systemMsg string
	var messages []claudeMessage

	for _, m := range req.Messages {
		if m.Role == "system" {
			systemMsg = m.Content
		} else {
			messages = append(messages, claudeMessage{
				Role:    m.Role,
				Content: m.Content,
			})
		}
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	claudeReq := claudeRequest{
		Model:       p.config.Model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		System:      systemMsg,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stop:        req.Stop,
		Stream:      false,
	}

	body, err := json.Marshal(claudeReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var claudeResp claudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var content string
	for _, c := range claudeResp.Content {
		if c.Type == "text" {
			content += c.Text
		}
	}

	return &CompletionResponse{
		ID:      claudeResp.ID,
		Model:   claudeResp.Model,
		Content: content,
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     claudeResp.Usage.InputTokens,
			CompletionTokens: claudeResp.Usage.OutputTokens,
			TotalTokens:      claudeResp.Usage.InputTokens + claudeResp.Usage.OutputTokens,
		},
		FinishReason: claudeResp.StopReason,
	}, nil
}

// CompleteStream sends a streaming completion request
func (p *ClaudeProvider) CompleteStream(ctx context.Context, req *CompletionRequest) (<-chan StreamChunk, error) {
	var systemMsg string
	var messages []claudeMessage

	for _, m := range req.Messages {
		if m.Role == "system" {
			systemMsg = m.Content
		} else {
			messages = append(messages, claudeMessage{
				Role:    m.Role,
				Content: m.Content,
			})
		}
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	claudeReq := claudeRequest{
		Model:       p.config.Model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		System:      systemMsg,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stop:        req.Stop,
		Stream:      true,
	}

	body, err := json.Marshal(claudeReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("x-api-key", p.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	ch := make(chan StreamChunk, 100)

	go func() {
		defer close(ch)
		defer resp.Body.Close()

		reader := bufio.NewReader(resp.Body)
		for {
			select {
			case <-ctx.Done():
				ch <- StreamChunk{Error: ctx.Err(), Done: true}
				return
			default:
			}

			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					ch <- StreamChunk{Error: err, Done: true}
				} else {
					ch <- StreamChunk{Done: true}
				}
				return
			}

			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")

			var event struct {
				Type  string `json:"type"`
				Index int    `json:"index,omitempty"`
				Delta struct {
					Type string `json:"type,omitempty"`
					Text string `json:"text,omitempty"`
				} `json:"delta,omitempty"`
			}

			if err := json.Unmarshal([]byte(data), &event); err != nil {
				continue
			}

			switch event.Type {
			case "content_block_delta":
				if event.Delta.Type == "text_delta" {
					ch <- StreamChunk{Content: event.Delta.Text}
				}
			case "message_stop":
				ch <- StreamChunk{Done: true}
				return
			}
		}
	}()

	return ch, nil
}

// ListModels returns available Claude models
func (p *ClaudeProvider) ListModels(ctx context.Context) ([]string, error) {
	// Claude doesn't have a models endpoint, return known models
	return []string{
		"claude-3-opus-20240229",
		"claude-3-sonnet-20240229",
		"claude-3-haiku-20240307",
		"claude-3-5-sonnet-20241022",
	}, nil
}
