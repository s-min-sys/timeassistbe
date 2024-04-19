package timeassist

type TimeType int

const (
	TimeTypeBegin TimeType = iota
	TimeTypeOnce
	RecycleTimeTypeYear
	RecycleTimeTypeMonth
	RecycleTimeTypeWeek
	RecycleTimeTypeDay
	RecycleTimeTypeHour
	RecycleTimeTypeMinute
	TimeTypeEnd
)
