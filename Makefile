.PHONY: help dev build test lint clean install-deps

# 默认目标
.DEFAULT_GOAL := help

# 变量定义
BINARY_NAME=cloudboot-core
AGENT_NAME=cb-agent
PROVIDER_MOCK=provider-mock
BUILD_DIR=build
CGO_ENABLED=1
LDFLAGS=-ldflags="-s -w"

## help: 显示帮助信息
help:
	@echo "CloudBoot NG - Makefile帮助"
	@echo ""
	@echo "可用命令:"
	@echo "  make dev              - 启动开发环境 (Tailwind watch + Air)"
	@echo "  make build            - 构建生产二进制文件"
	@echo "  make test             - 运行所有测试"
	@echo "  make lint             - 运行代码检查"
	@echo "  make clean            - 清理构建产物"
	@echo "  make install-deps     - 安装开发依赖（Tailwind、Air等）"
	@echo ""

## install-deps: 安装开发依赖
install-deps:
	@echo "📦 安装开发依赖..."
	@# 检查并安装 Tailwind CSS CLI (直接下载，无需 npm)
	@if ! command -v tailwindcss &> /dev/null; then \
		echo "⬇️  下载 Tailwind CSS CLI..."; \
		curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-arm64; \
		chmod +x tailwindcss-macos-arm64; \
		mv tailwindcss-macos-arm64 /usr/local/bin/tailwindcss; \
	fi
	@# 检查并安装 Air (热重载工具)
	@if ! command -v air &> /dev/null; then \
		echo "⬇️  安装 Air..."; \
		go install github.com/cosmtrek/air@latest; \
	fi
	@# 安装 Go 依赖
	@echo "📥 安装 Go 依赖..."
	@go mod download
	@echo "✅ 依赖安装完成"

## dev: 启动开发环境
dev:
	@echo "🚀 启动开发环境..."
	@# 确保输出目录存在
	@mkdir -p web/static/css
	@# 先执行一次 Tailwind 构建，确保 output.css 存在
	@echo "🎨 初始构建 Tailwind CSS..."
	@tailwindcss -i web/static/css/input.css -o web/static/css/output.css
	@# 启动 Tailwind CSS watch (后台)
	@echo "👀 启动 Tailwind CSS watch..."
	@tailwindcss -i web/static/css/input.css -o web/static/css/output.css --watch &
	@# 启动 Air (热重载)
	@echo "🔥 启动 Air 热重载..."
	@air

## build: 构建生产二进制
build:
	@echo "🔨 构建生产版本..."
	@mkdir -p $(BUILD_DIR)
	@# 构建 CSS
	@echo "🎨 编译 Tailwind CSS (minified)..."
	@tailwindcss -i web/static/css/input.css -o web/static/css/output.css --minify
	@# 构建 CloudBoot Core
	@echo "🏗️  构建 $(BINARY_NAME)..."
	@CGO_ENABLED=$(CGO_ENABLED) go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) cmd/server/main.go
	@# 构建 Agent
	@echo "🏗️  构建 $(AGENT_NAME)..."
	@CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(AGENT_NAME) cmd/agent/main.go
	@# 构建 Mock Provider
	@echo "🏗️  构建 $(PROVIDER_MOCK)..."
	@CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(PROVIDER_MOCK) cmd/provider-mock/main.go
	@echo "✅ 构建完成! 输出目录: $(BUILD_DIR)/"
	@ls -lh $(BUILD_DIR)/

## test: 运行测试
test:
	@echo "🧪 运行测试..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | tail -n 1

## lint: 代码检查
lint:
	@echo "🔍 运行代码检查..."
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint 未安装，使用 go vet 替代"; \
		go vet ./...; \
	fi

## clean: 清理构建产物
clean:
	@echo "🧹 清理构建产物..."
	@rm -rf $(BUILD_DIR)
	@rm -f web/static/css/output.css
	@rm -f coverage.out
	@echo "✅ 清理完成"

## run: 运行 CloudBoot Core (开发模式)
run:
	@echo "🚀 启动 CloudBoot Core..."
	@go run cmd/server/main.go
