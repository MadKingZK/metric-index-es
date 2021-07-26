package victoriametrics

import (
	"fmt"
	"testing"
)

func TestSend(t *testing.T) {
	url := "http://127.0.0.1:4242/api/put"
	contentType := "application/json"
	data := `{"timeseries":[{"metric":"zk123","value":45.34,"tags":{"t1":"v1","t2":"v2"},"timestamp":1625560733}]}`
	err := send(url, contentType, []byte(data))
	fmt.Println(err)
}
