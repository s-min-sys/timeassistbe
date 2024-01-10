package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"sort"
	"time"

	"github.com/gorilla/mux"
	"github.com/s-min-sys/timeassistbe/internal/autoimport"
	"github.com/s-min-sys/timeassistbe/internal/timeassist"
	"github.com/sgostarter/i/l"
	"github.com/sgostarter/libeasygo/pathutils"
	"github.com/sgostarter/libeasygo/stg/kv"
)

const (
	dataRoot = "data"
)

func main() {
	_ = pathutils.MustDirExists(dataRoot)

	logger := l.NewFileLoggerWrapper(filepath.Join(dataRoot, "task_log.txt"))
	logger.GetLogger().SetLevel(l.LevelDebug)
	logger.Info("new time assist start at:", time.Now())

	metaStorage, _ := kv.NewMemoryFileStorage(filepath.Join(dataRoot, "task_meta"))
	timer := timeassist.NewTaskTimer(filepath.Join(dataRoot, "task_timer"))
	taskTimer := timeassist.NewBizTimer(timer)

	taskList := timeassist.NewTaskList(filepath.Join(dataRoot, "task_list"), func(task *timeassist.TaskInfo, visible bool) {})

	taskManger := timeassist.NewTaskManager(metaStorage, taskTimer, taskList, logger)
	alarmManager := timeassist.NewAlarmManager(metaStorage, taskTimer, taskList, logger)

	timer.Start()

	autoimport.TryImportTaskConfigs("./import", "_task.yaml", taskManger, logger)
	autoimport.TryImportAlarmConfigs("./import", "_alarm.yaml", alarmManager, logger)

	r := mux.NewRouter()

	r.HandleFunc("/add/alarms", func(writer http.ResponseWriter, request *http.Request) {
		var alarms []timeassist.Alarm

		err := json.NewDecoder(request.Body).Decode(&alarms)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)

			return
		}

		var errMsg string
		var successCount, failedCount int

		for idx := 0; idx < len(alarms); idx++ {
			alarm := alarms[idx]

			err = alarmManager.Add(&alarm)
			if err != nil {
				errMsg += err.Error() + "\n"
				failedCount++
			} else {
				successCount++
			}
		}

		if successCount == 0 && failedCount > 0 {
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte(errMsg))
		} else {
			writer.WriteHeader(http.StatusOK)
		}
	}).Methods("POST")

	r.HandleFunc("/add/alarm", func(writer http.ResponseWriter, request *http.Request) {
		var respWrapper ResponseWrapper

		respWrapper.Apply(handleAddAlarm(request, alarmManager))

		httpResp(&respWrapper, writer)
	}).Methods("POST")

	r.HandleFunc("/tasks", func(writer http.ResponseWriter, request *http.Request) {
		var respWrapper ResponseWrapper

		tasks, code, msg := handleGetTasks(taskList)
		if respWrapper.Apply(code, msg) {
			respWrapper.Resp = tasks
		}

		httpResp(&respWrapper, writer)
	})
	r.HandleFunc("/tasks/{task_id}/done", func(writer http.ResponseWriter, request *http.Request) {
		_ = taskList.Remove(mux.Vars(request)["task_id"])

		var respWrapper ResponseWrapper

		respWrapper.Apply(CodeSuccess, "")

		httpResp(&respWrapper, writer)
	})
	r.HandleFunc("/tasks/{task_id}", func(writer http.ResponseWriter, request *http.Request) {
		var respWrapper ResponseWrapper

		respWrapper.Apply(CodeErrsNotImplemented, "")

		httpResp(&respWrapper, writer)
	}).Methods(http.MethodPost)

	s := &http.Server{
		Addr:        ":11110",
		Handler:     r,
		ReadTimeout: time.Second * 5,
	}

	_ = s.ListenAndServe()
}

//
//
//

func httpResp(respWrapper *ResponseWrapper, writer http.ResponseWriter) {
	writer.Header().Add("Content-Type", "application/json")

	writer.WriteHeader(http.StatusOK)

	d, _ := json.Marshal(respWrapper)

	_, _ = writer.Write(d)
}

func handleGetTasks(taskList timeassist.TaskList) (
	tasks []*timeassist.TaskInfo, code Code, msg string) {
	tasks, err := taskList.GetList()
	if err != nil {
		code = CodeErrInternal
		msg = err.Error()

		return
	}

	if tasks == nil {
		tasks = make([]*timeassist.TaskInfo, 0)
	}

	sort.Stable(timeassist.TaskInfos(tasks))

	return
}

func handleAddAlarm(request *http.Request, alarmManager timeassist.AlarmManager) (code Code, msg string) {
	var alarm timeassist.Alarm

	err := json.NewDecoder(request.Body).Decode(&alarm)
	if err != nil {
		code = CodeErrParse
		msg = err.Error()

		return
	}

	err = alarmManager.Add(&alarm)
	if err != nil {
		code = CodeErrParse
		msg = err.Error()

		return
	}

	code = CodeSuccess

	return
}

//
//
//

type Code int

const (
	CodeSuccess Code = iota
	CodeErrUnauthenticated
	CodeErrBadRequest
	CodeErrAuth
	CodeErrParse
	CodeErrInternal
	CodeErrPermission
	CodeErrNotFound
	CodeErrBanned
	CodeErrDisabled
	CodeErrUserExists
	CodeErrsNotImplemented
)

func (c Code) String() string {
	switch c {
	case CodeSuccess:
		return "成功"
	case CodeErrUnauthenticated:
		return "需要授权"
	case CodeErrBadRequest:
		return "缺少参数"
	case CodeErrAuth:
		return "非法凭证"
	case CodeErrParse:
		return "传输错误"
	case CodeErrInternal:
		return "服务器内部错误"
	case CodeErrPermission:
		return "没有对应权限"
	case CodeErrNotFound:
		return "指定对象不存在"
	case CodeErrBanned:
		return "用户被限制"
	case CodeErrDisabled:
		return "操作被禁止"
	case CodeErrUserExists:
		return "用户已经存在"
	case CodeErrsNotImplemented:
		return "未实现"
	}

	return fmt.Sprintf("未知错误%d", c)
}

func CodeToMessage(code Code, msg string) string {
	codeMsg := code.String()

	if msg != "" {
		codeMsg += ":" + msg
	}

	return codeMsg
}

type ResponseWrapper struct {
	Code    Code        `json:"code"`
	Message string      `json:"message"`
	Resp    interface{} `json:"resp,omitempty"`
}

func (wr *ResponseWrapper) Apply(code Code, msg string) bool {
	wr.Code = code
	wr.Message = CodeToMessage(code, msg)

	return code == CodeSuccess
}
