/**
 * AssertionsPanel - 断言验证面板
 * 允许用户创建和执行各种断言来验证设备/应用行为
 */
import { useState, useCallback, useEffect } from 'react';
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
  Alert,
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
  DeleteOutlined,
  ClearOutlined,
  QuestionCircleOutlined,
  SearchOutlined,
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
  GetSessionEventTypes,
  PreviewAssertionMatch,
  CreateStoredAssertionJSON,
  ListStoredAssertions,
  DeleteStoredAssertion,
  ExecuteStoredAssertionInSession,
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

interface StoredAssertionDisplay {
  id: string;
  name: string;
  type: string;
  createdAt: number;
}

const AssertionsPanel: React.FC<AssertionsPanelProps> = ({ sessionId, deviceId }) => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [results, setResults] = useState<AssertionResultDisplay[]>([]);
  const [customModalOpen, setCustomModalOpen] = useState(false);
  const [form] = Form.useForm();

  // UI/UX enhancement states
  const [availableEventTypes, setAvailableEventTypes] = useState<string[]>([]);
  const [loadingEventTypes, setLoadingEventTypes] = useState(false);
  const [previewCount, setPreviewCount] = useState<number | null>(null);
  const [previewLoading, setPreviewLoading] = useState(false);

  // Stored assertions
  const [storedAssertions, setStoredAssertions] = useState<StoredAssertionDisplay[]>([]);
  const [loadingStored, setLoadingStored] = useState(false);

  // Load stored assertions (global, not bound to session)
  const loadStoredAssertions = useCallback(async () => {
    setLoadingStored(true);
    try {
      // Pass empty sessionId to get all assertions (global)
      const assertions = await ListStoredAssertions('', '', false, 50);
      if (assertions) {
        setStoredAssertions(assertions.map((a: any) => ({
          id: a.id,
          name: a.name,
          type: a.type,
          createdAt: a.createdAt,
        })));
      } else {
        setStoredAssertions([]);
      }
    } catch (err) {
      console.error('Failed to load stored assertions:', err);
      setStoredAssertions([]);
    } finally {
      setLoadingStored(false);
    }
  }, []);

  // Load stored assertions on mount
  useEffect(() => {
    loadStoredAssertions();
  }, [loadStoredAssertions]);

  // Load event types when modal opens
  useEffect(() => {
    if (customModalOpen && sessionId) {
      setLoadingEventTypes(true);
      GetSessionEventTypes(sessionId)
        .then((types: string[]) => {
          setAvailableEventTypes(types || []);
        })
        .catch((err: unknown) => {
          console.error('Failed to load event types:', err);
        })
        .finally(() => {
          setLoadingEventTypes(false);
        });
    }
  }, [customModalOpen, sessionId]);

  // Preview match count when criteria changes
  const updatePreview = useCallback(async (eventTypes: string[], titleMatch: string) => {
    if (!sessionId) {
      setPreviewCount(null);
      return;
    }

    setPreviewLoading(true);
    try {
      const count = await PreviewAssertionMatch(sessionId, eventTypes || [], titleMatch || '');
      setPreviewCount(count);
    } catch (err) {
      console.error('Failed to preview match:', err);
      setPreviewCount(null);
    } finally {
      setPreviewLoading(false);
    }
  }, [sessionId]);

  // Debounced preview update
  const handleFormValuesChange = useCallback((changedValues: any, allValues: any) => {
    const eventTypes = allValues.eventTypes || [];
    const titleMatch = allValues.titleMatch || '';

    // Only update preview if we have some criteria
    if (eventTypes.length > 0 || titleMatch) {
      updatePreview(eventTypes, titleMatch);
    } else {
      setPreviewCount(null);
    }
  }, [updatePreview]);

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
      // Build assertion object - eventTypes is now an array from Select
      // Note: assertions are global, not bound to specific session
      const assertion: Record<string, any> = {
        id: `custom_${Date.now()}`,
        name: values.name || 'Custom Assertion',
        type: values.type,
        // Don't bind to session - assertions are global templates
        sessionId: '',
        deviceId: '',
        criteria: {
          types: Array.isArray(values.eventTypes) ? values.eventTypes : undefined,
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

      // 保存断言模板（不绑定 session）
      const templateJSON = JSON.stringify(assertion);
      try {
        await CreateStoredAssertionJSON(templateJSON, false);
        // 刷新已保存断言列表
        loadStoredAssertions();
      } catch (saveErr) {
        console.error('Failed to save assertion:', saveErr);
      }

      // 执行断言（需要当前 session）
      const execAssertion = { ...assertion, sessionId, deviceId };
      const result = await ExecuteAssertionJSON(JSON.stringify(execAssertion));
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
  }, [sessionId, deviceId, form, t, loadStoredAssertions]);

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

  // 删除单个结果
  const removeResult = useCallback((id: string) => {
    setResults(prev => prev.filter(r => r.id !== id));
  }, []);

  // 清空所有结果
  const clearAllResults = useCallback(() => {
    setResults([]);
  }, []);

  // 执行已保存的断言
  const executeStoredAssertionById = useCallback(async (assertionId: string) => {
    if (!sessionId) {
      message.warning(t('assertions.select_session_first'));
      return;
    }

    setLoading(true);
    try {
      // Execute with current session context (for global assertions)
      const result = await ExecuteStoredAssertionInSession(assertionId, sessionId, deviceId);
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

  // 删除已保存的断言
  const deleteStoredAssertionById = useCallback(async (assertionId: string) => {
    try {
      await DeleteStoredAssertion(assertionId);
      setStoredAssertions(prev => prev.filter(a => a.id !== assertionId));
      message.success(t('assertions.deleted'));
    } catch (err) {
      message.error(`${t('assertions.delete_failed')}: ${err}`);
    }
  }, [t]);

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
        <Space wrap size={[8, 8]}>
          {/* 内置快捷断言 */}
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
          {/* 已保存的断言 */}
          {storedAssertions.map(item => (
            <Button.Group key={item.id} size="small">
              <Tooltip title={t('assertions.execute')}>
                <Button
                  size="small"
                  loading={loading}
                  onClick={() => executeStoredAssertionById(item.id)}
                >
                  {item.name}
                </Button>
              </Tooltip>
              <Tooltip title={t('common.delete')}>
                <Button
                  size="small"
                  icon={<DeleteOutlined />}
                  onClick={(e) => {
                    e.stopPropagation();
                    deleteStoredAssertionById(item.id);
                  }}
                  danger
                />
              </Tooltip>
            </Button.Group>
          ))}
        </Space>
      </div>

      <Divider style={{ margin: '12px 0' }} />

      {/* Results */}
      <div style={{ flex: 1 }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 8 }}>
          <Text type="secondary" style={{ fontSize: 12 }}>
            {t('assertions.results')} ({results.length})
          </Text>
          {results.length > 0 && (
            <Tooltip title={t('assertions.clear_all')}>
              <Button
                type="text"
                size="small"
                icon={<ClearOutlined />}
                onClick={clearAllResults}
                style={{ fontSize: 12 }}
              />
            </Tooltip>
          )}
        </div>

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
                    <Text type="secondary" style={{ fontSize: 11 }}>
                      {new Date(item.executedAt).toLocaleTimeString()} | {item.duration}ms | {item.matchedCount} {t('assertions.matched')}
                    </Text>
                    <Tooltip title={t('common.delete')}>
                      <Button
                        type="text"
                        size="small"
                        icon={<DeleteOutlined />}
                        onClick={() => removeResult(item.id)}
                        style={{ color: '#ff4d4f' }}
                      />
                    </Tooltip>
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
        onCancel={() => {
          setCustomModalOpen(false);
          setPreviewCount(null);
          form.resetFields();
        }}
        footer={null}
        width={520}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={executeCustomAssertion}
          onValuesChange={handleFormValuesChange}
        >
          <Form.Item
            name="name"
            label={
              <Space>
                {t('assertions.assertion_name')}
                <Tooltip title={t('assertions.name_tooltip')}>
                  <QuestionCircleOutlined style={{ color: '#999' }} />
                </Tooltip>
              </Space>
            }
            rules={[{ required: true, message: t('assertions.name_required') }]}
          >
            <Input placeholder={t('assertions.name_placeholder')} />
          </Form.Item>

          <Form.Item
            name="type"
            label={
              <Space>
                {t('assertions.assertion_type')}
                <Tooltip title={t('assertions.type_tooltip')}>
                  <QuestionCircleOutlined style={{ color: '#999' }} />
                </Tooltip>
              </Space>
            }
            rules={[{ required: true, message: t('assertions.type_required') }]}
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
            label={
              <Space>
                {t('assertions.event_types')}
                <Tooltip title={t('assertions.event_types_tooltip')}>
                  <QuestionCircleOutlined style={{ color: '#999' }} />
                </Tooltip>
              </Space>
            }
          >
            <Select
              mode="multiple"
              placeholder={t('assertions.event_types_placeholder')}
              loading={loadingEventTypes}
              allowClear
              showSearch
              optionFilterProp="label"
              options={availableEventTypes.map(type => ({
                label: type,
                value: type,
              }))}
              notFoundContent={
                loadingEventTypes ? (
                  <Spin size="small" />
                ) : availableEventTypes.length === 0 ? (
                  <Text type="secondary">{t('assertions.no_event_types')}</Text>
                ) : null
              }
            />
          </Form.Item>

          <Form.Item
            name="titleMatch"
            label={
              <Space>
                {t('assertions.title_match')}
                <Tooltip title={t('assertions.title_match_tooltip')}>
                  <QuestionCircleOutlined style={{ color: '#999' }} />
                </Tooltip>
              </Space>
            }
            rules={[
              {
                validator: async (_, value) => {
                  if (value) {
                    try {
                      new RegExp(value);
                    } catch (e) {
                      throw new Error(t('assertions.invalid_regex'));
                    }
                  }
                },
              },
            ]}
          >
            <Input
              placeholder={t('assertions.title_match_placeholder')}
              prefix={<SearchOutlined style={{ color: '#999' }} />}
            />
          </Form.Item>

          {/* Match Preview */}
          {previewCount !== null && (
            <Alert
              type={previewCount > 0 ? 'info' : 'warning'}
              showIcon
              icon={previewLoading ? <Spin size="small" /> : undefined}
              message={
                previewLoading
                  ? t('assertions.previewing')
                  : t('assertions.preview_count', { count: previewCount })
              }
              style={{ marginBottom: 16 }}
            />
          )}

          <Form.Item noStyle shouldUpdate={(prev, curr) => prev.type !== curr.type}>
            {({ getFieldValue }) =>
              getFieldValue('type') === 'count' && (
                <Space style={{ width: '100%' }}>
                  <Form.Item
                    name="minCount"
                    label={t('assertions.min_count')}
                    style={{ marginBottom: 16 }}
                  >
                    <InputNumber min={0} style={{ width: 120 }} />
                  </Form.Item>
                  <Form.Item
                    name="maxCount"
                    label={t('assertions.max_count')}
                    style={{ marginBottom: 16 }}
                  >
                    <InputNumber min={0} style={{ width: 120 }} />
                  </Form.Item>
                </Space>
              )
            }
          </Form.Item>

          <Form.Item style={{ marginBottom: 0, marginTop: 8 }}>
            <Space>
              <Button type="primary" htmlType="submit" loading={loading}>
                {t('assertions.execute')}
              </Button>
              <Button onClick={() => {
                setCustomModalOpen(false);
                setPreviewCount(null);
                form.resetFields();
              }}>
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
