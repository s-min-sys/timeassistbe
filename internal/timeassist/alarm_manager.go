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

func NewAlarmManager(storage kv.Storage, timer BizTaskTimer, taskList ShowList, logger l.Wrapper) AlarmManager {
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
	taskList ShowList
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

	alarm.TimeLastAt = 0

	err = impl.storage.Set(alarm.ID, alarm)
	if err != nil {
		return
	}

	if show {
		_ = impl.taskList.Add(&ShowInfo{
			ID:        alarm.ID,
			Value:     alarm.Text,
			SubTitle:  av.String(alarm.AType, timeAt),
			AlarmFlag: alarmFlag,
			AlarmAt:   timeAt,
		})

		if rd != nil {
			err = impl.timer.AddTimer(time.Unix(rd.EndUTC, 0), rd)
		}

		alarm.TimeLastAt = rd.EndUTC
		_ = impl.storage.Set(alarm.ID, alarm)
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
	alarm := &Alarm{}

	ok, err := impl.storage.Get(id, alarm)
	if err == nil && ok {
		alarm.TimeLastAt = 0

		_ = impl.storage.Set(id, alarm)
	}

	return impl.taskList.Remove(id)
}

func (impl *alarmManagerImpl) init() {
	impl.timer.SetCallback(AlarmIDPre, impl.timerCb)
}

func (impl *alarmManagerImpl) timerCb(dRemoved *ShowItem) (at time.Time, data *ShowItem, err error) {
	alarm := &Alarm{}

	ok, err := impl.storage.Get(dRemoved.ID, alarm)
	if err != nil || !ok {
		return
	}

	timeNow := time.Now()
	timeLastAt := timeNow

	if alarm.TimeLastAt > 0 {
		timeLastAt = time.Unix(alarm.TimeLastAt, 0)
	}

	av, timeAt, rd, show, alarmFlag, err := alarm.GenRecycleDataEx(timeNow, timeLastAt)
	if err != nil {
		return
	}

	if alarmFlag {
		_ = impl.taskList.Add(&ShowInfo{
			ID:        alarm.ID,
			Value:     alarm.Text,
			SubTitle:  av.String(alarm.AType, timeLastAt) + " - 过期",
			AlarmFlag: alarmFlag,
			AlarmAt:   timeAt,
		})

		if rd != nil {
			at = time.Unix(rd.StartUTC, 0)
			data = rd
		}
	} else if show {
		_ = impl.taskList.Add(&ShowInfo{
			ID:        alarm.ID,
			Value:     alarm.Text,
			SubTitle:  av.String(alarm.AType, timeAt),
			AlarmFlag: alarmFlag,
			AlarmAt:   timeAt,
		})

		if rd != nil {
			at = time.Unix(rd.EndUTC, 0)
			data = rd
		}

		alarm.TimeLastAt = rd.EndUTC
		_ = impl.storage.Set(dRemoved.ID, alarm)
	} else {
		at = time.Unix(rd.StartUTC, 0)
		data = rd
	}

	return
}
