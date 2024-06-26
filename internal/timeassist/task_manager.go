package timeassist

import (
	"time"

	"github.com/sgostarter/i/l"
	"github.com/sgostarter/libeasygo/stg/kv"
)

// Callback 如果data存在， 则其ID必须为dRemoved的ID
type Callback func(dRemoved *ShowItem) (at time.Time, data *ShowItem, err error)

type TaskManager interface {
	Add(task *Task) error
	Remove(taskID string) error
	Done(taskID string) error
	TaskDone(taskID string)
}

func NewTaskManager(storage kv.Storage, timer BizTaskTimer, taskList ShowList, logger l.Wrapper) TaskManager {
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
		showList: taskList,
	}

	impl.init()

	return impl
}

type taskManagerImpl struct {
	logger   l.Wrapper
	storage  kv.Storage
	timer    BizTaskTimer
	showList ShowList
}

func (impl *taskManagerImpl) init() {
	impl.timer.SetCallback(TaskIDPre, impl.timerCb)
}

func (impl *taskManagerImpl) formatTaskSubTitle(task *Task, taskData *ShowItem) string {
	var timeLayout string

	switch task.TType {
	case RecycleTimeTypeMinute:
		timeLayout = "15时04分"
	case RecycleTimeTypeHour:
		timeLayout = "02号15时"
	case RecycleTimeTypeDay:
		timeLayout = "01月02号"
	case RecycleTimeTypeWeek:
		timeLayout = "01月02号"
	case RecycleTimeTypeMonth:
		timeLayout = "2006年01月02号" // nolint: goconst
	case RecycleTimeTypeYear:
		timeLayout = "2006年01月02号" // nolint: goconst
	case TimeTypeOnce:
		return ""
	}

	return time.Unix(taskData.StartUTC, 0).Format(timeLayout) + "-" +
		time.Unix(taskData.EndUTC, 0).Format(timeLayout)
}

func (impl *taskManagerImpl) TaskDone(taskID string) {
	if ParsePreOnID(taskID) != TaskIDPre {
		return
	}

	var task Task

	ok, err := impl.storage.Get(taskID, &task)
	if err != nil {
		return
	}

	if !ok {
		return
	}

	if task.TType == TimeTypeOnce {
		return
	}

	timeNow := time.Now()

	rd, _ := task.GenRecycleDataEx(timeNow)
	if rd == nil {
		return
	}

	if timeNow.Before(time.Unix(rd.StartUTC, 0)) {
		_ = impl.timer.AddTimer(time.Unix(rd.StartUTC, 0), rd)
	} else {
		_ = impl.timer.AddTimer(time.Unix(rd.EndUTC, 0), rd)
	}
}

func (impl *taskManagerImpl) timerCb(dRemoved *ShowItem) (at time.Time, data *ShowItem, err error) {
	showInfo, err := impl.showList.Get(dRemoved.ID)
	if err != nil {
		return
	}

	task := &Task{}

	ok, err := impl.storage.Get(dRemoved.ID, task)
	if err != nil {
		return
	}

	if !ok {
		if showInfo != nil {
			_ = impl.showList.Remove(dRemoved.ID)
		}

		return
	}

	timeNow := time.Now()

	if timeNow.Before(time.Unix(dRemoved.StartUTC, 0)) {
		at = time.Unix(dRemoved.StartUTC, 0)
		data = dRemoved

		if showInfo != nil {
			_ = impl.showList.Remove(dRemoved.ID)
		}

		return
	}

	if timeNow.Before(time.Unix(dRemoved.EndUTC, 0)) {
		at = time.Unix(dRemoved.EndUTC, 0)
		data = dRemoved

		_ = impl.showList.Add(&ShowInfo{
			ID:       task.ID,
			Value:    task.Text,
			SubTitle: impl.formatTaskSubTitle(task, dRemoved),
		})

		return
	}

	if !task.Auto {
		if showInfo != nil {
			showInfo.AlarmFlag = true

			_ = impl.showList.Add(showInfo)

			return
		}
	}

	rd, nowIsValid := task.GenRecycleDataEx(time.Unix(dRemoved.EndUTC, 0))
	if nowIsValid {
		_ = impl.showList.Add(&ShowInfo{
			ID:       task.ID,
			Value:    task.Text,
			SubTitle: impl.formatTaskSubTitle(task, rd),
		})

		at = time.Unix(rd.EndUTC, 0)
		data = rd
	} else {
		if showInfo != nil {
			_ = impl.showList.Remove(dRemoved.ID)
		}

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

	if task.TType == TimeTypeOnce {
		err = impl.showList.Add(&ShowInfo{
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
		err = impl.showList.Add(&ShowInfo{
			ID:       task.ID,
			Value:    task.Text,
			SubTitle: impl.formatTaskSubTitle(task, rd),
		})
		if err != nil {
			_ = impl.storage.Del(task.ID)

			return
		}

		err = impl.timer.AddTimer(time.Unix(rd.EndUTC, 0), rd)
	} else {
		err = impl.timer.AddTimer(time.Unix(rd.StartUTC, 0), rd)
	}

	if err != nil {
		_ = impl.storage.Del(task.ID)
		_ = impl.showList.Remove(task.ID)

		return
	}

	return
}

func (impl *taskManagerImpl) Done(taskID string) error {
	return impl.showList.Remove(taskID)
}

func (impl *taskManagerImpl) Remove(taskID string) error {
	_ = impl.showList.Remove(taskID)
	_ = impl.storage.Del(taskID)

	return nil
}
