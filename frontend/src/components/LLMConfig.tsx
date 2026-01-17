import React, { useEffect } from 'react';
import {
  Modal,
  Form,
  Input,
  Switch,
  Select,
  Button,
  Card,
  Space,
  Typography,
  Divider,
  Tag,
  Spin,
  Alert,
  Tooltip,
  Collapse,
  message,
} from 'antd';
import {
  RobotOutlined,
  CloudOutlined,
  DesktopOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  SyncOutlined,
  PlusOutlined,
  DeleteOutlined,
  ApiOutlined,
  ExperimentOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useAIStore, DiscoveredService, LLMSource } from '../stores';

const { Text, Paragraph } = Typography;
const { Panel } = Collapse;

interface LLMConfigProps {
  open: boolean;
  onClose: () => void;
}

const LLMConfig: React.FC<LLMConfigProps> = ({ open, onClose }) => {
  const { t } = useTranslation();
  const {
    serviceInfo,
    config,
    discoveredServices,
    isDiscovering,
    isLoading,
    isSaving,
    error,
    loadServiceInfo,
    loadConfig,
    setEnabled,
    setPreferredSource,
    discoverServices,
    setOpenAIConfig,
    setClaudeConfig,
    setCustomProviderConfig,
    switchProvider,
    testProvider,
    getAvailableModels,
    setFeature,
    addCustomEndpoint,
    removeCustomEndpoint,
    refreshProviders,
    // LLMConfig UI state
    testingProvider,
    currentProviderModels,
    isChangingModel,
    setTestingProvider,
    setCurrentProviderModels,
    setIsChangingModel,
  } = useAIStore();

  const [openaiForm] = Form.useForm();
  const [claudeForm] = Form.useForm();
  const [customForm] = Form.useForm();
  const [customEndpointForm] = Form.useForm();

  useEffect(() => {
    if (open) {
      loadServiceInfo();
      loadConfig();
      discoverServices();
    }
  }, [open]);

  // Load models for current provider
  useEffect(() => {
    if (serviceInfo?.provider?.endpoint) {
      // Find the service in discovered services
      const endpoint = serviceInfo.provider.endpoint.replace('/v1', '');
      const service = discoveredServices.find(s => s.endpoint === endpoint);
      if (service) {
        setCurrentProviderModels(service.models);
      }
    }
  }, [serviceInfo, discoveredServices]);

  useEffect(() => {
    if (config?.onlineProviders) {
      openaiForm.setFieldsValue({
        apiKey: config.onlineProviders.openai?.apiKey || '',
        model: config.onlineProviders.openai?.model || 'gpt-4o-mini',
        enabled: config.onlineProviders.openai?.enabled || false,
      });
      claudeForm.setFieldsValue({
        apiKey: config.onlineProviders.claude?.apiKey || '',
        model: config.onlineProviders.claude?.model || 'claude-3-haiku-20240307',
        enabled: config.onlineProviders.claude?.enabled || false,
      });
      customForm.setFieldsValue({
        endpoint: config.onlineProviders.custom?.endpoint || '',
        apiKey: config.onlineProviders.custom?.apiKey || '',
        model: config.onlineProviders.custom?.model || '',
        enabled: config.onlineProviders.custom?.enabled || false,
      });
    }
  }, [config]);

  const handleTest = async (providerType: string, formData: Record<string, any>) => {
    setTestingProvider(providerType);
    try {
      const testConfig: Record<string, string> = {};
      if (formData.apiKey) testConfig.api_key = formData.apiKey;
      if (formData.endpoint) testConfig.endpoint = formData.endpoint;
      if (formData.model) testConfig.model = formData.model;

      const result = await testProvider(providerType, testConfig);
      if (result.success) {
        message.success(result.message);
      } else {
        message.error(result.message);
      }
    } catch (err) {
      message.error(String(err));
    } finally {
      setTestingProvider(null);
    }
  };

  const handleUseService = async (service: DiscoveredService) => {
    try {
      const model = service.models[0] || '';
      await switchProvider(service.type, {
        endpoint: service.endpoint + '/v1',
        model: model,
      });
      message.success(`Switched to ${service.name}`);
      await loadServiceInfo();
    } catch (err) {
      message.error(`Failed to switch: ${String(err)}`);
    }
  };

  const handleChangeCurrentModel = async (model: string) => {
    if (!serviceInfo?.provider) return;
    setIsChangingModel(true);
    try {
      // Get provider type from name
      const providerName = serviceInfo.provider.name.toLowerCase();
      let providerType = 'ollama';
      if (providerName.includes('lm studio')) providerType = 'lmstudio';
      else if (providerName.includes('localai')) providerType = 'localai';

      await switchProvider(providerType, {
        endpoint: serviceInfo.provider.endpoint,
        model: model,
      });
      message.success(`Switched to model ${model}`);
      await loadServiceInfo();
    } catch (err) {
      message.error(`Failed to switch model: ${String(err)}`);
    } finally {
      setIsChangingModel(false);
    }
  };

  const handleSaveOpenAI = async () => {
    const values = await openaiForm.validateFields();
    await setOpenAIConfig(values.apiKey, values.model, values.enabled);
    message.success('OpenAI configuration saved');
  };

  const handleSaveClaude = async () => {
    const values = await claudeForm.validateFields();
    await setClaudeConfig(values.apiKey, values.model, values.enabled);
    message.success('Claude configuration saved');
  };

  const handleSaveCustom = async () => {
    const values = await customForm.validateFields();
    await setCustomProviderConfig(values.endpoint, values.apiKey, values.model, values.enabled);
    message.success('Custom provider configuration saved');
  };

  const handleAddCustomEndpoint = async () => {
    const values = await customEndpointForm.validateFields();
    await addCustomEndpoint(values.name, values.endpoint, values.type);
    customEndpointForm.resetFields();
    message.success('Custom endpoint added');
    discoverServices();
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'running':
        return <CheckCircleOutlined style={{ color: '#52c41a' }} />;
      case 'stopped':
        return <CloseCircleOutlined style={{ color: '#ff4d4f' }} />;
      default:
        return <CloseCircleOutlined style={{ color: '#d9d9d9' }} />;
    }
  };

  const getStatusTag = (status: string) => {
    switch (status) {
      case 'ready':
        return <Tag color="success">Ready</Tag>;
      case 'initializing':
        return <Tag color="processing">Initializing</Tag>;
      case 'no_provider':
        return <Tag color="warning">No Provider</Tag>;
      case 'error':
        return <Tag color="error">Error</Tag>;
      case 'disabled':
        return <Tag color="default">Disabled</Tag>;
      default:
        return <Tag>{status}</Tag>;
    }
  };

  return (
    <Modal
      title={
        <Space>
          <RobotOutlined />
          {t('ai.config_title', 'AI / LLM Configuration')}
        </Space>
      }
      open={open}
      onCancel={onClose}
      width={700}
      footer={
        <Button type="primary" onClick={onClose}>
          {t('common.close', 'Close')}
        </Button>
      }
    >
      <Spin spinning={isLoading || isSaving}>
        {error && (
          <Alert
            message="Error"
            description={error}
            type="error"
            showIcon
            closable
            style={{ marginBottom: 16 }}
          />
        )}

        {/* Service Status */}
        <Card size="small" style={{ marginBottom: 16 }}>
          <Space direction="vertical" style={{ width: '100%' }}>
            <Space style={{ justifyContent: 'space-between', width: '100%' }}>
              <Space>
                <Text strong>{t('ai.status', 'Status')}:</Text>
                {getStatusTag(serviceInfo?.status || 'disabled')}
              </Space>
              <Switch
                checked={config?.enabled}
                onChange={(checked) => setEnabled(checked)}
                checkedChildren={t('common.enabled', 'Enabled')}
                unCheckedChildren={t('common.disabled', 'Disabled')}
              />
            </Space>

            {serviceInfo?.provider && (
              <div style={{ marginTop: 8 }}>
                <Space style={{ marginBottom: 8 }}>
                  <Text type="secondary">{t('ai.current_provider', 'Current Provider')}:</Text>
                  <Tag color="blue">{serviceInfo.provider.name}</Tag>
                </Space>
                <div>
                  <Space>
                    <Text type="secondary">{t('ai.model', 'Model')}:</Text>
                    {currentProviderModels.length > 0 ? (
                      <Select
                        style={{ minWidth: 200 }}
                        value={serviceInfo.provider.model}
                        onChange={handleChangeCurrentModel}
                        loading={isChangingModel}
                        disabled={isChangingModel}
                        showSearch
                        filterOption={(input, option) =>
                          (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
                        }
                        options={currentProviderModels.map((model) => ({
                          label: model,
                          value: model,
                        }))}
                      />
                    ) : (
                      <Text>{serviceInfo.provider.model}</Text>
                    )}
                  </Space>
                </div>
              </div>
            )}
          </Space>
        </Card>

        {/* Preferred Source */}
        <Card
          size="small"
          title={t('ai.preferred_source', 'Preferred Source')}
          style={{ marginBottom: 16 }}
        >
          <Select
            style={{ width: '100%' }}
            value={config?.preferredSource || 'auto'}
            onChange={(value: LLMSource) => setPreferredSource(value)}
            options={[
              { label: t('ai.source_auto', 'Auto (Detect Best)'), value: 'auto' },
              { label: t('ai.source_local', 'Local Services First'), value: 'local' },
              { label: t('ai.source_online', 'Online APIs First'), value: 'online' },
            ]}
          />
        </Card>

        {/* Local Services Discovery */}
        <Card
          size="small"
          title={
            <Space>
              <DesktopOutlined />
              {t('ai.local_services', 'Local LLM Services')}
            </Space>
          }
          extra={
            <Button
              size="small"
              icon={<SyncOutlined spin={isDiscovering} />}
              onClick={() => discoverServices()}
            >
              {t('ai.refresh', 'Refresh')}
            </Button>
          }
          style={{ marginBottom: 16 }}
        >
          {discoveredServices.length === 0 ? (
            <Text type="secondary">
              {isDiscovering
                ? t('ai.discovering', 'Discovering services...')
                : t('ai.no_services', 'No local services found')}
            </Text>
          ) : (
            <Space direction="vertical" style={{ width: '100%' }}>
              {discoveredServices.map((service) => (
                <Card key={service.endpoint} size="small" style={{ marginBottom: 8 }}>
                  <Space style={{ justifyContent: 'space-between', width: '100%' }}>
                    <Space>
                      {getStatusIcon(service.status)}
                      <Text strong>{service.name}</Text>
                      <Text type="secondary">({service.endpoint})</Text>
                    </Space>
                    {service.status === 'running' && (
                      <Button
                        size="small"
                        type="primary"
                        onClick={() => handleUseService(service)}
                      >
                        {t('ai.use_service', 'Use')}
                      </Button>
                    )}
                  </Space>
                  {service.models.length > 0 && (
                    <div style={{ marginTop: 8 }}>
                      <Text type="secondary">{t('ai.models', 'Models')}: </Text>
                      {service.models.slice(0, 5).map((model) => (
                        <Tag key={model} style={{ marginBottom: 4 }}>
                          {model}
                        </Tag>
                      ))}
                      {service.models.length > 5 && (
                        <Tag>+{service.models.length - 5} more</Tag>
                      )}
                    </div>
                  )}
                </Card>
              ))}
            </Space>
          )}
        </Card>

        {/* Online Providers */}
        <Collapse defaultActiveKey={[]} style={{ marginBottom: 16 }}>
          <Panel
            header={
              <Space>
                <CloudOutlined />
                {t('ai.online_providers', 'Online Providers')}
              </Space>
            }
            key="online"
          >
            {/* OpenAI */}
            <Card size="small" title="OpenAI" style={{ marginBottom: 16 }}>
              <Form form={openaiForm} layout="vertical" size="small">
                <Form.Item name="enabled" valuePropName="checked">
                  <Switch checkedChildren="Enabled" unCheckedChildren="Disabled" />
                </Form.Item>
                <Form.Item
                  name="apiKey"
                  label="API Key"
                  rules={[{ required: openaiForm.getFieldValue('enabled') }]}
                >
                  <Input.Password placeholder="sk-..." />
                </Form.Item>
                <Form.Item name="model" label="Model">
                  <Select
                    options={[
                      { label: 'gpt-4o-mini', value: 'gpt-4o-mini' },
                      { label: 'gpt-4o', value: 'gpt-4o' },
                      { label: 'gpt-4-turbo', value: 'gpt-4-turbo' },
                      { label: 'gpt-3.5-turbo', value: 'gpt-3.5-turbo' },
                    ]}
                  />
                </Form.Item>
                <Space>
                  <Button onClick={handleSaveOpenAI}>Save</Button>
                  <Button
                    loading={testingProvider === 'openai'}
                    onClick={() => handleTest('openai', openaiForm.getFieldsValue())}
                  >
                    Test Connection
                  </Button>
                </Space>
              </Form>
            </Card>

            {/* Claude */}
            <Card size="small" title="Claude (Anthropic)" style={{ marginBottom: 16 }}>
              <Form form={claudeForm} layout="vertical" size="small">
                <Form.Item name="enabled" valuePropName="checked">
                  <Switch checkedChildren="Enabled" unCheckedChildren="Disabled" />
                </Form.Item>
                <Form.Item
                  name="apiKey"
                  label="API Key"
                  rules={[{ required: claudeForm.getFieldValue('enabled') }]}
                >
                  <Input.Password placeholder="sk-ant-..." />
                </Form.Item>
                <Form.Item name="model" label="Model">
                  <Select
                    options={[
                      { label: 'claude-3-haiku', value: 'claude-3-haiku-20240307' },
                      { label: 'claude-3-sonnet', value: 'claude-3-sonnet-20240229' },
                      { label: 'claude-3-opus', value: 'claude-3-opus-20240229' },
                      { label: 'claude-3.5-sonnet', value: 'claude-3-5-sonnet-20241022' },
                    ]}
                  />
                </Form.Item>
                <Space>
                  <Button onClick={handleSaveClaude}>Save</Button>
                  <Button
                    loading={testingProvider === 'claude'}
                    onClick={() => handleTest('claude', claudeForm.getFieldsValue())}
                  >
                    Test Connection
                  </Button>
                </Space>
              </Form>
            </Card>

            {/* Custom Provider */}
            <Card size="small" title="Custom (OpenAI Compatible)">
              <Form form={customForm} layout="vertical" size="small">
                <Form.Item name="enabled" valuePropName="checked">
                  <Switch checkedChildren="Enabled" unCheckedChildren="Disabled" />
                </Form.Item>
                <Form.Item
                  name="endpoint"
                  label="Endpoint"
                  rules={[{ required: customForm.getFieldValue('enabled') }]}
                >
                  <Input placeholder="https://api.example.com/v1" />
                </Form.Item>
                <Form.Item name="apiKey" label="API Key (Optional)">
                  <Input.Password />
                </Form.Item>
                <Form.Item name="model" label="Model">
                  <Input placeholder="model-name" />
                </Form.Item>
                <Space>
                  <Button onClick={handleSaveCustom}>Save</Button>
                  <Button
                    loading={testingProvider === 'custom'}
                    onClick={() => handleTest('custom', customForm.getFieldsValue())}
                  >
                    Test Connection
                  </Button>
                </Space>
              </Form>
            </Card>
          </Panel>
        </Collapse>

        {/* AI Features */}
        <Card
          size="small"
          title={
            <Space>
              <ExperimentOutlined />
              {t('ai.features', 'AI Features')}
            </Space>
          }
        >
          <Space direction="vertical" style={{ width: '100%' }}>
            <Space style={{ justifyContent: 'space-between', width: '100%' }}>
              <Text>{t('ai.feature_log_analysis', 'Smart Log Analysis')}</Text>
              <Switch
                checked={config?.features?.logAnalysis}
                onChange={(checked) => setFeature('logAnalysis', checked)}
              />
            </Space>
            <Space style={{ justifyContent: 'space-between', width: '100%' }}>
              <Text>{t('ai.feature_natural_search', 'Natural Language Search')}</Text>
              <Switch
                checked={config?.features?.naturalSearch}
                onChange={(checked) => setFeature('naturalSearch', checked)}
              />
            </Space>
            <Space style={{ justifyContent: 'space-between', width: '100%' }}>
              <Text>{t('ai.feature_workflow_gen', 'Workflow Auto-Generation')}</Text>
              <Switch
                checked={config?.features?.workflowGeneration}
                onChange={(checked) => setFeature('workflowGeneration', checked)}
              />
            </Space>
            <Space style={{ justifyContent: 'space-between', width: '100%' }}>
              <Text>{t('ai.feature_workflow_ai', 'AI-Assisted Workflow Execution')}</Text>
              <Switch
                checked={config?.features?.workflowAI}
                onChange={(checked) => setFeature('workflowAI', checked)}
              />
            </Space>
            <Space style={{ justifyContent: 'space-between', width: '100%' }}>
              <Text>{t('ai.feature_crash_analysis', 'Crash Root Cause Analysis')}</Text>
              <Switch
                checked={config?.features?.crashAnalysis}
                onChange={(checked) => setFeature('crashAnalysis', checked)}
              />
            </Space>
            <Space style={{ justifyContent: 'space-between', width: '100%' }}>
              <Text>{t('ai.feature_assertion_gen', 'Auto Assertion Generation')}</Text>
              <Switch
                checked={config?.features?.assertionGen}
                onChange={(checked) => setFeature('assertionGen', checked)}
              />
            </Space>
            <Space style={{ justifyContent: 'space-between', width: '100%' }}>
              <Text>{t('ai.feature_video_analysis', 'Video Frame Analysis')}</Text>
              <Switch
                checked={config?.features?.videoAnalysis}
                onChange={(checked) => setFeature('videoAnalysis', checked)}
              />
            </Space>
          </Space>
        </Card>
      </Spin>
    </Modal>
  );
};

export default LLMConfig;
