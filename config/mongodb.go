package config

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func NewMongoDB() *MongoDB {
	uri := getEnv("MONGO_URI", "mongodb://localhost:27017")
	dbName := getEnv("MONGO_DB_NAME", "audit_db")

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}

	log.Println("MongoDB connected successfully")

	return &MongoDB{
		Client:   client,
		Database: client.Database(dbName),
	}
}

func (m *MongoDB) Disconnect() error {
	return m.Client.Disconnect(context.TODO())
}
