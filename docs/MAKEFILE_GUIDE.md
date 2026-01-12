# Makefile 使用指南

gpasswd 项目提供了一个功能完整的 Makefile，简化开发、测试和发布流程。

---

## 快速开始

```bash
# 查看所有可用命令
make help

# 构建项目
make build

# 运行项目
make run

# 运行测试
make test
```

---

## 常用命令

### 开发阶段

| 命令 | 说明 | 使用场景 |
|------|------|---------|
| `make build` | 构建优化版本 | 正式构建 |
| `make build-dev` | 构建调试版本 | 开发调试 |
| `make run` | 构建并运行 | 快速测试 |
| `make clean` | 清理构建产物 | 清理环境 |
| `make watch` | 监听文件变化自动构建 | 开发模式（需要 entr） |

**示例**：
```bash
# 开发时快速迭代
make build-dev && ./build/gpasswd init

# 监听文件变化（需要安装 entr: brew install entr）
make watch
```

---

### 依赖管理

| 命令 | 说明 |
|------|------|
| `make deps` | 下载依赖 |
| `make tidy` | 整理 go.mod 和 go.sum |
| `make mod-upgrade` | 升级所有依赖到最新版本 |
| `make mod-vendor` | 创建 vendor 目录 |
| `make install-deps` | 安装开发工具（golangci-lint, delve） |

**示例**：
```bash
# 首次克隆后安装依赖
make deps

# 添加新依赖后整理
go get github.com/some/package@latest
make tidy
```

---

### 测试与质量

| 命令 | 说明 | 输出 |
|------|------|------|
| `make test` | 运行所有测试（含竞态检测） | 终端输出 |
| `make test-short` | 运行快速测试 | 终端输出 |
| `make coverage` | 生成覆盖率报告 | coverage.html |
| `make coverage-summary` | 显示覆盖率摘要 | 终端输出 |
| `make bench` | 运行性能测试 | 终端输出 |

**示例**：
```bash
# 运行测试并查看覆盖率
make coverage
# 自动在浏览器中打开 coverage.html

# 仅查看覆盖率百分比
make coverage-summary
```

---

### 代码质量检查

| 命令 | 说明 |
|------|------|
| `make fmt` | 格式化代码 |
| `make vet` | 运行 go vet |
| `make lint` | 运行 golangci-lint |
| `make check` | 运行所有检查（fmt + vet + lint + test） |

**示例**：
```bash
# 提交前运行完整检查
make check

# 如果 lint 失败，需要先安装
make install-deps
```

---

### 构建与发布

| 命令 | 说明 | 输出位置 |
|------|------|---------|
| `make build` | 构建当前平台 | build/gpasswd |
| `make build-all` | 构建所有平台 | build/gpasswd-* |
| `make build-darwin-amd64` | 构建 macOS Intel | build/gpasswd-darwin-amd64 |
| `make build-darwin-arm64` | 构建 macOS Apple Silicon | build/gpasswd-darwin-arm64 |
| `make build-linux-amd64` | 构建 Linux | build/gpasswd-linux-amd64 |
| `make release` | 创建发布版本 | build/release/ |

**示例**：
```bash
# 构建发布版本（包含压缩包和校验和）
make release

# 查看生成的文件
ls -lh build/release/
# 输出：
# gpasswd-0.1.0-dev-darwin-amd64
# gpasswd-0.1.0-dev-darwin-amd64.tar.gz
# gpasswd-0.1.0-dev-darwin-arm64
# gpasswd-0.1.0-dev-darwin-arm64.tar.gz
# checksums.txt
```

---

### 安装与卸载

| 命令 | 说明 | 需要权限 |
|------|------|---------|
| `make install` | 安装到 /usr/local/bin | sudo |
| `make uninstall` | 从 /usr/local/bin 卸载 | sudo |

**示例**：
```bash
# 安装到系统
make install
# 输入密码后可以全局使用 gpasswd

# 验证安装
which gpasswd
gpasswd --help

# 卸载
make uninstall
```

---

### 开发辅助

| 命令 | 说明 |
|------|------|
| `make size` | 显示二进制大小 |
| `make todo` | 查找代码中的 TODO 和 FIXME |
| `make version` | 显示版本信息 |
| `make info` | 显示项目信息 |

**示例**：
```bash
# 查看编译后的大小
make size
# 输出：
# Binary size:
# 1.4M build/gpasswd

# 查找待办事项
make todo
# 输出代码中所有 TODO 和 FIXME 注释
```

---

### 测试保管库管理

| 命令 | 说明 |
|------|------|
| `make init` | 初始化测试保管库 |
| `make clean-vault` | 删除测试保管库 |

**示例**：
```bash
# 创建测试保管库
make init

# 测试完成后清理
make clean-vault
```

---

## 完整工作流示例

### 日常开发

```bash
# 1. 拉取最新代码
git pull origin main

# 2. 安装/更新依赖
make deps

# 3. 创建功能分支
git checkout -b feature/add-crypto-module

# 4. 开发过程中监听变化（可选）
make watch

# 或者手动构建测试
make build-dev
./build/gpasswd

# 5. 编写测试
# ... 编辑 internal/crypto/kdf_test.go

# 6. 运行测试
make test

# 7. 检查覆盖率
make coverage

# 8. 提交前完整检查
make check

# 9. 提交代码
git add .
git commit -m "feat: add Argon2id key derivation"
git push origin feature/add-crypto-module
```

---

### 准备发布

```bash
# 1. 更新版本号
# 编辑 Makefile 中的 VERSION 变量
# VERSION=0.1.0

# 2. 运行完整检查
make check

# 3. 创建发布构建
make release

# 4. 验证构建产物
ls -lh build/release/
cat build/release/checksums.txt

# 5. 测试发布版本
./build/release/gpasswd-0.1.0-darwin-arm64 version

# 6. 创建 Git 标签
git tag v0.1.0
git push origin v0.1.0

# 7. 上传到 GitHub Releases
# 手动或通过 CI/CD
```

---

### 修复 Bug

```bash
# 1. 创建分支
git checkout -b fix/session-timeout

# 2. 编写失败的测试（TDD）
# 编辑 internal/session/manager_test.go
make test  # 应该失败

# 3. 修复代码
# 编辑 internal/session/manager.go
make test  # 应该通过

# 4. 确保覆盖率不下降
make coverage

# 5. 完整检查
make check

# 6. 提交
git commit -am "fix: correct session timeout behavior"
git push origin fix/session-timeout
```

---

### 性能优化

```bash
# 1. 运行基准测试（优化前）
make bench > bench-before.txt

# 2. 进行优化
# 编辑代码...

# 3. 运行基准测试（优化后）
make bench > bench-after.txt

# 4. 比较结果
# 使用 benchcmp 或手动对比

# 5. 确保功能正确
make test

# 6. 检查二进制大小
make size
```

---

## 高级用法

### 自定义构建参数

```bash
# 使用自定义版本号
make build VERSION=1.0.0-beta

# 使用自定义构建标志
GOFLAGS="-tags=debug" make build

# 禁用优化（保留符号表）
GOFLAGS="" make build
```

### 跨平台构建

```bash
# 在 macOS 上构建 Linux 版本
make build-linux-amd64

# 构建所有平台
make build-all

# 查看生成的文件
ls -lh build/
```

### 使用 vendor 模式

```bash
# 创建 vendor 目录
make mod-vendor

# 使用 vendor 构建（离线）
go build -mod=vendor -o build/gpasswd cmd/gpasswd/main.go
```

---

## 常见问题

### Q: make lint 报错 "golangci-lint not installed"

**A**: 安装开发依赖：
```bash
make install-deps
```

或手动安装：
```bash
# macOS
brew install golangci-lint

# Linux
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

---

### Q: make test 提示 CGO 相关错误

**A**: SQLite 需要 CGO，确保启用：
```bash
export CGO_ENABLED=1
make test
```

---

### Q: make watch 不工作

**A**: 需要安装 entr 工具：
```bash
# macOS
brew install entr

# Linux
sudo apt-get install entr  # Debian/Ubuntu
sudo yum install entr      # CentOS/RHEL
```

---

### Q: make coverage 不自动打开浏览器

**A**: 手动打开生成的文件：
```bash
make coverage
open coverage.html  # macOS
xdg-open coverage.html  # Linux
```

---

### Q: 如何清理所有构建产物和测试数据？

**A**: 运行完整清理：
```bash
make clean
make clean-vault
rm -rf ~/.gpasswd-test
```

---

## Makefile 变量说明

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `BINARY_NAME` | gpasswd | 二进制文件名 |
| `VERSION` | 0.1.0-dev | 版本号 |
| `BUILD_DIR` | build | 构建目录 |
| `CMD_PATH` | cmd/gpasswd/main.go | 入口文件 |
| `GOFLAGS` | -ldflags="-s -w" | Go 构建参数 |

可以通过环境变量覆盖：
```bash
VERSION=1.0.0 BUILD_DIR=dist make build
```

---

## 颜色输出说明

Makefile 使用颜色增强可读性：
- 🟢 **绿色**：成功、完成
- 🟡 **黄色**：警告、提示
- 🔴 **红色**：错误、失败

如果终端不支持颜色，输出仍然可读。

---

## 性能提示

1. **并行测试**：Go 测试默认并行运行
   ```bash
   # 指定并行度
   go test -parallel 4 ./...
   ```

2. **缓存构建**：Go 会缓存构建结果，重复构建很快

3. **增量编译**：只修改少量文件时，构建速度快

4. **使用 build-dev**：开发时不需要优化，构建更快
   ```bash
   make build-dev  # 比 make build 快 2-3 倍
   ```

---

## 集成 CI/CD

Makefile 可直接用于 CI/CD 流程：

### GitHub Actions 示例

```yaml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: make deps
      - run: make check
      - run: make coverage
      - uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
```

### GitLab CI 示例

```yaml
test:
  image: golang:1.21
  script:
    - make deps
    - make check
    - make coverage
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.xml
```

---

## 总结

gpasswd 的 Makefile 提供了：
- ✅ 30+ 个实用命令
- ✅ 彩色输出，易于阅读
- ✅ 完整的帮助文档（`make help`）
- ✅ 跨平台构建支持
- ✅ 开发、测试、发布一站式流程
- ✅ CI/CD 友好

**建议**：将 `make help` 加入书签，随时查看可用命令。

---

**相关文档**：
- [快速开始指南](QUICKSTART.md)
- [项目概览](PROJECT_OVERVIEW.md)
- [MVP 设计文档](MVP_DESIGN.md)
