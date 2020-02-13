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
		for _, val := range arr {
			queries = append(queries, elastic.NewMatchQuery(key, val))
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
