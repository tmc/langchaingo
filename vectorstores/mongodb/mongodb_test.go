package mongodb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/tmc/langchaingo/vectorstores/mongodb"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMongoDBConnection(t *testing.T) {
	mongoURL := ""
	store, err := mongodb.New(
		context.TODO(),
		mongodb.WithConnectionUri(mongoURL),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = store.Client.Disconnect(context.TODO()); err != nil {
		  panic(err)
		}
	  }()
	
	// Send a ping to confirm a successful connection
	if err := store.Client.Database("admin").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Err(); err != nil {
	panic(err)
	}
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")
}