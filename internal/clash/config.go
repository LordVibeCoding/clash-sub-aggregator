package clash

import (
	"fmt"

	"clash-sub-aggregator/internal/model"
	"clash-sub-aggregator/internal/subscription"

	"gopkg.in/yaml.v3"
)

// GenerateConfig 根据所有代理节点生成 mihomo 配置
func GenerateConfig(proxies []model.Proxy, cfg model.MihomoConfig) ([]byte, error) {
	proxyNames := make([]string, 0, len(proxies))
	for _, p := range proxies {
		proxyNames = append(proxyNames, subscription.ProxyName(p))
	}

	config := map[string]any{
		"mixed-port":          cfg.HTTPPort,
		"socks-port":          cfg.SocksPort,
		"allow-lan":           true,
		"bind-address":        "*",
		"mode":                "rule",
		"log-level":           "info",
		"external-controller": fmt.Sprintf("0.0.0.0:%d", cfg.ControllerPort),
		"proxies":             proxies,
		"proxy-groups": []map[string]any{
			{
				"name":    "PROXY",
				"type":    "select",
				"proxies": append(proxyNames, "DIRECT"),
			},
			{
				"name":     "AUTO",
				"type":     "url-test",
				"proxies":  proxyNames,
				"url":      "http://www.gstatic.com/generate_204",
				"interval": 300,
			},
		},
		"rules": []string{
			"MATCH,PROXY",
		},
	}

	if cfg.ControllerSecret != "" {
		config["secret"] = cfg.ControllerSecret
	}

	return yaml.Marshal(config)
}
