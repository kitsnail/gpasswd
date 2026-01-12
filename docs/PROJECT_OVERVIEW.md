# gpasswd 项目概览

## 项目状态

**版本**: 0.1.0-dev (MVP 开发阶段)
**创建日期**: 2026-01-13
**许可证**: MIT

---

## 已完成

✅ **文档**
- MVP 设计文档 (`docs/MVP_DESIGN.md`)
- 安全模型文档 (`docs/SECURITY.md`)
- README (`README.md`)
- 项目结构规划

✅ **基础架构**
- Go 模块初始化
- 目录结构创建
- 数据模型定义 (`internal/models/entry.go`)
- 配置管理 (`pkg/config/config.go`)
- 基础入口文件 (`cmd/gpasswd/main.go`)

---

## 待实现模块

### Phase 1: 核心加密模块 (优先级: 最高)

**目录**: `internal/crypto/`

需要实现的文件：
- `kdf.go` - Argon2id 密钥派生
- `cipher.go` - AES-256-GCM 加密/解密
- `password.go` - 强密码生成器

**功能**：
```go
// kdf.go
func DeriveKey(password string, salt []byte, params Argon2Params) ([]byte, error)

// cipher.go
func Encrypt(plaintext []byte, key []byte) (ciphertext, nonce []byte, err error)
func Decrypt(ciphertext, nonce, key []byte) (plaintext []byte, err error)

// password.go
func Generate(length int, options GenerateOptions) (string, error)
func CheckStrength(password string) StrengthResult
```

---

### Phase 2: 存储模块 (优先级: 最高)

**目录**: `internal/storage/`

需要实现的文件：
- `db.go` - SQLite 数据库初始化和连接管理
- `entry.go` - Entry CRUD 操作
- `metadata.go` - 元数据管理（salt, argon2 参数）

**功能**：
```go
// db.go
func InitDB(path string) (*DB, error)
func (db *DB) Close() error

// entry.go
func (db *DB) CreateEntry(entry *models.Entry, key []byte) error
func (db *DB) GetEntry(id string, key []byte) (*models.Entry, error)
func (db *DB) ListEntries(key []byte) ([]*models.Entry, error)
func (db *DB) UpdateEntry(entry *models.Entry, key []byte) error
func (db *DB) DeleteEntry(id string) error
func (db *DB) SearchEntries(keyword string, key []byte) ([]*models.Entry, error)

// metadata.go
func (db *DB) GetSalt() ([]byte, error)
func (db *DB) SetSalt(salt []byte) error
func (db *DB) GetArgon2Params() (Argon2Params, error)
func (db *DB) SetArgon2Params(params Argon2Params) error
```

**数据库 Schema**:
```sql
-- 见 docs/MVP_DESIGN.md
```

---

### Phase 3: 会话管理 (优先级: 高)

**目录**: `internal/session/`

需要实现的文件：
- `manager.go` - 会话和密钥管理

**功能**：
```go
type Manager struct {
    key         []byte
    lastAccess  time.Time
    timeout     time.Duration
    mu          sync.RWMutex
}

func NewManager(timeout time.Duration) *Manager
func (m *Manager) SetKey(key []byte)
func (m *Manager) GetKey() ([]byte, error) // 返回 ErrSessionExpired 如果超时
func (m *Manager) Lock() // 清除密钥
func (m *Manager) IsLocked() bool
```

---

### Phase 4: 剪贴板操作 (优先级: 中)

**目录**: `internal/clipboard/`

需要实现的文件：
- `clipboard.go` - macOS 剪贴板操作

**功能**：
```go
func Copy(text string) error
func CopyWithAutoClear(text string, duration time.Duration) error
func Clear() error
```

**实现**：使用 `exec.Command("pbcopy")`

---

### Phase 5: CLI 命令 (优先级: 最高)

**目录**: `internal/cli/`

需要实现的文件：
- `root.go` - Cobra 根命令
- `init.go` - `gpasswd init` 命令
- `add.go` - `gpasswd add` 命令
- `list.go` - `gpasswd list` 命令
- `show.go` - `gpasswd show` 命令
- `copy.go` - `gpasswd copy` 命令
- `reveal.go` - `gpasswd reveal` 命令
- `edit.go` - `gpasswd edit` 命令
- `delete.go` - `gpasswd delete` 命令
- `search.go` - `gpasswd search` 命令
- `generate.go` - `gpasswd generate` 命令
- `lock.go` - `gpasswd lock` 命令
- `version.go` - `gpasswd version` 命令

**依赖**：
- `github.com/spf13/cobra` - CLI 框架
- `github.com/AlecAivazis/survey/v2` - 交互式输入

---

## 依赖库

需要添加到 `go.mod`：

```bash
go get github.com/spf13/cobra@latest
go get github.com/spf13/viper@latest
go get github.com/AlecAivazis/survey/v2@latest
go get github.com/mattn/go-sqlite3@latest
go get golang.org/x/crypto/argon2@latest
go get github.com/google/uuid@latest
```

---

## 实施顺序建议

### Week 1: 基础设施
1. 加密模块 (`internal/crypto/`)
2. 存储模块 (`internal/storage/`)
3. 单元测试

### Week 2: 初始化和核心操作
4. `init` 命令（创建保管库）
5. `add` 命令（添加条目）
6. `list` 命令（列出条目）
7. `delete` 命令（删除条目）

### Week 3: 密码访问
8. 会话管理 (`internal/session/`)
9. 剪贴板模块 (`internal/clipboard/`)
10. `copy` 命令（复制密码）
11. `reveal` 命令（显示密码）
12. `generate` 命令（生成密码）

### Week 4: 搜索和完善
13. `search` 命令（搜索条目）
14. `edit` 命令（编辑条目）
15. `show` 命令（查看详情）
16. `lock` 命令（锁定会话）
17. 错误处理优化
18. 集成测试

---

## 测试策略

### 单元测试
- 每个模块独立测试
- 覆盖率目标: > 70%
- 使用 `go test ./...`

### 集成测试
- 端到端场景测试
- 临时测试数据库
- 自动清理

### 安全测试
- 加密/解密正确性
- 密钥派生一致性
- 会话超时机制
- 剪贴板清除验证

---

## 构建和发布

### 开发构建
```bash
go build -o gpasswd cmd/gpasswd/main.go
```

### 生产构建
```bash
go build -ldflags="-s -w" -o gpasswd cmd/gpasswd/main.go
```

### 跨平台构建（未来）
```bash
GOOS=darwin GOARCH=amd64 go build -o gpasswd-darwin-amd64 cmd/gpasswd/main.go
GOOS=darwin GOARCH=arm64 go build -o gpasswd-darwin-arm64 cmd/gpasswd/main.go
GOOS=linux GOARCH=amd64 go build -o gpasswd-linux-amd64 cmd/gpasswd/main.go
```

---

## 性能目标

- 初始化保管库: < 300ms
- 添加条目: < 200ms
- 搜索（1000 条目）: < 50ms
- 复制密码: < 100ms
- 内存占用: < 50MB（空闲时）

---

## 安全 Checklist

开发过程中需要确保：

- [ ] 主密码永不存储到磁盘
- [ ] 使用 `crypto/rand` 而非 `math/rand`
- [ ] Nonce 每次加密独立生成
- [ ] 敏感数据在内存中及时清零
- [ ] 文件权限设置为 0600
- [ ] SQL 参数化查询（防止注入）
- [ ] 错误信息不泄露敏感数据
- [ ] 依赖库定期更新
- [ ] 代码审查覆盖加密模块

---

## 文档维护

需要保持更新的文档：
- `README.md` - 用户文档
- `docs/MVP_DESIGN.md` - 设计文档
- `docs/SECURITY.md` - 安全模型
- `CHANGELOG.md` - 版本变更（发布时创建）
- 代码注释和 Godoc

---

## 发布计划

### v0.1.0 - MVP
- 基础 CRUD 功能
- 加密存储
- 会话管理
- 剪贴板集成
- 密码生成
- 搜索功能

### v0.2.0 - 安全增强
- 密钥文件支持
- 密码强度评估
- 审计日志
- 自动备份

### v0.3.0 - 便利性
- 导入/导出
- 模糊搜索
- 密码历史

---

## 贡献指南

1. Fork 仓库
2. 创建特性分支
3. 编写代码和测试
4. 确保测试通过
5. 提交 Pull Request

**代码规范**：
- 遵循 `gofmt` 格式
- 添加必要的注释
- 单元测试覆盖新代码
- 更新相关文档

---

## 资源链接

- **仓库**: https://github.com/kitsnail/gpasswd
- **问题跟踪**: https://github.com/kitsnail/gpasswd/issues
- **安全报告**: security@gpasswd.dev

---

**最后更新**: 2026-01-13
