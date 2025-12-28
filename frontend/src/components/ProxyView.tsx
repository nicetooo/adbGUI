import React, { useState, useEffect, useRef } from 'react';
import { Card, Button, InputNumber, Space, Typography, Tag, message, Table, Modal, Divider, Switch, Tooltip, Radio, Input, Drawer, Tabs } from 'antd';
import { PoweroffOutlined, PlayCircleOutlined, DeleteOutlined, SettingOutlined, LockOutlined, GlobalOutlined, DashboardOutlined, ThunderboltOutlined, ArrowUpOutlined, ArrowDownOutlined, ApiOutlined, SafetyCertificateOutlined, DownloadOutlined, HourglassOutlined } from '@ant-design/icons';
import DeviceSelector from './DeviceSelector';
// @ts-ignore
import { StartProxy, StopProxy, GetProxyStatus, GetLocalIP, RunAdbCommand, StartNetworkMonitor, StopNetworkMonitor, SetDeviceNetworkLimit, SetProxyLimit, SetProxyWSEnabled, SetProxyMITM, InstallProxyCert, SetProxyLatency, SetMITMBypassPatterns } from '../../wailsjs/go/main/App';
// @ts-ignore
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import { useVirtualizer } from '@tanstack/react-virtual';

const { Title, Text } = Typography;

interface ProxyViewProps {
  devices: any[];
  selectedDevice: string;
  setSelectedDevice: (id: string) => void;
  fetchDevices: () => Promise<void>;
  loading: boolean;
}

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
}

const ProxyView: React.FC<ProxyViewProps> = ({
  devices, selectedDevice, setSelectedDevice, fetchDevices, loading
}) => {
  const [isRunning, setIsRunning] = useState(false);
  const [port, setPort] = useState(8080);
  const [localIP, setLocalIP] = useState("");
  const [logs, setLogs] = useState<RequestLog[]>([]);
  const logsEndRef = useRef<HTMLDivElement>(null);
  const [wsEnabled, setWsEnabled] = useState(true); // Default matching backend
  const [mitmEnabled, setMitmEnabled] = useState(true);
  const [filterType, setFilterType] = useState<"ALL" | "HTTP" | "WS">("ALL");
  const [searchText, setSearchText] = useState("");
  const [latency, setLatency] = useState<number | null>(null);
  const [bypassPatterns, setBypassPatterns] = useState<string[]>([]);
  const [isBypassModalOpen, setIsBypassModalOpen] = useState(false);
  const [newPattern, setNewPattern] = useState("");
  
  const parentRef = useRef<HTMLDivElement>(null);
  const [selectedLog, setSelectedLog] = useState<RequestLog | null>(null);
  const [detailsDrawerOpen, setDetailsDrawerOpen] = useState(false);

  const [netStats, setNetStats] = useState<NetworkStats>({ rxSpeed: 0, txSpeed: 0, rxBytes: 0, txBytes: 0, time: 0 });
  // const [isMonitoring, setIsMonitoring] = useState(false); // Removed manual toggle
  const [dlLimit, setDlLimit] = useState<number | null>(null);
  const [ulLimit, setUlLimit] = useState<number | null>(null);

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
        if (selectedDevice) StopNetworkMonitor(selectedDevice); 
    };
  }, [selectedDevice]);

  useEffect(() => {
    // Initial status check
    GetProxyStatus().then((status: boolean) => setIsRunning(status));
    GetLocalIP().then((ip: string) => setLocalIP(ip));
    
    // Sync settings from backend
    // @ts-ignore
    import('../../wailsjs/go/main/App').then(m => {
        if (m.GetProxySettings) {
            m.GetProxySettings().then((settings: any) => {
                if (settings.wsEnabled !== undefined) setWsEnabled(settings.wsEnabled);
                if (settings.mitmEnabled !== undefined) setMitmEnabled(settings.mitmEnabled);
                if (settings.bypassPatterns !== undefined) setBypassPatterns(settings.bypassPatterns);
            });
        }
    });

    // Listen for logs
    EventsOn("proxy_request", (log: RequestLog) => {
      setLogs(prev => {
        const index = prev.findIndex(l => l.id === log.id);
        if (index > -1) {
            const newLogs = [...prev];
            newLogs[index] = { ...newLogs[index], ...log };
            return newLogs;
        }
        return [log, ...prev].slice(0, 5000);
      });
    });

    return () => {
      EventsOff("proxy-request");
    };
  }, []);
  
  useEffect(() => {
    // Auto-scroll logic could be added here if needed, but Table manages its own scroll usually.
    // If we want to stick to bottom, we could scroll the table body container.
    // For now, let's just let it be or use a "scroll to bottom" generic approach if user wants.
  }, [logs]);

  const handleStart = async () => {
    try {
      // 1. Start Server
      await SetProxyMITM(mitmEnabled);
      await SetProxyWSEnabled(wsEnabled);
      await StartProxy(port);
      setIsRunning(true);
      
      // 2. Automagically set device proxy if selected
      if (selectedDevice && localIP) {
          try {
              const cmd = `shell settings put global http_proxy ${localIP}:${port}`;
              await RunAdbCommand(selectedDevice, cmd);
              message.success(`Proxy started & Device linked to ${localIP}:${port}`);
          } catch (adbErr: any) {
              const errorStr = String(adbErr);
              if (errorStr.includes("WRITE_SECURE_SETTINGS")) {
                  message.warning("Proxy started, but device linking failed (Security Settings blocked)");
                  showSecurityError();
              } else {
                  message.error("Failed to link device: " + errorStr);
              }
          }
      } else {
          message.success("Proxy server started (Local mode)");
      }
    } catch (err) {
      message.error("Failed to start proxy: " + String(err));
    }
  };

  const handleStop = async () => {
    try {
      // 1. Automagically clear device proxy
      if (selectedDevice) {
          try {
              await RunAdbCommand(selectedDevice, "shell settings put global http_proxy :0");
          } catch (e) {}
      }
      
      // 2. Stop Server
      await StopProxy();
      setIsRunning(false);
      message.success("Capture stopped & Device proxy cleared");
    } catch (err) {
      message.error("Failed to stop proxy: " + String(err));
    }
  };

  const showSecurityError = () => {
    Modal.error({
        title: 'Permission Denied',
        content: (
            <div>
            <p>Failed to change settings due to permission restrictions.</p>
            <p><strong>Xiaomi (MIUI/HyperOS) Users:</strong></p>
            <p>Please enable <strong>"USB debugging (Security settings)"</strong> in Developer Options.</p>
            <p style={{ fontSize: '12px', color: '#999' }}>Note: This usually requires inserting a SIM card and signing in to a Mi Account.</p>
            <br/>
            <p><strong>Other Devices:</strong></p>
            <p>Some devices may block this via ADB. You may need to set the proxy manually in Wi-Fi settings.</p>
            </div>
        ),
    });
  };

  const columns = [
    {
      title: 'Time',
      dataIndex: 'time',
      key: 'time',
      width: 100,
    },
    {
      title: 'Method',
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
      title: 'URL / Host',
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
        setWsEnabled(checked);
        message.success(checked ? "WebSocket enabled" : "WebSocket disabled");
    } catch (err) {
        message.error("Failed to toggle WebSocket: " + String(err));
    }
  };

  const handleMITMToggle = async (checked: boolean) => {
    try {
        await SetProxyMITM(checked);
        setMitmEnabled(checked);
        if (checked) {
            message.info("HTTPS Decryption enabled. Please ensure CA Cert is installed on device.");
        } else {
            message.success("HTTPS Decryption disabled");
        }
    } catch (err) {
        message.error("Failed to toggle MITM: " + String(err));
    }
  };

  const handleInstallCert = async () => {
      if (!selectedDevice) return;
      try {
          const path = await InstallProxyCert(selectedDevice);
          Modal.success({
              title: 'Certificate Pushed',
              content: (
                  <div>
                      <p>CA Certificate has been pushed to: <br/><code>{path}</code></p>
                      <Divider style={{ margin: '12px 0' }} />
                      <p><strong>Next Steps on Device:</strong></p>
                      <ol style={{ paddingLeft: 20 }}>
                          <li>Open <strong>Settings</strong> &rarr; <strong>Security</strong> (or "Encryption & credentials")</li>
                          <li>Select <strong>Install from storage</strong> / <strong>Install a certificate</strong>.</li>
                          <li>Choose <strong>CA certificate</strong> (tap "Install anyway" if warned).</li>
                          <li>Browse to <strong>Downloads</strong> and select <code>adbGUI-CA.crt</code>.</li>
                      </ol>
                      <p style={{ fontSize: '12px', color: '#ff4d4f', marginTop: 8 }}>
                          * Note: On Android 11+, you might need to install it in "Wi-Fi certificate" and "VPN & app user certificate" separately depending on usage.
                      </p>
                  </div>
              )
          });
      } catch (err) {
          message.error("Failed to push cert: " + String(err));
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

          message.success(`Applied: DL ${dlLimit || 0}K, UL ${ulLimit || 0}K, Delay ${ms}ms`);
      } catch (err) {
          message.error("Failed to apply rules: " + String(err));
      }
  };

  const handleClearRules = async () => {
       try {
           await SetProxyLimit(0, 0);
           await SetProxyLatency(0);
           setDlLimit(null);
           setUlLimit(null);
           setLatency(null);
           message.success("All limits cleared");
       } catch (err) {
           message.error("Failed to reset: " + String(err));
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

const formatBody = (body: string) => {
    if (!body) return "";
    try {
        // Simple heuristic to check if it might be JSON
        if ((body.startsWith('{') && body.endsWith('}')) || (body.startsWith('[') && body.endsWith(']'))) {
            const parsed = JSON.parse(body);
            return JSON.stringify(parsed, null, 2);
        }
    } catch (e) {
        // Not valid JSON or too large to parse quickly
    }
    return body;
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

  const rowVirtualizer = useVirtualizer({
    count: filteredLogs.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 35, // Compact row height
    overscan: 20,
  });

  return (
    <div style={{ padding: '16px', height: '100%', display: 'flex', flexDirection: 'column', gap: '12px', overflow: 'hidden' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexShrink: 0 }}>
        <Space align="center" size="small">
            <GlobalOutlined style={{ fontSize: '18px' }} />
            <Title level={4} style={{ margin: 0 }}>Proxy & Network</Title>
        </Space>
        <Space>
             <DeviceSelector
                devices={devices}
                selectedDevice={selectedDevice}
                onDeviceChange={setSelectedDevice}
                onRefresh={fetchDevices}
                loading={loading}
            />
        </Space>
      </div>

      <Card size="small" bodyStyle={{ padding: '12px' }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
              {/* Row 1: Proxy & Device Settings */}
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <Space split={<Divider type="vertical" />}>
                      <Space size="small">
                          <Text strong>Server:</Text>
                          <Tag color={isRunning ? "success" : "default"} style={{ marginRight: 0 }}>{isRunning ? "ON" : "OFF"}</Tag>
                      </Space>
                      <Space size="small">
                          <Text type="secondary">IP:</Text>
                          <Text copyable style={{ fontSize: '13px' }}>{localIP || "Unknown"}</Text>
                      </Space>
                      <Space size="small">
                          <Text type="secondary">Port:</Text>
                          <InputNumber size="small" value={port} onChange={(v) => setPort(v || 8080)} disabled={isRunning} style={{ width: 60 }} />
                      </Space>
                      <Space size="small">
                         <Tooltip title="Allow WebSocket (ws://) connections">
                           <Space size={4}>
                             <ApiOutlined style={{ color: wsEnabled ? '#1890ff' : undefined }} />
                             <Switch size="small" checked={wsEnabled} onChange={handleWSToggle} />
                           </Space>
                         </Tooltip>
                      </Space>
                      <Space size="small">
                         <Tooltip title="Enable HTTPS Decryption (MITM). Requires CA Cert.">
                           <Space size={4}>
                             <SafetyCertificateOutlined style={{ color: mitmEnabled ? '#faad14' : undefined }} />
                             <Switch size="small" checked={mitmEnabled} onChange={handleMITMToggle} />
                           </Space>
                         </Tooltip>
                      </Space>
                       {mitmEnabled && (
                           <Space size="small">
                               <Button size="small" icon={<DownloadOutlined />} onClick={handleInstallCert}>
                                   Cert
                               </Button>
                               <Button size="small" icon={<SettingOutlined />} onClick={() => setIsBypassModalOpen(true)}>
                                   Rules
                               </Button>
                           </Space>
                       )}
                      
                      <Button 
                        type="primary" 
                        size="small" 
                        danger={isRunning}
                        icon={isRunning ? <PoweroffOutlined /> : <PlayCircleOutlined />} 
                        onClick={isRunning ? handleStop : handleStart}
                        style={{ height: 32, padding: '0 20px', borderRadius: 16 }}
                      >
                         {isRunning ? "Stop Capture" : "Start Capture"}
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
                        <Tag color="orange" style={{ marginRight: 0 }}>Limit</Tag>
                        
                        {/* Bandwidth */}
                        <InputNumber 
                            size="small" 
                            prefix={<ArrowDownOutlined style={{fontSize: 10, color: '#aaa'}} />}
                            suffix="K" 
                            placeholder="DL" 
                            min={0}
                            value={dlLimit}
                            onChange={setDlLimit}
                            style={{ width: 110 }} 
                            title="Download Limit (KB/s)"
                        />
                        <InputNumber 
                            size="small" 
                            prefix={<ArrowUpOutlined style={{fontSize: 10, color: '#aaa'}} />}
                            suffix="K" 
                            placeholder="UL" 
                            min={0}
                            value={ulLimit}
                            onChange={setUlLimit}
                            style={{ width: 110 }} 
                            title="Upload Limit (KB/s)"
                        />

                        {/* Latency */}
                        <InputNumber 
                            size="small" 
                            prefix={<HourglassOutlined style={{fontSize: 10, color: '#aaa'}} />}
                            suffix="ms" 
                            placeholder="Delay" 
                            min={0}
                            value={latency}
                            onChange={setLatency}
                            style={{ width: 120 }} 
                            title="Artificial Latency (ms)"
                        />
                        
                        {/* Actions */}
                        <Button type="primary" size="small" onClick={handleApplyRules}>Apply</Button>
                        <Button size="small" onClick={handleClearRules} icon={<DeleteOutlined />} title="Reset All" />
                    </Space>
                </div>
                </>
              )}
          </div>
      </Card>

      <Card 
        style={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden' }} 
        bodyStyle={{ flex: 1, overflow: 'hidden', padding: 0, display: 'flex', flexDirection: 'column' }} 
        size="small"
        title={
            <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                <Radio.Group size="small" value={filterType} onChange={e => setFilterType(e.target.value)} buttonStyle="solid">
                    <Radio.Button value="ALL">ALL</Radio.Button>
                    <Radio.Button value="HTTP">HTTP</Radio.Button>
                    <Radio.Button value="WS">WS</Radio.Button>
                </Radio.Group>
                <Input 
                    placeholder="Search URL / Method / Status" 
                    size="small" 
                    allowClear 
                    style={{ maxWidth: 400 }}
                    value={searchText}
                    onChange={e => setSearchText(e.target.value)}
                />
                <Button size="small" type="link" onClick={() => setLogs([])} icon={<DeleteOutlined />} style={{ padding: 0 }}>Clear Logs</Button>
            </div>
        }
      >
          {/* Virtual Table Header - Fixed widths */}
          <div style={{ display: 'grid', gridTemplateColumns: '80px 70px 60px 1fr 80px 80px', padding: '8px 12px', background: '#fafafa', borderBottom: '1px solid #f0f0f0', fontWeight: 'bold', fontSize: '12px', color: '#666' }}>
             <div>Time</div>
             <div>Method</div>
             <div>Stat</div>
             <div>URL</div>
             <div>Type</div>
             <div>Size</div>
          </div>

          <div 
            ref={parentRef} 
            style={{ flex: 1, overflow: 'auto', position: 'relative' }}
          >
             <div style={{ height: `${rowVirtualizer.getTotalSize()}px`, width: '100%', position: 'relative' }}>
                {rowVirtualizer.getVirtualItems().map((virtualRow) => {
                   const record = filteredLogs[virtualRow.index];
                   
                   return (
                     <div
                        key={virtualRow.key}
                        data-index={virtualRow.index}
                        ref={rowVirtualizer.measureElement}
                        style={{
                          position: 'absolute',
                          top: 0,
                          left: 0,
                          width: '100%',
                          transform: `translateY(${virtualRow.start}px)`,
                          background: virtualRow.index % 2 === 0 ? '#fff' : '#fafafa',
                          borderBottom: '1px solid #f0f0f0',
                        }}
                     >
                        {/* Main Row Content */}
                        <div 
                            onClick={() => { setSelectedLog(record); setDetailsDrawerOpen(true); }}
                            style={{ 
                                display: 'grid', 
                                gridTemplateColumns: '80px 70px 60px 1fr 80px 80px', 
                                padding: '6px 12px', 
                                fontSize: '12px', 
                                cursor: 'pointer',
                                alignItems: 'center'
                            }}
                        >
                            <div style={{ whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', color: '#888' }}>{record.time.split(' ')[1]}</div>
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
                            </div>
                            <div title={record.url} style={{ whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', color: '#1890ff' }}>
                                {record.url}
                                {record.isHttps && <LockOutlined style={{ fontSize: '10px', marginLeft: 4, color: '#52c41a' }} />}
                            </div>
                            <div style={{ color: '#888', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>{record.contentType?.split(';')[0].split('/')[1] || '-'}</div>
                            <div style={{ fontFamily: 'monospace', color: '#666' }}>{formatBytes(record.bodySize || 0)}</div>
                        </div>
                     </div>
                   );
                })}
             </div>
          </div>
      </Card>
      
      <Drawer
        title="Request Details"
        placement="right"
        onClose={() => setDetailsDrawerOpen(false)}
        open={detailsDrawerOpen}
        width="50%"
        style={{ minWidth: 500 }}
        bodyStyle={{ padding: '0' }}
      >
        {selectedLog && (
            <div style={{ padding: 16 }}>
                <Space direction="vertical" size="middle" style={{ width: '100%' }}>
                    <div>
                        <Text type="secondary" style={{ fontSize: 12 }}>URL</Text>
                        <div style={{ wordBreak: 'break-all', fontFamily: 'monospace', background: '#f5f5f5', padding: 8, borderRadius: 4, display: 'flex', alignItems: 'flex-start', gap: 8 }}>
                            <Tag color={selectedLog.method === 'GET' ? 'green' : 'blue'} style={{ flexShrink: 0 }}>{selectedLog.method}</Tag>
                            <Text copyable={{ text: selectedLog.url }} style={{ fontFamily: 'monospace', fontSize: '13px', flex: 1, wordBreak: 'break-all' }}>{selectedLog.url}</Text>
                        </div>
                    </div>

                    {(selectedLog.method === 'CONNECT') ? (
                       <div style={{ textAlign: 'center', padding: '40px 20px', background: '#f5f5f5', borderRadius: 8 }}>
                          <LockOutlined style={{ fontSize: 48, color: '#bfbfbf', marginBottom: 16 }} />
                          <br/>
                          <Text type="secondary" style={{ fontStyle: 'italic' }}>
                             HTTPS Tunnel (Master Connection). <br/>
                             Actual data requests within this tunnel will appear as separate rows if MITM is enabled.
                          </Text>
                       </div>
                    ) : (() => {
                       let queryParams: [string, string][] = [];
                       try {
                           const urlObj = new URL(selectedLog.url);
                           queryParams = Array.from(urlObj.searchParams.entries());
                       } catch (e) {
                           // Fallback for non-standard URLs
                           if (selectedLog.url.includes('?')) {
                               const search = selectedLog.url.split('?')[1];
                               queryParams = search.split('&').map(p => {
                                   const [k, v] = p.split('=');
                                   return [decodeURIComponent(k), decodeURIComponent(v || '')];
                               });
                           }
                       }
                       return (
                        <Tabs defaultActiveKey="request" items={[
                          {
                            key: 'request',
                            label: 'Request',
                            children: (
                              <Space direction="vertical" size="middle" style={{ width: '100%', paddingTop: 12 }}>
                                 {/* Query Params */}
                                 {queryParams.length > 0 && (
                                    <div>
                                       <div style={{ marginBottom: 8, fontSize: 12, fontWeight: 'bold', color: '#999', borderBottom: '1px solid #eee', paddingBottom: 4 }}>QUERY PARAMS</div>
                                       <div style={{ display: 'grid', gridTemplateColumns: 'auto 1fr', gap: '4px 12px' }}>
                                          {queryParams.map(([k, v], idx) => (
                                              <React.Fragment key={idx}>
                                                  <Text style={{ fontSize: '12px', color: '#888', textAlign: 'right', fontWeight: 500 }}>{k}:</Text>
                                                  <Text copyable={{ text: v }} style={{ fontSize: '12px', fontFamily: 'monospace', wordBreak: 'break-all', color: '#1677ff' }}>{v}</Text>
                                              </React.Fragment>
                                          ))}
                                       </div>
                                    </div>
                                 )}

                                 {/* Request Headers */}
                                 {selectedLog.headers && Object.keys(selectedLog.headers).length > 0 && (
                                    <div>
                                       <div style={{ marginBottom: 8, fontSize: 12, fontWeight: 'bold', color: '#999', borderBottom: '1px solid #eee', paddingBottom: 4 }}>HEADERS</div>
                                       <div style={{ display: 'grid', gridTemplateColumns: 'auto 1fr', gap: '4px 12px' }}>
                                          {Object.entries(selectedLog.headers).map(([k, v]) => (
                                              <React.Fragment key={k}>
                                                  <Text style={{ fontSize: '12px', color: '#888', textAlign: 'right', fontWeight: 500 }}>{k}:</Text>
                                                  <Text copyable={{ text: (v as string[]).join(', ') }} style={{ fontSize: '12px', fontFamily: 'monospace', wordBreak: 'break-all' }}>{(v as string[]).join(', ')}</Text>
                                              </React.Fragment>
                                          ))}
                                       </div>
                                    </div>
                                 )}
                                 
                                 {/* Request Body */}
                                 {selectedLog.previewBody && (
                                    <div>
                                         <div style={{ marginBottom: 8, fontSize: 12, fontWeight: 'bold', color: '#999', borderBottom: '1px solid #eee', paddingBottom: 4 }}>BODY</div>
                                         <div style={{ position: 'relative' }}>
                                             <div style={{ 
                                                 padding: '12px', 
                                                 background: '#fafafa', 
                                                 border: '1px solid #eee', 
                                                 borderRadius: '4px',
                                                 fontFamily: 'monospace', 
                                                 fontSize: '12px',
                                                 whiteSpace: 'pre-wrap',
                                                 overflow: 'auto',
                                                 maxHeight: '400px',
                                                 color: '#333'
                                             }}>
                                                 {formatBody(selectedLog.previewBody)}
                                             </div>
                                             <div style={{ position: 'absolute', top: 8, right: 8 }}>
                                                 <Text copyable={{ text: formatBody(selectedLog.previewBody) }} />
                                             </div>
                                         </div>
                                     </div>
                                  )}
                                  {!selectedLog.previewBody && (!selectedLog.headers || Object.keys(selectedLog.headers).length === 0) && (
                                      <Text type="secondary" style={{ display: 'block', textAlign: 'center', margin: '40px 0' }}>No request data.</Text>
                                  )}
                              </Space>
                            )
                          },
                          {
                            key: 'response',
                            label: `Response ${selectedLog.statusCode ? '(' + selectedLog.statusCode + ')' : ''}`,
                            children: selectedLog.statusCode ? (
                               <Space direction="vertical" size="middle" style={{ width: '100%', paddingTop: 12 }}>
                                  <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
                                     <Tag color={selectedLog.statusCode >= 400 ? 'error' : 'success'} style={{ fontSize: 14, padding: '2px 8px' }}>
                                        {selectedLog.statusCode}
                                     </Tag>
                                     {selectedLog.contentType && <Tag>{selectedLog.contentType.split(';')[0]}</Tag>}
                                  </div>

                                  {/* Response Headers */}
                                  {selectedLog.respHeaders && Object.keys(selectedLog.respHeaders).length > 0 && (
                                     <div>
                                        <div style={{ marginBottom: 8, fontSize: 12, fontWeight: 'bold', color: '#999', borderBottom: '1px solid #eee', paddingBottom: 4 }}>HEADERS</div>
                                        <div style={{ display: 'grid', gridTemplateColumns: 'auto 1fr', gap: '4px 12px' }}>
                                           {Object.entries(selectedLog.respHeaders).map(([k, v]) => (
                                               <React.Fragment key={k}>
                                                   <Text style={{ fontSize: '12px', color: '#888', textAlign: 'right', fontWeight: 500 }}>{k}:</Text>
                                                   <Text copyable={{ text: (v as string[]).join(', ') }} style={{ fontSize: '12px', fontFamily: 'monospace', wordBreak: 'break-all' }}>{(v as string[]).join(', ')}</Text>
                                               </React.Fragment>
                                           ))}
                                        </div>
                                     </div>
                                  )}

                                  {/* Response Body */}
                                  {selectedLog.respBody && (
                                     <div>
                                          <div style={{ marginBottom: 8, fontSize: 12, fontWeight: 'bold', color: '#999', borderBottom: '1px solid #eee', paddingBottom: 4 }}>BODY</div>
                                          <div style={{ position: 'relative' }}>
                                              <div style={{ 
                                                  padding: '12px', 
                                                  background: '#f9f9f9', 
                                                  border: '1px solid #eee', 
                                                  borderRadius: '4px',
                                                  fontFamily: 'monospace', 
                                                  fontSize: '12px',
                                                  whiteSpace: 'pre-wrap',
                                                  overflow: 'auto',
                                                  maxHeight: '500px',
                                                  color: '#333'
                                              }}>
                                                  {formatBody(selectedLog.respBody)}
                                              </div>
                                              <div style={{ position: 'absolute', top: 8, right: 8 }}>
                                                  <Text copyable={{ text: formatBody(selectedLog.respBody) }} />
                                              </div>
                                          </div>
                                      </div>
                                   )}
                                   {!selectedLog.respBody && (!selectedLog.respHeaders || Object.keys(selectedLog.respHeaders).length === 0) && (
                                      <Text type="secondary" style={{ display: 'block', textAlign: 'center', margin: '40px 0' }}>No response data.</Text>
                                   )}
                               </Space>
                            ) : (
                               <div style={{ padding: '80px 20px', textAlign: 'center' }}>
                                 <HourglassOutlined style={{ fontSize: 48, color: '#bfbfbf', marginBottom: 16 }} />
                                 <br/>
                                 <Text type="secondary" italic>Waiting for server response...</Text>
                                </div>
                            )
                          }
                        ]} />
                       );
                    })()}
                </Space>
            </div>
        )}
      </Drawer>

      <Modal
        title={
            <Space>
                <SettingOutlined />
                <span>MITM Bypass Rules</span>
            </Space>
        }
        open={isBypassModalOpen}
        onCancel={() => setIsBypassModalOpen(false)}
        footer={[
            <Button key="close" type="primary" onClick={() => setIsBypassModalOpen(false)}>Close</Button>
        ]}
      >
        <div style={{ marginBottom: 16 }}>
            <Text type="secondary">
                Add host keywords (e.g. "cdn", "static") that should bypass HTTPS decryption. 
                This is necessary for domains using Certificate Pinning (like TikTok images).
            </Text>
        </div>
        
        <Space.Compact style={{ width: '100%', marginBottom: 16 }}>
            <Input 
                placeholder="Add new keyword..." 
                value={newPattern} 
                onChange={e => setNewPattern(e.target.value)}
                onPressEnter={() => {
                    if (newPattern && !bypassPatterns.includes(newPattern)) {
                        const next = [...bypassPatterns, newPattern];
                        setBypassPatterns(next);
                        SetMITMBypassPatterns(next);
                        setNewPattern("");
                    }
                }}
            />
            <Button type="primary" onClick={() => {
                if (newPattern && !bypassPatterns.includes(newPattern)) {
                    const next = [...bypassPatterns, newPattern];
                    setBypassPatterns(next);
                    SetMITMBypassPatterns(next);
                    setNewPattern("");
                }
            }}>Add</Button>
        </Space.Compact>

        <div style={{ minHeight: 100, padding: 8, border: '1px solid #f0f0f0', borderRadius: 4, background: '#fafafa' }}>
            <Space wrap>
                {bypassPatterns.map(pat => (
                    <Tag 
                        key={pat} 
                        closable 
                        onClose={() => {
                            const next = bypassPatterns.filter(p => p !== pat);
                            setBypassPatterns(next);
                            SetMITMBypassPatterns(next);
                        }}
                    >
                        {pat}
                    </Tag>
                ))}
                {bypassPatterns.length === 0 && <Text type="secondary">No bypass rules defined.</Text>}
            </Space>
        </div>
        </Modal>
      
    </div>
  );
};

export default ProxyView;
