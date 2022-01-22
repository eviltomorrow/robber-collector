package conf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadCfg(t *testing.T) {
	_assert := assert.New(t)
	var (
		path = "../../tests/config-test.toml"
		cfg  = &Config{}
	)

	err := cfg.FindAndLoad(path, nil)
	_assert.Nil(err)

	_assert.Equal(false, cfg.Log.DisableTimestamp)
	_assert.Equal("info", cfg.Log.Level)
	_assert.Equal("text", cfg.Log.Format)
	_assert.Equal("../log/data.log", cfg.Log.FileName)
	_assert.Equal(20, cfg.Log.MaxSize)
	_assert.Equal("mongodb://127.0.0.1:27017", cfg.MongoDB.DSN)
	_assert.Equal("0.0.0.0", cfg.Server.Host)
	_assert.Equal(27320, cfg.Server.Port)
	_assert.Equal("tencent", cfg.Collect.Source)
	_assert.Equal(10, len(cfg.Collect.CodeList))
}
