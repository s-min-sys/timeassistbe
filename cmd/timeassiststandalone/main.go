package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/s-min-sys/timeassistbe/internal/autoimport"
	"github.com/s-min-sys/timeassistbe/internal/term"
	"github.com/s-min-sys/timeassistbe/internal/timeassist"
	"github.com/s-min-sys/timeassistbe/internal/ws"
	"github.com/sgostarter/i/l"
	"github.com/sgostarter/libeasygo/stg/kv"
	"go.uber.org/atomic"
)

func main() {
	logger := l.NewFileLoggerWrapper("task_log.txt")
	logger.GetLogger().SetLevel(l.LevelDebug)
	logger.Info("new time assist start at:", time.Now())

	activeTaskListChangeCh := make(chan interface{}, 10)

	timeMessages := term.NewTimeMessages()

	metaStorage, _ := kv.NewMemoryFileStorage("task_meta")
	timer := timeassist.NewTaskTimer("task_timer")
	taskTimer := timeassist.NewBizTimer(timer)

	taskList := timeassist.NewTaskList("task_list", func(task *timeassist.TaskInfo, visible bool) {
		pre := "remove"
		if visible {
			pre = "add"
		}

		message := fmt.Sprintf("%s task[%s] %s", pre, task.ID, task.Value)
		timeMessages.Add(message, time.Now().Add(time.Second*10))
		logger.Info(message)

		activeTaskListChangeCh <- true
	})

	taskManger := timeassist.NewTaskManager(metaStorage, taskTimer, taskList, logger)
	alarmManager := timeassist.NewAlarmManager(metaStorage, taskTimer, taskList, logger)

	timer.Start()

	autoimport.TryImportTaskConfigs("./import", "_task.yaml", taskManger, logger)
	autoimport.TryImportAlarmConfigs("./import", "_alarm.yaml", alarmManager, logger)

	var activeTaskListText atomic.String

	go func() {
		var lastActiveTaskList string

		for {
			time.Sleep(time.Second)

			tasks, err := taskList.GetList()
			if err != nil {
				logger.WithFields(l.ErrorField(err)).Error("get current list failed")

				continue
			}

			sort.Stable(timeassist.TaskInfos(tasks))

			ss := strings.Builder{}

			for _, task := range tasks {
				ss.WriteString(fmt.Sprintf("%s: %s [%s]\n", task.ID, task.Value, task.SubTitle))
			}

			if lastActiveTaskList != ss.String() {
				activeTaskListText.Store(ss.String())
				lastActiveTaskList = activeTaskListText.Load()

				activeTaskListChangeCh <- true
			}
		}
	}()

	consoleUI := term.NewConsoleIO()

	var enableTaskListUpdateAuto atomic.Bool

	enableTaskListUpdateAuto.Store(true)

	go func() {
		for {
			fnPrintUI := func() {
				if !enableTaskListUpdateAuto.Load() {
					return
				}

				term.CallClear()

				for _, message := range timeMessages.GetMessages() {
					consoleUI.Println("tip: " + message)
				}

				consoleUI.Println("-------------------------------")

				consoleUI.Println(activeTaskListText.Load())

				consoleUI.Println("=================================")
				consoleUI.Println(">>>> Type 'm' to manager")
			}

			fnPrintUI()

			<-activeTaskListChangeCh
			fnPrintUI()
		}
	}()

	go func() {
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

		s := &http.Server{
			Addr:        ":11110",
			Handler:     r,
			ReadTimeout: time.Second * 5,
		}

		_ = s.ListenAndServe()
	}()

	go func() {
		ws.NewWs(":11111", map[string]ws.Handler{
			"/ws/tasks": func(route string, conn *websocket.Conn) {
				var lastJSON string

				for {
					time.Sleep(time.Second)

					tasks, err := taskList.GetList()
					if err != nil {
						continue
					}

					sort.Stable(timeassist.TaskInfos(tasks))

					d, err := json.Marshal(tasks)
					if err != nil {
						continue
					}

					if lastJSON == string(d) {
						continue
					}

					lastJSON = string(d)

					err = conn.WriteMessage(websocket.TextMessage, d)
					if err != nil {
						break
					}
				}
			},
		}).Wait()
	}()

	for {
		s, _, _ := consoleUI.ReadString("")
		if s != "m" && s != "M" {
			enableTaskListUpdateAuto.Store(true)

			activeTaskListChangeCh <- true

			continue
		}

		enableTaskListUpdateAuto.Store(false)

	RETRY:
		consoleUI.Println(activeTaskListText.Load())
		consoleUI.Println("操作")
		consoleUI.Println("\t1: Done")
		consoleUI.Println("\t9: 取消")

		n, _, _ := consoleUI.ReadInt("请选择:")
		if n == 1 {
			id, _, _ := consoleUI.ReadString("请选择输入ID: ")

			var err error
			if timeassist.ParsePreOnID(id) == timeassist.TaskIDPre {
				err = taskManger.Done(id)
			} else {
				err = alarmManager.Done(id)
			}

			if err != nil {
				consoleUI.Println(fmt.Sprintf("task %s done failed: %s", id, err.Error()))
			} else {
				consoleUI.Println(fmt.Sprintf("task %s done!", id))
			}

			goto RETRY
		} else if n == 9 {
			enableTaskListUpdateAuto.Store(true)

			activeTaskListChangeCh <- true

			continue
		}
	}
}
