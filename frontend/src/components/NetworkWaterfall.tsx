/**
 * NetworkWaterfall - 网络请求瀑布图组件 (虚拟列表优化版)
 * 类似 Chrome DevTools 的网络时序图
 */
import React, { useMemo, useCallback, useState } from 'react';
import { Card, Typography, Empty, theme } from 'antd';
import { useTranslation } from 'react-i18next';
import VirtualList from './VirtualList';
import type { UnifiedEvent } from '../stores/eventTypes';

const { Text } = Typography;

interface NetworkWaterfallProps {
  events: UnifiedEvent[];
  sessionStart: number;
  sessionDuration: number;
  onEventClick?: (event: UnifiedEvent) => void;
  maxHeight?: number;
}

interface NetworkRequest {
  id: string;
  method: string;
  url: string;
  statusCode: number;
  startTime: number;
  duration: number;
  size?: number;
  contentType?: string;
  event: UnifiedEvent;
}

// 获取状态码颜色
const getStatusColor = (code: number, token: any): string => {
  if (code >= 500) return token.colorError;
  if (code >= 400) return token.colorWarning;
  if (code >= 300) return token.colorInfo;
  if (code >= 200) return token.colorSuccess;
  return token.colorTextQuaternary;
};

// 获取方法文字颜色
const getMethodTagColor = (method: string, token: any): string => {
  switch (method.toUpperCase()) {
    case 'GET': return token.colorInfo;
    case 'POST': return token.colorSuccess;
    case 'PUT': return token.colorWarning;
    case 'DELETE': return token.colorError;
    case 'PATCH': return '#722ed1';
    default: return token.colorTextSecondary;
  }
};

// 格式化大小
const formatSize = (bytes?: number): string => {
  if (!bytes) return '-';
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(2)} MB`;
};

// 提取 URL 路径
const extractPath = (url: string): string => {
  try {
    const parsed = new URL(url);
    return parsed.pathname + (parsed.search ? '?' + parsed.search.slice(0, 20) : '');
  } catch {
    return url.slice(0, 50);
  }
};

const ROW_HEIGHT = 16; // 仅作为估算值

const NetworkWaterfall: React.FC<NetworkWaterfallProps> = ({
  events,
  sessionStart,
  sessionDuration,
  onEventClick,
  maxHeight = 400,
}) => {
  const { t } = useTranslation();
  const { token } = theme.useToken();
  const [hoveredIndex, setHoveredIndex] = useState<number | null>(null);

  // 提取网络请求数据
  const requests = useMemo((): NetworkRequest[] => {
    return events
      .filter(e => e.source === 'network' && (e.type === 'http_request' || e.type === 'network_request'))
      .map(e => {
        // 后端 Data 是 json.RawMessage，可能是字符串需要解析
        let data: any = {};
        const rawData = e.data || e.detail;
        if (rawData) {
          if (typeof rawData === 'string') {
            try { data = JSON.parse(rawData); } catch { data = {}; }
          } else {
            data = rawData;
          }
        }
        return {
          id: e.id,
          method: data.method || 'GET',
          url: data.url || '',
          statusCode: data.statusCode || 0,
          startTime: e.relativeTime,
          duration: e.duration || data.duration || 0,
          size: data.responseBodySize || data.bodySize,
          contentType: data.contentType,
          event: e,
        };
      })
      .sort((a, b) => a.startTime - b.startTime);
  }, [events]);

  // 计算时间轴范围
  const timeRange = useMemo(() => {
    if (requests.length === 0) return { start: 0, end: sessionDuration, maxDuration: 1000 };
    const start = Math.min(...requests.map(r => r.startTime));
    const end = Math.max(...requests.map(r => r.startTime + r.duration));
    const maxDuration = Math.max(...requests.map(r => r.duration), 1); // 最大单次请求时长
    const padding = (end - start) * 0.05;
    return {
      start: Math.max(0, start - padding),
      end: Math.min(sessionDuration, end + padding),
      maxDuration,
    };
  }, [requests, sessionDuration]);

  const timeWidth = timeRange.end - timeRange.start || 1; // 避免除零

  // 点击处理
  const handleClick = useCallback((req: NetworkRequest) => {
    onEventClick?.(req.event);
  }, [onEventClick]);

  if (requests.length === 0) {
    return (
      <Card size="small" title={t('waterfall.title', 'Network Waterfall')}>
        <Empty description={t('waterfall.no_requests', 'No network requests')} />
      </Card>
    );
  }

  return (
    <Card 
      size="small" 
      title={
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Text strong>{t('waterfall.title', 'Network Waterfall')}</Text>
          <Text type="secondary" style={{ fontSize: 12 }}>
            {requests.length} requests
          </Text>
        </div>
      }
      bodyStyle={{ padding: 0 }}
    >
      {/* 时间刻度 */}
      <div style={{ 
        display: 'flex', 
        alignItems: 'center',
        borderBottom: `1px solid ${token.colorBorder}`,
        background: token.colorBgLayout,
        padding: '2px 8px',
        fontSize: 9,
        color: token.colorTextSecondary,
        height: 20,
      }}>
        <div style={{ width: 220, flexShrink: 0 }}>Request</div>
        <div style={{ flex: 1, display: 'flex', justifyContent: 'space-between' }}>
          <span>{Math.round(timeRange.start)}ms</span>
          <span>{Math.round((timeRange.start + timeWidth / 2))}ms</span>
          <span>{Math.round(timeRange.end)}ms</span>
        </div>
        <div style={{ width: 65, flexShrink: 0, textAlign: 'right' }}>Status</div>
      </div>

      {/* 虚拟列表容器 */}
      <VirtualList
        dataSource={requests}
        rowKey="id"
        rowHeight={ROW_HEIGHT}
        height={maxHeight}
        showBorder={false}
        onItemClick={(req) => handleClick(req)}
        renderItem={(req, index) => {
          const barLeft = ((req.startTime - timeRange.start) / timeWidth) * 100;
          // 宽度基于实际时长占时间轴的比例，最小3px，最大50%
          const rawWidth = (req.duration / timeWidth) * 100;
          const barWidth = Math.max(0.5, Math.min(50, rawWidth));
          const statusColor = getStatusColor(req.statusCode, token);
          const isHovered = hoveredIndex === index;

          return (
            <div
              onMouseEnter={() => setHoveredIndex(index)}
              onMouseLeave={() => setHoveredIndex(null)}
              style={{
                height: '100%',
                display: 'flex',
                alignItems: 'center',
                padding: '0 8px',
                borderBottom: `1px solid ${token.colorSplit}`,
                background: isHovered 
                  ? token.colorPrimaryBg 
                  : index % 2 === 0 
                    ? 'transparent' 
                    : token.colorFillQuaternary,
              }}
            >
              {/* 请求信息: METHOD /path */}
              <div 
                style={{ width: 220, flexShrink: 0, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', fontSize: 10 }}
                title={`${req.method} ${req.url}`}
              >
                <span style={{ color: getMethodTagColor(req.method, token), fontWeight: 500 }}>{req.method}</span>
                <span style={{ color: token.colorText }}> {req.url ? extractPath(req.url) : '-'}</span>
              </div>

              {/* 瀑布条 */}
              <div style={{ flex: 1, height: 12, position: 'relative', flexShrink: 0 }}>
                {/* 背景网格 */}
                <div style={{
                  position: 'absolute',
                  inset: 0,
                  borderLeft: `1px dashed ${token.colorBorderSecondary}`,
                  borderRight: `1px dashed ${token.colorBorderSecondary}`,
                }}>
                  <div style={{
                    position: 'absolute',
                    left: '50%',
                    top: 0,
                    bottom: 0,
                    borderLeft: `1px dashed ${token.colorBorderSecondary}`,
                  }} />
                </div>

                {/* 请求条 */}
                <div
                  style={{
                    position: 'absolute',
                    left: `${barLeft}%`,
                    width: `${barWidth}%`,
                    minWidth: 3,
                    top: 3,
                    height: 6,
                    background: statusColor,
                    borderRadius: 1,
                  }}
                />
              </div>

              {/* 状态和耗时 */}
              <div style={{ width: 65, textAlign: 'right', flexShrink: 0, display: 'flex', alignItems: 'center', justifyContent: 'flex-end', gap: 2 }}>
                <span 
                  style={{ 
                    fontSize: 9, 
                    color: statusColor,
                    fontWeight: 500,
                  }}
                >
                  {req.statusCode || '?'}
                </span>
                <span style={{ fontSize: 9, color: token.colorTextSecondary }}>
                  {req.duration}ms
                </span>
              </div>
            </div>
          );
        }}
      />

      {/* Hover 详情 */}
      {hoveredIndex !== null && requests[hoveredIndex] && (
        <div style={{ 
          padding: '8px 12px', 
          borderTop: `1px solid ${token.colorBorder}`,
          background: token.colorBgElevated,
          fontSize: 11,
        }}>
          <div style={{ display: 'flex', gap: 16, flexWrap: 'wrap' }}>
            <span><strong>URL:</strong> {requests[hoveredIndex].url}</span>
            <span><strong>Status:</strong> {requests[hoveredIndex].statusCode}</span>
            <span><strong>Duration:</strong> {requests[hoveredIndex].duration}ms</span>
            <span><strong>Size:</strong> {formatSize(requests[hoveredIndex].size)}</span>
            {requests[hoveredIndex].contentType && (
              <span><strong>Type:</strong> {requests[hoveredIndex].contentType}</span>
            )}
          </div>
        </div>
      )}

      {/* 图例 */}
      {hoveredIndex === null && (
        <div style={{ 
          padding: '8px 12px', 
          borderTop: `1px solid ${token.colorBorder}`,
          display: 'flex',
          gap: 16,
          fontSize: 10,
          color: token.colorTextSecondary,
        }}>
          <span><span style={{ display: 'inline-block', width: 10, height: 10, background: token.colorSuccess, marginRight: 4, borderRadius: 2 }} />2xx</span>
          <span><span style={{ display: 'inline-block', width: 10, height: 10, background: token.colorInfo, marginRight: 4, borderRadius: 2 }} />3xx</span>
          <span><span style={{ display: 'inline-block', width: 10, height: 10, background: token.colorWarning, marginRight: 4, borderRadius: 2 }} />4xx</span>
          <span><span style={{ display: 'inline-block', width: 10, height: 10, background: token.colorError, marginRight: 4, borderRadius: 2 }} />5xx</span>
        </div>
      )}
    </Card>
  );
};

export default NetworkWaterfall;
