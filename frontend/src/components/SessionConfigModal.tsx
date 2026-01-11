import React, { useState, useEffect } from 'react';
import {
  Modal,
  Form,
  Input,
  Switch,
  Select,
  InputNumber,
  Space,
  Typography,
  Divider,
  Card,
  Spin,
} from 'antd';
import {
  FileTextOutlined,
  VideoCameraOutlined,
  GlobalOutlined,
  DashboardOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { SessionConfig, defaultSessionConfig } from '../stores/eventTypes';
import { GetInstalledPackages } from '../../wailsjs/go/main/App';

const { Text } = Typography;

interface SessionConfigModalProps {
  open: boolean;
  onCancel: () => void;
  onStart: (name: string, config: SessionConfig) => void;
  defaultName?: string;
  deviceId?: string;
}

const SessionConfigModal: React.FC<SessionConfigModalProps> = ({
  open,
  onCancel,
  onStart,
  defaultName,
  deviceId,
}) => {
  const { t } = useTranslation();
  const [form] = Form.useForm();
  const [config, setConfig] = useState<SessionConfig>(defaultSessionConfig);
  const [packages, setPackages] = useState<string[]>([]);
  const [loadingPackages, setLoadingPackages] = useState(false);

  useEffect(() => {
    if (open) {
      // Reset form when modal opens
      const name = defaultName || `Session ${new Date().toLocaleTimeString()}`;
      form.setFieldsValue({ name });
      setConfig(defaultSessionConfig);

      // Fetch installed packages
      if (deviceId) {
        setLoadingPackages(true);
        GetInstalledPackages(deviceId, true)
          .then((pkgs) => {
            setPackages(pkgs || []);
          })
          .catch((err) => {
            console.error('Failed to fetch packages:', err);
            setPackages([]);
          })
          .finally(() => {
            setLoadingPackages(false);
          });
      }
    }
  }, [open, defaultName, form, deviceId]);

  const handleOk = () => {
    form.validateFields().then((values) => {
      onStart(values.name, config);
    });
  };

  const updateConfig = (path: string[], value: any) => {
    setConfig((prev) => {
      const newConfig = { ...prev };
      let current: any = newConfig;
      for (let i = 0; i < path.length - 1; i++) {
        current[path[i]] = { ...current[path[i]] };
        current = current[path[i]];
      }
      current[path[path.length - 1]] = value;
      return newConfig;
    });
  };

  const packageOptions = packages.map((pkg) => ({
    label: pkg,
    value: pkg,
  }));

  return (
    <Modal
      title={t('session.start_new')}
      open={open}
      onCancel={onCancel}
      onOk={handleOk}
      okText={t('session.start')}
      cancelText={t('common.cancel')}
      width={480}
    >
      <Form form={form} layout="vertical">
        <Form.Item
          name="name"
          label={t('session.session_name')}
          rules={[{ required: true, message: t('session.name_required') }]}
        >
          <Input placeholder={t('session.session_name')} />
        </Form.Item>
      </Form>

      <Divider style={{ margin: '16px 0' }} />

      {/* Logcat Config */}
      <Card
        size="small"
        style={{ marginBottom: 12 }}
        bodyStyle={{ padding: 12 }}
      >
        <Space style={{ width: '100%', justifyContent: 'space-between' }}>
          <Space>
            <FileTextOutlined style={{ fontSize: 18, color: '#52c41a' }} />
            <Text strong>{t('session.logcat')}</Text>
          </Space>
          <Switch
            checked={config.logcat.enabled}
            onChange={(checked) => updateConfig(['logcat', 'enabled'], checked)}
          />
        </Space>
        {config.logcat.enabled && (
          <div style={{ marginTop: 12 }}>
            <Form.Item label={t('session.package_name')} style={{ marginBottom: 8 }}>
              <Select
                showSearch
                allowClear
                placeholder={t('session.all_packages')}
                value={config.logcat.packageName || undefined}
                onChange={(value) =>
                  updateConfig(['logcat', 'packageName'], value || '')
                }
                options={packageOptions}
                loading={loadingPackages}
                notFoundContent={loadingPackages ? <Spin size="small" /> : t('logcat.apps_placeholder')}
                filterOption={(input, option) =>
                  (option?.label as string)?.toLowerCase().includes(input.toLowerCase())
                }
                style={{ width: '100%' }}
              />
            </Form.Item>
            <Space style={{ width: '100%' }}>
              <Form.Item label={t('session.pre_filter')} style={{ marginBottom: 0, flex: 1 }}>
                <Input
                  placeholder={t('session.filter_keyword')}
                  value={config.logcat.preFilter}
                  onChange={(e) =>
                    updateConfig(['logcat', 'preFilter'], e.target.value)
                  }
                />
              </Form.Item>
              <Form.Item
                label={t('session.exclude')}
                style={{ marginBottom: 0, flex: 1 }}
              >
                <Input
                  placeholder={t('session.exclude_keyword')}
                  value={config.logcat.excludeFilter}
                  onChange={(e) =>
                    updateConfig(['logcat', 'excludeFilter'], e.target.value)
                  }
                />
              </Form.Item>
            </Space>
          </div>
        )}
      </Card>

      {/* Recording Config */}
      <Card
        size="small"
        style={{ marginBottom: 12 }}
        bodyStyle={{ padding: 12 }}
      >
        <Space style={{ width: '100%', justifyContent: 'space-between' }}>
          <Space>
            <VideoCameraOutlined style={{ fontSize: 18, color: '#f5222d' }} />
            <Text strong>{t('session.screen_recording')}</Text>
          </Space>
          <Switch
            checked={config.recording.enabled}
            onChange={(checked) =>
              updateConfig(['recording', 'enabled'], checked)
            }
          />
        </Space>
        {config.recording.enabled && (
          <div style={{ marginTop: 12 }}>
            <Form.Item label={t('session.quality')} style={{ marginBottom: 0 }}>
              <Select
                value={config.recording.quality}
                onChange={(value) =>
                  updateConfig(['recording', 'quality'], value)
                }
                options={[
                  { label: t('session.quality_low'), value: 'low' },
                  { label: t('session.quality_medium'), value: 'medium' },
                  { label: t('session.quality_high'), value: 'high' },
                ]}
              />
            </Form.Item>
          </div>
        )}
      </Card>

      {/* Proxy Config */}
      <Card size="small" style={{ marginBottom: 12 }} bodyStyle={{ padding: 12 }}>
        <Space style={{ width: '100%', justifyContent: 'space-between' }}>
          <Space>
            <GlobalOutlined style={{ fontSize: 18, color: '#722ed1' }} />
            <Text strong>{t('session.network_proxy')}</Text>
          </Space>
          <Switch
            checked={config.proxy.enabled}
            onChange={(checked) => updateConfig(['proxy', 'enabled'], checked)}
          />
        </Space>
        {config.proxy.enabled && (
          <div style={{ marginTop: 12 }}>
            <Space style={{ width: '100%' }}>
              <Form.Item label={t('session.port')} style={{ marginBottom: 0 }}>
                <InputNumber
                  min={1024}
                  max={65535}
                  value={config.proxy.port}
                  onChange={(value) =>
                    updateConfig(['proxy', 'port'], value || 8080)
                  }
                />
              </Form.Item>
              <Form.Item label={t('session.https_mitm')} style={{ marginBottom: 0 }}>
                <Switch
                  checked={config.proxy.mitmEnabled}
                  onChange={(checked) =>
                    updateConfig(['proxy', 'mitmEnabled'], checked)
                  }
                />
              </Form.Item>
            </Space>
          </div>
        )}
      </Card>

      {/* Monitor Config */}
      <Card size="small" bodyStyle={{ padding: 12 }}>
        <Space style={{ width: '100%', justifyContent: 'space-between' }}>
          <Space>
            <DashboardOutlined style={{ fontSize: 18, color: '#13c2c2' }} />
            <Text strong>{t('session.device_monitor')}</Text>
          </Space>
          <Switch
            checked={config.monitor.enabled}
            onChange={(checked) => updateConfig(['monitor', 'enabled'], checked)}
          />
        </Space>
        <Text type="secondary" style={{ display: 'block', marginTop: 8, fontSize: 12 }}>
          {t('session.monitor_desc')}
        </Text>
      </Card>

      <Text
        type="secondary"
        style={{ display: 'block', marginTop: 12, fontSize: 12 }}
      >
        {t('session.auto_feature_note')}
      </Text>
    </Modal>
  );
};

export default SessionConfigModal;
