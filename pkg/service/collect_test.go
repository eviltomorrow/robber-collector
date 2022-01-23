package service

import (
	"log"
	"testing"
	"time"

	"github.com/eviltomorrow/robber-core/pkg/mongodb"
)

func TestCollectMetadataToMongoDB(t *testing.T) {
	mongodb.DSN = "mongodb://127.0.0.1:27017"
	if err := mongodb.Build(); err != nil {
		t.Fatalf("build mongodb connection failure, nest error: %v", err)
	}

	var (
		now = time.Now()
	)

	date, fetchCount, err := CollectMetadataToMongoDB(now, "net126", false)
	if err != nil {
		log.Fatal(err)
	}
	t.Logf("date: %s", date)
	t.Logf("fetchCount: %d", fetchCount)
}
