// ============================================================
// LyricCanvas — 健康检查处理器
// ============================================================
package handlers

import (
	"net/http"

	"lyriccanvas/models"
	"lyriccanvas/services"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	dsClient *services.DeepSeekClient
	store    *services.ProjectStore
}

func NewHealthHandler(dsClient *services.DeepSeekClient, store *services.ProjectStore) *HealthHandler {
	return &HealthHandler{dsClient: dsClient, store: store}
}

func (h *HealthHandler) Check(c *gin.Context) {
	deepseekOK := h.dsClient.ValidateKey()
	projectCount := 0
	if h.store != nil {
		projectCount = h.store.Count()
	}

	c.JSON(http.StatusOK, models.HealthResponse{
		Status:       "ok",
		DeepSeekOK:   deepseekOK,
		ProjectCount: projectCount,
	})
}
