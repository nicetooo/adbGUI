package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// ========================================
// EventPipeline - 事件处理管道
// ========================================

type EventPipeline struct {
	ctx      context.Context
	wailsCtx context.Context
	store    *EventStore
	mcpMode  bool // MCP mode - skip Wails EventsEmit calls

	// 事件通道
	eventChan chan UnifiedEvent

	// Session 管理
	sessions      map[string]*SessionState // sessionId -> state
	deviceSession map[string]string        // deviceId -> active sessionId
	sessionMu     sync.RWMutex

	// 批量发送到前端
	frontendBuffer   []UnifiedEvent
	frontendBufferMu sync.Mutex
	frontendTicker   *time.Ticker

	// 时间索引缓存 (LRU)
	timeIndexCache *TimeIndexLRUCache

	// 背压控制
	backpressure *BackpressureController

	// 插件管理器
	pluginManager *PluginManager

	// 停止信号
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// SessionState 会话运行时状态 (内存中)
type SessionState struct {
	Session      *DeviceSession
	StartTime    int64
	EventCount   int64
	LastEventAt  int64
	RecentEvents *RingBuffer // 最近事件缓冲
}

// RingBuffer 环形缓冲区
type RingBuffer struct {
	data  []UnifiedEvent
	size  int
	head  int
	count int
	mu    sync.RWMutex
}

func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		data: make([]UnifiedEvent, size),
		size: size,
	}
}

func (r *RingBuffer) Push(event UnifiedEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.data[r.head] = event
	r.head = (r.head + 1) % r.size
	if r.count < r.size {
		r.count++
	}
}

func (r *RingBuffer) GetRecent(n int) []UnifiedEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if n > r.count {
		n = r.count
	}
	if n == 0 {
		return nil
	}

	result := make([]UnifiedEvent, n)
	start := (r.head - n + r.size) % r.size
	for i := 0; i < n; i++ {
		result[i] = r.data[(start+i)%r.size]
	}
	return result
}

func (r *RingBuffer) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.head = 0
	r.count = 0
}

func (r *RingBuffer) Size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.count
}

// ========================================
// TimeIndexLRUCache - 时间索引 LRU 缓存
// ========================================

const DefaultTimeIndexCacheCapacity = 20 // 最多缓存 20 个 Session 的时间索引

// TimeIndexLRUCache 时间索引的 LRU 缓存
type TimeIndexLRUCache struct {
	capacity int
	cache    map[string]map[int]*TimeIndexEntry
	order    []string // 访问顺序，最近访问的在末尾
	mu       sync.RWMutex
}

// NewTimeIndexLRUCache 创建新的 LRU 缓存
func NewTimeIndexLRUCache(capacity int) *TimeIndexLRUCache {
	if capacity <= 0 {
		capacity = DefaultTimeIndexCacheCapacity
	}
	return &TimeIndexLRUCache{
		capacity: capacity,
		cache:    make(map[string]map[int]*TimeIndexEntry),
		order:    make([]string, 0, capacity),
	}
}

// Get 获取 session 的时间索引，同时更新访问顺序
func (c *TimeIndexLRUCache) Get(sessionID string) (map[int]*TimeIndexEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	index, exists := c.cache[sessionID]
	if exists {
		c.moveToEnd(sessionID)
	}
	return index, exists
}

// GetOrCreate 获取或创建 session 的时间索引
func (c *TimeIndexLRUCache) GetOrCreate(sessionID string) map[int]*TimeIndexEntry {
	c.mu.Lock()
	defer c.mu.Unlock()

	if index, exists := c.cache[sessionID]; exists {
		c.moveToEnd(sessionID)
		return index
	}

	// 创建新的
	c.evictIfNeeded()
	index := make(map[int]*TimeIndexEntry)
	c.cache[sessionID] = index
	c.order = append(c.order, sessionID)
	return index
}

// Set 设置 session 的时间索引
func (c *TimeIndexLRUCache) Set(sessionID string, index map[int]*TimeIndexEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.cache[sessionID]; exists {
		c.cache[sessionID] = index
		c.moveToEnd(sessionID)
		return
	}

	c.evictIfNeeded()
	c.cache[sessionID] = index
	c.order = append(c.order, sessionID)
}

// Delete 删除 session 的时间索引
func (c *TimeIndexLRUCache) Delete(sessionID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cache, sessionID)
	c.removeFromOrder(sessionID)
}

// GetAll 获取所有缓存的时间索引的深拷贝（用于持久化，避免与 updateTimeIndex 竞态）
func (c *TimeIndexLRUCache) GetAll() map[string]map[int]*TimeIndexEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]map[int]*TimeIndexEntry, len(c.cache))
	for sessionID, entries := range c.cache {
		copied := make(map[int]*TimeIndexEntry, len(entries))
		for sec, entry := range entries {
			e := *entry // 值拷贝
			copied[sec] = &e
		}
		result[sessionID] = copied
	}
	return result
}

// Size 返回当前缓存的 session 数量
func (c *TimeIndexLRUCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}

// moveToEnd 将 sessionID 移到访问顺序末尾（内部方法，需要持有锁）
func (c *TimeIndexLRUCache) moveToEnd(sessionID string) {
	c.removeFromOrder(sessionID)
	c.order = append(c.order, sessionID)
}

// removeFromOrder 从访问顺序中移除（内部方法，需要持有锁）
func (c *TimeIndexLRUCache) removeFromOrder(sessionID string) {
	for i, id := range c.order {
		if id == sessionID {
			c.order = append(c.order[:i], c.order[i+1:]...)
			break
		}
	}
}

// evictIfNeeded 如果超出容量则驱逐最老的条目（内部方法，需要持有锁）
func (c *TimeIndexLRUCache) evictIfNeeded() {
	for len(c.cache) >= c.capacity && len(c.order) > 0 {
		oldest := c.order[0]
		c.order = c.order[1:]
		delete(c.cache, oldest)
	}
}

// ========================================
// BackpressureController - 背压控制器
// ========================================

type BackpressureController struct {
	maxEventsPerSecond int

	windowEvents int64
	windowStart  time.Time
	levelCounts  map[EventLevel]int64

	// 采样器
	samplers map[string]*EventSampler

	// 统计
	droppedCount   int64
	sampledCount   int64
	aggregateCount int64

	mu sync.Mutex
}

type EventSampler struct {
	rate    int
	counter int64 // 虽然当前在锁内使用，但用 atomic 增强健壮性
}

func (s *EventSampler) ShouldKeep() bool {
	// 使用 atomic 确保线程安全，即使将来在锁外使用
	count := atomic.AddInt64(&s.counter, 1)
	return count%int64(s.rate) == 0
}

func NewBackpressureController(maxPerSecond int) *BackpressureController {
	return &BackpressureController{
		maxEventsPerSecond: maxPerSecond,
		windowStart:        time.Now(),
		levelCounts:        make(map[EventLevel]int64),
		samplers:           make(map[string]*EventSampler),
	}
}

// ShouldProcess 判断事件是否应该处理
func (b *BackpressureController) ShouldProcess(event UnifiedEvent) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 重置窗口
	now := time.Now()
	if now.Sub(b.windowStart) >= time.Second {
		b.windowStart = now
		b.windowEvents = 0
		b.levelCounts = make(map[EventLevel]int64)
	}

	b.windowEvents++
	b.levelCounts[event.Level]++

	// 关键事件永不丢弃
	if b.isCriticalEvent(event) {
		return true
	}

	// 未超载时全部保留
	if b.windowEvents <= int64(b.maxEventsPerSecond) {
		return true
	}

	// 中度超载 (2-5x): 采样 verbose
	if b.windowEvents <= int64(b.maxEventsPerSecond)*5 {
		if event.Level == LevelVerbose {
			sampler := b.getSampler(event, 10)
			if !sampler.ShouldKeep() {
				b.sampledCount++
				return false
			}
		}
		return true
	}

	// 重度超载 (>5x): verbose 丢弃, debug 采样
	if event.Level == LevelVerbose {
		b.droppedCount++
		return false
	}
	if event.Level == LevelDebug {
		sampler := b.getSampler(event, 5)
		if !sampler.ShouldKeep() {
			b.sampledCount++
			return false
		}
	}

	return true
}

func (b *BackpressureController) isCriticalEvent(event UnifiedEvent) bool {
	// Error/Fatal 级别
	if event.Level == LevelError || event.Level == LevelFatal {
		return true
	}
	// 网络请求
	if event.Source == SourceNetwork {
		return true
	}
	// 应用崩溃/ANR
	if event.Type == "app_crash" || event.Type == "app_anr" {
		return true
	}
	// Workflow 事件
	if event.Source == SourceWorkflow {
		return true
	}
	// 断言结果
	if event.Source == SourceAssertion {
		return true
	}
	// Session 生命周期
	if event.Type == "session_start" || event.Type == "session_end" {
		return true
	}
	return false
}

func (b *BackpressureController) getSampler(event UnifiedEvent, rate int) *EventSampler {
	key := fmt.Sprintf("%s:%s", event.Source, event.Type)
	if s, ok := b.samplers[key]; ok {
		return s
	}
	s := &EventSampler{rate: rate}
	b.samplers[key] = s
	return s
}

// GetStats 获取背压统计
func (b *BackpressureController) GetStats() map[string]int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return map[string]int64{
		"dropped":    b.droppedCount,
		"sampled":    b.sampledCount,
		"aggregated": b.aggregateCount,
	}
}

// ========================================
// EventPipeline 实现
// ========================================

// NewEventPipeline 创建事件管道
func NewEventPipeline(ctx, wailsCtx context.Context, store *EventStore, mcpMode bool) *EventPipeline {
	return &EventPipeline{
		ctx:            ctx,
		wailsCtx:       wailsCtx,
		store:          store,
		mcpMode:        mcpMode,
		eventChan:      make(chan UnifiedEvent, 10000),
		sessions:       make(map[string]*SessionState),
		deviceSession:  make(map[string]string),
		frontendBuffer: make([]UnifiedEvent, 0, 100),
		timeIndexCache: NewTimeIndexLRUCache(DefaultTimeIndexCacheCapacity),
		backpressure:   NewBackpressureController(2000),
		stopChan:       make(chan struct{}),
	}
}

// SetPluginManager 设置插件管理器
func (p *EventPipeline) SetPluginManager(pm *PluginManager) {
	p.pluginManager = pm
}

// Start 启动管道
func (p *EventPipeline) Start() {
	// 事件处理协程
	p.wg.Add(1)
	go p.processEvents()

	// 前端批量发送协程
	p.frontendTicker = time.NewTicker(500 * time.Millisecond)
	p.wg.Add(1)
	go p.frontendEmitter()

	// 时间索引持久化协程
	p.wg.Add(1)
	go p.timeIndexPersister()

	// 加载已有的活跃 Session
	p.loadActiveSessions()
}

// Stop 停止管道
func (p *EventPipeline) Stop() {
	close(p.stopChan)
	if p.frontendTicker != nil {
		p.frontendTicker.Stop()
	}

	// 等待所有协程结束
	p.wg.Wait()

	// 刷新剩余事件
	p.flushFrontendBuffer()
	p.persistTimeIndex()
}

// loadActiveSessions 加载已有的活跃 Session
func (p *EventPipeline) loadActiveSessions() {
	sessions, err := p.store.ListSessions("", 100)
	if err != nil {
		LogError("event").Err(err).Msg("Failed to load sessions")
		return
	}

	p.sessionMu.Lock()
	defer p.sessionMu.Unlock()

	for i := range sessions {
		session := &sessions[i]
		if session.Status == "active" {
			state := &SessionState{
				Session:      session,
				StartTime:    session.StartTime,
				EventCount:   int64(session.EventCount),
				RecentEvents: NewRingBuffer(1000),
			}
			p.sessions[session.ID] = state
			p.deviceSession[session.DeviceID] = session.ID
			p.timeIndexCache.GetOrCreate(session.ID)
		}
	}
}

// Emit 发送事件 (主入口)
func (p *EventPipeline) Emit(event UnifiedEvent) {
	// 自动同步 Data 到 Detail (向后兼容前端老代码)
	if len(event.Data) > 0 && len(event.Detail) == 0 {
		event.Detail = event.Data
	}

	// 背压检查
	if !p.backpressure.ShouldProcess(event) {
		return
	}

	select {
	case p.eventChan <- event:
		// 成功发送
	default:
		// 通道满了，对于非关键事件直接丢弃
		if event.Level == LevelVerbose || event.Level == LevelDebug {
			return
		}
		// 关键事件使用超时等待，避免永久阻塞 (500ms 避免 UI 卡顿)
		select {
		case p.eventChan <- event:
			// 成功发送
		case <-time.After(500 * time.Millisecond):
			// 超时，记录日志但不阻塞
			LogWarn("event_pipeline").
				Str("event_type", event.Type).
				Str("device_id", event.DeviceID).
				Str("level", string(event.Level)).
				Msg("Event channel send timeout, event dropped")
		}
	}
}

// EmitRaw 发送原始事件数据 (便捷方法)
func (p *EventPipeline) EmitRaw(deviceID string, source EventSource, eventType string,
	level EventLevel, title string, data interface{}) {

	dataBytes, err := json.Marshal(data)
	if err != nil {
		LogWarn("event_pipeline").Err(err).Str("eventType", eventType).Msg("Failed to marshal event data")
		dataBytes = []byte("{}")
	}

	event := UnifiedEvent{
		ID:        uuid.New().String(),
		DeviceID:  deviceID,
		Timestamp: time.Now().UnixMilli(),
		Source:    source,
		Type:      eventType,
		Level:     level,
		Title:     title,
		Data:      dataBytes,
	}

	// 从注册表获取 Category
	if info, ok := EventRegistry[eventType]; ok {
		event.Category = info.Category
	} else {
		event.Category = CategoryDiagnostic
	}

	p.Emit(event)
}

// processEvents 事件处理主循环
func (p *EventPipeline) processEvents() {
	defer p.wg.Done()

	for {
		select {
		case event := <-p.eventChan:
			p.processEvent(event)
		case <-p.stopChan:
			// 处理剩余事件
			for len(p.eventChan) > 0 {
				select {
				case event := <-p.eventChan:
					p.processEvent(event)
				default:
					return
				}
			}
			return
		}
	}
}

// processEvent 处理单个事件
func (p *EventPipeline) processEvent(event UnifiedEvent) {
	// 1. 尝试关联已有 Session (不自动创建)
	sessionID := p.GetActiveSessionID(event.DeviceID)

	// 2. 填充默认值
	if event.Category == "" {
		event.Category = GetCategoryForType(event.Type)
	}
	if event.Summary == "" {
		event.Summary = event.Title
	}
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	// 3. 如果没有活跃 Session，只推送到前端，不存储
	if sessionID == "" {
		// 无 Session 时仍推送到前端（实时显示）
		p.addToFrontendBuffer(event)
		return
	}

	event.SessionID = sessionID

	// 4. 获取 Session 状态并更新 (单次锁操作，优化锁粒度)
	p.sessionMu.Lock()
	state := p.sessions[sessionID]
	if state != nil {
		// 5. 计算相对时间
		event.RelativeTime = event.Timestamp - state.StartTime

		// 修正插件生成的历史事件：如果相对时间为负数，说明事件的时间戳早于 session 开始
		// 这通常发生在批量上报的历史事件中，将其设为 0 以保持时间线合理性
		if event.RelativeTime < 0 && event.Category == CategoryPlugin {
			event.RelativeTime = 0
		}

		// 6. 更新 Session 状态
		state.EventCount++
		state.LastEventAt = event.Timestamp
		state.RecentEvents.Push(event)
	}
	p.sessionMu.Unlock()

	// 7. 插件处理（生成派生事件）
	if p.pluginManager != nil {
		derivedEvents := p.pluginManager.ProcessEvent(event, sessionID)
		if len(derivedEvents) > 0 {
			log.Printf("[EventPipeline] 🔥 Plugin generated %d events for event %s (%s)", len(derivedEvents), event.ID, event.Type)
		}
		// 将派生事件递归发送到管道（会被标记，避免死循环）
		for _, derived := range derivedEvents {
			// 派生事件已由插件管理器设置好所有字段
			// 直接发送到管道，让它经过完整的处理流程
			p.Emit(derived)
		}
	} else {
		log.Printf("[EventPipeline] ⚠️ pluginManager is nil, skipping plugin processing")
	}

	// 8. 更新时间索引
	p.updateTimeIndex(event)

	// 9. 写入存储
	p.store.WriteEvent(event)

	// 10. 添加到前端缓冲
	p.addToFrontendBuffer(event)
}

// updateTimeIndex 更新时间索引
func (p *EventPipeline) updateTimeIndex(event UnifiedEvent) {
	second := int(event.RelativeTime / 1000)

	// LRU 缓存内部已有锁保护
	sessionIndex := p.timeIndexCache.GetOrCreate(event.SessionID)

	entry := sessionIndex[second]
	if entry == nil {
		entry = &TimeIndexEntry{
			Second:       second,
			EventCount:   0,
			FirstEventID: event.ID,
		}
		sessionIndex[second] = entry
	}

	entry.EventCount++
	if event.Level == LevelError || event.Level == LevelFatal {
		entry.HasError = true
	}
}

// addToFrontendBuffer 添加到前端缓冲
func (p *EventPipeline) addToFrontendBuffer(event UnifiedEvent) {
	p.frontendBufferMu.Lock()
	p.frontendBuffer = append(p.frontendBuffer, event)
	p.frontendBufferMu.Unlock()
}

// frontendEmitter 前端批量发送
func (p *EventPipeline) frontendEmitter() {
	defer p.wg.Done()

	for {
		select {
		case <-p.frontendTicker.C:
			p.flushFrontendBuffer()
		case <-p.stopChan:
			return
		}
	}
}

// flushFrontendBuffer 刷新前端缓冲
func (p *EventPipeline) flushFrontendBuffer() {
	p.frontendBufferMu.Lock()
	if len(p.frontendBuffer) == 0 {
		p.frontendBufferMu.Unlock()
		return
	}

	batch := p.frontendBuffer
	p.frontendBuffer = make([]UnifiedEvent, 0, 100)
	p.frontendBufferMu.Unlock()

	// 发送到前端 (统一使用 session-events-batch，兼容所有组件)
	if !p.mcpMode {
		wailsRuntime.EventsEmit(p.wailsCtx, "session-events-batch", batch)
	}
}

// timeIndexPersister 定期持久化时间索引
func (p *EventPipeline) timeIndexPersister() {
	defer p.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.persistTimeIndex()
		case <-p.stopChan:
			p.persistTimeIndex()
			return
		}
	}
}

// persistTimeIndex 持久化时间索引
func (p *EventPipeline) persistTimeIndex() {
	// LRU 缓存的 GetAll 方法返回快照
	allCache := p.timeIndexCache.GetAll()

	// 复制需要持久化的数据
	toSave := make(map[string][]TimeIndexEntry)
	for sessionID, entries := range allCache {
		for _, entry := range entries {
			toSave[sessionID] = append(toSave[sessionID], *entry)
		}
	}

	// 批量写入数据库
	for sessionID, entries := range toSave {
		for _, entry := range entries {
			if err := p.store.UpsertTimeIndex(sessionID, entry); err != nil {
				LogError("event").Err(err).Str("sessionId", sessionID).Msg("Failed to persist time index")
			}
		}
	}
}

// ========================================
// 公共 API
// ========================================

// StartSession 开始新 Session
func (p *EventPipeline) StartSession(deviceID, sessionType, name string, config *SessionConfig) string {
	SessionLog().Str("deviceId", deviceID).Str("type", sessionType).Str("name", name).Msg("Starting session")
	p.sessionMu.Lock()
	defer p.sessionMu.Unlock()

	// 结束旧 Session
	if oldID, ok := p.deviceSession[deviceID]; ok {
		SessionLog().Str("oldSessionId", oldID).Msg("Ending old session")
		if state := p.sessions[oldID]; state != nil {
			state.Session.EndTime = time.Now().UnixMilli()
			state.Session.Status = "completed"
			state.Session.EventCount = int(state.EventCount)
			p.store.UpdateSession(state.Session)
			if !p.mcpMode {
				wailsRuntime.EventsEmit(p.wailsCtx, "session-ended", state.Session)
			}
		}
	}

	sessionID := uuid.New().String()
	now := time.Now().UnixMilli()
	SessionLog().Str("sessionId", sessionID).Int64("startTime", now).Msg("Created new session")

	session := &DeviceSession{
		ID:        sessionID,
		DeviceID:  deviceID,
		Type:      sessionType,
		Name:      name,
		StartTime: now,
		Status:    "active",
		Metadata:  make(map[string]any),
	}

	// 设置配置
	if config != nil {
		session.Config = *config
	}

	state := &SessionState{
		Session:      session,
		StartTime:    now,
		RecentEvents: NewRingBuffer(1000),
	}

	p.sessions[sessionID] = state
	p.deviceSession[deviceID] = sessionID
	p.timeIndexCache.GetOrCreate(sessionID)

	p.store.CreateSession(session)
	SessionLog().Str("sessionId", sessionID).Msg("Session saved to store")
	if !p.mcpMode {
		wailsRuntime.EventsEmit(p.wailsCtx, "session-started", session)
	}

	// 发送 session_start 事件
	p.Emit(UnifiedEvent{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		DeviceID:  deviceID,
		Timestamp: now,
		Source:    SourceSystem,
		Category:  CategoryState,
		Type:      "session_start",
		Level:     LevelInfo,
		Title:     "Session started: " + name,
	})

	SessionLog().Str("sessionId", sessionID).Msg("Session started successfully")
	return sessionID
}

// EndSession 结束 Session
func (p *EventPipeline) EndSession(sessionID, status string) {
	p.sessionMu.Lock()
	defer p.sessionMu.Unlock()

	state := p.sessions[sessionID]
	if state == nil {
		return
	}

	now := time.Now().UnixMilli()
	state.Session.EndTime = now
	state.Session.Status = status
	state.Session.EventCount = int(state.EventCount)

	p.store.UpdateSession(state.Session)

	// 清理设备关联
	for deviceID, sid := range p.deviceSession {
		if sid == sessionID {
			delete(p.deviceSession, deviceID)
			break
		}
	}

	if !p.mcpMode {
		wailsRuntime.EventsEmit(p.wailsCtx, "session-ended", state.Session)
	}
}

// EnsureActiveSession ensures an active session exists for the device.
// If one already exists, returns its ID. Otherwise creates a new auto session.
// This replaces the old App.EnsureActiveSession() and works entirely within EventPipeline.
func (p *EventPipeline) EnsureActiveSession(deviceID string) string {
	// Fast path: check if session already exists (read lock)
	p.sessionMu.RLock()
	if sid, ok := p.deviceSession[deviceID]; ok {
		p.sessionMu.RUnlock()
		return sid
	}
	p.sessionMu.RUnlock()

	// Slow path: create auto session (uses StartSession which takes write lock)
	return p.StartSession(deviceID, "auto", "Auto Session "+time.Now().Format("15:04:05"), nil)
}

// GetActiveSessionID 获取设备的活跃 Session ID
func (p *EventPipeline) GetActiveSessionID(deviceID string) string {
	p.sessionMu.RLock()
	defer p.sessionMu.RUnlock()
	return p.deviceSession[deviceID]
}

// GetActiveSession 获取设备的活跃 Session
func (p *EventPipeline) GetActiveSession(deviceID string) *DeviceSession {
	p.sessionMu.RLock()
	defer p.sessionMu.RUnlock()

	sessionID := p.deviceSession[deviceID]
	if sessionID == "" {
		return nil
	}

	if state := p.sessions[sessionID]; state != nil {
		return state.Session
	}
	return nil
}

// GetSession 获取 Session (by ID)
func (p *EventPipeline) GetSession(sessionID string) *DeviceSession {
	p.sessionMu.RLock()
	defer p.sessionMu.RUnlock()

	if state := p.sessions[sessionID]; state != nil {
		return state.Session
	}
	return nil
}

// UpdateSessionVideoPath 更新 Session 的视频路径
func (p *EventPipeline) UpdateSessionVideoPath(sessionID, videoPath string) {
	p.sessionMu.Lock()
	defer p.sessionMu.Unlock()

	if state := p.sessions[sessionID]; state != nil {
		state.Session.VideoPath = videoPath
		p.store.UpdateSession(state.Session)
	}
}

// GetRecentEvents 获取最近事件 (从内存)
func (p *EventPipeline) GetRecentEvents(sessionID string, count int) []UnifiedEvent {
	p.sessionMu.RLock()
	state := p.sessions[sessionID]
	p.sessionMu.RUnlock()

	if state == nil {
		return nil
	}

	return state.RecentEvents.GetRecent(count)
}

// SetSessionMetadata 设置 Session 元数据
func (p *EventPipeline) SetSessionMetadata(sessionID, key string, value any) {
	p.sessionMu.Lock()
	defer p.sessionMu.Unlock()

	if state := p.sessions[sessionID]; state != nil {
		if state.Session.Metadata == nil {
			state.Session.Metadata = make(map[string]any)
		}
		state.Session.Metadata[key] = value
	}
}

// GetSessionMetadata 获取 Session 元数据
func (p *EventPipeline) GetSessionMetadata(sessionID, key string) any {
	p.sessionMu.RLock()
	defer p.sessionMu.RUnlock()

	if state := p.sessions[sessionID]; state != nil {
		return state.Session.Metadata[key]
	}
	return nil
}

// GetBackpressureStats 获取背压统计
func (p *EventPipeline) GetBackpressureStats() map[string]int64 {
	return p.backpressure.GetStats()
}

// GetPipelineStats returns pipeline statistics
func (p *EventPipeline) GetPipelineStats() map[string]interface{} {
	p.sessionMu.RLock()
	sessionCount := len(p.sessions)
	deviceCount := len(p.deviceSession)
	p.sessionMu.RUnlock()

	p.frontendBufferMu.Lock()
	bufferLen := len(p.frontendBuffer)
	p.frontendBufferMu.Unlock()

	return map[string]interface{}{
		"channelLen":      len(p.eventChan),
		"channelCap":      cap(p.eventChan),
		"activeSessions":  sessionCount,
		"deviceMappings":  deviceCount,
		"frontendBuffer":  bufferLen,
		"timeIndexCached": p.timeIndexCache.Size(),
	}
}
