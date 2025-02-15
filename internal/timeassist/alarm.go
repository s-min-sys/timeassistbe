package timeassist

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sgostarter/i/commerr"
)

/* AlarmValue
L[S]20230222092218 [阴阳年月日时分秒] TimeTypeOnce/AlarmTypeLunarOnce | ? 分
L[S]0222092420 [月日时分秒] RecycleTimeTypeYear/RecycleAlarmTypeLunarYear | 3 * 24 * 60 分
L[S]22092400 [日时分秒] RecycleTimeTypeMonth/RecycleAlarmTypeLunarMonth | 24 * 60 分
3092400[周时分秒] RecycleTimeTypeWeek | 24 * 60 分
092812[时分秒] RecycleTimeTypeDay | 60 分
2912[分秒] RecycleTimeTypeHour | 5分
23[秒] RecycleTimeTypeMinute | 0 分
*/

type Alarm struct {
	ID    string   `yaml:"ID" json:"id,omitempty"`
	AType TimeType `yaml:"AType,omitempty" json:"a_type,omitempty"`

	Text string `yaml:"Text" json:"text,omitempty"`

	Value    string `yaml:"Value,omitempty" json:"value,omitempty"` // @see AlarmValue
	TimeZone int    `yaml:"TimeZone,omitempty" json:"timeZone,omitempty"`

	ValidTime *ValidTime `yaml:"ValidRanges,omitempty" json:"valid_time,omitempty"`

	EarlyShowMinute int `yaml:"EarlyShowMinute,omitempty" json:"early_show_minute,omitempty"`

	//
	//
	//
	TimeLastAt int64 `yaml:"TimeLastAt,omitempty" json:"time_last_at,omitempty"` // ？
}

func (a *Alarm) Validate() (av *AlarmValue, err error) {
	if a.Text == "" || a.Value == "" {
		return nil, commerr.ErrInvalidArgument
	}

	av, err = ParseAlarmValue(a.Value, a.AType)

	return
}

func (a *Alarm) GenRecycleData() (av *AlarmValue, timeAt time.Time, rd *ShowItem, show, alarm bool, err error) {
	return a.GenRecycleDataEx(time.Now(), time.Now())
}

// GenRecycleDataEx
// nolint: gocyclo
func (a *Alarm) GenRecycleDataEx(timeNow, timeLastAt time.Time) (av *AlarmValue, timeAt time.Time, rd *ShowItem, show, alarm bool, err error) {
	av, err = a.Validate()
	if err != nil {
		return
	}

	rdNow := &ShowItem{
		ID: a.ID,
	}

	if a.TimeZone < -11 || a.TimeZone > 12 {
		a.TimeZone = 8
	}

	timeZone := time.FixedZone("X", a.TimeZone*3600)
	timeNow = timeNow.In(timeZone)
	timeLastAt = timeLastAt.In(timeZone)

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
	case TimeTypeOnce:
		if av.Lunar {
			timeAt = LunarToDateTime(av.Year, av.Month, fnFixLunarDayOfMonth(av.Year, av.Month, av.Day), av.Hour, av.Minute, av.Second)
		} else {
			timeAt = ToDateTime(av.Year, av.Month, fnFixDayOfMonth(av.Year, av.Month, av.Day), av.Hour, av.Minute, av.Second, timeZone)
		}

		showDuration = fnCalcDynamicDuration(timeNow, timeAt)
	case RecycleTimeTypeYear:
		if av.Lunar {
			year, _, _ := LunarYMD(timeNow)
			addYear := 0

		ReCalcLunarYear:
			day := fnFixLunarDayOfMonth(year, av.Month, av.Day)

			timeAt, err = LunarToDateTimeAndNextYear(year, av.Month, day, av.Hour, av.Minute, av.Second, addYear)
			if err != nil {
				return
			}

			if timeAt.Before(timeNow) {
				addYear++

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

		showDuration = time.Hour * 24 * 7
	case RecycleTimeTypeMonth:
		if av.Lunar {
			year, month, _ := LunarYMD(timeNow)

			addMonth := 0
		ReCalcLunarMonth:
			day := fnFixLunarDayOfMonth(year, month, av.Day)

			timeAt, err = LunarToDateTimeAndNextMonth(year, month, day, av.Hour, av.Minute, av.Second, addMonth)
			if err != nil {
				return
			}

			if timeAt.Before(timeNow) {
				addMonth++

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

		showDuration = time.Hour * 24 * 2
	case RecycleTimeTypeWeek:
		timeAt = ToDateTime(timeNow.Year(), int(timeNow.Month()), timeNow.Day(), av.Hour, av.Minute, av.Second, timeNow.Location())

		for int(timeAt.Weekday()) != av.Week || timeAt.Before(timeNow) {
			timeAt = DayAdd(timeAt, 1)
		}

		showDuration = time.Hour * 24
	case RecycleTimeTypeDay:
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
	case RecycleTimeTypeHour:
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
	case RecycleTimeTypeMinute:
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

		showDuration = time.Second * 5
	default:
		return
	}

	if a.ValidTime != nil {
		_, timeAt = a.ValidTime.FindAfterTime(timeAt)
	}

	if a.EarlyShowMinute > 0 {
		showDuration = time.Minute * time.Duration(a.EarlyShowMinute)
	}

	timeShow := timeAt.Add(-showDuration)

	rdNow.StartUTC = timeShow.Unix() // next show at
	rdNow.EndUTC = timeAt.Unix()     // next expire at

	if a.AType == TimeTypeOnce {
		if timeAt.Before(timeNow) {
			show = true
			alarm = true

			return
		}
	}

	rd = rdNow

	if timeShow.Before(timeNow) {
		show = true

		return
	}

	if timeLastAt.Before(timeNow) {
		alarm = true

		return
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

func (av *AlarmValue) StringNoNowTime(aType TimeType) (bool, string) {
	var days int

	var dayS string

	fnGetDay := func() string {
		if dayS != "" {
			return dayS
		}

		timeNow := time.Now()

		year := av.Year
		month := av.Month

		if year <= 0 {
			year = timeNow.Year()
		}

		if month <= 0 {
			month = int(timeNow.Month())
		}

		if av.Lunar {
			days = LunarGetDaysOfMonth(year, month)
		} else {
			days = GetDaysOfMonth(year, month)
		}

		switch av.Day {
		case -1:
			dayS = fmt.Sprintf("%02d日(最后一天)", days)
		case -2:
			dayS = fmt.Sprintf("%02d日(倒数第二天)", days)
		case -3:
			dayS = fmt.Sprintf("%02d日(倒数第三天)", days)
		default:
			dayS = fmt.Sprintf("%02d日", av.Day)
		}

		return dayS
	}

	if aType == TimeTypeOnce {
		if av.Lunar {
			return false, fmt.Sprintf("阴历%04d年%02d月%s%02d时%02d分%02d秒", av.Year, av.Month, fnGetDay(), av.Hour, av.Minute, av.Second)
		}

		return false, fmt.Sprintf("%04d年%02d月%s%02d时%02d分%02d秒", av.Year, av.Month, fnGetDay(), av.Hour, av.Minute, av.Second)
	}

	yx := "阳历"
	if av.Lunar {
		yx = "阴历"
	}

	var pre string

	switch aType {
	case RecycleTimeTypeYear:
		pre = yx + fmt.Sprintf("每年%02d月%s%02d时%02d分%02d秒", av.Month, fnGetDay(), av.Hour, av.Minute, av.Second)
	case RecycleTimeTypeMonth:
		pre = yx + fmt.Sprintf("每月%s%02d时%02d分%02d秒", fnGetDay(), av.Hour, av.Minute, av.Second)
	case RecycleTimeTypeWeek:
		week := fmt.Sprintf("周%d", av.Week)
		if av.Week == 0 {
			week = "周日"
		}

		pre = fmt.Sprintf("每周周%s%02d时%02d分%02d秒", week, av.Hour, av.Minute, av.Second)
	case RecycleTimeTypeDay:
		pre = fmt.Sprintf("每日%02d时%02d分%02d秒", av.Hour, av.Minute, av.Second)
	case RecycleTimeTypeHour:
		pre = fmt.Sprintf("每小时%02d分%02d秒", av.Minute, av.Second)
	case RecycleTimeTypeMinute:
		pre = fmt.Sprintf("每分%02d秒", av.Second)
	default:
		return true, ""
	}

	return true, pre
}

func (av *AlarmValue) String(aType TimeType, timeAt time.Time) string {
	shouldAppendTime, s := av.StringNoNowTime(aType)
	if shouldAppendTime {
		s += timeAt.Format("[2006年01月02日15时04分05秒]")
	}

	return s
}

func (av *AlarmValue) Valid(aType TimeType) bool {
	switch aType {
	case TimeTypeOnce:
		if av.Year <= 0 {
			return false
		}

		fallthrough
	case RecycleTimeTypeYear:
		if av.Month < 1 || av.Month > 12 {
			return false
		}

		fallthrough
	case RecycleTimeTypeWeek, RecycleTimeTypeMonth:
		if aType == RecycleTimeTypeWeek {
			if av.Week < 0 || av.Week > 6 {
				return false
			}
		} else {
			if av.Day != -1 && av.Day != -2 && av.Day != -3 {
				if av.Day < 1 || av.Day > 31 {
					return false
				}
			}
		}

		fallthrough
	case RecycleTimeTypeDay:
		if av.Hour < 0 || av.Hour > 23 {
			return false
		}

		fallthrough
	case RecycleTimeTypeHour:
		if av.Minute < 0 || av.Minute > 59 {
			return false
		}

		fallthrough
	case RecycleTimeTypeMinute:
		if av.Second < 0 || av.Second > 59 {
			return false
		}

		return true
	}

	return false
}

func ParseAlarmValue(value string, aType TimeType) (av *AlarmValue, err error) {
	switch aType {
	case TimeTypeOnce:
		av, err = parseAlarmValueOnce(value)
	case RecycleTimeTypeYear:
		av, err = parseAlarmValueYear(value)
	case RecycleTimeTypeMonth:
		av, err = parseAlarmValueMonth(value)
	case RecycleTimeTypeWeek:
		av, err = parseAlarmValueWeek(value)
	case RecycleTimeTypeDay:
		av, err = parseAlarmValueDay(value)
	case RecycleTimeTypeHour:
		av, err = parseAlarmValueHour(value)
	case RecycleTimeTypeMinute:
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

	if !av.Valid(TimeTypeOnce) {
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

	if !av.Valid(RecycleTimeTypeYear) {
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

	if !av.Valid(RecycleTimeTypeMonth) {
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

	if !av.Valid(RecycleTimeTypeWeek) {
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

	if !av.Valid(RecycleTimeTypeDay) {
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

	if !av.Valid(RecycleTimeTypeHour) {
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

	if !av.Valid(RecycleTimeTypeMinute) {
		err = commerr.ErrBadFormat
	}

	return
}
