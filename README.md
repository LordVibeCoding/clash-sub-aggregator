# 🌐 Clash 订阅聚合代理桥

> 将多个 Clash 代理订阅聚合为统一的代理入口，通过 RESTful API 集中管理节点、自由切换线路。适合需要在多台设备/服务器上共享代理的场景。

```
                          ┌─────────────────────┐
程序/脚本/Proxifier  ───→  │  VPS :7890 (代理)    │ ───→ 代理节点 ───→ 目标网站
                          │  mihomo 内核         │
管理端 (curl/脚本)   ───→  │  VPS :8080 (API)     │
                          └─────────────────────┘
```

## ✨ 特性

- 🔗 **多订阅聚合** — 添加多个 Clash 订阅链接，自动拉取解析合并
- 🎯 **地区过滤** — 只保留港/新/台/日等低延迟地区节点，自动过滤无关节点
- 🔀 **API 切换** — 通过 API 随时切换代理节点，无需重启服务
- 📡 **标准代理** — HTTP/SOCKS5 代理端口，任何支持代理的程序直接可用
- 🛡️ **Token 认证** — Bearer Token 保护管理 API
- ⚡ **自动重载** — 订阅变更后自动重新生成配置并重启 mihomo
- 🔄 **进程守护** — mihomo 异常退出自动检测，支持热重启

## 🚀 快速开始

### 环境要求

- Go 1.22+
- [mihomo](https://github.com/MetaCubeX/mihomo) v1.19.14+（需要 anytls 支持则 v1.19.14+）

### 安装 mihomo

```bash
# Linux amd64（推荐 v1.19.20+）
wget https://github.com/MetaCubeX/mihomo/releases/download/v1.19.20/mihomo-linux-amd64-v1.19.20.gz
gunzip mihomo-linux-amd64-v1.19.20.gz
chmod +x mihomo-linux-amd64-v1.19.20
mv mihomo-linux-amd64-v1.19.20 /usr/local/bin/mihomo
```

### 编译运行

```bash
git clone https://github.com/LordVibeCoding/clash-sub-aggregator.git
cd clash-sub-aggregator

# 修改配置（必须修改 token）
vim configs/app.yaml

# 编译运行
make run
```

### Docker 部署

```bash
docker compose up -d
```

### Systemd 部署（推荐）

```bash
make build

cat > /etc/systemd/system/clash-aggregator.service << EOF
[Unit]
Description=Clash Subscription Aggregator
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/clash-aggregator
ExecStart=/opt/clash-aggregator/bin/clash-sub-aggregator
Restart=always
RestartSec=5
Environment=CONFIG_PATH=configs/app.yaml

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now clash-aggregator
```

## ⚙️ 配置

```yaml
server:
  port: 8080                    # 管理 API 端口
  token: "your-secure-token"    # API 认证 Token（必须修改）

mihomo:
  binary: "/usr/local/bin/mihomo"
  config_dir: "./data"
  http_port: 7890               # HTTP/SOCKS5 混合代理端口
  socks_port: 7891              # SOCKS5 代理端口
  controller_port: 9090         # mihomo External Controller
  controller_secret: ""

data_dir: "./data"
```

## 📖 API

所有请求需携带：`Authorization: Bearer <token>`

### 订阅管理

```bash
# 添加订阅（自动拉取解析，成功后自动启动/重启 mihomo）
curl -X POST http://your-server:8080/api/subscriptions \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{"name": "机场A", "url": "https://example.com/subscribe?token=xxx"}'
# → {"id": "2776bc7a", "message": "订阅添加成功", "proxy_count": 39}

# 查看所有订阅
curl http://your-server:8080/api/subscriptions \
  -H "Authorization: Bearer your-token"

# 刷新所有订阅（重新拉取节点）
curl -X POST http://your-server:8080/api/subscriptions/refresh \
  -H "Authorization: Bearer your-token"

# 刷新单个订阅
curl -X POST http://your-server:8080/api/subscriptions/{id}/refresh \
  -H "Authorization: Bearer your-token"

# 删除订阅
curl -X DELETE http://your-server:8080/api/subscriptions/{id} \
  -H "Authorization: Bearer your-token"
```

### 代理控制

```bash
# 列出所有代理节点
curl http://your-server:8080/api/proxies \
  -H "Authorization: Bearer your-token"

# 切换节点（节点名需 URL 编码）
curl -X PUT "http://your-server:8080/api/proxies/PROXY/{节点名}" \
  -H "Authorization: Bearer your-token"
# → {"message": "已切换到 🇭🇰HKG 01"}

# 测试节点延迟
curl "http://your-server:8080/api/proxies/{节点名}/delay" \
  -H "Authorization: Bearer your-token"

# 服务状态
curl http://your-server:8080/api/status \
  -H "Authorization: Bearer your-token"
```

## 🖥️ 客户端接入

```bash
# 环境变量（适用于大多数 CLI 工具、脚本）
export http_proxy=http://your-server:7890
export https_proxy=http://your-server:7890

# Git
git config --global http.proxy http://your-server:7890

# Python
import requests
proxies = {"http": "http://your-server:7890", "https": "http://your-server:7890"}
requests.get("https://example.com", proxies=proxies)

# Node.js
HTTP_PROXY=http://your-server:7890 node app.js

# curl
curl -x http://your-server:7890 https://api.example.com

# Proxifier / 系统代理
# 类型: HTTP，地址: your-server，端口: 7890
```

## 🌏 节点过滤

默认只保留以下地区的节点：

| 地区 | 匹配关键词 |
|------|-----------|
| 🇭🇰 香港 | HK, HKG, 香港 |
| 🇸🇬 新加坡 | SG, SGP, 新加坡 |
| 🇹🇼 台湾 | TW, TWN, 台湾 |
| 🇯🇵 日本 | JP, JPN, 日本 |

修改过滤规则：编辑 `internal/subscription/parser.go` 中的 `allowedRegions`。

## 🏗️ 推荐部署架构

如果代理订阅的入口节点只接受中国大陆 IP，推荐以下架构：

```
                    管理 API（无需备案）
用户  ──────────→  海外 VPS (nginx 反代) ──→ 国内 VPS :8080

                    代理流量
程序  ──────────→  国内 VPS :7890 ──→ mihomo ──→ 代理节点
```

- 国内 VPS 运行 mihomo + 管理服务（能连上代理节点入口）
- 海外 VPS 用 nginx 反代管理 API（域名指向海外 VPS，不需要 ICP 备案）
- 代理端口直连国内 VPS IP

## 📁 项目结构

```
clash-sub-aggregator/
├── main.go                          # 入口
├── internal/
│   ├── api/
│   │   ├── router.go                # 路由定义
│   │   ├── middleware.go            # Token 认证
│   │   ├── subscription_handler.go  # 订阅 CRUD
│   │   └── proxy_handler.go         # 代理控制（转发 mihomo API）
│   ├── clash/
│   │   ├── config.go                # mihomo 配置生成
│   │   └── process.go               # mihomo 进程管理
│   ├── subscription/
│   │   ├── manager.go               # 订阅拉取 + 管理
│   │   └── parser.go                # 订阅解析 + 地区过滤
│   ├── model/
│   │   └── model.go                 # 数据模型
│   └── store/
│       └── store.go                 # JSON 文件持久化
├── configs/
│   └── app.yaml                     # 应用配置
├── data/                            # 运行时数据（自动生成）
├── Dockerfile
├── docker-compose.yaml
└── Makefile
```

## 📄 License

MIT
