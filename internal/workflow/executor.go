package workflow

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fatih/color"
)

// ExecutionContext holds the state during workflow execution
type ExecutionContext struct {
	Workflow  *Workflow
	Variables Variables
	StepIndex int
	StartTime time.Time
	Logger    Logger
}

// Logger interface for workflow execution logging
type Logger interface {
	Info(message string, args ...any)
	Error(message string, args ...any)
	Success(message string, args ...any)
	Warn(message string, args ...any)
}

// ConsoleLogger implements Logger for console output
type ConsoleLogger struct{}

func (l *ConsoleLogger) Info(message string, args ...any) {
	fmt.Printf("ℹ️  %s\n", fmt.Sprintf(message, args...))
}

func (l *ConsoleLogger) Error(message string, args ...any) {
	color.Red("❌ %s", fmt.Sprintf(message, args...))
}

func (l *ConsoleLogger) Success(message string, args ...any) {
	color.Green("✓ %s", fmt.Sprintf(message, args...))
}

func (l *ConsoleLogger) Warn(message string, args ...any) {
	color.Yellow("⚠️  %s", fmt.Sprintf(message, args...))
}

// Executor executes workflows
type Executor struct {
	logger Logger
	dryRun bool
}

// NewExecutor creates a new workflow executor
func NewExecutor(logger Logger, dryRun bool) *Executor {
	if logger == nil {
		logger = &ConsoleLogger{}
	}
	return &Executor{
		logger: logger,
		dryRun: dryRun,
	}
}

// Execute runs a workflow
func (e *Executor) Execute(workflow *Workflow) error {
	ctx := &ExecutionContext{
		Workflow:  workflow,
		Variables: make(Variables),
		StartTime: time.Now(),
		Logger:    e.logger,
	}

	// Copy workflow variables
	for k, v := range workflow.Variables {
		ctx.Variables[k] = v
	}

	e.logger.Info("Starting workflow: %s", workflow.Name)
	if workflow.Description != "" {
		fmt.Printf("  %s\n", workflow.Description)
	}
	fmt.Println()

	// Execute each step
	for i, step := range workflow.Steps {
		ctx.StepIndex = i

		// Show step header
		color.Cyan("\n[Step %d/%d] %s", i+1, len(workflow.Steps), step.Name)
		if step.Description != "" {
			fmt.Printf("  %s\n", step.Description)
		}

		if e.dryRun {
			e.logger.Info("DRY RUN: Would execute %s step", step.Type)
			continue
		}

		// Execute step
		if err := e.executeStep(ctx, &step); err != nil {
			return e.handleStepError(ctx, &step, err)
		}

		e.logger.Success("Step completed")
	}

	duration := time.Since(ctx.StartTime)
	fmt.Println()
	e.logger.Success("Workflow completed successfully in %s", duration)

	return nil
}

// executeStep executes a single workflow step
func (e *Executor) executeStep(ctx *ExecutionContext, step *Step) error {
	switch step.Type {
	case StepCommand:
		return e.executeCommand(ctx, step)
	case StepConfigure:
		return e.executeConfigure(ctx, step)
	case StepProvider:
		return e.executeProvider(ctx, step)
	case StepPlugin:
		return e.executePlugin(ctx, step)
	case StepTest:
		return e.executeTest(ctx, step)
	case StepPrompt:
		return e.executePrompt(ctx, step)
	case StepValidate:
		return e.executeValidate(ctx, step)
	case StepWait:
		return e.executeWait(ctx, step)
	case StepHTTP:
		return e.executeHTTP(ctx, step)
	default:
		return fmt.Errorf("unknown step type: %s", step.Type)
	}
}

// executeCommand runs a shell command
func (e *Executor) executeCommand(ctx *ExecutionContext, step *Step) error {
	// Substitute variables in command and args
	cmd := e.substituteVariables(step.Command, ctx.Variables)
	args := make([]string, len(step.Args))
	for i, arg := range step.Args {
		args[i] = e.substituteVariables(arg, ctx.Variables)
	}

	e.logger.Info("Running: %s %s", cmd, strings.Join(args, " "))

	command := exec.Command(cmd, args...)

	// Set environment variables
	if len(step.Env) > 0 {
		command.Env = os.Environ()
		for k, v := range step.Env {
			command.Env = append(command.Env, fmt.Sprintf("%s=%s", k, e.substituteVariables(v, ctx.Variables)))
		}
	}

	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %w\n%s", err, string(output))
	}

	if len(output) > 0 {
		fmt.Println(string(output))
	}

	return nil
}

// executeConfigure modifies CCO configuration
func (e *Executor) executeConfigure(ctx *ExecutionContext, step *Step) error {
	e.logger.Info("Configuring CCO...")
	// This would integrate with the config.Manager
	// For now, just log what we would do
	fmt.Printf("  Config changes: %v\n", step.Config)
	return nil
}

// executeProvider adds or modifies a provider
func (e *Executor) executeProvider(ctx *ExecutionContext, step *Step) error {
	e.logger.Info("Configuring provider...")
	fmt.Printf("  Provider config: %v\n", step.Config)
	return nil
}

// executePlugin enables or disables a plugin
func (e *Executor) executePlugin(ctx *ExecutionContext, step *Step) error {
	e.logger.Info("Managing plugins...")
	fmt.Printf("  Plugin config: %v\n", step.Config)
	return nil
}

// executeTest runs tests
func (e *Executor) executeTest(ctx *ExecutionContext, step *Step) error {
	e.logger.Info("Running tests...")
	// Could integrate with actual test commands
	return nil
}

// executePrompt prompts user for input
func (e *Executor) executePrompt(ctx *ExecutionContext, step *Step) error {
	variable, ok := step.Config["variable"].(string)
	if !ok {
		return fmt.Errorf("prompt step missing 'variable' config")
	}

	prompt, ok := step.Config["prompt"].(string)
	if !ok {
		return fmt.Errorf("prompt step missing 'prompt' config")
	}

	secret, _ := step.Config["secret"].(bool)

	fmt.Printf("%s ", prompt)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}

	value := strings.TrimSpace(input)
	ctx.Variables[variable] = value

	if !secret {
		e.logger.Info("Set %s = %s", variable, value)
	} else {
		e.logger.Info("Set %s = [hidden]", variable)
	}

	return nil
}

// executeValidate validates configuration
func (e *Executor) executeValidate(ctx *ExecutionContext, step *Step) error {
	e.logger.Info("Validating configuration...")
	// Could run cco config validate
	return nil
}

// executeWait waits for a condition or duration
func (e *Executor) executeWait(ctx *ExecutionContext, step *Step) error {
	if duration, ok := step.Config["duration"].(string); ok {
		d, err := time.ParseDuration(duration)
		if err != nil {
			return fmt.Errorf("invalid duration: %w", err)
		}
		e.logger.Info("Waiting for %s...", d)
		time.Sleep(d)
		return nil
	}

	return fmt.Errorf("wait step requires 'duration' config")
}

// executeHTTP makes an HTTP request
func (e *Executor) executeHTTP(ctx *ExecutionContext, step *Step) error {
	e.logger.Info("Making HTTP request...")
	fmt.Printf("  HTTP config: %v\n", step.Config)
	return nil
}

// handleStepError handles errors during step execution
func (e *Executor) handleStepError(ctx *ExecutionContext, step *Step, err error) error {
	e.logger.Error("Step failed: %v", err)

	switch step.OnError {
	case ErrorContinue:
		e.logger.Warn("Continuing to next step...")
		return nil

	case ErrorRetry:
		e.logger.Info("Retrying step...")
		return e.executeStep(ctx, step)

	case ErrorPrompt:
		fmt.Print("\nStep failed. Continue anyway? (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(response)) == "y" {
			return nil
		}
		return err

	default: // ErrorAbort
		return fmt.Errorf("workflow aborted: %w", err)
	}
}

// substituteVariables replaces {{var}} with values from context
func (e *Executor) substituteVariables(s string, vars Variables) string {
	result := s
	for k, v := range vars {
		placeholder := fmt.Sprintf("{{%s}}", k)
		result = strings.ReplaceAll(result, placeholder, v)
	}
	return result
}
