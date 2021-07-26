package metrics

import (
	"bytes"
	"context"
	"math/rand"
	"metric-index/config"
	es "metric-index/dao/elasticsearch"
	"metric-index/dao/redis"
	"time"

	"github.com/prometheus/prometheus/prompb"

	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

// Labels 标签key&value
type Labels map[string]string

// Metric 写入es的doc的结构体
type Metric struct {
	Labels
	Content string
}

// MetricStore 存储metric，metric=metricName+label
// 查询redis中是否有metric中的md5，如果没有则插入
// 需要在controller做wq处理，metric组装（调用WQMetricFilterAndAsm或者AsmMetric）
func MetricStore(wq *prompb.WriteRequest) {
	metrics := AsmMetric(wq)
	metricStrings := make([]string, len(metrics))
	for i := range metrics {
		metricStrings[i] = metrics[i].Content
	}

	result, err := redis.PipeExistsByGet(metricStrings)
	if err != nil {
		zap.L().Error("check metric key is exist from redis failed", zap.Error(err))
	}

	json := jsoniter.ConfigCompatibleWithStandardLibrary
	for i := range result {
		if !result[i] {
			metric, err := json.Marshal(metrics[i].Labels)
			if err != nil {
				zap.L().Error("metric struct marshal failed", zap.Error(err))
				continue
			}
			bi := es.BulkIndexerItem{
				Index:           config.Conf.MetricStore.Store.IndexName,
				Action:          "create",
				DocumentID:      "",
				Body:            nil,
				RetryOnConflict: nil,
				OnSuccess:       BulkOnSuccess,
				OnFailure:       nil,
			}
			bi.DocumentID = metrics[i].Content
			bi.Body = bytes.NewReader(metric)
			if err := es.Push(bi); err != nil {
				zap.L().Error("push metric to bulkindexer failed", zap.Error(err))
			}
		}
	}

	return
}

// BulkOnSuccess bulk成功时回调
func BulkOnSuccess(ctx context.Context, item es.BulkIndexerItem, res es.BulkIndexerResponseItem) {
	// 配合PipeExistsByGet打开
	var exTime time.Duration
	if config.Conf.MetricStore.Cache.IsExpire {
		exTime = time.Duration(config.Conf.MetricStore.Cache.Expire-
			config.Conf.MetricStore.Cache.DistInterval+
			rand.Intn(config.Conf.MetricStore.Cache.DistInterval)) *
			time.Second
	} else {
		exTime = time.Duration(-1) * time.Second
	}
	//if err := redis.Push(redis.CommitItem{
	//	Key:    item.DocumentID,
	//	Value:  1,
	//	ExTime: exTime,
	//}); err != nil {
	//	zap.L().Error("push metric to redis committer failed", zap.Error(err))
	//}

	if err := redis.Set(item.DocumentID, 1, exTime); err != nil {
		zap.L().Error("set metric from redis failed", zap.Error(err))
	}
}

// BulkOnFailure bulk失败时回调
func BulkOnFailure(ctx context.Context, item es.BulkIndexerItem, res es.BulkIndexerResponseItem, err error) {
	// 插入ES失败，则删除redis记录
	zap.L().Error("insert into elasticsearch failed", zap.Error(err), zap.Any("err", res.Error))

	// 配合PipeSetNX打开
	//if err = redis.Del(item.DocumentID); err != nil {
	//	zap.L().Error("delete notation from redis failed", zap.Error(err))
	//}
}
