import { useRef, useEffect, useMemo, useCallback } from "react";
import { Button, Input, Select, Space, Checkbox, message, Modal, Tooltip, Tag, theme } from "antd";
import { useTranslation } from "react-i18next";
import {
  PauseOutlined,
  PlayCircleOutlined,
  ClearOutlined,
  DownOutlined,
  InfoCircleOutlined,
  DeleteOutlined,
  StopOutlined,
} from "@ant-design/icons";
import VirtualList, { VirtualListHandle } from "./VirtualList";
import DeviceSelector from "./DeviceSelector";
import LogDetailPanel from "./LogDetailPanel";
import { useDeviceStore, useLogcatStore, FilterPreset, ParsedLog } from "../stores";
// @ts-ignore
import { main } from "../types/wails-models";
// @ts-ignore
import { ListPackages, StartApp, ForceStopApp, IsAppRunning } from "../../wailsjs/go/main/App";

const { Option } = Select;

// 日志级别颜色配置
const levelColors: Record<string, { text: string; bg: string; border: string }> = {
  'E': { text: '#ff4d4f', bg: 'rgba(255, 77, 79, 0.08)', border: '#ff4d4f' },
  'F': { text: '#cf1322', bg: 'rgba(207, 19, 34, 0.08)', border: '#cf1322' },
  'W': { text: '#faad14', bg: 'rgba(250, 173, 20, 0.08)', border: '#faad14' },
  'I': { text: '#1890ff', bg: 'transparent', border: '#1890ff' },
  'D': { text: '#52c41a', bg: 'transparent', border: '#52c41a' },
  'V': { text: '#8c8c8c', bg: 'transparent', border: '#8c8c8c' },
};

export default function LogcatView() {
  const { t } = useTranslation();
  const { token } = theme.useToken();
  const listRef = useRef<VirtualListHandle>(null);

  const { selectedDevice } = useDeviceStore();
  const {
    logs,
    isLogging,
    selectedPackage,
    logFilter,
    useRegex,
    preFilter,
    preUseRegex,
    excludeFilter,
    excludeUseRegex,
    clearLogs,
    setSelectedPackage,
    setLogFilter,
    setUseRegex,
    setPreFilter,
    setPreUseRegex,
    setExcludeFilter,
    setExcludeUseRegex,
    toggleLogcat,

    // Selection state
    selectedLogId,
    detailPanelOpen,
    selectLog,
    closeDetail,

    // UI State & Actions
    packages, setPackages,
    appRunningStatus, setAppRunningStatus,
    autoScroll, setAutoScroll,
    levelFilter, setLevelFilter,
    matchCase, setMatchCase,
    matchWholeWord, setMatchWholeWord,
    savedFilters, setSavedFilters,
    
    // Modal state
    isSaveModalOpen, setIsSaveModalOpen,
    newFilterName, setNewFilterName,
    openSaveModal, closeSaveModal,
    
    // AI Search state
    isAIParsing, setIsAIParsing,
    aiSearchText, setAiSearchText,
    aiPopoverOpen, setAiPopoverOpen,
  } = useLogcatStore();

  useEffect(() => {
    const fetchPackageList = async () => {
      if (!selectedDevice) return;
      try {
        const res = await ListPackages(selectedDevice, "user");
        setPackages(res || []);
      } catch (err) {
        console.error("Failed to fetch packages for logcat:", err);
      }
    };
    fetchPackageList();
  }, [selectedDevice]);

  useEffect(() => {
    let timer: number;
    const checkRunning = async () => {
      if (selectedDevice && selectedPackage) {
        try {
          const running = await IsAppRunning(selectedDevice, selectedPackage);
          setAppRunningStatus(running);
        } catch { }
      } else {
        setAppRunningStatus(false);
      }
    };

    checkRunning();
    // Poll every 5 seconds for status updates
    timer = window.setInterval(checkRunning, 5000);
    return () => clearInterval(timer);
  }, [selectedDevice, selectedPackage]);

  const handleStartApp = async () => {
    if (!selectedPackage || !selectedDevice) return;
    const hide = message.loading(t("app.launching", { name: selectedPackage }), 0);
    try {
      await ForceStopApp(selectedDevice, selectedPackage);
      await StartApp(selectedDevice, selectedPackage);
      message.success(t("app.start_app_success", { name: selectedPackage }));
    } catch (err) {
      message.error(t("app.start_app_failed") + ": " + String(err));
    } finally {
      hide();
    }
  };

  // 获取选中的日志对象
  const selectedLog = useMemo(() => {
    if (!selectedLogId) return null;
    return logs.find(log => log.id === selectedLogId) || null;
  }, [logs, selectedLogId]);

  // 处理日志行点击
  const handleLogClick = useCallback((log: ParsedLog) => {
    selectLog(log.id);
  }, [selectLog]);

  // 1. 深度编译过滤信息
  const filterInfo = useMemo(() => {
    const rawInput = logFilter || "";
    if (!rawInput.trim() && levelFilter.length === 0) {
      return { regex: null, invalid: false, pattern: "", highlighter: null };
    }

    try {
      let pattern = rawInput;
      if (!useRegex) {
        pattern = pattern.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
      }

      if (matchWholeWord) {
        pattern = `\\b(?:${pattern})\\b`;
      }

      const flags = matchCase ? "" : "i";
      return {
        regex: new RegExp(pattern, flags),
        highlighter: new RegExp(pattern, flags + "g"),
        invalid: false,
        pattern,
      };
    } catch (e) {
      return {
        regex: null,
        invalid: true,
        pattern: rawInput,
        highlighter: null,
      };
    }
  }, [logFilter, useRegex, matchCase, matchWholeWord]);

  // Filter Presets state handled by store
  // Persistent filters handled by store actions

  // Selected presets for Pre-Filter
  const {
    selectedPresetIds,
    selectedExcludePresetIds,
    setSelectedPresetIds,
    setSelectedExcludePresetIds,
  } = useLogcatStore();

  // Update pre-filter when presets selection changes
  useEffect(() => {
    if (selectedPresetIds.length === 0) {
      if (setPreFilter) setPreFilter("");
      return;
    }

    const selected = savedFilters.filter(f => selectedPresetIds.includes(f.id));
    if (selected.length === 0) return;

    const parts = selected.map(f => {
      if (f.isRegex) return `(${f.pattern})`;
      return `(${f.pattern.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")})`;
    });

    const combinedPattern = parts.join("|");
    if (setPreFilter) setPreFilter(combinedPattern);
    if (setPreUseRegex) setPreUseRegex(true);
  }, [selectedPresetIds, savedFilters, setPreFilter, setPreUseRegex]);

  // Update exclude-filter when presets selection changes
  useEffect(() => {
    if (selectedExcludePresetIds.length === 0) {
      if (setExcludeFilter) setExcludeFilter("");
      return;
    }

    const selected = savedFilters.filter(f => selectedExcludePresetIds.includes(f.id));
    if (selected.length === 0) return;

    const parts = selected.map(f => {
      if (f.isRegex) return `(${f.pattern})`;
      return `(${f.pattern.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")})`;
    });

    const combinedPattern = parts.join("|");
    if (setExcludeFilter) setExcludeFilter(combinedPattern);
    if (setExcludeUseRegex) setExcludeUseRegex(true);
  }, [selectedExcludePresetIds, savedFilters, setExcludeFilter, setExcludeUseRegex]);

  // Save Filter Modal state
  // Save Filter Modal state (moved to top)


  const handleSaveFilter = () => {
    if (!logFilter) {
      message.warning("Please enter a filter pattern first.");
      return;
    }
    setNewFilterName(logFilter);
    setIsSaveModalOpen(true);
  };

  const confirmSaveFilter = () => {
    if (!newFilterName.trim()) {
      message.error("Name cannot be empty");
      return;
    }
    const newPreset: FilterPreset = {
      id: Date.now().toString(),
      name: newFilterName,
      pattern: logFilter,
      isRegex: !!useRegex
    };
    setSavedFilters(prev => [...prev, newPreset]);
    message.success(t("logcat.filter_saved"));
    setIsSaveModalOpen(false);
  };

  // 2. 强力过滤引擎 - 现在使用 ParsedLog 结构
  const filteredLogs = useMemo(() => {
    if (!logFilter && levelFilter.length === 0) return logs;

    const { regex, invalid } = filterInfo;
    const hasLevelFilter = levelFilter.length > 0;

    // Always prefer using the pre-calculated regex if it's valid.
    // This ensures 'Match Case' and 'Whole Word' are respected.
    const useRegexEngine = !invalid && !!regex;

    return logs.filter((log: ParsedLog) => {
      // A. Level Filtering - 直接使用结构化的 level 字段
      if (hasLevelFilter) {
        if (!levelFilter.includes(log.level)) return false;
      }

      // B. Text/Regex Filtering - 搜索 raw 文本
      if (logFilter && !invalid) {
        const line = log.raw.replace(/\u001b\[[0-9;]*m/g, "");

        if (useRegexEngine) {
          // Use the unified regex engine (handles Case and Whole Word correctly)
          if (regex!.test(line)) return true;
          return false;
        } else {
          // Fallback to simple include (should rarely be hit if filterInfo is working)
          return line.toLowerCase().includes(logFilter.toLowerCase());
        }
      }
      return true;
    });
  }, [
    logs,
    levelFilter,
    logFilter,
    filterInfo,
    useRegex,
    matchCase,
    matchWholeWord,
  ]);

  const scrollToBottom = () => {
    setAutoScroll(true);
    listRef.current?.scrollToBottom();
  };

  // 获取日志级别颜色（用于 Checkbox）
  const getLogColor = (level: string) => {
    return levelColors[level]?.text || '#d4d4d4';
  };

  // 搜索高亮渲染
  const renderHighlight = useCallback((content: string) => {
    const { highlighter, invalid } = filterInfo;
    if (!logFilter || invalid || !highlighter) return content;

    const parts: React.ReactNode[] = [];
    let lastIndex = 0;
    let m;
    const activeHighlighter = new RegExp(highlighter.source, highlighter.flags);
    activeHighlighter.lastIndex = 0;

    while ((m = activeHighlighter.exec(content)) !== null) {
      if (m.index > lastIndex) {
        parts.push(content.substring(lastIndex, m.index));
      }
      parts.push(
        <mark key={m.index} style={{ backgroundColor: "#ffcc00", color: "#000", borderRadius: "2px", padding: "0 1px" }}>
          {m[0]}
        </mark>
      );
      lastIndex = activeHighlighter.lastIndex;
      if (m[0].length === 0) activeHighlighter.lastIndex++;
    }
    if (lastIndex < content.length) {
      parts.push(content.substring(lastIndex));
    }
    return parts.length > 0 ? parts : content;
  }, [filterInfo, logFilter]);

  // 渲染单条日志行 - 使用 ParsedLog 结构化数据
  const renderLogLine = useCallback((log: ParsedLog, isSelected: boolean) => {
    if (!log) return null;
    
    const colors = levelColors[log.level] || levelColors['V'];
    const levelTag = `${log.level}/${log.tag || 'Unknown'}`;
    const messageText = log.message || log.raw;

    return (
      <div style={{ display: 'flex', gap: 8, width: '100%', alignItems: 'center', height: 20 }}>
        {/* 时间 */}
        <span style={{
          width: 90, flexShrink: 0, color: '#888',
          fontFamily: '"JetBrains Mono", Menlo, Monaco, Consolas, monospace', fontSize: 11,
          whiteSpace: 'nowrap',
        }}>
          {log.timestamp || '--:--:--.---'}
        </span>
        {/* Level/Tag */}
        <span style={{
          width: 150, flexShrink: 0, color: colors.text, fontWeight: 500,
          fontFamily: '"JetBrains Mono", Menlo, Monaco, Consolas, monospace', fontSize: 11,
          overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap'
        }}>
          {renderHighlight(levelTag)}
        </span>
        {/* 消息 - 单行截断 */}
        <span style={{
          flex: 1, color: isSelected ? '#fff' : '#d4d4d4',
          fontFamily: '"JetBrains Mono", Menlo, Monaco, Consolas, monospace', fontSize: 11,
          overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
        }}>
          {renderHighlight(messageText)}
        </span>
      </div>
    );
  }, [renderHighlight]);

  const handleRemovePreset = (id: string) => {
    setSavedFilters(prev => prev.filter(p => p.id !== id));
    setSelectedPresetIds(prev => prev.filter(pid => pid !== id));
    setSelectedExcludePresetIds(prev => prev.filter(pid => pid !== id));
    message.success(t("logcat.filter_removed") || "Filter rule removed");
  };

  return (
    <div
      style={{
        padding: "16px 24px",
        flex: 1,
        display: "flex",
        flexDirection: "column",
        height: "100%",
        overflow: "hidden",
      }}
    >
      <div
        style={{
          marginBottom: 16,
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          flexShrink: 0,
        }}
      >
        <h2 style={{ margin: 0, color: token.colorText }}>{t("logcat.title")}</h2>
        <Space>
          <DeviceSelector />
          <span style={{ color: "#888", fontSize: "12px", marginLeft: 8 }}>{t("logcat.package") || "Package"}:</span>
          <Select
            showSearch
            value={selectedPackage}
            onChange={setSelectedPackage}
            style={{ width: 220 }}
            placeholder={t("logcat.apps_placeholder")}
            disabled={isLogging}
            allowClear
            filterOption={(input, option) => {
              const children = option?.children;
              if (!children) return false;
              const text = typeof children === 'string' ? children : String(children);
              return text.toLowerCase().indexOf(input.toLowerCase()) >= 0;
            }}
          >
            {packages.map((p) => (
              <Option key={p.name} value={p.name}>
                <Space>
                  {p.name}
                  {p.name === selectedPackage && appRunningStatus && (
                    <Tag color="success" style={{ fontSize: '10px', height: '16px', lineHeight: '14px', margin: 0, padding: '0 4px' }}>
                      RUNNING
                    </Tag>
                  )}
                </Space>
              </Option>
            ))}
          </Select>
          {selectedPackage && (
            <Tooltip title={appRunningStatus ? t("apps.force_stop") : t("apps.launch_app")}>
              <Button
                icon={appRunningStatus ? <StopOutlined /> : <PlayCircleOutlined />}
                onClick={appRunningStatus ? () => {
                  ForceStopApp(selectedDevice, selectedPackage).then(() => setAppRunningStatus(false));
                } : handleStartApp}
                style={{ color: appRunningStatus ? "#ff4d4f" : "#52c41a" }}
              />
            </Tooltip>
          )}
          <Button
            type={isLogging ? "primary" : "default"}
            danger={isLogging}
            icon={isLogging ? <PauseOutlined /> : <PlayCircleOutlined />}
            onClick={() => toggleLogcat(selectedDevice, selectedPackage)}
          >
            {isLogging ? t("logcat.stop") : t("logcat.start")}
          </Button>
          <Button icon={<ClearOutlined />} onClick={clearLogs}>
            {t("logcat.clear")}
          </Button>
        </Space>
      </div>

      <div
        style={{
          marginBottom: 12,
          display: "flex",
          flexDirection: "column",
          gap: 12,
          flexShrink: 0,
        }}
      >
        {/* Row 1: Fixed Pre-Filter (Advanced) - Moved up */}
        <div style={{ display: "flex", gap: 12, alignItems: "center", whiteSpace: "nowrap" }}>
          <span style={{ color: token.colorTextSecondary, fontSize: "12px", fontWeight: "bold", width: 80, minWidth: 80, textAlign: "right" }}>{t("logcat.pre_filter") || "Pre-Filter"}:</span>
          <div style={{ flex: 1 }}>
            <Select
              mode="multiple"
              style={{ width: "100%" }}
              placeholder={t("logcat.select_pre_filters") || "Select filters to strictly cache only matching logs (Pre-Filter)"}
              value={selectedPresetIds}
              onChange={setSelectedPresetIds}
              optionLabelProp="filterName"
              options={savedFilters.map(f => ({
                label: (
                  <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
                    <span>{f.name}</span>
                    <DeleteOutlined
                      className="delete-preset-icon"
                      style={{ color: "#999", fontSize: '12px' }}
                      onMouseEnter={(e) => e.currentTarget.style.color = "#ff4d4f"}
                      onMouseLeave={(e) => e.currentTarget.style.color = "#999"}
                      onClick={(e) => {
                        e.stopPropagation();
                        e.preventDefault();
                        handleRemovePreset(f.id);
                      }}
                    />
                  </div>
                ),
                value: f.id,
                filterName: f.name
              }))}
              maxTagCount="responsive"
              allowClear
            />
          </div>
          <div style={{ flexShrink: 0, display: "flex", alignItems: "center", gap: 4, fontSize: "11px", color: "#666", minWidth: 150 }}>
            <InfoCircleOutlined /> <span>{t("logcat.pre_filter_info") || "Only matching logs are buffered"}</span>
          </div>
        </div>

        {/* Row 1.5: Exclude Filter (New) */}
        <div style={{ display: "flex", gap: 12, alignItems: "center", whiteSpace: "nowrap" }}>
          <span style={{ color: "#ff4d4f", fontSize: "12px", fontWeight: "bold", width: 80, minWidth: 80, textAlign: "right" }}>{t("logcat.exclude_filter") || "Exclude"}:</span>
          <div style={{ flex: 1 }}>
            <Select
              mode="multiple"
              style={{ width: "100%" }}
              placeholder={t("logcat.select_exclude_filters") || "Select filters to exclude matching logs"}
              value={selectedExcludePresetIds}
              onChange={setSelectedExcludePresetIds}
              optionLabelProp="filterName"
              options={savedFilters.map(f => ({
                label: (
                  <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
                    <span>{f.name}</span>
                    <DeleteOutlined
                      className="delete-preset-icon"
                      style={{ color: "#999", fontSize: '12px' }}
                      onMouseEnter={(e) => e.currentTarget.style.color = "#ff4d4f"}
                      onMouseLeave={(e) => e.currentTarget.style.color = "#999"}
                      onClick={(e) => {
                        e.stopPropagation();
                        e.preventDefault();
                        handleRemovePreset(f.id);
                      }}
                    />
                  </div>
                ),
                value: f.id,
                filterName: f.name
              }))}
              maxTagCount="responsive"
              allowClear
            />
          </div>
          <div style={{ flexShrink: 0, display: "flex", alignItems: "center", gap: 4, fontSize: "11px", color: "#ff7875", minWidth: 150 }}>
            <StopOutlined /> <span>{t("logcat.exclude_filter_info") || "Matching logs will be dropped"}</span>
          </div>
        </div>

        {/* Row 2: Real-time View Filter (Most used) - Moved down */}
        <div style={{ display: "flex", gap: 12, alignItems: "center", whiteSpace: "nowrap" }}>
          <span style={{ color: token.colorTextSecondary, fontSize: "12px", fontWeight: "bold", width: 80, minWidth: 80, textAlign: "right" }}>{t("logcat.view_filter")}:</span>
          <div style={{ flex: 1, position: "relative" }}>
            <Input
              id="logcat-filter-input"
              placeholder={
                useRegex ? t("logcat.filter_regex") : t("logcat.filter_text")
              }
              value={logFilter}
              onChange={(e) => setLogFilter(e.target.value)}
              status={filterInfo.invalid ? "error" : ""}
              suffix={
                <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                  <div
                    style={{
                      display: "flex",
                      alignItems: "center",
                      gap: 6,
                      marginRight: 4,
                    }}
                  >
                    {logFilter && !filterInfo.invalid && (
                      <span
                        style={{
                          fontSize: "9px",
                          padding: "0 4px",
                          borderRadius: "2px",
                          backgroundColor: useRegex ? "#e6f7ff" : "#f5f5f5",
                          color: useRegex ? "#1677ff" : "#888",
                          border: `1px solid ${useRegex ? "#91d5ff" : "#d9d9d9"}`,
                          fontWeight: "bold",
                        }}
                      >
                        {useRegex ? "REG" : "TXT"}
                      </span>
                    )}
                    <span style={{ fontSize: "11px", color: "#888" }}>
                      {logFilter ? `${filteredLogs.length} / ${logs.length}` : logs.length}
                    </span>
                  </div>
                  <Space size={2} style={{ marginRight: -7 }}>
                    <Button
                      size="small"
                      type={matchCase ? "primary" : "default"}
                      style={{
                        fontSize: "11px", padding: "0 4px", height: 20, minWidth: 24, borderRadius: 2,
                        backgroundColor: matchCase ? "#1677ff" : "#f5f5f5", color: matchCase ? "#fff" : "#555",
                        border: "none", fontWeight: "bold",
                      }}
                      onClick={() => setMatchCase(!matchCase)}
                      title={t("logcat.match_case") || "Match Case (Aa)"}
                    > Aa </Button>
                    <Button
                      size="small"
                      type={matchWholeWord ? "primary" : "default"}
                      style={{
                        fontSize: "11px", padding: "0 4px", height: 20, minWidth: 24, borderRadius: 2,
                        backgroundColor: matchWholeWord ? "#1677ff" : "#f5f5f5", color: matchWholeWord ? "#fff" : "#555",
                        border: "none", fontWeight: "bold",
                      }}
                      onClick={() => setMatchWholeWord(!matchWholeWord)}
                      title={t("logcat.match_whole_word") || "Match Whole Word (W)"}
                    > W </Button>
                    <Button
                      size="small"
                      type={useRegex ? "primary" : "default"}
                      style={{
                        fontSize: "11px", padding: "0 4px", height: 20, minWidth: 24, borderRadius: 2,
                        backgroundColor: useRegex ? "#1677ff" : "#f5f5f5", color: useRegex ? "#fff" : "#555",
                        border: "none", fontWeight: "bold",
                      }}
                      onClick={() => setUseRegex(!useRegex)}
                      title={t("logcat.use_regex") || "Use Regular Expression (.*)"}
                    > .* </Button>
                    <Button
                      size="small"
                      type="primary"
                      style={{
                        fontSize: "11px", padding: "0 8px", height: 20, borderRadius: 2,
                        backgroundColor: "#52c41a", color: "#fff",
                        border: "none", fontWeight: "bold", marginLeft: 8
                      }}
                      onClick={handleSaveFilter}
                      title={t("logcat.save_filter")}
                    > {t("logcat.save") || "Save"} </Button>
                  </Space>
                </div>
              }
            />
            {logFilter && (
              <div
                style={{
                  position: "absolute", top: "100%", left: 0, fontSize: "10px",
                  color: filterInfo.invalid ? "#f5222d" : "#888", marginTop: 2,
                  fontFamily: "monospace", whiteSpace: "nowrap", overflow: "hidden",
                  textOverflow: "ellipsis", width: "100%", zIndex: 10
                }}
              >
                {filterInfo.invalid
                  ? t("logcat.invalid_regex")
                  : `${t("logcat.filter_pattern") || "Pattern"}: /${filterInfo.pattern}/${matchCase ? "" : "i"}`}
              </div>
            )}
          </div>

          <div style={{ flexShrink: 0 }}>
            <Checkbox.Group
              options={[
                { label: <span style={{ color: getLogColor("E") }}>{t("logcat.level.error")}</span>, value: "E" },
                { label: <span style={{ color: getLogColor("W") }}>{t("logcat.level.warn")}</span>, value: "W" },
                { label: <span style={{ color: getLogColor("I") }}>{t("logcat.level.info")}</span>, value: "I" },
                { label: <span style={{ color: getLogColor("D") }}>{t("logcat.level.debug")}</span>, value: "D" },
                { label: <span style={{ color: getLogColor("V") }}>{t("logcat.level.verbose")}</span>, value: "V" },
              ]}
              value={levelFilter}
              onChange={(vals) => setLevelFilter(vals as string[])}
              style={{ display: "flex", gap: "4px" }}
            />
          </div>
        </div>
      </div>

      {/* 日志列表 + 详情面板容器 */}
      <div
        style={{
          flex: 1,
          display: "flex",
          gap: 12,
          minHeight: 0,
          marginTop: 12,
        }}
      >
        {/* 日志列表 */}
        <div
          style={{
            flex: 1,
            position: "relative",
            minHeight: 0,
            backgroundColor: "#1e1e1e",
            borderRadius: "4px",
            overflow: "hidden",
          }}
        >
          <VirtualList
            ref={listRef}
            dataSource={filteredLogs}
            rowKey="id"
            rowHeight={28}
            autoScroll={autoScroll}
            onAutoScrollChange={setAutoScroll}
            selectedKey={selectedLogId}
            onItemClick={handleLogClick}
            showBorder={false}
            className="selectable"
            style={{ height: "100%", userSelect: "text", backgroundColor: "#1e1e1e" }}
            emptyText={t("logcat.no_logs")}
            renderItem={(log, index, isSelected) => {
              const colors = levelColors[log.level] || levelColors['V'];
              const isErrorOrWarn = log.level === 'E' || log.level === 'F' || log.level === 'W';
              
              return (
                <div
                  style={{
                    height: "100%",
                    padding: "4px 12px 4px 8px",
                    borderBottom: "1px solid #2d2d2d",
                    borderLeft: `3px solid ${colors.border}`,
                    color: "#d4d4d4",
                    fontFamily: '"JetBrains Mono", "Fira Code", monospace',
                    fontSize: "12px",
                    lineHeight: "20px",
                    backgroundColor: isSelected 
                      ? '#1677ff' 
                      : isErrorOrWarn 
                        ? colors.bg 
                        : 'transparent',
                    boxSizing: "border-box",
                    overflow: "hidden",
                  }}
                >
                  {renderLogLine(log, isSelected)}
                </div>
              );
            }}
          />

          {!autoScroll && filteredLogs.length > 0 && (
            <Button
              type="primary"
              shape="circle"
              icon={<DownOutlined />}
              size="large"
              onClick={scrollToBottom}
              style={{
                position: "absolute", bottom: 24, right: 24,
                boxShadow: "0 4px 12px rgba(0, 0, 0, 0.4)", zIndex: 100,
              }}
            />
          )}
        </div>

        {/* 详情面板 */}
        {detailPanelOpen && selectedLog && (
          <div style={{ width: 400, flexShrink: 0 }}>
            <LogDetailPanel log={selectedLog} onClose={closeDetail} />
          </div>
        )}
      </div>

      {/* Save Filter Modal */}
      <Modal
        title={t("logcat.save_filter")}
        open={isSaveModalOpen}
        onOk={confirmSaveFilter}
        onCancel={() => setIsSaveModalOpen(false)}
        okText={t("common.ok") || "OK"}
        cancelText={t("common.cancel") || "Cancel"}
        centered
      >
        <div style={{ marginBottom: 8 }}>{t("logcat.enter_filter_name")}</div>
        <Input
          value={newFilterName}
          onChange={e => setNewFilterName(e.target.value)}
          placeholder={t("logcat.filter_name_placeholder") || "Filter Name"}
          autoFocus
          onPressEnter={confirmSaveFilter}
        />
      </Modal>
    </div>
  );
}

