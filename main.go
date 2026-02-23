package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"clash-sub-aggregator/internal/api"
	"clash-sub-aggregator/internal/clash"
	"clash-sub-aggregator/internal/health"
	"clash-sub-aggregator/internal/model"
	"clash-sub-aggregator/internal/store"
	"clash-sub-aggregator/internal/subscription"

	"gopkg.in/yaml.v3"
)

func main() {
	cfg := loadConfig()

	// 初始化存储
	s, err := store.New(cfg.DataDir)
	if err != nil {
		log.Fatalf("初始化存储失败: %v", err)
	}

	// 初始化订阅管理器
	subMgr := subscription.NewManager(s)

	// 初始化 mihomo 进程管理
	proc := clash.NewProcess(cfg.Mihomo, subMgr)

	// 初始化健康检查器
	hc := health.New(proc.ControllerAddr, proc.IsRunning, subMgr, proc.Restart)
	subMgr.SetBlacklistChecker(hc)

	// 如果有订阅数据，自动启动 mihomo
	if len(subMgr.AllProxies()) > 0 {
		if err := proc.Start(); err != nil {
			log.Printf("自动启动 mihomo 失败: %v", err)
		}
	} else {
		log.Println("暂无代理节点，添加订阅后将自动启动 mihomo")
	}

	// 启动 API 服务
	router := api.NewRouter(cfg.Server.Token, subMgr, proc, hc)
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("管理 API 启动在 %s", addr)
	log.Printf("代理端口: HTTP/SOCKS5 :%d", cfg.Mihomo.HTTPPort)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

func loadConfig() model.AppConfig {
	cfg := model.AppConfig{
		Server:  model.ServerConfig{Port: 8080},
		Mihomo:  model.MihomoConfig{
			Binary:         "/usr/local/bin/mihomo",
			ConfigDir:      "./data",
			HTTPPort:       7890,
			SocksPort:      7891,
			ControllerPort: 9090,
		},
		DataDir: "./data",
	}

	configPath := "configs/app.yaml"
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		configPath = envPath
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("使用默认配置 (未找到 %s)", configPath)
		return cfg
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Printf("配置文件解析失败，使用默认配置: %v", err)
	}
	return cfg
}
