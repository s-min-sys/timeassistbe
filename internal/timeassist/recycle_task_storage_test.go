package timeassist

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

const (
	utTestStorageFile = "recycle_task_storage.txt"
)

func Test1(t *testing.T) {
	_ = os.Remove(utTestStorageFile)
	stg := NewRecycleTaskStorage(utTestStorageFile)
	task, err := stg.Get("1")
	assert.Nil(t, err)
	assert.Nil(t, task)

	err = stg.Add(&RecycleTask{
		ID:       "1",
		TType:    RecycleTaskTypeDays,
		Value:    2,
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
		},
		Text: "test 1",
	})
	assert.Nil(t, err)

	task, err = stg.Get("1")
	assert.Nil(t, err)
	assert.NotNil(t, task)
	t.Log(task)
}
