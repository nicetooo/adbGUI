package main

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

// PluginManager 插件管理器
type PluginManager struct {
	plugins map[string]*Plugin // ID -> Plugin
	mu      sync.RWMutex       // 读写锁

	store         *PluginStore   // 数据库存储
	eventPipeline *EventPipeline // 事件管道引用
}

// NewPluginManager 创建插件管理器
func NewPluginManager(store *PluginStore, eventPipeline *EventPipeline) *PluginManager {
	return &PluginManager{
		plugins:       make(map[string]*Plugin),
		store:         store,
		eventPipeline: eventPipeline,
	}
}

// LoadAllPlugins 从数据库加载所有插件
func (pm *PluginManager) LoadAllPlugins() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	plugins, err := pm.store.ListPlugins()
	if err != nil {
		return fmt.Errorf("加载插件列表失败: %w", err)
	}

	for _, plugin := range plugins {
		if plugin.Metadata.Enabled {
			if err := pm.loadPluginLocked(plugin); err != nil {
				log.Printf("[PluginManager] 加载插件 %s 失败: %v", plugin.Metadata.ID, err)
				continue
			}
			log.Printf("[PluginManager] 已加载插件: %s (%s)", plugin.Metadata.Name, plugin.Metadata.ID)
		}
	}

	return nil
}

// LoadPlugin 加载单个插件 (外部调用)
func (pm *PluginManager) LoadPlugin(plugin *Plugin) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	return pm.loadPluginLocked(plugin)
}

// loadPluginLocked 加载插件 (需持有锁)
func (pm *PluginManager) loadPluginLocked(plugin *Plugin) error {
	// 创建独立的 VM 实例
	vm := goja.New()

	// 初始化插件状态
	if plugin.State == nil {
		plugin.State = make(map[string]interface{})
	}

	// 注入辅助函数到 VM 全局
	if err := pm.injectHelpers(vm, plugin); err != nil {
		return fmt.Errorf("注入辅助函数失败: %w", err)
	}

	// 执行插件脚本
	_, err := vm.RunString(plugin.CompiledCode)
	if err != nil {
		return fmt.Errorf("插件脚本执行失败: %w", err)
	}

	// 获取插件对象
	pluginObj := vm.Get("plugin")
	if goja.IsUndefined(pluginObj) {
		return fmt.Errorf("未找到 plugin 对象")
	}

	obj := pluginObj.ToObject(vm)

	// 提取 onEvent 函数 (必需)
	onEventVal := obj.Get("onEvent")
	if goja.IsUndefined(onEventVal) {
		return fmt.Errorf("未找到 onEvent 函数")
	}
	onEventFunc, ok := goja.AssertFunction(onEventVal)
	if !ok {
		return fmt.Errorf("onEvent 不是函数")
	}

	// 提取 onInit 函数 (可选)
	var onInitFunc goja.Callable
	onInitVal := obj.Get("onInit")
	if !goja.IsUndefined(onInitVal) {
		onInitFunc, _ = goja.AssertFunction(onInitVal)
	}

	// 提取 onDestroy 函数 (可选)
	var onDestroyFunc goja.Callable
	onDestroyVal := obj.Get("onDestroy")
	if !goja.IsUndefined(onDestroyVal) {
		onDestroyFunc, _ = goja.AssertFunction(onDestroyVal)
	}

	// 保存到插件实例
	plugin.VM = vm
	plugin.OnEventFunc = onEventFunc
	plugin.OnInitFunc = onInitFunc
	plugin.OnDestroy = onDestroyFunc

	// 调用 onInit (如果存在)
	if onInitFunc != nil {
		context := pm.createPluginContext(vm, plugin)
		if _, err := onInitFunc(goja.Undefined(), context); err != nil {
			log.Printf("[PluginManager] 插件 %s 的 onInit 执行失败: %v", plugin.Metadata.ID, err)
		}
	}

	// 保存到内存
	pm.plugins[plugin.Metadata.ID] = plugin

	return nil
}

// UnloadPlugin 卸载插件
func (pm *PluginManager) UnloadPlugin(id string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	plugin, exists := pm.plugins[id]
	if !exists {
		return fmt.Errorf("插件不存在: %s", id)
	}

	// 调用 onDestroy (如果存在)
	if plugin.OnDestroy != nil {
		context := pm.createPluginContext(plugin.VM, plugin)
		if _, err := plugin.OnDestroy(goja.Undefined(), context); err != nil {
			log.Printf("[PluginManager] 插件 %s 的 onDestroy 执行失败: %v", id, err)
		}
	}

	// 从内存删除
	delete(pm.plugins, id)

	log.Printf("[PluginManager] 已卸载插件: %s", id)
	return nil
}

// ProcessEvent 处理事件 (从 EventPipeline 调用)
func (pm *PluginManager) ProcessEvent(event UnifiedEvent, sessionID string) []UnifiedEvent {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var allDerivedEvents []UnifiedEvent

	// 遍历所有启用的插件
	for _, plugin := range pm.plugins {
		if !plugin.Metadata.Enabled {
			continue
		}

		// 检查是否匹配事件
		if !plugin.MatchesEvent(event) {
			continue
		}

		// 执行插件
		result, err := pm.executePlugin(plugin, event, sessionID)
		if err != nil {
			log.Printf("[PluginManager] 插件 %s 执行失败: %v", plugin.Metadata.ID, err)
			// 生成错误事件
			errorEvent := pm.createPluginErrorEvent(plugin.Metadata.ID, event.ID, err)
			allDerivedEvents = append(allDerivedEvents, errorEvent)
			continue
		}

		// 收集派生事件
		if result != nil && len(result.DerivedEvents) > 0 {
			// 设置派生事件的字段
			for i := range result.DerivedEvents {
				derived := &result.DerivedEvents[i]
				derived.ID = uuid.New().String()
				derived.DeviceID = event.DeviceID
				derived.SessionID = sessionID
				derived.Timestamp = event.Timestamp
				derived.RelativeTime = event.RelativeTime
				derived.Category = CategoryPlugin
				derived.ParentEventID = event.ID
				derived.GeneratedByPlugin = plugin.Metadata.ID

				// 添加元数据标记
				if derived.Metadata == nil {
					derived.Metadata = make(map[string]interface{})
				}
				derived.Metadata["generatedBy"] = plugin.Metadata.ID
				derived.Metadata["pluginName"] = plugin.Metadata.Name
			}

			allDerivedEvents = append(allDerivedEvents, result.DerivedEvents...)
		}

		// TODO: 处理 tags 和 metadata (附加到原事件)
	}

	return allDerivedEvents
}

// executePlugin 执行单个插件 (带超时保护)
func (pm *PluginManager) executePlugin(plugin *Plugin, event UnifiedEvent, sessionID string) (*PluginResult, error) {
	// 超时控制
	timeout := 5 * time.Second
	resultChan := make(chan *PluginResult, 1)
	errorChan := make(chan error, 1)

	go func() {
		// 准备事件上下文
		context := pm.createEventContext(plugin.VM, plugin, event, sessionID)

		// 准备事件对象
		eventObj := plugin.VM.ToValue(event)

		// 调用 onEvent
		resultVal, err := plugin.OnEventFunc(goja.Undefined(), eventObj, context)
		if err != nil {
			errorChan <- err
			return
		}

		// 解析返回值
		result := &PluginResult{}
		if !goja.IsUndefined(resultVal) && !goja.IsNull(resultVal) {
			// 将 JS 对象转换为 Go 结构
			if err := plugin.VM.ExportTo(resultVal, result); err != nil {
				errorChan <- fmt.Errorf("解析插件返回值失败: %w", err)
				return
			}
		}

		resultChan <- result
	}()

	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errorChan:
		return nil, err
	case <-time.After(timeout):
		// 超时：中断 VM
		plugin.VM.Interrupt("timeout")
		return nil, fmt.Errorf("插件执行超时 (>%v)", timeout)
	}
}

// ========== 辅助函数注入 ==========

// injectHelpers 注入辅助函数到 VM
func (pm *PluginManager) injectHelpers(vm *goja.Runtime, plugin *Plugin) error {
	// matchURL: URL 通配符匹配
	vm.Set("matchURL", func(pattern, url string) bool {
		return matchURLPattern(pattern, url)
	})

	// matchRegex: 正则匹配
	vm.Set("matchRegex", func(regexStr, text string) interface{} {
		re, err := regexp.Compile(regexStr)
		if err != nil {
			return nil
		}
		matches := re.FindStringSubmatch(text)
		if matches == nil {
			return nil
		}
		return matches
	})

	// jsonPath: JSON Path 查询
	vm.Set("jsonPath", func(obj interface{}, path string) interface{} {
		// 将 obj 转换为 JSON 字符串
		jsonBytes, err := json.Marshal(obj)
		if err != nil {
			return nil
		}
		// 使用 gjson 查询
		result := gjson.GetBytes(jsonBytes, path)
		if !result.Exists() {
			return nil
		}
		return result.Value()
	})

	// base64Decode: Base64 解码
	vm.Set("base64Decode", func(encoded string) string {
		// TODO: 实现 Base64 解码
		return encoded
	})

	// parseURL: URL 解析
	vm.Set("parseURL", func(urlStr string) interface{} {
		// TODO: 实现 URL 解析
		return nil
	})

	// parseQuery: 查询参数解析
	vm.Set("parseQuery", func(query string) map[string]string {
		// TODO: 实现查询参数解析
		return nil
	})

	// formatTime: 时间格式化
	vm.Set("formatTime", func(timestamp int64, format string) string {
		t := time.UnixMilli(timestamp)
		// 简单格式化
		if format == "" {
			format = "2006-01-02 15:04:05"
		}
		return t.Format(format)
	})

	return nil
}

// createPluginContext 创建插件上下文 (onInit/onDestroy 用)
func (pm *PluginManager) createPluginContext(vm *goja.Runtime, plugin *Plugin) goja.Value {
	context := vm.NewObject()

	context.Set("pluginID", plugin.Metadata.ID)
	context.Set("config", plugin.Metadata.Config)

	// 暴露 state 对象（支持直接访问 context.state.xxx）
	context.Set("state", plugin.State)

	// log 函数
	context.Set("log", func(message string, level string) {
		if level == "" {
			level = "info"
		}
		log.Printf("[Plugin:%s] [%s] %s", plugin.Metadata.ID, level, message)
	})

	// setState / getState（保留向后兼容）
	context.Set("setState", func(key string, value interface{}) {
		plugin.State[key] = value
	})

	context.Set("getState", func(key string) interface{} {
		return plugin.State[key]
	})

	return context
}

// createEventContext 创建事件处理上下文 (onEvent 用)
func (pm *PluginManager) createEventContext(vm *goja.Runtime, plugin *Plugin, event UnifiedEvent, sessionID string) goja.Value {
	context := pm.createPluginContext(vm, plugin)
	obj := context.ToObject(vm)

	obj.Set("deviceID", event.DeviceID)
	obj.Set("sessionID", sessionID)

	// queryEvents 函数
	obj.Set("queryEvents", func(query map[string]interface{}) []UnifiedEvent {
		// TODO: 实现事件查询
		return []UnifiedEvent{}
	})

	return context
}

// createPluginErrorEvent 创建插件错误事件
func (pm *PluginManager) createPluginErrorEvent(pluginID, eventID string, err error) UnifiedEvent {
	return UnifiedEvent{
		ID:        uuid.New().String(),
		Timestamp: time.Now().UnixMilli(),
		Source:    SourceSystem,
		Category:  CategoryLog,
		Type:      "plugin_error",
		Level:     LevelError,
		Title:     fmt.Sprintf("插件执行错误: %s", pluginID),
		Summary:   err.Error(),
		Data: mustMarshal(map[string]interface{}{
			"pluginID":     pluginID,
			"error":        err.Error(),
			"triggerEvent": eventID,
		}),
	}
}

// ========== 管理方法 ==========

// GetPlugin 获取插件
func (pm *PluginManager) GetPlugin(id string) (*Plugin, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	plugin, exists := pm.plugins[id]
	if !exists {
		return nil, fmt.Errorf("插件不存在: %s", id)
	}

	return plugin, nil
}

// ListPlugins 列出所有插件
func (pm *PluginManager) ListPlugins() []*Plugin {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	plugins := make([]*Plugin, 0, len(pm.plugins))
	for _, plugin := range pm.plugins {
		plugins = append(plugins, plugin)
	}

	return plugins
}

// SavePlugin 保存插件
func (pm *PluginManager) SavePlugin(plugin *Plugin) error {
	// 保存到数据库
	if err := pm.store.SavePlugin(plugin); err != nil {
		return fmt.Errorf("保存插件到数据库失败: %w", err)
	}

	// 如果已加载，则重新加载
	if _, exists := pm.plugins[plugin.Metadata.ID]; exists {
		if err := pm.UnloadPlugin(plugin.Metadata.ID); err != nil {
			log.Printf("[PluginManager] 卸载旧插件失败: %v", err)
		}
	}

	// 如果启用，则加载
	if plugin.Metadata.Enabled {
		if err := pm.LoadPlugin(plugin); err != nil {
			return fmt.Errorf("加载插件失败: %w", err)
		}
	}

	return nil
}

// DeletePlugin 删除插件
func (pm *PluginManager) DeletePlugin(id string) error {
	// 卸载插件
	if err := pm.UnloadPlugin(id); err != nil {
		log.Printf("[PluginManager] 卸载插件失败: %v", err)
	}

	// 从数据库删除
	if err := pm.store.DeletePlugin(id); err != nil {
		return fmt.Errorf("从数据库删除插件失败: %w", err)
	}

	return nil
}

// TogglePlugin 启用/禁用插件
func (pm *PluginManager) TogglePlugin(id string, enabled bool) error {
	plugin, err := pm.store.GetPlugin(id)
	if err != nil {
		return err
	}

	plugin.Metadata.Enabled = enabled

	if err := pm.store.SavePlugin(plugin); err != nil {
		return err
	}

	if enabled {
		// 加载插件
		return pm.LoadPlugin(plugin)
	} else {
		// 卸载插件
		return pm.UnloadPlugin(id)
	}
}
