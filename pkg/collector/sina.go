package collector

import (
	"bufio"
	"fmt"
	"strings"
	"time"

	"github.com/eviltomorrow/robber-collector/pkg/model"
	"github.com/eviltomorrow/robber-core/pkg/httpclient"
	"github.com/eviltomorrow/robber-core/pkg/zlog"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	suspendNormal    = "正常"
	suspendOneHour   = "停牌一小时"
	suspendOneDay    = "停牌一天"
	suspendKeep      = "连续停牌"
	suspendMid       = "盘中停牌"
	suspendHalfOfDay = "停牌半天"
	suspendPause     = "暂停"
	suspendNoRecord  = "无该记录"
	suspendUnlisted  = "未上市"
	suspendDelist    = "退市"
	suspendUnknown   = "未知"
)

var (
	ErrSinaInvalidFormat = errors.New("invalid data format")
)

func FetchMetadataFromSina(codes []string) ([]*model.Metadata, error) {
	var (
		url    = fmt.Sprintf("https://hq.sinajs.cn/list=%s", strings.Join(codes, ","))
		header = httpclient.DefaultHeader
	)
	header["Referer"] = "https://finance.sina.com.cn"

	data, err := httpclient.GetHTTP(url, 20*time.Second, header)
	if err != nil {
		return nil, fmt.Errorf("url: %v, nest error: %v", url, err)
	}

	var result = make([]*model.Metadata, 0, len(codes))
	kv, err := parseSinaDataToMap(data)
	if err != nil {
		zlog.Error("parseSinaDataToMap failure", zap.String("data", data), zap.Error(err))
	}
	for key, val := range kv {
		metadata, err := parseSinaLineToMetadata(key, val)
		if err != nil {
			zlog.Error("parseSinaLineToMetadata failure", zap.String("key", key), zap.String("val", val), zap.Error(err))
		}
		if metadata != nil {
			result = append(result, metadata)
		}
	}
	return result, nil
}

func parseSinaDataToMap(data string) (map[string]string, error) {
	var result = make(map[string]string)

	var scanner = bufio.NewScanner(strings.NewReader(data))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		var text = strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}

		if !strings.HasPrefix(text, "var") || !strings.HasSuffix(text, ";") {
			return nil, fmt.Errorf("invalid line data")
		}

		var n = strings.Index(text, "=")
		if n == -1 {
			return nil, fmt.Errorf("invalid line data")
		}

		var code = strings.Replace(text[:n], "var hq_str_", "", -1)
		result[code] = text
	}
	return result, nil
}

func parseSinaLineToMetadata(code, data string) (*model.Metadata, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var begin = strings.Index(strings.TrimSpace(data), `"`)
	var end = strings.LastIndex(strings.TrimSpace(data), `"`)

	if begin == -1 || end == -1 || begin == end {
		return nil, ErrSinaInvalidFormat
	}

	var attr = strings.Split(data[begin+1:end], ",")
	if len(attr) == 1 {
		return nil, ErrSinaInvalidFormat
	}
	if len(attr) >= 2 && attr[len(attr)-1] == "" {
		attr = attr[:len(attr)-1]
	}
	switch {
	case strings.HasPrefix(code, "sh68"):
		if len(attr) != 34 {
			zlog.Warn("Invalid trade data", zap.String("code", code), zap.String("data", data), zap.Int("len", len(attr)))
			return nil, ErrSinaInvalidFormat
		}
	case strings.HasPrefix(code, "sh60"):
		if len(attr) != 33 {
			zlog.Warn("Invalid trade data", zap.String("code", code), zap.String("data", data), zap.Int("len", len(attr)))
			return nil, ErrSinaInvalidFormat
		}
	case strings.HasPrefix(code, "sz0"):
		if len(attr) != 33 {
			zlog.Warn("Invalid trade data", zap.String("code", code), zap.String("data", data), zap.Int("len", len(attr)))
			return nil, ErrSinaInvalidFormat
		}
	case strings.HasPrefix(code, "sz3"):
		if len(attr) != 33 {
			zlog.Warn("Invalid trade data", zap.String("code", code), zap.String("data", data), zap.Int("len", len(attr)))
			return nil, ErrSinaInvalidFormat
		}
	default:
		return nil, fmt.Errorf("no support code[%v]", code)
	}

	var md = &model.Metadata{
		Code: code,
	}
	for i, val := range attr {
		switch i {
		case 0:
			md.Name = val
		case 1:
			md.Open = atof64(md.Name, i, val)
		case 2:
			md.YesterdayClosed = atof64(md.Name, i, val)
		case 3:
			md.Latest = atof64(md.Name, i, val)
		case 4:
			md.High = atof64(md.Name, i, val)
		case 5:
			md.Low = atof64(md.Name, i, val)
		case 8:
			md.Volume = atou64(md.Name, i, val)
		case 9:
			md.Account = atof64(md.Name, i, val)
		case 30:
			md.Date = val
		case 31:
			md.Time = val
		case 32:
			md.Suspend = getSuspendDesc(val)
		default:
		}
	}
	return md, nil
}

// getSuspendDesc get suspend desc
func getSuspendDesc(val string) string {
	switch {
	case val == "00":
		return suspendNormal
	case val == "01":
		return suspendOneHour
	case val == "02":
		return suspendOneDay
	case val == "03":
		return suspendKeep
	case val == "04":
		return suspendMid
	case val == "05":
		return suspendHalfOfDay
	case val == "07":
		return suspendPause
	case val == "-1":
		return suspendNoRecord
	case val == "-2":
		return suspendUnlisted
	case val == "-3":
		return suspendDelist
	default:
		return suspendUnknown
	}
}
