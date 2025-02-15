package timeassist

import (
	"fmt"
	"sync"
	"time"

	"github.com/s-min-sys/timeassistbe/internal/trace"
	"github.com/sgostarter/libeasygo/stg/kv"
)

type D struct {
	Data *ShowItem
	At   time.Time
}

type TaskTimer interface {
	Start()
	AddTimer(at time.Time, data *ShowItem) error
	SetCallback(cb Callback)
	List() (items []D, err error)
}

// BizTaskTimer 危险 确保 idPre 不重复 且 ShowItem 的 ID 符合规则
type BizTaskTimer interface {
	AddTimer(at time.Time, data *ShowItem) error
	SetCallback(idPre string, cb Callback)
}

func NewBizTimer(timer TaskTimer) BizTaskTimer {
	impl := &bizTimerImpl{
		timer:      timer,
		idCheckers: make(map[string]Callback),
	}

	timer.SetCallback(impl.timerCB)

	return impl
}

type bizTimerImpl struct {
	timer TaskTimer

	idCheckerLock sync.Mutex
	idCheckers    map[string]Callback
}

func (impl *bizTimerImpl) AddTimer(at time.Time, data *ShowItem) error {
	return impl.timer.AddTimer(at, data)
}

func (impl *bizTimerImpl) SetCallback(idPre string, cb Callback) {
	impl.idCheckers[idPre] = cb
}

func (impl *bizTimerImpl) checkCallback(id string) Callback {
	impl.idCheckerLock.Lock()
	defer impl.idCheckerLock.Unlock()

	idPre := ParsePreOnID(id)

	for s, callback := range impl.idCheckers {
		if s == idPre {
			return callback
		}
	}

	return nil
}

func (impl *bizTimerImpl) timerCB(dRemoved *ShowItem) (at time.Time, data *ShowItem, err error) {
	cb := impl.checkCallback(dRemoved.ID)
	if cb == nil {
		return
	}

	return cb(dRemoved)
}

func NewTaskTimer(fileName string) TaskTimer {
	storage, err := kv.NewMemoryFileStorageEx(fileName, false)
	if err != nil {
		return nil
	}

	return &taskTimerImpl{
		storage: storage,
	}
}

type taskTimerImpl struct {
	storage kv.StorageTiny
	cb      Callback
}

func (impl *taskTimerImpl) check() {
	ds, err := impl.storage.GetMap(func(_ string) interface{} {
		return &D{}
	})
	if err != nil {
		return
	}

	timeNow := time.Now()

	var at time.Time

	var data *ShowItem

	for k, v := range ds {
		d, ok := v.(*D)
		if !ok {
			continue
		}

		if timeNow.Before(d.At) {
			continue
		}

		at, data, err = impl.cb(d.Data)

		if err == nil && data != nil && data.ID != "" {
			if data.ID != d.Data.ID {
				panic("mismatched id")
			}

			err = impl.storage.Set(data.ID, &D{
				Data: data,
				At:   at,
			})

			if err != nil {
				trace.Get().RecordMessage(data.ID, fmt.Sprintf("add timer %s  failed: %v", at.String(), err))
			} else {
				trace.Get().RecordTimeSchedule(data.ID, at)
			}
		} else {
			_ = impl.storage.Del(k)

			trace.Get().RecordRemoveTimeSchedule(d.Data.ID)
		}
	}
}

func (impl *taskTimerImpl) Start() {
	go func() {
		for {
			impl.check()

			time.Sleep(time.Second * 10)
		}
	}()
}

func (impl *taskTimerImpl) AddTimer(at time.Time, data *ShowItem) error {
	err := impl.storage.Set(data.ID, &D{
		Data: data,
		At:   at,
	})

	if err != nil {
		trace.Get().RecordMessage(data.ID, fmt.Sprintf("add timer %s  failed: %v", at.String(), err))
	} else {
		trace.Get().RecordTimeSchedule(data.ID, at)
	}

	return err
}

func (impl *taskTimerImpl) SetCallback(cb Callback) {
	impl.cb = cb
}

func (impl *taskTimerImpl) List() (items []D, err error) {
	ds, err := impl.storage.GetMap(func(_ string) interface{} {
		return &D{}
	})
	if err != nil {
		return
	}

	items = make([]D, 0, len(ds))

	for _, v := range ds {
		d, ok := v.(*D)
		if !ok {
			continue
		}

		items = append(items, *d)
	}

	return
}
