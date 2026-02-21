package ui

import (
	"fmt"
	"sort"
	"strings"
	"toni/internal/model"
	"toni/internal/util"

	"github.com/charmbracelet/lipgloss"
)

type wantToVisitColumn struct {
	key    string
	label  string
	width  int
	hidden bool
}

// WantToVisitModel represents the want to visit list screen.
type WantToVisitModel struct {
	allEntries []model.WantToVisitRow
	entries    []model.WantToVisitRow
	cursor     int
	offset     int

	columns      []wantToVisitColumn
	activeColumn int
	sortKey      string
	sortDesc     bool
	filterKey    string
	filterValue  string
}

// NewWantToVisitModel creates a new want to visit list model.
func NewWantToVisitModel(entries []model.WantToVisitRow) *WantToVisitModel {
	return &WantToVisitModel{
		allEntries: append([]model.WantToVisitRow(nil), entries...),
		entries:    append([]model.WantToVisitRow(nil), entries...),
		columns: []wantToVisitColumn{
			{key: "name", label: "name", width: 24},
			{key: "address", label: "address", width: 20},
			{key: "city", label: "city", width: 14},
			{key: "area", label: "area", width: 14},
			{key: "cuisine", label: "cuisine", width: 14},
			{key: "price", label: "price", width: 10},
			{key: "priority", label: "priority", width: 12},
			{key: "notes", label: "notes", width: 24},
		},
	}
}

func (m *WantToVisitModel) ApplyPrefs(prefs TablePrefs) {
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

func (m *WantToVisitModel) Prefs() TablePrefs {
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

func (m *WantToVisitModel) rebuild() {
	entries := append([]model.WantToVisitRow(nil), m.allEntries...)

	if m.filterKey != "" && m.filterValue != "" {
		filtered := make([]model.WantToVisitRow, 0, len(entries))
		target := strings.ToLower(strings.TrimSpace(m.filterValue))
		for _, r := range entries {
			if strings.EqualFold(strings.TrimSpace(m.getValue(r, m.filterKey)), target) {
				filtered = append(filtered, r)
			}
		}
		entries = filtered
	}

	if m.sortKey != "" {
		sort.SliceStable(entries, func(i, j int) bool {
			left := strings.ToLower(m.getValue(entries[i], m.sortKey))
			right := strings.ToLower(m.getValue(entries[j], m.sortKey))
			if left == right {
				return entries[i].ID > entries[j].ID
			}
			if m.sortDesc {
				return left > right
			}
			return left < right
		})
	}

	m.entries = entries
	m.clampCursor()
}

func (m *WantToVisitModel) clampCursor() {
	if len(m.entries) == 0 {
		m.cursor = 0
		m.offset = 0
		return
	}
	if m.cursor >= len(m.entries) {
		m.cursor = len(m.entries) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.offset > m.cursor {
		m.offset = m.cursor
	}
}

func (m *WantToVisitModel) getValue(row model.WantToVisitRow, key string) string {
	switch key {
	case "name":
		return row.RestaurantName
	case "address":
		return row.Address
	case "city":
		return row.City
	case "area":
		return row.Neighborhood
	case "cuisine":
		return row.Cuisine
	case "price":
		return row.PriceRange
	case "priority":
		if row.Priority == nil {
			return ""
		}
		return fmt.Sprintf("%02d", *row.Priority)
	case "notes":
		return row.Notes
	default:
		return ""
	}
}

func (m *WantToVisitModel) visibleColumnIndexes() []int {
	var idxs []int
	for i, c := range m.columns {
		if !c.hidden {
			idxs = append(idxs, i)
		}
	}
	return idxs
}

func (m *WantToVisitModel) ensureVisibleActiveColumn() {
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

func (m *WantToVisitModel) NextColumn() {
	start := m.activeColumn
	for {
		m.activeColumn = (m.activeColumn + 1) % len(m.columns)
		if !m.columns[m.activeColumn].hidden || m.activeColumn == start {
			return
		}
	}
}

func (m *WantToVisitModel) PrevColumn() {
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

func (m *WantToVisitModel) JumpToColumn(number int) bool {
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

func (m *WantToVisitModel) SortActiveColumn(desc bool) {
	m.sortKey = m.columns[m.activeColumn].key
	m.sortDesc = desc
	m.rebuild()
}

func (m *WantToVisitModel) CycleSortActiveColumn() string {
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

func (m *WantToVisitModel) HideActiveColumn() bool {
	if len(m.visibleColumnIndexes()) <= 1 {
		return false
	}
	m.columns[m.activeColumn].hidden = true
	m.ensureVisibleActiveColumn()
	return true
}

func (m *WantToVisitModel) ShowAllColumns() {
	for i := range m.columns {
		m.columns[i].hidden = false
	}
}

func (m *WantToVisitModel) FilterBySelectedValue() bool {
	if len(m.entries) == 0 {
		return false
	}
	key := m.columns[m.activeColumn].key
	value := strings.TrimSpace(m.getValue(m.entries[m.cursor], key))
	if value == "" {
		return false
	}
	m.filterKey = key
	m.filterValue = value
	m.rebuild()
	return true
}

func (m *WantToVisitModel) ClearFilter() bool {
	if m.filterKey == "" {
		return false
	}
	m.filterKey = ""
	m.filterValue = ""
	m.rebuild()
	return true
}

func (m *WantToVisitModel) CycleFilterBySelectedValue() string {
	if len(m.entries) == 0 {
		return "No rows to filter"
	}
	key := m.columns[m.activeColumn].key
	value := strings.TrimSpace(m.getValue(m.entries[m.cursor], key))
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

func (m *WantToVisitModel) TableMeta() string {
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

// CursorDown moves the cursor down.
func (m *WantToVisitModel) CursorDown() {
	if m.cursor < len(m.entries)-1 {
		m.cursor++
		if m.cursor >= m.offset+10 {
			m.offset++
		}
	}
}

// CursorUp moves the cursor up.
func (m *WantToVisitModel) CursorUp() {
	if m.cursor > 0 {
		m.cursor--
		if m.cursor < m.offset {
			m.offset--
		}
	}
}

// JumpToTop jumps to the top of the list.
func (m *WantToVisitModel) JumpToTop() {
	m.cursor = 0
	m.offset = 0
}

// JumpToBottom jumps to the bottom of the list.
func (m *WantToVisitModel) JumpToBottom() {
	if len(m.entries) > 0 {
		m.cursor = len(m.entries) - 1
		if m.cursor >= 10 {
			m.offset = m.cursor - 9
		}
	}
}

// SelectedEntry returns the currently selected entry.
func (m *WantToVisitModel) SelectedEntry() *model.WantToVisitRow {
	if len(m.entries) == 0 || m.cursor >= len(m.entries) {
		return nil
	}
	return &m.entries[m.cursor]
}

// View renders the want to visit list.
func (m *WantToVisitModel) View(width, height int) string {
	if len(m.entries) == 0 {
		emptyMsg := `    No places yet.
    Press  a  to add one!`
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
	var rows []string
	for i := m.offset; i < len(m.entries) && i < m.offset+visibleHeight; i++ {
		entry := m.entries[i]
		style := NormalRowStyle
		if i == m.cursor {
			style = SelectedRowStyle
		}

		cells := make([]string, 0, len(visible))
		aligns := make([]lipgloss.Position, 0, len(visible))
		for _, idx := range visible {
			col := m.columns[idx]
			switch col.key {
			case "name":
				cells = append(cells, util.TruncateString(entry.RestaurantName, col.width))
				aligns = append(aligns, lipgloss.Center)
			case "address":
				cells = append(cells, util.TruncateString(entry.Address, col.width))
				aligns = append(aligns, lipgloss.Center)
			case "city":
				cells = append(cells, util.TruncateString(entry.City, col.width))
				aligns = append(aligns, lipgloss.Center)
			case "area":
				cells = append(cells, util.TruncateString(entry.Neighborhood, col.width))
				aligns = append(aligns, lipgloss.Center)
			case "cuisine":
				cells = append(cells, util.TruncateString(entry.Cuisine, col.width))
				aligns = append(aligns, lipgloss.Center)
			case "price":
				priceCell := entry.PriceRange
				if priceCell == "" {
					priceCell = "—"
				}
				cells = append(cells, priceCell)
				aligns = append(aligns, lipgloss.Center)
			case "priority":
				priorityStr := "—"
				if entry.Priority != nil {
					priorityStr = fmt.Sprintf("%d/5", *entry.Priority)
					color := ColorMuted
					if *entry.Priority >= 4 {
						color = ColorRed
					} else if *entry.Priority >= 3 {
						color = ColorYellow
					}
					priorityStr = lipgloss.NewStyle().Foreground(color).Render(priorityStr)
				}
				cells = append(cells, priorityStr)
				aligns = append(aligns, lipgloss.Center)
			case "notes":
				cells = append(cells, util.TruncateString(entry.Notes, col.width))
				aligns = append(aligns, lipgloss.Left)
			}
		}

		rows = append(rows, renderTableRowWithAligns(cells, widths, aligns, style))
	}

	filterInfo := ""
	if m.filterKey != "" {
		filterInfo = fmt.Sprintf("  ·  filtered: %d/%d", len(m.entries), len(m.allEntries))
	}
	meta := m.TableMeta()
	if meta != "" {
		meta = "  ·  " + meta
	}
	rowPos := ""
	if len(m.entries) > 0 {
		rowPos = fmt.Sprintf("  ·  row %d/%d", m.cursor+1, len(m.entries))
	}
	status := StatusBarStyle.Render(fmt.Sprintf("Total places: %d%s%s%s", len(m.entries), rowPos, filterInfo, meta))

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
