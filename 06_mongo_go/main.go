package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/JagdeepSingh13/06_mongo_go/controllers"
	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	r := httprouter.New()
	client := getMongoClient()
	uc := controllers.NewUserController(client)

	r.GET("/user/:id", uc.GetUser)
	r.POST("/user", uc.CreateUser)
	r.DELETE("/user/:id", uc.DeleteUser)

	log.Println("Server running on localhost:9000")
	log.Fatal(http.ListenAndServe("localhost:9000", r))
}

func getMongoClient() *mongo.Client {
	// Define the MongoDB connection URI
	uri := "mongodb://127.0.0.1:27017"

	// Set client options
	clientOptions := options.Client().ApplyURI(uri)

	// Create a new client and connect to the server
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatalf("Failed to create MongoDB client: %v", err)
	}

	// Create a context with a timeout for the connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB
	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Ping the database to verify the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	log.Println("Connected to MongoDB successfully")
	return client
}
