package timeassist

import (
	"os"
	"time"
)

type Task struct {
	ID        string   `yaml:"ID" json:"id,omitempty"`
	TType     TimeType `yaml:"TType,omitempty" json:"t_type,omitempty"`
	LunarFlag bool     `yaml:"lunar_flag" json:"lunar_flag"`

	Text string `yaml:"Text" json:"text,omitempty"`

	//
	// recycle task
	//

	Value int  `yaml:"Value,omitempty" json:"value,omitempty"`
	Auto  bool `yaml:"Auto,omitempty" json:"auto,omitempty"`

	TimeZone  int        `yaml:"TimeZone,omitempty" json:"time_zone,omitempty"`
	ValidTime *ValidTime `yaml:"ValidRanges,omitempty" json:"valid_time,omitempty"`
}

func (ct *Task) Valid() (err error) {
	err = os.ErrInvalid

	if ct.ID == "" || ct.TType <= TimeTypeBegin || ct.TType >= TimeTypeEnd || ct.Text == "" {
		return
	}

	if ct.TType != TimeTypeOnce {
		if ct.Value <= 0 {
			return
		}
	}

	err = nil

	return
}

func (ct *Task) GenRecycleData() (rd *ShowItem, nowIsValid bool) {
	return ct.GenRecycleDataEx(time.Now())
}
func (ct *Task) GenRecycleDataEx(timeNow time.Time) (rd *ShowItem, nowIsValid bool) {
	rd, nowIsValid = ct.genRecycleDataEx(timeNow)
	if rd.EndUTC > time.Now().Unix() {
		return
	}

	for {
		timeNow = time.Unix(rd.EndUTC, 0)

		rd, _ = ct.genRecycleDataEx(timeNow)
		if rd.EndUTC > time.Now().Unix() {
			break
		}
	}

	return
}

func (ct *Task) genRecycleDataEx(timeNow time.Time) (rd *ShowItem, nowIsValid bool) {
	rd = &ShowItem{
		ID: ct.ID,
	}

	if ct.TimeZone < -11 || ct.TimeZone > 12 {
		ct.TimeZone = 8
	}

	nowIsValid = true

	timeNow = timeNow.In(time.FixedZone("X", ct.TimeZone*3600))

	fnFillRD := func(startProc func(t time.Time) time.Time, addProc func(t time.Time, minutes int) time.Time) {
		t := timeNow

		if ct.ValidTime != nil {
			var curIsValid bool

			curIsValid, t = ct.ValidTime.FindAfterTime(t)
			if !curIsValid && nowIsValid {
				nowIsValid = false
			}
		}

		st := startProc(t)

		et := addProc(st, ct.Value)
		if ct.ValidTime != nil {
			_, et = ct.ValidTime.FindAfterTime(et)
		}

		rd.StartUTC = st.Unix()
		rd.EndUTC = et.Unix()
	}

	switch ct.TType {
	case RecycleTimeTypeMinute:
		fnFillRD(MinuteStart, MinuteAdd)
	case RecycleTimeTypeHour:
		fnFillRD(HourStart, HourAdd)
	case RecycleTimeTypeDay:
		fnFillRD(DayStart, DayAdd)
	case RecycleTimeTypeWeek:
		fnFillRD(WeekStart, WeekAdd)
	case RecycleTimeTypeMonth:
		if ct.LunarFlag {
			fnFillRD(LunarMonthStart, LunarMonthAdd)
		} else {
			fnFillRD(MonthStart, MonthAdd)
		}
	case RecycleTimeTypeYear:
		if ct.LunarFlag {
			fnFillRD(LunarYearStart, LunarYearBeginAdd)
		} else {
			fnFillRD(YearStart, YearAdd)
		}
	default:
		panic("ct.TType")
	}

	return
}
