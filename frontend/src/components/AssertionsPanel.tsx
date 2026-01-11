/**
 * AssertionsPanel - 断言验证面板
 * 允许用户创建和执行各种断言来验证设备/应用行为
 */
import { useState, useCallback } from 'react';
import {
  Card,
  Button,
  Space,
  Typography,
  Tag,
  List,
  Input,
  Select,
  Form,
  Modal,
  message,
  Collapse,
  Result,
  Tooltip,
  Badge,
  Empty,
  Spin,
  Divider,
  InputNumber,
} from 'antd';
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  PlayCircleOutlined,
  PlusOutlined,
  ThunderboltOutlined,
  SafetyOutlined,
  BugOutlined,
  OrderedListOutlined,
  FieldTimeOutlined,
  NumberOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
// @ts-ignore
import {
  ExecuteAssertionJSON,
  QuickAssertExists,
  QuickAssertCount,
  QuickAssertNoErrors,
  QuickAssertNoCrashes,
  QuickAssertSequence,
  ListAssertionResults,
} from '../../wailsjs/go/main/App';
import type { main } from '../../wailsjs/go/models';

const { Text, Title } = Typography;
const { Panel } = Collapse;

interface AssertionsPanelProps {
  sessionId: string;
  deviceId: string;
}

interface AssertionResultDisplay {
  id: string;
  name: string;
  passed: boolean;
  message: string;
  executedAt: number;
  duration: number;
  matchedCount: number;
}

const AssertionsPanel: React.FC<AssertionsPanelProps> = ({ sessionId, deviceId }) => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [results, setResults] = useState<AssertionResultDisplay[]>([]);
  const [customModalOpen, setCustomModalOpen] = useState(false);
  const [form] = Form.useForm();

  // 执行快捷断言
  const executeQuickAssertion = useCallback(async (
    type: 'exists' | 'count' | 'noErrors' | 'noCrashes' | 'sequence',
    params?: any
  ) => {
    if (!sessionId) {
      message.warning(t('assertions.select_session_first'));
      return;
    }

    setLoading(true);
    try {
      let result: main.AssertionResult | null = null;

      switch (type) {
        case 'noErrors':
          result = await QuickAssertNoErrors(sessionId, deviceId);
          break;
        case 'noCrashes':
          result = await QuickAssertNoCrashes(sessionId, deviceId);
          break;
        case 'exists':
          result = await QuickAssertExists(
            sessionId,
            deviceId,
            params.eventType || '',
            params.titleMatch || ''
          );
          break;
        case 'count':
          result = await QuickAssertCount(
            sessionId,
            deviceId,
            params.eventType || '',
            params.minCount || 0,
            params.maxCount || 999999
          );
          break;
        case 'sequence':
          result = await QuickAssertSequence(
            sessionId,
            deviceId,
            params.eventTypes || []
          );
          break;
      }

      if (result) {
        const displayResult: AssertionResultDisplay = {
          id: result.id,
          name: result.assertionName,
          passed: result.passed,
          message: result.message,
          executedAt: result.executedAt,
          duration: result.duration,
          matchedCount: result.matchedEvents?.length || 0,
        };
        setResults(prev => [displayResult, ...prev]);

        if (result.passed) {
          message.success(`${t('assertions.passed_msg')}: ${result.assertionName}`);
        } else {
          message.error(`${t('assertions.failed_msg')}: ${result.message}`);
        }
      }
    } catch (err) {
      message.error(`${t('assertions.error_msg')}: ${err}`);
    } finally {
      setLoading(false);
    }
  }, [sessionId, deviceId, t]);

  // 执行自定义断言
  const executeCustomAssertion = useCallback(async (values: any) => {
    if (!sessionId) {
      message.warning(t('assertions.select_session_first'));
      return;
    }

    setLoading(true);
    try {
      // Build assertion object
      const assertion: Record<string, any> = {
        id: `custom_${Date.now()}`,
        name: values.name || 'Custom Assertion',
        type: values.type,
        sessionId: sessionId,
        deviceId: deviceId,
        criteria: {
          types: values.eventTypes ? values.eventTypes.split(',').map((s: string) => s.trim()) : undefined,
          titleMatch: values.titleMatch || undefined,
        },
        expected: {},
        createdAt: Date.now(),
      };

      // 根据类型设置期望值
      switch (values.type) {
        case 'exists':
          assertion.expected = { exists: true };
          break;
        case 'not_exists':
          assertion.expected = { exists: false };
          break;
        case 'count':
          assertion.expected = {
            minCount: values.minCount,
            maxCount: values.maxCount,
          };
          break;
      }

      // Use JSON version to avoid TypeScript class issues
      const result = await ExecuteAssertionJSON(JSON.stringify(assertion));
      if (result) {
        const displayResult: AssertionResultDisplay = {
          id: result.id,
          name: result.assertionName,
          passed: result.passed,
          message: result.message,
          executedAt: result.executedAt,
          duration: result.duration,
          matchedCount: result.matchedEvents?.length || 0,
        };
        setResults(prev => [displayResult, ...prev]);
        setCustomModalOpen(false);
        form.resetFields();

        if (result.passed) {
          message.success(`${t('assertions.passed_msg')}: ${result.assertionName}`);
        } else {
          message.error(`${t('assertions.failed_msg')}: ${result.message}`);
        }
      }
    } catch (err) {
      message.error(`${t('assertions.error_msg')}: ${err}`);
    } finally {
      setLoading(false);
    }
  }, [sessionId, deviceId, form, t]);

  // 加载历史结果
  const loadHistory = useCallback(async () => {
    if (!sessionId) return;

    try {
      const historyResults = await ListAssertionResults(sessionId, 50);
      if (historyResults) {
        setResults(historyResults.map(r => ({
          id: r.id,
          name: r.assertionName,
          passed: r.passed,
          message: r.message,
          executedAt: r.executedAt,
          duration: r.duration,
          matchedCount: r.matchedEvents?.length || 0,
        })));
      }
    } catch (err) {
      console.error('Failed to load assertion history:', err);
    }
  }, [sessionId]);

  // Quick assertion buttons
  const quickAssertions = [
    {
      key: 'noErrors',
      icon: <SafetyOutlined />,
      label: t('assertions.no_errors'),
      description: t('assertions.no_errors_desc'),
      color: 'green',
      onClick: () => executeQuickAssertion('noErrors'),
    },
    {
      key: 'noCrashes',
      icon: <BugOutlined />,
      label: t('assertions.no_crashes'),
      description: t('assertions.no_crashes_desc'),
      color: 'red',
      onClick: () => executeQuickAssertion('noCrashes'),
    },
  ];

  return (
    <Card
      size="small"
      title={
        <Space>
          <ThunderboltOutlined />
          <span>{t('assertions.title')}</span>
          <Badge count={results.filter(r => !r.passed).length} />
        </Space>
      }
      extra={
        <Button
          type="primary"
          size="small"
          icon={<PlusOutlined />}
          onClick={() => setCustomModalOpen(true)}
        >
          {t('assertions.custom')}
        </Button>
      }
      style={{ height: '100%', display: 'flex', flexDirection: 'column' }}
      bodyStyle={{ flex: 1, overflow: 'auto', padding: 12 }}
    >
      {/* Quick Assertions */}
      <div style={{ marginBottom: 16 }}>
        <Text type="secondary" style={{ fontSize: 12, marginBottom: 8, display: 'block' }}>
          {t('assertions.quick_checks')}
        </Text>
        <Space wrap>
          {quickAssertions.map(qa => (
            <Tooltip key={qa.key} title={qa.description}>
              <Button
                icon={qa.icon}
                size="small"
                loading={loading}
                onClick={qa.onClick}
              >
                {qa.label}
              </Button>
            </Tooltip>
          ))}
        </Space>
      </div>

      <Divider style={{ margin: '12px 0' }} />

      {/* Results */}
      <div style={{ flex: 1 }}>
        <Text type="secondary" style={{ fontSize: 12, marginBottom: 8, display: 'block' }}>
          {t('assertions.results')} ({results.length})
        </Text>

        {results.length === 0 ? (
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description={t('assertions.no_results')}
            style={{ marginTop: 24 }}
          />
        ) : (
          <List
            size="small"
            dataSource={results}
            renderItem={item => (
              <List.Item style={{ padding: '8px 0' }}>
                <div style={{ width: '100%' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    {item.passed ? (
                      <CheckCircleOutlined style={{ color: '#52c41a' }} />
                    ) : (
                      <CloseCircleOutlined style={{ color: '#ff4d4f' }} />
                    )}
                    <Text strong style={{ flex: 1, fontSize: 13 }}>
                      {item.name}
                    </Text>
                    <Tag color={item.passed ? 'success' : 'error'}>
                      {item.passed ? t('assertions.pass') : t('assertions.fail')}
                    </Tag>
                  </div>
                  <Text type="secondary" style={{ fontSize: 12, marginLeft: 22 }}>
                    {item.message}
                  </Text>
                  <div style={{ marginLeft: 22, marginTop: 4 }}>
                    <Text type="secondary" style={{ fontSize: 11 }}>
                      {new Date(item.executedAt).toLocaleTimeString()} |
                      {item.duration}ms |
                      {item.matchedCount} {t('assertions.matched')}
                    </Text>
                  </div>
                </div>
              </List.Item>
            )}
          />
        )}
      </div>

      {/* Custom Assertion Modal */}
      <Modal
        title={t('assertions.create_custom')}
        open={customModalOpen}
        onCancel={() => setCustomModalOpen(false)}
        footer={null}
        width={500}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={executeCustomAssertion}
        >
          <Form.Item
            name="name"
            label={t('assertions.assertion_name')}
            rules={[{ required: true }]}
          >
            <Input placeholder={t('assertions.name_placeholder')} />
          </Form.Item>

          <Form.Item
            name="type"
            label={t('assertions.assertion_type')}
            rules={[{ required: true }]}
          >
            <Select
              placeholder={t('assertions.select_type')}
              options={[
                { label: t('assertions.type_exists'), value: 'exists' },
                { label: t('assertions.type_not_exists'), value: 'not_exists' },
                { label: t('assertions.type_count'), value: 'count' },
              ]}
            />
          </Form.Item>

          <Form.Item
            name="eventTypes"
            label={t('assertions.event_types')}
            extra={t('assertions.event_types_extra')}
          >
            <Input placeholder={t('assertions.event_types_placeholder')} />
          </Form.Item>

          <Form.Item
            name="titleMatch"
            label={t('assertions.title_match')}
          >
            <Input placeholder={t('assertions.title_match_placeholder')} />
          </Form.Item>

          <Form.Item noStyle shouldUpdate={(prev, curr) => prev.type !== curr.type}>
            {({ getFieldValue }) =>
              getFieldValue('type') === 'count' && (
                <Space>
                  <Form.Item name="minCount" label={t('assertions.min_count')}>
                    <InputNumber min={0} />
                  </Form.Item>
                  <Form.Item name="maxCount" label={t('assertions.max_count')}>
                    <InputNumber min={0} />
                  </Form.Item>
                </Space>
              )
            }
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={loading}>
                {t('assertions.execute')}
              </Button>
              <Button onClick={() => setCustomModalOpen(false)}>
                {t('common.cancel')}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
};

export default AssertionsPanel;
