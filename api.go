package main

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"net/http"
	"os"
	"time"
)

type SearchCommand struct {
	Criteria string `json:"criteria"`
}

type SearchResult struct {
	Url       string    `json:"url"`
	CrawlDate time.Time `json:"crawlDate"`
}

func main() {
	log.Println("Initializing API")

	// load .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Unable to load .env file: ", err.Error())
	}
	log.Println("Loaded .env file")

	// initialize and validate database connection
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		log.Fatal("Unable to create database connection: ", err.Error())
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatal("Unable to connect to database: ", err.Error())
	}

	router := mux.NewRouter()

	// Register endpoints
	router.HandleFunc("/pages", searchPagesHandler(client)).Methods(http.MethodGet)
	router.HandleFunc("/count-pages", countPagesHandler(client)).Methods(http.MethodGet)

	// Register real-time endpoints
	router.HandleFunc("/pages-stream", searchPagesStreamHandler(client))

	log.Println("API will listen on: http://0.0.0.0:8080")
	if err := http.ListenAndServe("0.0.0.0:8080", router); err != nil {
		log.Println("fatal error while trying to listen: " + err.Error())
	}
}

func searchPagesHandler(client *mongo.Client) func(w http.ResponseWriter, r *http.Request) {
	//contentCollection := client.Database("trandoshan").Collection("pages")
	return func(w http.ResponseWriter, r *http.Request) {
		return //TODO
	}
}

func countPagesHandler(client *mongo.Client) func(w http.ResponseWriter, r *http.Request) {
	//contentCollection := client.Database("trandoshan").Collection("pages")
	return func(w http.ResponseWriter, r *http.Request) {
		return //TODO
	}
}

//TODO: add endpoint to interface with RabbitMQ API
//TODO: add endpoint to push url in todo queue

func searchPagesStreamHandler(client *mongo.Client) func(w http.ResponseWriter, r *http.Request) {
	//TODO: better upgrader
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
		return true
	}}
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Error while upgrading connection:", err.Error())
			return
		}
		defer c.Close()

		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				log.Println("Error while reading:", err.Error())
				break
			}

			// Unmarshal command
			var command SearchCommand
			if err := json.Unmarshal(message, &command); err != nil {
				log.Println("Error while unmarshaling command: " + err.Error())
				break
			}

			// todo process command here

			var results []SearchResult
			results = append(results, SearchResult{
				Url:       "http://sometzegregreg.onion",
				CrawlDate: time.Now(),
			})

			// Serialize result into json
			resultJson, err := json.Marshal(results)
			if err != nil {
				log.Println("Error while marshalling result: " + err.Error())
				break
			}

			// Send data back to client
			err = c.WriteMessage(mt, resultJson)
			if err != nil {
				log.Println("Error while writing response to client: " + err.Error())
				break
			}
		}
	}
}
