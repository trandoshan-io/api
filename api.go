package main

import (
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"log"
	"net/http"
)

func main() {
	log.Println("Initializing API")

	// load .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Unable to load .env file: ", err.Error())
	}
	log.Println("Loaded .env file")

	router := mux.NewRouter()

	// Register endpoints
	router.HandleFunc("/pages", searchPagesHandler).Methods(http.MethodGet)
	router.HandleFunc("/count-pages", countPagesHandler).Methods(http.MethodGet)

	log.Println("API will listen on: http://0.0.0.0:8080")
	if err := http.ListenAndServe("0.0.0.0:8080", router); err != nil {
		log.Println("fatal error while trying to listen: " + err.Error())
	}
}

func searchPagesHandler(w http.ResponseWriter, r *http.Request) {

}

func countPagesHandler(w http.ResponseWriter, r *http.Request) {

}

//TODO: add endpoint to interface with RabbitMQ API
//TODO: add endpoint to push url in todo queue
