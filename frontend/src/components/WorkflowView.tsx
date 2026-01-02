import React, { useState, useEffect, useCallback, useMemo } from "react";
import {
  Button,
  Space,
  Card,
  List,
  Modal,
  Input,
  message,
  theme,
  Empty,
  Popconfirm,
  Tooltip,
  Tag,
  Form,
  Select,
  InputNumber,
  Collapse,
  Divider,
  Drawer,
  Typography
} from "antd";
import { useTranslation } from "react-i18next";
import {
  PlayCircleOutlined,
  StopOutlined,
  DeleteOutlined,
  PlusOutlined,
  EditOutlined,
  RobotOutlined,
  AimOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  SwapOutlined,
  FormOutlined,
  BranchesOutlined,
  ReloadOutlined,
  PauseCircleOutlined,
  CaretRightOutlined,
  CopyOutlined,
  SaveOutlined,
  ArrowLeftOutlined,
  HomeOutlined,
  AppstoreOutlined,
  PoweroffOutlined,
  SoundOutlined,
  ExpandOutlined,
  LockOutlined,
} from "@ant-design/icons";
import {
  ReactFlow,
  MiniMap,
  Controls,
  Background,
  useNodesState,
  useEdgesState,
  addEdge,
  Handle,
  Position,
  Panel,
  Connection,
  Edge,
  Node,
  BackgroundVariant,
  MarkerType
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import dagre from 'dagre';

import DeviceSelector from "./DeviceSelector";
import ElementPicker, { ElementSelector } from "./ElementPicker";
import { useDeviceStore, useAutomationStore } from "../stores";

const { Text, Title } = Typography;

// Workflow types
interface WorkflowStep {
  id: string;
  type: string;
  name?: string;
  selector?: ElementSelector;
  value?: string;
  timeout?: number;
  onError?: string;
  loop?: number;
  postDelay?: number;
}

interface Workflow {
  id: string;
  name: string;
  description?: string;
  steps: WorkflowStep[];
  createdAt: string;
  updatedAt: string;
}

// Step type definitions
const STEP_TYPES = {
  ELEMENT_ACTIONS: [
    { key: 'click_element', icon: <AimOutlined />, color: 'blue' },
    { key: 'long_click_element', icon: <AimOutlined />, color: 'blue' },
    { key: 'input_text', icon: <FormOutlined />, color: 'cyan' },
    { key: 'swipe_element', icon: <SwapOutlined />, color: 'purple' },
  ],
  WAIT_CONDITIONS: [
    { key: 'wait_element', icon: <ClockCircleOutlined />, color: 'orange' },
    { key: 'wait_gone', icon: <ClockCircleOutlined />, color: 'orange' },
    { key: 'wait', icon: <ClockCircleOutlined />, color: 'default' },
  ],
  FLOW_CONTROL: [
    { key: 'if_exists', icon: <BranchesOutlined />, color: 'green' },
    { key: 'scroll_to', icon: <ReloadOutlined />, color: 'magenta' },
    { key: 'assert_element', icon: <CheckCircleOutlined />, color: 'lime' },
  ],
  SCRIPT_ACTIONS: [
    { key: 'script', icon: <PlayCircleOutlined />, color: 'geekblue' },
    { key: 'adb', icon: <RobotOutlined />, color: 'volcano' },
  ],
  SYSTEM_ACTIONS: [
    { key: 'key_back', icon: <ArrowLeftOutlined />, color: 'default' },
    { key: 'key_home', icon: <HomeOutlined />, color: 'default' },
    { key: 'key_recent', icon: <AppstoreOutlined />, color: 'default' },
    { key: 'key_power', icon: <PoweroffOutlined />, color: 'red' },
    { key: 'key_volume_up', icon: <SoundOutlined />, color: 'default' },
    { key: 'key_volume_down', icon: <SoundOutlined />, color: 'default' },
    { key: 'screen_on', icon: <ExpandOutlined />, color: 'default' },
    { key: 'screen_off', icon: <LockOutlined />, color: 'default' },
  ],
  NESTED: [
    { key: 'run_workflow', icon: <BranchesOutlined />, color: 'gold' },
  ],
};

const getStepTypeInfo = (type: string) => {
  for (const category of Object.values(STEP_TYPES)) {
    const found = category.find(s => s.key === type);
    if (found) return found;
  }
  return { key: type, icon: <RobotOutlined />, color: 'default' };
};

// Custom Node Component
const WorkflowNode = ({ data, selected }: any) => {
  const { step, isRunning, isCurrent } = data;
  const { token } = theme.useToken();
  const typeInfo = getStepTypeInfo(step.type);

  return (
    <div style={{ position: 'relative' }}>
      <Handle type="target" position={Position.Top} style={{ background: token.colorTextSecondary }} />
      <Card
        size="small"
        style={{
          width: 280,
          borderColor: isCurrent ? token.colorPrimary : selected ? token.colorPrimary : token.colorBorder,
          borderWidth: isCurrent || selected ? 2 : 1,
          boxShadow: isCurrent ? `0 0 10px ${token.colorPrimary}` : selected ? token.boxShadowSecondary : 'none',
          backgroundColor: isCurrent ? token.colorPrimaryBg : token.colorBgContainer, // Ensure container background
          color: token.colorText, // Explicit text color
        }}
        bodyStyle={{ padding: '8px 12px' }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <div style={{
            backgroundColor: isCurrent ? token.colorPrimaryBgHover : token.colorBgLayout, // Softer background for icon
            padding: 6,
            borderRadius: 6,
            display: 'flex',
            // color: typeInfo.color === 'default' ? token.colorText : token[(typeInfo.color + '6') as keyof typeof token] || typeInfo.color 
            // Simplified color mapping for better dark/light compatibility:
            color: typeInfo.color === 'default' ? token.colorText : typeInfo.color
          }}>
            {typeInfo.icon}
          </div>
          <div style={{ flex: 1, overflow: 'hidden' }}>
            <Text strong ellipsis style={{ color: token.colorText }}>{data.label || step.type}</Text>
            {step.selector && (
              <div style={{ fontSize: 11, color: token.colorTextSecondary, display: 'flex', alignItems: 'center', gap: 4 }}>
                <Tag style={{ margin: 0, fontSize: 10, lineHeight: '16px', padding: '0 4px' }}>{step.selector.type}</Tag>
                <Text ellipsis style={{ fontSize: 11, maxWidth: 120, color: token.colorTextSecondary }}>{step.selector.value}</Text>
              </div>
            )}
            {step.value && (
              <div style={{ fontSize: 11, color: token.colorTextSecondary }}>
                <Text ellipsis style={{ fontSize: 11, color: token.colorTextSecondary }}>{step.value}</Text>
              </div>
            )}
          </div>
        </div>
      </Card>
      <Handle type="source" position={Position.Bottom} style={{ background: token.colorTextSecondary }} />
    </div>
  );
};

const WorkflowView: React.FC = () => {
  const { selectedDevice } = useDeviceStore();
  const { scripts } = useAutomationStore();
  const { t } = useTranslation();
  const { token } = theme.useToken();

  const [workflows, setWorkflows] = useState<Workflow[]>([]);
  const [selectedWorkflow, setSelectedWorkflow] = useState<Workflow | null>(null);
  const [workflowModalVisible, setWorkflowModalVisible] = useState(false);
  const [workflowForm] = Form.useForm();

  // Execution state
  const [isRunning, setIsRunning] = useState(false);
  const [currentStepIndex, setCurrentStepIndex] = useState(-1);
  const [executionLogs, setExecutionLogs] = useState<string[]>([]);

  // React Flow state
  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);

  // Node Editing
  const [editingNodeId, setEditingNodeId] = useState<string | null>(null);
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [stepForm] = Form.useForm();

  // Element Picker
  const [elementPickerVisible, setElementPickerVisible] = useState(false);

  const nodeTypes = useMemo(() => ({ workflowNode: WorkflowNode }), []);

  // Load workflows
  useEffect(() => {
    loadWorkflows();
  }, []);

  const loadWorkflows = async () => {
    try {
      const result = await (window as any).go.main.App.LoadWorkflows();
      setWorkflows(result || []);
    } catch (err) {
      setWorkflows([]);
    }
  };

  // Layout Graph
  const getLayoutedElements = useCallback((nodes: Node[], edges: Edge[]) => {
    const dagreGraph = new dagre.graphlib.Graph();
    dagreGraph.setDefaultEdgeLabel(() => ({}));
    const nodeWidth = 280;
    const nodeHeight = 80;

    dagreGraph.setGraph({ rankdir: 'TB' });

    nodes.forEach((node) => {
      dagreGraph.setNode(node.id, { width: nodeWidth, height: nodeHeight });
    });

    edges.forEach((edge) => {
      dagreGraph.setEdge(edge.source, edge.target);
    });

    dagre.layout(dagreGraph);

    const layoutedNodes = nodes.map((node) => {
      const nodeWithPosition = dagreGraph.node(node.id);
      return {
        ...node,
        position: {
          x: nodeWithPosition.x - nodeWidth / 2,
          y: nodeWithPosition.y - nodeHeight / 2,
        },
      };
    });

    return { nodes: layoutedNodes, edges };
  }, []);

  // Convert Workflow Steps to Graph
  useEffect(() => {
    if (selectedWorkflow) {
      const newNodes: Node[] = selectedWorkflow.steps.map((step, index) => ({
        id: step.id,
        type: 'workflowNode',
        position: { x: 0, y: 0 }, // Initial position, will belayouted
        data: {
          step,
          label: step.name || t(`workflow.step_type.${step.type}`),
          isCurrent: index === currentStepIndex
        },
      }));

      const newEdges: Edge[] = selectedWorkflow.steps.slice(0, -1).map((step, index) => ({
        id: `e-${step.id}-${selectedWorkflow.steps[index + 1].id}`,
        source: step.id,
        target: selectedWorkflow.steps[index + 1].id,
        type: 'smoothstep',
        markerEnd: { type: MarkerType.ArrowClosed },
      }));

      const layouted = getLayoutedElements(newNodes, newEdges);
      setNodes(layouted.nodes);
      setEdges(layouted.edges);
    } else {
      setNodes([]);
      setEdges([]);
    }
  }, [selectedWorkflow, getLayoutedElements, t]); // Removed currentStepIndex dependency to prevent layout reset on run

  // Update visual state when running
  useEffect(() => {
    setNodes((nds) =>
      nds.map((node) => {
        const stepIndex = selectedWorkflow?.steps.findIndex(s => s.id === node.id);
        const isCurrent = stepIndex === currentStepIndex;
        // Only update if changed
        if (node.data.isCurrent !== isCurrent) {
          return {
            ...node,
            data: { ...(node.data as object), isCurrent, isRunning }
          }
        }
        return node;
      })
    );
  }, [currentStepIndex, isRunning, selectedWorkflow]);


  const onConnect = useCallback(
    (params: Connection) => setEdges((eds) => addEdge({ ...params, type: 'smoothstep', markerEnd: { type: MarkerType.ArrowClosed } }, eds)),
    [setEdges],
  );

  const handleNodeClick = (event: React.MouseEvent, node: Node) => {
    setEditingNodeId(node.id);
    const step = node.data.step as WorkflowStep;
    stepForm.resetFields();
    stepForm.setFieldsValue({
      ...step,
      selectorType: step.selector?.type,
      selectorValue: step.selector?.value,
      type: step.type // ensure type is set for conditional rendering
    });
    setDrawerVisible(true);
  };

  const handleCreateWorkflow = () => {
    workflowForm.resetFields();
    setSelectedWorkflow(null);
    setWorkflowModalVisible(true);
  };

  const handleCreateWorkflowSubmit = async (values: any) => {
    const newWorkflow: Workflow = {
      id: `wf_${Date.now()}`,
      name: values.name,
      description: values.description,
      steps: [],
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
    // Save immediatelly
    try {
      await (window as any).go.main.App.SaveWorkflow(newWorkflow);
      message.success(t("workflow.saved"));
      setWorkflowModalVisible(false);
      setWorkflows(prev => [...prev, newWorkflow]);
      setSelectedWorkflow(newWorkflow);
    } catch (err) {
      message.error(String(err));
    }
  }

  const handleSaveGraph = async (silent: boolean = false): Promise<Workflow | null> => {
    if (!selectedWorkflow) return null;

    const nodeMap = new Map(nodes.map(n => [n.id, n]));
    const outEdgesMap = new Map<string, string>(); // source -> target
    edges.forEach(e => outEdgesMap.set(e.source, e.target));

    // Find start node
    const incomings = new Set(edges.map(e => e.target));
    const starts = nodes.filter(n => !incomings.has(n.id));

    if (starts.length === 0 && nodes.length > 0) {
      message.warning("Cycle detected or no clear start. Saving best guess.");
    }

    const orderedSteps: WorkflowStep[] = [];
    let currentId: string | undefined = starts.length > 0 ? starts[0].id : (nodes[0]?.id);

    const visited = new Set<string>();
    while (currentId && !visited.has(currentId)) {
      visited.add(currentId);
      const node = nodeMap.get(currentId);
      if (node) {
        orderedSteps.push(node.data.step as WorkflowStep);
      }
      currentId = outEdgesMap.get(currentId);
    }

    const updatedWorkflow = {
      ...selectedWorkflow,
      steps: orderedSteps,
      updatedAt: new Date().toISOString(),
    };

    try {
      await (window as any).go.main.App.SaveWorkflow(updatedWorkflow);
      setSelectedWorkflow(updatedWorkflow);
      if (!silent) {
        message.success(t("workflow.saved"));
      }
      loadWorkflows();
      return updatedWorkflow;
    } catch (err) {
      if (!silent) {
        message.error(String(err));
      }
      return null;
    }
  };

  const handleAddStep = (type: string) => {
    if (!selectedWorkflow) return;

    const id = `step_${Date.now()}`;
    const newStep: WorkflowStep = {
      id,
      type,
      loop: 1,
      postDelay: 0,
      onError: 'stop'
    };

    const newNode: Node = {
      id,
      type: 'workflowNode',
      position: { x: 250, y: (nodes.length * 100) + 50 }, // Simple stacking
      data: {
        step: newStep,
        label: t(`workflow.step_type.${type}`),
        isCurrent: false
      }
    };

    // Auto connect to last node if exists
    if (nodes.length > 0) {
      const hasOutgoing = new Set(edges.map(e => e.source));
      const tails = nodes.filter(n => !hasOutgoing.has(n.id));

      if (tails.length === 1) {
        const lastNode = tails[0];
        const newEdge = {
          id: `e-${lastNode.id}-${id}`,
          source: lastNode.id,
          target: id,
          type: 'smoothstep',
          markerEnd: { type: MarkerType.ArrowClosed },
        };
        setEdges(eds => [...eds, newEdge]);
        newNode.position = { x: lastNode.position.x, y: lastNode.position.y + 120 };
      }
    }

    setNodes(nds => [...nds, newNode]);
    setEditingNodeId(id);
    stepForm.resetFields();
    stepForm.setFieldsValue({ type, onError: 'stop', loop: 1 });
    setDrawerVisible(true);
  };

  const handleUpdateStep = (values: any) => {
    if (!editingNodeId) return;

    setNodes(nds => nds.map(node => {
      if (node.id === editingNodeId) {
        const updatedStep: WorkflowStep = {
          ...(node.data.step as WorkflowStep),
          type: values.type,
          name: values.name,
          selector: values.selectorValue ? {
            type: values.selectorType || 'text',
            value: values.selectorValue,
          } : undefined,
          value: values.value,
          timeout: values.timeout,
          onError: values.onError,
          loop: values.loop,
          postDelay: values.postDelay,
        };
        return {
          ...node,
          data: {
            ...node.data,
            step: updatedStep,
            label: values.name || t(`workflow.step_type.${values.type}`)
          }
        };
      }
      return node;
    }));
    message.success(t("workflow.step_updated"));
  };

  const handleDeleteNode = () => {
    if (!editingNodeId) return;
    setNodes(nds => nds.filter(n => n.id !== editingNodeId));
    setEdges(eds => eds.filter(e => e.source !== editingNodeId && e.target !== editingNodeId));
    setDrawerVisible(false);
    setEditingNodeId(null);
  }

  const handleElementSelected = (selector: ElementSelector) => {
    stepForm.setFieldsValue({
      selectorType: selector.type,
      selectorValue: selector.value,
    });
    setElementPickerVisible(false);
    // Trigger form update
    handleUpdateStep(stepForm.getFieldsValue());
    message.success(t("workflow.selector_applied"));
  };

  // Runners
  const handleRunWorkflow = async () => {
    if (!selectedDevice || !selectedWorkflow) {
      message.warning(t("app.select_device"));
      return;
    }

    const deviceObj = useDeviceStore.getState().devices.find(d => d.id === selectedDevice);
    if (!deviceObj) {
      message.error("Device not found");
      return;
    }

    const workflowToRun = await handleSaveGraph(true);
    if (!workflowToRun) return;

    if (workflowToRun.steps.length === 0) {
      message.warning("Run aborted: Workflow has 0 steps. Please check graph connections.");
      return;
    }

    // Auto-fix legacy commands: prepend 'shell' if missing for 'input' commands
    // This allows running old workflows without re-converting
    const sanitizedSteps = workflowToRun.steps.map(s => {
      if (s.type === 'adb' && s.value && s.value.trim().startsWith('input ')) {
        return { ...s, value: `shell ${s.value}` };
      }
      return s;
    });
    const sanitizedWorkflow = { ...workflowToRun, steps: sanitizedSteps };

    setIsRunning(true);
    setCurrentStepIndex(0);
    setExecutionLogs([
      `[${new Date().toLocaleTimeString()}] ${t("workflow.started")}: ${sanitizedWorkflow.name}`,
      `[Info] Executing ${sanitizedWorkflow.steps.length} steps.`
    ]);

    // Setup event listeners using a Promise wrapper to wait for completion
    const executionPromise = new Promise<void>((resolve, reject) => {
      const runtime = (window as any).runtime;
      if (!runtime) {
        console.warn("Wails runtime not found");
        // Fallback for dev mode without backend
        setTimeout(resolve, 1000);
        return;
      }

      // Cleanup function
      const cleanUp = () => {
        runtime.EventsOff("workflow-completed", onComplete);
        runtime.EventsOff("workflow-error", onError);
        runtime.EventsOff("workflow-step-running", onStep);
      };

      const onComplete = (data: any) => {
        if (data.deviceId === deviceObj.id) {
          cleanUp();
          resolve();
        }
      };

      const onError = (data: any) => {
        if (data.deviceId === deviceObj.id) {
          cleanUp();
          reject(data.error);
        }
      };

      const onStep = (data: any) => {
        if (data.deviceId === deviceObj.id) {
          setCurrentStepIndex(data.stepIndex);
        }
      };

      runtime.EventsOn("workflow-completed", onComplete);
      runtime.EventsOn("workflow-error", onError);
      runtime.EventsOn("workflow-step-running", onStep);
    });

    try {
      await (window as any).go.main.App.RunWorkflow(deviceObj, sanitizedWorkflow);

      // Wait for the completion event
      await executionPromise;

      setExecutionLogs(prev => [...prev, `[${new Date().toLocaleTimeString()}] ${t("workflow.completed")}`]);
    } catch (err) {
      setExecutionLogs(prev => [...prev, `[${new Date().toLocaleTimeString()}] ${t("workflow.error")}: ${err}`]);
      message.error(String(err));
    } finally {
      setIsRunning(false);
      setCurrentStepIndex(-1);
    }
  };

  const handleStopWorkflow = async () => {
    try {
      await (window as any).go.main.App.StopWorkflow(selectedDevice);
      setIsRunning(false);
      setCurrentStepIndex(-1);
      setExecutionLogs(prev => [...prev, `[${new Date().toLocaleTimeString()}] ${t("workflow.stopped")}`]);
    } catch (err) {
      message.error(String(err));
    }
  };

  const handleDeleteWorkflow = async (id: string) => {
    try {
      await (window as any).go.main.App.DeleteWorkflow(id);
      message.success(t("workflow.deleted"));
      if (selectedWorkflow?.id === id) {
        setSelectedWorkflow(null);
      }
      loadWorkflows();
    } catch (err) {
      message.error(String(err));
    }
  };

  return (
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      {/* Header */}
      <div style={{ padding: "12px 24px", borderBottom: `1px solid ${token.colorBorderSecondary}`, display: "flex", justifyContent: "space-between", alignItems: "center", background: token.colorBgContainer }}>
        <Space>
          <Title level={4} style={{ margin: 0 }}>{t("workflow.title")}</Title>
          <Tag color="green">Visual Editor</Tag>
        </Space>
        <Space>
          <DeviceSelector />
          {selectedWorkflow && (
            <>
              <Button icon={<SaveOutlined />} onClick={() => handleSaveGraph()}>
                {t("workflow.save")}
              </Button>
              {isRunning ? (
                <Button danger icon={<StopOutlined />} onClick={handleStopWorkflow}>
                  {t("workflow.stop")}
                </Button>
              ) : (
                <Button type="primary" icon={<PlayCircleOutlined />} onClick={handleRunWorkflow} disabled={!selectedDevice}>
                  {t("workflow.run")}
                </Button>
              )}
            </>
          )}
        </Space>
      </div>

      <div style={{ flex: 1, display: "flex", overflow: "hidden" }}>
        {/* Left Sidebar: List */}
        <div style={{ width: 250, borderRight: `1px solid ${token.colorBorderSecondary}`, display: "flex", flexDirection: "column", background: token.colorBgContainer }}>
          <div style={{ padding: 12, borderBottom: `1px solid ${token.colorBorderSecondary}`, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Text strong>{t("workflow.workflows")}</Text>
            <Button type="primary" size="small" icon={<PlusOutlined />} onClick={handleCreateWorkflow} />
          </div>
          <div style={{ flex: 1, overflowY: 'auto' }}>
            {workflows.length === 0 ? (
              <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={t("workflow.no_workflows")} />
            ) : (
              <List
                size="small"
                dataSource={workflows}
                renderItem={(workflow) => (
                  <List.Item
                    onClick={() => {
                      if (selectedWorkflow?.id !== workflow.id) {
                        setSelectedWorkflow(workflow);
                      }
                    }}
                    style={{
                      cursor: 'pointer',
                      padding: '12px 16px',
                      borderLeft: selectedWorkflow?.id === workflow.id ? `3px solid ${token.colorPrimary}` : '3px solid transparent',
                      backgroundColor: selectedWorkflow?.id === workflow.id ? token.colorPrimaryBg : undefined
                    }}
                    actions={[
                      <Popconfirm key="del" title={t("workflow.delete_confirm")} onConfirm={(e) => { e?.stopPropagation(); handleDeleteWorkflow(workflow.id); }}>
                        <Button type="text" size="small" danger icon={<DeleteOutlined />} onClick={(e) => e.stopPropagation()} />
                      </Popconfirm>
                    ]}
                  >
                    <div style={{ width: '100%' }}>
                      <div style={{ fontWeight: 500 }}>{workflow.name}</div>
                      <div style={{ fontSize: 12, color: token.colorTextSecondary }}>{workflow.steps.length} steps</div>
                    </div>
                  </List.Item>
                )}
              />
            )}
          </div>
        </div>

        {/* Center: Graph */}
        <div style={{ flex: 1, position: 'relative' }}>
          {selectedWorkflow ? (
            <ReactFlow
              nodes={nodes}
              edges={edges}
              onNodesChange={onNodesChange}
              onEdgesChange={onEdgesChange}
              onConnect={onConnect}
              onNodeClick={handleNodeClick}
              nodeTypes={nodeTypes}
              fitView
              attributionPosition="bottom-right"
              style={{ backgroundColor: token.colorBgContainer }}
            >
              <Background
                color={token.colorTextQuaternary}
                gap={16}
                variant={BackgroundVariant.Dots}
                style={{ backgroundColor: token.colorBgLayout }}
              />
              <Controls
                style={{
                  backgroundColor: token.colorBgElevated,
                  border: `1px solid ${token.colorBorder}`,
                  fill: token.colorText,
                  borderRadius: 4,
                  padding: 2
                }}
              />
              <style>{`
                .react-flow__controls-button {
                  background-color: ${token.colorBgElevated} !important;
                  border-bottom: 1px solid ${token.colorBorder} !important;
                  fill: ${token.colorText} !important;
                }
                .react-flow__controls-button:hover {
                  background-color: ${token.colorBgLayout} !important;
                }
                .react-flow__controls-button:last-child {
                  border-bottom: none !important;
                }
                .react-flow__controls-button svg {
                  fill: ${token.colorText} !important;
                }
              `}</style>
              <MiniMap
                zoomable
                pannable
                style={{
                  backgroundColor: token.colorBgElevated,
                  border: `1px solid ${token.colorBorder}`,
                  borderRadius: 4
                }}
                maskColor={token.colorBgMask}
                nodeColor={token.colorPrimary}
              />
              <Panel position="top-left" style={{ margin: 16 }}>
                <Card
                  size="small"
                  style={{
                    width: 220,
                    boxShadow: token.boxShadowSecondary,
                    backgroundColor: token.colorBgElevated,
                    border: `1px solid ${token.colorBorder}`
                  }}
                  title={t("workflow.add_step")}
                  headStyle={{ borderBottom: `1px solid ${token.colorBorderSecondary}`, minHeight: 36, padding: '0 12px' }}
                  bodyStyle={{ padding: 12 }}
                >
                  <Collapse
                    ghost
                    size="small"
                    accordion
                    expandIconPosition="end"
                    defaultActiveKey={['1']}
                  >
                    <Collapse.Panel header={t("workflow.category.element_actions")} key="1">
                      <Space wrap size={[8, 8]}>
                        {STEP_TYPES.ELEMENT_ACTIONS.map(s => (
                          <Tooltip title={t(`workflow.step_type.${s.key}`)} key={s.key}>
                            <Button size="small" icon={s.icon} onClick={() => handleAddStep(s.key)} />
                          </Tooltip>
                        ))}
                      </Space>
                    </Collapse.Panel>
                    <Collapse.Panel header={t("workflow.category.wait_conditions")} key="2">
                      <Space wrap size={[8, 8]}>
                        {STEP_TYPES.WAIT_CONDITIONS.map(s => (
                          <Tooltip title={t(`workflow.step_type.${s.key}`)} key={s.key}>
                            <Button size="small" icon={s.icon} onClick={() => handleAddStep(s.key)} />
                          </Tooltip>
                        ))}
                      </Space>
                    </Collapse.Panel>
                    <Collapse.Panel header={t("workflow.category.flow_control")} key="3">
                      <Space wrap size={[8, 8]}>
                        {STEP_TYPES.FLOW_CONTROL.map(s => (
                          <Tooltip title={t(`workflow.step_type.${s.key}`)} key={s.key}>
                            <Button size="small" icon={s.icon} onClick={() => handleAddStep(s.key)} />
                          </Tooltip>
                        ))}
                      </Space>
                    </Collapse.Panel>
                    <Collapse.Panel header={t("workflow.category.script_actions")} key="4">
                      <Space wrap size={[8, 8]}>
                        {STEP_TYPES.SCRIPT_ACTIONS.map(s => (
                          <Tooltip title={t(`workflow.step_type.${s.key}`)} key={s.key}>
                            <Button size="small" icon={s.icon} onClick={() => handleAddStep(s.key)} />
                          </Tooltip>
                        ))}
                      </Space>
                    </Collapse.Panel>
                    <Collapse.Panel header={t("workflow.category.system_actions")} key="5">
                      <Space wrap size={[8, 8]}>
                        {STEP_TYPES.SYSTEM_ACTIONS.map(s => (
                          <Tooltip title={t(`workflow.step_type.${s.key}`)} key={s.key}>
                            <Button size="small" icon={s.icon} onClick={() => handleAddStep(s.key)} />
                          </Tooltip>
                        ))}
                      </Space>
                    </Collapse.Panel>
                    <Collapse.Panel header={t("workflow.category.nested")} key="6">
                      <Space wrap size={[8, 8]}>
                        {STEP_TYPES.NESTED.map(s => (
                          <Tooltip title={t(`workflow.step_type.${s.key}`)} key={s.key}>
                            <Button size="small" icon={s.icon} onClick={() => handleAddStep(s.key)} />
                          </Tooltip>
                        ))}
                      </Space>
                    </Collapse.Panel>
                  </Collapse>
                </Card>
              </Panel>

              {/* Execution Log Overlay */}
              {executionLogs.length > 0 && (
                <Panel position="bottom-center" style={{ width: '60%', marginBottom: 20 }}>
                  <Card
                    size="small"
                    style={{
                      maxHeight: 150,
                      overflowY: 'auto',
                      background: 'rgba(0,0,0,0.8)',
                      border: 'none',
                      color: '#fff',
                      position: 'relative'
                    }}
                    bodyStyle={{ padding: 8 }}
                  >
                    <div style={{ position: 'absolute', top: 4, right: 4, cursor: 'pointer' }} onClick={() => setExecutionLogs([])}>
                      <span style={{ fontSize: 16 }}>×</span>
                    </div>
                    {executionLogs.map((log, i) => (
                      <div key={i} style={{ fontSize: 11, fontFamily: 'monospace' }}>{log}</div>
                    ))}
                  </Card>
                </Panel>
              )}
            </ReactFlow>
          ) : (
            <div style={{ height: '100%', display: 'flex', justifyContent: 'center', alignItems: 'center', background: token.colorBgLayout }}>
              <Empty description={t("workflow.select_or_create")}>
                <Button type="primary" onClick={handleCreateWorkflow}>{t("workflow.create")}</Button>
              </Empty>
            </div>
          )}
        </div>

        {/* Right Drawer: Properties */}
        <Drawer
          title={t("workflow.edit_step")}
          placement="right"
          onClose={() => setDrawerVisible(false)}
          open={drawerVisible}
          width={400}
          mask={false}
          extra={
            <Button danger type="text" icon={<DeleteOutlined />} onClick={handleDeleteNode}>{t("workflow.delete")}</Button>
          }
        >
          <Form form={stepForm} layout="vertical" onValuesChange={(_, values) => handleUpdateStep(stepForm.getFieldsValue())}>
            <Form.Item name="type" label={t("workflow.step_type_label")} rules={[{ required: true }]}>
              <Select
                disabled
                options={[
                  ...STEP_TYPES.ELEMENT_ACTIONS.map(s => ({ label: t(`workflow.step_type.${s.key}`), value: s.key })),
                  ...STEP_TYPES.WAIT_CONDITIONS.map(s => ({ label: t(`workflow.step_type.${s.key}`), value: s.key })),
                  ...STEP_TYPES.FLOW_CONTROL.map(s => ({ label: t(`workflow.step_type.${s.key}`), value: s.key })),
                  ...STEP_TYPES.SCRIPT_ACTIONS.map(s => ({ label: t(`workflow.step_type.${s.key}`), value: s.key })),
                  ...STEP_TYPES.SYSTEM_ACTIONS.map(s => ({ label: t(`workflow.step_type.${s.key}`), value: s.key })),
                  ...STEP_TYPES.NESTED.map(s => ({ label: t(`workflow.step_type.${s.key}`), value: s.key })),
                ]}
              />
            </Form.Item>
            <Form.Item name="name" label={t("workflow.name")}>
              <Input placeholder="Custom step name" />
            </Form.Item>

            <Form.Item noStyle shouldUpdate={(prev, curr) => prev.type !== curr.type}>
              {({ getFieldValue }) => {
                const type = getFieldValue('type');
                const needsSelector = ['click_element', 'long_click_element', 'input_text', 'swipe_element', 'wait_element', 'wait_gone', 'if_exists', 'scroll_to', 'assert_element'].includes(type);
                const needsValue = ['input_text', 'wait', 'adb', 'script'].includes(type);
                const isWorkflow = type === 'run_workflow';

                return (
                  <>
                    {needsSelector && (
                      <>
                        <div style={{ display: 'flex', gap: 8, marginBottom: 16 }}>
                          <Button
                            type="primary"
                            ghost
                            icon={<AimOutlined />}
                            onClick={() => setElementPickerVisible(true)}
                            disabled={!selectedDevice}
                            style={{ flex: 1 }}
                          >
                            {t("workflow.pick_element")}
                          </Button>
                        </div>
                        <div style={{ display: 'flex', gap: 8 }}>
                          <Form.Item name="selectorType" label={t("workflow.selector_type")} style={{ width: 100 }}>
                            <Select
                              options={[
                                { label: 'Text', value: 'text' },
                                { label: 'ID', value: 'id' },
                                { label: 'XPath', value: 'xpath' },
                                { label: 'Bounds', value: 'bounds' },
                                { label: t("workflow.selector_advanced"), value: 'advanced' },
                              ]}
                            />
                          </Form.Item>
                          <Form.Item name="selectorValue" label={t("workflow.selector_value")} style={{ flex: 1 }}>
                            <Input placeholder={t("workflow.selector_placeholder")} />
                          </Form.Item>
                        </div>
                      </>
                    )}

                    {isWorkflow && (
                      <Form.Item name="value" label={t("workflow.select_workflow")} rules={[{ required: true }]}>
                        <Select
                          placeholder={t("workflow.select_workflow")}
                          options={workflows
                            .filter(w => w.id !== selectedWorkflow?.id) // Prevent self-selection
                            .map(w => ({ label: w.name, value: w.id }))
                          }
                        />
                      </Form.Item>
                    )}

                    {needsValue && (
                      <Form.Item name="value" label={t("workflow.value")}>
                        {type === 'script' ? (
                          <Select
                            placeholder={t("workflow.select_script")}
                            options={scripts.map(s => ({ label: s.name, value: s.name }))}
                          />
                        ) : type === 'wait' ? (
                          <InputNumber addonAfter="ms" min={100} step={100} style={{ width: '100%' }} />
                        ) : (
                          <Input placeholder={type === 'adb' ? 'shell input keyevent 4' : t("workflow.value_placeholder")} />
                        )}
                      </Form.Item>
                    )}
                  </>
                );
              }}
            </Form.Item>

            <div style={{ display: 'flex', gap: 16 }}>
              <Form.Item name="timeout" label={t("workflow.timeout")} style={{ flex: 1 }}>
                <InputNumber addonAfter="ms" min={100} step={1000} style={{ width: '100%' }} placeholder="5000" />
              </Form.Item>
              <Form.Item name="loop" label={t("workflow.loop")} style={{ width: 100 }}>
                <InputNumber min={1} style={{ width: '100%' }} placeholder="1" />
              </Form.Item>
            </div>

            <Form.Item name="onError" label={t("workflow.on_error")}>
              <Select
                options={[
                  { label: t("workflow.error_stop"), value: 'stop' },
                  { label: t("workflow.error_continue"), value: 'continue' },
                ]}
                placeholder={t("workflow.error_stop")}
              />
            </Form.Item>

            <Button type="primary" block onClick={() => handleUpdateStep(stepForm.getFieldsValue())}>
              {t("workflow.update")}
            </Button>
          </Form>
        </Drawer>
      </div>

      {/* Workflow Creation Modal */}
      <Modal
        title={t("workflow.create")}
        open={workflowModalVisible}
        onOk={() => workflowForm.submit()}
        onCancel={() => setWorkflowModalVisible(false)}
      >
        <Form form={workflowForm} layout="vertical" onFinish={handleCreateWorkflowSubmit}>
          <Form.Item
            name="name"
            label={t("workflow.name")}
            rules={[{ required: true }]}
          >
            <Input placeholder={t("workflow.name_placeholder")} />
          </Form.Item>
          <Form.Item name="description" label={t("workflow.description")}>
            <Input.TextArea placeholder={t("workflow.description_placeholder")} rows={2} />
          </Form.Item>
        </Form>
      </Modal>

      {/* Element Picker Modal */}
      <ElementPicker
        visible={elementPickerVisible}
        onSelect={handleElementSelected}
        onCancel={() => setElementPickerVisible(false)}
      />

    </div>
  );
};

export default WorkflowView;
