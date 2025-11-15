package plugins

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// Registry manages all registered plugins
type Registry struct {
	mu                  sync.RWMutex
	requestTransformers []RequestTransformer
	responseTransformers []ResponseTransformer
	streamTransformers  []StreamTransformer
	metadataPlugins     []MetadataPlugin
}

// NewRegistry creates a new plugin registry
func NewRegistry() *Registry {
	return &Registry{
		requestTransformers:  make([]RequestTransformer, 0),
		responseTransformers: make([]ResponseTransformer, 0),
		streamTransformers:   make([]StreamTransformer, 0),
		metadataPlugins:      make([]MetadataPlugin, 0),
	}
}

// RegisterRequestTransformer adds a request transformer plugin
func (r *Registry) RegisterRequestTransformer(plugin RequestTransformer) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.requestTransformers = append(r.requestTransformers, plugin)
	r.sortRequestTransformers()
}

// RegisterResponseTransformer adds a response transformer plugin
func (r *Registry) RegisterResponseTransformer(plugin ResponseTransformer) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.responseTransformers = append(r.responseTransformers, plugin)
	r.sortResponseTransformers()
}

// RegisterStreamTransformer adds a stream transformer plugin
func (r *Registry) RegisterStreamTransformer(plugin StreamTransformer) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.streamTransformers = append(r.streamTransformers, plugin)
	r.sortStreamTransformers()
}

// RegisterMetadataPlugin adds a metadata plugin
func (r *Registry) RegisterMetadataPlugin(plugin MetadataPlugin) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.metadataPlugins = append(r.metadataPlugins, plugin)
}

// ApplyRequestTransformers applies all request transformers in order
func (r *Registry) ApplyRequestTransformers(ctx context.Context, request []byte) ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := request
	for _, plugin := range r.requestTransformers {
		if !plugin.Enabled(ctx) {
			continue
		}

		transformed, err := plugin.TransformRequest(ctx, result)
		if err != nil {
			return nil, fmt.Errorf("plugin %s failed: %w", plugin.Name(), err)
		}

		result = transformed
	}

	return result, nil
}

// ApplyResponseTransformers applies all response transformers in order
func (r *Registry) ApplyResponseTransformers(ctx context.Context, response []byte) ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := response
	for _, plugin := range r.responseTransformers {
		if !plugin.Enabled(ctx) {
			continue
		}

		transformed, err := plugin.TransformResponse(ctx, result)
		if err != nil {
			return nil, fmt.Errorf("plugin %s failed: %w", plugin.Name(), err)
		}

		result = transformed
	}

	return result, nil
}

// ApplyStreamTransformers applies all stream transformers to a chunk
func (r *Registry) ApplyStreamTransformers(ctx context.Context, chunk []byte) ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := chunk
	for _, plugin := range r.streamTransformers {
		if !plugin.Enabled(ctx) {
			continue
		}

		transformed, err := plugin.TransformChunk(ctx, result)
		if err != nil {
			return nil, fmt.Errorf("plugin %s failed: %w", plugin.Name(), err)
		}

		result = transformed
	}

	return result, nil
}

// NotifyRequest notifies all metadata plugins about a request
func (r *Registry) NotifyRequest(ctx context.Context, metadata RequestMetadata) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, plugin := range r.metadataPlugins {
		if plugin.Enabled(ctx) {
			plugin.OnRequest(ctx, metadata)
		}
	}
}

// NotifyResponse notifies all metadata plugins about a response
func (r *Registry) NotifyResponse(ctx context.Context, metadata ResponseMetadata) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, plugin := range r.metadataPlugins {
		if plugin.Enabled(ctx) {
			plugin.OnResponse(ctx, metadata)
		}
	}
}

// ListPlugins returns information about all registered plugins
func (r *Registry) ListPlugins() map[string][]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string][]string)

	result["request_transformers"] = make([]string, 0, len(r.requestTransformers))
	for _, p := range r.requestTransformers {
		result["request_transformers"] = append(result["request_transformers"],
			fmt.Sprintf("%s - %s (priority: %d)", p.Name(), p.Description(), p.Priority()))
	}

	result["response_transformers"] = make([]string, 0, len(r.responseTransformers))
	for _, p := range r.responseTransformers {
		result["response_transformers"] = append(result["response_transformers"],
			fmt.Sprintf("%s - %s (priority: %d)", p.Name(), p.Description(), p.Priority()))
	}

	result["stream_transformers"] = make([]string, 0, len(r.streamTransformers))
	for _, p := range r.streamTransformers {
		result["stream_transformers"] = append(result["stream_transformers"],
			fmt.Sprintf("%s - %s (priority: %d)", p.Name(), p.Description(), p.Priority()))
	}

	result["metadata_plugins"] = make([]string, 0, len(r.metadataPlugins))
	for _, p := range r.metadataPlugins {
		result["metadata_plugins"] = append(result["metadata_plugins"],
			fmt.Sprintf("%s - %s", p.Name(), p.Description()))
	}

	return result
}

// Helper functions to sort plugins by priority

func (r *Registry) sortRequestTransformers() {
	sort.Slice(r.requestTransformers, func(i, j int) bool {
		return r.requestTransformers[i].Priority() < r.requestTransformers[j].Priority()
	})
}

func (r *Registry) sortResponseTransformers() {
	sort.Slice(r.responseTransformers, func(i, j int) bool {
		return r.responseTransformers[i].Priority() < r.responseTransformers[j].Priority()
	})
}

func (r *Registry) sortStreamTransformers() {
	sort.Slice(r.streamTransformers, func(i, j int) bool {
		return r.streamTransformers[i].Priority() < r.streamTransformers[j].Priority()
	})
}
