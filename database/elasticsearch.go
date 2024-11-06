package database

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	elastic "github.com/olivere/elastic/v7"
	"rxcsoft.cn/utils/config"
)

var (
	env *config.DB
)

// StartElastic 初始化Elastic
func StartElastic(c config.DB) {
	if env == nil {
		env = &c
	}
}

func buildClientOption() (u []string) {
	host := strings.Split(env.Host, ",")

	var urls []string

	for _, url := range host {
		urls = append(urls, url)
	}

	return urls
}

// NewESClient 创建一个客户端
func NewESClient() *elastic.Client {
	client, err := elastic.NewClient(
		elastic.SetURL(buildClientOption()...),
		elastic.SetBasicAuth(env.Username, env.Password),
		elastic.SetSniff(false))
	if err != nil {
		panic(err)
	}

	return client
}

// CreateESIndexByJson 创建索引
// @param indexName string
// @param mapping string
// const mapping = `
// {
//     "mappings": {
//         "properties": {
//             "id": {
//                 "type": "long"
//             },
//             "title": {
//                 "type": "text"
//             },
//             "genres": {
//                 "type": "keyword"
//             }
//         }
//     }
// }`
func ExistsESIndex(indexName string) bool {
	client := NewESClient()
	defer client.Stop()

	ctx := context.Background()

	exists, err := client.IndexExists(indexName).Do(ctx)
	if err != nil {
		return false
	}

	if exists {
		return true
	}
	return false
}

// CreateESIndex 创建索引
// @param indexName string
// @param mapping string
// const mapping = `
// {
//     "mappings": {
//         "properties": {
//             "id": {
//                 "type": "long"
//             },
//             "title": {
//                 "type": "text"
//             },
//             "genres": {
//                 "type": "keyword"
//             }
//         }
//     }
// }`
func CreateESIndex(indexName, mapping string, alias bool) error {
	client := NewESClient()
	defer client.Stop()

	ctx := context.Background()

	if alias {
		suffix := uuid.New()
		name := indexName + "_" + suffix.String()
		exists, err := client.IndexExists(name).Do(ctx)
		if err != nil {
			return err
		}

		if !exists {
			result, err := client.CreateIndex(name).BodyString(mapping).Do(ctx)
			if err != nil {
				log.Errorf("ES create index error: %v", err)
				return err
			}

			//  创建别名
			_, e1 := client.Alias().
				Action(elastic.NewAliasAddAction(indexName).Index(name)).
				Do(ctx)
			if e1 != nil {
				return e1
			}

			log.Infof("ES create index ok: %v", result.Index)
		}

		return nil
	}

	exists, err := client.IndexExists(indexName).Do(ctx)
	if err != nil {
		return err
	}

	if !exists {
		result, err := client.CreateIndex(indexName).BodyString(mapping).Do(ctx)
		if err != nil {
			log.Errorf("ES create index error: %v", err)
			return err
		}

		log.Infof("ES create index ok: %v", result.Index)
	}
	return nil
}

// CreateESIndexByJson 创建索引
// @param indexName string
// @param mapping string
// const mapping = `
// {
//     "mappings": {
//         "properties": {
//             "id": {
//                 "type": "long"
//             },
//             "title": {
//                 "type": "text"
//             },
//             "genres": {
//                 "type": "keyword"
//             }
//         }
//     }
// }`
func CreateESIndexByJson(indexName string, mapping interface{}, alias bool) error {
	client := NewESClient()
	defer client.Stop()

	ctx := context.Background()

	if alias {
		suffix := uuid.New()
		name := indexName + "_" + suffix.String()
		exists, err := client.IndexExists(name).Do(ctx)
		if err != nil {
			return err
		}

		if !exists {
			result, err := client.CreateIndex(name).BodyJson(mapping).Do(ctx)
			if err != nil {
				log.Errorf("ES create index error: %v", err)
				return err
			}

			//  创建别名
			_, e1 := client.Alias().
				Action(elastic.NewAliasAddAction(indexName).Index(name)).
				Do(ctx)
			if e1 != nil {
				return e1
			}

			log.Infof("ES create index ok: %v", result.Index)
		}
		return nil
	}

	exists, err := client.IndexExists(indexName).Do(ctx)
	if err != nil {
		return err
	}

	if !exists {
		result, err := client.CreateIndex(indexName).BodyJson(mapping).Do(ctx)
		if err != nil {
			log.Errorf("ES create index error: %v", err)
			return err
		}

		log.Infof("ES create index ok: %v", result.Index)
	}
	return nil
}

// GetESIndexName 通过别名获取 index 名
func GetESIndexName(alias string) ([]string, error) {
	client := NewESClient()
	defer client.Stop()

	ctx := context.Background()
	index, err := client.Aliases().Alias(alias).Do(ctx)
	if err != nil {
		return []string{}, err
	}

	return index.IndicesByAlias(alias), nil
}

// GetESIndexAlias 通过index名获取别名
func GetESIndexAlias(indexName string) ([]string, error) {
	client := NewESClient()
	defer client.Stop()

	ctx := context.Background()
	alias, err := client.Aliases().Index(indexName).Do(ctx)
	if err != nil {
		return []string{}, err
	}

	var result []string

	for _, alia := range alias.Indices {
		for _, name := range alia.Aliases {
			result = append(result, name.AliasName)
		}

	}

	return result, nil

}

// UpdateESIndex 重建索引
func UpdateESIndex(oldIndexName, newIndexName, alias string, mapping interface{}) error {
	client := NewESClient()
	defer client.Stop()

	ctx := context.Background()

	// 1. 创建别名
	_, e1 := client.Alias().
		Action(elastic.NewAliasAddAction(alias).Index(oldIndexName)).
		Do(ctx)
	if e1 != nil {
		return e1
	}

	// 2. 创建新索引
	exists, e2 := client.IndexExists(newIndexName).Do(ctx)
	if e2 != nil {
		return e2
	}

	if !exists {
		result, err := client.CreateIndex(newIndexName).BodyJson(mapping).Do(ctx)
		if err != nil {
			log.Errorf("ES create index error: %v", err)
			return err
		}

		log.Infof("ES create index ok: %v", result.Index)
	}

	// 3. 复制数据
	_, e4 := client.Reindex().
		Source(elastic.NewReindexSource().Index(oldIndexName)).Size(5000).
		Destination(elastic.NewReindexDestination().Index(newIndexName).Routing("=cat")).
		Do(ctx)
	if e4 != nil {
		return e4
	}

	// 4. 修改别名
	_, e3 := client.Alias().
		Action(elastic.NewAliasAddAction(alias).Index(newIndexName), elastic.NewAliasRemoveAction(alias).Index(oldIndexName)).
		Do(ctx)
	if e3 != nil {
		return e3
	}

	// 删除旧索引
	_, e5 := client.DeleteIndex(oldIndexName).
		Do(ctx)
	if e5 != nil {
		return e5
	}

	return nil
}

// RecreateIndex 更新索引
func RecreateIndex(indexName string, script *elastic.Script, query elastic.Query) error {
	client := NewESClient()
	defer client.Stop()

	ctx := context.Background()

	result, err := client.UpdateByQuery().
		Index(indexName).
		Script(script).
		Query(query).
		Do(ctx)
	if err != nil {
		return err
	}

	if result.Total > 0 {
		log.Infof("ES recreate index ok: %v", result.Total)
	}
	return nil
}

// DeleteESIndex 删除索引
func DeleteESIndex(indexName string) error {
	client := NewESClient()
	defer client.Stop()

	ctx := context.Background()
	result, err := client.DeleteIndex(indexName).Do(ctx)
	if err != nil {
		log.Errorf("ES delete index error: %v", err)
		return err
	}

	log.Infof("ES delete index ok: %v", result.Acknowledged)
	return nil
}

// ESFlush 刷新索引，保证写入成功
func ESFlush(indexName string) error {
	client := NewESClient()
	defer client.Stop()

	ctx := context.Background()
	result, err := client.Flush(indexName).Do(ctx)
	if err != nil {
		log.Errorf("ES flush index error: %v", err)
		return err
	}

	log.Infof("ES flush index ok: %v", result.Shards)
	return nil
}

// ESInsert 插入数据（json serialization）
func ESInsert(indexName string, id string, body interface{}) error {
	client := NewESClient()
	defer client.Stop()

	ctx := context.Background()
	result, err := client.Index().
		Index(indexName).
		Id(id).
		BodyJson(body).
		Do(ctx)
	if err != nil {
		log.Errorf("ES insert interface in index error: %v", err)
		return err
	}

	log.Infof("ES insert interface in index ok: %v", result.Result)
	return nil
}

// ESInsert 插入数据(json string)
func ESInsertByString(indexName string, id string, body string) error {
	client := NewESClient()
	defer client.Stop()

	ctx := context.Background()
	result, err := client.Index().
		Index(indexName).
		Id(id).
		BodyString(body).
		Do(ctx)
	if err != nil {
		log.Errorf("ES insert string in index error: %v", err)
		return err
	}

	log.Infof("ES insert string in index ok: %v", result.Result)
	return nil
}

// ESGet 获取单个文档
func ESGet(indexName string, id string) (map[string]interface{}, error) {
	client := NewESClient()
	defer client.Stop()

	ctx := context.Background()
	result, err := client.Get().
		Index(indexName).
		Id(id).
		Do(ctx)
	if err != nil {
		log.Errorf("ES got document error: %v", err)
		return nil, err
	}

	if result.Found {
		log.Infof("ES got document %s in version %d from index %v, source %s\n", result.Id, result.Version, result.Index, result.Source)

		var s map[string]interface{}

		e := json.Unmarshal(result.Source, &s)
		if e != nil {
			log.Errorf("ES got document error: %v", e)
		}

		return s, nil
	}
	return nil, nil
}

// ESDelete 删除单个文档
func ESDelete(indexName string, id string) (e error) {
	client := NewESClient()
	defer client.Stop()

	ctx := context.Background()
	result, err := client.Delete().
		Index(indexName).
		Id(id).
		Do(ctx)
	if err != nil {
		log.Errorf("ES delete document error: %v", err)
		return err
	}

	log.Infof("ES delete document %s in version %d from index %s\n", result.Id, result.Version, result.Index)
	return nil
}

// ESDelete 删除index 下所有文档
func ESDeleteAll(indexName string) (e error) {
	client := NewESClient()
	defer client.Stop()

	ctx := context.Background()
	result, err := client.DeleteByQuery().
		Index(indexName).
		Query(elastic.NewMatchAllQuery()).
		Do(ctx)
	if err != nil {
		log.Errorf("ES delete all document error: %v", err)
		return err
	}

	log.Infof("ES delete all document %d in from index\n", result.Total)
	return nil
}

// ESUpdate 更新单个文档
func ESUpdate(indexName string, id string, doc map[string]interface{}) (e error) {
	client := NewESClient()
	defer client.Stop()

	ctx := context.Background()
	result, err := client.Update().
		Index(indexName).
		Id(id).
		Doc(doc).
		FetchSource(true).
		Do(ctx)
	if err != nil {
		log.Errorf("ES update document error: %v", err)
		return err
	}

	log.Infof("ES update document %s in version %d from index %s\n", result.Id, result.Version, result.Index)
	return nil
}

// ESUpsert 更新值，没有的情况下使用默认传入的值
func ESUpsert(indexName string, id string, script *elastic.Script, doc map[string]interface{}) (e error) {
	client := NewESClient()
	defer client.Stop()

	ctx := context.Background()
	result, err := client.Update().
		Index(indexName).
		Id(id).
		Script(script).
		Upsert(doc).
		Do(ctx)
	if err != nil {
		log.Errorf("ES update with script document error: %v", err)
		return err
	}

	log.Infof("ES update with script document %s in version %d from index %s\n", result.Id, result.Version, result.Index)
	return nil
}

// ESSearch 检索
// // searchResult is of type SearchResult and returns hits, suggestions,
// // and all kinds of other information from Elasticsearch.
// fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)
//
// // Each is a convenience function that iterates over hits in a search result.
// // It makes sure you don't need to check for nil values in the response.
// // However, it ignores errors in serialization. If you want full control
// // over iterating the hits, see below.
// var ttyp Tweet
// for _, item := range searchResult.Each(reflect.TypeOf(ttyp)) {
//     t := item.(Tweet)
//     fmt.Printf("Tweet by %s: %s\n", t.User, t.Message)
// }
// // TotalHits is another convenience function that works even when something goes wrong.
// fmt.Printf("Found a total of %d tweets\n", searchResult.TotalHits())
//
// // Here's how you iterate through results with full control over each step.
// if searchResult.TotalHits() > 0 {
//     fmt.Printf("Found a total of %d tweets\n", searchResult.TotalHits())
//
//     // Iterate through results
//     for _, hit := range searchResult.Hits.Hits {
//         // hit.Index contains the name of the index
//
//         // Deserialize hit.Source into a Tweet (could also be just a map[string]interface{}).
//         var t Tweet
//         err := json.Unmarshal(hit.Source, &t)
//         if err != nil {
//             // Deserialization failed
//         }
//
//         // Work with tweet
//         fmt.Printf("Tweet by %s: %s\n", t.User, t.Message)
//     }
// } else {
//     // No hits
//     fmt.Print("Found no tweets\n")
// }
func ESSearch(indexName string, termQuery elastic.Query, sort string, start, size int) (result *elastic.SearchResult, err error) {
	client := NewESClient()
	defer client.Stop()

	ctx := context.Background()

	searchResult, e := client.Search().
		Index(indexName).
		Query(termQuery).
		Sort(sort, true).
		From(start).Size(size).
		Pretty(true).
		Do(ctx) // execute
	if e != nil {
		log.Errorf("ES search error: %v", e)
		return nil, err
	}

	// searchResult is of type SearchResult and returns hits, suggestions,
	log.Infof("ES search took %d milliseconds\n", searchResult.TookInMillis)
	return searchResult, nil
}
