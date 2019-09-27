package main

import (
	"encoding/base64"
	"encoding/json"
	"github.com/creekorful/microgo/pkg/httputil"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"time"
)

// Json object returned to the client
type SearchResult struct {
	Id        string    `json:"id"`
	Url       string    `json:"url"`
	Title     string    `json:"title"`
	CrawlDate time.Time `json:"crawlDate"`
}

type ClientCommand struct {
	Command string `json:"command"`
	Payload string `json:"payload"`
}

func searchPagesHandler(client *mongo.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// get search criterias
		url := r.FormValue("url")
		searchCriteria := r.FormValue("criteria")

		// search for pages
		var pages []SearchResult
		err := searchPages(client, url, searchCriteria, func(data *PageData) {
			pages = append(pages, SearchResult{Id: data.Id.Hex(), Url: data.Url, Title: data.Title, CrawlDate: data.CrawlDate})
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

func getCrawledUrls(client *mongo.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// search for pages
		var crawledUrls []string
		err := searchPages(client, "", "", func(data *PageData) {
			crawledUrls = append(crawledUrls, data.Url)
		})

		if err != nil {
			log.Println("Error while searching pages: " + err.Error())
		}

		// Write json response
		if err := httputil.WriteJsonResponse(w, 200, crawledUrls); err != nil {
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

func pagesStreamHandler(client *mongo.Client) func(w http.ResponseWriter, r *http.Request) {
	//TODO: better upgrader with security check
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
		return true
	}}

	return func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Error while upgrading connection:" + err.Error())
			return
		}
		defer c.Close()

		for {
			// Read message from client
			mt, message, err := c.ReadMessage()
			if err != nil {
				log.Println("Error while reading message from client: " + err.Error())
				break
			}

			var command ClientCommand
			if err := json.Unmarshal(message, &command); err != nil {
				log.Println("Error while un-marshalling command: " + err.Error())
				break
			}

			// Manage command
			switch {
			// search command
			case command.Command == "search":
				_ = searchPages(client, "", command.Payload, func(data *PageData) {
					pageBytes, err := json.Marshal(SearchResult{Id: data.Id.Hex(), Url: data.Url, Title: data.Title, CrawlDate: data.CrawlDate})
					if err != nil {
						log.Println("Error while marshalling page: " + err.Error())
						return
					}

					_ = c.WriteMessage(mt, pageBytes)
				})
			// get page content command
			case command.Command == "get-content":
				page, err := getPage(client, command.Payload)
				if err != nil {
					log.Println("Error while getting page: " + err.Error())
					return
				}

				// Encode page content in base64
				//TODO: improve
				contentBase64 := base64.StdEncoding.EncodeToString([]byte(page.Content))
				_ = c.WriteMessage(mt, []byte(contentBase64))
			}

			// Write message to client
			err = c.WriteMessage(mt, message)
			if err != nil {
				log.Println("Error while writing message to client: " + err.Error())
				break
			}
		}
	}
}

//TODO: add endpoint to interface with NATS API
//TODO: add endpoint to push url in todo queue
