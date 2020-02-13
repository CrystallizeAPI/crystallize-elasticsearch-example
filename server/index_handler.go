package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/crystallizeapi/crystallize-elasticsearch-example/service"
	"github.com/crystallizeapi/crystallize-elasticsearch-example/types"
)

type Item struct {
	Item types.CatalogueItem `json:"item"`
}

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	begin := time.Now()
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var item Item
	if err = json.Unmarshal(body, &item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	client, err := service.CreateClient()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	indexService := service.IndexService{}
	indexService.Index(ctx, client, item.Item)

	res, err := json.Marshal(item.Item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(res)

	fmt.Printf("Indexed 1 item in %d ms\n", time.Since(begin).Milliseconds())
}
