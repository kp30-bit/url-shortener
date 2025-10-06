package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kp30-bit/url-shortener/config"
	"github.com/kp30-bit/url-shortener/internal/controller"
	repository "github.com/kp30-bit/url-shortener/internal/repository/mongo"
	"github.com/kp30-bit/url-shortener/internal/usecase"
)

// App struct
type App struct {
	Router      *gin.Engine
	URLUsecase  usecase.URLUsecase
	MongoClient *repository.MongoClient
	Config      *config.Config
}

// Singletons
var (
	cfgInstance *config.Config
	cfgOnce     sync.Once
	mongoClient *repository.MongoClient
	mongoDB     *repository.MongoDB
	mongoOnce   sync.Once
)

// GetConfig singleton
func GetConfig() *config.Config {
	cfgOnce.Do(func() {
		c, err := config.LoadConfig()
		if err != nil {
			log.Fatalf("❌ Failed to load configuration: %v", err)
		}
		cfgInstance = c
		log.Println("✅ Configuration loaded successfully")
	})
	return cfgInstance
}

// GetMongo singleton
func GetMongo() (*repository.MongoClient, *repository.MongoDB) {
	mongoOnce.Do(func() {
		cfg := GetConfig()
		client, db, err := repository.InitMongo(cfg.MongoURI, cfg.MongoDBName)
		if err != nil {
			log.Fatalf("❌ Failed to initialize MongoDB: %v", err)
		}
		mongoClient = client
		mongoDB = db
		log.Println("✅ Connected to MongoDB successfully")
	})
	return mongoClient, mongoDB
}

// NewApp initializes application
func NewApp() *App {
	cfg := GetConfig()
	client, db := GetMongo()
	urlUsecase := usecase.NewURLUsecase(db, cfg)
	router := gin.Default()

	// Register routes
	controller.RegisterURLRoutes(router, urlUsecase)

	return &App{
		Router:      router,
		URLUsecase:  urlUsecase,
		MongoClient: client,
		Config:      cfg,
	}
}

func main() {
	app := NewApp()
	cfg := GetConfig()

	// Start HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: app.Router,
	}

	go func() {
		log.Printf("🚀 Server is running on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server error: %v", err)
		}
	}()

	// Graceful shutdown
	gracefulShutdown(srv, mongoClient)
}

func gracefulShutdown(srv *http.Server, client *repository.MongoClient) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown HTTP server first
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}
	log.Println("✅ HTTP server stopped")

	// Disconnect MongoDB
	if err := client.Disconnect(ctx); err != nil {
		log.Fatalf("❌ Error disconnecting MongoDB: %v", err)
	}
	log.Println("✅ MongoDB connection closed")
}
