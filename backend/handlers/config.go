// ============================================================
// LyricCanvas — 配置管理 HTTP 处理器（API Key 设置等）
// ============================================================
package handlers

import (
	"net/http"
	"os"

	"lyriccanvas/config"
	"lyriccanvas/services"

	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	dsClient *services.DeepSeekClient
}

func NewConfigHandler(dsClient *services.DeepSeekClient) *ConfigHandler {
	return &ConfigHandler{dsClient: dsClient}
}

type ConfigResponse struct {
	HasAPIKey     bool   `json:"hasApiKey"`
	APIKeyPreview string `json:"apiKeyPreview"` // 脱敏显示，如 "sk-****f9"
	Port          string `json:"port"`
	BaseURL       string `json:"baseUrl"`
}

type SetAPIKeyRequest struct {
	APIKey string `json:"apiKey"`
}

// GetConfig 获取当前配置信息（GET /api/config）
func (h *ConfigHandler) GetConfig(c *gin.Context) {
	key := os.Getenv("DEEPSEEK_API_KEY")
	preview := ""
	hasKey := key != ""
	if hasKey && len(key) > 12 {
		preview = key[:5] + "****" + key[len(key)-2:]
	} else if hasKey {
		preview = key[:3] + "****"
	}

	c.JSON(http.StatusOK, ConfigResponse{
		HasAPIKey:     hasKey,
		APIKeyPreview: preview,
		Port:          os.Getenv("PORT"),
		BaseURL:       os.Getenv("DEEPSEEK_BASE_URL"),
	})
}

// SetAPIKey 设置 API Key（PUT /api/config/apikey）
func (h *ConfigHandler) SetAPIKey(c *gin.Context) {
	var req SetAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.APIKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供有效的 API Key"})
		return
	}

	// 写入 .env 文件
	if err := config.SetEnvValue("DEEPSEEK_API_KEY", req.APIKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存配置失败: " + err.Error()})
		return
	}

	// 更新运行时的 DeepSeek 客户端
	h.dsClient.SetAPIKey(req.APIKey)

	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "API Key 已保存并生效"})
}
