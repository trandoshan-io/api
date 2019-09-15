package main

import (
	"context"
	"github.com/creekorful/microgo/pkg/httputil"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
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

func searchPagesHandler(client *mongo.Client) func(w http.ResponseWriter, r *http.Request) {
	pageCollection := client.Database("trandoshan").Collection("pages")
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// get search criteria
		searchCriteria := r.FormValue("criteria")

		// setup production context
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

		// Query the database for result
		filter := bson.D{{"content", primitive.Regex{Pattern: searchCriteria, Options: "i"}}}
		cur, err := pageCollection.Find(ctx, filter)
		if err != nil {
			log.Println("Error while querying database: " + err.Error())
			return
		}
		defer cur.Close(ctx)

		var pages []SearchResult
		for cur.Next(ctx) {
			var page PageData
			err := cur.Decode(&page)
			if err != nil {
				log.Println("Error while decoding result: " + err.Error())
				break
			}
			pages = append(pages, SearchResult{Id: page.Id.Hex(), Url: page.Url, CrawlDate: page.CrawlDate})
		}
		if err := cur.Err(); err != nil {
			log.Println("Error with cursor: " + err.Error())
			return
		}

		// Write json response
		if err := httputil.WriteJsonResponse(w, 200, pages); err != nil {
			log.Println("Error while writing response to client: " + err.Error())
		}
	}
}

func viewPageContentHandler(client *mongo.Client) func(w http.ResponseWriter, r *http.Request) {
	pageCollection := client.Database("trandoshan").Collection("pages")
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Get page ID
		pageId, err := primitive.ObjectIDFromHex(mux.Vars(r)["page-id"])
		if err != nil {
			log.Println("Unable to convert objectId from hex value: " + err.Error())
			return
		}

		// Setup production context
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

		// Query database for result
		var page PageData
		if err := pageCollection.FindOne(ctx, bson.M{"_id": pageId}).Decode(&page); err != nil {
			log.Println("Error while decoding result: " + err.Error())
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