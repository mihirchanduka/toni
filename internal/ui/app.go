package ui

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
	"toni/internal/db"
	"toni/internal/model"
	"toni/internal/search"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const (
	minUsableWidth  = 72
	minUsableHeight = 18
)

// Model is the root Bubble Tea model.
type Model struct {
	db               *sql.DB
	yelpClient       *search.YelpClient
	termCapabilities TerminalCapabilities
	screen           model.Screen
	mode             model.Mode
	gState           GState

	width  int
	height int

	error        string
	info         string
	showingHelp  bool
	columnJump   bool
	returnScreen model.Screen

	// Screen models
	visits            *VisitsModel
	restaurants       *RestaurantsModel
	wantToVisit       *WantToVisitModel
	visitDetail       *VisitDetailModel
	restaurantDetail  *RestaurantDetailModel
	wantToVisitDetail *WantToVisitDetailModel
	visitForm         *VisitFormModel
	restaurantForm    *RestaurantFormModel
	wantToVisitForm   *WantToVisitFormModel

	keys      KeyMap
	formKeys  FormKeyMap
	prefs     UIPreferences
	undoStack []undoAction
	redoStack []undoAction
}

// New creates a new root model.
func New(database *sql.DB, yelpClient *search.YelpClient, termCaps TerminalCapabilities) Model {
	return Model{
		db:               database,
		yelpClient:       yelpClient,
		termCapabilities: termCaps,
		screen:           model.ScreenVisits,
		mode:             model.ModeNav,
		gState:           GStateIdle,
		keys:             DefaultKeyMap(),
		formKeys:         DefaultFormKeyMap(),
		prefs:            loadUIPreferences(),
		returnScreen:     model.ScreenVisits,
	}
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return loadVisitsCmd(m.db, "")
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.mode == model.ModeNav && m.columnJump {
			switch msg.String() {
			case "esc":
				m.columnJump = false
				m.info = ""
				return m, nil
			}
			if n, err := strconv.Atoi(msg.String()); err == nil {
				table := m.currentTable()
				if table != nil && table.JumpToColumn(n) {
					m.columnJump = false
					m.info = fmt.Sprintf("Jumped to column %d", n)
					m.persistCurrentTablePrefs()
					return m, nil
				}
				m.info = fmt.Sprintf("Column %d unavailable", n)
				return m, nil
			}
		}

		// Handle ctrl+c globally
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Handle help toggle
		if msg.String() == "?" && m.mode == model.ModeNav {
			m.showingHelp = !m.showingHelp
			return m, nil
		}

		if m.showingHelp {
			if msg.String() == "esc" || msg.String() == "?" {
				m.showingHelp = false
			}
			return m, nil
		}

		// Route to mode-specific handlers
		if m.mode == model.ModeNav {
			return m.handleNavMode(msg)
		}
		return m.handleInsertMode(msg)

	case model.ErrorMsg:
		m.error = msg.Err.Error()
		return m, nil

	case model.VisitsLoadedMsg:
		m.visits = NewVisitsModel(msg.Visits)
		m.visits.ApplyPrefs(m.prefs.Visits)
		m.error = ""
		return m, nil

	case model.RestaurantsLoadedMsg:
		m.restaurants = NewRestaurantsModel(msg.Restaurants)
		m.restaurants.ApplyPrefs(m.prefs.Restaurants)
		m.error = ""
		return m, nil

	case model.VisitDetailLoadedMsg:
		m.visitDetail = NewVisitDetailModel(msg.Visit, msg.Restaurant)
		m.screen = model.ScreenVisitDetail
		m.error = ""
		return m, nil

	case model.RestaurantDetailLoadedMsg:
		m.restaurantDetail = NewRestaurantDetailModel(msg.Detail)
		m.screen = model.ScreenRestaurantDetail
		m.error = ""
		return m, nil

	case model.VisitSavedMsg:
		if action := m.buildVisitSaveAction(msg); action != nil {
			m.pushUndoAction(*action)
		}
		m.mode = model.ModeNav
		m.screen = model.ScreenVisits
		m.visitForm = nil
		m.info = "Visit saved"
		return m, tea.Batch(
			loadVisitsCmd(m.db, ""),
			loadRestaurantsCmd(m.db, ""),
			loadWantToVisitCmd(m.db),
		)

	case model.RestaurantSavedMsg:
		if action := m.buildRestaurantSaveAction(msg); action != nil {
			m.pushUndoAction(*action)
		}
		m.mode = model.ModeNav
		m.screen = model.ScreenRestaurants
		m.restaurantForm = nil
		m.info = "Restaurant saved"
		return m, tea.Batch(
			loadRestaurantsCmd(m.db, ""),
			loadVisitsCmd(m.db, ""),
			loadWantToVisitCmd(m.db),
		)

	case model.FormCancelledMsg:
		m.mode = model.ModeNav
		m.visitForm = nil
		m.restaurantForm = nil
		m.wantToVisitForm = nil
		m.screen = m.returnScreen
		return m, nil

	case model.DeleteVisitMsg:
		m.pushUndoAction(m.buildDeleteVisitAction(msg))
		m.screen = model.ScreenVisits
		m.visitDetail = nil
		m.info = "Visit deleted (u to undo)"
		return m, tea.Batch(
			loadVisitsCmd(m.db, ""),
			loadRestaurantsCmd(m.db, ""),
			loadWantToVisitCmd(m.db),
		)

	case model.DeleteRestaurantMsg:
		m.pushUndoAction(m.buildDeleteRestaurantAction(msg))
		m.screen = model.ScreenRestaurants
		m.restaurantDetail = nil
		m.info = "Restaurant deleted (u to undo)"
		return m, tea.Batch(
			loadRestaurantsCmd(m.db, ""),
			loadVisitsCmd(m.db, ""),
			loadWantToVisitCmd(m.db),
		)

	case model.WantToVisitLoadedMsg:
		m.wantToVisit = NewWantToVisitModel(msg.WantToVisit)
		m.wantToVisit.ApplyPrefs(m.prefs.WantToVisit)
		m.error = ""
		return m, nil

	case model.WantToVisitSavedMsg:
		if action := m.buildWantToVisitSaveAction(msg); action != nil {
			m.pushUndoAction(*action)
		}
		m.mode = model.ModeNav
		m.screen = model.ScreenWantToVisit
		m.wantToVisitForm = nil
		m.info = "Want-to-visit entry saved"
		return m, tea.Batch(
			loadWantToVisitCmd(m.db),
			loadRestaurantsCmd(m.db, ""),
		)

	case model.DeleteWantToVisitMsg:
		m.pushUndoAction(m.buildDeleteWantToVisitAction(msg))
		m.screen = model.ScreenWantToVisit
		m.wantToVisitDetail = nil
		m.info = "Want-to-visit entry deleted (u to undo)"
		return m, loadWantToVisitCmd(m.db)

	case model.ConvertToVisitMsg:
		m.pushUndoAction(m.buildConvertAction(msg))
		// Convert want_to_visit to visit - open visit form with restaurant pre-filled
		m.mode = model.ModeInsert
		m.screen = model.ScreenVisitForm
		m.returnScreen = model.ScreenWantToVisit
		m.visitForm = NewVisitFormModel(m.db, m.yelpClient, msg.RestaurantID)
		m.wantToVisitDetail = nil
		m.info = "Converted to visit (u to undo)"
		return m, loadWantToVisitCmd(m.db)

	case undoAppliedMsg:
		return m, m.applyUndoResult(msg)

	case wantToVisitDetailLoadedMsg:
		m.wantToVisitDetail = NewWantToVisitDetailModel(msg.entry, msg.restaurant)
		m.screen = model.ScreenWantToVisitDetail
		m.error = ""
		return m, nil

	default:
		// Pass all other messages to forms
		if m.mode == model.ModeInsert {
			return m.handleInsertMode(msg)
		}
	}

	// Delegate to current screen
	return m.updateCurrentScreen(msg)
}

// View renders the UI.
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	if m.width < minUsableWidth || m.height < minUsableHeight {
		return renderTerminalTooSmall(m.width, m.height)
	}
	if m.showingHelp {
		return RenderFullHelp(m.width, m.height)
	}

	var content string
	var breadcrumbParts []string

	// Determine if this screen should show tabs
	showTabs := m.screen == model.ScreenVisits ||
		m.screen == model.ScreenRestaurants ||
		m.screen == model.ScreenWantToVisit

	switch m.screen {
	case model.ScreenVisits:
		breadcrumbParts = []string{"Visits"}
	case model.ScreenRestaurants:
		breadcrumbParts = []string{"Restaurants"}
	case model.ScreenWantToVisit:
		breadcrumbParts = []string{"Want to Visit"}
	case model.ScreenVisitDetail:
		breadcrumbParts = []string{"Visits", "Detail"}
		if m.visitDetail != nil {
			breadcrumbParts = []string{"Visits", m.visitDetail.restaurant.Name}
		}
	case model.ScreenRestaurantDetail:
		breadcrumbParts = []string{"Restaurants", "Detail"}
		if m.restaurantDetail != nil {
			breadcrumbParts = []string{"Restaurants", m.restaurantDetail.detail.Restaurant.Name}
		}
	case model.ScreenWantToVisitDetail:
		breadcrumbParts = []string{"Want to Visit", "Detail"}
		if m.wantToVisitDetail != nil {
			breadcrumbParts = []string{"Want to Visit", m.wantToVisitDetail.restaurant.Name}
		}
	case model.ScreenVisitForm:
		breadcrumbParts = []string{"Visits", "Form"}
	case model.ScreenRestaurantForm:
		breadcrumbParts = []string{"Restaurants", "Form"}
	case model.ScreenWantToVisitForm:
		breadcrumbParts = []string{"Want to Visit", "Form"}
	}

	header := renderHeader(breadcrumbParts, m.width)
	tabs := ""
	if showTabs {
		tabs = renderTabs(m.screen, m.width)
	}
	footer := RenderHelp(m.screen, m.mode, m.width)
	var banners []string
	if m.error != "" {
		banners = append(banners, ErrorStyle.Width(m.width).Render("Error: "+m.error))
	}
	if m.info != "" {
		banners = append(banners, SuccessStyle.Width(m.width).Render(m.info))
	}

	contentHeight := m.height - lipgloss.Height(header) - lipgloss.Height(footer)
	if tabs != "" {
		contentHeight -= lipgloss.Height(tabs)
	}
	for _, b := range banners {
		contentHeight -= lipgloss.Height(b)
	}
	if contentHeight < 3 {
		return renderTerminalTooSmall(m.width, m.height)
	}

	switch m.screen {
	case model.ScreenVisits:
		if m.visits != nil {
			content = m.visits.View(m.width, contentHeight)
		}
	case model.ScreenRestaurants:
		if m.restaurants != nil {
			content = m.restaurants.View(m.width, contentHeight)
		}
	case model.ScreenWantToVisit:
		if m.wantToVisit != nil {
			content = m.wantToVisit.View(m.width, contentHeight)
		}
	case model.ScreenVisitDetail:
		if m.visitDetail != nil {
			content = m.visitDetail.View(m.width, contentHeight)
		}
	case model.ScreenRestaurantDetail:
		if m.restaurantDetail != nil {
			content = m.restaurantDetail.View(m.width, contentHeight)
		}
	case model.ScreenWantToVisitDetail:
		if m.wantToVisitDetail != nil {
			content = m.wantToVisitDetail.View(m.width, contentHeight)
		}
	case model.ScreenVisitForm:
		if m.visitForm != nil {
			content = m.visitForm.View(m.width, contentHeight)
		}
	case model.ScreenRestaurantForm:
		if m.restaurantForm != nil {
			content = m.restaurantForm.View(m.width, contentHeight)
		}
	case model.ScreenWantToVisitForm:
		if m.wantToVisitForm != nil {
			content = m.wantToVisitForm.View(m.width, contentHeight)
		}
	}

	// Ensure content fills the available height to anchor footer at bottom
	contentStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(contentHeight)
	content = contentStyle.Render(content)

	parts := []string{header}
	if tabs != "" {
		parts = append(parts, tabs)
	}
	parts = append(parts, banners...)
	parts = append(parts, content, footer)
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func renderTerminalTooSmall(width, height int) string {
	panel := lipgloss.JoinVertical(
		lipgloss.Center,
		ErrorStyle.Render("Terminal too small"),
		"",
		BreadcrumbStyle.Render(fmt.Sprintf(
			"%dx%d detected, need at least %dx%d",
			width, height, minUsableWidth, minUsableHeight,
		)),
	)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, panel)
}

func (m Model) switchTopLevel(target model.Screen) (tea.Model, tea.Cmd) {
	m.screen = target
	switch target {
	case model.ScreenVisits:
		return m, nil
	case model.ScreenWantToVisit:
		if m.wantToVisit == nil {
			return m, loadWantToVisitCmd(m.db)
		}
		return m, nil
	case model.ScreenRestaurants:
		if m.restaurants == nil {
			return m, loadRestaurantsCmd(m.db, "")
		}
	}
	return m, nil
}

func nextTopLevelScreen(current model.Screen) model.Screen {
	switch current {
	case model.ScreenVisits:
		return model.ScreenWantToVisit
	case model.ScreenWantToVisit:
		return model.ScreenRestaurants
	default:
		return model.ScreenVisits
	}
}

func prevTopLevelScreen(current model.Screen) model.Screen {
	switch current {
	case model.ScreenVisits:
		return model.ScreenRestaurants
	case model.ScreenWantToVisit:
		return model.ScreenVisits
	default:
		return model.ScreenWantToVisit
	}
}

func renderTabs(screen model.Screen, width int) string {
	// Define tabs
	tabs := []struct {
		name   string
		screen model.Screen
	}{
		{"Visits", model.ScreenVisits},
		{"Want to Visit", model.ScreenWantToVisit},
		{"Restaurants", model.ScreenRestaurants},
	}

	var tabStrings []string
	for _, tab := range tabs {
		tabStyle := lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(ColorMuted)

		if screen == tab.screen {
			tabStyle = tabStyle.
				Foreground(ColorText).
				Bold(true).
				Underline(true)
		}

		tabStrings = append(tabStrings, tabStyle.Render(tab.name))
	}

	tabBar := lipgloss.JoinHorizontal(lipgloss.Left, tabStrings...)
	return lipgloss.NewStyle().
		Width(width).
		Padding(0, 2).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(ColorMuted).
		Render(tabBar)
}

func renderHeader(breadcrumbParts []string, width int) string {
	// Build raw strings first so truncation never touches styled ANSI content.
	leftRaw := "toni"
	if len(breadcrumbParts) > 0 {
		leftRaw += " › " + strings.Join(breadcrumbParts, " › ")
	}
	rightRaw := " " + time.Now().Format("Mon 02 Jan") + " "

	box := HeaderBoxStyle.Copy()
	innerWidth := width - box.GetHorizontalFrameSize()
	if innerWidth < 1 {
		innerWidth = 1
	}

	// Keep the header stable by fitting plain text into available width.
	rightLen := lipgloss.Width(rightRaw)
	if rightLen >= innerWidth {
		rightRaw = ""
		rightLen = 0
	}
	maxLeft := innerWidth - rightLen
	if rightLen > 0 {
		maxLeft-- // ensure at least one spacer between left and date
	}
	// Subtract 2 to account for potential wide char issues or rounding
	if maxLeft > 2 {
		maxLeft -= 2
	}
	if maxLeft < 0 {
		maxLeft = 0
	}
	leftRaw = ansi.Truncate(leftRaw, maxLeft, "…")

	left := HeaderStyle.Render(leftRaw)
	right := BreadcrumbStyle.Render(rightRaw)

	// Calculate padding
	leftLen := lipgloss.Width(left)
	padding := innerWidth - leftLen - rightLen
	if padding < 0 {
		padding = 0
	}

	headerContent := left + strings.Repeat(" ", padding) + right
	return box.Width(width).Render(
		lipgloss.NewStyle().Width(innerWidth).Render(headerContent),
	)
}

// handleNavMode handles navigation mode input.
func (m Model) handleNavMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if t := m.currentTable(); t != nil {
		switch msg.String() {
		case "tab":
			t.NextColumn()
			m.persistCurrentTablePrefs()
			return m, nil
		case "shift+tab":
			t.PrevColumn()
			m.persistCurrentTablePrefs()
			return m, nil
		case "/":
			m.columnJump = true
			m.info = "Jump to column: press 1-9 (esc to cancel)"
			return m, nil
		case "s":
			m.info = t.CycleSortActiveColumn()
			m.persistCurrentTablePrefs()
			return m, nil
		case "c":
			if t.HideActiveColumn() {
				m.info = "Column hidden"
				m.persistCurrentTablePrefs()
			} else {
				m.info = "Cannot hide last visible column"
			}
			return m, nil
		case "C":
			t.ShowAllColumns()
			m.info = "All columns shown"
			m.persistCurrentTablePrefs()
			return m, nil
		case "n":
			m.info = t.CycleFilterBySelectedValue()
			m.persistCurrentTablePrefs()
			return m, nil
		}
	}

	switch msg.String() {
	case "u":
		if len(m.undoStack) == 0 {
			m.info = "Nothing to undo"
			return m, nil
		}
		return m, m.undoCmd()
	case "ctrl+r":
		if len(m.redoStack) == 0 {
			m.info = "Nothing to redo"
			return m, nil
		}
		return m, m.redoCmd()
	}

	// Handle "gg" state machine
	if msg.String() == "g" {
		if m.gState == GStateIdle {
			m.gState = GStateFirstG
			return m, nil
		} else if m.gState == GStateFirstG {
			m.gState = GStateIdle
			// Jump to top
			return m.handleJumpToTop()
		}
	} else {
		// Any other key resets the state
		if m.gState == GStateFirstG {
			m.gState = GStateIdle
		}
	}

	// Screen-specific navigation
	switch m.screen {
	case model.ScreenVisits:
		return m.handleVisitsNav(msg)
	case model.ScreenRestaurants:
		return m.handleRestaurantsNav(msg)
	case model.ScreenWantToVisit:
		return m.handleWantToVisitNav(msg)
	case model.ScreenVisitDetail:
		return m.handleVisitDetailNav(msg)
	case model.ScreenRestaurantDetail:
		return m.handleRestaurantDetailNav(msg)
	case model.ScreenWantToVisitDetail:
		return m.handleWantToVisitDetailNav(msg)
	}

	return m, nil
}

func (m *Model) currentTable() tableController {
	switch m.screen {
	case model.ScreenVisits:
		if m.visits != nil {
			return m.visits
		}
	case model.ScreenRestaurants:
		if m.restaurants != nil {
			return m.restaurants
		}
	case model.ScreenWantToVisit:
		if m.wantToVisit != nil {
			return m.wantToVisit
		}
	}
	return nil
}

func (m *Model) persistCurrentTablePrefs() {
	switch m.screen {
	case model.ScreenVisits:
		if m.visits != nil {
			m.prefs.Visits = m.visits.Prefs()
		}
	case model.ScreenRestaurants:
		if m.restaurants != nil {
			m.prefs.Restaurants = m.restaurants.Prefs()
		}
	case model.ScreenWantToVisit:
		if m.wantToVisit != nil {
			m.prefs.WantToVisit = m.wantToVisit.Prefs()
		}
	}
	_ = saveUIPreferences(m.prefs)
}

// handleInsertMode handles insert/edit mode input.
func (m Model) handleInsertMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.screen {
	case model.ScreenVisitForm:
		if m.visitForm != nil {
			newForm, cmd := m.visitForm.Update(msg)
			m.visitForm = &newForm
			return m, cmd
		}
	case model.ScreenRestaurantForm:
		if m.restaurantForm != nil {
			// Restaurant form only handles KeyMsg
			if keyMsg, ok := msg.(tea.KeyMsg); ok {
				newForm, cmd := m.restaurantForm.Update(keyMsg)
				m.restaurantForm = &newForm
				return m, cmd
			}
		}
	case model.ScreenWantToVisitForm:
		if m.wantToVisitForm != nil {
			newForm, cmd := m.wantToVisitForm.Update(msg)
			m.wantToVisitForm = &newForm
			return m, cmd
		}
	}
	return m, nil
}

func (m Model) updateCurrentScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) handleJumpToTop() (tea.Model, tea.Cmd) {
	if m.visits != nil && m.screen == model.ScreenVisits {
		m.visits.JumpToTop()
	}
	if m.restaurants != nil && m.screen == model.ScreenRestaurants {
		m.restaurants.JumpToTop()
	}
	if m.wantToVisit != nil && m.screen == model.ScreenWantToVisit {
		m.wantToVisit.JumpToTop()
	}
	return m, nil
}

// Navigation handlers for each screen
func (m Model) handleVisitsNav(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.visits == nil {
		return m, nil
	}

	switch {
	case msg.String() == "q":
		return m, tea.Quit
	case msg.String() == "left" || msg.String() == "b":
		return m.switchTopLevel(prevTopLevelScreen(m.screen))
	case msg.String() == "right" || msg.String() == "f":
		return m.switchTopLevel(nextTopLevelScreen(m.screen))
	case msg.String() == "B":
		return m.switchTopLevel(model.ScreenVisits)
	case msg.String() == "F":
		return m.switchTopLevel(model.ScreenRestaurants)
	case msg.String() == "r":
		m.screen = model.ScreenRestaurants
		if m.restaurants == nil {
			return m, loadRestaurantsCmd(m.db, "")
		}
		return m, nil
	case msg.String() == "w":
		m.screen = model.ScreenWantToVisit
		if m.wantToVisit == nil {
			return m, loadWantToVisitCmd(m.db)
		}
		return m, nil
	case msg.String() == "a":
		m.returnScreen = model.ScreenVisits
		m.mode = model.ModeInsert
		m.screen = model.ScreenVisitForm
		m.visitForm = NewVisitFormModel(m.db, m.yelpClient, 0)
		return m, nil
	case msg.String() == "enter" || msg.String() == "l":
		if len(m.visits.rows) > 0 && m.visits.cursor < len(m.visits.rows) {
			visitID := m.visits.rows[m.visits.cursor].ID
			return m, loadVisitDetailCmd(m.db, visitID)
		}
		return m, nil
	case msg.String() == "j" || msg.String() == "down":
		m.visits.MoveDown()
		return m, nil
	case msg.String() == "k" || msg.String() == "up":
		m.visits.MoveUp()
		return m, nil
	case msg.String() == "G":
		m.visits.JumpToBottom()
		return m, nil
	case msg.String() == "ctrl+d" || msg.String() == "pgdown":
		m.visits.HalfPageDown(m.height)
		return m, nil
	case msg.String() == "ctrl+u" || msg.String() == "pgup":
		m.visits.HalfPageUp(m.height)
		return m, nil
	}

	return m, nil
}

func (m Model) handleRestaurantsNav(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.restaurants == nil {
		return m, nil
	}

	switch {
	case msg.String() == "q":
		return m, tea.Quit
	case msg.String() == "left" || msg.String() == "b":
		return m.switchTopLevel(prevTopLevelScreen(m.screen))
	case msg.String() == "right" || msg.String() == "f":
		return m.switchTopLevel(nextTopLevelScreen(m.screen))
	case msg.String() == "B":
		return m.switchTopLevel(model.ScreenVisits)
	case msg.String() == "F":
		return m.switchTopLevel(model.ScreenRestaurants)
	case msg.String() == "h":
		return m.switchTopLevel(model.ScreenVisits)
	case msg.String() == "w":
		m.screen = model.ScreenWantToVisit
		if m.wantToVisit == nil {
			return m, loadWantToVisitCmd(m.db)
		}
		return m, nil
	case msg.String() == "a":
		m.returnScreen = model.ScreenRestaurants
		m.mode = model.ModeInsert
		m.screen = model.ScreenRestaurantForm
		m.restaurantForm = NewRestaurantFormModel(m.db, 0)
		return m, nil
	case msg.String() == "v":
		if len(m.restaurants.rows) > 0 && m.restaurants.cursor < len(m.restaurants.rows) {
			restaurantID := m.restaurants.rows[m.restaurants.cursor].ID
			m.returnScreen = model.ScreenRestaurants
			m.mode = model.ModeInsert
			m.screen = model.ScreenVisitForm
			m.visitForm = NewVisitFormModel(m.db, m.yelpClient, restaurantID)
			return m, nil
		}
		return m, nil
	case msg.String() == "enter" || msg.String() == "l":
		if len(m.restaurants.rows) > 0 && m.restaurants.cursor < len(m.restaurants.rows) {
			restaurantID := m.restaurants.rows[m.restaurants.cursor].ID
			return m, loadRestaurantDetailCmd(m.db, restaurantID)
		}
		return m, nil
	case msg.String() == "j" || msg.String() == "down":
		m.restaurants.MoveDown()
		return m, nil
	case msg.String() == "k" || msg.String() == "up":
		m.restaurants.MoveUp()
		return m, nil
	case msg.String() == "G":
		m.restaurants.JumpToBottom()
		return m, nil
	case msg.String() == "ctrl+d" || msg.String() == "pgdown":
		m.restaurants.HalfPageDown(m.height)
		return m, nil
	case msg.String() == "ctrl+u" || msg.String() == "pgup":
		m.restaurants.HalfPageUp(m.height)
		return m, nil
	}

	return m, nil
}

func (m Model) handleVisitDetailNav(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.String() == "h" || msg.String() == "esc" || msg.String() == "b":
		m.screen = model.ScreenVisits
		m.visitDetail = nil
		return m, nil
	case msg.String() == "e":
		if m.visitDetail != nil {
			m.returnScreen = model.ScreenVisitDetail
			m.mode = model.ModeInsert
			m.screen = model.ScreenVisitForm
			m.visitForm = NewVisitFormModel(m.db, m.yelpClient, 0)
			m.visitForm.LoadVisit(m.visitDetail.visit)
			return m, nil
		}
		return m, nil
	case msg.String() == "d":
		if m.visitDetail != nil {
			return m, deleteVisitCmd(m.db, m.visitDetail.visit.ID)
		}
		return m, nil
	}
	return m, nil
}

func (m Model) handleRestaurantDetailNav(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.String() == "h" || msg.String() == "esc" || msg.String() == "b":
		m.screen = model.ScreenRestaurants
		m.restaurantDetail = nil
		return m, nil
	case msg.String() == "v":
		if m.restaurantDetail != nil {
			m.returnScreen = model.ScreenRestaurantDetail
			m.mode = model.ModeInsert
			m.screen = model.ScreenVisitForm
			m.visitForm = NewVisitFormModel(m.db, m.yelpClient, m.restaurantDetail.detail.Restaurant.ID)
			return m, nil
		}
		return m, nil
	case msg.String() == "e":
		if m.restaurantDetail != nil {
			m.returnScreen = model.ScreenRestaurantDetail
			m.mode = model.ModeInsert
			m.screen = model.ScreenRestaurantForm
			m.restaurantForm = NewRestaurantFormModel(m.db, 0)
			m.restaurantForm.LoadRestaurant(m.restaurantDetail.detail.Restaurant)
			return m, nil
		}
		return m, nil
	case msg.String() == "d":
		if m.restaurantDetail != nil {
			return m, deleteRestaurantCmd(m.db, m.restaurantDetail.detail.Restaurant.ID)
		}
		return m, nil
	}
	return m, nil
}

func (m Model) handleWantToVisitNav(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.wantToVisit == nil {
		return m, nil
	}

	switch {
	case msg.String() == "q":
		return m, tea.Quit
	case msg.String() == "left" || msg.String() == "b":
		return m.switchTopLevel(prevTopLevelScreen(m.screen))
	case msg.String() == "right" || msg.String() == "f":
		return m.switchTopLevel(nextTopLevelScreen(m.screen))
	case msg.String() == "B":
		return m.switchTopLevel(model.ScreenVisits)
	case msg.String() == "F":
		return m.switchTopLevel(model.ScreenRestaurants)
	case msg.String() == "v":
		m.screen = model.ScreenVisits
		return m, nil
	case msg.String() == "r":
		m.screen = model.ScreenRestaurants
		if m.restaurants == nil {
			return m, loadRestaurantsCmd(m.db, "")
		}
		return m, nil
	case msg.String() == "a":
		m.returnScreen = model.ScreenWantToVisit
		m.mode = model.ModeInsert
		m.screen = model.ScreenWantToVisitForm
		m.wantToVisitForm = NewWantToVisitFormModel(m.db, m.yelpClient, 0)
		return m, nil
	case msg.String() == "enter" || msg.String() == "l":
		entry := m.wantToVisit.SelectedEntry()
		if entry != nil {
			return m, loadWantToVisitDetailCmd(m.db, entry.ID)
		}
		return m, nil
	case msg.String() == "j" || msg.String() == "down":
		m.wantToVisit.CursorDown()
		return m, nil
	case msg.String() == "k" || msg.String() == "up":
		m.wantToVisit.CursorUp()
		return m, nil
	case msg.String() == "G":
		m.wantToVisit.JumpToBottom()
		return m, nil
	case msg.String() == "ctrl+d" || msg.String() == "pgdown":
		// Half page down (approximate)
		for i := 0; i < m.height/2; i++ {
			m.wantToVisit.CursorDown()
		}
		return m, nil
	case msg.String() == "ctrl+u" || msg.String() == "pgup":
		// Half page up (approximate)
		for i := 0; i < m.height/2; i++ {
			m.wantToVisit.CursorUp()
		}
		return m, nil
	}

	return m, nil
}

func (m Model) handleWantToVisitDetailNav(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.String() == "h" || msg.String() == "esc" || msg.String() == "b":
		m.screen = model.ScreenWantToVisit
		m.wantToVisitDetail = nil
		return m, nil
	case msg.String() == "c":
		// Convert to visit - mark as visited
		if m.wantToVisitDetail != nil {
			return m, convertToVisitCmd(m.db, m.wantToVisitDetail.entry.ID)
		}
		return m, nil
	case msg.String() == "e":
		if m.wantToVisitDetail != nil {
			m.returnScreen = model.ScreenWantToVisitDetail
			m.mode = model.ModeInsert
			m.screen = model.ScreenWantToVisitForm
			m.wantToVisitForm = NewWantToVisitFormModel(m.db, m.yelpClient, 0)
			m.wantToVisitForm.LoadWantToVisit(m.wantToVisitDetail.entry)
			return m, nil
		}
		return m, nil
	case msg.String() == "d":
		if m.wantToVisitDetail != nil {
			return m, deleteWantToVisitCmd(m.db, m.wantToVisitDetail.entry.ID)
		}
		return m, nil
	}
	return m, nil
}

// Commands

func loadVisitsCmd(database *sql.DB, filter string) tea.Cmd {
	return func() tea.Msg {
		visits, err := db.ListVisits(database, filter)
		if err != nil {
			return model.ErrorMsg{Err: err}
		}
		return model.VisitsLoadedMsg{Visits: visits}
	}
}

func loadRestaurantsCmd(database *sql.DB, filter string) tea.Cmd {
	return func() tea.Msg {
		restaurants, err := db.ListRestaurants(database, filter)
		if err != nil {
			return model.ErrorMsg{Err: err}
		}
		return model.RestaurantsLoadedMsg{Restaurants: restaurants}
	}
}

func loadVisitDetailCmd(database *sql.DB, visitID int64) tea.Cmd {
	return func() tea.Msg {
		visit, err := db.GetVisit(database, visitID)
		if err != nil {
			return model.ErrorMsg{Err: fmt.Errorf("failed to load visit: %w", err)}
		}

		restaurant, err := db.GetRestaurant(database, visit.RestaurantID)
		if err != nil {
			return model.ErrorMsg{Err: fmt.Errorf("failed to load restaurant: %w", err)}
		}

		return model.VisitDetailLoadedMsg{
			Visit:      visit,
			Restaurant: restaurant,
		}
	}
}

func loadRestaurantDetailCmd(database *sql.DB, restaurantID int64) tea.Cmd {
	return func() tea.Msg {
		detail, err := db.GetRestaurantWithStats(database, restaurantID)
		if err != nil {
			return model.ErrorMsg{Err: fmt.Errorf("failed to load restaurant: %w", err)}
		}

		return model.RestaurantDetailLoadedMsg{Detail: detail}
	}
}

func deleteVisitCmd(database *sql.DB, visitID int64) tea.Cmd {
	return func() tea.Msg {
		visit, err := db.GetVisit(database, visitID)
		if err != nil {
			return model.ErrorMsg{Err: fmt.Errorf("failed to load visit before delete: %w", err)}
		}

		err = db.DeleteVisit(database, visitID)
		if err != nil {
			return model.ErrorMsg{Err: fmt.Errorf("failed to delete visit: %w", err)}
		}
		return model.DeleteVisitMsg{ID: visitID, Deleted: visit}
	}
}

func deleteRestaurantCmd(database *sql.DB, restaurantID int64) tea.Cmd {
	return func() tea.Msg {
		restaurant, err := db.GetRestaurant(database, restaurantID)
		if err != nil {
			return model.ErrorMsg{Err: fmt.Errorf("failed to load restaurant before delete: %w", err)}
		}
		visits, err := db.GetVisitsByRestaurant(database, restaurantID)
		if err != nil {
			return model.ErrorMsg{Err: fmt.Errorf("failed to load restaurant visits before delete: %w", err)}
		}
		wantToVisitEntries, err := db.GetWantToVisitByRestaurant(database, restaurantID)
		if err != nil {
			return model.ErrorMsg{Err: fmt.Errorf("failed to load related want_to_visit before delete: %w", err)}
		}

		err = db.DeleteRestaurant(database, restaurantID)
		if err != nil {
			return model.ErrorMsg{Err: fmt.Errorf("failed to delete restaurant: %w", err)}
		}
		return model.DeleteRestaurantMsg{
			ID:                 restaurantID,
			Deleted:            restaurant,
			DeletedVisits:      visits,
			DeletedWantToVisit: wantToVisitEntries,
		}
	}
}

func loadWantToVisitCmd(database *sql.DB) tea.Cmd {
	return func() tea.Msg {
		entries, err := db.GetWantToVisitList(database, "")
		if err != nil {
			return model.ErrorMsg{Err: err}
		}
		return model.WantToVisitLoadedMsg{WantToVisit: entries}
	}
}

func loadWantToVisitDetailCmd(database *sql.DB, wtvID int64) tea.Cmd {
	return func() tea.Msg {
		wtv, err := db.GetWantToVisit(database, wtvID)
		if err != nil {
			return model.ErrorMsg{Err: fmt.Errorf("failed to load want_to_visit: %w", err)}
		}

		restaurant, err := db.GetRestaurant(database, wtv.RestaurantID)
		if err != nil {
			return model.ErrorMsg{Err: fmt.Errorf("failed to load restaurant: %w", err)}
		}

		// Return a special message type for want_to_visit detail
		return wantToVisitDetailLoadedMsg{
			entry:      wtv,
			restaurant: restaurant,
		}
	}
}

func deleteWantToVisitCmd(database *sql.DB, wtvID int64) tea.Cmd {
	return func() tea.Msg {
		wtv, err := db.GetWantToVisit(database, wtvID)
		if err != nil {
			return model.ErrorMsg{Err: fmt.Errorf("failed to load want_to_visit before delete: %w", err)}
		}

		err = db.DeleteWantToVisit(database, wtvID)
		if err != nil {
			return model.ErrorMsg{Err: fmt.Errorf("failed to delete want_to_visit: %w", err)}
		}
		return model.DeleteWantToVisitMsg{ID: wtvID, Deleted: wtv}
	}
}

func convertToVisitCmd(database *sql.DB, wtvID int64) tea.Cmd {
	return func() tea.Msg {
		wtv, err := db.GetWantToVisit(database, wtvID)
		if err != nil {
			return model.ErrorMsg{Err: fmt.Errorf("failed to load convert source: %w", err)}
		}
		restaurantID, err := db.ConvertWantToVisitToVisit(database, wtvID)
		if err != nil {
			return model.ErrorMsg{Err: fmt.Errorf("failed to convert: %w", err)}
		}
		return model.ConvertToVisitMsg{
			WantToVisitID: wtvID,
			RestaurantID:  restaurantID,
			Deleted:       wtv,
		}
	}
}

// Local message type for want_to_visit detail
type wantToVisitDetailLoadedMsg struct {
	entry      model.WantToVisit
	restaurant model.Restaurant
}
