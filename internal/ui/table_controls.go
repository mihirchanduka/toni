package ui

type tableController interface {
	NextColumn()
	PrevColumn()
	JumpToColumn(number int) bool
	SortActiveColumn(desc bool)
	HideActiveColumn() bool
	ShowAllColumns()
	FilterBySelectedValue() bool
	ClearFilter() bool
	TableMeta() string
}
