# gpasswd MVP 设计文档

## 项目定位

**本地优先的 macOS 命令行密码管理工具**

### 核心理念
- **本地优先**：所有数据存储在本地，无云依赖
- **KISS 原则**：保持简单，专注核心功能
- **安全第一**：强加密，零信任设计
- **CLI 友好**：为开发者和高级用户设计
- **可扩展**：架构清晰，便于后续迭代

---

## MVP 功能范围

### 核心功能（Must Have）

1. **初始化保管库**
   - 设置主密码
   - 创建加密数据库
   - 生成本地配置

2. **基本 CRUD 操作**
   - 添加账号条目（交互式）
   - 列出所有条目（简洁展示）
   - 查看条目详情
   - 编辑条目
   - 删除条目

3. **密码访问**
   - 复制密码到剪贴板（自动清除）
   - 显示密码（带二次确认）

4. **密码生成**
   - 生成强随机密码
   - 可配置长度和字符类型

5. **搜索功能**
   - 按名称/标签搜索
   - 模糊匹配

6. **安全机制**
   - 主密码验证
   - 会话超时（可配置）
   - 剪贴板自动清除

### 明确不包含（Out of Scope）

❌ 云同步
❌ 浏览器集成
❌ GUI 界面
❌ 多用户/团队功能
❌ 附件存储
❌ TOTP/2FA 生成
❌ 导入/导出（第一版）
❌ 密码健康检查
❌ 与 Have I Been Pwned 集成

---

## 技术架构

### 技术栈

```
语言：       Go 1.21+
加密库：     crypto/aes, golang.org/x/crypto/argon2
存储：       SQLite (github.com/mattn/go-sqlite3)
CLI框架：    cobra (github.com/spf13/cobra)
配置：       viper (github.com/spf13/viper)
交互：       survey (github.com/AlecAivazis/survey/v2)
```

### 为什么选择这些技术？

**Go**：
- 单二进制文件，分发简单
- 优秀的标准库（crypto）
- 性能好，内存安全

**SQLite**：
- 无需额外进程
- 事务支持
- 全文搜索（FTS5）
- 成熟稳定

**Cobra**：
- Go CLI 事实标准
- 子命令管理清晰
- 自动生成帮助文档

**Argon2**：
- 比 PBKDF2 更安全
- 抗暴力破解能力强
- 现代密码哈希标准

---

## 安全设计

### 加密方案

```
主密码 (User Input)
    ↓
Argon2id (时间成本=3, 内存成本=64MB, 并行度=4, Salt=随机32字节)
    ↓
派生密钥 (32字节)
    ↓
AES-256-GCM 加密
    ↓
加密数据存储
```

**关键点**：
- 主密码**永不存储**，每次使用时从用户输入派生
- 每个条目使用独立的随机 Nonce（GCM 模式）
- Salt 在初始化时生成，存储在数据库元数据中

### 会话管理

```
用户输入主密码
    ↓
派生加密密钥
    ↓
密钥加密存储在内存中（[]byte）
    ↓
设置定时器（默认 5 分钟）
    ↓
超时后清零密钥（runtime.MemProfileRate）
    ↓
下次操作需重新输入主密码
```

**安全措施**：
- 密钥仅存在内存，不写入磁盘
- 超时后主动清零内存
- 使用 `defer` 确保退出时清理

### 剪贴板安全

```
复制密码到剪贴板
    ↓
启动后台 goroutine
    ↓
等待 30 秒（可配置）
    ↓
清空剪贴板
    ↓
提示用户已清除
```

---

## 数据模型

### 数据库 Schema

```sql
-- 元数据表（存储全局配置）
CREATE TABLE metadata (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
-- 存储：salt, argon2_params, created_at, updated_at

-- 条目表（存储加密的密码条目）
CREATE TABLE entries (
    id TEXT PRIMARY KEY,              -- UUID
    encrypted_data BLOB NOT NULL,     -- AES-256-GCM 加密的 JSON
    nonce BLOB NOT NULL,              -- GCM Nonce (12 bytes)
    search_text TEXT,                 -- 明文搜索字段（仅名称、标签）
    created_at INTEGER NOT NULL,      -- Unix 时间戳
    updated_at INTEGER NOT NULL
);

-- 全文搜索索引
CREATE VIRTUAL TABLE entries_fts USING fts5(
    entry_id UNINDEXED,
    search_text,
    content=entries,
    content_rowid=rowid
);
```

### Entry 数据结构（加密前）

```go
type Entry struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`       // 例如："Gmail Work"
    Category  string    `json:"category"`   // 例如："email", "api-key", "website"
    Username  string    `json:"username"`   // 可选
    Password  string    `json:"password"`   // 敏感字段
    URL       string    `json:"url"`        // 可选
    Notes     string    `json:"notes"`      // 可选，加密存储
    Tags      []string  `json:"tags"`       // 例如：["work", "google"]
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

**加密策略**：
- `password`, `username`, `url`, `notes` 全部加密
- `name`, `category`, `tags` 生成 `search_text` 明文存储（便于搜索）
- 权衡：搜索便利性 vs 隐私（MVP 选择便利性）

---

## 项目目录结构

```
gpasswd/
├── cmd/
│   └── gpasswd/
│       └── main.go                # 入口文件
├── internal/
│   ├── cli/                       # CLI 命令实现
│   │   ├── root.go               # 根命令
│   │   ├── init.go               # init 命令
│   │   ├── add.go                # add 命令
│   │   ├── list.go               # list 命令
│   │   ├── get.go                # get 命令
│   │   ├── copy.go               # copy 命令
│   │   ├── edit.go               # edit 命令
│   │   ├── delete.go             # delete 命令
│   │   ├── search.go             # search 命令
│   │   └── generate.go           # generate 命令
│   ├── crypto/                    # 加密模块
│   │   ├── kdf.go                # 密钥派生（Argon2）
│   │   ├── cipher.go             # AES-GCM 加密/解密
│   │   └── password.go           # 密码生成器
│   ├── storage/                   # 存储模块
│   │   ├── db.go                 # SQLite 数据库操作
│   │   ├── entry.go              # Entry CRUD
│   │   └── metadata.go           # 元数据管理
│   ├── session/                   # 会话管理
│   │   └── manager.go            # 会话和密钥管理
│   ├── clipboard/                 # 剪贴板操作
│   │   └── clipboard.go          # 复制和自动清除
│   └── models/                    # 数据模型
│       └── entry.go              # Entry 结构定义
├── pkg/
│   └── config/                    # 配置管理
│       └── config.go             # 配置文件读写
├── docs/
│   ├── MVP_DESIGN.md             # 本文档
│   └── SECURITY.md               # 安全模型说明
├── scripts/
│   └── install.sh                # 安装脚本
├── .gitignore
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

**设计原则**：
- `internal/`：不对外暴露的内部实现
- `pkg/`：可复用的公共包
- `cmd/`：可执行文件入口
- 清晰的模块划分，便于测试和维护

---

## CLI 命令设计

### 命令列表

```bash
# 初始化
gpasswd init

# 添加条目（交互式）
gpasswd add

# 列出所有条目
gpasswd list
gpasswd list --category email

# 查看条目详情（隐藏密码）
gpasswd show <name|id>

# 复制密码到剪贴板
gpasswd copy <name|id>

# 显示密码（明文，需二次确认）
gpasswd reveal <name|id>

# 编辑条目（交互式）
gpasswd edit <name|id>

# 删除条目（需确认）
gpasswd delete <name|id>

# 搜索条目
gpasswd search <keyword>

# 生成密码
gpasswd generate [--length 20] [--no-symbols]

# 锁定会话（清除内存中的密钥）
gpasswd lock

# 版本信息
gpasswd version
```

### 交互式输入示例

```bash
$ gpasswd add
[gpasswd] Master Password: ****

Title: Gmail Work Account
Category: (email) email
Username: john@company.com
Password: (generate/input)? generate
  ✓ Generated: xK9$mP2@vL4#nR8&qT3!
URL (optional): https://mail.google.com
Tags (comma-separated): work,google,primary
Notes (optional): Main work email, recovery: john.doe@personal.com

✓ Entry saved successfully (ID: a3f2c8e1-...)

$ gpasswd copy "Gmail Work"
[gpasswd] Master Password: ****
✓ Password copied to clipboard (will clear in 30 seconds)
```

---

## 配置文件设计

### 配置文件位置

```
~/.gpasswd/
├── config.yaml          # 用户配置
└── vault.db             # 加密数据库
```

### config.yaml 内容

```yaml
# 会话配置
session:
  timeout: 300           # 会话超时时间（秒），0 = 永不超时

# 剪贴板配置
clipboard:
  clear_timeout: 30      # 剪贴板清除时间（秒）

# 密码生成器默认配置
password_generator:
  length: 20             # 默认长度
  use_uppercase: true
  use_lowercase: true
  use_digits: true
  use_symbols: true
  exclude_ambiguous: false  # 排除易混淆字符（0/O, 1/l/I）

# 安全配置
security:
  failed_attempts_limit: 5     # 失败尝试限制
  lockout_duration: 30         # 锁定时间（秒）

# Argon2 参数（高级用户可调整）
argon2:
  time_cost: 3           # 迭代次数
  memory_cost: 65536     # 内存成本（KB），64MB
  parallelism: 4         # 并行度

# 显示配置
display:
  show_timestamps: true
  date_format: "2006-01-02 15:04"
```

---

## 用户流程

### 首次使用

```
1. 用户下载 gpasswd 二进制文件
2. 运行 `gpasswd init`
3. 设置主密码（要求：12+ 字符，包含大小写数字符号）
4. 系统生成 salt 和 Argon2 参数
5. 创建空数据库
6. 初始化配置文件
7. 提示用户备份 ~/.gpasswd/ 目录
```

### 日常使用

```
1. 添加新账号：`gpasswd add`（交互式输入）
2. 需要密码时：`gpasswd copy "Gmail Work"`
3. 密码自动复制到剪贴板，30 秒后清除
4. 会话保持 5 分钟，期间不需重新输入主密码
```

### 主密码管理

```
- 如果忘记主密码：无法恢复（设计如此）
- 建议：主密码写在纸上，存放安全地点
- 未来版本：可选的密钥文件作为第二因素
```

---

## 安全威胁模型

### 威胁场景 1：攻击者获得数据库文件

**防御**：
- ✅ 数据使用 AES-256-GCM 加密
- ✅ 密钥由 Argon2id 派生（暴力破解成本高）
- ✅ 主密码永不存储
- ⚠️ 如果主密码弱，存在暴力破解风险

**建议**：强制主密码复杂度要求

### 威胁场景 2：键盘记录器

**防御**：
- ❌ 工具层面无法防御
- 建议：依赖系统安全（FileVault、防病毒软件）

### 威胁场景 3：内存转储

**防御**：
- ✅ 会话超时自动清除密钥
- ✅ 使用 `defer` 确保退出时清理
- ⚠️ 仍存在窗口期风险

### 威胁场景 4：剪贴板监控

**防御**：
- ✅ 自动清除剪贴板（30 秒）
- ⚠️ 窗口期内仍可能被读取
- 建议：敏感环境使用 `gpasswd reveal`（终端显示）

### 威胁场景 5：数据库损坏

**防御**：
- ✅ SQLite 事务保证一致性
- 建议：用户定期备份 `~/.gpasswd/` 目录
- 未来：自动备份功能

---

## 实现计划

### Phase 1: 基础架构（Week 1）

- [ ] 项目初始化
- [ ] 加密模块实现（Argon2 + AES-GCM）
- [ ] 数据库模块实现（SQLite schema + CRUD）
- [ ] 单元测试

### Phase 2: CLI 核心命令（Week 2）

- [ ] init 命令
- [ ] add 命令（交互式）
- [ ] list 命令
- [ ] show 命令
- [ ] delete 命令

### Phase 3: 密码访问与生成（Week 3）

- [ ] copy 命令（含剪贴板清除）
- [ ] reveal 命令
- [ ] generate 命令
- [ ] 会话管理实现

### Phase 4: 搜索与完善（Week 4）

- [ ] search 命令
- [ ] edit 命令
- [ ] 配置文件支持
- [ ] 错误处理优化
- [ ] 文档和 README

### Phase 5: 测试与发布

- [ ] 集成测试
- [ ] 安全审查
- [ ] 性能测试（1000+ 条目）
- [ ] 打包和分发

---

## 成功指标

MVP 被认为成功的标准：

1. ✅ 能安全存储至少 100 个密码条目
2. ✅ 从输入命令到复制密码到剪贴板 < 2 秒
3. ✅ 主密码验证失败 5 次后自动锁定
4. ✅ 剪贴板自动清除功能工作正常
5. ✅ 代码覆盖率 > 70%
6. ✅ 通过基础安全审查（无硬编码密钥、无明文存储）

---

## 未来迭代方向（Post-MVP）

### v0.2 - 增强安全
- 可选的密钥文件（类似 1Password Secret Key）
- 密码强度评估
- 审计日志

### v0.3 - 便利性
- 导入/导出（1Password, LastPass, CSV）
- 快速模糊搜索（fzf 集成）
- 密码历史记录

### v0.4 - 高级功能
- TOTP 生成器
- SSH 密钥管理
- 附件存储（加密文件）

### v0.5 - 同步（可选）
- Git-based 同步（加密 repo）
- WebDAV 支持
- 端到端加密云备份

---

## 开源与许可

- **许可证**：MIT License
- **仓库**：GitHub（公开）
- **贡献**：欢迎 PR 和 Issue
- **安全披露**：SECURITY.md 说明漏洞报告流程

---

## 参考资料

1. [1Password Security Design](https://1password.com/files/1Password-White-Paper.pdf)
2. [Argon2 RFC 9106](https://datatracker.ietf.org/doc/html/rfc9106)
3. [AES-GCM NIST SP 800-38D](https://csrc.nist.gov/publications/detail/sp/800-38d/final)
4. [SQLite Encryption Extension](https://www.sqlite.org/see/doc/trunk/www/readme.wiki)
5. [OWASP Password Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)

---

**文档版本**：v1.0
**最后更新**：2026-01-13
**作者**：gpasswd 项目组
