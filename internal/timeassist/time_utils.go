package timeassist

import (
	"fmt"
	"time"

	"github.com/6tail/lunar-go/SolarUtil"
	"github.com/6tail/lunar-go/calendar"
	"github.com/sgostarter/libeasygo/cuserror"
)

func MinuteStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, t.Location())
}

func MinuteEnd(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 59, 0, t.Location())
}

func MinuteAdd(t time.Time, minutes int) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute()+minutes, t.Second(), t.Nanosecond(), t.Location())
}

func HourStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
}

func HourEnd(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 59, 59, 0, t.Location())
}

func HourAdd(t time.Time, hours int) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour()+hours, t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

func DayStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func DayEnd(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
}

func DayAdd(t time.Time, days int) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day()+days, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

func WeekStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day()-int(t.Weekday()), 0, 0, 0, 0, t.Location())
}

func WeekEnd(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day()-int(t.Weekday())+6, 23, 59, 59, 0, t.Location())
}

func WeekAdd(t time.Time, weeks int) time.Time {
	sw := calendar.NewSolarWeekFromYmd(t.Year(), int(t.Month()), t.Day(), 1)

	sw = sw.Next(weeks, false)

	return time.Date(sw.GetYear(), time.Month(sw.GetMonth()), sw.GetDay(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

func MonthStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func MonthEnd(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), SolarUtil.GetDaysOfMonth(t.Year(), int(t.Month())), 23, 59, 59, 0, t.Location())
}

func MonthAdd(t time.Time, months int) time.Time {
	return time.Date(t.Year(), t.Month()+time.Month(months), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

func LunarMonthStart(t time.Time) time.Time {
	tz8 := time.FixedZone("z8", 8*3600)

	t8 := t
	t8 = t8.In(tz8)

	st := calendar.NewSolarFromYmd(t8.Year(), int(t8.Month()), t8.Day())
	lst := st.GetLunar()
	lst = calendar.NewLunarFromYmd(lst.GetYear(), lst.GetMonth(), 1)

	st = lst.GetSolar()

	return time.Date(st.GetYear(), time.Month(st.GetMonth()), st.GetDay(), 0, 0, 0, 0, tz8).In(t.Location())
}

func LunarMonthEnd(t time.Time) time.Time {
	tz8 := time.FixedZone("z8", 8*3600)

	t8 := t
	t8 = t8.In(tz8)

	st := calendar.NewSolarFromYmd(t8.Year(), int(t8.Month()), t8.Day())
	lst := st.GetLunar()
	lst = calendar.NewLunarFromYmd(lst.GetYear(), lst.GetMonth(), calendar.NewLunarMonthFromYm(lst.GetYear(), lst.GetMonth()).GetDayCount())

	st = lst.GetSolar()

	return time.Date(st.GetYear(), time.Month(st.GetMonth()), st.GetDay(), 23, 59, 59, 0, tz8).In(t.Location())
}

func LunarMonthAdd(t time.Time, months int) time.Time {
	tz8 := time.FixedZone("z8", 8*3600)

	t8 := t
	t8 = t8.In(tz8)

	st := calendar.NewSolarFromYmd(t8.Year(), int(t8.Month()), t8.Day())
	lst := st.GetLunar()

	lm := calendar.NewLunarMonthFromYm(lst.GetYear(), lst.GetMonth()).Next(months)

	dayCount := calendar.NewLunarMonthFromYm(lm.GetYear(), lm.GetMonth()).GetDayCount()
	if lst.GetDay() < dayCount {
		dayCount = lst.GetDay()
	}

	lst = calendar.NewLunarFromYmd(lm.GetYear(), lm.GetMonth(), dayCount)

	st = lst.GetSolar()

	return time.Date(st.GetYear(), time.Month(st.GetMonth()), st.GetDay(), t8.Hour(), t8.Minute(), t8.Second(), 0, tz8).In(t.Location())
}

func LunarMonthEndAdd(t time.Time, months int) time.Time {
	tz8 := time.FixedZone("z8", 8*3600)

	t8 := t
	t8 = t8.In(tz8)

	st := calendar.NewSolarFromYmd(t8.Year(), int(t8.Month()), t8.Day())
	lst := st.GetLunar()

	lm := calendar.NewLunarMonthFromYm(lst.GetYear(), lst.GetMonth()).Next(months)

	lst = calendar.NewLunarFromYmd(lm.GetYear(), lm.GetMonth(), calendar.NewLunarMonthFromYm(lm.GetYear(), lm.GetMonth()).GetDayCount())

	st = lst.GetSolar()

	return time.Date(st.GetYear(), time.Month(st.GetMonth()), st.GetDay(), 23, 59, 59, 0, tz8).In(t.Location())
}

func YearStart(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

func YearEnd(t time.Time) time.Time {
	return time.Date(t.Year(), 12, SolarUtil.GetDaysOfMonth(t.Year(), 12), 23, 59, 59, 0, t.Location())
}

func YearAdd(t time.Time, years int) time.Time {
	return time.Date(t.Year()+years, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

func LunarYearStart(t time.Time) time.Time {
	tz8 := time.FixedZone("z8", 8*3600)

	t8 := t
	t8 = t8.In(tz8)

	st := calendar.NewSolarFromYmd(t8.Year(), int(t8.Month()), t8.Day())
	lst := st.GetLunar()
	lst = calendar.NewLunarFromYmd(lst.GetYear(), 1, 1)

	st = lst.GetSolar()

	return time.Date(st.GetYear(), time.Month(st.GetMonth()), st.GetDay(), 0, 0, 0, 0, tz8).In(t.Location())
}

func LunarYearEnd(t time.Time) time.Time {
	tz8 := time.FixedZone("z8", 8*3600)

	t8 := t
	t8 = t8.In(tz8)

	st := calendar.NewSolarFromYmd(t8.Year(), int(t8.Month()), t8.Day())
	lst := st.GetLunar()
	lst = calendar.NewLunarFromYmd(lst.GetYear(), 12, calendar.NewLunarMonthFromYm(lst.GetYear(), 12).GetDayCount())

	st = lst.GetSolar()

	return time.Date(st.GetYear(), time.Month(st.GetMonth()), st.GetDay(), 23, 59, 59, 0, tz8).In(t.Location())
}

func LunarYearBeginAdd(t time.Time, years int) time.Time {
	if years == 0 {
		return t
	}

	tz8 := time.FixedZone("z8", 8*3600)

	t8 := t
	t8 = t8.In(tz8)

	st := calendar.NewSolarFromYmd(t8.Year(), int(t8.Month()), t8.Day())
	lst := st.GetLunar()

	ly := calendar.NewLunarYear(lst.GetYear()).Next(years)

	lst = calendar.NewLunarFromYmd(ly.GetYear(), 1, 1)

	st = lst.GetSolar()

	return time.Date(st.GetYear(), time.Month(1), 1, 0, 0, 0, 0, tz8).In(t.Location())
}

func LunarYearEndAdd(t time.Time, years int) time.Time {
	tz8 := time.FixedZone("z8", 8*3600)

	t8 := t
	t8 = t8.In(tz8)

	st := calendar.NewSolarFromYmd(t8.Year(), int(t8.Month()), t8.Day())
	lst := st.GetLunar()

	ly := calendar.NewLunarYear(lst.GetYear()).Next(years)

	lst = calendar.NewLunarFromYmd(ly.GetYear(), 12, calendar.NewLunarMonthFromYm(ly.GetYear(), 12).GetDayCount())

	st = lst.GetSolar()

	return time.Date(st.GetYear(), time.Month(st.GetMonth()), st.GetDay(), 23, 59, 59, 0, tz8).In(t.Location())
}

func IsWorkDay(t time.Time) bool {
	switch t.Weekday() {
	case time.Sunday, time.Saturday:
		return false
	}

	return true
}

func WeekIndexInMonth(t time.Time) int {
	return calendar.NewSolarWeekFromDate(t, 1).GetIndex()
}

func ToDateTime(year, month, day, hour, minute, second int, location *time.Location) time.Time {
	return time.Date(year, time.Month(month), day, hour, minute, second, 0, location)
}

func LunarToDateTime(year, month, day, hour, minute, second int) time.Time {
	solar := calendar.NewLunar(year, month, day, hour, minute, second).GetSolar()

	return time.Date(solar.GetYear(), time.Month(solar.GetMonth()), solar.GetDay(), solar.GetHour(), solar.GetMinute(),
		solar.GetSecond(), 0, time.FixedZone("z8", 8*3600))
}

func lunarMonthToDateTime(m *calendar.LunarMonth, lunarDay, hour, minute, second int) (t time.Time, err error) {
	if m == nil {
		err = cuserror.NewWithErrorMsg("no m")

		return
	}

	days := m.GetDayCount()

	if lunarDay < 0 {
		switch lunarDay {
		case -1:
			lunarDay = days
		case -2:
			lunarDay = days - 1
		case -3:
			lunarDay = days - 2
		default:
			err = cuserror.NewWithErrorMsg(fmt.Sprintf("wrong lunar day %v", lunarDay))

			return
		}
	}

	if lunarDay < 1 {
		err = cuserror.NewWithErrorMsg("lunar day must bigger than 0")

		return
	}

	if lunarDay > days {
		lunarDay = days
	}

	t = LunarToDateTime(m.GetYear(), m.GetMonth(), lunarDay, hour, minute, second)

	return
}

func LunarToDateTimeAndNextYear(lunarYear, lunarMonth, lunarDay, hour, minute, second int, years int) (t time.Time, err error) {
	y := calendar.NewLunarYear(lunarYear)
	y = y.Next(years)

	m := y.GetMonth(lunarMonth)

	return lunarMonthToDateTime(m, lunarDay, hour, minute, second)
}

func LunarToDateTimeAndNextMonth(lunarYear, lunarMonth, lunarDay, hour, minute, second int, months int) (t time.Time, err error) {
	m := calendar.NewLunarMonthFromYm(lunarYear, lunarMonth)
	m = m.Next(months)

	return lunarMonthToDateTime(m, lunarDay, hour, minute, second)
}

func LunarToDateTimeAndNextDay(year, month, day, hour, minute, second int, days int) time.Time {
	lunar := calendar.NewLunar(year, month, day, hour, minute, second)

	lunar = lunar.Next(days)

	solar := lunar.GetSolar()

	return time.Date(solar.GetYear(), time.Month(solar.GetMonth()), solar.GetDay(), solar.GetHour(), solar.GetMinute(),
		solar.GetSecond(), 0, time.FixedZone("z8", 8*3600))
}

func GetDaysOfMonth(year int, month int) int {
	return SolarUtil.GetDaysOfMonth(year, month)
}

func LunarYMD(t time.Time) (year, month, day int) {
	l := calendar.NewSolarFromDate(t.In(time.Local)).GetLunar()

	year = l.GetYear()
	month = l.GetMonth()
	day = l.GetDay()

	return
}

func LunarGetDaysOfMonth(year int, month int) int {
	return calendar.NewLunarYear(year).GetMonth(month).GetDayCount()
}
