import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

// Wails runtime bindings
const GetAIServiceInfo = (window as any).go?.main?.App?.GetAIServiceInfo;
const GetAIConfig = (window as any).go?.main?.App?.GetAIConfig;
const SetAIEnabled = (window as any).go?.main?.App?.SetAIEnabled;
const SetAIPreferredSource = (window as any).go?.main?.App?.SetAIPreferredSource;
const DiscoverLLMServices = (window as any).go?.main?.App?.DiscoverLLMServices;
const SetOpenAIConfig = (window as any).go?.main?.App?.SetOpenAIConfig;
const SetClaudeConfig = (window as any).go?.main?.App?.SetClaudeConfig;
const SetCustomProviderConfig = (window as any).go?.main?.App?.SetCustomProviderConfig;
const SwitchAIProvider = (window as any).go?.main?.App?.SwitchAIProvider;
const TestAIProvider = (window as any).go?.main?.App?.TestAIProvider;
const GetAvailableModels = (window as any).go?.main?.App?.GetAvailableModels;
const SetAIFeature = (window as any).go?.main?.App?.SetAIFeature;
const AddCustomLLMEndpoint = (window as any).go?.main?.App?.AddCustomLLMEndpoint;
const RemoveCustomLLMEndpoint = (window as any).go?.main?.App?.RemoveCustomLLMEndpoint;
const RefreshAIProviders = (window as any).go?.main?.App?.RefreshAIProviders;
const AIComplete = (window as any).go?.main?.App?.AIComplete;
const AIGenerateWorkflow = (window as any).go?.main?.App?.AIGenerateWorkflow;
const AIParseNaturalQuery = (window as any).go?.main?.App?.AIParseNaturalQuery;
const AIAnalyzeLog = (window as any).go?.main?.App?.AIAnalyzeLog;
const AIAnalyzeCrash = (window as any).go?.main?.App?.AIAnalyzeCrash;
const AISuggestAssertions = (window as any).go?.main?.App?.AISuggestAssertions;

// Types
export type LLMSource = 'auto' | 'local' | 'online';

export interface AIProviderInfo {
  name: string;
  model: string;
  endpoint: string;
  type: 'local' | 'online';
}

export interface DiscoveredService {
  name: string;
  type: string;
  endpoint: string;
  status: 'running' | 'stopped' | 'not_found' | 'unreachable';
  models: string[];
}

export interface ProviderConfig {
  enabled: boolean;
  apiKey?: string;
  endpoint?: string;
  model?: string;
}

export interface AIFeaturesConfig {
  logAnalysis: boolean;
  naturalSearch: boolean;
  workflowGeneration: boolean;
  workflowAI: boolean;
  crashAnalysis: boolean;
  assertionGen: boolean;
  videoAnalysis: boolean;
}

export interface AIConfig {
  enabled: boolean;
  preferredSource: LLMSource;
  localServices?: {
    autoDetect: boolean;
    customEndpoints?: Array<{
      name: string;
      endpoint: string;
      type: string;
    }>;
  };
  onlineProviders?: {
    openai: ProviderConfig;
    claude: ProviderConfig;
    custom: ProviderConfig;
  };
  activeProvider?: {
    type: string;
    name: string;
    endpoint: string;
    model: string;
  };
  features?: AIFeaturesConfig;
}

export interface AIServiceInfo {
  status: 'initializing' | 'ready' | 'no_provider' | 'error' | 'disabled';
  enabled: boolean;
  provider?: AIProviderInfo;
  error?: string;
  features?: AIFeaturesConfig;
}

interface AIState {
  // Service status
  serviceInfo: AIServiceInfo | null;
  config: AIConfig | null;

  // Discovered services
  discoveredServices: DiscoveredService[];
  isDiscovering: boolean;

  // Loading states
  isLoading: boolean;
  isSaving: boolean;

  // Error state
  error: string | null;

  // Actions
  loadServiceInfo: () => Promise<void>;
  loadConfig: () => Promise<void>;
  setEnabled: (enabled: boolean) => Promise<void>;
  setPreferredSource: (source: LLMSource) => Promise<void>;
  discoverServices: () => Promise<void>;

  // Provider configuration
  setOpenAIConfig: (apiKey: string, model: string, enabled: boolean) => Promise<void>;
  setClaudeConfig: (apiKey: string, model: string, enabled: boolean) => Promise<void>;
  setCustomProviderConfig: (endpoint: string, apiKey: string, model: string, enabled: boolean) => Promise<void>;

  // Provider switching
  switchProvider: (providerType: string, config: Record<string, string>) => Promise<void>;
  testProvider: (providerType: string, config: Record<string, string>) => Promise<{ success: boolean; message: string }>;
  getAvailableModels: (providerType: string, config: Record<string, string>) => Promise<string[]>;

  // Feature toggles
  setFeature: (feature: string, enabled: boolean) => Promise<void>;

  // Custom endpoints
  addCustomEndpoint: (name: string, endpoint: string, serviceType: string) => Promise<void>;
  removeCustomEndpoint: (name: string) => Promise<void>;

  // Refresh
  refreshProviders: () => Promise<void>;

  // AI completion (for testing)
  complete: (messages: Array<{ role: string; content: string }>, options?: Record<string, any>) => Promise<string>;

  // AI analysis features
  generateWorkflow: (sessionID: string, config?: any) => Promise<any>;
  parseNaturalQuery: (query: string, sessionID: string) => Promise<any>;
  analyzeLog: (tag: string, message: string, level: string) => Promise<any>;
  analyzeCrash: (crashEventID: string, sessionID: string) => Promise<any>;
  suggestAssertions: (sessionID: string) => Promise<any[]>;
}

export const useAIStore = create<AIState>()(
  immer((set, get) => ({
    serviceInfo: null,
    config: null,
    discoveredServices: [],
    isDiscovering: false,
    isLoading: false,
    isSaving: false,
    error: null,

    loadServiceInfo: async () => {
      if (!GetAIServiceInfo) return;

      set((state) => {
        state.isLoading = true;
        state.error = null;
      });

      try {
        const info = await GetAIServiceInfo();
        set((state) => {
          state.serviceInfo = info;
          state.isLoading = false;
        });
      } catch (err) {
        set((state) => {
          state.error = String(err);
          state.isLoading = false;
        });
      }
    },

    loadConfig: async () => {
      if (!GetAIConfig) return;

      set((state) => {
        state.isLoading = true;
        state.error = null;
      });

      try {
        const config = await GetAIConfig();
        set((state) => {
          state.config = config;
          state.isLoading = false;
        });
      } catch (err) {
        set((state) => {
          state.error = String(err);
          state.isLoading = false;
        });
      }
    },

    setEnabled: async (enabled: boolean) => {
      if (!SetAIEnabled) return;

      set((state) => {
        state.isSaving = true;
        state.error = null;
      });

      try {
        await SetAIEnabled(enabled);
        // Reload config and service info
        await get().loadConfig();
        await get().loadServiceInfo();
        set((state) => {
          state.isSaving = false;
        });
      } catch (err) {
        set((state) => {
          state.error = String(err);
          state.isSaving = false;
        });
      }
    },

    setPreferredSource: async (source: LLMSource) => {
      if (!SetAIPreferredSource) return;

      set((state) => {
        state.isSaving = true;
        state.error = null;
      });

      try {
        await SetAIPreferredSource(source);
        await get().loadConfig();
        set((state) => {
          state.isSaving = false;
        });
      } catch (err) {
        set((state) => {
          state.error = String(err);
          state.isSaving = false;
        });
      }
    },

    discoverServices: async () => {
      if (!DiscoverLLMServices) return;

      set((state) => {
        state.isDiscovering = true;
        state.error = null;
      });

      try {
        const services = await DiscoverLLMServices();
        set((state) => {
          state.discoveredServices = services || [];
          state.isDiscovering = false;
        });
      } catch (err) {
        set((state) => {
          state.error = String(err);
          state.isDiscovering = false;
        });
      }
    },

    setOpenAIConfig: async (apiKey: string, model: string, enabled: boolean) => {
      if (!SetOpenAIConfig) return;

      set((state) => {
        state.isSaving = true;
        state.error = null;
      });

      try {
        await SetOpenAIConfig(apiKey, model, enabled);
        await get().loadConfig();
        set((state) => {
          state.isSaving = false;
        });
      } catch (err) {
        set((state) => {
          state.error = String(err);
          state.isSaving = false;
        });
      }
    },

    setClaudeConfig: async (apiKey: string, model: string, enabled: boolean) => {
      if (!SetClaudeConfig) return;

      set((state) => {
        state.isSaving = true;
        state.error = null;
      });

      try {
        await SetClaudeConfig(apiKey, model, enabled);
        await get().loadConfig();
        set((state) => {
          state.isSaving = false;
        });
      } catch (err) {
        set((state) => {
          state.error = String(err);
          state.isSaving = false;
        });
      }
    },

    setCustomProviderConfig: async (endpoint: string, apiKey: string, model: string, enabled: boolean) => {
      if (!SetCustomProviderConfig) return;

      set((state) => {
        state.isSaving = true;
        state.error = null;
      });

      try {
        await SetCustomProviderConfig(endpoint, apiKey, model, enabled);
        await get().loadConfig();
        set((state) => {
          state.isSaving = false;
        });
      } catch (err) {
        set((state) => {
          state.error = String(err);
          state.isSaving = false;
        });
      }
    },

    switchProvider: async (providerType: string, config: Record<string, string>) => {
      if (!SwitchAIProvider) return;

      set((state) => {
        state.isSaving = true;
        state.error = null;
      });

      try {
        await SwitchAIProvider(providerType, config);
        await get().loadServiceInfo();
        await get().loadConfig();
        set((state) => {
          state.isSaving = false;
        });
      } catch (err) {
        set((state) => {
          state.error = String(err);
          state.isSaving = false;
        });
        throw err;
      }
    },

    testProvider: async (providerType: string, config: Record<string, string>) => {
      if (!TestAIProvider) {
        return { success: false, message: 'API not available' };
      }

      try {
        const [success, message] = await TestAIProvider(providerType, config);
        return { success, message };
      } catch (err) {
        return { success: false, message: String(err) };
      }
    },

    getAvailableModels: async (providerType: string, config: Record<string, string>) => {
      if (!GetAvailableModels) return [];

      try {
        const models = await GetAvailableModels(providerType, config);
        return models || [];
      } catch (err) {
        console.error('Failed to get available models:', err);
        return [];
      }
    },

    setFeature: async (feature: string, enabled: boolean) => {
      if (!SetAIFeature) return;

      set((state) => {
        state.isSaving = true;
        state.error = null;
      });

      try {
        await SetAIFeature(feature, enabled);
        await get().loadConfig();
        await get().loadServiceInfo();
        set((state) => {
          state.isSaving = false;
        });
      } catch (err) {
        set((state) => {
          state.error = String(err);
          state.isSaving = false;
        });
      }
    },

    addCustomEndpoint: async (name: string, endpoint: string, serviceType: string) => {
      if (!AddCustomLLMEndpoint) return;

      set((state) => {
        state.isSaving = true;
        state.error = null;
      });

      try {
        await AddCustomLLMEndpoint(name, endpoint, serviceType);
        await get().loadConfig();
        set((state) => {
          state.isSaving = false;
        });
      } catch (err) {
        set((state) => {
          state.error = String(err);
          state.isSaving = false;
        });
      }
    },

    removeCustomEndpoint: async (name: string) => {
      if (!RemoveCustomLLMEndpoint) return;

      set((state) => {
        state.isSaving = true;
        state.error = null;
      });

      try {
        await RemoveCustomLLMEndpoint(name);
        await get().loadConfig();
        set((state) => {
          state.isSaving = false;
        });
      } catch (err) {
        set((state) => {
          state.error = String(err);
          state.isSaving = false;
        });
      }
    },

    refreshProviders: async () => {
      if (!RefreshAIProviders) return;

      set((state) => {
        state.isLoading = true;
        state.error = null;
      });

      try {
        await RefreshAIProviders();
        await get().loadServiceInfo();
        await get().discoverServices();
        set((state) => {
          state.isLoading = false;
        });
      } catch (err) {
        set((state) => {
          state.error = String(err);
          state.isLoading = false;
        });
      }
    },

    complete: async (messages: Array<{ role: string; content: string }>, options?: Record<string, any>) => {
      if (!AIComplete) {
        throw new Error('AI completion not available');
      }

      try {
        const result = await AIComplete(messages, options || {});
        return result;
      } catch (err) {
        throw new Error(String(err));
      }
    },

    generateWorkflow: async (sessionID: string, config?: any) => {
      if (!AIGenerateWorkflow) {
        throw new Error('Workflow generation not available');
      }

      try {
        const result = await AIGenerateWorkflow(sessionID, config || null);
        return result;
      } catch (err) {
        throw new Error(String(err));
      }
    },

    parseNaturalQuery: async (query: string, sessionID: string) => {
      if (!AIParseNaturalQuery) {
        throw new Error('Natural query parsing not available');
      }

      try {
        const result = await AIParseNaturalQuery(query, sessionID);
        return result;
      } catch (err) {
        throw new Error(String(err));
      }
    },

    analyzeLog: async (tag: string, message: string, level: string) => {
      if (!AIAnalyzeLog) {
        throw new Error('Log analysis not available');
      }

      try {
        const result = await AIAnalyzeLog(tag, message, level);
        return result;
      } catch (err) {
        throw new Error(String(err));
      }
    },

    analyzeCrash: async (crashEventID: string, sessionID: string) => {
      if (!AIAnalyzeCrash) {
        throw new Error('Crash analysis not available');
      }

      try {
        const result = await AIAnalyzeCrash(crashEventID, sessionID);
        return result;
      } catch (err) {
        throw new Error(String(err));
      }
    },

    suggestAssertions: async (sessionID: string) => {
      if (!AISuggestAssertions) {
        throw new Error('Assertion suggestion not available');
      }

      try {
        const result = await AISuggestAssertions(sessionID);
        return result || [];
      } catch (err) {
        throw new Error(String(err));
      }
    },
  }))
);
