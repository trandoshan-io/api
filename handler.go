package main

import (
	"github.com/creekorful/microgo/pkg/httputil"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
)

func searchPagesHandler(client *mongo.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// get search criteria
		searchCriteria := r.FormValue("criteria")

		// search for pages
		var pages []SearchResult
		err := searchPages(client, searchCriteria, func(data *PageData) {
			pages = append(pages, SearchResult{Id: data.Id.Hex(), Url: data.Url, CrawlDate: data.CrawlDate})
		})

		if err != nil {
			log.Println("Error while searching pages: " + err.Error())
		}

		// Write json response
		if err := httputil.WriteJsonResponse(w, 200, pages); err != nil {
			log.Println("Error while writing response to client: " + err.Error())
		}
	}
}

func viewPageContentHandler(client *mongo.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Get page
		page, err := getPage(client, mux.Vars(r)["page-id"])
		if err != nil {
			log.Println("Error while getting page: " + err.Error())
			return
		}

		// Write json response
		if err := httputil.WriteJsonResponse(w, 200, page.Content); err != nil {
			log.Println("Error while writing response to client: " + err.Error())
		}
	}
}

func countPagesHandler(client *mongo.Client) func(w http.ResponseWriter, r *http.Request) {
	//pageCollection := client.Database("trandoshan").Collection("pages")
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		return //TODO
	}
}

//TODO: add endpoint to interface with RabbitMQ API
//TODO: add endpoint to push url in todo queue
