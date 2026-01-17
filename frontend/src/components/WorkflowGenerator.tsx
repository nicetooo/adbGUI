import React, { useCallback, useEffect } from 'react';
import {
  Card,
  Button,
  Progress,
  List,
  Tag,
  Space,
  Typography,
  Modal,
  Form,
  Input,
  Select,
  Switch,
  Collapse,
  Spin,
  Alert,
  Tooltip,
  message,
  theme,
} from 'antd';
import {
  PlayCircleOutlined,
  PauseCircleOutlined,
  ThunderboltOutlined,
  EditOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  SettingOutlined,
  RobotOutlined,
  BranchesOutlined,
  VideoCameraOutlined,
  LoadingOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useWorkflowGeneratorStore } from '../stores/workflowGeneratorStore';

// Wails event listener
const EventsOn = (window as any).runtime?.EventsOn;
const EventsOff = (window as any).runtime?.EventsOff;

const { Text, Title } = Typography;
const { Panel } = Collapse;

// Types
interface WorkflowStep {
  id: string;
  type: string;
  name: string;
  selector?: {
    type: string;
    value: string;
  };
  value?: string;
  timeout?: number;
}

interface GeneratedWorkflow {
  sessionId: string;
  name: string;
  description: string;
  steps: WorkflowStep[];
  suggestions: string[];
  confidence: number;
  loops: LoopPattern[];
  branches: BranchPattern[];
}

interface LoopPattern {
  startIndex: number;
  endIndex: number;
  iterations: number;
  confidence: number;
}

interface BranchPattern {
  condition: string;
  thenSteps: number[];
  elseSteps: number[];
  confidence: number;
}

interface WorkflowGeneratorProps {
  open?: boolean;
  onClose?: () => void;
  sessionId: string;
  deviceId?: string;
  onGenerated?: (workflow: GeneratedWorkflow) => void;
  onSave?: (workflow: GeneratedWorkflow) => void;
}

// Wails bindings
const AIGenerateWorkflow = (window as any).go?.main?.App?.AIGenerateWorkflow;
const SaveWorkflow = (window as any).go?.main?.App?.SaveWorkflow;

const WorkflowGenerator: React.FC<WorkflowGeneratorProps> = ({
  open = true,
  onClose,
  sessionId,
  deviceId,
  onGenerated,
  onSave,
}) => {
  const { t } = useTranslation();
  const { token } = theme.useToken();

  // Store state
  const {
    isGenerating,
    progress,
    progressMessage,
    progressStage,
    generatedWorkflow,
    editModalVisible,
    editingStep,
    editingIndex,
    configModalVisible,
    error,
    aiLogs,
    config,
    setIsGenerating,
    setProgress,
    setProgressMessage,
    setProgressStage,
    setGeneratedWorkflow,
    openEditModal,
    closeEditModal,
    openConfigModal,
    closeConfigModal,
    setError,
    addAiLog,
    clearAiLogs,
    setConfig,
    resetGeneration,
    updateEditingStep,
  } = useWorkflowGeneratorStore();

  // Listen for progress events
  useEffect(() => {
    if (!EventsOn) return;

    const handleProgress = (data: {
      stage: string;
      percent: number;
      message: string;
      frameBase64?: string;
      frameIndex?: number;
      frameTimeMs?: number;
      sceneType?: string;
      description?: string;
      ocrText?: string[];
    }) => {
      setProgress(data.percent);
      setProgressMessage(data.message);
      setProgressStage(data.stage);

      // Log AI-related stages for debugging
      const aiStages = ['ai_prompt', 'ai_response', 'ai_error', 'ai_enhancement', 'frame_analyzed', 'frame_analysis_error'];
      if (aiStages.includes(data.stage)) {
        addAiLog({
          stage: data.stage,
          message: data.message,
          timestamp: Date.now(),
          frameBase64: data.frameBase64,
          frameIndex: data.frameIndex,
          frameTimeMs: data.frameTimeMs,
          sceneType: data.sceneType,
          description: data.description,
          ocrText: data.ocrText,
        });
      }
    };

    EventsOn('workflow-gen-progress', handleProgress);

    return () => {
      if (EventsOff) {
        EventsOff('workflow-gen-progress');
      }
    };
  }, [setProgress, setProgressMessage, setProgressStage, addAiLog]);

  // Generate workflow
  const handleGenerate = useCallback(async () => {
    if (!AIGenerateWorkflow) {
      message.error('AI Workflow generation not available');
      return;
    }

    setIsGenerating(true);
    setProgress(0);
    setProgressMessage('Starting workflow generation...');
    setProgressStage('starting');
    setError(null);
    clearAiLogs(); // Clear previous logs

    try {
      const result = await AIGenerateWorkflow(sessionId, {
        ...config,
        UseVideoAnalysis: config.useVideoAnalysis,
      });

      setProgress(100);
      setProgressMessage('Complete!');

      if (result) {
        setGeneratedWorkflow(result);
        onGenerated?.(result);
        message.success(t('workflow.generated_success', 'Workflow generated successfully'));
      }
    } catch (err) {
      setError(String(err));
      message.error(t('workflow.generation_failed', 'Failed to generate workflow'));
    } finally {
      setIsGenerating(false);
    }
  }, [sessionId, config, onGenerated, t, setIsGenerating, setProgress, setProgressMessage, setProgressStage, setError, clearAiLogs, setGeneratedWorkflow]);

  // Save workflow
  const handleSave = useCallback(async () => {
    if (!generatedWorkflow || !SaveWorkflow) return;

    try {
      const workflowToSave = {
        id: `workflow-${Date.now()}`,
        name: generatedWorkflow.name || 'Generated Workflow',
        description: generatedWorkflow.description || '',
        steps: generatedWorkflow.steps || [],
        variables: {},
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };

      await SaveWorkflow(workflowToSave);
      onSave?.(generatedWorkflow);
      message.success(t('workflow.saved', 'Workflow saved'));
    } catch (err) {
      message.error(t('workflow.save_failed', 'Failed to save workflow'));
    }
  }, [generatedWorkflow, onSave, t]);

  // Edit step
  const handleEditStep = (step: WorkflowStep, index: number) => {
    openEditModal({ ...step }, index);
  };

  // Save step edit
  const handleSaveStepEdit = () => {
    if (!generatedWorkflow || !editingStep || editingIndex < 0) return;

    const newSteps = [...(generatedWorkflow.steps || [])];
    newSteps[editingIndex] = editingStep;

    setGeneratedWorkflow({
      ...generatedWorkflow,
      steps: newSteps,
    });

    closeEditModal();
  };

  // Delete step
  const handleDeleteStep = (index: number) => {
    if (!generatedWorkflow) return;

    const newSteps = (generatedWorkflow.steps || []).filter((_, i) => i !== index);
    setGeneratedWorkflow({
      ...generatedWorkflow,
      steps: newSteps,
    });
  };

  // Get step type icon
  const getStepTypeIcon = (type: string) => {
    switch (type) {
      case 'click':
      case 'tap':
        return <PlayCircleOutlined />;
      case 'input':
      case 'type':
        return <EditOutlined />;
      case 'wait':
        return <PauseCircleOutlined />;
      case 'swipe':
        return <ThunderboltOutlined />;
      default:
        return <CheckCircleOutlined />;
    }
  };

  // Get step type color
  const getStepTypeColor = (type: string) => {
    switch (type) {
      case 'click':
      case 'tap':
        return 'blue';
      case 'input':
      case 'type':
        return 'green';
      case 'wait':
        return 'orange';
      case 'swipe':
        return 'purple';
      case 'assert':
        return 'cyan';
      default:
        return 'default';
    }
  };

  // AI is no longer available - feature disabled
  const isAIAvailable = false;

  const content = (
    <>
      {!isAIAvailable && (
        <Alert
          type="warning"
          message={t('ai.not_configured', 'AI service not configured')}
          description={t('ai.configure_hint', 'Please configure an AI provider in settings to use this feature.')}
          style={{ marginBottom: 16 }}
        />
      )}

      {isGenerating && (
        <div style={{ padding: 24 }}>
          <div style={{ textAlign: 'center' }}>
            <Spin
              size="large"
              indicator={
                progressStage.startsWith('video') ? (
                  <VideoCameraOutlined style={{ fontSize: 32 }} spin />
                ) : progressStage === 'ai_enhancement' || progressStage.startsWith('ai_') ? (
                  <RobotOutlined style={{ fontSize: 32 }} spin />
                ) : (
                  <LoadingOutlined style={{ fontSize: 32 }} spin />
                )
              }
            />
            <Progress
              percent={progress}
              status={progressStage === 'video_error' ? 'exception' : 'active'}
              style={{ marginTop: 16 }}
            />
            <Text type="secondary" style={{ display: 'block', marginTop: 8 }}>
              {progressMessage || t('workflow.analyzing', 'Analyzing session events...')}
            </Text>
            {progressStage.startsWith('video') && (
              <Tag color="purple" style={{ marginTop: 8 }}>
                <VideoCameraOutlined /> {t('workflow.video_analysis', 'Video Analysis')}
              </Tag>
            )}
            {(progressStage === 'ai_enhancement' || progressStage.startsWith('ai_')) && (
              <Tag color="blue" style={{ marginTop: 8 }}>
                <RobotOutlined /> {t('workflow.ai_naming', 'AI Naming Steps')}
              </Tag>
            )}
          </div>

          {/* Real-time AI Logs during generation */}
          {aiLogs.length > 0 && (
            <div style={{
              marginTop: 16,
              maxHeight: 500,
              overflow: 'auto',
              border: `1px solid ${token.colorBorderSecondary}`,
              borderRadius: token.borderRadius,
              padding: 8,
              background: token.colorBgContainer,
            }}>
              <Text strong style={{ marginBottom: 8, display: 'block' }}>
                <RobotOutlined /> AI Interaction Logs ({aiLogs.length})
              </Text>
              {aiLogs.map((log, index) => (
                <div
                  key={index}
                  style={{
                    marginBottom: 8,
                    padding: 8,
                    background: log.stage.includes('error')
                      ? token.colorErrorBg
                      : log.stage === 'frame_analyzed'
                        ? token.colorInfoBg
                        : token.colorSuccessBg,
                    borderRadius: token.borderRadius,
                    border: `1px solid ${log.stage.includes('error')
                      ? token.colorErrorBorder
                      : log.stage === 'frame_analyzed'
                        ? token.colorInfoBorder
                        : token.colorSuccessBorder}`,
                  }}
                >
                  <div style={{ marginBottom: 4 }}>
                    <Tag color={
                      log.stage.includes('error') ? 'red'
                        : log.stage === 'ai_prompt' ? 'blue'
                          : log.stage === 'ai_response' ? 'green'
                            : log.stage === 'frame_analyzed' ? 'geekblue'
                              : 'purple'
                    }>
                      {log.stage}
                    </Tag>
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      {new Date(log.timestamp).toLocaleTimeString()}
                    </Text>
                    {log.sceneType && (
                      <Tag color="cyan" style={{ marginLeft: 4 }}>{log.sceneType}</Tag>
                    )}
                  </div>
                  {/* Show frame image if available */}
                  {log.frameBase64 && (
                    <div style={{ marginBottom: 8, display: 'flex', gap: 12 }}>
                      <div style={{ flexShrink: 0 }}>
                        <img
                          src={log.frameBase64.startsWith('data:') ? log.frameBase64 : `data:image/jpeg;base64,${log.frameBase64}`}
                          alt={`Frame ${log.frameIndex}`}
                          style={{
                            width: 160,
                            height: 'auto',
                            borderRadius: token.borderRadius,
                            border: `1px solid ${token.colorBorderSecondary}`,
                          }}
                        />
                        <Text type="secondary" style={{ fontSize: 10, display: 'block', textAlign: 'center', marginTop: 4 }}>
                          Frame #{log.frameIndex} @ {log.frameTimeMs}ms
                        </Text>
                      </div>
                      <div style={{ flex: 1, minWidth: 0 }}>
                        {log.description && (
                          <div style={{ marginBottom: 4 }}>
                            <Text strong style={{ fontSize: 12 }}>Description:</Text>
                            <Text style={{ fontSize: 12, display: 'block' }}>{log.description}</Text>
                          </div>
                        )}
                        {log.ocrText && log.ocrText.length > 0 && (
                          <div>
                            <Text strong style={{ fontSize: 12 }}>OCR Text:</Text>
                            <div style={{ fontSize: 11, color: token.colorTextSecondary }}>
                              {log.ocrText.slice(0, 5).map((text, i) => (
                                <Tag key={i} style={{ marginBottom: 2, fontSize: 10 }}>{text}</Tag>
                              ))}
                              {log.ocrText.length > 5 && <Text type="secondary">+{log.ocrText.length - 5} more</Text>}
                            </div>
                          </div>
                        )}
                      </div>
                    </div>
                  )}
                  {/* Show message for non-frame logs */}
                  {!log.frameBase64 && (
                    <pre
                      style={{
                        margin: 0,
                        fontSize: 11,
                        whiteSpace: 'pre-wrap',
                        wordBreak: 'break-word',
                        maxHeight: 200,
                        overflow: 'auto',
                        background: token.colorFillTertiary,
                        padding: 8,
                        borderRadius: token.borderRadius,
                        color: token.colorText,
                      }}
                    >
                      {log.message}
                    </pre>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {error && (
        <Alert
          type="error"
          message={t('workflow.error', 'Generation Error')}
          description={error}
          closable
          onClose={() => setError(null)}
          style={{ marginBottom: 16 }}
        />
      )}

      {generatedWorkflow && !isGenerating && (
        <>
          {/* Workflow Info */}
          <div style={{ marginBottom: 16 }}>
            <Title level={5}>
              {generatedWorkflow.name || 'Generated Workflow'}
            </Title>
            <Text type="secondary">{generatedWorkflow.description}</Text>
            <div style={{ marginTop: 8 }}>
              <Tag color="blue">
                {generatedWorkflow.steps?.length || 0} {t('workflow.steps', 'steps')}
              </Tag>
              <Tag color={(generatedWorkflow.confidence || 0) > 0.8 ? 'green' : 'orange'}>
                {t('workflow.confidence', 'Confidence')}: {((generatedWorkflow.confidence || 0) * 100).toFixed(0)}%
              </Tag>
              {generatedWorkflow.loops && generatedWorkflow.loops.length > 0 && (
                <Tag color="purple" icon={<BranchesOutlined />}>
                  {generatedWorkflow.loops.length} {t('workflow.loops', 'loops')}
                </Tag>
              )}
              {(generatedWorkflow as any).metadata?.hasVideoContext ? (
                <Tooltip title={`${(generatedWorkflow as any).metadata?.videoFrameCount || 0} frames analyzed`}>
                  <Tag color="geekblue" icon={<VideoCameraOutlined />}>
                    {t('workflow.video_enhanced', 'Video Enhanced')} ({(generatedWorkflow as any).metadata?.videoFrameCount || 0})
                  </Tag>
                </Tooltip>
              ) : (generatedWorkflow as any).metadata?.useVideoAnalysis ? (
                <Tooltip title={
                  !(generatedWorkflow as any).metadata?.sessionVideoPath
                    ? t('workflow.no_video_hint', 'Record with video to enable video analysis')
                    : t('workflow.video_failed_hint', 'Video analysis failed - check if ffmpeg is installed')
                }>
                  <Tag color="default" icon={<VideoCameraOutlined />}>
                    {!(generatedWorkflow as any).metadata?.sessionVideoPath
                      ? t('workflow.no_video', 'No Video')
                      : t('workflow.video_failed', 'Video Failed')}
                  </Tag>
                </Tooltip>
              ) : null}
              {(generatedWorkflow as any).metadata?.aiEnhanced && (
                <Tag color="cyan" icon={<RobotOutlined />}>
                  {t('workflow.ai_enhanced', 'AI Enhanced')}
                </Tag>
              )}
            </div>
          </div>

          {/* Suggestions */}
          {generatedWorkflow.suggestions && generatedWorkflow.suggestions.length > 0 && (
            <Collapse style={{ marginBottom: 16 }}>
              <Panel
                header={
                  <Space>
                    <ExclamationCircleOutlined />
                    {t('workflow.suggestions', 'AI Suggestions')} ({generatedWorkflow.suggestions.length})
                  </Space>
                }
                key="suggestions"
              >
                <List
                  size="small"
                  dataSource={generatedWorkflow.suggestions}
                  renderItem={(suggestion) => (
                    <List.Item>
                      <Text type="secondary">{suggestion}</Text>
                    </List.Item>
                  )}
                />
              </Panel>
            </Collapse>
          )}

          {/* AI Interaction Logs - for debugging */}
          {aiLogs.length > 0 && (
            <Collapse style={{ marginBottom: 16 }}>
              <Panel
                header={
                  <Space>
                    <RobotOutlined />
                    AI Interaction Logs ({aiLogs.length})
                    {aiLogs.filter(l => l.frameBase64).length > 0 && (
                      <Tag color="geekblue">
                        <VideoCameraOutlined /> {aiLogs.filter(l => l.frameBase64).length} frames
                      </Tag>
                    )}
                  </Space>
                }
                key="ai-logs"
              >
                <div style={{ maxHeight: 500, overflow: 'auto' }}>
                  {aiLogs.map((log, index) => (
                    <div
                      key={index}
                      style={{
                        marginBottom: 12,
                        padding: 8,
                        background: log.stage.includes('error')
                          ? token.colorErrorBg
                          : log.stage === 'frame_analyzed'
                            ? token.colorInfoBg
                            : token.colorSuccessBg,
                        borderRadius: token.borderRadius,
                        border: `1px solid ${log.stage.includes('error')
                          ? token.colorErrorBorder
                          : log.stage === 'frame_analyzed'
                            ? token.colorInfoBorder
                            : token.colorSuccessBorder}`,
                      }}
                    >
                      <div style={{ marginBottom: 4 }}>
                        <Tag color={
                          log.stage.includes('error') ? 'red'
                            : log.stage === 'ai_prompt' ? 'blue'
                              : log.stage === 'ai_response' ? 'green'
                                : log.stage === 'frame_analyzed' ? 'geekblue'
                                  : 'purple'
                        }>
                          {log.stage}
                        </Tag>
                        <Text type="secondary" style={{ fontSize: 12 }}>
                          {new Date(log.timestamp).toLocaleTimeString()}
                        </Text>
                        {log.sceneType && (
                          <Tag color="cyan" style={{ marginLeft: 4 }}>{log.sceneType}</Tag>
                        )}
                      </div>
                      {/* Show frame image if available */}
                      {log.frameBase64 && (
                        <div style={{ marginBottom: 8, display: 'flex', gap: 12 }}>
                          <div style={{ flexShrink: 0 }}>
                            <img
                              src={log.frameBase64.startsWith('data:') ? log.frameBase64 : `data:image/jpeg;base64,${log.frameBase64}`}
                              alt={`Frame ${log.frameIndex}`}
                              style={{
                                width: 180,
                                height: 'auto',
                                borderRadius: token.borderRadius,
                                border: `1px solid ${token.colorBorderSecondary}`,
                              }}
                            />
                            <Text type="secondary" style={{ fontSize: 10, display: 'block', textAlign: 'center', marginTop: 4 }}>
                              Frame #{log.frameIndex} @ {log.frameTimeMs}ms
                            </Text>
                          </div>
                          <div style={{ flex: 1, minWidth: 0 }}>
                            {log.description && (
                              <div style={{ marginBottom: 4 }}>
                                <Text strong style={{ fontSize: 12 }}>Description:</Text>
                                <Text style={{ fontSize: 12, display: 'block' }}>{log.description}</Text>
                              </div>
                            )}
                            {log.ocrText && log.ocrText.length > 0 && (
                              <div>
                                <Text strong style={{ fontSize: 12 }}>OCR Text:</Text>
                                <div style={{ fontSize: 11, color: token.colorTextSecondary }}>
                                  {log.ocrText.slice(0, 8).map((text, i) => (
                                    <Tag key={i} style={{ marginBottom: 2, fontSize: 10 }}>{text}</Tag>
                                  ))}
                                  {log.ocrText.length > 8 && <Text type="secondary">+{log.ocrText.length - 8} more</Text>}
                                </div>
                              </div>
                            )}
                          </div>
                        </div>
                      )}
                      {/* Show message for non-frame logs */}
                      {!log.frameBase64 && (
                        <pre
                          style={{
                            margin: 0,
                            fontSize: 11,
                            whiteSpace: 'pre-wrap',
                            wordBreak: 'break-word',
                            maxHeight: 300,
                            overflow: 'auto',
                            background: token.colorFillTertiary,
                            padding: 8,
                            borderRadius: token.borderRadius,
                            color: token.colorText,
                          }}
                        >
                          {log.message}
                        </pre>
                      )}
                    </div>
                  ))}
                </div>
              </Panel>
            </Collapse>
          )}

          {/* Steps List */}
          <List
            size="small"
            bordered
            dataSource={generatedWorkflow.steps || []}
            renderItem={(step, index) => (
              <List.Item
                actions={[
                  <Tooltip title={t('common.edit', 'Edit')}>
                    <Button
                      type="text"
                      size="small"
                      icon={<EditOutlined />}
                      onClick={() => handleEditStep(step, index)}
                    />
                  </Tooltip>,
                  <Tooltip title={t('common.delete', 'Delete')}>
                    <Button
                      type="text"
                      size="small"
                      danger
                      onClick={() => handleDeleteStep(index)}
                    >
                      X
                    </Button>
                  </Tooltip>,
                ]}
              >
                <List.Item.Meta
                  avatar={
                    <Tag color={getStepTypeColor(step.type)}>
                      {getStepTypeIcon(step.type)} {step.type}
                    </Tag>
                  }
                  title={
                    <Space>
                      <Text strong>#{index + 1}</Text>
                      <Text>{step.name}</Text>
                    </Space>
                  }
                  description={
                    <Space wrap>
                      {step.selector && (
                        <Text code>
                          {step.selector.type}: {step.selector.value.slice(0, 40)}
                          {step.selector.value.length > 40 ? '...' : ''}
                        </Text>
                      )}
                      {step.value && <Text type="secondary">Value: {step.value}</Text>}
                      {step.timeout && <Text type="secondary">Timeout: {step.timeout}ms</Text>}
                    </Space>
                  }
                />
              </List.Item>
            )}
          />

          {/* Action Buttons */}
          <div style={{ marginTop: 16, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => setGeneratedWorkflow(null)}>
                {t('common.reset', 'Reset')}
              </Button>
              <Button type="primary" onClick={handleSave}>
                {t('workflow.save', 'Save Workflow')}
              </Button>
            </Space>
          </div>
        </>
      )}

      {/* Edit Step Modal */}
      <Modal
        title={t('workflow.edit_step', 'Edit Step')}
        open={editModalVisible}
        onOk={handleSaveStepEdit}
        onCancel={closeEditModal}
      >
        {editingStep && (
          <Form layout="vertical">
            <Form.Item label={t('workflow.step_name', 'Step Name')}>
              <Input
                value={editingStep.name}
                onChange={(e) => updateEditingStep({ name: e.target.value })}
              />
            </Form.Item>
            <Form.Item label={t('workflow.step_type', 'Step Type')}>
              <Select
                value={editingStep.type}
                onChange={(value) => updateEditingStep({ type: value })}
              >
                <Select.Option value="click">Click</Select.Option>
                <Select.Option value="tap">Tap</Select.Option>
                <Select.Option value="input">Input</Select.Option>
                <Select.Option value="swipe">Swipe</Select.Option>
                <Select.Option value="wait">Wait</Select.Option>
                <Select.Option value="back">Back</Select.Option>
                <Select.Option value="home">Home</Select.Option>
              </Select>
            </Form.Item>
            {editingStep.selector && (
              <>
                <Form.Item label={t('workflow.selector_type', 'Selector Type')}>
                  <Select
                    value={editingStep.selector.type}
                    onChange={(value) =>
                      updateEditingStep({
                        selector: { ...editingStep.selector!, type: value },
                      })
                    }
                  >
                    <Select.Option value="id">ID</Select.Option>
                    <Select.Option value="text">Text</Select.Option>
                    <Select.Option value="xpath">XPath</Select.Option>
                    <Select.Option value="class">Class</Select.Option>
                  </Select>
                </Form.Item>
                <Form.Item label={t('workflow.selector_value', 'Selector Value')}>
                  <Input.TextArea
                    value={editingStep.selector.value}
                    onChange={(e) =>
                      updateEditingStep({
                        selector: { ...editingStep.selector!, value: e.target.value },
                      })
                    }
                    rows={2}
                  />
                </Form.Item>
              </>
            )}
            {(editingStep.type === 'input' || editingStep.type === 'type') && (
              <Form.Item label={t('workflow.input_value', 'Input Value')}>
                <Input
                  value={editingStep.value}
                  onChange={(e) => updateEditingStep({ value: e.target.value })}
                />
              </Form.Item>
            )}
            <Form.Item label={t('workflow.timeout', 'Timeout (ms)')}>
              <Input
                type="number"
                value={editingStep.timeout}
                onChange={(e) =>
                  updateEditingStep({ timeout: parseInt(e.target.value) || 5000 })
                }
              />
            </Form.Item>
          </Form>
        )}
      </Modal>

      {/* Config Modal */}
      <Modal
        title={t('workflow.generation_config', 'Generation Settings')}
        open={configModalVisible}
        onOk={closeConfigModal}
        onCancel={closeConfigModal}
      >
        <Form layout="vertical">
          <Form.Item label={t('workflow.include_waits', 'Include Wait Steps')}>
            <Switch
              checked={config.includeWaits}
              onChange={(checked) => setConfig({ includeWaits: checked })}
            />
          </Form.Item>
          <Form.Item label={t('workflow.optimize_selectors', 'Optimize Selectors')}>
            <Switch
              checked={config.optimizeSelectors}
              onChange={(checked) => setConfig({ optimizeSelectors: checked })}
            />
          </Form.Item>
          <Form.Item label={t('workflow.detect_loops', 'Detect Repeat Patterns')}>
            <Switch
              checked={config.detectLoops}
              onChange={(checked) => setConfig({ detectLoops: checked })}
            />
          </Form.Item>
          <Form.Item label={t('workflow.detect_branches', 'Detect Branches')}>
            <Switch
              checked={config.detectBranches}
              onChange={(checked) => setConfig({ detectBranches: checked })}
            />
          </Form.Item>
          <Form.Item label={t('workflow.generate_assertions', 'Generate Assertions')}>
            <Switch
              checked={config.generateAssertions}
              onChange={(checked) => setConfig({ generateAssertions: checked })}
            />
          </Form.Item>
          <Form.Item label={t('workflow.min_confidence', 'Minimum Confidence')}>
            <Select
              value={config.minConfidence}
              onChange={(value) => setConfig({ minConfidence: value })}
            >
              <Select.Option value={0.5}>50%</Select.Option>
              <Select.Option value={0.7}>70%</Select.Option>
              <Select.Option value={0.8}>80%</Select.Option>
              <Select.Option value={0.9}>90%</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item
            label={
              <Space>
                <VideoCameraOutlined />
                {t('workflow.use_video_analysis', 'Use Video Frame Analysis')}
              </Space>
            }
            tooltip={t('workflow.video_analysis_hint', 'Extract video frames at each touch event for better context understanding')}
          >
            <Switch
              checked={config.useVideoAnalysis}
              onChange={(checked) => setConfig({ useVideoAnalysis: checked })}
            />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );

  // If used as a modal (with open prop)
  if (onClose !== undefined) {
    return (
      <Modal
        title={
          <Space>
            <RobotOutlined />
            <span>{t('workflow.ai_generator', 'AI Workflow Generator')}</span>
          </Space>
        }
        open={open}
        onCancel={onClose}
        width={800}
        footer={[
          <Button key="settings" icon={<SettingOutlined />} onClick={openConfigModal} disabled={isGenerating}>
            {t('common.settings', 'Settings')}
          </Button>,
          <Button
            key="generate"
            type="primary"
            icon={<ThunderboltOutlined />}
            onClick={handleGenerate}
            loading={isGenerating}
            disabled={!isAIAvailable}
          >
            {t('workflow.generate', 'Generate Workflow')}
          </Button>,
          generatedWorkflow && (
            <Button key="save" type="primary" onClick={handleSave}>
              {t('workflow.save', 'Save Workflow')}
            </Button>
          ),
        ].filter(Boolean)}
        destroyOnClose
      >
        {content}
      </Modal>
    );
  }

  // Standalone card mode
  return (
    <Card
      title={
        <Space>
          <RobotOutlined />
          <span>{t('workflow.ai_generator', 'AI Workflow Generator')}</span>
        </Space>
      }
      extra={
        <Space>
          <Button
            icon={<SettingOutlined />}
            onClick={openConfigModal}
            disabled={isGenerating}
          >
            {t('common.settings', 'Settings')}
          </Button>
          <Button
            type="primary"
            icon={<ThunderboltOutlined />}
            onClick={handleGenerate}
            loading={isGenerating}
            disabled={!isAIAvailable}
          >
            {t('workflow.generate', 'Generate Workflow')}
          </Button>
        </Space>
      }
    >
      {content}
    </Card>
  );
};

export default WorkflowGenerator;
