package timeassist

import (
	"time"

	"github.com/s-min-sys/timeassistbe/internal/utils"
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

type ShowInfo struct {
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

	//
	LeftTimeS string `json:"left_time_s"`
}

func (showInfo *ShowInfo) AutoFill() {
	switch ParsePreOnID(showInfo.ID) {
	case TaskIDPre:
		showInfo.VOTaskType = VOTaskTypeTask
		showInfo.LeftTimeS = ""
	case AlarmIDPre:
		showInfo.VOTaskType = VOTaskTypeAlarm
		showInfo.LeftTimeS = utils.LeftTimeString(showInfo.AlarmAt)
	default:
		showInfo.VOTaskType = VOTaskTypeUnknown
	}
}

type ShowInfoListChangeObserver func(task *ShowInfo, visible bool)

type ShowList interface {
	SetOb(ob ShowInfoListChangeObserver) error
	Add(taskInfo *ShowInfo) error // 如果存在，也不要返回错误
	Get(taskID string) (taskInfo *ShowInfo, err error)
	Remove(taskID string) error // 如果不存在，也不要返回错误
	GetList() ([]*ShowInfo, error)
}

func NewShowList(fileName string, ob ShowInfoListChangeObserver) ShowList {
	storage, err := kv.NewMemoryFileStorageEx(fileName, false)
	if err != nil {
		return nil
	}

	return &showListImpl{
		storage:        storage,
		changeObserver: ob,
	}
}

type showListImpl struct {
	storage        kv.StorageTiny
	changeObserver ShowInfoListChangeObserver
}

func (impl *showListImpl) SetOb(_ ShowInfoListChangeObserver) error {
	return commerr.ErrUnavailable
}

func (impl *showListImpl) Add(taskInfo *ShowInfo) (err error) {
	if taskInfo == nil || taskInfo.ID == "" {
		err = commerr.ErrInvalidArgument

		return
	}

	taskInfo.AutoFill()

	var taskInfoOld ShowInfo

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
			impl.changeObserver(&ShowInfo{
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

	if impl.changeObserver != nil {
		impl.changeObserver(&tmpTaskInfo, true)
	}

	return
}

func (impl *showListImpl) Get(taskID string) (taskInfo *ShowInfo, err error) {
	taskInfo = &ShowInfo{}

	ok, err := impl.storage.Get(taskID, taskInfo)
	if err != nil {
		return
	}

	if !ok {
		taskInfo = nil
	}

	return
}

func (impl *showListImpl) Remove(taskID string) (err error) {
	var taskInfo ShowInfo

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
		impl.changeObserver(&ShowInfo{
			ID:    taskID,
			Value: taskInfo.Value,
		}, false)
	}

	return
}

func (impl *showListImpl) GetList() (tasks []*ShowInfo, err error) {
	ds, err := impl.storage.GetList(func(_ string) interface{} {
		return &ShowInfo{}
	})
	if err != nil {
		return
	}

	for _, d := range ds {
		taskInfo, ok := d.(*ShowInfo)
		if !ok {
			continue
		}

		taskInfo.AutoFill()

		tasks = append(tasks, taskInfo)
	}

	return
}
