# gpasswd

> **g**o **passwd** - 安全的本地密码管理 CLI 工具

一个专为 macOS 开发者设计的命令行密码管理器，本地存储，强加密，简单易用。

---

## 特性

- **本地优先**：所有数据加密存储在本地，无云依赖
- **军事级加密**：AES-256-GCM + Argon2id 密钥派生
- **CLI 友好**：为终端用户优化的交互体验
- **快速搜索**：基于 SQLite FTS5 的全文搜索
- **自动清除**：剪贴板自动清除，防止密码泄露
- **会话管理**：智能会话超时，平衡安全与便利
- **开源透明**：代码公开，安全可审计

---

## 快速开始

### 安装

```bash
# 方式 1：从源码构建
git clone https://github.com/kitsnail/gpasswd.git
cd gpasswd
go build -o gpasswd cmd/gpasswd/main.go
sudo mv gpasswd /usr/local/bin/

# 方式 2：使用 go install（即将支持）
go install github.com/kitsnail/gpasswd@latest

# 方式 3：下载预编译二进制（即将支持）
# 访问 Releases 页面下载
```

### 初始化

```bash
# 首次使用，初始化保管库
gpasswd init

# 设置主密码（至少 12 字符，包含大小写数字符号）
Enter master password: ****
Confirm master password: ****
✓ Vault initialized at ~/.gpasswd/
```

### 基本使用

```bash
# 添加新账号（交互式）
gpasswd add

# 列出所有账号
gpasswd list

# 复制密码到剪贴板（30秒后自动清除）
gpasswd copy "Gmail Work"

# 搜索账号
gpasswd search github

# 生成强密码
gpasswd generate --length 20

# 查看帮助
gpasswd --help
```

---

## 命令概览

| 命令 | 说明 |
|------|------|
| `gpasswd init` | 初始化保管库并设置主密码 |
| `gpasswd add` | 添加新的密码条目（交互式） |
| `gpasswd list [--category TYPE]` | 列出所有或指定分类的条目 |
| `gpasswd show <name>` | 查看条目详情（隐藏密码） |
| `gpasswd copy <name>` | 复制密码到剪贴板（自动清除） |
| `gpasswd reveal <name>` | 在终端显示密码（需确认） |
| `gpasswd edit <name>` | 编辑条目 |
| `gpasswd delete <name>` | 删除条目（需确认） |
| `gpasswd search <keyword>` | 搜索条目 |
| `gpasswd generate [OPTIONS]` | 生成强密码 |
| `gpasswd lock` | 立即锁定会话 |
| `gpasswd version` | 显示版本信息 |

---

## 使用示例

### 添加账号

```bash
$ gpasswd add
[gpasswd] Master Password: ****

Title: GitHub Personal
Category: (website)
Username: john@example.com
Password: (generate/input)? generate
  ✓ Generated: xK9$mP2@vL4#nR8&qT3!
URL (optional): https://github.com
Tags (comma-separated): code,work
Notes (optional): Personal GitHub account with 2FA enabled

✓ Entry saved successfully (ID: a3f2c8e1-...)
```

### 获取密码

```bash
$ gpasswd copy "GitHub Personal"
[gpasswd] Master Password: ****
✓ Password copied to clipboard (will clear in 30 seconds)

# 30秒后...
✓ Clipboard cleared
```

### 搜索账号

```bash
$ gpasswd search github
Found 3 entries:
  1. GitHub Personal (website) - john@example.com
  2. GitHub Work (website) - john@company.com
  3. GitHub API Token (api-key) - **hidden**
```

### 生成密码

```bash
$ gpasswd generate --length 24 --no-symbols
Generated password: Xk9mP2vL4nR8qT3wY6zA5bC7

$ gpasswd generate --length 16
Generated password: xK9$mP2@vL4#nR8&
```

---

## 配置

配置文件位置：`~/.gpasswd/config.yaml`

```yaml
# 会话配置
session:
  timeout: 300           # 会话超时（秒），0 = 永不超时

# 剪贴板配置
clipboard:
  clear_timeout: 30      # 剪贴板清除时间（秒）

# 密码生成器默认配置
password_generator:
  length: 20
  use_uppercase: true
  use_lowercase: true
  use_digits: true
  use_symbols: true
  exclude_ambiguous: false  # 排除易混淆字符（0/O, 1/l/I）

# 安全配置
security:
  failed_attempts_limit: 5
  lockout_duration: 30   # 锁定时间（秒）

# Argon2 参数（高级用户）
argon2:
  time_cost: 3
  memory_cost: 65536     # KB（64MB）
  parallelism: 4
```

---

## 安全性

### 加密设计

- **密钥派生**：Argon2id（64MB 内存，3 迭代）
- **数据加密**：AES-256-GCM（认证加密）
- **零知识架构**：主密码永不存储，仅在内存中派生密钥
- **每条目独立加密**：每个条目使用独立的 Nonce

### 威胁模型

✅ **防护的攻击**：
- 数据库文件泄露（加密保护）
- 暴力破解（Argon2id 高成本）
- 数据篡改（GCM 认证标签）
- 剪贴板遗留（自动清除）

❌ **无法防护的攻击**：
- 键盘记录器（依赖系统安全）
- 物理访问已解锁的电脑
- 内存转储（会话期间）

详细安全模型请参阅：[SECURITY.md](docs/SECURITY.md)

### 最佳实践

1. **使用强主密码**：至少 12 字符，建议使用密码短语
2. **定期备份**：`~/.gpasswd/` 目录包含所有数据
3. **启用 FileVault**：macOS 磁盘加密
4. **离开时锁定**：执行 `gpasswd lock` 或设置较短超时
5. **私密环境使用**：避免在公共场所使用 `reveal` 命令

---

## 架构

### 技术栈

- **语言**：Go 1.21+
- **加密**：`crypto/aes`, `golang.org/x/crypto/argon2`
- **存储**：SQLite 3
- **CLI 框架**：Cobra
- **交互**：Survey

### 项目结构

```
gpasswd/
├── cmd/gpasswd/          # 命令行入口
├── internal/
│   ├── cli/              # CLI 命令实现
│   ├── crypto/           # 加密模块
│   ├── storage/          # 存储模块
│   ├── session/          # 会话管理
│   └── clipboard/        # 剪贴板操作
├── pkg/config/           # 配置管理
└── docs/                 # 文档
    ├── MVP_DESIGN.md     # MVP 设计文档
    └── SECURITY.md       # 安全模型
```

---

## 开发

### 构建

```bash
# 开发构建
go build -o gpasswd cmd/gpasswd/main.go

# 生产构建（缩小体积）
go build -ldflags="-s -w" -o gpasswd cmd/gpasswd/main.go

# 运行测试
go test ./...

# 测试覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 贡献

欢迎贡献代码、报告 Bug 或提出建议！

1. Fork 本仓库
2. 创建特性分支：`git checkout -b feature/amazing-feature`
3. 提交更改：`git commit -m 'Add amazing feature'`
4. 推送分支：`git push origin feature/amazing-feature`
5. 创建 Pull Request

**注意**：
- 所有加密相关代码必须经过审查
- 请遵循 Go 代码规范
- 添加必要的单元测试

---

## 路线图

### v0.1.0 - MVP（当前）
- [x] 基础 CRUD 操作
- [x] 加密存储（AES-256-GCM + Argon2id）
- [x] 会话管理
- [x] 剪贴板集成
- [ ] 密码生成器
- [ ] 搜索功能

### v0.2.0 - 安全增强
- [ ] 可选的密钥文件（第二因素）
- [ ] 密码强度评估
- [ ] 审计日志（加密）
- [ ] 自动备份

### v0.3.0 - 便利性
- [ ] 导入/导出（1Password, LastPass, CSV）
- [ ] 快速模糊搜索（fzf 集成）
- [ ] 密码历史记录
- [ ] 浏览器集成（Native Messaging）

### v0.4.0 - 高级功能
- [ ] TOTP 生成器
- [ ] SSH 密钥管理
- [ ] 附件存储
- [ ] 密码健康监控

### v0.5.0 - 同步（可选）
- [ ] Git-based 同步
- [ ] WebDAV 支持
- [ ] 端到端加密云备份

---

## 常见问题

### 忘记主密码怎么办？

无法恢复。这是零知识架构的设计特性，确保即使开发者也无法访问你的数据。

**建议**：将主密码写在纸上，存放在安全的地方（如保险箱）。

### 如何在多台电脑间同步？

MVP 版本不支持同步。你可以手动备份 `~/.gpasswd/` 目录。

未来版本将支持：
- Git 仓库同步（数据已加密，可安全推送）
- WebDAV 同步
- 加密云存储（Dropbox/iCloud）

### 与 1Password/LastPass 比较？

| 特性 | gpasswd | 1Password | LastPass |
|------|---------|-----------|----------|
| 本地存储 | ✅ | 部分 | ❌ |
| CLI 友好 | ✅ | ✅ | ❌ |
| 开源 | ✅ | 部分 | ❌ |
| 免费 | ✅ | ❌ | 部分 |
| 浏览器集成 | 未来 | ✅ | ✅ |
| 云同步 | 未来 | ✅ | ✅ |
| 移动端 | ❌ | ✅ | ✅ |

**定位**：gpasswd 专注于本地 CLI 使用，适合开发者和注重隐私的用户。

### 支持 Windows/Linux 吗？

MVP 版本仅支持 macOS（剪贴板操作使用 `pbcopy`）。

未来版本计划跨平台支持，只需替换剪贴板模块。

### 数据存储在哪里？

```
~/.gpasswd/
├── vault.db       # 加密的 SQLite 数据库
└── config.yaml    # 配置文件
```

**备份建议**：定期备份此目录，或使用 Time Machine。

---

## 安全漏洞报告

如果你发现安全漏洞，请**不要公开披露**。

**报告方式**：
- 邮件：security@gpasswd.dev
- GitHub Security Advisory（私有）

我们承诺 48 小时内响应，并遵循负责任披露原则。

---

## 致谢

灵感来源于：
- [1Password](https://1password.com/) - 安全设计
- [pass](https://www.passwordstore.org/) - Unix 哲学
- [Bitwarden](https://bitwarden.com/) - 开源精神

依赖的优秀项目：
- [Cobra](https://github.com/spf13/cobra) - CLI 框架
- [Argon2](https://github.com/P-H-C/phc-winner-argon2) - 密钥派生
- [SQLite](https://www.sqlite.org/) - 嵌入式数据库

---

## 许可证

[MIT License](LICENSE)

Copyright (c) 2026 gpasswd contributors

---

## 联系

- **GitHub**：https://github.com/kitsnail/gpasswd
- **问题反馈**：https://github.com/kitsnail/gpasswd/issues
- **讨论区**：https://github.com/kitsnail/gpasswd/discussions

---

**注意**：gpasswd 目前处于 MVP 开发阶段，不建议在生产环境使用。正式版本发布前，数据格式可能发生变化。
