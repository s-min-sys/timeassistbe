package timeassist

import (
	"os"
	"time"
)

type TaskType int

const (
	TaskTypeBegin TaskType = iota
	RecycleTaskTypeMinutes
	RecycleTaskTypeHours
	RecycleTaskTypeDays
	RecycleTaskTypeWeeks
	RecycleTaskTypeMonths
	RecycleTaskTypeLunarMonths
	RecycleTaskTypeYears
	RecycleTaskTypeLunarYears
	OnceTask
	TaskTypeEnd
)

func (o TaskType) String() string {
	switch o {
	case RecycleTaskTypeMinutes:
		return "RecycleTaskTypeMinutes"
	case RecycleTaskTypeHours:
		return "RecycleTaskTypeHours"
	case RecycleTaskTypeDays:
		return "RecycleTaskTypeDays"
	case RecycleTaskTypeWeeks:
		return "RecycleTaskTypeWeeks"
	case RecycleTaskTypeMonths:
		return "RecycleTaskTypeMonths"
	case RecycleTaskTypeLunarMonths:
		return "RecycleTaskTypeLunarMonths"
	case RecycleTaskTypeYears:
		return "RecycleTaskTypeYears"
	case RecycleTaskTypeLunarYears:
		return "RecycleTaskTypeLunarYears"
	case OnceTask:
		return "OnceTask"
	}

	return "RecycleTaskTypUnknown"
}

type Task struct {
	ID    string   `yaml:"ID" json:"id,omitempty"`
	TType TaskType `yaml:"TType,omitempty" json:"t_type,omitempty"`

	Text string `yaml:"Text" json:"text,omitempty"`

	//
	// recycle task
	//

	Value int  `yaml:"Value,omitempty" json:"value,omitempty"`
	Auto  bool `yaml:"Auto,omitempty" json:"auto,omitempty"`

	TimeZone  int        `yaml:"TimeZone,omitempty" json:"timeZone,omitempty"`
	ValidTime *ValidTime `yaml:"ValidRanges,omitempty" json:"valid_time,omitempty"`
}

type TaskData struct {
	ID       string   `yaml:"ID"`
	TType    TaskType `yaml:"TType" json:"t_type,omitempty"`
	StartUTC int64    `yaml:"StartUTC"`
	EndUTC   int64    `yaml:"EndUTC"`
}

func (ct *Task) Valid() (err error) {
	err = os.ErrInvalid

	if ct.ID == "" || ct.TType <= TaskTypeBegin || ct.TType >= TaskTypeEnd || ct.Text == "" {
		return
	}

	if ct.TType != OnceTask {
		if ct.Value <= 0 {
			return
		}
	}

	err = nil

	return
}

func (ct *Task) GenRecycleData() (rd *TaskData, nowIsValid bool) {
	return ct.GenRecycleDataEx(time.Now())
}
func (ct *Task) GenRecycleDataEx(timeNow time.Time) (rd *TaskData, nowIsValid bool) {
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

func (ct *Task) genRecycleDataEx(timeNow time.Time) (rd *TaskData, nowIsValid bool) {
	rd = &TaskData{
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
	case RecycleTaskTypeMinutes:
		fnFillRD(MinuteStart, MinuteAdd)
	case RecycleTaskTypeHours:
		fnFillRD(HourStart, HourAdd)
	case RecycleTaskTypeDays:
		fnFillRD(DayStart, DayAdd)
	case RecycleTaskTypeWeeks:
		fnFillRD(WeekStart, WeekAdd)
	case RecycleTaskTypeMonths:
		fnFillRD(MonthStart, MonthAdd)
	case RecycleTaskTypeLunarMonths:
		fnFillRD(LunarMonthStart, LunarMonthAdd)
	case RecycleTaskTypeYears:
		fnFillRD(YearStart, YearAdd)
	case RecycleTaskTypeLunarYears:
		fnFillRD(LunarYearStart, LunarYearBeginAdd)
	default:
		panic("ct.TType")
	}

	return
}
