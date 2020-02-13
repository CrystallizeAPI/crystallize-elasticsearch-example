package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/crystallizeapi/crystallize-elasticsearch-example/service"
	"github.com/olivere/elastic/v7"
)

func HandleSearch(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()

	var queries []elastic.Query
	for key, arr := range qs {
		if len(arr) == 1 {
			queries = append(queries, elastic.NewMatchQuery(key, arr[0]))
		} else {
			// Here we want to handle an array of possibilities for the same key. For
			// example: ?id=123&id=456 will search for matching items with each id.
			var subQueries []elastic.Query
			for _, val := range arr {
				subQueries = append(subQueries, elastic.NewMatchQuery(key, val))
			}
			queries = append(queries, elastic.NewBoolQuery().Should(subQueries...))
		}
	}

	query := elastic.NewBoolQuery().Must(queries...)
	searchService := service.SearchService{}

	ctx := context.Background()
	client, err := service.CreateClient()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := searchService.Search(ctx, client, query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}
