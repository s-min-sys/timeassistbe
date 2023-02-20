package timeassist

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenRecycleTaskYaml(t *testing.T) {
	rt := &RecycleTask{
		TType:    RecycleTaskTypeMinutes,
		Value:    10,
		Auto:     true,
		TimeZone: 8,
		ValidTime: &ValidTime{
			ValidDaysInWeek: &ValidRanges{
				ValidRanges: []ValidRange{
					{
						Start: 1,
						End:   6,
					},
				},
			},
			ValidHoursInDay: &ValidRanges{
				ValidRanges: []ValidRange{
					{
						Start: 9,
						End:   13,
					},
					{
						Start: 14,
						End:   22,
					},
				},
			},
		},
		Text: "message",
	}

	rts := []*RecycleTask{
		rt,
	}

	d, err := yaml.Marshal(rts)
	assert.Nil(t, err)
	fmt.Print("\n" + string(d))
}

func utCheck1(t *testing.T) func(b, eB bool, startUTC, endUTC int64, sY, sMonth, sD, sH, sMinute, sS, eY, eMonth, eD, eH, eMinute, eS int) {
	return func(b, eB bool, startUTC, endUTC int64, sY, sMonth, sD, sH, sMinute, sS, eY, eMonth, eD, eH, eMinute, eS int) {
		assert.EqualValues(t, eB, b)
		startT := time.Unix(startUTC, 0)
		endT := time.Unix(endUTC, 0)

		assert.EqualValues(t, sY, startT.Year())
		assert.EqualValues(t, sMonth, startT.Month())
		assert.EqualValues(t, sD, startT.Day())
		assert.EqualValues(t, sH, startT.Hour())
		assert.EqualValues(t, sMinute, startT.Minute())
		assert.EqualValues(t, sS, startT.Second())

		assert.EqualValues(t, eY, endT.Year())
		assert.EqualValues(t, eMonth, endT.Month())
		assert.EqualValues(t, eD, endT.Day())
		assert.EqualValues(t, eH, endT.Hour())
		assert.EqualValues(t, eMinute, endT.Minute())
		assert.EqualValues(t, eS, endT.Second())
	}
}

func TestRecycleTask1(t *testing.T) {
	vt := &ValidTime{
		ValidMonthsInYear: nil,
		ValidWeeksInMonth: nil,
		ValidDaysInMonth:  nil,
		ValidDaysInWeek:   nil,
		ValidHoursInDay: &ValidRanges{
			ValidRanges: []ValidRange{
				{
					Start: 12,
					End:   14,
				},
			},
		},
	}

	ct := &RecycleTask{
		TType:     RecycleTaskTypeMinutes,
		Value:     10,
		Auto:      false,
		TimeZone:  8,
		ValidTime: vt,
	}

	timeNow := time.Date(2023, 2, 11, 10, 20, 10, 0, time.Local)

	rd, nowIsValid := ct.GenRecycleDataEx(timeNow)

	fnCheck1 := func(b, eB bool, startUTC, endUTC int64, sY, sMonth, sD, sH, sMinute, sS, eY, eMonth, eD, eH, eMinute, eS int) {
		assert.EqualValues(t, eB, b)
		startT := time.Unix(startUTC, 0)
		endT := time.Unix(endUTC, 0)

		assert.EqualValues(t, sY, startT.Year())
		assert.EqualValues(t, sMonth, startT.Month())
		assert.EqualValues(t, sD, startT.Day())
		assert.EqualValues(t, sH, startT.Hour())
		assert.EqualValues(t, sMinute, startT.Minute())
		assert.EqualValues(t, sS, startT.Second())

		assert.EqualValues(t, eY, endT.Year())
		assert.EqualValues(t, eMonth, endT.Month())
		assert.EqualValues(t, eD, endT.Day())
		assert.EqualValues(t, eH, endT.Hour())
		assert.EqualValues(t, eMinute, endT.Minute())
		assert.EqualValues(t, eS, endT.Second())
	}

	fnCheck1(nowIsValid, false, rd.StartUTC, rd.EndUTC,
		2023, 2, 11, 12, 0, 0,
		2023, 2, 11, 12, 10, 0)
	//
	//
	//

	vt.Reset()
	vt.ValidHoursInDay = &ValidRanges{
		ValidRanges: []ValidRange{
			{
				Start: 12,
				End:   14,
			},
		},
	}

	timeNow = time.Date(2023, 2, 11, 13, 20, 10, 0, time.Local)

	rd, nowIsValid = ct.GenRecycleDataEx(timeNow)

	fnCheck1(nowIsValid, true, rd.StartUTC, rd.EndUTC,
		2023, 2, 11, 13, 20, 0,
		2023, 2, 11, 13, 30, 0)

	//
	//
	//

	vt.Reset()
	vt.ValidDaysInWeek = &ValidRanges{
		ValidRanges: []ValidRange{
			{
				Start: 6,
				End:   7,
			},
		},
	}

	timeNow = time.Date(2023, 2, 11, 13, 20, 10, 0, time.Local)

	rd, nowIsValid = ct.GenRecycleDataEx(timeNow)

	fnCheck1(nowIsValid, true, rd.StartUTC, rd.EndUTC,
		2023, 2, 11, 13, 20, 0,
		2023, 2, 11, 13, 30, 0)

	//
	//
	//

	vt.Reset()
	vt.ValidDaysInWeek = &ValidRanges{
		ValidRanges: []ValidRange{
			{
				Start: 5,
				End:   6,
			},
		},
	}

	timeNow = time.Date(2023, 2, 11, 13, 20, 10, 0, time.Local)

	rd, nowIsValid = ct.GenRecycleDataEx(timeNow)

	fnCheck1(nowIsValid, false, rd.StartUTC, rd.EndUTC,
		2023, 2, 17, 0, 0, 0,
		2023, 2, 17, 0, 10, 0)
}

func TestRecycleTask2(t *testing.T) {
	vt := &ValidTime{
		ValidMonthsInYear: nil,
		ValidWeeksInMonth: nil,
		ValidDaysInMonth:  nil,
		ValidDaysInWeek: &ValidRanges{
			ValidRanges: []ValidRange{
				{
					Start: 1,
					End:   6,
				},
			},
		},
		ValidHoursInDay: &ValidRanges{
			ValidRanges: []ValidRange{
				{
					Start: 9,
					End:   12,
				},
				{
					Start: 14,
					End:   23,
				},
			},
		},
	}

	ct := &RecycleTask{
		TType:     RecycleTaskTypeHours,
		Value:     3,
		Auto:      false,
		TimeZone:  8,
		ValidTime: vt,
	}

	tz := time.FixedZone("UT", 8*3600)

	timeNow := time.Date(2023, 2, 10, 10, 20, 10, 0, tz)

	rd, nowIsValid := ct.GenRecycleDataEx(timeNow)

	fnCheck := utCheck1(t)

	ss := strings.Builder{}
	ss.WriteString("\n")

	for idx := 0; idx < 100; idx++ {
		tD := utTestRecycleTask2Data1[idx]
		fnCheck(nowIsValid, tD.b, rd.StartUTC, rd.EndUTC,
			tD.sY, tD.sMon, tD.sD, tD.sH, tD.sMin, tD.sS, tD.eY, tD.eMon, tD.eD, tD.eH, tD.eMin, tD.eS)

		// t.Log(nowIsValid, time.Unix(rd.StartUTC, 0).In(tz), time.Unix(rd.EndUTC, 0).In(tz))

		ss.WriteString("{\n")
		ss.WriteString(strconv.FormatBool(nowIsValid) + ",")
		t4p := time.Unix(rd.StartUTC, 0).In(tz)
		ss.WriteString(fmt.Sprintf("%d,%d,%d,%d,%d,%d,", t4p.Year(), t4p.Month(), t4p.Day(), t4p.Hour(), t4p.Minute(), t4p.Second()))
		t4p = time.Unix(rd.EndUTC, 0).In(tz)
		ss.WriteString(fmt.Sprintf("%d,%d,%d,%d,%d,%d,\n", t4p.Year(), t4p.Month(), t4p.Day(), t4p.Hour(), t4p.Minute(), t4p.Second()))
		ss.WriteString("},\n")

		rd, nowIsValid = ct.GenRecycleDataEx(time.Unix(rd.EndUTC, 0))
	}

	// t.Log(ss.String())
}
