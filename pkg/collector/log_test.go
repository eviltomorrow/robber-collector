package collector

import "testing"

func TestFetchMetadataFromLogForSina(t *testing.T) {
	p, err := FetchMetadataFromLogForSina("/home/shepard/tmp/data.log")
	if err != nil {
		t.Fatal(err)
	}
	for data := range p {
		t.Logf("%v\r\n", data)
	}
}
