package elasticsearch

import (
	"context"
	"errors"
	"monica-adaptor/config"

	jsoniter "github.com/json-iterator/go"
	"github.com/olivere/elastic/v7"
	"go.uber.org/zap"
)

// Metric 插入ES的metric数据格式
type Metric struct {
	Content string `json:"content"`
}

var client *elastic.Client

// Init 初始化Elasticsearch
func Init() (err error) {
	client, err = elastic.NewClient(elastic.SetURL(config.Conf.MetricStore.Store.URL...),
		elastic.SetBasicAuth(config.Conf.MetricStore.Store.UserName, config.Conf.MetricStore.Store.Password))
	if err != nil {
		zap.L().Error("Elastic init err", zap.Error(err))
		return err
	}
	return nil
}

// InsertDoc 插入单个doc
func InsertDoc(indexName string, doc interface{}) (res *elastic.IndexResponse, err error) {
	ctx := context.Background()
	res, err = client.Index().Index(indexName).BodyJson(doc).Do(ctx)
	return
}

//BulkAPI 批量同步到ES
func BulkAPI(indexName string, metrics []*Metric) ([]string, error) {
	if len(metrics) <= 0 {
		return nil, nil
	}

	if client == nil {
		return nil, errors.New("client is nil")
	}

	json := jsoniter.ConfigCompatibleWithStandardLibrary
	ctx := context.Background()
	bulkReq := client.Bulk()
	req := elastic.NewBulkIndexRequest()
	for i := range metrics {
		//req := elastic.NewBulkIndexRequest().Index(indexName).Id(metrics[i].Content).Doc(metrics[i])
		metric, err := json.Marshal(metrics[i])
		if err != nil {
			zap.L().Error("Marshal metric failed", zap.Error(err))
		}
		req.Index(indexName).Id(metrics[i].Content).Doc(string(metric))
		if req == nil {
			zap.L().Error("BulkAPI panic!", zap.Any("data", map[string]interface{}{"data": metrics[i]}))
			continue
		}
		bulkReq = bulkReq.Add(req)
	}

	bulkResopne, err := bulkReq.Do(ctx)
	if err != nil {
		zap.L().Error("ElasticBulkApi do err", zap.Error(err))
		return nil, err
	}

	// failedID := getBulkFailedIDs(bulkResopne)
	successIDs := getBulkSuccessIDs(bulkResopne)
	return successIDs, nil
}

func getBulkSuccessIDs(r *elastic.BulkResponse) (IDs []string) {
	if r.Items == nil {
		return nil
	}
	for i := range r.Items {
		for _, result := range r.Items[i] {
			if !(result.Status >= 200 && result.Status <= 299) {
				zap.L().Error("bulk exec elastic failed")
				continue
			}
			IDs = append(IDs, result.Id)
		}
	}
	return
}

func getBulkFailedIDs(r *elastic.BulkResponse) (IDs []string) {
	if r.Items == nil {
		return nil
	}
	for i := range r.Items {
		for _, result := range r.Items[i] {
			if !(result.Status >= 200 && result.Status <= 299) {
				zap.L().Error("bulk exec elastic failed")
				IDs = append(IDs, result.Id)
			}
		}
	}
	return
}
