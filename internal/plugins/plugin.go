package plugins

import (
	"context"
	"encoding/json"
)

// Plugin represents a transformable component that can modify requests and responses
type Plugin interface {
	// Name returns the plugin name
	Name() string

	// Description returns a short description of what the plugin does
	Description() string

	// Priority returns the execution priority (lower = earlier execution)
	Priority() int

	// Enabled checks if the plugin should be active
	Enabled(ctx context.Context) bool
}

// RequestTransformer can modify requests before they're sent to providers
type RequestTransformer interface {
	Plugin

	// TransformRequest modifies the request body
	// Returns the modified request or the original if no transformation needed
	TransformRequest(ctx context.Context, request []byte) ([]byte, error)
}

// ResponseTransformer can modify responses before they're sent to clients
type ResponseTransformer interface {
	Plugin

	// TransformResponse modifies the response body
	// Returns the modified response or the original if no transformation needed
	TransformResponse(ctx context.Context, response []byte) ([]byte, error)
}

// StreamTransformer can modify streaming responses chunk by chunk
type StreamTransformer interface {
	Plugin

	// TransformChunk modifies a single SSE chunk
	TransformChunk(ctx context.Context, chunk []byte) ([]byte, error)
}

// MetadataPlugin can inspect and log request/response metadata
type MetadataPlugin interface {
	Plugin

	// OnRequest is called when a request is received
	OnRequest(ctx context.Context, metadata RequestMetadata)

	// OnResponse is called when a response is received
	OnResponse(ctx context.Context, metadata ResponseMetadata)
}

// RequestMetadata contains information about a request
type RequestMetadata struct {
	Provider     string
	Model        string
	InputTokens  int
	HasStreaming bool
	Raw          json.RawMessage
}

// ResponseMetadata contains information about a response
type ResponseMetadata struct {
	Provider      string
	Model         string
	OutputTokens  int
	Status        int
	DurationMs    int64
	CachedTokens  int
	Raw           json.RawMessage
}

// Context keys for plugin data
type contextKey string

const (
	// ProviderKey is the context key for the provider name
	ProviderKey contextKey = "provider"
	// ModelKey is the context key for the model name
	ModelKey contextKey = "model"
	// RequestIDKey is the context key for request tracking
	RequestIDKey contextKey = "request_id"
)

// GetProvider retrieves the provider name from context
func GetProvider(ctx context.Context) string {
	if v := ctx.Value(ProviderKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetModel retrieves the model name from context
func GetModel(ctx context.Context) string {
	if v := ctx.Value(ModelKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
	if v := ctx.Value(RequestIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
