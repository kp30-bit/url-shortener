package config

import (
	"log"
)

type Config struct {
	Port        string
	MongoURI    string
	MongoDBName string
}

func LoadConfig() (*Config, error) {
	port := "8080"
	mongoURI := "mongodb+srv://kamalpratik:youwillneverwalkalone@cluster0.lu5o0r2.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0"
	dbName := "Cluster0"

	log.Printf("📦 Loaded Config: %v,%v", port, dbName)
	return &Config{
		Port:        port,
		MongoURI:    mongoURI,
		MongoDBName: dbName,
	}, nil
}
