package health

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"clash-sub-aggregator/internal/subscription"
)

// Checker 健康检查器，定时测速并维护黑名单
type Checker struct {
	controllerAddr func() string
	isRunning      func() bool
	subMgr         *subscription.Manager
	restart        func() error

	mu            sync.RWMutex
	blacklist     map[string]time.Time // 节点名 → 加入黑名单时间
	lastCheckAt   time.Time
	lastCheckCost time.Duration
	checking      bool

	client *http.Client
	stopCh chan struct{}
}

func New(
	controllerAddr func() string,
	isRunning func() bool,
	subMgr *subscription.Manager,
	restart func() error,
) *Checker {
	return &Checker{
		controllerAddr: controllerAddr,
		isRunning:      isRunning,
		subMgr:         subMgr,
		restart:        restart,
		blacklist:      make(map[string]time.Time),
		client:         &http.Client{Timeout: 6 * time.Second},
	}
}

// Start 启动定时健康检查（30 分钟间隔）
func (c *Checker) Start() {
	c.stopCh = make(chan struct{})
	go func() {
		// 启动后 1 分钟执行首次检查
		select {
		case <-time.After(1 * time.Minute):
		case <-c.stopCh:
			return
		}
		c.CheckAll()

		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.CheckAll()
			case <-c.stopCh:
				return
			}
		}
	}()
	log.Println("健康检查器已启动，间隔 30 分钟")
}

// Stop 停止定时检查
func (c *Checker) Stop() {
	if c.stopCh != nil {
		close(c.stopCh)
	}
}

// CheckAll 对所有节点执行健康检查
func (c *Checker) CheckAll() {
	c.mu.Lock()
	if c.checking {
		c.mu.Unlock()
		log.Println("健康检查正在进行中，跳过")
		return
	}
	c.checking = true
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		c.checking = false
		c.mu.Unlock()
	}()

	if !c.isRunning() {
		log.Println("mihomo 未运行，跳过健康检查")
		return
	}

	start := time.Now()
	proxies := c.subMgr.AllProxiesUnfiltered()
	if len(proxies) == 0 {
		return
	}

	log.Printf("开始健康检查，共 %d 个节点", len(proxies))

	// 收集节点名
	names := make([]string, 0, len(proxies))
	for _, p := range proxies {
		name := subscription.ProxyName(p)
		if name != "unknown" {
			names = append(names, name)
		}
	}

	// 并发测速，最多 5 个并行
	type result struct {
		name    string
		healthy bool
	}
	results := make(chan result, len(names))
	sem := make(chan struct{}, 5)

	var wg sync.WaitGroup
	for _, name := range names {
		wg.Add(1)
		go func(n string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			healthy := c.checkNode(n)
			results <- result{name: n, healthy: healthy}
		}(name)
	}
	wg.Wait()
	close(results)

	// 更新黑名单
	changed := false
	c.mu.Lock()
	for r := range results {
		_, inBlacklist := c.blacklist[r.name]
		if !r.healthy && !inBlacklist {
			c.blacklist[r.name] = time.Now()
			changed = true
			log.Printf("节点不可用，加入黑名单: %s", r.name)
		} else if r.healthy && inBlacklist {
			delete(c.blacklist, r.name)
			changed = true
			log.Printf("节点已恢复，移出黑名单: %s", r.name)
		}
	}
	c.lastCheckAt = time.Now()
	c.lastCheckCost = time.Since(start)
	blacklistCount := len(c.blacklist)
	c.mu.Unlock()

	log.Printf("健康检查完成，耗时 %s，黑名单 %d 个节点", c.lastCheckCost.Round(time.Second), blacklistCount)

	// 黑名单有变化时重启 mihomo 重新生成配置
	if changed {
		log.Println("黑名单变化，重启 mihomo 更新配置")
		if err := c.restart(); err != nil {
			log.Printf("重启 mihomo 失败: %v", err)
		}
	}
}

// checkNode 测试单个节点延迟
func (c *Checker) checkNode(name string) bool {
	apiURL := fmt.Sprintf("%s/proxies/%s/delay?timeout=5000&url=http://www.gstatic.com/generate_204",
		c.controllerAddr(), url.PathEscape(name))

	resp, err := c.client.Get(apiURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return false
	}

	// mihomo 返回 {"delay": 123} 或 {"message": "..."}
	if _, ok := body["delay"]; ok {
		return true
	}
	return false
}

// IsBlacklisted 检查节点是否在黑名单中
func (c *Checker) IsBlacklisted(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.blacklist[name]
	return ok
}

// Status 返回健康检查状态
func (c *Checker) Status() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	bl := make([]map[string]any, 0, len(c.blacklist))
	for name, since := range c.blacklist {
		bl = append(bl, map[string]any{
			"name":  name,
			"since": since.Format("2006-01-02 15:04:05"),
		})
	}

	status := map[string]any{
		"blacklist_count": len(c.blacklist),
		"blacklist":       bl,
		"checking":        c.checking,
	}
	if !c.lastCheckAt.IsZero() {
		status["last_check_at"] = c.lastCheckAt.Format("2006-01-02 15:04:05")
		status["last_check_cost"] = c.lastCheckCost.String()
	}
	return status
}
