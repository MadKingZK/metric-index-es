package elasticsearch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"metric-index/config"
	"time"

	"go.uber.org/zap"

	"metric-index/utils/fasthttp"

	"github.com/cenkalti/backoff/v4"
	elastic "github.com/elastic/go-elasticsearch/v7"
	jsoniter "github.com/json-iterator/go"
)

var es *elastic.Client
var bi BulkIndexer

type customJSONDecoder struct{}

func (d customJSONDecoder) UnmarshalFromReader(r io.Reader, blk *BulkIndexerResponse) error {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	return json.NewDecoder(r).Decode(blk)
}

// Init 初始化esclient和bulkindexer
func Init() (err error) {
	retryBackoff := backoff.NewExponentialBackOff()
	conf := elastic.Config{
		Addresses:    config.Conf.MetricStore.Store.URL,
		Transport:    &fasthttp.Transport{},
		DisableRetry: false,
		RetryBackoff: func(i int) time.Duration {
			if i == 1 {
				retryBackoff.Reset()
			}
			return retryBackoff.NextBackOff()
		},
	}

	if es, err = elastic.NewClient(conf); err != nil {
		return
	}

	res, err := es.Info()
	if err != nil {
		return
	}
	if res.IsError() {
		return errors.New("connect ES failed")
	}
	fmt.Println("============-> Connect ES Success <-============")
	fmt.Println(res.String())

	if err := InitBulkIndexer(); err != nil {
		fmt.Println("xxxxxxxxxxxx-> Init BulkIndexer Err <-xxxxxxxxxxxx")
	}

	fmt.Println("============-> Init BulkIndexer Success <-============")
	return
}

// InitBulkIndexer 初始化 bulkindexer
func InitBulkIndexer() (err error) {
	bi, err = NewBulkIndexer(BulkIndexerConfig{
		Index:         config.Conf.MetricStore.Store.IndexName,                   // The default index name
		Client:        es,                                                        // The Elasticsearch client
		NumWorkers:    config.Conf.MetricStore.Store.WorkerNum,                   // The number of worker goroutines
		FlushBytes:    config.Conf.MetricStore.Store.FlushBytes,                  // The flush threshold in bytes
		FlushInterval: config.Conf.MetricStore.Store.FlushInterval * time.Second, // The periodic flush interval
		Decoder:       customJSONDecoder{},
	})
	if err != nil {
		zap.L().Error("Error creating the indexer: %s", zap.Error(err))
	}
	return
}

// CloseBulkIndexer 关闭bulkindexer，在main中defer调用
func CloseBulkIndexer() (err error) {
	if err = bi.Close(context.Background()); err != nil {
		zap.L().Error("Close BulkIndexer Failed", zap.Error(err))
	}
	return
}

// Push 推送消息
func Push(bulkIndexerItem BulkIndexerItem) (err error) {
	err = bi.Add(context.Background(), bulkIndexerItem)

	return
}

// BulkStats 获取es bulk api状态
func BulkStats() BulkIndexerStats {
	stats := bi.Stats()
	return stats
}
