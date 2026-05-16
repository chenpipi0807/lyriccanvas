// ============================================================
// LyricCanvas — 押韵词联想 HTTP 处理器
// ============================================================
package handlers

import (
	"net/http"

	"lyriccanvas/models"
	"lyriccanvas/services"

	"github.com/gin-gonic/gin"
)

type RhymeHandler struct {
	svc *services.RhymeService
}

func NewRhymeHandler(svc *services.RhymeService) *RhymeHandler {
	return &RhymeHandler{svc: svc}
}

// Query 单字押韵查询 GET /api/rhyme?char=花
func (h *RhymeHandler) Query(c *gin.Context) {
	char := c.Query("char")
	if char == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 char 参数"})
		return
	}

	result, err := h.svc.Query(char)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// QueryBatch 批量押韵查询 POST /api/rhyme/batch
func (h *RhymeHandler) QueryBatch(c *gin.Context) {
	var req models.BatchRhymeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	results, err := h.svc.QueryBatch(req.Chars)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}
