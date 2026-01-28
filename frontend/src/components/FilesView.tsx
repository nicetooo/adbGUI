import React, { useRef, useEffect } from "react";
import {
  Button,
  Space,
  Tooltip,
  Input,
  Checkbox,
  Dropdown,
  App,
  theme,
  Progress,
  Spin,
} from "antd";
import { useTranslation } from "react-i18next";
import VirtualTable from "./VirtualTable";
import {
  FolderOutlined,
  FileOutlined,
  CopyOutlined,
  ScissorOutlined,
  MoreOutlined,
  ArrowLeftOutlined,
  FolderAddOutlined,
  SnippetsOutlined,
  PictureOutlined,
  PlaySquareOutlined,
  VerticalAlignBottomOutlined,
  CloudUploadOutlined,
} from "@ant-design/icons";
// @ts-ignore
import {
  GetThumbnail,
  ListFiles,
  DeleteFile,
  MoveFile,
  CopyFile,
  Mkdir,
  OpenFileOnHost,
  CancelOpenFile,
  DownloadFile,
  UploadFile,
} from "../../wailsjs/go/main/App";
import { OnFileDrop, OnFileDropOff } from "../../wailsjs/runtime/runtime";
import DeviceSelector from "./DeviceSelector";
import { useDeviceStore, useFilesStore, useThumbnailStore } from "../stores";

const NameWithThumbnail = ({
  deviceId,
  record,
  onClick,
}: {
  deviceId: string;
  record: any;
  onClick: () => void;
}) => {
  const { thumbnails, loadingPaths, setThumbnail, setLoading } = useThumbnailStore();
  const thumb = thumbnails[record.path] || null;
  const loading = loadingPaths.has(record.path);

  const name = record.name.toLowerCase();
  const isImage =
    name.endsWith(".jpg") ||
    name.endsWith(".jpeg") ||
    name.endsWith(".png") ||
    name.endsWith(".webp") ||
    name.endsWith(".gif");
  const isVideo =
    name.endsWith(".mp4") ||
    name.endsWith(".mkv") ||
    name.endsWith(".mov") ||
    name.endsWith(".avi");

  React.useEffect(() => {
    if ((isImage || isVideo) && deviceId && !thumb && !loading) {
      setLoading(record.path, true);
      GetThumbnail(deviceId, record.path, record.modTime)
        .then((res: string) => {
          setThumbnail(record.path, res);
        })
        .catch(() => {
          setLoading(record.path, false);
        });
    }
  }, [deviceId, record.path, record.modTime, thumb, loading, setThumbnail, setLoading]);

  const icon = thumb ? (
    <img
      src={thumb}
      style={{
        width: 24,
        height: 24,
        objectFit: "cover",
        borderRadius: 2,
        display: "block",
      }}
    />
  ) : record.isDir ? (
    <FolderOutlined style={{ color: "#1890ff" }} />
  ) : isImage ? (
    <PictureOutlined style={{ color: "#52c41a" }} />
  ) : isVideo ? (
    <PlaySquareOutlined style={{ color: "#eb2f96" }} />
  ) : (
    <FileOutlined />
  );

  const content = (
    <Space
      onClick={onClick}
      style={{
        cursor: "pointer",
        width: "100%",
      }}
    >
      {icon}
      <span>{record.name}</span>
    </Space>
  );

  if (thumb) {
    return (
      <Tooltip
        title={
          <img
            src={thumb}
            style={{
              maxWidth: 400,
              maxHeight: 400,
              display: "block",
              borderRadius: 4,
            }}
          />
        }
        mouseEnterDelay={0.3}
        overlayStyle={{ maxWidth: "none" }}
        overlayInnerStyle={{ padding: 0, backgroundColor: "transparent" }}
      >
        {content}
      </Tooltip>
    );
  }

  return content;
};

interface FilesViewProps {
  initialPath?: string;
}

const FocusInput = ({
  defaultValue,
  onChange,
  placeholder,
  selectAll,
}: any) => {
  const inputRef = useRef<any>(null);
  useEffect(() => {
    const timer = setTimeout(() => {
      inputRef.current?.focus();
      if (selectAll) {
        inputRef.current?.select();
      }
    }, 200);
    return () => clearTimeout(timer);
  }, [selectAll]);

  return (
    <Input
      ref={inputRef}
      defaultValue={defaultValue}
      onChange={onChange}
      placeholder={placeholder}
      style={{ marginTop: 16 }}
      onPressEnter={() => {
        const okBtn = document.querySelector(
          ".ant-modal-confirm-btns .ant-btn-primary"
        ) as HTMLButtonElement;
        if (okBtn) okBtn.click();
      }}
    />
  );
};

const FilesView: React.FC<FilesViewProps> = ({
  initialPath,
}) => {
  const { t } = useTranslation();
  const { token } = theme.useToken();
  const { modal, message } = App.useApp();
  const { selectedDevice } = useDeviceStore();

  // Use filesStore instead of useState
  const {
    currentPath,
    fileList,
    filesLoading,
    clipboard,
    showHiddenFiles,
    isDraggingOver,
    isUploading,
    uploadingFileName,
    uploadProgress,
    setCurrentPath,
    setFileList,
    setFilesLoading,
    setClipboard,
    setShowHiddenFiles,
    setIsDraggingOver,
    setIsUploading,
    setUploadingFileName,
    setUploadProgress,
  } = useFilesStore();

  // Only set initialPath on first mount if provided
  useEffect(() => {
    if (initialPath) {
      setCurrentPath(initialPath);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // Run only once on mount

  const fetchFiles = async (path: string) => {
    if (!selectedDevice) return;
    setFilesLoading(true);
    try {
      const res = await ListFiles(selectedDevice, path);
      setFileList(res || []);
      setCurrentPath(path);
    } catch (err) {
      message.error(t("app.list_files_failed") + ": " + String(err));
    } finally {
      setFilesLoading(false);
    }
  };

  useEffect(() => {
    if (selectedDevice) {
      fetchFiles(currentPath);
    }
  }, [selectedDevice]);

  // Listen for file drop events
  useEffect(() => {
    const handleFileDrop = async (_x: number, _y: number, paths: string[]) => {
      console.log("[FilesView] OnFileDrop triggered, paths:", paths);
      console.log("[FilesView] selectedDevice:", selectedDevice, "currentPath:", currentPath);
      setIsDraggingOver(false);
      
      if (paths.length === 0) {
        console.log("[FilesView] No paths provided");
        return;
      }

      if (!selectedDevice) {
        console.log("[FilesView] No device selected");
        message.error(t("apps.no_device_selected") || "Please select a device first");
        return;
      }

      // Use /sdcard as default if currentPath is root or empty (root is read-only)
      const isRootRedirect = !currentPath || currentPath === "/";
      const uploadDir = isRootRedirect ? "/sdcard" : currentPath;
      console.log("[FilesView] Upload directory:", uploadDir, "redirected:", isRootRedirect);

      // Notify user if redirecting to /sdcard
      if (isRootRedirect) {
        message.info(t("files.upload_redirect_sdcard"));
      }

      // Start upload
      setIsUploading(true);
      setUploadProgress({ current: 0, total: paths.length });

      let successCount = 0;
      let failedFiles: string[] = [];

      for (let i = 0; i < paths.length; i++) {
        const localPath = paths[i];
        const fileName = localPath.split("/").pop() || localPath.split("\\").pop() || localPath;
        setUploadingFileName(fileName);
        setUploadProgress({ current: i, total: paths.length });

        try {
          const remotePath = uploadDir + (uploadDir.endsWith("/") ? "" : "/") + fileName;
          console.log("[FilesView] Uploading", localPath, "to", remotePath);
          await UploadFile(selectedDevice, localPath, remotePath);
          console.log("[FilesView] Upload success:", fileName);
          successCount++;
        } catch (err) {
          console.error(`[FilesView] Failed to upload ${fileName}:`, err);
          failedFiles.push(fileName);
        }
      }

      setIsUploading(false);
      setUploadingFileName("");
      setUploadProgress(null);

      // Show result message
      if (failedFiles.length === 0) {
        message.success(t("files.upload_success", { count: successCount }));
      } else if (successCount > 0) {
        message.warning(t("files.upload_partial", { success: successCount, failed: failedFiles.length }));
      } else {
        message.error(t("files.upload_failed"));
      }

      // Refresh file list - navigate to upload directory if redirected
      if (isRootRedirect && successCount > 0) {
        fetchFiles(uploadDir);
      } else {
        fetchFiles(currentPath);
      }
    };

    // Register file drop handler
    console.log("[FilesView] Registering OnFileDrop handler, device:", selectedDevice);
    OnFileDrop(handleFileDrop, true);

    // Cleanup on unmount
    return () => {
      console.log("[FilesView] Unregistering OnFileDrop handler");
      OnFileDropOff();
    };
  }, [selectedDevice, currentPath, t, setIsDraggingOver, setIsUploading, setUploadingFileName, setUploadProgress]);

  // Handle drag visual feedback
  useEffect(() => {
    let dragCounter = 0;
    let dropTimeout: ReturnType<typeof setTimeout> | null = null;

    const handleDragEnter = (e: DragEvent) => {
      e.preventDefault();
      e.stopPropagation();
      dragCounter++;
      if (e.dataTransfer?.types.includes("Files")) {
        setIsDraggingOver(true);
        // Safety timeout: reset after 3 seconds if drop doesn't happen
        if (dropTimeout) clearTimeout(dropTimeout);
        dropTimeout = setTimeout(() => {
          setIsDraggingOver(false);
          dragCounter = 0;
        }, 3000);
      }
    };

    const handleDragLeave = (e: DragEvent) => {
      e.preventDefault();
      e.stopPropagation();
      dragCounter--;
      if (dragCounter <= 0) {
        dragCounter = 0;
        setIsDraggingOver(false);
        if (dropTimeout) {
          clearTimeout(dropTimeout);
          dropTimeout = null;
        }
      }
    };

    const handleDragOver = (e: DragEvent) => {
      e.preventDefault();
      e.stopPropagation();
    };

    const handleDrop = (e: DragEvent) => {
      e.preventDefault();
      e.stopPropagation();
      dragCounter = 0;
      // Reset state immediately - OnFileDrop will handle the actual upload
      setIsDraggingOver(false);
      if (dropTimeout) {
        clearTimeout(dropTimeout);
        dropTimeout = null;
      }
    };

    // Listen on document for better drop detection
    document.addEventListener("dragenter", handleDragEnter);
    document.addEventListener("dragleave", handleDragLeave);
    document.addEventListener("dragover", handleDragOver);
    document.addEventListener("drop", handleDrop);

    return () => {
      document.removeEventListener("dragenter", handleDragEnter);
      document.removeEventListener("dragleave", handleDragLeave);
      document.removeEventListener("dragover", handleDragOver);
      document.removeEventListener("drop", handleDrop);
      if (dropTimeout) clearTimeout(dropTimeout);
    };
  }, [setIsDraggingOver]);

  const handleFileAction = async (action: string, file: any) => {
    if (!selectedDevice) return;
    try {
      switch (action) {
        case "open":
          const openKey = `open_${file.path}`;
          message.loading({
            content: (
              <span>
                {t("app.launching", { name: file.name })}
                <Button
                  type="link"
                  size="small"
                  onClick={async () => {
                    await CancelOpenFile(file.path);
                    message.destroy(openKey);
                  }}
                >
                  {t("common.cancel")}
                </Button>
              </span>
            ),
            key: openKey,
            duration: 0,
          });
          try {
            await OpenFileOnHost(selectedDevice, file.path);
            message.destroy(openKey);
          } catch (err) {
            if (String(err).includes("cancelled")) {
              message.info(t("app.open_cancelled"), 2);
            } else {
              message.error(t("app.open_failed") + ": " + String(err), 3);
            }
            message.destroy(openKey);
          }
          break;
        case "download":
          try {
            const savePath = await DownloadFile(selectedDevice, file.path);
            if (savePath) {
              message.success(t("app.export_success", { path: savePath }));
            }
          } catch (err) {
            message.error(t("app.export_failed") + ": " + String(err));
          }
          break;
        case "delete":
          await DeleteFile(selectedDevice, file.path);
          message.success(t("app.delete_success", { name: file.name }));
          fetchFiles(currentPath);
          break;
        case "copy":
          setClipboard({ path: file.path, type: "copy" });
          message.info(t("app.copy_success", { name: file.name }));
          break;
        case "cut":
          setClipboard({ path: file.path, type: "cut" });
          message.info(t("app.cut_success", { name: file.name }));
          break;
        case "paste":
          if (!clipboard) return;
          const dest =
            currentPath +
            (currentPath.endsWith("/") ? "" : "/") +
            clipboard.path.split("/").pop();
          if (clipboard.type === "copy") {
            await CopyFile(selectedDevice, clipboard.path, dest);
            message.success(
              t("app.paste_success", { name: clipboard.path.split("/").pop() })
            );
          } else {
            await MoveFile(selectedDevice, clipboard.path, dest);
            message.success(
              t("app.move_success", { name: clipboard.path.split("/").pop() })
            );
            setClipboard(null);
          }
          fetchFiles(currentPath);
          break;
        case "rename":
          let newName = file.name;
          modal.confirm({
            title: t("common.rename"),
            okText: t("common.ok"),
            cancelText: t("common.cancel"),
            autoFocusButton: null,
            content: (
              <FocusInput
                defaultValue={file.name}
                onChange={(e: any) => (newName = e.target.value)}
                placeholder={t("common.enter_new_name")}
                selectAll={true}
              />
            ),
            onOk: async () => {
              if (newName && newName !== file.name) {
                try {
                  const parentPath = currentPath.endsWith("/")
                    ? currentPath
                    : currentPath + "/";
                  const newPath = parentPath + newName;
                  await MoveFile(selectedDevice, file.path, newPath);
                  message.success(t("app.rename_success", { name: newName }));
                  fetchFiles(currentPath);
                } catch (err) {
                  message.error(t("app.rename_failed") + ": " + String(err));
                  throw err;
                }
              }
            },
          });
          break;
        case "mkdir":
          let folderName = "";
          modal.confirm({
            title: t("common.new_directory"),
            okText: t("common.ok"),
            cancelText: t("common.cancel"),
            autoFocusButton: null,
            content: (
              <FocusInput
                onChange={(e: any) => (folderName = e.target.value)}
                placeholder={t("common.folder_name")}
              />
            ),
            onOk: async () => {
              if (folderName) {
                try {
                  const parentPath = currentPath.endsWith("/")
                    ? currentPath
                    : currentPath + "/";
                  const newPath = parentPath + folderName;
                  await Mkdir(selectedDevice, newPath);
                  message.success(t("app.mkdir_success", { name: folderName }));
                  fetchFiles(currentPath);
                } catch (err) {
                  message.error(t("app.mkdir_failed") + ": " + String(err));
                  throw err;
                }
              }
            },
          });
          break;
      }
    } catch (err) {
      message.error(t("app.command_failed") + ": " + String(err));
    }
  };

  const fileColumns = [
    {
      title: t("files.name"),
      key: "name",
      sorter: (a: any, b: any) => a.name.localeCompare(b.name),
      render: (_: any, record: any) => (
        <NameWithThumbnail
          deviceId={selectedDevice}
          record={record}
          onClick={() => {
            if (record.isDir) {
              fetchFiles(record.path);
            } else {
              handleFileAction("open", record);
            }
          }}
        />
      ),
    },
    {
      title: t("files.size"),
      dataIndex: "size",
      key: "size",
      width: 100,
      sorter: (a: any, b: any) => a.size - b.size,
      render: (size: number, record: any) =>
        record.isDir
          ? "-"
          : size > 1024 * 1024
            ? (size / (1024 * 1024)).toFixed(2) + " MB"
            : (size / 1024).toFixed(2) + " KB",
    },
    {
      title: t("files.mode"),
      dataIndex: "mode",
      key: "mode",
      width: 110,
      render: (mode: string) => (
        <span style={{ fontFamily: "monospace", fontSize: 12 }}>{mode || "-"}</span>
      ),
    },
    {
      title: t("files.time"),
      dataIndex: "modTime",
      key: "modTime",
      width: 180,
      defaultSortOrder: "descend" as const,
      sorter: (a: any, b: any) => {
        if (a.modTime === "N/A" && b.modTime === "N/A") return 0;
        if (a.modTime === "N/A") return -1;
        if (b.modTime === "N/A") return 1;
        return a.modTime.localeCompare(b.modTime);
      },
    },
    {
      title: t("files.action"),
      key: "action",
      width: 180,
      render: (_: any, record: any) => (
        <Space>
          {!record.isDir && (
            <Tooltip title={t("common.download")}>
              <Button
                size="small"
                icon={<VerticalAlignBottomOutlined />}
                onClick={() => handleFileAction("download", record)}
              />
            </Tooltip>
          )}
          <Tooltip title={t("common.copy")}>
            <Button
              size="small"
              icon={<CopyOutlined />}
              onClick={() => handleFileAction("copy", record)}
            />
          </Tooltip>
          <Tooltip title={t("common.cut")}>
            <Button
              size="small"
              icon={<ScissorOutlined />}
              onClick={() => handleFileAction("cut", record)}
            />
          </Tooltip>
          <Dropdown
            menu={{
              items: [
                {
                  key: "open",
                  label: t("files.open_on_host"),
                  disabled: record.isDir,
                  onClick: () => handleFileAction("open", record),
                },
                {
                  key: "rename",
                  label: t("files.rename"),
                  onClick: () => handleFileAction("rename", record),
                },
                {
                  key: "delete",
                  label: t("files.delete"),
                  danger: true,
                  onClick: () => {
                    modal.confirm({
                      title: t("files.delete_confirm_title"),
                      okText: t("common.delete"),
                      okType: "danger",
                      cancelText: t("common.cancel"),
                      content: t("files.delete_confirm_content", {
                        name: record.name,
                      }),
                      onOk: () => handleFileAction("delete", record),
                    });
                  },
                },
              ],
            }}
          >
            <Button size="small" icon={<MoreOutlined />} />
          </Dropdown>
        </Space>
      ),
    },
  ];

  const filteredFiles = fileList.filter(
    (f) => showHiddenFiles || !f.name.startsWith(".")
  );

  return (
    <div
      style={{
        padding: "16px 24px",
        height: "100%",
        display: "flex",
        flexDirection: "column",
        overflow: "hidden",
        position: "relative",
        "--wails-drop-target": "drop",
      } as React.CSSProperties}
    >
      {/* Drag overlay */}
      {(isDraggingOver || isUploading) && (
        <div
          style={{
            position: "absolute",
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            backgroundColor: isUploading ? "rgba(0, 0, 0, 0.7)" : "rgba(22, 119, 255, 0.15)",
            border: isUploading ? "none" : `3px dashed ${token.colorPrimary}`,
            borderRadius: 8,
            display: "flex",
            flexDirection: "column",
            alignItems: "center",
            justifyContent: "center",
            zIndex: 1000,
            pointerEvents: isUploading ? "auto" : "none",
          }}
        >
          {isUploading ? (
            <>
              <Spin size="large" />
              <div style={{ marginTop: 16, color: "#fff", fontSize: 16 }}>
                {t("files.uploading", { name: uploadingFileName })}
              </div>
              {uploadProgress && (
                <div style={{ width: 200, marginTop: 12 }}>
                  <Progress
                    percent={Math.round(((uploadProgress.current + 1) / uploadProgress.total) * 100)}
                    size="small"
                    status="active"
                    format={() => `${uploadProgress.current + 1}/${uploadProgress.total}`}
                  />
                </div>
              )}
            </>
          ) : (
            <>
              <CloudUploadOutlined style={{ fontSize: 64, color: token.colorPrimary }} />
              <div style={{ marginTop: 16, color: token.colorPrimary, fontSize: 18, fontWeight: 500 }}>
                {t("files.drop_files_here")}
              </div>
              <div style={{ marginTop: 8, color: token.colorTextSecondary, fontSize: 14 }}>
                {t("files.drop_files_hint", { path: currentPath })}
              </div>
            </>
          )}
        </div>
      )}
      <div
        style={{
          marginBottom: 16,
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          flexShrink: 0,
        }}
      >
        <h2 style={{ margin: 0, color: token.colorText }}>{t("files.title")}</h2>
        <DeviceSelector />
      </div>

      <div
        style={{
          marginBottom: 12,
          display: "flex",
          gap: 16,
          alignItems: "center",
          flexShrink: 0,
        }}
      >
        <div style={{ flex: 1, display: "flex", gap: 8 }}>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => {
              // Navigate to parent directory
              const normalizedPath = currentPath.replace(/\/+$/, ""); // Remove trailing slashes
              const parts = normalizedPath.split("/").filter(Boolean);
              parts.pop();
              const parentPath = parts.length > 0 ? "/" + parts.join("/") : "/";
              console.log("[FilesView] Navigating back from", currentPath, "to", parentPath);
              fetchFiles(parentPath);
            }}
            disabled={!currentPath || currentPath === "/" || currentPath === ""}
          />
          <Input
            value={currentPath}
            onChange={(e) => setCurrentPath(e.target.value)}
            onPressEnter={() => fetchFiles(currentPath)}
            className="selectable"
            style={{ flex: 1, userSelect: "text" }}
          />
        </div>
        <Space>
          <Checkbox
            checked={showHiddenFiles}
            onChange={(e: any) => setShowHiddenFiles(e.target.checked)}
          >
            {t("files.show_hidden")}
          </Checkbox>
          <Button
            icon={<FolderAddOutlined />}
            onClick={() => handleFileAction("mkdir", null)}
          >
            {t("files.new_folder")}
          </Button>
          <Button
            icon={<SnippetsOutlined />}
            disabled={!clipboard}
            onClick={() => handleFileAction("paste", null)}
            type={clipboard ? "primary" : "default"}
          >
            {t("files.paste")}{" "}
            {clipboard &&
              `(${clipboard.type === "copy" ? t("common.copy") : t("common.cut")
              })`}
          </Button>
        </Space>
      </div>

      <div
        className="selectable"
        style={{
          flex: 1,
          overflow: "hidden",
          backgroundColor: token.colorBgContainer,
          borderRadius: "4px",
          border: `1px solid ${token.colorBorderSecondary}`,
          display: "flex",
          flexDirection: "column",
          userSelect: "text",
        }}
      >
        <VirtualTable
          columns={fileColumns}
          dataSource={filteredFiles}
          rowKey="path"
          loading={filesLoading}
          scroll={{ y: "calc(100vh - 170px)" }}
          onRow={(record: any) => ({
            onDoubleClick: () => {
              if (record.isDir) {
                fetchFiles(record.path);
              } else {
                handleFileAction("open", record);
              }
            },
          })}
          style={{ flex: 1 }}
        />
      </div>
    </div>
  );
};

export { FocusInput };
export default FilesView;

