package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/Davincible/claude-code-open/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `Manage the LLM proxy router configuration.`,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration interactively",
	Long:  `Initialize configuration by prompting for provider details.`,
	RunE:  runConfigInit,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Display the current configuration.`,
	RunE:  runConfigShow,
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration",
	Long:  `Validate the current configuration for errors.`,
	RunE:  runConfigValidate,
}

var configGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate example YAML configuration",
	Long:  `Generate an example YAML configuration file with all available providers.`,
	RunE:  runConfigGenerate,
}

func init() {
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configGenerateCmd)

	// Add flags for generate command
	configGenerateCmd.Flags().BoolP("force", "f", false, "Overwrite existing configuration file")
}

func runConfigInit(cmd *cobra.Command, _ []string) error {
	// Welcome message
	color.Blue("═══════════════════════════════════════════════════════════")
	color.Blue("  🤖 CCO Configuration Wizard")
	color.Blue("═══════════════════════════════════════════════════════════")
	fmt.Println()
	color.Yellow("This wizard will help you set up your CCO configuration.")
	color.Cyan("You can add one or more providers interactively.\n")

	reader := bufio.NewReader(os.Stdin)
	var providers []config.Provider

	// Show available providers
	color.Green("Available providers:")
	fmt.Println("  1. OpenRouter   - Access to multiple models (Claude, GPT, etc.)")
	fmt.Println("  2. OpenAI       - GPT-4, GPT-3.5, and other OpenAI models")
	fmt.Println("  3. Anthropic    - Claude models directly")
	fmt.Println("  4. Ollama       - Local models (no API key needed)")
	fmt.Println("  5. DeepSeek     - Coding-focused models")
	fmt.Println("  6. Groq         - Ultra-fast inference")
	fmt.Println("  7. NVIDIA       - Nemotron models")
	fmt.Println("  8. Gemini       - Google's Gemini models")
	fmt.Println("  9. Custom       - Your own provider")
	fmt.Println()

	addMore := true
	for addMore {
		// Get provider selection
		fmt.Print("Select provider (1-9): ")
		selection, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error reading selection: %w", err)
		}
		selection = strings.TrimSpace(selection)

		provider := config.Provider{}

		switch selection {
		case "1":
			provider = promptProviderDetails(reader, "openrouter", config.DefaultProviderURLs["openrouter"])
		case "2":
			provider = promptProviderDetails(reader, "openai", config.DefaultProviderURLs["openai"])
		case "3":
			provider = promptProviderDetails(reader, "anthropic", config.DefaultProviderURLs["anthropic"])
		case "4":
			provider = promptProviderDetails(reader, "ollama", config.DefaultProviderURLs["ollama"])
		case "5":
			provider = promptProviderDetails(reader, "deepseek", config.DefaultProviderURLs["deepseek"])
		case "6":
			provider = promptProviderDetails(reader, "groq", config.DefaultProviderURLs["groq"])
		case "7":
			provider = promptProviderDetails(reader, "nvidia", config.DefaultProviderURLs["nvidia"])
		case "8":
			provider = promptProviderDetails(reader, "gemini", config.DefaultProviderURLs["gemini"])
		case "9":
			provider = promptCustomProvider(reader)
		default:
			color.Red("Invalid selection. Please choose 1-9.")
			continue
		}

		if provider.Name != "" {
			providers = append(providers, provider)
			color.Green("✓ Added provider: %s\n", provider.Name)
		}

		// Ask if user wants to add more
		fmt.Print("Add another provider? (y/N): ")
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error reading response: %w", err)
		}
		response = strings.ToLower(strings.TrimSpace(response))
		addMore = response == "y" || response == "yes"
	}

	if len(providers) == 0 {
		color.Yellow("No providers configured. Exiting.")
		return nil
	}

	// Optional router API key
	fmt.Print("\nRouter API Key (optional, press Enter to skip): ")
	routerAPIKey, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("error reading router API key: %w", err)
	}
	routerAPIKey = strings.TrimSpace(routerAPIKey)

	// Set default model
	defaultModel := fmt.Sprintf("%s/%s", providers[0].Name, providers[0].DefaultModels[0])
	fmt.Printf("\nDefault model (press Enter for %s): ", defaultModel)
	customDefault, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("error reading default model: %w", err)
	}
	customDefault = strings.TrimSpace(customDefault)
	if customDefault != "" {
		defaultModel = customDefault
	}

	// Create configuration
	cfg := &config.Config{
		Host:      config.DefaultHost,
		Port:      config.DefaultPort,
		APIKey:    routerAPIKey,
		Providers: providers,
		Router: config.RouterConfig{
			Default: defaultModel,
		},
	}

	// Apply defaults
	cfgMgr.ApplyDefaults(cfg)

	// Show summary
	fmt.Println()
	color.Blue("═══════════════════════════════════════════════════════════")
	color.Blue("  Configuration Summary")
	color.Blue("═══════════════════════════════════════════════════════════")
	fmt.Printf("  Host: %s\n", cfg.Host)
	fmt.Printf("  Port: %d\n", cfg.Port)
	fmt.Printf("  Providers: %d\n", len(cfg.Providers))
	for _, p := range cfg.Providers {
		fmt.Printf("    - %s (%s)\n", p.Name, p.APIBase)
	}
	fmt.Printf("  Default Model: %s\n", cfg.Router.Default)
	fmt.Println()

	// Confirm save
	fmt.Print("Save this configuration? (Y/n): ")
	confirm, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("error reading confirmation: %w", err)
	}
	confirm = strings.ToLower(strings.TrimSpace(confirm))

	if confirm == "n" || confirm == "no" {
		color.Yellow("Configuration not saved.")
		return nil
	}

	// Save configuration
	if err := cfgMgr.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	color.Green("\n✓ Configuration saved successfully!")
	color.Cyan("\nNext steps:")
	fmt.Println("  1. Run 'cco config show' to view your configuration")
	fmt.Println("  2. Run 'cco start' to start the proxy server")
	fmt.Println("  3. Run 'eval \"$(cco activate)\"' to set up shell environment")

	return nil
}

func promptProviderDetails(reader *bufio.Reader, name, defaultURL string) config.Provider {
	color.Cyan("\nConfiguring %s", strings.ToUpper(name))

	// Get API key
	var apiKey string
	if name == "ollama" {
		color.Yellow("Ollama doesn't require an API key (using 'ollama' as placeholder)")
		apiKey = "ollama"
	} else {
		fmt.Printf("API Key for %s: ", name)
		key, err := reader.ReadString('\n')
		if err != nil {
			color.Red("Error reading API key")
			return config.Provider{}
		}
		apiKey = strings.TrimSpace(key)

		if apiKey == "" {
			color.Yellow("Skipping %s (no API key provided)", name)
			return config.Provider{}
		}
	}

	// Use default URL
	fmt.Printf("API URL (press Enter for default: %s): ", defaultURL)
	url, err := reader.ReadString('\n')
	if err != nil {
		color.Red("Error reading URL")
		return config.Provider{}
	}
	url = strings.TrimSpace(url)
	if url == "" {
		url = defaultURL
	}

	return config.Provider{
		Name:          name,
		APIKey:        apiKey,
		APIBase:       url,
		DefaultModels: config.DefaultProviderModels[name],
	}
}

func promptCustomProvider(reader *bufio.Reader) config.Provider {
	color.Cyan("\nConfiguring Custom Provider")

	fmt.Print("Provider Name: ")
	name, err := reader.ReadString('\n')
	if err != nil {
		color.Red("Error reading provider name")
		return config.Provider{}
	}
	name = strings.TrimSpace(name)

	if name == "" {
		color.Yellow("Skipping custom provider (no name provided)")
		return config.Provider{}
	}

	fmt.Print("API Key: ")
	apiKey, err := reader.ReadString('\n')
	if err != nil {
		color.Red("Error reading API key")
		return config.Provider{}
	}
	apiKey = strings.TrimSpace(apiKey)

	fmt.Print("API Base URL: ")
	url, err := reader.ReadString('\n')
	if err != nil {
		color.Red("Error reading URL")
		return config.Provider{}
	}
	url = strings.TrimSpace(url)

	fmt.Print("Default Model (optional): ")
	model, err := reader.ReadString('\n')
	if err != nil {
		color.Red("Error reading model")
		return config.Provider{}
	}
	model = strings.TrimSpace(model)

	var models []string
	if model != "" {
		models = []string{model}
	}

	return config.Provider{
		Name:          name,
		APIKey:        apiKey,
		APIBase:       url,
		DefaultModels: models,
	}
}

func runConfigShow(cmd *cobra.Command, _ []string) error {
	if !cfgMgr.Exists() {
		color.Yellow("No configuration found. Run 'cco config init' or 'cco config generate' to create one.")
		return nil
	}

	cfg, err := cfgMgr.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	color.Blue("Current Configuration:")
	fmt.Printf("  %-15s: %s\n", "Host", cfg.Host)
	fmt.Printf("  %-15s: %d\n", "Port", cfg.Port)
	fmt.Printf("  %-15s: %s\n", "API Key", maskString(cfg.APIKey))
	fmt.Printf("  %-15s: %s\n", "Config Path", cfgMgr.GetPath())

	// Show config file type
	configType := "JSON"
	if cfgMgr.HasYAML() {
		configType = "YAML"
	}

	fmt.Printf("  %-15s: %s\n", "Format", configType)

	fmt.Println("\nProviders:")

	for _, provider := range cfg.Providers {
		fmt.Printf("  - Name: %s\n", provider.Name)
		fmt.Printf("    URL: %s\n", provider.APIBase)
		fmt.Printf("    API Key: %s\n", maskString(provider.APIKey.(string)))

		if len(provider.DefaultModels) > 0 {
			fmt.Printf("    Default Models: %v\n", provider.DefaultModels)
		}

		if len(provider.ModelWhitelist) > 0 {
			fmt.Printf("    Model Whitelist: %v\n", provider.ModelWhitelist)
		}

		if len(provider.Models) > 0 {
			fmt.Printf("    Models: %v\n", provider.Models)
		}

		fmt.Println()
	}

	fmt.Println("Router Configuration:")
	fmt.Printf("  %-15s: %s\n", "Default", cfg.Router.Default)

	if cfg.Router.Think != "" {
		fmt.Printf("  %-15s: %s\n", "Think", cfg.Router.Think)
	}

	if cfg.Router.Background != "" {
		fmt.Printf("  %-15s: %s\n", "Background", cfg.Router.Background)
	}

	if cfg.Router.LongContext != "" {
		fmt.Printf("  %-15s: %s\n", "Long Context", cfg.Router.LongContext)
	}

	if cfg.Router.WebSearch != "" {
		fmt.Printf("  %-15s: %s\n", "Web Search", cfg.Router.WebSearch)
	}

	return nil
}

func runConfigValidate(cmd *cobra.Command, _ []string) error {
	if !cfgMgr.Exists() {
		return errors.New("no configuration found")
	}

	cfg, err := cfgMgr.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validation logic
	var validationErrors []string

	if len(cfg.Providers) == 0 {
		validationErrors = append(validationErrors, "no providers configured")
	}

	for i, provider := range cfg.Providers {
		if provider.Name == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("provider %d: name is required", i))
		}

		if provider.APIBase == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("provider %d: API base URL is required", i))
		}

		if provider.APIKey == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("provider %d: API key is required", i))
		}
	}

	if cfg.Router.Default == "" {
		validationErrors = append(validationErrors, "default router model is required")
	}

	if len(validationErrors) > 0 {
		color.Red("Configuration validation failed:")

		for _, err := range validationErrors {
			fmt.Printf("  - %s\n", err)
		}

		return errors.New("configuration validation failed")
	}

	color.Green("Configuration is valid!")

	return nil
}

func runConfigGenerate(cmd *cobra.Command, _ []string) error {
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return err
	}

	// Check if config already exists
	if cfgMgr.Exists() && !force {
		configType := "JSON"
		if cfgMgr.HasYAML() {
			configType = "YAML"
		}

		color.Yellow("Configuration file already exists (%s format): %s", configType, cfgMgr.GetPath())
		color.Cyan("Use --force to overwrite, or 'cco config show' to view current config")

		return nil
	}

	// Generate example YAML config
	if err := cfgMgr.CreateExampleYAML(); err != nil {
		return fmt.Errorf("failed to create example configuration: %w", err)
	}

	color.Green("Example YAML configuration created: %s", cfgMgr.GetYAMLPath())
	color.Cyan("\nNext steps:")
	fmt.Println("1. Edit the configuration file to add your API keys")
	fmt.Println("2. Customize provider settings and model whitelists as needed")
	fmt.Println("3. Run 'cco config validate' to check your configuration")
	fmt.Println("4. Start the router with 'cco start'")

	color.Yellow("\nNote: The configuration includes all 8 supported providers:")
	fmt.Println("- OpenRouter (access to multiple models)")
	fmt.Println("- OpenAI (GPT models)")
	fmt.Println("- Anthropic (Claude models)")
	fmt.Println("- Nvidia (Nemotron models)")
	fmt.Println("- Google Gemini (Gemini models)")
	fmt.Println("- Ollama (local models)")
	fmt.Println("- DeepSeek (coding-focused models)")
	fmt.Println("- Groq (ultra-fast inference)")

	return nil
}

func maskString(s string) string {
	if s == "" {
		return "(not set)"
	}

	if len(s) <= 8 {
		return strings.Repeat("*", len(s))
	}

	return s[:4] + strings.Repeat("*", len(s)-8) + s[len(s)-4:]
}
