# All Services Launcher Makefile

.PHONY: build run clean test deps

# 构建可执行文件
build:
	go build -o all-services .

# 运行服务
run: build
	./all-services

# 清理构建产物
clean:
	rm -f all-services

# 下载依赖
deps:
	go mod tidy
	go mod download

# 测试编译
test-build:
	go build -o /tmp/all-services-test .
	rm -f /tmp/all-services-test

# 交叉编译
build-linux:
	GOOS=linux GOARCH=amd64 go build -o all-services-linux .

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o all-services-darwin .

build-windows:
	GOOS=windows GOARCH=amd64 go build -o all-services.exe .

# 构建所有平台
build-all: build-linux build-darwin build-windows

# 帮助信息
help:
	@echo "Available targets:"
	@echo "  build       - 构建可执行文件"
	@echo "  run         - 构建并运行服务"
	@echo "  clean       - 清理构建产物"
	@echo "  deps        - 下载项目依赖"
	@echo "  test-build  - 测试编译（不保留可执行文件）"
	@echo "  build-linux - 交叉编译 Linux 版本"
	@echo "  build-darwin- 交叉编译 macOS 版本"
	@echo "  build-windows- 交叉编译 Windows 版本"
	@echo "  build-all   - 构建所有平台版本"
	@echo "  help        - 显示此帮助信息"
