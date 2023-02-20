package timeassist

import (
	"time"
)

type ValidRange struct {
	Start int `yaml:"Start" json:"start"`
	End   int `yaml:"End" json:"end"`
}

type ValidRanges struct {
	ValidRanges []ValidRange `yaml:"ValidRanges" json:"valid_ranges,omitempty"`
}

func (o *ValidRanges) IsValid(value int) bool {
	if len(o.ValidRanges) == 0 {
		return true
	}

	for _, hoursRange := range o.ValidRanges {
		if value >= hoursRange.Start && value < hoursRange.End {
			return true
		}
	}

	return false
}

type ValidTime struct {
	ValidMonthsInYear *ValidRanges `yaml:"ValidMonthsInYear,omitempty" json:"valid_months_in_year,omitempty"`
	ValidWeeksInMonth *ValidRanges `yaml:"ValidWeeksInMonth,omitempty" json:"valid_weeks_in_month,omitempty"`
	ValidDaysInMonth  *ValidRanges `yaml:"ValidDaysInMonth,omitempty" json:"valid_days_in_month,omitempty"`
	ValidDaysInWeek   *ValidRanges `yaml:"ValidDaysInWeek,omitempty" json:"valid_days_in_week,omitempty"`
	ValidHoursInDay   *ValidRanges `yaml:"ValidHoursInDay,omitempty" json:"valid_hours_in_day,omitempty"`
}

func (vt *ValidTime) Reset() {
	vt.ValidMonthsInYear = nil
	vt.ValidWeeksInMonth = nil
	vt.ValidDaysInMonth = nil
	vt.ValidDaysInWeek = nil
	vt.ValidHoursInDay = nil
}

func (vt *ValidTime) FindAfterTime(t time.Time) (tIsOk bool, nextT time.Time) {
	tIsOk = true
	nextT = t

RETRY:
	if vt.ValidMonthsInYear != nil {
		for !vt.ValidMonthsInYear.IsValid(int(nextT.Month())) {
			tIsOk = false

			nextT = MonthStart(MonthAdd(nextT, 1))
		}
	}

	if vt.ValidWeeksInMonth != nil {
		var needRetry bool

		for !vt.ValidWeeksInMonth.IsValid(WeekIndexInMonth(nextT)) {
			needRetry = true

			nextT = WeekStart(WeekAdd(nextT, 1))
		}

		if needRetry {
			tIsOk = false

			goto RETRY
		}
	}

	if vt.ValidDaysInMonth != nil {
		var needRetry bool

		for !vt.ValidDaysInMonth.IsValid(nextT.Day()) {
			needRetry = true

			nextT = DayStart(DayAdd(nextT, 1))
		}

		if needRetry {
			tIsOk = false

			goto RETRY
		}
	}

	if vt.ValidDaysInWeek != nil {
		var needRetry bool

		for !vt.ValidDaysInWeek.IsValid(int(nextT.Weekday())) {
			needRetry = true

			nextT = DayStart(DayAdd(nextT, 1))
		}

		if needRetry {
			tIsOk = false

			goto RETRY
		}
	}

	if vt.ValidHoursInDay != nil {
		var needRetry bool

		for !vt.ValidHoursInDay.IsValid(nextT.Hour()) {
			needRetry = true

			nextT = HourStart(HourAdd(nextT, 1))
		}

		if needRetry {
			tIsOk = false

			goto RETRY
		}
	}

	return
}
