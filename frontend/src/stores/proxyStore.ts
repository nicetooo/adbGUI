import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

export interface RequestLog {
  id: string;
  time: string;
  clientIp: string;
  method: string;
  url: string;
  headers?: Record<string, string[]>;
  body?: string;
  isHttps: boolean;
  statusCode?: number;
  contentType?: string;
  bodySize?: number;
  previewBody?: string;
  respHeaders?: Record<string, string[]>;
  respBody?: string;
  isWs?: boolean;
  mocked?: boolean;
  timestamp?: number;
  duration?: number;
  partialUpdate?: boolean;
}

export interface NetworkStats {
  rxSpeed: number;
  txSpeed: number;
  rxBytes: number;
  txBytes: number;
  time: number;
}

interface MockRule {
  id: string;
  urlPattern: string;
  method?: string;
  statusCode: number;
  headers: Record<string, string>;
  body: string;
  delay?: number;
  description?: string;
  enabled: boolean;
  createdAt?: number;
}

// Data for pre-filling a mock rule from another view (e.g. EventTimeline)
export interface PendingMockData {
  urlPattern: string;
  method?: string;
  statusCode: number;
  contentType?: string;
  body?: string;
  description?: string;
}

interface ProxyState {
  // 核心数据
  logs: RequestLog[];
  isRunning: boolean;
  port: number;
  localIP: string;
  netStats: NetworkStats;
  
  // 配置
  wsEnabled: boolean;
  mitmEnabled: boolean;
  latency: number | null;
  bypassPatterns: string[];
  dlLimit: number | null;
  ulLimit: number | null;
  
  // 过滤
  filterType: "ALL" | "HTTP" | "WS";
  searchText: string;
  
  // UI 状态
  selectedLog: RequestLog | null;
  detailsDrawerOpen: boolean;
  isBypassModalOpen: boolean;
  newPattern: string;
  
  // Resend 请求相关
  resendModalOpen: boolean;
  resendLoading: boolean;
  resendResponse: any;
  
  // Mock 规则相关
  mockModalOpen: boolean;
  mockRules: MockRule[];
  editingMockRule: MockRule | null;
  pendingMockData: PendingMockData | null;
  
  // 证书状态
  certTrustStatus: string | null;
  
  // AI 搜索相关
  isAIParsing: boolean;
  aiSearchText: string;
  aiPopoverOpen: boolean;
  
  // 操作方法
  addLog: (log: RequestLog) => void;
  updateLog: (id: string, updates: Partial<RequestLog>) => void;
  clearLogs: () => void;
  setProxyRunning: (running: boolean) => void;
  setPort: (port: number) => void;
  setLocalIP: (ip: string) => void;
  setNetStats: (stats: NetworkStats) => void;
  toggleWS: () => void;
  toggleMITM: () => void;
  setFilterType: (type: "ALL" | "HTTP" | "WS") => void;
  setSearchText: (text: string) => void;
  setLatency: (ms: number | null) => void;
  setSpeedLimits: (dl: number | null, ul: number | null) => void;
  addBypassPattern: (pattern: string) => void;
  removeBypassPattern: (pattern: string) => void;
  selectLog: (log: RequestLog | null) => void;
  setDetailsDrawerOpen: (open: boolean) => void;
  setBypassModalOpen: (open: boolean) => void;
  setNewPattern: (pattern: string) => void;
  
  // Resend 请求
  setResendModalOpen: (open: boolean) => void;
  setResendLoading: (loading: boolean) => void;
  setResendResponse: (response: any) => void;
  openResendModal: () => void;
  closeResendModal: () => void;
  
  // Mock 规则
  setMockModalOpen: (open: boolean) => void;
  setMockRules: (rules: MockRule[]) => void;
  setEditingMockRule: (rule: MockRule | null) => void;
  openMockModal: () => void;
  closeMockModal: () => void;
  setPendingMockData: (data: PendingMockData | null) => void;
  openMockWithPrefill: (data: PendingMockData) => void;
  
  // 证书状态
  setCertTrustStatus: (status: string | null) => void;
  
  // AI 搜索
  setIsAIParsing: (parsing: boolean) => void;
  setAiSearchText: (text: string) => void;
  setAiPopoverOpen: (open: boolean) => void;
}

export const useProxyStore = create<ProxyState>()(
  immer((set) => ({
    // 初始状态
    logs: [],
    isRunning: false,
    port: 8080,
    localIP: "",
    netStats: { rxSpeed: 0, txSpeed: 0, rxBytes: 0, txBytes: 0, time: 0 },
    
    wsEnabled: true,
    mitmEnabled: true,
    latency: null,
    bypassPatterns: [],
    dlLimit: null,
    ulLimit: null,
    
    filterType: "ALL",
    searchText: "",
    
    selectedLog: null,
    detailsDrawerOpen: false,
    isBypassModalOpen: false,
    newPattern: "",
    
    // Resend 请求
    resendModalOpen: false,
    resendLoading: false,
    resendResponse: null,
    
    // Mock 规则
    mockModalOpen: false,
    mockRules: [],
    editingMockRule: null,
    pendingMockData: null,
    
    // 证书状态
    certTrustStatus: null,
    
    // AI 搜索
    isAIParsing: false,
    aiSearchText: "",
    aiPopoverOpen: false,
    
    // 操作方法
    addLog: (log: RequestLog) => set((state: ProxyState) => {
      state.logs.unshift(log);
      if (state.logs.length > 5000) {
        state.logs.pop();
      }
    }),
    
    updateLog: (id: string, updates: Partial<RequestLog>) => set((state: ProxyState) => {
      const log = state.logs.find(l => l.id === id);
      if (log) {
        Object.assign(log, updates);
        // Also update selectedLog if it's the same log
        if (state.selectedLog && state.selectedLog.id === id) {
          Object.assign(state.selectedLog, updates);
        }
      }
    }),
    
    clearLogs: () => set({ logs: [] }),
    
    setProxyRunning: (running: boolean) => set({ isRunning: running }),
    
    setPort: (port: number) => set({ port }),
    
    setLocalIP: (ip: string) => set({ localIP: ip }),
    
    setNetStats: (stats: NetworkStats) => set({ netStats: stats }),
    
    toggleWS: () => set((state: ProxyState) => {
      state.wsEnabled = !state.wsEnabled;
    }),
    
    toggleMITM: () => set((state: ProxyState) => {
      state.mitmEnabled = !state.mitmEnabled;
    }),
    
    setFilterType: (type: "ALL" | "HTTP" | "WS") => set({ filterType: type }),
    
    setSearchText: (text: string) => set({ searchText: text }),
    
    setLatency: (ms: number | null) => set({ latency: ms }),
    
    setSpeedLimits: (dl: number | null, ul: number | null) => set({ dlLimit: dl, ulLimit: ul }),
    
    addBypassPattern: (pattern: string) => set((state: ProxyState) => {
      state.bypassPatterns.push(pattern);
    }),
    
    removeBypassPattern: (pattern: string) => set((state: ProxyState) => {
      state.bypassPatterns = state.bypassPatterns.filter(p => p !== pattern);
    }),
    
    selectLog: (log: RequestLog | null) => set({ selectedLog: log }),
    
    setDetailsDrawerOpen: (open: boolean) => set({ detailsDrawerOpen: open }),
    
    setBypassModalOpen: (open: boolean) => set({ isBypassModalOpen: open }),
    
    setNewPattern: (pattern: string) => set({ newPattern: pattern }),
    
    // Resend 请求
    setResendModalOpen: (open: boolean) => set({ resendModalOpen: open }),
    
    setResendLoading: (loading: boolean) => set({ resendLoading: loading }),
    
    setResendResponse: (response: any) => set({ resendResponse: response }),
    
    openResendModal: () => set({ resendModalOpen: true, resendResponse: null }),
    
    closeResendModal: () => set({ resendModalOpen: false, resendLoading: false, resendResponse: null }),
    
    // Mock 规则
    setMockModalOpen: (open: boolean) => set({ mockModalOpen: open }),
    
    setMockRules: (rules: MockRule[]) => set({ mockRules: rules }),
    
    setEditingMockRule: (rule: MockRule | null) => set({ editingMockRule: rule }),
    
    openMockModal: () => set({ mockModalOpen: true }),
    
    closeMockModal: () => set({ mockModalOpen: false, editingMockRule: null }),
    
    setPendingMockData: (data: PendingMockData | null) => set({ pendingMockData: data }),
    
    openMockWithPrefill: (data: PendingMockData) => set({
      pendingMockData: data,
      mockModalOpen: true,
      editingMockRule: null,
    }),
    
    // 证书状态
    setCertTrustStatus: (status: string | null) => set({ certTrustStatus: status }),
    
    // AI 搜索
    setIsAIParsing: (parsing: boolean) => set({ isAIParsing: parsing }),
    
    setAiSearchText: (text: string) => set({ aiSearchText: text }),
    
    setAiPopoverOpen: (open: boolean) => set({ aiPopoverOpen: open }),
  }))
);
