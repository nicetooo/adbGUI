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
      title="Start New Session"
      open={open}
      onCancel={onCancel}
      onOk={handleOk}
      okText="Start Session"
      cancelText="Cancel"
      width={480}
    >
      <Form form={form} layout="vertical">
        <Form.Item
          name="name"
          label="Session Name"
          rules={[{ required: true, message: 'Please enter a session name' }]}
        >
          <Input placeholder="Session name" />
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
            <Text strong>Logcat</Text>
          </Space>
          <Switch
            checked={config.logcat.enabled}
            onChange={(checked) => updateConfig(['logcat', 'enabled'], checked)}
          />
        </Space>
        {config.logcat.enabled && (
          <div style={{ marginTop: 12 }}>
            <Form.Item label="Package Name" style={{ marginBottom: 8 }}>
              <Select
                showSearch
                allowClear
                placeholder="All packages (optional)"
                value={config.logcat.packageName || undefined}
                onChange={(value) =>
                  updateConfig(['logcat', 'packageName'], value || '')
                }
                options={packageOptions}
                loading={loadingPackages}
                notFoundContent={loadingPackages ? <Spin size="small" /> : 'No packages found'}
                filterOption={(input, option) =>
                  (option?.label as string)?.toLowerCase().includes(input.toLowerCase())
                }
                style={{ width: '100%' }}
              />
            </Form.Item>
            <Space style={{ width: '100%' }}>
              <Form.Item label="Pre-filter" style={{ marginBottom: 0, flex: 1 }}>
                <Input
                  placeholder="Filter keyword"
                  value={config.logcat.preFilter}
                  onChange={(e) =>
                    updateConfig(['logcat', 'preFilter'], e.target.value)
                  }
                />
              </Form.Item>
              <Form.Item
                label="Exclude"
                style={{ marginBottom: 0, flex: 1 }}
              >
                <Input
                  placeholder="Exclude keyword"
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
            <Text strong>Screen Recording</Text>
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
            <Form.Item label="Quality" style={{ marginBottom: 0 }}>
              <Select
                value={config.recording.quality}
                onChange={(value) =>
                  updateConfig(['recording', 'quality'], value)
                }
                options={[
                  { label: 'Low (480p, 2Mbps)', value: 'low' },
                  { label: 'Medium (720p, 4Mbps)', value: 'medium' },
                  { label: 'High (1080p, 8Mbps)', value: 'high' },
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
            <Text strong>Network Proxy</Text>
          </Space>
          <Switch
            checked={config.proxy.enabled}
            onChange={(checked) => updateConfig(['proxy', 'enabled'], checked)}
          />
        </Space>
        {config.proxy.enabled && (
          <div style={{ marginTop: 12 }}>
            <Space style={{ width: '100%' }}>
              <Form.Item label="Port" style={{ marginBottom: 0 }}>
                <InputNumber
                  min={1024}
                  max={65535}
                  value={config.proxy.port}
                  onChange={(value) =>
                    updateConfig(['proxy', 'port'], value || 8080)
                  }
                />
              </Form.Item>
              <Form.Item label="HTTPS MITM" style={{ marginBottom: 0 }}>
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
            <Text strong>Device Monitor</Text>
          </Space>
          <Switch
            checked={config.monitor.enabled}
            onChange={(checked) => updateConfig(['monitor', 'enabled'], checked)}
          />
        </Space>
        <Text type="secondary" style={{ display: 'block', marginTop: 8, fontSize: 12 }}>
          Monitor battery, network, screen state, app lifecycle, touch events, and performance metrics
        </Text>
      </Card>

      <Text
        type="secondary"
        style={{ display: 'block', marginTop: 12, fontSize: 12 }}
      >
        Selected features will start automatically with the session and stop
        when the session ends.
      </Text>
    </Modal>
  );
};

export default SessionConfigModal;
