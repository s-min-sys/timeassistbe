package timeassist

import (
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/sgostarter/i/commerr"
	"github.com/sgostarter/libeasygo/stg/kv"
)

type VOTaskType int

const (
	VOTaskTypeUnknown VOTaskType = iota
	VOTaskTypeTask
	VOTaskTypeAlarm
)

type TaskInfo struct {
	ID    string `json:"id"`
	Value string `json:"value"`

	/*
		once task:
		recycle task: start - end
		alarm: future/outdate
	*/
	SubTitle string `json:"sub_title"`

	//
	// alarm
	//

	AlarmFlag bool      `json:"alarm_flag,omitempty"`
	AlarmAt   time.Time `json:"alarm_at,omitempty"`

	//
	//
	//
	VOTaskType VOTaskType `json:"vo_task_type"`

	//
	//
	//
	NotifyID string `json:"notify_id,omitempty"`
}

func (taskInfo *TaskInfo) AutoFill() {
	switch ParsePreOnID(taskInfo.ID) {
	case TaskIDPre:
		taskInfo.VOTaskType = VOTaskTypeTask
	case AlarmIDPre:
		taskInfo.VOTaskType = VOTaskTypeAlarm
	default:
		taskInfo.VOTaskType = VOTaskTypeUnknown
	}
}

type TaskInfos []*TaskInfo

func (s TaskInfos) Len() int           { return len(s) }
func (s TaskInfos) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s TaskInfos) Less(i, j int) bool { return s[i].Value < s[j].Value }

type TaskListChangeObserver func(task *TaskInfo, visible bool)

type TaskList interface {
	SetOb(ob TaskListChangeObserver) error
	Add(taskInfo *TaskInfo) error // 如果存在，也不要返回错误
	Get(taskID string) (taskInfo *TaskInfo, err error)
	Remove(taskID string) error // 如果不存在，也不要返回错误
	GetList() ([]*TaskInfo, error)
}

func NewTaskList(fileName string, ob TaskListChangeObserver) TaskList {
	storage, err := kv.NewMemoryFileStorageEx(fileName, false)
	if err != nil {
		return nil
	}

	return &taskListImpl{
		storage:        storage,
		changeObserver: ob,
	}
}

type taskListImpl struct {
	storage        kv.StorageTiny
	changeObserver TaskListChangeObserver
}

func (impl *taskListImpl) SetOb(_ TaskListChangeObserver) error {
	return commerr.ErrUnavailable
}

func (impl *taskListImpl) Add(taskInfo *TaskInfo) (err error) {
	if taskInfo == nil || taskInfo.ID == "" {
		err = commerr.ErrInvalidArgument

		return
	}

	taskInfo.AutoFill()

	var taskInfoOld TaskInfo

	ok, err := impl.storage.Get(taskInfo.ID, &taskInfoOld)
	if err != nil {
		return
	}

	var forceUpdateNotifyID bool

	if ok {
		if taskInfo.VOTaskType == VOTaskTypeAlarm && taskInfo.AlarmFlag && !taskInfoOld.AlarmFlag {
			forceUpdateNotifyID = true
		}

		taskInfo.NotifyID = taskInfoOld.NotifyID

		if impl.changeObserver != nil {
			impl.changeObserver(&TaskInfo{
				ID:    taskInfoOld.ID,
				Value: taskInfoOld.Value,
			}, false)
		}
	}

	if forceUpdateNotifyID || taskInfo.NotifyID == "" {
		taskInfo.NotifyID = uuid.NewV4().String()
	}

	err = impl.storage.Set(taskInfo.ID, &taskInfo)
	if err != nil {
		return
	}

	tmpTaskInfo := *taskInfo
	impl.changeObserver(&tmpTaskInfo, true)

	return
}

func (impl *taskListImpl) Get(taskID string) (taskInfo *TaskInfo, err error) {
	taskInfo = &TaskInfo{}

	ok, err := impl.storage.Get(taskID, taskInfo)
	if err != nil {
		return
	}

	if !ok {
		taskInfo = nil
	}

	return
}

func (impl *taskListImpl) Remove(taskID string) (err error) {
	var taskInfo TaskInfo

	ok, err := impl.storage.Get(taskID, &taskInfo)
	if err != nil {
		return
	}

	if !ok {
		return
	}

	err = impl.storage.Del(taskID)
	if err != nil {
		return
	}

	if impl.changeObserver != nil {
		impl.changeObserver(&TaskInfo{
			ID:    taskID,
			Value: taskInfo.Value,
		}, false)
	}

	return
}

func (impl *taskListImpl) GetList() (tasks []*TaskInfo, err error) {
	ds, err := impl.storage.GetList(func(key string) interface{} {
		return &TaskInfo{}
	})
	if err != nil {
		return
	}

	for _, d := range ds {
		taskInfo, ok := d.(*TaskInfo)
		if !ok {
			continue
		}

		taskInfo.AutoFill()

		tasks = append(tasks, taskInfo)
	}

	return
}
