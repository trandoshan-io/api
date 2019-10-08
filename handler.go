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
type searchResult struct {
	Id        string    `json:"id"`
	Url       string    `json:"url"`
	Title     string    `json:"title"`
	CrawlDate time.Time `json:"crawlDate"`
}

type clientCommand struct {
	Command string `json:"command"`
	Payload string `json:"payload"`
}

func searchResourcesHandler(client *mongo.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// get search criterias
		url := r.FormValue("url")
		searchCriteria := r.FormValue("criteria")

		// search for resources
		var resources []searchResult
		err := searchResources(client, url, searchCriteria, func(data *resourceData) {
			resources = append(resources, searchResult{Id: data.Id.Hex(), Url: data.Url, Title: data.Title, CrawlDate: data.CrawlDate})
		})

		if err != nil {
			log.Printf("Error while searching resources: %s", err)
		}

		// Write json response
		if err := httputil.WriteJsonResponse(w, 200, resources); err != nil {
			log.Printf("Error while writing response to client: %s", err)
		}
	}
}

func getCrawledUrls(client *mongo.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// search for resources
		var crawledUrls []string
		err := searchResources(client, "", "", func(data *resourceData) {
			crawledUrls = append(crawledUrls, data.Url)
		})

		if err != nil {
			log.Printf("Error while searching resource: %s", err)
		}

		// Write json response
		if err := httputil.WriteJsonResponse(w, 200, crawledUrls); err != nil {
			log.Printf("Error while writing response to client: %s", err)
		}
	}
}

func viewResourceContentHandler(client *mongo.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Get resource
		resource, err := getResource(client, mux.Vars(r)["resource-id"])
		if err != nil {
			log.Printf("Error while getting resource: %s", err)
			return
		}

		// Write json response
		if err := httputil.WriteJsonResponse(w, 200, resource.Content); err != nil {
			log.Printf("Error while writing response to client: %s", err)
		}
	}
}

func resourcesStreamHandler(client *mongo.Client) func(w http.ResponseWriter, r *http.Request) {
	//TODO: better upgrader with security check
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
		return true
	}}

	return func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Error while upgrading connection: %s", err)
			return
		}
		defer c.Close()

		for {
			// Read message from client
			mt, message, err := c.ReadMessage()
			if err != nil {
				log.Printf("Error while reading message from client: %s", err)
				break
			}

			var command clientCommand
			if err := json.Unmarshal(message, &command); err != nil {
				log.Printf("Error while un-marshalling command: %s", err)
				break
			}

			// Manage command
			switch {
			// search command
			case command.Command == "search":
				_ = searchResources(client, "", command.Payload, func(data *resourceData) {
					resourceBytes, err := json.Marshal(searchResult{Id: data.Id.Hex(), Url: data.Url, Title: data.Title, CrawlDate: data.CrawlDate})
					if err != nil {
						log.Printf("Error while marshalling resource: %s", err)
						return
					}

					_ = c.WriteMessage(mt, resourceBytes)
				})
				// get resource content command
			case command.Command == "get-content":
				resource, err := getResource(client, command.Payload)
				if err != nil {
					log.Printf("Error while getting resource: %s", err)
					return
				}

				// Encode resource content in base64
				//TODO: improve
				contentBase64 := base64.StdEncoding.EncodeToString([]byte(resource.Content))
				_ = c.WriteMessage(mt, []byte(contentBase64))
			}

			// Write message to client
			err = c.WriteMessage(mt, message)
			if err != nil {
				log.Printf("Error while writing message to client: %s", err)
				break
			}
		}
	}
}

//TODO: add endpoint to interface with NATS API
//TODO: add endpoint to push url in todo queue
