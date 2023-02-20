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
)

func main() {
	logger := l.NewFileLoggerWrapper("task_log.txt")
	logger.GetLogger().SetLevel(l.LevelDebug)
	logger.Info("new time assist start at:", time.Now())

	taskStorage := timeassist.NewRecycleTaskStorage("task_meta")
	taskTimer := timeassist.NewRecycleTaskTimer("task_timer")
	taskList := timeassist.NewRecycleTaskList("task_list", func(task *timeassist.TaskInfo, visible bool) {})
	taskManger := timeassist.NewRecycleTaskManager(taskStorage, taskTimer, taskList, logger)

	autoimport.TryImportConfigs("./import", taskManger, logger)

	r := mux.NewRouter()

	r.HandleFunc("/tasks", func(writer http.ResponseWriter, request *http.Request) {
		tasks, err := taskManger.GetCurrentList()
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
