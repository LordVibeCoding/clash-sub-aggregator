package model

import "time"

// Subscription 订阅源
type Subscription struct {
	ID        string    `json:"id" yaml:"id"`
	Name      string    `json:"name" yaml:"name"`
	URL       string    `json:"url" yaml:"url"`
	Proxies   []Proxy   `json:"proxies,omitempty" yaml:"proxies,omitempty"`
	UpdatedAt time.Time `json:"updated_at" yaml:"updated_at"`
	CreatedAt time.Time `json:"created_at" yaml:"created_at"`
}

// Proxy 代理节点，用 map 保留所有字段（不同协议字段差异大）
type Proxy = map[string]any

// AppConfig 应用配置
type AppConfig struct {
	Server   ServerConfig `yaml:"server"`
	Mihomo   MihomoConfig `yaml:"mihomo"`
	DataDir  string       `yaml:"data_dir"`
}

type ServerConfig struct {
	Port  int    `yaml:"port"`
	Token string `yaml:"token"`
}

type MihomoConfig struct {
	Binary           string `yaml:"binary"`
	ConfigDir        string `yaml:"config_dir"`
	HTTPPort         int    `yaml:"http_port"`
	SocksPort        int    `yaml:"socks_port"`
	ControllerPort   int    `yaml:"controller_port"`
	ControllerSecret string `yaml:"controller_secret"`
}

// StoreData JSON 持久化结构
type StoreData struct {
	Subscriptions []Subscription `json:"subscriptions"`
}
