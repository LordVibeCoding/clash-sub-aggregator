package subscription

import (
	"encoding/base64"
	"fmt"
	"strings"

	"clash-sub-aggregator/internal/model"

	"gopkg.in/yaml.v3"
)

// clashConfig 用于解析订阅返回的 Clash YAML
type clashConfig struct {
	Proxies []model.Proxy `yaml:"proxies"`
}

// Parse 解析订阅内容，支持 Clash YAML 和 Base64 编码格式
func Parse(raw []byte) ([]model.Proxy, error) {
	content := strings.TrimSpace(string(raw))

	// 尝试 base64 解码
	if decoded, err := base64.StdEncoding.DecodeString(content); err == nil {
		content = strings.TrimSpace(string(decoded))
	} else if decoded, err := base64.RawStdEncoding.DecodeString(content); err == nil {
		content = strings.TrimSpace(string(decoded))
	}

	// 解析为 Clash YAML
	var cfg clashConfig
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		return nil, fmt.Errorf("解析订阅内容失败: %w", err)
	}

	if len(cfg.Proxies) == 0 {
		return nil, fmt.Errorf("订阅内容中没有找到代理节点")
	}

	return FilterByRegion(cfg.Proxies), nil
}

// allowedRegions 只保留这些地区的节点
var allowedRegions = []string{
	"HK", "HKG", "香港", "🇭🇰",
	"SG", "SGP", "新加坡", "🇸🇬",
	"TW", "TWN", "台湾", "🇹🇼",
	"JP", "JPN", "日本", "🇯🇵",
}

// supportedTypes mihomo 支持的代理协议
var supportedTypes = map[string]bool{
	"ss": true, "ssr": true, "vmess": true, "vless": true,
	"trojan": true, "hysteria": true, "hysteria2": true,
	"wireguard": true, "tuic": true, "socks5": true, "http": true,
	"snell": true, "ssh": true,
}

// FilterByRegion 过滤只保留港/新/台/日节点，并排除不支持的协议
func FilterByRegion(proxies []model.Proxy) []model.Proxy {
	var filtered []model.Proxy
	for _, p := range proxies {
		// 排除不支持的协议类型
		if t, ok := p["type"].(string); ok && !supportedTypes[strings.ToLower(t)] {
			continue
		}
		name := strings.ToUpper(ProxyName(p))
		for _, region := range allowedRegions {
			if strings.Contains(name, strings.ToUpper(region)) {
				filtered = append(filtered, p)
				break
			}
		}
	}
	return filtered
}

// ProxyName 从 proxy map 中提取名称
func ProxyName(p model.Proxy) string {
	if name, ok := p["name"].(string); ok {
		return name
	}
	return "unknown"
}
