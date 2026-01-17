package main

import (
	"context"
	"sync"
	"time"
)

// ========================================
// Logcat AI Analyzer - 日志智能分析
// ========================================

// LogcatAnalyzerConfig 日志分析配置
type LogcatAnalyzerConfig struct {
	Enabled           bool    `json:"enabled"`
	AutoAnalyze       bool    `json:"autoAnalyze"`       // 自动分析高级别日志
	MinLevel          string  `json:"minLevel"`          // 最低分析级别 (error, warn, info)
	BatchSize         int     `json:"batchSize"`         // 批量分析大小
	AnalysisInterval  int     `json:"analysisInterval"`  // 分析间隔（秒）
	CacheResults      bool    `json:"cacheResults"`      // 缓存分析结果
	MaxCacheSize      int     `json:"maxCacheSize"`      // 最大缓存条目
}

// DefaultLogcatAnalyzerConfig 默认配置
func DefaultLogcatAnalyzerConfig() LogcatAnalyzerConfig {
	return LogcatAnalyzerConfig{
		Enabled:          true,
		AutoAnalyze:      true,
		MinLevel:         "error",
		BatchSize:        10,
		AnalysisInterval: 5,
		CacheResults:     true,
		MaxCacheSize:     1000,
	}
}

// LogcatAnalyzer 日志分析器
type LogcatAnalyzer struct {
	app       *App
	config    LogcatAnalyzerConfig
	queue     []LogAnalysisRequest
	queueMu   sync.Mutex
	cache     map[string]*LogAnalysisResult
	cacheMu   sync.RWMutex
	stopChan  chan struct{}
	isRunning bool
	runMu     sync.Mutex
}

// LogAnalysisRequest 日志分析请求
type LogAnalysisRequest struct {
	DeviceID  string `json:"deviceId"`
	SessionID string `json:"sessionId"`
	EventID   string `json:"eventId"`
	Tag       string `json:"tag"`
	Message   string `json:"message"`
	Level     string `json:"level"`
	Timestamp int64  `json:"timestamp"`
}

// NewLogcatAnalyzer 创建日志分析器
func NewLogcatAnalyzer(app *App, config LogcatAnalyzerConfig) *LogcatAnalyzer {
	return &LogcatAnalyzer{
		app:      app,
		config:   config,
		queue:    make([]LogAnalysisRequest, 0),
		cache:    make(map[string]*LogAnalysisResult),
		stopChan: make(chan struct{}),
	}
}

// Start 启动分析器
func (la *LogcatAnalyzer) Start() {
	la.runMu.Lock()
	if la.isRunning {
		la.runMu.Unlock()
		return
	}
	la.isRunning = true
	la.stopChan = make(chan struct{})
	la.runMu.Unlock()

	go la.analyzeLoop()
}

// Stop 停止分析器
func (la *LogcatAnalyzer) Stop() {
	la.runMu.Lock()
	defer la.runMu.Unlock()

	if !la.isRunning {
		return
	}

	close(la.stopChan)
	la.isRunning = false
}

// QueueAnalysis 添加日志到分析队列
func (la *LogcatAnalyzer) QueueAnalysis(req LogAnalysisRequest) {
	if !la.config.Enabled || !la.config.AutoAnalyze {
		return
	}

	// 检查级别过滤
	if !la.shouldAnalyze(req.Level) {
		return
	}

	// 检查缓存
	if la.config.CacheResults {
		la.cacheMu.RLock()
		cacheKey := la.getCacheKey(req)
		if _, exists := la.cache[cacheKey]; exists {
			la.cacheMu.RUnlock()
			return
		}
		la.cacheMu.RUnlock()
	}

	la.queueMu.Lock()
	la.queue = append(la.queue, req)
	// 限制队列大小
	if len(la.queue) > la.config.MaxCacheSize {
		la.queue = la.queue[len(la.queue)-la.config.MaxCacheSize:]
	}
	la.queueMu.Unlock()
}

// GetCachedAnalysis 获取缓存的分析结果
func (la *LogcatAnalyzer) GetCachedAnalysis(tag, message, level string) *LogAnalysisResult {
	if !la.config.CacheResults {
		return nil
	}

	la.cacheMu.RLock()
	defer la.cacheMu.RUnlock()

	cacheKey := la.getCacheKeyFromParams(tag, message, level)
	return la.cache[cacheKey]
}

// AnalyzeNow 立即分析日志
func (la *LogcatAnalyzer) AnalyzeNow(ctx context.Context, tag, message, level string) (*LogAnalysisResult, error) {
	la.app.aiServiceMu.RLock()
	aiService := la.app.aiService
	la.app.aiServiceMu.RUnlock()

	if aiService == nil || !aiService.IsReady() {
		return nil, nil
	}

	// 使用 App 的 AnalyzeLog 方法
	result, err := la.app.AnalyzeLog(tag, message, level)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	if la.config.CacheResults && result != nil {
		la.cacheMu.Lock()
		cacheKey := la.getCacheKeyFromParams(tag, message, level)
		la.cache[cacheKey] = result
		// 限制缓存大小
		if len(la.cache) > la.config.MaxCacheSize {
			// 简单策略：删除一半
			count := 0
			for k := range la.cache {
				if count >= la.config.MaxCacheSize/2 {
					break
				}
				delete(la.cache, k)
				count++
			}
		}
		la.cacheMu.Unlock()
	}

	return result, nil
}

// analyzeLoop 分析循环
func (la *LogcatAnalyzer) analyzeLoop() {
	ticker := time.NewTicker(time.Duration(la.config.AnalysisInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-la.stopChan:
			return
		case <-ticker.C:
			la.processBatch()
		}
	}
}

// processBatch 处理一批日志
func (la *LogcatAnalyzer) processBatch() {
	la.queueMu.Lock()
	if len(la.queue) == 0 {
		la.queueMu.Unlock()
		return
	}

	// 取一批日志
	batchSize := la.config.BatchSize
	if batchSize > len(la.queue) {
		batchSize = len(la.queue)
	}

	batch := la.queue[:batchSize]
	la.queue = la.queue[batchSize:]
	la.queueMu.Unlock()

	// 获取 AI 服务
	la.app.aiServiceMu.RLock()
	aiService := la.app.aiService
	la.app.aiServiceMu.RUnlock()

	if aiService == nil || !aiService.IsReady() {
		return
	}

	// 分析每条日志
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, req := range batch {
		select {
		case <-la.stopChan:
			return
		case <-ctx.Done():
			return
		default:
		}

		// 使用 App 的 AnalyzeLog 方法
		result, err := la.app.AnalyzeLog(req.Tag, req.Message, req.Level)
		if err != nil {
			continue
		}

		if result != nil && la.config.CacheResults {
			la.cacheMu.Lock()
			cacheKey := la.getCacheKey(req)
			la.cache[cacheKey] = result
			la.cacheMu.Unlock()

			// 如果检测到重要内容，发送事件
			if result.Classification == "error" || result.Severity > 0.7 {
				la.emitAnalysisResult(req, result)
			}
		}
	}
}

// shouldAnalyze 检查是否应该分析此级别的日志
func (la *LogcatAnalyzer) shouldAnalyze(level string) bool {
	levelPriority := map[string]int{
		"verbose": 0,
		"debug":   1,
		"info":    2,
		"warn":    3,
		"error":   4,
		"fatal":   5,
	}

	minPriority := levelPriority[la.config.MinLevel]
	logPriority := levelPriority[level]

	return logPriority >= minPriority
}

// getCacheKey 获取缓存键
func (la *LogcatAnalyzer) getCacheKey(req LogAnalysisRequest) string {
	return la.getCacheKeyFromParams(req.Tag, req.Message, req.Level)
}

// getCacheKeyFromParams 从参数生成缓存键
func (la *LogcatAnalyzer) getCacheKeyFromParams(tag, message, level string) string {
	// 使用 tag + level + message前50字符 作为键
	msgKey := message
	if len(msgKey) > 50 {
		msgKey = msgKey[:50]
	}
	return tag + ":" + level + ":" + msgKey
}

// emitAnalysisResult 发送分析结果事件
func (la *LogcatAnalyzer) emitAnalysisResult(req LogAnalysisRequest, result *LogAnalysisResult) {
	// 通过事件管道发送分析结果
	if la.app.eventPipeline != nil && req.SessionID != "" {
		data := map[string]interface{}{
			"originalEventId": req.EventID,
			"tag":             req.Tag,
			"message":         req.Message,
			"analysis":        result,
		}

		la.app.eventPipeline.EmitRaw(
			req.DeviceID,
			SourceLogcat,
			"logcat_analysis",
			LevelInfo,
			"AI Analysis: "+result.Classification,
			data,
		)
	}
}

// ========================================
// App Methods for Log Analysis
// ========================================

// logcatAnalyzer 存储在 App 中
var logcatAnalyzerMu sync.RWMutex
var globalLogcatAnalyzer *LogcatAnalyzer

// GetLogcatAnalyzer 获取或创建日志分析器
func (a *App) GetLogcatAnalyzer() *LogcatAnalyzer {
	logcatAnalyzerMu.RLock()
	if globalLogcatAnalyzer != nil {
		logcatAnalyzerMu.RUnlock()
		return globalLogcatAnalyzer
	}
	logcatAnalyzerMu.RUnlock()

	logcatAnalyzerMu.Lock()
	defer logcatAnalyzerMu.Unlock()

	if globalLogcatAnalyzer == nil {
		config := DefaultLogcatAnalyzerConfig()
		globalLogcatAnalyzer = NewLogcatAnalyzer(a, config)
		globalLogcatAnalyzer.Start()
	}

	return globalLogcatAnalyzer
}

// AnalyzeLogEntry 分析单条日志
func (a *App) AnalyzeLogEntry(tag, message, level string) (*LogAnalysisResult, error) {
	analyzer := a.GetLogcatAnalyzer()

	// 先检查缓存
	cached := analyzer.GetCachedAnalysis(tag, message, level)
	if cached != nil {
		return cached, nil
	}

	// 实时分析
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return analyzer.AnalyzeNow(ctx, tag, message, level)
}

// SetLogAnalysisConfig 设置日志分析配置
func (a *App) SetLogAnalysisConfig(config LogcatAnalyzerConfig) {
	logcatAnalyzerMu.Lock()
	defer logcatAnalyzerMu.Unlock()

	if globalLogcatAnalyzer != nil {
		globalLogcatAnalyzer.Stop()
		globalLogcatAnalyzer = nil
	}

	globalLogcatAnalyzer = NewLogcatAnalyzer(a, config)
	if config.Enabled && config.AutoAnalyze {
		globalLogcatAnalyzer.Start()
	}
}

// GetLogAnalysisConfig 获取日志分析配置
func (a *App) GetLogAnalysisConfig() LogcatAnalyzerConfig {
	logcatAnalyzerMu.RLock()
	defer logcatAnalyzerMu.RUnlock()

	if globalLogcatAnalyzer != nil {
		return globalLogcatAnalyzer.config
	}

	return DefaultLogcatAnalyzerConfig()
}

// QueueLogForAnalysis 将日志加入分析队列
func (a *App) QueueLogForAnalysis(deviceID, sessionID, eventID, tag, message, level string) {
	analyzer := a.GetLogcatAnalyzer()
	analyzer.QueueAnalysis(LogAnalysisRequest{
		DeviceID:  deviceID,
		SessionID: sessionID,
		EventID:   eventID,
		Tag:       tag,
		Message:   message,
		Level:     level,
		Timestamp: time.Now().UnixMilli(),
	})
}
