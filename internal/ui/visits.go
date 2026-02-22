package ui

import (
	"fmt"
	"sort"
	"strings"
	"toni/internal/model"
	"toni/internal/util"
	"unicode"

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

	viewportHeight int

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

func (m *VisitsModel) CycleSortActiveColumn() string {
	activeKey := m.columns[m.activeColumn].key
	activeLabel := strings.ToUpper(m.columns[m.activeColumn].label)
	switch {
	case m.sortKey != activeKey:
		m.sortKey = activeKey
		m.sortDesc = false
		m.rebuild()
		return fmt.Sprintf("Sorted %s ascending", activeLabel)
	case !m.sortDesc:
		m.sortDesc = true
		m.rebuild()
		return fmt.Sprintf("Sorted %s descending", activeLabel)
	default:
		m.sortKey = ""
		m.sortDesc = false
		m.rebuild()
		return "Sorting cleared"
	}
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

func (m *VisitsModel) CycleFilterBySelectedValue() string {
	if len(m.rows) == 0 {
		return "No rows to filter"
	}
	key := m.columns[m.activeColumn].key
	value := strings.TrimSpace(m.getValue(m.rows[m.cursor], key))
	if value == "" {
		return "No filterable value in selected cell"
	}

	if m.filterKey == key && strings.EqualFold(strings.TrimSpace(m.filterValue), value) {
		m.filterKey = ""
		m.filterValue = ""
		m.rebuild()
		return "Filter cleared"
	}

	m.filterKey = key
	m.filterValue = value
	m.rebuild()
	return "Filter applied from selected value"
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
		label := formatHeaderLabel(col.label)
		if idx == m.activeColumn {
			label = renderActiveHeaderLabel(label)
		}
		if m.sortKey == col.key {
			if m.sortDesc {
				label += " ↓"
			} else {
				label += " ↑"
			}
		}
		cellWidth := max(col.width+2, lipgloss.Width(label)+4)
		totalFixed += cellWidth
		widths = append(widths, cellWidth)
		headers = append(headers, label)
	}

	if len(widths) > 0 {
		sepTotal := (len(widths) - 1) * tableSeparatorWidth()
		extra := width - totalFixed - sepTotal - 2
		if extra > 0 {
			widths[len(widths)-1] += extra
		}
	}

	headerStyle := TableHeaderStyle.Bold(true)
	header := renderTableRow(headers, widths, headerStyle)
	divider := renderTableDivider(widths)

	visibleHeight := height - 3
	m.viewportHeight = visibleHeight
	var rows []string

	for i := m.offset; i < len(m.rows) && i < m.offset+visibleHeight; i++ {
		row := m.rows[i]
		style := NormalRowStyle
		if i == m.cursor {
			style = SelectedRowStyle
		}

		cells := make([]string, 0, len(visible))
		aligns := make([]lipgloss.Position, 0, len(visible))
		for _, idx := range visible {
			col := m.columns[idx]
			switch col.key {
			case "date":
				cells = append(cells, util.FormatDateHuman(row.VisitedOn))
				aligns = append(aligns, lipgloss.Center)
			case "name":
				cells = append(cells, util.TruncateString(row.RestaurantName, col.width))
				aligns = append(aligns, lipgloss.Center)
			case "address":
				cells = append(cells, util.TruncateString(row.Address, col.width))
				aligns = append(aligns, lipgloss.Center)
			case "city":
				cells = append(cells, util.TruncateString(row.City, col.width))
				aligns = append(aligns, lipgloss.Center)
			case "price":
				priceCell := row.PriceRange
				if priceCell == "" {
					priceCell = "—"
				}
				cells = append(cells, priceCell)
				aligns = append(aligns, lipgloss.Center)
			case "rating":
				ratingCell := util.FormatRatingWithStar(row.Rating)
				if row.Rating != nil {
					ratingCell = lipgloss.NewStyle().Foreground(ColorYellow).Render(ratingCell)
				}
				cells = append(cells, ratingCell)
				aligns = append(aligns, lipgloss.Center)
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
				aligns = append(aligns, lipgloss.Center)
			case "notes":
				cells = append(cells, util.TruncateString(row.Notes, col.width))
				aligns = append(aligns, lipgloss.Left)
			}
		}

		rows = append(rows, renderTableRowWithAligns(cells, widths, aligns, style))
	}

	filterInfo := ""
	if m.filterKey != "" {
		filterInfo = fmt.Sprintf("  ·  filtered: %d/%d", len(m.rows), len(m.allRows))
	}
	meta := m.TableMeta()
	if meta != "" {
		meta = "  ·  " + meta
	}
	rowPos := ""
	if len(m.rows) > 0 {
		rowPos = fmt.Sprintf("  ·  row %d/%d", m.cursor+1, len(m.rows))
	}
	status := StatusBarStyle.Render(fmt.Sprintf("Total visits: %d%s%s%s", len(m.rows), rowPos, filterInfo, meta))

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		divider,
		strings.Join(rows, "\n"),
	)
	statusHeight := lipgloss.Height(status)
	contentHeight := lipgloss.Height(content)
	spacerHeight := max(0, height-contentHeight-statusHeight)
	spacer := lipgloss.NewStyle().Height(spacerHeight).Render("")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		spacer,
		status,
	)
}

// MoveDown moves the cursor down.
func (m *VisitsModel) MoveDown() {
	if m.cursor < len(m.rows)-1 {
		m.cursor++
		vh := m.viewportHeight
		if vh == 0 {
			vh = 10
		}
		if m.cursor >= m.offset+vh {
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
		vh := m.viewportHeight
		if vh == 0 {
			vh = 10
		}
		if m.cursor >= vh {
			m.offset = m.cursor - vh + 1
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
	vh := m.viewportHeight
	if vh == 0 {
		vh = 10
	}
	if m.cursor >= m.offset+vh {
		m.offset = m.cursor - vh + 1
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
	return renderTableRowWithAligns(cells, widths, nil, style)
}

func renderTableRowWithAligns(cells []string, widths []int, aligns []lipgloss.Position, style lipgloss.Style) string {
	var parts []string
	sep := renderColumnSeparator()
	for i, cell := range cells {
		if i >= len(widths) {
			continue
		}
		align := lipgloss.Center
		if i < len(aligns) {
			align = aligns[i]
		}
		cellStyle := style.Copy().Width(widths[i]).Align(align)
		parts = append(parts, cellStyle.Render(cell))
		if i < len(cells)-1 {
			parts = append(parts, sep)
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, parts...)
}

func renderColumnSeparator() string {
	return TableSeparatorStyle.Render(" │ ")
}

func tableSeparatorWidth() int {
	return lipgloss.Width(renderColumnSeparator())
}

func renderTableDivider(widths []int) string {
	if len(widths) == 0 {
		return ""
	}
	sep := TableDividerStyle.Render("─┼─")
	parts := make([]string, 0, len(widths)*2)
	for i, width := range widths {
		if width < 1 {
			width = 1
		}
		parts = append(parts, TableDividerStyle.Render(strings.Repeat("─", width)))
		if i < len(widths)-1 {
			parts = append(parts, sep)
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, parts...)
}

func formatHeaderLabel(label string) string {
	s := strings.TrimSpace(strings.ToLower(label))
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func renderActiveHeaderLabel(label string) string {
	return lipgloss.NewStyle().Foreground(ColorYellow).Bold(true).Render(label)
}
