package dbconnection

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jerrutledge/caption-search-go-api/episode"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Connect() (*mongo.Collection, error) {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Printf("'MONGODB_URI' %s", "not found")
		log.Fatal("You must set your 'MONGODB_URI' environmental variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	collection := client.Database("caption-search").Collection("episodes")
	return collection, err
}

func HelloResponse(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello!")
}

func SearchResponse(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	params, present := query["q"]
	if !present || len(params) == 0 {
		ReturnError(w)
		return
	}
	var queryString string = params[0]
	w.Header().Set("Content-Type", "application/json")
	// allow certain sites to make requests
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000") // local development
	// w.Header().Set("Access-Control-Allow-Origin", "http://caption-search.jeremyrutledge.com") // web api
	// w.Header().Set("Access-Control-Allow-Origin", "http://jeremyrutledge.com")                // web api
	// allow gets
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	coll, err := Connect()
	if err != nil {
		ReturnError(w)
		return
	}
	var response = episode.SearchResults{Err: false}
	err, response.Results = episode.Search(coll, queryString)
	if err != nil {
		ReturnError(w)
		return
	}
	fmt.Println("QUERY: " + queryString + " results: " + fmt.Sprint(len(response.Results)))
	data, err := json.Marshal(response)
	if err != nil {
		ReturnError(w)
		return
	}
	w.Write(data)
}

func ReturnError(w http.ResponseWriter) {
	var response = episode.SearchResults{Err: true}
	data, err := json.Marshal(response)
	if err != nil {
		// TODO: handle error
		fmt.Println("Unhandled MongoDB error")
		log.Fatal(err)
	}
	w.Write(data)
}
