package server

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/crystallizeapi/crystallize-elasticsearch-example/service"
	"github.com/crystallizeapi/crystallize-elasticsearch-example/types"
)

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	var item types.CatalogueItem
	if err = json.Unmarshal(body, &item); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	client, err := service.CreateClient()
	if err != nil {
		log.Fatal(err)
	}

	indexService := service.IndexService{}
	indexService.Index(ctx, client, item)

	res, err := json.Marshal(item)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}
