package service

import (
	"testing"

	"github.com/eviltomorrow/robber-core/pkg/mongodb"
)

func TestPushMetadataToRepository(t *testing.T) {
	mongodb.DSN = "mongodb://127.0.0.1:27017"
	if err := mongodb.Build(); err != nil {
		t.Fatalf("build mongodb connection failure, nest error: %v", err)
	}

	total, stock, day, week, err := PushMetadataToRepository("net126", "2022-01-21")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("total: %v\r\n", total)
	t.Logf("stock: %v\r\n", stock)
	t.Logf("day: %v\r\n", day)
	t.Logf("week: %v\r\n", week)
}
