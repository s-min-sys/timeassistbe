package trace

import (
	"os"
	"path"
	"sync"
	"time"

	"github.com/sgostarter/libeasygo/pathutils"
)

var (
	_traceOnce sync.Once
	_trace     TaskTrace
)

type TaskTrace interface {
	RecordMessage(taskID string, message string)
	RecordTimeSchedule(taskID string, at time.Time)
	RecordRemoveTimeSchedule(taskID string)
}

func Get() TaskTrace {
	_traceOnce.Do(func() {
		_trace = newTaskTrace("trace")
	})

	return _trace
}

func newTaskTrace(root string) *taskTraceImpl {
	return &taskTraceImpl{
		root:  root,
		files: make(map[string]*os.File),
	}
}

type taskTraceImpl struct {
	root      string
	filesLock sync.Mutex
	files     map[string]*os.File
}

func (tt *taskTraceImpl) get(taskID string) *os.File {
	tt.filesLock.Lock()
	defer tt.filesLock.Unlock()

	if tt.files[taskID] != nil {
		return tt.files[taskID]
	}

	_ = pathutils.MustDirExists(tt.root)

	f, err := os.OpenFile(path.Join(tt.root, taskID), os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		panic("file:" + path.Join(tt.root, taskID))

		return nil
	}

	tt.files[taskID] = f

	return f
}

func (tt *taskTraceImpl) writeString(taskID string, message string) {
	_, _ = tt.get(taskID).WriteString(time.Now().Format("2006/01/02 15:04:05 =>") + message + "\n")
}

func (tt *taskTraceImpl) RecordMessage(taskID string, message string) {
	tt.writeString(taskID, message)
}

func (tt *taskTraceImpl) RecordTimeSchedule(taskID string, at time.Time) {
	tt.writeString(taskID, "Set Schedule:"+at.String())
}

func (tt *taskTraceImpl) RecordRemoveTimeSchedule(taskID string) {
	_, _ = tt.get(taskID).WriteString("Remove Schedule")
}
