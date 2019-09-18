package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

// Database page mapping
type PageData struct {
	Id        primitive.ObjectID `bson:"_id"`
	Url       string             `bson:"url"`
	Title     string             `bson:"title"`
	CrawlDate time.Time          `bson:"crawlDate"`
	Content   string             `bson:"content"`
}

type ForbiddenExtensionData struct {
	Extension string `bson:"extension"`
}

// Search pages using search criterias
// callback: callback triggered when a page match given search criteria
// TODO: use channel instead
func searchPages(client *mongo.Client, url string, searchCriteria string, callback func(data *PageData)) error {
	// Setup production context and acquire database collection
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	pageCollection := client.Database("trandoshan").Collection("pages")

	// add search criterias
	// todo: fix or query not working
	filters := bson.M{
		"$or": []bson.M{
			{
				"title":   bson.M{"$regex": primitive.Regex{Pattern: searchCriteria, Options: "i",}},
				"content": bson.M{"$regex": primitive.Regex{Pattern: searchCriteria, Options: "i",}},
			},
		}}

	// add url criteria if set
	if url != "" {
		filters["url"] = bson.M{"url": url}
	}

	// Query the database for result
	cur, err := pageCollection.Find(ctx, filters)
	if err != nil {
		return fmt.Errorf("Error while querying database: " + err.Error())
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var page PageData
		err := cur.Decode(&page)
		if err != nil {
			// if there is a decoding error dot not return error since other results may be decoded
			log.Println("Error while decoding result: " + err.Error())
			// go to next iteration
			break
		}
		// trigger callback with decoded page
		callback(&page)
	}
	if err := cur.Err(); err != nil {
		return fmt.Errorf("Error with cursor: " + err.Error())
	}

	return nil
}

// Get page using his object-id
func getPage(client *mongo.Client, objectIdHex string) (*PageData, error) {
	objectId, err := primitive.ObjectIDFromHex(objectIdHex)
	if err != nil {
		return nil, fmt.Errorf("Unable to convert objectId from hex value: " + err.Error())
	}

	// Setup production context and acquire database collection
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	pageCollection := client.Database("trandoshan").Collection("pages")

	// Query database for result
	var page PageData
	if err := pageCollection.FindOne(ctx, bson.M{"_id": objectId}).Decode(&page); err != nil {
		return nil, fmt.Errorf("Error while decoding result: " + err.Error())
	}

	return &page, nil
}

// Get all forbidden extensions
// callback: callback triggered when a extension is found
// TODO: use channel instead
func getForbiddenExtensions(client *mongo.Client, callback func(string)) error {
	// Setup production context and acquire database collection
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	pageCollection := client.Database("trandoshan").Collection("forbiddenExtensions")

	// Query the database for result
	cur, err := pageCollection.Find(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("Error while querying database: " + err.Error())
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var forbiddenExtension ForbiddenExtensionData
		err := cur.Decode(&forbiddenExtension)
		if err != nil {
			// if there is a decoding error dot not return error since other results may be decoded
			log.Println("Error while decoding result: " + err.Error())
			// go to next iteration
			break
		}
		// trigger callback with decoded page
		//TODO: string pointer instead?
		callback(forbiddenExtension.Extension)
	}
	if err := cur.Err(); err != nil {
		return fmt.Errorf("Error with cursor: " + err.Error())
	}

	return nil
}
