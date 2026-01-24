/**
 * SessionStats - Session 统计仪表板组件
 * 显示事件分布、级别统计、时间密度等可视化图表
 */
import React, { useMemo } from 'react';
import { Card, Row, Col, Statistic, Typography, theme, Empty } from 'antd';
import {
  PieChart, Pie, Cell, ResponsiveContainer,
  BarChart, Bar, XAxis, YAxis, Tooltip,
  AreaChart, Area, CartesianGrid,
  Legend,
} from 'recharts';
import {
  BugOutlined,
  ClockCircleOutlined,
  ThunderboltOutlined,
  ApiOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import type { UnifiedEvent, EventSource, EventLevel, TimeIndexEntry } from '../stores/eventTypes';
import { sourceConfig, levelConfig } from '../stores/eventTypes';

const { Text, Title } = Typography;

interface SessionStatsProps {
  events: UnifiedEvent[];
  timeIndex: TimeIndexEntry[];
  sessionDuration: number;
}

// Ant Design 颜色映射到实际颜色值
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

const SessionStats: React.FC<SessionStatsProps> = ({ events, timeIndex, sessionDuration }) => {
  const { t } = useTranslation();
  const { token } = theme.useToken();

  // 计算来源分布
  const sourceData = useMemo(() => {
    const counts: Record<string, number> = {};
    events.forEach(e => {
      counts[e.source] = (counts[e.source] || 0) + 1;
    });
    return Object.entries(counts)
      .map(([source, count]) => ({
        name: sourceConfig[source as EventSource]?.label || source,
        value: count,
        color: colorMap[sourceConfig[source as EventSource]?.color || 'default'],
      }))
      .sort((a, b) => b.value - a.value);
  }, [events]);

  // 计算级别分布
  const levelData = useMemo(() => {
    const counts: Record<string, number> = {};
    events.forEach(e => {
      counts[e.level] = (counts[e.level] || 0) + 1;
    });
    const levels: EventLevel[] = ['verbose', 'debug', 'info', 'warn', 'error', 'fatal'];
    return levels
      .filter(level => counts[level] > 0)
      .map(level => ({
        name: levelConfig[level]?.label || level,
        count: counts[level] || 0,
        color: colorMap[levelConfig[level]?.color || 'default'],
      }));
  }, [events]);

  // 计算时间密度数据（每秒事件数）
  const densityData = useMemo(() => {
    if (!timeIndex.length) return [];
    return timeIndex.map(entry => ({
      time: entry.second,
      events: entry.eventCount,
      hasError: entry.hasError,
    }));
  }, [timeIndex]);

  // 关键指标
  const stats = useMemo(() => {
    const errorCount = events.filter(e => e.level === 'error' || e.level === 'fatal').length;
    const crashCount = events.filter(e => e.type === 'app_crash').length;
    const anrCount = events.filter(e => e.type === 'app_anr').length;
    const networkEvents = events.filter(e => e.source === 'network');
    const avgResponseTime = networkEvents.length > 0
      ? Math.round(networkEvents.reduce((sum, e) => sum + (e.duration || 0), 0) / networkEvents.length)
      : 0;
    
    return {
      total: events.length,
      errors: errorCount,
      crashes: crashCount,
      anrs: anrCount,
      networkRequests: networkEvents.length,
      avgResponseTime,
      errorRate: events.length > 0 ? ((errorCount / events.length) * 100).toFixed(1) : '0',
    };
  }, [events]);

  if (events.length === 0) {
    return (
      <Card size="small">
        <Empty description={t('timeline.no_events')} />
      </Card>
    );
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      {/* 关键指标卡片 */}
      <Row gutter={[12, 12]}>
        <Col span={6}>
          <Card size="small" bordered={false} style={{ background: token.colorBgContainer }}>
            <Statistic
              title={t('stats.total_events', 'Total Events')}
              value={stats.total}
              valueStyle={{ color: token.colorPrimary }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small" bordered={false} style={{ background: token.colorBgContainer }}>
            <Statistic
              title={t('stats.errors', 'Errors')}
              value={stats.errors}
              valueStyle={{ color: stats.errors > 0 ? token.colorError : token.colorTextSecondary }}
              prefix={<BugOutlined />}
              suffix={<Text type="secondary" style={{ fontSize: 12 }}>({stats.errorRate}%)</Text>}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small" bordered={false} style={{ background: token.colorBgContainer }}>
            <Statistic
              title={t('stats.crashes_anrs', 'Crashes/ANRs')}
              value={stats.crashes + stats.anrs}
              valueStyle={{ color: (stats.crashes + stats.anrs) > 0 ? token.colorError : token.colorTextSecondary }}
              prefix={<ThunderboltOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small" bordered={false} style={{ background: token.colorBgContainer }}>
            <Statistic
              title={t('stats.avg_response', 'Avg Response')}
              value={stats.avgResponseTime}
              suffix="ms"
              valueStyle={{ color: token.colorText }}
              prefix={<ApiOutlined />}
            />
          </Card>
        </Col>
      </Row>

      {/* 图表区域 */}
      <Row gutter={[12, 12]}>
        {/* 来源分布饼图 */}
        <Col span={8}>
          <Card 
            size="small" 
            title={<Text strong style={{ fontSize: 13 }}>{t('stats.source_distribution', 'Event Sources')}</Text>}
            bodyStyle={{ padding: '8px 12px' }}
          >
            <div style={{ height: 180 }}>
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={sourceData}
                    cx="50%"
                    cy="50%"
                    innerRadius={40}
                    outerRadius={70}
                    paddingAngle={2}
                    dataKey="value"
                  >
                    {sourceData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.color} />
                    ))}
                  </Pie>
                  <Tooltip
                    formatter={(value) => [value ?? 0, 'Count']}
                    contentStyle={{
                      background: token.colorBgElevated,
                      border: `1px solid ${token.colorBorder}`,
                      borderRadius: 4,
                      color: token.colorText,
                    }}
                    labelStyle={{ color: token.colorText }}
                    itemStyle={{ color: token.colorText }}
                  />
                  <Legend
                    layout="vertical"
                    align="right"
                    verticalAlign="middle"
                    iconSize={8}
                    wrapperStyle={{ fontSize: 11, color: token.colorText }}
                    formatter={(value) => <span style={{ color: token.colorText }}>{value}</span>}
                  />
                </PieChart>
              </ResponsiveContainer>
            </div>
          </Card>
        </Col>

        {/* 级别分布条形图 */}
        <Col span={8}>
          <Card 
            size="small" 
            title={<Text strong style={{ fontSize: 13 }}>{t('stats.level_distribution', 'Event Levels')}</Text>}
            bodyStyle={{ padding: '8px 12px' }}
          >
            <div style={{ height: 180 }}>
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={levelData} layout="vertical" margin={{ left: 10, right: 20 }}>
                  <CartesianGrid 
                    strokeDasharray="3 3" 
                    stroke={token.colorBorderSecondary}
                    horizontal={false}
                  />
                  <XAxis 
                    type="number" 
                    tick={{ fontSize: 10, fill: token.colorTextSecondary }} 
                    axisLine={{ stroke: token.colorBorder }}
                    tickLine={{ stroke: token.colorBorder }}
                  />
                  <YAxis 
                    type="category" 
                    dataKey="name" 
                    tick={{ fontSize: 10, fill: token.colorText }} 
                    width={50}
                    axisLine={{ stroke: token.colorBorder }}
                    tickLine={{ stroke: token.colorBorder }}
                  />
                  <Tooltip
                    contentStyle={{
                      background: token.colorBgElevated,
                      border: `1px solid ${token.colorBorder}`,
                      borderRadius: 4,
                      fontSize: 12,
                      color: token.colorText,
                    }}
                    labelStyle={{ color: token.colorText }}
                    itemStyle={{ color: token.colorText }}
                    cursor={{ fill: token.colorPrimaryBg, opacity: 0.3 }}
                  />
                  <Bar dataKey="count" radius={[0, 4, 4, 0]}>
                    {levelData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.color} />
                    ))}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            </div>
          </Card>
        </Col>

        {/* 时间密度面积图 */}
        <Col span={8}>
          <Card 
            size="small" 
            title={<Text strong style={{ fontSize: 13 }}>{t('stats.time_density', 'Event Density')}</Text>}
            bodyStyle={{ padding: '8px 12px' }}
          >
            <div style={{ height: 180 }}>
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={densityData} margin={{ left: -20, right: 10 }}>
                  <CartesianGrid 
                    strokeDasharray="3 3" 
                    stroke={token.colorBorderSecondary}
                    vertical={false}
                  />
                  <XAxis 
                    dataKey="time" 
                    tick={{ fontSize: 10, fill: token.colorTextSecondary }}
                    tickFormatter={(v) => `${Math.floor(v / 60)}:${(v % 60).toString().padStart(2, '0')}`}
                    axisLine={{ stroke: token.colorBorder }}
                    tickLine={{ stroke: token.colorBorder }}
                  />
                  <YAxis 
                    tick={{ fontSize: 10, fill: token.colorTextSecondary }}
                    axisLine={{ stroke: token.colorBorder }}
                    tickLine={{ stroke: token.colorBorder }}
                  />
                  <Tooltip
                    labelFormatter={(v) => `${Math.floor(Number(v) / 60)}:${(Number(v) % 60).toString().padStart(2, '0')}`}
                    formatter={(value) => [value ?? 0, 'Events']}
                    contentStyle={{
                      background: token.colorBgElevated,
                      border: `1px solid ${token.colorBorder}`,
                      borderRadius: 4,
                      fontSize: 12,
                      color: token.colorText,
                    }}
                    labelStyle={{ color: token.colorText }}
                    itemStyle={{ color: token.colorText }}
                    cursor={{ stroke: token.colorPrimary, strokeDasharray: '3 3' }}
                  />
                  <Area
                    type="monotone"
                    dataKey="events"
                    stroke={token.colorPrimary}
                    fill={token.colorPrimaryBg}
                    strokeWidth={2}
                  />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default SessionStats;
