package metrics

import (
	apimetrics "metric-index/api/metrics"
	"metric-index/config"
	"strings"

	"github.com/prometheus/prometheus/prompb"
)

// WQMetricFilter remote 过滤metric，不符合规范的，直接丢弃
func WQMetricFilter(wq *prompb.WriteRequest) {
	var newTS []*prompb.TimeSeries
	cfg := config.Conf.MetricFilter
	var drop bool
	for _, ts := range wq.Timeseries {
		drop = false
		for _, label := range ts.Labels {
			if label.Name == "__name__" {
				if !cfg.MetricNameRegex.MatchString(label.Value) {
					drop = true
					break
				}
			} else if !cfg.LabelNameRegex.MatchString(label.Name) ||
				!cfg.LabelValueRegex.MatchString(label.Value) {
				drop = true
				break
			}
		}
		if !drop {
			newTS = append(newTS, ts)
		}
	}

	wq.Timeseries = newTS
	return
}

// WQMetricFilterAndAsm 过滤metrics并组装，不符合规范的，直接丢弃，不组装进Metric Slice
// 在即需要组装metrics做缓存和存储，又需要过滤metrics判断是否规范时，WQMetricFilterAndAsm可以减少遍历
func WQMetricFilterAndAsm(wq *prompb.WriteRequest) (Metric []string) {
	var newTS []*prompb.TimeSeries
	cfg := config.Conf.MetricFilter
	var drop bool
	for _, ts := range wq.Timeseries {
		drop = false
		var labelStr string
		for _, label := range ts.Labels {
			if label.Name == "__name__" {
				if !cfg.MetricNameRegex.MatchString(label.Value) {
					drop = true
					break
				}
				labelStr += label.Value + "{"
			} else {
				if !cfg.LabelNameRegex.MatchString(label.Name) ||
					!cfg.LabelValueRegex.MatchString(label.Value) {
					drop = true
					break
				}
				labelStr += label.Name + "=" + "\"" + label.Value + "\"" + ","
			}
		}
		if !drop {
			newTS = append(newTS, ts)
			labelStr = strings.TrimRight(labelStr, ",") + "}"
			Metric = append(Metric, labelStr)
		}
	}

	wq.Timeseries = newTS
	return
}

// MetricFilter 主动发送metrics请求的过滤器
func MetricFilter(wq *apimetrics.WriteReq) {
	var newTS []*apimetrics.TimeSeries
	cfg := config.Conf.MetricFilter
	var drop bool
	for _, ts := range wq.Timeseries {
		drop = false
		if !cfg.MetricNameRegex.MatchString(ts.MetricName) {
			continue
		}

		for name, value := range ts.Labels {
			if !cfg.LabelNameRegex.MatchString(name) ||
				!cfg.LabelValueRegex.MatchString(value) {
				drop = true
				break
			}
		}
		if !drop {
			newTS = append(newTS, ts)
		}
	}

	wq.Timeseries = newTS
	return
}

// MetricFilterAndAsm 过滤metrics并组装，不符合规范的，直接丢弃，不组装进Metric Slice
// 在即需要组装metrics做缓存和存储，又需要过滤metrics判断是否规范时，MetricFilterAndAsm可以减少遍历
func MetricFilterAndAsm(req *apimetrics.WriteReq) (Metric []string) {
	var newTS []*apimetrics.TimeSeries
	cfg := config.Conf.MetricFilter
	var drop bool
	for i := range req.Timeseries {
		drop = false
		var labelStr string
		if !cfg.MetricNameRegex.MatchString(req.Timeseries[i].MetricName) {
			continue
		}
		labelStr += req.Timeseries[i].MetricName + "{"

		for name, value := range req.Timeseries[i].Labels {
			if !cfg.LabelNameRegex.MatchString(name) ||
				!cfg.LabelValueRegex.MatchString(value) {
				drop = true
				break
			}
			labelStr += name + "=" + "\"" + value + "\"" + ","
		}
		if !drop {
			newTS = append(newTS, req.Timeseries[i])
			labelStr = strings.TrimRight(labelStr, ",") + "}"
			Metric = append(Metric, labelStr)
		}
	}

	req.Timeseries = newTS
	return
}
