import { useEffect } from "react";
import {
  Layout,
  Menu,
  Button,
  message,
  notification,
  Dropdown,
} from "antd";
import { useTranslation } from "react-i18next";
import LogcatView from "./components/LogcatView";
import DevicesView from "./components/DevicesView";
import AppsView from "./components/AppsView";
import FilesView from "./components/FilesView";
import ShellView from "./components/ShellView";
import MirrorView from "./components/MirrorView";
import ProxyView from "./components/ProxyView";
import DeviceInfoModal from "./components/DeviceInfoModal";
import AboutModal from "./components/AboutModal";
import WirelessConnectModal from "./components/WirelessConnectModal";
import FeedbackModal from "./components/FeedbackModal";
import {
  MobileOutlined,
  AppstoreOutlined,
  CodeOutlined,
  FileTextOutlined,
  DesktopOutlined,
  FolderOutlined,
  GithubOutlined,
  BugOutlined,
  InfoCircleOutlined,
  TranslationOutlined,
  GlobalOutlined,
  SunOutlined,
  MoonOutlined,
} from "@ant-design/icons";
import "./App.css";
import { useTheme } from "./ThemeContext";
import {
  useDeviceStore,
  useMirrorStore,
  useLogcatStore,
  useUIStore,
  VIEW_KEYS,
  openRecordPath,
} from "./stores";
// @ts-ignore
import { OpenPath } from "../wailsjs/go/main/App";

// @ts-ignore
const BrowserOpenURL = (window as any).runtime.BrowserOpenURL;
// @ts-ignore
const EventsOn = (window as any).runtime.EventsOn;
// @ts-ignore
const EventsOff = (window as any).runtime.EventsOff;

const { Content, Sider } = Layout;

function App() {
  const { mode, setMode, isDark } = useTheme();
  const { t, i18n } = useTranslation();

  // Device store
  const {
    devices,
    historyDevices,
    selectedDevice,
    loading,
    busyDevices,
    deviceInfoVisible,
    deviceInfoLoading,
    selectedDeviceInfo,
    fetchDevices,
    setSelectedDevice,
    handleFetchDeviceInfo,
    closeDeviceInfo,
    handleAdbConnect,
    handleAdbPair,
    handleSwitchToWireless,
    handleAdbDisconnect,
    handleRemoveHistoryDevice,
    handleOpenSettings,
    handleTogglePin,
  } = useDeviceStore();

  // Mirror store
  const { mirrorStatuses, recordStatuses, subscribeToEvents: subscribeMirrorEvents, updateDurations } = useMirrorStore();

  // Logcat store
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
    setLogs,
    setSelectedPackage,
    setLogFilter,
    setUseRegex,
    setPreFilter,
    setPreUseRegex,
    setExcludeFilter,
    setExcludeUseRegex,
    toggleLogcat,
    stopLogcat,
  } = useLogcatStore();

  // UI store
  const {
    selectedKey,
    aboutVisible,
    wirelessConnectVisible,
    feedbackVisible,
    appVersion,
    setSelectedKey,
    showAbout,
    hideAbout,
    showWirelessConnect,
    hideWirelessConnect,
    showFeedback,
    hideFeedback,
    init: initUI,
    subscribeToEvents: subscribeUIEvents,
  } = useUIStore();

  // Wrap store actions with message feedback
  const handleAdbConnectWithFeedback = async (address: string) => {
    try {
      await handleAdbConnect(address);
    } catch (err) {
      message.error(t("app.connect_failed") + ": " + String(err));
      throw err;
    }
  };

  const handleAdbPairWithFeedback = async (address: string, code: string) => {
    try {
      await handleAdbPair(address, code);
      message.success(t("app.pairing_success"));
    } catch (err) {
      message.error(t("app.pairing_failed") + ": " + String(err));
      throw err;
    }
  };

  const handleSwitchToWirelessWithFeedback = async (deviceId: string) => {
    const hide = message.loading(t("app.switching_to_wireless"), 0);
    try {
      await handleSwitchToWireless(deviceId);
      message.success(t("app.switch_success"));
    } catch (err) {
      message.error(t("app.switch_failed") + ": " + String(err));
    } finally {
      hide();
    }
  };

  const handleAdbDisconnectWithFeedback = async (deviceId: string) => {
    try {
      await handleAdbDisconnect(deviceId);
      message.success(t("app.disconnect_success"));
    } catch (err) {
      message.error(t("app.disconnect_failed") + ": " + String(err));
    }
  };

  const handleRemoveHistoryDeviceWithFeedback = async (deviceId: string) => {
    try {
      await handleRemoveHistoryDevice(deviceId);
      message.success(t("app.remove_success"));
    } catch (err) {
      message.error(t("app.remove_failed") + ": " + String(err));
    }
  };

  const handleOpenSettingsWithFeedback = async (deviceId: string, action?: string, data?: string) => {
    const hide = message.loading(t("app.opening_settings"), 0);
    try {
      await handleOpenSettings(deviceId, action, data);
      message.success(t("app.open_settings_success"));
    } catch (err) {
      message.error(t("app.open_settings_failed") + ": " + String(err));
    } finally {
      hide();
    }
  };

  const handleTogglePinWithFeedback = async (serial: string) => {
    try {
      await handleTogglePin(serial);
    } catch (err) {
      message.error(String(err));
    }
  };

  const handleFetchDeviceInfoWithFeedback = async (deviceId: string) => {
    try {
      await handleFetchDeviceInfo(deviceId);
    } catch (err) {
      message.error(t("app.fetch_device_info_failed") + ": " + String(err));
    }
  };

  const fetchDevicesWithFeedback = async (silent: boolean = false) => {
    try {
      await fetchDevices(silent);
    } catch (err) {
      if (!silent) {
        message.error(t("app.fetch_devices_failed") + ": " + String(err));
      }
    }
  };

  const toggleLogcatWithFeedback = async (pkg: string) => {
    if (!selectedDevice) {
      message.error(t("app.no_device_selected"));
      return;
    }
    try {
      await toggleLogcat(selectedDevice, pkg);
    } catch (err) {
      message.error(t("app.logcat_failed") + ": " + String(err));
    }
  };

  const handleJumpToLogcat = async (pkg: string) => {
    setSelectedPackage(pkg);
    setLogFilter("");
    setSelectedKey(VIEW_KEYS.LOGCAT);

    if (isLogging) {
      stopLogcat();
      setTimeout(() => toggleLogcatWithFeedback(pkg), 300);
    } else {
      setTimeout(() => toggleLogcatWithFeedback(pkg), 100);
    }
  };

  // Initialize stores and subscribe to events
  useEffect(() => {
    fetchDevicesWithFeedback();
    initUI();

    // Mirror events subscription with notification
    const unsubMirror = subscribeMirrorEvents((deviceId, path) => {
      notification.success({
        message: t("app.recording_saved"),
        description: t("app.recording_saved_desc"),
        btn: (
          <Button
            type="primary"
            size="small"
            onClick={() => {
              if (path) {
                OpenPath(path);
              }
              notification.destroy();
            }}
          >
            {t("app.show_in_folder")}
          </Button>
        ),
        key: "scrcpy-record-saved-" + deviceId,
        duration: 5,
      });
    });

    // UI events subscription
    const unsubUI = subscribeUIEvents((deviceId) => {
      setSelectedDevice(deviceId);
    });

    return () => {
      unsubMirror();
      unsubUI();
      stopLogcat();
    };
  }, []);

  // Screenshot progress listener
  useEffect(() => {
    const msgKey = "screenshot-msg";
    const unregister = EventsOn("screenshot-progress", (stepKey: string, data?: any) => {
      switch (stepKey) {
        case "screenshot_success":
          message.success({ content: t("app.screenshot_success", { path: data }), key: msgKey });
          break;
        case "screenshot_error":
          message.error({ content: t("app.command_failed") + ": " + String(data), key: msgKey });
          break;
        case "screenshot_off":
          message.warning({ content: t("app.screenshot_off"), key: msgKey });
          break;
        default:
          message.loading({ content: t(`app.${stepKey}`), key: msgKey, duration: 0 });
      }
    });

    return () => unregister();
  }, [t]);

  // Duration update timer
  useEffect(() => {
    const timer = setInterval(updateDurations, 1000);
    return () => clearInterval(timer);
  }, [updateDurations]);

  // Poll devices when on devices view
  useEffect(() => {
    if (selectedKey !== VIEW_KEYS.DEVICES) {
      return;
    }

    let timeoutId: any;
    let isActive = true;

    const poll = async () => {
      if (!isActive) return;
      await fetchDevicesWithFeedback(true);
      if (isActive) {
        timeoutId = setTimeout(poll, 3000);
      }
    };

    timeoutId = setTimeout(poll, 3000);

    return () => {
      isActive = false;
      clearTimeout(timeoutId);
    };
  }, [selectedKey]);

  const renderContent = () => {
    switch (selectedKey) {
      case VIEW_KEYS.DEVICES:
        return (
          <DevicesView
            devices={devices}
            historyDevices={historyDevices}
            loading={loading}
            fetchDevices={fetchDevicesWithFeedback}
            busyDevices={busyDevices}
            setSelectedKey={setSelectedKey}
            setSelectedDevice={setSelectedDevice}
            handleStartScrcpy={async (deviceId) => {
              setSelectedDevice(deviceId);
              setSelectedKey(VIEW_KEYS.MIRROR);
            }}
            handleFetchDeviceInfo={handleFetchDeviceInfoWithFeedback}
            onShowWirelessConnect={showWirelessConnect}
            handleSwitchToWireless={handleSwitchToWirelessWithFeedback}
            handleAdbConnect={handleAdbConnectWithFeedback}
            handleAdbDisconnect={handleAdbDisconnectWithFeedback}
            handleRemoveHistoryDevice={handleRemoveHistoryDeviceWithFeedback}
            handleOpenSettings={handleOpenSettingsWithFeedback}
            handleTogglePin={handleTogglePinWithFeedback}
            mirrorStatuses={mirrorStatuses}
            recordStatuses={recordStatuses}
          />
        );
      case VIEW_KEYS.FILES:
        return (
          <FilesView
            devices={devices}
            selectedDevice={selectedDevice}
            setSelectedDevice={setSelectedDevice}
            fetchDevices={fetchDevicesWithFeedback}
            loading={loading}
          />
        );
      case VIEW_KEYS.APPS:
        return (
          <AppsView
            devices={devices}
            selectedDevice={selectedDevice}
            setSelectedDevice={setSelectedDevice}
            fetchDevices={fetchDevicesWithFeedback}
            loading={loading}
            setSelectedKey={setSelectedKey}
            handleJumpToLogcat={handleJumpToLogcat}
          />
        );
      case VIEW_KEYS.SHELL:
        return (
          <ShellView
            devices={devices}
            selectedDevice={selectedDevice}
            setSelectedDevice={setSelectedDevice}
            fetchDevices={fetchDevicesWithFeedback}
            loading={loading}
          />
        );
      case VIEW_KEYS.LOGCAT:
        return (
          <LogcatView
            devices={devices}
            selectedDevice={selectedDevice}
            setSelectedDevice={setSelectedDevice}
            fetchDevices={fetchDevicesWithFeedback}
            isLogging={isLogging}
            toggleLogcat={toggleLogcatWithFeedback}
            selectedPackage={selectedPackage}
            setSelectedPackage={setSelectedPackage}
            logs={logs}
            setLogs={setLogs}
            logFilter={logFilter}
            setLogFilter={setLogFilter}
            useRegex={useRegex}
            setUseRegex={setUseRegex}
            preFilter={preFilter}
            setPreFilter={setPreFilter}
            preUseRegex={preUseRegex}
            setPreUseRegex={setPreUseRegex}
            excludeFilter={excludeFilter}
            setExcludeFilter={setExcludeFilter}
            excludeUseRegex={excludeUseRegex}
            setExcludeUseRegex={setExcludeUseRegex}
          />
        );
      case VIEW_KEYS.MIRROR:
        return (
          <MirrorView
            devices={devices}
            selectedDevice={selectedDevice}
            setSelectedDevice={setSelectedDevice}
            fetchDevices={fetchDevicesWithFeedback}
            loading={loading}
            mirrorStatuses={mirrorStatuses}
            recordStatuses={recordStatuses}
          />
        );
      case VIEW_KEYS.PROXY:
        return (
          <ProxyView
            devices={devices}
            selectedDevice={selectedDevice}
            setSelectedDevice={setSelectedDevice}
            fetchDevices={fetchDevicesWithFeedback}
            loading={loading}
          />
        );

      default:
        return <div style={{ padding: 24 }}>{t("app.select_option")}</div>;
    }
  };

  return (
    <Layout style={{ minHeight: "100vh" }}>
      <Sider
        width={220}
        theme="light"
        style={{
          height: "100vh",
          position: "fixed",
          left: 0,
          top: 0,
          bottom: 0,
          display: "flex",
          flexDirection: "column",
          backgroundColor: isDark ? '#2C2C2E' : "#F5F5F7",
          borderRight: isDark ? "1px solid rgba(255,255,255,0.06)" : "1px solid rgba(0,0,0,0.06)",
        }}
      >
        <div
          style={{ display: "flex", flexDirection: "column", height: "100%" }}
        >
          <div className="drag-handle" style={{ height: 38, width: "100%", flexShrink: 0 }} />
          <div style={{ flex: 1, overflowY: "auto", padding: "0 8px" }}>
            <Menu
              theme={isDark ? "dark" : "light"}
              selectedKeys={[selectedKey]}
              mode="inline"
              onClick={({ key }) => setSelectedKey(key as any)}
              style={{
                backgroundColor: "transparent",
                borderRight: "none",
              }}
              items={[
                { key: VIEW_KEYS.DEVICES, icon: <MobileOutlined />, label: t("menu.devices") },
                { key: VIEW_KEYS.MIRROR, icon: <DesktopOutlined />, label: t("menu.mirror") },
                { key: VIEW_KEYS.APPS, icon: <AppstoreOutlined />, label: t("menu.apps") },
                { key: VIEW_KEYS.FILES, icon: <FolderOutlined />, label: t("menu.files") },
                { key: VIEW_KEYS.SHELL, icon: <CodeOutlined />, label: t("menu.shell") },
                { key: VIEW_KEYS.LOGCAT, icon: <FileTextOutlined />, label: t("menu.logcat") },
                { key: VIEW_KEYS.PROXY, icon: <GlobalOutlined />, label: t("menu.proxy") || "Proxy" },
              ]}
            />
          </div>
          <div
            style={{
              padding: "8px 16px",
              borderTop: isDark ? "1px solid rgba(255, 255, 255, 0.06)" : "1px solid rgba(0, 0, 0, 0.06)",
              display: "flex",
              justifyContent: "center",
              gap: "4px",
              flexWrap: "wrap",
            }}
          >
            <Dropdown
              menu={{
                items: [
                  {
                    key: "light",
                    label: t("app.theme_light") || "Light",
                    icon: <SunOutlined />,
                    onClick: () => setMode("light")
                  },
                  {
                    key: "dark",
                    label: t("app.theme_dark") || "Dark",
                    icon: <MoonOutlined />,
                    onClick: () => setMode("dark")
                  },
                  {
                    key: "system",
                    label: t("app.theme_system") || "System",
                    icon: <DesktopOutlined />,
                    onClick: () => setMode("system")
                  },
                ],
                selectedKeys: [mode],
              }}
              placement="top"
            >
              <Button
                type="text"
                size="small"
                icon={
                  mode === 'light' ? <SunOutlined style={{ fontSize: "16px", color: isDark ? "rgba(255,255,255,0.65)" : "rgba(0,0,0,0.65)" }} /> :
                  mode === 'dark' ? <MoonOutlined style={{ fontSize: "16px", color: isDark ? "rgba(255,255,255,0.65)" : "rgba(0,0,0,0.65)" }} /> :
                  <DesktopOutlined style={{ fontSize: "16px", color: isDark ? "rgba(255,255,255,0.65)" : "rgba(0,0,0,0.65)" }} />
                }
                title={t("app.change_theme") || "Change Theme"}
              />
            </Dropdown>

            <Dropdown
              menu={{
                items: [
                  { key: "en", label: "English", onClick: () => i18n.changeLanguage("en") },
                  { key: "zh", label: "简体中文", onClick: () => i18n.changeLanguage("zh") },
                  { key: "zh-TW", label: "繁體中文 (台灣)", onClick: () => i18n.changeLanguage("zh-TW") },
                  { key: "ja", label: "日本語", onClick: () => i18n.changeLanguage("ja") },
                  { key: "ko", label: "한국어", onClick: () => i18n.changeLanguage("ko") },
                ],
                selectedKeys: [i18n.language],
              }}
              placement="top"
            >
              <Button
                type="text"
                size="small"
                icon={<TranslationOutlined style={{ fontSize: "16px", color: isDark ? "rgba(255,255,255,0.65)" : "rgba(0,0,0,0.65)" }} />}
                title={t("app.change_language")}
              />
            </Dropdown>
            <Button
              type="text"
              size="small"
              icon={<InfoCircleOutlined style={{ fontSize: "16px", color: isDark ? "rgba(255,255,255,0.65)" : "rgba(0,0,0,0.65)" }} />}
              onClick={showAbout}
              title={t("app.about")}
            />
            <Button
              type="text"
              size="small"
              icon={<GithubOutlined style={{ fontSize: "16px", color: isDark ? "rgba(255,255,255,0.65)" : "rgba(0,0,0,0.65)" }} />}
              onClick={() => BrowserOpenURL && BrowserOpenURL("https://github.com/niceto0/gaze")}
              title={t("app.github")}
            />
            <Button
              type="text"
              size="small"
              icon={<BugOutlined style={{ fontSize: "16px", color: isDark ? "rgba(255,255,255,0.65)" : "rgba(0,0,0,0.65)" }} />}
              onClick={showFeedback}
              title={t("app.feedback")}
            />
          </div>
        </div>
      </Sider>
      <Layout className="site-layout" style={{ marginLeft: 220 }}>
        <Content style={{ margin: "0", height: "100vh", overflow: "hidden", display: "flex", flexDirection: "column" }}>
          <div className="drag-handle" style={{ height: 38, width: "100%", flexShrink: 0 }} />
          {renderContent()}
        </Content>
      </Layout>

      <DeviceInfoModal
        visible={deviceInfoVisible}
        onCancel={closeDeviceInfo}
        deviceInfo={selectedDeviceInfo}
        loading={deviceInfoLoading}
        onRefresh={() => selectedDeviceInfo && handleFetchDeviceInfoWithFeedback(selectedDeviceInfo.serial || selectedDevice)}
      />

      <AboutModal visible={aboutVisible} onCancel={hideAbout} />

      <WirelessConnectModal
        visible={wirelessConnectVisible}
        onCancel={hideWirelessConnect}
        onConnect={handleAdbConnectWithFeedback}
        onPair={handleAdbPairWithFeedback}
      />

      <FeedbackModal
        visible={feedbackVisible}
        onCancel={hideFeedback}
        appVersion={appVersion}
        deviceInfo={devices.find(d => d.id === selectedDevice) ? `${devices.find(d => d.id === selectedDevice)?.brand} ${devices.find(d => d.id === selectedDevice)?.model} (ID: ${selectedDevice})` : "None"}
      />
    </Layout>
  );
}

export default App;
