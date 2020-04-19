package db

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Customer_struct struct {
	Customer_id    string
	Customer_name  string
	Customer_type  string
	Customer_email string
}

var CollectionMongoDB *mongo.Collection

func Test777(i int) int {
	j := i * 2
	return j
}

func GetCollectionMongoBD(Database string, Collection string, HostConnect string) *mongo.Collection {

	clientOptions := options.Client().ApplyURI(HostConnect)
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		fmt.Println(err.Error())
	}
	err = client.Connect(context.Background())
	if err != nil {
		fmt.Println(err.Error())
	}

	err = client.Ping(context.TODO(), readpref.Primary())
	if err != nil {
		fmt.Println("Couldn't connect to the database", err.Error())
	} else {
		fmt.Println("Connected MongoDB!")
	}

	return client.Database(Database).Collection(Collection)
}
