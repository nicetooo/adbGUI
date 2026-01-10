import React, { useState, useEffect, useCallback, useMemo } from 'react';
import {
  Table,
  Card,
  Button,
  Space,
  Modal,
  Tag,
  Tooltip,
  Popconfirm,
  message,
  Typography,
  Input,
  Select,
  DatePicker,
  Statistic,
  Row,
  Col,
  Descriptions,
  Empty,
} from 'antd';
import {
  PlayCircleOutlined,
  DeleteOutlined,
  InfoCircleOutlined,
  ReloadOutlined,
  SearchOutlined,
  VideoCameraOutlined,
  ClockCircleOutlined,
  FileTextOutlined,
  FilterOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { useDeviceStore } from '../stores';
import type { DeviceSession } from '../stores/eventTypes';

const { Text, Title } = Typography;
const { RangePicker } = DatePicker;

// Format duration from ms to readable string
const formatDuration = (ms: number): string => {
  if (!ms || ms <= 0) return '-';
  const seconds = Math.floor(ms / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);

  if (hours > 0) {
    return `${hours}h ${minutes % 60}m ${seconds % 60}s`;
  } else if (minutes > 0) {
    return `${minutes}m ${seconds % 60}s`;
  } else {
    return `${seconds}s`;
  }
};

// Format timestamp to readable date
const formatTime = (timestamp: number): string => {
  if (!timestamp) return '-';
  return new Date(timestamp).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
};

// Get status color
const getStatusColor = (status: string): string => {
  switch (status) {
    case 'active': return 'processing';
    case 'completed': return 'success';
    case 'failed': return 'error';
    case 'cancelled': return 'warning';
    default: return 'default';
  }
};

// Get status text
const getStatusText = (status: string): string => {
  switch (status) {
    case 'active': return '录制中';
    case 'completed': return '已完成';
    case 'failed': return '失败';
    case 'cancelled': return '已取消';
    default: return status;
  }
};

interface SessionManagerProps {
  style?: React.CSSProperties;
}

const SessionManager: React.FC<SessionManagerProps> = ({ style }) => {
  const { selectedDevice } = useDeviceStore();

  // State
  const [sessions, setSessions] = useState<DeviceSession[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedSession, setSelectedSession] = useState<DeviceSession | null>(null);
  const [detailModalOpen, setDetailModalOpen] = useState(false);
  const [videoModalOpen, setVideoModalOpen] = useState(false);
  const [videoPath, setVideoPath] = useState<string>('');
  const [searchText, setSearchText] = useState('');
  const [statusFilter, setStatusFilter] = useState<string | null>(null);
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);

  // Load sessions
  const loadSessions = useCallback(async () => {
    setLoading(true);
    try {
      // Load all sessions (empty deviceId = all devices)
      const list = await (window as any).go.main.App.ListStoredSessions('', 100);
      setSessions(list || []);
    } catch (err) {
      console.error('Failed to load sessions:', err);
      message.error('加载 Session 列表失败');
    } finally {
      setLoading(false);
    }
  }, []);

  // Initial load
  useEffect(() => {
    loadSessions();
  }, [loadSessions]);

  // Delete session
  const handleDelete = useCallback(async (sessionId: string) => {
    try {
      await (window as any).go.main.App.DeleteStoredSession(sessionId);
      message.success('删除成功');
      loadSessions();
    } catch (err) {
      console.error('Failed to delete session:', err);
      message.error('删除失败');
    }
  }, [loadSessions]);

  // Batch delete
  const handleBatchDelete = useCallback(async () => {
    if (selectedRowKeys.length === 0) return;

    try {
      for (const key of selectedRowKeys) {
        await (window as any).go.main.App.DeleteStoredSession(key as string);
      }
      message.success(`成功删除 ${selectedRowKeys.length} 个 Session`);
      setSelectedRowKeys([]);
      loadSessions();
    } catch (err) {
      console.error('Failed to batch delete:', err);
      message.error('批量删除失败');
    }
  }, [selectedRowKeys, loadSessions]);

  // View session detail
  const handleViewDetail = useCallback((session: DeviceSession) => {
    setSelectedSession(session);
    setDetailModalOpen(true);
  }, []);

  // Play video
  const handlePlayVideo = useCallback((session: DeviceSession) => {
    if (session.videoPath) {
      setVideoPath(session.videoPath);
      setSelectedSession(session);
      setVideoModalOpen(true);
    } else {
      message.warning('该 Session 没有录屏文件');
    }
  }, []);

  // Filter sessions
  const filteredSessions = useMemo(() => {
    return sessions.filter(session => {
      // Search filter
      if (searchText) {
        const search = searchText.toLowerCase();
        if (!session.name.toLowerCase().includes(search) &&
            !session.id.toLowerCase().includes(search) &&
            !session.deviceId.toLowerCase().includes(search)) {
          return false;
        }
      }
      // Status filter
      if (statusFilter && session.status !== statusFilter) {
        return false;
      }
      return true;
    });
  }, [sessions, searchText, statusFilter]);

  // Statistics
  const stats = useMemo(() => {
    const total = sessions.length;
    const active = sessions.filter(s => s.status === 'active').length;
    const completed = sessions.filter(s => s.status === 'completed').length;
    const withVideo = sessions.filter(s => !!s.videoPath).length;
    const totalEvents = sessions.reduce((sum, s) => sum + (s.eventCount || 0), 0);
    return { total, active, completed, withVideo, totalEvents };
  }, [sessions]);

  // Table columns
  const columns: ColumnsType<DeviceSession> = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      width: 200,
      ellipsis: true,
      render: (name, record) => (
        <Space>
          <Text strong>{name || 'Unnamed'}</Text>
          {record.videoPath && (
            <Tooltip title="有录屏">
              <VideoCameraOutlined style={{ color: '#1890ff' }} />
            </Tooltip>
          )}
        </Space>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status) => (
        <Tag color={getStatusColor(status)}>{getStatusText(status)}</Tag>
      ),
    },
    {
      title: '开始时间',
      dataIndex: 'startTime',
      key: 'startTime',
      width: 180,
      sorter: (a, b) => a.startTime - b.startTime,
      defaultSortOrder: 'descend',
      render: (time) => formatTime(time),
    },
    {
      title: '时长',
      key: 'duration',
      width: 100,
      render: (_, record) => {
        const duration = record.endTime > 0
          ? record.endTime - record.startTime
          : Date.now() - record.startTime;
        return formatDuration(duration);
      },
    },
    {
      title: '事件数',
      dataIndex: 'eventCount',
      key: 'eventCount',
      width: 80,
      sorter: (a, b) => (a.eventCount || 0) - (b.eventCount || 0),
      render: (count) => count || 0,
    },
    {
      title: '设备',
      dataIndex: 'deviceId',
      key: 'deviceId',
      width: 150,
      ellipsis: true,
      render: (deviceId) => (
        <Tooltip title={deviceId}>
          <Text type="secondary" style={{ fontSize: 12 }}>
            {deviceId?.split('.')[0] || deviceId}
          </Text>
        </Tooltip>
      ),
    },
    {
      title: '操作',
      key: 'actions',
      width: 150,
      fixed: 'right',
      render: (_, record) => (
        <Space size="small">
          {record.videoPath && (
            <Tooltip title="播放录屏">
              <Button
                type="text"
                size="small"
                icon={<PlayCircleOutlined />}
                onClick={() => handlePlayVideo(record)}
              />
            </Tooltip>
          )}
          <Tooltip title="查看详情">
            <Button
              type="text"
              size="small"
              icon={<InfoCircleOutlined />}
              onClick={() => handleViewDetail(record)}
            />
          </Tooltip>
          <Popconfirm
            title="确定删除该 Session？"
            description="删除后将无法恢复"
            onConfirm={() => handleDelete(record.id)}
            okText="删除"
            cancelText="取消"
            okButtonProps={{ danger: true }}
          >
            <Tooltip title="删除">
              <Button
                type="text"
                size="small"
                danger
                icon={<DeleteOutlined />}
              />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div style={{ padding: 16, height: '100%', overflow: 'auto', ...style }}>
      {/* Statistics Cards */}
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={6}>
          <Card size="small">
            <Statistic title="总 Sessions" value={stats.total} prefix={<FileTextOutlined />} />
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small">
            <Statistic
              title="录制中"
              value={stats.active}
              valueStyle={{ color: '#1890ff' }}
              prefix={<ClockCircleOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small">
            <Statistic
              title="有录屏"
              value={stats.withVideo}
              valueStyle={{ color: '#52c41a' }}
              prefix={<VideoCameraOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small">
            <Statistic title="总事件数" value={stats.totalEvents} />
          </Card>
        </Col>
      </Row>

      {/* Main Table Card */}
      <Card
        title="Session 管理"
        extra={
          <Space>
            <Input
              placeholder="搜索名称/ID"
              prefix={<SearchOutlined />}
              value={searchText}
              onChange={e => setSearchText(e.target.value)}
              style={{ width: 200 }}
              allowClear
            />
            <Select
              placeholder="状态筛选"
              value={statusFilter}
              onChange={setStatusFilter}
              style={{ width: 120 }}
              allowClear
              options={[
                { label: '录制中', value: 'active' },
                { label: '已完成', value: 'completed' },
                { label: '失败', value: 'failed' },
                { label: '已取消', value: 'cancelled' },
              ]}
            />
            <Button icon={<ReloadOutlined />} onClick={loadSessions}>
              刷新
            </Button>
            {selectedRowKeys.length > 0 && (
              <Popconfirm
                title={`确定删除选中的 ${selectedRowKeys.length} 个 Session？`}
                onConfirm={handleBatchDelete}
                okText="删除"
                cancelText="取消"
                okButtonProps={{ danger: true }}
              >
                <Button danger>
                  批量删除 ({selectedRowKeys.length})
                </Button>
              </Popconfirm>
            )}
          </Space>
        }
      >
        <Table
          columns={columns}
          dataSource={filteredSessions}
          rowKey="id"
          loading={loading}
          size="small"
          scroll={{ x: 1000 }}
          pagination={{
            pageSize: 20,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条`,
          }}
          rowSelection={{
            selectedRowKeys,
            onChange: setSelectedRowKeys,
          }}
          locale={{
            emptyText: <Empty description="暂无 Session 数据" />,
          }}
        />
      </Card>

      {/* Detail Modal */}
      <Modal
        title="Session 详情"
        open={detailModalOpen}
        onCancel={() => setDetailModalOpen(false)}
        footer={[
          <Button key="close" onClick={() => setDetailModalOpen(false)}>
            关闭
          </Button>,
          selectedSession?.videoPath && (
            <Button
              key="play"
              type="primary"
              icon={<PlayCircleOutlined />}
              onClick={() => {
                setDetailModalOpen(false);
                handlePlayVideo(selectedSession!);
              }}
            >
              播放录屏
            </Button>
          ),
        ].filter(Boolean)}
        width={700}
      >
        {selectedSession && (
          <Descriptions bordered column={2} size="small">
            <Descriptions.Item label="ID" span={2}>
              <Text copyable={{ text: selectedSession.id }}>
                {selectedSession.id}
              </Text>
            </Descriptions.Item>
            <Descriptions.Item label="名称">{selectedSession.name || '-'}</Descriptions.Item>
            <Descriptions.Item label="状态">
              <Tag color={getStatusColor(selectedSession.status)}>
                {getStatusText(selectedSession.status)}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="开始时间">
              {formatTime(selectedSession.startTime)}
            </Descriptions.Item>
            <Descriptions.Item label="结束时间">
              {selectedSession.endTime > 0 ? formatTime(selectedSession.endTime) : '进行中'}
            </Descriptions.Item>
            <Descriptions.Item label="时长">
              {formatDuration(
                selectedSession.endTime > 0
                  ? selectedSession.endTime - selectedSession.startTime
                  : Date.now() - selectedSession.startTime
              )}
            </Descriptions.Item>
            <Descriptions.Item label="事件数量">
              {selectedSession.eventCount || 0}
            </Descriptions.Item>
            <Descriptions.Item label="设备 ID" span={2}>
              <Text copyable={{ text: selectedSession.deviceId }}>
                {selectedSession.deviceId}
              </Text>
            </Descriptions.Item>
            <Descriptions.Item label="录屏文件" span={2}>
              {selectedSession.videoPath ? (
                <Text copyable={{ text: selectedSession.videoPath }}>
                  {selectedSession.videoPath}
                </Text>
              ) : (
                <Text type="secondary">无</Text>
              )}
            </Descriptions.Item>
            {selectedSession.videoDuration && selectedSession.videoDuration > 0 && (
              <Descriptions.Item label="录屏时长">
                {formatDuration(selectedSession.videoDuration)}
              </Descriptions.Item>
            )}
            <Descriptions.Item label="配置" span={2}>
              <Space>
                {selectedSession.config?.logcat?.enabled && <Tag>Logcat</Tag>}
                {selectedSession.config?.recording?.enabled && <Tag color="blue">录屏</Tag>}
                {selectedSession.config?.proxy?.enabled && <Tag color="green">代理</Tag>}
                {selectedSession.config?.monitor?.enabled && <Tag color="orange">监控</Tag>}
              </Space>
            </Descriptions.Item>
          </Descriptions>
        )}
      </Modal>

      {/* Video Player Modal */}
      <Modal
        title={
          <Space>
            <VideoCameraOutlined />
            <span>录屏回放</span>
            {selectedSession && <Text type="secondary">- {selectedSession.name}</Text>}
          </Space>
        }
        open={videoModalOpen}
        onCancel={() => {
          setVideoModalOpen(false);
          setVideoPath('');
        }}
        footer={null}
        width={900}
        centered
        destroyOnClose
      >
        {videoPath && (
          <div style={{ textAlign: 'center' }}>
            <video
              controls
              autoPlay
              style={{
                maxWidth: '100%',
                maxHeight: '70vh',
                borderRadius: 8,
                backgroundColor: '#000',
              }}
            >
              <source src={`file://${videoPath}`} type="video/mp4" />
              您的浏览器不支持视频播放
            </video>
            <div style={{ marginTop: 12 }}>
              <Text type="secondary" copyable={{ text: videoPath }}>
                {videoPath}
              </Text>
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
};

export default SessionManager;
