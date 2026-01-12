# 配置文件说明文档

本文档详细说明 gpasswd 项目中各个配置文件的用途和最佳实践。

---

## 📋 配置文件清单

### Git 相关配置

#### 1. `.gitignore` ✅ 已完善
**用途**: 指定 Git 应该忽略的文件和目录

**亮点**：
- ✅ **450+ 行**完整的忽略规则
- ✅ **分类清晰**：按功能分 15 个大类
- ✅ **安全优先**：重点保护敏感数据
- ✅ **跨平台**：覆盖 macOS/Linux/Windows
- ✅ **详细注释**：每个规则都有说明

**关键保护**：
```gitignore
# 用户数据（绝对不能提交）
.gpasswd/
vault.db
*.key
config.yaml
.env

# 密钥文件
*.pem
*.p12
*.pfx
master.key
```

**白名单**（明确保留）：
```gitignore
!.gitignore
!.env.example
!config.example.yaml
!go.mod
!go.sum
```

---

#### 2. `.gitattributes` ✅ 新增
**用途**: 定义文件属性，确保跨平台一致性

**功能**：
- ✅ **行尾规范化**：自动转换为 LF（Unix 风格）
- ✅ **二进制文件标记**：防止错误合并
- ✅ **语言统计**：GitHub 语言识别
- ✅ **导出控制**：`git archive` 排除测试文件
- ✅ **Diff 优化**：为不同文件类型设置最佳 diff 策略

**示例**：
```gitattributes
# 所有文本文件使用 LF
*.go text eol=lf
*.md text eol=lf

# 敏感文件作为二进制（防止 diff 泄露）
*.env binary
vault.db binary

# 不包含在归档中
*_test.go export-ignore
docs/ export-ignore
```

---

### 编辑器配置

#### 3. `.editorconfig` ✅ 新增
**用途**: 统一不同编辑器的代码风格

**支持的编辑器**：
- VS Code
- IntelliJ IDEA / GoLand
- Vim / Neovim
- Sublime Text
- Atom
- 等 [支持 EditorConfig 的所有编辑器](https://editorconfig.org/#pre-installed)

**配置要点**：
```editorconfig
# Go 使用 Tab（Go 官方规范）
[*.go]
indent_style = tab
indent_size = 4

# YAML 使用 2 空格
[*.{yaml,yml}]
indent_style = space
indent_size = 2

# Makefile 必须使用 Tab
[Makefile]
indent_style = tab
```

**好处**：
- ✅ 团队协作时代码风格一致
- ✅ 自动配置，无需手动设置
- ✅ 适配 Go 官方规范（gofmt 使用 tab）

---

### 代码质量工具

#### 4. `.golangci.yml` ✅ 新增
**用途**: golangci-lint 代码检查工具的配置

**启用的检查器**（25+ 个）：

**Bug 和性能**：
- `errcheck` - 检查未处理的错误
- `gosimple` - 简化代码建议
- `govet` - Go 官方 vet 工具
- `ineffassign` - 检测无效赋值
- `staticcheck` - 静态分析
- `unused` - 未使用的代码

**安全检查**：
- `gosec` - 安全漏洞检测
  - 弱加密算法（MD5, SHA1, DES, RC4）
  - TLS 配置错误
  - 弱随机数
  - SQL 注入风险

**代码风格**：
- `gofmt` - 格式检查
- `goimports` - 导入排序
- `revive` - 快速可配置的 linter
- `gocritic` - 代码质量诊断

**复杂度控制**：
- `gocyclo` - 圈复杂度（上限 15）
- `dupl` - 代码重复检测

**使用方式**：
```bash
# 手动运行
make lint

# 或直接使用
golangci-lint run

# 自动修复（如果支持）
golangci-lint run --fix
```

**特别配置**：
```yaml
# 允许测试文件中的重复代码
- path: _test\.go
  linters:
    - dupl
    - goconst

# 允许 config.go 中使用用户路径（G304 警告）
- linters:
    - gosec
  text: "G304:"
  path: "pkg/config/config.go"
```

---

### 示例配置文件

#### 5. `config.example.yaml` ✅ 新增
**用途**: 用户配置文件模板

**包含的配置**：
- ✅ 会话超时设置
- ✅ 剪贴板清除时间
- ✅ 密码生成器默认参数
- ✅ 安全策略（失败尝试限制）
- ✅ Argon2id 密钥派生参数
- ✅ 显示偏好设置
- ✅ 未来功能的占位符（备份、审计、搜索）

**使用方式**：
```bash
# 复制到用户目录
cp config.example.yaml ~/.gpasswd/config.yaml

# 编辑自定义设置
vim ~/.gpasswd/config.yaml
```

**重要提示**：
```yaml
# ⚠️ 警告：修改 Argon2 参数会导致现有保管库无法访问！
argon2:
  time_cost: 3        # 不要随意修改
  memory_cost: 65536  # 不要随意修改
  parallelism: 4      # 不要随意修改
```

---

#### 6. `.env.example` ✅ 新增
**用途**: 环境变量模板（用于开发和测试）

**关键环境变量**：

**开发模式**：
```bash
GPASSWD_DEBUG=true
GPASSWD_TEST_MODE=true
GPASSWD_LOG_LEVEL=debug
```

**自定义路径**：
```bash
GPASSWD_VAULT_DIR=/custom/path
GPASSWD_CONFIG_FILE=/custom/config.yaml
```

**测试用参数**（仅测试环境）：
```bash
# ⚠️ 警告：永远不要在生产环境使用！
GPASSWD_ARGON2_TIME=1
GPASSWD_ARGON2_MEMORY=8192
GPASSWD_SKIP_PASSWORD_VALIDATION=false
```

**使用方式**：
```bash
# 开发环境
cp .env.example .env
source .env
make run

# 或直接在命令行设置
GPASSWD_DEBUG=true make run
```

**安全提示**：
- ❌ **永远不要提交 `.env` 文件**（已在 `.gitignore` 中）
- ✅ 仅提交 `.env.example` 作为模板
- ✅ 生产环境不要使用测试参数

---

## 🔧 配置文件最佳实践

### 开发阶段

1. **首次设置**：
   ```bash
   # 安装开发工具
   make install-deps

   # 复制配置示例
   cp .env.example .env
   cp config.example.yaml ~/.gpasswd/config.yaml

   # 编辑器会自动读取 .editorconfig
   ```

2. **代码提交前**：
   ```bash
   # 格式化代码
   make fmt

   # 运行 linter
   make lint

   # 运行测试
   make test

   # 或一键检查
   make check
   ```

3. **修改配置文件**：
   ```bash
   # 修改后验证配置文件
   golangci-lint run --config .golangci.yml

   # 测试 .gitignore
   git status --ignored

   # 测试 .gitattributes
   git check-attr -a <file>
   ```

---

### 团队协作

1. **新成员入职**：
   - 克隆仓库后，编辑器自动读取 `.editorconfig`
   - 复制 `.env.example` 到 `.env` 并自定义
   - 运行 `make install-deps` 安装 golangci-lint

2. **代码审查**：
   - CI/CD 自动运行 `make check`
   - 要求所有 PR 通过 golangci-lint 检查
   - 覆盖率不低于当前水平

3. **配置变更**：
   - 修改 `.golangci.yml` 需团队讨论
   - 更新 `.gitignore` 时注意不要误删重要文件
   - `.editorconfig` 变更需全员重启编辑器

---

## 📊 配置文件对比表

| 文件 | 用途 | 影响范围 | 是否提交 | 优先级 |
|------|------|---------|---------|--------|
| `.gitignore` | Git 忽略规则 | 本地 + 远程 | ✅ 是 | 最高 |
| `.gitattributes` | 文件属性 | 本地 + 远程 | ✅ 是 | 高 |
| `.editorconfig` | 编辑器配置 | 本地 | ✅ 是 | 中 |
| `.golangci.yml` | 代码检查 | 本地 + CI | ✅ 是 | 高 |
| `config.example.yaml` | 配置模板 | 文档 | ✅ 是 | 中 |
| `.env.example` | 环境变量模板 | 文档 | ✅ 是 | 中 |
| `config.yaml` | 实际配置 | 本地用户 | ❌ 否 | - |
| `.env` | 实际环境变量 | 本地开发 | ❌ 否 | - |

---

## 🛡️ 安全检查清单

### 提交前必查

- [ ] `.env` 文件没有被提交
- [ ] `config.yaml` 没有被提交
- [ ] `vault.db` 没有被提交
- [ ] 没有密钥文件（`.key`, `.pem`）
- [ ] 没有包含测试密码的代码
- [ ] `make lint` 通过所有检查
- [ ] `make test` 所有测试通过

### 配置文件安全

```bash
# 检查敏感文件是否被正确忽略
git check-ignore -v .env
git check-ignore -v vault.db
git check-ignore -v config.yaml

# 应该输出对应的 .gitignore 规则
# 如果没有输出，说明文件可能会被提交！
```

### Git 历史检查

```bash
# 检查历史提交中是否包含敏感文件
git log --all --full-history --source -- vault.db
git log --all --full-history --source -- .env
git log --all --full-history --source -- "*.key"

# 应该没有任何输出
```

---

## 🔄 配置文件更新流程

### 更新 `.gitignore`

```bash
# 1. 修改 .gitignore
vim .gitignore

# 2. 测试是否生效
git status --ignored

# 3. 清除 Git 缓存（如果文件已被跟踪）
git rm --cached <file>

# 4. 提交更改
git add .gitignore
git commit -m "chore: update .gitignore"
```

### 更新 `.golangci.yml`

```bash
# 1. 修改配置
vim .golangci.yml

# 2. 测试新配置
golangci-lint run --config .golangci.yml

# 3. 如果有大量新错误，考虑分批修复
golangci-lint run --new-from-rev=HEAD~1

# 4. 提交
git add .golangci.yml
git commit -m "chore: update linter configuration"
```

### 更新示例配置

```bash
# 1. 修改示例文件
vim config.example.yaml

# 2. 验证 YAML 格式
yamllint config.example.yaml

# 3. 更新文档
vim docs/PROJECT_OVERVIEW.md

# 4. 提交
git add config.example.yaml docs/
git commit -m "docs: update configuration examples"
```

---

## 📚 相关资源

### 官方文档
- [Git Attributes](https://git-scm.com/docs/gitattributes)
- [EditorConfig](https://editorconfig.org/)
- [golangci-lint](https://golangci-lint.run/)

### 最佳实践
- [GitHub .gitignore 模板](https://github.com/github/gitignore)
- [Go 项目布局](https://github.com/golang-standards/project-layout)
- [Effective Go](https://go.dev/doc/effective_go)

### 工具
- [gitignore.io](https://www.toptal.com/developers/gitignore) - 生成 .gitignore
- [EditorConfig 插件](https://editorconfig.org/#download) - 各编辑器插件
- [golangci-lint 安装](https://golangci-lint.run/usage/install/)

---

## 🎯 快速参考

### 常用命令

```bash
# 检查被忽略的文件
git status --ignored

# 检查文件属性
git check-attr -a <file>

# 运行 linter
make lint

# 格式化代码
make fmt

# 完整检查
make check
```

### 配置文件位置

```
gpasswd/
├── .editorconfig           # 编辑器配置
├── .env.example            # 环境变量示例
├── .gitattributes          # Git 属性
├── .gitignore              # Git 忽略规则
├── .golangci.yml           # Linter 配置
└── config.example.yaml     # 应用配置示例

用户目录/
├── ~/.gpasswd/
│   ├── config.yaml         # 实际配置（不提交）
│   └── vault.db            # 数据库（不提交）
└── .env                    # 实际环境变量（不提交）
```

---

**最后更新**: 2026-01-13
**维护者**: gpasswd 项目组
