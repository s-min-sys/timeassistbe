package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/s-min-sys/timeassistbe/internal/autoimport"
	"github.com/s-min-sys/timeassistbe/internal/term"
	"github.com/s-min-sys/timeassistbe/internal/timeassist"
	"github.com/s-min-sys/timeassistbe/internal/ws"
	"github.com/sgostarter/i/l"
	"go.uber.org/atomic"
)

func main() {
	logger := l.NewFileLoggerWrapper("task_log.txt")
	logger.GetLogger().SetLevel(l.LevelDebug)
	logger.Info("new time assist start at:", time.Now())

	activeTaskListChangeCh := make(chan interface{}, 10)

	timeMessages := term.NewTimeMessages()

	taskStorage := timeassist.NewRecycleTaskStorage("task_meta")
	taskTimer := timeassist.NewRecycleTaskTimer("task_timer")
	taskList := timeassist.NewRecycleTaskList("task_list", func(task *timeassist.TaskInfo, visible bool) {
		pre := "remove"
		if visible {
			pre = "add"
		}

		message := fmt.Sprintf("%s task[%s] %s", pre, task.ID, task.Value)
		timeMessages.Add(message, time.Now().Add(time.Second*10))
		logger.Info(message)

		activeTaskListChangeCh <- true
	})
	taskManger := timeassist.NewRecycleTaskManager(taskStorage, taskTimer, taskList, logger)

	autoimport.TryImportConfigs("./import", taskManger, logger)

	var activeTaskListText atomic.String

	go func() {
		var lastActiveTaskList string

		for {
			time.Sleep(time.Second)

			tasks, err := taskManger.GetCurrentList()
			if err != nil {
				logger.WithFields(l.ErrorField(err)).Error("get current list failed")
				continue
			}

			sort.Stable(timeassist.TaskInfos(tasks))

			ss := strings.Builder{}

			for _, task := range tasks {
				ss.WriteString(fmt.Sprintf("%s: %s\n", task.ID, task.Value))
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

			select {
			case <-activeTaskListChangeCh:
				fnPrintUI()
			}
		}
	}()

	go func() {
		r := mux.NewRouter()

		fakeTaskIdx := 0
		r.HandleFunc("/fake/tasks", func(writer http.ResponseWriter, request *http.Request) {
			var tasks []*timeassist.TaskInfo

			tasks = append(tasks, &timeassist.TaskInfo{
				ID:    strconv.Itoa(fakeTaskIdx),
				Value: fmt.Sprintf("TITLE-%d", fakeTaskIdx),
			})

			fakeTaskIdx++

			tasks = append(tasks, &timeassist.TaskInfo{
				ID:    strconv.Itoa(fakeTaskIdx),
				Value: fmt.Sprintf("TITLE-%d", fakeTaskIdx),
			})

			fakeTaskIdx++

			d, _ := json.Marshal(tasks)

			_, _ = writer.Write(d)
		})
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
				var lastJson string

				for {
					time.Sleep(time.Second)

					tasks, err := taskManger.GetCurrentList()
					if err != nil {
						continue
					}

					sort.Stable(timeassist.TaskInfos(tasks))

					d, err := json.Marshal(tasks)
					if err != nil {
						continue
					}

					if lastJson == string(d) {
						continue
					}

					lastJson = string(d)

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
			if err := taskManger.Done(id); err != nil {
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
