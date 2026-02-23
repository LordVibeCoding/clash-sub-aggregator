package subscription

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"clash-sub-aggregator/internal/model"
	"clash-sub-aggregator/internal/store"

	"github.com/google/uuid"
)

type Manager struct {
	store  *store.Store
	client *http.Client
}

func NewManager(s *store.Store) *Manager {
	return &Manager{
		store: s,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Add 添加订阅并立即拉取解析
func (m *Manager) Add(name, url string) (model.Subscription, error) {
	sub := model.Subscription{
		ID:        uuid.New().String()[:8],
		Name:      name,
		URL:       url,
		CreatedAt: time.Now(),
	}

	proxies, err := m.fetch(url)
	if err != nil {
		return model.Subscription{}, fmt.Errorf("拉取订阅失败: %w", err)
	}
	sub.Proxies = proxies
	sub.UpdatedAt = time.Now()

	if err := m.store.Add(sub); err != nil {
		return model.Subscription{}, err
	}
	return sub, nil
}

func (m *Manager) Delete(id string) error {
	return m.store.Delete(id)
}

func (m *Manager) List() []model.Subscription {
	return m.store.List()
}

// RefreshAll 刷新所有订阅
func (m *Manager) RefreshAll() (int, error) {
	subs := m.store.List()
	total := 0
	var lastErr error

	for _, sub := range subs {
		proxies, err := m.fetch(sub.URL)
		if err != nil {
			lastErr = fmt.Errorf("刷新 %s 失败: %w", sub.Name, err)
			continue
		}
		sub.Proxies = proxies
		sub.UpdatedAt = time.Now()
		if err := m.store.Update(sub); err != nil {
			lastErr = err
			continue
		}
		total += len(proxies)
	}
	return total, lastErr
}

// RefreshOne 刷新单个订阅
func (m *Manager) RefreshOne(id string) (model.Subscription, error) {
	sub, ok := m.store.Get(id)
	if !ok {
		return model.Subscription{}, fmt.Errorf("订阅 %s 不存在", id)
	}

	proxies, err := m.fetch(sub.URL)
	if err != nil {
		return model.Subscription{}, fmt.Errorf("拉取订阅失败: %w", err)
	}
	sub.Proxies = proxies
	sub.UpdatedAt = time.Now()

	if err := m.store.Update(sub); err != nil {
		return model.Subscription{}, err
	}
	return sub, nil
}

// AllProxies 获取所有订阅的代理节点（去重）
func (m *Manager) AllProxies() []model.Proxy {
	subs := m.store.List()
	seen := make(map[string]bool)
	var all []model.Proxy

	for _, sub := range subs {
		for _, p := range sub.Proxies {
			name := ProxyName(p)
			if !seen[name] {
				seen[name] = true
				all = append(all, p)
			}
		}
	}
	return all
}

func (m *Manager) fetch(url string) ([]model.Proxy, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "clash-sub-aggregator/1.0")

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return Parse(body)
}
