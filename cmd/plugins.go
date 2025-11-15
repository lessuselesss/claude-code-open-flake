package cmd

import (
	"fmt"

	"github.com/Davincible/claude-code-open/internal/config"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var pluginsCmd = &cobra.Command{
	Use:   "plugins",
	Short: "Manage request/response transformation plugins",
	Long:  `List, enable, and configure plugins that transform requests and responses.`,
}

var pluginsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available plugins",
	Long:  `Display all available plugins and their current status.`,
	RunE:  runPluginsList,
}

var pluginsEnableCmd = &cobra.Command{
	Use:   "enable [plugin-name]",
	Short: "Enable a plugin",
	Long:  `Enable a specific plugin by name.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runPluginsEnable,
}

var pluginsDisableCmd = &cobra.Command{
	Use:   "disable [plugin-name]",
	Short: "Disable a plugin",
	Long:  `Disable a specific plugin by name.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runPluginsDisable,
}

func init() {
	pluginsCmd.AddCommand(pluginsListCmd)
	pluginsCmd.AddCommand(pluginsEnableCmd)
	pluginsCmd.AddCommand(pluginsDisableCmd)
}

func runPluginsList(cmd *cobra.Command, _ []string) error {
	color.Blue("═══════════════════════════════════════════════════════════")
	color.Blue("  📦 Available Plugins")
	color.Blue("═══════════════════════════════════════════════════════════")
	fmt.Println()

	// Load current config to show status
	cfg, err := cfgMgr.Load()
	var pluginsConfig config.PluginsConfig
	if err != nil {
		color.Yellow("No configuration found. Showing defaults.")
		// Use empty plugins config
		pluginsConfig = config.PluginsConfig{}
	} else {
		pluginsConfig = cfg.Plugins
	}

	plugins := []struct {
		name        string
		description string
		enabled     bool
		configKey   string
	}{
		{
			name:        "token-counter",
			description: "Tracks and logs token usage statistics for each request",
			enabled:     pluginsConfig.TokenCounter,
			configKey:   "plugins.token_counter",
		},
		{
			name:        "system-prompt-injector",
			description: "Injects a custom system prompt into all requests",
			enabled:     pluginsConfig.SystemPrompt != "",
			configKey:   "plugins.system_prompt",
		},
		{
			name:        "response-filter",
			description: "Filters sensitive content from responses based on word lists",
			enabled:     pluginsConfig.ResponseFilterEnabled,
			configKey:   "plugins.response_filter_enabled",
		},
	}

	color.Cyan("Built-in Plugins:")
	fmt.Println()

	for _, p := range plugins {
		status := color.RedString("✗ Disabled")
		if p.enabled {
			status = color.GreenString("✓ Enabled")
		}

		fmt.Printf("  %s\n", color.New(color.Bold).Sprint(p.name))
		fmt.Printf("    Status: %s\n", status)
		fmt.Printf("    Description: %s\n", p.description)
		fmt.Printf("    Config: %s\n", color.CyanString(p.configKey))
		fmt.Println()
	}

	color.Yellow("\nPlugin Configuration:")
	fmt.Println()
	fmt.Println("Plugins can be configured in your config.yaml:")
	fmt.Println()

	color.Cyan("  plugins:")
	fmt.Println("    token_counter: true")
	fmt.Println("    system_prompt: \"You are a helpful assistant...\"")
	fmt.Println("    response_filter_enabled: false")
	fmt.Println("    filter_words:")
	fmt.Println("      - \"sensitive-word\"")
	fmt.Println("    filter_replacement: \"[FILTERED]\"")
	fmt.Println()

	color.Green("Use 'cco plugins enable/disable [name]' to toggle plugins")
	color.Cyan("Or edit your config.yaml directly for more control")

	return nil
}

func runPluginsEnable(cmd *cobra.Command, args []string) error {
	pluginName := args[0]

	cfg, err := cfgMgr.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	switch pluginName {
	case "token-counter":
		cfg.Plugins.TokenCounter = true
		color.Green("✓ Enabled token-counter plugin")
	case "system-prompt-injector":
		return fmt.Errorf("system-prompt-injector requires a system prompt. Edit config.yaml and set plugins.system_prompt")
	case "response-filter":
		cfg.Plugins.ResponseFilterEnabled = true
		color.Green("✓ Enabled response-filter plugin")
		if len(cfg.Plugins.FilterWords) == 0 {
			color.Yellow("⚠ Warning: response-filter is enabled but no filter_words are configured")
			color.Cyan("  Edit config.yaml and add filter_words to the plugins section")
		}
	default:
		return fmt.Errorf("unknown plugin: %s", pluginName)
	}

	if err := cfgMgr.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	color.Cyan("\nRestart the server for changes to take effect:")
	fmt.Println("  cco stop && cco start")

	return nil
}

func runPluginsDisable(cmd *cobra.Command, args []string) error {
	pluginName := args[0]

	cfg, err := cfgMgr.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	switch pluginName {
	case "token-counter":
		cfg.Plugins.TokenCounter = false
		color.Green("✓ Disabled token-counter plugin")
	case "system-prompt-injector":
		cfg.Plugins.SystemPrompt = ""
		color.Green("✓ Disabled system-prompt-injector plugin")
	case "response-filter":
		cfg.Plugins.ResponseFilterEnabled = false
		color.Green("✓ Disabled response-filter plugin")
	default:
		return fmt.Errorf("unknown plugin: %s", pluginName)
	}

	if err := cfgMgr.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	color.Cyan("\nRestart the server for changes to take effect:")
	fmt.Println("  cco stop && cco start")

	return nil
}
