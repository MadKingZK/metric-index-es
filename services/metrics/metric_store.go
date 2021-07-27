package metrics

import (
	"bytes"
	"context"
	"math/rand"
	"metric-index/config"
	es "metric-index/dao/elasticsearch"
	"metric-index/dao/gocache"
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
	Content   string
	IsInCache bool
}

// Store 存储metric，metric=metricName+label
// 查询redis中是否有metric中的md5，如果没有则插入
// 需要在controller做wq处理，metric组装（调用WQMetricFilterAndAsm或者AsmMetric）
func Store(wq *prompb.WriteRequest) {
	metrics := Assembler(wq)
	metricStrings := make([]string, 0, len(metrics))
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	for i := range metrics {
		_, found := gocache.Get(metrics[i].Content)
		metrics[i].IsInCache = found
		if !found {
			metricStrings = append(metricStrings, metrics[i].Content)
		}
	}

	result, err := redis.PipeExistsByGet(metricStrings)
	if err != nil {
		zap.L().Error("check metric key is exist from redis failed", zap.Error(err))
	}

	j := 0
	for i := 0; i < len(metrics) && j < len(result); i++ {
		if metrics[i].IsInCache {
			continue
		}
		if result[j] {
			gocache.SetDefault(metrics[i].Content, 1)
		} else {
			metric, err := json.Marshal(metrics[j].Labels)
			if err != nil {
				zap.L().Error("metric struct marshal failed", zap.Error(err))
				j++
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
			bi.DocumentID = metrics[j].Content
			bi.Body = bytes.NewReader(metric)
			if err := es.Push(bi); err != nil {
				zap.L().Error("push metric to bulkindexer failed", zap.Error(err))
			}
		}
		j++
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

	gocache.SetDefault(item.DocumentID, 1)
	if err := redis.Set(item.DocumentID, 1, exTime); err != nil {
		zap.L().Error("set metric from redis failed", zap.Error(err))
	}

	//if err := redis.Push(redis.CommitItem{
	//	Key:    item.DocumentID,
	//	Value:  1,
	//	ExTime: exTime,
	//}); err != nil {
	//	zap.L().Error("push metric to redis committer failed", zap.Error(err))
	//}

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
