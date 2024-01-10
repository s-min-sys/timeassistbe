package timeassist

import (
	"time"

	"github.com/sgostarter/i/l"
	"github.com/sgostarter/libeasygo/stg/kv"
)

// Callback 如果data存在， 则其ID必须为dRemoved的ID
type Callback func(dRemoved *TaskData) (at time.Time, data *TaskData, err error)

type TaskManager interface {
	Add(task *Task) error
	Remove(taskID string) error
	Done(taskID string) error
}

func NewTaskManager(storage kv.Storage, timer BizTaskTimer, taskList TaskList, logger l.Wrapper) TaskManager {
	if logger == nil {
		logger = l.NewNopLoggerWrapper()
	}

	if storage == nil || timer == nil || taskList == nil {
		logger.Fatal("invalid construct parameters")
	}

	impl := &taskManagerImpl{
		logger:   logger.WithFields(l.StringField(l.ClsKey, "taskManagerImpl")),
		storage:  storage,
		timer:    timer,
		taskList: taskList,
	}

	impl.init()

	return impl
}

type taskManagerImpl struct {
	logger   l.Wrapper
	storage  kv.Storage
	timer    BizTaskTimer
	taskList TaskList
}

func (impl *taskManagerImpl) init() {
	impl.timer.SetCallback(TaskIDPre, impl.timerCb)
}

func (impl *taskManagerImpl) formatTaskSubTitle(task *Task, taskData *TaskData) string {
	var timeLayout string

	switch task.TType {
	case RecycleTaskTypeMinutes:
		timeLayout = "15时04分"
	case RecycleTaskTypeHours:
		timeLayout = "02号15时"
	case RecycleTaskTypeDays:
		timeLayout = "01月02号"
	case RecycleTaskTypeWeeks:
		timeLayout = "01月02号[Mon]"
	case RecycleTaskTypeMonths:
		timeLayout = "2006年01月02号" // nolint: goconst
	case RecycleTaskTypeLunarMonths:
		timeLayout = "2006年01月02号" // nolint: goconst
	case RecycleTaskTypeYears:
		timeLayout = "2006年01月02号" // nolint: goconst
	case RecycleTaskTypeLunarYears:
		timeLayout = "2006年01月02号" // nolint: goconst
	case OnceTask:
		return ""
	}

	return time.Unix(taskData.StartUTC, 0).Format(timeLayout) + "-" +
		time.Unix(taskData.EndUTC, 0).Format(timeLayout)
}

func (impl *taskManagerImpl) timerCb(dRemoved *TaskData) (at time.Time, data *TaskData, err error) {
	_ = impl.taskList.Remove(dRemoved.ID)

	task := &Task{}

	ok, err := impl.storage.Get(dRemoved.ID, task)
	if err != nil || !ok {
		return
	}

	rd, nowIsValid := task.GenRecycleDataEx(time.Unix(dRemoved.EndUTC, 0))
	if nowIsValid {
		_ = impl.taskList.Add(&TaskInfo{
			ID:       task.ID,
			Value:    task.Text,
			SubTitle: impl.formatTaskSubTitle(task, rd),
		})

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

	return
}

//
//
//

func (impl *taskManagerImpl) Add(task *Task) (err error) {
	if task == nil {
		return
	}

	task.ID = FixTaskID(task.ID)

	err = task.Valid()
	if err != nil {
		return
	}

	err = impl.storage.Set(task.ID, task)
	if err != nil {
		return
	}

	if task.TType == OnceTask {
		err = impl.taskList.Add(&TaskInfo{
			ID:       task.ID,
			Value:    task.Text,
			SubTitle: "单次任务",
		})
		if err != nil {
			_ = impl.storage.Del(task.ID)
		}

		return
	}

	rd, nowIsValid := task.GenRecycleData()
	if nowIsValid {
		err = impl.taskList.Add(&TaskInfo{
			ID:       task.ID,
			Value:    task.Text,
			SubTitle: impl.formatTaskSubTitle(task, rd),
		})
		if err != nil {
			_ = impl.storage.Del(task.ID)

			return
		}

		if task.Auto {
			err = impl.timer.AddTimer(time.Unix(rd.EndUTC, 0), rd)
		}
	} else {
		err = impl.timer.AddTimer(time.Unix(rd.StartUTC, 0), rd)
	}

	if err != nil {
		_ = impl.storage.Del(task.ID)
		_ = impl.taskList.Remove(task.ID)

		return
	}

	return
}

func (impl *taskManagerImpl) Done(taskID string) error {
	return impl.taskList.Remove(taskID)
}

func (impl *taskManagerImpl) Remove(taskID string) error {
	_ = impl.taskList.Remove(taskID)
	_ = impl.storage.Del(taskID)

	return nil
}
