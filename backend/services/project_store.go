// ============================================================
// LyricCanvas — 本地 JSON 文件项目存储
// ============================================================
package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"lyriccanvas/models"

	"github.com/google/uuid"
)

type ProjectStore struct {
	mu      sync.RWMutex
	dataDir string
}

func NewProjectStore(dataDir string) (*ProjectStore, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}
	return &ProjectStore{dataDir: dataDir}, nil
}

func (s *ProjectStore) filePath(id string) string {
	return filepath.Join(s.dataDir, id+".json")
}

// List 获取所有项目列表（只含摘要）
func (s *ProjectStore) List() ([]models.ProjectListItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.dataDir)
	if err != nil {
		return nil, fmt.Errorf("读取数据目录失败: %w", err)
	}

	var items []models.ProjectListItem
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		proj, err := s.loadUnsafe(entry.Name()[:len(entry.Name())-5])
		if err != nil {
			continue // 跳过损坏文件
		}
		items = append(items, models.ProjectListItem{
			ID:        proj.ID,
			Name:      proj.Name,
			CreatedAt: proj.CreatedAt,
			UpdatedAt: proj.UpdatedAt,
			NodeCount: len(proj.Canvas.Nodes),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].UpdatedAt.After(items[j].UpdatedAt)
	})

	return items, nil
}

// Get 获取单个项目完整数据
func (s *ProjectStore) Get(id string) (*models.Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.loadUnsafe(id)
}

func (s *ProjectStore) loadUnsafe(id string) (*models.Project, error) {
	data, err := os.ReadFile(s.filePath(id))
	if err != nil {
		return nil, fmt.Errorf("项目不存在: %s", id)
	}
	var proj models.Project
	if err := json.Unmarshal(data, &proj); err != nil {
		return nil, fmt.Errorf("项目数据解析失败: %w", err)
	}
	return &proj, nil
}

// Create 新建项目
func (s *ProjectStore) Create(name string) (*models.Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	proj := &models.Project{
		ID:        "proj_" + uuid.New().String()[:8],
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
		Canvas: models.Canvas{
			Viewport: models.Viewport{X: 0, Y: 0, Zoom: 1},
			Nodes:    []models.LyricNode{},
			Groups:   []models.GroupLabel{},
			DraftZone: models.DraftZone{
				X: 100, Y: 100, Width: 600, Height: 800,
			},
		},
		ChatHistory: []models.ChatMessage{},
	}

	if err := s.saveUnsafe(proj); err != nil {
		return nil, err
	}
	return proj, nil
}

// Update 更新项目
func (s *ProjectStore) Update(id string, req *models.UpdateProjectRequest) (*models.Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	proj, err := s.loadUnsafe(id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		proj.Name = req.Name
	}
	if req.Canvas != nil {
		proj.Canvas = *req.Canvas
	}
	if req.ChatHistory != nil {
		proj.ChatHistory = req.ChatHistory
	}
	proj.UpdatedAt = time.Now()

	if err := s.saveUnsafe(proj); err != nil {
		return nil, err
	}
	return proj, nil
}

// AppendChatMessage 追加一条对话记录
func (s *ProjectStore) AppendChatMessage(projectID string, msg models.ChatMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	proj, err := s.loadUnsafe(projectID)
	if err != nil {
		return err
	}
	proj.ChatHistory = append(proj.ChatHistory, msg)
	proj.UpdatedAt = time.Now()
	return s.saveUnsafe(proj)
}

// Delete 删除项目
func (s *ProjectStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.Remove(s.filePath(id)); err != nil {
		return fmt.Errorf("删除项目失败: %w", err)
	}
	return nil
}

func (s *ProjectStore) saveUnsafe(proj *models.Project) error {
	data, err := json.MarshalIndent(proj, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化项目数据失败: %w", err)
	}
	if err := os.WriteFile(s.filePath(proj.ID), data, 0644); err != nil {
		return fmt.Errorf("写入项目文件失败: %w", err)
	}
	return nil
}

// Count 获取项目总数
func (s *ProjectStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entries, _ := os.ReadDir(s.dataDir)
	count := 0
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			count++
		}
	}
	return count
}
