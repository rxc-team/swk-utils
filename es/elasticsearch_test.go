package es

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/olivere/elastic/v7"
	"rxcsoft.cn/utils/config"
)

var cf config.DB = config.DB{
	Host: "http://192.168.1.26:9200/",
	//Username:       "",
	//Password:       "",
}

func TestCreateESIndex(t *testing.T) {
	StartElastic(cf)
	const mapping = `
	{
	    "mappings": {
	        "properties": {
	            "id": {
	                "type": "text"
	            },
	            "title": {
	                "type": "keyword"
	            },
	            "genres": {
	                "type": "keyword"
	            }
	        }
	    }
	}`
	err := CreateESIndex("test7", mapping, false)
	if err != nil {
		t.Errorf("es create index has error: %v", err)
	}
}

func TestNewESClient(t *testing.T) {
	StartElastic(cf)

	client := NewESClient()
	defer client.Stop()

	ctx := context.Background()
	ex, er := client.IndexExists("test6").Do(ctx)
	if er != nil {
		t.Errorf("es has error: %v", er)
	}
	if ex {
		t.Logf("es ok: %v", ex)
	}
}

func TestUpdateESIndex(t *testing.T) {
	StartElastic(cf)
	const mapping = `
	{
	    "mappings": {
	        "properties": {
	            "id": {
	                "type": "text"
	            },
	            "title": {
	                "type": "keyword"
	            }
	        }
	    }
	}`
	err := UpdateESIndex("test1", "test2", "aaaa", mapping)
	if err != nil {
		t.Errorf("es update index has error: %v", err)
	}
}

func TestESInsert(t *testing.T) {
	StartElastic(cf)

	err := ESInsert("test6", "2", map[string]interface{}{
		"id":     "0002",
		"title":  "北京人民大会d",
		"121212": "v 多大",
		"genres": []string{"ddd", "ccc"},
	})

	if err != nil {
		t.Errorf("es create doc has error: %v", err)
	}
}

func TestESGet(t *testing.T) {
	StartElastic(cf)

	re, err := ESGet("item_5e82b1bbe389ee7c5f29cbd3_97e4d0cf-9b65-4cc0-861f-4776310de84e", "5e82b4a9e389ee7fdaca3ff9")

	if err != nil {
		if elastic.IsNotFound(err) {
			t.Logf("es create doc has error: not found")

			return
		}
		t.Errorf("es create doc has error: %v", err)
		return
	}
	if len(re) == 0 {
		t.Errorf("es create doc has error: not found")
		return
	}

	t.Logf("%v", re)
}

func TestESDelete(t *testing.T) {
	StartElastic(cf)

	err := ESDelete("test6", "2")

	if err != nil {
		t.Errorf("es delete doc has error: %v", err)
	}
}

func TestESDeleteAll(t *testing.T) {
	StartElastic(cf)

	err := ESDeleteAll("test6")

	if err != nil {
		t.Errorf("es delete doc has error: %v", err)
	}
}

func TestDeleteESIndex(t *testing.T) {
	StartElastic(cf)

	err := DeleteESIndex("test")

	if err != nil {
		t.Errorf("es delete index has error: %v", err)
	}
}

func TestESUpdate(t *testing.T) {
	StartElastic(cf)
	err := ESUpdate("test", "5", map[string]interface{}{
		"name": "wujianhua",
	})
	if err != nil {
		t.Errorf("es  update doc has error: %v", err)
	}
}

func TestESUpdateWithScript(t *testing.T) {
	StartElastic(cf)

	script := elastic.NewScript("ctx._source.id += params.num").Param("num", 3)
	err := ESUpsert("test", "5", script, map[string]interface{}{
		"id":    "99999",
		"title": "北京人民大会堂",
	})
	if err != nil {
		t.Errorf("es  update doc with script has error: %v", err)
	}
}

func TestESSearch(t *testing.T) {
	StartElastic(cf)
	termQuery := elastic.NewTermQuery("genres", "ddd")
	result, err := ESSearch("test6", termQuery, "genres", 0, 20)
	if err != nil {
		t.Errorf("es  update doc with script has error: %v", err)
	}

	type Tweet struct {
		ID     string   `json:"id"`
		Title  string   `json:"title"`
		Genres []string `json:"genres"`
	}

	// Here's how you iterate through results with full control over each step.
	if result.TotalHits() > 0 {

		// Iterate through results
		for _, hit := range result.Hits.Hits {
			// hit.Index contains the name of the index

			// Deserialize hit.Source into a Tweet (could also be just a map[string]interface{}).
			var it Tweet
			err := json.Unmarshal(hit.Source, &it)
			if err != nil {
				// Deserialization failed
			}
			t.Logf("%v", it)
		}
	} else {
		// No hits
		t.Logf("Found no tweets\n")
	}
}
