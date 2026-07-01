package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// AppConfig 总配置
type AppConfig struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	WxWork   WxWorkConfig   `yaml:"wxwork"`
}

// ServerConfig 服务配置
type ServerConfig struct {
	Port string `yaml:"port"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// WxWorkConfig 企业微信配置
type WxWorkConfig struct {
	CorpID      string `yaml:"corp_id"`
	AgentID     string `yaml:"agent_id"`
	Secret      string `yaml:"secret"`
	RedirectURI string `yaml:"redirect_uri"`
}

var Cfg AppConfig

// Load 从 YAML 文件加载配置
func Load(path string) {
	// 默认值
	Cfg = AppConfig{
		Server: ServerConfig{
			Port: "8485",
		},
		Database: DatabaseConfig{
			Driver: "sqlite",
			DSN:    "exam.db",
		},
		Redis: RedisConfig{
			Host: "127.0.0.1",
			Port: "6379",
			DB:   0,
		},
	}

	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("⚠️ 配置文件 %s 读取失败，使用默认配置: %v", path, err)
		return
	}

	if err := yaml.Unmarshal(data, &Cfg); err != nil {
		log.Fatalf("配置文件解析失败: %v", err)
	}

	log.Println("✅ 配置文件加载成功:", path)
}
