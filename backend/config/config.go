// ============================================================
// LyricCanvas — Go 后端配置加载
// ============================================================
package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	DataDir           string
	DeepSeekAPIKey    string
	DeepSeekBaseURL   string
	DeepSeekModel     string
	AllowedOrigins    string
}

func Load() (*Config, error) {
	// .env 文件加载失败不致命（可能用系统环境变量）
	_ = godotenv.Load()

	cfg := &Config{
		Port:            getEnv("PORT", "8080"),
		DataDir:         getEnv("DATA_DIR", "./data"),
		DeepSeekAPIKey:  os.Getenv("DEEPSEEK_API_KEY"),
		DeepSeekBaseURL: getEnv("DEEPSEEK_BASE_URL", "https://api.deepseek.com"),
		DeepSeekModel:   getEnv("DEEPSEEK_DEFAULT_MODEL", "deepseek-chat"),
		AllowedOrigins:  getEnv("ALLOWED_ORIGINS", "http://localhost:5173"),
	}

	if cfg.DeepSeekAPIKey == "" {
		return cfg, fmt.Errorf("DEEPSEEK_API_KEY 未设置，请在 .env 文件或环境变量中配置")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
