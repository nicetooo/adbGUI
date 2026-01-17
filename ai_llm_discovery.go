package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ========================================
// LLM Service Discovery - 本地 LLM 服务自动检测
// ========================================

// LLMServiceType represents different LLM service types
type LLMServiceType string

const (
	LLMServiceOllama   LLMServiceType = "ollama"
	LLMServiceLMStudio LLMServiceType = "lmstudio"
	LLMServiceLocalAI  LLMServiceType = "localai"
	LLMServiceCustom   LLMServiceType = "custom"
)

// LLMServiceStatus represents the status of an LLM service
type LLMServiceStatus string

const (
	LLMServiceRunning     LLMServiceStatus = "running"
	LLMServiceStopped     LLMServiceStatus = "stopped"
	LLMServiceNotFound    LLMServiceStatus = "not_found"
	LLMServiceUnreachable LLMServiceStatus = "unreachable"
)

// LocalLLMService represents a discovered local LLM service
type LocalLLMService struct {
	Name     string           `json:"name"`     // "Ollama", "LM Studio", "LocalAI"
	Type     LLMServiceType   `json:"type"`     // Service type
	Endpoint string           `json:"endpoint"` // "http://localhost:11434"
	Status   LLMServiceStatus `json:"status"`   // "running", "stopped", "not_found"
	Models   []string         `json:"models"`   // Available models
	Version  string           `json:"version"`  // Service version if available
}

// OpenAIEndpoint returns the OpenAI-compatible API endpoint
func (s *LocalLLMService) OpenAIEndpoint() string {
	switch s.Type {
	case LLMServiceOllama:
		return strings.TrimSuffix(s.Endpoint, "/") + "/v1"
	case LLMServiceLMStudio:
		return strings.TrimSuffix(s.Endpoint, "/") + "/v1"
	case LLMServiceLocalAI:
		return strings.TrimSuffix(s.Endpoint, "/") + "/v1"
	default:
		return strings.TrimSuffix(s.Endpoint, "/") + "/v1"
	}
}

// GetDefaultModel returns the default/first available model
func (s *LocalLLMService) GetDefaultModel() string {
	if len(s.Models) > 0 {
		return s.Models[0]
	}
	return ""
}

// LLMDiscovery handles discovery of local LLM services
type LLMDiscovery struct {
	// Default ports for different services
	defaultPorts map[LLMServiceType]int
	// HTTP client for probing
	client *http.Client
	// Custom endpoints to check
	customEndpoints []CustomEndpoint
	mu              sync.RWMutex
}

// CustomEndpoint represents a user-configured custom endpoint
type CustomEndpoint struct {
	Name     string         `json:"name"`
	Endpoint string         `json:"endpoint"`
	Type     LLMServiceType `json:"type"`
}

// NewLLMDiscovery creates a new LLM discovery instance
func NewLLMDiscovery() *LLMDiscovery {
	return &LLMDiscovery{
		defaultPorts: map[LLMServiceType]int{
			LLMServiceOllama:   11434,
			LLMServiceLMStudio: 1234,
			LLMServiceLocalAI:  8080,
		},
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		customEndpoints: []CustomEndpoint{},
	}
}

// AddCustomEndpoint adds a custom endpoint to discover
func (d *LLMDiscovery) AddCustomEndpoint(endpoint CustomEndpoint) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.customEndpoints = append(d.customEndpoints, endpoint)
}

// DiscoverAll discovers all available LLM services
func (d *LLMDiscovery) DiscoverAll(ctx context.Context) []LocalLLMService {
	var services []LocalLLMService
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Discover standard services in parallel
	for serviceType, port := range d.defaultPorts {
		wg.Add(1)
		go func(st LLMServiceType, p int) {
			defer wg.Done()
			endpoint := fmt.Sprintf("http://localhost:%d", p)
			service := d.probeService(ctx, st, endpoint)
			mu.Lock()
			services = append(services, service)
			mu.Unlock()
		}(serviceType, port)
	}

	// Discover custom endpoints
	d.mu.RLock()
	customEndpoints := make([]CustomEndpoint, len(d.customEndpoints))
	copy(customEndpoints, d.customEndpoints)
	d.mu.RUnlock()

	for _, custom := range customEndpoints {
		wg.Add(1)
		go func(c CustomEndpoint) {
			defer wg.Done()
			service := d.probeService(ctx, c.Type, c.Endpoint)
			service.Name = c.Name
			mu.Lock()
			services = append(services, service)
			mu.Unlock()
		}(custom)
	}

	wg.Wait()
	return services
}

// DiscoverOllama specifically discovers Ollama service
func (d *LLMDiscovery) DiscoverOllama(ctx context.Context) *LocalLLMService {
	endpoint := fmt.Sprintf("http://localhost:%d", d.defaultPorts[LLMServiceOllama])
	service := d.probeService(ctx, LLMServiceOllama, endpoint)
	return &service
}

// DiscoverLMStudio specifically discovers LM Studio service
func (d *LLMDiscovery) DiscoverLMStudio(ctx context.Context) *LocalLLMService {
	endpoint := fmt.Sprintf("http://localhost:%d", d.defaultPorts[LLMServiceLMStudio])
	service := d.probeService(ctx, LLMServiceLMStudio, endpoint)
	return &service
}

// DiscoverLocalAI specifically discovers LocalAI service
func (d *LLMDiscovery) DiscoverLocalAI(ctx context.Context) *LocalLLMService {
	endpoint := fmt.Sprintf("http://localhost:%d", d.defaultPorts[LLMServiceLocalAI])
	service := d.probeService(ctx, LLMServiceLocalAI, endpoint)
	return &service
}

// probeService probes a specific endpoint for LLM service
func (d *LLMDiscovery) probeService(ctx context.Context, serviceType LLMServiceType, endpoint string) LocalLLMService {
	service := LocalLLMService{
		Name:     d.getServiceName(serviceType),
		Type:     serviceType,
		Endpoint: endpoint,
		Status:   LLMServiceNotFound,
		Models:   []string{},
	}

	switch serviceType {
	case LLMServiceOllama:
		d.probeOllama(ctx, &service)
	case LLMServiceLMStudio:
		d.probeLMStudio(ctx, &service)
	case LLMServiceLocalAI:
		d.probeLocalAI(ctx, &service)
	default:
		d.probeGenericOpenAI(ctx, &service)
	}

	return service
}

// probeOllama probes Ollama service
func (d *LLMDiscovery) probeOllama(ctx context.Context, service *LocalLLMService) {
	// Try Ollama-specific API first: GET /api/tags
	tagsURL := strings.TrimSuffix(service.Endpoint, "/") + "/api/tags"
	req, err := http.NewRequestWithContext(ctx, "GET", tagsURL, nil)
	if err != nil {
		service.Status = LLMServiceUnreachable
		return
	}

	resp, err := d.client.Do(req)
	if err != nil {
		service.Status = LLMServiceUnreachable
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		service.Status = LLMServiceRunning

		// Parse models
		var result struct {
			Models []struct {
				Name string `json:"name"`
			} `json:"models"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
			for _, m := range result.Models {
				service.Models = append(service.Models, m.Name)
			}
		}
		return
	}

	// Try OpenAI-compatible endpoint as fallback
	d.probeGenericOpenAI(ctx, service)
}

// probeLMStudio probes LM Studio service
func (d *LLMDiscovery) probeLMStudio(ctx context.Context, service *LocalLLMService) {
	// LM Studio uses OpenAI-compatible API
	d.probeGenericOpenAI(ctx, service)
}

// probeLocalAI probes LocalAI service
func (d *LLMDiscovery) probeLocalAI(ctx context.Context, service *LocalLLMService) {
	// LocalAI uses OpenAI-compatible API
	d.probeGenericOpenAI(ctx, service)
}

// probeGenericOpenAI probes using OpenAI-compatible /v1/models endpoint
func (d *LLMDiscovery) probeGenericOpenAI(ctx context.Context, service *LocalLLMService) {
	modelsURL := strings.TrimSuffix(service.Endpoint, "/") + "/v1/models"
	req, err := http.NewRequestWithContext(ctx, "GET", modelsURL, nil)
	if err != nil {
		service.Status = LLMServiceUnreachable
		return
	}

	resp, err := d.client.Do(req)
	if err != nil {
		service.Status = LLMServiceUnreachable
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		service.Status = LLMServiceRunning

		// Parse OpenAI models response
		var result struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
			for _, m := range result.Data {
				service.Models = append(service.Models, m.ID)
			}
		}
		return
	}

	service.Status = LLMServiceStopped
}

// getServiceName returns a human-readable name for the service type
func (d *LLMDiscovery) getServiceName(serviceType LLMServiceType) string {
	switch serviceType {
	case LLMServiceOllama:
		return "Ollama"
	case LLMServiceLMStudio:
		return "LM Studio"
	case LLMServiceLocalAI:
		return "LocalAI"
	case LLMServiceCustom:
		return "Custom"
	default:
		return string(serviceType)
	}
}

// ProbeEndpoint probes a specific endpoint and returns the service info
func (d *LLMDiscovery) ProbeEndpoint(ctx context.Context, endpoint string, serviceType LLMServiceType) LocalLLMService {
	return d.probeService(ctx, serviceType, endpoint)
}

// GetDefaultPort returns the default port for a service type
func (d *LLMDiscovery) GetDefaultPort(serviceType LLMServiceType) int {
	if port, ok := d.defaultPorts[serviceType]; ok {
		return port
	}
	return 0
}
