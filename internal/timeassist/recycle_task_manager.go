package timeassist

import (
	"errors"
	"github.com/sgostarter/i/l"
	"time"

	uuid "github.com/satori/go.uuid"
)

var (
	ErrNotFound = errors.New("not found")
)

type RecycleTaskStorage interface {
	Add(task *RecycleTask) error
	Get(taskID string) (*RecycleTask, error)
	Remove(taskID string) error
}

type Callback func(dRemoved *RecycleData) (at time.Time, data *RecycleData, err error)

type RecycleTaskTimer interface {
	Start()
	AddTimer(at time.Time, data *RecycleData) error
	SetCallback(cb Callback)
}

type TaskInfo struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

type TaskInfos []*TaskInfo

func (s TaskInfos) Len() int           { return len(s) }
func (s TaskInfos) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s TaskInfos) Less(i, j int) bool { return s[i].Value < s[j].Value }

type TaskListChangeObserver func(task *TaskInfo, visible bool)

type RecycleTaskList interface {
	SetOb(ob TaskListChangeObserver) error
	Add(taskID, value string) error // 如果存在，也不要返回错误
	Remove(taskID string) error     // 如果不存在，也不要返回错误
	GetList() ([]*TaskInfo, error)
}

type RecycleTaskManager interface {
	Add(task *RecycleTask) error
	Remove(taskID string) error
	GetCurrentList() ([]*TaskInfo, error)
	Done(taskID string) error
}

func NewRecycleTaskManager(storage RecycleTaskStorage, timer RecycleTaskTimer, taskList RecycleTaskList, logger l.Wrapper) RecycleTaskManager {
	if logger == nil {
		logger = l.NewNopLoggerWrapper()
	}

	if storage == nil || taskList == nil {
		logger.Fatal("invalid construct parameters")
	}

	impl := &recycleTaskManagerImpl{
		logger:   logger.WithFields(l.StringField(l.ClsKey, "recycleTaskManagerImpl")),
		storage:  storage,
		timer:    timer,
		taskList: taskList,
	}

	impl.init()

	return impl
}

type recycleTaskManagerImpl struct {
	logger   l.Wrapper
	storage  RecycleTaskStorage
	timer    RecycleTaskTimer
	taskList RecycleTaskList
}

func (impl *recycleTaskManagerImpl) Done(taskID string) error {
	return impl.taskList.Remove(taskID)
}

func (impl *recycleTaskManagerImpl) init() {
	impl.timer.SetCallback(impl.timerCb)
	impl.timer.Start()
}

func (impl *recycleTaskManagerImpl) timerCb(dRemoved *RecycleData) (at time.Time, data *RecycleData, err error) {
	task, err := impl.storage.Get(dRemoved.ID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			_ = impl.taskList.Remove(task.ID)
			_ = impl.storage.Remove(task.ID)
		}

		return
	}

	err = impl.taskList.Remove(dRemoved.ID)
	if err != nil {
		_ = impl.storage.Remove(task.ID)

		return
	}

	if task == nil {
		return
	}

	rd, nowIsValid := task.GenRecycleDataEx(time.Unix(dRemoved.EndUTC, 0))

	if nowIsValid {
		err = impl.taskList.Add(task.ID, task.Text)
		if err != nil {
			_ = impl.storage.Remove(task.ID)

			return
		}

		if task.Auto {
			at = time.Unix(rd.EndUTC, 0)
			data = rd
		}
	} else {
		if time.Now().Unix() < rd.StartUTC {
			at = time.Unix(rd.StartUTC, 0)
		} else {
			at = time.Unix(rd.EndUTC, 0)
		}

		data = rd
	}

	if err != nil {
		_ = impl.storage.Remove(task.ID)
		_ = impl.taskList.Remove(task.ID)

		return
	}

	return
}

func (impl *recycleTaskManagerImpl) Add(task *RecycleTask) (err error) {
	if task == nil {
		return
	}

	if task.ID == "" {
		task.ID = uuid.NewV4().String()
	}

	err = task.Valid()
	if err != nil {
		return
	}

	err = impl.storage.Add(task)
	if err != nil {
		return
	}

	rd, nowIsValid := task.GenRecycleData()

	if nowIsValid {
		err = impl.taskList.Add(task.ID, task.Text)
		if err != nil {
			_ = impl.storage.Remove(task.ID)

			return
		}

		if task.Auto {
			err = impl.timer.AddTimer(time.Unix(rd.EndUTC, 0), rd)
		}
	} else {
		err = impl.timer.AddTimer(time.Unix(rd.StartUTC, 0), rd)
	}

	if err != nil {
		_ = impl.storage.Remove(task.ID)
		_ = impl.taskList.Remove(task.ID)

		return
	}

	return
}

func (impl *recycleTaskManagerImpl) Remove(taskID string) error {
	_ = impl.taskList.Remove(taskID)
	_ = impl.storage.Remove(taskID)

	return nil
}

func (impl *recycleTaskManagerImpl) GetCurrentList() ([]*TaskInfo, error) {
	return impl.taskList.GetList()
}
