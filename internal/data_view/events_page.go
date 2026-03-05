package data_view

type EventsPageData struct {
	PageData
	Featured    []EventView
	MonthGroups []MonthGroup
}
