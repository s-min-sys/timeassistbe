package timeassist

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sgostarter/i/commerr"
)

type AlarmType int

const (
	AlarmTypeBegin AlarmType = iota
	AlarmTypeOnce
	RecycleAlarmTypeYear
	RecycleAlarmTypeMonth
	RecycleAlarmTypeWeek
	RecycleAlarmTypeDay
	RecycleAlarmTypeHour
	RecycleAlarmTypeMinute
	AlarmTypeEnd
)

/* AlarmValue
L[S]20230222092218 [阴阳年月日时分秒] AlarmTypeOnce/AlarmTypeLunarOnce | ? 分
L[S]0222092420 [月日时分秒] RecycleAlarmTypeYear/RecycleAlarmTypeLunarYear | 3 * 24 * 60 分
L[S]22092400 [日时分秒] RecycleAlarmTypeMonth/RecycleAlarmTypeLunarMonth | 24 * 60 分
3092400[周时分秒] RecycleAlarmTypeWeek | 24 * 60 分
092812[时分秒] RecycleAlarmTypeDay | 60 分
2912[分秒] RecycleAlarmTypeHour | 5分
23[秒] RecycleAlarmTypeMinute | 0 分
*/

type Alarm struct {
	ID    string    `yaml:"ID" json:"id,omitempty"`
	AType AlarmType `yaml:"AType,omitempty" json:"a_type,omitempty"`

	Text string `yaml:"Text" json:"text,omitempty"`

	Value    string `yaml:"Value,omitempty" json:"value,omitempty"` // @see AlarmValue
	TimeZone int    `yaml:"TimeZone,omitempty" json:"timeZone,omitempty"`
}

func (a *Alarm) Validate() (av *AlarmValue, err error) {
	if a.ID == "" || a.Text == "" || a.Value == "" {
		return nil, commerr.ErrInvalidArgument
	}

	av, err = ParseAlarmValue(a.Value, a.AType)

	return
}

func (a *Alarm) GenRecycleData() (av *AlarmValue, timeAt time.Time, rd *TaskData, show, alarm bool, err error) {
	return a.GenRecycleDataEx(time.Now(), time.Now())
}

func (a *Alarm) GenRecycleDataEx(timeNow, timeLastAt time.Time) (av *AlarmValue, timeAt time.Time, rd *TaskData, show, alarm bool, err error) {
	av, err = a.Validate()
	if err != nil {
		return
	}

	rdNow := &TaskData{
		ID: a.ID,
	}

	if a.TimeZone < -11 || a.TimeZone > 12 {
		a.TimeZone = 8
	}

	timeZone := time.FixedZone("X", a.TimeZone*3600)
	timeNow = timeNow.In(timeZone)

	var showDuration time.Duration

	fnCalcDynamicDuration := func(tNow, tAt time.Time) (d time.Duration) {
		if tAt.Before(tNow) {
			return
		}
		if tAt.Sub(tNow) >= 3*24*time.Hour {
			d = time.Hour * 24 * 3
		} else if tAt.Sub(tNow) >= 24*time.Hour {
			d = time.Hour * 24
		} else if tAt.Sub(tNow) >= time.Hour {
			d = time.Hour
		} else {
			d = time.Minute
		}

		return
	}

	fnFixDayOfMonth := func(year, month, day int) int {
		days := GetDaysOfMonth(year, month)
		if day > days || day == -1 {
			return days
		}

		return day
	}

	fnFixLunarDayOfMonth := func(year, month, day int) int {
		days := LunarGetDaysOfMonth(year, month)
		if day > days || day == -1 {
			return days
		}

		return day
	}

	switch a.AType {
	case AlarmTypeOnce:
		if av.Lunar {
			timeAt = ToLunarDateTime(av.Year, av.Month, fnFixLunarDayOfMonth(av.Year, av.Month, av.Day), av.Hour, av.Minute, av.Second)
		} else {
			timeAt = ToDateTime(av.Year, av.Month, fnFixDayOfMonth(av.Year, av.Month, av.Day), av.Hour, av.Minute, av.Second, timeZone)
		}
		showDuration = fnCalcDynamicDuration(timeNow, timeAt)
	case RecycleAlarmTypeYear:
		if av.Lunar {
			year, _, _ := LunarYMD(timeNow)
		ReCalcLunarYear:
			day := fnFixLunarDayOfMonth(year, av.Month, av.Day)

			timeAt = ToLunarDateTime(year, av.Month, day, av.Hour, av.Minute, av.Second)
			if timeAt.Before(timeNow) {
				year++
				goto ReCalcLunarYear
			}
		} else {
			year := timeNow.Year()

		ReCalcYear:
			day := fnFixDayOfMonth(year, av.Month, av.Day)

			timeAt = ToDateTime(year, av.Month, day, av.Hour, av.Minute, av.Second, timeZone)
			if timeAt.Before(timeNow) {
				year++
				goto ReCalcYear
			}

		}

		showDuration = time.Hour * 24 * 3
	case RecycleAlarmTypeMonth:
		if av.Lunar {
			year, month, _ := LunarYMD(timeNow)

		ReCalcLunarMonth:
			day := fnFixLunarDayOfMonth(year, month, av.Day)

			timeAt = ToLunarDateTime(year, month, day, av.Hour, av.Minute, av.Second)
			if timeAt.Before(timeNow) {
				month++
				goto ReCalcLunarMonth
			}
		} else {
			year := timeNow.Year()
			month := timeNow.Month()

		ReCalcMonth:
			day := fnFixDayOfMonth(year, int(month), av.Day)

			timeAt = ToDateTime(year, int(month), day, av.Hour, av.Minute, av.Second, timeZone)
			if timeAt.Before(timeNow) {
				month++
				goto ReCalcMonth
			}
		}

		showDuration = time.Hour * 24
	case RecycleAlarmTypeWeek:
		timeAt = timeNow
		for int(timeAt.Weekday()) != av.Week {
			timeAt = DayAdd(timeAt, 1)
		}

		timeAt = ToDateTime(timeAt.Year(), int(timeAt.Month()), timeAt.Day(), av.Hour, av.Minute, av.Second, timeAt.Location())
		showDuration = time.Hour * 2
	case RecycleAlarmTypeDay:
		year := timeNow.Year()
		month := timeNow.Month()
		day := timeNow.Day()

	ReCalcDay:
		timeAt = ToDateTime(year, int(month), day, av.Hour, av.Minute, av.Second, timeZone)
		if timeAt.Before(timeNow) {
			day++
			goto ReCalcDay
		}

		showDuration = time.Hour
	case RecycleAlarmTypeHour:
		year := timeNow.Year()
		month := timeNow.Month()
		day := timeNow.Day()
		hour := timeNow.Hour()

	ReCalcHour:
		timeAt = ToDateTime(year, int(month), day, hour, av.Minute, av.Second, timeZone)
		if timeAt.Before(timeNow) {
			hour++
			goto ReCalcHour
		}

		showDuration = time.Minute * 5
	case RecycleAlarmTypeMinute:
		year := timeNow.Year()
		month := timeNow.Month()
		day := timeNow.Day()
		hour := timeNow.Hour()
		minute := timeNow.Minute()

	ReCalcMinute:
		timeAt = ToDateTime(year, int(month), day, hour, minute, av.Second, timeZone)
		if timeAt.Before(timeNow) {
			minute++
			goto ReCalcMinute
		}

		showDuration = time.Minute * 5
	default:
		return
	}

	timeShow := timeAt.Add(-showDuration)

	rdNow.StartUTC = timeShow.Unix()
	rdNow.EndUTC = timeAt.Unix()

	if a.AType == AlarmTypeOnce {
		if timeAt.Before(timeNow) {
			show = true
			alarm = true

			return
		}
	}

	rd = rdNow

	if timeLastAt.Before(timeNow) {
		show = true
		alarm = true

		return
	}

	if timeShow.Before(timeNow) {
		show = true
	}

	return
}

type AlarmValue struct {
	Lunar  bool
	Year   int
	Month  int
	Day    int
	Week   int
	Hour   int
	Minute int
	Second int
}

func (av *AlarmValue) String(aType AlarmType, timeAt time.Time) string {
	if aType == AlarmTypeOnce {
		if av.Lunar {
			var day string
			if av.Day == -1 {
				day = fmt.Sprintf("%02d日(最后一天)", GetDaysOfMonth(av.Year, av.Month))
			} else {
				day = fmt.Sprintf("%02d日", av.Day)
			}
			return fmt.Sprintf("阴历%04d年%02d月%s%02d时%02d分%02d秒", av.Year, av.Month, day, av.Hour, av.Minute, av.Second)
		}

		var day string
		if av.Day == -1 {
			day = fmt.Sprintf("%02d日(最后一天)", LunarGetDaysOfMonth(av.Year, av.Month))
		} else {
			day = fmt.Sprintf("%02d日", av.Day)
		}
		return fmt.Sprintf("%04d年%02d月%s%02d时%02d分%02d秒", av.Year, av.Month, day, av.Hour, av.Minute, av.Second)

	}

	var day string
	if av.Day == -1 {
		if av.Lunar {
			day = fmt.Sprintf("%02d日(最后一天)", LunarGetDaysOfMonth(av.Year, av.Month))
		} else {
			day = fmt.Sprintf("%02d日(最后一天)", GetDaysOfMonth(av.Year, av.Month))
		}
	} else {
		day = fmt.Sprintf("%02d日", av.Day)
	}

	yx := "阳历"
	if av.Lunar {
		yx = "阴历"
	}

	var pre string

	switch aType {
	case RecycleAlarmTypeYear:
		pre = yx + fmt.Sprintf("每年%02d月%s%02d时%02d分%02d秒", av.Month, day, av.Hour, av.Minute, av.Second)
	case RecycleAlarmTypeMonth:
		pre = yx + fmt.Sprintf("每月%s%02d时%02d分%02d秒", day, av.Hour, av.Minute, av.Second)
	case RecycleAlarmTypeWeek:
		week := fmt.Sprintf("周%d", av.Week)
		if av.Week == 0 {
			week = "周日"
		}
		pre = fmt.Sprintf("每周周%s%02d时%02d分%02d秒", week, av.Hour, av.Minute, av.Second)
	case RecycleAlarmTypeDay:
		pre = fmt.Sprintf("每日%02d时%02d分%02d秒", av.Hour, av.Minute, av.Second)
	case RecycleAlarmTypeHour:
		pre = fmt.Sprintf("每小时%02d分%02d秒", av.Minute, av.Second)
	case RecycleAlarmTypeMinute:
		pre = fmt.Sprintf("每分%02d秒", av.Second)
	default:
		return ""
	}

	return pre + timeAt.Format("[2006年01月02日15时04分05秒]")
}

func (av *AlarmValue) Valid(aType AlarmType) bool {
	switch aType {
	case AlarmTypeOnce:
		if av.Year <= 0 {
			return false
		}
		fallthrough
	case RecycleAlarmTypeYear:
		if av.Month < 1 || av.Month > 12 {
			return false
		}
		fallthrough
	case RecycleAlarmTypeWeek, RecycleAlarmTypeMonth:
		if aType == RecycleAlarmTypeWeek {
			if av.Week < 0 || av.Week > 5 {
				return false
			}
		} else {
			if av.Day < 1 || av.Day > 31 || av.Day == -1 /*last day of month*/ {
				return false
			}
		}
		fallthrough
	case RecycleAlarmTypeDay:
		if av.Hour < 0 || av.Hour > 23 {
			return false
		}
		fallthrough
	case RecycleAlarmTypeHour:
		if av.Minute < 0 || av.Minute > 59 {
			return false
		}
		fallthrough
	case RecycleAlarmTypeMinute:
		if av.Second < 0 || av.Second > 59 {
			return false
		}

		return true
	}

	return false
}

func ParseAlarmValue(value string, aType AlarmType) (av *AlarmValue, err error) {
	switch aType {
	case AlarmTypeOnce:
		av, err = parseAlarmValueOnce(value)
	case RecycleAlarmTypeYear:
		av, err = parseAlarmValueYear(value)
	case RecycleAlarmTypeMonth:
		av, err = parseAlarmValueMonth(value)
	case RecycleAlarmTypeWeek:
		av, err = parseAlarmValueWeek(value)
	case RecycleAlarmTypeDay:
		av, err = parseAlarmValueDay(value)
	case RecycleAlarmTypeHour:
		av, err = parseAlarmValueHour(value)
	case RecycleAlarmTypeMinute:
		av, err = parseAlarmValueMinute(value)
	default:
		err = commerr.ErrUnimplemented
	}

	return
}

func parseAlarmValueOnce(value string) (av *AlarmValue, err error) {
	// L[S]20230222092218
	if len(value) < 1 {
		err = commerr.ErrBadFormat

		return
	}

	lunar := strings.ToUpper(value[0:1])
	if lunar == "L" || lunar == "S" {
		value = value[1:]
	} else {
		lunar = "S"
	}

	if len(value) < 4 {
		err = commerr.ErrBadFormat

		return
	}

	year, err := strconv.Atoi(value[0:4])
	if err != nil {
		return
	}

	av, err = parseAlarmValueYear(value[4:])
	if err != nil {
		return
	}

	av.Lunar = lunar == "L"
	av.Year = year

	if !av.Valid(AlarmTypeOnce) {
		err = commerr.ErrBadFormat
	}

	return
}

func parseAlarmValueYear(value string) (av *AlarmValue, err error) {
	// L[S]0222092218
	if len(value) < 1 {
		err = commerr.ErrBadFormat

		return
	}

	lunar := strings.ToUpper(value[0:1])
	if lunar == "L" || lunar == "S" {
		value = value[1:]
	} else {
		lunar = "S"
	}

	if len(value) < 2 {
		err = commerr.ErrBadFormat

		return
	}

	month, err := strconv.Atoi(value[0:2])
	if err != nil {
		return
	}

	av, err = parseAlarmValueMonth(value[2:])
	if err != nil {
		return
	}

	av.Lunar = lunar == "L"
	av.Month = month

	if !av.Valid(RecycleAlarmTypeYear) {
		err = commerr.ErrBadFormat
	}

	return
}

func parseAlarmValueMonth(value string) (av *AlarmValue, err error) {
	// L[S]22092400
	if len(value) < 1 {
		err = commerr.ErrBadFormat

		return
	}

	lunar := strings.ToUpper(value[0:1])
	if lunar == "L" || lunar == "S" {
		value = value[1:]
	} else {
		lunar = "S"
	}

	if len(value) < 2 {
		err = commerr.ErrBadFormat

		return
	}

	day, err := strconv.Atoi(value[0:2])
	if err != nil {
		return
	}

	av, err = parseAlarmValueDay(value[2:])
	if err != nil {
		return
	}

	av.Lunar = lunar == "L"
	av.Day = day

	if !av.Valid(RecycleAlarmTypeMonth) {
		err = commerr.ErrBadFormat
	}

	return
}

func parseAlarmValueWeek(value string) (av *AlarmValue, err error) {
	if len(value) < 1 {
		err = commerr.ErrBadFormat

		return
	}

	// 3092400
	week, err := strconv.Atoi(value[0:1])
	if err != nil {
		return
	}

	av, err = parseAlarmValueDay(value[1:])
	if err != nil {
		return
	}

	av.Week = week

	if !av.Valid(RecycleAlarmTypeWeek) {
		err = commerr.ErrBadFormat
	}

	return
}

func parseAlarmValueDay(value string) (av *AlarmValue, err error) {
	if len(value) < 2 {
		err = commerr.ErrBadFormat

		return
	}

	// 092812
	hour, err := strconv.Atoi(value[0:2])
	if err != nil {
		return
	}

	av, err = parseAlarmValueHour(value[2:])
	if err != nil {
		return
	}

	av.Hour = hour

	if !av.Valid(RecycleAlarmTypeDay) {
		err = commerr.ErrBadFormat
	}

	return
}

func parseAlarmValueHour(value string) (av *AlarmValue, err error) {
	if len(value) < 2 {
		err = commerr.ErrBadFormat

		return
	}

	// 2912
	minute, err := strconv.Atoi(value[0:2])
	if err != nil {
		return
	}

	av, err = parseAlarmValueMinute(value[2:])
	if err != nil {
		return
	}

	av.Minute = minute

	if !av.Valid(RecycleAlarmTypeHour) {
		err = commerr.ErrBadFormat
	}

	return
}

func parseAlarmValueMinute(value string) (av *AlarmValue, err error) {
	if len(value) < 2 {
		err = commerr.ErrBadFormat

		return
	}

	// 23
	second, err := strconv.Atoi(value)
	if err != nil {
		return
	}

	av = &AlarmValue{
		Second: second,
	}

	if !av.Valid(RecycleAlarmTypeMinute) {
		err = commerr.ErrBadFormat
	}

	return
}
