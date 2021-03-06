package conf

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Log     Log     `json:"log" toml:"log"`
	MongoDB MongoDB `json:"mongodb" toml:"mongodb"`
	Etcd    Etcd    `json:"etcd" toml:"etcd"`
	Server  Server  `json:"server" toml:"server"`
	Collect Collect `json:"collect" toml:"collect"`
}

type Log struct {
	DisableTimestamp bool   `json:"disable-timestamp" toml:"disable-timestamp"`
	Level            string `json:"level" toml:"level"`
	Format           string `json:"format" toml:"format"`
	FileName         string `json:"filename" toml:"filename"`
	MaxSize          int    `json:"maxsize" toml:"maxsize"`
}

type MongoDB struct {
	DSN string `json:"dsn" toml:"dsn"`
}

type Etcd struct {
	Endpoints []string `json:"endpoints" toml:"endpoints"`
}

type Server struct {
	Host string `json:"host" toml:"host"`
	Port int    `json:"port" toml:"port"`
}

type Collect struct {
	Source   string   `json:"source" toml:"source"`
	CodeList []string `json:"code-list" toml:"code-list"`
}

func (c *Config) FindAndLoad(path string, override func(cfg *Config)) error {
	findPath := func() (string, error) {
		var possibleConf = []string{
			path,
			"./etc/config.toml",
			"../etc/config.toml",
		}
		for _, path := range possibleConf {
			if path == "" {
				continue
			}
			if _, err := os.Stat(path); err == nil {
				fp, err := filepath.Abs(path)
				if err == nil {
					return fp, nil
				}
				return path, nil
			}
		}
		return "", fmt.Errorf("not found conf file, possible conf %v", possibleConf)
	}
	conf, err := findPath()
	if err != nil {
		return err
	}

	if _, err := toml.DecodeFile(conf, c); err != nil {
		return err
	}
	return nil
}

func (cg *Config) String() string {
	buf, _ := json.Marshal(cg)
	return string(buf)
}

var Global = &Config{
	Log: Log{
		DisableTimestamp: false,
		Level:            "info",
		Format:           "text",
		FileName:         "/tmp/robber-collector/data.log",
		MaxSize:          20,
	},
	MongoDB: MongoDB{
		DSN: "mongodb://127.0.0.1:27017",
	},
	Etcd: Etcd{
		Endpoints: []string{
			"127.0.0.1:2379",
		},
	},
	Server: Server{
		Host: "0.0.0.0",
		Port: 27320,
	},
	Collect: Collect{
		Source: "sina",
		CodeList: []string{
			"sh688***",
			"sh605***",
			"sh603***",
			"sh601***",
			"sh600***",
			"sz300***",
			"sz0030**",
			"sz002***",
			"sz001**",
			"sz000***",
		},
	},
}
