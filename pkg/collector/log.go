package collector

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/eviltomorrow/robber-collector/pkg/model"
)

func FetchMetadataFromLogForNet126(path string) (chan *model.Metadata, error) {
	return nil, fmt.Errorf("not implement")
}

func FetchMetadataFromLogForSina(path string) (chan *model.Metadata, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	var p = make(chan *model.Metadata, 128)
	go func() {

		var (
			scanner      = bufio.NewScanner(f)
			total, count int
		)
		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			var text = scanner.Text()
			if strings.Contains(text, "Invalid trade data") && strings.Contains(text, "WARN") {
				total++
				var (
					begin, end int
					code       string
				)
				for i, c := range text {
					if c == '[' {
						begin = i
					}
					if c == ']' {
						end = i
						if begin+1 > end {
							continue
						}
						var data = text[begin+1 : end]
						if strings.HasPrefix(data, "code=") {
							code = strings.TrimPrefix(data, "code=")
						}
						if strings.HasPrefix(data, "data=") {
							var (
								m = strings.Index(data, "\\\"")
								n = strings.LastIndex(data, "\\\"")
							)
							if m == -1 || n == -1 || m+1 == n {
								continue
							}
							md, err := parseSinaLineToMetadata(code, fmt.Sprintf("\"%s\"", data[m+2:n]), map[string]int{
								"sh68": 33,
								"sz3":  34,
							})
							if err != nil {
								log.Printf("parse sina data failure, nest error: %v", err)
							} else {
								count++
								p <- md
							}
						}
					}
				}
			}
		}
		close(p)
		f.Close()
		log.Printf("From log, total: %v, count: %v\r\n", total, count)
	}()
	return p, nil
}
