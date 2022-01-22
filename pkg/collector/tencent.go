package collector

import (
	"fmt"
	"strings"
	"time"

	"github.com/eviltomorrow/robber-collector/pkg/model"
	"github.com/eviltomorrow/robber-core/pkg/httpclient"
)

func FetchMetadataFromTencent(codes []string) ([]*model.Metadata, error) {
	data, err := httpclient.GetHTTP(fmt.Sprintf("http://hq.sinajs.cn/list=%s", strings.Join(codes, ",")), 20*time.Second, httpclient.DefaultHeader)
	if err != nil {
		return nil, err
	}

	var result = make([]*model.Metadata, 0, len(codes))
	for key, val := range parseTencentDataToMap(data) {
		metadata, err := parseTencentLineToMetadata(key, val)
		if err != nil {
			_ = err
		}
		result = append(result, metadata)
	}
	return result, nil
}

func parseTencentDataToMap(data string) map[string]string {
	return nil
}

func parseTencentLineToMetadata(code, data string) (*model.Metadata, error) {
	return nil, nil
}
