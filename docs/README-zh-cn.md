<p align="center">
 <a href="../README.md">English</a> /
 <a href="./README-fa.md">فارسی</a> /
 <a href="./README-ru.md">Русский</a> /
 <a href="./README-zh-cn.md">简体中文</a>
</p>

<p align="center">
  <img src="./assets/gamaj-mark.svg" alt="Gamaj" width="104" height="104">
</p>

<h1>GAMAJ</h1>

### Panel

**Gamaj（گمج）** 是一个自托管的 Xray 代理账号管理平台：网页面板、完整的 REST
API、命令行工具以及多节点分发。全部功能运行在一个原生二进制文件和 systemd
服务之上。

当前版本：**`is.0.0.1`** · 默认端口：**616**

<p align="center">
  <a href="https://t.me/TheGamaj">Telegram 频道</a> ·
  <a href="https://t.me/GamajGP">支持群组</a> ·
  <a href="https://t.me/AsliCode">代码频道</a>
</p>

## 功能

- **网页面板**（React + Chakra UI），明暗主题，四种语言：英语、فارسی、俄语、简体中文
- **完整的 REST API**，并支持为机器人和第三方集成创建 **API 密钥**
- **多节点**：基于 gRPC 与双向证书认证，节点证书自动签发
- 协议：**VLESS**、**VMess**、**Trojan**、**Shadowsocks**；节点还支持 OpenVPN、WireGuard、L2TP、PPTP、SSTP 与 Remote Access
- **单个入站多用户**、单个用户多入站、单端口 fallback
- **流量、到期时间和并发连接限制**，支持周期流量重置（每日/每周/…）
- **订阅链接**：兼容 v2ray 客户端、Clash 与 sing-box，自动生成分享链接和二维码
- **Telegram 集成**：报告、备份和用户通知
- **备份与恢复**（数据库与完整归档），可按计划发送到 Telegram
- **系统监控**：CPU、内存、磁盘、在线用户、各节点流量、Xray 日志
- **多管理员**：`full_access`、`sudo`、`standard`、`seller` 角色，两步验证、独立限额与 API 密钥
- **外部应用托管**：在同一台服务器上安装和管理第三方面板类应用
- **HAProxy 与主机管理**：主机模板、地址、证书以及节点上的主机操作
- **完全自有品牌**：一个名称、一个端口，没有容器层

## 环境要求

- 带 systemd 的 Linux 服务器（Debian/Ubuntu、CentOS/Rocky/Alma、Fedora、Arch、Alpine）与 root 权限
- 架构：`linux-386`、`linux-amd64`、`linux-arm64`、`linux-armv5`、`linux-armv6`、`linux-armv7`、`linux-s390x`
- 数据库：SQLite（默认）、MySQL 或 MariaDB

## 安装

```bash
curl -fsSL https://raw.githubusercontent.com/TheGamaj/Panel/Asli/scripts/gamaj/gamaj.sh | sudo bash -s -- install
```

指定版本：

```bash
curl -fsSL https://raw.githubusercontent.com/TheGamaj/Panel/Asli/scripts/gamaj/gamaj.sh | sudo bash -s -- install --version is.0.0.1
```

选择数据库：

```bash
... | sudo bash -s -- install --database sqlite
... | sudo bash -s -- install --database mysql
... | sudo bash -s -- install --database mariadb
```

不要以 `sudo bash -c "$(curl ...)"` 方式运行安装脚本：脚本正文可能超过 Linux
的参数长度限制。请始终把 curl 的输出通过管道交给 bash。

Gamaj 只以原生二进制加 systemd 服务的方式安装：没有容器安装，也没有安装模式
开关，一条安装路径即为全部功能。

## 安装之后

| 项目 | 路径 |
|---|---|
| 面板文件 | `/opt/gamaj` |
| 配置文件 | `/opt/gamaj/.env` |
| 数据与数据库 | `/var/lib/gamaj` |
| 服务 | `gamaj.service` |

创建第一个管理员：

```bash
sudo gamaj cli admin create --role full_access
```

打开面板：

- 使用域名与 TLS：`https://YOUR_DOMAIN:616/dashboard/`
- 本地测试可用 SSH 隧道：`ssh -L 616:localhost:616 user@serverip`，然后访问 `http://localhost:616/dashboard/`

包含 TLS 与防火墙说明的完整步骤见 [INSTALL.md](INSTALL.md)。

## 管理

```bash
sudo gamaj status          # 服务状态
sudo gamaj restart         # 重启
sudo gamaj logs            # 查看日志
sudo gamaj backup          # 备份
sudo gamaj update          # 更新
sudo gamaj core-update     # 更新 Xray-core
sudo gamaj edit-env        # 编辑 /opt/gamaj/.env
sudo gamaj ssl             # 管理证书
sudo gamaj uninstall       # 卸载
```

## 配置

所有设置位于 `/opt/gamaj/.env`，所有键名都以 `GAMAJ_` 开头。完整列表见
[`.env.example`](../.env.example)。常用项：

| 键 | 含义 | 默认值 |
|---|---|---|
| `GAMAJ_DATABASE_URL` | 数据库连接（必填） | — |
| `GAMAJ_HOST` | 监听地址 | `0.0.0.0` |
| `GAMAJ_PORT` | 面板端口 | `616` |
| `GAMAJ_DATA_DIR` | 数据目录 | `/var/lib/gamaj` |
| `GAMAJ_XRAY_JSON` | Xray 配置文件 | `/var/lib/gamaj/xray_config.json` |
| `GAMAJ_USER_AUTODELETE_DAYS` | 过期用户保留天数 | `0`（关闭） |
| `GAMAJ_JWT_ACCESS_TOKEN_EXPIRE_MINUTES` | 面板会话时长 | `1440` |
| `GAMAJ_WEBHOOK_ADDRESS` | 通知 webhook 地址 | — |

## 添加节点

在另一台服务器上：

```bash
curl -fsSL https://raw.githubusercontent.com/TheGamaj/Panel/Asli/scripts/gamaj/gamaj-node.sh | sudo bash -s -- install
```

然后在面板中打开 **Nodes → 添加节点**。面板会签发证书与配置，节点会自动上报
流量、在线用户并执行主机操作。

## 安装机器人

Telegram 机器人位于独立的仓库：[TheGamaj/Bot](https://github.com/TheGamaj/Bot)，只通过面板 API 工作。所有支付设置都保存在面板中，机器人仅通过 API 使用它们。

推荐方式：在面板打开 **Applications** 并安装 **Gamaj Bot**——面板会自动下载二进制文件、创建 `gm_...` API 密钥、写入配置并注册 Webhook。

手动安装：

```bash
git clone https://github.com/TheGamaj/Bot.git Gamaj-Bot
cd Gamaj-Bot
sudo ./scripts/install.sh    # installs Go and builds the binary automatically
sudo nano /opt/gamaj-bot/.gamaj-bot.json   # panel_url、api_key、bot_token、admin_id
sudo systemctl restart gamaj-bot
```

## 命令行

`gamaj cli` 可从终端管理用户、管理员、入站和数据库，例如 `gamaj cli user list`、
`gamaj cli admin set-password`、`gamaj cli migrate up`。完整参考见
[cli/README.md](cli/README.md)。

## 开发

```bash
cd dashboard && npm ci && VITE_BASE_API=/api/ npm run build && cd ..
bash scripts/build_binary.sh        # dist/gamaj-server 与 dist/gamaj-cli

go build ./... && go vet ./internal/... ./cmd/...
go test ./internal/app/system/ ./internal/app/api/ ./internal/gateway/ \
        ./internal/app/backup/ ./internal/app/migrations/ ./cmd/... \
        ./internal/platform/envguard/

sudo bash scripts/tests/e2e-binary-install.sh   # 端到端安装检查
```

如果代码树中出现不带 `GAMAJ_` 前缀的环境变量名（或旧版遗留名称），
`go test ./internal/platform/envguard/` 会让构建失败。

## 文档

- [INSTALL.md](INSTALL.md) — 安装与升级
- [cli/README.md](cli/README.md) — 命令行参考
- [dashboard/README.md](dashboard/README.md) — 面板说明
- [scripts/gamaj/README.md](scripts/gamaj/README.md) — 安装脚本
- [CONTRIBUTING.md](CONTRIBUTING.md) — 贡献规范
- [tutorials/](../tutorials/) — 面板内置的分步教程

## 许可证

GNU Affero General Public License v3.0，见 [LICENSE](../LICENSE)。
