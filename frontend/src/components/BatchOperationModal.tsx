import React, { useState, useEffect } from "react";
import {
  Modal,
  Button,
  Space,
  Select,
  Input,
  Progress,
  List,
  Tag,
  message,
  Divider,
  Typography,
} from "antd";
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  LoadingOutlined,
  DeleteOutlined,
  ClearOutlined,
  StopOutlined,
  CodeOutlined,
  UploadOutlined,
  ReloadOutlined,
  CloudUploadOutlined,
} from "@ant-design/icons";
import { useTranslation } from "react-i18next";
import { useDeviceStore, Device, BatchOperation, BatchResult } from "../stores";

const { Text } = Typography;

interface BatchOperationModalProps {
  open: boolean;
  onClose: () => void;
  selectedDeviceIds: string[];
  devices: Device[];
}

type OperationType = BatchOperation['type'];

const BatchOperationModal: React.FC<BatchOperationModalProps> = ({
  open,
  onClose,
  selectedDeviceIds,
  devices,
}) => {
  const { t } = useTranslation();
  const {
    batchOperationInProgress,
    batchResults,
    executeBatchOperation,
    selectAPKForBatch,
    selectFileForBatch,
    subscribeToBatchEvents,
    clearSelection,
  } = useDeviceStore();

  const [operationType, setOperationType] = useState<OperationType>("shell");
  const [packageName, setPackageName] = useState("");
  const [apkPath, setApkPath] = useState("");
  const [command, setCommand] = useState("");
  const [localPath, setLocalPath] = useState("");
  const [remotePath, setRemotePath] = useState("/sdcard/");
  const [executed, setExecuted] = useState(false);

  // Subscribe to batch progress events
  useEffect(() => {
    if (open) {
      const unsubscribe = subscribeToBatchEvents();
      return () => unsubscribe();
    }
  }, [open, subscribeToBatchEvents]);

  // Reset state when modal opens
  useEffect(() => {
    if (open) {
      setExecuted(false);
    }
  }, [open]);

  const getDeviceLabel = (deviceId: string) => {
    const device = devices.find(d => d.id === deviceId);
    if (device) {
      return device.brand && device.model
        ? `${device.brand} ${device.model}`
        : device.model || device.id;
    }
    return deviceId;
  };

  const handleSelectAPK = async () => {
    try {
      const path = await selectAPKForBatch();
      if (path) {
        setApkPath(path);
      }
    } catch (err) {
      message.error(String(err));
    }
  };

  const handleSelectFile = async () => {
    try {
      const path = await selectFileForBatch();
      if (path) {
        setLocalPath(path);
      }
    } catch (err) {
      message.error(String(err));
    }
  };

  const handleExecute = async () => {
    // Validate operation-specific fields
    switch (operationType) {
      case "install":
        if (!apkPath) {
          message.error(t("batch.select_apk") || "Please select an APK file");
          return;
        }
        break;
      case "uninstall":
      case "clear":
      case "stop":
        if (!packageName) {
          message.error(t("batch.enter_package") || "Please enter a package name");
          return;
        }
        break;
      case "shell":
        if (!command) {
          message.error(t("batch.enter_command") || "Please enter a command");
          return;
        }
        break;
      case "push":
        if (!localPath || !remotePath) {
          message.error(t("batch.enter_paths") || "Please select a file and enter remote path");
          return;
        }
        break;
      case "reboot":
        // No additional fields needed
        break;
    }

    const op: BatchOperation = {
      type: operationType,
      deviceIds: selectedDeviceIds,
      packageName: packageName,
      apkPath: apkPath,
      command: command,
      localPath: localPath,
      remotePath: remotePath,
    };

    setExecuted(true);
    try {
      const result = await executeBatchOperation(op);
      if (result.failureCount === 0) {
        message.success(
          t("batch.all_success", { count: result.successCount }) ||
          `All ${result.successCount} operations completed successfully`
        );
      } else if (result.successCount === 0) {
        message.error(
          t("batch.all_failed", { count: result.failureCount }) ||
          `All ${result.failureCount} operations failed`
        );
      } else {
        message.warning(
          t("batch.partial_success", { success: result.successCount, failed: result.failureCount }) ||
          `${result.successCount} succeeded, ${result.failureCount} failed`
        );
      }
    } catch (err) {
      message.error(String(err));
    }
  };

  const handleClose = () => {
    if (executed && batchResults.length > 0) {
      clearSelection();
    }
    onClose();
  };

  const operationOptions = [
    { value: "shell", label: t("batch.shell") || "Shell Command", icon: <CodeOutlined /> },
    { value: "install", label: t("batch.install") || "Install APK", icon: <CloudUploadOutlined /> },
    { value: "uninstall", label: t("batch.uninstall") || "Uninstall App", icon: <DeleteOutlined /> },
    { value: "clear", label: t("batch.clear") || "Clear App Data", icon: <ClearOutlined /> },
    { value: "stop", label: t("batch.stop") || "Force Stop App", icon: <StopOutlined /> },
    { value: "push", label: t("batch.push") || "Push File", icon: <UploadOutlined /> },
    { value: "reboot", label: t("batch.reboot") || "Reboot", icon: <ReloadOutlined /> },
  ];

  const renderOperationInputs = () => {
    switch (operationType) {
      case "install":
        return (
          <Space.Compact style={{ width: "100%" }}>
            <Input
              value={apkPath}
              placeholder={t("batch.apk_path") || "APK file path"}
              readOnly
              style={{ flex: 1 }}
            />
            <Button onClick={handleSelectAPK}>
              {t("common.browse") || "Browse"}
            </Button>
          </Space.Compact>
        );
      case "uninstall":
      case "clear":
      case "stop":
        return (
          <Input
            value={packageName}
            onChange={e => setPackageName(e.target.value)}
            placeholder="com.example.app"
          />
        );
      case "shell":
        return (
          <Input
            value={command}
            onChange={e => setCommand(e.target.value)}
            placeholder="ls -la /sdcard/"
          />
        );
      case "push":
        return (
          <Space direction="vertical" style={{ width: "100%" }}>
            <Space.Compact style={{ width: "100%" }}>
              <Input
                value={localPath}
                placeholder={t("batch.local_path") || "Local file path"}
                readOnly
                style={{ flex: 1 }}
              />
              <Button onClick={handleSelectFile}>
                {t("common.browse") || "Browse"}
              </Button>
            </Space.Compact>
            <Input
              value={remotePath}
              onChange={e => setRemotePath(e.target.value)}
              placeholder="/sdcard/"
              addonBefore={t("batch.remote") || "Remote"}
            />
          </Space>
        );
      case "reboot":
        return (
          <Text type="secondary">
            {t("batch.reboot_warning") || "This will reboot all selected devices"}
          </Text>
        );
      default:
        return null;
    }
  };

  const successCount = batchResults.filter(r => r.success).length;
  const failureCount = batchResults.filter(r => !r.success).length;
  const progress = selectedDeviceIds.length > 0
    ? Math.round((batchResults.length / selectedDeviceIds.length) * 100)
    : 0;

  return (
    <Modal
      title={t("batch.title") || "Batch Operation"}
      open={open}
      onCancel={handleClose}
      width={600}
      footer={[
        <Button key="cancel" onClick={handleClose}>
          {t("common.close") || "Close"}
        </Button>,
        <Button
          key="execute"
          type="primary"
          onClick={handleExecute}
          loading={batchOperationInProgress}
          disabled={batchOperationInProgress || selectedDeviceIds.length === 0}
        >
          {t("batch.execute") || "Execute"}
        </Button>,
      ]}
    >
      <Space direction="vertical" style={{ width: "100%" }} size="middle">
        {/* Device count */}
        <div>
          <Text strong>
            {t("batch.target_devices", { count: selectedDeviceIds.length }) ||
              `Target: ${selectedDeviceIds.length} device(s)`}
          </Text>
        </div>

        {/* Operation type selector */}
        <Select
          value={operationType}
          onChange={setOperationType}
          style={{ width: "100%" }}
          disabled={batchOperationInProgress}
          options={operationOptions.map(op => ({
            value: op.value,
            label: (
              <Space>
                {op.icon}
                {op.label}
              </Space>
            ),
          }))}
        />

        {/* Operation-specific inputs */}
        {renderOperationInputs()}

        {/* Progress section */}
        {(batchOperationInProgress || batchResults.length > 0) && (
          <>
            <Divider style={{ margin: "12px 0" }} />

            <Progress
              percent={progress}
              status={batchOperationInProgress ? "active" : (failureCount > 0 ? "exception" : "success")}
              format={() => `${batchResults.length}/${selectedDeviceIds.length}`}
            />

            {batchResults.length > 0 && (
              <div style={{ marginBottom: 8 }}>
                <Space>
                  <Tag color="success" icon={<CheckCircleOutlined />}>
                    {successCount} {t("batch.succeeded") || "succeeded"}
                  </Tag>
                  {failureCount > 0 && (
                    <Tag color="error" icon={<CloseCircleOutlined />}>
                      {failureCount} {t("batch.failed") || "failed"}
                    </Tag>
                  )}
                </Space>
              </div>
            )}

            {/* Results list */}
            <List
              size="small"
              dataSource={batchResults}
              style={{ maxHeight: 200, overflow: "auto" }}
              renderItem={(result: BatchResult) => (
                <List.Item>
                  <Space style={{ width: "100%", justifyContent: "space-between" }}>
                    <Space>
                      {result.success ? (
                        <CheckCircleOutlined style={{ color: "#52c41a" }} />
                      ) : (
                        <CloseCircleOutlined style={{ color: "#ff4d4f" }} />
                      )}
                      <Text>{getDeviceLabel(result.deviceId)}</Text>
                    </Space>
                    {result.error && (
                      <Text type="danger" ellipsis style={{ maxWidth: 200 }}>
                        {result.error}
                      </Text>
                    )}
                    {result.success && result.output && (
                      <Text type="secondary" ellipsis style={{ maxWidth: 200 }}>
                        {result.output.substring(0, 50)}
                      </Text>
                    )}
                  </Space>
                </List.Item>
              )}
            />

            {/* Pending devices */}
            {batchOperationInProgress && batchResults.length < selectedDeviceIds.length && (
              <List
                size="small"
                dataSource={selectedDeviceIds.filter(
                  id => !batchResults.find(r => r.deviceId === id)
                )}
                style={{ maxHeight: 100, overflow: "auto" }}
                renderItem={(deviceId: string) => (
                  <List.Item>
                    <Space>
                      <LoadingOutlined style={{ color: "#1677ff" }} />
                      <Text type="secondary">{getDeviceLabel(deviceId)}</Text>
                    </Space>
                  </List.Item>
                )}
              />
            )}
          </>
        )}
      </Space>
    </Modal>
  );
};

export default BatchOperationModal;
