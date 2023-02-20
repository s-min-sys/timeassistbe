package timeassist

import (
	"testing"
	"time"
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
