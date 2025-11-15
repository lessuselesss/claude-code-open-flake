package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/Davincible/claude-code-open/internal/config"
)

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "Interactive model browser",
	Long:  `Browse available models from all configured providers interactively.`,
	RunE:  runModels,
}

var modelsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available models",
	Long:  `List all available models from configured providers in plain text.`,
	RunE:  runModelsList,
}

func init() {
	modelsCmd.AddCommand(modelsListCmd)

	// Add flag to list models for a specific provider
	modelsListCmd.Flags().StringP("provider", "p", "", "Filter models by provider name")
}

func runModels(cmd *cobra.Command, _ []string) error {
	// Load configuration
	cfg, err := cfgMgr.Load()
	if err != nil {
		color.Yellow("No configuration found. Using default provider list.")
		cfg = &config.Config{
			Providers: []config.Provider{
				{Name: "openrouter", DefaultModels: config.DefaultProviderModels["openrouter"]},
				{Name: "openai", DefaultModels: config.DefaultProviderModels["openai"]},
				{Name: "anthropic", DefaultModels: config.DefaultProviderModels["anthropic"]},
				{Name: "nvidia", DefaultModels: config.DefaultProviderModels["nvidia"]},
				{Name: "gemini", DefaultModels: config.DefaultProviderModels["gemini"]},
				{Name: "ollama", DefaultModels: config.DefaultProviderModels["ollama"]},
				{Name: "deepseek", DefaultModels: config.DefaultProviderModels["deepseek"]},
				{Name: "groq", DefaultModels: config.DefaultProviderModels["groq"]},
			},
		}
	}

	// Apply defaults to ensure all providers have models
	cfgMgr.ApplyDefaults(cfg)

	// Start the TUI
	p := tea.NewProgram(initialModel(cfg), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running model browser: %w", err)
	}

	return nil
}

func runModelsList(cmd *cobra.Command, _ []string) error {
	providerFilter, _ := cmd.Flags().GetString("provider")

	// Load configuration
	cfg, err := cfgMgr.Load()
	if err != nil {
		color.Yellow("No configuration found. Using default provider list.")
		cfg = &config.Config{
			Providers: []config.Provider{
				{Name: "openrouter", DefaultModels: config.DefaultProviderModels["openrouter"]},
				{Name: "openai", DefaultModels: config.DefaultProviderModels["openai"]},
				{Name: "anthropic", DefaultModels: config.DefaultProviderModels["anthropic"]},
				{Name: "nvidia", DefaultModels: config.DefaultProviderModels["nvidia"]},
				{Name: "gemini", DefaultModels: config.DefaultProviderModels["gemini"]},
				{Name: "ollama", DefaultModels: config.DefaultProviderModels["ollama"]},
				{Name: "deepseek", DefaultModels: config.DefaultProviderModels["deepseek"]},
				{Name: "groq", DefaultModels: config.DefaultProviderModels["groq"]},
			},
		}
	}

	cfgMgr.ApplyDefaults(cfg)

	for _, provider := range cfg.Providers {
		// Skip if provider filter is set and doesn't match
		if providerFilter != "" && provider.Name != providerFilter {
			continue
		}

		color.Blue("\n%s", strings.ToUpper(provider.Name))

		models := provider.GetAllowedModels()
		if len(models) == 0 {
			models = provider.DefaultModels
		}

		for _, model := range models {
			fmt.Printf("  %s/%s\n", provider.Name, model)
		}
	}

	fmt.Println()
	return nil
}

// Model selector TUI

type modelItem struct {
	provider string
	model    string
}

type model struct {
	cfg           *config.Config
	providers     []config.Provider
	cursor        int
	selected      int
	items         []modelItem
	viewportStart int
	viewportSize  int
	quitting      bool
}

type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
	Quit   key.Binding
	Help   key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter", " "),
		key.WithHelp("enter/space", "select"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q/esc", "quit"),
	),
}

func initialModel(cfg *config.Config) model {
	// Build list of all provider/model combinations
	var items []modelItem

	for _, provider := range cfg.Providers {
		models := provider.GetAllowedModels()
		if len(models) == 0 {
			models = provider.DefaultModels
		}

		for _, modelName := range models {
			items = append(items, modelItem{
				provider: provider.Name,
				model:    modelName,
			})
		}
	}

	return model{
		cfg:          cfg,
		providers:    cfg.Providers,
		items:        items,
		viewportSize: 20, // Default viewport size
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Adjust viewport based on terminal size
		m.viewportSize = msg.Height - 10 // Leave room for header and footer

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
				// Scroll viewport if needed
				if m.cursor < m.viewportStart {
					m.viewportStart = m.cursor
				}
			}

		case key.Matches(msg, keys.Down):
			if m.cursor < len(m.items)-1 {
				m.cursor++
				// Scroll viewport if needed
				if m.cursor >= m.viewportStart+m.viewportSize {
					m.viewportStart = m.cursor - m.viewportSize + 1
				}
			}

		case key.Matches(msg, keys.Select):
			m.selected = m.cursor
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.quitting {
		if m.selected >= 0 && m.selected < len(m.items) {
			item := m.items[m.selected]
			return m.renderSelectedModel(item)
		}
		return ""
	}

	var s strings.Builder

	// Header
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED")).
		Background(lipgloss.Color("#1E1E2E")).
		Padding(0, 1)

	s.WriteString(titleStyle.Render("🤖 CCO Model Browser"))
	s.WriteString("\n\n")

	// Provider stats
	s.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A6ADC8")).
		Render(fmt.Sprintf("Providers: %d | Models: %d", len(m.providers), len(m.items))))
	s.WriteString("\n\n")

	// Model list (viewport)
	itemStyle := lipgloss.NewStyle().PaddingLeft(2)
	selectedStyle := lipgloss.NewStyle().
		PaddingLeft(1).
		Foreground(lipgloss.Color("#89B4FA")).
		Bold(true)

	currentProvider := ""
	visibleIndex := 0

	for i, item := range m.items {
		// Skip items outside viewport
		if i < m.viewportStart {
			continue
		}
		if visibleIndex >= m.viewportSize {
			break
		}

		// Show provider header when it changes
		if item.provider != currentProvider {
			currentProvider = item.provider

			providerHeaderStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F9E2AF")).
				Bold(true).
				MarginTop(1)

			s.WriteString(providerHeaderStyle.Render(fmt.Sprintf("▶ %s", strings.ToUpper(item.provider))))
			s.WriteString("\n")
			visibleIndex++
		}

		// Render model item
		cursor := " "
		if i == m.cursor {
			cursor = "●"
		}

		line := fmt.Sprintf("%s %s/%s", cursor, item.provider, item.model)

		if i == m.cursor {
			s.WriteString(selectedStyle.Render(line))
		} else {
			s.WriteString(itemStyle.Render(line))
		}
		s.WriteString("\n")
		visibleIndex++
	}

	// Footer with help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6C7086")).
		MarginTop(2)

	s.WriteString(helpStyle.Render("↑/↓: Navigate | Enter: Select | q: Quit"))

	return s.String()
}

func (m model) renderSelectedModel(item modelItem) string {
	var s strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#A6E3A1"))

	s.WriteString(titleStyle.Render("✓ Model Selected"))
	s.WriteString("\n\n")

	// Model info
	s.WriteString(lipgloss.NewStyle().Bold(true).Render("Provider: "))
	s.WriteString(item.provider)
	s.WriteString("\n")

	s.WriteString(lipgloss.NewStyle().Bold(true).Render("Model: "))
	s.WriteString(item.model)
	s.WriteString("\n\n")

	// Usage example
	s.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#89B4FA")).
		Render("Usage with CCO:"))
	s.WriteString("\n\n")

	codeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A6E3A1")).
		Background(lipgloss.Color("#1E1E2E")).
		Padding(0, 1)

	fullModelName := fmt.Sprintf("%s/%s", item.provider, item.model)

	// Find router config for this model
	routerUsage := ""
	if m.cfg.Router.Default == fullModelName {
		routerUsage = " (currently set as default)"
	}

	s.WriteString(codeStyle.Render(fmt.Sprintf("export CCO_MODEL=\"%s\"%s", fullModelName, routerUsage)))
	s.WriteString("\n\n")

	s.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#89B4FA")).
		Render("Update config (YAML):"))
	s.WriteString("\n\n")

	yamlExample := fmt.Sprintf("router:\n  default: \"%s\"", fullModelName)
	s.WriteString(codeStyle.Render(yamlExample))
	s.WriteString("\n")

	return s.String()
}
