import React from "react";
import { Modal, Typography, Divider, Space, Button } from "antd";
import { InfoCircleOutlined, GithubOutlined, MailOutlined } from "@ant-design/icons";

import logo from "../assets/images/logo.svg";

const { Title, Text, Paragraph } = Typography;

interface AboutModalProps {
  visible: boolean;
  onCancel: () => void;
}

const AboutModal: React.FC<AboutModalProps> = ({ visible, onCancel }) => {
  // @ts-ignore
  const BrowserOpenURL = (window as any).runtime.BrowserOpenURL;

  return (
    <Modal
      open={visible}
      onCancel={onCancel}
      footer={[
        <Button key="close" type="primary" onClick={onCancel}>
          Close
        </Button>,
      ]}
      width={450}
      centered
      title={
        <Space>
          <InfoCircleOutlined style={{ color: "#1890ff" }} />
          <span>About ADB GUI</span>
        </Space>
      }
    >
      <div style={{ textAlign: "center", marginBottom: 24 }}>
        <img src={logo} alt="Logo" style={{ width: 80, marginBottom: 16 }} />
        <Title level={3} style={{ margin: 0 }}>ADB GUI</Title>
        <Text type="secondary">Version 1.0.0</Text>
      </div>

      <Paragraph>
        A powerful, modern, and self-contained Android management tool built with 
        <strong> Wails</strong>, <strong>React</strong>, and <strong>Ant Design</strong>.
      </Paragraph>
      <Paragraph style={{ fontStyle: "italic", color: "#faad14" }}>
        ✨ This application is a product of pure <strong>vibecoding</strong>.
      </Paragraph>

      <Divider />

      <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
        <div style={{ display: "flex", justifyContent: "space-between" }}>
          <Text strong>Author:</Text>
          <Space>
            <Text>Nice</Text>
            <Button 
              type="link" 
              size="small" 
              icon={<MailOutlined />} 
              onClick={() => BrowserOpenURL && BrowserOpenURL("mailto:naserdin@protonmail.com")}
            />
          </Space>
        </div>
        
        <div style={{ display: "flex", justifyContent: "space-between" }}>
          <Text strong>GitHub:</Text>
          <Button 
            type="link" 
            size="small" 
            icon={<GithubOutlined />} 
            onClick={() => BrowserOpenURL && BrowserOpenURL("https://github.com/nicetooo/adbGUI")}
          >
            nicetooo/adbGUI
          </Button>
        </div>

        <div style={{ display: "flex", justifyContent: "space-between" }}>
          <Text strong>License:</Text>
          <Text>MIT License</Text>
        </div>
      </div>

      <Divider />
      
      <Paragraph style={{ fontSize: 12, color: "#888", textAlign: "center" }}>
        Made with ❤️ for Android Developers & Power Users
      </Paragraph>
    </Modal>
  );
};

export default AboutModal;

