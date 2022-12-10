// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"cloud.google.com/go/logging"
	"github.com/jerrutledge/caption-search-go-api/dbconnection"
	"github.com/jerrutledge/caption-search-go-api/episode"
)

func (a *App) Handler(w http.ResponseWriter, r *http.Request) {
	a.log.Log(logging.Entry{
		Severity: logging.Info,
		HTTPRequest: &logging.HTTPRequest{
			Request: r,
		},
		Labels:  map[string]string{"arbitraryField": "custom entry"},
		Payload: "Structured logging example.",
	})
	fmt.Fprintf(w, "Hello World, and welcome to Caption Search!\n")
}

func (a *App) SearchLogHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// allow certain sites to make requests
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000") // local development
	// w.Header().Set("Access-Control-Allow-Origin", "http://caption-search.jeremyrutledge.com") // web api
	// w.Header().Set("Access-Control-Allow-Origin", "http://jeremyrutledge.com")                // web api
	// allow gets
	w.Header().Set("Access-Control-Allow-Methods", "GET")

	// begin response
	query := r.URL.Query()
	params, present := query["q"]
	if !present || len(params) == 0 {
		a.ReturnError(w, r)
		return
	}
	var queryString string = params[0]
	coll, err := dbconnection.Connect()
	if err != nil {
		a.ReturnError(w, r)
		return
	}
	var response = episode.SearchResults{Err: false}
	err, response.Results = episode.Search(coll, queryString)
	if err != nil {
		a.ReturnError(w, r)
		return
	}
	data, err := json.Marshal(response)
	if err != nil {
		a.ReturnError(w, r)
		return
	}
	a.log.Log(logging.Entry{
		Severity: logging.Info,
		HTTPRequest: &logging.HTTPRequest{
			Request: r,
		},
		Labels:  map[string]string{"query": queryString, "hits": fmt.Sprint(len(response.Results))},
		Payload: "sorry",
	})
	w.Write(data)
}

func (a *App) ReturnError(w http.ResponseWriter, r *http.Request) {
	a.log.Log(logging.Entry{
		Severity: logging.Info,
		HTTPRequest: &logging.HTTPRequest{
			Request: r,
		},
		Labels: map[string]string{"error": "true"},
	})
	var response = episode.SearchResults{Err: true}
	data, err := json.Marshal(response)
	if err != nil {
		return
	}
	w.Write(data)
	return
}
