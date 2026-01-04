import { useEffect, useRef, memo, useMemo, useCallback } from 'react';
import {
  Card,
  Tag,
  Select,
  Button,
  Empty,
  Space,
  Typography,
  Tooltip,
  Badge,
  Segmented,
  theme,
  Input,
} from 'antd';

import { useVirtualizer } from '@tanstack/react-virtual';
import {
  ReloadOutlined,
  ClearOutlined,
  FilterOutlined,
  HistoryOutlined,
  PlayCircleOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import {
  useSessionStore,
  useDeviceStore,
  useTimelineStore,
  formatEventTime,
  formatDuration,
  getEventIcon,
  getEventColor,
  categoryColors,
  levelStyles,
  type SessionEvent,
  type Session,
} from '../stores';

const { Text, Title } = Typography;

const TimelineView = () => {

  // Timeline View Component
  const { t } = useTranslation();
  const { token } = theme.useToken();

  // Device store
  const { devices, selectedDevice, setSelectedDevice } = useDeviceStore();

  // Session store
  const {
    sessions,
    activeSessionId,
    timeline,
    filter,
    isTimelineOpen,
    selectedEventId,
    loadTimeline,
    clearTimeline,
    setFilter,
    getFilteredTimeline,
    getSessions,
    subscribeToEvents,
    selectEvent,
  } = useSessionStore();

  // Timeline store
  const {
    sessionList,
    selectedSessionId,
    autoScroll,
    setSessionList,
    setSelectedSessionId,
    setAutoScroll,
    toggleAutoScroll,
  } = useTimelineStore();

  const listRef = useRef<HTMLDivElement>(null);

  // Category filter options
  const categoryOptions = [
    { label: 'Workflow', value: 'workflow' },
    { label: 'Log', value: 'log' },
    { label: 'Network', value: 'network' },
    { label: 'System', value: 'system' },
  ];

  // Level filter options
  const levelOptions = [
    { label: 'Error', value: 'error' },
    { label: 'Warn', value: 'warn' },
    { label: 'Info', value: 'info' },
    { label: 'Debug', value: 'debug' },
  ];

  // Segmented options - memoized to prevent re-creation
  const segmentedOptions = useMemo(() => [
    { label: t('timeline.auto_scroll') || 'Auto Scroll', value: 'on' },
    { label: t('timeline.manual') || 'Manual', value: 'off' },
  ], [t]);

  // Device options - memoized
  const deviceOptions = useMemo(() =>
    devices.map((d) => ({
      label: d.model || d.id,
      value: d.id,
    })),
    [devices]
  );

  // Session options - memoized
  const sessionOptions = useMemo(() =>
    sessionList.map((s) => ({
      label: (
        <Space>
          <PlayCircleOutlined />
          <span>{s.name}</span>
          <Text type="secondary" style={{ fontSize: 11 }}>
            {new Date(s.startTime).toLocaleString()}
          </Text>
        </Space>
      ),
      value: s.id,
    })),
    [sessionList]
  );

  // Subscribe to real-time events
  useEffect(() => {

    const unsubscribe = subscribeToEvents();
    return () => unsubscribe();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // subscribeToEvents is stable from Zustand

  // Load sessions when device changes
  const isLoadingSessionsRef = useRef(false);

  useEffect(() => {
    // Prevent concurrent loads
    if (isLoadingSessionsRef.current) return;

    const loadSessions = async () => {
      if (selectedDevice) {
        isLoadingSessionsRef.current = true;
        try {
          const sessions = await getSessions(selectedDevice, 50);
          setSessionList(sessions);
          // Auto-select first session if none selected
          if (sessions.length > 0 && !selectedSessionId) {
            setSelectedSessionId(sessions[0].id);
          }
        } finally {
          isLoadingSessionsRef.current = false;
        }
      } else {
        setSessionList([]);
        setSelectedSessionId(null);
      }
    };
    loadSessions();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedDevice]); // Only depend on selectedDevice

  // Load timeline when session changes
  // Memoize filter key to prevent unnecessary re-runs
  const filterKey = useMemo(() => JSON.stringify(filter), [
    filter.categories,
    filter.types,
    filter.levels,
    filter.stepId,
    filter.startTime,
    filter.endTime,
    filter.searchText,
  ]);

  useEffect(() => {
    if (selectedSessionId) {
      loadTimeline(selectedSessionId, filter);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedSessionId, filterKey]);

  // Get filtered timeline - memoized to prevent unnecessary re-work
  const filteredTimeline = useMemo(() => {
    return getFilteredTimeline();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [timeline, filterKey]);

  // Virtualizer callbacks - memoized to prevent recreation
  const getScrollElement = useCallback(() => listRef.current, []);
  const estimateSize = useCallback(() => 60, []);
  // Note: getItemKey uses filteredTimeline but we don't include it in deps
  // because the function is called with the current timeline state anyway
  const getItemKey = useCallback(
    (index: number) => {
      const item = filteredTimeline[index];
      return item?.id || index;
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    []
  );

  // Virtualizer
  const rowVirtualizer = useVirtualizer({
    count: filteredTimeline.length,
    getScrollElement,
    estimateSize,
    overscan: 5,
    getItemKey,
  });

  // Auto-scroll to bottom
  useEffect(() => {
    if (autoScroll && listRef.current && filteredTimeline.length > 0) {
      // Defer scroll to avoid flushSync warning
      queueMicrotask(() => {
        rowVirtualizer.scrollToIndex(filteredTimeline.length - 1, { align: 'end' });
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [timeline.length, autoScroll]); // Remove filteredTimeline.length and rowVirtualizer to prevent loops

  // Get session status badge
  const getSessionStatusBadge = useCallback((session: Session) => {
    switch (session.status) {
      case 'active':
        return <Badge status="processing" text={t('timeline.active') || 'Active'} />;
      case 'completed':
        return <Badge status="success" text={t('timeline.completed') || 'Completed'} />;
      case 'failed':
        return <Badge status="error" text={t('timeline.failed') || 'Failed'} />;
      case 'cancelled':
        return <Badge status="warning" text={t('timeline.cancelled') || 'Cancelled'} />;
      default:
        return <Badge status="default" text={session.status} />;
    }
  }, [t]);

  // Memoized callbacks to prevent re-renders
  const handleSearchChange = useCallback((value: string) => {
    setFilter({ searchText: value });
  }, [setFilter]);

  const handleCategoryChange = useCallback((values: string[]) => {
    setFilter({ categories: values });
  }, [setFilter]);

  const handleLevelChange = useCallback((values: string[]) => {
    setFilter({ levels: values });
  }, [setFilter]);

  const handleAutoScrollChange = useCallback((value: string | number) => {
    setAutoScroll(value === 'on');
  }, [setAutoScroll]);

  const handleRefresh = useCallback(() => {
    if (selectedSessionId) {
      // Read filter directly from store to avoid dependency
      const currentFilter = useSessionStore.getState().filter;
      loadTimeline(selectedSessionId, currentFilter);
    }
  }, [selectedSessionId, loadTimeline]);







  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', padding: 16 }}>
      {/* Header */}
      <div style={{ marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0, marginBottom: 12, color: token.colorText }}>
          <HistoryOutlined style={{ marginRight: 8 }} />
          {t('timeline.title') || 'Session Timeline'}
        </Title>

        {/* Device & Session Selection */}
        <Space wrap style={{ marginBottom: 12 }}>
          <Select
            style={{ width: 200 }}
            placeholder={t('timeline.select_device') || 'Select Device'}
            value={selectedDevice || undefined}
            onChange={setSelectedDevice}
            options={deviceOptions}
          />
          <Select
            style={{ width: 300 }}
            placeholder={t('timeline.select_session') || 'Select Session'}
            value={selectedSessionId || undefined}
            onChange={setSelectedSessionId}
            options={sessionOptions}
          />
        </Space>

        {/* Filters */}
        <div style={{ display: 'flex', alignItems: 'center', gap: 12, flexWrap: 'wrap' }}>
          <Input.Search
            placeholder={t('timeline.search_placeholder') || "Search..."}
            allowClear
            style={{ width: 200 }}
            onSearch={handleSearchChange}
          />
          <FilterOutlined style={{ color: token.colorTextSecondary }} />
          <Select
            mode="multiple"
            style={{ minWidth: 200 }}
            placeholder={t('timeline.filter_category') || 'Category'}
            value={filter.categories || []}
            onChange={handleCategoryChange}
            options={categoryOptions}
            allowClear
          />
          <Select
            mode="multiple"
            style={{ minWidth: 150 }}
            placeholder={t('timeline.filter_level') || 'Level'}
            value={filter.levels || []}
            onChange={handleLevelChange}
            options={levelOptions}
            allowClear
          />
          <Segmented
            size="small"
            options={segmentedOptions}
            value={autoScroll ? 'on' : 'off'}
            onChange={handleAutoScrollChange}
          />
          <Tooltip title={t('timeline.refresh') || 'Refresh'}>
            <Button
              icon={<ReloadOutlined />}
              size="small"
              onClick={handleRefresh}
            />
          </Tooltip>
          <Tooltip title={t('timeline.clear') || 'Clear'}>
            <Button
              icon={<ClearOutlined />}
              size="small"
              onClick={clearTimeline}
            />
          </Tooltip>
        </div>
      </div>

      {/* Session Info Card */}
      {selectedSessionId && sessionList.find((s) => s.id === selectedSessionId) && (
        <Card size="small" style={{ marginBottom: 12 }}>
          {(() => {
            const session = sessionList.find((s) => s.id === selectedSessionId)!;
            return (
              <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                {getSessionStatusBadge(session)}
                <Text strong style={{ color: token.colorText }}>{session.name}</Text>
                <Tag>{session.type}</Tag>
                <Text type="secondary">
                  {t('timeline.events') || 'Events'}: {session.eventCount}
                </Text>
                {session.endTime > 0 && (
                  <Text type="secondary">
                    {t('timeline.duration') || 'Duration'}:{' '}
                    {formatDuration(session.endTime - session.startTime)}
                  </Text>
                )}
              </div>
            );
          })()}
        </Card>
      )}



      {/* Timeline List */}
      <div
        ref={listRef}
        style={{
          flex: 1,
          overflow: 'auto',
          backgroundColor: token.colorBgContainer,
          borderRadius: 8,
          border: `1px solid ${token.colorBorder}`,
        }}
      >
        {filteredTimeline.length === 0 ? (
          <Empty
            description={t('timeline.no_events') || 'No events'}
            style={{ marginTop: 48 }}
          />
        ) : (
          <div
            style={{
              height: `${rowVirtualizer.getTotalSize()}px`,
              width: '100%',
              position: 'relative',
            }}
          >
            {rowVirtualizer.getVirtualItems().map((virtualRow) => {
              const event = filteredTimeline[virtualRow.index];
              if (!event) return null; // Added check for robustness
              return (
                <EventRow
                  key={virtualRow.key}
                  index={virtualRow.index}
                  event={event}
                  isSelected={event.id === selectedEventId}
                  onSelect={selectEvent}
                  measureElement={rowVirtualizer.measureElement}
                  style={{
                    position: 'absolute',
                    top: 0,
                    left: 0,
                    width: '100%',
                    transform: `translateY(${virtualRow.start}px)`,
                  }}
                />
              );
            })}
          </div>
        )}
      </div>

      {/* Stats Footer */}
      <div
        style={{
          marginTop: 8,
          display: 'flex',
          justifyContent: 'space-between',
          color: token.colorTextSecondary,
          fontSize: 12,
        }}
      >
        <span>
          {t('timeline.showing') || 'Showing'} {filteredTimeline.length}{' '}
          {t('timeline.of') || 'of'} {timeline.length} {t('timeline.events') || 'events'}
        </span>
        {activeSessionId && (
          <Badge status="processing" text={t('timeline.live') || 'Live'} />
        )}
      </div>
    </div >
  );
};

// Separated EventRow component to prevent re-creation on every render (causing flicker)
const EventRow = memo(({
  index,
  event,
  isSelected,
  onSelect,
  style,
  measureElement,
}: {
  index: number;
  event: SessionEvent;
  isSelected: boolean;
  onSelect: (id: string) => void;
  style: React.CSSProperties;
  measureElement: (el: HTMLElement | null) => void;
}) => {
  const { token } = theme.useToken();
  const icon = getEventIcon(event);
  const color = getEventColor(event);

  return (
    <div
      ref={measureElement}
      data-index={index}
      onClick={() => onSelect(event.id)}
      style={{
        ...style,
        cursor: 'pointer',
        backgroundColor: isSelected ? token.colorPrimaryBg : 'transparent',
        borderLeft: `3px solid ${color}`,
        padding: '8px 12px',
        borderBottom: `1px solid ${token.colorSplit}`,
        display: 'flex',
        flexDirection: 'column',
        justifyContent: 'center',
        height: 'auto',
        minHeight: style.height,
      }}
    >
      <div style={{ width: '100%' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
          <span style={{ fontSize: 16 }}>{icon}</span>
          <Text strong style={{ flex: 1, color: token.colorText }}>
            {event.title}
          </Text>
          <Text type="secondary" style={{ fontSize: 12 }}>
            {formatEventTime(event.timestamp)}
          </Text>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <Tag
            color={categoryColors[event.category]}
            style={{ margin: 0, fontSize: 11 }}
          >
            {event.category}
          </Tag>
          <Tag
            color={levelStyles[event.level]?.color}
            style={{ margin: 0, fontSize: 11 }}
          >
            {event.level}
          </Tag>
          {event.duration !== undefined && event.duration > 0 && (
            <Text type="secondary" style={{ fontSize: 11 }}>
              {formatDuration(event.duration)}
            </Text>
          )}
          {event.stepId && (
            <Text type="secondary" style={{ fontSize: 11 }}>
              Step: {event.stepId.slice(0, 8)}
            </Text>
          )}
        </div>
        {event.detail && typeof event.detail === 'object' && (
          <div style={{ marginTop: 4 }}>
            {(event.detail as any).error && (
              <Text type="danger" style={{ fontSize: 12 }}>
                {(event.detail as any).error}
              </Text>
            )}
            {(event.detail as any).url && (
              <Text
                type="secondary"
                style={{ fontSize: 11, display: 'block' }}
                ellipsis
              >
                {(event.detail as any).url}
              </Text>
            )}
            {(event.detail as any).result !== undefined && (
              <Text type="secondary" style={{ fontSize: 11 }}>
                Result: {String((event.detail as any).result)}
              </Text>
            )}
          </div>
        )}
      </div>
    </div>
  );
});

export default TimelineView;
