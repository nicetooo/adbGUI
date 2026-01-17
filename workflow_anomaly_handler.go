package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ========================================
// Workflow Anomaly Handler - 异常检测和处理
// ========================================

// AnomalyType 异常类型
type AnomalyType string

const (
	AnomalyPermissionDialog AnomalyType = "permission_dialog"
	AnomalySystemDialog     AnomalyType = "system_dialog"
	AnomalyANRDialog        AnomalyType = "anr_dialog"
	AnomalyErrorDialog      AnomalyType = "error_dialog"
	AnomalyElementNotFound  AnomalyType = "element_not_found"
	AnomalyUIChanged        AnomalyType = "ui_changed"
	AnomalyTimeout          AnomalyType = "timeout"
	AnomalyUnknown          AnomalyType = "unknown"
)

// SuggestedAction 建议的操作
type SuggestedAction struct {
	ID          string         `json:"id"`
	Label       string         `json:"label"`       // 显示文本
	Description string         `json:"description"` // 详细描述
	Steps       []WorkflowStep `json:"steps"`       // 要执行的步骤
	Priority    int            `json:"priority"`    // 推荐优先级 (越小越优先)
	AutoExecute bool           `json:"autoExecute"` // 是否可以自动执行
}

// AnomalyAnalysis AI 异常分析结果
type AnomalyAnalysis struct {
	AnomalyType      AnomalyType       `json:"anomalyType"`
	Description      string            `json:"description"`
	Confidence       float32           `json:"confidence"`
	SuggestedActions []SuggestedAction `json:"suggestedActions"`
	AutoResolvable   bool              `json:"autoResolvable"`
	DialogTitle      string            `json:"dialogTitle,omitempty"`
	DialogMessage    string            `json:"dialogMessage,omitempty"`
	DetectedButtons  []string          `json:"detectedButtons,omitempty"`
}

// WorkflowAnomalyHandler 异常处理器
type WorkflowAnomalyHandler struct {
	app       *App
	aiService *AIService
}

// NewWorkflowAnomalyHandler 创建异常处理器
func NewWorkflowAnomalyHandler(app *App) *WorkflowAnomalyHandler {
	var aiService *AIService
	if app != nil {
		app.aiServiceMu.RLock()
		aiService = app.aiService
		app.aiServiceMu.RUnlock()
	}
	return &WorkflowAnomalyHandler{
		app:       app,
		aiService: aiService,
	}
}

// AnalyzeAnomaly 分析当前屏幕状态，检测异常
func (h *WorkflowAnomalyHandler) AnalyzeAnomaly(ctx context.Context, deviceID string, expectedStep *WorkflowStep, errorMsg string) (*AnomalyAnalysis, error) {
	// 获取当前 UI 层级
	hierarchyResult, err := h.app.GetUIHierarchy(deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get UI hierarchy: %w", err)
	}

	// 使用原始 XML 进行分析
	hierarchy := hierarchyResult.RawXML

	// 尝试 AI 分析
	if h.aiService != nil && h.aiService.IsReady() {
		analysis, err := h.analyzeWithAI(ctx, hierarchy, expectedStep, errorMsg)
		if err == nil && analysis != nil {
			return analysis, nil
		}
	}

	// 回退到基于规则的分析
	return h.analyzeWithRules(hierarchy, expectedStep, errorMsg)
}

// analyzeWithAI 使用 AI 分析异常
func (h *WorkflowAnomalyHandler) analyzeWithAI(ctx context.Context, hierarchy string, expectedStep *WorkflowStep, errorMsg string) (*AnomalyAnalysis, error) {
	prompt := fmt.Sprintf(`Analyze this Android UI state and identify the anomaly.

Expected step: %s (type: %s)
Error message: %s

Current UI hierarchy:
%s

Return a JSON object:
{
  "anomalyType": "permission_dialog|system_dialog|anr_dialog|error_dialog|element_not_found|ui_changed|timeout|unknown",
  "description": "Human-readable description",
  "confidence": 0.9,
  "autoResolvable": true,
  "dialogTitle": "Dialog title if any",
  "dialogMessage": "Dialog message if any",
  "detectedButtons": ["Button1", "Button2"],
  "suggestedActions": [
    {
      "id": "action1",
      "label": "Click Allow",
      "description": "Grant the permission and continue",
      "priority": 1,
      "autoExecute": true
    }
  ]
}

Analyze for:
1. Is there a permission request dialog? (camera, location, storage, etc.)
2. Is there a system dialog? (ANR, crash, low battery, etc.)
3. Is the expected element just not visible?
4. Has the UI layout changed?

Return only valid JSON.`, expectedStep.Name, expectedStep.Type, errorMsg, truncateString(hierarchy, 3000))

	resp, err := h.aiService.Complete(ctx, &CompletionRequest{
		Messages: []ChatMessage{
			{Role: "system", Content: "You are an Android UI analyzer. Analyze the UI state and identify anomalies."},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   800,
		Temperature: 0.2,
	})

	if err != nil {
		return nil, err
	}

	var analysis AnomalyAnalysis
	response := strings.TrimSpace(resp.Content)
	if strings.HasPrefix(response, "```") {
		response = strings.TrimPrefix(response, "```json")
		response = strings.TrimPrefix(response, "```")
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)
	}

	if err := json.Unmarshal([]byte(response), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// 为建议的操作生成实际步骤
	h.enrichSuggestedActions(&analysis)

	return &analysis, nil
}

// analyzeWithRules 使用规则分析异常
func (h *WorkflowAnomalyHandler) analyzeWithRules(hierarchy string, expectedStep *WorkflowStep, errorMsg string) (*AnomalyAnalysis, error) {
	analysis := &AnomalyAnalysis{
		Confidence:       0.7,
		SuggestedActions: []SuggestedAction{},
	}

	lowerHierarchy := strings.ToLower(hierarchy)

	// 检测权限对话框
	if h.isPermissionDialog(lowerHierarchy) {
		analysis.AnomalyType = AnomalyPermissionDialog
		analysis.Description = "Permission request dialog detected"
		analysis.AutoResolvable = true
		analysis.DetectedButtons = h.findButtons(hierarchy, []string{"allow", "deny", "允许", "拒绝", "don't allow"})

		// 添加建议操作
		if containsAny(lowerHierarchy, []string{"allow", "允许"}) {
			analysis.SuggestedActions = append(analysis.SuggestedActions, SuggestedAction{
				ID:          "allow_permission",
				Label:       "Allow Permission",
				Description: "Grant the requested permission and continue",
				Priority:    1,
				AutoExecute: true,
				Steps: []WorkflowStep{
					{
						Type: "click_element",
						Name: "Click Allow",
						Selector: &ElementSelector{
							Type: "text",
							Value:    "Allow",
						},
						Timeout: 5000,
					},
				},
			})
		}

		if containsAny(lowerHierarchy, []string{"deny", "拒绝", "don't allow"}) {
			analysis.SuggestedActions = append(analysis.SuggestedActions, SuggestedAction{
				ID:          "deny_permission",
				Label:       "Deny Permission",
				Description: "Deny the permission request",
				Priority:    2,
				AutoExecute: false,
				Steps: []WorkflowStep{
					{
						Type: "click_element",
						Name: "Click Deny",
						Selector: &ElementSelector{
							Type: "text",
							Value:    "Deny",
						},
						Timeout: 5000,
					},
				},
			})
		}

		return analysis, nil
	}

	// 检测 ANR 对话框
	if h.isANRDialog(lowerHierarchy) {
		analysis.AnomalyType = AnomalyANRDialog
		analysis.Description = "Application Not Responding dialog detected"
		analysis.AutoResolvable = true
		analysis.DetectedButtons = h.findButtons(hierarchy, []string{"wait", "close", "等待", "关闭"})

		analysis.SuggestedActions = append(analysis.SuggestedActions, SuggestedAction{
			ID:          "wait_anr",
			Label:       "Wait",
			Description: "Wait for the application to respond",
			Priority:    1,
			AutoExecute: true,
			Steps: []WorkflowStep{
				{
					Type: "click_element",
					Name: "Click Wait",
					Selector: &ElementSelector{
						Type: "text",
						Value:    "Wait",
					},
					Timeout: 5000,
				},
			},
		})

		analysis.SuggestedActions = append(analysis.SuggestedActions, SuggestedAction{
			ID:          "close_anr",
			Label:       "Close App",
			Description: "Force close the application",
			Priority:    2,
			AutoExecute: false,
			Steps: []WorkflowStep{
				{
					Type: "click_element",
					Name: "Click Close",
					Selector: &ElementSelector{
						Type: "text",
						Value:    "Close",
					},
					Timeout: 5000,
				},
			},
		})

		return analysis, nil
	}

	// 检测系统对话框
	if h.isSystemDialog(lowerHierarchy) {
		analysis.AnomalyType = AnomalySystemDialog
		analysis.Description = "System dialog detected"
		analysis.AutoResolvable = true
		analysis.DetectedButtons = h.findButtons(hierarchy, []string{"ok", "cancel", "yes", "no", "确定", "取消"})

		// 优先点击确定/OK
		for _, btn := range analysis.DetectedButtons {
			lowerBtn := strings.ToLower(btn)
			if lowerBtn == "ok" || lowerBtn == "确定" || lowerBtn == "yes" {
				analysis.SuggestedActions = append(analysis.SuggestedActions, SuggestedAction{
					ID:          "confirm_dialog",
					Label:       btn,
					Description: "Confirm and dismiss the dialog",
					Priority:    1,
					AutoExecute: true,
					Steps: []WorkflowStep{
						{
							Type: "click_element",
							Name: "Click " + btn,
							Selector: &ElementSelector{
								Type: "text",
								Value:    btn,
							},
							Timeout: 5000,
						},
					},
				})
				break
			}
		}

		return analysis, nil
	}

	// 检测错误对话框
	if h.isErrorDialog(lowerHierarchy) {
		analysis.AnomalyType = AnomalyErrorDialog
		analysis.Description = "Error dialog detected"
		analysis.AutoResolvable = false
		analysis.DetectedButtons = h.findButtons(hierarchy, []string{"ok", "retry", "cancel", "确定", "重试", "取消"})

		return analysis, nil
	}

	// 元素未找到
	if strings.Contains(errorMsg, "not found") || strings.Contains(errorMsg, "timeout") {
		analysis.AnomalyType = AnomalyElementNotFound
		analysis.Description = fmt.Sprintf("Expected element not found: %s", expectedStep.Name)
		analysis.AutoResolvable = false

		// 建议等待或跳过
		analysis.SuggestedActions = append(analysis.SuggestedActions, SuggestedAction{
			ID:          "retry_with_wait",
			Label:       "Retry with Wait",
			Description: "Wait 2 seconds and retry",
			Priority:    1,
			AutoExecute: true,
			Steps: []WorkflowStep{
				{
					Type:    "wait",
					Name:    "Wait 2 seconds",
					Timeout: 2000,
				},
			},
		})

		analysis.SuggestedActions = append(analysis.SuggestedActions, SuggestedAction{
			ID:          "skip_step",
			Label:       "Skip Step",
			Description: "Skip this step and continue",
			Priority:    2,
			AutoExecute: false,
			Steps:       []WorkflowStep{},
		})

		return analysis, nil
	}

	// 未知异常
	analysis.AnomalyType = AnomalyUnknown
	analysis.Description = fmt.Sprintf("Unknown anomaly: %s", errorMsg)
	analysis.AutoResolvable = false

	return analysis, nil
}

// isPermissionDialog 检测是否为权限对话框
func (h *WorkflowAnomalyHandler) isPermissionDialog(hierarchy string) bool {
	permissionKeywords := []string{
		"permission",
		"allow",
		"deny",
		"权限",
		"允许",
		"拒绝",
		"com.android.permissioncontroller",
		"com.google.android.permissioncontroller",
		"access your",
		"访问",
	}
	return containsAny(hierarchy, permissionKeywords)
}

// isANRDialog 检测是否为 ANR 对话框
func (h *WorkflowAnomalyHandler) isANRDialog(hierarchy string) bool {
	anrKeywords := []string{
		"isn't responding",
		"not responding",
		"应用无响应",
		"没有响应",
		"anr",
		"wait or close",
		"等待或关闭",
	}
	return containsAny(hierarchy, anrKeywords)
}

// isSystemDialog 检测是否为系统对话框
func (h *WorkflowAnomalyHandler) isSystemDialog(hierarchy string) bool {
	// 检查是否有 AlertDialog 特征
	dialogKeywords := []string{
		"alertdialog",
		"android:id/alertTitle",
		"android:id/message",
		"android:id/button1",
		"android:id/button2",
	}
	return containsAny(hierarchy, dialogKeywords)
}

// isErrorDialog 检测是否为错误对话框
func (h *WorkflowAnomalyHandler) isErrorDialog(hierarchy string) bool {
	errorKeywords := []string{
		"error",
		"failed",
		"错误",
		"失败",
		"exception",
		"crash",
		"unfortunately",
		"has stopped",
		"已停止",
	}
	return containsAny(hierarchy, errorKeywords)
}

// findButtons 在 UI 层级中查找按钮
func (h *WorkflowAnomalyHandler) findButtons(hierarchy string, candidates []string) []string {
	found := []string{}
	lowerHierarchy := strings.ToLower(hierarchy)

	for _, candidate := range candidates {
		if strings.Contains(lowerHierarchy, strings.ToLower(candidate)) {
			found = append(found, candidate)
		}
	}

	return found
}

// enrichSuggestedActions 为建议的操作添加实际步骤
func (h *WorkflowAnomalyHandler) enrichSuggestedActions(analysis *AnomalyAnalysis) {
	for i := range analysis.SuggestedActions {
		action := &analysis.SuggestedActions[i]
		if len(action.Steps) == 0 && action.Label != "" {
			// 根据标签生成点击步骤
			action.Steps = []WorkflowStep{
				{
					Type: "click_element",
					Name: "Click " + action.Label,
					Selector: &ElementSelector{
						Type: "text",
						Value:    action.Label,
					},
					Timeout: 5000,
				},
			}
		}
	}
}

// ExecuteSuggestedAction 执行建议的操作
func (h *WorkflowAnomalyHandler) ExecuteSuggestedAction(ctx context.Context, deviceID string, action *SuggestedAction) error {
	for _, step := range action.Steps {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := h.executeStep(ctx, deviceID, &step); err != nil {
			return fmt.Errorf("failed to execute step %s: %w", step.Name, err)
		}

		// 步骤间短暂等待
		time.Sleep(300 * time.Millisecond)
	}

	return nil
}

// executeStep 执行单个步骤
func (h *WorkflowAnomalyHandler) executeStep(ctx context.Context, deviceID string, step *WorkflowStep) error {
	switch step.Type {
	case "click_element":
		if step.Selector != nil {
			cfg := DefaultElementActionConfig()
			return h.app.ClickElement(ctx, deviceID, step.Selector, &cfg)
		}
		return fmt.Errorf("no selector specified for click_element")

	case "wait":
		timeout := step.Timeout
		if timeout <= 0 {
			timeout = 1000
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(timeout) * time.Millisecond):
			return nil
		}

	default:
		return fmt.Errorf("unsupported step type: %s", step.Type)
	}
}

// Helper functions

func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// ========================================
// Anomaly Learning - 学习用户选择
// ========================================

// AnomalyResolution 记录用户对异常的处理方式
type AnomalyResolution struct {
	ID          string          `json:"id"`
	DeviceID    string          `json:"deviceId"`
	WorkflowID  string          `json:"workflowId"`
	StepIndex   int             `json:"stepIndex"`
	AnomalyType AnomalyType     `json:"anomalyType"`
	DialogTitle string          `json:"dialogTitle,omitempty"`
	Action      SuggestedAction `json:"action"`
	Timestamp   int64           `json:"timestamp"`
	Remember    bool            `json:"remember"` // 是否记住此选择
}

// AnomalyLearner 异常处理学习器
type AnomalyLearner struct {
	resolutions map[string][]AnomalyResolution // key: anomalyType + dialogTitle
}

// NewAnomalyLearner 创建学习器
func NewAnomalyLearner() *AnomalyLearner {
	return &AnomalyLearner{
		resolutions: make(map[string][]AnomalyResolution),
	}
}

// RecordResolution 记录用户选择
func (l *AnomalyLearner) RecordResolution(resolution AnomalyResolution) {
	key := string(resolution.AnomalyType) + ":" + resolution.DialogTitle
	l.resolutions[key] = append(l.resolutions[key], resolution)
}

// GetPreferredAction 获取用户偏好的处理方式
func (l *AnomalyLearner) GetPreferredAction(anomalyType AnomalyType, dialogTitle string) *SuggestedAction {
	key := string(anomalyType) + ":" + dialogTitle

	resolutions, ok := l.resolutions[key]
	if !ok || len(resolutions) == 0 {
		return nil
	}

	// 统计各操作的使用次数
	actionCount := make(map[string]int)
	actionMap := make(map[string]*SuggestedAction)

	for _, r := range resolutions {
		if r.Remember {
			actionCount[r.Action.ID]++
			actionMap[r.Action.ID] = &r.Action
		}
	}

	// 返回最常用的操作
	var bestAction *SuggestedAction
	var maxCount int

	for id, count := range actionCount {
		if count > maxCount {
			maxCount = count
			bestAction = actionMap[id]
		}
	}

	return bestAction
}
