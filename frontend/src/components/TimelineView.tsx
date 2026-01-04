import { useEffect, useRef, memo, useMemo, useCallback } from 'react';
import dayjs from 'dayjs';
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
  DatePicker,
  Tabs,
  Descriptions,
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

  // RangePicker presets generator function - returns dayjs objects for compatibility
  const getRangePresets = () => {
    return [
      { label: '1m', value: [dayjs().subtract(1, 'minute'), dayjs()] as [dayjs.Dayjs, dayjs.Dayjs] },
      { label: '5m', value: [dayjs().subtract(5, 'minute'), dayjs()] as [dayjs.Dayjs, dayjs.Dayjs] },
      { label: '15m', value: [dayjs().subtract(15, 'minute'), dayjs()] as [dayjs.Dayjs, dayjs.Dayjs] },
      { label: '1h', value: [dayjs().subtract(1, 'hour'), dayjs()] as [dayjs.Dayjs, dayjs.Dayjs] },
    ];
  };

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
  const estimateSize = useCallback(() => 32, []); // Compact row height
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

  const handleTimeRangeChange = useCallback((dates: any) => {
    if (!dates || dates.length < 2) {
      // Clear time filter
      setFilter({ startTime: undefined, endTime: undefined });
    } else {
      // Set custom time range
      setFilter({
        startTime: dates[0]?.valueOf() || undefined,
        endTime: dates[1]?.valueOf() || undefined
      });
    }
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

        {/* Device Selection */}
        <Space wrap style={{ marginBottom: 12 }}>
          <Select
            style={{ width: 200 }}
            placeholder={t('timeline.select_device') || 'Select Device'}
            value={selectedDevice || undefined}
            onChange={setSelectedDevice}
            options={deviceOptions}
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
          <DatePicker.RangePicker
            size="small"
            showTime={{ format: 'HH:mm' }}
            format="MM-DD HH:mm"
            presets={getRangePresets()}
            onChange={handleTimeRangeChange}
            allowClear
            style={{ width: 260 }}
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

      {/* Split Layout: Timeline List + Detail Panel */}
      <div style={{ flex: 1, display: 'flex', gap: 12, overflow: 'hidden' }}>
        {/* Timeline List */}
        <div
          ref={listRef}
          style={{
            flex: selectedEventId ? 2 : 1,
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
                if (!event) return null;
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

        {/* Detail Panel */}
        {selectedEventId && (() => {
          const selectedEvent = filteredTimeline.find(e => e.id === selectedEventId);
          if (!selectedEvent) return null;

          return (
            <Card
              size="small"
              title={<Text strong>Event Details</Text>}
              style={{ flex: 1, overflow: 'auto', minWidth: 300 }}
              styles={{ body: { padding: 12 } }}
              extra={
                <Button type="text" size="small" onClick={() => selectEvent('')}>
                  ✕
                </Button>
              }
            >
              <Space direction="vertical" size="small" style={{ width: '100%' }}>
                {/* Title */}
                <div>
                  <Text type="secondary" style={{ fontSize: 11 }}>Title</Text>
                  <div style={{ wordBreak: 'break-all' }}>{selectedEvent.title}</div>
                </div>

                {/* Meta */}
                <div style={{ display: 'flex', gap: 8 }}>
                  <Tag color={selectedEvent.level === 'error' ? 'red' : selectedEvent.level === 'warn' ? 'orange' : 'blue'}>
                    {selectedEvent.level}
                  </Tag>
                  <Tag color="purple">{selectedEvent.category}</Tag>
                  <Tag>{selectedEvent.type}</Tag>
                </div>

                {/* Time */}
                <div>
                  <Text type="secondary" style={{ fontSize: 11 }}>Time</Text>
                  <div style={{ fontFamily: 'monospace', fontSize: 12 }}>
                    {new Date(selectedEvent.timestamp).toLocaleString()}
                  </div>
                </div>

                {/* Duration */}
                {selectedEvent.duration !== undefined && selectedEvent.duration > 0 && (
                  <div>
                    <Text type="secondary" style={{ fontSize: 11 }}>Duration</Text>
                    <div>{formatDuration(selectedEvent.duration)}</div>
                  </div>
                )}

                {/* Detail */}
                {/* Detail Content */}
                {(() => {
                  if (!selectedEvent.detail) return null;

                  // Network Event - Tabbed View
                  if (selectedEvent.category === 'network') {
                    const d = selectedEvent.detail as any;
                    const items = [
                      {
                        key: 'summary',
                        label: 'Summary',
                        children: (
                          <Descriptions column={1} size="small" bordered>
                            <Descriptions.Item label="URL"><Text copyable={{ text: d.url }} style={{ fontSize: 11 }}>{d.url}</Text></Descriptions.Item>
                            <Descriptions.Item label="Method">{d.method}</Descriptions.Item>
                            <Descriptions.Item label="Status Code">
                              <Tag color={d.statusCode >= 400 ? 'red' : 'green'}>{d.statusCode}</Tag>
                            </Descriptions.Item>
                            <Descriptions.Item label="Duration">{d.duration ? `${d.duration} ms` : '-'}</Descriptions.Item>
                            <Descriptions.Item label="Type">{d.contentType || '-'}</Descriptions.Item>
                            <Descriptions.Item label="Size">{d.bodySize ? `${d.bodySize} bytes` : '-'}</Descriptions.Item>
                          </Descriptions>
                        )
                      },
                      {
                        key: 'request',
                        label: 'Request',
                        children: (
                          <div>
                            {d.requestHeaders && (
                              <div style={{ marginBottom: 12 }}>
                                <Text strong style={{ fontSize: 12, display: 'block', marginBottom: 4 }}>Headers</Text>
                                <div style={{ background: token.colorFillAlter, padding: 8, borderRadius: 4, fontFamily: 'monospace', fontSize: 11 }}>
                                  {Object.entries(d.requestHeaders).map(([k, v]) => (
                                    <div key={k} style={{ wordBreak: 'break-all' }}>
                                      <span style={{ color: token.colorTextSecondary }}>{k}:</span> {Array.isArray(v) ? v.join(', ') : String(v)}
                                    </div>
                                  ))}
                                </div>
                              </div>
                            )}
                            <div>
                              <Text strong style={{ fontSize: 12, display: 'block', marginBottom: 4 }}>Body</Text>
                              <pre style={{ margin: 0, fontSize: 11, background: token.colorFillAlter, padding: 8, borderRadius: 4, whiteSpace: 'pre-wrap', wordBreak: 'break-all', overflowX: 'hidden' }}>{d.requestBody || '<Empty>'}</pre>
                            </div>
                          </div>
                        )
                      },
                      {
                        key: 'response',
                        label: 'Response',
                        children: (
                          <div>
                            {d.responseHeaders && (
                              <div style={{ marginBottom: 12 }}>
                                <Text strong style={{ fontSize: 12, display: 'block', marginBottom: 4 }}>Headers</Text>
                                <div style={{ background: token.colorFillAlter, padding: 8, borderRadius: 4, fontFamily: 'monospace', fontSize: 11 }}>
                                  {Object.entries(d.responseHeaders).map(([k, v]) => (
                                    <div key={k} style={{ wordBreak: 'break-all' }}>
                                      <span style={{ color: token.colorTextSecondary }}>{k}:</span> {Array.isArray(v) ? v.join(', ') : String(v)}
                                    </div>
                                  ))}
                                </div>
                              </div>
                            )}
                            <div>
                              <Text strong style={{ fontSize: 12, display: 'block', marginBottom: 4 }}>Body</Text>
                              <pre style={{ margin: 0, fontSize: 11, background: token.colorFillAlter, padding: 8, borderRadius: 4, whiteSpace: 'pre-wrap', wordBreak: 'break-all', overflowX: 'hidden' }}>{d.responseBody || '<Empty>'}</pre>
                            </div>
                          </div>
                        )
                      }
                    ];
                    return <Tabs items={items} size="small" style={{ marginTop: 0 }} />;
                  }

                  // Default - JSON View
                  return (
                    <div>
                      <Text type="secondary" style={{ fontSize: 11 }}>Detail</Text>
                      <pre style={{
                        fontSize: 11,
                        background: token.colorFillAlter,
                        padding: 8,
                        borderRadius: 4,
                        overflowX: 'hidden',
                        margin: '4px 0 0',
                        whiteSpace: 'pre-wrap',
                        wordBreak: 'break-all',
                      }}>
                        {JSON.stringify(selectedEvent.detail, null, 2)}
                      </pre>
                    </div>
                  );
                })()}
              </Space>
            </Card>
          );
        })()}
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

// Compact single-line EventRow
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

  // Level colors
  const levelColorMap: Record<string, string> = {
    error: 'red',
    warn: 'orange',
    info: 'blue',
    debug: 'default',
    verbose: 'default',
  };

  // Category colors
  const categoryColorMap: Record<string, string> = {
    workflow: 'purple',
    log: 'cyan',
    network: 'green',
    system: 'default',
  };

  // Background color based on level for quick visual
  const rowBg = isSelected
    ? token.colorPrimaryBg
    : event.level === 'error'
      ? 'rgba(255, 77, 79, 0.08)'
      : event.level === 'warn'
        ? 'rgba(250, 173, 20, 0.06)'
        : 'transparent';

  return (
    <div
      ref={measureElement}
      data-index={index}
      onClick={() => onSelect(event.id)}
      style={{
        ...style,
        cursor: 'pointer',
        backgroundColor: rowBg,
        borderLeft: `3px solid ${levelColorMap[event.level] === 'red' ? '#ff4d4f' : levelColorMap[event.level] === 'orange' ? '#faad14' : token.colorBorder}`,
        padding: '4px 8px',
        borderBottom: `1px solid ${token.colorSplit}`,
        display: 'flex',
        alignItems: 'center',
        gap: 8,
        height: 32,
        fontSize: 12,
      }}
    >
      {/* Time */}
      <Text type="secondary" style={{ fontSize: 11, fontFamily: 'monospace', width: 80, flexShrink: 0 }}>
        {formatEventTime(event.timestamp)}
      </Text>

      {/* Category Tag */}
      <Tag
        color={categoryColorMap[event.category] || 'default'}
        style={{ margin: 0, fontSize: 10, lineHeight: '16px', padding: '0 4px' }}
      >
        {event.category.toUpperCase().slice(0, 3)}
      </Tag>

      {/* Level Tag */}
      <Tag
        color={levelColorMap[event.level] || 'default'}
        style={{ margin: 0, fontSize: 10, lineHeight: '16px', padding: '0 4px' }}
      >
        {event.category === 'network' && (event.detail as any)?.method
          ? (event.detail as any).method.toUpperCase()
          : event.level.toUpperCase().slice(0, 3)}
      </Tag>

      {/* Title - truncated */}
      <Text
        ellipsis
        style={{
          flex: 1,
          color: event.level === 'error' ? '#ff4d4f' : event.level === 'warn' ? '#faad14' : token.colorText,
          fontSize: 12,
        }}
      >
        {event.category === 'network' && (event.detail as any)?.method && event.title.startsWith((event.detail as any).method + ' ')
          ? event.title.substring(((event.detail as any).method as string).length + 1)
          : event.title}
      </Text>

      {/* Duration if present */}
      {event.duration !== undefined && event.duration > 0 && (
        <Text type="secondary" style={{ fontSize: 10, flexShrink: 0 }}>
          {formatDuration(event.duration)}
        </Text>
      )}
    </div>
  );
});

export default TimelineView;
