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

func Test11(t *testing.T) {
	b := []byte{0x01, 0x02, 0x00, 0x08}
	t.Log(b)
	s := string(b)
	t.Log(s)

	s2 := s
	t.Log(s2)

	s2 += "^"
	b2 := []byte(s2)
	t.Log(b2)
}

func TestNewRecycleTaskTimer(t *testing.T) {
	_ = os.Remove(utTestTimerFile)

	timer := NewTaskTimer(utTestTimerFile)

	timer.SetCallback(func(dRemoved *TaskData) (at time.Time, data *TaskData, err error) {
		t.Log("timeNow:", time.Now(), ", data", dRemoved)

		return
	})

	err := timer.AddTimer(time.Now().Add(time.Minute), &TaskData{
		ID:       "1",
		StartUTC: time.Now().Unix(),
		EndUTC:   time.Now().Add(time.Second * 40).Unix(),
	})
	assert.Nil(t, err)

	err = timer.AddTimer(time.Now().Add(time.Second*10), &TaskData{
		ID:       "2",
		StartUTC: time.Now().Unix(),
		EndUTC:   time.Now().Add(time.Second * 40).Unix(),
	})
	assert.Nil(t, err)

	time.Sleep(time.Minute * 2)
}
