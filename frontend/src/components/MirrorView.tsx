import React from "react";
import { Button, Space, Tag, Card, Switch, Tooltip, Slider, Select } from "antd";
import {
  StopOutlined,
  DesktopOutlined,
  DownloadOutlined,
  PlayCircleOutlined,
  FileTextOutlined,
  SettingOutlined,
  MobileOutlined,
} from "@ant-design/icons";
import DeviceSelector from "./DeviceSelector";
// @ts-ignore
import { main } from "../wailsjs/go/models";

const { Option } = Select;

interface Device {
  id: string;
  state: string;
  model: string;
  brand: string;
}

interface MirrorViewProps {
  devices: Device[];
  selectedDevice: string;
  setSelectedDevice: (id: string) => void;
  fetchDevices: () => Promise<void>;
  loading: boolean;
  isMirroring: boolean;
  mirrorDuration: number;
  handleStartScrcpy: (deviceId: string) => Promise<void>;
  handleStopScrcpy: (deviceId: string) => Promise<void>;
  scrcpyConfig: main.ScrcpyConfig;
  setScrcpyConfig: (config: main.ScrcpyConfig) => void;
  updateScrcpyConfig: (newConfig: main.ScrcpyConfig) => Promise<void>;
  shouldRecord: boolean;
  setShouldRecord: (val: boolean) => void;
  isRecording: boolean;
  recordDuration: number;
  handleStartMidSessionRecord: () => Promise<void>;
  handleStopMidSessionRecord: () => Promise<void>;
}

const MirrorView: React.FC<MirrorViewProps> = ({
  devices,
  selectedDevice,
  setSelectedDevice,
  fetchDevices,
  loading,
  isMirroring,
  mirrorDuration,
  handleStartScrcpy,
  handleStopScrcpy,
  scrcpyConfig,
  setScrcpyConfig,
  updateScrcpyConfig,
  shouldRecord,
  setShouldRecord,
  isRecording,
  recordDuration,
  handleStartMidSessionRecord,
  handleStopMidSessionRecord,
}) => {
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
          <h2 style={{ margin: 0 }}>Mirror Screen</h2>
          <Tag color="blue">Powered by Scrcpy</Tag>
        </div>
        <Space>
          <DeviceSelector
            devices={devices}
            selectedDevice={selectedDevice}
            onDeviceChange={setSelectedDevice}
            onRefresh={fetchDevices}
            loading={loading}
          />
          {isMirroring ? (
            <Button
              type="primary"
              danger
              size="large"
              icon={<StopOutlined />}
              onClick={() => handleStopScrcpy(selectedDevice)}
              disabled={!selectedDevice}
              style={{ height: "40px", borderRadius: "8px" }}
            >
              Stop Mirroring
            </Button>
          ) : (
            <Button
              type="primary"
              size="large"
              icon={<DesktopOutlined />}
              onClick={() => handleStartScrcpy(selectedDevice)}
              disabled={!selectedDevice}
              style={{ height: "40px", borderRadius: "8px" }}
            >
              Start Mirroring
            </Button>
          )}
        </Space>
      </div>

      {isMirroring && (
        <div
          style={{
            marginBottom: 16,
            padding: "8px 16px",
            backgroundColor: "#e6f7ff",
            border: "1px solid #91d5ff",
            borderRadius: "8px",
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
          }}
        >
          <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
            <Tag color="processing">MIRRORING</Tag>
            <span style={{ fontWeight: "bold", fontFamily: "monospace" }}>
              {new Date(mirrorDuration * 1000).toISOString().substr(11, 8)}
            </span>
          </div>
          <div style={{ fontSize: "12px", color: "#1890ff" }}>
            Active mirror session
          </div>
        </div>
      )}

      <div style={{ flex: 1, overflowY: "auto", paddingRight: 8 }}>
        <div
          style={{
            display: "flex",
            gap: "16px",
            alignItems: "flex-start",
          }}
        >
          {/* Left Column */}
          <div
            style={{
              flex: 1,
              display: "flex",
              flexDirection: "column",
              gap: "16px",
            }}
          >
            {/* Recording */}
            <Card
              title={
                <Space>
                  <DownloadOutlined />
                  Recording
                </Space>
              }
              size="small"
              style={{
                border:
                  shouldRecord || isRecording ? "1px solid #ff4d4f" : undefined,
                backgroundColor:
                  shouldRecord || isRecording ? "#fff1f0" : undefined,
              }}
            >
              <div
                style={{
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                  marginBottom: isRecording ? 12 : 0,
                }}
              >
                <Space direction="vertical" size={0}>
                  <span
                    style={{
                      fontWeight:
                        shouldRecord || isRecording ? "bold" : "normal",
                    }}
                  >
                    Record Screen
                  </span>
                  <div style={{ fontSize: "11px", color: "#888" }}>
                    {isRecording
                      ? "Recording in progress..."
                      : shouldRecord
                      ? "Save dialog will show up"
                      : "Independent of other settings"}
                  </div>
                </Space>
                <Switch
                  size="small"
                  checked={shouldRecord || isRecording}
                  onChange={async (v) => {
                    setShouldRecord(v);
                    if (isMirroring) {
                      if (v) {
                        await handleStartMidSessionRecord();
                      } else {
                        await handleStopMidSessionRecord();
                      }
                    }
                  }}
                  style={{
                    backgroundColor:
                      shouldRecord || isRecording ? "#ff4d4f" : undefined,
                  }}
                />
              </div>

              {isRecording && (
                <div
                  style={{
                    padding: "8px",
                    backgroundColor: "rgba(255, 77, 79, 0.05)",
                    borderRadius: "4px",
                    border: "1px dashed #ffa39e",
                  }}
                >
                  <div
                    style={{
                      display: "flex",
                      justifyContent: "space-between",
                      alignItems: "center",
                      marginBottom: 4,
                    }}
                  >
                    <Tag
                      color="error"
                      icon={<div className="record-dot" />}
                      style={{ margin: 0 }}
                    >
                      RECORD
                    </Tag>
                    <span
                      style={{
                        fontWeight: "bold",
                        fontFamily: "monospace",
                        color: "#cf1322",
                      }}
                    >
                      {new Date(recordDuration * 1000).toISOString().substr(11, 8)}
                    </span>
                  </div>
                  {scrcpyConfig.recordPath && (
                    <div
                      style={{
                        fontSize: "10px",
                        color: "#888",
                        wordBreak: "break-all",
                        marginTop: 4,
                      }}
                    >
                      File: {scrcpyConfig.recordPath.split(/[\\/]/).pop()}
                    </div>
                  )}
                </div>
              )}
            </Card>

            {/* Video Settings */}
            <Card
              title={
                <Space>
                  <PlayCircleOutlined />
                  Video Settings
                </Space>
              }
              size="small"
              className="mirror-card"
            >
              <div style={{ marginBottom: 16 }}>
                <div
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                  }}
                >
                  <span>Max Resolution</span>
                  <Tag>{scrcpyConfig.maxSize || "Auto"}</Tag>
                </div>
                <Slider
                  min={0}
                  max={2560}
                  step={128}
                  value={scrcpyConfig.maxSize}
                  onChange={(v) =>
                    setScrcpyConfig({ ...scrcpyConfig, maxSize: v })
                  }
                  marks={{
                    0: "Auto",
                    1024: "1024",
                    1920: "1080",
                    2560: "2K",
                  }}
                />
              </div>

              <div style={{ marginBottom: 16, marginTop: 24 }}>
                <div
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                  }}
                >
                  <span>Bit Rate (Mbps)</span>
                  <Tag>{scrcpyConfig.bitRate}M</Tag>
                </div>
                <Slider
                  min={1}
                  max={64}
                  value={scrcpyConfig.bitRate}
                  onChange={(v) =>
                    setScrcpyConfig({ ...scrcpyConfig, bitRate: v })
                  }
                />
              </div>

              <div style={{ marginBottom: 16 }}>
                <div
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                  }}
                >
                  <span>Max FPS</span>
                  <Tag>{scrcpyConfig.maxFps}</Tag>
                </div>
                <Slider
                  min={15}
                  max={144}
                  value={scrcpyConfig.maxFps}
                  onChange={(v) =>
                    setScrcpyConfig({ ...scrcpyConfig, maxFps: v })
                  }
                />
              </div>

              <div
                style={{
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                }}
              >
                <span>Video Codec</span>
                <Select
                  size="small"
                  value={scrcpyConfig.videoCodec}
                  onChange={(v) =>
                    updateScrcpyConfig({ ...scrcpyConfig, videoCodec: v })
                  }
                  style={{ width: 100 }}
                >
                  <Option value="h264">H.264</Option>
                  <Option value="h265">H.265</Option>
                  <Option value="av1">AV1</Option>
                </Select>
              </div>
            </Card>

            {/* Audio Settings */}
            <Card
              title={
                <Space>
                  <FileTextOutlined />
                  Audio Settings
                </Space>
              }
              size="small"
            >
              <Space direction="vertical" style={{ width: "100%" }}>
                <div className="setting-item">
                  <span>Disable Audio</span>
                  <Switch
                    size="small"
                    checked={scrcpyConfig.noAudio}
                    onChange={(v) =>
                      updateScrcpyConfig({ ...scrcpyConfig, noAudio: v })
                    }
                  />
                </div>
                <div className="setting-item">
                  <span>Audio Codec</span>
                  <Select
                    size="small"
                    disabled={scrcpyConfig.noAudio}
                    value={scrcpyConfig.audioCodec}
                    onChange={(v) =>
                      updateScrcpyConfig({ ...scrcpyConfig, audioCodec: v })
                    }
                    style={{ width: 100 }}
                  >
                    <Option value="opus">Opus</Option>
                    <Option value="aac">AAC</Option>
                    <Option value="flac">FLAC</Option>
                    <Option value="raw">RAW</Option>
                  </Select>
                </div>
              </Space>
            </Card>
          </div>

          {/* Right Column */}
          <div
            style={{
              flex: 1,
              display: "flex",
              flexDirection: "column",
              gap: "16px",
            }}
          >
            {/* Window Options */}
            <Card
              title={
                <Space>
                  <SettingOutlined />
                  Window Options
                </Space>
              }
              size="small"
            >
              <Space direction="vertical" style={{ width: "100%" }}>
                <div className="setting-item">
                  <span>Always On Top</span>
                  <Switch
                    size="small"
                    checked={scrcpyConfig.alwaysOnTop}
                    onChange={(v) =>
                      updateScrcpyConfig({
                        ...scrcpyConfig,
                        alwaysOnTop: v,
                      })
                    }
                  />
                </div>
                <div className="setting-item">
                  <span>Fullscreen</span>
                  <Switch
                    size="small"
                    checked={scrcpyConfig.fullscreen}
                    onChange={(v) =>
                      updateScrcpyConfig({
                        ...scrcpyConfig,
                        fullscreen: v,
                      })
                    }
                  />
                </div>
                <div className="setting-item">
                  <span>Borderless</span>
                  <Switch
                    size="small"
                    checked={scrcpyConfig.windowBorderless}
                    onChange={(v) =>
                      updateScrcpyConfig({
                        ...scrcpyConfig,
                        windowBorderless: v,
                      })
                    }
                  />
                </div>
              </Space>
            </Card>

            {/* Control & Interaction */}
            <Card
              title={
                <Space>
                  <MobileOutlined />
                  Control & Interaction
                </Space>
              }
              size="small"
            >
              <Space direction="vertical" style={{ width: "100%" }}>
                <div className="setting-item">
                  <Tooltip title="Prevent device from sleeping">
                    <span>Stay Awake</span>
                  </Tooltip>
                  <Switch
                    size="small"
                    checked={scrcpyConfig.stayAwake}
                    onChange={(v) =>
                      updateScrcpyConfig({ ...scrcpyConfig, stayAwake: v })
                    }
                  />
                </div>
                <div className="setting-item">
                  <Tooltip title="Disable keyboard/mouse control">
                    <span>Read Only</span>
                  </Tooltip>
                  <Switch
                    size="small"
                    checked={scrcpyConfig.readOnly}
                    onChange={(v) =>
                      updateScrcpyConfig({ ...scrcpyConfig, readOnly: v })
                    }
                  />
                </div>
                <div className="setting-item">
                  <span>Show Touches</span>
                  <Switch
                    size="small"
                    checked={scrcpyConfig.showTouches}
                    onChange={(v) =>
                      updateScrcpyConfig({
                        ...scrcpyConfig,
                        showTouches: v,
                      })
                    }
                  />
                </div>
              </Space>
            </Card>

            {/* Power Management */}
            <Card
              title={
                <Space>
                  <StopOutlined />
                  Power Management
                </Space>
              }
              size="small"
            >
              <Space direction="vertical" style={{ width: "100%" }}>
                <div className="setting-item">
                  <Tooltip title="Turn off device screen while mirroring">
                    <span>Turn Screen Off</span>
                  </Tooltip>
                  <Switch
                    size="small"
                    checked={scrcpyConfig.turnScreenOff}
                    onChange={(v) =>
                      updateScrcpyConfig({
                        ...scrcpyConfig,
                        turnScreenOff: v,
                      })
                    }
                  />
                </div>
                <div className="setting-item">
                  <Tooltip title="Power off device when closing mirror">
                    <span>Power Off On Close</span>
                  </Tooltip>
                  <Switch
                    size="small"
                    checked={scrcpyConfig.powerOffOnClose}
                    onChange={(v) =>
                      updateScrcpyConfig({
                        ...scrcpyConfig,
                        powerOffOnClose: v,
                      })
                    }
                  />
                </div>
              </Space>
            </Card>
          </div>
        </div>
      </div>
    </div>
  );
};

export default MirrorView;

