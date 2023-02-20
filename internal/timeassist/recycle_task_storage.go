package timeassist

import (
	"errors"
	"gopkg.in/yaml.v3"
	"os"
	"sync"
)

var (
	ErrorNotImplement = errors.New("not implement")
)

func NewRecycleTaskStorage(fileName string) RecycleTaskStorage {
	impl := &recycleTaskStorageImpl{
		fileName: fileName,
		m:        make(map[string]*RecycleTask),
	}

	impl.init()

	return impl
}

type recycleTaskStorageImpl struct {
	fileName string

	mLock sync.Mutex
	m     map[string]*RecycleTask
}

func (impl *recycleTaskStorageImpl) init() {
	_ = impl.load()
}

func (impl *recycleTaskStorageImpl) load() error {
	if impl.fileName == "" {
		return ErrorNotImplement
	}

	d, err := os.ReadFile(impl.fileName)
	if err != nil {
		return err
	}

	var m map[string]*RecycleTask
	err = yaml.Unmarshal(d, &m)
	if err != nil {
		return err
	}

	impl.m = m

	if impl.m == nil {
		impl.m = make(map[string]*RecycleTask)
	}

	return nil
}

func (impl *recycleTaskStorageImpl) save() error {
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

func (impl *recycleTaskStorageImpl) Add(task *RecycleTask) error {
	impl.mLock.Lock()
	defer impl.mLock.Unlock()

	impl.m[task.ID] = task

	return impl.save()
}

func (impl *recycleTaskStorageImpl) Get(taskID string) (*RecycleTask, error) {
	impl.mLock.Lock()
	defer impl.mLock.Unlock()

	if task, ok := impl.m[taskID]; ok {
		return task, nil
	}

	return nil, nil
}

func (impl *recycleTaskStorageImpl) Remove(taskID string) error {
	impl.mLock.Lock()
	defer impl.mLock.Unlock()

	delete(impl.m, taskID)

	return impl.save()
}
