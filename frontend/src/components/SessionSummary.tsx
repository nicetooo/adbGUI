/**
 * SessionSummary - AI-powered session summary component
 * Displays an overview of session activity with key findings and suggestions
 */
import React, { useState, useCallback } from 'react';
import {
  Modal,
  Button,
  Card,
  Typography,
  Spin,
  Alert,
  Tag,
  Space,
  Timeline,
  Statistic,
  Row,
  Col,
  Divider,
  List,
  Progress,
  Empty,
} from 'antd';
import {
  RobotOutlined,
  FileTextOutlined,
  CheckCircleOutlined,
  WarningOutlined,
  CloseCircleOutlined,
  InfoCircleOutlined,
  ClockCircleOutlined,
  ThunderboltOutlined,
  ApiOutlined,
  BugOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useAIStore } from '../stores/aiStore';

const { Title, Text, Paragraph } = Typography;

// Types matching backend
interface SessionFinding {
  type: 'error' | 'warning' | 'info' | 'success';
  title: string;
  description: string;
  eventIds?: string[];
  severity: number;
}

interface TimelineHighlight {
  timestamp: number;
  relativeMs: number;
  title: string;
  description: string;
  eventId?: string;
}

interface SessionStats {
  totalEvents: number;
  eventsBySource: Record<string, number>;
  eventsByLevel: Record<string, number>;
  errorCount: number;
  warningCount: number;
  crashCount: number;
  networkRequests: number;
  touchEvents: number;
}

interface SessionSummaryData {
  sessionId: string;
  duration: number;
  eventCount: number;
  overview: string;
  keyFindings: SessionFinding[];
  timeline: TimelineHighlight[];
  statistics: SessionStats;
  suggestions: string[];
  generatedAt: number;
}

interface SessionSummaryProps {
  sessionId: string;
  onEventClick?: (eventId: string) => void;
  /** Custom trigger element. If not provided, renders a default button */
  trigger?: React.ReactNode;
  /** If true, only render the modal (use with external trigger via ref) */
  triggerless?: boolean;
}

// Wails binding
const AISummarizeSession = (window as any).go?.main?.App?.AISummarizeSession;

// Get finding icon
const getFindingIcon = (type: string) => {
  switch (type) {
    case 'error':
      return <CloseCircleOutlined style={{ color: '#ff4d4f' }} />;
    case 'warning':
      return <WarningOutlined style={{ color: '#faad14' }} />;
    case 'success':
      return <CheckCircleOutlined style={{ color: '#52c41a' }} />;
    default:
      return <InfoCircleOutlined style={{ color: '#1890ff' }} />;
  }
};

// Get finding color
const getFindingColor = (type: string) => {
  switch (type) {
    case 'error':
      return 'error';
    case 'warning':
      return 'warning';
    case 'success':
      return 'success';
    default:
      return 'processing';
  }
};

// Format duration
const formatDuration = (ms: number): string => {
  const seconds = Math.floor(ms / 1000);
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ${seconds % 60}s`;
  const hours = Math.floor(minutes / 60);
  return `${hours}h ${minutes % 60}m`;
};

// Format relative time
const formatRelativeTime = (ms: number): string => {
  const seconds = Math.floor(ms / 1000);
  if (seconds < 60) return `+${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  return `+${minutes}m ${seconds % 60}s`;
};

const SessionSummary: React.FC<SessionSummaryProps> = ({ sessionId, onEventClick, trigger, triggerless }) => {
  const { t } = useTranslation();
  const { serviceInfo } = useAIStore();

  const [isOpen, setIsOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [summary, setSummary] = useState<SessionSummaryData | null>(null);
  const [error, setError] = useState<string | null>(null);

  const isAIAvailable = serviceInfo?.status === 'ready';

  const handleGenerateSummary = useCallback(async () => {
    if (!AISummarizeSession || !sessionId) {
      setError(t('session_summary.api_not_available', 'Summary API not available'));
      return;
    }

    setIsLoading(true);
    setError(null);
    setIsOpen(true);

    try {
      const result = await AISummarizeSession(sessionId);
      if (result) {
        setSummary(result);
      } else {
        setError(t('session_summary.no_data', 'No summary data returned'));
      }
    } catch (err) {
      setError(String(err));
    } finally {
      setIsLoading(false);
    }
  }, [sessionId, t]);

  const handleClose = () => {
    setIsOpen(false);
  };

  // Render trigger element
  const renderTrigger = () => {
    if (triggerless) {
      return null;
    }
    if (trigger) {
      return React.cloneElement(trigger as React.ReactElement, {
        onClick: handleGenerateSummary,
      });
    }
    return (
      <Button
        icon={<FileTextOutlined />}
        onClick={handleGenerateSummary}
        disabled={!sessionId}
        title={t('session_summary.generate', 'Generate Session Summary')}
      >
        {t('session_summary.summary', 'Summary')}
      </Button>
    );
  };

  return (
    <>
      {renderTrigger()}

      <Modal
        title={
          <Space>
            <RobotOutlined />
            {t('session_summary.title', 'Session Summary')}
            {isAIAvailable && <Tag color="blue">AI</Tag>}
          </Space>
        }
        open={isOpen}
        onCancel={handleClose}
        width={800}
        footer={[
          <Button key="close" onClick={handleClose}>
            {t('common.close', 'Close')}
          </Button>,
        ]}
      >
        {isLoading && (
          <div style={{ textAlign: 'center', padding: 40 }}>
            <Spin size="large" />
            <Text style={{ display: 'block', marginTop: 16 }} type="secondary">
              {t('session_summary.generating', 'Analyzing session...')}
            </Text>
          </div>
        )}

        {error && (
          <Alert
            type="error"
            message={t('session_summary.error', 'Failed to generate summary')}
            description={error}
            closable
            onClose={() => setError(null)}
          />
        )}

        {summary && !isLoading && (
          <div>
            {/* Overview */}
            <Card size="small" style={{ marginBottom: 16 }}>
              <Paragraph>{summary.overview}</Paragraph>
              <Row gutter={16}>
                <Col span={6}>
                  <Statistic
                    title={t('session_summary.duration', 'Duration')}
                    value={formatDuration(summary.duration)}
                    prefix={<ClockCircleOutlined />}
                  />
                </Col>
                <Col span={6}>
                  <Statistic
                    title={t('session_summary.events', 'Events')}
                    value={summary.eventCount}
                    prefix={<ThunderboltOutlined />}
                  />
                </Col>
                <Col span={6}>
                  <Statistic
                    title={t('session_summary.errors', 'Errors')}
                    value={summary.statistics.errorCount}
                    valueStyle={{ color: summary.statistics.errorCount > 0 ? '#ff4d4f' : undefined }}
                    prefix={<BugOutlined />}
                  />
                </Col>
                <Col span={6}>
                  <Statistic
                    title={t('session_summary.network', 'Network')}
                    value={summary.statistics.networkRequests}
                    prefix={<ApiOutlined />}
                  />
                </Col>
              </Row>
            </Card>

            {/* Key Findings */}
            {summary.keyFindings.length > 0 && (
              <>
                <Title level={5}>{t('session_summary.key_findings', 'Key Findings')}</Title>
                <List
                  size="small"
                  dataSource={summary.keyFindings}
                  renderItem={(finding) => (
                    <List.Item>
                      <List.Item.Meta
                        avatar={getFindingIcon(finding.type)}
                        title={
                          <Space>
                            <Text strong>{finding.title}</Text>
                            <Tag color={getFindingColor(finding.type)}>{finding.type}</Tag>
                            {finding.severity > 0.7 && (
                              <Progress
                                type="circle"
                                percent={Math.round(finding.severity * 100)}
                                size={24}
                                strokeColor="#ff4d4f"
                              />
                            )}
                          </Space>
                        }
                        description={finding.description}
                      />
                    </List.Item>
                  )}
                  style={{ marginBottom: 16 }}
                />
              </>
            )}

            {/* Statistics */}
            <Title level={5}>{t('session_summary.statistics', 'Statistics')}</Title>
            <Card size="small" style={{ marginBottom: 16 }}>
              <Row gutter={[8, 8]}>
                {Object.entries(summary.statistics.eventsBySource).map(([source, count]) => (
                  <Col key={source} span={6}>
                    <Tag>{source}: {count}</Tag>
                  </Col>
                ))}
              </Row>
            </Card>

            {/* Timeline Highlights */}
            {summary.timeline.length > 0 && (
              <>
                <Title level={5}>{t('session_summary.timeline', 'Timeline Highlights')}</Title>
                <Timeline
                  items={summary.timeline.slice(0, 10).map((item) => ({
                    color: item.title === 'Error' || item.title === 'App Crash' ? 'red' : 'blue',
                    children: (
                      <div
                        style={{ cursor: item.eventId ? 'pointer' : 'default' }}
                        onClick={() => item.eventId && onEventClick?.(item.eventId)}
                      >
                        <Text type="secondary" style={{ fontSize: 11 }}>
                          {formatRelativeTime(item.relativeMs)}
                        </Text>
                        <br />
                        <Text strong>{item.title}</Text>
                        <br />
                        <Text type="secondary" ellipsis style={{ maxWidth: 400 }}>
                          {item.description}
                        </Text>
                      </div>
                    ),
                  }))}
                />
              </>
            )}

            {/* Suggestions */}
            {summary.suggestions && summary.suggestions.length > 0 && (
              <>
                <Divider />
                <Title level={5}>{t('session_summary.suggestions', 'Suggestions')}</Title>
                <List
                  size="small"
                  dataSource={summary.suggestions}
                  renderItem={(suggestion, index) => (
                    <List.Item>
                      <Text>
                        <Tag color="blue">{index + 1}</Tag>
                        {suggestion}
                      </Text>
                    </List.Item>
                  )}
                />
              </>
            )}

            {/* Generated timestamp */}
            <Divider />
            <Text type="secondary" style={{ fontSize: 11 }}>
              {t('session_summary.generated_at', 'Generated at')}: {new Date(summary.generatedAt).toLocaleString()}
            </Text>
          </div>
        )}

        {!summary && !isLoading && !error && (
          <Empty description={t('session_summary.click_to_generate', 'Click the button to generate summary')} />
        )}
      </Modal>
    </>
  );
};

export default SessionSummary;
