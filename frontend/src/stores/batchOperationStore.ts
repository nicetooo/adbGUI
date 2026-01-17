import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

type OperationType = 'shell' | 'install' | 'uninstall' | 'clear' | 'stop' | 'push' | 'reboot';

interface BatchOperationModalState {
  operationType: OperationType;
  packageName: string;
  apkPath: string;
  command: string;
  localPath: string;
  remotePath: string;
  executed: boolean;

  // Actions
  setOperationType: (type: OperationType) => void;
  setPackageName: (name: string) => void;
  setApkPath: (path: string) => void;
  setCommand: (cmd: string) => void;
  setLocalPath: (path: string) => void;
  setRemotePath: (path: string) => void;
  setExecuted: (executed: boolean) => void;
  reset: () => void;
}

export const useBatchOperationStore = create<BatchOperationModalState>()(
  immer((set) => ({
    operationType: 'shell',
    packageName: '',
    apkPath: '',
    command: '',
    localPath: '',
    remotePath: '/sdcard/',
    executed: false,

    setOperationType: (type) => set((state) => {
      state.operationType = type;
    }),

    setPackageName: (name) => set((state) => {
      state.packageName = name;
    }),

    setApkPath: (path) => set((state) => {
      state.apkPath = path;
    }),

    setCommand: (cmd) => set((state) => {
      state.command = cmd;
    }),

    setLocalPath: (path) => set((state) => {
      state.localPath = path;
    }),

    setRemotePath: (path) => set((state) => {
      state.remotePath = path;
    }),

    setExecuted: (executed) => set((state) => {
      state.executed = executed;
    }),

    reset: () => set((state) => {
      state.operationType = 'shell';
      state.packageName = '';
      state.apkPath = '';
      state.command = '';
      state.localPath = '';
      state.remotePath = '/sdcard/';
      state.executed = false;
    }),
  }))
);
