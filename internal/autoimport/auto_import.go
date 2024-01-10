package autoimport

import (
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/s-min-sys/timeassistbe/internal/timeassist"
	"github.com/sgostarter/i/l"
	"gopkg.in/yaml.v3"
)

func TryImportTaskConfigs(root string, fileSuffix string, taskManger timeassist.TaskManager, logger l.Wrapper) {
	logger.Debug("TryImportTaskConfigs root:", root)

	_ = filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, fileSuffix) {
			return nil
		}

		tryImportTaskConfigs(path, taskManger, logger)

		return nil
	})
}

func tryImportTaskConfigs(file string, taskManger timeassist.TaskManager, logger l.Wrapper) {
	d, err := os.ReadFile(file)
	if err != nil && len(d) == 0 {
		return
	}

	var tasks []timeassist.Task

	err = yaml.Unmarshal(d, &tasks)
	if err != nil {
		logger.WithFields(l.ErrorField(err)).Error("invalid import config file format")

		return
	}

	if len(tasks) == 0 {
		return
	}

	for idx := 0; idx < len(tasks); idx++ {
		task := tasks[idx]
		if err = taskManger.Add(&task); err != nil {
			logger.WithFields(l.ErrorField(err), l.StringField("id", task.ID)).Error("try import task failed")
		} else {
			logger.WithFields(l.StringField("id", task.ID)).Error("try import task successful")
		}
	}

	err = os.Rename(file, file+"."+strconv.FormatInt(time.Now().Unix(), 10))
	if err != nil {
		logger.WithFields(l.ErrorField(err)).Error("rename failed")

		return
	}
}

func TryImportAlarmConfigs(root string, fileSuffix string, alarmManager timeassist.AlarmManager, logger l.Wrapper) {
	logger.Debug("TryImportAlarmConfigs root:", root)

	_ = filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, fileSuffix) {
			return nil
		}

		tryImportAlarmConfigs(path, alarmManager, logger)

		return nil
	})
}

func tryImportAlarmConfigs(file string, alarmManager timeassist.AlarmManager, logger l.Wrapper) {
	d, err := os.ReadFile(file)
	if err != nil && len(d) == 0 {
		return
	}

	var alarms []timeassist.Alarm

	err = yaml.Unmarshal(d, &alarms)
	if err != nil {
		logger.WithFields(l.ErrorField(err)).Error("invalid import config file format")

		return
	}

	if len(alarms) == 0 {
		return
	}

	for idx := 0; idx < len(alarms); idx++ {
		alarm := alarms[idx]
		if err = alarmManager.Add(&alarm); err != nil {
			logger.WithFields(l.ErrorField(err), l.StringField("id", alarm.ID)).Error("try import alarm failed")
		} else {
			logger.WithFields(l.StringField("id", alarm.ID)).Error("try import alarm successful")
		}
	}

	err = os.Rename(file, file+"."+strconv.FormatInt(time.Now().Unix(), 10))
	if err != nil {
		logger.WithFields(l.ErrorField(err)).Error("rename failed")

		return
	}
}
