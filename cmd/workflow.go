package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/Davincible/claude-code-open/internal/workflow"
)

var workflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Manage and execute workflows",
	Long: `Create, manage, and execute multi-step workflows for common operations.

Workflows are pre-defined plans that automate complex tasks like:
- Setting up new providers
- Configuring production environments
- Running health checks
- Backing up configurations`,
	Aliases: []string{"wf"},
}

var workflowListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available workflows",
	Long:  `Display all workflows including built-in templates and custom workflows.`,
	RunE:  runWorkflowList,
}

var workflowRunCmd = &cobra.Command{
	Use:   "run [workflow-name]",
	Short: "Execute a workflow",
	Long: `Execute a workflow by name.

Examples:
  cco workflow run setup-ollama
  cco workflow run production-setup
  cco workflow run add-provider`,
	Args: cobra.ExactArgs(1),
	RunE: runWorkflowRun,
}

var workflowShowCmd = &cobra.Command{
	Use:   "show [workflow-name]",
	Short: "Show workflow details",
	Long:  `Display the full details of a workflow including all steps.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runWorkflowShow,
}

var workflowInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize default workflow templates",
	Long:  `Create default workflow templates in ~/.claude-code-open/workflows/`,
	RunE:  runWorkflowInit,
}

var workflowDeleteCmd = &cobra.Command{
	Use:   "delete [workflow-name]",
	Short: "Delete a workflow",
	Long:  `Remove a workflow from the workflows directory.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runWorkflowDelete,
}

var (
	dryRun bool
)

func init() {
	workflowCmd.AddCommand(workflowListCmd)
	workflowCmd.AddCommand(workflowRunCmd)
	workflowCmd.AddCommand(workflowShowCmd)
	workflowCmd.AddCommand(workflowInitCmd)
	workflowCmd.AddCommand(workflowDeleteCmd)

	workflowRunCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be executed without running")
}

func runWorkflowList(cmd *cobra.Command, _ []string) error {
	wfMgr := workflow.NewManager(baseDir)

	workflows, err := wfMgr.List()
	if err != nil {
		return fmt.Errorf("failed to list workflows: %w", err)
	}

	if len(workflows) == 0 {
		color.Yellow("No workflows found.")
		fmt.Println()
		color.Cyan("Run 'cco workflow init' to create default workflow templates.")
		return nil
	}

	color.Blue("═══════════════════════════════════════════════════════════")
	color.Blue("  📋 Available Workflows")
	color.Blue("═══════════════════════════════════════════════════════════")
	fmt.Println()

	for _, wf := range workflows {
		color.New(color.Bold).Printf("  %s", wf.Name)
		if wf.Version != "" {
			fmt.Printf(" (v%s)", wf.Version)
		}
		fmt.Println()

		if wf.Description != "" {
			fmt.Printf("    %s\n", wf.Description)
		}

		fmt.Printf("    Steps: %d\n", len(wf.Steps))
		if wf.Author != "" {
			fmt.Printf("    Author: %s\n", wf.Author)
		}

		fmt.Println()
	}

	color.Green("Use 'cco workflow run <name>' to execute a workflow")
	color.Cyan("Use 'cco workflow show <name>' to see workflow details")

	return nil
}

func runWorkflowRun(cmd *cobra.Command, args []string) error {
	workflowName := args[0]

	wfMgr := workflow.NewManager(baseDir)

	// Load workflow
	wf, err := wfMgr.Load(workflowName)
	if err != nil {
		return fmt.Errorf("failed to load workflow '%s': %w", workflowName, err)
	}

	// Show workflow header
	color.Blue("═══════════════════════════════════════════════════════════")
	color.Blue("  🚀 Executing Workflow: %s", wf.Name)
	color.Blue("═══════════════════════════════════════════════════════════")
	fmt.Println()

	if dryRun {
		color.Yellow("DRY RUN MODE - No changes will be made")
		fmt.Println()
	}

	// Execute workflow
	executor := workflow.NewExecutor(nil, dryRun)
	if err := executor.Execute(wf); err != nil {
		return fmt.Errorf("workflow execution failed: %w", err)
	}

	return nil
}

func runWorkflowShow(cmd *cobra.Command, args []string) error {
	workflowName := args[0]

	wfMgr := workflow.NewManager(baseDir)

	wf, err := wfMgr.Load(workflowName)
	if err != nil {
		return fmt.Errorf("failed to load workflow '%s': %w", workflowName, err)
	}

	// Display workflow details
	color.Blue("═══════════════════════════════════════════════════════════")
	color.Blue("  📋 Workflow: %s", wf.Name)
	color.Blue("═══════════════════════════════════════════════════════════")
	fmt.Println()

	if wf.Description != "" {
		fmt.Printf("Description: %s\n", wf.Description)
	}

	if wf.Version != "" {
		fmt.Printf("Version: %s\n", wf.Version)
	}

	if wf.Author != "" {
		fmt.Printf("Author: %s\n", wf.Author)
	}

	fmt.Printf("Created: %s\n", wf.Created.Format("2006-01-02 15:04:05"))
	fmt.Println()

	// Display variables
	if len(wf.Variables) > 0 {
		color.Cyan("Variables:")
		for k, v := range wf.Variables {
			if v == "" {
				fmt.Printf("  %s: (will be prompted)\n", k)
			} else {
				fmt.Printf("  %s: %s\n", k, v)
			}
		}
		fmt.Println()
	}

	// Display steps
	color.Cyan("Steps:")
	for i, step := range wf.Steps {
		fmt.Printf("\n%d. %s", i+1, color.New(color.Bold).Sprint(step.Name))
		fmt.Printf(" [%s]\n", step.Type)

		if step.Description != "" {
			fmt.Printf("   %s\n", step.Description)
		}

		if step.Command != "" {
			fmt.Printf("   Command: %s %s\n", step.Command, strings.Join(step.Args, " "))
		}

		if len(step.Config) > 0 {
			fmt.Printf("   Config: %v\n", step.Config)
		}

		if step.OnError != "" {
			fmt.Printf("   On Error: %s\n", step.OnError)
		}
	}

	fmt.Println()
	color.Green("Use 'cco workflow run %s' to execute this workflow", wf.Name)

	return nil
}

func runWorkflowInit(cmd *cobra.Command, _ []string) error {
	color.Blue("═══════════════════════════════════════════════════════════")
	color.Blue("  🔧 Initializing Default Workflows")
	color.Blue("═══════════════════════════════════════════════════════════")
	fmt.Println()

	wfMgr := workflow.NewManager(baseDir)

	if err := wfMgr.CreateDefault(); err != nil {
		return fmt.Errorf("failed to create default workflows: %w", err)
	}

	workflows, err := wfMgr.List()
	if err != nil {
		return fmt.Errorf("failed to list workflows: %w", err)
	}

	color.Green("✓ Created %d default workflow templates:", len(workflows))
	fmt.Println()

	for _, wf := range workflows {
		fmt.Printf("  • %s\n", color.New(color.Bold).Sprint(wf.Name))
		if wf.Description != "" {
			fmt.Printf("    %s\n", wf.Description)
		}
	}

	fmt.Println()
	color.Cyan("Workflows location: %s/workflows/", baseDir)
	color.Green("\nRun 'cco workflow list' to see all workflows")
	color.Green("Run 'cco workflow run <name>' to execute a workflow")

	return nil
}

func runWorkflowDelete(cmd *cobra.Command, args []string) error {
	workflowName := args[0]

	wfMgr := workflow.NewManager(baseDir)

	// Confirm deletion
	fmt.Printf("Delete workflow '%s'? (y/N): ", workflowName)
	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) != "y" {
		color.Yellow("Deletion cancelled")
		return nil
	}

	if err := wfMgr.Delete(workflowName); err != nil {
		return fmt.Errorf("failed to delete workflow: %w", err)
	}

	color.Green("✓ Workflow '%s' deleted successfully", workflowName)

	return nil
}
