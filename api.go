package main

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"net/http"
	"os"
	"time"
)

// Json object returned to the client
type SearchResult struct {
	Id        string    `json:"id"`
	Url       string    `json:"url"`
	CrawlDate time.Time `json:"crawlDate"`
}

// Database page mapping
type PageData struct {
	Id        primitive.ObjectID `bson:"_id"`
	Url       string             `bson:"url"`
	CrawlDate time.Time          `bson:"crawlDate"`
	Content   string             `bson:"content"`
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
	router.HandleFunc("/pages", searchPagesHandler(client)).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/pages/{page-id}/content", viewPageContentHandler(client)).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/count-pages", countPagesHandler(client)).Methods(http.MethodGet, http.MethodOptions)

	// todo add websocket endpoint

	router.Use(mux.CORSMethodMiddleware(router))

	log.Println("API will listen on: http://0.0.0.0:8080")
	if err := http.ListenAndServe("0.0.0.0:8080", router); err != nil {
		log.Println("fatal error while trying to listen: " + err.Error())
	}
}