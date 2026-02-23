<div align="center">

# 🌐 Clash 订阅聚合代理桥

<p>
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go&logoColor=white" />
  <img src="https://img.shields.io/badge/mihomo-v1.19.20-6C5CE7?style=flat-square" />
  <img src="https://img.shields.io/badge/License-MIT-green?style=flat-square" />
  <img src="https://img.shields.io/github/stars/LordVibeCoding/clash-sub-aggregator?style=flat-square&color=yellow" />
</p>

<p><strong>将多个 Clash 代理订阅聚合为统一代理入口<br/>通过 RESTful API 集中管理节点、自由切换线路</strong></p>

<br/>

<table>
<tr>
<td align="center">🔗<br/><strong>多订阅聚合</strong></td>
<td align="center">🎯<br/><strong>地区过滤</strong></td>
<td align="center">🔀<br/><strong>API 切换</strong></td>
<td align="center">📡<br/><strong>标准代理</strong></td>
<td align="center">🛡️<br/><strong>Token 认证</strong></td>
<td align="center">🔄<br/><strong>进程守护</strong></td>
<td align="center">💊<br/><strong>健康检查</strong></td>
</tr>
</table>

</div>

---

## 工作原理

```
                          ┌─────────────────────┐
程序/脚本/Proxifier  ───→  │  VPS :7890 (代理)    │ ───→ 代理节点 ───→ 目标网站
                          │  mihomo 内核         │
管理端 (curl/脚本)   ───→  │  VPS :8080 (API)     │
                          └─────────────────────┘
```

<details>
<summary><strong>📡 推荐部署架构（代理节点限制中国 IP 时）</strong></summary>
<br/>

```
                    管理 API（无需备案）
用户  ──────────→  海外 VPS (nginx 反代) ──→ 国内 VPS :8080

                    代理流量
程序  ──────────→  国内 VPS :7890 ──→ mihomo ──→ 代理节点
```

- 国内 VPS 运行 mihomo + 管理服务（能连上代理节点入口）
- 海外 VPS 用 nginx 反代管理 API（域名指向海外 VPS，不需要 ICP 备案）
- 代理端口直连国内 VPS IP

</details>

---

## 🚀 快速开始

### 安装 mihomo

```bash
wget https://github.com/MetaCubeX/mihomo/releases/download/v1.19.20/mihomo-linux-amd64-v1.19.20.gz
gunzip mihomo-linux-amd64-v1.19.20.gz
chmod +x mihomo-linux-amd64-v1.19.20
mv mihomo-linux-amd64-v1.19.20 /usr/local/bin/mihomo
```

### 编译运行

```bash
git clone https://github.com/LordVibeCoding/clash-sub-aggregator.git
cd clash-sub-aggregator
vim configs/app.yaml  # 修改 token
make run
```

### Docker

```bash
docker compose up -d
```

<details>
<summary><strong>📋 Systemd 部署（推荐生产环境）</strong></summary>

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

</details>

---

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
  controller_port: 9090
  controller_secret: ""

data_dir: "./data"
```

---

## 📖 API

> 所有请求需携带 `Authorization: Bearer <token>`

<table>
<tr><th>方法</th><th>路径</th><th>说明</th></tr>
<tr><td><code>POST</code></td><td><code>/api/subscriptions</code></td><td>添加订阅</td></tr>
<tr><td><code>GET</code></td><td><code>/api/subscriptions</code></td><td>列出所有订阅</td></tr>
<tr><td><code>DELETE</code></td><td><code>/api/subscriptions/:id</code></td><td>删除订阅</td></tr>
<tr><td><code>POST</code></td><td><code>/api/subscriptions/refresh</code></td><td>刷新所有订阅</td></tr>
<tr><td><code>POST</code></td><td><code>/api/subscriptions/:id/refresh</code></td><td>刷新单个订阅</td></tr>
<tr><td><code>GET</code></td><td><code>/api/proxies</code></td><td>列出所有代理节点</td></tr>
<tr><td><code>PUT</code></td><td><code>/api/proxies/:group/:name</code></td><td>切换节点</td></tr>
<tr><td><code>GET</code></td><td><code>/api/proxies/:name/delay</code></td><td>测试节点延迟</td></tr>
<tr><td><code>GET</code></td><td><code>/api/status</code></td><td>服务状态</td></tr>
<tr><td><code>GET</code></td><td><code>/api/health</code></td><td>健康检查状态（黑名单、上次检查时间）</td></tr>
<tr><td><code>POST</code></td><td><code>/api/health/check</code></td><td>手动触发健康检查</td></tr>
</table>

<details>
<summary><strong>📝 API 使用示例</strong></summary>

### 订阅管理

```bash
# 添加订阅
curl -X POST http://your-server:8080/api/subscriptions \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{"name": "机场A", "url": "https://example.com/subscribe?token=xxx"}'
# → {"id": "2776bc7a", "message": "订阅添加成功", "proxy_count": 39}

# 查看所有订阅
curl http://your-server:8080/api/subscriptions \
  -H "Authorization: Bearer your-token"

# 刷新所有订阅
curl -X POST http://your-server:8080/api/subscriptions/refresh \
  -H "Authorization: Bearer your-token"

# 删除订阅
curl -X DELETE http://your-server:8080/api/subscriptions/{id} \
  -H "Authorization: Bearer your-token"
```

### 代理控制

```bash
# 列出所有节点
curl http://your-server:8080/api/proxies \
  -H "Authorization: Bearer your-token"

# 切换节点（节点名需 URL 编码）
curl -X PUT "http://your-server:8080/api/proxies/PROXY/{节点名}" \
  -H "Authorization: Bearer your-token"
# → {"message": "已切换到 🇭🇰HKG 01"}

# 测试延迟
curl "http://your-server:8080/api/proxies/{节点名}/delay" \
  -H "Authorization: Bearer your-token"
```

</details>

---

## 🖥️ 客户端接入

<table>
<tr><th>方式</th><th>配置</th></tr>
<tr>
<td><strong>环境变量</strong></td>
<td>

```bash
export http_proxy=http://your-server:7890
export https_proxy=http://your-server:7890
```

</td>
</tr>
<tr>
<td><strong>Git</strong></td>
<td>

```bash
git config --global http.proxy http://your-server:7890
```

</td>
</tr>
<tr>
<td><strong>Python</strong></td>
<td>

```python
import requests
proxies = {"http": "http://your-server:7890", "https": "http://your-server:7890"}
requests.get("https://example.com", proxies=proxies)
```

</td>
</tr>
<tr>
<td><strong>Node.js</strong></td>
<td>

```bash
HTTP_PROXY=http://your-server:7890 node app.js
```

</td>
</tr>
<tr>
<td><strong>curl</strong></td>
<td>

```bash
curl -x http://your-server:7890 https://api.example.com
```

</td>
</tr>
<tr>
<td><strong>Proxifier</strong></td>
<td>类型 HTTP，地址 your-server，端口 7890</td>
</tr>
</table>

---

## 💊 节点健康检查

服务内置自动健康检查机制，每 30 分钟对所有代理节点测速：

- 连接失败或超时（5s）的节点自动加入黑名单，从活跃配置中移除
- 下次检查时恢复的节点自动移出黑名单，重新加入配置
- 黑名单变化时自动重启 mihomo 重载配置
- 最多 5 个节点并行测速，避免压力过大
- 黑名单仅存内存，服务重启后重新检测

```bash
# 查看健康状态（黑名单列表、上次检查时间）
curl http://your-server:8080/api/health \
  -H "Authorization: Bearer your-token"
# → {"blacklist": [...], "blacklist_count": 3, "last_check_at": "2026-02-24 02:30:00", ...}

# 手动触发一次健康检查（异步执行）
curl -X POST http://your-server:8080/api/health/check \
  -H "Authorization: Bearer your-token"
# → {"message": "健康检查已触发"}
```

---

## 🌏 节点过滤

默认只保留低延迟地区节点：

<table>
<tr><th>地区</th><th>匹配关键词</th></tr>
<tr><td>🇭🇰 香港</td><td><code>HK</code> <code>HKG</code> <code>香港</code></td></tr>
<tr><td>🇸🇬 新加坡</td><td><code>SG</code> <code>SGP</code> <code>新加坡</code></td></tr>
<tr><td>🇹🇼 台湾</td><td><code>TW</code> <code>TWN</code> <code>台湾</code></td></tr>
<tr><td>🇯🇵 日本</td><td><code>JP</code> <code>JPN</code> <code>日本</code></td></tr>
</table>

> 修改过滤规则：编辑 `internal/subscription/parser.go` 中的 `allowedRegions`

---

## 📁 项目结构

```
clash-sub-aggregator/
├── main.go                          # 入口
├── internal/
│   ├── api/
│   │   ├── router.go                # 路由定义
│   │   ├── middleware.go            # Token 认证
│   │   ├── subscription_handler.go  # 订阅 CRUD
│   │   ├── proxy_handler.go         # 代理控制
│   │   └── health_handler.go        # 健康检查 API
│   ├── health/
│   │   └── checker.go               # 定时健康检查 + 黑名单
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
├── configs/app.yaml                 # 应用配置
├── Dockerfile
├── docker-compose.yaml
└── Makefile
```

---

<div align="center">

MIT License

</div>
