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
  }))
);
