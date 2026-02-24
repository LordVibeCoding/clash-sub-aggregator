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

// NodeResult 单个节点的测速结果
type NodeResult struct {
	Name  string `json:"name"`
	Delay int    `json:"delay"`
	Error string `json:"error,omitempty"`
}

// Checker 健康检查器，按需测速并维护黑名单
type Checker struct {
	controllerAddr func() string
	isRunning      func() bool
	subMgr         *subscription.Manager
	restart        func() error

	mu            sync.RWMutex
	blacklist     map[string]time.Time
	lastCheckAt   time.Time
	lastCheckCost time.Duration
	lastResults   []NodeResult
	checking      bool

	client *http.Client
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

// CheckAll 对所有节点执行健康检查，返回测速结果
func (c *Checker) CheckAll() []NodeResult {
	c.mu.Lock()
	if c.checking {
		c.mu.Unlock()
		log.Println("健康检查正在进行中，跳过")
		return nil
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
		return nil
	}

	start := time.Now()
	proxies := c.subMgr.AllProxiesUnfiltered()
	if len(proxies) == 0 {
		return nil
	}

	log.Printf("开始健康检查，共 %d 个节点", len(proxies))

	names := make([]string, 0, len(proxies))
	for _, p := range proxies {
		name := subscription.ProxyName(p)
		if name != "unknown" {
			names = append(names, name)
		}
	}

	// 并发测速，最多 5 个并行
	results := make(chan NodeResult, len(names))
	sem := make(chan struct{}, 5)

	var wg sync.WaitGroup
	for _, name := range names {
		wg.Add(1)
		go func(n string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			delay := c.checkNode(n)
			r := NodeResult{Name: n, Delay: delay}
			if delay == 0 {
				r.Error = "timeout"
			}
			results <- r
		}(name)
	}
	wg.Wait()
	close(results)

	// 收集结果 + 更新黑名单
	var allResults []NodeResult
	changed := false
	c.mu.Lock()
	for r := range results {
		allResults = append(allResults, r)
		_, inBlacklist := c.blacklist[r.Name]
		if r.Delay == 0 && !inBlacklist {
			c.blacklist[r.Name] = time.Now()
			changed = true
			log.Printf("节点不可用，加入黑名单: %s", r.Name)
		} else if r.Delay > 0 && inBlacklist {
			delete(c.blacklist, r.Name)
			changed = true
			log.Printf("节点已恢复，移出黑名单: %s", r.Name)
		}
	}
	c.lastCheckAt = time.Now()
	c.lastCheckCost = time.Since(start)
	c.lastResults = allResults
	blacklistCount := len(c.blacklist)
	c.mu.Unlock()

	log.Printf("健康检查完成，耗时 %s，黑名单 %d 个节点", c.lastCheckCost.Round(time.Second), blacklistCount)

	if changed {
		log.Println("黑名单变化，重启 mihomo 更新配置")
		if err := c.restart(); err != nil {
			log.Printf("重启 mihomo 失败: %v", err)
		}
	}

	return allResults
}

// checkNode 测试单个节点延迟，返回延迟毫秒数，0 表示失败
func (c *Checker) checkNode(name string) int {
	apiURL := fmt.Sprintf("%s/proxies/%s/delay?timeout=5000&url=http://www.gstatic.com/generate_204",
		c.controllerAddr(), url.PathEscape(name))

	resp, err := c.client.Get(apiURL)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return 0
	}

	if d, ok := body["delay"]; ok {
		if df, ok := d.(float64); ok {
			return int(df)
		}
	}
	return 0
}

// IsBlacklisted 检查节点是否在黑名单中
func (c *Checker) IsBlacklisted(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.blacklist[name]
	return ok
}

// Status 返回健康检查状态，包含有效节点及延迟
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

	// 有效节点（有延迟的）
	var available []map[string]any
	for _, r := range c.lastResults {
		if r.Delay > 0 {
			available = append(available, map[string]any{
				"name":  r.Name,
				"delay": r.Delay,
			})
		}
	}

	status := map[string]any{
		"blacklist_count":  len(c.blacklist),
		"blacklist":        bl,
		"available_count":  len(available),
		"available":        available,
		"checking":         c.checking,
	}
	if !c.lastCheckAt.IsZero() {
		status["last_check_at"] = c.lastCheckAt.Format("2006-01-02 15:04:05")
		status["last_check_cost"] = c.lastCheckCost.String()
	}
	return status
}
