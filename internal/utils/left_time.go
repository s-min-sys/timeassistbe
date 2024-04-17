package utils

import (
	"fmt"
	"time"
)

func LeftTimeString(at time.Time) string {
	timeNow := time.Now()
	if timeNow.After(at) {
		return "已经过期"
	}

	d := at.Sub(timeNow)
	if d > time.Hour*24*30 {
		return fmt.Sprintf("约%d月%d天", d/(time.Hour*24*30), d%(time.Hour*24*30)/(time.Hour*24))
	}

	if d > time.Hour*24 {
		return fmt.Sprintf("%d天%d小时", d/(time.Hour*24), d%(time.Hour*24)/time.Hour)
	}

	if d > time.Hour {
		return fmt.Sprintf("%d小时%d分", d/time.Hour, (d%time.Hour)/time.Minute)
	}

	if d > time.Minute {
		return fmt.Sprintf("%d分%d秒", d/time.Minute, (d%time.Minute)/time.Second)
	}

	return fmt.Sprintf("%d秒", d/time.Second)
}
