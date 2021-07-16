package metrics

import (
	"bytes"
	"unsafe"

	"github.com/prometheus/prometheus/prompb"
)

// AsmMetric 组装metric
func AsmMetric(wq *prompb.WriteRequest) []string {
	metric := make([]string, len(wq.Timeseries))
	var b bytes.Buffer
	for i := range wq.Timeseries {
		b.Reset()
		b.WriteString(wq.Timeseries[i].Labels[0].Value)
		b.WriteString("{")
		for j, l := 1, len(wq.Timeseries[i].Labels); j < l; j++ {
			b.WriteString(wq.Timeseries[i].Labels[j].Name)
			b.WriteString(`"`)
			b.WriteString(wq.Timeseries[i].Labels[j].Value)
			b.WriteString(`",`)
		}
		b.WriteString(`",`)
		metric[i] = bytes2str(b.Bytes())
	}
	return metric
}

func bytes2str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
