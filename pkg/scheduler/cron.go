package scheduler

import (
	"time"

	"github.com/eviltomorrow/robber-core/pkg/zlog"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

var DefaultCronSpec = "05 17 * * MON,TUE,WED,THU,FRI"

func Run() {
	var c = cron.New()
	_, err := c.AddFunc(DefaultCronSpec, func() {
		var (
			now = time.Now()
		)

		zlog.Info("Fetch metadata begin", zap.String("date", now.Format("2006-01-02")))

		zlog.Info("Fetch metadata complete", zap.String("date", ""), zap.Int64("fetch-count", 0), zap.Int64("push-count", 0), zap.Duration("cost", time.Since(now)))

	})
	if err != nil {
		zlog.Fatal("Cron add func failure", zap.Error(err))
	}
	c.Start()
}
