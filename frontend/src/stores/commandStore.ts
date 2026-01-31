import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';
import { VIEW_KEYS, ViewKey } from './types';

// --- Result types ---

export type CommandResultType =
  | 'page'
  | 'command'
  | 'device'
  | 'proxy'
  | 'workflow'
  | 'session';

export interface CommandResult {
  id: string;
  type: CommandResultType;
  title: string;
  subtitle?: string;
  icon?: string;
  /** Action to execute when selected */
  action: () => void;
}

export interface CommandGroup {
  type: CommandResultType;
  label: string;
  results: CommandResult[];
}

// --- Fuzzy matching ---

/**
 * Simple fuzzy match: checks if all characters in `query` appear in `text` in order.
 * Returns a score (higher = better match), or -1 if no match.
 *
 * Scoring:
 *  - Exact prefix match: +100
 *  - Word boundary match: +50
 *  - Contains (substring): +30
 *  - Subsequence only: +10
 *  - Each matched char: +1
 */
export function fuzzyMatch(query: string, text: string): number {
  if (!query) return 0;
  const q = query.toLowerCase();
  const t = text.toLowerCase();

  // Exact prefix
  if (t.startsWith(q)) return 100 + q.length;

  // Contains substring
  if (t.includes(q)) return 30 + q.length;

  // Word boundary match: check if query matches start of words
  const words = t.split(/[\s\-_/:.]+/);
  const wordStarts = words.map((w) => w[0] || '').join('');
  if (wordStarts.includes(q)) return 50 + q.length;

  // Subsequence match
  let qi = 0;
  let matched = 0;
  for (let ti = 0; ti < t.length && qi < q.length; ti++) {
    if (t[ti] === q[qi]) {
      qi++;
      matched++;
    }
  }
  if (qi === q.length) return 10 + matched;

  return -1; // no match
}

// --- Static data builders ---

export interface PageItem {
  key: ViewKey;
  nameKey: string; // i18n key like "menu.devices"
  keywords: string[]; // additional match keywords (English)
}

export const PAGE_ITEMS: PageItem[] = [
  { key: VIEW_KEYS.DEVICES, nameKey: 'menu.devices', keywords: ['devices', 'phone', 'android', '设备', '手机', 'デバイス', '기기'] },
  { key: VIEW_KEYS.APPS, nameKey: 'menu.apps', keywords: ['apps', 'applications', 'packages', 'apk', '应用', '安装', 'アプリ', '앱'] },
  { key: VIEW_KEYS.SHELL, nameKey: 'menu.shell', keywords: ['shell', 'terminal', 'adb', 'command', '终端', '命令', 'ターミナル', '터미널'] },
  { key: VIEW_KEYS.LOGCAT, nameKey: 'menu.logcat', keywords: ['logcat', 'logs', 'logging', 'debug', '日志', '调试', 'ログ', '로그'] },
  { key: VIEW_KEYS.MIRROR, nameKey: 'menu.mirror', keywords: ['mirror', 'screen', 'scrcpy', 'display', '投屏', '屏幕', 'ミラー', '미러'] },
  { key: VIEW_KEYS.FILES, nameKey: 'menu.files', keywords: ['files', 'filesystem', 'storage', 'folder', '文件', '存储', 'ファイル', '파일'] },
  { key: VIEW_KEYS.PROXY, nameKey: 'menu.proxy', keywords: ['proxy', 'network', 'http', 'https', 'traffic', 'charles', '代理', '网络', '抓包', 'プロキシ', '프록시'] },
  { key: VIEW_KEYS.RECORDING, nameKey: 'menu.recording', keywords: ['recording', 'screenrecord', 'video', 'capture', '录制', '录屏', '録画', '녹화'] },
  { key: VIEW_KEYS.WORKFLOW, nameKey: 'menu.workflow', keywords: ['workflow', 'automation', 'test', 'flow', '工作流', '自动化', 'ワークフロー', '워크플로'] },
  { key: VIEW_KEYS.INSPECTOR, nameKey: 'menu.inspector', keywords: ['inspector', 'ui', 'layout', 'hierarchy', 'element', '检查', '布局', '元素', 'インスペクタ', '인스펙터'] },
  { key: VIEW_KEYS.EVENTS, nameKey: 'menu.events', keywords: ['events', 'timeline', 'session', '事件', '时间线', 'イベント', '이벤트'] },
  { key: VIEW_KEYS.SESSIONS, nameKey: 'menu.sessions', keywords: ['sessions', 'history', 'list', '会话', '历史', 'セッション', '세션'] },
  { key: VIEW_KEYS.PERFORMANCE, nameKey: 'menu.performance', keywords: ['performance', 'perf', 'cpu', 'memory', 'fps', 'battery', '性能', '内存', '电池', 'パフォーマンス', '성능'] },
];

export interface CommandItem {
  id: string;
  nameKey: string; // i18n key like "command.cmd_toggle_theme"
  keywords: string[]; // additional match keywords
  /** action ID - actual action is bound at runtime */
}

export const COMMAND_ITEMS: CommandItem[] = [
  { id: 'toggle_theme', nameKey: 'command.cmd_toggle_theme', keywords: ['theme', 'dark', 'light', 'mode', '主题', '深色', '浅色', '暗色', 'テーマ', '테마'] },
  { id: 'wireless_connect', nameKey: 'command.cmd_wireless_connect', keywords: ['wireless', 'wifi', 'connect', 'pair', '无线', '连接', '配对', 'ワイヤレス', '무선'] },
  { id: 'mock_rules', nameKey: 'command.cmd_mock_rules', keywords: ['mock', 'rules', 'fake', 'response', '管理', '规则', '列表', 'ルール', '규칙'] },
  { id: 'add_mock_rule', nameKey: 'command.cmd_add_mock_rule', keywords: ['mock', 'add', 'create', 'new', 'rule', 'fake', '添加', '新建', '创建', '追加', '추가'] },
  { id: 'map_remote_rules', nameKey: 'command.cmd_map_remote_rules', keywords: ['map', 'remote', 'redirect', 'forward', '远程', '映射', '重定向', 'リダイレクト', '리다이렉트'] },
  { id: 'add_map_remote_rule', nameKey: 'command.cmd_add_map_remote_rule', keywords: ['map', 'remote', 'add', 'create', 'new', 'redirect', '添加', '新建', '创建', '追加', '추가'] },
  { id: 'rewrite_rules', nameKey: 'command.cmd_rewrite_rules', keywords: ['rewrite', 'modify', 'replace', '重写', '修改', '替换', 'リライト', '리라이트'] },
  { id: 'add_rewrite_rule', nameKey: 'command.cmd_add_rewrite_rule', keywords: ['rewrite', 'add', 'create', 'new', 'modify', '添加', '新建', '创建', '追加', '추가'] },
  { id: 'breakpoint_rules', nameKey: 'command.cmd_breakpoint_rules', keywords: ['breakpoint', 'intercept', 'pause', 'debug', '断点', '拦截', '调试', 'ブレーク', '브레이크'] },
  { id: 'add_breakpoint_rule', nameKey: 'command.cmd_add_breakpoint_rule', keywords: ['breakpoint', 'add', 'create', 'new', 'intercept', '添加', '新建', '创建', '追加', '추가'] },
  { id: 'proto_files', nameKey: 'command.cmd_proto_files', keywords: ['proto', 'protobuf', 'grpc', 'schema', '协议', 'プロト', '프로토'] },
  { id: 'new_workflow', nameKey: 'command.cmd_new_workflow', keywords: ['new', 'create', 'workflow', 'automation', '新建', '创建', '工作流', '自动化', 'ワークフロー', '워크플로'] },
  { id: 'about', nameKey: 'command.cmd_about', keywords: ['about', 'version', 'info', '关于', '版本', 'バージョン', '정보'] },
  { id: 'feedback', nameKey: 'command.cmd_feedback', keywords: ['feedback', 'bug', 'report', 'issue', '反馈', '报告', 'フィードバック', '피드백'] },
];

// --- Store ---

interface CommandState {
  // Visibility
  isOpen: boolean;
  query: string;
  selectedIndex: number;

  // Actions
  open: () => void;
  close: () => void;
  toggle: () => void;
  setQuery: (query: string) => void;
  setSelectedIndex: (index: number) => void;
  moveSelection: (delta: number, maxIndex: number) => void;
  reset: () => void;
}

export const useCommandStore = create<CommandState>()(
  immer((set) => ({
    isOpen: false,
    query: '',
    selectedIndex: 0,

    open: () =>
      set((state) => {
        state.isOpen = true;
        state.query = '';
        state.selectedIndex = 0;
      }),

    close: () =>
      set((state) => {
        state.isOpen = false;
        state.query = '';
        state.selectedIndex = 0;
      }),

    toggle: () =>
      set((state) => {
        if (state.isOpen) {
          state.isOpen = false;
          state.query = '';
          state.selectedIndex = 0;
        } else {
          state.isOpen = true;
          state.query = '';
          state.selectedIndex = 0;
        }
      }),

    setQuery: (query) =>
      set((state) => {
        state.query = query;
        state.selectedIndex = 0;
      }),

    setSelectedIndex: (index) =>
      set((state) => {
        state.selectedIndex = index;
      }),

    moveSelection: (delta, maxIndex) =>
      set((state) => {
        let next = state.selectedIndex + delta;
        if (next < 0) next = maxIndex;
        if (next > maxIndex) next = 0;
        state.selectedIndex = next;
      }),

    reset: () =>
      set((state) => {
        state.query = '';
        state.selectedIndex = 0;
      }),
  }))
);
