// ============================================================
// LyricCanvas — 项目 CRUD HTTP 处理器
// ============================================================
package handlers

import (
	"net/http"

	"lyriccanvas/models"
	"lyriccanvas/services"

	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	store *services.ProjectStore
}

func NewProjectHandler(store *services.ProjectStore) *ProjectHandler {
	return &ProjectHandler{store: store}
}

// List 获取项目列表
func (h *ProjectHandler) List(c *gin.Context) {
	items, err := h.store.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if items == nil {
		items = []models.ProjectListItem{}
	}
	c.JSON(http.StatusOK, items)
}

// Get 获取单个项目
func (h *ProjectHandler) Get(c *gin.Context) {
	id := c.Param("id")
	proj, err := h.store.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, proj)
}

// Create 新建项目
func (h *ProjectHandler) Create(c *gin.Context) {
	var req models.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}
	if req.Name == "" {
		req.Name = "未命名歌曲"
	}

	proj, err := h.store.Create(req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, proj)
}

// Update 更新项目
func (h *ProjectHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req models.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	proj, err := h.store.Update(id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, proj)
}

// Delete 删除项目
func (h *ProjectHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.store.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
