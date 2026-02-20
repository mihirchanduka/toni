package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type OnboardingSettings struct {
	Completed   bool `json:"completed"`
	YelpEnabled bool `json:"yelp_enabled"`
}

func onboardingPath(configDir string) string {
	return filepath.Join(configDir, "onboarding.json")
}

func loadOnboardingSettings(configDir string) (OnboardingSettings, error) {
	path := onboardingPath(configDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return OnboardingSettings{}, nil
		}
		return OnboardingSettings{}, err
	}

	var settings OnboardingSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return OnboardingSettings{}, err
	}
	return settings, nil
}

func saveOnboardingSettings(configDir string, settings OnboardingSettings) error {
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(onboardingPath(configDir), data, 0644)
}

func secureYelpKeyPath(configDir string) string {
	return filepath.Join(configDir, "yelp_api_key")
}

func saveSecureYelpAPIKey(configDir, key string) error {
	if strings.TrimSpace(key) == "" {
		return nil
	}
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}
	// Owner read/write only.
	return os.WriteFile(secureYelpKeyPath(configDir), []byte(strings.TrimSpace(key)+"\n"), 0600)
}

func loadSecureYelpAPIKey(configDir string) (string, error) {
	data, err := os.ReadFile(secureYelpKeyPath(configDir))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func shouldRunOnboarding(settings OnboardingSettings) bool {
	if settings.Completed {
		return false
	}
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

type onboardingStep int

const (
	stepEnable onboardingStep = iota
	stepKey
	stepDone
)

type onboardingModel struct {
	step        onboardingStep
	enable      bool
	existingKey string
	keyInput    textinput.Model
	settings    OnboardingSettings
	capturedKey string
	status      string
	width       int
	height      int
}

var (
	obColorSurface = lipgloss.Color("#2A332C")
	obColorMuted   = lipgloss.Color("#7E8C80")
	obColorText    = lipgloss.Color("#D6E0D3")
	obColorAccent  = lipgloss.Color("#8FA082")
	obColorDanger  = lipgloss.Color("#f38ba8")

	obTitleStyle = lipgloss.NewStyle().
			Foreground(obColorAccent).
			Bold(true)

	obHeaderStyle = lipgloss.NewStyle().
			Foreground(obColorAccent).
			Bold(true).
			Padding(0, 1).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(obColorMuted)

	obTabsStyle = lipgloss.NewStyle().
			Padding(0, 2).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(obColorMuted)

	obTabInactive = lipgloss.NewStyle().
			Foreground(obColorMuted).
			Padding(0, 2)

	obTabActive = lipgloss.NewStyle().
			Foreground(obColorText).
			Bold(true).
			Underline(true).
			Padding(0, 2)

	obPanelStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(obColorMuted).
			Padding(1, 2)

	obInputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(obColorAccent).
			Padding(0, 1)

	obLabelStyle = lipgloss.NewStyle().
			Foreground(obColorAccent).
			Bold(true)

	obMutedStyle = lipgloss.NewStyle().
			Foreground(obColorMuted)

	obOptionStyle = lipgloss.NewStyle().
			Foreground(obColorText)

	obOptionSelected = lipgloss.NewStyle().
				Foreground(obColorAccent).
				Bold(true)

	obWarnStyle = lipgloss.NewStyle().
			Foreground(obColorDanger)

	obFooterStyle = lipgloss.NewStyle().
			Foreground(obColorMuted).
			Padding(0, 1).
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(obColorMuted)
)

func newOnboardingModel(existingKey string) onboardingModel {
	in := textinput.New()
	in.Placeholder = "Paste YELP API key here"
	in.CharLimit = 300
	in.Prompt = "api> "
	in.TextStyle = lipgloss.NewStyle().Foreground(obColorText)
	in.PlaceholderStyle = lipgloss.NewStyle().Foreground(obColorMuted)
	in.Cursor.Style = lipgloss.NewStyle().Foreground(obColorText).Background(obColorAccent)
	in.Focus()

	return onboardingModel{
		step:        stepEnable,
		enable:      true,
		existingKey: strings.TrimSpace(existingKey),
		keyInput:    in,
		settings: OnboardingSettings{
			Completed:   true,
			YelpEnabled: true,
		},
	}
}

func (m onboardingModel) Init() tea.Cmd { return nil }

func (m onboardingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch m.step {
		case stepEnable:
			switch msg.String() {
			case "y", "Y":
				m.enable = true
				return m.nextStep()
			case "n", "N":
				m.enable = false
				return m.nextStep()
			case "up", "k":
				m.enable = true
				return m, nil
			case "down", "j":
				m.enable = false
				return m, nil
			case "left", "h":
				m.enable = true
				return m, nil
			case "right", "l":
				m.enable = false
				return m, nil
			case "enter":
				// Enter commits the currently selected option (m.enable)
				return m.nextStep()
			case "ctrl+c", "q":
				m.settings.Completed = true
				m.settings.YelpEnabled = false
				m.status = "Setup canceled. YELP autocomplete disabled."
				m.step = stepDone
				return m, tea.Quit
			default:
				// Swallow any other keys silently (no error flash)
				return m, nil
			}
		case stepKey:
			switch msg.String() {
			case "enter":
				key := strings.TrimSpace(m.keyInput.Value())
				if key == "" {
					m.settings.YelpEnabled = false
					m.status = "No key entered. YELP autocomplete disabled."
				} else {
					m.settings.YelpEnabled = true
					m.capturedKey = key
					m.status = "YELP API key saved."
				}
				m.step = stepDone
				return m, tea.Quit
			case "esc":
				m.settings.YelpEnabled = false
				m.status = "Skipped key setup. YELP autocomplete disabled."
				m.step = stepDone
				return m, tea.Quit
			case "ctrl+c", "q":
				m.settings.YelpEnabled = false
				m.status = "Setup canceled. YELP autocomplete disabled."
				m.step = stepDone
				return m, tea.Quit
			}
			var cmd tea.Cmd
			m.keyInput, cmd = m.keyInput.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

func (m onboardingModel) nextStep() (tea.Model, tea.Cmd) {
	if !m.enable {
		m.settings.YelpEnabled = false
		m.status = "YELP autocomplete disabled."
		m.step = stepDone
		return m, tea.Quit
	}
	if m.existingKey != "" {
		m.settings.YelpEnabled = true
		m.capturedKey = m.existingKey
		m.status = "Using existing YELP_API_KEY from environment/flags."
		m.step = stepDone
		return m, tea.Quit
	}
	m.step = stepKey
	return m, nil
}

func (m onboardingModel) View() string {
	width := m.width
	height := m.height
	if width <= 0 {
		width = 100
	}
	if height <= 0 {
		height = 28
	}

	header := m.renderHeader(width)
	tabs := m.renderTabs(width)
	footer := m.renderFooter(width)

	contentHeight := height - 6
	if contentHeight < 8 {
		contentHeight = 8
	}
	content := m.renderContent(width, contentHeight)
	ui := lipgloss.JoinVertical(lipgloss.Left, header, tabs, content, footer)

	return lipgloss.NewStyle().
		Foreground(obColorText).
		Width(width).
		Height(height).
		Render(ui)
}

func (m onboardingModel) renderHeader(width int) string {
	left := "  " + obTitleStyle.Render("toni") + " " + obMutedStyle.Render("› Setup")
	right := obMutedStyle.Render(time.Now().Format("Mon 02 Jan")) + "  "
	padding := width - lipgloss.Width(left) - lipgloss.Width(right)
	if padding < 0 {
		padding = 0
	}
	return obHeaderStyle.Width(width).Render(left + strings.Repeat(" ", padding) + right)
}

func (m onboardingModel) renderTabs(width int) string {
	enableTab := obTabInactive.Render("Enable API")
	keyTab := obTabInactive.Render("YELP API Key")
	if m.step == stepEnable {
		enableTab = obTabActive.Render("Enable API")
	}
	if m.step == stepKey {
		keyTab = obTabActive.Render("YELP API Key")
	}
	return obTabsStyle.Width(width).Render(lipgloss.JoinHorizontal(lipgloss.Left, "  ", enableTab, keyTab))
}

func (m onboardingModel) renderFooter(width int) string {
	switch m.step {
	case stepEnable:
		return obFooterStyle.Width(width).Render("↑↓/jk to navigate  y/n enter to confirm  q cancel")
	case stepKey:
		return obFooterStyle.Width(width).Render("enter save  esc skip  q cancel")
	default:
		return obFooterStyle.Width(width).Render("Setup complete")
	}
}

func (m onboardingModel) renderContent(width, height int) string {
	cardWidth := min(92, width-6)
	if cardWidth < 40 {
		cardWidth = width - 2
	}

	var body string
	switch m.step {
	case stepEnable:
		question := obLabelStyle.Render("Use YELP API for restaurant autocomplete?")
		on := "Enable YELP autocomplete"
		off := "Disable YELP autocomplete"

		var onDisplay, offDisplay string
		if m.enable {
			onDisplay = "  " + obOptionSelected.Render("→ "+on)
			offDisplay = "    " + obOptionStyle.Render(off)
		} else {
			onDisplay = "    " + obOptionStyle.Render(on)
			offDisplay = "  " + obOptionSelected.Render("→ "+off)
		}

		body = lipgloss.JoinVertical(
			lipgloss.Left,
			question,
			"",
			onDisplay,
			offDisplay,
			"",
			obMutedStyle.Render("Use arrow keys or j/k to navigate, y/n or Enter to confirm"),
			obMutedStyle.Render("You can change this later in ~/.toni/onboarding.json"),
		)
	case stepKey:
		input := obInputStyle.Width(max(30, cardWidth-14)).Render(m.keyInput.View())
		body = lipgloss.JoinVertical(
			lipgloss.Left,
			obLabelStyle.Render("Get a YELP API key:"),
			"",
			obMutedStyle.Render("Install/setup help:"),
			obMutedStyle.Render("https://github.com/mihirchanduka/toni"),
			"",
			obMutedStyle.Render("1) https://www.yelp.com/developers/v3/manage_app"),
			obMutedStyle.Render("2) Create an app"),
			obMutedStyle.Render("3) Copy API key"),
			"",
			obLabelStyle.Render("YELP API Key"),
			input,
			"",
			obMutedStyle.Render("Press Enter to save, Esc to skip."),
		)
	default:
		msg := obMutedStyle.Render(m.status)
		if strings.Contains(strings.ToLower(m.status), "disabled") {
			msg = obWarnStyle.Render(m.status)
		}
		body = lipgloss.JoinVertical(lipgloss.Left, obLabelStyle.Render("Onboarding Complete"), "", msg)
	}

	card := obPanelStyle.Width(cardWidth).Render(body)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Top, card)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func runOnboarding(configDir string, existingKey string) (OnboardingSettings, error) {
	model := newOnboardingModel(existingKey)
	prog := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := prog.Run()
	if err != nil {
		return OnboardingSettings{}, fmt.Errorf("onboarding tui failed: %w", err)
	}
	m, ok := finalModel.(onboardingModel)
	if !ok {
		return OnboardingSettings{}, fmt.Errorf("unexpected onboarding model type")
	}
	if strings.TrimSpace(m.capturedKey) != "" {
		if err := saveSecureYelpAPIKey(configDir, m.capturedKey); err != nil {
			return OnboardingSettings{}, err
		}
	}
	if err := saveOnboardingSettings(configDir, m.settings); err != nil {
		return OnboardingSettings{}, err
	}
	return m.settings, nil
}
