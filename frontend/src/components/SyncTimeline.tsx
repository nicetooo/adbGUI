import React, { useRef, useEffect, useCallback, useState } from 'react';
import { Card, Space, Typography, Tag, Tooltip, Button } from 'antd';
import {
  PlayCircleOutlined,
  PauseCircleOutlined,
  SyncOutlined,
  WarningOutlined,
  BugOutlined,
  ApiOutlined,
  HighlightOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';

const { Text } = Typography;

interface TimelineEvent {
  id: string;
  timestamp: number;
  relativeTime: number;
  type: string;
  source: string;
  level: string;
  title: string;
}

interface SyncTimelineProps {
  events: TimelineEvent[];
  videoDurationMs: number;
  videoOffsetMs: number;
  currentTimeMs: number;
  isPlaying: boolean;
  onSeek: (timeMs: number) => void;
  onEventClick?: (event: TimelineEvent) => void;
  height?: number;
}

const getEventIcon = (source: string, type: string) => {
  switch (source) {
    case 'touch':
      return <HighlightOutlined style={{ fontSize: 10 }} />;
    case 'network':
      return <ApiOutlined style={{ fontSize: 10 }} />;
    case 'app':
      if (type === 'app_crash' || type === 'app_anr') {
        return <BugOutlined style={{ fontSize: 10 }} />;
      }
      return <WarningOutlined style={{ fontSize: 10 }} />;
    default:
      return null;
  }
};

const getEventColor = (level: string, type: string): string => {
  if (type === 'app_crash' || type === 'app_anr') return '#ff4d4f';
  switch (level) {
    case 'error':
    case 'fatal':
      return '#ff4d4f';
    case 'warn':
      return '#faad14';
    case 'info':
      return '#1890ff';
    default:
      return '#8c8c8c';
  }
};

const SyncTimeline: React.FC<SyncTimelineProps> = ({
  events,
  videoDurationMs,
  videoOffsetMs,
  currentTimeMs,
  isPlaying,
  onSeek,
  onEventClick,
  height = 80,
}) => {
  const { t } = useTranslation();
  const containerRef = useRef<HTMLDivElement>(null);
  const [containerWidth, setContainerWidth] = useState(0);

  // Update container width on resize
  useEffect(() => {
    const updateWidth = () => {
      if (containerRef.current) {
        setContainerWidth(containerRef.current.offsetWidth);
      }
    };

    updateWidth();
    window.addEventListener('resize', updateWidth);
    return () => window.removeEventListener('resize', updateWidth);
  }, []);

  // Calculate position for a given time
  const timeToPosition = useCallback(
    (timeMs: number) => {
      if (videoDurationMs === 0) return 0;
      return (timeMs / videoDurationMs) * containerWidth;
    },
    [videoDurationMs, containerWidth]
  );

  // Handle click on timeline
  const handleTimelineClick = useCallback(
    (e: React.MouseEvent) => {
      if (!containerRef.current) return;
      const rect = containerRef.current.getBoundingClientRect();
      const x = e.clientX - rect.left;
      const percentage = x / containerWidth;
      const timeMs = percentage * videoDurationMs;
      onSeek(Math.max(0, Math.min(timeMs, videoDurationMs)));
    },
    [containerWidth, videoDurationMs, onSeek]
  );

  // Format time for display
  const formatTime = (ms: number) => {
    const totalSeconds = Math.floor(ms / 1000);
    const minutes = Math.floor(totalSeconds / 60);
    const seconds = totalSeconds % 60;
    return `${minutes}:${seconds.toString().padStart(2, '0')}`;
  };

  // Group events by approximate position to prevent overlap
  const groupedEvents = events.reduce((acc, event) => {
    const position = Math.floor(timeToPosition(event.relativeTime - videoOffsetMs) / 10) * 10;
    if (!acc[position]) {
      acc[position] = [];
    }
    acc[position].push(event);
    return acc;
  }, {} as Record<number, TimelineEvent[]>);

  return (
    <Card size="small" bodyStyle={{ padding: '8px 12px' }}>
      <Space direction="vertical" style={{ width: '100%' }} size="small">
        {/* Timeline header */}
        <Space style={{ justifyContent: 'space-between', width: '100%' }}>
          <Space size="small">
            <SyncOutlined style={{ color: '#1890ff' }} />
            <Text strong>{t('timeline.sync_timeline', 'Sync Timeline')}</Text>
          </Space>
          <Space size="small">
            <Tag color="blue">{events.length} {t('timeline.events', 'events')}</Tag>
            <Text type="secondary">
              {formatTime(currentTimeMs)} / {formatTime(videoDurationMs)}
            </Text>
          </Space>
        </Space>

        {/* Timeline track */}
        <div
          ref={containerRef}
          onClick={handleTimelineClick}
          style={{
            position: 'relative',
            height: height,
            backgroundColor: '#f5f5f5',
            borderRadius: 4,
            cursor: 'pointer',
            overflow: 'hidden',
          }}
        >
          {/* Progress bar */}
          <div
            style={{
              position: 'absolute',
              left: 0,
              top: 0,
              height: '100%',
              width: `${(currentTimeMs / videoDurationMs) * 100}%`,
              backgroundColor: 'rgba(24, 144, 255, 0.2)',
              transition: 'width 0.1s linear',
            }}
          />

          {/* Event markers */}
          {Object.entries(groupedEvents).map(([position, eventGroup]) => {
            const positionNum = parseInt(position);
            const firstEvent = eventGroup[0];
            const hasError = eventGroup.some(
              (e) => e.level === 'error' || e.level === 'fatal' || e.type === 'app_crash'
            );

            return (
              <Tooltip
                key={position}
                title={
                  <div>
                    {eventGroup.slice(0, 5).map((e) => (
                      <div key={e.id} style={{ marginBottom: 4 }}>
                        <Tag color={getEventColor(e.level, e.type)} style={{ marginRight: 4 }}>
                          {e.type}
                        </Tag>
                        <span>{e.title.slice(0, 50)}</span>
                      </div>
                    ))}
                    {eventGroup.length > 5 && (
                      <Text type="secondary">+{eventGroup.length - 5} more</Text>
                    )}
                  </div>
                }
              >
                <div
                  onClick={(e) => {
                    e.stopPropagation();
                    if (onEventClick) {
                      onEventClick(firstEvent);
                    }
                    onSeek(firstEvent.relativeTime - videoOffsetMs);
                  }}
                  style={{
                    position: 'absolute',
                    left: positionNum,
                    top: '50%',
                    transform: 'translate(-50%, -50%)',
                    width: 16,
                    height: 16,
                    borderRadius: '50%',
                    backgroundColor: hasError ? '#ff4d4f' : '#1890ff',
                    border: '2px solid white',
                    boxShadow: '0 2px 4px rgba(0,0,0,0.2)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    color: 'white',
                    cursor: 'pointer',
                    zIndex: 5,
                  }}
                >
                  {eventGroup.length > 1 && (
                    <Text style={{ fontSize: 8, color: 'white', fontWeight: 'bold' }}>
                      {eventGroup.length}
                    </Text>
                  )}
                </div>
              </Tooltip>
            );
          })}

          {/* Playhead */}
          <div
            style={{
              position: 'absolute',
              left: timeToPosition(currentTimeMs),
              top: 0,
              height: '100%',
              width: 2,
              backgroundColor: '#1890ff',
              zIndex: 10,
              transition: 'left 0.1s linear',
            }}
          >
            {/* Playhead handle */}
            <div
              style={{
                position: 'absolute',
                top: -4,
                left: -6,
                width: 14,
                height: 14,
                borderRadius: '50%',
                backgroundColor: '#1890ff',
                border: '2px solid white',
                boxShadow: '0 2px 4px rgba(0,0,0,0.3)',
              }}
            />
          </div>

          {/* Time markers */}
          {[0, 0.25, 0.5, 0.75, 1].map((fraction) => (
            <div
              key={fraction}
              style={{
                position: 'absolute',
                left: `${fraction * 100}%`,
                bottom: 4,
                transform: 'translateX(-50%)',
                fontSize: 10,
                color: '#999',
              }}
            >
              {formatTime(fraction * videoDurationMs)}
            </div>
          ))}
        </div>

        {/* Legend */}
        <Space size="small" wrap>
          <Space size={4}>
            <div
              style={{
                width: 10,
                height: 10,
                borderRadius: '50%',
                backgroundColor: '#ff4d4f',
              }}
            />
            <Text type="secondary" style={{ fontSize: 12 }}>
              {t('timeline.error', 'Error/Crash')}
            </Text>
          </Space>
          <Space size={4}>
            <div
              style={{
                width: 10,
                height: 10,
                borderRadius: '50%',
                backgroundColor: '#1890ff',
              }}
            />
            <Text type="secondary" style={{ fontSize: 12 }}>
              {t('timeline.normal', 'Normal')}
            </Text>
          </Space>
        </Space>
      </Space>
    </Card>
  );
};

export default SyncTimeline;
