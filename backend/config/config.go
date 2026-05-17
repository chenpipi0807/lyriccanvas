// ============================================================
// LyricCanvas — Go 后端配置加载
// ============================================================
package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"

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

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// SetEnvValue 写入 .env 文件中的指定键值（保留其他行不变）
func SetEnvValue(key, value string) error {
	const envFile = ".env"

	// 读取已有内容
	var lines []string
	found := false

	f, err := os.Open(envFile)
	if err == nil {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(strings.TrimSpace(line), key+"=") || strings.HasPrefix(strings.TrimSpace(line), key+" =") {
				lines = append(lines, key+"="+value)
				found = true
			} else {
				lines = append(lines, line)
			}
		}
		f.Close()
	}

	if !found {
		lines = append(lines, key+"="+value)
	}

	// 写回
	out := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(envFile, []byte(out), 0644); err != nil {
		return fmt.Errorf("写入 .env 失败: %w", err)
	}

	// 同步更新当前进程环境变量
	os.Setenv(key, value)
	return nil
}
