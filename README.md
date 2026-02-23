# 🌐 Clash 订阅聚合代理桥

> 一个轻量级的 Clash 订阅聚合管理服务，将多个代理订阅聚合为统一的代理入口，通过 RESTful API 实现节点的集中管理与自由切换。

```
程序/脚本  →  proxy.example.com:7890  →  Clash 内核  →  代理节点  →  目标网站
                                            ↑
                              管理 API :8080 (增删订阅/切换节点/测速)
```

## ✨ 特性

- 🔗 **多订阅聚合** — 支持添加多个 Clash 订阅链接，自动拉取解析合并
- 🎯 **节点过滤** — 自动过滤只保留港/新/台/日等低延迟地区节点
- 🔀 **自由切换** — 通过 API 随时切换代理节点，无需重启
- 📡 **HTTP 代理** — 对外暴露标准 HTTP/SOCKS5 代理端口，任何程序直接可用
- 🛡️ **Token 认证** — API 访问需要 Bearer Token，防止未授权使用
- ⚡ **自动管理** — mihomo 进程自动启动/重启，订阅变更后自动重载配置
- 🐳 **Docker 支持** — 提供 Dockerfile 和 docker-compose，一键部署

## 📦 架构

```
┌──────────────────────────────────────────┐
│            Management API (Go)           │
│                                          │
│  POST   /api/subscriptions        添加订阅│
│  GET    /api/subscriptions        列出订阅│
│  DELETE /api/subscriptions/:id    删除订阅│
│  POST   /api/subscriptions/refresh 刷新  │
│  GET    /api/proxies              列出节点│
│  PUT    /api/proxies/:group/:name 切换   │
│  GET    /api/proxies/:name/delay  测延迟 │
│  GET    /api/status               状态   │
├──────────────────────────────────────────┤
│         mihomo (Clash.Meta) 内核          │
│                                          │
│  :7890  HTTP/SOCKS5 混合代理端口          │
│  :7891  SOCKS5 代理端口                   │
│  :9090  External Controller              │
└──────────────────────────────────────────┘
```

## 🚀 快速开始

### 环境要求

- Go 1.22+
- [mihomo](https://github.com/MetaCubeX/mihomo) (Clash.Meta)

### 安装 mihomo

```bash
# Linux amd64
wget https://github.com/MetaCubeX/mihomo/releases/download/v1.19.0/mihomo-linux-amd64-v1.19.0.gz
gunzip mihomo-linux-amd64-v1.19.0.gz
chmod +x mihomo-linux-amd64-v1.19.0
mv mihomo-linux-amd64-v1.19.0 /usr/local/bin/mihomo
```

### 编译运行

```bash
git clone https://github.com/LordVibeCoding/clash-sub-aggregator.git
cd clash-sub-aggregator

# 编辑配置
cp configs/app.yaml configs/app.yaml
vim configs/app.yaml  # 修改 token

# 编译运行
make run
```

### Docker 部署

```bash
docker compose up -d
```

## ⚙️ 配置

```yaml
server:
  port: 8080                    # 管理 API 端口
  token: "your-secure-token"    # API 认证 Token

mihomo:
  binary: "/usr/local/bin/mihomo"
  config_dir: "./data"
  http_port: 7890               # HTTP/SOCKS5 代理端口
  socks_port: 7891              # SOCKS5 代理端口
  controller_port: 9090
  controller_secret: ""

data_dir: "./data"
```

## 📖 API 使用

所有请求需携带认证头：`Authorization: Bearer <token>`

### 添加订阅

```bash
curl -X POST http://your-server:8080/api/subscriptions \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{"name": "机场A", "url": "https://example.com/subscribe?token=xxx"}'
```

```json
{"id": "2776bc7a", "message": "订阅添加成功", "proxy_count": 39}
```

### 查看订阅列表

```bash
curl http://your-server:8080/api/subscriptions \
  -H "Authorization: Bearer your-token"
```

### 列出所有代理节点

```bash
curl http://your-server:8080/api/proxies \
  -H "Authorization: Bearer your-token"
```

### 切换节点

```bash
# 将 PROXY 组切换到指定节点（节点名需 URL 编码）
curl -X PUT "http://your-server:8080/api/proxies/PROXY/%F0%9F%87%AD%F0%9F%87%B0HKG%2001" \
  -H "Authorization: Bearer your-token"
```

```json
{"message": "已切换到 🇭🇰HKG 01"}
```

### 测试节点延迟

```bash
curl "http://your-server:8080/api/proxies/%F0%9F%87%AD%F0%9F%87%B0HKG%2001/delay" \
  -H "Authorization: Bearer your-token"
```

### 刷新订阅

```bash
# 刷新所有订阅
curl -X POST http://your-server:8080/api/subscriptions/refresh \
  -H "Authorization: Bearer your-token"

# 刷新单个订阅
curl -X POST http://your-server:8080/api/subscriptions/订阅ID/refresh \
  -H "Authorization: Bearer your-token"
```

### 删除订阅

```bash
curl -X DELETE http://your-server:8080/api/subscriptions/订阅ID \
  -H "Authorization: Bearer your-token"
```

## 🖥️ 客户端使用

设置好代理后，所有流量将通过选定的代理节点转发：

```bash
# 环境变量方式（适用于大多数 CLI 工具）
export http_proxy=http://your-server:7890
export https_proxy=http://your-server:7890

# Git
git config --global http.proxy http://your-server:7890

# Python requests
import requests
proxies = {"http": "http://your-server:7890", "https": "http://your-server:7890"}
requests.get("https://example.com", proxies=proxies)

# Node.js (with undici or global-agent)
HTTP_PROXY=http://your-server:7890 node app.js

# curl 临时使用
curl -x http://your-server:7890 https://api.example.com
```

## 🌏 节点过滤

默认只保留以下地区的节点，减少干扰：

| 地区 | 匹配关键词 |
|------|-----------|
| 🇭🇰 香港 | HK, HKG, 香港 |
| 🇸🇬 新加坡 | SG, SGP, 新加坡 |
| 🇹🇼 台湾 | TW, TWN, 台湾 |
| 🇯🇵 日本 | JP, JPN, 日本 |

如需修改过滤规则，编辑 `internal/subscription/parser.go` 中的 `allowedRegions` 变量。

## 📁 项目结构

```
clash-sub-aggregator/
├── main.go                          # 入口
├── internal/
│   ├── api/
│   │   ├── router.go                # 路由定义
│   │   ├── middleware.go            # Token 认证中间件
│   │   ├── subscription_handler.go  # 订阅管理接口
│   │   └── proxy_handler.go         # 代理控制接口
│   ├── clash/
│   │   ├── config.go                # mihomo 配置生成
│   │   └── process.go               # mihomo 进程管理
│   ├── subscription/
│   │   ├── manager.go               # 订阅管理逻辑
│   │   └── parser.go                # 订阅解析 + 节点过滤
│   ├── model/
│   │   └── model.go                 # 数据模型
│   └── store/
│       └── store.go                 # JSON 文件持久化
├── configs/
│   └── app.yaml                     # 应用配置
├── data/                            # 运行时数据
├── Dockerfile
├── docker-compose.yaml
└── Makefile
```

## 🛡️ 安全建议

- 修改默认 Token，使用强随机字符串
- 通过防火墙限制管理 API（8080）的访问来源
- 代理端口（7890）按需开放
- 定期刷新订阅保持节点可用性

## 📄 License

MIT
