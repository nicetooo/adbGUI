# Gaze - Makefile
# 简化开发和构建流程

.PHONY: help dev build clean test

help:
	@echo "Gaze Build Commands:"
	@echo "  make dev    - Start development server with FTS5 enabled"
	@echo "  make build  - Build production binary with FTS5 enabled"
	@echo "  make clean  - Clean build artifacts"
	@echo "  make test   - Run Go tests"

# 开发模式（启用 FTS5 全文搜索）
dev:
	@echo "🚀 Starting Wails dev with FTS5 enabled..."
	wails dev -tags fts5

# 生产构建（启用 FTS5）
build:
	@echo "🔨 Building for production with FTS5 enabled..."
	wails build -tags fts5

# 清理构建产物
clean:
	@echo "🧹 Cleaning build directory..."
	rm -rf build/bin

# 运行测试
test:
	@echo "🧪 Running Go tests..."
	go test -tags fts5 ./... -v

# 快速测试（不含 verbose 输出）
test-quick:
	@echo "🧪 Running Go tests (quick)..."
	go test -tags fts5 ./...
