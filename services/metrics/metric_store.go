package metrics

import (
	"math/rand"
	"monica-adaptor/config"
	"monica-adaptor/dao/elasticsearch"
	"monica-adaptor/dao/redis"
	"time"

	"go.uber.org/zap"
)

// MetricStore 存储metric，metric=metricName+label
// 查询redis中是否有metric中的md5，如果没有则插入
// 需要在controller做wq处理，metric组装（调用WQMetricFilterAndAsm或者AsmMetric）
func MetricStore(metrics []string) {
	expire := config.Conf.MetricStore.Cache.Expire
	distInterval := config.Conf.MetricStore.Cache.DistInterval
	var exTime time.Duration

	result, err := redis.PipeExistsByGet(metrics)
	if err != nil {
		zap.L().Error("check metric key is exist from redis failed", zap.Error(err))
	}

	needInsertMetrics := make([]*elasticsearch.Metric, 0, len(result))
	for i := range result {
		if !result[i] {
			needInsertMetrics = append(needInsertMetrics, &elasticsearch.Metric{Content: metrics[i]})
		}
	}

	if len(needInsertMetrics) == 0 {
		return
	}

	if config.Conf.MetricStore.Cache.IsExpire {
		exTime = time.Duration(expire-distInterval+rand.Intn(distInterval)) * time.Second
	} else {
		exTime = time.Duration(-1) * time.Second
	}

	bulkInsertMetric(needInsertMetrics, exTime)
	return
}

func bulkInsertMetric(metrics []*elasticsearch.Metric, exTime time.Duration) {
	successIDs, err := elasticsearch.BulkAPI(config.Conf.MetricStore.Store.IndexName, metrics)
	if err != nil && successIDs == nil {
		zap.L().Error("bulkinsert metrics into elastic failed", zap.Error(err))
		return
	}

	kvs := make(map[string]interface{}, len(successIDs))
	for i := range successIDs {
		kvs[successIDs[i]] = 1
	}
	if err = redis.PipeSet(kvs, exTime); err != nil {
		zap.L().Error("pipeset metric into redis failed", zap.Error(err))
	}
	return
}

func insertMetric(metric string) {
	var metricStruct = new(elasticsearch.Metric)
	metricStruct.Content = metric
	_, err := elasticsearch.InsertDoc(config.Conf.MetricStore.Store.IndexName, metricStruct)
	if err == nil {
		return
	}

	// 插入ES失败，则删除redis记录
	zap.L().Error("insert into elasticsearch failed", zap.Error(err))
	if err = redis.Del(metric); err != nil {
		zap.L().Error("delete notation from redis failed", zap.Error(err))
	}

}
