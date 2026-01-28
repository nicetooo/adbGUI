import React, { useEffect, useRef } from "react";
import {
  Button,
  Space,
  Input,
  Radio,
  message,
  Modal,
  Tag,
  Dropdown,
  theme,
  Spin,
} from "antd";
import VirtualTable from "./VirtualTable";
import { useTranslation } from "react-i18next";
import { useTheme } from "../ThemeContext";
import {
  ReloadOutlined,
  PlayCircleOutlined,
  FileTextOutlined,
  DownloadOutlined,
  InfoCircleOutlined,
  StopOutlined,
  CloseCircleOutlined,
  CheckCircleOutlined,
  ClearOutlined,
  DeleteOutlined,
  MoreOutlined,
  SettingOutlined,
  CloudUploadOutlined,
  LoadingOutlined,
} from "@ant-design/icons";
import DeviceSelector from "./DeviceSelector";
import AppInfoModal from "./AppInfoModal";
import { useDeviceStore, useLogcatStore, useUIStore, VIEW_KEYS, useAppsStore } from "../stores";
// @ts-ignore
import {
  GetAppInfo,
  StartActivity,
  ListPackages,
  UninstallApp,
  ClearAppData,
  ForceStopApp,
  StartApp,
  EnableApp,
  DisableApp,
  ExportAPK,
  OpenSettings,
  InstallPackage,
} from "../../wailsjs/go/main/App";
// @ts-ignore
import { OnFileDrop, OnFileDropOff } from "../../wailsjs/runtime/runtime";
// @ts-ignore
import { main } from "../types/wails-models";

const AppsView: React.FC = () => {
  const { t } = useTranslation();
  const { token } = theme.useToken();
  const { isDark } = useTheme();
  const { selectedDevice } = useDeviceStore();
  const { setSelectedPackage, setLogFilter, toggleLogcat, stopLogcat, isLogging } = useLogcatStore();
  const { setSelectedKey } = useUIStore();
  const containerRef = useRef<HTMLDivElement>(null);

  // Use appsStore instead of useState
  const {
    tableHeight,
    packages,
    appsLoading,
    packageFilter,
    typeFilter,
    infoModalVisible,
    infoLoading,
    selectedAppInfo,
    permissionSearch,
    activitySearch,
    isDraggingOver,
    isInstalling,
    installingFileName,
    setTableHeight,
    setPackages,
    setAppsLoading,
    setPackageFilter,
    setTypeFilter,
    setInfoModalVisible,
    setInfoLoading,
    setSelectedAppInfo,
    setPermissionSearch,
    setActivitySearch,
    setIsDraggingOver,
    setIsInstalling,
  } = useAppsStore();

  const handleJumpToLogcat = async (pkg: string) => {
    setSelectedPackage(pkg);
    setLogFilter("");
    setSelectedKey(VIEW_KEYS.LOGCAT);

    if (isLogging) {
      stopLogcat();
      setTimeout(() => toggleLogcat(selectedDevice, pkg), 300);
    } else {
      setTimeout(() => toggleLogcat(selectedDevice, pkg), 100);
    }
  };

  // Handle package installation from drag and drop (APK, XAPK, AAB)
  const handleInstallPackage = async (packagePath: string) => {
    console.log("[AppsView] handleInstallPackage called with:", packagePath);
    console.log("[AppsView] selectedDevice:", selectedDevice);
    
    if (!selectedDevice) {
      message.error(t("apps.no_device_selected") || "Please select a device first");
      return;
    }

    const fileName = packagePath.split("/").pop() || packagePath.split("\\").pop() || packagePath;
    console.log("[AppsView] Starting install for:", fileName);
    setIsInstalling(true, fileName);

    try {
      console.log("[AppsView] Calling InstallPackage...");
      const result = await InstallPackage(selectedDevice, packagePath);
      console.log("[AppsView] InstallPackage result:", result);
      if (result.toLowerCase().includes("success")) {
        message.success(t("apps.install_success", { name: fileName }) || `Successfully installed ${fileName}`);
        // Refresh package list after installation (with delay for Android to register the app)
        setTimeout(() => {
          console.log("[AppsView] Refreshing packages after install, typeFilter:", typeFilter, "device:", selectedDevice);
          fetchPackages(typeFilter, selectedDevice);
        }, 1000);
      } else {
        message.error(t("apps.install_failed") + ": " + result);
      }
    } catch (err) {
      console.error("[AppsView] InstallPackage error:", err);
      message.error(t("apps.install_failed") + ": " + String(err));
    } finally {
      setIsInstalling(false);
    }
  };

  // Listen for file drop events
  useEffect(() => {
    const handleFileDrop = (_x: number, _y: number, paths: string[]) => {
      console.log("[AppsView] OnFileDrop triggered, paths:", paths);
      setIsDraggingOver(false);
      
      // Filter for installable package files (APK, XAPK, AAB)
      const packageFiles = paths.filter(path => {
        const lowerPath = path.toLowerCase();
        return lowerPath.endsWith(".apk") || lowerPath.endsWith(".xapk") || lowerPath.endsWith(".aab");
      });
      
      console.log("[AppsView] Filtered package files:", packageFiles);
      
      if (packageFiles.length === 0) {
        message.warning(t("apps.no_package_files") || "No installable files found (supported: .apk, .xapk, .aab)");
        return;
      }

      // Install package files sequentially
      const installSequentially = async () => {
        for (const packagePath of packageFiles) {
          console.log("[AppsView] Installing:", packagePath);
          await handleInstallPackage(packagePath);
        }
      };
      installSequentially();
    };

    // Register file drop handler with drop target
    console.log("[AppsView] Registering OnFileDrop handler");
    OnFileDrop(handleFileDrop, true);

    // Cleanup on unmount
    return () => {
      console.log("[AppsView] Unregistering OnFileDrop handler");
      OnFileDropOff();
    };
  }, [selectedDevice, typeFilter, t]);

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
      // Reset state immediately - OnFileDrop will handle the actual installation
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
  }, []);

  const fetchPackages = async (packageType?: string, deviceId?: string) => {
    const targetDevice = deviceId || selectedDevice;
    if (!targetDevice) return;
    const typeToFetch = packageType || typeFilter;
    setAppsLoading(true);
    try {
      const res = await ListPackages(targetDevice, typeToFetch);
      if (typeToFetch === "all") {
        setPackages(res || []);
      } else if (typeToFetch === "system") {
        const userPackages = packages.filter((p: main.AppPackage) => p.type === "user");
        setPackages([...userPackages, ...(res || [])]);
      } else {
        setPackages(res || []);
      }
    } catch (err) {
      message.error(t("app.list_packages_failed") + ": " + String(err));
    } finally {
      setAppsLoading(false);
    }
  };

  useEffect(() => {
    if (selectedDevice) {
      fetchPackages();
    }
  }, [selectedDevice]);

  const handleUninstall = async (packageName: string) => {
    try {
      await UninstallApp(selectedDevice, packageName);
      message.success(t("app.uninstall_success", { name: packageName }));
      fetchPackages(typeFilter, selectedDevice);
    } catch (err) {
      message.error(t("app.uninstall_failed") + ": " + String(err));
    }
  };

  const handleClearData = async (packageName: string) => {
    try {
      await ClearAppData(selectedDevice, packageName);
      message.success(t("app.clear_data_success", { name: packageName }));
    } catch (err) {
      message.error(t("app.clear_data_failed") + ": " + String(err));
    }
  };

  const handleForceStop = async (packageName: string) => {
    try {
      await ForceStopApp(selectedDevice, packageName);
      message.success(t("app.force_stop_success", { name: packageName }));
    } catch (err) {
      message.error(t("app.force_stop_failed") + ": " + String(err));
    }
  };

  const handleStartApp = async (packageName: string) => {
    const hide = message.loading(t("app.launching", { name: packageName }), 0);
    try {
      await ForceStopApp(selectedDevice, packageName);
      await StartApp(selectedDevice, packageName);
      message.success(t("app.start_app_success", { name: packageName }));
    } catch (err) {
      message.error(t("app.start_app_failed") + ": " + String(err));
    } finally {
      hide();
    }
  };

  const handleToggleState = async (packageName: string, currentState: string) => {
    try {
      if (currentState === "enabled") {
        await DisableApp(selectedDevice, packageName);
        message.success(t("app.disabled_success", { name: packageName }));
      } else {
        await EnableApp(selectedDevice, packageName);
        message.success(t("app.enabled_success", { name: packageName }));
      }
      fetchPackages(typeFilter, selectedDevice);
    } catch (err) {
      message.error(t("app.change_state_failed") + ": " + String(err));
    }
  };

  const handleExportAPK = async (packageName: string) => {
    const hideMessage = message.loading(t("app.exporting", { name: packageName }), 0);
    try {
      const res = await ExportAPK(selectedDevice, packageName);
      if (res) {
        message.success(t("app.export_success", { path: res }));
      }
    } catch (err) {
      message.error(t("app.export_failed") + ": " + String(err));
    } finally {
      hideMessage();
    }
  };

  const handleOpenSettings = async (deviceId: string, action: string = "", data: string = "") => {
    const hide = message.loading(t("app.opening_settings"), 0);
    try {
      await OpenSettings(deviceId, action, data);
      message.success(t("app.open_settings_success"));
    } catch (err) {
      message.error(t("app.open_settings_failed") + ": " + String(err));
    } finally {
      hide();
    }
  };

  const handleFetchAppInfo = async (packageName: string, force: boolean = false) => {
    if (!selectedDevice) return;

    if (!force) {
      setSelectedAppInfo(packages.find((p) => p.name === packageName) || null);
    }

    setPermissionSearch("");
    setInfoModalVisible(true);
    setInfoLoading(true);
    try {
      const res = await GetAppInfo(selectedDevice, packageName, force);
      setSelectedAppInfo(res);
    } catch (err) {
      message.error(t("app.fetch_app_info_failed") + ": " + String(err));
    } finally {
      setInfoLoading(false);
    }
  };

  const handleStartActivity = async (activityName: string) => {
    const hide = message.loading(t("app.launching", { name: activityName }), 0);
    try {
      await StartActivity(selectedDevice, activityName);
      message.success(t("app.start_activity_success"));
    } catch (err) {
      message.error(t("app.start_activity_failed") + ": " + String(err));
    } finally {
      hide();
    }
  };

  useEffect(() => {
    const updateHeight = () => {
      if (containerRef.current) {
        const offset = 160;
        const height = containerRef.current.clientHeight - offset;
        setTableHeight(height > 200 ? height : 400);
      }
    };

    updateHeight();
    const timer = setTimeout(updateHeight, 100);
    window.addEventListener("resize", updateHeight);
    return () => {
      window.removeEventListener("resize", updateHeight);
      clearTimeout(timer);
    };
  }, []);

  const filteredPackages = packages.filter((p) => {
    const matchesName = p.name
      .toLowerCase()
      .includes(packageFilter.toLowerCase());
    const matchesType = typeFilter === "all" || p.type === typeFilter;
    return matchesName && matchesType;
  });

  const appColumns = [
    {
      title: t("apps.title"),
      key: "app",
      render: (_: any, record: main.AppPackage) => {
        const firstLetter = (record.label || record.name)
          .charAt(0)
          .toUpperCase();
        const colors = [
          "#f56a00", "#7265e6", "#ffbf00", "#00a2ae", "#1890ff",
          "#52c41a", "#eb2f96", "#fadb14", "#fa541c", "#13c2c2",
        ];
        const color = colors[
          Math.abs(record.name.split("").reduce((a: number, b: string) => (a << 5) - a + b.charCodeAt(0), 0)) % colors.length
        ];

        return (
          <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
            <div
              style={{
                width: 36, height: 36, borderRadius: 8, backgroundColor: color,
                display: "flex", alignItems: "center", justifyContent: "center",
                position: "relative", overflow: "hidden", flexShrink: 0,
                fontSize: "18px", fontWeight: "bold", color: "#fff",
                boxShadow: "0 2px 4px rgba(0,0,0,0.1)",
              }}
            >
              <span style={{ position: "relative", zIndex: 1 }}>{firstLetter}</span>
            </div>
            <div style={{ display: "flex", flexDirection: "column", lineHeight: 1.1 }}>
              <span style={{ fontWeight: 600, fontSize: "14px", color: token.colorText }}>
                {record.label || record.name}
              </span>
              <span style={{ fontSize: "10px", color: token.colorTextSecondary, fontFamily: "monospace" }}>
                {record.name}
              </span>
            </div>
          </div>
        );
      },
    },
    {
      title: t("apps.type"),
      dataIndex: "type",
      key: "type",
      width: 100,
      render: (type: string) => (
        <Tag color={type === "system" ? "orange" : "blue"}>
          {type === "system" ? t("apps.system") : t("apps.user")}
        </Tag>
      ),
    },
    {
      title: t("apps.state"),
      dataIndex: "state",
      key: "state",
      width: 100,
      render: (state: string) => (
        <Tag color={state === "enabled" ? "green" : "red"}>
          {state === "enabled" ? t("apps.enable").toUpperCase() : t("apps.disable").toUpperCase()}
        </Tag>
      ),
    },
    {
      title: t("apps.action"),
      key: "action",
      width: 380,
      render: (_: any, record: main.AppPackage) => {
        return (
          <Space size={4}>
            <Button size="small" icon={<PlayCircleOutlined />} onClick={() => handleStartApp(record.name)} title={t("apps.launch_app")} />
            <Button
              size="small"
              icon={<FileTextOutlined />}
              onClick={() => handleJumpToLogcat(record.name)}
              title={t("menu.logcat")}
            />
            <Button size="small" icon={<CloseCircleOutlined />} onClick={() => handleForceStop(record.name)} title={t("apps.force_stop")} />
            <Button
              size="small"
              icon={record.state === "enabled" ? <StopOutlined /> : <CheckCircleOutlined />}
              onClick={() => handleToggleState(record.name, record.state)}
              title={record.state === "enabled" ? t("apps.disable") : t("apps.enable")}
            />
            <Button size="small" icon={<SettingOutlined />} onClick={() => handleOpenSettings(selectedDevice, "android.settings.APPLICATION_DETAILS_SETTINGS", `package:${record.name}`)} title={t("apps.app_settings")} />
            <Button size="small" icon={<DownloadOutlined />} onClick={() => handleExportAPK(record.name)} title={t("apps.export")} />
            <Button size="small" icon={<InfoCircleOutlined />} onClick={() => handleFetchAppInfo(record.name)} title={t("app_info.title")} />

            <Dropdown
              menu={{
                items: [
                  {
                    key: "clear",
                    icon: <ClearOutlined />,
                    label: t("apps.clear_data"),
                    danger: true,
                    onClick: () => {
                      Modal.confirm({
                        title: t("apps.clear_data_confirm_title"),
                        content: t("apps.clear_data_confirm_content", { name: record.name }),
                        okText: t("apps.clear_data"),
                        okType: "danger",
                        cancelText: t("common.cancel"),
                        onOk: () => handleClearData(record.name),
                      });
                    },
                  },
                  {
                    key: "uninstall",
                    icon: <DeleteOutlined />,
                    label: t("apps.uninstall"),
                    danger: true,
                    onClick: () => {
                      Modal.confirm({
                        title: t("apps.uninstall_confirm_title"),
                        content: t("apps.uninstall_confirm_content", { name: record.name }),
                        okText: t("apps.uninstall"),
                        okType: "danger",
                        cancelText: t("common.cancel"),
                        onOk: () => handleUninstall(record.name),
                      });
                    },
                  },
                ],
              }}
              trigger={["click"]}
            >
              <Button size="small" icon={<MoreOutlined />} />
            </Dropdown>
          </Space>
        );
      },
    },
  ];

  return (
    <div
      ref={containerRef}
      style={{
        display: "flex",
        flexDirection: "column",
        height: "100%",
        position: "relative",
        padding: "16px 24px",
        // @ts-ignore - Wails drop target CSS property
        "--wails-drop-target": "drop",
      } as React.CSSProperties}
    >
      {/* Drag and drop overlay */}
      {(isDraggingOver || isInstalling) && (
        <div
          style={{
            position: "absolute",
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            backgroundColor: isDark ? "rgba(0, 0, 0, 0.85)" : "rgba(255, 255, 255, 0.95)",
            zIndex: 1000,
            display: "flex",
            flexDirection: "column",
            alignItems: "center",
            justifyContent: "center",
            borderRadius: token.borderRadiusLG,
            border: isDraggingOver ? `3px dashed ${token.colorPrimary}` : "none",
            transition: "all 0.2s ease",
          }}
        >
          {isInstalling ? (
            <>
              <Spin indicator={<LoadingOutlined style={{ fontSize: 48, color: token.colorPrimary }} spin />} />
              <div style={{ marginTop: 16, fontSize: 18, color: token.colorText }}>
                {t("apps.installing") || "Installing..."}
              </div>
              <div style={{ marginTop: 8, fontSize: 14, color: token.colorTextSecondary }}>
                {installingFileName}
              </div>
            </>
          ) : (
            <>
              <CloudUploadOutlined style={{ fontSize: 64, color: token.colorPrimary }} />
              <div style={{ marginTop: 16, fontSize: 18, color: token.colorText }}>
                {t("apps.drop_package_here") || "Drop files here to install"}
              </div>
              <div style={{ marginTop: 8, fontSize: 14, color: token.colorTextSecondary }}>
                {t("apps.drop_package_hint") || "Supports .apk, .xapk, .aab files"}
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
        <h2 style={{ margin: 0 }}>{t("apps.title")}</h2>
        <DeviceSelector />
      </div>
      <Space style={{ marginBottom: 16, flexShrink: 0 }}>
        <Input
          placeholder={t("apps.filter_placeholder")}
          value={packageFilter}
          onChange={(e) => setPackageFilter(e.target.value)}
          style={{ width: 300 }}
        />
        <Radio.Group
          value={typeFilter}
          onChange={(e) => {
            const newType = e.target.value;
            setTypeFilter(newType);
            fetchPackages(newType, selectedDevice);
          }}
        >
          <Radio.Button value="all">{t("apps.all")}</Radio.Button>
          <Radio.Button value="user">{t("apps.user")}</Radio.Button>
          <Radio.Button value="system">{t("apps.system")}</Radio.Button>
        </Radio.Group>
        <Button
          icon={<ReloadOutlined />}
          onClick={() => fetchPackages()}
          loading={appsLoading}
        >
          {t("common.refresh") || "Refresh"}
        </Button>
      </Space>
      <div
        className="selectable"
        style={{
          flex: 1,
          overflow: "hidden",
          backgroundColor: token.colorBgContainer,
          borderRadius: token.borderRadiusLG,
          border: `1px solid ${token.colorBorderSecondary}`,
          display: "flex",
          flexDirection: "column",
          userSelect: "text",
          boxShadow: isDark ? "0 2px 8px rgba(0,0,0,0.2)" : "0 2px 8px rgba(0,0,0,0.05)",
        }}
      >
        <VirtualTable
          columns={appColumns}
          dataSource={filteredPackages}
          rowKey="name"
          loading={appsLoading}
          scroll={{ y: tableHeight }}
          style={{ flex: 1 }}
        />
      </div>
      <AppInfoModal
        open={infoModalVisible}
        onCancel={() => setInfoModalVisible(false)}
        selectedAppInfo={selectedAppInfo}
        infoLoading={infoLoading}
        handleFetchAppInfo={handleFetchAppInfo}
        permissionSearch={permissionSearch}
        setPermissionSearch={setPermissionSearch}
        activitySearch={activitySearch}
        setActivitySearch={setActivitySearch}
        handleStartActivity={handleStartActivity}
        getContainer={() => containerRef.current || document.body}
      />
    </div>
  );
};

export default AppsView;

