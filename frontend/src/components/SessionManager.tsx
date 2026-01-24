import React, { useEffect, useCallback, useMemo } from 'react';
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
  EditOutlined,
  EyeOutlined,
  PauseCircleOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import type { ColumnsType } from 'antd/es/table';
import { useDeviceStore, useUIStore, useEventStore, VIEW_KEYS, useSessionManagerStore } from '../stores';
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

// Get status text (will be replaced with i18n inside component)
const getStatusText = (status: string, t?: (key: string) => string): string => {
  if (!t) return status;
  switch (status) {
    case 'active': return t('session_manager.status_active');
    case 'completed': return t('session_manager.status_completed');
    case 'failed': return t('session_manager.status_failed');
    case 'cancelled': return t('session_manager.status_cancelled');
    default: return status;
  }
};

interface SessionManagerProps {
  style?: React.CSSProperties;
}

const SessionManager: React.FC<SessionManagerProps> = ({ style }) => {
  const { t } = useTranslation();
  const { selectedDevice } = useDeviceStore();
  const { setSelectedKey } = useUIStore();
  const { loadSession, endSession } = useEventStore();

  // Session Manager Store
  const {
    sessions,
    loading,
    selectedSession,
    selectedRowKeys,
    searchText,
    statusFilter,
    detailModalOpen,
    renameModalOpen,
    renameSession,
    newName,
    setSessions,
    setLoading,
    setSelectedRowKeys,
    setSearchText,
    setStatusFilter,
    openDetailModal,
    closeDetailModal,
    openRenameModal,
    closeRenameModal,
    setNewName,
  } = useSessionManagerStore();

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
      message.success(t('session_manager.delete_success'));
      loadSessions();
    } catch (err) {
      console.error('Failed to delete session:', err);
      message.error(t('session_manager.delete_failed'));
    }
  }, [loadSessions, t]);

  // Batch delete
  const handleBatchDelete = useCallback(async () => {
    if (selectedRowKeys.length === 0) return;

    try {
      for (const key of selectedRowKeys) {
        await (window as any).go.main.App.DeleteStoredSession(key as string);
      }
      message.success(t('session_manager.delete_success'));
      setSelectedRowKeys([]);
      loadSessions();
    } catch (err) {
      console.error('Failed to batch delete:', err);
      message.error(t('session_manager.delete_failed'));
    }
  }, [selectedRowKeys, loadSessions, t]);

  // View session detail
  const handleViewDetail = useCallback((session: DeviceSession) => {
    openDetailModal(session);
  }, [openDetailModal]);

  // Play video with system player
  const handlePlayVideo = useCallback(async (session: DeviceSession) => {
    if (session.videoPath) {
      try {
        await (window as any).go.main.App.OpenFile(session.videoPath);
        message.success(t('session_manager.opening_video'));
      } catch (err) {
        console.error('Failed to open video:', err);
        message.error(t('session_manager.open_video_failed') + ': ' + (err as Error).message);
      }
    } else {
      message.warning(t('session_manager.no_video'));
    }
  }, [t]);

  // Open rename modal
  const handleOpenRename = useCallback((session: DeviceSession) => {
    openRenameModal(session);
  }, [openRenameModal]);

  // Rename session
  const handleRename = useCallback(async () => {
    if (!renameSession || !newName.trim()) return;

    try {
      await (window as any).go.main.App.RenameStoredSession(renameSession.id, newName.trim());
      message.success(t('session_manager.rename_success'));
      closeRenameModal();
      loadSessions();
    } catch (err) {
      console.error('Failed to rename session:', err);
      message.error(t('session_manager.rename_failed'));
    }
  }, [renameSession, newName, closeRenameModal, loadSessions, t]);

  // View session in Events tab
  const handleViewInEvents = useCallback(async (session: DeviceSession) => {
    try {
      await loadSession(session.id);
      setSelectedKey(VIEW_KEYS.EVENTS);
    } catch (err) {
      console.error('Failed to load session:', err);
      message.error(t('session_manager.load_failed'));
    }
  }, [loadSession, setSelectedKey, t]);

  // End active session
  const handleEndSession = useCallback(async (sessionId: string) => {
    try {
      await endSession(sessionId, 'completed');
      message.success(t('session_manager.end_success'));
      loadSessions();
    } catch (err) {
      console.error('Failed to end session:', err);
      message.error(t('session_manager.end_failed'));
    }
  }, [endSession, loadSessions, t]);

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
      title: t('session_manager.col_name'),
      dataIndex: 'name',
      key: 'name',
      width: 200,
      ellipsis: true,
      render: (name, record) => (
        <Space>
          <Text strong>{name || t('session_manager.unnamed')}</Text>
          {record.videoPath && (
            <Tooltip title={t('session_manager.has_video')}>
              <VideoCameraOutlined style={{ color: '#1890ff' }} />
            </Tooltip>
          )}
        </Space>
      ),
    },
    {
      title: t('session_manager.col_status'),
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status) => (
        <Tag color={getStatusColor(status)}>{getStatusText(status, t)}</Tag>
      ),
    },
    {
      title: t('session_manager.col_start_time'),
      dataIndex: 'startTime',
      key: 'startTime',
      width: 180,
      sorter: (a, b) => a.startTime - b.startTime,
      defaultSortOrder: 'descend',
      render: (time) => formatTime(time),
    },
    {
      title: t('session_manager.col_duration'),
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
      title: t('session_manager.col_event_count'),
      dataIndex: 'eventCount',
      key: 'eventCount',
      width: 80,
      sorter: (a, b) => (a.eventCount || 0) - (b.eventCount || 0),
      render: (count) => count || 0,
    },
    {
      title: t('session_manager.col_device'),
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
      title: t('session_manager.col_actions'),
      key: 'actions',
      width: 280,
      fixed: 'right',
      render: (_, record) => (
        <Space size="small">
          {record.status === 'active' && (
            <Popconfirm
              title={t('session_manager.end_confirm')}
              onConfirm={() => handleEndSession(record.id)}
              okText={t('common.ok')}
              cancelText={t('common.cancel')}
            >
              <Tooltip title={t('session_manager.end_session')}>
                <Button
                  type="text"
                  size="small"
                  danger
                  icon={<PauseCircleOutlined />}
                />
              </Tooltip>
            </Popconfirm>
          )}
          <Tooltip title={t('session_manager.view_events')}>
            <Button
              type="text"
              size="small"
              icon={<EyeOutlined />}
              onClick={() => handleViewInEvents(record)}
            />
          </Tooltip>
          {record.videoPath && (
            <Tooltip title={t('session_manager.play_recording')}>
              <Button
                type="text"
                size="small"
                icon={<PlayCircleOutlined />}
                onClick={() => handlePlayVideo(record)}
              />
            </Tooltip>
          )}
          <Tooltip title={t('session_manager.rename')}>
            <Button
              type="text"
              size="small"
              icon={<EditOutlined />}
              onClick={() => handleOpenRename(record)}
            />
          </Tooltip>
          <Tooltip title={t('session_manager.view_detail')}>
            <Button
              type="text"
              size="small"
              icon={<InfoCircleOutlined />}
              onClick={() => handleViewDetail(record)}
            />
          </Tooltip>
          <Popconfirm
            title={t('session_manager.delete_single_confirm')}
            description={t('session_manager.delete_warning')}
            onConfirm={() => handleDelete(record.id)}
            okText={t('common.delete')}
            cancelText={t('common.cancel')}
            okButtonProps={{ danger: true }}
          >
            <Tooltip title={t('common.delete')}>
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
            <Statistic title={t('session_manager.stat_total')} value={stats.total} prefix={<FileTextOutlined />} />
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small">
            <Statistic
              title={t('session_manager.stat_active')}
              value={stats.active}
              valueStyle={{ color: '#1890ff' }}
              prefix={<ClockCircleOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small">
            <Statistic
              title={t('session_manager.stat_with_video')}
              value={stats.withVideo}
              valueStyle={{ color: '#52c41a' }}
              prefix={<VideoCameraOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small">
            <Statistic title={t('session_manager.stat_total_events')} value={stats.totalEvents} />
          </Card>
        </Col>
      </Row>

      {/* Main Table Card */}
      <Card
        title={t('session_manager.title')}
        extra={
          <Space>
            <Input
              placeholder={t('session_manager.search_placeholder')}
              prefix={<SearchOutlined />}
              value={searchText}
              onChange={e => setSearchText(e.target.value)}
              style={{ width: 200 }}
              allowClear
            />
            <Select
              placeholder={t('session_manager.status_filter')}
              value={statusFilter}
              onChange={setStatusFilter}
              style={{ width: 120 }}
              allowClear
              options={[
                { label: t('session_manager.status_active'), value: 'active' },
                { label: t('session_manager.status_completed'), value: 'completed' },
                { label: t('session_manager.status_failed'), value: 'failed' },
                { label: t('session_manager.status_cancelled'), value: 'cancelled' },
              ]}
            />
            <Button icon={<ReloadOutlined />} onClick={loadSessions}>
              {t('session_manager.refresh')}
            </Button>
            {selectedRowKeys.length > 0 && (
              <Popconfirm
                title={t('session_manager.delete_confirm', { count: selectedRowKeys.length })}
                onConfirm={handleBatchDelete}
                okText={t('common.delete')}
                cancelText={t('common.cancel')}
                okButtonProps={{ danger: true }}
              >
                <Button danger>
                  {t('session_manager.batch_delete')} ({selectedRowKeys.length})
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
            showTotal: (total) => t('session_manager.total_count', { total }),
          }}
          rowSelection={{
            selectedRowKeys,
            onChange: setSelectedRowKeys,
          }}
          locale={{
            emptyText: <Empty description={t('session_manager.no_data')} />,
          }}
        />
      </Card>

      {/* Detail Modal */}
      <Modal
        title={t('session_manager.detail_title')}
        open={detailModalOpen}
        onCancel={closeDetailModal}
        footer={[
          <Button key="close" onClick={closeDetailModal}>
            {t('session_manager.close')}
          </Button>,
          selectedSession?.videoPath && (
            <Button
              key="play"
              type="primary"
              icon={<PlayCircleOutlined />}
              onClick={() => {
                closeDetailModal();
                handlePlayVideo(selectedSession!);
              }}
            >
              {t('session_manager.play_video')}
            </Button>
          ),
        ].filter(Boolean)}
        width={700}
      >
        {selectedSession && (
          <Descriptions bordered column={2} size="small">
            <Descriptions.Item label={t('session_manager.session_id')} span={2}>
              <Text copyable={{ text: selectedSession.id }}>
                {selectedSession.id}
              </Text>
            </Descriptions.Item>
            <Descriptions.Item label={t('session_manager.col_name')}>{selectedSession.name || '-'}</Descriptions.Item>
            <Descriptions.Item label={t('session_manager.col_status')}>
              <Tag color={getStatusColor(selectedSession.status)}>
                {getStatusText(selectedSession.status, t)}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label={t('session_manager.start_time')}>
              {formatTime(selectedSession.startTime)}
            </Descriptions.Item>
            <Descriptions.Item label={t('session_manager.end_time')}>
              {selectedSession.endTime > 0 ? formatTime(selectedSession.endTime) : t('session_manager.status_active')}
            </Descriptions.Item>
            <Descriptions.Item label={t('session_manager.col_duration')}>
              {formatDuration(
                selectedSession.endTime > 0
                  ? selectedSession.endTime - selectedSession.startTime
                  : Date.now() - selectedSession.startTime
              )}
            </Descriptions.Item>
            <Descriptions.Item label={t('session_manager.event_count')}>
              {selectedSession.eventCount || 0}
            </Descriptions.Item>
            <Descriptions.Item label={t('session_manager.col_device')} span={2}>
              <Text copyable={{ text: selectedSession.deviceId }}>
                {selectedSession.deviceId}
              </Text>
            </Descriptions.Item>
            <Descriptions.Item label={t('session_manager.video_path')} span={2}>
              {selectedSession.videoPath ? (
                <Text copyable={{ text: selectedSession.videoPath }}>
                  {selectedSession.videoPath}
                </Text>
              ) : (
                <Text type="secondary">{t('session_manager.no_video_info')}</Text>
              )}
            </Descriptions.Item>
            {selectedSession.videoDuration && selectedSession.videoDuration > 0 && (
              <Descriptions.Item label={t('session_manager.video_duration')}>
                {formatDuration(selectedSession.videoDuration)}
              </Descriptions.Item>
            )}
            <Descriptions.Item label={t('session_manager.config_info')} span={2}>
              <Space>
                {selectedSession.config?.logcat?.enabled && <Tag>{t('session.logcat')}</Tag>}
                {selectedSession.config?.recording?.enabled && <Tag color="blue">{t('session.screen_recording')}</Tag>}
                {selectedSession.config?.proxy?.enabled && <Tag color="green">{t('session.network_proxy')}</Tag>}
                {selectedSession.config?.monitor?.enabled && <Tag color="orange">{t('session.device_monitor')}</Tag>}
              </Space>
            </Descriptions.Item>
          </Descriptions>
        )}
      </Modal>

      {/* Rename Modal */}
      <Modal
        title={t('session_manager.rename_title')}
        open={renameModalOpen}
        onCancel={closeRenameModal}
        onOk={handleRename}
        okText={t('common.ok')}
        cancelText={t('common.cancel')}
        width={400}
      >
        <Input
          placeholder={t('session_manager.name_placeholder')}
          value={newName}
          onChange={(e) => setNewName(e.target.value)}
          onPressEnter={handleRename}
          autoFocus
        />
      </Modal>

    </div>
  );
};

export default SessionManager;
