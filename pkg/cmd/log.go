package cmd

import (
	"log"
	"time"

	"github.com/eviltomorrow/robber-collector/pkg/collector"
	"github.com/eviltomorrow/robber-collector/pkg/model"
	"github.com/eviltomorrow/robber-core/pkg/mongodb"
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Recover data from log",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			now = time.Now()
		)
		if logPath == "" {
			log.Fatalf("Invalid log-path")
		}
		setupCfg()
		setupVars()

		if err := mongodb.Build(); err != nil {
			log.Fatalf("Build mongodb connection failure, nest error: %v\r\n", err)
		}
		var total int64 = 0
		switch source {
		case "sina":
			var (
				source  = "sina"
				timeout = 10 * time.Second
			)
			p, err := collector.FetchMetadataFromLogForSina(logPath)
			if err != nil {
				log.Fatalf("Fetch metadata from log failure, nest error: %v\r\n", err)
			}

			var metadata = make([]*model.Metadata, 0, 64)
			for md := range p {
				_, err := model.DeleteMetadataByDate(mongodb.DB, source, md.Code, md.Date, timeout)
				if err != nil {
					log.Fatalf("DeleteMetadataByDate failure, nest error: %v, code: %v", err, md.Code)
				}
				if md.Volume != 0 && md.Open != 0 {
					metadata = append(metadata, md)
				}

				if len(metadata) >= 64 {
					affected, err := model.InsertMetadataMany(mongodb.DB, source, metadata, timeout)
					if err != nil {
						log.Fatalf("InsertMetadataMany failure, nest error: %v", err)
					}
					total += affected
					metadata = metadata[:0]
				}
			}
			if len(metadata) != 0 {
				affected, err := model.InsertMetadataMany(mongodb.DB, source, metadata, timeout)
				if err != nil {
					log.Fatalf("InsertMetadataMany failure, nest error: %v", err)
				}
				total += affected
			}
		// case "net126":
		default:
			log.Fatalf("Not support source[%v]", source)
		}
		log.Printf("[Complete] total count: %v, cost: %v\r\n", total, time.Since(now))
	},
}

var (
	logPath string
	source  string
)

func init() {
	logCmd.Flags().StringVarP(&cfgPath, "config", "c", "config.toml", "robber-collector's config file")
	logCmd.Flags().StringVarP(&logPath, "path", "p", "", "robber-collector's log recovery log file")
	logCmd.Flags().StringVarP(&source, "source", "s", "", "robber-collector's log recovery source")
	rootCmd.AddCommand(logCmd)
}
