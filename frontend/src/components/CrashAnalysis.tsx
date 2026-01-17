import React, { useCallback } from 'react';
import {
  Card,
  Button,
  List,
  Tag,
  Space,
  Typography,
  Collapse,
  Spin,
  Alert,
  Timeline,
  Progress,
  Tooltip,
  Divider,
  Badge,
  Empty,
} from 'antd';
import {
  BugOutlined,
  ThunderboltOutlined,
  ClockCircleOutlined,
  WarningOutlined,
  FileSearchOutlined,
  ReloadOutlined,
  RobotOutlined,
  FireOutlined,
  ApiOutlined,
  MobileOutlined,
  DatabaseOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useAIStore } from '../stores/aiStore';
import { useCrashAnalysisStore } from '../stores/crashAnalysisStore';

const { Text, Title, Paragraph } = Typography;
const { Panel } = Collapse;

// Types
interface CauseCandidate {
  eventId: string;
  eventType: string;
  eventTitle: string;
  timestamp: number;
  explanation: string;
  probability: number;
  category: string;
}

interface RelatedEvent {
  id: string;
  type: string;
  title: string;
  timestamp: number;
  source: string;
  level: string;
}

interface RootCauseAnalysis {
  crashEventId: string;
  crashType: string;
  crashMessage: string;
  crashTime: number;
  probableCauses: CauseCandidate[];
  relatedEvents: RelatedEvent[];
  summary: string;
  confidence: number;
  recommendations: string[];
}

interface CrashEvent {
  id: string;
  type: string;
  title: string;
  timestamp: number;
  data: {
    exception?: string;
    stackTrace?: string;
    packageName?: string;
    processName?: string;
  };
}

interface CrashAnalysisProps {
  sessionId: string;
  crashEvents: CrashEvent[];
  onEventClick?: (eventId: string) => void;
}

// Wails bindings
const AIAnalyzeCrash = (window as any).go?.main?.App?.AIAnalyzeCrash;

// Get cause category icon
const getCategoryIcon = (category: string) => {
  switch (category) {
    case 'network':
      return <ApiOutlined />;
    case 'memory':
      return <DatabaseOutlined />;
    case 'state':
      return <MobileOutlined />;
    case 'log':
      return <FileSearchOutlined />;
    default:
      return <WarningOutlined />;
  }
};

// Get probability color
const getProbabilityColor = (probability: number) => {
  if (probability >= 0.7) return '#ff4d4f';
  if (probability >= 0.4) return '#faad14';
  return '#8c8c8c';
};

const CrashAnalysis: React.FC<CrashAnalysisProps> = ({
  sessionId,
  crashEvents,
  onEventClick,
}) => {
  const { t } = useTranslation();
  const { serviceInfo } = useAIStore();

  // State from store
  const {
    selectedCrash,
    analysis,
    isAnalyzing,
    error,
    setSelectedCrash,
    setAnalysis,
    setIsAnalyzing,
    setError,
  } = useCrashAnalysisStore();

  // Check if AI is available
  const isAIAvailable = serviceInfo?.status === 'ready';

  // Analyze crash
  const handleAnalyzeCrash = useCallback(
    async (crashEvent: CrashEvent) => {
      if (!AIAnalyzeCrash) {
        setError('AI crash analysis not available');
        return;
      }

      setSelectedCrash(crashEvent);
      setIsAnalyzing(true);
      setError(null);
      setAnalysis(null);

      try {
        const result = await AIAnalyzeCrash(crashEvent.id, sessionId);
        if (result) {
          setAnalysis(result);
        }
      } catch (err) {
        setError(String(err));
      } finally {
        setIsAnalyzing(false);
      }
    },
    [sessionId]
  );

  // Format time
  const formatTime = (timestamp: number) => {
    const date = new Date(timestamp);
    return date.toLocaleTimeString();
  };

  // Format relative time
  const formatRelativeTime = (timestamp: number, referenceTime: number) => {
    const diff = (referenceTime - timestamp) / 1000;
    if (diff < 60) return `${diff.toFixed(1)}s before`;
    if (diff < 3600) return `${(diff / 60).toFixed(1)}m before`;
    return `${(diff / 3600).toFixed(1)}h before`;
  };

  if (crashEvents.length === 0) {
    return (
      <Card>
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description={t('crash.no_crashes', 'No crash events in this session')}
        />
      </Card>
    );
  }

  return (
    <div style={{ display: 'flex', gap: 16 }}>
      {/* Crash List */}
      <Card
        title={
          <Space>
            <BugOutlined />
            <span>{t('crash.events', 'Crash Events')}</span>
            <Badge count={crashEvents.length} style={{ backgroundColor: '#ff4d4f' }} />
          </Space>
        }
        style={{ width: 300, flexShrink: 0 }}
        bodyStyle={{ padding: 0 }}
      >
        <List
          dataSource={crashEvents}
          renderItem={(crash) => (
            <List.Item
              style={{
                cursor: 'pointer',
                padding: '12px 16px',
                backgroundColor:
                  selectedCrash?.id === crash.id ? '#e6f7ff' : undefined,
              }}
              onClick={() => handleAnalyzeCrash(crash)}
            >
              <List.Item.Meta
                avatar={
                  <FireOutlined
                    style={{ color: '#ff4d4f', fontSize: 20 }}
                  />
                }
                title={
                  <Text ellipsis style={{ width: 200 }}>
                    {crash.type === 'app_anr' ? 'ANR' : 'Crash'}:{' '}
                    {crash.data?.packageName || 'Unknown'}
                  </Text>
                }
                description={
                  <Space direction="vertical" size={0}>
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      {formatTime(crash.timestamp)}
                    </Text>
                    <Text
                      type="secondary"
                      ellipsis
                      style={{ fontSize: 12, width: 200 }}
                    >
                      {crash.title}
                    </Text>
                  </Space>
                }
              />
            </List.Item>
          )}
        />
      </Card>

      {/* Analysis Panel */}
      <Card
        title={
          <Space>
            <RobotOutlined />
            <span>{t('crash.ai_analysis', 'AI Root Cause Analysis')}</span>
          </Space>
        }
        extra={
          selectedCrash && (
            <Button
              icon={<ReloadOutlined />}
              onClick={() => handleAnalyzeCrash(selectedCrash)}
              loading={isAnalyzing}
              disabled={!isAIAvailable}
            >
              {t('crash.reanalyze', 'Re-analyze')}
            </Button>
          )
        }
        style={{ flex: 1 }}
      >
        {!isAIAvailable && (
          <Alert
            type="warning"
            message={t('ai.not_configured', 'AI service not configured')}
            style={{ marginBottom: 16 }}
          />
        )}

        {!selectedCrash && (
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description={t('crash.select_crash', 'Select a crash event to analyze')}
          />
        )}

        {isAnalyzing && (
          <div style={{ textAlign: 'center', padding: 40 }}>
            <Spin size="large" />
            <Text
              type="secondary"
              style={{ display: 'block', marginTop: 16 }}
            >
              {t('crash.analyzing', 'Analyzing crash root cause...')}
            </Text>
          </div>
        )}

        {error && (
          <Alert
            type="error"
            message={t('crash.analysis_failed', 'Analysis Failed')}
            description={error}
            closable
            onClose={() => setError(null)}
          />
        )}

        {analysis && !isAnalyzing && (
          <>
            {/* Summary */}
            <Alert
              type="info"
              message={
                <Space>
                  <Text strong>{t('crash.summary', 'Summary')}</Text>
                  <Tag color={analysis.confidence > 0.7 ? 'green' : 'orange'}>
                    {t('crash.confidence', 'Confidence')}:{' '}
                    {(analysis.confidence * 100).toFixed(0)}%
                  </Tag>
                </Space>
              }
              description={analysis.summary}
              style={{ marginBottom: 16 }}
            />

            {/* Crash Info */}
            <Card size="small" style={{ marginBottom: 16 }}>
              <Space direction="vertical">
                <Space>
                  <Tag color="error">{analysis.crashType}</Tag>
                  <Text type="secondary">{formatTime(analysis.crashTime)}</Text>
                </Space>
                <Text code style={{ display: 'block' }}>
                  {analysis.crashMessage}
                </Text>
              </Space>
            </Card>

            {/* Probable Causes */}
            <Collapse defaultActiveKey={['causes']} style={{ marginBottom: 16 }}>
              <Panel
                header={
                  <Space>
                    <ThunderboltOutlined />
                    {t('crash.probable_causes', 'Probable Causes')} (
                    {analysis.probableCauses.length})
                  </Space>
                }
                key="causes"
              >
                <Timeline>
                  {analysis.probableCauses.map((cause, index) => (
                    <Timeline.Item
                      key={cause.eventId}
                      color={getProbabilityColor(cause.probability)}
                      dot={getCategoryIcon(cause.category)}
                    >
                      <Card
                        size="small"
                        hoverable
                        onClick={() => onEventClick?.(cause.eventId)}
                        style={{ cursor: 'pointer' }}
                      >
                        <Space
                          direction="vertical"
                          size={4}
                          style={{ width: '100%' }}
                        >
                          <Space wrap>
                            <Tag color="blue">{cause.eventType}</Tag>
                            <Tag>{cause.category}</Tag>
                            <Progress
                              type="circle"
                              percent={Math.round(cause.probability * 100)}
                              size={40}
                              strokeColor={getProbabilityColor(
                                cause.probability
                              )}
                            />
                          </Space>
                          <Text strong>{cause.eventTitle}</Text>
                          <Text type="secondary">
                            {formatRelativeTime(
                              cause.timestamp,
                              analysis.crashTime
                            )}
                          </Text>
                          <Paragraph
                            type="secondary"
                            style={{ marginBottom: 0 }}
                          >
                            {cause.explanation}
                          </Paragraph>
                        </Space>
                      </Card>
                    </Timeline.Item>
                  ))}
                </Timeline>
              </Panel>
            </Collapse>

            {/* Related Events */}
            <Collapse style={{ marginBottom: 16 }}>
              <Panel
                header={
                  <Space>
                    <ClockCircleOutlined />
                    {t('crash.related_events', 'Related Events')} (
                    {analysis.relatedEvents.length})
                  </Space>
                }
                key="events"
              >
                <List
                  size="small"
                  dataSource={analysis.relatedEvents}
                  renderItem={(event) => (
                    <List.Item
                      style={{ cursor: 'pointer' }}
                      onClick={() => onEventClick?.(event.id)}
                    >
                      <List.Item.Meta
                        title={
                          <Space>
                            <Tag>{event.type}</Tag>
                            <Tag
                              color={
                                event.level === 'error'
                                  ? 'red'
                                  : event.level === 'warn'
                                  ? 'orange'
                                  : 'default'
                              }
                            >
                              {event.level}
                            </Tag>
                            <Text>{event.title}</Text>
                          </Space>
                        }
                        description={
                          <Text type="secondary">
                            {formatRelativeTime(
                              event.timestamp,
                              analysis.crashTime
                            )}{' '}
                            - {event.source}
                          </Text>
                        }
                      />
                    </List.Item>
                  )}
                />
              </Panel>
            </Collapse>

            {/* Recommendations */}
            {analysis.recommendations &&
              analysis.recommendations.length > 0 && (
                <Collapse>
                  <Panel
                    header={
                      <Space>
                        <FileSearchOutlined />
                        {t('crash.recommendations', 'Recommendations')}
                      </Space>
                    }
                    key="recommendations"
                  >
                    <List
                      size="small"
                      dataSource={analysis.recommendations}
                      renderItem={(rec, index) => (
                        <List.Item>
                          <Space>
                            <Badge
                              count={index + 1}
                              style={{ backgroundColor: '#1890ff' }}
                            />
                            <Text>{rec}</Text>
                          </Space>
                        </List.Item>
                      )}
                    />
                  </Panel>
                </Collapse>
              )}
          </>
        )}
      </Card>
    </div>
  );
};

export default CrashAnalysis;
