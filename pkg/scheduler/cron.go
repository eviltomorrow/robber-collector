package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/eviltomorrow/robber-collector/pkg/service"
	"github.com/eviltomorrow/robber-core/pkg/zlog"
	"github.com/eviltomorrow/robber-notification/pkg/client"
	"github.com/eviltomorrow/robber-notification/pkg/pb"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

var (
	DefaultCronSpec = "05 17 * * MON,TUE,WED,THU,FRI"
	Source          = "sina"
)

func Run() {
	var c = cron.New()
	_, err := c.AddFunc(DefaultCronSpec, func() {
		var (
			now = time.Now()
			ok  = false
		)
		defer func() {
			if !ok {
				client, close, err := client.NewClientForNotification()
				if err != nil {
					zlog.Error("NewClientForNotification failure", zap.Error(err))
					return
				}
				defer close()

				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				defer cancel()

				if _, err := client.SendEmail(ctx, &pb.Mail{
					To: []*pb.Contact{
						{Name: "shepard", Address: "eviltomorrow@163.com"},
					},
					Subject: "[Warnning] Robber-collector panic notify",
					Body:    fmt.Sprintf("<h4>Fetch metadata failure, please check in log, date: %v</h4>", time.Now().Format("2006-01-02 15:04:05")),
				}); err != nil {
					zlog.Error("SendEmail failure", zap.Error(err))
				}
			}
		}()

		zlog.Info("Fetch metadata begin", zap.String("date", now.Format("2006-01-02")))
		date, fetchCount, err := service.CollectMetadataToMongoDB(now, Source, true)
		if err != nil {
			zlog.Error("CollectMetadataToMongoDB failure", zap.String("date", date), zap.Error(err))
			return
		}

		total, stock, day, week, err := service.PushMetadataToRepository(Source, date)
		if err != nil {
			zlog.Error("PushMetadataToRepository failure", zap.String("date", date), zap.Error(err))
			return
		}

		if err := service.PushTaskInfoToRepository(date, total, stock, day, week); err != nil {
			zlog.Error("PushTaskInfoToRepository failure", zap.String("date", date), zap.Error(err))
			return
		}
		ok = true
		zlog.Info("Fetch metadata complete", zap.String("date", date), zap.Int64("fetch-count", fetchCount), zap.Int64("push-count", total), zap.Duration("cost", time.Since(now)))

	})
	if err != nil {
		zlog.Fatal("Cron add func failure", zap.Error(err))
	}
	c.Start()
}
