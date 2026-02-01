# Gaze

**Wails**、**React**、**Ant Design** で構築された、強力でモダンな自己完結型の Android デバイス管理・自動化ツールです。デバイスの全行動を追跡するための統一された **Session-Event** アーキテクチャ、テスト自動化のためのビジュアル **Workflow** エンジン、そして AI によるデバイス制御のための完全な **MCP** (Model Context Protocol) 統合を備えています。


[English](README.md) | [简体中文](locales/README.zh-CN.md) | [繁體中文](locales/README.zh-TW.md) | [日本語](locales/README.ja.md) | [한국어](locales/README.ko.md) | [Español](locales/README.es.md) | [Português](locales/README.pt.md) | [Français](locales/README.fr.md) | [Deutsch](locales/README.de.md) | [Русский](locales/README.ru.md) | [Tiếng Việt](locales/README.vi.md) | [العربية](locales/README.ar.md)

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey.svg)
[![Website](https://img.shields.io/website?up_message=online&url=https%3A%2F%2Fgaze.nicetooo.com)](https://gaze.nicetooo.com)

## なぜ Gaze なのか？

- **モダンで高速**: Wails (Go + React) で構築されており、最小限のリソース消費でネイティブに近い体験を提供します。
- **真の自己完結型**: システムに `adb`、`scrcpy`、`aapt`、`ffmpeg`、`ffprobe` をインストールする必要はありません。すべてバンドル済みで、すぐに使用できます。
- **信頼性の高いファイル転送**: macOS で不安定になりがちな *Android File Transfer* の堅牢な代替手段です。
- **マルチデバイス対応**: 複数デバイスの独立した同時バックグラウンド録画をサポートします。
- **Session-Event アーキテクチャ**: すべてのデバイス活動（ログ、ネットワーク、タッチ、アプリライフサイクル）を単一のタイムライン上で統一的に追跡します。
- **ビジュアルワークフロー自動化**: ドラッグ＆ドロップのノードエディタで複雑なテストフローを構築できます — コード不要です。
- **MCP による AI 連携**: Model Context Protocol を通じて 50 以上のツールを公開し、Claude Desktop や Cursor などの AI クライアントとシームレスに統合できます。
- **開発者ファースト**: 開発者による、開発者のための統合 Logcat、Shell、MITM Proxy、UI Inspector を搭載しています。

## アプリのスクリーンショット

| デバイス管理 | 画面ミラーリング |
|:---:|:---:|
| <img src="screenshots/devices.png" width="400" /> | <img src="screenshots/mirror.png" width="400" /> |

| ファイルマネージャー | アプリ管理 |
|:---:|:---:|
| <img src="screenshots/files.png" width="400" /> | <img src="screenshots/apps.png" width="400" /> |

| パフォーマンスモニター | セッションタイムライン |
|:---:|:---:|
| <img src="screenshots/performance.png" width="400" /> | <img src="screenshots/session.png" width="400" /> |

| セッション一覧 | Logcat ビューア |
|:---:|:---:|
| <img src="screenshots/session-list.png" width="400" /> | <img src="screenshots/logcat.png" width="400" /> |

| ビジュアルワークフローエディター | ワークフロー一覧 |
|:---:|:---:|
| <img src="screenshots/workflow-editor.png" width="400" /> | <img src="screenshots/workflow.png" width="400" /> |

| UI インスペクター | タッチ録画 |
|:---:|:---:|
| <img src="screenshots/ui-inspector.png" width="400" /> | <img src="screenshots/recording.png" width="400" /> |

| ネットワークプロキシ (MITM) | ADB シェル |
|:---:|:---:|
| <img src="screenshots/proxy.png" width="400" /> | <img src="screenshots/shell.png" width="400" /> |

---

## 機能

### デバイス管理
- **統合デバイスリスト**: USB/Wi-Fi 自動統合により、物理デバイスとワイヤレスデバイスをシームレスに管理できます。
- **ワイヤレス接続**: mDNS サポート付きの IP/ポートペアリングで簡単に接続できます。
- **デバイス履歴**: 過去に接続したオフラインデバイスに素早くアクセスできます。
- **デバイスの固定**: よく使うデバイスを固定して、常にリストの先頭に表示できます。
- **デバイス監視**: バッテリー、ネットワーク、画面状態の変化をリアルタイムで追跡します。
- **一括操作**: 複数デバイスに対して同時に操作を実行できます。

### アプリ管理
- **完全なパッケージ制御**: インストール（ドラッグ＆ドロップ）、アンインストール、有効化、無効化、強制停止、データ消去。
- **APK 管理**: インストール済み APK のエクスポート、一括インストール。
- **スマートフィルタリング**: システム/ユーザーアプリで検索・フィルタリング。
- **クイックアクション**: アプリの起動やログへの直接ジャンプ。

### スクリーンミラーリング (Scrcpy)
- **高性能**: Scrcpy による低遅延ミラーリング。
- **録画**: 複数デバイスの同時録画をサポートする独立したバックグラウンド録画と、ワンクリックでのフォルダアクセス。
- **オーディオ転送**: デバイスの音声を PC にストリーミング（Android 11 以降）。
- **カスタマイズ**: 解像度、ビットレート、FPS、コーデック（H.264/H.265）を調整可能。
- **制御**: マルチタッチ対応、スリープ防止、画面オフモード。

### ファイル管理
- **多機能エクスプローラー**: 閲覧、コピー、切り取り、貼り付け、名前変更、削除、フォルダ作成。
- **ドラッグ＆ドロップ**: ウィンドウにファイルをドラッグするだけでアップロード。
- **ダウンロード**: デバイスから PC への簡単なファイル転送。
- **プレビュー**: ホストマシン上でファイルを直接開く。

### 高度な Logcat
- **リアルタイムストリーミング**: 自動スクロール制御付きのライブログビューア。
- **強力なフィルタリング**: ログレベル、タグ、PID、またはカスタム正規表現によるフィルタリング。
- **アプリ中心**: 特定のアプリケーションのログを自動フィルタリング。
- **JSON 整形**: 検出された JSON ログセグメントの自動整形表示。

### ネットワークとプロキシ (MITM)
- **自動キャプチャ**: ワンクリックで HTTP/HTTPS プロキシサーバーを起動し、ADB 経由でデバイスのプロキシ設定を自動構成。
- **HTTPS 復号化 (MITM)**: CA 証明書の自動生成と展開による SSL トラフィックの復号化をサポート。
- **WebSocket 対応**: リアルタイムの WebSocket トラフィックをキャプチャして検査可能。
- **大容量データの処理**: 切り捨てなしの完全なボディキャプチャ（最大 100MB）をサポートし、5000 件のログバッファを搭載。
- **トラフィックシェーピング**: デバイスごとのダウンロード/アップロード帯域幅制限と人工的な遅延により、実際のネットワーク環境をシミュレート。
- **ビジュアルメトリクス**: 選択したデバイスの RX/TX 速度をリアルタイムに監視。

### Session とイベント追跡
- **統一イベントパイプライン**: すべてのデバイス活動（ログ、ネットワークリクエスト、タッチイベント、アプリライフサイクル、アサーション）がイベントとしてキャプチャされ、Session タイムラインに関連付けられます。
- **自動 Session 管理**: イベント発生時に Session が自動作成されるほか、カスタム設定（logcat、録画、プロキシ、監視）を使った手動作成も可能です。
- **イベントタイムライン**: 時間ベースのインデックスとナビゲーションによる、すべてのイベントのマルチレーン表示。
- **全文検索**: SQLite FTS5 を使用した全イベントの横断検索。
- **バックプレッシャー制御**: 高負荷時の自動イベントサンプリングにより、重要なイベント（エラー、ネットワーク、ワークフロー）を保護。
- **イベントアサーション**: イベントストリームに対するアサーションを定義・評価し、自動検証を実現。
- **ビデオ同期**: ビジュアルデバッグのために、イベントタイムスタンプに同期したビデオフレームを抽出。

### UI インスペクターと自動化
- **UI 階層インスペクター**: 任意の画面の完全な UI ツリーを閲覧・分析。
- **要素ピッカー**: クリックで UI 要素を選択し、そのプロパティ（resource-id、テキスト、bounds、クラス）を検査。
- **タッチ録画**: タッチ操作を記録し、自動化スクリプトとして再生。
- **要素ベースのアクション**: セレクター（id、text、contentDesc、className、xpath）を使用して、UI 要素のクリック、長押し、テキスト入力、スワイプ、待機、アサートを実行。

### ビジュアルワークフローエンジン
- **ノードベースエディタ**: XYFlow を搭載したドラッグ＆ドロップインターフェースで自動化フローをビジュアルに構築。
- **30 以上のステップタイプ**: タップ、スワイプ、要素操作、アプリ制御、キーイベント、画面制御、待機、ADB コマンド、変数、分岐、サブワークフロー、Session 制御。
- **条件分岐**: exists/not_exists/text_equals/text_contains 条件によるインテリジェントなフローの作成。
- **変数と式**: 算術式サポート（`{{count}} + 1`）付きのワークフロー変数を使用。
- **ステップバイステップデバッグ**: 一時停止、ステップ実行、各ワークフローステップでの変数状態の検査。
- **Session 統合**: 包括的なテストレポートのために、ワークフロー内で追跡 Session を開始・停止。

### ADB シェル
- **統合コンソール**: アプリ内で直接 ADB コマンドを実行。
- **コマンド履歴**: 過去に実行したコマンドに素早くアクセス。

### システムトレイ
- **クイックアクセス**: メニューバー/システムトレイからミラーリング制御やデバイス状態を確認。
- **デバイスの固定**: よく使うデバイスをリストとトレイメニューの先頭に固定。
- **トレイ機能**: 固定デバイスの Logcat、Shell、ファイルマネージャーにトレイから直接アクセス。
- **録画インジケーター**: 録画中はトレイに赤いドットを表示。
- **バックグラウンド動作**: アプリをバックグラウンドで実行し、即座にアクセス可能。

---

## MCP 統合 (Model Context Protocol)

Gaze には 50 以上のツールと 5 つのリソースを公開する組み込み **MCP サーバー**が含まれており、AI クライアントが自然言語で Android デバイスを完全に制御できます。これにより、Gaze は AI と Android の橋渡し役となります。

### 対応 AI クライアント

| クライアント | トランスポート | 設定 |
|--------|-----------|---------------|
| **Claude Desktop** | SSE | `claude_desktop_config.json` |
| **Claude Code (CLI)** | SSE | `.claude/settings.json` |
| **Cursor** | SSE | Cursor MCP 設定 |

### クイックセットアップ

MCP サーバーは Gaze と共に `http://localhost:23816/mcp/sse` で自動的に起動します。

**Claude Desktop** (`claude_desktop_config.json`):
```json
{
  "mcpServers": {
    "gaze": {
      "url": "http://localhost:23816/mcp/sse"
    }
  }
}
```

**Claude Code**:
```bash
claude mcp add gaze --transport sse http://localhost:23816/mcp/sse
```

**Cursor**: Cursor の MCP 設定で MCP サーバー URL `http://localhost:23816/mcp/sse` を追加してください。

### MCP ツール (50 以上)

| カテゴリ | ツール | 説明 |
|----------|-------|-------------|
| **デバイス** | `device_list`, `device_info`, `device_connect`, `device_disconnect`, `device_pair`, `device_wireless`, `device_ip` | デバイスの検出、接続、情報取得 |
| **CLI ツール** | `adb_execute`, `aapt_execute`, `ffmpeg_execute`, `ffprobe_execute` | バンドルされた CLI ツール（ADB、AAPT、FFmpeg、FFprobe）の実行 |
| **アプリ** | `app_list`, `app_info`, `app_start`, `app_stop`, `app_running`, `app_install`, `app_uninstall`, `app_clear_data` | アプリケーションの完全なライフサイクル管理 |
| **スクリーン** | `screen_screenshot`, `screen_record_start`, `screen_record_stop`, `screen_recording_status` | スクリーンショット（base64）と録画制御 |
| **UI 自動化** | `ui_hierarchy`, `ui_search`, `ui_tap`, `ui_swipe`, `ui_input`, `ui_resolution` | UI 検査、要素操作、入力 |
| **Session** | `session_create`, `session_end`, `session_active`, `session_list`, `session_events`, `session_stats` | Session ライフサイクルとイベントクエリ |
| **Workflow** | `workflow_list`, `workflow_get`, `workflow_create`, `workflow_update`, `workflow_delete`, `workflow_run`, `workflow_stop`, `workflow_pause`, `workflow_resume`, `workflow_step_next`, `workflow_status`, `workflow_execute_step` | Workflow の完全な CRUD、実行、デバッグ |
| **プロキシ** | `proxy_start`, `proxy_stop`, `proxy_status` | ネットワークプロキシ制御 |
| **ビデオ** | `video_frame`, `video_metadata`, `session_video_frame`, `session_video_info` | ビデオフレームの抽出とメタデータ |

### MCP リソース

| URI | 説明 |
|-----|-------------|
| `gaze://devices` | 接続されたデバイスのリスト |
| `gaze://devices/{deviceId}` | デバイスの詳細情報 |
| `gaze://sessions` | アクティブおよび最近の Session |
| `workflow://list` | 保存されたすべての Workflow |
| `workflow://{workflowId}` | ステップを含む Workflow の詳細 |

### AI は Gaze で何ができるのか？

MCP 統合により、AI クライアントは以下のことが可能です：
- **テストの自動化**: 自然言語の指示で UI テストワークフローを作成・実行。
- **問題のデバッグ**: スクリーンショットの撮影、UI 階層の検査、ログの読み取り、ネットワークトラフィックの分析。
- **デバイス管理**: アプリのインストール、ファイル転送、複数デバイスにわたる設定の構成。
- **Workflow の構築**: 分岐ロジックと変数管理を持つ複雑な自動化ワークフローの生成。
- **Session の監視**: イベントベースの Session 記録による長期的なデバイス動作の追跡。

---

## 内蔵バイナリ

このアプリケーションは完全に自己完結型です。以下をバンドルしています：
- **ADB** (Android Debug Bridge)
- **Scrcpy** (スクリーンミラーリングと録画)
- **AAPT** (Android Asset Packaging Tool)
- **FFmpeg** (ビデオ/オーディオ処理)
- **FFprobe** (メディア分析)

起動時にこれらは一時ディレクトリに展開され、自動的に使用されます。システムの PATH を設定する必要はありません。

---

## Xiaomi/Poco/Redmi ユーザーへの重要な注意

Scrcpy で**タッチ操作**を有効にするには、以下の手順が必要です：
1. **開発者オプション**に移動します。
2. **USB デバッグ**を有効にします。
3. **USB デバッグ（セキュリティ設定）**を有効にします。
   *(注意: ほとんどの Xiaomi デバイスでは、SIM カードと Mi アカウントのログインが必要です)*

---

## はじめに

### 前提条件
- **Go** (v1.23 以降)
- **Node.js** (v18 LTS)
- **Wails CLI** (v2.9.2)
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@v2.9.2
  ```

### 開発
```bash
wails dev
```

### ビルド
```bash
wails build
```
コンパイルされたアプリケーションは `build/bin` に生成されます。

### テストの実行
```bash
go test ./...
```

### リリース
このプロジェクトは GitHub Actions を使用してマルチプラットフォームビルドを自動化しています。新しいリリースを作成するには：
1. コミットにタグを付ける: `git tag v1.0.0`
2. タグをプッシュする: `git push origin v1.0.0`
GitHub Action が macOS、Windows、Linux 用に自動的にビルドし、アーティファクトをリリースページにアップロードします。

---

## アーキテクチャ概要

```
                    +-----------------+
                    |   Wails (GUI)   |
                    +--------+--------+
                             |
              +--------------+--------------+
              |                             |
     +--------v--------+          +--------v--------+
     |  React Frontend |          |   Go Backend    |
     |  (Ant Design,   |          |  (App, Device,  |
     |   Zustand,      |          |   Automation,   |
     |   XYFlow)       |          |   Workflow)     |
     +-----------------+          +--------+--------+
                                           |
                         +-----------------+-----------------+
                         |                 |                 |
                +--------v------+  +-------v-------+  +-----v-------+
                | Event Pipeline|  |  MCP Server   |  |   Proxy     |
                | (Session,     |  |  (50+ tools,  |  |  (MITM,     |
                |  SQLite,      |  |   5 resources)|  |   goproxy)  |
                |  FTS5)        |  +---------------+  +-------------+
                +---------------+
```

---

## 技術スタック

| レイヤー | 技術 |
|-------|-----------|
| **デスクトップフレームワーク** | Wails v2 |
| **バックエンド** | Go 1.23+ |
| **フロントエンド** | React 18, TypeScript, Ant Design 6 |
| **状態管理** | Zustand |
| **Workflow エディタ** | XYFlow + Dagre |
| **データベース** | SQLite (WAL mode, FTS5) |
| **プロキシ** | goproxy |
| **MCP** | mcp-go (Model Context Protocol) |
| **国際化** | i18next (5 言語) |
| **ロギング** | zerolog |
| **チャート** | Recharts |

---

## トラブルシューティング

### macOS: 「App が壊れているため開けません」
GitHub からアプリをダウンロードした際に *「Gaze.app は壊れているため開けません」* というエラーが表示される場合、これは macOS Gatekeeper の隔離機能によるものです。

これを解決するには、ターミナルで以下のコマンドを実行してください：
```bash
sudo xattr -cr /path/to/Gaze.app
```
*(`/path/to/Gaze.app` はダウンロードしたアプリケーションの実際のパスに置き換えてください)*

> **または自分でビルドする:** Gatekeeper を回避したくない場合は、ローカルで[ソースからアプリをビルド](#はじめに)することも簡単にできます。数分で完了します！

### Windows: 「Windows によって PC が保護されました」
青い SmartScreen ポップアップがアプリの起動をブロックする場合：
1. **詳細情報**をクリックします。
2. **実行**をクリックします。

---

## ライセンス
このプロジェクトは MIT ライセンスの下でライセンスされています。
