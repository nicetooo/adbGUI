# Gaze - 构建指南

## 快速开始

### 使用 Makefile（推荐）
```bash
make dev     # 启动开发服务器
make build   # 构建生产版本
make test    # 运行测试
```

### 直接使用 Wails CLI
```bash
wails dev -tags fts5      # 开发模式
wails build -tags fts5    # 生产构建
```

## 重要：FTS5 全文搜索

项目使用 SQLite FTS5 扩展实现高性能全文搜索。**必须**在编译时添加 `-tags fts5` 参数才能启用此功能。

### 未启用 FTS5 的影响
- ✅ 基本搜索功能正常（降级为 LIKE 查询）
- ❌ 无法使用高级搜索语法（AND、OR、NOT、短语匹配）
- ❌ 大数据量时搜索性能下降

### 启用 FTS5 的好处
- ✅ 支持布尔搜索：`"error" AND "network"`
- ✅ 支持短语匹配：`"send message"`
- ✅ 相关性排序（BM25 算法）
- ✅ 10000+ 事件时性能优势明显

### 验证 FTS5 是否启用
应用启动时查看日志：
```
[INF] FTS5 full-text search enabled [module=event_store]
```

如果看到警告：
```
[WRN] FTS5 not available, using LIKE fallback for search
```
说明编译时缺少 `-tags fts5` 参数。

## 开发环境要求

### 必需
- Go 1.21+
- Node.js 18+
- Wails CLI v2.11.0+

### 平台特定
- **macOS**: Xcode Command Line Tools
- **Linux**: `gcc`, `pkg-config`, `libgtk-3-dev`, `libwebkit2gtk-4.0-dev`
- **Windows**: MinGW-w64 or MSVC

## 构建标签说明

### fts5
启用 SQLite FTS5 全文搜索扩展。

**使用**:
```bash
wails dev -tags fts5
wails build -tags fts5
```

**影响**: 
- 增加约 500KB 二进制大小
- 首次创建 FTS 索引时略增写入延迟（几乎可忽略）
- 大幅提升搜索性能和功能

## 常见问题

### Q: 为什么不默认启用 FTS5？
A: `github.com/mattn/go-sqlite3` 驱动默认不包含扩展功能，需要通过 build tags 显式启用。未来可能通过预编译脚本自动处理。

### Q: 已有的事件数据会自动建立 FTS 索引吗？
A: 不会。FTS 索引通过触发器同步**新事件**。对于已存在的事件：
- **方案 1**: 删除旧数据库重新录制（推荐）
- **方案 2**: 手动重建索引（需要额外工具，暂未实现）

### Q: 能否在运行时切换 FTS5？
A: 不能。FTS5 是编译时决定的，无法运行时切换。

## 发布构建

生产发布时**必须**使用 `-tags fts5`:
```bash
make build
# 或
wails build -tags fts5 -clean
```

## 参考资料
- [Wails 文档](https://wails.io/docs/introduction)
- [go-sqlite3 Build Tags](https://github.com/mattn/go-sqlite3#feature-flags)
- [SQLite FTS5 文档](https://www.sqlite.org/fts5.html)
