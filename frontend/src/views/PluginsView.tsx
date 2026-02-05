import React, { useEffect } from "react";
import {
  Card,
  Button,
  Switch,
  Typography,
  Space,
  Empty,
  Spin,
  Alert,
  Tag,
  Popconfirm,
  Modal,
  App,
} from "antd";
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  CodeOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  BugOutlined,
} from "@ant-design/icons";
import { usePluginStore } from "../stores/pluginStore";
import { useTranslation } from "react-i18next";
import PluginEditor from "../components/PluginEditor";

const { Title, Text, Paragraph } = Typography;

const PluginsView: React.FC = () => {
  const { t } = useTranslation();
  const { message } = App.useApp();
  const {
    plugins,
    loading,
    error,
    editorOpen,
    currentPlugin,
    loadPlugins,
    getPlugin,
    deletePlugin,
    togglePlugin,
    openEditor,
    closeEditor,
  } = usePluginStore();

  useEffect(() => {
    loadPlugins();
  }, [loadPlugins]);

  const handleToggle = async (id: string, enabled: boolean) => {
    try {
      await togglePlugin(id, enabled);
      message.success(
        enabled ? t("plugins.enabled_success") : t("plugins.disabled_success")
      );
    } catch (error) {
      message.error(t("plugins.toggle_failed"));
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await deletePlugin(id);
      message.success(t("plugins.deleted_success"));
    } catch (error) {
      message.error(t("plugins.delete_failed"));
    }
  };

  const handleEdit = async (plugin: any, initialTab: string = "basic") => {
    // 从后端获取完整的插件信息（包含源码）
    const fullPlugin = await getPlugin(plugin.metadata.id);
    if (fullPlugin) {
      openEditor(fullPlugin, initialTab);
    } else {
      message.error(t("plugins.load_failed"));
    }
  };

  const handleCreate = () => {
    openEditor();
  };

  return (
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      {/* Header - Fixed */}
      <div
        style={{
          padding: "24px",
          paddingBottom: "16px",
          flexShrink: 0,
          borderBottom: "1px solid rgba(5, 5, 5, 0.06)",
        }}
      >
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
          }}
        >
          <div>
            <Title level={2} style={{ margin: 0 }}>
              {t("plugins.title")}
            </Title>
            <Text type="secondary">{t("plugins.subtitle")}</Text>
          </div>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={handleCreate}
            size="large"
          >
            {t("plugins.create")}
          </Button>
        </div>

        {/* Error Alert */}
        {error && (
          <Alert
            message={t("plugins.error")}
            description={error}
            type="error"
            closable
            style={{ marginTop: "16px" }}
          />
        )}
      </div>

      {/* Scrollable Content */}
      <div style={{ flex: 1, overflowY: "auto", padding: "24px" }}>
        {/* Loading */}
        {loading && (
          <div style={{ textAlign: "center", padding: "48px" }}>
            <Spin size="large" />
          </div>
        )}

        {/* Empty State */}
        {!loading && plugins.length === 0 && (
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description={t("plugins.empty")}
            style={{ marginTop: "64px" }}
          >
            <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
              {t("plugins.create_first")}
            </Button>
          </Empty>
        )}

        {/* Plugin Cards */}
        {!loading && plugins.length > 0 && (
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "repeat(auto-fill, minmax(400px, 1fr))",
              gap: "16px",
            }}
          >
          {plugins.map((plugin) => (
            <Card
              key={plugin.metadata.id}
              hoverable
              style={{
                opacity: plugin.metadata.enabled ? 1 : 0.6,
                display: "flex",
                flexDirection: "column",
                height: "100%",
              }}
              styles={{
                body: {
                  flex: 1,
                  display: "flex",
                  flexDirection: "column",
                },
              }}
              actions={[
                <Button
                  type="text"
                  icon={<EditOutlined />}
                  onClick={() => handleEdit(plugin, "basic")}
                >
                  {t("plugins.edit")}
                </Button>,
                <Button
                  type="text"
                  icon={<CodeOutlined />}
                  onClick={() => handleEdit(plugin, "code")}
                >
                  {t("plugins.view_code")}
                </Button>,
                <Popconfirm
                  title={t("plugins.delete_confirm")}
                  onConfirm={() => handleDelete(plugin.metadata.id)}
                  okText={t("common.yes")}
                  cancelText={t("common.no")}
                >
                  <Button type="text" danger icon={<DeleteOutlined />}>
                    {t("plugins.delete")}
                  </Button>
                </Popconfirm>,
              ]}
            >
              <Card.Meta
                title={
                  <Space>
                    <span>{plugin.metadata.name}</span>
                    {plugin.metadata.enabled ? (
                      <Tag icon={<CheckCircleOutlined />} color="success">
                        {t("plugins.enabled")}
                      </Tag>
                    ) : (
                      <Tag icon={<CloseCircleOutlined />} color="default">
                        {t("plugins.disabled")}
                      </Tag>
                    )}
                  </Space>
                }
                description={
                  <div
                    style={{
                      display: "flex",
                      flexDirection: "column",
                      height: "100%",
                    }}
                  >
                    <Paragraph
                      ellipsis={{ rows: 2 }}
                      style={{ marginBottom: "12px", minHeight: "44px" }}
                    >
                      {plugin.metadata.description || t("plugins.no_description")}
                    </Paragraph>
                    <div style={{ flex: 1 }}>
                      <div>
                        <Text type="secondary" style={{ fontSize: "12px" }}>
                          {t("plugins.version")}: {plugin.metadata.version}
                        </Text>
                        {plugin.metadata.author && (
                          <Text
                            type="secondary"
                            style={{ fontSize: "12px", marginLeft: "12px" }}
                          >
                            {t("plugins.author")}: {plugin.metadata.author}
                          </Text>
                        )}
                      </div>
                      {/* Filters section - always show structure for consistent layout */}
                      <div style={{ marginTop: "8px", minHeight: "60px" }}>
                        {plugin.metadata.filters?.sources &&
                        plugin.metadata.filters.sources.length > 0 ? (
                          <div style={{ marginBottom: "4px" }}>
                            <Text type="secondary" style={{ fontSize: "12px" }}>
                              {t("plugins.sources")}:{" "}
                            </Text>
                            {plugin.metadata.filters.sources.map((s) => (
                              <Tag key={s} style={{ fontSize: "11px" }}>
                                {s}
                              </Tag>
                            ))}
                          </div>
                        ) : null}
                        {plugin.metadata.filters?.types &&
                        plugin.metadata.filters.types.length > 0 ? (
                          <div style={{ marginTop: "4px" }}>
                            <Text type="secondary" style={{ fontSize: "12px" }}>
                              {t("plugins.types")}:{" "}
                            </Text>
                            {plugin.metadata.filters.types.map((t) => (
                              <Tag key={t} style={{ fontSize: "11px" }}>
                                {t}
                              </Tag>
                            ))}
                          </div>
                        ) : null}
                        {(!plugin.metadata.filters?.sources ||
                          plugin.metadata.filters.sources.length === 0) &&
                          (!plugin.metadata.filters?.types ||
                            plugin.metadata.filters.types.length === 0) && (
                            <Text
                              type="secondary"
                              style={{
                                fontSize: "12px",
                                fontStyle: "italic",
                                opacity: 0.5,
                              }}
                            >
                              {t("plugins.no_filters")}
                            </Text>
                          )}
                      </div>
                    </div>
                    <div
                      style={{
                        display: "flex",
                        justifyContent: "space-between",
                        alignItems: "center",
                        marginTop: "12px",
                        paddingTop: "12px",
                        borderTop: "1px solid rgba(255, 255, 255, 0.1)",
                      }}
                    >
                      <Switch
                        checked={plugin.metadata.enabled}
                        onChange={(checked) =>
                          handleToggle(plugin.metadata.id, checked)
                        }
                        checkedChildren={t("plugins.on")}
                        unCheckedChildren={t("plugins.off")}
                      />
                      <Text type="secondary" style={{ fontSize: "11px" }}>
                        {plugin.language === "typescript" ? "TypeScript" : "JavaScript"}
                      </Text>
                    </div>
                  </div>
                }
              />
            </Card>
          ))}
        </div>
      )}
      </div>

      {/* Plugin Editor Modal */}
      {editorOpen && (
        <PluginEditor
          open={editorOpen}
          plugin={currentPlugin}
          onClose={closeEditor}
        />
      )}
    </div>
  );
};

export default PluginsView;
