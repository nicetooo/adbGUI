import React, { useState } from "react";
import { Modal, Button, Descriptions, Card, Tabs, Table, Tag, Input } from "antd";
import { InfoCircleOutlined, ReloadOutlined, SearchOutlined } from "@ant-design/icons";
// @ts-ignore
import { main } from "../../wailsjs/go/models";

interface DeviceInfoModalProps {
  visible: boolean;
  onCancel: () => void;
  deviceInfo: main.DeviceInfo | null;
  loading: boolean;
  onRefresh: () => void;
}

const DeviceInfoModal: React.FC<DeviceInfoModalProps> = ({
  visible,
  onCancel,
  deviceInfo,
  loading,
  onRefresh,
}) => {
  const [searchText, setSearchText] = useState("");

  const propColumns = [
    {
      title: "Property",
      dataIndex: "key",
      key: "key",
      width: "40%",
      render: (text: string) => <code style={{ fontSize: "11px" }}>{text}</code>,
    },
    {
      title: "Value",
      dataIndex: "value",
      key: "value",
      render: (text: string) => <span style={{ fontSize: "12px" }}>{text}</span>,
    },
  ];

  const propData = deviceInfo?.props
    ? Object.entries(deviceInfo.props)
        .map(([key, value]) => ({
          key,
          value: String(value),
        }))
        .filter(item => 
          item.key.toLowerCase().includes(searchText.toLowerCase()) || 
          item.value.toLowerCase().includes(searchText.toLowerCase())
        )
        .sort((a, b) => a.key.localeCompare(b.key))
    : [];

  return (
    <Modal
      title={
        <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", paddingRight: "32px" }}>
          <span>
            <InfoCircleOutlined style={{ marginRight: 8, color: "#1890ff" }} />
            Device Information
          </span>
          <Button 
            size="small" 
            icon={<ReloadOutlined spin={loading} />} 
            onClick={onRefresh}
            disabled={loading}
          >
            Refresh
          </Button>
        </div>
      }
      open={visible}
      onCancel={onCancel}
      footer={[
        <Button key="close" onClick={onCancel}>
          Close
        </Button>,
      ]}
      width={700}
      bodyStyle={{ padding: "12px 24px 24px" }}
    >
      {loading && !deviceInfo ? (
        <div style={{ padding: "60px 0", textAlign: "center" }}>
          <ReloadOutlined spin style={{ fontSize: 32, color: "#1890ff", marginBottom: 16 }} />
          <div>Fetching device information...</div>
        </div>
      ) : (
        deviceInfo && (
          <div className="selectable" style={{ userSelect: "text" }}>
            <Tabs defaultActiveKey="basic">
              <Tabs.TabPane tab="Basic Info" key="basic">
                <Card size="small" bordered={false}>
                  <Descriptions bordered column={1} size="small">
                    <Descriptions.Item label="Model">{deviceInfo.model || "N/A"}</Descriptions.Item>
                    <Descriptions.Item label="Brand">{deviceInfo.brand || "N/A"}</Descriptions.Item>
                    <Descriptions.Item label="Manufacturer">{deviceInfo.manufacturer || "N/A"}</Descriptions.Item>
                    <Descriptions.Item label="Android Version">
                      <Tag color="blue">{deviceInfo.androidVer || "N/A"}</Tag>
                      (SDK {deviceInfo.sdk || "N/A"})
                    </Descriptions.Item>
                    <Descriptions.Item label="Serial Number">{deviceInfo.serial || "N/A"}</Descriptions.Item>
                    <Descriptions.Item label="CPU Architecture">{deviceInfo.abi || "N/A"}</Descriptions.Item>
                    <Descriptions.Item label="Hardware">{deviceInfo.cpu || "N/A"}</Descriptions.Item>
                    <Descriptions.Item label="Memory">{deviceInfo.memory || "N/A"}</Descriptions.Item>
                    <Descriptions.Item label="Resolution">{deviceInfo.resolution || "N/A"}</Descriptions.Item>
                    <Descriptions.Item label="Density">{deviceInfo.density || "N/A"}</Descriptions.Item>
                  </Descriptions>
                </Card>
              </Tabs.TabPane>
              <Tabs.TabPane tab="All Properties" key="props">
                <div style={{ marginBottom: 16 }}>
                  <Input
                    placeholder="Search properties (key or value)..."
                    prefix={<SearchOutlined />}
                    value={searchText}
                    onChange={e => setSearchText(e.target.value)}
                    allowClear
                  />
                </div>
                <Table
                  dataSource={propData}
                  columns={propColumns}
                  size="small"
                  pagination={false}
                  scroll={{ y: 400 }}
                  rowKey="key"
                />
              </Tabs.TabPane>
            </Tabs>
          </div>
        )
      )}
    </Modal>
  );
};

export default DeviceInfoModal;

