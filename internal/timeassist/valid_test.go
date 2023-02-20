package timeassist

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValid1(t *testing.T) {
	vt := &ValidTime{
		ValidMonthsInYear: nil,
		ValidWeeksInMonth: nil,
		ValidDaysInMonth:  nil,
		ValidDaysInWeek:   nil,
		ValidHoursInDay: &ValidRanges{
			ValidRanges: []ValidRange{
				{
					Start: 19,
					End:   22,
				},
			},
		},
	}

	nowT := time.Date(2023, 2, 11, 19, 13, 0, 0, time.Local)

	tIsOk, newT := vt.FindAfterTime(nowT)
	assert.True(t, tIsOk)
	assert.EqualValues(t, nowT, newT)

	vt.ValidHoursInDay = &ValidRanges{
		ValidRanges: []ValidRange{
			{
				Start: 20,
				End:   22,
			},
		},
	}

	tIsOk, newT = vt.FindAfterTime(nowT)
	assert.False(t, tIsOk)
	assert.EqualValues(t, 2023, newT.Year())
	assert.EqualValues(t, 2, newT.Month())
	assert.EqualValues(t, 11, newT.Day())
	assert.EqualValues(t, 20, newT.Hour())
	assert.EqualValues(t, 0, newT.Minute())
}

func TestValid2(t *testing.T) {
	t.Log(int(time.Now().Weekday()))
}
