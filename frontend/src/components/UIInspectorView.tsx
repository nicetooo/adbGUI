import React, { useState, useEffect } from "react";
import {
    Button,
    Space,
    Card,
    Input,
    message,
    theme,
    Empty,
    Tooltip,
    Divider,
    Tree,
    Select,
} from "antd";
import { useTranslation } from "react-i18next";
import {
    SearchOutlined,
    BlockOutlined,
    ReloadOutlined,
    ArrowDownOutlined,
    InfoCircleOutlined,
    CodeOutlined,
    PlayCircleOutlined,
    CopyOutlined,
} from "@ant-design/icons";
import { useDeviceStore, useAutomationStore } from "../stores";

const DetailItem: React.FC<{
    label: string;
    value: string | boolean;
    copyable?: boolean;
}> = ({ label, value, copyable }) => {
    const { token } = theme.useToken();
    const valStr = String(value ?? "");
    const isFalsey = value === null || value === undefined || value === "" || valStr === "false";

    // Don't hide Clickable/Long Clickable rows just because they are false
    const isActionRow = label === "Clickable" || label === "Long Clickable";
    if (isFalsey && !isActionRow) return null;

    return (
        <div style={{ marginBottom: 10 }}>
            <div style={{ color: token.colorTextSecondary, fontSize: 11, marginBottom: 4 }}>{label}</div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <div style={{
                    wordBreak: 'break-all',
                    backgroundColor: token.colorFillSecondary,
                    padding: '4px 10px',
                    borderRadius: 4,
                    fontSize: 12,
                    flex: 1,
                    minHeight: 28,
                    display: 'flex',
                    alignItems: 'center',
                    border: `1px solid ${token.colorBorderSecondary}`
                }}>
                    {valStr}
                </div>
                {copyable && (
                    <Tooltip title="Copy">
                        <Button
                            type="text"
                            size="small"
                            style={{ width: 28, height: 28, padding: 0 }}
                            icon={<CodeOutlined style={{ fontSize: 14 }} />}
                            onClick={() => {
                                navigator.clipboard.writeText(valStr);
                                message.success("Copied");
                            }}
                        />
                    </Tooltip>
                )}
            </div>
        </div>
    );
};

const UIInspectorView: React.FC = () => {
    const { selectedDevice } = useDeviceStore();
    const {
        uiHierarchy,
        isFetchingHierarchy,
        fetchUIHierarchy,
        checkAndRefreshUIHierarchy,
    } = useAutomationStore();

    const { t } = useTranslation();
    const { token } = theme.useToken();

    const [selectedNode, setSelectedNode] = useState<any>(null);
    const [searchText, setSearchText] = useState("");
    const [expandedKeys, setExpandedKeys] = useState<React.Key[]>([]);
    const [autoExpandParent, setAutoExpandParent] = useState(true);
    const [selectedAction, setSelectedAction] = useState("click");
    const [inputText, setInputText] = useState("");

    const generateStepJson = (node: any, type: "click" | "long_click") => {
        const step = {
            type: "check",
            value: node.text ? "text" : (node.resourceId ? "id" : "description"),
            checkValue: node.text || node.resourceId || node.contentDesc || "",
            waitTimeout: 5000,
            onFailure: "stop",
            // We can add the action to perform on success if we extend the task engine
            // For now, it's a template for the user
            label: `${type === 'click' ? 'Click' : 'Long Click'}: ${node.text || node.resourceId || node.contentDesc}`
        };
        return JSON.stringify(step, null, 2);
    };

    useEffect(() => {
        if (selectedDevice && !uiHierarchy) {
            fetchUIHierarchy(selectedDevice);
        }
    }, [selectedDevice]);

    // Memoize tree data to avoid unnecessary recursion on every render
    const treeData = React.useMemo(() => {
        if (!uiHierarchy) return [];

        // Recursive search and highlight
        const processTreeData = (node: any, path: string = "0"): any => {
            if (!node) return null;

            const key = `${path}-${node.class}-${node.bounds}`;
            const children = (node.nodes || []).map((child: any, i: number) => processTreeData(child, `${path}-${i}`)).filter(Boolean);

            const lowerSearch = searchText.toLowerCase().trim();
            const match = lowerSearch && (
                String(node.text || "").toLowerCase().includes(lowerSearch) ||
                String(node.resourceId || "").toLowerCase().includes(lowerSearch) ||
                String(node.class || "").toLowerCase().includes(lowerSearch) ||
                String(node.contentDesc || "").toLowerCase().includes(lowerSearch)
            );

            // If searching but no match and no matching children, prune this branch
            if (lowerSearch && !match && children.length === 0) return null;

            return {
                key,
                title: (
                    <Space size={4}>
                        <span style={{ color: token.colorPrimary, fontSize: 11 }}>[{node.class.split('.').pop()}]</span>
                        {node.text && (
                            <span style={{
                                fontWeight: 500,
                                backgroundColor: match && String(node.text).toLowerCase().includes(lowerSearch) ? token.colorWarningBg : 'transparent'
                            }}>
                                "{node.text}"
                            </span>
                        )}
                        {node.resourceId && (
                            <span style={{
                                color: token.colorTextSecondary,
                                fontSize: 10,
                                backgroundColor: match && String(node.resourceId).toLowerCase().includes(lowerSearch) ? token.colorWarningBg : 'transparent'
                            }}>
                                #{node.resourceId.split('/').pop()}
                            </span>
                        )}
                    </Space>
                ),
                data: node,
                children,
            };
        };

        const processed = processTreeData(uiHierarchy);
        return processed ? [processed] : [];
    }, [uiHierarchy, searchText, token]);

    // Automatically expand results when searching or on first load
    useEffect(() => {
        if (!uiHierarchy) return;

        if (searchText.trim()) {
            const keys: React.Key[] = [];
            const getAllKeys = (data: any[]) => {
                data.forEach(item => {
                    if (item && item.key) {
                        keys.push(item.key);
                        if (item.children) getAllKeys(item.children);
                    }
                });
            };
            getAllKeys(treeData);
            setExpandedKeys(keys);
            setAutoExpandParent(true);
        } else if (expandedKeys.length === 0) {
            // Expand first 2 levels by default if nothing is expanded
            const keys: React.Key[] = [];
            const getInitialKeys = (data: any[], depth: number) => {
                if (depth > 2) return;
                data.forEach(item => {
                    if (item && item.key) {
                        keys.push(item.key);
                        if (item.children) getInitialKeys(item.children, depth + 1);
                    }
                });
            };
            getInitialKeys(treeData, 0);
            setExpandedKeys(keys);
        }
    }, [searchText, treeData, uiHierarchy]);

    return (
        <div style={{ padding: 16, height: "100%", overflow: 'hidden' }}>
            <Card
                title={
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <Space>
                            <BlockOutlined />
                            {t("automation.ui_hierarchy")}
                        </Space>
                        <Space>
                            <Input
                                placeholder={t("common.search")}
                                prefix={<SearchOutlined />}
                                size="small"
                                value={searchText}
                                onChange={e => setSearchText(e.target.value)}
                                style={{ width: 200 }}
                                allowClear
                            />
                            <Button
                                icon={<ReloadOutlined />}
                                size="small"
                                loading={isFetchingHierarchy}
                                onClick={() => selectedDevice && fetchUIHierarchy(selectedDevice)}
                            >
                                {t("common.refresh")}
                            </Button>
                        </Space>
                    </div>
                }
                styles={{ body: { padding: 0, height: 'calc(100% - 48px)', display: 'flex', flexDirection: 'column' } }}
                style={{ height: '100%', display: 'flex', flexDirection: 'column' }}
            >
                <div style={{ display: 'flex', flex: 1, minHeight: 0 }}>
                    <div style={{ flex: 1, overflow: 'auto', padding: 12, borderRight: `1px solid ${token.colorBorderSecondary}` }}>
                        {uiHierarchy ? (
                            <Tree
                                showLine
                                switcherIcon={<ArrowDownOutlined style={{ fontSize: 10 }} />}
                                onSelect={(_, info: any) => {
                                    const node = info.node.data;
                                    setSelectedNode(node);
                                    // Auto select input action for EditText
                                    if (node.class && node.class.toLowerCase().includes("edittext")) {
                                        setSelectedAction("input");
                                    } else {
                                        setSelectedAction("click");
                                    }
                                }}
                                treeData={treeData}
                                expandedKeys={expandedKeys}
                                autoExpandParent={autoExpandParent}
                                onExpand={setExpandedKeys}
                            />
                        ) : (
                            <Empty
                                description={selectedDevice ? t("automation.no_hierarchy_info") : t("app.select_device")}
                                style={{ marginTop: 100 }}
                            />
                        )}
                    </div>
                    <div style={{ width: 350, overflow: 'auto', padding: 12, backgroundColor: token.colorFillAlter }}>
                        {selectedNode ? (
                            <div style={{ fontSize: 13 }}>
                                <h4 style={{ marginBottom: 12, borderBottom: `1px solid ${token.colorBorder}`, paddingBottom: 8 }}>
                                    {t("automation.element_details")}
                                </h4>
                                <DetailItem label="Text" value={selectedNode.text} copyable />
                                <DetailItem label="ID" value={selectedNode.resourceId} copyable />
                                <DetailItem label="Class" value={selectedNode.class} />
                                <DetailItem label="Content Desc" value={selectedNode.contentDesc} />
                                <DetailItem label="Bounds" value={selectedNode.bounds} />
                                <Divider style={{ margin: '12px 0' }} />
                                <DetailItem label="Clickable" value={selectedNode.clickable} />
                                <DetailItem label="Enabled" value={selectedNode.enabled} />
                                <DetailItem label="Focused" value={selectedNode.focused} />
                                <DetailItem label="Scrollable" value={selectedNode.scrollable} />
                                <DetailItem label="Checkable" value={selectedNode.checkable} />
                                <DetailItem label="Long Clickable" value={selectedNode.longClickable} />

                                <div style={{ marginTop: 20 }}>
                                    <div style={{ color: token.colorTextSecondary, fontSize: 11, marginBottom: 8, fontWeight: 500 }}>
                                        {t("automation.node_actions")}
                                    </div>
                                    <Space.Compact style={{ width: '100%', marginBottom: 8 }}>
                                        <Select
                                            value={selectedAction}
                                            onChange={setSelectedAction}
                                            style={{ width: '40%' }}
                                            size="small"
                                            options={[
                                                { label: t("automation.click"), value: "click" },
                                                { label: t("automation.long_click"), value: "long_click" },
                                                { label: t("automation.swipe_up"), value: "swipe_up" },
                                                { label: t("automation.swipe_down"), value: "swipe_down" },
                                                { label: t("automation.swipe_left"), value: "swipe_left" },
                                                { label: t("automation.swipe_right"), value: "swipe_right" },
                                                { label: t("automation.back"), value: "back" },
                                                { label: t("automation.home"), value: "home" },
                                                { label: t("automation.recent"), value: "recent" },
                                                { label: t("automation.input"), value: "input" }
                                            ]}
                                        />
                                        <Button
                                            type="primary"
                                            size="small"
                                            style={{ width: '60%' }}
                                            icon={<PlayCircleOutlined />}
                                            onClick={async () => {
                                                if (selectedDevice && selectedNode.bounds) {
                                                    try {
                                                        if (selectedAction === "input") {
                                                            await (window as any).go.main.App.InputNodeText(selectedDevice, selectedNode.bounds, inputText);
                                                        } else {
                                                            await (window as any).go.main.App.PerformNodeAction(selectedDevice, selectedNode.bounds, selectedAction);
                                                        }
                                                        message.success(t("automation.action_success"));
                                                        // Background refresh with diffing
                                                        setTimeout(async () => {
                                                            if (selectedDevice) {
                                                                const changed = await checkAndRefreshUIHierarchy(selectedDevice);
                                                                if (changed) {
                                                                    message.info(t("automation.ui_updated"));
                                                                }
                                                            }
                                                        }, 800);
                                                    } catch (err) {
                                                        message.error(String(err));
                                                    }
                                                }
                                            }}
                                        >
                                            {t("automation.simulate")}
                                        </Button>
                                    </Space.Compact>

                                    {selectedAction === "input" && (
                                        <Input
                                            size="small"
                                            placeholder={t("automation.input_placeholder")}
                                            value={inputText}
                                            onChange={e => setInputText(e.target.value)}
                                            style={{ marginBottom: 8 }}
                                            onPressEnter={async () => {
                                                if (selectedDevice && selectedNode.bounds && inputText) {
                                                    try {
                                                        await (window as any).go.main.App.InputNodeText(selectedDevice, selectedNode.bounds, inputText);
                                                        message.success(t("automation.action_success"));
                                                        // Background refresh with diffing
                                                        setTimeout(async () => {
                                                            if (selectedDevice) {
                                                                const changed = await checkAndRefreshUIHierarchy(selectedDevice);
                                                                if (changed) {
                                                                    message.info(t("automation.ui_updated"));
                                                                }
                                                            }
                                                        }, 800);
                                                    } catch (err) {
                                                        message.error(String(err));
                                                    }
                                                }
                                            }}
                                        />
                                    )}

                                    <Button
                                        block
                                        size="small"
                                        icon={<CopyOutlined />}
                                        onClick={() => {
                                            navigator.clipboard.writeText(generateStepJson(selectedNode, selectedAction as any));
                                            message.success(t("automation.copy_step_success"));
                                        }}
                                    >
                                        {t("automation.copy_as_step")}
                                    </Button>

                                    <Divider style={{ margin: '16px 0' }} />

                                    <Button
                                        type="primary"
                                        ghost
                                        block
                                        size="small"
                                        onClick={() => {
                                            const value = selectedNode.text || selectedNode.resourceId || selectedNode.contentDesc;
                                            if (value) {
                                                navigator.clipboard.writeText(value);
                                                message.success(t("common.copy_success", { name: value }));
                                            }
                                        }}
                                    >
                                        {t("automation.copy_selector")}
                                    </Button>
                                </div>
                            </div>
                        ) : (
                            <div style={{ textAlign: 'center', color: token.colorTextDescription, marginTop: 100 }}>
                                {t("automation.select_element_tip")}
                            </div>
                        )}
                    </div>
                </div>
            </Card>
        </div>
    );
};

export default UIInspectorView;
