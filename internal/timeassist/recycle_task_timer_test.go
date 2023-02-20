package timeassist

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

const (
	utTestTimerFile = "recycle_task_timer.txt"
)

func TestNewRecycleTaskTimer(t *testing.T) {
	_ = os.Remove(utTestTimerFile)

	timer := NewRecycleTaskTimer(utTestTimerFile)

	timer.SetCallback(func(dRemoved *RecycleData) (at time.Time, data *RecycleData, err error) {
		t.Log("timeNow:", time.Now(), ", data", dRemoved)

		return
	})

	err := timer.AddTimer(time.Now().Add(time.Minute), &RecycleData{
		ID:       "1",
		StartUTC: time.Now().Unix(),
		EndUTC:   time.Now().Add(time.Second * 40).Unix(),
	})
	assert.Nil(t, err)

	err = timer.AddTimer(time.Now().Add(time.Second*10), &RecycleData{
		ID:       "2",
		StartUTC: time.Now().Unix(),
		EndUTC:   time.Now().Add(time.Second * 40).Unix(),
	})
	assert.Nil(t, err)

	time.Sleep(time.Minute * 2)
}
