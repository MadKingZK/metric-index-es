package metrics

import (
	"fmt"
	"testing"
	"time"
)

func TestWrite(t *testing.T) {

	for i := 0; i < 100; i++ {
		JobChannel <- []string{fmt.Sprintf("i:%d", i)}
		time.Sleep(time.Millisecond * 20)
	}
	close(JobChannel)
}
