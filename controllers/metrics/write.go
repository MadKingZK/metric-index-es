package metrics

import (
	"io/ioutil"
	"math"
	"monica-adaptor/library"
	"monica-adaptor/library/cgroup"
	"monica-adaptor/services/metrics"
	"monica-adaptor/services/remote/victoriametrics"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/prometheus/prometheus/prompb"
)

var JobChannel = make(chan []string, math.MaxUint16)
var numCpu int

func init() {
	numCpu = cgroup.AvailableCPUs()
	for i := 0; i < numCpu; i++ {
		go workChannel()
	}
}

func workChannel() {
	defer func() {
		if r := recover(); r != nil {
			zap.L().Fatal("work channel fun panic")
		}
	}()
	m := make([]string, 0, math.MaxUint16/numCpu+1)
	timer := library.Get(time.Millisecond * 100)
	for {
		select {
		case value, ok := <-JobChannel:
			if !ok {
				library.Put(timer)
				return
			}
			m = append(m, value...)
		case <-timer.C:
			metrics.MetricStore(m)
			m = m[:0]
			library.Put(timer)
		}
	}
}

// Write 接收remote write
func Write(c *gin.Context) {
	cmpBody, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		zap.L().Error("read request.body failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": http.StatusText(http.StatusBadRequest),
			"data":    "",
		})
		return
	}
	//c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(cmpBody))

	// 解压cmp_body
	body, err := snappy.Decode(nil, cmpBody)
	if err != nil {
		zap.L().Error("uncompress request.body failed", zap.Error(err))
		c.JSON(http.StatusTooManyRequests, gin.H{
			"code":    http.StatusTooManyRequests,
			"message": http.StatusText(http.StatusTooManyRequests),
			"data":    "",
		})
		return
	}

	var wq = new(prompb.WriteRequest)
	if err = proto.Unmarshal(body, wq); err != nil {
		panic(err)
	}

	metricSlice := metrics.AsmMetric(wq)
	JobChannel <- metricSlice
	//metricSlice := metrics.WQMetricFilterAndAsm(wq)
	//if len(wq.Timeseries) == 0 {
	//	c.JSON(http.StatusOK, gin.H{
	//		"code":    http.StatusOK,
	//		"message": "success",
	//		"data":    "",
	//	})
	//	return
	//}

	// 发送metrics到victoriaMetrics，不能异步发送
	// Prometheus对remote wirte有错误处理，失败时会retry重试，阻塞等待
	//if err = victoriametrics.ReqForward(cmpBody); err != nil {
	if err = victoriametrics.Write(wq); err != nil {
		zap.L().Error("send to vm failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": http.StatusText(http.StatusInternalServerError),
			"data":    "",
		})
	}
	zap.L().Info("send to vm success")

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "success",
		"data":    "",
	})
	return
}
