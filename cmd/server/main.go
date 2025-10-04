package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kp30-bit/url-shortener/config"
	"github.com/kp30-bit/url-shortener/internal/controller"
	repository "github.com/kp30-bit/url-shortener/internal/repository/mongo"
	"github.com/kp30-bit/url-shortener/internal/usecase"
)

func main() {
	cfg := loadConfig()

	client, mongoDB, err := initMongoDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialise mongo with error : %v", err)
	}

	router := setupRouter(mongoDB)
	startServer(router, cfg)
	handleGracefulShutdown(client)

}

func loadConfig() *config.Config {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}
	log.Println("✅ Configuration loaded successfully")
	return cfg
}

func initMongoDB(cfg *config.Config) (*repository.MongoClient, *repository.MongoDB, error) {
	client, mongoDB, err := repository.InitMongo(cfg.MongoURI, cfg.MongoDBName)
	if err == nil {
		log.Println("✅ Connected to MongoDB successfully")
	}
	return client, mongoDB, err
}

func setupRouter(mongoDB *repository.MongoDB) *gin.Engine {
	r := gin.Default()

	// Initialize UseCase Layer
	urlUsecase := usecase.NewURLUsecase(mongoDB)

	// Register all routes
	controller.RegisterURLRoutes(r, urlUsecase)

	return r
}

func startServer(router *gin.Engine, cfg *config.Config) {
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("🚀 Server is running on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server error: %v", err)
		}
	}()
}

func handleGracefulShutdown(client *repository.MongoClient) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Disconnect(ctx); err != nil {
		log.Fatalf("❌ Error disconnecting MongoDB: %v", err)
	}
	log.Println("✅ MongoDB connection closed.")
}
