package timeassist

import (
	"gopkg.in/yaml.v3"
	"os"
	"sync"
)

func NewRecycleTaskList(fileName string, ob TaskListChangeObserver) RecycleTaskList {
	impl := &recycleTaskListImpl{
		fileName:       fileName,
		m:              make(map[string]string),
		changeObserver: ob,
	}

	impl.init()

	return impl
}

type recycleTaskListImpl struct {
	fileName string

	mLock          sync.Mutex
	m              map[string]string
	changeObserver TaskListChangeObserver
}

func (impl *recycleTaskListImpl) SetOb(ob TaskListChangeObserver) error {
	return ErrorNotImplement
}

func (impl *recycleTaskListImpl) init() {
	_ = impl.load()
}

func (impl *recycleTaskListImpl) load() error {
	if impl.fileName == "" {
		return ErrorNotImplement
	}

	d, err := os.ReadFile(impl.fileName)
	if err != nil {
		return err
	}

	var m map[string]string
	err = yaml.Unmarshal(d, &m)
	if err != nil {
		return err
	}

	impl.m = m

	if impl.m == nil {
		impl.m = make(map[string]string)
	}

	return nil
}

func (impl *recycleTaskListImpl) save() error {
	if impl.fileName == "" {
		return ErrorNotImplement
	}

	d, err := yaml.Marshal(impl.m)
	if err != nil {
		return err
	}

	err = os.WriteFile(impl.fileName, d, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func (impl *recycleTaskListImpl) Add(taskID, value string) error {
	impl.mLock.Lock()
	defer impl.mLock.Unlock()

	if v, ok := impl.m[taskID]; ok {
		if v == value {
			return nil
		}
		delete(impl.m, taskID)

		if impl.changeObserver != nil {
			impl.changeObserver(&TaskInfo{
				ID:    taskID,
				Value: v,
			}, false)
		}
	}

	impl.m[taskID] = value

	impl.changeObserver(&TaskInfo{
		ID:    taskID,
		Value: value,
	}, true)

	return impl.save()
}

func (impl *recycleTaskListImpl) Remove(taskID string) error {
	impl.mLock.Lock()
	defer impl.mLock.Unlock()

	v, ok := impl.m[taskID]
	if !ok {
		return nil
	}

	delete(impl.m, taskID)

	if impl.changeObserver != nil {
		impl.changeObserver(&TaskInfo{
			ID:    taskID,
			Value: v,
		}, false)
	}

	return impl.save()
}

func (impl *recycleTaskListImpl) GetList() (tasks []*TaskInfo, err error) {
	for s, s2 := range impl.m {
		tasks = append(tasks, &TaskInfo{
			ID:    s,
			Value: s2,
		})
	}

	return
}
