package configuration

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// most important
var DB *mongo.Client

func InitDataBase() error {
	//client
	mongo_uri := os.Getenv("MONGO_URI")
	if mongo_uri == "" {
		return fmt.Errorf("MONGO_URI environment variable is not set")
	}

	clientOption := options.Client().ApplyURI(mongo_uri)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	//connect to mongodb
	client, err := mongo.Connect(context.TODO(), clientOption)

	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Mongodb connected successfully")
	DB = client
	return nil

}

func GetCollection(col_name string) (*mongo.Collection, error) {
	db_name := os.Getenv("DB_NAME")

	if DB == nil {
		return nil, fmt.Errorf("database client is not initialized")
	}

	if col_name == "" {
		return nil, fmt.Errorf("collection name is not mentioned")

	}

	if db_name == "" {
		return nil, fmt.Errorf("db name  is not initialized")
	}
	return DB.Database(db_name).Collection(col_name), nil
}
