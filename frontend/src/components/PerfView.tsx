import { useEffect, useRef, useCallback } from "react";
import {
  Card,
  Button,
  Space,
  Statistic,
  Row,
  Col,
  Tag,
  Typography,
  Empty,
  Tooltip,
  Spin,
  Table,
  Progress,
  Descriptions,
  message,
} from "antd";
import type { ColumnsType } from "antd/es/table";
import {
  DashboardOutlined,
  ThunderboltOutlined,
  DatabaseOutlined,
  WifiOutlined,
  HeatMapOutlined,
  CameraOutlined,
  LoadingOutlined,
} from "@ant-design/icons";
import { useTranslation } from "react-i18next";
import {
  XAxis,
  YAxis,
  Tooltip as RechartsTooltip,
  ResponsiveContainer,
  BarChart,
  Bar,
  Cell,
} from "recharts";
import { useTheme } from "../ThemeContext";
import { useDeviceStore } from "../stores/deviceStore";
import { usePerfStore, ProcessPerfData, ProcessDetail } from "../stores/perfStore";
import DeviceSelector from "./DeviceSelector";

const { Text, Title } = Typography;

// ========================================
// Chart Configurations
// ========================================

const CHART_COLORS = {
  cpu: "#1677ff",
  memory: "#52c41a",
  fps: "#fa8c16",
  netRx: "#13c2c2",
  netTx: "#eb2f96",
  battery: "#fadb14",
  temp: "#f5222d",
};

// ========================================
// Sub Components
// ========================================

function MetricCard({
  title,
  value,
  suffix,
  precision,
  icon,
  color,
  alert,
  subValue,
  subLabel,
}: {
  title: string;
  value: number;
  suffix?: string;
  precision?: number;
  icon: React.ReactNode;
  color: string;
  alert?: boolean;
  subValue?: string;
  subLabel?: string;
}) {
  return (
    <Card
      size="small"
      style={{
        borderLeft: `3px solid ${alert ? "#f5222d" : color}`,
        backgroundColor: alert ? "rgba(245,34,45,0.04)" : undefined,
      }}
    >
      <Statistic
        title={
          <Space size={4}>
            {icon}
            <span>{title}</span>
            {alert && <Tag color="error" style={{ fontSize: 10, lineHeight: "16px", padding: "0 4px" }}>!</Tag>}
          </Space>
        }
        value={value}
        suffix={suffix}
        precision={precision ?? 1}
        valueStyle={{ color: alert ? "#f5222d" : color, fontSize: 24 }}
      />
      {subValue && (
        <Text type="secondary" style={{ fontSize: 12 }}>
          {subLabel}: {subValue}
        </Text>
      )}
    </Card>
  );
}

// ========================================
// Process Table Columns
// ========================================

function cpuColor(cpu: number): string {
  if (cpu >= 50) return "#f5222d";
  if (cpu >= 20) return "#fa8c16";
  if (cpu > 0) return "#1677ff";
  return "inherit";
}

function memColor(kb: number): string {
  const mb = kb / 1024;
  if (mb >= 500) return "#f5222d";
  if (mb >= 200) return "#fa8c16";
  if (mb > 0) return "#52c41a";
  return "inherit";
}


// Memory category colors for the bar chart
const MEM_CAT_COLORS: Record<string, string> = {
  "Graphics": "#722ed1",
  "Code": "#13c2c2",
  "Native Heap": "#fa8c16",
  "Java Heap": "#52c41a",
  "Stack": "#1677ff",
  "System": "#eb2f96",
  "Private Other": "#faad14",
  "Unknown": "#8c8c8c",
};

function oomLabel(score: number): { text: string; color: string } {
  if (score <= 0) return { text: "Foreground", color: "green" };
  if (score <= 100) return { text: "Visible", color: "blue" };
  if (score <= 200) return { text: "Perceptible", color: "cyan" };
  if (score <= 700) return { text: "Previous", color: "orange" };
  return { text: "Cached", color: "default" };
}

function ProcessDetailPanel({ detail, loading, isDark }: { detail: ProcessDetail | null; loading: boolean; isDark: boolean }) {
  const { t } = useTranslation();

  if (loading) {
    return (
      <div style={{ textAlign: "center", padding: 32 }}>
        <Spin indicator={<LoadingOutlined spin />} />
        <div style={{ marginTop: 8 }}>
          <Text type="secondary">{t("perf.detail_loading")}</Text>
        </div>
      </div>
    );
  }

  if (!detail) return null;

  // Build memory breakdown chart data, sorted by PSS desc
  const memChartData = (detail.memory || [])
    .filter((m) => m.pssKB > 0)
    .sort((a, b) => b.pssKB - a.pssKB)
    .map((m) => ({
      name: m.name,
      pss: +(m.pssKB / 1024).toFixed(1),
      color: MEM_CAT_COLORS[m.name] || "#8c8c8c",
    }));

  const oom = oomLabel(detail.oomScoreAdj);

  const javaUsedPct = detail.javaHeapSizeKB > 0
    ? Math.round((detail.javaHeapAllocKB / detail.javaHeapSizeKB) * 100)
    : 0;
  const nativeUsedPct = detail.nativeHeapSizeKB > 0
    ? Math.round((detail.nativeHeapAllocKB / detail.nativeHeapSizeKB) * 100)
    : 0;

  return (
    <div style={{ padding: "12px 16px" }}>
      <Row gutter={16}>
        {/* Left: Memory breakdown bar chart */}
        <Col span={14}>
          <Text strong style={{ fontSize: 13, display: "block", marginBottom: 8 }}>
            {t("perf.detail_memory_breakdown")} (PSS: {(detail.totalPssKB / 1024).toFixed(1)} MB)
          </Text>
          <ResponsiveContainer width="100%" height={memChartData.length * 32 + 20}>
            <BarChart data={memChartData} layout="vertical" margin={{ top: 0, right: 40, bottom: 0, left: 90 }}>
              <XAxis type="number" tick={{ fontSize: 11, fill: isDark ? "rgba(255,255,255,0.45)" : "rgba(0,0,0,0.45)" }} unit=" MB" />
              <YAxis type="category" dataKey="name" tick={{ fontSize: 11, fill: isDark ? "rgba(255,255,255,0.65)" : "rgba(0,0,0,0.65)" }} width={85} />
              <RechartsTooltip
                contentStyle={{
                  backgroundColor: isDark ? "#1f1f1f" : "#fff",
                  border: `1px solid ${isDark ? "rgba(255,255,255,0.1)" : "rgba(0,0,0,0.1)"}`,
                  borderRadius: 6, fontSize: 12,
                }}
                formatter={(value: number | undefined) => [`${value ?? 0} MB`, "PSS"]}
              />
              <Bar dataKey="pss" radius={[0, 4, 4, 0]} barSize={18}>
                {memChartData.map((entry, i) => (
                  <Cell key={i} fill={entry.color} />
                ))}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </Col>

        {/* Right: Process info + Heap details */}
        <Col span={10}>
          <Text strong style={{ fontSize: 13, display: "block", marginBottom: 8 }}>
            {t("perf.detail_process_info")}
          </Text>
          <Descriptions size="small" column={2} colon={false}
            labelStyle={{ fontSize: 12, color: isDark ? "rgba(255,255,255,0.45)" : "rgba(0,0,0,0.45)", padding: "2px 8px 2px 0" }}
            contentStyle={{ fontSize: 12, padding: "2px 0" }}
          >
            <Descriptions.Item label={t("perf.detail_threads")}>{detail.threads}</Descriptions.Item>
            <Descriptions.Item label={t("perf.detail_fd")}>{detail.fdSize}</Descriptions.Item>
            <Descriptions.Item label={t("perf.detail_activities")}>{detail.objects.activities}</Descriptions.Item>
            <Descriptions.Item label={t("perf.detail_views")}>{detail.objects.views}</Descriptions.Item>
            <Descriptions.Item label="WebViews">{detail.objects.webViews}</Descriptions.Item>
            <Descriptions.Item label="Binders">{detail.objects.localBinders + detail.objects.proxyBinders}</Descriptions.Item>
            <Descriptions.Item label="Swap">{detail.vmSwapKB > 0 ? `${(detail.vmSwapKB / 1024).toFixed(1)} MB` : "0"}</Descriptions.Item>
            <Descriptions.Item label="OOM">
              <Tag color={oom.color} style={{ fontSize: 11, lineHeight: "18px", padding: "0 4px", margin: 0 }}>{oom.text} ({detail.oomScoreAdj})</Tag>
            </Descriptions.Item>
          </Descriptions>

          {/* Heap usage bars */}
          {detail.javaHeapSizeKB > 0 && (
            <div style={{ marginTop: 12 }}>
              <Text style={{ fontSize: 12 }}>Java Heap: {(detail.javaHeapAllocKB / 1024).toFixed(1)} / {(detail.javaHeapSizeKB / 1024).toFixed(1)} MB</Text>
              <Progress percent={javaUsedPct} size="small" strokeColor="#52c41a" />
            </div>
          )}
          {detail.nativeHeapSizeKB > 0 && (
            <div style={{ marginTop: 4 }}>
              <Text style={{ fontSize: 12 }}>Native Heap: {(detail.nativeHeapAllocKB / 1024).toFixed(1)} / {(detail.nativeHeapSizeKB / 1024).toFixed(1)} MB</Text>
              <Progress percent={nativeUsedPct} size="small" strokeColor="#fa8c16" />
            </div>
          )}
        </Col>
      </Row>
    </div>
  );
}

// ========================================
// Main Component
// ========================================

function PerfView() {
  const { t } = useTranslation();
  const { isDark } = useTheme();
  const { selectedDevice } = useDeviceStore();

  const {
    isMonitoring,
    config,
    latestSample,
    cpuAlertThreshold,
    memAlertThreshold,
    selectedPid,
    processDetail,
    processDetailLoading,
    tableContainerHeight,
    startMonitoring,
    stopMonitoring,
    setSelectedMetricTab,
    setTableContainerHeight,
    subscribeToEvents,
    fetchProcessDetail,
    clearProcessDetail,
  } = usePerfStore();

  const monitoringDeviceRef = useRef<string | null>(null);
  const tableContainerRef = useRef<HTMLDivElement>(null);

  // Subscribe to backend events
  useEffect(() => {
    const unsubscribe = subscribeToEvents();
    return () => unsubscribe();
  }, []);

  // Measure table container height dynamically via ResizeObserver
  // Must re-run when hasData changes, because the table div is conditionally rendered
  const hasData = latestSample !== null;
  useEffect(() => {
    const el = tableContainerRef.current;
    if (!el) return;
    const observer = new ResizeObserver((entries) => {
      for (const entry of entries) {
        const height = Math.floor(entry.contentRect.height);
        if (height > 0) {
          setTableContainerHeight(height);
        }
      }
    });
    observer.observe(el);
    return () => observer.disconnect();
  }, [hasData]);

  // Auto-start / auto-stop monitoring
  useEffect(() => {
    if (!selectedDevice) return;

    const deviceId = selectedDevice;
    monitoringDeviceRef.current = deviceId;

    startMonitoring(deviceId, config).catch((err) => {
      console.error("Auto-start perf monitor failed:", err);
    });

    return () => {
      if (monitoringDeviceRef.current) {
        stopMonitoring(monitoringDeviceRef.current).catch(() => {});
        monitoringDeviceRef.current = null;
      }
    };
  }, [selectedDevice]);

  const handleSnapshot = useCallback(async () => {
    if (!selectedDevice) {
      message.warning(t("perf.select_device_first"));
      return;
    }
    const sample = await usePerfStore.getState().takeSnapshot(selectedDevice, "");
    if (sample) {
      message.success(t("perf.snapshot_taken"));
    }
  }, [selectedDevice]);

  // Handle row click to load process detail
  const handleRowClick = useCallback((record: ProcessPerfData) => {
    if (!selectedDevice) return;
    if (selectedPid === record.pid) {
      // Toggle off
      clearProcessDetail();
    } else {
      fetchProcessDetail(selectedDevice, record.pid);
    }
  }, [selectedDevice, selectedPid]);

  const sample = latestSample!;
  const processes = latestSample?.processes || [];

  // Alert states
  const cpuAlert = hasData && latestSample != null && latestSample.cpuUsage > cpuAlertThreshold;
  const memAlert = hasData && latestSample != null && latestSample.memUsage > memAlertThreshold;

  // Process table columns
  const processColumns: ColumnsType<ProcessPerfData> = [
    {
      title: t("perf.col_name"),
      dataIndex: "name",
      key: "name",
      ellipsis: true,
      sorter: (a, b) => a.name.localeCompare(b.name),
      render: (name: string) => (
        <Text style={{ fontFamily: "monospace", fontSize: 12 }} copyable={{ text: name }}>
          {name}
        </Text>
      ),
    },
    {
      title: t("perf.col_cpu"),
      dataIndex: "cpu",
      key: "cpu",
      width: 90,
      defaultSortOrder: "descend",
      sorter: (a, b) => a.cpu - b.cpu,
      render: (cpu: number) => (
        <span style={{ color: cpuColor(cpu), fontWeight: cpu >= 20 ? 600 : 400, fontFamily: "monospace" }}>
          {cpu > 0 ? `${cpu.toFixed(1)}%` : "0%"}
        </span>
      ),
    },
    {
      title: t("perf.col_memory"),
      dataIndex: "memoryKB",
      key: "memoryKB",
      width: 100,
      sorter: (a, b) => a.memoryKB - b.memoryKB,
      render: (kb: number) => (
        <span style={{ color: memColor(kb), fontFamily: "monospace" }}>
          {kb > 0 ? `${(kb / 1024).toFixed(1)} MB` : "-"}
        </span>
      ),
    },
    {
      title: t("perf.col_user"),
      dataIndex: "linuxUser",
      key: "linuxUser",
      width: 100,
      ellipsis: true,
      sorter: (a, b) => (a.linuxUser || "").localeCompare(b.linuxUser || ""),
      render: (user: string) => (
        <Text type="secondary" style={{ fontFamily: "monospace", fontSize: 11 }}>{user || "-"}</Text>
      ),
    },
    {
      title: "PID",
      dataIndex: "pid",
      key: "pid",
      width: 70,
      sorter: (a, b) => a.pid - b.pid,
      render: (pid: number) => (
        <Text type="secondary" style={{ fontFamily: "monospace", fontSize: 12 }}>{pid}</Text>
      ),
    },
  ];

  // Table scroll height: measured container height minus ant-table header (~39px for size="small")
  const ANT_TABLE_HEADER_HEIGHT = 39;
  const tableScrollY = tableContainerHeight
    ? tableContainerHeight - ANT_TABLE_HEADER_HEIGHT
    : 300; // Fallback before ResizeObserver fires

  return (
    <div style={{ padding: "0 20px 20px", height: "100%", display: "flex", flexDirection: "column", overflow: "hidden" }}>
      {/* Header */}
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 12, flexShrink: 0 }}>
        <Space>
          <Title level={4} style={{ margin: 0 }}>
            <DashboardOutlined /> {t("perf.title")}
          </Title>
          {isMonitoring && (
            <Tag color="processing" icon={<span className="recording-dot" />}>
              {t("perf.monitoring")}
            </Tag>
          )}
          {hasData && processes.length > 0 && (
            <Tag>{processes.length} {t("perf.processes")}</Tag>
          )}
        </Space>
        <Space>
          <DeviceSelector />
          <Tooltip title={t("perf.snapshot_tooltip")}>
            <Button
              icon={<CameraOutlined />}
              onClick={handleSnapshot}
              disabled={!selectedDevice}
              size="small"
            />
          </Tooltip>

        </Space>
      </div>

      {!selectedDevice ? (
        <Empty description={t("perf.select_device_first")} style={{ marginTop: 100 }} />
      ) : !hasData ? (
        <div style={{ textAlign: "center", marginTop: 120 }}>
          <Spin size="large" />
          <div style={{ marginTop: 16 }}>
            <Text type="secondary">{t("perf.collecting_data")}</Text>
          </div>
        </div>
      ) : (
        <>
          {/* System Overview Cards */}
          <Row gutter={[12, 12]} style={{ marginBottom: 12, flexShrink: 0 }}>
            <Col span={4}>
              <MetricCard
                title={t("perf.cpu")}
                value={sample.cpuUsage}
                suffix="%"
                icon={<ThunderboltOutlined />}
                color={CHART_COLORS.cpu}
                alert={cpuAlert}
                subValue={sample.cpuCores > 0 ? `${sample.cpuCores} ${t("perf.cores")}` : undefined}
                subLabel={t("perf.cpu_temp")}
              />
            </Col>
            <Col span={4}>
              <MetricCard
                title={t("perf.memory")}
                value={sample.memUsage}
                suffix="%"
                icon={<DatabaseOutlined />}
                color={CHART_COLORS.memory}
                alert={memAlert}
                subValue={`${sample.memUsedMB}/${sample.memTotalMB} MB`}
                subLabel={t("perf.used")}
              />
            </Col>
            <Col span={4}>
              <MetricCard
                title={t("perf.fps_label")}
                value={sample.fps}
                suffix="fps"
                precision={0}
                icon={<HeatMapOutlined />}
                color={CHART_COLORS.fps}
              />
            </Col>
            <Col span={4}>
              <MetricCard
                title={t("perf.network")}
                value={sample.netRxKBps + sample.netTxKBps}
                suffix="KB/s"
                precision={0}
                icon={<WifiOutlined />}
                color={CHART_COLORS.netRx}
                subValue={`↓${sample.netRxKBps.toFixed(0)} ↑${sample.netTxKBps.toFixed(0)}`}
                subLabel="RX/TX"
              />
            </Col>
            <Col span={4}>
              <MetricCard
                title={t("perf.battery")}
                value={sample.batteryLevel}
                suffix="%"
                precision={0}
                icon={<ThunderboltOutlined />}
                color={CHART_COLORS.battery}
                subValue={sample.batteryTemp > 0 ? `${sample.batteryTemp.toFixed(1)}°C` : undefined}
                subLabel={t("perf.temperature")}
              />
            </Col>
            <Col span={4}>
              <MetricCard
                title={t("perf.cpu_temp")}
                value={sample.cpuTempC}
                suffix="°C"
                icon={<HeatMapOutlined />}
                color={sample.cpuTempC > 70 ? "#f5222d" : CHART_COLORS.temp}
                alert={sample.cpuTempC > 70}
              />
            </Col>
          </Row>

          {/* Process Table (Task Manager) — fills remaining space */}
          <Card
            size="small"
            title={t("perf.process_list")}
            style={{ marginBottom: 12, flex: 1, minHeight: 0, display: "flex", flexDirection: "column" }}
            bodyStyle={{ padding: 0, flex: 1, minHeight: 0, display: "flex", flexDirection: "column", overflow: "hidden" }}
          >
            <div ref={tableContainerRef} style={{ flex: 1, minHeight: 0, overflow: "hidden" }}>
              <Table<ProcessPerfData>
                dataSource={processes}
                columns={processColumns}
                rowKey="pid"
                pagination={false}
                size="small"
                scroll={{ y: tableScrollY }}
                locale={{ emptyText: t("perf.no_processes") }}
                onRow={(record) => ({
                  onClick: () => handleRowClick(record),
                  style: { cursor: "pointer", backgroundColor: selectedPid === record.pid ? (isDark ? "rgba(22,119,255,0.08)" : "rgba(22,119,255,0.04)") : undefined },
                })}
                expandable={{
                  expandedRowKeys: selectedPid ? [selectedPid] : [],
                  expandedRowRender: () => (
                    <ProcessDetailPanel detail={processDetail} loading={processDetailLoading} isDark={isDark} />
                  ),
                  showExpandColumn: false,
                }}
              />
            </div>
          </Card>


        </>
      )}
    </div>
  );
}

export default PerfView;
