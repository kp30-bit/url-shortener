package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port        string
	MongoURI    string
	MongoDBName string
}

func Load() (*Config, error) {
	port := os.Getenv("PORT")
	mongoURI := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DB_NAME")

	if port == "" {
		port = "8080"
	}
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}
	if dbName == "" {
		dbName = "url_shortener"
	}

	fmt.Println("📦 Loaded Config: ", port, dbName)
	return &Config{
		Port:        port,
		MongoURI:    mongoURI,
		MongoDBName: dbName,
	}, nil
}
