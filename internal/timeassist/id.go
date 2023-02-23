package timeassist

import (
	"strings"

	uuid "github.com/satori/go.uuid"
)

const (
	TaskIDPre  = "T"
	AlarmIDPre = "A"
)

func ParsePreOnID(id string) string {
	if id == "" {
		return ""
	}

	pre := id[0:1]
	if pre != TaskIDPre && pre != AlarmIDPre {
		pre = ""
	}

	return pre
}

func FixTaskID(id string) string {
	return fixID(id, TaskIDPre)
}

func FixAlarmID(id string) string {
	return fixID(id, AlarmIDPre)
}

func fixID(id string, pre string) string {
	if id == "" {
		id = uuid.NewV4().String()
	}

	if !strings.HasPrefix(id, pre) {
		id = pre + id
	}

	return id
}
