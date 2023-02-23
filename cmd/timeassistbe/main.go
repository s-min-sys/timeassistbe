package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/mux"
	"github.com/s-min-sys/timeassistbe/internal/autoimport"
	"github.com/s-min-sys/timeassistbe/internal/timeassist"
	"github.com/sgostarter/i/l"
	"github.com/sgostarter/libeasygo/stg/kv"
)

func main() {
	logger := l.NewFileLoggerWrapper("task_log.txt")
	logger.GetLogger().SetLevel(l.LevelDebug)
	logger.Info("new time assist start at:", time.Now())

	metaStorage, _ := kv.NewMemoryFileStorage("task_meta")
	timer := timeassist.NewTaskTimer("task_timer")
	taskTimer := timeassist.NewBizTimer(timer)

	taskList := timeassist.NewTaskList("task_list", func(task *timeassist.TaskInfo, visible bool) {})

	taskManger := timeassist.NewTaskManager(metaStorage, taskTimer, taskList, logger)
	alarmManager := timeassist.NewAlarmManager(metaStorage, taskTimer, taskList, logger)

	timer.Start()

	autoimport.TryImportTaskConfigs("./import", "_task.yaml", taskManger, logger)
	autoimport.TryImportAlarmConfigs("./import", "_alarm.yaml", alarmManager, logger)
	r := mux.NewRouter()

	r.HandleFunc("/tasks", func(writer http.ResponseWriter, request *http.Request) {
		tasks, err := taskList.GetList()
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte(err.Error()))

			return
		}

		if tasks == nil {
			tasks = make([]*timeassist.TaskInfo, 0)
		}

		sort.Stable(timeassist.TaskInfos(tasks))

		d, _ := json.Marshal(tasks)

		_, _ = writer.Write(d)
	})
	r.HandleFunc("/tasks/{task_id}/done", func(writer http.ResponseWriter, request *http.Request) {
		_ = taskList.Remove(mux.Vars(request)["task_id"])

		writer.WriteHeader(http.StatusOK)
	})
	r.HandleFunc("/tasks/{task_id}", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNotImplemented)
	}).Methods(http.MethodPost)

	s := &http.Server{
		Addr:        ":11110",
		Handler:     r,
		ReadTimeout: time.Second * 5,
	}

	_ = s.ListenAndServe()
}
