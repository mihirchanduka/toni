package ui

import (
	"fmt"
	"sort"
	"strings"
	"toni/internal/model"
	"toni/internal/util"

	"github.com/charmbracelet/lipgloss"
)

type restaurantColumn struct {
	key    string
	label  string
	width  int
	hidden bool
}

// RestaurantsModel represents the restaurants list screen.
type RestaurantsModel struct {
	allRows []model.RestaurantRow
	rows    []model.RestaurantRow
	cursor  int
	offset  int

	viewportHeight int

	columns      []restaurantColumn
	activeColumn int
	sortKey      string
	sortDesc     bool
	filterKey    string
	filterValue  string
}

// NewRestaurantsModel creates a new restaurants model.
func NewRestaurantsModel(rows []model.RestaurantRow) *RestaurantsModel {
	return &RestaurantsModel{
		allRows: append([]model.RestaurantRow(nil), rows...),
		rows:    append([]model.RestaurantRow(nil), rows...),
		columns: []restaurantColumn{
			{key: "name", label: "name", width: 24},
			{key: "address", label: "address", width: 20},
			{key: "city", label: "city", width: 14},
			{key: "area", label: "area", width: 14},
			{key: "cuisine", label: "cuisine", width: 14},
			{key: "price", label: "price", width: 10},
			{key: "rating", label: "rating", width: 8},
			{key: "visits", label: "visits", width: 8},
			{key: "last", label: "last", width: 14},
		},
	}
}

func (m *RestaurantsModel) ApplyPrefs(prefs TablePrefs) {
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

func (m *RestaurantsModel) Prefs() TablePrefs {
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

func (m *RestaurantsModel) rebuild() {
	rows := append([]model.RestaurantRow(nil), m.allRows...)

	if m.filterKey != "" && m.filterValue != "" {
		filtered := make([]model.RestaurantRow, 0, len(rows))
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

func (m *RestaurantsModel) clampCursor() {
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

func (m *RestaurantsModel) getValue(row model.RestaurantRow, key string) string {
	switch key {
	case "name":
		return row.Name
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
	case "rating":
		if row.AvgRating == nil {
			return ""
		}
		return fmt.Sprintf("%05.2f", *row.AvgRating)
	case "visits":
		return fmt.Sprintf("%06d", row.VisitCount)
	case "last":
		return row.LastVisit
	default:
		return ""
	}
}

func (m *RestaurantsModel) visibleColumnIndexes() []int {
	var idxs []int
	for i, c := range m.columns {
		if !c.hidden {
			idxs = append(idxs, i)
		}
	}
	return idxs
}

func (m *RestaurantsModel) ensureVisibleActiveColumn() {
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

func (m *RestaurantsModel) NextColumn() {
	start := m.activeColumn
	for {
		m.activeColumn = (m.activeColumn + 1) % len(m.columns)
		if !m.columns[m.activeColumn].hidden || m.activeColumn == start {
			return
		}
	}
}

func (m *RestaurantsModel) PrevColumn() {
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

func (m *RestaurantsModel) JumpToColumn(number int) bool {
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

func (m *RestaurantsModel) SortActiveColumn(desc bool) {
	m.sortKey = m.columns[m.activeColumn].key
	m.sortDesc = desc
	m.rebuild()
}

func (m *RestaurantsModel) CycleSortActiveColumn() string {
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

func (m *RestaurantsModel) HideActiveColumn() bool {
	if len(m.visibleColumnIndexes()) <= 1 {
		return false
	}
	m.columns[m.activeColumn].hidden = true
	m.ensureVisibleActiveColumn()
	return true
}

func (m *RestaurantsModel) ShowAllColumns() {
	for i := range m.columns {
		m.columns[i].hidden = false
	}
}

func (m *RestaurantsModel) FilterBySelectedValue() bool {
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

func (m *RestaurantsModel) ClearFilter() bool {
	if m.filterKey == "" {
		return false
	}
	m.filterKey = ""
	m.filterValue = ""
	m.rebuild()
	return true
}

func (m *RestaurantsModel) CycleFilterBySelectedValue() string {
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

func (m *RestaurantsModel) TableMeta() string {
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

// View renders the restaurants list.
func (m *RestaurantsModel) View(width, height int) string {
	if len(m.rows) == 0 {
		emptyMsg := `    No restaurants yet.
    Press  a  to add your first restaurant!`
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

	totalRating := 0.0
	ratedCount := 0
	for _, r := range m.rows {
		if r.AvgRating != nil {
			totalRating += *r.AvgRating
			ratedCount++
		}
	}

	for i := m.offset; i < len(m.rows) && i < m.offset+visibleHeight; i++ {
		row := m.rows[i]
		style := NormalRowStyle
		if i == m.cursor {
			style = SelectedRowStyle
		}

		cells := make([]string, 0, len(visible))
		for _, idx := range visible {
			col := m.columns[idx]
			switch col.key {
			case "name":
				cells = append(cells, util.TruncateString(row.Name, col.width))
			case "address":
				cells = append(cells, util.TruncateString(row.Address, col.width))
			case "city":
				cells = append(cells, util.TruncateString(row.City, col.width))
			case "area":
				cells = append(cells, util.TruncateString(row.Neighborhood, col.width))
			case "cuisine":
				cells = append(cells, util.TruncateString(row.Cuisine, col.width))
			case "price":
				priceCell := row.PriceRange
				if priceCell == "" {
					priceCell = "—"
				}
				cells = append(cells, priceCell)
			case "rating":
				avgRatingCell := "—"
				if row.AvgRating != nil {
					avgRatingCell = lipgloss.NewStyle().Foreground(ColorYellow).Render(fmt.Sprintf("%.1f★", *row.AvgRating))
				}
				cells = append(cells, avgRatingCell)
			case "visits":
				cells = append(cells, fmt.Sprintf("%d", row.VisitCount))
			case "last":
				lastVisitCell := "—"
				if row.LastVisit != "" {
					lastVisitCell = util.FormatDateHuman(row.LastVisit)
				}
				cells = append(cells, lastVisitCell)
			}
		}
		rows = append(rows, renderTableRow(cells, widths, style))
	}

	overallAvg := ""
	if ratedCount > 0 {
		overallAvg = fmt.Sprintf("  ·  avg rating %.1f", totalRating/float64(ratedCount))
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
	status := StatusBarStyle.Render(fmt.Sprintf("%d restaurants%s%s%s%s", len(m.rows), rowPos, overallAvg, filterInfo, meta))

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
func (m *RestaurantsModel) MoveDown() {
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
func (m *RestaurantsModel) MoveUp() {
	if m.cursor > 0 {
		m.cursor--
		if m.cursor < m.offset {
			m.offset--
		}
	}
}

// JumpToTop jumps to the first item.
func (m *RestaurantsModel) JumpToTop() {
	m.cursor = 0
	m.offset = 0
}

// JumpToBottom jumps to the last item.
func (m *RestaurantsModel) JumpToBottom() {
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
func (m *RestaurantsModel) HalfPageDown(pageSize int) {
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
func (m *RestaurantsModel) HalfPageUp(pageSize int) {
	halfPage := pageSize / 2
	m.cursor -= halfPage
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
}
