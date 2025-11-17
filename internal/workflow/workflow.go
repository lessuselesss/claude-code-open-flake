package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Workflow represents a multi-step operation plan
type Workflow struct {
	Name        string    `json:"name" yaml:"name"`
	Description string    `json:"description" yaml:"description"`
	Version     string    `json:"version" yaml:"version"`
	Author      string    `json:"author,omitempty" yaml:"author,omitempty"`
	Created     time.Time `json:"created" yaml:"created"`
	Steps       []Step    `json:"steps" yaml:"steps"`
	Variables   Variables `json:"variables,omitempty" yaml:"variables,omitempty"`
}

// Step represents a single action in a workflow
type Step struct {
	Name        string            `json:"name" yaml:"name"`
	Type        StepType          `json:"type" yaml:"type"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	Command     string            `json:"command,omitempty" yaml:"command,omitempty"`
	Args        []string          `json:"args,omitempty" yaml:"args,omitempty"`
	Env         map[string]string `json:"env,omitempty" yaml:"env,omitempty"`
	Config      map[string]any    `json:"config,omitempty" yaml:"config,omitempty"`
	OnError     ErrorAction       `json:"on_error,omitempty" yaml:"on_error,omitempty"`
	Condition   string            `json:"condition,omitempty" yaml:"condition,omitempty"`
}

// StepType defines the type of workflow step
type StepType string

const (
	StepCommand      StepType = "command"       // Execute shell command
	StepConfigure    StepType = "configure"     // Modify CCO config
	StepProvider     StepType = "provider"      // Add/modify provider
	StepPlugin       StepType = "plugin"        // Enable/disable plugin
	StepTest         StepType = "test"          // Run tests
	StepPrompt       StepType = "prompt"        // Prompt user for input
	StepWait         StepType = "wait"          // Wait for condition
	StepHTTP         StepType = "http"          // HTTP request
	StepValidate     StepType = "validate"      // Validate configuration
)

// ErrorAction defines what to do on step failure
type ErrorAction string

const (
	ErrorContinue ErrorAction = "continue" // Continue to next step
	ErrorAbort    ErrorAction = "abort"    // Abort workflow
	ErrorRetry    ErrorAction = "retry"    // Retry the step
	ErrorPrompt   ErrorAction = "prompt"   // Ask user what to do
)

// Variables holds workflow variables
type Variables map[string]string

// Manager handles workflow operations
type Manager struct {
	workflowDir string
}

// NewManager creates a new workflow manager
func NewManager(baseDir string) *Manager {
	return &Manager{
		workflowDir: filepath.Join(baseDir, "workflows"),
	}
}

// List returns all available workflows
func (m *Manager) List() ([]Workflow, error) {
	if err := os.MkdirAll(m.workflowDir, 0750); err != nil {
		return nil, fmt.Errorf("create workflow directory: %w", err)
	}

	entries, err := os.ReadDir(m.workflowDir)
	if err != nil {
		return nil, fmt.Errorf("read workflow directory: %w", err)
	}

	var workflows []Workflow
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		workflow, err := m.Load(entry.Name()[:len(entry.Name())-5])
		if err != nil {
			continue // Skip invalid workflows
		}

		workflows = append(workflows, *workflow)
	}

	return workflows, nil
}

// Load loads a workflow by name
func (m *Manager) Load(name string) (*Workflow, error) {
	path := filepath.Join(m.workflowDir, name+".json")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read workflow file: %w", err)
	}

	var workflow Workflow
	if err := json.Unmarshal(data, &workflow); err != nil {
		return nil, fmt.Errorf("parse workflow: %w", err)
	}

	return &workflow, nil
}

// Save saves a workflow
func (m *Manager) Save(workflow *Workflow) error {
	if err := os.MkdirAll(m.workflowDir, 0750); err != nil {
		return fmt.Errorf("create workflow directory: %w", err)
	}

	path := filepath.Join(m.workflowDir, workflow.Name+".json")

	data, err := json.MarshalIndent(workflow, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal workflow: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write workflow file: %w", err)
	}

	return nil
}

// Delete removes a workflow
func (m *Manager) Delete(name string) error {
	path := filepath.Join(m.workflowDir, name+".json")
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("delete workflow: %w", err)
	}
	return nil
}

// CreateDefault creates default workflow templates
func (m *Manager) CreateDefault() error {
	defaults := []Workflow{
		{
			Name:        "setup-ollama",
			Description: "Set up Ollama provider for local models",
			Version:     "1.0.0",
			Created:     time.Now(),
			Steps: []Step{
				{
					Name:        "Check Ollama Installation",
					Type:        StepCommand,
					Description: "Verify Ollama is installed",
					Command:     "which",
					Args:        []string{"ollama"},
					OnError:     ErrorAbort,
				},
				{
					Name:        "Start Ollama Service",
					Type:        StepCommand,
					Description: "Ensure Ollama service is running",
					Command:     "ollama",
					Args:        []string{"serve"},
					OnError:     ErrorContinue,
				},
				{
					Name:        "Configure Ollama Provider",
					Type:        StepProvider,
					Description: "Add Ollama to CCO configuration",
					Config: map[string]any{
						"name":    "ollama",
						"url":     "http://localhost:11434/v1/chat/completions",
						"api_key": "ollama",
					},
				},
				{
					Name:        "Test Ollama Connection",
					Type:        StepTest,
					Description: "Verify Ollama is responding",
				},
			},
		},
		{
			Name:        "production-setup",
			Description: "Configure CCO for production deployment",
			Version:     "1.0.0",
			Created:     time.Now(),
			Steps: []Step{
				{
					Name:        "Generate Strong API Key",
					Type:        StepCommand,
					Description: "Create secure router API key",
					Command:     "openssl",
					Args:        []string{"rand", "-hex", "32"},
				},
				{
					Name:        "Enable Security Plugins",
					Type:        StepPlugin,
					Description: "Enable production security features",
					Config: map[string]any{
						"token_counter": true,
						"rate_limiter":  true,
					},
				},
				{
					Name:        "Validate Configuration",
					Type:        StepValidate,
					Description: "Ensure configuration is production-ready",
				},
				{
					Name:        "Start CCO Server",
					Type:        StepCommand,
					Description: "Launch CCO in production mode",
					Command:     "cco",
					Args:        []string{"start", "--verbose"},
				},
			},
		},
		{
			Name:        "add-provider",
			Description: "Interactive workflow to add a new provider",
			Version:     "1.0.0",
			Created:     time.Now(),
			Variables: Variables{
				"provider_name": "",
				"api_key":       "",
				"api_url":       "",
				"model":         "",
			},
			Steps: []Step{
				{
					Name:        "Get Provider Name",
					Type:        StepPrompt,
					Description: "Ask user for provider name",
					Config: map[string]any{
						"variable": "provider_name",
						"prompt":   "Provider name (e.g., openai, anthropic):",
					},
				},
				{
					Name:        "Get API Key",
					Type:        StepPrompt,
					Description: "Ask user for API key",
					Config: map[string]any{
						"variable": "api_key",
						"prompt":   "API Key:",
						"secret":   true,
					},
				},
				{
					Name:        "Get API URL",
					Type:        StepPrompt,
					Description: "Ask user for API base URL",
					Config: map[string]any{
						"variable": "api_url",
						"prompt":   "API Base URL:",
					},
				},
				{
					Name:        "Get Default Model",
					Type:        StepPrompt,
					Description: "Ask user for default model",
					Config: map[string]any{
						"variable": "model",
						"prompt":   "Default model name:",
					},
				},
				{
					Name:        "Add Provider",
					Type:        StepConfigure,
					Description: "Add provider to configuration",
					Config: map[string]any{
						"action": "add_provider",
					},
				},
				{
					Name:        "Test Provider",
					Type:        StepTest,
					Description: "Verify provider connection",
				},
			},
		},
		{
			Name:        "backup-config",
			Description: "Backup current CCO configuration",
			Version:     "1.0.0",
			Created:     time.Now(),
			Steps: []Step{
				{
					Name:        "Create Backup Directory",
					Type:        StepCommand,
					Description: "Ensure backup directory exists",
					Command:     "mkdir",
					Args:        []string{"-p", "~/.cco-backups"},
				},
				{
					Name:        "Backup Configuration",
					Type:        StepCommand,
					Description: "Copy config to backup location",
					Command:     "cp",
					Args: []string{
						"~/.claude-code-open/config.yaml",
						"~/.cco-backups/config-$(date +%Y%m%d-%H%M%S).yaml",
					},
				},
			},
		},
		{
			Name:        "health-check",
			Description: "Run comprehensive health check on CCO setup",
			Version:     "1.0.0",
			Created:     time.Now(),
			Steps: []Step{
				{
					Name:        "Check CCO Status",
					Type:        StepCommand,
					Description: "Verify CCO server is running",
					Command:     "cco",
					Args:        []string{"status"},
				},
				{
					Name:        "Validate Configuration",
					Type:        StepValidate,
					Description: "Check configuration validity",
				},
				{
					Name:        "Test Each Provider",
					Type:        StepTest,
					Description: "Send test request to each provider",
				},
				{
					Name:        "Check Plugin Status",
					Type:        StepCommand,
					Description: "List active plugins",
					Command:     "cco",
					Args:        []string{"plugins", "list"},
				},
			},
		},
	}

	for _, workflow := range defaults {
		if err := m.Save(&workflow); err != nil {
			return fmt.Errorf("save default workflow %s: %w", workflow.Name, err)
		}
	}

	return nil
}
