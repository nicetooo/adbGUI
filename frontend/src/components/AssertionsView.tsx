/**
 * AssertionsView - 统一断言管理页面
 * Tab 1: 单个断言的 CRUD + 快捷断言
 * Tab 2: 断言集的 CRUD + 执行历史
 */
import { useEffect, useCallback } from "react";
import {
  Button,
  Table,
  Space,
  Tag,
  Modal,
  Form,
  Input,
  Select,
  Drawer,
  List,
  Progress,
  Popconfirm,
  Empty,
  Spin,
  App,
  theme,
  Typography,
  Tooltip,
  Badge,
  Tabs,
  InputNumber,
  Alert,
} from "antd";
import {
  PlusOutlined,
  PlayCircleOutlined,
  DeleteOutlined,
  EditOutlined,
  EyeOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ExclamationCircleOutlined,
  ReloadOutlined,
  SafetyOutlined,
  BugOutlined,
  SearchOutlined,
  QuestionCircleOutlined,
  ThunderboltOutlined,
} from "@ant-design/icons";
import { useTranslation } from "react-i18next";
import DeviceSelector from "./DeviceSelector";
import { useDeviceStore } from "../stores/deviceStore";
import { useEventStore } from "../stores/eventStore";
import {
  useAssertionsPanelStore,
  type AssertionSetDisplay,
  type AssertionSetResultDisplay,
} from "../stores/assertionsPanelStore";
import {
  ListAssertionSets,
  CreateAssertionSet,
  UpdateAssertionSet,
  DeleteAssertionSet,
  ExecuteAssertionSet,
  GetAssertionSetResults,
  ListStoredAssertions,
  CreateStoredAssertionJSON,
  DeleteStoredAssertion,
  ExecuteStoredAssertionInSession,
  GetStoredAssertion,
  UpdateStoredAssertionJSON,
  GetSessionEventTypes,
  PreviewAssertionMatch,
  QuickAssertNoErrors,
  QuickAssertNoCrashes,
} from "../../wailsjs/go/main/App";
import { formatDuration } from "../stores/eventTypes";

const { Text } = Typography;

// ============================================================
// Tab 1: Individual Assertions
// ============================================================

const AssertionsTab: React.FC = () => {
  const { t } = useTranslation();
  const { token } = theme.useToken();
  const { message } = App.useApp();
  const { selectedDevice } = useDeviceStore();
  const activeSessionId = useEventStore((s) => s.activeSessionId);

  const {
    loading,
    results,
    storedAssertions,
    loadingStored,
    customModalOpen,
    editModalOpen,
    editingAssertion,
    availableEventTypes,
    loadingEventTypes,
    previewCount,
    previewLoading,
    setLoading,
    addResult,
    setResults,
    clearAllResults,
    setStoredAssertions,
    removeStoredAssertion,
    setLoadingStored,
    setCustomModalOpen,
    setEditModalOpen,
    setEditingAssertion,
    setAvailableEventTypes,
    setLoadingEventTypes,
    setPreviewCount,
    setPreviewLoading,
  } = useAssertionsPanelStore();

  const [form] = Form.useForm();
  const [editForm] = Form.useForm();

  // Load stored assertions
  const loadStoredAssertions = useCallback(async () => {
    setLoadingStored(true);
    try {
      const assertions = await ListStoredAssertions("", "", false, 100);
      setStoredAssertions(
        (assertions || []).map((a: any) => ({
          id: a.id,
          name: a.name,
          type: a.type,
          createdAt: a.createdAt,
        }))
      );
    } catch {
      setStoredAssertions([]);
    } finally {
      setLoadingStored(false);
    }
  }, [setStoredAssertions, setLoadingStored]);

  useEffect(() => {
    loadStoredAssertions();
  }, [loadStoredAssertions]);

  // Load event types when create modal opens
  useEffect(() => {
    if (customModalOpen && activeSessionId) {
      setLoadingEventTypes(true);
      GetSessionEventTypes(activeSessionId)
        .then((types: string[]) => setAvailableEventTypes(types || []))
        .catch(() => {})
        .finally(() => setLoadingEventTypes(false));
    }
  }, [customModalOpen, activeSessionId, setLoadingEventTypes, setAvailableEventTypes]);

  // Preview match count
  const updatePreview = useCallback(
    async (eventTypes: string[], titleMatch: string) => {
      if (!activeSessionId) {
        setPreviewCount(null);
        return;
      }
      setPreviewLoading(true);
      try {
        const count = await PreviewAssertionMatch(activeSessionId, eventTypes || [], titleMatch || "");
        setPreviewCount(count);
      } catch {
        setPreviewCount(null);
      } finally {
        setPreviewLoading(false);
      }
    },
    [activeSessionId, setPreviewCount, setPreviewLoading]
  );

  const handleFormValuesChange = useCallback(
    (_changedValues: any, allValues: any) => {
      const eventTypes = allValues.eventTypes || [];
      const titleMatch = allValues.titleMatch || "";
      if (eventTypes.length > 0 || titleMatch) {
        updatePreview(eventTypes, titleMatch);
      } else {
        setPreviewCount(null);
      }
    },
    [updatePreview, setPreviewCount]
  );

  // Create assertion (save + optionally execute)
  const handleCreateAssertion = useCallback(
    async (values: any) => {
      setLoading(true);
      try {
        const assertion: Record<string, any> = {
          id: `custom_${Date.now()}`,
          name: values.name || "Custom Assertion",
          type: values.type,
          sessionId: "",
          deviceId: "",
          criteria: {
            types: Array.isArray(values.eventTypes) ? values.eventTypes : undefined,
            titleMatch: values.titleMatch || undefined,
          },
          expected: {},
          createdAt: Date.now(),
        };
        switch (values.type) {
          case "exists":
            assertion.expected = { exists: true };
            break;
          case "not_exists":
            assertion.expected = { exists: false };
            break;
          case "count":
            assertion.expected = { minCount: values.minCount, maxCount: values.maxCount };
            break;
        }
        await CreateStoredAssertionJSON(JSON.stringify(assertion), false);
        message.success(t("assertions.created"));
        setCustomModalOpen(false);
        setPreviewCount(null);
        form.resetFields();
        loadStoredAssertions();
      } catch (err: any) {
        message.error(`${t("assertions.save_failed")}: ${err}`);
      } finally {
        setLoading(false);
      }
    },
    [form, t, message, setLoading, setCustomModalOpen, setPreviewCount, loadStoredAssertions]
  );

  // Execute stored assertion
  const handleExecute = useCallback(
    async (assertionId: string) => {
      if (!activeSessionId) {
        message.warning(t("assertions.select_session_first"));
        return;
      }
      setLoading(true);
      try {
        const result = await ExecuteStoredAssertionInSession(assertionId, activeSessionId, selectedDevice || "");
        if (result) {
          addResult({
            id: result.id,
            name: result.assertionName,
            passed: result.passed,
            message: result.message,
            executedAt: result.executedAt,
            duration: result.duration,
            matchedCount: result.matchedEvents?.length || 0,
          });
          if (result.passed) {
            message.success(`${t("assertions.passed_msg")}: ${result.assertionName}`);
          } else {
            message.error(`${t("assertions.failed_msg")}: ${result.message}`);
          }
        }
      } catch (err: any) {
        message.error(`${t("assertions.error_msg")}: ${err}`);
      } finally {
        setLoading(false);
      }
    },
    [activeSessionId, selectedDevice, t, message, setLoading, addResult]
  );

  // Quick assertions
  const handleQuickAssertion = useCallback(
    async (type: "noErrors" | "noCrashes") => {
      if (!activeSessionId) {
        message.warning(t("assertions.select_session_first"));
        return;
      }
      setLoading(true);
      try {
        const result =
          type === "noErrors"
            ? await QuickAssertNoErrors(activeSessionId, selectedDevice || "")
            : await QuickAssertNoCrashes(activeSessionId, selectedDevice || "");
        if (result) {
          addResult({
            id: result.id,
            name: result.assertionName,
            passed: result.passed,
            message: result.message,
            executedAt: result.executedAt,
            duration: result.duration,
            matchedCount: result.matchedEvents?.length || 0,
          });
          if (result.passed) {
            message.success(`${t("assertions.passed_msg")}: ${result.assertionName}`);
          } else {
            message.error(`${t("assertions.failed_msg")}: ${result.message}`);
          }
        }
      } catch (err: any) {
        message.error(`${t("assertions.error_msg")}: ${err}`);
      } finally {
        setLoading(false);
      }
    },
    [activeSessionId, selectedDevice, t, message, setLoading, addResult]
  );

  // Delete
  const handleDelete = useCallback(
    async (id: string) => {
      try {
        await DeleteStoredAssertion(id);
        removeStoredAssertion(id);
        message.success(t("assertions.deleted"));
      } catch (err: any) {
        message.error(`${t("assertions.delete_failed")}: ${err}`);
      }
    },
    [t, message, removeStoredAssertion]
  );

  // Open edit modal
  const openEditModal = useCallback(
    async (assertionId: string) => {
      try {
        const stored = await GetStoredAssertion(assertionId);
        if (stored) {
          const criteria = stored.criteria
            ? typeof stored.criteria === "string"
              ? JSON.parse(stored.criteria)
              : stored.criteria
            : {};
          const expected = stored.expected
            ? typeof stored.expected === "string"
              ? JSON.parse(stored.expected)
              : stored.expected
            : {};
          setEditingAssertion({ id: stored.id, name: stored.name, type: stored.type, criteria, expected, createdAt: stored.createdAt });
          editForm.setFieldsValue({
            name: stored.name,
            type: stored.type,
            eventTypes: criteria.types || [],
            titleMatch: criteria.titleMatch || "",
            minCount: expected.minCount,
            maxCount: expected.maxCount,
          });
          setEditModalOpen(true);
          // Load event types for edit form
          if (activeSessionId) {
            setLoadingEventTypes(true);
            GetSessionEventTypes(activeSessionId)
              .then((types: string[]) => setAvailableEventTypes(types || []))
              .catch(() => {})
              .finally(() => setLoadingEventTypes(false));
          }
        }
      } catch (err: any) {
        message.error(`${t("assertions.load_failed")}: ${err}`);
      }
    },
    [editForm, t, message, activeSessionId, setEditingAssertion, setEditModalOpen, setLoadingEventTypes, setAvailableEventTypes]
  );

  // Save edited assertion
  const saveEditedAssertion = useCallback(
    async (values: any) => {
      if (!editingAssertion) return;
      try {
        const assertion: Record<string, any> = {
          id: editingAssertion.id,
          name: values.name || "Custom Assertion",
          type: values.type,
          sessionId: "",
          deviceId: "",
          criteria: {
            types: Array.isArray(values.eventTypes) ? values.eventTypes : undefined,
            titleMatch: values.titleMatch || undefined,
          },
          expected: {},
          createdAt: editingAssertion.createdAt,
        };
        switch (values.type) {
          case "exists":
            assertion.expected = { exists: true };
            break;
          case "not_exists":
            assertion.expected = { exists: false };
            break;
          case "count":
            assertion.expected = { minCount: values.minCount, maxCount: values.maxCount };
            break;
        }
        await UpdateStoredAssertionJSON(editingAssertion.id, JSON.stringify(assertion));
        message.success(t("assertions.saved"));
        setEditModalOpen(false);
        editForm.resetFields();
        setEditingAssertion(null);
        loadStoredAssertions();
      } catch (err: any) {
        message.error(`${t("assertions.save_failed")}: ${err}`);
      }
    },
    [editingAssertion, editForm, t, message, setEditModalOpen, setEditingAssertion, loadStoredAssertions]
  );

  const deleteFromEditModal = useCallback(async () => {
    if (!editingAssertion) return;
    try {
      await DeleteStoredAssertion(editingAssertion.id);
      message.success(t("assertions.deleted"));
      setEditModalOpen(false);
      editForm.resetFields();
      setEditingAssertion(null);
      removeStoredAssertion(editingAssertion.id);
    } catch (err: any) {
      message.error(`${t("assertions.delete_failed")}: ${err}`);
    }
  }, [editingAssertion, editForm, t, message, setEditModalOpen, setEditingAssertion, removeStoredAssertion]);

  const typeLabel = (type: string) => {
    switch (type) {
      case "exists": return t("assertions.type_exists");
      case "not_exists": return t("assertions.type_not_exists");
      case "count": return t("assertions.type_count");
      default: return type;
    }
  };

  const typeColor = (type: string) => {
    switch (type) {
      case "exists": return "blue";
      case "not_exists": return "orange";
      case "count": return "purple";
      default: return "default";
    }
  };

  // Assertion form fields (shared between create & edit)
  const renderAssertionFormFields = (isEdit: boolean) => (
    <>
      <Form.Item
        name="name"
        label={
          <Space>
            {t("assertions.assertion_name")}
            <Tooltip title={t("assertions.name_tooltip")}>
              <QuestionCircleOutlined style={{ color: "#999" }} />
            </Tooltip>
          </Space>
        }
        rules={[{ required: true, message: t("assertions.name_required") }]}
      >
        <Input placeholder={t("assertions.name_placeholder")} />
      </Form.Item>
      <Form.Item
        name="type"
        label={
          <Space>
            {t("assertions.assertion_type")}
            <Tooltip title={t("assertions.type_tooltip")}>
              <QuestionCircleOutlined style={{ color: "#999" }} />
            </Tooltip>
          </Space>
        }
        rules={[{ required: true, message: t("assertions.type_required") }]}
      >
        <Select
          placeholder={t("assertions.select_type")}
          options={[
            { label: t("assertions.type_exists"), value: "exists" },
            { label: t("assertions.type_not_exists"), value: "not_exists" },
            { label: t("assertions.type_count"), value: "count" },
          ]}
        />
      </Form.Item>
      <Form.Item
        name="eventTypes"
        label={
          <Space>
            {t("assertions.event_types")}
            <Tooltip title={t("assertions.event_types_tooltip")}>
              <QuestionCircleOutlined style={{ color: "#999" }} />
            </Tooltip>
          </Space>
        }
      >
        <Select
          mode="multiple"
          placeholder={t("assertions.event_types_placeholder")}
          loading={loadingEventTypes}
          allowClear
          showSearch
          optionFilterProp="label"
          options={availableEventTypes.map((type) => ({ label: type, value: type }))}
          notFoundContent={
            loadingEventTypes ? (
              <Spin size="small" />
            ) : availableEventTypes.length === 0 ? (
              <Text type="secondary">{t("assertions.no_event_types")}</Text>
            ) : null
          }
        />
      </Form.Item>
      <Form.Item
        name="titleMatch"
        label={
          <Space>
            {t("assertions.title_match")}
            <Tooltip title={t("assertions.title_match_tooltip")}>
              <QuestionCircleOutlined style={{ color: "#999" }} />
            </Tooltip>
          </Space>
        }
        rules={
          !isEdit
            ? [
                {
                  validator: async (_, value) => {
                    if (value) {
                      try {
                        new RegExp(value);
                      } catch {
                        throw new Error(t("assertions.invalid_regex"));
                      }
                    }
                  },
                },
              ]
            : undefined
        }
      >
        <Input placeholder={t("assertions.title_match_placeholder")} prefix={<SearchOutlined style={{ color: "#999" }} />} />
      </Form.Item>
      {/* Preview (create only) */}
      {!isEdit && previewCount !== null && (
        <Alert
          type={previewCount > 0 ? "info" : "warning"}
          showIcon
          icon={previewLoading ? <Spin size="small" /> : undefined}
          message={previewLoading ? t("assertions.previewing") : t("assertions.preview_count", { count: previewCount })}
          style={{ marginBottom: 16 }}
        />
      )}
      <Form.Item noStyle shouldUpdate={(prev, curr) => prev.type !== curr.type}>
        {({ getFieldValue }) =>
          getFieldValue("type") === "count" && (
            <Space style={{ width: "100%" }}>
              <Form.Item name="minCount" label={t("assertions.min_count")} style={{ marginBottom: 16 }}>
                <InputNumber min={0} style={{ width: 120 }} />
              </Form.Item>
              <Form.Item name="maxCount" label={t("assertions.max_count")} style={{ marginBottom: 16 }}>
                <InputNumber min={0} style={{ width: 120 }} />
              </Form.Item>
            </Space>
          )
        }
      </Form.Item>
    </>
  );

  // Assertions table columns
  const columns = [
    {
      title: t("assertions.assertion_name"),
      dataIndex: "name",
      key: "name",
      render: (name: string) => <Text strong>{name}</Text>,
    },
    {
      title: t("assertions.assertion_type"),
      dataIndex: "type",
      key: "type",
      width: 140,
      render: (type: string) => <Tag color={typeColor(type)}>{typeLabel(type)}</Tag>,
    },
    {
      title: t("assertion_sets.created"),
      dataIndex: "createdAt",
      key: "createdAt",
      width: 180,
      render: (ts: number) => (ts ? new Date(ts).toLocaleString() : "-"),
    },
    {
      title: t("assertion_sets.actions"),
      key: "actions",
      width: 180,
      render: (_: any, record: any) => (
        <Space>
          <Tooltip title={t("assertions.execute")}>
            <Button
              type="primary"
              size="small"
              icon={<PlayCircleOutlined />}
              onClick={() => handleExecute(record.id)}
              disabled={!activeSessionId}
              loading={loading}
            />
          </Tooltip>
          <Tooltip title={t("common.edit")}>
            <Button size="small" icon={<EditOutlined />} onClick={() => openEditModal(record.id)} />
          </Tooltip>
          <Popconfirm title={t("assertion_sets.delete_confirm")} onConfirm={() => handleDelete(record.id)}>
            <Tooltip title={t("common.delete")}>
              <Button size="small" danger icon={<DeleteOutlined />} />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div style={{ display: "flex", flexDirection: "column", height: "100%" }}>
      {/* Toolbar */}
      <Space style={{ marginBottom: 16, flexWrap: "wrap" }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setCustomModalOpen(true)}>
          {t("assertions.create_custom")}
        </Button>
        <Button icon={<ReloadOutlined />} onClick={loadStoredAssertions}>
          {t("assertion_sets.refresh")}
        </Button>
        <Tooltip title={t("assertions.no_errors_desc")}>
          <Button icon={<SafetyOutlined />} loading={loading} onClick={() => handleQuickAssertion("noErrors")} disabled={!activeSessionId}>
            {t("assertions.no_errors")}
          </Button>
        </Tooltip>
        <Tooltip title={t("assertions.no_crashes_desc")}>
          <Button icon={<BugOutlined />} loading={loading} onClick={() => handleQuickAssertion("noCrashes")} disabled={!activeSessionId}>
            {t("assertions.no_crashes")}
          </Button>
        </Tooltip>
        {!activeSessionId && (
          <Text type="warning" style={{ fontSize: 12 }}>
            {t("assertion_sets.no_active_session")}
          </Text>
        )}
      </Space>

      {/* Table */}
      <div style={{ flex: 1, overflow: "auto" }}>
        <Table
          dataSource={storedAssertions}
          columns={columns}
          rowKey="id"
          loading={loadingStored}
          pagination={false}
          size="middle"
          locale={{
            emptyText: <Empty description={t("assertions.no_stored")} image={Empty.PRESENTED_IMAGE_SIMPLE} />,
          }}
        />
      </div>

      {/* Execution results (collapsible bottom section) */}
      {results.length > 0 && (
        <div
          style={{
            marginTop: 12,
            padding: "8px 12px",
            background: token.colorBgContainer,
            borderRadius: token.borderRadius,
            border: `1px solid ${token.colorBorderSecondary}`,
            maxHeight: 200,
            overflow: "auto",
          }}
        >
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 8 }}>
            <Text type="secondary" style={{ fontSize: 12 }}>
              <ThunderboltOutlined /> {t("assertions.results")} ({results.length})
            </Text>
            <Button type="text" size="small" onClick={clearAllResults} style={{ fontSize: 12 }}>
              {t("assertions.clear_all")}
            </Button>
          </div>
          {results.map((item) => (
            <div
              key={item.id}
              style={{
                display: "flex",
                alignItems: "center",
                gap: 8,
                padding: "4px 0",
                fontSize: 12,
              }}
            >
              {item.passed ? (
                <CheckCircleOutlined style={{ color: token.colorSuccess }} />
              ) : (
                <CloseCircleOutlined style={{ color: token.colorError }} />
              )}
              <Text strong style={{ fontSize: 12 }}>
                {item.name}
              </Text>
              <Text type="secondary" style={{ fontSize: 11, flex: 1 }}>
                {item.message}
              </Text>
              <Text type="secondary" style={{ fontSize: 11 }}>
                {formatDuration(item.duration)}
              </Text>
            </div>
          ))}
        </div>
      )}

      {/* Create Assertion Modal */}
      <Modal
        title={t("assertions.create_custom")}
        open={customModalOpen}
        onCancel={() => {
          setCustomModalOpen(false);
          setPreviewCount(null);
          form.resetFields();
        }}
        footer={null}
        width={560}
      >
        <Form form={form} layout="vertical" onFinish={handleCreateAssertion} onValuesChange={handleFormValuesChange}>
          {renderAssertionFormFields(false)}
          <Form.Item style={{ marginBottom: 0, marginTop: 8 }}>
            <Space>
              <Button type="primary" htmlType="submit" loading={loading}>
                {t("assertions.save_and_create")}
              </Button>
              <Button
                onClick={() => {
                  setCustomModalOpen(false);
                  setPreviewCount(null);
                  form.resetFields();
                }}
              >
                {t("common.cancel")}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* Edit Assertion Modal */}
      <Modal
        title={t("assertions.edit_assertion")}
        open={editModalOpen}
        onCancel={() => {
          setEditModalOpen(false);
          editForm.resetFields();
          setEditingAssertion(null);
        }}
        footer={null}
        width={560}
      >
        <Form form={editForm} layout="vertical" onFinish={saveEditedAssertion}>
          {renderAssertionFormFields(true)}
          <Form.Item style={{ marginBottom: 0, marginTop: 8 }}>
            <div style={{ display: "flex", justifyContent: "space-between" }}>
              <Button danger onClick={deleteFromEditModal}>
                <DeleteOutlined /> {t("common.delete")}
              </Button>
              <Space>
                <Button
                  onClick={() => {
                    setEditModalOpen(false);
                    editForm.resetFields();
                    setEditingAssertion(null);
                  }}
                >
                  {t("common.cancel")}
                </Button>
                <Button type="primary" htmlType="submit">
                  {t("common.save")}
                </Button>
              </Space>
            </div>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

// ============================================================
// Tab 2: Assertion Sets (preserved from AssertionSetsView)
// ============================================================

const AssertionSetsTab: React.FC = () => {
  const { t } = useTranslation();
  const { token } = theme.useToken();
  const { message } = App.useApp();
  const { selectedDevice } = useDeviceStore();
  const activeSessionId = useEventStore((s) => s.activeSessionId);

  const {
    assertionSets,
    loadingSets,
    selectedSet,
    setExecutionResults,
    loadingExecution,
    setCreateModalOpen,
    setDetailModalOpen,
    editingSet,
    storedAssertions,
    loadingStored,
    setAssertionSets,
    removeAssertionSet,
    setLoadingSets,
    setSelectedSet,
    setSetExecutionResults,
    setLoadingExecution,
    setSetCreateModalOpen,
    setSetDetailModalOpen,
    setEditingSet,
    setStoredAssertions,
    setLoadingStored,
  } = useAssertionsPanelStore();

  const [form] = Form.useForm();

  const loadSets = useCallback(async () => {
    setLoadingSets(true);
    try {
      const sets = await ListAssertionSets();
      setAssertionSets(
        (sets || []).map((s: any) => ({
          id: s.id,
          name: s.name,
          description: s.description || "",
          assertions: s.assertions || [],
          createdAt: s.createdAt,
          updatedAt: s.updatedAt,
        }))
      );
    } catch (err: any) {
      message.error(t("assertion_sets.load_failed") + ": " + err.message);
    } finally {
      setLoadingSets(false);
    }
  }, [setAssertionSets, setLoadingSets, message, t]);

  const loadStoredAssertions = useCallback(async () => {
    setLoadingStored(true);
    try {
      const assertions = await ListStoredAssertions("", "", true, 100);
      setStoredAssertions(
        (assertions || []).map((a: any) => ({
          id: a.id,
          name: a.name,
          type: a.type,
          createdAt: a.createdAt,
        }))
      );
    } catch {
      // silent
    } finally {
      setLoadingStored(false);
    }
  }, [setStoredAssertions, setLoadingStored]);

  useEffect(() => {
    loadSets();
    loadStoredAssertions();
  }, [loadSets, loadStoredAssertions]);

  const handleSave = useCallback(async () => {
    try {
      const values = await form.validateFields();
      if (editingSet) {
        await UpdateAssertionSet(editingSet.id, values.name, values.description || "", values.assertions || []);
        message.success(t("assertion_sets.updated"));
      } else {
        await CreateAssertionSet(values.name, values.description || "", values.assertions || []);
        message.success(t("assertion_sets.created"));
      }
      setSetCreateModalOpen(false);
      setEditingSet(null);
      form.resetFields();
      loadSets();
    } catch (err: any) {
      if (err?.errorFields) return;
      message.error(t("assertion_sets.save_failed") + ": " + err.message);
    }
  }, [editingSet, form, loadSets, message, t, setSetCreateModalOpen, setEditingSet]);

  const handleDelete = useCallback(
    async (id: string) => {
      try {
        await DeleteAssertionSet(id);
        removeAssertionSet(id);
        message.success(t("assertion_sets.deleted"));
      } catch (err: any) {
        message.error(t("assertion_sets.delete_failed") + ": " + err.message);
      }
    },
    [removeAssertionSet, message, t]
  );

  const handleExecute = useCallback(
    async (set: AssertionSetDisplay) => {
      if (!activeSessionId) {
        message.warning(t("assertions.select_session_first"));
        return;
      }
      if (!selectedDevice) {
        message.warning(t("assertion_sets.select_device_first"));
        return;
      }
      setSelectedSet(set);
      setSetDetailModalOpen(true);
      setLoadingExecution(true);
      try {
        const result = await ExecuteAssertionSet(set.id, activeSessionId, selectedDevice);
        if (result) {
          const mapped: AssertionSetResultDisplay = {
            id: result.id,
            setId: result.setId,
            setName: result.setName,
            sessionId: result.sessionId,
            deviceId: result.deviceId,
            executionId: result.executionId,
            startTime: result.startTime,
            endTime: result.endTime,
            duration: result.duration,
            status: result.status,
            summary: result.summary,
            results: (result.results || []).map((r: any) => ({
              id: r.id,
              assertionId: r.assertionId,
              assertionName: r.assertionName,
              sessionId: r.sessionId,
              passed: r.passed,
              message: r.message,
              actualValue: r.actualValue,
              expectedValue: r.expectedValue,
              executedAt: r.executedAt,
              duration: r.duration,
            })),
            executedAt: result.executedAt,
          };
          setSetExecutionResults([mapped, ...setExecutionResults]);
          message.success(`${t("assertion_sets.execution_complete")}: ${result.summary.passed}/${result.summary.total} ${t("assertions.pass")}`);
        }
      } catch (err: any) {
        message.error(t("assertion_sets.execution_failed") + ": " + err.message);
      } finally {
        setLoadingExecution(false);
      }
    },
    [activeSessionId, selectedDevice, setSelectedSet, setSetDetailModalOpen, setLoadingExecution, setSetExecutionResults, setExecutionResults, message, t]
  );

  const handleViewDetail = useCallback(
    async (set: AssertionSetDisplay) => {
      setSelectedSet(set);
      setSetDetailModalOpen(true);
      setLoadingExecution(true);
      try {
        const results = await GetAssertionSetResults(set.id, 20);
        setSetExecutionResults(
          (results || []).map((r: any) => ({
            id: r.id,
            setId: r.setId,
            setName: r.setName,
            sessionId: r.sessionId,
            deviceId: r.deviceId,
            executionId: r.executionId,
            startTime: r.startTime,
            endTime: r.endTime,
            duration: r.duration,
            status: r.status,
            summary: r.summary || { total: 0, passed: 0, failed: 0, error: 0, passRate: 0 },
            results: (r.results || []).map((ar: any) => ({
              id: ar.id,
              assertionId: ar.assertionId,
              assertionName: ar.assertionName,
              sessionId: ar.sessionId,
              passed: ar.passed,
              message: ar.message,
              actualValue: ar.actualValue,
              expectedValue: ar.expectedValue,
              executedAt: ar.executedAt,
              duration: ar.duration,
            })),
            executedAt: r.executedAt,
          }))
        );
      } catch (err: any) {
        message.error(t("assertion_sets.load_history_failed") + ": " + err.message);
      } finally {
        setLoadingExecution(false);
      }
    },
    [setSelectedSet, setSetDetailModalOpen, setLoadingExecution, setSetExecutionResults, message, t]
  );

  const handleOpenCreate = useCallback(() => {
    setEditingSet(null);
    form.resetFields();
    setSetCreateModalOpen(true);
  }, [form, setEditingSet, setSetCreateModalOpen]);

  const handleOpenEdit = useCallback(
    (set: AssertionSetDisplay) => {
      setEditingSet(set);
      form.setFieldsValue({ name: set.name, description: set.description, assertions: set.assertions });
      setSetCreateModalOpen(true);
    },
    [form, setEditingSet, setSetCreateModalOpen]
  );

  const statusColor = (status: string) => {
    switch (status) {
      case "passed": return "success";
      case "failed": return "error";
      case "partial": return "warning";
      default: return "default";
    }
  };

  const statusIcon = (status: string) => {
    switch (status) {
      case "passed": return <CheckCircleOutlined style={{ color: token.colorSuccess }} />;
      case "failed": return <CloseCircleOutlined style={{ color: token.colorError }} />;
      case "partial": return <ExclamationCircleOutlined style={{ color: token.colorWarning }} />;
      default: return <ExclamationCircleOutlined />;
    }
  };

  const columns = [
    {
      title: t("assertion_sets.name"),
      dataIndex: "name",
      key: "name",
      render: (name: string, record: AssertionSetDisplay) => (
        <div>
          <Text strong>{name}</Text>
          {record.description && (
            <div>
              <Text type="secondary" style={{ fontSize: 12 }}>
                {record.description}
              </Text>
            </div>
          )}
        </div>
      ),
    },
    {
      title: t("assertion_sets.assertion_count"),
      dataIndex: "assertions",
      key: "count",
      width: 120,
      align: "center" as const,
      render: (assertions: string[]) => <Badge count={assertions?.length || 0} showZero style={{ backgroundColor: token.colorPrimary }} />,
    },
    {
      title: t("assertion_sets.created"),
      dataIndex: "createdAt",
      key: "createdAt",
      width: 180,
      render: (ts: number) => (ts ? new Date(ts).toLocaleString() : "-"),
    },
    {
      title: t("assertion_sets.actions"),
      key: "actions",
      width: 200,
      render: (_: any, record: AssertionSetDisplay) => (
        <Space>
          <Tooltip title={t("assertion_sets.execute")}>
            <Button type="primary" size="small" icon={<PlayCircleOutlined />} onClick={() => handleExecute(record)} disabled={!activeSessionId || !selectedDevice} />
          </Tooltip>
          <Tooltip title={t("assertion_sets.view_detail")}>
            <Button size="small" icon={<EyeOutlined />} onClick={() => handleViewDetail(record)} />
          </Tooltip>
          <Tooltip title={t("assertion_sets.edit")}>
            <Button size="small" icon={<EditOutlined />} onClick={() => handleOpenEdit(record)} />
          </Tooltip>
          <Popconfirm title={t("assertion_sets.delete_confirm")} onConfirm={() => handleDelete(record.id)}>
            <Tooltip title={t("assertion_sets.delete")}>
              <Button size="small" danger icon={<DeleteOutlined />} />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div style={{ display: "flex", flexDirection: "column", height: "100%" }}>
      {/* Toolbar */}
      <Space style={{ marginBottom: 16 }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleOpenCreate}>
          {t("assertion_sets.create")}
        </Button>
        <Button icon={<ReloadOutlined />} onClick={loadSets}>
          {t("assertion_sets.refresh")}
        </Button>
        {!activeSessionId && (
          <Text type="warning" style={{ fontSize: 12 }}>
            {t("assertion_sets.no_active_session")}
          </Text>
        )}
      </Space>

      {/* Table */}
      <div style={{ flex: 1, overflow: "auto" }}>
        <Table
          dataSource={assertionSets}
          columns={columns}
          rowKey="id"
          loading={loadingSets}
          pagination={false}
          size="middle"
          locale={{ emptyText: <Empty description={t("assertion_sets.empty")} image={Empty.PRESENTED_IMAGE_SIMPLE} /> }}
        />
      </div>

      {/* Create/Edit Set Modal */}
      <Modal
        title={editingSet ? t("assertion_sets.edit_title") : t("assertion_sets.create_title")}
        open={setCreateModalOpen}
        onOk={handleSave}
        onCancel={() => {
          setSetCreateModalOpen(false);
          setEditingSet(null);
          form.resetFields();
        }}
        okText={t("assertion_sets.save")}
        width={600}
        destroyOnHidden
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label={t("assertion_sets.name")} rules={[{ required: true, message: t("assertion_sets.name_required") }]}>
            <Input placeholder={t("assertion_sets.name_placeholder")} />
          </Form.Item>
          <Form.Item name="description" label={t("assertion_sets.description")}>
            <Input.TextArea rows={2} placeholder={t("assertion_sets.description_placeholder")} />
          </Form.Item>
          <Form.Item name="assertions" label={t("assertion_sets.select_assertions")}>
            <Select
              mode="multiple"
              placeholder={t("assertion_sets.select_assertions_placeholder")}
              loading={loadingStored}
              optionFilterProp="label"
              options={(storedAssertions || []).map((a) => ({ label: a.name, value: a.id }))}
            />
          </Form.Item>
        </Form>
      </Modal>

      {/* Detail Drawer */}
      <Drawer
        title={
          selectedSet ? (
            <Space>
              {selectedSet.name}
              <Tag>
                {(selectedSet.assertions || []).length} {t("assertion_sets.assertions_label")}
              </Tag>
            </Space>
          ) : (
            t("assertion_sets.detail")
          )
        }
        open={setDetailModalOpen}
        onClose={() => {
          setSetDetailModalOpen(false);
          setSelectedSet(null);
          setSetExecutionResults([]);
        }}
        styles={{ wrapper: { width: 640 } }}
        extra={
          selectedSet && (
            <Button
              type="primary"
              icon={<PlayCircleOutlined />}
              onClick={() => selectedSet && handleExecute(selectedSet)}
              disabled={!activeSessionId || !selectedDevice}
              loading={loadingExecution}
            >
              {t("assertion_sets.execute")}
            </Button>
          )
        }
      >
        {selectedSet && (
          <div>
            {selectedSet.description && (
              <Text type="secondary" style={{ display: "block", marginBottom: 16 }}>
                {selectedSet.description}
              </Text>
            )}
            <h4>{t("assertion_sets.execution_history")}</h4>
            {loadingExecution ? (
              <div style={{ textAlign: "center", padding: 32 }}>
                <Spin />
              </div>
            ) : setExecutionResults.length === 0 ? (
              <Empty description={t("assertion_sets.no_history")} image={Empty.PRESENTED_IMAGE_SIMPLE} />
            ) : (
              <List
                dataSource={setExecutionResults}
                renderItem={(result: AssertionSetResultDisplay) => (
                  <List.Item
                    style={{
                      borderLeft: `3px solid ${result.status === "passed" ? token.colorSuccess : result.status === "failed" ? token.colorError : token.colorWarning}`,
                      paddingLeft: 12,
                      marginBottom: 12,
                      background: token.colorBgContainer,
                      borderRadius: token.borderRadius,
                    }}
                  >
                    <div style={{ width: "100%" }}>
                      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 8 }}>
                        <Space>
                          {statusIcon(result.status)}
                          <Tag color={statusColor(result.status)}>{result.status.toUpperCase()}</Tag>
                          <Text type="secondary" style={{ fontSize: 12 }}>
                            {new Date(result.executedAt).toLocaleString()}
                          </Text>
                        </Space>
                        <Space>
                          <Text style={{ fontSize: 12 }}>{formatDuration(result.duration)}</Text>
                          <Progress
                            type="circle"
                            size={32}
                            percent={Math.round(result.summary.passRate)}
                            strokeColor={result.summary.passRate === 100 ? token.colorSuccess : result.summary.passRate >= 50 ? token.colorWarning : token.colorError}
                          />
                        </Space>
                      </div>
                      <Space style={{ marginBottom: 8 }}>
                        <Tag color="blue">
                          {t("assertion_sets.total")}: {result.summary.total}
                        </Tag>
                        <Tag color="green">
                          {t("assertions.pass")}: {result.summary.passed}
                        </Tag>
                        {result.summary.failed > 0 && (
                          <Tag color="red">
                            {t("assertions.fail")}: {result.summary.failed}
                          </Tag>
                        )}
                        {result.summary.error > 0 && (
                          <Tag color="orange">
                            {t("assertion_sets.errors")}: {result.summary.error}
                          </Tag>
                        )}
                      </Space>
                      <div>
                        {(result.results || []).map((ar, idx) => (
                          <div
                            key={ar.id || idx}
                            style={{
                              padding: "4px 8px",
                              borderRadius: 4,
                              marginBottom: 2,
                              background: ar.passed ? token.colorSuccessBg : token.colorErrorBg,
                              fontSize: 12,
                              display: "flex",
                              justifyContent: "space-between",
                            }}
                          >
                            <Space size={4}>
                              {ar.passed ? <CheckCircleOutlined style={{ color: token.colorSuccess }} /> : <CloseCircleOutlined style={{ color: token.colorError }} />}
                              <Text style={{ fontSize: 12 }} strong={!ar.passed}>
                                {ar.assertionName}
                              </Text>
                              <Text type="secondary" style={{ fontSize: 11 }}>
                                {ar.message}
                              </Text>
                            </Space>
                            <Text type="secondary" style={{ fontSize: 11 }}>
                              {formatDuration(ar.duration)}
                            </Text>
                          </div>
                        ))}
                      </div>
                    </div>
                  </List.Item>
                )}
              />
            )}
          </div>
        )}
      </Drawer>
    </div>
  );
};

// ============================================================
// Main View: Tabs
// ============================================================

const AssertionsView: React.FC = () => {
  const { t } = useTranslation();

  return (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        height: "100%",
        position: "relative",
        padding: "16px 24px",
      }}
    >
      {/* Header */}
      <div
        style={{
          marginBottom: 16,
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
        }}
      >
        <h2 style={{ margin: 0 }}>{t("assertions.title")}</h2>
        <DeviceSelector />
      </div>

      {/* Tabs */}
      <Tabs
        defaultActiveKey="assertions"
        style={{ flex: 1, display: "flex", flexDirection: "column" }}
        items={[
          {
            key: "assertions",
            label: (
              <span>
                <ThunderboltOutlined /> {t("assertions.tab_assertions")}
              </span>
            ),
            children: <AssertionsTab />,
          },
          {
            key: "sets",
            label: (
              <span>
                <CheckCircleOutlined /> {t("assertions.tab_sets")}
              </span>
            ),
            children: <AssertionSetsTab />,
          },
        ]}
      />
    </div>
  );
};

export default AssertionsView;
