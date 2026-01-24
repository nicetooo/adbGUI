/**
 * EventLanes - 事件泳道图组件 (Canvas 优化版本)
 * 按 Source 分组显示事件，使用 Canvas 渲染大量事件点
 */
import React, { useMemo, useState, useCallback, useRef, useEffect } from 'react';
import { Card, Typography, Tooltip, Empty, theme, Switch, Tag } from 'antd';
import {
  FileTextOutlined,
  GlobalOutlined,
  MobileOutlined,
  AppstoreOutlined,
  AimOutlined,
  NodeIndexOutlined,
  DownOutlined,
  RightOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import type { UnifiedEvent, EventSource } from '../stores/eventTypes';
import { sourceConfig, formatRelativeTime } from '../stores/eventTypes';

const { Text } = Typography;

interface EventLanesProps {
  events: UnifiedEvent[];
  sessionDuration: number;
  onEventClick?: (event: UnifiedEvent) => void;
  height?: number;
}

// Source 图标映射
const sourceIcons: Record<EventSource, React.ReactNode> = {
  logcat: <FileTextOutlined />,
  network: <GlobalOutlined />,
  device: <MobileOutlined />,
  app: <AppstoreOutlined />,
  ui: <MobileOutlined />,
  touch: <AimOutlined />,
  workflow: <NodeIndexOutlined />,
  perf: <MobileOutlined />,
  system: <MobileOutlined />,
  assertion: <MobileOutlined />,
};

// Ant Design 颜色映射
const colorMap: Record<string, string> = {
  green: '#52c41a',
  purple: '#722ed1',
  cyan: '#13c2c2',
  blue: '#1890ff',
  magenta: '#eb2f96',
  orange: '#fa8c16',
  geekblue: '#2f54eb',
  gold: '#faad14',
  red: '#ff4d4f',
  default: '#8c8c8c',
};

// 获取事件颜色（根据级别，使用主题 token）
const getEventDotColor = (event: UnifiedEvent, token: any): string => {
  if (event.level === 'error' || event.level === 'fatal') return token.colorError;
  if (event.level === 'warn') return token.colorWarning;
  if (event.type === 'app_crash' || event.type === 'app_anr') return token.colorError;
  const sourceColor = sourceConfig[event.source]?.color || 'default';
  return colorMap[sourceColor];
};

interface LaneData {
  source: EventSource;
  label: string;
  color: string;
  icon: React.ReactNode;
  events: UnifiedEvent[];
  expanded: boolean;
}

// Canvas 渲染的泳道组件
interface CanvasLaneProps {
  events: UnifiedEvent[];
  sessionDuration: number;
  height: number;
  token: any;
  onEventClick?: (event: UnifiedEvent) => void;
}

const CanvasLane: React.FC<CanvasLaneProps> = ({ events, sessionDuration, height, token, onEventClick }) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const [hoveredEvent, setHoveredEvent] = useState<UnifiedEvent | null>(null);
  const [tooltipPos, setTooltipPos] = useState({ x: 0, y: 0 });

  // 预计算事件位置
  const eventPositions = useMemo(() => {
    return events.map(event => ({
      event,
      x: (event.relativeTime / sessionDuration),
      isImportant: event.level === 'error' || event.level === 'fatal' || 
                   event.type === 'app_crash' || event.type === 'app_anr',
      color: getEventDotColor(event, token),
    }));
  }, [events, sessionDuration, token]);

  // Canvas 渲染
  useEffect(() => {
    const canvas = canvasRef.current;
    const container = containerRef.current;
    if (!canvas || !container) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    // 设置 canvas 尺寸
    const rect = container.getBoundingClientRect();
    const dpr = window.devicePixelRatio || 1;
    canvas.width = rect.width * dpr;
    canvas.height = height * dpr;
    canvas.style.width = `${rect.width}px`;
    canvas.style.height = `${height}px`;
    ctx.scale(dpr, dpr);

    // 清空画布
    ctx.clearRect(0, 0, rect.width, height);

    // 绘制背景网格
    ctx.strokeStyle = token.colorBorderSecondary;
    ctx.setLineDash([3, 3]);
    ctx.lineWidth = 1;
    [0, 0.25, 0.5, 0.75, 1].forEach(ratio => {
      const x = ratio * rect.width;
      ctx.beginPath();
      ctx.moveTo(x, 0);
      ctx.lineTo(x, height);
      ctx.stroke();
    });
    ctx.setLineDash([]);

    // 绘制事件点 - 先绘制普通事件，再绘制重要事件（确保重要事件在上层）
    const centerY = height / 2;
    
    // 普通事件
    eventPositions.filter(e => !e.isImportant).forEach(({ x, color }) => {
      const px = x * rect.width;
      ctx.beginPath();
      ctx.arc(px, centerY, 3, 0, Math.PI * 2);
      ctx.fillStyle = color;
      ctx.globalAlpha = 0.7;
      ctx.fill();
    });

    // 重要事件（方形，更大）
    ctx.globalAlpha = 1;
    eventPositions.filter(e => e.isImportant).forEach(({ x, color }) => {
      const px = x * rect.width;
      ctx.fillStyle = color;
      ctx.fillRect(px - 5, centerY - 5, 10, 10);
    });

  }, [eventPositions, height, token]);

  // 鼠标移动处理 - 查找最近的事件
  const handleMouseMove = useCallback((e: React.MouseEvent<HTMLCanvasElement>) => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const rect = canvas.getBoundingClientRect();
    const x = (e.clientX - rect.left) / rect.width;
    const threshold = 10 / rect.width; // 10px 的点击区域

    // 找到最近的事件
    let closest: UnifiedEvent | null = null;
    let closestDist = threshold;

    for (const { event, x: ex, isImportant } of eventPositions) {
      const dist = Math.abs(x - ex);
      // 重要事件优先
      if (dist < closestDist || (dist < threshold && isImportant && !closest)) {
        closestDist = dist;
        closest = event;
      }
    }

    if (closest !== hoveredEvent) {
      setHoveredEvent(closest);
      if (closest) {
        setTooltipPos({ x: e.clientX, y: e.clientY });
      }
    }
  }, [eventPositions, hoveredEvent]);

  const handleMouseLeave = useCallback(() => {
    setHoveredEvent(null);
  }, []);

  const handleClick = useCallback((e: React.MouseEvent<HTMLCanvasElement>) => {
    if (hoveredEvent && onEventClick) {
      onEventClick(hoveredEvent);
    }
  }, [hoveredEvent, onEventClick]);

  return (
    <div ref={containerRef} style={{ flex: 1, position: 'relative', height }}>
      <canvas
        ref={canvasRef}
        style={{ cursor: hoveredEvent ? 'pointer' : 'default' }}
        onMouseMove={handleMouseMove}
        onMouseLeave={handleMouseLeave}
        onClick={handleClick}
      />
      {hoveredEvent && (
        <div
          style={{
            position: 'fixed',
            left: tooltipPos.x + 10,
            top: tooltipPos.y - 40,
            background: token.colorBgElevated,
            border: `1px solid ${token.colorBorder}`,
            borderRadius: 4,
            padding: '4px 8px',
            fontSize: 11,
            zIndex: 1000,
            pointerEvents: 'none',
            boxShadow: token.boxShadow,
            maxWidth: 300,
          }}
        >
          <div style={{ color: token.colorText }}>{hoveredEvent.title}</div>
          <div style={{ color: token.colorTextSecondary }}>{formatRelativeTime(hoveredEvent.relativeTime)}</div>
        </div>
      )}
    </div>
  );
};

// Mini preview 使用 Canvas
const MiniCanvasPreview: React.FC<{
  events: UnifiedEvent[];
  sessionDuration: number;
  token: any;
}> = ({ events, sessionDuration, token }) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    const container = containerRef.current;
    if (!canvas || !container) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    const rect = container.getBoundingClientRect();
    const dpr = window.devicePixelRatio || 1;
    canvas.width = rect.width * dpr;
    canvas.height = 20 * dpr;
    canvas.style.width = `${rect.width}px`;
    canvas.style.height = '20px';
    ctx.scale(dpr, dpr);

    ctx.clearRect(0, 0, rect.width, 20);

    const centerY = 10;
    events.forEach(event => {
      const x = (event.relativeTime / sessionDuration) * rect.width;
      const color = getEventDotColor(event, token);
      ctx.beginPath();
      ctx.arc(x, centerY, 1.5, 0, Math.PI * 2);
      ctx.fillStyle = color;
      ctx.globalAlpha = 0.8;
      ctx.fill();
    });
  }, [events, sessionDuration, token]);

  return (
    <div ref={containerRef} style={{ flex: 1, height: 20 }}>
      <canvas ref={canvasRef} />
    </div>
  );
};

const EventLanes: React.FC<EventLanesProps> = ({
  events,
  sessionDuration,
  onEventClick,
  height = 300,
}) => {
  const { t } = useTranslation();
  const { token } = theme.useToken();

  // 展开状态
  const [expandedLanes, setExpandedLanes] = useState<Set<EventSource>>(new Set(['logcat', 'network', 'app']));
  const [showAllSources, setShowAllSources] = useState(false);

  // 按 Source 分组
  const lanes = useMemo((): LaneData[] => {
    const groups: Record<string, UnifiedEvent[]> = {};
    events.forEach(e => {
      if (!groups[e.source]) groups[e.source] = [];
      groups[e.source].push(e);
    });

    // 按事件数量排序
    const sortedSources = Object.entries(groups)
      .sort((a, b) => b[1].length - a[1].length)
      .map(([source]) => source as EventSource);

    // 如果不显示所有，只显示前5个有事件的 source
    const displaySources = showAllSources ? sortedSources : sortedSources.slice(0, 5);

    return displaySources.map(source => ({
      source,
      label: sourceConfig[source]?.label || source,
      color: colorMap[sourceConfig[source]?.color || 'default'],
      icon: sourceIcons[source] || <MobileOutlined />,
      events: groups[source] || [],
      expanded: expandedLanes.has(source),
    }));
  }, [events, expandedLanes, showAllSources]);

  // 切换展开状态
  const toggleLane = useCallback((source: EventSource) => {
    setExpandedLanes(prev => {
      const next = new Set(prev);
      if (next.has(source)) {
        next.delete(source);
      } else {
        next.add(source);
      }
      return next;
    });
  }, []);

  if (events.length === 0) {
    return (
      <Card size="small" title={t('lanes.title', 'Event Lanes')}>
        <Empty description={t('timeline.no_events')} />
      </Card>
    );
  }

  const expandedLaneHeight = 44;

  return (
    <Card 
      size="small" 
      title={
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Text strong>{t('lanes.title', 'Event Lanes')}</Text>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <Text type="secondary" style={{ fontSize: 11 }}>
              {t('lanes.show_all', 'Show all')}
            </Text>
            <Switch 
              size="small" 
              checked={showAllSources} 
              onChange={setShowAllSources} 
            />
          </div>
        </div>
      }
      bodyStyle={{ padding: 0 }}
    >
      {/* 时间刻度 */}
      <div style={{ 
        display: 'flex', 
        borderBottom: `1px solid ${token.colorBorder}`,
        background: token.colorBgLayout,
        padding: '4px 8px',
        fontSize: 10,
        color: token.colorTextSecondary,
      }}>
        <div style={{ width: 120, flexShrink: 0 }}>Source</div>
        <div style={{ flex: 1, display: 'flex', justifyContent: 'space-between' }}>
          <span>0:00</span>
          <span>{formatRelativeTime(sessionDuration / 2)}</span>
          <span>{formatRelativeTime(sessionDuration)}</span>
        </div>
      </div>

      {/* 泳道列表 */}
      <div style={{ maxHeight: height, overflow: 'auto' }}>
        {lanes.map(lane => (
          <div key={lane.source}>
            {/* 泳道头 */}
            <div
              onClick={() => toggleLane(lane.source)}
              style={{
                display: 'flex',
                alignItems: 'center',
                padding: '4px 8px',
                borderBottom: `1px solid ${token.colorBorderSecondary}`,
                cursor: 'pointer',
                background: token.colorBgContainer,
              }}
            >
              {/* 展开图标 */}
              <span style={{ width: 16, color: token.colorTextSecondary }}>
                {lane.expanded ? <DownOutlined style={{ fontSize: 10 }} /> : <RightOutlined style={{ fontSize: 10 }} />}
              </span>

              {/* Source 信息 */}
              <div style={{ width: 104, display: 'flex', alignItems: 'center', gap: 6 }}>
                <span style={{ color: lane.color }}>{lane.icon}</span>
                <Text style={{ fontSize: 11 }}>{lane.label}</Text>
                <Tag style={{ margin: 0, fontSize: 9, padding: '0 4px' }}>{lane.events.length}</Tag>
              </div>

              {/* 迷你事件预览（未展开时显示） */}
              {!lane.expanded && (
                <MiniCanvasPreview 
                  events={lane.events} 
                  sessionDuration={sessionDuration} 
                  token={token}
                />
              )}
            </div>

            {/* 展开的泳道内容 */}
            {lane.expanded && (
              <div
                style={{
                  display: 'flex',
                  padding: '8px 8px 8px 28px',
                  borderBottom: `1px solid ${token.colorBorderSecondary}`,
                  background: token.colorBgLayout,
                  minHeight: expandedLaneHeight,
                }}
              >
                <CanvasLane
                  events={lane.events}
                  sessionDuration={sessionDuration}
                  height={expandedLaneHeight - 16}
                  token={token}
                  onEventClick={onEventClick}
                />
              </div>
            )}
          </div>
        ))}
      </div>

      {/* 图例 */}
      <div style={{ 
        padding: '8px 12px', 
        borderTop: `1px solid ${token.colorBorder}`,
        display: 'flex',
        gap: 16,
        fontSize: 10,
        color: token.colorTextSecondary,
      }}>
        <span>
          <span style={{ display: 'inline-block', width: 6, height: 6, background: token.colorTextQuaternary, marginRight: 4, borderRadius: '50%' }} />
          Normal
        </span>
        <span>
          <span style={{ display: 'inline-block', width: 6, height: 6, background: token.colorWarning, marginRight: 4, borderRadius: '50%' }} />
          Warning
        </span>
        <span>
          <span style={{ display: 'inline-block', width: 8, height: 8, background: token.colorError, marginRight: 4, borderRadius: 2 }} />
          Error/Crash
        </span>
      </div>
    </Card>
  );
};

export default EventLanes;
