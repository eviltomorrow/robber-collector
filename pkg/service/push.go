package service

import (
	"context"
	"fmt"
	"time"

	"github.com/eviltomorrow/robber-collector/pkg/model"
	"github.com/eviltomorrow/robber-core/pkg/mongodb"
	client_repository "github.com/eviltomorrow/robber-repository/pkg/client"
	pb_repository "github.com/eviltomorrow/robber-repository/pkg/pb"
)

func PushMetadataToRepository(source, date string) (int64, int64, int64, int64, error) {
	var (
		offset  int64 = 0
		limit   int64 = 100
		count   int64 = 0
		lastID  string
		timeout = 20 * time.Second
	)

	rstub, cancel, err := client_repository.NewClientForRepository()
	if err != nil {
		return 0, 0, 0, 0, err
	}
	defer cancel()

	req, err := rstub.PushData(context.Background())
	if err != nil {
		return 0, 0, 0, 0, err
	}

	for {
		data, err := model.SelectMetadataRange(mongodb.DB, offset, limit, source, date, lastID, timeout)
		if err != nil {
			return 0, 0, 0, 0, err
		}
		for _, d := range data {
			if d.Volume == 0 {
				continue
			}
			err := req.Send(&pb_repository.Metadata{
				Code:            d.Code,
				Name:            d.Name,
				Open:            d.Open,
				YesterdayClosed: d.YesterdayClosed,
				Latest:          d.Latest,
				High:            d.High,
				Low:             d.Low,
				Volume:          d.Volume,
				Account:         d.Account,
				Date:            d.Date,
				Time:            d.Time,
				Suspend:         d.Suspend,
			})
			if err != nil {
				_, e1 := req.CloseAndRecv()
				return 0, 0, 0, 0, fmt.Errorf("%v, nest error: %v", err, e1)
			}
			count++
		}

		if len(data) < int(limit) {
			break
		}
		if len(data) != 0 {
			lastID = data[len(data)-1].ObjectID
		}
		offset += limit
	}

	resp, err := req.CloseAndRecv()
	if err != nil {
		return 0, 0, 0, 0, err
	}
	return count, resp.Stock, resp.Day, resp.Week, err
}

func PushTaskInfoToRepository(date string, metadata, stock, day, week int64) error {
	rstub, cancel, err := client_repository.NewClientForRepository()
	if err != nil {
		return err
	}
	defer cancel()

	if _, err := rstub.Complete(context.Background(), &pb_repository.Task{
		Date:          date,
		MetadataCount: metadata,
		StockCount:    stock,
		DayCount:      day,
		WeekCount:     week,
	}); err != nil {
		return err
	}
	return nil
}
