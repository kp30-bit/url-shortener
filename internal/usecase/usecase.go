package usecase

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/kp30-bit/url-shortener/config"
	repository "github.com/kp30-bit/url-shortener/internal/repository/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gin-gonic/gin"
)

type URLUsecase interface {
	ShortenURLHandler(c *gin.Context)
	GetOriginalURLHandler(c *gin.Context)
	ListAllURLsHandler(c *gin.Context)
	DeleteURLHandler(c *gin.Context)
	GetAnalyticsHandler(c *gin.Context)
}

type urlUsecase struct {
	db  *repository.MongoDB
	cfg *config.Config
}

func NewURLUsecase(mongoDB *repository.MongoDB, cfg *config.Config) URLUsecase {
	return &urlUsecase{
		db:  mongoDB,
		cfg: cfg,
	}
}

type ShortenURLResponse struct {
	ShortID  string `json:"short_id"`
	ShortURL string `json:"short_url"`
}

type ShortenURLRequest struct {
	OriginalURL string `json:"original_url" binding:"required,url"`
}

type URLMapping struct {
	ShortID     string    `bson:"short_id" json:"short_id"`
	OriginalURL string    `bson:"original_url" json:"original_url"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
	Clicks      int       `bson:"clicks" json:"clicks"`
}

func (u *urlUsecase) ShortenURLHandler(c *gin.Context) {
	var req ShortenURLRequest

	// Step 1: Parse and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Invalid input: %v", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	collection := u.db.Database.Collection("urls")
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()

	// Step 2: Check if original URL already exists
	var existing URLMapping
	err := collection.FindOne(ctx, bson.M{"original_url": req.OriginalURL}).Decode(&existing)
	if err == nil {
		// URL already exists, return existing short URL
		shortURL := fmt.Sprintf("http://%s/api/url/%s", u.cfg.Host, existing.ShortID)
		c.JSON(http.StatusOK, ShortenURLResponse{
			ShortID:  existing.ShortID,
			ShortURL: shortURL,
		})
		return
	} else if err != mongo.ErrNoDocuments {
		// Some other DB error
		log.Printf("error : Database error: %v", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
		return
	}

	// Step 3: Generate a new unique short ID
	shortID := GenerateShortID(8)

	// Step 4: Insert new mapping into MongoDB
	_, err = collection.InsertOne(ctx, map[string]interface{}{
		"short_id":     shortID,
		"original_url": req.OriginalURL,
		"created_at":   time.Now(),
	})
	if err != nil {
		log.Printf("Failed to store URL mapping: %v", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store URL mapping"})
		return
	}

	// Step 5: Return the new short URL
	shortURL := fmt.Sprintf("http://%s/api/url/%s", u.cfg.Host, shortID)
	c.JSON(http.StatusOK, ShortenURLResponse{
		ShortID:  shortID,
		ShortURL: shortURL,
	})
}

func (u *urlUsecase) GetOriginalURLHandler(c *gin.Context) {
	shortID := c.Param("shortID")
	if shortID == "" || len(shortID) != 8 { // assuming Base62 8-char ID
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid shortID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()

	collection := u.db.Database.Collection("urls")
	var result URLMapping
	err := collection.FindOne(ctx, bson.M{"short_id": shortID}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "short URL not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	// Optional: Increment click counter asynchronously
	go func() {
		_, _ = collection.UpdateOne(context.Background(),
			bson.M{"short_id": shortID},
			bson.M{"$inc": bson.M{"clicks": 1}})
	}()

	// Redirect to original URL
	c.Redirect(http.StatusFound, result.OriginalURL)
}

func (u *urlUsecase) ListAllURLsHandler(c *gin.Context) {
	collection := u.db.Database.Collection("urls")
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()

	// Optional: Pagination parameters from query string
	page := 1
	limit := 20

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	skip := (page - 1) * limit

	// MongoDB find options
	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(limit))
	findOptions.SetSort(bson.M{"created_at": -1}) // newest first

	// Query MongoDB
	cursor, err := collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch URLs: " + err.Error()})
		return
	}
	defer cursor.Close(ctx)

	var urls []URLMapping
	if err := cursor.All(ctx, &urls); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse URL records: " + err.Error()})
		return
	}

	// Return response
	c.JSON(http.StatusOK, urls)
}

func GenerateShortID(length int) string {
	charset := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	rand.Seed(time.Now().UnixNano())
	id := make([]rune, length)
	for i := range id {
		id[i] = charset[rand.Intn(len(charset))]
	}
	return string(id)
}

func (u *urlUsecase) DeleteURLHandler(c *gin.Context) {
	shortID := c.Param("shortID")
	if shortID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "shortID missing in request"})
		return
	}

	collection := u.db.Database.Collection("urls")
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()

	// Delete the document with the given shortID
	result, err := collection.DeleteOne(ctx, bson.M{"short_id": shortID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete URL: " + err.Error()})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Short URL deleted successfully"})
}

type Analytics struct {
	TotalURLs   int `json:"total_urls"`
	TotalClicks int `json:"total_clicks"`
}

func (u *urlUsecase) GetAnalyticsHandler(c *gin.Context) {
	collection := u.db.Collection("urls")
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()

	// Total URLs
	totalURLs, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get total URLs"})
		return
	}

	// Total Clicks (aggregate sum)
	cursor, err := collection.Aggregate(ctx, bson.A{
		bson.M{"$group": bson.M{"_id": nil, "totalClicks": bson.M{"$sum": "$clicks"}}},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get total clicks"})
		return
	}
	var result []bson.M
	if err := cursor.All(ctx, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse analytics"})
		return
	}

	totalClicks := 0
	if len(result) > 0 {
		totalClicks = int(result[0]["totalClicks"].(int32)) // mongo returns int32
	}

	c.JSON(http.StatusOK, Analytics{
		TotalURLs:   int(totalURLs),
		TotalClicks: totalClicks,
	})
}
