import React, { useEffect, useCallback, useMemo } from 'react';
import { Card, Button, InputNumber, Space, Typography, Tag, Divider, Switch, Tooltip, Radio, Input, Tabs, theme, Form, Table, Popconfirm, Popover, Spin, App, Modal, Select, AutoComplete } from 'antd';
import { PoweroffOutlined, PlayCircleOutlined, DeleteOutlined, SettingOutlined, LockOutlined, GlobalOutlined, ArrowUpOutlined, ArrowDownOutlined, ApiOutlined, SafetyCertificateOutlined, DownloadOutlined, HourglassOutlined, CopyOutlined, BlockOutlined, SendOutlined, CloseOutlined, PlusOutlined, EditOutlined, CodeOutlined, CloudDownloadOutlined, FolderOpenOutlined, LoadingOutlined, MinusCircleOutlined } from '@ant-design/icons';
import VirtualList from './VirtualList';
import DeviceSelector from './DeviceSelector';
import JsonViewer from './JsonViewer';
import JsonEditor from './JsonEditor';
import { useDeviceStore, useProxyStore, RequestLog as StoreRequestLog } from '../stores';
// @ts-ignore
import { StartProxy, StopProxy, GetProxyStatus, GetLocalIP, RunAdbCommand, StartNetworkMonitor, StopNetworkMonitor, SetProxyLimit, SetProxyWSEnabled, SetProxyMITM, InstallProxyCert, SetProxyLatency, SetMITMBypassPatterns, SetProxyDevice, ResendRequest, AddMockRule, RemoveMockRule, GetMockRules, ToggleMockRule, CheckCertTrust, SetupProxyForDevice, CleanupProxyForDevice, GetProtoFiles, AddProtoFile, UpdateProtoFile, RemoveProtoFile, GetProtoMappings, AddProtoMapping, UpdateProtoMapping, RemoveProtoMapping, GetProtoMessageTypes, LoadProtoFromURL, LoadProtoFromDisk } from '../../wailsjs/go/main/App';
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
    isProtobuf?: boolean;
    isReqProtobuf?: boolean;
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
    const [protoFileForm] = Form.useForm();
    const [protoMappingForm] = Form.useForm();

    // Watch conditions array to reactively disable/hide fields based on type & operator
    const watchedConditions: Array<{ type?: string; operator?: string }> | undefined = Form.useWatch('conditions', mockForm);

    // Additional proxy store state
    const {
        resendModalOpen,
        resendLoading,
        resendResponse,
        mockListModalOpen,
        mockEditModalOpen,
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
        setMockRules,
        setEditingMockRule,
        openMockListModal,
        closeMockListModal,
        openMockEditModal,
        closeMockEditModal,
        setMockEditModalOpen,
        pendingMockData,
        setPendingMockData,
        mockConditionHints,
        setMockConditionHints,
        setCertTrustStatus,
        setIsAIParsing,
        setAiSearchText,
        setAiPopoverOpen,
        protoFiles,
        protoMappings,
        protoMessageTypes,
        protoListModalOpen,
        protoEditFileModalOpen,
        protoEditMappingModalOpen,
        editingProtoFile,
        editingProtoMapping,
        setProtoFiles,
        setProtoMappings,
        setProtoMessageTypes,
        openProtoListModal,
        closeProtoListModal,
        openProtoEditFileModal,
        closeProtoEditFileModal,
        openProtoEditMappingModal,
        closeProtoEditMappingModal,
        protoImportLoading,
        setProtoImportLoading,
        protoImportURLModalOpen,
        protoImportURL,
        openProtoImportURLModal,
        closeProtoImportURLModal,
        setProtoImportURL,
    } = useProxyStore();

    // Watch hidden _conditionHints field stored in the form (survives HMR / store resets)
    const formStoredHints: string | undefined = Form.useWatch('_conditionHints', mockForm);

    // Compute condition key presets.
    // Priority: 1) store hints (set when creating mock from a request)
    //           2) form-persisted hints (survives HMR)
    //           3) scan captured logs
    const conditionKeyPresets = useMemo(() => {
        // Priority 1: store-level hints (freshly set from createMockFromRequest)
        if (mockConditionHints && mockConditionHints.headers.length > 0) {
            return mockConditionHints;
        }
        // Priority 2: form-persisted hints (survives store resets / HMR)
        if (formStoredHints) {
            try {
                const parsed = JSON.parse(formStoredHints);
                if (parsed && parsed.headers && parsed.headers.length > 0) {
                    return parsed as { headers: Array<{ key: string; value: string }>; queryParams: Array<{ key: string; value: string }> };
                }
            } catch (_) { /* ignore */ }
        }
        // Priority 3: derive from all captured logs
        const headerMap = new Map<string, string>();
        const queryMap = new Map<string, string>();
        const cap = Math.min(logs.length, 200);
        for (let i = 0; i < cap; i++) {
            const log = logs[i];
            const hd = log.headers || (log as any).requestHeaders;
            if (hd) {
                for (const [k, v] of Object.entries(hd)) {
                    if (!headerMap.has(k)) {
                        headerMap.set(k, Array.isArray(v) ? (v as string[]).join(', ') : String(v || ''));
                    }
                }
            }
            try {
                const u = new URL(log.url);
                u.searchParams.forEach((val, key) => {
                    if (!queryMap.has(key)) queryMap.set(key, val);
                });
            } catch (_) { /* ignore */ }
        }
        const headers = Array.from(headerMap.entries())
            .map(([key, value]) => ({ key, value }))
            .sort((a, b) => a.key.localeCompare(b.key));
        const queryParams = Array.from(queryMap.entries())
            .map(([key, value]) => ({ key, value }))
            .sort((a, b) => a.key.localeCompare(b.key));
        return { headers, queryParams };
    }, [mockConditionHints, formStoredHints, logs.length]);

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

    // Consume pendingMockData from another view (e.g. EventTimeline mock button)
    useEffect(() => {
        if (pendingMockData) {
            setEditingMockRule(null);
            mockForm.resetFields();
            mockForm.setFieldsValue({
                urlPattern: pendingMockData.urlPattern,
                method: pendingMockData.method,
                statusCode: pendingMockData.statusCode,
                contentType: pendingMockData.contentType || 'application/json',
                body: pendingMockData.body || '',
                delay: 0,
                description: pendingMockData.description || '',
            });
            setMockEditModalOpen(true);
            setPendingMockData(null);
        }
    }, [pendingMockData]);

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
                    isProtobuf: detail.isProtobuf || false,
                    isReqProtobuf: detail.isReqProtobuf || false,
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
            setMockRules((rules || []) as any);
        } catch (err) {
            console.error('Failed to load mock rules:', err);
        }
    };

    // Open mock list modal
    const handleOpenMockListModal = () => {
        loadMockRules();
        openMockListModal();
    };

    // Add or update mock rule
    const handleSaveMockRule = async () => {
        try {
            const values = await mockForm.validateFields();
            // Filter out incomplete conditions (must have at least type and operator)
            const conditions = (values.conditions || []).filter(
                (c: any) => c && c.type && c.operator
            );
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
                createdAt: editingMockRule?.createdAt || 0,
                conditions,
            };

            if (editingMockRule) {
                // Update existing - for now just remove and re-add
                await RemoveMockRule(editingMockRule.id);
            }
            await AddMockRule(rule as any);

            message.success(editingMockRule ? t('proxy.mock_rule_updated') : t('proxy.mock_rule_added'));
            mockForm.resetFields();
            closeMockEditModal();
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

    // Edit mock rule — open edit modal with prefilled form
    const startEditMockRule = (rule: any) => {
        setMockConditionHints(null);
        openMockEditModal(rule);
        mockForm.resetFields();
        mockForm.setFieldsValue({
            urlPattern: rule.urlPattern,
            method: rule.method,
            statusCode: rule.statusCode,
            contentType: rule.headers?.['Content-Type'] || 'application/json',
            body: rule.body,
            delay: rule.delay,
            description: rule.description,
            conditions: rule.conditions || [],
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

        // Extract condition hints from the captured request
        const hintHeaders: Array<{ key: string; value: string }> = [];
        const headerData = log.headers || (log as any).requestHeaders;
        if (headerData) {
            for (const [key, values] of Object.entries(headerData)) {
                const val = Array.isArray(values) ? (values as string[]).join(', ') : String(values || '');
                hintHeaders.push({ key, value: val });
            }
            hintHeaders.sort((a, b) => a.key.localeCompare(b.key));
        }

        const hintQueryParams: Array<{ key: string; value: string }> = [];
        try {
            const urlObj = new URL(log.url);
            urlObj.searchParams.forEach((value, key) => {
                hintQueryParams.push({ key, value });
            });
        } catch (e) {
            const qIdx = log.url.indexOf('?');
            if (qIdx >= 0) {
                try {
                    const params = new URLSearchParams(log.url.substring(qIdx + 1));
                    params.forEach((value, key) => {
                        hintQueryParams.push({ key, value });
                    });
                } catch (_) { /* ignore */ }
            }
        }

        const hints = { headers: hintHeaders, queryParams: hintQueryParams };
        setMockConditionHints(hints);

        openMockEditModal(null);
        mockForm.resetFields();
        mockForm.setFieldsValue({
            urlPattern,
            method: log.method,
            statusCode: log.statusCode || 200,
            contentType: log.contentType?.split(';')[0] || 'application/json',
            body: log.respBody || '',
            delay: 0,
            description: `Mock for ${log.method} ${urlPattern}`,
            _conditionHints: JSON.stringify(hints),
        });
    };

    // --- Proto management handlers ---
    const loadProtoData = async () => {
        try {
            const [files, mappings, types] = await Promise.all([
                GetProtoFiles(),
                GetProtoMappings(),
                GetProtoMessageTypes(),
            ]);
            setProtoFiles(files || []);
            setProtoMappings(mappings || []);
            setProtoMessageTypes(types || []);
        } catch (err) {
            console.error('Failed to load proto data:', err);
        }
    };

    const handleOpenProtoListModal = () => {
        loadProtoData();
        openProtoListModal();
    };

    const handleSaveProtoFile = async () => {
        try {
            const values = await protoFileForm.validateFields();
            if (editingProtoFile) {
                await UpdateProtoFile(editingProtoFile.id, values.name, values.content);
                message.success(t('proxy.proto_file_updated'));
            } else {
                await AddProtoFile(values.name, values.content);
                message.success(t('proxy.proto_file_added'));
            }
            protoFileForm.resetFields();
            closeProtoEditFileModal();
            loadProtoData();
        } catch (err: any) {
            message.error(String(err));
        }
    };

    const handleDeleteProtoFile = async (id: string) => {
        try {
            await RemoveProtoFile(id);
            message.success(t('proxy.proto_file_deleted'));
            loadProtoData();
        } catch (err) {
            message.error(String(err));
        }
    };

    const startEditProtoFile = (file: any) => {
        openProtoEditFileModal(file);
        protoFileForm.resetFields();
        protoFileForm.setFieldsValue({
            name: file.name,
            content: file.content,
        });
    };

    const handleImportProtoFromDisk = async () => {
        try {
            setProtoImportLoading(true);
            const ids = await LoadProtoFromDisk();
            if (ids && ids.length > 0) {
                message.success(t('proxy.proto_files_imported', { count: ids.length }));
                loadProtoData();
            }
        } catch (err: any) {
            message.error(String(err));
        } finally {
            setProtoImportLoading(false);
        }
    };

    const handleImportProtoFromURL = () => {
        openProtoImportURLModal();
    };

    const handleConfirmImportProtoFromURL = async () => {
        const url = protoImportURL;
        if (!url || url === 'https://') return;
        try {
            setProtoImportLoading(true);
            const ids = await LoadProtoFromURL(url);
            if (ids && ids.length > 0) {
                message.success(t('proxy.proto_files_imported', { count: ids.length }));
                loadProtoData();
            }
            closeProtoImportURLModal();
        } catch (err: any) {
            message.error(String(err));
        } finally {
            setProtoImportLoading(false);
        }
    };

    const handleSaveProtoMapping = async () => {
        try {
            const values = await protoMappingForm.validateFields();
            if (editingProtoMapping) {
                await UpdateProtoMapping(
                    editingProtoMapping.id,
                    values.urlPattern,
                    values.messageType,
                    values.direction || 'response',
                    values.description || ''
                );
                message.success(t('proxy.proto_mapping_updated'));
            } else {
                await AddProtoMapping(
                    values.urlPattern,
                    values.messageType,
                    values.direction || 'response',
                    values.description || ''
                );
                message.success(t('proxy.proto_mapping_added'));
            }
            protoMappingForm.resetFields();
            closeProtoEditMappingModal();
            loadProtoData();
        } catch (err: any) {
            message.error(String(err));
        }
    };

    const handleDeleteProtoMapping = async (id: string) => {
        try {
            await RemoveProtoMapping(id);
            message.success(t('proxy.proto_mapping_deleted'));
            loadProtoData();
        } catch (err) {
            message.error(String(err));
        }
    };

    const startEditProtoMapping = (mapping: any) => {
        openProtoEditMappingModal(mapping);
        protoMappingForm.resetFields();
        protoMappingForm.setFieldsValue({
            urlPattern: mapping.urlPattern,
            messageType: mapping.messageType,
            direction: mapping.direction,
            description: mapping.description,
        });
    };

    const filteredLogs = logs.filter(log => {
        // Filter by type (ALL, HTTP, WS)
        if (filterType === "HTTP" && log.isWs) return false;
        if (filterType === "WS" && !log.isWs) return false;

        // Filter by search text (deep search: URL, method, status, headers, body, response)
        if (searchText) {
            const lowerSearch = searchText.toLowerCase();

            // Basic fields
            if (log.url.toLowerCase().includes(lowerSearch)) return true;
            if (log.method.toLowerCase().includes(lowerSearch)) return true;
            if (String(log.statusCode || '').includes(lowerSearch)) return true;
            if ((log.contentType || '').toLowerCase().includes(lowerSearch)) return true;
            if ((log.clientIp || '').toLowerCase().includes(lowerSearch)) return true;

            // Request headers (key + values)
            if (log.headers) {
                for (const [key, values] of Object.entries(log.headers)) {
                    if (key.toLowerCase().includes(lowerSearch)) return true;
                    if (values?.some(v => v.toLowerCase().includes(lowerSearch))) return true;
                }
            }

            // Response headers (key + values)
            if (log.respHeaders) {
                for (const [key, values] of Object.entries(log.respHeaders)) {
                    if (key.toLowerCase().includes(lowerSearch)) return true;
                    if (values?.some(v => v.toLowerCase().includes(lowerSearch))) return true;
                }
            }

            // Request body
            if ((log.previewBody || '').toLowerCase().includes(lowerSearch)) return true;

            // Response body
            if ((log.respBody || '').toLowerCase().includes(lowerSearch)) return true;

            return false;
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
                            <Button.Group size="small">
                                <Tooltip title={t('proxy.mock_rules')}>
                                    <Button icon={<BlockOutlined />} onClick={handleOpenMockListModal}>
                                        Mock
                                    </Button>
                                </Tooltip>
                                <Tooltip title={t('proxy.add_mock_rule')}>
                                    <Button icon={<PlusOutlined />} onClick={() => { mockForm.resetFields(); setMockConditionHints(null); openMockEditModal(null); }} />
                                </Tooltip>
                            </Button.Group>
                            <Tooltip title={t('proxy.proto_management')}>
                                <Button size="small" icon={<CodeOutlined />} onClick={handleOpenProtoListModal}>
                                    Proto
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
                                const reqHasHeaders = selectedLog.headers && Object.keys(selectedLog.headers).length > 0;
                                const respHasHeaders = selectedLog.respHeaders && Object.keys(selectedLog.respHeaders).length > 0;
                                const headerGrid = (headers: Record<string, string[]>) => (
                                    <div style={{ display: 'grid', gridTemplateColumns: 'auto 1fr', gap: '4px 12px', paddingTop: 8 }}>
                                        {Object.entries(headers).map(([k, v]) => (
                                            <React.Fragment key={k}>
                                                <Text style={{ fontSize: '12px', color: token.colorTextSecondary, textAlign: 'right', fontWeight: 500 }}>{k}:</Text>
                                                <Text copyable={{ text: (v as string[]).join(', ') }} style={{ fontSize: '12px', fontFamily: 'monospace', wordBreak: 'break-all' }}>{(v as string[]).join(', ')}</Text>
                                            </React.Fragment>
                                        ))}
                                    </div>
                                );

                                return (
                                    <Tabs defaultActiveKey={selectedLog.previewBody ? 'reqBody' : 'reqHeaders'} size="small" items={[
                                        {
                                            key: 'reqHeaders',
                                            label: t('proxy.req_headers'),
                                            children: (
                                                <div style={{ paddingTop: 8 }}>
                                                    {reqHasHeaders ? headerGrid(selectedLog.headers as Record<string, string[]>) : (
                                                        <Text type="secondary" style={{ display: 'block', textAlign: 'center', margin: '40px 0' }}>{t('proxy.no_req_data')}</Text>
                                                    )}
                                                </div>
                                            )
                                        },
                                        ...(queryParams.length > 0 ? [{
                                            key: 'queryParams',
                                            label: <span>{t('proxy.query_params_tab')} <Tag style={{ marginLeft: 4, fontSize: 10, lineHeight: '16px', padding: '0 4px' }}>{queryParams.length}</Tag></span>,
                                            children: (
                                                <div style={{ paddingTop: 8 }}>
                                                    <div style={{
                                                        border: `1px solid ${token.colorBorderSecondary}`,
                                                        borderRadius: 6,
                                                        overflow: 'hidden',
                                                    }}>
                                                        <div style={{
                                                            display: 'grid',
                                                            gridTemplateColumns: 'minmax(100px, auto) 1fr',
                                                            fontSize: 12,
                                                        }}>
                                                            <div style={{
                                                                padding: '6px 12px',
                                                                fontWeight: 600,
                                                                color: token.colorTextSecondary,
                                                                background: token.colorFillAlter,
                                                                borderBottom: `1px solid ${token.colorBorderSecondary}`,
                                                                borderRight: `1px solid ${token.colorBorderSecondary}`,
                                                            }}>{t('proxy.param_name')}</div>
                                                            <div style={{
                                                                padding: '6px 12px',
                                                                fontWeight: 600,
                                                                color: token.colorTextSecondary,
                                                                background: token.colorFillAlter,
                                                                borderBottom: `1px solid ${token.colorBorderSecondary}`,
                                                            }}>{t('proxy.param_value')}</div>
                                                            {queryParams.map(([k, v], idx) => (
                                                                <React.Fragment key={idx}>
                                                                    <div style={{
                                                                        padding: '5px 12px',
                                                                        fontWeight: 500,
                                                                        color: token.colorText,
                                                                        borderBottom: idx < queryParams.length - 1 ? `1px solid ${token.colorBorderSecondary}` : 'none',
                                                                        borderRight: `1px solid ${token.colorBorderSecondary}`,
                                                                        wordBreak: 'break-all',
                                                                    }}>{k}</div>
                                                                    <div style={{
                                                                        padding: '5px 12px',
                                                                        borderBottom: idx < queryParams.length - 1 ? `1px solid ${token.colorBorderSecondary}` : 'none',
                                                                    }}>
                                                                        <Text copyable={{ text: v }} style={{ fontSize: 12, fontFamily: 'monospace', wordBreak: 'break-all', color: token.colorLink }}>{v}</Text>
                                                                    </div>
                                                                </React.Fragment>
                                                            ))}
                                                        </div>
                                                    </div>
                                                </div>
                                            )
                                        }] : []),
                                        {
                                            key: 'reqBody',
                                            label: <span>{t('proxy.req_body')}{selectedLog.isReqProtobuf ? ' [PB]' : ''}</span>,
                                            children: (
                                                <div style={{ paddingTop: 8 }}>
                                                    {selectedLog.isReqProtobuf && (
                                                        <div style={{ marginBottom: 8 }}><Tag color="geekblue">Protobuf Decoded</Tag></div>
                                                    )}
                                                    {selectedLog.previewBody ? (
                                                        <JsonViewer data={formatBody(selectedLog.previewBody)} fontSize={12} />
                                                    ) : (
                                                        <Text type="secondary" style={{ display: 'block', textAlign: 'center', margin: '40px 0' }}>{t('proxy.no_body')}</Text>
                                                    )}
                                                </div>
                                            )
                                        },
                                        {
                                            key: 'respHeaders',
                                            label: `${t('proxy.resp_headers')} ${selectedLog.statusCode ? '(' + selectedLog.statusCode + ')' : ''}`,
                                            children: (
                                                <div style={{ paddingTop: 8 }}>
                                                    {selectedLog.statusCode ? (
                                                        <>
                                                            <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 12 }}>
                                                                <Tag color={selectedLog.statusCode >= 400 ? 'error' : 'success'} style={{ fontSize: 14, padding: '2px 8px' }}>
                                                                    {selectedLog.statusCode}
                                                                </Tag>
                                                                {selectedLog.mocked && <Tag color="magenta">Mocked</Tag>}
                                                                {selectedLog.contentType && <Tag>{selectedLog.contentType.split(';')[0]}</Tag>}
                                                            </div>
                                                            {respHasHeaders ? headerGrid(selectedLog.respHeaders as Record<string, string[]>) : (
                                                                <Text type="secondary" style={{ display: 'block', textAlign: 'center', margin: '40px 0' }}>{t('proxy.no_resp_data')}</Text>
                                                            )}
                                                        </>
                                                    ) : (
                                                        <div style={{ padding: '60px 20px', textAlign: 'center' }}>
                                                            <HourglassOutlined style={{ fontSize: 48, color: token.colorTextDisabled, marginBottom: 16 }} />
                                                            <br />
                                                            <Text type="secondary" italic>{t('proxy.waiting_for_response')}</Text>
                                                        </div>
                                                    )}
                                                </div>
                                            )
                                        },
                                        {
                                            key: 'respBody',
                                            label: <span>{t('proxy.resp_body')}{selectedLog.isProtobuf ? ' [PB]' : ''}</span>,
                                            children: (
                                                <div style={{ paddingTop: 8 }}>
                                                    {selectedLog.statusCode ? (
                                                        <>
                                                            {selectedLog.isProtobuf && (
                                                                <div style={{ marginBottom: 8 }}><Tag color="geekblue">Protobuf Decoded</Tag></div>
                                                            )}
                                                            {selectedLog.respBody ? (
                                                                <JsonViewer data={formatBody(selectedLog.respBody)} fontSize={12} />
                                                            ) : (
                                                                <Text type="secondary" style={{ display: 'block', textAlign: 'center', margin: '40px 0' }}>{t('proxy.no_body')}</Text>
                                                            )}
                                                        </>
                                                    ) : (
                                                        <div style={{ padding: '60px 20px', textAlign: 'center' }}>
                                                            <HourglassOutlined style={{ fontSize: 48, color: token.colorTextDisabled, marginBottom: 16 }} />
                                                            <br />
                                                            <Text type="secondary" italic>{t('proxy.waiting_for_response')}</Text>
                                                        </div>
                                                    )}
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
                        <JsonEditor height={120} placeholder="Request body" />
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
                        <div style={{ flex: 1, overflow: 'auto', minHeight: 0 }}>
                            <JsonViewer data={formatBody(resendResponse.body, resendResponse.contentType)} fontSize={12} />
                        </div>
                    </div>
                )}
            </Modal>

            {/* Mock Rules List Modal */}
            <Modal
                title={t('proxy.mock_rules')}
                open={mockListModalOpen}
                onCancel={closeMockListModal}
                width={800}
                footer={null}
                style={{ top: 32, paddingBottom: 0 }}
            >
                <div style={{ marginBottom: 12 }}>
                    <Button type="primary" icon={<PlusOutlined />} onClick={() => { mockForm.resetFields(); setMockConditionHints(null); openMockEditModal(null); }}>
                        {t('proxy.add_mock_rule')}
                    </Button>
                </div>
                <div style={{ maxHeight: 'calc(100vh - 220px)', overflow: 'auto' }}>
                    {mockRules.length === 0 ? (
                        <Text type="secondary" style={{ display: 'block', textAlign: 'center', padding: 32 }}>{t('proxy.no_mock_rules')}</Text>
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
                                { title: t('proxy.conditions'), dataIndex: 'conditions', width: 90, render: (v: any[]) => v && v.length > 0 ? <Tag color="blue">{v.length}</Tag> : '-' },
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

            {/* Mock Rule Edit Modal */}
            <Modal
                title={editingMockRule ? t('proxy.edit_mock_rule') : t('proxy.add_mock_rule')}
                open={mockEditModalOpen}
                onCancel={() => { closeMockEditModal(); mockForm.resetFields(); }}
                width={900}
                footer={null}
                style={{ top: 32, paddingBottom: 0 }}
            >
                <Form form={mockForm} layout="vertical" size="small">
                    <Space wrap style={{ width: '100%' }}>
                        <Form.Item name="urlPattern" label={t('proxy.url_pattern')} rules={[{ required: true }]} style={{ marginBottom: 8, minWidth: 280 }}>
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
                        <JsonEditor height={'calc(100vh - 560px)'} placeholder='{"success": true}' />
                    </Form.Item>
                    <Form.Item name="description" label={t('proxy.description')} style={{ marginBottom: 12 }}>
                        <Input placeholder={t('proxy.description')} />
                    </Form.Item>

                    {/* Conditions editor */}
                    <Divider style={{ margin: '8px 0' }}>{t('proxy.conditions')}</Divider>
                    <Form.List name="conditions">
                        {(fields, { add, remove }) => (
                            <>
                                {fields.map(({ key, name, ...restField }) => {
                                    const rowType = watchedConditions?.[name]?.type;
                                    const rowOperator = watchedConditions?.[name]?.operator;
                                    const isBodyType = rowType === 'body';
                                    const isExistenceOp = rowOperator === 'exists' || rowOperator === 'not_exists';

                                    // Filter AutoComplete presets based on selected condition type
                                    const keyOptions = (() => {
                                        const showHeaders = rowType !== 'query';
                                        const showQuery = rowType !== 'header';
                                        const groups: Array<{ label: React.ReactNode; options: Array<{ value: string; label: React.ReactNode }> }> = [];
                                        if (showHeaders && conditionKeyPresets.headers.length > 0) {
                                            groups.push({
                                                label: <span style={{ fontWeight: 600, fontSize: 11, color: '#888' }}>Headers</span>,
                                                options: conditionKeyPresets.headers.map(h => ({
                                                    value: h.key,
                                                    label: (
                                                        <span>
                                                            <strong>{h.key}</strong>
                                                            <span style={{ color: '#888', marginLeft: 6, fontSize: 12 }}>
                                                                {h.value.length > 30 ? h.value.substring(0, 30) + '...' : h.value}
                                                            </span>
                                                        </span>
                                                    ),
                                                })),
                                            });
                                        }
                                        if (showQuery && conditionKeyPresets.queryParams.length > 0) {
                                            groups.push({
                                                label: <span style={{ fontWeight: 600, fontSize: 11, color: '#888' }}>Query Params</span>,
                                                options: conditionKeyPresets.queryParams.map(q => ({
                                                    value: q.key,
                                                    label: (
                                                        <span>
                                                            <strong>{q.key}</strong>
                                                            <span style={{ color: '#888' }}> = {q.value}</span>
                                                        </span>
                                                    ),
                                                })),
                                            });
                                        }
                                        return groups;
                                    })();

                                    return (
                                    <Space key={key} align="baseline" style={{ display: 'flex', marginBottom: 4 }} wrap>
                                        <Form.Item {...restField} name={[name, 'type']} style={{ marginBottom: 0, width: 100 }}>
                                            <Select placeholder={t('proxy.condition_type')} size="small"
                                                onChange={() => {
                                                    mockForm.setFieldValue(['conditions', name, 'key'], '');
                                                    mockForm.setFieldValue(['conditions', name, 'value'], '');
                                                }}
                                            >
                                                <Select.Option value="header">{t('proxy.condition_header')}</Select.Option>
                                                <Select.Option value="query">{t('proxy.condition_query')}</Select.Option>
                                                <Select.Option value="body">{t('proxy.condition_body')}</Select.Option>
                                            </Select>
                                        </Form.Item>

                                        {/* Key field — disabled for body type, AutoComplete with presets filtered by type */}
                                        <Form.Item {...restField} name={[name, 'key']} style={{ marginBottom: 0, width: 200 }}>
                                            <AutoComplete
                                                options={isBodyType ? [] : keyOptions}
                                                placeholder={isBodyType ? '-' : t('proxy.condition_key')}
                                                size="small"
                                                disabled={isBodyType}
                                                filterOption={(input, option) =>
                                                    ((option as { value?: string })?.value ?? '').toLowerCase().includes(input.toLowerCase())
                                                }
                                                onSelect={(val: string) => {
                                                    const hint = conditionKeyPresets.headers.find(h => h.key === val)
                                                        || conditionKeyPresets.queryParams.find(q => q.key === val);
                                                    if (hint) {
                                                        mockForm.setFieldValue(['conditions', name, 'value'], hint.value);
                                                    }
                                                }}
                                            />
                                        </Form.Item>

                                        <Form.Item {...restField} name={[name, 'operator']} style={{ marginBottom: 0, width: 120 }}>
                                            <Select placeholder={t('proxy.condition_operator')} size="small"
                                                onChange={(val: string) => {
                                                    // Clear value when switching to existence operators
                                                    if (val === 'exists' || val === 'not_exists') {
                                                        mockForm.setFieldValue(['conditions', name, 'value'], '');
                                                    }
                                                }}
                                            >
                                                <Select.Option value="equals">{t('proxy.op_equals')}</Select.Option>
                                                <Select.Option value="contains">{t('proxy.op_contains')}</Select.Option>
                                                <Select.Option value="regex">{t('proxy.op_regex')}</Select.Option>
                                                <Select.Option value="exists">{t('proxy.op_exists')}</Select.Option>
                                                <Select.Option value="not_exists">{t('proxy.op_not_exists')}</Select.Option>
                                            </Select>
                                        </Form.Item>
                                        {/* Value field — hidden for exists/not_exists operators */}
                                        {!isExistenceOp && (
                                        <Form.Item {...restField} name={[name, 'value']} style={{ marginBottom: 0, flex: 1, minWidth: 140 }}>
                                            <Input placeholder={t('proxy.condition_value')} size="small" />
                                        </Form.Item>
                                        )}
                                        <MinusCircleOutlined onClick={() => remove(name)} style={{ color: '#ff4d4f' }} />
                                    </Space>
                                    );
                                })}
                                <Button type="dashed" onClick={() => add({ type: 'header', key: '', operator: 'equals', value: '' })} icon={<PlusOutlined />} size="small" style={{ marginBottom: 12 }}>
                                    {t('proxy.add_condition')}
                                </Button>
                            </>
                        )}
                    </Form.List>

                    <Space>
                        <Button type="primary" onClick={handleSaveMockRule}>
                            {editingMockRule ? t('common.save') : t('proxy.add_mock_rule')}
                        </Button>
                        <Button onClick={() => { closeMockEditModal(); mockForm.resetFields(); }}>
                            {t('common.cancel')}
                        </Button>
                    </Space>
                </Form>
            </Modal>

            {/* Proto Management Modal */}
            <Modal
                title={t('proxy.proto_management')}
                open={protoListModalOpen}
                onCancel={closeProtoListModal}
                width={900}
                footer={null}
                style={{ top: 32, paddingBottom: 0 }}
            >
                <Tabs size="small" items={[
                    {
                        key: 'files',
                        label: t('proxy.proto_files'),
                        children: (
                            <div>
                                <div style={{ marginBottom: 12 }}>
                                    <Space wrap>
                                        <Button type="primary" icon={<PlusOutlined />} onClick={() => { protoFileForm.resetFields(); openProtoEditFileModal(null); }}>
                                            {t('proxy.add_proto_file')}
                                        </Button>
                                        <Button icon={<FolderOpenOutlined />} loading={protoImportLoading} onClick={handleImportProtoFromDisk}>
                                            {t('proxy.import_local_file')}
                                        </Button>
                                        <Button icon={<CloudDownloadOutlined />} loading={protoImportLoading} onClick={handleImportProtoFromURL}>
                                            {t('proxy.import_from_url')}
                                        </Button>
                                    </Space>
                                </div>
                                <div style={{ maxHeight: 'calc(100vh - 320px)', overflow: 'auto' }}>
                                    {protoFiles.length === 0 ? (
                                        <Text type="secondary" style={{ display: 'block', textAlign: 'center', padding: 32 }}>{t('proxy.no_proto_files')}</Text>
                                    ) : (
                                        <Table
                                            dataSource={protoFiles}
                                            rowKey="id"
                                            size="small"
                                            pagination={false}
                                            columns={[
                                                { title: t('proxy.proto_file_name'), dataIndex: 'name', ellipsis: true },
                                                {
                                                    title: t('proxy.proto_file_size'),
                                                    dataIndex: 'content',
                                                    width: 100,
                                                    render: (content: string) => `${(content || '').length} chars`
                                                },
                                                {
                                                    title: '',
                                                    width: 100,
                                                    render: (_: any, record: any) => (
                                                        <Space size="small">
                                                            <Button type="text" size="small" icon={<EditOutlined />} onClick={() => startEditProtoFile(record)} />
                                                            <Popconfirm title={t('common.delete') + '?'} onConfirm={() => handleDeleteProtoFile(record.id)}>
                                                                <Button type="text" size="small" danger icon={<DeleteOutlined />} />
                                                            </Popconfirm>
                                                        </Space>
                                                    )
                                                }
                                            ]}
                                        />
                                    )}
                                </div>
                            </div>
                        )
                    },
                    {
                        key: 'mappings',
                        label: t('proxy.proto_mappings'),
                        children: (
                            <div>
                                <div style={{ marginBottom: 12 }}>
                                    <Button type="primary" icon={<PlusOutlined />} onClick={() => { protoMappingForm.resetFields(); openProtoEditMappingModal(null); }}>
                                        {t('proxy.add_proto_mapping')}
                                    </Button>
                                </div>
                                <div style={{ maxHeight: 'calc(100vh - 320px)', overflow: 'auto' }}>
                                    {protoMappings.length === 0 ? (
                                        <Text type="secondary" style={{ display: 'block', textAlign: 'center', padding: 32 }}>{t('proxy.no_proto_mappings')}</Text>
                                    ) : (
                                        <Table
                                            dataSource={protoMappings}
                                            rowKey="id"
                                            size="small"
                                            pagination={false}
                                            columns={[
                                                { title: t('proxy.url_pattern'), dataIndex: 'urlPattern', ellipsis: true },
                                                { title: t('proxy.message_type'), dataIndex: 'messageType', ellipsis: true },
                                                { title: t('proxy.direction'), dataIndex: 'direction', width: 100, render: (dir: string) => dir === 'both' ? t('proxy.proto_both') : dir === 'request' ? t('proxy.request') : dir === 'response' ? t('proxy.response') : dir },
                                                { title: t('proxy.description'), dataIndex: 'description', ellipsis: true },
                                                {
                                                    title: '',
                                                    width: 100,
                                                    render: (_: any, record: any) => (
                                                        <Space size="small">
                                                            <Button type="text" size="small" icon={<EditOutlined />} onClick={() => startEditProtoMapping(record)} />
                                                            <Popconfirm title={t('common.delete') + '?'} onConfirm={() => handleDeleteProtoMapping(record.id)}>
                                                                <Button type="text" size="small" danger icon={<DeleteOutlined />} />
                                                            </Popconfirm>
                                                        </Space>
                                                    )
                                                }
                                            ]}
                                        />
                                    )}
                                </div>
                            </div>
                        )
                    }
                ]} />
            </Modal>

            {/* Proto File Edit Modal */}
            <Modal
                title={editingProtoFile ? t('proxy.edit_proto_file') : t('proxy.add_proto_file')}
                open={protoEditFileModalOpen}
                onCancel={() => { closeProtoEditFileModal(); protoFileForm.resetFields(); }}
                width={800}
                footer={null}
                style={{ top: 32, paddingBottom: 0 }}
            >
                <Form form={protoFileForm} layout="vertical" size="small" style={{ marginTop: 8 }}>
                    <Form.Item name="name" label={t('proxy.proto_file_name')} rules={[{ required: true }]}>
                        <Input placeholder="user.proto" />
                    </Form.Item>
                    <Form.Item name="content" label={t('proxy.proto_file_content')} rules={[{ required: true }]}>
                        <Input.TextArea
                            rows={18}
                            placeholder={'syntax = "proto3";\n\npackage example;\n\nmessage UserResponse {\n  int32 id = 1;\n  string name = 2;\n  string email = 3;\n}'}
                            style={{ fontFamily: "'Fira Code', monospace", fontSize: 12 }}
                        />
                    </Form.Item>
                    <Space>
                        <Button type="primary" onClick={handleSaveProtoFile}>
                            {editingProtoFile ? t('common.save') : t('proxy.add_proto_file')}
                        </Button>
                        <Button onClick={() => { closeProtoEditFileModal(); protoFileForm.resetFields(); }}>
                            {t('common.cancel')}
                        </Button>
                    </Space>
                </Form>
            </Modal>

            {/* Proto Mapping Edit Modal */}
            <Modal
                title={editingProtoMapping ? t('proxy.edit_proto_mapping') : t('proxy.add_proto_mapping')}
                open={protoEditMappingModalOpen}
                onCancel={() => { closeProtoEditMappingModal(); protoMappingForm.resetFields(); }}
                width={600}
                footer={null}
                style={{ top: 32 }}
            >
                <Form form={protoMappingForm} layout="vertical" size="small" style={{ marginTop: 8 }}>
                    <Form.Item name="urlPattern" label={t('proxy.url_pattern')} rules={[{ required: true }]}>
                        <Input placeholder="*/api/user/*" />
                    </Form.Item>
                    <Form.Item name="messageType" label={t('proxy.message_type')} rules={[{ required: true }]}>
                        {protoMessageTypes.length > 0 ? (
                            <Input
                                placeholder="example.UserResponse"
                                suffix={
                                    <Popover
                                        trigger="click"
                                        content={
                                            <div style={{ maxHeight: 300, overflow: 'auto' }}>
                                                {protoMessageTypes.map(msgType => (
                                                    <div
                                                        key={msgType}
                                                        style={{ padding: '4px 8px', cursor: 'pointer', fontSize: 12, fontFamily: 'monospace', borderRadius: 4 }}
                                                        onMouseEnter={(e) => { (e.target as HTMLElement).style.background = token.colorFillSecondary; }}
                                                        onMouseLeave={(e) => { (e.target as HTMLElement).style.background = 'transparent'; }}
                                                        onClick={() => {
                                                            protoMappingForm.setFieldValue('messageType', msgType);
                                                        }}
                                                    >
                                                        {msgType}
                                                    </div>
                                                ))}
                                            </div>
                                        }
                                    >
                                        <Button type="text" size="small" style={{ fontSize: 10 }}>types</Button>
                                    </Popover>
                                }
                            />
                        ) : (
                            <Input placeholder="example.UserResponse" />
                        )}
                    </Form.Item>
                    <Form.Item name="direction" label={t('proxy.direction')} initialValue="response">
                        <Radio.Group buttonStyle="solid" size="small">
                            <Radio.Button value="response">{t('proxy.response')}</Radio.Button>
                            <Radio.Button value="request">{t('proxy.request')}</Radio.Button>
                            <Radio.Button value="both">{t('proxy.proto_both')}</Radio.Button>
                        </Radio.Group>
                    </Form.Item>
                    <Form.Item name="description" label={t('proxy.description')}>
                        <Input placeholder={t('proxy.description')} />
                    </Form.Item>
                    <Space>
                        <Button type="primary" onClick={handleSaveProtoMapping}>
                            {editingProtoMapping ? t('common.save') : t('proxy.add_proto_mapping')}
                        </Button>
                        <Button onClick={() => { closeProtoEditMappingModal(); protoMappingForm.resetFields(); }}>
                            {t('common.cancel')}
                        </Button>
                    </Space>
                </Form>
            </Modal>

            {/* Proto Import URL Modal */}
            <Modal
                title={t('proxy.import_from_url')}
                open={protoImportURLModalOpen}
                onCancel={closeProtoImportURLModal}
                onOk={handleConfirmImportProtoFromURL}
                confirmLoading={protoImportLoading}
                okText={t('proxy.import')}
                cancelText={t('common.cancel')}
                width={500}
            >
                <div style={{ marginTop: 16 }}>
                    <Input
                        placeholder="https://raw.githubusercontent.com/..."
                        value={protoImportURL}
                        onChange={(e) => setProtoImportURL(e.target.value)}
                        onPressEnter={handleConfirmImportProtoFromURL}
                        autoFocus
                    />
                </div>
            </Modal>

        </div>
    );
};

export default ProxyView;
