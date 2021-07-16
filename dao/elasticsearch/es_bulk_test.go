package elasticsearch

import (
	"context"
	"fmt"
	"monica-adaptor/config"
	"testing"

	"github.com/olivere/elastic/v7"
	"go.uber.org/zap"
)

func TestESBulkAPI(t *testing.T) {
	metrics := make([]string, 0, 10)
	metrics = append(metrics,
		`{"content": "abpcnqehmowytila19{instance_id\"i-abcdefghigklmn\",inner_addr\"192.168.1.1\",project_id\"fjajebngiajvjjyn\""}`)
	metrics = append(metrics,
		`{"content": "bbpcnqehmowytila19{instance_id\"i-abcdefghigklmn\",inner_addr\"192.168.1.1\",project_id\"fjajebngiajvjjyn\""}`)
	metrics = append(metrics,
		`{"content": "cbpcnqehmowytila19{instance_id\"i-abcdefghigklmn\",inner_addr\"192.168.1.1\",project_id\"fjajebngiajvjjyn\""}`)
	metrics = append(metrics,
		`{"content": "dbpcnqehmowytila19{instance_id\"i-abcdefghigklmn\",inner_addr\"192.168.1.1\",project_id\"fjajebngiajvjjyn\""}`)
	metrics = append(metrics,
		`{"content": "ebpcnqehmowytila19{instance_id\"i-abcdefghigklmn\",inner_addr\"192.168.1.1\",project_id\"fjajebngiajvjjyn\""}`)
	metrics = append(metrics,
		`{"content": "fbpcnqehmowytila19{instance_id\"i-abcdefghigklmn\",inner_addr\"192.168.1.1\",project_id\"fjajebngiajvjjyn\""}`)
	metrics = append(metrics,
		`{"content": "gbpcnqehmowytila19{instance_id\"i-abcdefghigklmn\",inner_addr\"192.168.1.1\",project_id\"fjajebngiajvjjyn\""}`)
	metrics = append(metrics,
		`{"content": "hbpcnqehmowytila19{instance_id\"i-abcdefghigklmn\",inner_addr\"192.168.1.1\",project_id\"fjajebngiajvjjyn\""}`)
	metrics = append(metrics,
		`{"content": "ibpcnqehmowytila19{instance_id\"i-abcdefghigklmn\",inner_addr\"192.168.1.1\",project_id\"fjajebngiajvjjyn\""}`)
	metrics = append(metrics,
		`{"content": "jbpcnqehmowytila19{instance_id\"i-abcdefghigklmn\",inner_addr\"192.168.1.1\",project_id\"fjajebngiajvjjyn\""}`)

	_ = config.Init()
	_ = Init()

	if client == nil {
		panic("client is nil")
		return
	}

	ctx := context.Background()
	bulkReq := client.Bulk()
	req := elastic.NewBulkIndexRequest()
	for i := range metrics {
		//req := elastic.NewBulkIndexRequest().Index(indexName).Id(metrics[i].Content).Doc(metrics[i])
		req.Index(config.Conf.MetricStore.Store.IndexName).Id(metrics[i]).Doc(metrics[i])
		if req == nil {
			zap.L().Error("BulkAPI panic!", zap.String("data", metrics[i]))
			continue
		}
		bulkReq = bulkReq.Add(req)
	}
	bulkResopne, err := bulkReq.Do(ctx)
	if err != nil {
		panic("ElasticBulkApi do err")
		return
	}

	// failedID := getBulkFailedIDs(bulkResopne)
	parseBulkResp(bulkResopne)
}

func parseBulkResp(r *elastic.BulkResponse) {
	if r.Items == nil {
		return
	}
	for i := range r.Items {
		for _, result := range r.Items[i] {
			fmt.Println(result.Index, result.Result, result.Id, result.Error, result.Status)
		}
	}
	return
}
