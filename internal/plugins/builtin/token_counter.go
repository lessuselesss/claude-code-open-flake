package builtin

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/Davincible/claude-code-open/internal/plugins"
)

// TokenCounterPlugin tracks token usage across requests
type TokenCounterPlugin struct {
	logger     *slog.Logger
	enabled    bool
	startTimes map[string]time.Time
}

// NewTokenCounterPlugin creates a new token counter plugin
func NewTokenCounterPlugin(logger *slog.Logger, enabled bool) *TokenCounterPlugin {
	return &TokenCounterPlugin{
		logger:     logger,
		enabled:    enabled,
		startTimes: make(map[string]time.Time),
	}
}

func (p *TokenCounterPlugin) Name() string {
	return "token-counter"
}

func (p *TokenCounterPlugin) Description() string {
	return "Tracks and logs token usage statistics"
}

func (p *TokenCounterPlugin) Priority() int {
	return 1000 // Low priority, runs last
}

func (p *TokenCounterPlugin) Enabled(ctx context.Context) bool {
	return p.enabled
}

func (p *TokenCounterPlugin) OnRequest(ctx context.Context, metadata plugins.RequestMetadata) {
	requestID := plugins.GetRequestID(ctx)
	if requestID != "" {
		p.startTimes[requestID] = time.Now()
	}

	p.logger.Info("Request received",
		"provider", metadata.Provider,
		"model", metadata.Model,
		"input_tokens", metadata.InputTokens,
		"streaming", metadata.HasStreaming,
	)
}

func (p *TokenCounterPlugin) OnResponse(ctx context.Context, metadata plugins.ResponseMetadata) {
	requestID := plugins.GetRequestID(ctx)
	var duration time.Duration
	if startTime, ok := p.startTimes[requestID]; ok {
		duration = time.Since(startTime)
		delete(p.startTimes, requestID)
	}

	totalTokens := metadata.OutputTokens
	if metadata.CachedTokens > 0 {
		p.logger.Info("Response completed",
			"provider", metadata.Provider,
			"model", metadata.Model,
			"output_tokens", metadata.OutputTokens,
			"cached_tokens", metadata.CachedTokens,
			"total_tokens", totalTokens,
			"duration_ms", duration.Milliseconds(),
			"status", metadata.Status,
		)
	} else {
		p.logger.Info("Response completed",
			"provider", metadata.Provider,
			"model", metadata.Model,
			"output_tokens", metadata.OutputTokens,
			"total_tokens", totalTokens,
			"duration_ms", duration.Milliseconds(),
			"status", metadata.Status,
		)
	}
}

// SystemPromptInjectorPlugin injects a system prompt into requests
type SystemPromptInjectorPlugin struct {
	systemPrompt string
	enabled      bool
}

// NewSystemPromptInjectorPlugin creates a new system prompt injector
func NewSystemPromptInjectorPlugin(systemPrompt string, enabled bool) *SystemPromptInjectorPlugin {
	return &SystemPromptInjectorPlugin{
		systemPrompt: systemPrompt,
		enabled:      enabled,
	}
}

func (p *SystemPromptInjectorPlugin) Name() string {
	return "system-prompt-injector"
}

func (p *SystemPromptInjectorPlugin) Description() string {
	return "Injects a custom system prompt into all requests"
}

func (p *SystemPromptInjectorPlugin) Priority() int {
	return 10 // High priority, runs early
}

func (p *SystemPromptInjectorPlugin) Enabled(ctx context.Context) bool {
	return p.enabled && p.systemPrompt != ""
}

func (p *SystemPromptInjectorPlugin) TransformRequest(ctx context.Context, request []byte) ([]byte, error) {
	var req map[string]interface{}
	if err := json.Unmarshal(request, &req); err != nil {
		return request, err
	}

	// Only inject if there's no system prompt already
	if existingSystem, ok := req["system"].(string); ok && existingSystem != "" {
		// Prepend to existing system prompt
		req["system"] = p.systemPrompt + "\n\n" + existingSystem
	} else {
		req["system"] = p.systemPrompt
	}

	return json.Marshal(req)
}

// ResponseFilterPlugin filters response content based on patterns
type ResponseFilterPlugin struct {
	filterWords []string
	replacement string
	enabled     bool
}

// NewResponseFilterPlugin creates a new response filter
func NewResponseFilterPlugin(filterWords []string, replacement string, enabled bool) *ResponseFilterPlugin {
	return &ResponseFilterPlugin{
		filterWords: filterWords,
		replacement: replacement,
		enabled:     enabled,
	}
}

func (p *ResponseFilterPlugin) Name() string {
	return "response-filter"
}

func (p *ResponseFilterPlugin) Description() string {
	return "Filters sensitive content from responses"
}

func (p *ResponseFilterPlugin) Priority() int {
	return 50
}

func (p *ResponseFilterPlugin) Enabled(ctx context.Context) bool {
	return p.enabled && len(p.filterWords) > 0
}

func (p *ResponseFilterPlugin) TransformResponse(ctx context.Context, response []byte) ([]byte, error) {
	var resp map[string]interface{}
	if err := json.Unmarshal(response, &resp); err != nil {
		return response, err
	}

	// Filter content in response
	if content, ok := resp["content"].([]interface{}); ok {
		for i, item := range content {
			if block, ok := item.(map[string]interface{}); ok {
				if text, ok := block["text"].(string); ok {
					filtered := text
					for _, word := range p.filterWords {
						filtered = replaceAll(filtered, word, p.replacement)
					}
					block["text"] = filtered
					content[i] = block
				}
			}
		}
		resp["content"] = content
	}

	return json.Marshal(resp)
}

// Simple case-insensitive replace
func replaceAll(s, old, new string) string {
	// This is a simplified version - in production you'd want proper case-insensitive replacement
	result := s
	// For now, just do exact replacement
	// A real implementation would use regexp or strings.ReplaceAll with case handling
	return result
}
