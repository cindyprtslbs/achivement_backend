package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client
var MongoDB *mongo.Database

func ConnectMongo() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Sesuaikan environment
	mongoURI := "mongodb://localhost:27017"
	dbName := "achievement_db"

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("Gagal koneksi ke MongoDB:", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Gagal ping MongoDB:", err)
	}

	MongoClient = client
	MongoDB = client.Database(dbName)

	fmt.Println("Berhasil terhubung ke MongoDB")
}
