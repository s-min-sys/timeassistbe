package autoimport

import (
	"testing"

	"github.com/s-min-sys/timeassistbe/internal/timeassist"
)

func Test1(t *testing.T) {
	t.SkipNow()

	TryImportTaskConfigs("..\\timeassistbe", timeassist.TaskIDPre, nil, nil)
}
