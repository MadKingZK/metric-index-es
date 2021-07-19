package metrics

import (
	apimetrics "monica-adaptor/api/metrics"
	"monica-adaptor/services/metrics"
	"monica-adaptor/services/remote/victoriametrics"
	"net/http"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

// Metrics 处理主动发送metric请求
func Metrics(c *gin.Context) {
	req := new(apimetrics.WriteReq)
	if err := c.ShouldBindJSON(&req); err != nil {
		zap.L().Error("bad request, req struct can not bind json", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": http.StatusText(http.StatusBadRequest),
			"data":    "",
		})
		return
	}

	var metricSlice []string
	metricSlice = metrics.MetricFilterAndAsm(req)
	if len(req.Timeseries) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"message": "success",
			"data":    "",
		})
		return
	}

	go metrics.MetricStore(metricSlice)

	// 发送metrics到victoriaMetrics
	if err := victoriametrics.Send(req.Timeseries); err != nil {
		zap.L().Error("send to vm failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": http.StatusText(http.StatusInternalServerError),
			"data":    "",
		})
		return
	}
	zap.L().Info("send to vm success")

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "success",
		"data":    "",
	})
	return
}
