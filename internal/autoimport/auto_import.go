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

func TryImportConfigs(root string, taskManger timeassist.RecycleTaskManager, logger l.Wrapper) {
	logger.Debug("TryImportConfigs root:", root)

	_ = filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".yaml") {
			return nil
		}

		logger.Debug("TryImportConfigs =>:", path)
		tryImportConfigs(path, taskManger, logger)

		return nil
	})
}

func tryImportConfigs(file string, taskManger timeassist.RecycleTaskManager, logger l.Wrapper) {
	d, err := os.ReadFile(file)
	if err != nil && len(d) == 0 {
		return
	}

	var recycleTasks []timeassist.RecycleTask

	err = yaml.Unmarshal(d, &recycleTasks)
	if err != nil {
		logger.WithFields(l.ErrorField(err)).Error("invalid import config file format")

		return
	}

	if len(recycleTasks) == 0 {
		return
	}

	for _, task := range recycleTasks {
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
