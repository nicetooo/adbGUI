/**
 * EventTimeline - 新事件系统的时间线组件
 * 使用 eventStore 和 SQLite 持久化
 */
import { useEffect, useRef, useState, useMemo, useCallback, memo } from 'react';
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
  Input,
  Descriptions,
  Segmented,
  theme,
  Spin,
  Slider,
  Popover,
  Checkbox,
  Tabs,
  Modal,
  List,
} from 'antd';
import { useVirtualizer } from '@tanstack/react-virtual';
import {
  ReloadOutlined,
  FilterOutlined,
  HistoryOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  BookOutlined,
  DownloadOutlined,
  SearchOutlined,
  CloseOutlined,
  VerticalAlignBottomOutlined,
  ExpandOutlined,
  ThunderboltOutlined,
  LeftOutlined,
  RightOutlined,
  DeleteOutlined,
  PushpinOutlined,
} from '@ant-design/icons';
import AssertionsPanel from './AssertionsPanel';
import SmartSearchInput, { NLQueryResult } from './SmartSearchInput';
import SessionSummary from './SessionSummary';
import { useTranslation } from 'react-i18next';
import {
  useEventStore,
  useDeviceStore,
  useCurrentBookmarks,
  type UnifiedEvent,
  type DeviceSession,
  type EventSource,
  type EventCategory,
  type EventLevel,
  type SessionConfig,
  type Bookmark,
  sourceConfig,
  categoryConfig,
  levelConfig,
  formatRelativeTime,
  formatTimestamp,
  getEventIcon,
  getEventColor,
  isCriticalEvent,
} from '../stores';
import SessionConfigModal from './SessionConfigModal';

const { Text, Title } = Typography;

// ========================================
// 时间轴可视化组件
// ========================================

interface TimeIndexEntry {
  second: number;
  eventCount: number;
  firstEventId: string;
  hasError: boolean;
}

interface TimeRulerProps {
  sessionStart: number;
  sessionDuration: number;
  timeIndex: TimeIndexEntry[];
  currentTime: number;
  onSeek: (time: number) => void;
  timeRange: { start: number; end: number } | null;
  onTimeRangeChange: (range: { start: number; end: number } | null) => void;
  bookmarks?: Bookmark[];
  onBookmarkClick?: (bookmark: Bookmark) => void;
}

const TimeRuler = memo(({ sessionStart, sessionDuration, timeIndex, currentTime, onSeek, timeRange, onTimeRangeChange, bookmarks = [], onBookmarkClick }: TimeRulerProps) => {
  const { token } = theme.useToken();
  const containerRef = useRef<HTMLDivElement>(null);
  const [isDragging, setIsDragging] = useState(false);
  const [dragStart, setDragStart] = useState<number | null>(null);
  const [dragEnd, setDragEnd] = useState<number | null>(null);

  const totalSeconds = Math.ceil(sessionDuration / 1000);

  // Convert array to map for quick lookup
  const indexMap = useMemo(() => {
    const map = new Map<number, TimeIndexEntry>();
    timeIndex.forEach(entry => map.set(entry.second, entry));
    return map;
  }, [timeIndex]);

  const maxCount = useMemo(() => {
    let max = 1;
    timeIndex.forEach(entry => {
      if (entry.eventCount > max) max = entry.eventCount;
    });
    return max;
  }, [timeIndex]);

  // Convert x position to time
  const posToTime = useCallback((clientX: number) => {
    if (!containerRef.current) return 0;
    const rect = containerRef.current.getBoundingClientRect();
    const x = Math.max(0, Math.min(clientX - rect.left, rect.width));
    const ratio = x / rect.width;
    return Math.round(ratio * sessionDuration);
  }, [sessionDuration]);

  // Handle mouse events for drag selection
  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    if (e.button !== 0) return; // Only left click
    e.preventDefault();
    const time = posToTime(e.clientX);
    setIsDragging(true);
    setDragStart(time);
    setDragEnd(time);
  }, [posToTime]);

  const handleMouseMove = useCallback((e: MouseEvent) => {
    if (!isDragging) return;
    const time = posToTime(e.clientX);
    setDragEnd(time);
  }, [isDragging, posToTime]);

  const handleMouseUp = useCallback((e: MouseEvent) => {
    if (!isDragging || dragStart === null || dragEnd === null) {
      setIsDragging(false);
      return;
    }

    const start = Math.min(dragStart, dragEnd);
    const end = Math.max(dragStart, dragEnd);

    // If range is too small (< 100ms), treat as click to seek
    if (end - start < 100) {
      onTimeRangeChange(null);
      onSeek(start);
    } else {
      onTimeRangeChange({ start, end });
    }

    setIsDragging(false);
    setDragStart(null);
    setDragEnd(null);
  }, [isDragging, dragStart, dragEnd, onSeek, onTimeRangeChange]);

  // Add/remove global mouse event listeners
  useEffect(() => {
    if (isDragging) {
      window.addEventListener('mousemove', handleMouseMove);
      window.addEventListener('mouseup', handleMouseUp);
      return () => {
        window.removeEventListener('mousemove', handleMouseMove);
        window.removeEventListener('mouseup', handleMouseUp);
      };
    }
  }, [isDragging, handleMouseMove, handleMouseUp]);

  // Calculate selection display (either from drag or from timeRange prop)
  const selectionStart = isDragging && dragStart !== null && dragEnd !== null
    ? Math.min(dragStart, dragEnd)
    : timeRange?.start ?? null;
  const selectionEnd = isDragging && dragStart !== null && dragEnd !== null
    ? Math.max(dragStart, dragEnd)
    : timeRange?.end ?? null;

  const hasSelection = selectionStart !== null && selectionEnd !== null && selectionEnd > selectionStart;

  return (
    <div style={{ position: 'relative' }}>
      <div
        ref={containerRef}
        onMouseDown={handleMouseDown}
        style={{
          height: 50,
          background: token.colorBgContainer,
          borderRadius: 4,
          position: 'relative',
          cursor: 'crosshair',
          border: `1px solid ${token.colorBorder}`,
          overflow: 'hidden',
          userSelect: 'none',
        }}
      >
        {/* Event density bars - 使用绝对定位让每个条形正确对应时间位置 */}
        <div style={{ position: 'relative', height: 36 }}>
          {Array.from({ length: Math.max(totalSeconds, 1) }).map((_, i) => {
            const entry = indexMap.get(i);
            const height = entry ? Math.max(4, (entry.eventCount / maxCount) * 30) : 0;
            const hasError = entry?.hasError;
            const secStart = i * 1000;
            const secEnd = (i + 1) * 1000;
            const isInSelection = hasSelection && secStart < selectionEnd! && secEnd > selectionStart!;

            // 使用百分比定位，确保条形位置与时间轴对齐
            const leftPercent = (i / totalSeconds) * 100;
            const widthPercent = (1 / totalSeconds) * 100;

            return (
              <div
                key={i}
                style={{
                  position: 'absolute',
                  left: `${leftPercent}%`,
                  width: `calc(${widthPercent}% - 1px)`,
                  height,
                  bottom: 0,
                  background: hasError ? token.colorError : token.colorPrimary,
                  opacity: isInSelection ? 1 : (hasError ? 0.9 : 0.7),
                  borderRadius: 1,
                }}
                title={entry ? `${i}s: ${entry.eventCount} events` : `${i}s: 0 events`}
              />
            );
          })}
        </div>

        {/* Selection overlay */}
        {hasSelection && sessionDuration > 0 && (
          <div
            style={{
              position: 'absolute',
              left: `${(selectionStart! / sessionDuration) * 100}%`,
              width: `${((selectionEnd! - selectionStart!) / sessionDuration) * 100}%`,
              top: 0,
              bottom: 0,
              background: token.colorPrimaryBg,
              borderLeft: `2px solid ${token.colorPrimary}`,
              borderRight: `2px solid ${token.colorPrimary}`,
              pointerEvents: 'none',
            }}
          />
        )}

        {/* Current position indicator (yellow) */}
        {sessionDuration > 0 && (
          <div
            style={{
              position: 'absolute',
              left: `${(currentTime / sessionDuration) * 100}%`,
              top: 0,
              bottom: 0,
              width: 2,
              background: token.colorWarning,
              pointerEvents: 'none',
              zIndex: 10,
            }}
          />
        )}

        {/* Bookmark markers */}
        {bookmarks.map(bookmark => (
          <Tooltip
            key={bookmark.id}
            title={`${bookmark.label} (${formatRelativeTime(bookmark.relativeTime)})`}
          >
            <div
              onClick={(e) => {
                e.stopPropagation();
                onBookmarkClick?.(bookmark);
              }}
              style={{
                position: 'absolute',
                left: `${(bookmark.relativeTime / sessionDuration) * 100}%`,
                top: 0,
                width: 0,
                height: 0,
                borderLeft: '6px solid transparent',
                borderRight: '6px solid transparent',
                borderTop: `10px solid ${bookmark.color || token.colorSuccess}`,
                transform: 'translateX(-6px)',
                cursor: 'pointer',
                zIndex: 15,
              }}
            />
          </Tooltip>
        ))}

        {/* Time labels */}
        <div style={{
          position: 'absolute',
          bottom: 2,
          left: 4,
          fontSize: 10,
          color: token.colorTextSecondary,
          pointerEvents: 'none',
        }}>
          {formatRelativeTime(0)}
        </div>
        <div style={{
          position: 'absolute',
          bottom: 2,
          right: 4,
          fontSize: 10,
          color: token.colorTextSecondary,
          pointerEvents: 'none',
        }}>
          {formatRelativeTime(sessionDuration)}
        </div>

        {/* Selection time labels */}
        {hasSelection && (
          <>
            <div style={{
              position: 'absolute',
              top: 2,
              left: `${(selectionStart! / sessionDuration) * 100}%`,
              transform: 'translateX(-50%)',
              fontSize: 10,
              fontWeight: 500,
              color: token.colorPrimary,
              background: token.colorBgContainer,
              padding: '0 2px',
              borderRadius: 2,
              pointerEvents: 'none',
            }}>
              {formatRelativeTime(selectionStart!)}
            </div>
            <div style={{
              position: 'absolute',
              top: 2,
              left: `${(selectionEnd! / sessionDuration) * 100}%`,
              transform: 'translateX(-50%)',
              fontSize: 10,
              fontWeight: 500,
              color: token.colorPrimary,
              background: token.colorBgContainer,
              padding: '0 2px',
              borderRadius: 2,
              pointerEvents: 'none',
            }}>
              {formatRelativeTime(selectionEnd!)}
            </div>
          </>
        )}
      </div>

      {/* Selection info and clear button */}
      {timeRange && (
        <div style={{
          display: 'flex',
          alignItems: 'center',
          gap: 8,
          marginTop: 4,
          fontSize: 12,
        }}>
          <Text type="secondary">
            {formatRelativeTime(timeRange.start)} - {formatRelativeTime(timeRange.end)}
            {' '}({((timeRange.end - timeRange.start) / 1000).toFixed(1)}s)
          </Text>
          <Button
            type="link"
            size="small"
            onClick={() => onTimeRangeChange(null)}
            style={{ padding: 0, height: 'auto' }}
          >
            <CloseOutlined />
          </Button>
        </div>
      )}

    </div>
  );
});

TimeRuler.displayName = 'TimeRuler';

// ========================================
// 事件行组件
// ========================================

interface EventRowProps {
  event: UnifiedEvent;
  isSelected: boolean;
  onClick: () => void;
  style?: React.CSSProperties;
}

const EventRow = memo(({ event, isSelected, onClick, style }: EventRowProps) => {
  const { token } = theme.useToken();
  const color = getEventColor(event);
  const icon = getEventIcon(event);
  const isCritical = isCriticalEvent(event);
  const isWarning = event.level === 'warn';

  // Determine background color based on state and level
  const getBackground = () => {
    if (isSelected) return token.colorPrimaryBg;
    if (isCritical) return token.colorErrorBg;
    if (isWarning) return token.colorWarningBg;
    return 'transparent';
  };

  return (
    <div
      onClick={onClick}
      style={{
        ...style,
        display: 'flex',
        alignItems: 'center',
        padding: '4px 8px',
        cursor: 'pointer',
        background: getBackground(),
        borderLeft: `3px solid ${color}`,
        borderBottom: `1px solid ${token.colorBorderSecondary}`,
        transition: 'background 0.15s',
      }}
    >
      {/* Time */}
      <Text
        type="secondary"
        style={{
          width: 80,
          flexShrink: 0,
          fontSize: 11,
          fontFamily: 'monospace'
        }}
      >
        {formatRelativeTime(event.relativeTime)}
      </Text>

      {/* Icon */}
      <span style={{ width: 24, textAlign: 'center', fontSize: 14 }}>
        {icon}
      </span>

      {/* Source tag */}
      <Tag
        color={sourceConfig[event.source]?.color || '#888'}
        style={{
          margin: '0 8px',
          fontSize: 10,
          padding: '0 4px',
          lineHeight: '18px',
        }}
      >
        {event.source}
      </Tag>

      {/* Title */}
      <Text
        ellipsis
        style={{
          flex: 1,
          fontSize: 12,
          color: isCritical ? token.colorError : token.colorText,
        }}
      >
        {event.title}
      </Text>

      {/* Duration if present */}
      {event.duration && event.duration > 0 && (
        <Text type="secondary" style={{ fontSize: 10, marginLeft: 8 }}>
          {event.duration}ms
        </Text>
      )}

      {/* Level indicator */}
      {(event.level === 'error' || event.level === 'warn') && (
        <Badge
          color={levelConfig[event.level].color}
          style={{ marginLeft: 8 }}
        />
      )}
    </div>
  );
});

EventRow.displayName = 'EventRow';

// ========================================
// 事件详情抽屉
// ========================================

interface EventDetailProps {
  event: UnifiedEvent | null;
  onClose: () => void;
}

// Network event detail component
const NetworkEventDetail = memo(({ data, token }: { data: any; token: any }) => {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState('overview');

  const statusColor = data.statusCode >= 500 ? '#ff4d4f' :
    data.statusCode >= 400 ? '#faad14' :
    data.statusCode >= 300 ? '#1890ff' : '#52c41a';

  const formatSize = (bytes: number) => {
    if (!bytes) return '-';
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(2)} MB`;
  };

  const formatHeaders = (headers: Record<string, string> | null) => {
    if (!headers || Object.keys(headers).length === 0) {
      return <Text type="secondary">{t('timeline.no_headers')}</Text>;
    }
    return (
      <div style={{ fontSize: 12 }}>
        {Object.entries(headers).map(([key, value]) => (
          <div key={key} style={{ marginBottom: 4 }}>
            <Text strong style={{ color: token.colorPrimary }}>{key}: </Text>
            <Text copyable={{ text: value }} style={{ wordBreak: 'break-all' }}>{value}</Text>
          </div>
        ))}
      </div>
    );
  };

  const formatBody = (body: string | null, contentType: string | null) => {
    if (!body) return <Text type="secondary">{t('timeline.no_body')}</Text>;

    // Try to format as JSON if content type suggests it
    if (contentType?.includes('json')) {
      try {
        const parsed = JSON.parse(body);
        return (
          <pre style={{
            background: token.colorBgLayout,
            padding: 12,
            borderRadius: 4,
            overflow: 'auto',
            maxHeight: 300,
            fontSize: 11,
            whiteSpace: 'pre-wrap',
            wordBreak: 'break-all',
            margin: 0,
          }}>
            {JSON.stringify(parsed, null, 2)}
          </pre>
        );
      } catch {
        // Fall through to plain text
      }
    }

    return (
      <pre style={{
        background: token.colorBgLayout,
        padding: 12,
        borderRadius: 4,
        overflow: 'auto',
        maxHeight: 300,
        fontSize: 11,
        whiteSpace: 'pre-wrap',
        wordBreak: 'break-all',
        margin: 0,
      }}>
        {body}
      </pre>
    );
  };

  const tabItems = [
    {
      key: 'overview',
      label: t('timeline.overview'),
      children: (
        <Descriptions column={1} size="small" bordered>
          <Descriptions.Item label={t('timeline.url')}>
            <Text copyable style={{ fontSize: 11, wordBreak: 'break-all' }}>
              {data.url}
            </Text>
          </Descriptions.Item>
          <Descriptions.Item label={t('timeline.method')}>
            <Tag color="blue">{data.method}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label={t('timeline.status')}>
            <Tag color={statusColor}>{data.statusCode}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label={t('timeline.duration')}>
            {data.duration ? `${data.duration}ms` : '-'}
          </Descriptions.Item>
          <Descriptions.Item label={t('timeline.size')}>
            {formatSize(data.bodySize)}
          </Descriptions.Item>
          <Descriptions.Item label={t('timeline.content_type')}>
            {data.contentType || '-'}
          </Descriptions.Item>
          <Descriptions.Item label={t('timeline.protocol')}>
            {data.isHttps ? 'HTTPS' : 'HTTP'}
            {data.isWs && ' (WebSocket)'}
          </Descriptions.Item>
        </Descriptions>
      ),
    },
    {
      key: 'reqHeaders',
      label: t('timeline.request_headers'),
      children: formatHeaders(data.requestHeaders),
    },
    {
      key: 'reqBody',
      label: t('timeline.request_body'),
      children: formatBody(data.requestBody, data.requestHeaders?.['Content-Type'] || data.requestHeaders?.['content-type']),
    },
    {
      key: 'resHeaders',
      label: t('timeline.response_headers'),
      children: formatHeaders(data.responseHeaders),
    },
    {
      key: 'resBody',
      label: t('timeline.response_body'),
      children: formatBody(data.responseBody, data.contentType),
    },
  ];

  return (
    <Tabs
      activeKey={activeTab}
      onChange={setActiveTab}
      size="small"
      items={tabItems}
      style={{ marginTop: -8 }}
    />
  );
});

NetworkEventDetail.displayName = 'NetworkEventDetail';

const EventDetail = memo(({ event, onClose }: EventDetailProps) => {
  const { token } = theme.useToken();

  if (!event) return null;

  const getData = () => {
    if (!event.data) return null;
    try {
      return typeof event.data === 'string' ? JSON.parse(event.data) : event.data;
    } catch {
      return null;
    }
  };

  const data = getData();
  const isNetworkEvent = event.type === 'network_request' || event.category === 'network';

  const renderData = () => {
    if (!event.data) return <Text type="secondary">No data</Text>;

    // Use specialized view for network events
    if (isNetworkEvent && data) {
      return <NetworkEventDetail data={data} token={token} />;
    }

    // Generic JSON view for other events
    try {
      return (
        <pre style={{
          background: token.colorBgLayout,
          padding: 12,
          borderRadius: 4,
          overflow: 'auto',
          maxHeight: 'calc(100vh - 400px)',
          fontSize: 11,
          whiteSpace: 'pre-wrap',
          wordBreak: 'break-all',
          margin: 0,
        }}>
          {JSON.stringify(data, null, 2)}
        </pre>
      );
    } catch {
      return (
        <pre style={{
          background: token.colorBgLayout,
          padding: 12,
          borderRadius: 4,
          overflow: 'auto',
          maxHeight: 'calc(100vh - 400px)',
          fontSize: 11,
          whiteSpace: 'pre-wrap',
          wordBreak: 'break-all',
          margin: 0,
        }}>
          {String(event.data)}
        </pre>
      );
    }
  };

  // For network events, show a simplified header
  const getTitle = () => {
    if (isNetworkEvent && data) {
      const statusColor = data.statusCode >= 500 ? '#ff4d4f' :
        data.statusCode >= 400 ? '#faad14' :
        data.statusCode >= 300 ? '#1890ff' : '#52c41a';
      try {
        const urlPath = new URL(data.url).pathname;
        return (
          <Space size={4} wrap>
            <Tag color="blue" style={{ margin: 0 }}>{data.method}</Tag>
            <Tag color={statusColor} style={{ margin: 0 }}>{data.statusCode}</Tag>
            <Text style={{ fontSize: 12 }} ellipsis={{ tooltip: urlPath }}>{urlPath}</Text>
          </Space>
        );
      } catch {
        return <Text ellipsis>{event.title}</Text>;
      }
    }
    return (
      <Space size={4}>
        <span>{getEventIcon(event)}</span>
        <Text ellipsis style={{ maxWidth: 280 }}>{event.title}</Text>
      </Space>
    );
  };

  return (
    <Card
      size="small"
      title={getTitle()}
      extra={
        <Button type="text" size="small" onClick={onClose}>
          <CloseOutlined />
        </Button>
      }
      style={{ height: '100%', display: 'flex', flexDirection: 'column' }}
      bodyStyle={{ flex: 1, overflow: 'auto', padding: 12 }}
    >
      {!isNetworkEvent && (
        <Descriptions column={1} size="small" bordered style={{ marginBottom: 12 }}>
          <Descriptions.Item label="Time">
            <Text style={{ fontSize: 11 }}>
              {formatTimestamp(event.timestamp)} ({formatRelativeTime(event.relativeTime)})
            </Text>
          </Descriptions.Item>
          <Descriptions.Item label="Source">
            <Tag color={sourceConfig[event.source]?.color} style={{ margin: 0 }}>
              {sourceConfig[event.source]?.label || event.source}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="Level">
            <Tag color={levelConfig[event.level]?.color} style={{ margin: 0 }}>
              {levelConfig[event.level]?.label || event.level}
            </Tag>
          </Descriptions.Item>
          {event.duration && (
            <Descriptions.Item label="Duration">{event.duration}ms</Descriptions.Item>
          )}
        </Descriptions>
      )}

      {isNetworkEvent ? (
        renderData()
      ) : (
        <div>
          <Text strong style={{ fontSize: 12 }}>Data</Text>
          <div style={{ marginTop: 8 }}>
            {renderData()}
          </div>
        </div>
      )}
    </Card>
  );
});

EventDetail.displayName = 'EventDetail';

// ========================================
// 筛选器弹出框
// ========================================

interface FilterPopoverProps {
  sources: EventSource[];
  categories: EventCategory[];
  levels: EventLevel[];
  onSourcesChange: (sources: EventSource[]) => void;
  onCategoriesChange: (categories: EventCategory[]) => void;
  onLevelsChange: (levels: EventLevel[]) => void;
}

const FilterPopover = memo(({
  sources,
  categories,
  levels,
  onSourcesChange,
  onCategoriesChange,
  onLevelsChange,
}: FilterPopoverProps) => {
  const { t } = useTranslation();
  const allSources: EventSource[] = ['logcat', 'network', 'device', 'app', 'ui', 'touch', 'workflow', 'perf', 'system', 'assertion'];
  const allCategories: EventCategory[] = ['log', 'network', 'state', 'interaction', 'automation', 'diagnostic'];
  const allLevels: EventLevel[] = ['verbose', 'debug', 'info', 'warn', 'error', 'fatal'];

  return (
    <div style={{ width: 300 }}>
      <div style={{ marginBottom: 12 }}>
        <Text strong style={{ display: 'block', marginBottom: 8 }}>{t('timeline.sources')}</Text>
        <Checkbox.Group
          value={sources}
          onChange={(v) => onSourcesChange(v as EventSource[])}
          options={allSources.map(s => ({
            label: <Tag color={sourceConfig[s]?.color}>{sourceConfig[s]?.label}</Tag>,
            value: s,
          }))}
          style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}
        />
      </div>

      <div style={{ marginBottom: 12 }}>
        <Text strong style={{ display: 'block', marginBottom: 8 }}>{t('timeline.categories')}</Text>
        <Checkbox.Group
          value={categories}
          onChange={(v) => onCategoriesChange(v as EventCategory[])}
          options={allCategories.map(c => ({
            label: <Tag color={categoryConfig[c]?.color}>{categoryConfig[c]?.label}</Tag>,
            value: c,
          }))}
          style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}
        />
      </div>

      <div>
        <Text strong style={{ display: 'block', marginBottom: 8 }}>{t('timeline.levels')}</Text>
        <Checkbox.Group
          value={levels}
          onChange={(v) => onLevelsChange(v as EventLevel[])}
          options={allLevels.map(l => ({
            label: <Tag color={levelConfig[l]?.color}>{levelConfig[l]?.label}</Tag>,
            value: l,
          }))}
          style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}
        />
      </div>
    </div>
  );
});

FilterPopover.displayName = 'FilterPopover';

// ========================================
// 主组件
// ========================================

const EventTimeline = () => {
  const { t } = useTranslation();
  const { token } = theme.useToken();
  const listRef = useRef<HTMLDivElement>(null);

  // Device store
  const { devices, selectedDevice, setSelectedDevice } = useDeviceStore();

  // Event store
  const {
    sessions,
    sessionsVersion,
    activeSessionId,
    visibleEvents,
    totalEventCount,
    filteredEventCount,
    filter,
    timeIndex,
    isLoading,
    autoScroll,
    loadSession,
    startSession,
    endSession,
    setFilter,
    applyFilter,
    loadEventsInRange,
    createBookmark,
    deleteBookmark,
    subscribeToEvents,
    setAutoScroll,
  } = useEventStore();

  // Use dedicated hook for bookmarks to ensure proper re-renders
  const sessionBookmarks = useCurrentBookmarks();

  // Local state
  const [selectedEventId, setSelectedEventId] = useState<string | null>(null);
  const [detailOpen, setDetailOpen] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [filterSources, setFilterSources] = useState<EventSource[]>([]);
  const [filterCategories, setFilterCategories] = useState<EventCategory[]>([]);
  const [filterLevels, setFilterLevels] = useState<EventLevel[]>([]);
  const [timeRange, setTimeRange] = useState<{ start: number; end: number } | null>(null);
  const [currentTime, setCurrentTime] = useState(0);
  const [sessionList, setSessionList] = useState<DeviceSession[]>([]);
  const [assertionsPanelOpen, setAssertionsPanelOpen] = useState(false);
  const [configModalOpen, setConfigModalOpen] = useState(false);
  const [bookmarkModalOpen, setBookmarkModalOpen] = useState(false);
  const [bookmarkLabel, setBookmarkLabel] = useState('');

  // Subscribe to events
  useEffect(() => {
    const unsubscribe = subscribeToEvents();
    return () => unsubscribe();
  }, [subscribeToEvents]);

  // Also listen for session changes to update the dropdown list
  // Note: We use the same reference for cleanup since EventsOffAll removes by handler
  useEffect(() => {
    const EventsOn = (window as any).runtime?.EventsOn;
    if (!EventsOn) return;

    let isMounted = true;
    const currentDevice = selectedDevice;

    const refreshSessionList = async () => {
      if (isMounted && currentDevice) {
        const list = await (window as any).go.main.App.ListStoredSessions(currentDevice, 50);
        if (isMounted) {
          setSessionList(list || []);
        }
      }
    };

    EventsOn('session-started', refreshSessionList);
    EventsOn('session-ended', refreshSessionList);

    return () => {
      isMounted = false;
      // Note: Don't use EventsOff here as it removes ALL handlers for the event
      // which would break subscribeToEvents. The handlers will be garbage collected
      // when this effect cleans up.
    };
  }, [selectedDevice]);

  // Load sessions when device changes
  useEffect(() => {
    const loadSessions = async () => {
      if (selectedDevice) {
        const list = await (window as any).go.main.App.ListStoredSessions(selectedDevice, 50);
        setSessionList(list || []);
      } else {
        setSessionList([]);
      }
    };
    loadSessions();
  }, [selectedDevice]);

  // Get current session - use sessionsVersion to force recompute
  const currentSession = useMemo(() => {
    if (!activeSessionId) {
      return null;
    }
    const session = sessions.get(activeSessionId);
    return session || null;
  }, [sessions, activeSessionId, sessionsVersion]);

  // Session duration
  const sessionDuration = useMemo(() => {
    if (!currentSession) return 0;
    const end = currentSession.endTime || Date.now();
    return end - currentSession.startTime;
  }, [currentSession]);

  // Handle AI parsed query result
  const handleParsedQuery = useCallback((result: NLQueryResult | null) => {
    if (!result || !result.query) {
      // Clear AI filters when no result
      return;
    }

    const { query } = result;

    // Note: Source filtering uses AND logic, so we only set it when there's exactly ONE source
    // Multiple sources would filter to events that have ALL those sources (impossible)
    // For multiple sources, we add them to keywords for text search instead
    const validSourceNames = ['logcat', 'network', 'device', 'app', 'ui', 'touch', 'workflow', 'perf', 'system', 'assertion'];
    const collectedSources: Set<string> = new Set();

    // Collect sources from query.sources
    if (query.sources?.length) {
      query.sources.forEach((s: string) => {
        if (validSourceNames.includes(s)) {
          collectedSources.add(s);
        }
      });
    }

    // Also check types for source-like values
    if (query.types?.length) {
      const typeToSource: Record<string, string> = {
        'touch': 'touch',
        'gesture': 'touch',
        'tap': 'touch',
        'swipe': 'touch',
        'device': 'device',
        'ui': 'ui',
        'network': 'network',
        'http': 'network',
        'api': 'network',
        'logcat': 'logcat',
        'log': 'logcat',
        'app': 'app',
        'crash': 'app',
        'anr': 'app',
        'workflow': 'workflow',
        'perf': 'perf',
        'performance': 'perf',
        'system': 'system',
        'assertion': 'assertion',
      };

      query.types.forEach((t: string) => {
        const mappedSource = typeToSource[t.toLowerCase()];
        if (mappedSource) {
          collectedSources.add(mappedSource);
        } else if (validSourceNames.includes(t)) {
          collectedSources.add(t);
        }
      });
    }

    // Only set source filter if exactly one source is specified
    // Multiple sources with AND logic would return nothing
    if (collectedSources.size === 1) {
      setFilterSources(Array.from(collectedSources) as EventSource[]);
    }

    // Apply levels from AI query - levels use OR logic, safe to set multiple
    if (query.levels?.length) {
      const validLevels = query.levels.filter((l: string) =>
        ['verbose', 'debug', 'info', 'warn', 'error', 'fatal'].includes(l)
      ) as EventLevel[];
      if (validLevels.length > 0) {
        setFilterLevels(validLevels);
      }
    }

    // Apply time range from AI query
    if (query.timeRange?.last && currentSession) {
      // Parse time like "5m", "1h", "30s"
      const match = query.timeRange.last.match(/^(\d+)(s|m|h|d)$/);
      if (match) {
        const value = parseInt(match[1]);
        const unit = match[2];
        const msMultiplier = { s: 1000, m: 60000, h: 3600000, d: 86400000 };
        const duration = value * (msMultiplier[unit as keyof typeof msMultiplier] || 60000);
        const startTime = Math.max(0, sessionDuration - duration);
        setTimeRange({ start: startTime, end: sessionDuration });
      }
    }

    // Apply keywords as search text
    if (query.keywords?.length) {
      const searchPattern = query.keywords.join('|');
      setSearchText(searchPattern);
    }
  }, [currentSession, sessionDuration, setSearchText]);

  // Get time index for current session (returns array)
  const sessionTimeIndex = useMemo((): TimeIndexEntry[] => {
    if (!activeSessionId) return [];
    return timeIndex.get(activeSessionId) || [];
  }, [timeIndex, activeSessionId]);


  // Selected event - 完整的事件数据（包含 data 字段）
  const [selectedEventFull, setSelectedEventFull] = useState<UnifiedEvent | null>(null);

  // Virtual list
  const rowVirtualizer = useVirtualizer({
    count: visibleEvents.length,
    getScrollElement: () => listRef.current,
    estimateSize: () => 36,
    overscan: 10,
  });

  // Auto-scroll - debounced to prevent jumping
  const lastScrollRef = useRef<number>(0);
  useEffect(() => {
    if (autoScroll && listRef.current && visibleEvents.length > 0) {
      const now = Date.now();
      // Only scroll if 200ms has passed since last scroll
      if (now - lastScrollRef.current > 200) {
        lastScrollRef.current = now;
        requestAnimationFrame(() => {
          rowVirtualizer.scrollToIndex(visibleEvents.length - 1, { align: 'end', behavior: 'auto' });
        });
      }
    }
  }, [visibleEvents.length, autoScroll, rowVirtualizer]);


  // Update current time based on scroll position - only when auto-scroll is enabled
  useEffect(() => {
    if (autoScroll && visibleEvents.length > 0) {
      const lastVisible = visibleEvents[visibleEvents.length - 1];
      if (lastVisible) {
        setCurrentTime(lastVisible.relativeTime);
      }
    }
  }, [visibleEvents, autoScroll]);

  // Handle filter changes
  useEffect(() => {
    setFilter({
      sources: filterSources.length > 0 ? filterSources : undefined,
      categories: filterCategories.length > 0 ? filterCategories : undefined,
      levels: filterLevels.length > 0 ? filterLevels : undefined,
      searchText: searchText || undefined,
      startTime: timeRange?.start,
      endTime: timeRange?.end,
    });
    applyFilter();
  }, [filterSources, filterCategories, filterLevels, searchText, timeRange, setFilter, applyFilter]);

  // Scroll to selected event after filter changes
  useEffect(() => {
    if (selectedEventId && visibleEvents.length > 0) {
      const selectedIndex = visibleEvents.findIndex(e => e.id === selectedEventId);
      if (selectedIndex >= 0) {
        // Event is in the filtered list, scroll to it
        rowVirtualizer.scrollToIndex(selectedIndex, { align: 'center', behavior: 'auto' });
      }
    }
  }, [visibleEvents, selectedEventId, rowVirtualizer]);

  // Scroll to first event when time range is selected
  useEffect(() => {
    if (timeRange && visibleEvents.length > 0 && !selectedEventId) {
      // Scroll to the first event in the filtered list
      rowVirtualizer.scrollToIndex(0, { align: 'start', behavior: 'auto' });
    }
  }, [timeRange, visibleEvents.length, selectedEventId, rowVirtualizer]);

  // Handle time range change - scroll to start of range
  const handleTimeRangeChange = useCallback((range: { start: number; end: number } | null) => {
    setTimeRange(range);
    if (range) {
      // Update current time indicator to range start
      setCurrentTime(range.start);
      setAutoScroll(false);
    }
  }, [setAutoScroll]);

  // Handlers
  const handleSessionSelect = useCallback((sessionId: string) => {
    if (sessionId) {
      setTimeRange(null); // Clear time range when switching sessions
      loadSession(sessionId);
    }
  }, [loadSession]);

  const handleEventClick = useCallback(async (event: UnifiedEvent) => {
    setSelectedEventId(event.id);
    setDetailOpen(true);
    // 更新黄色指示器位置到该事件的时间点
    setCurrentTime(event.relativeTime);
    // 加载完整的事件数据（包含 data 字段）
    try {
      const fullEvent = await (window as any).go.main.App.GetStoredEvent(event.id);
      setSelectedEventFull(fullEvent);
    } catch (err) {
      console.error('[EventTimeline] Failed to load event details:', err);
      setSelectedEventFull(event); // 降级使用列表中的数据
    }
  }, []);

  const handleSeek = useCallback((time: number) => {
    const targetTime = Math.round(time);
    setAutoScroll(false);
    // 立即更新黄线位置，提供即时视觉反馈
    setCurrentTime(targetTime);

    // 获取当前状态
    const { visibleEvents: currentEvents, activeSessionId, totalEventCount } = useEventStore.getState();

    // 二分查找并滚动的函数
    const scrollToTime = (events: UnifiedEvent[]) => {
      if (events.length === 0) return;

      let left = 0;
      let right = events.length - 1;

      while (left < right) {
        const mid = Math.floor((left + right) / 2);
        if (events[mid].relativeTime < targetTime) {
          left = mid + 1;
        } else {
          right = mid;
        }
      }

      let closestIndex = left;
      if (left > 0) {
        const diffLeft = Math.abs(events[left].relativeTime - targetTime);
        const diffPrev = Math.abs(events[left - 1].relativeTime - targetTime);
        if (diffPrev < diffLeft) {
          closestIndex = left - 1;
        }
      }

      const foundEvent = events[closestIndex];
      if (foundEvent) {
        setCurrentTime(foundEvent.relativeTime);
      }
      rowVirtualizer.scrollToIndex(closestIndex, { align: 'center', behavior: 'smooth' });
    };

    // 检查是否需要重新加载
    const needReload = currentEvents.length === 0 ||
      currentEvents.length < totalEventCount * 0.9 ||
      (currentEvents.length > 0 && (
        targetTime < currentEvents[0].relativeTime - 1000 ||
        targetTime > currentEvents[currentEvents.length - 1].relativeTime + 1000
      ));

    if (needReload && activeSessionId) {
      // 需要重新加载，异步执行
      loadSession(activeSessionId).then(() => {
        const { visibleEvents: events } = useEventStore.getState();
        scrollToTime(events);
      });
    } else {
      // 直接用当前事件滚动
      scrollToTime(currentEvents);
    }
  }, [setAutoScroll, rowVirtualizer, loadSession]);

  const handleOpenBookmarkModal = useCallback(() => {
    if (activeSessionId && typeof currentTime === 'number') {
      setBookmarkLabel(`${t('timeline.bookmark')} ${formatRelativeTime(currentTime)}`);
      setBookmarkModalOpen(true);
    }
  }, [activeSessionId, currentTime, t]);

  const handleConfirmBookmark = useCallback(() => {
    if (activeSessionId && typeof currentTime === 'number' && bookmarkLabel.trim()) {
      createBookmark(currentTime, bookmarkLabel.trim());
      setBookmarkModalOpen(false);
      setBookmarkLabel('');
    }
  }, [activeSessionId, currentTime, bookmarkLabel, createBookmark]);

  const handleBookmarkClick = useCallback((bookmark: Bookmark) => {
    // Jump to the bookmark time
    setCurrentTime(bookmark.relativeTime);
    handleSeek(bookmark.relativeTime);
  }, [handleSeek]);

  const handleDeleteBookmark = useCallback((bookmarkId: string) => {
    deleteBookmark(bookmarkId);
  }, [deleteBookmark]);

  const handleStartSession = useCallback(() => {
    if (selectedDevice) {
      setConfigModalOpen(true);
    }
  }, [selectedDevice]);

  const handleStartWithConfig = useCallback(async (name: string, config: SessionConfig) => {
    setConfigModalOpen(false);
    if (selectedDevice) {
      try {
        const sessionId = await (window as any).go.main.App.StartSessionWithConfig(
          selectedDevice,
          name,
          config
        );
        // Reload the session to get the updated data
        if (sessionId) {
          await loadSession(sessionId);
          // Also refresh the session list so it appears in dropdown
          const list = await (window as any).go.main.App.ListStoredSessions(selectedDevice, 50);
          setSessionList(list || []);
        }
      } catch (err) {
        console.error('[EventTimeline] StartSessionWithConfig error:', err);
      }
    }
  }, [selectedDevice, loadSession]);

  const handleEndSession = useCallback(async () => {
    if (activeSessionId) {
      await endSession(activeSessionId, 'completed');
    }
  }, [activeSessionId, endSession]);

  // Device options
  const deviceOptions = useMemo(() =>
    devices.map(d => ({
      label: d.model || d.id,
      value: d.id,
    })),
    [devices]
  );

  // Session options
  const sessionOptions = useMemo(() =>
    sessionList.map(s => ({
      label: (
        <Space>
          {s.status === 'active' ? <PlayCircleOutlined /> : <HistoryOutlined />}
          <span>{s.name}</span>
          <Text type="secondary" style={{ fontSize: 10 }}>
            {new Date(s.startTime).toLocaleTimeString()}
          </Text>
        </Space>
      ),
      value: s.id,
    })),
    [sessionList]
  );

  // Get active filter count
  const activeFilterCount = filterSources.length + filterCategories.length + filterLevels.length;

  return (
    <div style={{ height: '100%', display: 'flex', padding: 16, gap: 16 }}>
      {/* Main Timeline Area */}
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0 }}>
        {/* Header */}
        <div style={{ marginBottom: 12 }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 12 }}>
            <Title level={4} style={{ margin: 0, color: token.colorText }}>
              <HistoryOutlined style={{ marginRight: 8 }} />
              {t('timeline.title')}
            </Title>

            <Space>
              <Button
                type={assertionsPanelOpen ? 'primary' : 'default'}
                icon={<ThunderboltOutlined />}
                onClick={() => setAssertionsPanelOpen(!assertionsPanelOpen)}
              >
                {t('timeline.assertions')}
              </Button>
              {currentSession?.status === 'active' ? (
                <Button
                  danger
                  icon={<PauseCircleOutlined />}
                  onClick={handleEndSession}
                >
                  {t('timeline.end_session')}
                </Button>
              ) : (
                <Button
                  type="primary"
                  icon={<PlayCircleOutlined />}
                  onClick={handleStartSession}
                  disabled={!selectedDevice}
                >
                  {t('timeline.start_session')}
                </Button>
              )}
            </Space>
          </div>

        {/* Selectors */}
        <Space wrap style={{ marginBottom: 12 }}>
          <Select
            style={{ width: 180 }}
            placeholder={t('timeline.select_device')}
            value={selectedDevice || undefined}
            onChange={setSelectedDevice}
            options={deviceOptions}
          />
          <Select
            style={{ width: 280 }}
            placeholder={t('timeline.select_session')}
            value={activeSessionId || undefined}
            onChange={handleSessionSelect}
            options={sessionOptions}
          />
        </Space>

        {/* Filters */}
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
          <SmartSearchInput
            value={searchText}
            onChange={setSearchText}
            onParsedQuery={handleParsedQuery}
            sessionId={activeSessionId || undefined}
            searchContext="events"
            style={{ width: 280 }}
          />

          <Popover
            content={
              <FilterPopover
                sources={filterSources}
                categories={filterCategories}
                levels={filterLevels}
                onSourcesChange={setFilterSources}
                onCategoriesChange={setFilterCategories}
                onLevelsChange={setFilterLevels}
              />
            }
            trigger="click"
            placement="bottomLeft"
          >
            <Badge count={activeFilterCount} size="small">
              <Button icon={<FilterOutlined />}>{t('timeline.filters')}</Button>
            </Badge>
          </Popover>

          <Popover
            trigger="click"
            placement="bottomRight"
            content={
              <div style={{ width: 280 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
                  <Text strong>{t('timeline.bookmarks')}</Text>
                  <Button
                    type="primary"
                    size="small"
                    icon={<PushpinOutlined />}
                    onClick={(e) => {
                      e.stopPropagation();
                      handleOpenBookmarkModal();
                    }}
                    disabled={!activeSessionId}
                  >
                    {t('timeline.add_bookmark')}
                  </Button>
                </div>
                {sessionBookmarks.length === 0 ? (
                  <Empty
                    image={Empty.PRESENTED_IMAGE_SIMPLE}
                    description={t('timeline.no_bookmarks')}
                    style={{ margin: '12px 0' }}
                  />
                ) : (
                  <List
                    size="small"
                    dataSource={sessionBookmarks}
                    style={{ maxHeight: 200, overflow: 'auto' }}
                    renderItem={item => (
                      <List.Item
                        style={{ padding: '4px 0', cursor: 'pointer' }}
                        onClick={() => handleBookmarkClick(item)}
                        actions={[
                          <Tooltip key="delete" title={t('common.delete')}>
                            <Button
                              type="text"
                              size="small"
                              icon={<DeleteOutlined />}
                              onClick={(e) => {
                                e.stopPropagation();
                                handleDeleteBookmark(item.id);
                              }}
                              danger
                            />
                          </Tooltip>
                        ]}
                      >
                        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                          <PushpinOutlined style={{ color: item.color || '#52c41a' }} />
                          <div>
                            <Text ellipsis style={{ maxWidth: 150 }}>{item.label}</Text>
                            <br />
                            <Text type="secondary" style={{ fontSize: 11 }}>
                              {formatRelativeTime(item.relativeTime)}
                            </Text>
                          </div>
                        </div>
                      </List.Item>
                    )}
                  />
                )}
              </div>
            }
          >
            <Badge count={sessionBookmarks.length} size="small" offset={[-5, 5]}>
              <Button icon={<BookOutlined />}>
                {t('timeline.bookmark')}
              </Button>
            </Badge>
          </Popover>

          <div style={{ flex: 1 }} />

          <Segmented
            value={autoScroll ? 'live' : 'manual'}
            onChange={v => setAutoScroll(v === 'live')}
            options={[
              { label: t('timeline.live'), value: 'live', icon: <PlayCircleOutlined /> },
              { label: t('timeline.manual'), value: 'manual', icon: <PauseCircleOutlined /> },
            ]}
          />

          <Button
            icon={<VerticalAlignBottomOutlined />}
            onClick={() => {
              if (visibleEvents.length > 0) {
                rowVirtualizer.scrollToIndex(visibleEvents.length - 1, { align: 'end' });
              }
            }}
          >
            {t('timeline.bottom')}
          </Button>
        </div>
      </div>

      {/* Time Ruler */}
      {currentSession && (
        <TimeRuler
          sessionStart={currentSession.startTime}
          sessionDuration={sessionDuration}
          timeIndex={sessionTimeIndex}
          currentTime={currentTime}
          onSeek={handleSeek}
          timeRange={timeRange}
          onTimeRangeChange={handleTimeRangeChange}
          bookmarks={sessionBookmarks}
          onBookmarkClick={handleBookmarkClick}
        />
      )}

      {/* Event count */}
      <div style={{ padding: '8px 0', display: 'flex', alignItems: 'center', gap: 12 }}>
        <Text type="secondary">
          {filteredEventCount} / {totalEventCount} events
          {visibleEvents.length < filteredEventCount && ` (showing ${visibleEvents.length})`}
        </Text>
        {currentSession?.status === 'active' && (
          <Badge status="processing" text={t('timeline.recording')} />
        )}
      </div>

      {/* Event List and Detail Panel */}
      <div style={{ flex: 1, display: 'flex', gap: 12, overflow: 'hidden' }}>
        {/* Event List */}
        <Card
          size="small"
          style={{ flex: 1, overflow: 'hidden', minWidth: 0 }}
          bodyStyle={{ padding: 0, height: '100%' }}
        >
          {isLoading ? (
            <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}>
              <Spin />
            </div>
          ) : visibleEvents.length === 0 ? (
            <Empty
              description={t('timeline.no_events')}
              style={{ marginTop: 60 }}
            />
          ) : (
            <div
              ref={listRef}
              style={{
                height: '100%',
                overflow: 'auto',
              }}
            >
              <div
                style={{
                  height: rowVirtualizer.getTotalSize(),
                  width: '100%',
                  position: 'relative',
                }}
              >
                {rowVirtualizer.getVirtualItems().map(virtualRow => {
                  const event = visibleEvents[virtualRow.index];
                  if (!event) return null;

                  return (
                    <EventRow
                      key={event.id}
                      event={event}
                      isSelected={event.id === selectedEventId}
                      onClick={() => handleEventClick(event)}
                      style={{
                        position: 'absolute',
                        top: 0,
                        left: 0,
                        width: '100%',
                        height: virtualRow.size,
                        transform: `translateY(${virtualRow.start}px)`,
                      }}
                    />
                  );
                })}
              </div>
            </div>
          )}
        </Card>

        {/* Event Detail Panel */}
        {selectedEventFull && (
          <div style={{ width: '50%', minWidth: 500, flexShrink: 0 }}>
            <EventDetail
              event={selectedEventFull}
              onClose={() => {
                setDetailOpen(false);
                setSelectedEventId(null);
                setSelectedEventFull(null);
              }}
            />
          </div>
        )}
      </div>

        {/* Session Config Modal */}
        <SessionConfigModal
          open={configModalOpen}
          onCancel={() => setConfigModalOpen(false)}
          onStart={handleStartWithConfig}
          deviceId={selectedDevice || ''}
        />

        {/* Bookmark Modal */}
        <Modal
          title={t('timeline.add_bookmark')}
          open={bookmarkModalOpen}
          onCancel={() => {
            setBookmarkModalOpen(false);
            setBookmarkLabel('');
          }}
          onOk={handleConfirmBookmark}
          okText={t('common.ok')}
          cancelText={t('common.cancel')}
          width={400}
        >
          <div style={{ marginBottom: 16 }}>
            <Text type="secondary">
              {t('timeline.bookmark_time')}: {formatRelativeTime(currentTime)}
            </Text>
          </div>
          <Input
            placeholder={t('timeline.bookmark_label_placeholder')}
            value={bookmarkLabel}
            onChange={(e) => setBookmarkLabel(e.target.value)}
            onPressEnter={handleConfirmBookmark}
            autoFocus
          />
        </Modal>
      </div>

      {/* Assertions Panel */}
      {assertionsPanelOpen && (
        <div style={{ width: 320, flexShrink: 0 }}>
          <AssertionsPanel
            sessionId={activeSessionId || ''}
            deviceId={selectedDevice || ''}
          />
        </div>
      )}
    </div>
  );
};

export default EventTimeline;
