package collector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollectMetadataFromSina(t *testing.T) {
	_assert := assert.New(t)
	var codes = []string{
		"sh601012", "sz300002", "sz000001", "sh688887",
		"sh688887",
	}
	data, err := FetchMetadataFromSina(codes)
	_assert.Nil(err)
	t.Logf("data: %s\r\n", data)
}
