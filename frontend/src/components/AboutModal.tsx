import React, { useEffect, useState } from "react";
import { Modal, Typography, Divider, Space, Button, Spin, Tooltip } from "antd";
import {
  InfoCircleOutlined,
  GithubOutlined,
  MailOutlined,
  FolderOpenOutlined,
  DatabaseOutlined,
  VideoCameraOutlined,
  HddOutlined,
  ReloadOutlined,
  ToolOutlined,
  ApartmentOutlined,
  CodeOutlined,
  FileTextOutlined,
  ClearOutlined,
} from "@ant-design/icons";
import { useTranslation } from "react-i18next";

import { useUIStore } from "../stores";
import { GetStorageInfo, OpenDataDir } from "../../wailsjs/go/main/App";

const { Text } = Typography;

interface StorageInfo {
  dataDir: string;
  totalSize: number;
  dbSize: number;
  recordingSize: number;
  binSize: number;
  workflowSize: number;
  scriptSize: number;
  logSize: number;
  cacheSize: number;
  otherSize: number;
}

interface StorageCategory {
  key: keyof StorageInfo;
  labelKey: string;
  color: string;
  icon: React.ReactNode;
}

const STORAGE_CATEGORIES: StorageCategory[] = [
  { key: "dbSize", labelKey: "about.storage_database", color: "#1677ff", icon: <DatabaseOutlined /> },
  { key: "recordingSize", labelKey: "about.storage_recordings", color: "#52c41a", icon: <VideoCameraOutlined /> },
  { key: "binSize", labelKey: "about.storage_bin", color: "#722ed1", icon: <ToolOutlined /> },
  { key: "workflowSize", labelKey: "about.storage_workflows", color: "#13c2c2", icon: <ApartmentOutlined /> },
  { key: "scriptSize", labelKey: "about.storage_scripts", color: "#eb2f96", icon: <CodeOutlined /> },
  { key: "logSize", labelKey: "about.storage_logs", color: "#fa8c16", icon: <FileTextOutlined /> },
  { key: "cacheSize", labelKey: "about.storage_cache", color: "#a0d911", icon: <ClearOutlined /> },
  { key: "otherSize", labelKey: "about.storage_other", color: "#8c8c8c", icon: <HddOutlined /> },
];

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  const value = bytes / Math.pow(1024, i);
  return `${value.toFixed(i === 0 ? 0 : 1)} ${units[i]}`;
}

interface AboutModalProps {
  open: boolean;
  onCancel: () => void;
}

const AboutModal: React.FC<AboutModalProps> = ({ open, onCancel }) => {
  const { t } = useTranslation();
  const { appVersion } = useUIStore();
  const [storageInfo, setStorageInfo] = useState<StorageInfo | null>(null);
  const [storageLoading, setStorageLoading] = useState(false);

  // @ts-ignore
  const BrowserOpenURL = (window as any).runtime.BrowserOpenURL;

  const loadStorageInfo = async () => {
    setStorageLoading(true);
    try {
      const info = await GetStorageInfo();
      setStorageInfo(info as unknown as StorageInfo);
    } catch {
      // ignore
    } finally {
      setStorageLoading(false);
    }
  };

  useEffect(() => {
    if (open) {
      loadStorageInfo();
    }
  }, [open]);

  const StorageBar: React.FC<{ info: StorageInfo }> = ({ info }) => {
    const total = info.totalSize || 1;
    const visible = STORAGE_CATEGORIES.filter((c) => (info[c.key] as number) > 0);

    return (
      <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
        {/* Stacked bar */}
        <div
          style={{
            display: "flex",
            height: 8,
            borderRadius: 4,
            overflow: "hidden",
            background: "var(--ant-color-fill-tertiary, #f0f0f0)",
          }}
        >
          {visible.map((c) => {
            const size = info[c.key] as number;
            const pct = (size / total) * 100;
            return (
              <Tooltip key={c.key} title={`${t(c.labelKey)}: ${formatBytes(size)}`}>
                <div style={{ width: `${pct}%`, background: c.color, minWidth: 2 }} />
              </Tooltip>
            );
          })}
        </div>

        {/* Legend */}
        <div style={{ display: "flex", gap: 10, flexWrap: "wrap", fontSize: 12 }}>
          {visible.map((c) => (
            <Space key={c.key} size={4}>
              <span style={{ color: c.color, fontSize: 12 }}>{c.icon}</span>
              <Text type="secondary" style={{ fontSize: 12 }}>
                {t(c.labelKey)} {formatBytes(info[c.key] as number)}
              </Text>
            </Space>
          ))}
        </div>
      </div>
    );
  };

  return (
    <Modal
      open={open}
      onCancel={onCancel}
      footer={[
        <Button key="close" type="primary" onClick={onCancel}>
          {t("common.close")}
        </Button>,
      ]}
      width={480}
      centered
      styles={{ body: { maxHeight: "calc(80vh - 100px)", overflowY: "auto" } }}
      title={
        <Space>
          <InfoCircleOutlined style={{ color: "#1890ff" }} />
          <span>{t("about.title")}</span>
        </Space>
      }
    >
      <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
        <div style={{ display: "flex", justifyContent: "space-between" }}>
          <Text strong>{t("about.version")}:</Text>
          <Text>{appVersion || "..."}</Text>
        </div>

        <div style={{ display: "flex", justifyContent: "space-between" }}>
          <Text strong>{t("about.author")}:</Text>
          <Space>
            <Text>Nice</Text>
            <Button
              type="link"
              size="small"
              icon={<MailOutlined />}
              onClick={() => BrowserOpenURL && BrowserOpenURL("mailto:naserdin@protonmail.com")}
            />
          </Space>
        </div>

        <div style={{ display: "flex", justifyContent: "space-between" }}>
          <Text strong>{t("about.github")}:</Text>
          <Button
            type="link"
            size="small"
            icon={<GithubOutlined />}
            onClick={() => BrowserOpenURL && BrowserOpenURL("https://github.com/nicethings/Gaze")}
          >
            Gaze
          </Button>
        </div>

        <div style={{ display: "flex", justifyContent: "space-between" }}>
          <Text strong>{t("about.license")}:</Text>
          <Text>{t("about.mit_license")}</Text>
        </div>
      </div>

      <Divider />

      {/* Storage Info Section */}
      <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
          <Text strong>
            <HddOutlined style={{ marginRight: 6 }} />
            {t("about.storage")}
          </Text>
          <Space size={4}>
            {storageInfo && (
              <Text type="secondary" style={{ fontSize: 13 }}>
                {formatBytes(storageInfo.totalSize)}
              </Text>
            )}
            <Button
              type="text"
              size="small"
              icon={<ReloadOutlined spin={storageLoading} />}
              onClick={loadStorageInfo}
              disabled={storageLoading}
            />
          </Space>
        </div>

        {storageLoading && !storageInfo ? (
          <div style={{ textAlign: "center", padding: 12 }}>
            <Spin size="small" />
          </div>
        ) : storageInfo ? (
          <>
            <StorageBar info={storageInfo} />

            <div
              style={{
                display: "flex",
                alignItems: "center",
                gap: 6,
                marginTop: 4,
                padding: "6px 8px",
                borderRadius: 6,
                background: "var(--ant-color-fill-quaternary, #fafafa)",
                cursor: "pointer",
              }}
              onClick={() => {
                if (storageInfo.dataDir) {
                  OpenDataDir().catch(() => {});
                }
              }}
              title={storageInfo.dataDir}
            >
              <FolderOpenOutlined style={{ color: "var(--ant-color-text-secondary)", flexShrink: 0 }} />
              <Text
                type="secondary"
                style={{
                  fontSize: 12,
                  fontFamily: "monospace",
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                  whiteSpace: "nowrap",
                }}
              >
                {storageInfo.dataDir}
              </Text>
            </div>
          </>
        ) : null}
      </div>

    </Modal>
  );
};

export default AboutModal;
