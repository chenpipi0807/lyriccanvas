// ============================================================
// LyricCanvas — 入口（单 exe 模式：内嵌前端 dist）
// ============================================================
package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"lyriccanvas/config"
	"lyriccanvas/handlers"
	"lyriccanvas/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

//go:embed frontend-dist/*
var frontendEmbed embed.FS

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}
	if cfg.DeepSeekAPIKey == "" {
		log.Printf("⚠️ DeepSeek API Key 未设置，请在页面右上角设置中填写")
	}

	store, err := services.NewProjectStore(cfg.DataDir)
	if err != nil {
		log.Fatalf("存储初始化失败: %v", err)
	}

	dsClient := services.NewDeepSeekClient(cfg.DeepSeekAPIKey, cfg.DeepSeekBaseURL)

	rhymeSvc, err := services.NewRhymeService()
	if err != nil {
		log.Printf("⚠️ 押韵服务初始化失败: %v", err)
	}

	projHandler := handlers.NewProjectHandler(store)
	chatHandler := handlers.NewChatHandler(dsClient, store)
	healthHandler := handlers.NewHealthHandler(dsClient, store)
	configHandler := handlers.NewConfigHandler(dsClient)
	var rhymeHandler *handlers.RhymeHandler
	if rhymeSvc != nil {
		rhymeHandler = handlers.NewRhymeHandler(rhymeSvc)
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	origins := []string{cfg.AllowedOrigins, "http://localhost:8848", "http://127.0.0.1:8848"}
	r.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	api := r.Group("/api")
	{
		api.GET("/health", healthHandler.Check)
		api.GET("/config", configHandler.GetConfig)
		api.PUT("/config/apikey", configHandler.SetAPIKey)
		api.GET("/projects", projHandler.List)
		api.POST("/projects", projHandler.Create)
		api.GET("/projects/:id", projHandler.Get)
		api.PUT("/projects/:id", projHandler.Update)
		api.DELETE("/projects/:id", projHandler.Delete)
		api.POST("/chat", chatHandler.Send)
		api.GET("/chat/history/:projectId", chatHandler.GetHistory)
		api.DELETE("/chat/history/:projectId", chatHandler.ClearHistory)
		api.GET("/chat/system-prompt", chatHandler.GetSystemPrompt)
		api.GET("/chat/templates", chatHandler.GetTemplates)
		api.GET("/chat/actions", chatHandler.GetActions)
		api.POST("/lyrics/parse", chatHandler.ParseLyrics)

		if rhymeHandler != nil {
			api.GET("/rhyme", rhymeHandler.Query)
			api.POST("/rhyme/batch", rhymeHandler.QueryBatch)
		}
	}

	staticFS, err := fs.Sub(frontendEmbed, "frontend-dist")
	if err != nil {
		log.Printf("⚠️ 未嵌入前端文件 (开发模式, 请用 npm run dev): %v", err)
	} else {
		r.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			if strings.HasPrefix(path, "/api/") {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			f, err := staticFS.Open(strings.TrimPrefix(path, "/"))
			if err == nil {
				f.Close()
				c.FileFromFS(path, http.FS(staticFS))
				return
			}
			c.FileFromFS("/", http.FS(staticFS))
		})
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🎵 LyricCanvas → http://localhost%s", addr)
	log.Printf("📁 数据: %s  🤖 默认模型: %s (前端可切换)", cfg.DataDir, cfg.DeepSeekModel)

	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
