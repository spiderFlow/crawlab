package schedule

import (
	"github.com/crawlab-team/crawlab/core/interfaces"
	"github.com/crawlab-team/crawlab/core/utils"
	"github.com/robfig/cron/v3"
	"strings"
)

type CronLogger struct {
	interfaces.Logger
}

func (l *CronLogger) Info(msg string, keysAndValues ...interface{}) {
	p := l.getPlaceholder(len(keysAndValues))
	l.Infof("cron: %s %s", msg, p)
}

func (l *CronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	p := l.getPlaceholder(len(keysAndValues))
	l.Errorf("cron: %s %v %s", msg, err, p)
}

func (l *CronLogger) getPlaceholder(n int) (s string) {
	var arr []string
	for i := 0; i < n; i++ {
		arr = append(arr, "%v")
	}
	return strings.Join(arr, " ")
}

func NewCronLogger() cron.Logger {
	return &CronLogger{
		Logger: utils.NewLogger("Cron"),
	}
}
