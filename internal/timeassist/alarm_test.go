package timeassist

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// nolint
func utErrIsNil(f bool) func(assert.TestingT, error, ...interface{}) bool {
	return func(t assert.TestingT, err error, i ...interface{}) bool {
		if f {
			assert.Nil(t, err)

			return true
		}

		assert.NotNil(t, err)
		return false
	}
}

// nolint
func Test_parseAlarmValueOnce(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name    string
		args    args
		wantAv  *AlarmValue
		wantErr assert.ErrorAssertionFunc
	}{
		{"", args{"20230222135710"}, &AlarmValue{
			Lunar:  false,
			Year:   2023,
			Month:  2,
			Day:    22,
			Week:   0,
			Hour:   13,
			Minute: 57,
			Second: 10,
		}, utErrIsNil(true)},
		{"", args{"20230222135760"}, &AlarmValue{}, utErrIsNil(false)},
		{"", args{"L20230222135710"}, &AlarmValue{
			Lunar:  true,
			Year:   2023,
			Month:  2,
			Day:    22,
			Week:   0,
			Hour:   13,
			Minute: 57,
			Second: 10,
		}, utErrIsNil(true)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAv, err := parseAlarmValueOnce(tt.args.value)
			if !tt.wantErr(t, err, fmt.Sprintf("parseAlarmValueOnce(%v)", tt.args.value)) {
				return
			}
			assert.Equalf(t, tt.wantAv, gotAv, "parseAlarmValueOnce(%v)", tt.args.value)
		})
	}
}

// nolint
func Test_parseAlarmValueWeek(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name    string
		args    args
		wantAv  *AlarmValue
		wantErr assert.ErrorAssertionFunc
	}{
		{"", args{"3092400"}, &AlarmValue{
			Lunar:  false,
			Year:   0,
			Month:  0,
			Day:    0,
			Week:   3,
			Hour:   9,
			Minute: 24,
			Second: 0,
		}, utErrIsNil(true)},
		{"", args{"0092400"}, &AlarmValue{
			Lunar:  false,
			Year:   0,
			Month:  0,
			Day:    0,
			Week:   0,
			Hour:   9,
			Minute: 24,
			Second: 0,
		}, utErrIsNil(true)},
		{"", args{"192400"}, &AlarmValue{
			Lunar:  false,
			Year:   0,
			Month:  0,
			Day:    0,
			Week:   3,
			Hour:   9,
			Minute: 24,
			Second: 0,
		}, utErrIsNil(false)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAv, err := parseAlarmValueWeek(tt.args.value)
			if !tt.wantErr(t, err, fmt.Sprintf("parseAlarmValueWeek(%v)", tt.args.value)) {
				return
			}
			assert.Equalf(t, tt.wantAv, gotAv, "parseAlarmValueWeek(%v)", tt.args.value)
		})
	}
}

// nolint
func TestAlarm_GenRecycleDataEx(t *testing.T) {
	tz := time.FixedZone("X", 8*3600)

	type fields struct {
		ID       string
		AType    AlarmType
		Text     string
		Value    string
		TimeZone int
	}
	type args struct {
		timeNow time.Time
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wantRd         *TaskData
		wantNowIsValid bool
	}{
		{"", fields{
			ID:       "1",
			AType:    AlarmTypeOnce,
			Text:     "1",
			Value:    "S20230222092222",
			TimeZone: 8,
		}, args{
			timeNow: time.Date(2023, 2, 22, 9, 22, 20, 0, tz),
		}, &TaskData{
			ID:       "1",
			StartUTC: time.Date(2023, 2, 22, 9, 21, 22, 0, tz).Unix(),
			EndUTC:   time.Date(2023, 2, 22, 9, 22, 22, 0, tz).Unix(),
		}, true},
		{"", fields{
			ID:       "1",
			AType:    AlarmTypeOnce,
			Text:     "1",
			Value:    "20230222092220",
			TimeZone: 8,
		}, args{
			timeNow: time.Date(2023, 2, 12, 19, 22, 40, 0, tz),
		}, &TaskData{
			ID:       "1",
			StartUTC: time.Date(2023, 2, 19, 9, 22, 20, 0, tz).Unix(),
			EndUTC:   time.Date(2023, 2, 22, 9, 22, 20, 0, tz).Unix(),
		}, false},
		{"", fields{
			ID:       "1",
			AType:    AlarmTypeOnce,
			Text:     "1",
			Value:    "20230222092220",
			TimeZone: 8,
		}, args{
			timeNow: time.Date(2023, 2, 23, 19, 22, 40, 0, tz),
		}, nil, true},
		{"", fields{
			ID:       "1",
			AType:    AlarmTypeOnce,
			Text:     "1",
			Value:    "L20230203092220",
			TimeZone: 8,
		}, args{
			timeNow: time.Date(2023, 2, 3, 15, 22, 40, 0, tz),
		}, &TaskData{
			ID:       "1",
			StartUTC: time.Date(2023, 2, 19, 9, 22, 20, 0, tz).Unix(),
			EndUTC:   time.Date(2023, 2, 22, 9, 22, 20, 0, tz).Unix(),
		}, false},
		{"", fields{
			ID:       "1",
			AType:    AlarmTypeOnce,
			Text:     "1",
			Value:    "L20230203092220",
			TimeZone: 8,
		}, args{
			timeNow: time.Date(2023, 2, 20, 15, 22, 40, 0, tz),
		}, &TaskData{
			ID:       "1",
			StartUTC: time.Date(2023, 2, 21, 9, 22, 20, 0, tz).Unix(),
			EndUTC:   time.Date(2023, 2, 22, 9, 22, 20, 0, tz).Unix(),
		}, false},

		//
		//
		//

		{"", fields{
			ID:       "1",
			AType:    RecycleAlarmTypeYear,
			Text:     "1",
			Value:    "S0222092220",
			TimeZone: 8,
		}, args{
			timeNow: time.Date(2023, 2, 20, 15, 22, 40, 0, tz),
		}, &TaskData{
			ID:       "1",
			StartUTC: time.Date(2023, 2, 15, 9, 22, 20, 0, tz).Unix(),
			EndUTC:   time.Date(2023, 2, 22, 9, 22, 20, 0, tz).Unix(),
		}, true},
		{"", fields{
			ID:       "1",
			AType:    RecycleAlarmTypeYear,
			Text:     "1",
			Value:    "S0222092220",
			TimeZone: 8,
		}, args{
			timeNow: time.Date(2023, 2, 22, 15, 22, 40, 0, tz),
		}, &TaskData{
			ID:       "1",
			StartUTC: time.Date(2024, 2, 15, 9, 22, 20, 0, tz).Unix(),
			EndUTC:   time.Date(2024, 2, 22, 9, 22, 20, 0, tz).Unix(),
		}, false},

		//
		//
		//

		{"", fields{
			ID:       "1",
			AType:    RecycleAlarmTypeWeek,
			Text:     "1",
			Value:    "2092220",
			TimeZone: 8,
		}, args{
			timeNow: time.Date(2023, 2, 20, 15, 22, 40, 0, tz),
		}, &TaskData{
			ID:       "1",
			StartUTC: time.Date(2023, 2, 20, 9, 22, 20, 0, tz).Unix(),
			EndUTC:   time.Date(2023, 2, 21, 9, 22, 20, 0, tz).Unix(),
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Alarm{
				ID:       tt.fields.ID,
				AType:    tt.fields.AType,
				Text:     tt.fields.Text,
				Value:    tt.fields.Value,
				TimeZone: tt.fields.TimeZone,
			}
			_, _, gotRd, gotNowIsValid, _, _ := a.GenRecycleDataEx(tt.args.timeNow, tt.args.timeNow)
			assert.Equalf(t, tt.wantRd, gotRd, "GenRecycleDataEx(%v)", tt.args.timeNow)
			assert.Equalf(t, tt.wantNowIsValid, gotNowIsValid, "GenRecycleDataEx(%v)", tt.args.timeNow)
		})
	}
}
