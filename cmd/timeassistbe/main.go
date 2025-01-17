package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/s-min-sys/notifier-share/pkg"
	"github.com/s-min-sys/timeassistbe/internal/autoimport"
	"github.com/s-min-sys/timeassistbe/internal/timeassist"
	"github.com/s-min-sys/timeassistbe/internal/utils"
	"github.com/sgostarter/i/l"
	"github.com/sgostarter/libconfig"
	"github.com/sgostarter/libeasygo/pathutils"
	"github.com/sgostarter/libeasygo/ptl"
	"github.com/sgostarter/libeasygo/stg/kv"
	"golang.org/x/exp/slices"
)

const (
	dataRoot = "data"
)

type Config struct {
	Listens   string `yaml:"Listens"`
	NotifyURL string `yaml:"NotifyURL"`
}

func main() {
	var cfg Config

	_, err := libconfig.Load("config.yaml", &cfg)
	if err != nil {
		panic(err)
	}

	_ = pathutils.MustDirExists(dataRoot)

	logger := l.NewConsoleLoggerWrapper()
	//logger := l.NewFileLoggerWrapper(filepath.Join(dataRoot, "task_log.txt"))
	logger.GetLogger().SetLevel(l.LevelDebug)
	logger.Info("new time assist start at:", time.Now())

	metaStorage, _ := kv.NewMemoryFileStorageEx(filepath.Join(dataRoot, "task_meta"), false)
	timer := timeassist.NewTaskTimer(filepath.Join(dataRoot, "task_timer"))
	taskTimer := timeassist.NewBizTimer(timer)

	showList := timeassist.NewShowList(filepath.Join(dataRoot, "task_list"), func(task *timeassist.ShowInfo, visible bool) {
		if !visible {
			return
		}

		notifyAlarm(logger, cfg.NotifyURL, task)
	})

	taskManger := timeassist.NewTaskManager(metaStorage, taskTimer, showList, logger)
	alarmManager := timeassist.NewAlarmManager(metaStorage, taskTimer, showList, logger)

	timer.Start()

	autoimport.TryImportTaskConfigs("./import", "_task.yaml", taskManger, logger)
	autoimport.TryImportAlarmConfigs("./import", "_alarm.yaml", alarmManager, logger)

	r := mux.NewRouter()

	r.HandleFunc("/alarms/add", func(writer http.ResponseWriter, request *http.Request) {
		var alarms []timeassist.Alarm

		err = json.NewDecoder(request.Body).Decode(&alarms)
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
	}).Methods(http.MethodPost)

	r.HandleFunc("/alarm/add", func(writer http.ResponseWriter, request *http.Request) {
		var respWrapper ResponseWrapper

		respWrapper.Apply(handleAddAlarm(request, alarmManager))

		httpResp(&respWrapper, writer)
	}).Methods(http.MethodPost)

	r.HandleFunc("/alarm/remove", func(writer http.ResponseWriter, request *http.Request) {
		var respWrapper ResponseWrapper

		respWrapper.Apply(handleRemoveAlarm(request, alarmManager))

		httpResp(&respWrapper, writer)
	}).Methods(http.MethodPost)

	r.HandleFunc("/alarms/detail", func(writer http.ResponseWriter, request *http.Request) {
		var respWrapper ResponseWrapper

		items, code, msg := handleGetAlarms(request, timer, metaStorage)
		if respWrapper.Apply(code, msg) {
			respWrapper.Resp = items
		}

		httpResp(&respWrapper, writer)
	}).Methods(http.MethodGet)

	//
	r.HandleFunc("/task/add", func(writer http.ResponseWriter, request *http.Request) {
		var respWrapper ResponseWrapper

		respWrapper.Apply(handleAddTask(request, taskManger))

		httpResp(&respWrapper, writer)
	}).Methods(http.MethodPost)

	r.HandleFunc("/task/detail", func(writer http.ResponseWriter, request *http.Request) {
		var respWrapper ResponseWrapper

		items, code, msg := handleGetRTasks(request, timer, metaStorage)
		if respWrapper.Apply(code, msg) {
			respWrapper.Resp = items
		}

		httpResp(&respWrapper, writer)
	}).Methods(http.MethodGet)

	r.HandleFunc("/shows", func(writer http.ResponseWriter, _ *http.Request) {
		var respWrapper ResponseWrapper

		tasks, code, msg := handleGetTasks(showList)
		if respWrapper.Apply(code, msg) {
			respWrapper.Resp = tasks
		}

		httpResp(&respWrapper, writer)
	})

	r.HandleFunc("/shows/{task_id}/done", func(writer http.ResponseWriter, request *http.Request) {
		var respWrapper ResponseWrapper

		respWrapper.Apply(handleTaskDone(request, showList, taskManger))

		httpResp(&respWrapper, writer)
	})

	doNotify(logger, cfg.NotifyURL, "time assist be started")

	fnListen := func(listen string) {
		srv := &http.Server{
			Addr:        listen,
			ReadTimeout: time.Second,
			Handler:     r,
		}

		logger.WithFields(l.StringField("listen", listen)).Debug("start listen")

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.WithFields(l.ErrorField(err), l.StringField("listen", listen)).Error("listen failed")
		}
	}

	listens := strings.Split(cfg.Listens, " ")

	for idx := 0; idx < len(listens)-1; idx++ {
		go fnListen(listens[idx])
	}

	fnListen(listens[len(listens)-1])
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

func handleGetTasks(taskList timeassist.ShowList) (
	tasks []*timeassist.ShowInfo, code Code, msg string) {
	tasks, err := taskList.GetList()
	if err != nil {
		code = CodeErrInternal
		msg = err.Error()

		return
	}

	if tasks == nil {
		tasks = make([]*timeassist.ShowInfo, 0)
	}

	slices.SortFunc(tasks, func(a, b *timeassist.ShowInfo) int {
		if a.VOTaskType != b.VOTaskType {
			if a.VOTaskType == timeassist.VOTaskTypeTask {
				return 1
			}

			return -1
		}

		if a.VOTaskType == timeassist.VOTaskTypeTask {
			return strings.Compare(a.Value, b.Value)
		}

		return a.AlarmAt.Compare(b.AlarmAt)
	})

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

func handleRemoveAlarm(request *http.Request, alarmManager timeassist.AlarmManager) (code Code, msg string) {
	id := request.URL.Query().Get("id")
	if id == "" {
		code = CodeErrBadRequest

		return
	}

	err := alarmManager.Remove(id)
	if err != nil {
		code = CodeErrInternal
		msg = err.Error()

		return
	}

	code = CodeSuccess

	return
}

type AlarmItem struct {
	ID       string `json:"id"`
	CheckAt  int64  `json:"check_at"`
	CheckAtS string `json:"check_at_s"`

	ShowAt    int64  `json:"show_at"`
	ShowAtS   string `json:"show_at_s"`
	ExpireAt  int64  `json:"expire_at"`
	ExpireAtS string `json:"expire_at_s"`

	Text     string `json:"text"`
	Value    string `json:"value"`
	AValue   string `json:"a_value"`
	LeftTime string `json:"left_time"`
}

func handleGetAlarms(_ *http.Request, t timeassist.TaskTimer, storage kv.StorageTiny) (aItems []AlarmItem, code Code, msg string) {
	items, err := t.List()
	if err != nil {
		code = CodeErrInternal
		msg = err.Error()

		return
	}

	aItems = make([]AlarmItem, 0, len(items))

	fnFormatTime := func(t time.Time) string {
		return t.Format("2006-01-02 15:04:05")
	}

	fnFormatTimeStamp := func(tm int64) string {
		return fnFormatTime(time.Unix(tm, 0))
	}

	for _, d := range items {
		idPre := timeassist.ParsePreOnID(d.Data.ID)
		if idPre != timeassist.AlarmIDPre {
			continue
		}

		alarm := &timeassist.Alarm{}

		ok, e := storage.Get(d.Data.ID, alarm)
		if e != nil || !ok {
			continue
		}

		av, e := timeassist.ParseAlarmValue(alarm.Value, alarm.AType)
		if e != nil {
			continue
		}

		_, aValue := av.StringNoNowTime(alarm.AType)
		aItems = append(aItems, AlarmItem{
			ID:        d.Data.ID,
			CheckAt:   d.At.Unix(),
			CheckAtS:  fnFormatTime(d.At),
			ShowAt:    d.Data.StartUTC,
			ShowAtS:   fnFormatTimeStamp(d.Data.StartUTC),
			ExpireAt:  d.Data.EndUTC,
			ExpireAtS: fnFormatTimeStamp(d.Data.EndUTC),
			Text:      alarm.Text,
			Value:     alarm.Value,
			AValue:    aValue,
			LeftTime:  utils.LeftTimeString(time.Unix(d.Data.EndUTC, 0)),
		})
	}

	slices.SortFunc(aItems, func(a, b AlarmItem) int {
		if a.ExpireAt < b.ExpireAt {
			return -1
		}

		if a.ExpireAt > b.ExpireAt {
			return 1
		}

		return 0
	})

	code = CodeSuccess

	return
}

func handleAddTask(request *http.Request, taskManager timeassist.TaskManager) (code Code, msg string) {
	var task timeassist.Task

	err := json.NewDecoder(request.Body).Decode(&task)
	if err != nil {
		code = CodeErrParse
		msg = err.Error()

		return
	}

	err = taskManager.Add(&task)
	if err != nil {
		code = CodeErrParse
		msg = err.Error()

		return
	}

	code = CodeSuccess

	return
}

func handleTaskDone(request *http.Request, taskList timeassist.ShowList, taskManager timeassist.TaskManager) (code Code, msg string) {
	taskID := mux.Vars(request)["task_id"]

	if timeassist.ParsePreOnID(taskID) == timeassist.TaskIDPre {
		taskManager.TaskDone(taskID)
	}

	_ = taskList.Remove(taskID)

	code = CodeSuccess

	return
}

func handleGetRTasks(_ *http.Request, t timeassist.TaskTimer, storage kv.StorageTiny) (aItems []AlarmItem, code Code, msg string) {
	items, err := t.List()
	if err != nil {
		code = CodeErrInternal
		msg = err.Error()

		return
	}

	aItems = make([]AlarmItem, 0, len(items))

	fnFormatTime := func(t time.Time) string {
		return t.Format("2006-01-02 15:04:05")
	}

	fnFormatTimeStamp := func(tm int64) string {
		return fnFormatTime(time.Unix(tm, 0))
	}

	for _, d := range items {
		idPre := timeassist.ParsePreOnID(d.Data.ID)
		if idPre != timeassist.TaskIDPre {
			continue
		}

		task := &timeassist.Task{}

		ok, e := storage.Get(d.Data.ID, task)
		if e != nil || !ok {
			continue
		}

		aItems = append(aItems, AlarmItem{
			ID:        d.Data.ID,
			CheckAt:   d.At.Unix(),
			CheckAtS:  fnFormatTime(d.At),
			ShowAt:    d.Data.StartUTC,
			ShowAtS:   fnFormatTimeStamp(d.Data.StartUTC),
			ExpireAt:  d.Data.EndUTC,
			ExpireAtS: fnFormatTimeStamp(d.Data.EndUTC),
			Text:      task.Text,
			Value:     "",
			AValue:    task.Desc(),
			LeftTime:  utils.LeftTimeString(time.Unix(d.Data.EndUTC, 0)),
		})
	}

	slices.SortFunc(aItems, func(a, b AlarmItem) int {
		if a.ExpireAt < b.ExpireAt {
			return -1
		}

		if a.ExpireAt > b.ExpireAt {
			return 1
		}

		return 0
	})

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

func notifyAlarm(logger l.Wrapper, notifyURL string, task *timeassist.ShowInfo) {
	var text string

	if task.VOTaskType == timeassist.VOTaskTypeAlarm {
		text = fmt.Sprintf("闹钟: %s %s - %s", task.Value, task.SubTitle, task.AlarmAt.Format("2006-01-02 15:04:05"))
		if task.AlarmFlag {
			text += " 已经过期"
		}
	} else {
		text = fmt.Sprintf("任务: %s %s", task.Value, task.SubTitle)
		if task.AlarmFlag {
			text += " 已经过期"
		}
	}

	doNotify(logger, notifyURL, text)
}

func doNotify(logger l.Wrapper, notifyURL, text string) {
	go func() {
		fnSend := func(senderID pkg.SenderID, receiverType pkg.ReceiverType, text string) {
			code, errMsg := pkg.SendTextMessage(notifyURL, &pkg.TextMessage{
				SenderID:     senderID,
				ReceiverType: receiverType,
				Text:         text,
			})
			if code != ptl.CodeSuccess {
				logger.WithFields(l.StringField("errMsg", errMsg), l.StringField("senderID", string(senderID)),
					l.IntField("receiverType", int(receiverType)), l.StringField("text", text)).
					Info("send failed")
			}
		}

		fnSend(pkg.SenderIDTelegram, pkg.ReceiverTypeAdminUsers, text)
		fnSend(pkg.SenderIDTelegram, pkg.ReceiverTypeAdminGroups, text)

		fnSend(pkg.SenderIDWeChat, pkg.ReceiverTypeAdminUsers, text)
		fnSend(pkg.SenderIDWeChat, pkg.ReceiverTypeAdminGroups, text)
	}()
}
