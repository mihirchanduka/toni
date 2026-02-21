package ui

type tableController interface {
	NextColumn()
	PrevColumn()
	JumpToColumn(number int) bool
	CycleSortActiveColumn() string
	SortActiveColumn(desc bool)
	HideActiveColumn() bool
	ShowAllColumns()
	CycleFilterBySelectedValue() string
	FilterBySelectedValue() bool
	ClearFilter() bool
	TableMeta() string
}
