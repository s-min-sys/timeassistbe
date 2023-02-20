package timeassist

import (
	"gopkg.in/yaml.v3"
	"os"
	"sync"
	"time"
)

func NewRecycleTaskTimer(fileName string) RecycleTaskTimer {
	impl := &recycleTaskTimerImpl{
		fileName: fileName,
		m:        make(map[string]*D),
	}

	impl.init()

	return impl
}

type D struct {
	Data *RecycleData
	At   time.Time
}

type recycleTaskTimerImpl struct {
	fileName string
	cb       Callback

	mLock sync.Mutex
	m     map[string]*D
}

func (impl *recycleTaskTimerImpl) init() {
	impl.load()
}

func (impl *recycleTaskTimerImpl) check() {
	impl.mLock.Lock()
	defer impl.mLock.Unlock()

	newM := make(map[string]*D)

	var needSave bool

	timeNow := time.Now()
	for s, d := range impl.m {
		if timeNow.After(d.At) {
			needSave = true

			delete(impl.m, s)

			at, data, err := impl.cb(d.Data)
			if err == nil && data != nil && data.ID != "" {
				newM[data.ID] = &D{
					Data: data,
					At:   at,
				}
			}
		}
	}

	for s, d := range newM {
		impl.m[s] = d
	}

	if needSave {
		impl.save()
	}
}

func (impl *recycleTaskTimerImpl) Start() {
	go func() {
		for {
			impl.check()

			time.Sleep(time.Minute)
		}
	}()
}

func (impl *recycleTaskTimerImpl) AddTimer(at time.Time, data *RecycleData) error {
	impl.mLock.Lock()
	defer impl.mLock.Unlock()

	impl.m[data.ID] = &D{
		Data: data,
		At:   at,
	}

	impl.save()

	return nil
}

func (impl *recycleTaskTimerImpl) SetCallback(cb Callback) {
	impl.cb = cb
}

func (impl *recycleTaskTimerImpl) load() {
	if impl.fileName == "" {
		return
	}

	d, err := os.ReadFile(impl.fileName)
	if err != nil {
		return
	}

	var m map[string]*D

	err = yaml.Unmarshal(d, &m)
	if err != nil {
		return
	}

	impl.m = m

	if impl.m == nil {
		impl.m = make(map[string]*D)
	}
}

func (impl *recycleTaskTimerImpl) save() {
	if impl.fileName == "" {
		return
	}

	d, err := yaml.Marshal(impl.m)
	if err != nil {
		return
	}

	_ = os.WriteFile(impl.fileName, d, os.ModePerm)
}
