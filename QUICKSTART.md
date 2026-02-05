# 快速开始

## 开发环境设置

### 1. 安装依赖
```bash
# macOS
brew install go node wails

# 安装 Go 依赖
go mod download
```

### 2. 启动开发服务器
```bash
make dev
```

**就这么简单！** 应用会自动启动，FTS5 全文搜索已启用。

## 常用命令

| 命令 | 说明 |
|------|------|
| `make dev` | 启动开发服务器（热重载） |
| `make build` | 构建生产版本 |
| `make test` | 运行测试 |
| `make clean` | 清理构建产物 |

## 构建配置

所有命令都自动包含 `-tags fts5`，无需手动添加。

如果你需要直接使用 `wails` 或 `go` 命令，记得加上 tag：
```bash
wails dev -tags fts5
wails build -tags fts5
go test -tags fts5 ./...
```

## 验证 FTS5 是否启用

启动应用后查看日志：
```
[INF] FTS5 full-text search enabled ✅
```

如果看到警告：
```
[WRN] FTS5 not available, using LIKE fallback
```
说明编译时缺少 `-tags fts5`，请使用 `make dev` 重新启动。

## 更多信息

- 详细构建说明：[README.build.md](README.build.md)
- 构建配置文档：[.build-config.md](.build-config.md)
- 项目文档：[CLAUDE.md](CLAUDE.md)
