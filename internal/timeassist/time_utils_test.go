package timeassist

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWeekStart(t *testing.T) {
	t.Log(WeekStart(time.Now()).In(time.Local))
	t.Log(WeekStart(time.Now().In(time.FixedZone("Z", -11*3600))).In(time.Local))

	t.Log(MonthStart(time.Now()).In(time.Local))
	t.Log(MonthStart(time.Now().In(time.FixedZone("Z", -11*3600))).In(time.Local))

	t.Log(MonthAdd(time.Now(), 20).In(time.Local))
	t.Log(WeekAdd(WeekEnd(time.Now()), 2).In(time.Local))

	t.Log(MonthEnd(time.Now()).In(time.Local))

	t.Log(LunarMonthStart(time.Now()))
	t.Log(LunarMonthEnd(time.Now()))

	for idx := 0; idx < 10; idx++ {
		t.Log("idx => ", idx, LunarMonthEndAdd(time.Now(), idx))
	}

	t.Log(LunarYearEndAdd(time.Now(), 0))
	t.Log(LunarYearEndAdd(time.Now(), -1))

	t.Log(IsWorkDay(time.Now()))
}

func TestLunarToDateTimeAndNextMonth(t *testing.T) {
	for idx := 0; idx < 20; idx++ {
		timeAt, err := LunarToDateTimeAndNextMonth(2000, 10, 1, 1, 1, 1, idx)
		assert.Nil(t, err)

		t.Log(timeAt)
	}
}

func TestLunarMulMonth(t *testing.T) {
	timeAt, err := LunarToDateTimeAndNextMonth(2023, 2, 1, 1, 1, 1, 0)
	assert.Nil(t, err)
	t.Log(timeAt)

	timeAt, err = LunarToDateTimeAndNextMonth(2023, 2, 1, 1, 1, 1, 1)
	assert.Nil(t, err)
	t.Log(timeAt)

	timeAt, err = LunarToDateTimeAndNextMonth(2023, 3, 1, 1, 1, 1, 0)
	assert.Nil(t, err)
	t.Log(timeAt)
}

func TestToDateTimeAdd(t *testing.T) {
	month := 10

	for idx := 0; idx < 20; idx++ {
		timeAt := ToDateTime(2000, month+idx, 1, 1, 1, 1, time.Local)
		t.Log(timeAt)
	}
}
