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

const (
	trandoshanDatabase = "trandoshan"
	resourceCollection = "resources"
)

// Database resource mapping
type resourceData struct {
	Id        primitive.ObjectID `bson:"_id"`
	Url       string             `bson:"url"`
	Title     string             `bson:"title"`
	CrawlDate time.Time          `bson:"crawlDate"`
	Content   string             `bson:"content"`
}

// Search resources using search criterias
// and trigger given callback when a resource match given search criteria
// TODO: use channel instead
func searchResources(client *mongo.Client, url string, searchCriteria string, callback func(data *resourceData)) error {
	// Setup production context and acquire database collection
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	resourceCollection := client.Database(trandoshanDatabase).Collection(resourceCollection)

	// add search criterias
	// todo: fix or query not working
	filters := bson.M{
		"$or": []bson.M{
			{
				"title":   bson.M{"$regex": primitive.Regex{Pattern: searchCriteria, Options: "i"}},
				"content": bson.M{"$regex": primitive.Regex{Pattern: searchCriteria, Options: "i"}},
			},
		}}

	// add url criteria if set
	if url != "" {
		filters["url"] = bson.M{"url": url}
	}

	// Query the database for result
	cur, err := resourceCollection.Find(ctx, filters)
	if err != nil {
		return fmt.Errorf("error while querying database: %s", err)
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var resource resourceData
		err := cur.Decode(&resource)
		if err != nil {
			// if there is a decoding error dot not return error since other results may be decoded
			log.Printf("error while decoding result: %s", err)
			// go to next iteration
			break
		}
		// trigger callback with decoded resource
		callback(&resource)
	}
	if err := cur.Err(); err != nil {
		return fmt.Errorf("error with cursor: %s", err)
	}

	return nil
}

// Get resource using his object-id
func getResource(client *mongo.Client, objectIdHex string) (*resourceData, error) {
	objectId, err := primitive.ObjectIDFromHex(objectIdHex)
	if err != nil {
		return nil, fmt.Errorf("unable to convert objectId from hex value: %s", err)
	}

	// Setup production context and acquire database collection
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	resourceCollection := client.Database(trandoshanDatabase).Collection(resourceCollection)

	// Query database for result
	var resource resourceData
	if err := resourceCollection.FindOne(ctx, bson.M{"_id": objectId}).Decode(&resource); err != nil {
		return nil, fmt.Errorf("error while decoding result: %s", err)
	}

	return &resource, nil
}
