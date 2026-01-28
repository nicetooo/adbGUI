import React, { useEffect, useCallback } from 'react';
import { Card, Button, InputNumber, Space, Typography, Tag, Divider, Switch, Tooltip, Radio, Input, Tabs, theme, Form, Table, Popconfirm, Popover, Spin, App, Modal } from 'antd';
import { PoweroffOutlined, PlayCircleOutlined, DeleteOutlined, SettingOutlined, LockOutlined, GlobalOutlined, ArrowUpOutlined, ArrowDownOutlined, ApiOutlined, SafetyCertificateOutlined, DownloadOutlined, HourglassOutlined, CopyOutlined, BlockOutlined, SendOutlined, CloseOutlined, PlusOutlined, EditOutlined } from '@ant-design/icons';
import VirtualList from './VirtualList';
import DeviceSelector from './DeviceSelector';
import { useDeviceStore, useProxyStore, RequestLog as StoreRequestLog } from '../stores';
// @ts-ignore
import { StartProxy, StopProxy, GetProxyStatus, GetLocalIP, RunAdbCommand, StartNetworkMonitor, StopNetworkMonitor, SetProxyLimit, SetProxyWSEnabled, SetProxyMITM, InstallProxyCert, SetProxyLatency, SetMITMBypassPatterns, SetProxyDevice, ResendRequest, AddMockRule, RemoveMockRule, GetMockRules, ToggleMockRule, CheckCertTrust, SetupProxyForDevice, CleanupProxyForDevice } from '../../wailsjs/go/main/App';
// @ts-ignore
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import { useTranslation } from 'react-i18next';

const { Title, Text } = Typography;

interface NetworkStats {
    rxSpeed: number;
    txSpeed: number;
    rxBytes: number;
    txBytes: number;
    time: number;
}

interface RequestLog {
    id: string;
    time: string;
    clientIp: string;
    method: string;
    url: string;
    headers: Record<string, string[]>;
    body: string; // Legacy field, might be unused now
    isHttps: boolean;

    // New fields
    statusCode?: number;
    contentType?: string;
    bodySize?: number;
    previewBody?: string;
    respHeaders?: Record<string, string[]>;
    respBody?: string;
    isWs?: boolean;
    mocked?: boolean;
}

const ProxyView: React.FC = () => {
    const { t } = useTranslation();
    const { token } = theme.useToken();
    const { modal, message } = App.useApp();
    const { selectedDevice } = useDeviceStore();

    // Use proxyStore instead of useState
    const {
        isRunning,
        port,
        localIP,
        logs,
        wsEnabled,
        mitmEnabled,
        filterType,
        searchText,
        latency,
        bypassPatterns,
        isBypassModalOpen,
        newPattern,
        selectedLog,
        netStats,
        dlLimit,
        ulLimit,
        setProxyRunning,
        setPort,
        setLocalIP,
        addLog,
        updateLog,
        clearLogs,
        toggleWS,
        toggleMITM,
        setFilterType,
        setSearchText,
        setLatency,
        setBypassModalOpen,
        setNewPattern,
        selectLog,
        setNetStats,
        setSpeedLimits,
        addBypassPattern,
        removeBypassPattern,
    } = useProxyStore();

    const [resendForm] = Form.useForm();
    const [mockForm] = Form.useForm();

    // Additional proxy store state
    const {
        resendModalOpen,
        resendLoading,
        resendResponse,
        mockModalOpen,
        mockRules,
        editingMockRule,
        certTrustStatus,
        isAIParsing,
        aiSearchText,
        aiPopoverOpen,
        setResendModalOpen,
        setResendLoading,
        setResendResponse,
        closeResendModal,
        setMockModalOpen,
        setMockRules,
        setEditingMockRule,
        closeMockModal,
        setCertTrustStatus,
        setIsAIParsing,
        setAiSearchText,
        setAiPopoverOpen,
    } = useProxyStore();

    // Check cert trust status when MITM is enabled and proxy is running
    useEffect(() => {
        if (mitmEnabled && isRunning && selectedDevice) {
            setCertTrustStatus('checking');
            CheckCertTrust(selectedDevice).then((status: string) => {
                setCertTrustStatus(status);
            }).catch(() => {
                setCertTrustStatus('unknown');
            });
        } else {
            setCertTrustStatus(null);
        }
    }, [mitmEnabled, isRunning, selectedDevice]);

    // Re-check cert status periodically when pending
    useEffect(() => {
        if (certTrustStatus === 'pending' && selectedDevice) {
            const interval = setInterval(() => {
                CheckCertTrust(selectedDevice).then((status: string) => {
                    if (status !== 'pending') {
                        setCertTrustStatus(status);
                    }
                }).catch(() => {});
            }, 2000); // Check every 2 seconds
            return () => clearInterval(interval);
        }
    }, [certTrustStatus, selectedDevice]);

    useEffect(() => {
        // Listen for network stats
        EventsOn("network-stats", (stats: NetworkStats) => {
            setNetStats(stats);
        });

        return () => {
            EventsOff("network-stats");
        };
    }, []);

    // Auto-start network monitor when device is selected
    useEffect(() => {
        if (selectedDevice) {
            StartNetworkMonitor(selectedDevice);
        }
        return () => {
            if (selectedDevice) {
                StopNetworkMonitor(selectedDevice);
            }
        };
    }, [selectedDevice]);

    useEffect(() => {
        // Initial status check
        GetProxyStatus().then((status: boolean) => setProxyRunning(status));
        GetLocalIP().then((ip: string) => setLocalIP(ip));

        // Listen for proxy status changes (e.g. started by session config)
        const handleProxyStatus = (data: { running: boolean; port: number }) => {
            setProxyRunning(data.running);
            if (data.port) {
                setPort(data.port);
            }
        };
        EventsOn("proxy-status-changed", handleProxyStatus);

        // Sync settings from backend
        // Note: These settings are managed by the store and backend
        // @ts-ignore
        import('../../wailsjs/go/main/App').then(m => {
            // Settings are already synced via store
        });

        // Listen for network events from session (unified event source)
        const handleSessionBatch = (events: any[]) => {
            // Filter for network events only
            const networkEvents = events.filter((e: any) => e.category === 'network');

            for (const event of networkEvents) {
                // UnifiedEvent uses 'data' field (JSON), parse it
                let detail: any = {};
                if (event.data) {
                    try {
                        detail = typeof event.data === 'string' ? JSON.parse(event.data) : event.data;
                    } catch (e) {
                        console.error('Failed to parse event data:', e);
                    }
                }
                const timeStr = event.timestamp ? new Date(event.timestamp).toLocaleTimeString() : '';
                // Convert session event to RequestLog format
                const log: RequestLog = {
                    id: detail.id || event.id,
                    time: timeStr,
                    clientIp: detail.clientIp || '',
                    method: detail.method || 'UNKNOWN',
                    url: detail.url || event.title,
                    headers: detail.requestHeaders || {},
                    body: detail.requestBody || '',
                    isHttps: detail.isHttps || false,
                    statusCode: detail.statusCode,
                    contentType: detail.contentType,
                    bodySize: detail.bodySize,
                    previewBody: detail.requestBody,
                    respHeaders: detail.responseHeaders,
                    respBody: detail.responseBody,
                    isWs: detail.isWs || false,
                    mocked: detail.mocked || false,
                };

                // Use getState() to get current logs (avoid stale closure)
                const currentLogs = useProxyStore.getState().logs;
                const existingLog = currentLogs.find(l => l.id === log.id);
                if (existingLog) {
                    updateLog(log.id, log);
                } else {
                    addLog(log);
                }
            }
        };

        EventsOn("session-events-batch", handleSessionBatch);

        return () => {
            EventsOff("session-events-batch");
            EventsOff("proxy-status-changed");
        };
    }, []);

    useEffect(() => {
        // Auto-scroll logic could be added here if needed, but Table manages its own scroll usually.
        // If we want to stick to bottom, we could scroll the table body container.
        // For now, let's just let it be or use a "scroll to bottom" generic approach if user wants.
    }, [logs]);

    const handleStart = async () => {
        try {
            // 1. Start Server (binds to localhost only for security)
            await SetProxyMITM(mitmEnabled);
            await SetProxyWSEnabled(wsEnabled);
            await StartProxy(port);
            setProxyRunning(true);

            // 2. Set proxy device for event association
            if (selectedDevice) {
                await SetProxyDevice(selectedDevice);
            }

            // 3. Setup adb reverse + device proxy if device selected
            if (selectedDevice) {
                try {
                    await SetupProxyForDevice(selectedDevice, port);
                    message.success(t('proxy.start_success', { ip: '127.0.0.1', port }));
                } catch (adbErr: any) {
                    const errorStr = String(adbErr);
                    if (errorStr.includes("WRITE_SECURE_SETTINGS")) {
                        message.warning(t('proxy.link_failed', { error: "Security Settings blocked" }));
                        showSecurityError();
                    } else {
                        message.error(t('proxy.link_failed', { error: errorStr }));
                    }
                }
            } else {
                message.success(t('proxy.start_local_success'));
            }
        } catch (err) {
            message.error(t('app.command_failed') + ": " + String(err));
        }
    };

    const handleStop = async () => {
        try {
            // 1. Cleanup adb reverse and device proxy
            if (selectedDevice) {
                try {
                    await CleanupProxyForDevice(selectedDevice, port);
                } catch (e) { }
            }

            // 2. Stop Server
            await StopProxy();
            setProxyRunning(false);
            message.success(t('proxy.stop_success'));
        } catch (err) {
            message.error(t('app.command_failed') + ": " + String(err));
        }
    };

    const showSecurityError = () => {
        modal.error({
            title: t('proxy.permission_denied'),
            content: (
                <div>
                    <p>{t('proxy.permission_denied_desc')}</p>
                    <p><strong>{t('proxy.xiaomi_users')}</strong></p>
                    <p>{t('proxy.xiaomi_desc')}</p>
                    <p style={{ fontSize: '12px', color: '#999' }}>{t('proxy.xiaomi_note')}</p>
                    <br />
                    <p><strong>{t('proxy.other_devices')}</strong></p>
                    <p>{t('proxy.other_desc')}</p>
                </div>
            ),
        });
    };

    const columns = [
        {
            title: t('proxy.col_time'),
            dataIndex: 'time',
            key: 'time',
            width: 100,
        },
        {
            title: t('proxy.col_method'),
            dataIndex: 'method',
            key: 'method',
            width: 100,
            render: (method: string, record: RequestLog) => {
                let color = 'default';
                if (method === 'GET') color = 'green';
                else if (method === 'POST') color = 'blue';
                else if (method === 'PUT') color = 'orange';
                else if (method === 'DELETE') color = 'red';
                else if (method === 'CONNECT') color = 'purple';
                else if (method === 'WS') color = 'cyan';

                return (
                    <Space>
                        <Tag color={color}>{method}</Tag>
                        {record.isHttps && <LockOutlined style={{ color: '#52c41a' }} title="HTTPS Tunnel" />}
                    </Space>
                );
            }
        },
        {
            title: t('proxy.col_url'),
            dataIndex: 'url',
            key: 'url',
            ellipsis: true,
            render: (url: string) => (
                <Text style={{ fontFamily: "'Fira Code', monospace", fontSize: '13px' }} title={url}>{url}</Text>
            )
        },
        {
            title: 'Client IP',
            dataIndex: 'clientIp',
            key: 'clientIp',
            width: 140,
        },
    ];

    const handleWSToggle = async (checked: boolean) => {
        try {
            await SetProxyWSEnabled(checked);
            toggleWS();
            message.success(checked ? t('proxy.on') : t('proxy.off'));
        } catch (err) {
            message.error(t('app.command_failed') + ": " + String(err));
        }
    };

    const handleMITMToggle = async (checked: boolean) => {
        try {
            await SetProxyMITM(checked);
            toggleMITM();
            if (checked) {
                message.info(t('proxy.mitm_tooltip'));
            } else {
                message.success(t('proxy.off'));
            }
        } catch (err) {
            message.error(t('app.command_failed') + ": " + String(err));
        }
    };

    const handleInstallCert = async () => {
        if (!selectedDevice) return;
        try {
            const path = await InstallProxyCert(selectedDevice);
            modal.success({
                title: t('proxy.cert_pushed'),
                content: (
                    <div>
                        <p>{t('proxy.cert_pushed_desc', { path })}</p>
                        <Divider style={{ margin: '12px 0' }} />
                        <p><strong>{t('proxy.next_steps')}</strong></p>
                        <ol style={{ paddingLeft: 20 }}>
                            <li>{t('proxy.step_1')}</li>
                            <li>{t('proxy.step_2')}</li>
                            <li>{t('proxy.step_3')}</li>
                            <li>{t('proxy.step_4')}</li>
                        </ol>
                        <p style={{ fontSize: '12px', color: '#ff4d4f', marginTop: 8 }}>
                            {t('proxy.cert_warning')}
                        </p>
                    </div>
                )
            });
        } catch (err) {
            message.error(t('app.command_failed') + ": " + String(err));
        }
    };

    const handleApplyRules = async () => {
        try {
            // Apply Speed Limits
            const dlBytes = (dlLimit || 0) * 1024;
            const ulBytes = (ulLimit || 0) * 1024;
            await SetProxyLimit(ulBytes, dlBytes);

            // Apply Latency
            const ms = latency || 0;
            await SetProxyLatency(ms);

            message.success(t('proxy.apply'));
        } catch (err) {
            message.error(t('app.command_failed') + ": " + String(err));
        }
    };

    const handleClearRules = async () => {
        try {
            await SetProxyLimit(0, 0);
            await SetProxyLatency(0);
            setSpeedLimits(null, null);
            setLatency(null);
            message.success(t('proxy.off'));
        } catch (err) {
            message.error(t('app.command_failed') + ": " + String(err));
        }
    };

    const formatSpeed = (speed: number) => {
        if (!speed || speed === 0) return '0 B/s';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(speed) / Math.log(k));
        return parseFloat((speed / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i] + '/s';
    };

    const formatBytes = (bytes: number) => {
        if (!bytes || bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    };

    const formatBody = (body: string, contentType?: string) => {
        if (!body) return "";
        const trimmed = body.trim();

        // Check if it's JSON by content-type or by content
        const isJsonType = contentType?.includes('application/json');
        const looksLikeJson = (trimmed.startsWith('{') && trimmed.endsWith('}')) ||
                             (trimmed.startsWith('[') && trimmed.endsWith(']'));

        if (isJsonType || looksLikeJson) {
            try {
                const parsed = JSON.parse(trimmed);
                return JSON.stringify(parsed, null, 2);
            } catch (e) {
                // Not valid JSON
            }
        }
        return body;
    };

    // Generate cURL command from request
    const generateCurl = (log: StoreRequestLog): string => {
        const parts: string[] = ['curl'];

        // Method
        if (log.method !== 'GET') {
            parts.push(`-X ${log.method}`);
        }

        // Headers
        if (log.headers) {
            for (const [key, values] of Object.entries(log.headers)) {
                // Skip pseudo-headers and host (included in URL)
                if (key.startsWith(':') || key.toLowerCase() === 'host') continue;
                for (const value of values) {
                    parts.push(`-H '${key}: ${value.replace(/'/g, "'\\''")}'`);
                }
            }
        }

        // Body
        if (log.previewBody && ['POST', 'PUT', 'PATCH'].includes(log.method)) {
            const escaped = log.previewBody.replace(/'/g, "'\\''");
            parts.push(`-d '${escaped}'`);
        }

        // URL (quoted)
        parts.push(`'${log.url}'`);

        return parts.join(' \\\n  ');
    };

    const handleCopyCurl = (log: StoreRequestLog) => {
        const curl = generateCurl(log);
        navigator.clipboard.writeText(curl).then(() => {
            message.success(t('proxy.copied_curl'));
        }).catch(() => {
            message.error(t('proxy.copy_failed'));
        });
    };

    // Open resend modal with pre-filled data
    const handleOpenResendModal = (log: StoreRequestLog) => {
        const headersStr = log.headers
            ? Object.entries(log.headers).map(([k, v]) => `${k}: ${v.join(', ')}`).join('\n')
            : '';
        resendForm.setFieldsValue({
            method: log.method,
            url: log.url,
            headers: headersStr,
            body: log.previewBody || '',
        });
        setResendResponse(null);
        setResendModalOpen(true);
    };

    // Handle resend request
    const handleResend = async () => {
        try {
            const values = await resendForm.validateFields();
            setResendLoading(true);

            // Parse headers string to map
            const headersMap: Record<string, string> = {};
            if (values.headers) {
                values.headers.split('\n').forEach((line: string) => {
                    const idx = line.indexOf(':');
                    if (idx > 0) {
                        const key = line.substring(0, idx).trim();
                        const val = line.substring(idx + 1).trim();
                        if (key) headersMap[key] = val;
                    }
                });
            }

            const response = await ResendRequest(values.method, values.url, headersMap, values.body || '');
            setResendResponse(response);
            message.success(t('proxy.resend_success'));
        } catch (err: any) {
            message.error(t('proxy.resend_failed') + ': ' + String(err));
        } finally {
            setResendLoading(false);
        }
    };

    // Load mock rules
    const loadMockRules = async () => {
        try {
            const rules = await GetMockRules();
            setMockRules(rules || []);
        } catch (err) {
            console.error('Failed to load mock rules:', err);
        }
    };

    // Open mock rules modal
    const handleOpenMockModal = () => {
        loadMockRules();
        setMockModalOpen(true);
    };

    // Add or update mock rule
    const handleSaveMockRule = async () => {
        try {
            const values = await mockForm.validateFields();
            const rule = {
                id: editingMockRule?.id || '',
                urlPattern: values.urlPattern,
                method: values.method || '',
                statusCode: values.statusCode || 200,
                headers: { 'Content-Type': values.contentType || 'application/json' },
                body: values.body || '',
                delay: values.delay || 0,
                description: values.description || '',
                enabled: true,
            };

            if (editingMockRule) {
                // Update existing - for now just remove and re-add
                await RemoveMockRule(editingMockRule.id);
            }
            await AddMockRule(rule);

            message.success(editingMockRule ? t('proxy.mock_rule_updated') : t('proxy.mock_rule_added'));
            mockForm.resetFields();
            setEditingMockRule(null);
            loadMockRules();
        } catch (err: any) {
            message.error(String(err));
        }
    };

    // Delete mock rule
    const handleDeleteMockRule = async (id: string) => {
        try {
            await RemoveMockRule(id);
            message.success(t('proxy.mock_rule_deleted'));
            loadMockRules();
        } catch (err) {
            message.error(String(err));
        }
    };

    // Toggle mock rule
    const handleToggleMockRule = async (id: string, enabled: boolean) => {
        try {
            await ToggleMockRule(id, enabled);
            loadMockRules();
        } catch (err) {
            message.error(String(err));
        }
    };

    // Edit mock rule
    const startEditMockRule = (rule: any) => {
        setEditingMockRule(rule);
        mockForm.setFieldsValue({
            urlPattern: rule.urlPattern,
            method: rule.method,
            statusCode: rule.statusCode,
            contentType: rule.headers?.['Content-Type'] || 'application/json',
            body: rule.body,
            delay: rule.delay,
            description: rule.description,
        });
    };

    // Create mock from request
    const createMockFromRequest = (log: StoreRequestLog) => {
        // Extract URL pattern (use path with wildcard for query params)
        let urlPattern = log.url;
        try {
            const urlObj = new URL(log.url);
            urlPattern = `*${urlObj.pathname}*`;
        } catch (e) {
            // Use full URL as pattern if parsing fails
            urlPattern = `*${log.url.split('?')[0]}*`;
        }

        setEditingMockRule(null);
        mockForm.resetFields();
        mockForm.setFieldsValue({
            urlPattern,
            method: log.method,
            statusCode: log.statusCode || 200,
            contentType: log.contentType?.split(';')[0] || 'application/json',
            body: log.respBody || '',
            delay: 0,
            description: `Mock for ${log.method} ${urlPattern}`,
        });
        setMockModalOpen(true);
    };

    const filteredLogs = logs.filter(log => {
        // Filter by type (ALL, HTTP, WS)
        if (filterType === "HTTP" && log.isWs) return false;
        if (filterType === "WS" && !log.isWs) return false;

        // Filter by search text
        if (searchText) {
            const lowerSearch = searchText.toLowerCase();
            return (
                log.url.toLowerCase().includes(lowerSearch) ||
                log.method.toLowerCase().includes(lowerSearch) ||
                String(log.statusCode || '').includes(lowerSearch)
            );
        }
        return true;
    });

    return (
        <div style={{ padding: '16px', height: '100%', display: 'flex', flexDirection: 'column', gap: '12px', overflow: 'hidden' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexShrink: 0 }}>
                <Space align="center" size="small">
                    <GlobalOutlined style={{ fontSize: '18px' }} />
                    <Title level={4} style={{ margin: 0 }}>{t('proxy.title')}</Title>
                </Space>
                <DeviceSelector />
            </div>

            <Card size="small" styles={{ body: { padding: '12px' } }}>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
                    {/* Row 1: Proxy & Device Settings */}
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <Space split={<Divider type="vertical" />}>
                            <Space size="small">
                                <Text strong>{t('proxy.server')}:</Text>
                                <Tag color={isRunning ? "success" : "default"} style={{ marginRight: 0 }}>{isRunning ? t('proxy.on') : t('proxy.off')}</Tag>
                            </Space>
                            <Space size="small">
                                <Text type="secondary">{t('proxy.ip')}:</Text>
                                <Text copyable style={{ fontSize: '13px' }}>{localIP || "Unknown"}</Text>
                            </Space>
                            <Space size="small">
                                <Text type="secondary">{t('proxy.port')}:</Text>
                                <InputNumber size="small" value={port} onChange={(v) => setPort(v || 8080)} disabled={isRunning} style={{ width: 60 }} />
                            </Space>
                            <Space size="small">
                                <Tooltip title={t('proxy.ws_tooltip')}>
                                    <Space size={4}>
                                        <ApiOutlined style={{ color: wsEnabled ? '#1890ff' : undefined }} />
                                        <Switch size="small" checked={wsEnabled} onChange={handleWSToggle} />
                                    </Space>
                                </Tooltip>
                            </Space>
                            <Space size="small">
                                <Tooltip title={t('proxy.mitm_tooltip')}>
                                    <Space size={4}>
                                        <SafetyCertificateOutlined style={{ color: mitmEnabled ? '#faad14' : undefined }} />
                                        <Switch size="small" checked={mitmEnabled} onChange={handleMITMToggle} />
                                    </Space>
                                </Tooltip>
                            </Space>
                            {mitmEnabled && (
                                <Space size="small">
                                    <Tooltip title={
                                        certTrustStatus === 'trusted' ? t('proxy.cert_trusted') :
                                        certTrustStatus === 'not_trusted' ? t('proxy.cert_not_trusted') :
                                        certTrustStatus === 'pending' ? t('proxy.cert_pending') :
                                        certTrustStatus === 'checking' ? t('proxy.cert_checking') :
                                        t('proxy.cert_install_hint')
                                    }>
                                        <Button size="small" icon={<DownloadOutlined />} onClick={handleInstallCert}>
                                            Cert
                                        </Button>
                                    </Tooltip>
                                    {certTrustStatus && certTrustStatus !== 'no_proxy' && certTrustStatus !== 'unknown' && (
                                        <Tag
                                            color={
                                                certTrustStatus === 'trusted' ? 'success' :
                                                certTrustStatus === 'not_trusted' ? 'error' :
                                                certTrustStatus === 'pending' ? 'warning' :
                                                certTrustStatus === 'checking' ? 'processing' :
                                                'default'
                                            }
                                            style={{ marginRight: 0 }}
                                        >
                                            {certTrustStatus === 'trusted' ? '✓' :
                                             certTrustStatus === 'not_trusted' ? '✗' :
                                             certTrustStatus === 'pending' ? '?' :
                                             certTrustStatus === 'checking' ? '...' : '?'}
                                        </Tag>
                                    )}
                                    <Button size="small" icon={<SettingOutlined />} onClick={() => setBypassModalOpen(true)}>
                                        Rules
                                    </Button>
                                </Space>
                            )}
                            <Tooltip title={t('proxy.mock_rules')}>
                                <Button size="small" icon={<BlockOutlined />} onClick={handleOpenMockModal}>
                                    Mock
                                </Button>
                            </Tooltip>

                            <Button
                                type="primary"
                                size="small"
                                danger={isRunning}
                                icon={isRunning ? <PoweroffOutlined /> : <PlayCircleOutlined />}
                                onClick={isRunning ? handleStop : handleStart}
                                style={{ height: 32, padding: '0 20px', borderRadius: 16 }}
                            >
                                {isRunning ? t('proxy.stop_capture') : t('proxy.start_capture')}
                            </Button>
                        </Space>
                    </div>

                    {/* Row 2: Network Monitor & Limit (conditionally rendered) */}
                    {selectedDevice && (
                        <>
                            <Divider style={{ margin: '0' }} />

                            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                                {/* Left: Monitor */}
                                <Space size="middle">
                                    <Space size="large">
                                        <Space size={4}>
                                            <ArrowDownOutlined style={{ fontSize: '12px', color: '#52c41a' }} />
                                            <Text type="secondary" style={{ fontSize: '10px' }}>RX</Text>
                                            <Text style={{ color: '#52c41a', fontFamily: 'monospace', fontWeight: 600, fontSize: '12px', minWidth: '70px' }}>{formatSpeed(netStats.rxSpeed)}</Text>
                                        </Space>
                                        <Space size={4}>
                                            <ArrowUpOutlined style={{ fontSize: '12px', color: '#1890ff' }} />
                                            <Text type="secondary" style={{ fontSize: '10px' }}>TX</Text>
                                            <Text style={{ color: '#1890ff', fontFamily: 'monospace', fontWeight: 600, fontSize: '12px', minWidth: '70px' }}>{formatSpeed(netStats.txSpeed)}</Text>
                                        </Space>
                                    </Space>
                                </Space>

                                <Divider type="vertical" style={{ height: 32 }} />

                                {/* Right: Controls */}
                                <Space size="small" align="center">
                                    <Tag color="orange" style={{ marginRight: 0 }}>{t('proxy.limit')}</Tag>

                                    {/* Bandwidth */}
                                    <InputNumber
                                        size="small"
                                        prefix={<ArrowDownOutlined style={{ fontSize: 10, color: '#aaa' }} />}
                                        suffix="K"
                                        placeholder="DL"
                                        min={0}
                                        value={dlLimit}
                                        onChange={(val) => setSpeedLimits(val, ulLimit)}
                                        style={{ width: 110 }}
                                        title={t('proxy.dl_limit')}
                                    />
                                    <InputNumber
                                        size="small"
                                        prefix={<ArrowUpOutlined style={{ fontSize: 10, color: '#aaa' }} />}
                                        suffix="K"
                                        placeholder="UL"
                                        min={0}
                                        value={ulLimit}
                                        onChange={(val) => setSpeedLimits(dlLimit, val)}
                                        style={{ width: 110 }}
                                        title={t('proxy.ul_limit')}
                                    />

                                    {/* Latency */}
                                    <InputNumber
                                        size="small"
                                        prefix={<HourglassOutlined style={{ fontSize: 10, color: '#aaa' }} />}
                                        suffix="ms"
                                        placeholder="Delay"
                                        min={0}
                                        value={latency}
                                        onChange={setLatency}
                                        style={{ width: 120 }}
                                        title={t('proxy.latency')}
                                    />

                                    {/* Actions */}
                                    <Button type="primary" size="small" onClick={handleApplyRules}>{t('proxy.apply')}</Button>
                                    <Button size="small" onClick={handleClearRules} icon={<DeleteOutlined />} title={t('common.clear')} />
                                </Space>
                            </div>
                        </>
                    )}
                </div>
            </Card>

            {/* Request List and Detail Panel Container */}
            <div style={{ flex: 1, display: 'flex', gap: 12, overflow: 'hidden' }}>
                <Card
                    style={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden', minWidth: 0 }}
                    styles={{ body: { flex: 1, overflow: 'hidden', padding: 0, display: 'flex', flexDirection: 'column' } }}
                    size="small"
                    title={
                        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                            <Radio.Group size="small" value={filterType} onChange={e => setFilterType(e.target.value)} buttonStyle="solid">
                                <Radio.Button value="ALL">{t('proxy.filter_all')}</Radio.Button>
                                <Radio.Button value="HTTP">{t('proxy.filter_http')}</Radio.Button>
                                <Radio.Button value="WS">{t('proxy.filter_ws')}</Radio.Button>
                            </Radio.Group>
                            <Input
                                placeholder={t('proxy.search_placeholder')}
                                size="small"
                                allowClear
                                style={{ maxWidth: 400 }}
                                value={searchText}
                                onChange={e => setSearchText(e.target.value)}
                            />
                            <Button size="small" type="link" onClick={() => clearLogs()} icon={<DeleteOutlined />} style={{ padding: 0 }}>{t('proxy.clear_logs')}</Button>
                        </div>
                    }
                >
                {/* Virtual Table Header - Fixed widths */}
                <div style={{ display: 'grid', gridTemplateColumns: '80px 70px 80px 1fr 80px 80px', padding: '8px 12px', background: token.colorFillAlter, borderBottom: `1px solid ${token.colorBorderSecondary}`, fontWeight: 'bold', fontSize: '12px', color: token.colorTextSecondary }}>
                    <div>{t('proxy.col_time')}</div>
                    <div>{t('proxy.col_method')}</div>
                    <div>{t('proxy.col_stat')}</div>
                    <div>{t('proxy.col_url')}</div>
                    <div>{t('proxy.col_type')}</div>
                    <div>{t('proxy.col_size')}</div>
                </div>

                <VirtualList
                    dataSource={filteredLogs}
                    rowKey="id"
                    rowHeight={35}
                    overscan={20}
                    selectedKey={selectedLog?.id}
                    onItemClick={selectLog}
                    showBorder={false}
                    style={{ flex: 1 }}
                    renderItem={(record, index, isSelected) => (
                        <div
                            style={{
                                display: 'grid',
                                gridTemplateColumns: '80px 70px 80px 1fr 80px 80px',
                                padding: '6px 12px',
                                fontSize: '12px',
                                alignItems: 'center',
                                height: '100%',
                                background: isSelected 
                                    ? token.colorPrimaryBg 
                                    : index % 2 === 0 
                                        ? token.colorBgContainer 
                                        : token.colorFillAlter,
                                borderBottom: `1px solid ${token.colorBorderSecondary}`,
                            }}
                        >
                            <div style={{ whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', color: '#888' }}>{record.time}</div>
                            <div>
                                {record.method === 'CONNECT' ? (
                                    <Tag color="purple" style={{ marginRight: 0, transform: 'scale(0.8)', transformOrigin: 'left center' }}>TUNNEL</Tag>
                                ) : record.isWs ? (
                                    <Tag color="cyan" style={{ marginRight: 0, transform: 'scale(0.8)', transformOrigin: 'left center' }}>WS</Tag>
                                ) : (
                                    <Tag color={record.statusCode && record.statusCode >= 400 ? 'red' : record.method === 'GET' ? 'green' : record.method === 'POST' ? 'blue' : 'default'} style={{ marginRight: 0, transform: 'scale(0.8)', transformOrigin: 'left center' }}>{record.method}</Tag>
                                )}
                            </div>
                            <div>
                                <Tag color={record.statusCode && record.statusCode >= 400 ? 'red' : record.statusCode && record.statusCode >= 300 ? 'orange' : 'success'} style={{ marginRight: 0, transform: 'scale(0.8)', transformOrigin: 'left center' }}>{record.statusCode || '-'}</Tag>
                                {record.mocked && <Tag color="magenta" style={{ marginLeft: 2, transform: 'scale(0.7)', transformOrigin: 'left center' }}>M</Tag>}
                            </div>
                            <div title={record.url} style={{ whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', color: token.colorLink }}>
                                {record.url}
                                {record.isHttps && <LockOutlined style={{ fontSize: '10px', marginLeft: 4, color: '#52c41a' }} />}
                            </div>
                            <div style={{ color: '#888', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>{record.contentType?.split(';')[0].split('/')[1] || '-'}</div>
                            <div style={{ fontFamily: 'monospace', color: '#666' }}>{formatBytes(record.bodySize || 0)}</div>
                        </div>
                    )}
                />
            </Card>

                {/* Detail Panel */}
                {selectedLog && (
                    <Card
                        size="small"
                        style={{ width: '50%', minWidth: 400, flexShrink: 0, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}
                        styles={{ body: { flex: 1, overflow: 'auto', padding: 16 } }}
                        title={
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                <Text strong>{t('proxy.details')}</Text>
                                <Space size="small">
                                    <Button
                                        type="text"
                                        size="small"
                                        icon={<CopyOutlined />}
                                        onClick={() => handleCopyCurl(selectedLog)}
                                    >
                                        cURL
                                    </Button>
                                    <Button
                                        type="text"
                                        size="small"
                                        icon={<SendOutlined />}
                                        onClick={() => handleOpenResendModal(selectedLog)}
                                    >
                                        {t('proxy.resend')}
                                    </Button>
                                    <Button
                                        type="text"
                                        size="small"
                                        icon={<BlockOutlined />}
                                        onClick={() => createMockFromRequest(selectedLog)}
                                    >
                                        Mock
                                    </Button>
                                    <Button
                                        type="text"
                                        size="small"
                                        icon={<CloseOutlined />}
                                        onClick={() => selectLog(null)}
                                    />
                                </Space>
                            </div>
                        }
                    >
                        <Space direction="vertical" size="middle" style={{ width: '100%' }}>
                            <div style={{ wordBreak: 'break-all', fontFamily: 'monospace', background: token.colorFillTertiary, padding: 8, borderRadius: 4, display: 'flex', alignItems: 'flex-start', gap: 8 }}>
                                <Tag color={selectedLog.method === 'GET' ? 'green' : 'blue'} style={{ flexShrink: 0 }}>{selectedLog.method}</Tag>
                                <Text copyable={{ text: selectedLog.url }} style={{ fontFamily: 'monospace', fontSize: '13px', flex: 1, wordBreak: 'break-all' }}>{selectedLog.url}</Text>
                            </div>

                            {(selectedLog.method === 'CONNECT') ? (
                                <div style={{ textAlign: 'center', padding: '40px 20px', background: token.colorFillAlter, borderRadius: 8 }}>
                                    <LockOutlined style={{ fontSize: 48, color: token.colorTextDisabled, marginBottom: 16 }} />
                                    <br />
                                    <Text type="secondary" style={{ fontStyle: 'italic' }}>
                                        {t('proxy.tunnel_info')}
                                    </Text>
                                </div>
                            ) : (() => {
                                let queryParams: [string, string][] = [];
                                try {
                                    const urlObj = new URL(selectedLog.url);
                                    queryParams = Array.from(urlObj.searchParams.entries());
                                } catch (e) {
                                    if (selectedLog.url.includes('?')) {
                                        const search = selectedLog.url.split('?')[1];
                                        queryParams = search.split('&').map(p => {
                                            const [k, v] = p.split('=');
                                            return [decodeURIComponent(k), decodeURIComponent(v || '')];
                                        });
                                    }
                                }
                                return (
                                    <Tabs defaultActiveKey="request" size="small" items={[
                                        {
                                            key: 'request',
                                            label: t('proxy.request'),
                                            children: (
                                                <Space direction="vertical" size="middle" style={{ width: '100%', paddingTop: 8 }}>
                                                    {queryParams.length > 0 && (
                                                        <div>
                                                            <div style={{ marginBottom: 8, fontSize: 12, fontWeight: 'bold', color: token.colorTextSecondary, borderBottom: `1px solid ${token.colorBorderSecondary}`, paddingBottom: 4 }}>{t('proxy.query_params')}</div>
                                                            <div style={{ display: 'grid', gridTemplateColumns: 'auto 1fr', gap: '4px 12px' }}>
                                                                {queryParams.map(([k, v], idx) => (
                                                                    <React.Fragment key={idx}>
                                                                        <Text style={{ fontSize: '12px', color: token.colorTextSecondary, textAlign: 'right', fontWeight: 500 }}>{k}:</Text>
                                                                        <Text copyable={{ text: v }} style={{ fontSize: '12px', fontFamily: 'monospace', wordBreak: 'break-all', color: token.colorLink }}>{v}</Text>
                                                                    </React.Fragment>
                                                                ))}
                                                            </div>
                                                        </div>
                                                    )}
                                                    {selectedLog.headers && Object.keys(selectedLog.headers).length > 0 && (
                                                        <div>
                                                            <div style={{ marginBottom: 8, fontSize: 12, fontWeight: 'bold', color: token.colorTextSecondary, borderBottom: `1px solid ${token.colorBorderSecondary}`, paddingBottom: 4 }}>{t('proxy.headers')}</div>
                                                            <div style={{ display: 'grid', gridTemplateColumns: 'auto 1fr', gap: '4px 12px' }}>
                                                                {Object.entries(selectedLog.headers).map(([k, v]) => (
                                                                    <React.Fragment key={k}>
                                                                        <Text style={{ fontSize: '12px', color: token.colorTextSecondary, textAlign: 'right', fontWeight: 500 }}>{k}:</Text>
                                                                        <Text copyable={{ text: (v as string[]).join(', ') }} style={{ fontSize: '12px', fontFamily: 'monospace', wordBreak: 'break-all' }}>{(v as string[]).join(', ')}</Text>
                                                                    </React.Fragment>
                                                                ))}
                                                            </div>
                                                        </div>
                                                    )}
                                                    {selectedLog.previewBody && (
                                                        <div>
                                                            <div style={{ marginBottom: 8, fontSize: 12, fontWeight: 'bold', color: token.colorTextSecondary, borderBottom: `1px solid ${token.colorBorderSecondary}`, paddingBottom: 4 }}>{t('proxy.body')}</div>
                                                            <div style={{ position: 'relative' }}>
                                                                <pre style={{
                                                                    padding: '12px',
                                                                    background: token.colorFillAlter,
                                                                    border: `1px solid ${token.colorBorderSecondary}`,
                                                                    borderRadius: '4px',
                                                                    fontFamily: 'monospace',
                                                                    fontSize: '12px',
                                                                    whiteSpace: 'pre-wrap',
                                                                    overflow: 'auto',
                                                                    maxHeight: '300px',
                                                                    margin: 0,
                                                                    wordBreak: 'break-all'
                                                                }}>
                                                                    {formatBody(selectedLog.previewBody)}
                                                                </pre>
                                                                <div style={{ position: 'absolute', top: 8, right: 8 }}>
                                                                    <Text copyable={{ text: formatBody(selectedLog.previewBody) }} />
                                                                </div>
                                                            </div>
                                                        </div>
                                                    )}
                                                    {!selectedLog.previewBody && (!selectedLog.headers || Object.keys(selectedLog.headers).length === 0) && (
                                                        <Text type="secondary" style={{ display: 'block', textAlign: 'center', margin: '40px 0' }}>{t('proxy.no_req_data')}</Text>
                                                    )}
                                                </Space>
                                            )
                                        },
                                        {
                                            key: 'response',
                                            label: `${t('proxy.response')} ${selectedLog.statusCode ? '(' + selectedLog.statusCode + ')' : ''}`,
                                            children: selectedLog.statusCode ? (
                                                <Space direction="vertical" size="middle" style={{ width: '100%', paddingTop: 8 }}>
                                                    <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                                                        <Tag color={selectedLog.statusCode >= 400 ? 'error' : 'success'} style={{ fontSize: 14, padding: '2px 8px' }}>
                                                            {selectedLog.statusCode}
                                                        </Tag>
                                                        {selectedLog.mocked && <Tag color="magenta">Mocked</Tag>}
                                                        {selectedLog.contentType && <Tag>{selectedLog.contentType.split(';')[0]}</Tag>}
                                                    </div>
                                                    {selectedLog.respHeaders && Object.keys(selectedLog.respHeaders).length > 0 && (
                                                        <div>
                                                            <div style={{ marginBottom: 8, fontSize: 12, fontWeight: 'bold', color: token.colorTextSecondary, borderBottom: `1px solid ${token.colorBorderSecondary}`, paddingBottom: 4 }}>{t('proxy.headers')}</div>
                                                            <div style={{ display: 'grid', gridTemplateColumns: 'auto 1fr', gap: '4px 12px' }}>
                                                                {Object.entries(selectedLog.respHeaders).map(([k, v]) => (
                                                                    <React.Fragment key={k}>
                                                                        <Text style={{ fontSize: '12px', color: token.colorTextSecondary, textAlign: 'right', fontWeight: 500 }}>{k}:</Text>
                                                                        <Text copyable={{ text: (v as string[]).join(', ') }} style={{ fontSize: '12px', fontFamily: 'monospace', wordBreak: 'break-all' }}>{(v as string[]).join(', ')}</Text>
                                                                    </React.Fragment>
                                                                ))}
                                                            </div>
                                                        </div>
                                                    )}
                                                    {selectedLog.respBody && (
                                                        <div>
                                                            <div style={{ marginBottom: 8, fontSize: 12, fontWeight: 'bold', color: token.colorTextSecondary, borderBottom: `1px solid ${token.colorBorderSecondary}`, paddingBottom: 4 }}>{t('proxy.body')}</div>
                                                            <div style={{ position: 'relative' }}>
                                                                <pre style={{
                                                                    padding: '12px',
                                                                    background: token.colorFillAlter,
                                                                    border: `1px solid ${token.colorBorderSecondary}`,
                                                                    borderRadius: '4px',
                                                                    fontFamily: 'monospace',
                                                                    fontSize: '12px',
                                                                    whiteSpace: 'pre-wrap',
                                                                    overflow: 'auto',
                                                                    maxHeight: '400px',
                                                                    margin: 0,
                                                                    wordBreak: 'break-all'
                                                                }}>
                                                                    {formatBody(selectedLog.respBody)}
                                                                </pre>
                                                                <div style={{ position: 'absolute', top: 8, right: 8 }}>
                                                                    <Text copyable={{ text: formatBody(selectedLog.respBody) }} />
                                                                </div>
                                                            </div>
                                                        </div>
                                                    )}
                                                    {!selectedLog.respBody && (!selectedLog.respHeaders || Object.keys(selectedLog.respHeaders).length === 0) && (
                                                        <Text type="secondary" style={{ display: 'block', textAlign: 'center', margin: '40px 0' }}>{t('proxy.no_resp_data')}</Text>
                                                    )}
                                                </Space>
                                            ) : (
                                                <div style={{ padding: '60px 20px', textAlign: 'center' }}>
                                                    <HourglassOutlined style={{ fontSize: 48, color: token.colorTextDisabled, marginBottom: 16 }} />
                                                    <br />
                                                    <Text type="secondary" italic>{t('proxy.waiting_for_response')}</Text>
                                                </div>
                                            )
                                        }
                                    ]} />
                                );
                            })()}
                        </Space>
                    </Card>
                )}
            </div>

            <Modal
                title={
                    <Space>
                        <SettingOutlined />
                        <span>{t('proxy.rules_title')}</span>
                    </Space>
                }
                open={isBypassModalOpen}
                onCancel={() => setBypassModalOpen(false)}
                footer={[
                    <Button key="close" type="primary" onClick={() => setBypassModalOpen(false)}>{t('common.close')}</Button>
                ]}
            >
                <div style={{ marginBottom: 16 }}>
                    <Text type="secondary">
                        {t('proxy.rules_desc')}
                    </Text>
                </div>

                <Space.Compact style={{ width: '100%', marginBottom: 16 }}>
                    <Input
                        placeholder={t('proxy.add_keyword')}
                        value={newPattern}
                        onChange={e => setNewPattern(e.target.value)}
                        onPressEnter={() => {
                            if (newPattern && !bypassPatterns.includes(newPattern)) {
                                const next = [...bypassPatterns, newPattern];
                                addBypassPattern(newPattern.trim());
                                SetMITMBypassPatterns(next);
                                setNewPattern("");
                            }
                        }}
                    />
                    <Button type="primary" onClick={() => {
                        if (newPattern && !bypassPatterns.includes(newPattern)) {
                            const next = [...bypassPatterns, newPattern];
                            addBypassPattern(newPattern.trim());
                            SetMITMBypassPatterns(next);
                            setNewPattern("");
                        }
                    }}> {t('proxy.add')} </Button>
                </Space.Compact>

                <div style={{ minHeight: 100, padding: 8, border: '1px solid #f0f0f0', borderRadius: 4, background: '#fafafa' }}>
                    <Space wrap>
                        {bypassPatterns.map(pat => (
                            <Tag
                                key={pat}
                                closable
                                onClose={async () => {
                                    removeBypassPattern(pat);
                                    const next = bypassPatterns.filter(p => p !== pat);
                                    await SetMITMBypassPatterns(next);
                                }}
                            >
                                {pat}
                            </Tag>
                        ))}
                        {bypassPatterns.length === 0 && <Text type="secondary">{t('proxy.no_rules')}</Text>}
                    </Space>
                </div>
            </Modal>

            {/* Resend Request Modal */}
            <Modal
                title={t('proxy.resend_request')}
                open={resendModalOpen}
                onCancel={() => setResendModalOpen(false)}
                width="50%"
                styles={{ body: { height: '70vh', display: 'flex', flexDirection: 'column', overflow: 'hidden' } }}
                footer={[
                    <Button key="cancel" onClick={() => setResendModalOpen(false)}>{t('common.cancel')}</Button>,
                    <Button key="send" type="primary" loading={resendLoading} onClick={handleResend} icon={<SendOutlined />}>
                        {t('proxy.resend')}
                    </Button>
                ]}
            >
                <Form form={resendForm} layout="vertical" style={{ marginTop: 16, flexShrink: 0 }}>
                    <Space.Compact style={{ width: '100%' }}>
                        <Form.Item name="method" noStyle>
                            <Input style={{ width: 100 }} />
                        </Form.Item>
                        <Form.Item name="url" noStyle rules={[{ required: true }]}>
                            <Input style={{ flex: 1 }} placeholder="URL" />
                        </Form.Item>
                    </Space.Compact>
                    <Form.Item name="headers" label={t('proxy.headers')} style={{ marginTop: 16 }}>
                        <Input.TextArea rows={3} placeholder="Header-Name: value" style={{ fontFamily: 'monospace', fontSize: 12 }} />
                    </Form.Item>
                    <Form.Item name="body" label={t('proxy.body')}>
                        <Input.TextArea rows={4} placeholder="Request body" style={{ fontFamily: 'monospace', fontSize: 12 }} />
                    </Form.Item>
                </Form>

                {resendResponse && (
                    <div style={{ flex: 1, marginTop: 16, padding: 12, background: token.colorFillAlter, borderRadius: 8, display: 'flex', flexDirection: 'column', overflow: 'hidden', minHeight: 0 }}>
                        <Space style={{ marginBottom: 8, flexShrink: 0 }}>
                            <Tag color={resendResponse.statusCode >= 400 ? 'error' : 'success'}>{resendResponse.statusCode}</Tag>
                            {resendResponse.mocked && <Tag color="magenta">Mocked</Tag>}
                            <Text type="secondary">{resendResponse.duration}ms</Text>
                            <Text type="secondary">{formatBytes(resendResponse.bodySize)}</Text>
                            {resendResponse.contentType && <Text type="secondary">{resendResponse.contentType.split(';')[0]}</Text>}
                        </Space>
                        <pre style={{ flex: 1, overflow: 'auto', padding: 8, margin: 0, background: token.colorBgContainer, borderRadius: 4, fontFamily: 'monospace', fontSize: 12, whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
                            {formatBody(resendResponse.body, resendResponse.contentType)}
                        </pre>
                    </div>
                )}
            </Modal>

            {/* Mock Rules Modal */}
            <Modal
                title={t('proxy.mock_rules')}
                open={mockModalOpen}
                onCancel={() => { setMockModalOpen(false); setEditingMockRule(null); mockForm.resetFields(); }}
                width={800}
                footer={null}
            >
                <div style={{ marginBottom: 16 }}>
                    <Form form={mockForm} layout="vertical" size="small">
                        <Space wrap style={{ width: '100%' }}>
                            <Form.Item name="urlPattern" label={t('proxy.url_pattern')} rules={[{ required: true }]} style={{ marginBottom: 8, minWidth: 250 }}>
                                <Input placeholder="*/api/*" />
                            </Form.Item>
                            <Form.Item name="method" label={t('proxy.col_method')} style={{ marginBottom: 8, width: 100 }}>
                                <Input placeholder="GET" />
                            </Form.Item>
                            <Form.Item name="statusCode" label={t('proxy.status_code')} style={{ marginBottom: 8, width: 80 }}>
                                <InputNumber min={100} max={599} placeholder="200" />
                            </Form.Item>
                            <Form.Item name="delay" label={t('proxy.delay_ms')} style={{ marginBottom: 8, width: 80 }}>
                                <InputNumber min={0} placeholder="0" />
                            </Form.Item>
                        </Space>
                        <Form.Item name="contentType" label="Content-Type" style={{ marginBottom: 8 }}>
                            <Input placeholder="application/json" />
                        </Form.Item>
                        <Form.Item name="body" label={t('proxy.response_body')} style={{ marginBottom: 8 }}>
                            <Input.TextArea rows={4} placeholder='{"success": true}' style={{ fontFamily: 'monospace', fontSize: 12 }} />
                        </Form.Item>
                        <Form.Item name="description" label={t('proxy.description')} style={{ marginBottom: 8 }}>
                            <Input placeholder={t('proxy.description')} />
                        </Form.Item>
                        <Button type="primary" icon={<PlusOutlined />} onClick={handleSaveMockRule}>
                            {editingMockRule ? t('common.save') : t('proxy.add')}
                        </Button>
                        {editingMockRule && (
                            <Button style={{ marginLeft: 8 }} onClick={() => { setEditingMockRule(null); mockForm.resetFields(); }}>
                                {t('common.cancel')}
                            </Button>
                        )}
                    </Form>
                </div>

                <Divider style={{ margin: '12px 0' }} />

                <div style={{ maxHeight: 300, overflow: 'auto' }}>
                    {mockRules.length === 0 ? (
                        <Text type="secondary" style={{ display: 'block', textAlign: 'center', padding: 20 }}>{t('proxy.no_mock_rules')}</Text>
                    ) : (
                        <Table
                            dataSource={mockRules}
                            rowKey="id"
                            size="small"
                            pagination={false}
                            columns={[
                                { title: t('proxy.url_pattern'), dataIndex: 'urlPattern', ellipsis: true },
                                { title: t('proxy.col_method'), dataIndex: 'method', width: 80, render: (v: string) => v || '*' },
                                { title: t('proxy.status_code'), dataIndex: 'statusCode', width: 80 },
                                { title: t('proxy.delay_ms'), dataIndex: 'delay', width: 80, render: (v: number) => v ? `${v}ms` : '-' },
                                {
                                    title: t('proxy.mock_enabled'),
                                    dataIndex: 'enabled',
                                    width: 80,
                                    render: (v: boolean, record: any) => (
                                        <Switch size="small" checked={v} onChange={(checked) => handleToggleMockRule(record.id, checked)} />
                                    )
                                },
                                {
                                    title: '',
                                    width: 100,
                                    render: (_: any, record: any) => (
                                        <Space size="small">
                                            <Button type="text" size="small" icon={<EditOutlined />} onClick={() => startEditMockRule(record)} />
                                            <Popconfirm title={t('common.delete') + '?'} onConfirm={() => handleDeleteMockRule(record.id)}>
                                                <Button type="text" size="small" danger icon={<DeleteOutlined />} />
                                            </Popconfirm>
                                        </Space>
                                    )
                                }
                            ]}
                        />
                    )}
                </div>
            </Modal>

        </div>
    );
};

export default ProxyView;
