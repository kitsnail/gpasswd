# 快速开始指南

本指南帮助开发者快速上手 gpasswd 项目的开发。

---

## 环境要求

- **Go**: 1.21 或更高版本
- **macOS**: 当前仅支持 macOS（剪贴板功能依赖）
- **SQLite**: 通过 CGO 自动编译
- **Git**: 版本管理

---

## 克隆仓库

```bash
git clone https://github.com/kitsnail/gpasswd.git
cd gpasswd
```

---

## 安装依赖

```bash
# 安装 Go 依赖
go mod download

# 如果需要添加新依赖
go get github.com/spf13/cobra@latest
go get github.com/spf13/viper@latest
go get github.com/AlecAivazis/survey/v2@latest
go get github.com/mattn/go-sqlite3@latest
go get golang.org/x/crypto/argon2@latest
go get github.com/google/uuid@latest

# 整理依赖
go mod tidy
```

---

## 构建项目

```bash
# 开发构建
go build -o gpasswd cmd/gpasswd/main.go

# 运行
./gpasswd

# 或直接运行
go run cmd/gpasswd/main.go
```

---

## 运行测试

```bash
# 运行所有测试
go test ./...

# 运行测试并查看覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# 运行特定包的测试
go test ./internal/crypto/

# 详细输出
go test -v ./...
```

---

## 开发工作流

### 1. 创建新分支

```bash
git checkout -b feature/your-feature-name
```

### 2. 编写代码

遵循项目结构：
- 新功能代码放在 `internal/` 或 `pkg/`
- CLI 命令放在 `internal/cli/`
- 单元测试文件以 `_test.go` 结尾

### 3. 编写测试

```go
// internal/crypto/kdf_test.go
package crypto

import "testing"

func TestDeriveKey(t *testing.T) {
    password := "test_password"
    salt := []byte("test_salt_32_bytes_long_value")

    key, err := DeriveKey(password, salt, DefaultArgon2Params())
    if err != nil {
        t.Fatalf("DeriveKey failed: %v", err)
    }

    if len(key) != 32 {
        t.Errorf("Expected key length 32, got %d", len(key))
    }
}
```

### 4. 运行格式化和检查

```bash
# 格式化代码
go fmt ./...

# 静态检查
go vet ./...

# 使用 golangci-lint（推荐）
golangci-lint run
```

### 5. 提交代码

```bash
git add .
git commit -m "feat: add Argon2id key derivation"
git push origin feature/your-feature-name
```

### 6. 创建 Pull Request

在 GitHub 上创建 PR，等待代码审查。

---

## 调试技巧

### 使用 Delve 调试器

```bash
# 安装 Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 调试程序
dlv debug cmd/gpasswd/main.go

# 在 Delve 中设置断点
(dlv) break main.main
(dlv) continue
```

### 打印调试

```go
import "log"

log.Printf("Debug: key length = %d\n", len(key))
```

### 临时测试数据

```bash
# 创建临时测试目录
export GPASSWD_TEST_DIR=$(mktemp -d)
echo "Test directory: $GPASSWD_TEST_DIR"

# 测试完成后清理
rm -rf $GPASSWD_TEST_DIR
```

---

## 常见任务

### 添加新的 CLI 命令

1. 在 `internal/cli/` 创建新文件，如 `mycommand.go`
2. 实现命令：

```go
package cli

import (
    "github.com/spf13/cobra"
)

var myCommandCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "Short description",
    Long:  `Long description`,
    RunE: func(cmd *cobra.Command, args []string) error {
        // 实现逻辑
        return nil
    },
}

func init() {
    rootCmd.AddCommand(myCommandCmd)
}
```

3. 添加测试 `mycommand_test.go`

### 添加新的配置项

1. 修改 `pkg/config/config.go` 中的 `Config` 结构体
2. 更新 `DefaultConfig()` 函数
3. 更新 `docs/MVP_DESIGN.md` 中的配置示例

### 修改数据库 Schema

1. 修改 `internal/storage/db.go` 中的 `InitDB()` 函数
2. 添加迁移逻辑（如果需要向后兼容）
3. 更新相关的 CRUD 操作
4. 更新文档

---

## 项目结构导航

```
gpasswd/
├── cmd/gpasswd/          # 👈 命令行入口
│   └── main.go
├── internal/             # 👈 内部包（不对外暴露）
│   ├── cli/              # CLI 命令实现
│   ├── crypto/           # 加密模块
│   ├── storage/          # 数据库操作
│   ├── session/          # 会话管理
│   ├── clipboard/        # 剪贴板操作
│   └── models/           # 数据模型
├── pkg/                  # 👈 公共包（可被外部使用）
│   └── config/           # 配置管理
├── docs/                 # 👈 文档
│   ├── MVP_DESIGN.md     # MVP 设计
│   ├── SECURITY.md       # 安全模型
│   ├── PROJECT_OVERVIEW.md # 项目概览
│   └── QUICKSTART.md     # 本文档
└── scripts/              # 👈 脚本
    └── README.md
```

---

## 常见问题

### Q: 编译时提示 SQLite 相关错误

**A**: SQLite 需要 CGO，确保启用：
```bash
export CGO_ENABLED=1
go build cmd/gpasswd/main.go
```

### Q: 如何测试剪贴板功能？

**A**: 在测试中使用 mock：
```go
// clipboard_test.go
func TestCopy(t *testing.T) {
    // 使用 mockExecCommand 替代真实的 exec.Command
}
```

### Q: 如何处理敏感数据的测试？

**A**: 使用临时文件和延迟清理：
```go
func TestWithTempVault(t *testing.T) {
    tmpDir := t.TempDir() // 自动清理
    vaultPath := filepath.Join(tmpDir, "vault.db")

    db, err := InitDB(vaultPath)
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()

    // 测试逻辑...
}
```

---

## 开发资源

### Go 学习资源
- [Effective Go](https://go.dev/doc/effective_go)
- [Go by Example](https://gobyexample.com/)
- [Go 标准库文档](https://pkg.go.dev/std)

### 安全资源
- [OWASP Cheat Sheets](https://cheatsheetseries.owasp.org/)
- [Go Security Guidelines](https://golang.org/security/)
- [Argon2 RFC 9106](https://datatracker.ietf.org/doc/html/rfc9106)

### 依赖库文档
- [Cobra](https://github.com/spf13/cobra)
- [Viper](https://github.com/spf13/viper)
- [Survey](https://github.com/AlecAivazis/survey)
- [go-sqlite3](https://github.com/mattn/go-sqlite3)

---

## 下一步

1. 阅读 `docs/MVP_DESIGN.md` 了解整体设计
2. 阅读 `docs/SECURITY.md` 了解安全模型
3. 从实现 `internal/crypto/` 模块开始
4. 逐步完成 `docs/PROJECT_OVERVIEW.md` 中的实施计划

---

**祝开发顺利！**

有问题？提交 Issue 或在 Discussions 中讨论。
