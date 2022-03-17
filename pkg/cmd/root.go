package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/eviltomorrow/robber-collector/internal/conf"
	"github.com/eviltomorrow/robber-collector/internal/server"
	"github.com/eviltomorrow/robber-collector/pkg/client"
	"github.com/eviltomorrow/robber-collector/pkg/scheduler"
	"github.com/eviltomorrow/robber-core/pkg/mongodb"
	"github.com/eviltomorrow/robber-core/pkg/pid"
	"github.com/eviltomorrow/robber-core/pkg/system"
	"github.com/eviltomorrow/robber-core/pkg/zlog"
	"github.com/eviltomorrow/robber-core/pkg/znet"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var rootCmd = &cobra.Command{
	Use:   "robber-collector",
	Short: "",
	Long:  "  \r\nrobber-collector server running",
	Run: func(cmd *cobra.Command, args []string) {
		if err := pid.CreatePidFile("robber-collecotr.pid"); err != nil {
			log.Fatalf("[Fatal] Robber-collector create pid file failure, nest error: %v\r\n", err)
		}

		if pprofMode {
			go func() {
				port, err := znet.GetFreePort()
				if err != nil {
					log.Fatalf("[Fatal] Get free port failure, nest error: %v\r\n", err)
				}

				log.Printf("[Debug] Debug mode is started, pprof service is listend on http://%s:%d/debug/pprof", system.IP, port)
				if err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
					log.Fatalf("[Fatal] ListenAndServe pprof server failure, nest error: %v\r\n", err)
				}
			}()
		}

		setupCfg()
		setupVars()

		if err := mongodb.Build(); err != nil {
			zlog.Fatal("Build mongodb connection failure", zap.Error(err))
		}

		go scheduler.Run()
		if err := server.StartupGRPC(); err != nil {
			zlog.Fatal("Startup GRPC service failure", zap.Error(err))
		}
		registerCleanFuncs()
		blockingUntilTermination()
	},
}

var (
	cleanFuncs []func() error
	cfg        = conf.Global
	cfgPath    = ""
	pprofMode  bool
)

func init() {
	rootCmd.CompletionOptions = cobra.CompletionOptions{
		DisableDefaultCmd: true,
	}
	rootCmd.Flags().StringVarP(&cfgPath, "config", "c", "config.toml", "robber-collector's config file")
	rootCmd.Flags().BoolVarP(&pprofMode, "pprof", "p", false, "robber-collector's pprof mode")
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func blockingUntilTermination() {
	var ch = make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)
	switch <-ch {
	case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
	case syscall.SIGUSR1:
	case syscall.SIGUSR2:
	default:
	}

	for _, f := range cleanFuncs {
		f()
	}
}

func registerCleanFuncs() {
	cleanFuncs = append(cleanFuncs, server.RevokeEtcdConn)
	cleanFuncs = append(cleanFuncs, server.ShutdownGRPC)
	cleanFuncs = append(cleanFuncs, pid.DestroyFile)
}

func setupCfg() {
	if err := cfg.FindAndLoad(cfgPath, nil); err != nil {
		log.Fatalf("[Fatal] FindAndLoad failure, nest error: %v\r\n", err)
	}
	global, prop, err := zlog.InitLogger(&zlog.Config{
		Level:            cfg.Log.Level,
		Format:           cfg.Log.Format,
		DisableTimestamp: cfg.Log.DisableTimestamp,
		File: zlog.FileLogConfig{
			Filename:   cfg.Log.FileName,
			MaxSize:    cfg.Log.MaxSize,
			MaxDays:    30,
			MaxBackups: 30,
		},
		DisableStacktrace:   true,
		DisableErrorVerbose: true,
	})
	if err != nil {
		log.Fatalf("[Fatal] Setup log config failure, nest error: %v\r\n", err)
	}
	zlog.ReplaceGlobals(global, prop)

	zlog.Info("Global Config info", zap.String("cfg", cfg.String()))
}

func setupVars() {
	server.Host = cfg.Server.Host
	server.Port = cfg.Server.Port
	server.Source = cfg.Collect.Source
	server.Endpoints = cfg.Etcd.Endpoints

	mongodb.DSN = cfg.MongoDB.DSN

	client.EtcdEndpoints = cfg.Etcd.Endpoints

	scheduler.Source = cfg.Collect.Source
}
