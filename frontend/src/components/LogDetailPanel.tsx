import { memo, useCallback } from 'react';
import { Card, Tag, Typography, Button, Space, message, Tabs } from 'antd';
import { CloseOutlined, CopyOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useLogcatStore, type ParsedLog } from '../stores/logcatStore';

const { Text } = Typography;

interface LogDetailPanelProps {
  log: ParsedLog | null;
  onClose: () => void;
}

// 日志级别配置
const levelConfig: Record<string, { color: string; label: string }> = {
  'V': { color: '#8c8c8c', label: 'VERBOSE' },
  'D': { color: '#52c41a', label: 'DEBUG' },
  'I': { color: '#1890ff', label: 'INFO' },
  'W': { color: '#faad14', label: 'WARNING' },
  'E': { color: '#ff4d4f', label: 'ERROR' },
  'F': { color: '#cf1322', label: 'FATAL' },
};

const LogDetailPanel = memo(({ log, onClose }: LogDetailPanelProps) => {
  const { t } = useTranslation();
  const activeTab = useLogcatStore((s) => s.activeDetailTab);
  const setActiveTab = useLogcatStore((s) => s.setActiveDetailTab);

  const handleCopy = useCallback((text: string) => {
    navigator.clipboard.writeText(text);
    message.success(t('common.copied') || 'Copied!');
  }, [t]);

  if (!log) return null;

  const config = levelConfig[log.level] || levelConfig['V'];
  const messageText = log.message || log.raw;

  // Tab 内容的公共样式
  const contentStyle: React.CSSProperties = {
    height: '100%',
    overflow: 'auto',
    backgroundColor: '#1e1e1e',
    borderRadius: 4,
    padding: 12,
  };

  const textStyle: React.CSSProperties = {
    margin: 0,
    fontFamily: '"JetBrains Mono", "Fira Code", monospace',
    fontSize: 12,
    lineHeight: 1.6,
    whiteSpace: 'pre-wrap',
    wordBreak: 'break-all',
    color: '#d4d4d4',
  };

  const tabItems = [
    {
      key: 'message',
      label: t('logcat.detail.message') || 'Message',
      children: (
        <div style={contentStyle}>
          <pre style={textStyle}>{messageText}</pre>
        </div>
      ),
    },
    {
      key: 'raw',
      label: t('logcat.detail.raw') || 'Raw',
      children: (
        <div style={contentStyle}>
          <pre style={textStyle}>{log.raw}</pre>
        </div>
      ),
    },
  ];

  return (
    <Card
      size="small"
      title={
        <Space size={8}>
          <Tag color={config.color} style={{ margin: 0 }}>
            {config.label}
          </Tag>
          <Text strong style={{ fontSize: 13 }} ellipsis>
            {log.tag || 'Unknown'}
          </Text>
        </Space>
      }
      extra={
        <Space size={4}>
          <Button
            type="text"
            size="small"
            icon={<CopyOutlined />}
            onClick={() => handleCopy(activeTab === 'message' ? messageText : log.raw)}
          />
          <Button type="text" size="small" icon={<CloseOutlined />} onClick={onClose} />
        </Space>
      }
      style={{
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        borderLeft: `3px solid ${config.color}`,
      }}
      styles={{
        body: {
          flex: 1,
          overflow: 'hidden',
          padding: 0,
          display: 'flex',
          flexDirection: 'column',
        }
      }}
    >
      {/* 紧凑的元信息行 */}
      <div style={{ 
        padding: '8px 12px', 
        borderBottom: '1px solid #303030',
        display: 'flex',
        gap: 16,
        fontSize: 11,
        color: '#888',
        flexShrink: 0,
        backgroundColor: '#1a1a1a',
      }}>
        <span>
          <Text type="secondary" style={{ fontSize: 11 }}>Time: </Text>
          <Text style={{ fontFamily: 'monospace', fontSize: 11 }}>{log.date} {log.timestamp}</Text>
        </span>
        {log.pid && (
          <span>
            <Text type="secondary" style={{ fontSize: 11 }}>PID: </Text>
            <Text style={{ fontFamily: 'monospace', fontSize: 11 }}>{log.pid}</Text>
          </span>
        )}
        {log.packageName && (
          <span style={{ flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
            <Text type="secondary" style={{ fontSize: 11 }}>Pkg: </Text>
            <Text style={{ fontFamily: 'monospace', fontSize: 11 }}>{log.packageName}</Text>
          </span>
        )}
      </div>

      {/* Tabs 占满剩余空间 */}
      <Tabs
        activeKey={activeTab}
        onChange={(key) => setActiveTab(key as 'message' | 'raw')}
        items={tabItems}
        size="small"
        style={{ 
          flex: 1, 
          display: 'flex', 
          flexDirection: 'column',
          overflow: 'hidden',
        }}
        tabBarStyle={{ 
          margin: 0, 
          padding: '0 12px',
          flexShrink: 0,
        }}
        className="log-detail-tabs"
      />

      {/* 添加样式让 Tab 内容占满高度 */}
      <style>{`
        .log-detail-tabs .ant-tabs-content {
          flex: 1;
          height: 100%;
        }
        .log-detail-tabs .ant-tabs-content-holder {
          flex: 1;
          overflow: hidden;
        }
        .log-detail-tabs .ant-tabs-tabpane {
          height: 100%;
          padding: 8px 12px 12px;
        }
      `}</style>
    </Card>
  );
});

LogDetailPanel.displayName = 'LogDetailPanel';

export default LogDetailPanel;
