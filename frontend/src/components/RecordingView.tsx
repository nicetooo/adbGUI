import React, { useEffect, useRef } from "react";
import {
  Button,
  Space,
  Tag,
  Card,
  List,
  Modal,
  Input,
  InputNumber,
  Select,
  message,
  theme,
  Progress,
  Empty,
  Popconfirm,
  Tooltip,
  Checkbox,
  Radio,
  Alert,
  Badge,
} from "antd";
import VirtualList from "./VirtualList";
import { useTranslation } from "react-i18next";
import {
  PlayCircleOutlined,
  StopOutlined,
  DeleteOutlined,
  SaveOutlined,
  CaretRightOutlined,
  RobotOutlined,
  EditOutlined,
  InfoCircleOutlined,
  VideoCameraOutlined,
  BranchesOutlined,
  ThunderboltOutlined,
  AimOutlined,
  LoadingOutlined,
  CheckCircleOutlined,
  ArrowUpOutlined,
  ArrowDownOutlined,
  PlusOutlined,
  CloseOutlined,
  CheckOutlined,
  ClockCircleOutlined,
} from "@ant-design/icons";
import DeviceSelector from "./DeviceSelector";
import { useDeviceStore, useAutomationStore, TouchScript } from "../stores";
import { convertScriptToWorkflow } from "../stores/automationStore";
import { formatDurationMMSS } from "../stores/eventTypes";
import { main } from "../types/wails-models";

const RecordingView: React.FC = () => {
  const { selectedDevice } = useDeviceStore();
  const {
    isRecording,
    isPlaying,
    recordingDeviceId,
    playingDeviceId,
    recordingDuration,
    recordedActionCount,
    currentScript,
    scripts,
    playbackProgress,
    recordingMode,
    isWaitingForSelector,
    pendingSelectorData,
    isPreCapturing,
    isAnalyzing,
    startRecording,
    stopRecording,
    playScript,
    stopPlayback,
    loadScripts,
    saveScript,
    deleteScript,
    deleteScripts,
    renameScript,
    setCurrentScript,
    updateRecordingDuration,
    subscribeToEvents,
    submitSelectorChoice,
    setRecordingMode,
    // RecordingView UI state from store
    selectedScriptNames,
    saveModalVisible,
    scriptName,
    renameModalVisible,
    editingScriptName,
    newScriptName,
    selectedScript,
    setSelectedScriptNames,
    setSaveModalVisible,
    setScriptName,
    openRenameModal,
    closeRenameModal,
    closeSaveModal,
    setSelectedScript,
    setNewScriptName,
    // Playback settings
    playbackSpeed,
    smartTapTimeoutMs,
    setPlaybackSpeed,
    setSmartTapTimeoutMs,
    // Script editing
    editingEventIndex,
    setEditingEventIndex,
    updateScriptEvent,
    deleteScriptEvent,
    moveScriptEvent,
    insertWaitEvent,
    isScriptDirty,
    setScriptDirty,
  } = useAutomationStore();

  const { t } = useTranslation();
  const { token } = theme.useToken();

  const durationIntervalRef = useRef<number | null>(null);

  useEffect(() => {
    const unsubscribe = subscribeToEvents();
    loadScripts();
    return () => {
      unsubscribe();
      if (durationIntervalRef.current) {
        clearInterval(durationIntervalRef.current);
      }
    };
  }, []);

  useEffect(() => {
    if (isRecording) {
      durationIntervalRef.current = window.setInterval(() => {
        updateRecordingDuration();
      }, 1000);
    } else {
      if (durationIntervalRef.current) {
        clearInterval(durationIntervalRef.current);
        durationIntervalRef.current = null;
      }
    }
    return () => {
      if (durationIntervalRef.current) {
        clearInterval(durationIntervalRef.current);
      }
    };
  }, [isRecording]);

  const handleStartRecording = async (mode: 'fast' | 'precise' = 'fast') => {
    if (!selectedDevice) {
      message.warning(t("app.select_device"));
      return;
    }
    try {
      await startRecording(selectedDevice, mode);
      message.success(t("automation.recording"));
    } catch (err) {
      message.error(String(err));
    }
  };

  const handleStopRecording = async () => {
    try {
      const script = await stopRecording();
      if (script && script.events && script.events.length > 0) {
        setCurrentScript(script);
        message.success(t("automation.events_count", { count: script.events.length }));
      } else {
        message.warning(t("automation.no_events"));
      }
    } catch (err) {
      message.error(String(err));
    }
  };

  const handleSaveScript = async () => {
    if (!currentScript) return;
    if (!scriptName.trim()) {
      message.warning(t("automation.enter_name"));
      return;
    }

    try {
      const scriptToSave = main.TouchScript.createFrom({
        ...currentScript,
        name: scriptName.trim(),
      });
      await saveScript(scriptToSave);
      message.success(t("automation.script_saved"));
      closeSaveModal();
      setCurrentScript(null);
    } catch (err) {
      message.error(String(err));
    }
  };

  const handlePlayScript = async (script: TouchScript) => {
    if (!selectedDevice) {
      message.warning(t("app.select_device"));
      return;
    }
    try {
      // Inject playback settings from store into the script
      const scriptWithSettings = main.TouchScript.createFrom({
        ...script,
        playbackSpeed: playbackSpeed,
        smartTapTimeoutMs: smartTapTimeoutMs,
      });
      await playScript(selectedDevice, scriptWithSettings);
      setSelectedScript(script);
    } catch (err) {
      message.error(String(err));
    }
  };

  const handleStopPlayback = () => {
    stopPlayback();
    setSelectedScript(null);
  };

  const handleDeleteScript = async (name: string) => {
    try {
      await deleteScript(name);
      message.success(t("automation.script_deleted"));
    } catch (err) {
      message.error(String(err));
    }
  };

  const handleBulkDeleteScripts = async () => {
    if (selectedScriptNames.length === 0) return;
    try {
      await deleteScripts(selectedScriptNames);
      message.success(t("automation.scripts_deleted_msg", { count: selectedScriptNames.length }));
      setSelectedScriptNames([]);
    } catch (err) {
      message.error(String(err));
    }
  };

  const handleRenameScript = async () => {
    if (!newScriptName) return;
    try {
      await renameScript(editingScriptName, newScriptName);
      message.success(t("automation.script_renamed"));
      closeRenameModal();
    } catch (err) {
      message.error(String(err));
    }
  };

  const handleExecuteSingleEvent = async (event: any, script: TouchScript) => {
    if (!selectedDevice) {
      message.warning(t("app.select_device"));
      return;
    }
    try {
      const touchEvent = main.TouchEvent.createFrom(event);
      await (window as any).go.main.App.ExecuteSingleTouchEvent(
        selectedDevice,
        touchEvent,
        script.resolution || ""
      );
      message.success(t("recording.event_executed"));
    } catch (err) {
      message.error(String(err));
    }
  };

  const handleConvertToWorkflow = async (script: TouchScript) => {
    try {
      await convertScriptToWorkflow(script, t);
      message.success(t("automation.converted_success"));
    } catch (err) {
      const errMsg = String(err);
      if (errMsg.includes("no_events_to_convert")) {
        message.warning(t("automation.no_events_to_convert"));
      } else {
        message.error(t("automation.convert_failed") + ": " + errMsg);
      }
    }
  };



  const formatEventDescription = (event: main.TouchEvent, index: number) => {
    // Build element info suffix if available
    let elementSuffix = '';
    if (event.selector && event.selector.value) {
      elementSuffix = ` [${event.selector.type}: ${event.selector.value}]`;
    }

    switch (event.type) {
      case "tap":
        return `${index + 1}. tap (${event.x}, ${event.y})${elementSuffix} @ ${event.timestamp}ms`;
      case "long_press":
        return `${index + 1}. long press (${event.x}, ${event.y})${elementSuffix} ${event.duration ?? 0}ms @ ${event.timestamp}ms`;
      case "swipe":
        return `${index + 1}. swipe (${event.x}, ${event.y}) → (${event.x2}, ${event.y2})${elementSuffix}${event.duration ? ` ${event.duration}ms` : ''} @ ${event.timestamp}ms`;
      case "wait":
        return `${index + 1}. wait ${event.duration ?? 0}ms`;
      default:
        return `${index + 1}. unknown`;
    }
  };

  const handleSaveEditedScript = async () => {
    const script = currentScript || selectedScript;
    if (!script) return;
    try {
      const scriptToSave = main.TouchScript.createFrom({ ...script });
      await saveScript(scriptToSave);
      message.success(t("recording.editor.changes_saved"));
      setScriptDirty(false);
      setEditingEventIndex(null);
    } catch (err) {
      message.error(String(err));
    }
  };

  const handleDiscardChanges = async () => {
    // Reload scripts to discard in-memory changes
    await loadScripts();
    setScriptDirty(false);
    setEditingEventIndex(null);
    // If editing a selectedScript, reset it from the reloaded list
    if (selectedScript) {
      const reloaded = useAutomationStore.getState().scripts.find(s => s.name === selectedScript.name);
      if (reloaded) {
        setSelectedScript(reloaded);
      }
    }
    setCurrentScript(null);
    message.info(t("recording.editor.changes_discarded"));
  };

  const handleDeleteEvent = (index: number) => {
    deleteScriptEvent(index);
    message.success(t("recording.editor.event_deleted"));
  };

  const handleMoveEvent = (fromIndex: number, direction: 'up' | 'down') => {
    const toIndex = direction === 'up' ? fromIndex - 1 : fromIndex + 1;
    moveScriptEvent(fromIndex, toIndex);
    message.success(t("recording.editor.event_moved"));
  };

  const handleInsertWait = (afterIndex: number) => {
    insertWaitEvent(afterIndex, 1000);
    message.success(t("recording.editor.wait_inserted"));
  };

  const isDeviceRecording = isRecording && recordingDeviceId === selectedDevice;
  const isDevicePlaying = isPlaying && playingDeviceId === selectedDevice;

  return (
    <div
      style={{
        padding: "16px 24px",
        height: "100%",
        display: "flex",
        flexDirection: "column",
        overflow: "hidden",
      }}
    >
      {/* Header */}
      <div
        style={{
          marginBottom: 16,
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          flexShrink: 0,
        }}
      >
        <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
          <h2 style={{ margin: 0, color: token.colorText }}>{t("recording.title")}</h2>
          <Tag color="blue">{t("recording.touch_script")}</Tag>
        </div>
        <DeviceSelector />
      </div>

      {/* Main content */}
      <div style={{ flex: 1, overflowY: "auto", paddingRight: 8 }}>
        <div style={{ display: "flex", gap: 16, alignItems: "flex-start" }}>
          {/* Left Column - Recording Control */}
          <div style={{ flex: 1, display: "flex", flexDirection: "column", gap: 16, overflow: "hidden" }}>
            {/* Recording Card */}
            <Card
              title={
                <Space>
                  <VideoCameraOutlined />
                  {t("recording.record_control")}
                </Space>
              }
              size="small"
              style={{
                border: isDeviceRecording ? `1px solid ${token.colorError}` : undefined,
                backgroundColor: isDeviceRecording ? token.colorErrorBg : undefined,
              }}
            >
              <Space direction="vertical" style={{ width: "100%" }}>
                <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
                  <Space direction="vertical" size={0}>
                    <span style={{ fontWeight: 500 }}>{t("recording.mode")}</span>
                    <Radio.Group
                      value={recordingMode}
                      onChange={(e) => setRecordingMode(e.target.value)}
                      disabled={isDeviceRecording}
                      size="small"
                      optionType="button"
                      buttonStyle="solid"
                    >
                      <Tooltip title={t("recording.fast_mode_desc")}>
                        <Radio.Button value="fast">
                          <ThunderboltOutlined /> {t("recording.fast")}
                        </Radio.Button>
                      </Tooltip>
                      <Tooltip title={t("recording.precise_mode_desc")}>
                        <Radio.Button value="precise">
                          <AimOutlined /> {t("recording.precise")}
                        </Radio.Button>
                      </Tooltip>
                    </Radio.Group>
                  </Space>

                  {isDeviceRecording ? (
                    <Button
                      type="primary"
                      danger
                      icon={<StopOutlined />}
                      onClick={handleStopRecording}
                      size="large"
                    >
                      {t("recording.stop_record")}
                    </Button>
                  ) : (
                    <Button
                      type="primary"
                      icon={<PlayCircleOutlined />}
                      onClick={() => handleStartRecording(recordingMode)}
                      disabled={!selectedDevice || isPlaying}
                      size="large"
                    >
                      {t("recording.start_record")}
                    </Button>
                  )}
                </div>

                {isDeviceRecording && (
                  <div
                    style={{
                      padding: 12,
                      backgroundColor: token.colorErrorBg,
                      borderRadius: 8,
                      border: `1px dashed ${token.colorErrorBorder}`,
                    }}
                  >
                    <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
                      <Tag color="error" icon={<div className="record-dot" />}>
                        {t("recording.recording")}
                      </Tag>
                      <span style={{ fontWeight: "bold", fontFamily: "monospace" }}>
                        {formatDurationMMSS(recordingDuration)} · {recordedActionCount} {t("recording.events")}
                      </span>
                    </div>

                    {recordingMode === 'precise' && (
                      <div style={{
                        marginTop: 12,
                        borderTop: `1px solid ${token.colorErrorBorder}`,
                        paddingTop: 8,
                        display: 'flex',
                        alignItems: 'center',
                        gap: 8,
                        color: token.colorError
                      }}>
                        {isPreCapturing ? (
                          <>
                            <LoadingOutlined spin />
                            <span>{t("recording.status.dumping")}</span>
                          </>
                        ) : isAnalyzing ? (
                          <>
                            <LoadingOutlined spin />
                            <span>{t("recording.status.analyzing")}</span>
                          </>
                        ) : isWaitingForSelector ? (
                          <>
                            <AimOutlined />
                            <span>{t("recording.status.waiting_selector")}</span>
                          </>
                        ) : (
                          <>
                            <CheckCircleOutlined style={{ color: token.colorSuccess }} />
                            <span style={{ color: token.colorSuccess }}>{t("recording.status.ready")}</span>
                          </>
                        )}
                      </div>
                    )}

                    <div style={{ fontSize: 12, color: token.colorTextSecondary, marginTop: 8 }}>
                      {t("recording.touch_device_tip")}
                    </div>
                  </div>
                )}

                {currentScript && currentScript.events && currentScript.events.length > 0 && (
                  <div
                    style={{
                      padding: 12,
                      backgroundColor: token.colorSuccessBg,
                      borderRadius: 8,
                      border: `1px solid ${token.colorSuccessBorder}`,
                    }}
                  >
                    <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
                      <span>
                        {t("recording.events_count", { count: currentScript.events.length })}
                      </span>
                      <Space>
                        <Tooltip title={t("recording.test_tooltip")}>
                          <Button
                            icon={<CaretRightOutlined />}
                            size="small"
                            onClick={() => handlePlayScript(currentScript)}
                            disabled={isPlaying || !selectedDevice}
                          >
                            {t("recording.test_script")}
                          </Button>
                        </Tooltip>
                        <Tooltip title={t("automation.convert_to_workflow")}>
                          <Button
                            icon={<BranchesOutlined />}
                            size="small"
                            onClick={() => handleConvertToWorkflow(currentScript)}
                            disabled={isRecording || isPlaying}
                          >
                            {t("automation.convert")}
                          </Button>
                        </Tooltip>
                        <Button
                          type="primary"
                          icon={<SaveOutlined />}
                          size="small"
                          onClick={() => setSaveModalVisible(true)}
                        >
                          {t("recording.save_script")}
                        </Button>
                      </Space>
                    </div>
                  </div>
                )}
              </Space>
            </Card>

            {/* Playback Progress */}
            {isDevicePlaying && playbackProgress && (
              <Card
                title={
                  <Space>
                    <CaretRightOutlined />
                    {t("recording.playing")}
                  </Space>
                }
                size="small"
                style={{
                  border: `1px solid ${token.colorPrimaryBorder}`,
                  backgroundColor: token.colorPrimaryBg,
                }}
              >
                <div style={{ marginBottom: 12 }}>
                  <Progress
                    percent={Math.round((playbackProgress.current / playbackProgress.total) * 100)}
                    status="active"
                    format={() => `${playbackProgress.current}/${playbackProgress.total}`}
                  />
                </div>
                <Button danger icon={<StopOutlined />} onClick={handleStopPlayback}>
                  {t("recording.stop")}
                </Button>
              </Card>
            )}

            {/* Script Preview / Editor */}
            {(currentScript || selectedScript) && (
              <Card
                title={
                  <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
                    <Space>
                      <EditOutlined />
                      {t("recording.editor.title")}
                      {isScriptDirty && (
                        <Tag color="warning">{t("recording.editor.unsaved_changes")}</Tag>
                      )}
                    </Space>
                    {isScriptDirty && (
                      <Space size={4}>
                        <Button
                          type="primary"
                          size="small"
                          icon={<SaveOutlined />}
                          onClick={handleSaveEditedScript}
                          disabled={!(currentScript || selectedScript)?.name}
                        >
                          {t("recording.editor.save_changes")}
                        </Button>
                        <Button
                          size="small"
                          icon={<CloseOutlined />}
                          onClick={handleDiscardChanges}
                        >
                          {t("recording.editor.discard_changes")}
                        </Button>
                      </Space>
                    )}
                  </div>
                }
                size="small"
                style={{ maxHeight: 400, overflow: "auto" }}
              >
                <div style={{ fontFamily: "monospace", fontSize: 12 }}>
                  {((currentScript || selectedScript)?.events || []).map((event, idx) => {
                    const activeScript = currentScript || selectedScript;
                    const totalEvents = activeScript?.events?.length || 0;
                    const isEditing = editingEventIndex === idx;

                    return (
                      <div key={idx}>
                        <div
                          style={{
                            padding: "6px 0",
                            borderBottom: `1px solid ${token.colorBorderSecondary}`,
                            color:
                              isDevicePlaying && playbackProgress && idx < playbackProgress.current
                                ? token.colorSuccess
                                : token.colorText,
                            backgroundColor: isEditing ? token.colorPrimaryBg : undefined,
                            borderRadius: isEditing ? 4 : 0,
                          }}
                        >
                          {/* Event row */}
                          <div style={{ display: "flex", alignItems: "center", gap: 4 }}>
                            <Tooltip title={t("recording.execute_event")}>
                              <Button
                                type="text"
                                size="small"
                                icon={<ThunderboltOutlined />}
                                onClick={() => handleExecuteSingleEvent(event, activeScript!)}
                                disabled={!selectedDevice || isRecording || isPlaying}
                                style={{ padding: "0 4px", minWidth: 24 }}
                              />
                            </Tooltip>
                            <span
                              style={{ flex: 1, wordBreak: "break-all", overflowWrap: "anywhere", cursor: "pointer" }}
                              onClick={() => setEditingEventIndex(isEditing ? null : idx)}
                            >
                              {formatEventDescription(event, idx)}
                            </span>
                            <Space size={0}>
                              <Tooltip title={t("recording.editor.move_up")}>
                                <Button
                                  type="text" size="small"
                                  icon={<ArrowUpOutlined />}
                                  disabled={idx === 0 || isPlaying}
                                  onClick={() => handleMoveEvent(idx, 'up')}
                                  style={{ padding: "0 2px", minWidth: 20 }}
                                />
                              </Tooltip>
                              <Tooltip title={t("recording.editor.move_down")}>
                                <Button
                                  type="text" size="small"
                                  icon={<ArrowDownOutlined />}
                                  disabled={idx >= totalEvents - 1 || isPlaying}
                                  onClick={() => handleMoveEvent(idx, 'down')}
                                  style={{ padding: "0 2px", minWidth: 20 }}
                                />
                              </Tooltip>
                              <Tooltip title={t("recording.editor.edit_event")}>
                                <Button
                                  type="text" size="small"
                                  icon={<EditOutlined />}
                                  onClick={() => setEditingEventIndex(isEditing ? null : idx)}
                                  style={{ padding: "0 2px", minWidth: 20, color: isEditing ? token.colorPrimary : undefined }}
                                />
                              </Tooltip>
                              <Tooltip title={t("recording.editor.insert_wait_after")}>
                                <Button
                                  type="text" size="small"
                                  icon={<ClockCircleOutlined />}
                                  onClick={() => handleInsertWait(idx)}
                                  disabled={isPlaying}
                                  style={{ padding: "0 2px", minWidth: 20 }}
                                />
                              </Tooltip>
                              <Popconfirm
                                title={t("recording.editor.delete_event_confirm")}
                                onConfirm={() => handleDeleteEvent(idx)}
                                okText={t("common.ok")}
                                cancelText={t("common.cancel")}
                              >
                                <Tooltip title={t("recording.editor.delete_event")}>
                                  <Button
                                    type="text" size="small" danger
                                    icon={<DeleteOutlined />}
                                    disabled={isPlaying}
                                    style={{ padding: "0 2px", minWidth: 20 }}
                                  />
                                </Tooltip>
                              </Popconfirm>
                            </Space>
                          </div>

                          {/* Inline editor */}
                          {isEditing && (
                            <div style={{
                              margin: "8px 0 4px 32px",
                              padding: 8,
                              backgroundColor: token.colorFillAlter,
                              borderRadius: 6,
                              display: "flex",
                              flexWrap: "wrap",
                              gap: 8,
                              alignItems: "center",
                            }}>
                              {event.type !== 'wait' && (
                                <>
                                  <Space size={4}>
                                    <span>{t("recording.editor.x")}:</span>
                                    <InputNumber
                                      size="small" style={{ width: 70 }}
                                      value={event.x}
                                      onChange={(v: number | null) => v !== null && updateScriptEvent(idx, { x: v })}
                                    />
                                  </Space>
                                  <Space size={4}>
                                    <span>{t("recording.editor.y")}:</span>
                                    <InputNumber
                                      size="small" style={{ width: 70 }}
                                      value={event.y}
                                      onChange={(v: number | null) => v !== null && updateScriptEvent(idx, { y: v })}
                                    />
                                  </Space>
                                </>
                              )}
                              {event.type === 'swipe' && (
                                <>
                                  <Space size={4}>
                                    <span>{t("recording.editor.x2")}:</span>
                                    <InputNumber
                                      size="small" style={{ width: 70 }}
                                      value={event.x2}
                                      onChange={(v: number | null) => v !== null && updateScriptEvent(idx, { x2: v })}
                                    />
                                  </Space>
                                  <Space size={4}>
                                    <span>{t("recording.editor.y2")}:</span>
                                    <InputNumber
                                      size="small" style={{ width: 70 }}
                                      value={event.y2}
                                      onChange={(v: number | null) => v !== null && updateScriptEvent(idx, { y2: v })}
                                    />
                                  </Space>
                                </>
                              )}
                              {(event.type === 'long_press' || event.type === 'swipe' || event.type === 'wait') && (
                                <Space size={4}>
                                  <span>{t("recording.editor.duration_ms")}:</span>
                                  <InputNumber
                                    size="small" style={{ width: 80 }}
                                    min={0}
                                    value={event.duration}
                                    onChange={(v: number | null) => v !== null && updateScriptEvent(idx, { duration: v })}
                                  />
                                </Space>
                              )}
                              <Button
                                type="text" size="small"
                                icon={<CheckOutlined />}
                                onClick={() => setEditingEventIndex(null)}
                                style={{ color: token.colorPrimary }}
                              >
                                {t("common.ok")}
                              </Button>
                            </div>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>
              </Card>
            )}
          </div>

          {/* Right Column - Script List */}
          <div style={{ flex: 1, display: "flex", flexDirection: "column", overflow: "hidden" }}>
            <Card
              title={
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <Space>
                    <Checkbox
                      indeterminate={selectedScriptNames.length > 0 && selectedScriptNames.length < scripts.length}
                      checked={scripts.length > 0 && selectedScriptNames.length === scripts.length}
                      onChange={(e) => setSelectedScriptNames(e.target.checked ? scripts.map(s => s.name) : [])}
                    />
                    <SaveOutlined />
                    {t("recording.saved_scripts")}
                    {selectedScriptNames.length > 0 && (
                      <Popconfirm
                        title={t("recording.delete_selected_confirm", { count: selectedScriptNames.length })}
                        onConfirm={handleBulkDeleteScripts}
                        okText={t("common.ok")}
                        cancelText={t("common.cancel")}
                      >
                        <Button danger size="small" icon={<DeleteOutlined />}>
                          {t("recording.delete_selected")} ({selectedScriptNames.length})
                        </Button>
                      </Popconfirm>
                    )}
                  </Space>
                  <Tooltip title={t("recording.scaling_info")}>
                    <Tag icon={<InfoCircleOutlined />} color="blue" style={{ margin: 0 }}>
                      {t("recording.auto_scaling")}
                    </Tag>
                  </Tooltip>
                </div>
              }
              size="small"
              styles={{ body: { padding: 0 } }}
            >
              {/* Playback Settings */}
              <div style={{
                display: 'flex',
                alignItems: 'center',
                gap: 12,
                padding: '6px 12px',
                borderBottom: `1px solid ${token.colorBorderSecondary}`,
                fontSize: 12,
                flexWrap: 'wrap',
              }}>
                <Space size={4} align="center">
                  <ThunderboltOutlined style={{ color: token.colorTextSecondary }} />
                  <span style={{ color: token.colorTextSecondary }}>{t("recording.playback_speed")}:</span>
                  <Select
                    size="small"
                    value={playbackSpeed}
                    onChange={setPlaybackSpeed}
                    style={{ width: 80 }}
                    options={[
                      { value: 0.5, label: '0.5x' },
                      { value: 1, label: '1x' },
                      { value: 2, label: '2x' },
                      { value: 5, label: '5x' },
                    ]}
                  />
                </Space>
                <Space size={4} align="center">
                  <AimOutlined style={{ color: token.colorTextSecondary }} />
                  <span style={{ color: token.colorTextSecondary }}>{t("recording.smart_tap_timeout")}:</span>
                  <InputNumber
                    size="small"
                    value={smartTapTimeoutMs}
                    onChange={(v) => setSmartTapTimeoutMs(v || 5000)}
                    min={1000}
                    max={30000}
                    step={1000}
                    style={{ width: 80 }}
                    suffix="ms"
                  />
                </Space>
              </div>
              <VirtualList<TouchScript>
                dataSource={scripts}
                rowKey="name"
                height="calc(100vh - 200px)"
                rowHeight={60}
                emptyText={t("recording.no_scripts")}
                selectedKey={selectedScript?.name}
                onItemClick={(script) => {
                  setSelectedScript(script);
                  setCurrentScript(null);
                }}
                showBorder={true}
                renderItem={(script, _index, isSelected) => (
                  <div
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'space-between',
                      padding: '8px 16px',
                      height: '100%',
                      backgroundColor: isSelected ? token.colorPrimaryBg : undefined,
                    }}
                  >
                    <div style={{ display: 'flex', alignItems: 'center', flex: 1, minWidth: 0 }}>
                      <Checkbox
                        checked={selectedScriptNames.includes(script.name)}
                        onChange={(e) => {
                          e.stopPropagation();
                          const name = script.name;
                          if (e.target.checked) {
                            setSelectedScriptNames([...selectedScriptNames, name]);
                          } else {
                            setSelectedScriptNames(selectedScriptNames.filter(n => n !== name));
                          }
                        }}
                        onClick={(e) => e.stopPropagation()}
                        style={{ marginRight: 12 }}
                      />
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <div style={{ wordBreak: 'break-all', fontWeight: 500 }}>{script.name}</div>
                        <div style={{ fontSize: 11, color: token.colorTextSecondary }}>
                          {script.events?.length || 0} {t("recording.events")} · {script.resolution}
                          {script.deviceModel ? ` · ${script.deviceModel}` : null}
                        </div>
                      </div>
                    </div>
                    <Space size={4}>
                      <Tooltip title={t("recording.play")}>
                        <Button
                          type="text"
                          size="small"
                          icon={<CaretRightOutlined />}
                          onClick={(e) => {
                            e.stopPropagation();
                            handlePlayScript(script);
                          }}
                          disabled={isRecording || isPlaying || !selectedDevice}
                        />
                      </Tooltip>
                      <Tooltip title={t("common.rename")}>
                        <Button
                          type="text"
                          size="small"
                          icon={<EditOutlined />}
                          onClick={(e) => {
                            e.stopPropagation();
                            openRenameModal(script.name);
                          }}
                          disabled={isRecording || isPlaying}
                        />
                      </Tooltip>
                      <Tooltip title={t("automation.convert_to_workflow")}>
                        <Button
                          type="text"
                          size="small"
                          icon={<BranchesOutlined />}
                          onClick={(e) => {
                            e.stopPropagation();
                            handleConvertToWorkflow(script);
                          }}
                          disabled={isRecording || isPlaying}
                        />
                      </Tooltip>
                      <Popconfirm
                        title={t("recording.delete_confirm", { name: script.name })}
                        onConfirm={() => handleDeleteScript(script.name)}
                        okText={t("common.ok")}
                        cancelText={t("common.cancel")}
                      >
                        <Tooltip title={t("recording.delete")}>
                          <Button
                            type="text"
                            size="small"
                            danger
                            icon={<DeleteOutlined />}
                            onClick={(e) => e.stopPropagation()}
                          />
                        </Tooltip>
                      </Popconfirm>
                    </Space>
                  </div>
                )}
              />
            </Card>
          </div>
        </div>
      </div>

      {/* Save Script Modal */}
      <Modal
        title={t("recording.save_script")}
        open={saveModalVisible}
        onOk={handleSaveScript}
        onCancel={closeSaveModal}
        okText={t("common.ok")}
        cancelText={t("common.cancel")}
      >
        <Input
          placeholder={t("recording.script_name")}
          value={scriptName}
          onChange={(e) => setScriptName(e.target.value)}
          onPressEnter={handleSaveScript}
        />
      </Modal>

      {/* Rename Script Modal */}
      <Modal
        title={t("recording.rename_script")}
        open={renameModalVisible}
        onOk={handleRenameScript}
        onCancel={closeRenameModal}
        okText={t("common.ok")}
        cancelText={t("common.cancel")}
      >
        <Input
          placeholder={t("recording.new_script_name")}
          value={newScriptName}
          onChange={(e) => setNewScriptName(e.target.value)}
          onPressEnter={handleRenameScript}
        />
      </Modal>

      {/* Selector Choice Modal (Precise Recording Mode) */}
      <Modal
        title={
          <Space>
            <AimOutlined style={{ color: token.colorPrimary }} />
            {t("recording.choose_selector")}
          </Space>
        }
        open={isWaitingForSelector}
        onCancel={() => {
          // If closed, default to coordinates to let recording continue
          const coordSuggestion = pendingSelectorData?.suggestions?.find((s: any) => s.type === 'coordinates');
          if (coordSuggestion) {
            submitSelectorChoice(coordSuggestion.type, coordSuggestion.value);
          } else if (pendingSelectorData?.elementInfo) {
            // Fallback manually if suggestion list is missing for some reason
            const { x, y } = pendingSelectorData.elementInfo;
            submitSelectorChoice('coordinates', `${x},${y}`);
          }
        }}
        footer={
          <Button
            onClick={() => {
              const coordSuggestion = pendingSelectorData?.suggestions?.find((s: any) => s.type === 'coordinates');
              if (coordSuggestion) {
                submitSelectorChoice(coordSuggestion.type, coordSuggestion.value);
              }
            }}
          >
            {t("recording.use_coordinates_and_close")}
          </Button>
        }
        closable={true}
        maskClosable={false}
        width={600}
        centered
      >
        <Space direction="vertical" style={{ width: "100%" }} size="middle">
          <Alert
            message={t("recording.precise_mode_active")}
            description={t("recording.please_choose_selector")}
            type="info"
            showIcon
          />

          {pendingSelectorData?.elementInfo && (
            <Card size="small" title={t("recording.element_info")} className="element-info-card">
              <div style={{ display: 'grid', gridTemplateColumns: '80px 1fr', gap: '8px' }}>
                <span style={{ color: token.colorTextDescription }}>{t("recording.class")}:</span>
                <span style={{ fontFamily: 'monospace', fontSize: '12px' }}>{pendingSelectorData.elementInfo.class}</span>

                <span style={{ color: token.colorTextDescription }}>{t("recording.bounds")}:</span>
                <span style={{ fontFamily: 'monospace', fontSize: '12px' }}>{pendingSelectorData.elementInfo.bounds}</span>
              </div>
            </Card>
          )}

          <List
            header={<strong>{t("recording.suggestions")}</strong>}
            dataSource={pendingSelectorData?.suggestions || []}
            renderItem={(suggestion: any) => {
              const getPriorityColor = (p: number) => {
                if (p >= 5) return 'success';
                if (p >= 4) return 'processing';
                if (p >= 3) return 'warning';
                return 'default';
              };

              return (
                <List.Item
                  style={{
                    cursor: 'pointer',
                    borderRadius: 8,
                    marginBottom: 8,
                    border: `1px solid ${token.colorBorderSecondary}`,
                    transition: 'all 0.2s'
                  }}
                  className="selector-option-item"
                  onClick={() => submitSelectorChoice(suggestion.type, suggestion.value)}
                >
                  <List.Item.Meta
                    avatar={
                      <div style={{
                        width: 32,
                        height: 32,
                        borderRadius: '50%',
                        backgroundColor: token.colorPrimaryBg,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        color: token.colorPrimary
                      }}>
                        {suggestion.type === 'text' && <span>T</span>}
                        {suggestion.type === 'id' && <span>ID</span>}
                        {suggestion.type === 'desc' && <span>D</span>}
                        {suggestion.type === 'xpath' && <span>X</span>}
                        {suggestion.type === 'coordinates' && <span>XY</span>}
                        {suggestion.type === 'class' && <span>C</span>}
                      </div>
                    }
                    title={
                      <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                        <span style={{ fontWeight: 600 }}>{suggestion.type.toUpperCase()}</span>
                        <Tag color={getPriorityColor(suggestion.priority)}>
                          Priority: {suggestion.priority}
                        </Tag>
                      </div>
                    }
                    description={
                      <div>
                        <div style={{
                          fontFamily: 'monospace',
                          backgroundColor: token.colorFillAlter,
                          padding: '4px 8px',
                          borderRadius: 4,
                          marginBottom: 4,
                          wordBreak: 'break-all'
                        }}>
                          {suggestion.value}
                        </div>
                        <div style={{ fontSize: '12px', color: token.colorTextDescription }}>
                          {suggestion.description}
                        </div>
                      </div>
                    }
                  />
                </List.Item>
              );
            }}
          />
        </Space>
      </Modal>
    </div>
  );
};

export default RecordingView;
