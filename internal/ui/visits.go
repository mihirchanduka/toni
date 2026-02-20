package ui

import (
	"fmt"
	"sort"
	"strings"
	"toni/internal/model"
	"toni/internal/util"

	"github.com/charmbracelet/lipgloss"
)

type visitColumn struct {
	key    string
	label  string
	width  int
	hidden bool
}

// VisitsModel represents the visits list screen.
type VisitsModel struct {
	allRows []model.VisitRow
	rows    []model.VisitRow
	cursor  int
	offset  int

	columns      []visitColumn
	activeColumn int
	sortKey      string
	sortDesc     bool
	filterKey    string
	filterValue  string
}

// NewVisitsModel creates a new visits model.
func NewVisitsModel(rows []model.VisitRow) *VisitsModel {
	m := &VisitsModel{
		allRows: append([]model.VisitRow(nil), rows...),
		rows:    append([]model.VisitRow(nil), rows...),
		columns: []visitColumn{
			{key: "date", label: "date", width: 12},
			{key: "name", label: "name", width: 24},
			{key: "address", label: "address", width: 20},
			{key: "city", label: "city", width: 14},
			{key: "price", label: "price", width: 10},
			{key: "rating", label: "rating", width: 8},
			{key: "return", label: "return", width: 8},
			{key: "notes", label: "notes", width: 24},
		},
		activeColumn: 0,
	}
	return m
}

func (m *VisitsModel) ApplyPrefs(prefs TablePrefs) {
	if prefs.SortKey != "" {
		m.sortKey = prefs.SortKey
		m.sortDesc = prefs.SortDesc
	}
	hidden := make(map[string]bool, len(prefs.HiddenColumns))
	for _, c := range prefs.HiddenColumns {
		hidden[c] = true
	}
	for i := range m.columns {
		m.columns[i].hidden = hidden[m.columns[i].key]
	}
	if prefs.ActiveColumn != "" {
		for i, c := range m.columns {
			if c.key == prefs.ActiveColumn {
				m.activeColumn = i
				break
			}
		}
	}
	m.ensureVisibleActiveColumn()
	m.rebuild()
}

func (m *VisitsModel) Prefs() TablePrefs {
	var hidden []string
	for _, c := range m.columns {
		if c.hidden {
			hidden = append(hidden, c.key)
		}
	}
	return TablePrefs{
		SortKey:       m.sortKey,
		SortDesc:      m.sortDesc,
		HiddenColumns: hidden,
		ActiveColumn:  m.columns[m.activeColumn].key,
	}
}

func (m *VisitsModel) rebuild() {
	rows := append([]model.VisitRow(nil), m.allRows...)

	if m.filterKey != "" && m.filterValue != "" {
		filtered := make([]model.VisitRow, 0, len(rows))
		target := strings.ToLower(strings.TrimSpace(m.filterValue))
		for _, r := range rows {
			if strings.EqualFold(strings.TrimSpace(m.getValue(r, m.filterKey)), target) {
				filtered = append(filtered, r)
			}
		}
		rows = filtered
	}

	if m.sortKey != "" {
		sort.SliceStable(rows, func(i, j int) bool {
			left := strings.ToLower(m.getValue(rows[i], m.sortKey))
			right := strings.ToLower(m.getValue(rows[j], m.sortKey))
			if left == right {
				return rows[i].ID > rows[j].ID
			}
			if m.sortDesc {
				return left > right
			}
			return left < right
		})
	}

	m.rows = rows
	m.clampCursor()
}

func (m *VisitsModel) clampCursor() {
	if len(m.rows) == 0 {
		m.cursor = 0
		m.offset = 0
		return
	}
	if m.cursor >= len(m.rows) {
		m.cursor = len(m.rows) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.offset > m.cursor {
		m.offset = m.cursor
	}
}

func (m *VisitsModel) getValue(row model.VisitRow, key string) string {
	switch key {
	case "date":
		return row.VisitedOn
	case "name":
		return row.RestaurantName
	case "address":
		return row.Address
	case "city":
		return row.City
	case "price":
		return row.PriceRange
	case "rating":
		if row.Rating == nil {
			return ""
		}
		return fmt.Sprintf("%05.2f", *row.Rating)
	case "return":
		if row.WouldReturn == nil {
			return ""
		}
		if *row.WouldReturn {
			return "yes"
		}
		return "no"
	case "notes":
		return row.Notes
	default:
		return ""
	}
}

func (m *VisitsModel) NextColumn() {
	start := m.activeColumn
	for {
		m.activeColumn = (m.activeColumn + 1) % len(m.columns)
		if !m.columns[m.activeColumn].hidden || m.activeColumn == start {
			return
		}
	}
}

func (m *VisitsModel) PrevColumn() {
	start := m.activeColumn
	for {
		m.activeColumn--
		if m.activeColumn < 0 {
			m.activeColumn = len(m.columns) - 1
		}
		if !m.columns[m.activeColumn].hidden || m.activeColumn == start {
			return
		}
	}
}

func (m *VisitsModel) JumpToColumn(number int) bool {
	if number < 1 || number > len(m.columns) {
		return false
	}
	idx := number - 1
	if m.columns[idx].hidden {
		return false
	}
	m.activeColumn = idx
	return true
}

func (m *VisitsModel) SortActiveColumn(desc bool) {
	m.sortKey = m.columns[m.activeColumn].key
	m.sortDesc = desc
	m.rebuild()
}

func (m *VisitsModel) HideActiveColumn() bool {
	if len(m.visibleColumnIndexes()) <= 1 {
		return false
	}
	m.columns[m.activeColumn].hidden = true
	m.ensureVisibleActiveColumn()
	return true
}

func (m *VisitsModel) ShowAllColumns() {
	for i := range m.columns {
		m.columns[i].hidden = false
	}
}

func (m *VisitsModel) FilterBySelectedValue() bool {
	if len(m.rows) == 0 {
		return false
	}
	key := m.columns[m.activeColumn].key
	value := strings.TrimSpace(m.getValue(m.rows[m.cursor], key))
	if value == "" {
		return false
	}
	m.filterKey = key
	m.filterValue = value
	m.rebuild()
	return true
}

func (m *VisitsModel) ClearFilter() bool {
	if m.filterKey == "" {
		return false
	}
	m.filterKey = ""
	m.filterValue = ""
	m.rebuild()
	return true
}

func (m *VisitsModel) TableMeta() string {
	col := strings.ToUpper(m.columns[m.activeColumn].label)
	parts := []string{fmt.Sprintf("col %s", col)}
	if m.sortKey != "" {
		order := "asc"
		if m.sortDesc {
			order = "desc"
		}
		parts = append(parts, fmt.Sprintf("sort %s %s", strings.ToUpper(m.sortKey), order))
	}
	if m.filterKey != "" {
		parts = append(parts, fmt.Sprintf("filter %s=%q", strings.ToUpper(m.filterKey), m.filterValue))
	}
	return strings.Join(parts, "  ·  ")
}

func (m *VisitsModel) visibleColumnIndexes() []int {
	var idxs []int
	for i, c := range m.columns {
		if !c.hidden {
			idxs = append(idxs, i)
		}
	}
	return idxs
}

func (m *VisitsModel) ensureVisibleActiveColumn() {
	if !m.columns[m.activeColumn].hidden {
		return
	}
	for i := range m.columns {
		if !m.columns[i].hidden {
			m.activeColumn = i
			return
		}
	}
	m.columns[0].hidden = false
	m.activeColumn = 0
}

// View renders the visits list.
func (m *VisitsModel) View(width, height int) string {
	if len(m.rows) == 0 {
		emptyMsg := `    No visits yet.
    Press  a  to log your first meal.`
		return EmptyStateStyle.
			Width(width).
			Height(height).
			Render(emptyMsg)
	}

	visible := m.visibleColumnIndexes()
	if len(visible) == 0 {
		return EmptyStateStyle.Width(width).Height(height).Render("No visible columns. Press C to show all columns.")
	}

	widths := make([]int, 0, len(visible))
	headers := make([]string, 0, len(visible))
	totalFixed := 0
	for _, idx := range visible {
		col := m.columns[idx]
		label := strings.ToUpper(col.label)
		if idx == m.activeColumn {
			label = "❋ " + label
		}
		if m.sortKey == col.key {
			if m.sortDesc {
				label += " ↓"
			} else {
				label += " ↑"
			}
		}
		cellWidth := max(col.width, lipgloss.Width(label)+2)
		totalFixed += cellWidth
		widths = append(widths, cellWidth)
		headers = append(headers, label)
	}

	if len(widths) > 0 {
		extra := width - totalFixed - 4
		if extra > 0 {
			widths[len(widths)-1] += extra
		}
	}

	headerStyle := TableHeaderStyle.Bold(true)
	header := renderTableRow(headers, widths, headerStyle)

	visibleHeight := height - 3
	var rows []string

	for i := m.offset; i < len(m.rows) && i < m.offset+visibleHeight; i++ {
		row := m.rows[i]
		style := NormalRowStyle
		if i%2 == 1 {
			style = style.Background(lipgloss.Color("#232B24"))
		}
		if i == m.cursor {
			style = SelectedRowStyle
		}

		cells := make([]string, 0, len(visible))
		for _, idx := range visible {
			col := m.columns[idx]
			switch col.key {
			case "date":
				cells = append(cells, util.FormatDateHuman(row.VisitedOn))
			case "name":
				cells = append(cells, util.TruncateString(row.RestaurantName, col.width-2))
			case "address":
				cells = append(cells, util.TruncateString(row.Address, col.width-2))
			case "city":
				cells = append(cells, util.TruncateString(row.City, col.width-2))
			case "price":
				priceCell := row.PriceRange
				if priceCell == "" {
					priceCell = "—"
				}
				cells = append(cells, priceCell)
			case "rating":
				ratingCell := util.FormatRatingWithStar(row.Rating)
				if row.Rating != nil {
					ratingCell = lipgloss.NewStyle().Foreground(ColorYellow).Render(ratingCell)
				}
				cells = append(cells, ratingCell)
			case "return":
				returnCell := util.FormatWouldReturnSymbol(row.WouldReturn)
				returnStyle := lipgloss.NewStyle().Foreground(ColorMuted)
				if row.WouldReturn != nil {
					if *row.WouldReturn {
						returnStyle = returnStyle.Foreground(ColorGreen)
					} else {
						returnStyle = returnStyle.Foreground(ColorRed)
					}
				}
				cells = append(cells, returnStyle.Render(returnCell))
			case "notes":
				cells = append(cells, util.TruncateString(row.Notes, col.width-2))
			}
		}

		rows = append(rows, renderTableRow(cells, widths, style))
	}

	filterInfo := ""
	if m.filterKey != "" {
		filterInfo = fmt.Sprintf("  ·  filtered: %d/%d", len(m.rows), len(m.allRows))
	}
	meta := m.TableMeta()
	if meta != "" {
		meta = "  ·  " + meta
	}
	status := StatusBarStyle.Render(fmt.Sprintf("Total visits: %d%s%s", len(m.rows), filterInfo, meta))

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		strings.Join(rows, "\n"),
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		"",
		status,
	)
}

// MoveDown moves the cursor down.
func (m *VisitsModel) MoveDown() {
	if m.cursor < len(m.rows)-1 {
		m.cursor++
		if m.cursor >= m.offset+10 {
			m.offset++
		}
	}
}

// MoveUp moves the cursor up.
func (m *VisitsModel) MoveUp() {
	if m.cursor > 0 {
		m.cursor--
		if m.cursor < m.offset {
			m.offset--
		}
	}
}

// JumpToTop jumps to the first item.
func (m *VisitsModel) JumpToTop() {
	m.cursor = 0
	m.offset = 0
}

// JumpToBottom jumps to the last item.
func (m *VisitsModel) JumpToBottom() {
	if len(m.rows) > 0 {
		m.cursor = len(m.rows) - 1
		if m.cursor >= 10 {
			m.offset = m.cursor - 9
		}
	}
}

// HalfPageDown moves down half a page.
func (m *VisitsModel) HalfPageDown(pageSize int) {
	halfPage := pageSize / 2
	m.cursor += halfPage
	if m.cursor >= len(m.rows) {
		m.cursor = len(m.rows) - 1
	}
	if m.cursor >= m.offset+10 {
		m.offset = m.cursor - 9
	}
}

// HalfPageUp moves up half a page.
func (m *VisitsModel) HalfPageUp(pageSize int) {
	halfPage := pageSize / 2
	m.cursor -= halfPage
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
}

// Helper function to render a table row
func renderTableRow(cells []string, widths []int, style lipgloss.Style) string {
	var parts []string
	for i, cell := range cells {
		if i >= len(widths) {
			continue
		}
		parts = append(parts, style.Width(widths[i]).Render(cell))
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, parts...)
}
