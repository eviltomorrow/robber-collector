package service

import (
	"fmt"
	"time"

	"github.com/eviltomorrow/robber-collector/pkg/collector"
	"github.com/eviltomorrow/robber-collector/pkg/model"
	"github.com/eviltomorrow/robber-core/pkg/mongodb"
	"github.com/eviltomorrow/robber-core/pkg/zmath"
)

var (
	FetchFactories = map[string]func([]string) ([]*model.Metadata, error){
		"sina":   collector.FetchMetadataFromSina,
		"net126": collector.FetchMetadataFromNet126,
	}
)

func CollectMetadataToMongoDB(today time.Time, source string, randSleep bool) (string, int64, error) {
	var (
		date = today.Format("2006-01-02")

		retrytimes = 0
		count      int64
		timeout    = 10 * time.Second
		size       = 30
		codes      = make([]string, 0, size)
	)

	fetchMetadata, ok := FetchFactories[source]
	if !ok {
		return "", 0, fmt.Errorf("not found fetch func, source = [%s]", source)
	}
	for code := range collector.GenRangeCode() {
		codes = append(codes, code)
		if len(codes) == size {
		retry_1:
			metadata, err := fetchMetadata(codes)
			if err != nil {
				retrytimes++
				if retrytimes == 10 {
					return date, count, fmt.Errorf("FetchMeatadata failure, nest error: %v, source: [%v], codes: %v", err, source, codes)
				} else {
					time.Sleep(30 * time.Second)
					goto retry_1
				}
			}
			retrytimes = 0
			codes = codes[:0]

			for _, md := range metadata {
				_, err := model.DeleteMetadataByDate(mongodb.DB, source, md.Code, md.Date, timeout)
				if err != nil {
					return date, count, fmt.Errorf("DeleteMetadataByDate failure, nest error: %v, code: %v", err, md.Code)
				}
			}
			affected, err := model.InsertMetadataMany(mongodb.DB, source, metadata, timeout)
			if err != nil {
				return date, count, fmt.Errorf("InsertMetadataMany failure, nest error: %v", err)
			}
			count += affected
			if randSleep {
				time.Sleep(time.Duration(zmath.GenRandInt(10, 30)) * time.Second)
			}
		}
	}

	if len(codes) != 0 {
	retry_2:
		metadata, err := fetchMetadata(codes)
		if err != nil {
			retrytimes++
			if retrytimes == 10 {
				return date, count, fmt.Errorf("FetchMeatadata failure, nest error: %v, source: [%v], codes: %v", err, source, codes)
			} else {
				time.Sleep(30 * time.Second)
				goto retry_2
			}
		}

		if len(metadata) != 0 {
			for _, md := range metadata {
				_, err := model.DeleteMetadataByDate(mongodb.DB, source, md.Code, md.Date, timeout)
				if err != nil {
					return date, count, fmt.Errorf("DeleteMetadataByDate failure, nest error: %v, code: %v", err, md.Code)
				}
			}
			affected, err := model.InsertMetadataMany(mongodb.DB, source, metadata, timeout)
			if err != nil {
				return date, count, fmt.Errorf("InsertMetadataMany failure, nest error: %v", err)
			}
			count += affected
		}
	}
	return date, count, nil
}
