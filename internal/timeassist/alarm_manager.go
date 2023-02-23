package timeassist

import (
	"time"

	"github.com/sgostarter/i/l"
	"github.com/sgostarter/libeasygo/stg/kv"
)

type AlarmManager interface {
	Add(alarm *Alarm) error
	Remove(id string) error
	Done(id string) error
}

func NewAlarmManager(storage kv.Storage, timer BizTaskTimer, taskList TaskList, logger l.Wrapper) AlarmManager {
	if logger == nil {
		logger = l.NewNopLoggerWrapper()
	}

	if storage == nil || timer == nil || taskList == nil {
		logger.Fatal("invalid construct parameters")
	}

	impl := &alarmManagerImpl{
		logger:   logger.WithFields(l.StringField(l.ClsKey, "alarmManagerImpl")),
		storage:  storage,
		timer:    timer,
		taskList: taskList,
	}

	impl.init()

	return impl
}

type alarmManagerImpl struct {
	logger   l.Wrapper
	storage  kv.Storage
	timer    BizTaskTimer
	taskList TaskList
}

func (impl *alarmManagerImpl) Add(alarm *Alarm) (err error) {
	if alarm == nil {
		return
	}

	alarm.ID = FixAlarmID(alarm.ID)

	av, timeAt, rd, show, alarmFlag, err := alarm.GenRecycleData()
	if err != nil {
		return
	}

	err = impl.storage.Set(alarm.ID, alarm)
	if err != nil {
		return
	}

	if show {
		var expire string
		if alarmFlag {
			expire = " - 已过期"
		}

		_ = impl.taskList.Add(&TaskInfo{
			ID:        alarm.ID,
			Value:     alarm.Text,
			SubTitle:  av.String(alarm.AType, timeAt) + expire,
			AlarmFlag: alarmFlag,
			AlarmLast: timeAt,
		})

		if rd != nil {
			err = impl.timer.AddTimer(time.Unix(rd.EndUTC, 0), rd)
		}
	} else {
		err = impl.timer.AddTimer(time.Unix(rd.StartUTC, 0), rd)
	}

	if err != nil {
		_ = impl.storage.Del(alarm.ID)
	}

	return
}

func (impl *alarmManagerImpl) Remove(id string) error {
	_ = impl.taskList.Remove(id)
	_ = impl.storage.Del(id)

	return nil
}

func (impl *alarmManagerImpl) Done(id string) error {
	return impl.taskList.Remove(id)
}

func (impl *alarmManagerImpl) init() {
	impl.timer.SetCallback(AlarmIDPre, impl.timerCb)
}

func (impl *alarmManagerImpl) timerCb(dRemoved *TaskData) (at time.Time, data *TaskData, err error) {
	alarm := &Alarm{}

	ok, err := impl.storage.Get(dRemoved.ID, alarm)
	if err != nil || !ok {
		return
	}

	timeNow := time.Now()
	timeLastAt := timeNow

	tryUpdateTimeLastAt := true

	if taskInList, _ := impl.taskList.Get(dRemoved.ID); taskInList != nil {
		timeLastAt = taskInList.AlarmLast

		tryUpdateTimeLastAt = false
	}

	av, timeAt, rd, show, alarmFlag, err := alarm.GenRecycleDataEx(timeNow, timeLastAt)
	if err != nil {
		return
	}

	if show {
		var expire string
		if alarmFlag {
			if tryUpdateTimeLastAt && timeLastAt == timeAt {
				expire = " - 已过期"
			} else {
				expire = " - 有过期"
			}
		}

		alarmLast := timeAt
		if !tryUpdateTimeLastAt {
			alarmLast = timeLastAt
		}
		_ = impl.taskList.Add(&TaskInfo{
			ID:        alarm.ID,
			Value:     alarm.Text,
			SubTitle:  av.String(alarm.AType, timeAt) + expire,
			AlarmFlag: alarmFlag,
			AlarmLast: alarmLast,
		})

		if rd != nil {
			at = time.Unix(rd.EndUTC, 0)
			data = rd
		}

	} else {
		_ = impl.taskList.Remove(dRemoved.ID)

		at = time.Unix(rd.StartUTC, 0)

		data = rd
	}

	return
}
