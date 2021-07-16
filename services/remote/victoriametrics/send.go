package victoriametrics

import (
	"encoding/json"
	apimetrics "monica-adaptor/api/metrics"
	"monica-adaptor/config"
)

// Send 发送metrics到远端服务
func Send(req []*apimetrics.TimeSeries) (err error) {
	data, err := json.Marshal(req)
	if err != nil {
		return
	}

	err = send(config.Conf.Remote.Send.URL, config.Conf.Remote.Send.ContentType, data)

	return
}
