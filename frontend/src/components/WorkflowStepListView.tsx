/**
 * WorkflowStepListView - Draggable sortable list view for workflow steps.
 * 
 * Shows workflow steps in linear execution order (following successStepId from start),
 * with drag-and-drop reordering support. Reordering updates the connection graph.
 */
import React, { useMemo, useCallback } from "react";
import { useTranslation } from "react-i18next";
import { theme, Tag, Typography, Tooltip, Button, Empty, message } from "antd";
import {
  HolderOutlined,
  AimOutlined,
  SwapOutlined,
  FormOutlined,
  ClockCircleOutlined,
  PlayCircleOutlined,
  CheckCircleOutlined,
  CaretRightOutlined,
  ForkOutlined,
  ReloadOutlined,
  IdcardOutlined,
  RobotOutlined,
  ArrowLeftOutlined,
  HomeOutlined,
  AppstoreOutlined,
  PoweroffOutlined,
  SoundOutlined,
  ExpandOutlined,
  LockOutlined,
  BranchesOutlined,
  StopOutlined,
  DeleteOutlined,
  SettingOutlined,
  ThunderboltOutlined,
} from "@ant-design/icons";
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  DragEndEvent,
} from "@dnd-kit/core";
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy,
  useSortable,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";

import type { Workflow, WorkflowStep } from "../types/workflow";

const { Text } = Typography;

// Step type icon/color map (must match WorkflowView's STEP_TYPES)
const STEP_TYPE_META: Record<string, { icon: React.ReactNode; color: string }> = {
  start: { icon: <CaretRightOutlined />, color: 'green' },
  tap: { icon: <AimOutlined />, color: 'magenta' },
  swipe: { icon: <SwapOutlined />, color: 'magenta' },
  click_element: { icon: <AimOutlined />, color: 'blue' },
  long_click_element: { icon: <AimOutlined />, color: 'blue' },
  input_text: { icon: <FormOutlined />, color: 'cyan' },
  swipe_element: { icon: <SwapOutlined />, color: 'purple' },
  wait_element: { icon: <ClockCircleOutlined />, color: 'orange' },
  wait_gone: { icon: <ClockCircleOutlined />, color: 'orange' },
  wait: { icon: <ClockCircleOutlined />, color: 'default' },
  branch: { icon: <ForkOutlined />, color: 'purple' },
  scroll_to: { icon: <ReloadOutlined />, color: 'magenta' },
  assert_element: { icon: <CheckCircleOutlined />, color: 'lime' },
  set_variable: { icon: <IdcardOutlined />, color: 'orange' },
  read_to_variable: { icon: <FormOutlined />, color: 'cyan' },
  script: { icon: <PlayCircleOutlined />, color: 'geekblue' },
  adb: { icon: <RobotOutlined />, color: 'volcano' },
  key_back: { icon: <ArrowLeftOutlined />, color: 'default' },
  key_home: { icon: <HomeOutlined />, color: 'default' },
  key_recent: { icon: <AppstoreOutlined />, color: 'default' },
  key_power: { icon: <PoweroffOutlined />, color: 'red' },
  key_volume_up: { icon: <SoundOutlined />, color: 'default' },
  key_volume_down: { icon: <SoundOutlined />, color: 'default' },
  screen_on: { icon: <ExpandOutlined />, color: 'default' },
  screen_off: { icon: <LockOutlined />, color: 'default' },
  run_workflow: { icon: <BranchesOutlined />, color: 'gold' },
  launch_app: { icon: <PlayCircleOutlined />, color: 'green' },
  stop_app: { icon: <StopOutlined />, color: 'red' },
  clear_app: { icon: <DeleteOutlined />, color: 'orange' },
  open_settings: { icon: <SettingOutlined />, color: 'blue' },
  start_session: { icon: <PlayCircleOutlined />, color: 'cyan' },
  end_session: { icon: <StopOutlined />, color: 'volcano' },
};

const getStepMeta = (type: string) => STEP_TYPE_META[type] || { icon: <RobotOutlined />, color: 'default' };

/**
 * Get the step summary text for display in the list view.
 */
function getStepSummary(step: WorkflowStep): string | null {
  if (step.element?.selector) {
    return `${step.element.selector.type}: ${step.element.selector.value}`;
  }
  if (step.tap) return `(${step.tap.x}, ${step.tap.y})`;
  if (step.swipe) {
    if (step.swipe.direction) return `${step.swipe.direction} ${step.swipe.distance || ''}px`;
    return `(${step.swipe.x}, ${step.swipe.y}) -> (${step.swipe.x2}, ${step.swipe.y2})`;
  }
  if (step.app?.packageName) return step.app.packageName;
  if (step.wait?.durationMs) return `${step.wait.durationMs}ms`;
  if (step.script?.scriptName) return step.script.scriptName;
  if (step.adb?.command) return step.adb.command;
  if (step.variable) return `${step.variable.name} = ${step.variable.value}`;
  if (step.readToVariable) return `${step.readToVariable.selector?.value} -> ${step.readToVariable.variableName}`;
  if (step.branch) return step.branch.condition;
  return null;
}

/**
 * Compute the linear execution order by following successStepId from start.
 * Any steps not reachable via the main success chain are appended at the end.
 */
function getLinearOrder(steps: WorkflowStep[]): WorkflowStep[] {
  if (steps.length === 0) return [];

  const stepMap = new Map<string, WorkflowStep>();
  steps.forEach(s => stepMap.set(s.id, s));

  // Find the start node
  const startStep = steps.find(s => s.type === 'start') || steps[0];

  const ordered: WorkflowStep[] = [];
  const visited = new Set<string>();

  // Walk the success chain
  let current: WorkflowStep | undefined = startStep;
  while (current && !visited.has(current.id)) {
    visited.add(current.id);
    ordered.push(current);
    const nextId: string | undefined = current.connections?.successStepId;
    current = nextId ? stepMap.get(nextId) : undefined;
  }

  // Append any unvisited steps (branch targets, disconnected nodes, etc.)
  steps.forEach(s => {
    if (!visited.has(s.id)) {
      ordered.push(s);
    }
  });

  return ordered;
}

// ==================== SortableStepItem ====================

interface SortableStepItemProps {
  step: WorkflowStep;
  index: number;
  isActive: boolean;
  isCurrent: boolean;
  isStart: boolean;
  onClick: (step: WorkflowStep) => void;
}

const SortableStepItem: React.FC<SortableStepItemProps> = ({
  step,
  index,
  isActive,
  isCurrent,
  isStart,
  onClick,
}) => {
  const { t } = useTranslation();
  const { token } = theme.useToken();
  const meta = getStepMeta(step.type);
  const summary = getStepSummary(step);

  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id: step.id,
    disabled: isStart, // Start node cannot be dragged
  });

  const style: React.CSSProperties = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
    zIndex: isDragging ? 100 : 'auto',
  };

  return (
    <div ref={setNodeRef} style={style}>
      <div
        onClick={() => onClick(step)}
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 8,
          padding: '10px 12px',
          borderRadius: 8,
          border: `1px solid ${isActive ? token.colorPrimary : token.colorBorderSecondary}`,
          borderLeftWidth: 3,
          borderLeftColor: isCurrent ? token.colorPrimary : isActive ? token.colorPrimary : 'transparent',
          background: isCurrent
            ? token.colorPrimaryBg
            : isActive
              ? token.colorPrimaryBgHover
              : token.colorBgContainer,
          cursor: 'pointer',
          transition: 'all 0.2s',
          marginBottom: 6,
        }}
      >
        {/* Drag Handle */}
        {!isStart ? (
          <div
            {...attributes}
            {...listeners}
            style={{
              cursor: isDragging ? 'grabbing' : 'grab',
              color: token.colorTextQuaternary,
              display: 'flex',
              alignItems: 'center',
              padding: '4px 2px',
              flexShrink: 0,
            }}
          >
            <HolderOutlined style={{ fontSize: 14 }} />
          </div>
        ) : (
          <div style={{ width: 22, flexShrink: 0 }} />
        )}

        {/* Step Index */}
        <div style={{
          width: 24,
          height: 24,
          borderRadius: 12,
          background: isStart ? token.colorSuccess : token.colorFillSecondary,
          color: isStart ? '#fff' : token.colorTextSecondary,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          fontSize: 11,
          fontWeight: 600,
          flexShrink: 0,
        }}>
          {index + 1}
        </div>

        {/* Step Icon */}
        <div style={{
          backgroundColor: token.colorBgLayout,
          padding: 6,
          borderRadius: 6,
          display: 'flex',
          color: meta.color === 'default' ? token.colorText : meta.color,
          flexShrink: 0,
        }}>
          {meta.icon}
        </div>

        {/* Step Info */}
        <div style={{ flex: 1, overflow: 'hidden', minWidth: 0 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
            <Text strong ellipsis style={{ fontSize: 13, color: token.colorText }}>
              {step.name || t(`workflow.step_type.${step.type}`)}
            </Text>
            <Tag
              style={{ margin: 0, fontSize: 10, lineHeight: '16px', padding: '0 4px', flexShrink: 0 }}
              color={meta.color === 'default' ? undefined : meta.color}
            >
              {step.type}
            </Tag>
          </div>
          {summary && (
            <Tooltip title={summary}>
              <Text
                ellipsis
                style={{ fontSize: 11, color: token.colorTextSecondary, display: 'block' }}
              >
                {summary}
              </Text>
            </Tooltip>
          )}
        </div>

        {/* Connection indicator */}
        {step.type === 'branch' && (
          <Tag color="purple" style={{ margin: 0, fontSize: 10, lineHeight: '16px', padding: '0 4px' }}>
            {t("workflow.step_type.branch")}
          </Tag>
        )}
      </div>

      {/* Arrow connector between steps */}
      {!isDragging && (
        <div style={{
          display: 'flex',
          justifyContent: 'center',
          height: 16,
          position: 'relative',
        }}>
          <div style={{
            width: 1,
            height: '100%',
            background: token.colorBorderSecondary,
          }} />
          <div style={{
            position: 'absolute',
            bottom: 0,
            width: 0,
            height: 0,
            borderLeft: '4px solid transparent',
            borderRight: '4px solid transparent',
            borderTop: `5px solid ${token.colorBorderSecondary}`,
          }} />
        </div>
      )}
    </div>
  );
};

// ==================== WorkflowStepListView ====================

interface WorkflowStepListViewProps {
  workflow: Workflow;
  currentStepId: string | null;
  editingNodeId: string | null;
  onStepClick: (step: WorkflowStep) => void;
  onReorder: (orderedSteps: WorkflowStep[]) => void;
}

const WorkflowStepListView: React.FC<WorkflowStepListViewProps> = ({
  workflow,
  currentStepId,
  editingNodeId,
  onStepClick,
  onReorder,
}) => {
  const { t } = useTranslation();
  const { token } = theme.useToken();

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: { distance: 5 },
    }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  );

  // Compute linear order from graph
  const linearSteps = useMemo(
    () => getLinearOrder(workflow.steps),
    [workflow.steps]
  );

  const stepIds = useMemo(() => linearSteps.map(s => s.id), [linearSteps]);

  const handleDragEnd = useCallback(
    (event: DragEndEvent) => {
      const { active, over } = event;
      if (!over || active.id === over.id) return;

      const oldIndex = linearSteps.findIndex(s => s.id === active.id);
      const newIndex = linearSteps.findIndex(s => s.id === over.id);
      if (oldIndex === -1 || newIndex === -1) return;

      // Don't allow moving before start node
      const startIndex = linearSteps.findIndex(s => s.type === 'start');
      if (newIndex <= startIndex && active.id !== 'start') {
        message.warning(t("workflow.drag_cannot_before_start"));
        return;
      }

      const reordered = arrayMove(linearSteps, oldIndex, newIndex);
      onReorder(reordered);
    },
    [linearSteps, onReorder, t]
  );

  if (workflow.steps.length === 0) {
    return (
      <div style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        height: '100%',
        color: token.colorTextSecondary,
      }}>
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description={t("workflow.no_steps")}
        />
      </div>
    );
  }

  return (
    <div style={{
      height: '100%',
      overflow: 'auto',
      padding: 16,
    }}>
      {/* List header */}
      <div style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        marginBottom: 12,
        paddingBottom: 8,
        borderBottom: `1px solid ${token.colorBorderSecondary}`,
      }}>
        <Text type="secondary" style={{ fontSize: 12 }}>
          {t("workflow.steps_list_header", { count: workflow.steps.length })}
        </Text>
        <Text type="secondary" style={{ fontSize: 11 }}>
          {t("workflow.drag_to_reorder")}
        </Text>
      </div>

      {/* Sortable list */}
      <DndContext
        sensors={sensors}
        collisionDetection={closestCenter}
        onDragEnd={handleDragEnd}
      >
        <SortableContext items={stepIds} strategy={verticalListSortingStrategy}>
          {linearSteps.map((step, index) => (
            <SortableStepItem
              key={step.id}
              step={step}
              index={index}
              isActive={editingNodeId === step.id}
              isCurrent={currentStepId === step.id}
              isStart={step.type === 'start'}
              onClick={onStepClick}
            />
          ))}
        </SortableContext>
      </DndContext>
    </div>
  );
};

export default WorkflowStepListView;
