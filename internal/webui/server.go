package webui

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/Davincible/claude-code-open/internal/config"
	"github.com/Davincible/claude-code-open/internal/plugins"
)

//go:embed static/*
var staticFiles embed.FS

// Server provides a web UI for managing CCO
type Server struct {
	config   *config.Manager
	plugins  *plugins.Registry
	logger   *slog.Logger
	stats    *Stats
	mux      *http.ServeMux
}

// Stats tracks request statistics
type Stats struct {
	mu              sync.RWMutex
	TotalRequests   int64                  `json:"total_requests"`
	TotalTokens     int64                  `json:"total_tokens"`
	RequestsByModel map[string]int64       `json:"requests_by_model"`
	TokensByModel   map[string]int64       `json:"tokens_by_model"`
	RecentRequests  []RequestLog           `json:"recent_requests"`
	StartTime       time.Time              `json:"start_time"`
}

// RequestLog stores information about a single request
type RequestLog struct {
	Timestamp    time.Time `json:"timestamp"`
	Provider     string    `json:"provider"`
	Model        string    `json:"model"`
	InputTokens  int       `json:"input_tokens"`
	OutputTokens int       `json:"output_tokens"`
	DurationMs   int64     `json:"duration_ms"`
	Status       int       `json:"status"`
}

// NewServer creates a new web UI server
func NewServer(cfg *config.Manager, pluginRegistry *plugins.Registry, logger *slog.Logger) *Server {
	s := &Server{
		config:  cfg,
		plugins: pluginRegistry,
		logger:  logger,
		stats: &Stats{
			RequestsByModel: make(map[string]int64),
			TokensByModel:   make(map[string]int64),
			RecentRequests:  make([]RequestLog, 0, 100),
			StartTime:       time.Now(),
		},
		mux: http.NewServeMux(),
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// Serve static files
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		s.logger.Error("Failed to create static file system", "error", err)
		return
	}

	s.mux.Handle("/", http.FileServer(http.FS(staticFS)))

	// API endpoints
	s.mux.HandleFunc("/api/config", s.handleConfig)
	s.mux.HandleFunc("/api/providers", s.handleProviders)
	s.mux.HandleFunc("/api/models", s.handleModels)
	s.mux.HandleFunc("/api/plugins", s.handlePlugins)
	s.mux.HandleFunc("/api/stats", s.handleStats)
	s.mux.HandleFunc("/api/health", s.handleHealth)
}

// ServeHTTP implements http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// API Handlers

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	cfg := s.config.Get()

	// Mask API keys before sending
	maskedCfg := *cfg
	for i := range maskedCfg.Providers {
		maskedCfg.Providers[i].APIKey = maskAPIKey(maskedCfg.Providers[i].APIKey)
	}
	if maskedCfg.APIKey != "" {
		maskedCfg.APIKey = maskString(maskedCfg.APIKey)
	}

	s.jsonResponse(w, maskedCfg)
}

func (s *Server) handleProviders(w http.ResponseWriter, r *http.Request) {
	cfg := s.config.Get()

	type ProviderInfo struct {
		Name    string   `json:"name"`
		URL     string   `json:"url"`
		HasKey  bool     `json:"has_key"`
		Models  int      `json:"models"`
		Enabled bool     `json:"enabled"`
	}

	providers := make([]ProviderInfo, 0, len(cfg.Providers))
	for _, p := range cfg.Providers {
		hasKey := false
		if key, ok := p.APIKey.(string); ok && key != "" {
			hasKey = true
		}

		models := len(p.GetAllowedModels())
		if models == 0 {
			models = len(p.DefaultModels)
		}

		providers = append(providers, ProviderInfo{
			Name:    p.Name,
			URL:     p.APIBase,
			HasKey:  hasKey,
			Models:  models,
			Enabled: hasKey,
		})
	}

	s.jsonResponse(w, providers)
}

func (s *Server) handleModels(w http.ResponseWriter, r *http.Request) {
	cfg := s.config.Get()

	type ModelInfo struct {
		Provider string `json:"provider"`
		Model    string `json:"model"`
		FullName string `json:"full_name"`
	}

	models := make([]ModelInfo, 0)
	for _, p := range cfg.Providers {
		modelList := p.GetAllowedModels()
		if len(modelList) == 0 {
			modelList = p.DefaultModels
		}

		for _, m := range modelList {
			models = append(models, ModelInfo{
				Provider: p.Name,
				Model:    m,
				FullName: fmt.Sprintf("%s/%s", p.Name, m),
			})
		}
	}

	s.jsonResponse(w, models)
}

func (s *Server) handlePlugins(w http.ResponseWriter, r *http.Request) {
	if s.plugins == nil {
		s.jsonResponse(w, map[string]interface{}{
			"error": "plugin registry not available",
		})
		return
	}

	pluginList := s.plugins.ListPlugins()
	s.jsonResponse(w, pluginList)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	s.stats.mu.RLock()
	defer s.stats.mu.RUnlock()

	type StatsResponse struct {
		TotalRequests   int64              `json:"total_requests"`
		TotalTokens     int64              `json:"total_tokens"`
		RequestsByModel map[string]int64   `json:"requests_by_model"`
		TokensByModel   map[string]int64   `json:"tokens_by_model"`
		RecentRequests  []RequestLog       `json:"recent_requests"`
		Uptime          string             `json:"uptime"`
		StartTime       time.Time          `json:"start_time"`
	}

	uptime := time.Since(s.stats.StartTime)

	resp := StatsResponse{
		TotalRequests:   s.stats.TotalRequests,
		TotalTokens:     s.stats.TotalTokens,
		RequestsByModel: s.stats.RequestsByModel,
		TokensByModel:   s.stats.TokensByModel,
		RecentRequests:  s.stats.RecentRequests,
		Uptime:          uptime.String(),
		StartTime:       s.stats.StartTime,
	}

	s.jsonResponse(w, resp)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.jsonResponse(w, map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// LogRequest adds a request to the statistics
func (s *Server) LogRequest(provider, model string, inputTokens, outputTokens int, durationMs int64, status int) {
	s.stats.mu.Lock()
	defer s.stats.mu.Unlock()

	s.stats.TotalRequests++
	s.stats.TotalTokens += int64(inputTokens + outputTokens)

	modelKey := fmt.Sprintf("%s/%s", provider, model)
	s.stats.RequestsByModel[modelKey]++
	s.stats.TokensByModel[modelKey] += int64(inputTokens + outputTokens)

	// Add to recent requests (keep last 100)
	log := RequestLog{
		Timestamp:    time.Now(),
		Provider:     provider,
		Model:        model,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		DurationMs:   durationMs,
		Status:       status,
	}

	s.stats.RecentRequests = append(s.stats.RecentRequests, log)
	if len(s.stats.RecentRequests) > 100 {
		s.stats.RecentRequests = s.stats.RecentRequests[1:]
	}
}

// Helper functions

func (s *Server) jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.logger.Error("Failed to encode JSON response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func maskAPIKey(key interface{}) string {
	if key == nil {
		return ""
	}

	switch v := key.(type) {
	case string:
		return maskString(v)
	case []interface{}:
		return fmt.Sprintf("[%d keys configured]", len(v))
	default:
		return "[configured]"
	}
}

func maskString(s string) string {
	if s == "" {
		return ""
	}
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "****" + s[len(s)-4:]
}
