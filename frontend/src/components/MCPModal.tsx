import React from 'react';
import { Modal } from 'antd';
import { useTranslation } from 'react-i18next';
import { useUIStore } from '../stores/uiStore';
import MCPInfoSection from './MCPInfoSection';

const MCPModal: React.FC = () => {
  const { t } = useTranslation();
  const { mcpVisible, hideMCP } = useUIStore();

  return (
    <Modal
      title={t('mcp.title')}
      open={mcpVisible}
      onCancel={hideMCP}
      footer={null}
      width={1000}
      style={{ top: 20 }}
    >
      <MCPInfoSection />
    </Modal>
  );
};

export default MCPModal;
