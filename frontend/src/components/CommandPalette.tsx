/**
 * CommandPalette - Global search & command palette (Cmd+K / Ctrl+K)
 *
 * Searches across: pages, commands, devices, proxy requests, workflows, sessions.
 * Supports fuzzy matching, keyboard navigation, and grouped results.
 */

import { useEffect, useRef, useMemo, useCallback } from 'react';
import { Input, Modal, Typography, theme as antTheme, Tag } from 'antd';
import {
  SearchOutlined,
  AppstoreOutlined,
  ThunderboltOutlined,
  MobileOutlined,
  GlobalOutlined,
  BranchesOutlined,
  DatabaseOutlined,
  DesktopOutlined,
  CodeOutlined,
  FileTextOutlined,
  FolderOutlined,
  BlockOutlined,
  VideoCameraOutlined,
  DashboardOutlined,
  LineChartOutlined,
  SunOutlined,
  MoonOutlined,
  WifiOutlined,
  BugOutlined,
  InfoCircleOutlined,
  ApiOutlined,
  EditOutlined,
  PauseCircleOutlined,
  FileProtectOutlined,
  PlusOutlined,
  CommentOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import {
  useCommandStore,
  fuzzyMatch,
  PAGE_ITEMS,
  COMMAND_ITEMS,
  CommandResult,
  CommandGroup,
  CommandResultType,
} from '../stores/commandStore';
import { useDeviceStore } from '../stores/deviceStore';
import { useProxyStore } from '../stores/proxyStore';
import { useWorkflowStore } from '../stores/workflowStore';
import { useEventStore } from '../stores/eventStore';
import { useUIStore } from '../stores/uiStore';
import { VIEW_KEYS } from '../stores/types';
import { useTheme } from '../ThemeContext';

const { Text } = Typography;

// --- Icon mapping ---

const PAGE_ICONS: Record<string, React.ReactNode> = {
  [VIEW_KEYS.DEVICES]: <MobileOutlined />,
  [VIEW_KEYS.APPS]: <AppstoreOutlined />,
  [VIEW_KEYS.SHELL]: <CodeOutlined />,
  [VIEW_KEYS.LOGCAT]: <FileTextOutlined />,
  [VIEW_KEYS.MIRROR]: <DesktopOutlined />,
  [VIEW_KEYS.FILES]: <FolderOutlined />,
  [VIEW_KEYS.PROXY]: <GlobalOutlined />,
  [VIEW_KEYS.RECORDING]: <VideoCameraOutlined />,
  [VIEW_KEYS.WORKFLOW]: <BranchesOutlined />,
  [VIEW_KEYS.INSPECTOR]: <BlockOutlined />,
  [VIEW_KEYS.EVENTS]: <DashboardOutlined />,
  [VIEW_KEYS.SESSIONS]: <DatabaseOutlined />,
  [VIEW_KEYS.PERFORMANCE]: <LineChartOutlined />,
};

const COMMAND_ICONS: Record<string, React.ReactNode> = {
  toggle_theme: <SunOutlined />,
  wireless_connect: <WifiOutlined />,
  mock_rules: <ApiOutlined />,
  add_mock_rule: <PlusOutlined />,
  map_remote_rules: <GlobalOutlined />,
  add_map_remote_rule: <PlusOutlined />,
  rewrite_rules: <EditOutlined />,
  add_rewrite_rule: <PlusOutlined />,
  breakpoint_rules: <PauseCircleOutlined />,
  add_breakpoint_rule: <PlusOutlined />,
  proto_files: <FileProtectOutlined />,
  new_workflow: <PlusOutlined />,
  about: <InfoCircleOutlined />,
  feedback: <CommentOutlined />,
};

const TYPE_ICONS: Record<CommandResultType, React.ReactNode> = {
  page: <AppstoreOutlined />,
  command: <ThunderboltOutlined />,
  device: <MobileOutlined />,
  proxy: <GlobalOutlined />,
  workflow: <BranchesOutlined />,
  session: <DatabaseOutlined />,
};

// --- Max results per type ---
const MAX_RESULTS: Record<CommandResultType, number> = {
  page: 13,
  command: 11,
  device: 5,
  proxy: 10,
  workflow: 5,
  session: 5,
};

// --- Component ---

export default function CommandPalette() {
  const { t } = useTranslation();
  const { token } = antTheme.useToken();
  const { isDark, mode, setMode } = useTheme();
  const inputRef = useRef<any>(null);
  const listRef = useRef<HTMLDivElement>(null);

  // Command store
  const { isOpen, query, selectedIndex, close, setQuery, setSelectedIndex, moveSelection } =
    useCommandStore();

  // Other stores (read-only)
  const devices = useDeviceStore((s) => s.devices);
  const historyDevices = useDeviceStore((s) => s.historyDevices);
  const setSelectedDevice = useDeviceStore((s) => s.setSelectedDevice);
  const proxyLogs = useProxyStore((s) => s.logs);
  const workflows = useWorkflowStore((s) => s.workflows);
  const sessions = useEventStore((s) => s.sessions);
  const setSelectedKey = useUIStore((s) => s.setSelectedKey);
  const showAbout = useUIStore((s) => s.showAbout);
  const showWirelessConnect = useUIStore((s) => s.showWirelessConnect);
  const showFeedback = useUIStore((s) => s.showFeedback);
  const openMockListModal = useProxyStore((s) => s.openMockListModal);
  const openMockEditModal = useProxyStore((s) => s.openMockEditModal);
  const openMapRemoteListModal = useProxyStore((s) => s.openMapRemoteListModal);
  const openMapRemoteEditModal = useProxyStore((s) => s.openMapRemoteEditModal);
  const openRewriteListModal = useProxyStore((s) => s.openRewriteListModal);
  const openRewriteEditModal = useProxyStore((s) => s.openRewriteEditModal);
  const openBreakpointListModal = useProxyStore((s) => s.openBreakpointListModal);
  const openBreakpointEditModal = useProxyStore((s) => s.openBreakpointEditModal);
  const openProtoListModal = useProxyStore((s) => s.openProtoListModal);
  const setWorkflowModalVisible = useWorkflowStore((s) => s.setWorkflowModalVisible);
  const selectWorkflow = useWorkflowStore((s) => s.selectWorkflow);
  const selectProxyLog = useProxyStore((s) => s.selectLog);

  // --- Build command actions map ---
  const commandActions = useMemo(
    () =>
      ({
        toggle_theme: () => {
          setMode(isDark ? 'light' : 'dark');
        },
        wireless_connect: () => showWirelessConnect(),
        mock_rules: () => {
          setSelectedKey(VIEW_KEYS.PROXY);
          setTimeout(() => openMockListModal(), 100);
        },
        add_mock_rule: () => {
          setSelectedKey(VIEW_KEYS.PROXY);
          setTimeout(() => openMockEditModal(null), 100);
        },
        map_remote_rules: () => {
          setSelectedKey(VIEW_KEYS.PROXY);
          setTimeout(() => openMapRemoteListModal(), 100);
        },
        add_map_remote_rule: () => {
          setSelectedKey(VIEW_KEYS.PROXY);
          setTimeout(() => openMapRemoteEditModal(null), 100);
        },
        rewrite_rules: () => {
          setSelectedKey(VIEW_KEYS.PROXY);
          setTimeout(() => openRewriteListModal(), 100);
        },
        add_rewrite_rule: () => {
          setSelectedKey(VIEW_KEYS.PROXY);
          setTimeout(() => openRewriteEditModal(null), 100);
        },
        breakpoint_rules: () => {
          setSelectedKey(VIEW_KEYS.PROXY);
          setTimeout(() => openBreakpointListModal(), 100);
        },
        add_breakpoint_rule: () => {
          setSelectedKey(VIEW_KEYS.PROXY);
          setTimeout(() => openBreakpointEditModal(null), 100);
        },
        proto_files: () => {
          setSelectedKey(VIEW_KEYS.PROXY);
          setTimeout(() => openProtoListModal(), 100);
        },
        new_workflow: () => {
          setSelectedKey(VIEW_KEYS.WORKFLOW);
          setTimeout(() => setWorkflowModalVisible(true), 100);
        },
        about: () => showAbout(),
        feedback: () => showFeedback(),
      }) as Record<string, () => void>,
    [
      isDark,
      setMode,
      showWirelessConnect,
      setSelectedKey,
      openMockListModal,
      openMockEditModal,
      openMapRemoteListModal,
      openMapRemoteEditModal,
      openRewriteListModal,
      openRewriteEditModal,
      openBreakpointListModal,
      openBreakpointEditModal,
      openProtoListModal,
      setWorkflowModalVisible,
      showAbout,
      showFeedback,
    ]
  );

  // --- Search logic ---
  const groups = useMemo((): CommandGroup[] => {
    const q = query.trim();
    const result: CommandGroup[] = [];

    // 1. Pages
    const pageResults: CommandResult[] = [];
    for (const page of PAGE_ITEMS) {
      const translatedName = t(page.nameKey);
      const allText = [translatedName, ...page.keywords].join(' ');
      const score = q ? fuzzyMatch(q, allText) : 0;
      if (score >= 0) {
        pageResults.push({
          id: `page:${page.key}`,
          type: 'page',
          title: translatedName,
          icon: page.key,
          action: () => setSelectedKey(page.key),
        });
      }
    }
    if (pageResults.length > 0) {
      result.push({ type: 'page', label: t('command.group_pages'), results: pageResults.slice(0, MAX_RESULTS.page) });
    }

    // 2. Commands
    const cmdResults: CommandResult[] = [];
    for (const cmd of COMMAND_ITEMS) {
      const translatedName = t(cmd.nameKey);
      const allText = [translatedName, ...cmd.keywords].join(' ');
      const score = q ? fuzzyMatch(q, allText) : 0;
      if (score >= 0 && cmd.id in commandActions) {
        cmdResults.push({
          id: `cmd:${cmd.id}`,
          type: 'command',
          title: translatedName,
          icon: cmd.id,
          action: commandActions[cmd.id],
        });
      }
    }
    if (cmdResults.length > 0) {
      result.push({ type: 'command', label: t('command.group_commands'), results: cmdResults.slice(0, MAX_RESULTS.command) });
    }

    // 3. Devices (only when there's a query, to avoid clutter)
    if (q) {
      const deviceResults: CommandResult[] = [];
      const allDevices = [
        ...devices.map((d) => ({ ...d, isOnline: true })),
        ...historyDevices
          .filter((h) => !devices.some((d) => d.id === h.id))
          .map((h) => ({ ...h, state: 'offline', isOnline: false })),
      ];
      for (const dev of allDevices) {
        const text = [dev.model, dev.brand, dev.serial, dev.id].filter(Boolean).join(' ');
        const score = fuzzyMatch(q, text);
        if (score >= 0) {
          deviceResults.push({
            id: `device:${dev.id}`,
            type: 'device',
            title: dev.model || dev.serial || dev.id,
            subtitle: dev.brand ? `${dev.brand} · ${dev.serial}` : dev.serial,
            action: () => {
              setSelectedDevice(dev.id);
              setSelectedKey(VIEW_KEYS.DEVICES);
            },
          });
        }
      }
      if (deviceResults.length > 0) {
        result.push({
          type: 'device',
          label: t('command.group_devices'),
          results: deviceResults.slice(0, MAX_RESULTS.device),
        });
      }
    }

    // 4. Workflows (only when there's a query)
    if (q) {
      const wfResults: CommandResult[] = [];
      for (const wf of workflows) {
        const text = [wf.name, wf.description || ''].join(' ');
        const score = fuzzyMatch(q, text);
        if (score >= 0) {
          wfResults.push({
            id: `wf:${wf.id}`,
            type: 'workflow',
            title: wf.name,
            subtitle: wf.description || `${wf.steps?.length || 0} steps`,
            action: () => {
              setSelectedKey(VIEW_KEYS.WORKFLOW);
              setTimeout(() => selectWorkflow(wf.id), 100);
            },
          });
        }
      }
      if (wfResults.length > 0) {
        result.push({
          type: 'workflow',
          label: t('command.group_workflows'),
          results: wfResults.slice(0, MAX_RESULTS.workflow),
        });
      }
    }

    // 5. Sessions (only when there's a query)
    if (q) {
      const sessionResults: CommandResult[] = [];
      const sessionArray = Array.from(sessions.values());
      for (const sess of sessionArray) {
        const text = [sess.name, sess.status, sess.deviceId].filter(Boolean).join(' ');
        const score = fuzzyMatch(q, text);
        if (score >= 0) {
          sessionResults.push({
            id: `session:${sess.id}`,
            type: 'session',
            title: sess.name || sess.id,
            subtitle: sess.status,
            action: () => {
              setSelectedKey(VIEW_KEYS.SESSIONS);
            },
          });
        }
      }
      if (sessionResults.length > 0) {
        result.push({
          type: 'session',
          label: t('command.group_sessions'),
          results: sessionResults.slice(0, MAX_RESULTS.session),
        });
      }
    }

    // 6. Proxy requests (only when query length >= 2, since there can be many)
    if (q && q.length >= 2) {
      const proxyResults: CommandResult[] = [];
      // Search last 500 logs max for performance
      const logsToSearch = proxyLogs.slice(-500);
      for (const log of logsToSearch) {
        const text = [log.method, log.url, String(log.statusCode || '')].join(' ');
        const score = fuzzyMatch(q, text);
        if (score >= 0) {
          proxyResults.push({
            id: `proxy:${log.id}`,
            type: 'proxy',
            title: `${log.method} ${log.url}`,
            subtitle: log.statusCode ? `${log.statusCode}` : 'pending',
            action: () => {
              setSelectedKey(VIEW_KEYS.PROXY);
              setTimeout(() => selectProxyLog(log), 100);
            },
          });
        }
        if (proxyResults.length >= MAX_RESULTS.proxy) break;
      }
      if (proxyResults.length > 0) {
        result.push({
          type: 'proxy',
          label: t('command.group_proxy'),
          results: proxyResults,
        });
      }
    }

    return result;
  }, [
    query,
    t,
    devices,
    historyDevices,
    workflows,
    sessions,
    proxyLogs,
    commandActions,
    setSelectedKey,
    setSelectedDevice,
    selectWorkflow,
    selectProxyLog,
  ]);

  // Flatten results for index-based navigation
  const flatResults = useMemo(() => groups.flatMap((g) => g.results), [groups]);

  // --- Execute selected result ---
  const executeResult = useCallback(
    (result: CommandResult) => {
      close();
      // Delay slightly so the modal closes before the action fires
      setTimeout(() => result.action(), 50);
    },
    [close]
  );

  // --- Keyboard handler ---
  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      switch (e.key) {
        case 'ArrowDown':
          e.preventDefault();
          moveSelection(1, flatResults.length - 1);
          break;
        case 'ArrowUp':
          e.preventDefault();
          moveSelection(-1, flatResults.length - 1);
          break;
        case 'Enter':
          e.preventDefault();
          if (flatResults[selectedIndex]) {
            executeResult(flatResults[selectedIndex]);
          }
          break;
        case 'Escape':
          e.preventDefault();
          close();
          break;
      }
    },
    [flatResults, selectedIndex, moveSelection, executeResult, close]
  );

  // Focus input when opened
  useEffect(() => {
    if (isOpen) {
      setTimeout(() => inputRef.current?.focus(), 50);
    }
  }, [isOpen]);

  // Scroll selected item into view
  useEffect(() => {
    if (!listRef.current) return;
    const el = listRef.current.querySelector(`[data-index="${selectedIndex}"]`);
    if (el) {
      el.scrollIntoView({ block: 'nearest' });
    }
  }, [selectedIndex]);

  // --- Render helpers ---

  const getResultIcon = (result: CommandResult): React.ReactNode => {
    if (result.type === 'page' && result.icon) {
      return PAGE_ICONS[result.icon] || TYPE_ICONS.page;
    }
    if (result.type === 'command' && result.icon) {
      return COMMAND_ICONS[result.icon] || TYPE_ICONS.command;
    }
    return TYPE_ICONS[result.type];
  };

  const getStatusColor = (subtitle?: string): string | undefined => {
    if (!subtitle) return undefined;
    const s = subtitle.toLowerCase();
    if (s === 'active' || s === 'running') return 'green';
    if (s === 'completed') return 'blue';
    if (s === 'failed' || s === 'error') return 'red';
    return undefined;
  };

  // Track flat index across groups
  let flatIndex = 0;

  return (
    <Modal
      open={isOpen}
      onCancel={close}
      footer={null}
      closable={false}
      width={620}
      centered={false}
      style={{ top: '15%' }}
      styles={{
        body: { padding: 0 },
        mask: { backdropFilter: 'blur(2px)' },
      }}
      maskClosable
      destroyOnHidden
    >
      <div onKeyDown={handleKeyDown}>
        {/* Search input */}
        <div
          style={{
            padding: '12px 16px',
            borderBottom: `1px solid ${token.colorBorderSecondary}`,
          }}
        >
          <Input
            ref={inputRef}
            prefix={<SearchOutlined style={{ color: token.colorTextPlaceholder }} />}
            placeholder={t('command.placeholder')}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            variant="borderless"
            size="large"
            style={{ fontSize: 16 }}
            suffix={
              <Tag
                style={{
                  margin: 0,
                  fontSize: 11,
                  padding: '0 6px',
                  lineHeight: '20px',
                  color: token.colorTextDescription,
                  borderColor: token.colorBorderSecondary,
                }}
              >
                ESC
              </Tag>
            }
          />
        </div>

        {/* Results */}
        <div
          ref={listRef}
          style={{
            maxHeight: 420,
            overflowY: 'auto',
            padding: '4px 0',
          }}
        >
          {flatResults.length === 0 && query.trim() ? (
            <div
              style={{
                padding: '24px 16px',
                textAlign: 'center',
                color: token.colorTextDescription,
              }}
            >
              {t('command.no_results')}
            </div>
          ) : (
            groups.map((group) => (
              <div key={group.type}>
                {/* Group header */}
                <div
                  style={{
                    padding: '8px 16px 4px',
                    fontSize: 11,
                    fontWeight: 600,
                    color: token.colorTextDescription,
                    textTransform: 'uppercase',
                    letterSpacing: '0.05em',
                  }}
                >
                  {group.label}
                </div>

                {/* Group items */}
                {group.results.map((result) => {
                  const currentFlatIndex = flatIndex++;
                  const isSelected = currentFlatIndex === selectedIndex;

                  return (
                    <div
                      key={result.id}
                      data-index={currentFlatIndex}
                      onClick={() => executeResult(result)}
                      onMouseEnter={() => setSelectedIndex(currentFlatIndex)}
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: 12,
                        padding: '8px 16px',
                        cursor: 'pointer',
                        borderRadius: 6,
                        margin: '0 6px',
                        backgroundColor: isSelected
                          ? token.colorFillSecondary
                          : 'transparent',
                        transition: 'background-color 0.1s',
                      }}
                    >
                      {/* Icon */}
                      <span
                        style={{
                          fontSize: 16,
                          color: isSelected ? token.colorPrimary : token.colorTextSecondary,
                          width: 20,
                          textAlign: 'center',
                          flexShrink: 0,
                        }}
                      >
                        {getResultIcon(result)}
                      </span>

                      {/* Title + subtitle */}
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <Text
                          ellipsis
                          style={{
                            display: 'block',
                            fontSize: 13,
                            lineHeight: '20px',
                            color: isSelected ? token.colorText : token.colorText,
                          }}
                        >
                          {result.title}
                        </Text>
                      </div>

                      {/* Subtitle / badge */}
                      {result.subtitle && (
                        <span style={{ flexShrink: 0 }}>
                          {getStatusColor(result.subtitle) ? (
                            <Tag
                              color={getStatusColor(result.subtitle)}
                              style={{ margin: 0, fontSize: 11 }}
                            >
                              {result.subtitle}
                            </Tag>
                          ) : (
                            <Text
                              type="secondary"
                              style={{ fontSize: 12 }}
                              ellipsis
                            >
                              {result.subtitle}
                            </Text>
                          )}
                        </span>
                      )}
                    </div>
                  );
                })}
              </div>
            ))
          )}
        </div>

        {/* Footer hints */}
        {flatResults.length > 0 && (
          <div
            style={{
              padding: '8px 16px',
              borderTop: `1px solid ${token.colorBorderSecondary}`,
              display: 'flex',
              gap: 16,
              fontSize: 11,
              color: token.colorTextDescription,
            }}
          >
            <span>↑↓ {t('command.hint_navigate')}</span>
            <span>↵ {t('command.hint_select')}</span>
            <span>ESC {t('command.hint_esc')}</span>
          </div>
        )}
      </div>
    </Modal>
  );
}
