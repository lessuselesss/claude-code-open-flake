package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/Davincible/claude-code-open/internal/plugins"
	"github.com/Davincible/claude-code-open/internal/plugins/builtin"
	"github.com/Davincible/claude-code-open/internal/webui"
)

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Start the web UI dashboard",
	Long: `Start a web-based dashboard for managing CCO configuration,
viewing statistics, and monitoring requests.`,
	RunE: runUI,
}

var (
	uiPort int
	uiHost string
)

func init() {
	uiCmd.Flags().IntVarP(&uiPort, "port", "p", 6971, "Port for web UI server")
	uiCmd.Flags().StringVarP(&uiHost, "host", "H", "127.0.0.1", "Host for web UI server")
}

func runUI(cmd *cobra.Command, _ []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	logFile, _ := cmd.Flags().GetBool("log-file")

	setupLogging(verbose, logFile)

	color.Blue("═══════════════════════════════════════════════════════════")
	color.Blue("  🎨 Starting CCO Web UI")
	color.Blue("═══════════════════════════════════════════════════════════")
	fmt.Println()

	// Load configuration
	cfg, err := cfgMgr.Load()
	if err != nil {
		color.Yellow("Warning: Could not load configuration: %v", err)
		color.Cyan("Using minimal configuration. Run 'cco config init' to set up providers.")
		// Create minimal config
		minimalCfg := cfgMgr.Get()
		cfg = minimalCfg
	}

	// Create plugin registry (even if empty)
	pluginRegistry := plugins.NewRegistry()

	// Initialize plugins from config
	if cfg.Plugins.TokenCounter {
		tokenCounter := builtin.NewTokenCounterPlugin(logger, true)
		pluginRegistry.RegisterMetadataPlugin(tokenCounter)
		logger.Info("Registered plugin", "name", "token-counter")
	}

	if cfg.Plugins.SystemPrompt != "" {
		systemPromptInjector := builtin.NewSystemPromptInjectorPlugin(cfg.Plugins.SystemPrompt, true)
		pluginRegistry.RegisterRequestTransformer(systemPromptInjector)
		logger.Info("Registered plugin", "name", "system-prompt-injector")
	}

	if cfg.Plugins.ResponseFilterEnabled && len(cfg.Plugins.FilterWords) > 0 {
		responseFilter := builtin.NewResponseFilterPlugin(
			cfg.Plugins.FilterWords,
			cfg.Plugins.FilterReplacement,
			true,
		)
		pluginRegistry.RegisterResponseTransformer(responseFilter)
		logger.Info("Registered plugin", "name", "response-filter")
	}

	// Create web UI server
	uiServer := webui.NewServer(cfgMgr, pluginRegistry, logger)

	addr := fmt.Sprintf("%s:%d", uiHost, uiPort)

	color.Green("\n✓ Web UI server starting")
	color.Cyan("\nAccess the dashboard:")
	fmt.Printf("  http://%s:%d\n", uiHost, uiPort)
	if uiHost == "127.0.0.1" || uiHost == "localhost" {
		fmt.Printf("  http://localhost:%d\n", uiPort)
	}

	color.Yellow("\nPress Ctrl+C to stop")
	fmt.Println()

	// Setup graceful shutdown
	srv := &http.Server{
		Addr:    addr,
		Handler: uiServer,
	}

	// Channel to listen for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		logger.Info("Web UI server listening", "address", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	<-stop

	color.Yellow("\n\nShutting down web UI server...")
	if err := srv.Close(); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	color.Green("✓ Web UI server stopped")

	return nil
}
