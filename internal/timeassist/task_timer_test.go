package timeassist

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

	timer.SetCallback(func(dRemoved *ShowItem) (at time.Time, data *ShowItem, err error) {
		t.Log("timeNow:", time.Now(), ", data", dRemoved)

		return
	})

	err := timer.AddTimer(time.Now().Add(time.Minute), &ShowItem{
		ID:       "1",
		StartUTC: time.Now().Unix(),
		EndUTC:   time.Now().Add(time.Second * 40).Unix(),
	})
	assert.Nil(t, err)

	err = timer.AddTimer(time.Now().Add(time.Second*10), &ShowItem{
		ID:       "2",
		StartUTC: time.Now().Unix(),
		EndUTC:   time.Now().Add(time.Second * 40).Unix(),
	})
	assert.Nil(t, err)

	timer.Start()

	time.Sleep(time.Minute * 2)
}
