package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
)

// UploadToMongo upload the processed / enriched documents to MongoDB.
func UploadToMongo(dataType []CompanyDataEnriched) error {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatalln("No URI has been set.")
	}

	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return fmt.Errorf("error connecting to MongoDB: %v", err)
	}
	defer func(client *mongo.Client, ctx context.Context) {
		mongoErr := client.Disconnect(ctx)
		if mongoErr != nil {
			log.Printf("error disconnecting from MongoDB: %v\n", mongoErr)
		}
	}(client, context.TODO())

	databaseName := os.Getenv("MONGO_DB")
	if databaseName == "" {
		databaseName = "mongo" // Default database name
	}

	collectionName := os.Getenv("MONGO_COLLECTION")
	if collectionName == "" {
		collectionName = "enriched_data" // Default collection name
	}

	collection := client.Database(databaseName).Collection(collectionName)

	// Delete all documents in the collection
	_, err = collection.DeleteMany(context.TODO(), bson.D{})
	if err != nil {
		return fmt.Errorf("error deleting documents: %v", err)
	}

	var documents []interface{}
	for _, someDoc := range dataType {
		documents = append(documents, someDoc)
	}

	// Insert the new documents
	_, err = collection.InsertMany(context.TODO(), documents)
	if err != nil {
		return fmt.Errorf("error inserting documents: %v", err)
	}

	log.Println("Successfully deleted and uploaded to MongoDB")
	return nil
}

// FetchDocumentsFromMongo fetches the unprocessed documents from MongoDB
func FetchDocumentsFromMongo(client *mongo.Client) ([]CompanyDataOld, error) {
	databaseName := os.Getenv("MONGO_DB")
	if databaseName == "" {
		databaseName = "mongo" // Default database name
	}

	collectionName := os.Getenv("MONGO_COLLECTION")
	if collectionName == "" {
		collectionName = "raw_data" // Default collection name
	}

	collection := client.Database(databaseName).Collection(collectionName)

	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, fmt.Errorf("error inserting documents: %v", err)
	}

	var results []interface{}
	if err = cursor.All(context.TODO(), &results); err != nil {
		log.Panic(err)
	}

	fmt.Println("Successfully fetched from MongoDB")

	return results, nil
}
