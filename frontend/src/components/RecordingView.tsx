import React, { useState, useEffect, useRef } from "react";
import {
  Button,
  Space,
  Tag,
  Card,
  List,
  Modal,
  Input,
  message,
  theme,
  Progress,
  Empty,
  Popconfirm,
  Tooltip,
  Checkbox,
} from "antd";
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
} from "@ant-design/icons";
import DeviceSelector from "./DeviceSelector";
import { useDeviceStore, useAutomationStore, TouchScript } from "../stores";
import { main } from "../../wailsjs/go/models";

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
  } = useAutomationStore();

  const [selectedScriptNames, setSelectedScriptNames] = useState<string[]>([]);

  const { t } = useTranslation();
  const { token } = theme.useToken();

  const [saveModalVisible, setSaveModalVisible] = useState(false);
  const [scriptName, setScriptName] = useState("");

  const [renameModalVisible, setRenameModalVisible] = useState(false);
  const [editingScriptName, setEditingScriptName] = useState("");
  const [newScriptName, setNewScriptName] = useState("");

  const [selectedScript, setSelectedScript] = useState<TouchScript | null>(null);
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

  const handleStartRecording = async () => {
    if (!selectedDevice) {
      message.warning(t("app.select_device"));
      return;
    }
    try {
      await startRecording(selectedDevice);
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
      setSaveModalVisible(false);
      setScriptName("");
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
      await playScript(selectedDevice, script);
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
      setRenameModalVisible(false);
      setEditingScriptName("");
      setNewScriptName("");
    } catch (err) {
      message.error(String(err));
    }
  };

  const handleConvertToWorkflow = async (script: TouchScript) => {
    if (!script.events || script.events.length === 0) {
      message.warning(t("automation.no_events_to_convert"));
      return;
    }

    const steps: main.WorkflowStep[] = [];
    let lastTimestamp = 0;

    script.events.forEach((event, idx) => {
      // Add wait step if there is a significant delay
      const delay = event.timestamp - lastTimestamp;
      if (delay > 50) { // 50ms threshold
        steps.push({
          id: `step_wait_${Date.now()}_${idx}`,
          type: 'wait',
          value: String(delay),
          loop: 1,
          postDelay: 0,
        } as main.WorkflowStep);
      }
      lastTimestamp = event.timestamp;

      // Add Action Step
      const id = `step_action_${Date.now()}_${idx}`;
      if (event.type === 'tap') {
        steps.push({
          id,
          type: 'adb',
          name: `Tap ${idx + 1}`,
          value: `shell input tap ${event.x} ${event.y}`,
          loop: 1,
          postDelay: 0,
          onError: 'stop'
        } as main.WorkflowStep);
      } else if (event.type === 'swipe') {
        steps.push({
          id,
          type: 'adb',
          name: `Swipe ${idx + 1}`,
          value: `shell input swipe ${event.x} ${event.y} ${event.x2 || event.x} ${event.y2 || event.y} ${event.duration || 300}`,
          loop: 1,
          postDelay: 0,
          onError: 'stop'
        } as main.WorkflowStep);
      } else if (event.type === 'wait') {
        // Explicit wait event in script
        steps.push({
          id,
          type: 'wait',
          name: `Wait ${idx + 1}`,
          value: String(event.duration || 1000),
          loop: 1,
          postDelay: 0,
        } as main.WorkflowStep);
      }
    });

    const newWorkflow = new (main as any).Workflow({
      id: `wf_converted_${Date.now()}`,
      name: `${script.name} (Converted)`,
      description: `Converted from recording on ${new Date().toLocaleString()}`,
      steps: steps,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    });

    try {
      await (window as any).go.main.App.SaveWorkflow(newWorkflow);
      message.success(t("automation.converted_success"));
    } catch (err) {
      message.error(t("automation.convert_failed") + ": " + String(err));
    }
  };

  const formatDuration = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins.toString().padStart(2, "0")}:${secs.toString().padStart(2, "0")}`;
  };

  const formatEventDescription = (event: any, index: number) => {
    switch (event.type) {
      case "tap":
        return `${index + 1}. tap (${event.x}, ${event.y}) @ ${event.timestamp}ms`;
      case "swipe":
        return `${index + 1}. swipe (${event.x}, ${event.y}) → (${event.x2}, ${event.y2}) ${event.duration}ms @ ${event.timestamp}ms`;
      case "wait":
        return `${index + 1}. wait ${event.duration}ms`;
      default:
        return `${index + 1}. unknown`;
    }
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
          <div style={{ flex: 1, display: "flex", flexDirection: "column", gap: 16 }}>
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
                  <span>{t("recording.touch_on_device")}</span>
                  {isDeviceRecording ? (
                    <Button
                      type="primary"
                      danger
                      icon={<StopOutlined />}
                      onClick={handleStopRecording}
                    >
                      {t("recording.stop_record")}
                    </Button>
                  ) : (
                    <Button
                      type="primary"
                      icon={<PlayCircleOutlined />}
                      onClick={handleStartRecording}
                      disabled={!selectedDevice || isPlaying}
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
                        {formatDuration(recordingDuration)} · {recordedActionCount} {t("recording.events")}
                      </span>
                    </div>
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

            {/* Script Preview */}
            {(currentScript || selectedScript) && (
              <Card
                title={t("recording.script_preview")}
                size="small"
                style={{ maxHeight: 300, overflow: "auto" }}
              >
                <div style={{ fontFamily: "monospace", fontSize: 12 }}>
                  {((currentScript || selectedScript)?.events || []).map((event, idx) => (
                    <div
                      key={idx}
                      style={{
                        padding: "4px 0",
                        borderBottom: `1px solid ${token.colorBorderSecondary}`,
                        color:
                          isDevicePlaying && playbackProgress && idx < playbackProgress.current
                            ? token.colorSuccess
                            : token.colorText,
                      }}
                    >
                      {formatEventDescription(event, idx)}
                    </div>
                  ))}
                </div>
              </Card>
            )}
          </div>

          {/* Right Column - Script List */}
          <div style={{ flex: 1, display: "flex", flexDirection: "column" }}>
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
              bodyStyle={{ padding: 0, overflowY: "auto", maxHeight: "calc(100vh - 200px)" }}
            >
              {scripts.length === 0 ? (
                <Empty description={t("recording.no_scripts")} image={Empty.PRESENTED_IMAGE_SIMPLE} />
              ) : (
                <List
                  size="small"
                  dataSource={scripts}
                  renderItem={(script) => (
                    <List.Item
                      style={{
                        cursor: "pointer",
                        padding: "8px 16px",
                        backgroundColor:
                          selectedScript?.name === script.name ? token.colorPrimaryBg : undefined,
                      }}
                      onClick={() => {
                        setSelectedScript(script);
                        setCurrentScript(null);
                      }}
                      actions={[
                        <Tooltip title={t("recording.play")} key="play">
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
                        </Tooltip>,
                        <Tooltip title={t("common.rename")} key="rename">
                          <Button
                            type="text"
                            size="small"
                            icon={<EditOutlined />}
                            onClick={(e) => {
                              e.stopPropagation();
                              setEditingScriptName(script.name);
                              setNewScriptName(script.name);
                              setRenameModalVisible(true);
                            }}
                            disabled={isRecording || isPlaying}
                          />
                        </Tooltip>,
                        <Tooltip title={t("automation.convert_to_workflow")} key="convert">
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
                        </Tooltip>,
                        <Popconfirm
                          key="delete"
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
                        </Popconfirm>,
                      ]}
                    >
                      <div style={{ display: 'flex', alignItems: 'center', width: '100%' }}>
                        <Checkbox
                          checked={selectedScriptNames.includes(script.name)}
                          onChange={(e) => {
                            e.stopPropagation();
                            const name = script.name;
                            setSelectedScriptNames(prev =>
                              e.target.checked ? [...prev, name] : prev.filter(n => n !== name)
                            );
                          }}
                          style={{ marginRight: 12 }}
                        />
                        <List.Item.Meta
                          title={<span style={{ wordBreak: 'break-all' }}>{script.name}</span>}
                          description={
                            <span style={{ fontSize: 11, color: token.colorTextSecondary }}>
                              {script.events?.length || 0} {t("recording.events")} · {script.resolution}
                              {(script as any).deviceModel ? ` · ${(script as any).deviceModel}` : null}
                            </span>
                          }
                        />
                      </div>
                    </List.Item>
                  )}
                />
              )}
            </Card>
          </div>
        </div>
      </div>

      {/* Save Script Modal */}
      <Modal
        title={t("recording.save_script")}
        open={saveModalVisible}
        onOk={handleSaveScript}
        onCancel={() => {
          setSaveModalVisible(false);
          setScriptName("");
        }}
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
        onCancel={() => {
          setRenameModalVisible(false);
          setEditingScriptName("");
          setNewScriptName("");
        }}
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
    </div>
  );
};

export default RecordingView;
