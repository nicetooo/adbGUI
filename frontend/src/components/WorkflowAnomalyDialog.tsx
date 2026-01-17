import React, { useState, useMemo } from 'react';
import {
  Modal,
  Card,
  Button,
  Radio,
  Checkbox,
  Space,
  Typography,
  Tag,
  Alert,
  Divider,
  List,
  Image,
  Tooltip,
} from 'antd';
import {
  WarningOutlined,
  ExclamationCircleOutlined,
  QuestionCircleOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  PauseCircleOutlined,
  ReloadOutlined,
  FastForwardOutlined,
  LockOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';

const { Text, Title, Paragraph } = Typography;

// Types matching backend
interface SuggestedAction {
  id: string;
  label: string;
  description: string;
  priority: number;
  autoExecute: boolean;
  steps: any[];
}

interface AnomalyAnalysis {
  anomalyType: string;
  description: string;
  confidence: number;
  suggestedActions: SuggestedAction[];
  autoResolvable: boolean;
  dialogTitle?: string;
  dialogMessage?: string;
  detectedButtons?: string[];
}

interface WorkflowStep {
  id: string;
  type: string;
  name: string;
}

interface WorkflowAnomalyDialogProps {
  visible: boolean;
  anomaly: AnomalyAnalysis | null;
  step: WorkflowStep | null;
  stepIndex: number;
  screenshot?: string; // Base64 image
  onAction: (response: AnomalyResponse) => void;
  onCancel: () => void;
}

interface AnomalyResponse {
  action: 'execute' | 'skip' | 'pause' | 'stop' | 'retry';
  steps?: any[];
  updateWorkflow: boolean;
  remember: boolean;
}

// Get anomaly type icon
const getAnomalyIcon = (type: string) => {
  switch (type) {
    case 'permission_dialog':
      return <LockOutlined style={{ color: '#faad14', fontSize: 24 }} />;
    case 'system_dialog':
    case 'anr_dialog':
      return <ExclamationCircleOutlined style={{ color: '#ff4d4f', fontSize: 24 }} />;
    case 'error_dialog':
      return <CloseCircleOutlined style={{ color: '#ff4d4f', fontSize: 24 }} />;
    case 'element_not_found':
      return <QuestionCircleOutlined style={{ color: '#faad14', fontSize: 24 }} />;
    case 'ui_changed':
      return <WarningOutlined style={{ color: '#1890ff', fontSize: 24 }} />;
    default:
      return <ExclamationCircleOutlined style={{ color: '#8c8c8c', fontSize: 24 }} />;
  }
};

// Get anomaly type color
const getAnomalyColor = (type: string) => {
  switch (type) {
    case 'permission_dialog':
      return 'warning';
    case 'system_dialog':
    case 'anr_dialog':
    case 'error_dialog':
      return 'error';
    case 'element_not_found':
      return 'warning';
    case 'ui_changed':
      return 'processing';
    default:
      return 'default';
  }
};

// Get anomaly type label
const getAnomalyLabel = (type: string, t: any) => {
  const labels: Record<string, string> = {
    permission_dialog: t('anomaly.permission_dialog', 'Permission Request'),
    system_dialog: t('anomaly.system_dialog', 'System Dialog'),
    anr_dialog: t('anomaly.anr_dialog', 'App Not Responding'),
    error_dialog: t('anomaly.error_dialog', 'Error Dialog'),
    element_not_found: t('anomaly.element_not_found', 'Element Not Found'),
    ui_changed: t('anomaly.ui_changed', 'UI Changed'),
    timeout: t('anomaly.timeout', 'Timeout'),
    unknown: t('anomaly.unknown', 'Unknown Issue'),
  };
  return labels[type] || type;
};

const WorkflowAnomalyDialog: React.FC<WorkflowAnomalyDialogProps> = ({
  visible,
  anomaly,
  step,
  stepIndex,
  screenshot,
  onAction,
  onCancel,
}) => {
  const { t } = useTranslation();
  const [selectedAction, setSelectedAction] = useState<string>('');
  const [rememberChoice, setRememberChoice] = useState(false);
  const [updateWorkflow, setUpdateWorkflow] = useState(false);

  // Sort actions by priority
  const sortedActions = useMemo(() => {
    if (!anomaly?.suggestedActions) return [];
    return [...anomaly.suggestedActions].sort((a, b) => a.priority - b.priority);
  }, [anomaly]);

  // Handle action selection
  const handleConfirm = () => {
    if (!selectedAction) return;

    const action = sortedActions.find((a) => a.id === selectedAction);
    if (!action) return;

    onAction({
      action: 'execute',
      steps: action.steps,
      updateWorkflow,
      remember: rememberChoice,
    });
  };

  // Handle skip
  const handleSkip = () => {
    onAction({
      action: 'skip',
      updateWorkflow: false,
      remember: false,
    });
  };

  // Handle retry
  const handleRetry = () => {
    onAction({
      action: 'retry',
      updateWorkflow: false,
      remember: false,
    });
  };

  // Handle pause
  const handlePause = () => {
    onAction({
      action: 'pause',
      updateWorkflow: false,
      remember: false,
    });
  };

  // Handle stop
  const handleStop = () => {
    onAction({
      action: 'stop',
      updateWorkflow: false,
      remember: false,
    });
  };

  if (!anomaly || !step) return null;

  return (
    <Modal
      title={
        <Space>
          {getAnomalyIcon(anomaly.anomalyType)}
          <span>{t('workflow.anomaly_detected', 'Workflow Anomaly Detected')}</span>
        </Space>
      }
      open={visible}
      onCancel={onCancel}
      width={700}
      footer={null}
      maskClosable={false}
    >
      {/* Step Info */}
      <Alert
        type="info"
        message={
          <Space>
            <Text strong>
              {t('workflow.step', 'Step')} #{stepIndex + 1}:
            </Text>
            <Text>{step.name}</Text>
            <Tag>{step.type}</Tag>
          </Space>
        }
        style={{ marginBottom: 16 }}
      />

      {/* Anomaly Description */}
      <Card size="small" style={{ marginBottom: 16 }}>
        <Space direction="vertical" style={{ width: '100%' }}>
          <Space>
            <Tag color={getAnomalyColor(anomaly.anomalyType)}>
              {getAnomalyLabel(anomaly.anomalyType, t)}
            </Tag>
            <Tag color={anomaly.confidence > 0.8 ? 'green' : 'orange'}>
              {t('workflow.confidence', 'Confidence')}: {(anomaly.confidence * 100).toFixed(0)}%
            </Tag>
          </Space>
          <Paragraph>{anomaly.description}</Paragraph>

          {/* Dialog info if available */}
          {(anomaly.dialogTitle || anomaly.dialogMessage) && (
            <Card size="small" style={{ background: '#fafafa' }}>
              {anomaly.dialogTitle && <Text strong>{anomaly.dialogTitle}</Text>}
              {anomaly.dialogMessage && (
                <Paragraph type="secondary" style={{ marginTop: 8, marginBottom: 0 }}>
                  {anomaly.dialogMessage}
                </Paragraph>
              )}
              {anomaly.detectedButtons && anomaly.detectedButtons.length > 0 && (
                <div style={{ marginTop: 8 }}>
                  <Text type="secondary">{t('workflow.detected_buttons', 'Detected buttons')}: </Text>
                  {anomaly.detectedButtons.map((btn) => (
                    <Tag key={btn}>{btn}</Tag>
                  ))}
                </div>
              )}
            </Card>
          )}
        </Space>
      </Card>

      {/* Screenshot if available */}
      {screenshot && (
        <Card size="small" title={t('workflow.current_screen', 'Current Screen')} style={{ marginBottom: 16 }}>
          <Image
            src={`data:image/png;base64,${screenshot}`}
            alt="Screenshot"
            style={{ maxWidth: '100%', maxHeight: 300 }}
          />
        </Card>
      )}

      {/* Suggested Actions */}
      <Divider>{t('workflow.suggested_actions', 'Suggested Actions')}</Divider>

      <Radio.Group
        value={selectedAction}
        onChange={(e) => setSelectedAction(e.target.value)}
        style={{ width: '100%' }}
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          {sortedActions.map((action, index) => (
            <Card
              key={action.id}
              size="small"
              hoverable
              style={{
                cursor: 'pointer',
                borderColor: selectedAction === action.id ? '#1890ff' : undefined,
              }}
              onClick={() => setSelectedAction(action.id)}
            >
              <Radio value={action.id}>
                <Space>
                  <Text strong>{action.label}</Text>
                  {index === 0 && (
                    <Tag color="green">{t('workflow.recommended', 'Recommended')}</Tag>
                  )}
                  {action.autoExecute && (
                    <Tag color="blue">{t('workflow.auto_execute', 'Auto')}</Tag>
                  )}
                </Space>
              </Radio>
              <Paragraph
                type="secondary"
                style={{ marginLeft: 24, marginBottom: 0, marginTop: 4 }}
              >
                {action.description}
              </Paragraph>
            </Card>
          ))}
        </Space>
      </Radio.Group>

      {/* Options */}
      <Divider />
      <Space direction="vertical">
        <Checkbox
          checked={rememberChoice}
          onChange={(e) => setRememberChoice(e.target.checked)}
        >
          <Tooltip title={t('workflow.remember_hint', 'Automatically apply this choice for similar situations')}>
            {t('workflow.remember_choice', 'Remember this choice for future')}
          </Tooltip>
        </Checkbox>
        <Checkbox
          checked={updateWorkflow}
          onChange={(e) => setUpdateWorkflow(e.target.checked)}
        >
          <Tooltip title={t('workflow.update_workflow_hint', 'Add handling logic to the workflow')}>
            {t('workflow.add_to_workflow', 'Add this handling to workflow')}
          </Tooltip>
        </Checkbox>
      </Space>

      {/* Action Buttons */}
      <Divider />
      <div style={{ display: 'flex', justifyContent: 'space-between' }}>
        <Space>
          <Button icon={<FastForwardOutlined />} onClick={handleSkip}>
            {t('workflow.skip_step', 'Skip Step')}
          </Button>
          <Button icon={<ReloadOutlined />} onClick={handleRetry}>
            {t('workflow.retry', 'Retry')}
          </Button>
        </Space>
        <Space>
          <Button icon={<PauseCircleOutlined />} onClick={handlePause}>
            {t('workflow.pause', 'Pause')}
          </Button>
          <Button danger icon={<CloseCircleOutlined />} onClick={handleStop}>
            {t('workflow.stop', 'Stop')}
          </Button>
          <Button
            type="primary"
            icon={<CheckCircleOutlined />}
            onClick={handleConfirm}
            disabled={!selectedAction}
          >
            {t('workflow.apply_action', 'Apply Action')}
          </Button>
        </Space>
      </div>
    </Modal>
  );
};

export default WorkflowAnomalyDialog;
