import React from 'react';
import { Select, Button, Divider } from 'antd';
import { ReloadOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useDeviceStore } from '../stores';

const { Option } = Select;

interface DeviceSelectorProps {
  style?: React.CSSProperties;
}

const DeviceSelector: React.FC<DeviceSelectorProps> = ({ style = {} }) => {
  const { t } = useTranslation();
  const { devices, selectedDevice, loading, setSelectedDevice, fetchDevices } = useDeviceStore();

  const handleRefresh = async () => {
    try {
      await fetchDevices();
    } catch {
      // Error handled by store
    }
  };

  return (
    <Select
      value={selectedDevice || undefined}
      onChange={setSelectedDevice}
      style={{ width: 220, ...style }}
      placeholder={t("device_selector.placeholder")}
      popupRender={(menu) => (
        <>
          {menu}
          <Divider style={{ margin: '8px 0' }} />
          <div style={{ padding: '0 8px 4px' }}>
            <Button
              type="text"
              icon={<ReloadOutlined />}
              onClick={handleRefresh}
              loading={loading}
              style={{ width: '100%', textAlign: 'left' }}
            >
              {t("device_selector.refresh")}
            </Button>
          </div>
        </>
      )}
    >
      {devices.map((d) => (
        <Option key={d.id} value={d.id}>
          {d.brand ? `${d.brand} ${d.model}` : d.model || d.id}
        </Option>
      ))}
    </Select>
  );
};

export default DeviceSelector;
