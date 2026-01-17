import { Modal, Input, Checkbox, Tabs, Typography, Space } from "antd";
import { useEffect } from "react";
import { useTranslation } from "react-i18next";
import { useFeedbackStore } from "../stores/feedbackStore";

const { TextArea } = Input;
const { Text } = Typography;

interface FeedbackModalProps {
  open: boolean;
  onCancel: () => void;
  appVersion: string;
  deviceInfo: string;
}

const FeedbackModal = ({ open, onCancel, appVersion, deviceInfo }: FeedbackModalProps) => {
  const { t } = useTranslation();
  const {
    backendLogs,
    frontendLogs,
    selectedBackendLogs,
    selectedFrontendLogs,
    description,
    setDescription,
    toggleAllBackend,
    toggleAllFrontend,
    toggleBackendLog,
    toggleFrontendLog,
    fetchLogs,
  } = useFeedbackStore();

  useEffect(() => {
    if (open) {
      fetchLogs();
    }
  }, [open, fetchLogs]);

  const handleOk = () => {
    let feedbackText = `### Description\n${description || "(No description provided)"}\n\n`;
    feedbackText += `### Environment\n- App Version: ${appVersion}\n- Device: ${deviceInfo}\n- OS: ${navigator.platform}\n\n`;

    if (selectedBackendLogs.size > 0) {
      feedbackText += `### Backend Logs\n\`\`\`\n${Array.from(selectedBackendLogs)
        .sort((a, b) => a - b)
        .map(i => backendLogs[i])
        .join("\n")}\n\`\`\`\n\n`;
    }

    if (selectedFrontendLogs.size > 0) {
      feedbackText += `### Frontend Logs\n\`\`\`\n${Array.from(selectedFrontendLogs)
        .sort((a, b) => a - b)
        .map(i => frontendLogs[i])
        .join("\n")}\n\`\`\`\n\n`;
    }

    const body = encodeURIComponent(feedbackText);
    const url = `https://github.com/nicetooo/Gaze/issues/new?body=${body}`;
    // @ts-ignore
    window.runtime.BrowserOpenURL(url);
    onCancel();
  };

  return (
    <Modal
      title={t("app.feedback")}
      open={open}
      onCancel={onCancel}
      onOk={handleOk}
      width={800}
      okText={t("app.submit_to_github")}
      cancelText={t("common.cancel")}
      centered
      styles={{ body: { maxHeight: "calc(80vh - 100px)", overflowY: "auto" } }}
    >
      <Space orientation="vertical" style={{ width: "100%" }} size="middle">
        <div>
          <Text strong>{t("app.feedback_description")}</Text>
          <TextArea
            rows={4}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder={t("app.feedback_placeholder")}
            style={{ marginTop: 8 }}
          />
        </div>

        <Tabs
          defaultActiveKey="1"
          items={[
            {
              key: "1",
              label: `${t("app.backend_logs")} (${backendLogs.length})`,
              children: (
                <div style={{ maxHeight: 300, overflowY: "auto", border: "1px solid #f0f0f0", borderRadius: 4, padding: 8 }}>
                  <div style={{ marginBottom: 8, paddingBottom: 8, borderBottom: "1px solid #f0f0f0" }}>
                    <Checkbox
                      onChange={(e) => toggleAllBackend(e.target.checked)}
                      checked={selectedBackendLogs.size === backendLogs.length && backendLogs.length > 0}
                      indeterminate={selectedBackendLogs.size > 0 && selectedBackendLogs.size < backendLogs.length}
                    >
                      {t("app.select_all")}
                    </Checkbox>
                  </div>
                  <div style={{ display: "flex", flexDirection: "column", gap: "4px" }}>
                    {backendLogs.map((log, index) => (
                      <div key={index} style={{ display: "flex", alignItems: "flex-start", padding: "4px 0" }}>
                        <Checkbox
                          checked={selectedBackendLogs.has(index)}
                          onChange={(e) => toggleBackendLog(index, e.target.checked)}
                        >
                          <Text code style={{ fontSize: "12px", whiteSpace: "pre-wrap" }}>{log}</Text>
                        </Checkbox>
                      </div>
                    ))}
                  </div>
                </div>
              ),
            },
            {
              key: "2",
              label: `${t("app.frontend_logs")} (${frontendLogs.length})`,
              children: (
                <div style={{ maxHeight: 300, overflowY: "auto", border: "1px solid #f0f0f0", borderRadius: 4, padding: 8 }}>
                  <div style={{ marginBottom: 8, paddingBottom: 8, borderBottom: "1px solid #f0f0f0" }}>
                    <Checkbox
                      onChange={(e) => toggleAllFrontend(e.target.checked)}
                      checked={selectedFrontendLogs.size === frontendLogs.length && frontendLogs.length > 0}
                      indeterminate={selectedFrontendLogs.size > 0 && selectedFrontendLogs.size < frontendLogs.length}
                    >
                      {t("app.select_all")}
                    </Checkbox>
                  </div>
                  <div style={{ display: "flex", flexDirection: "column", gap: "4px" }}>
                    {frontendLogs.map((log, index) => (
                      <div key={index} style={{ display: "flex", alignItems: "flex-start", padding: "4px 0" }}>
                        <Checkbox
                          checked={selectedFrontendLogs.has(index)}
                          onChange={(e) => toggleFrontendLog(index, e.target.checked)}
                        >
                          <Text code style={{ fontSize: "12px", whiteSpace: "pre-wrap" }}>{log}</Text>
                        </Checkbox>
                      </div>
                    ))}
                  </div>
                </div>
              ),
            },
          ]}
        />
        <Text type="secondary" style={{ fontSize: "12px" }}>
          * {t("app.feedback_note")}
        </Text>
      </Space>
    </Modal>
  );
};

export default FeedbackModal;
