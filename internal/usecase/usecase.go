package usecase

import (
	"context"
	"math/rand"
	"net/http"
	"time"

	repository "github.com/kp30-bit/url-shortener/internal/repository/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/gin-gonic/gin"
)

type URLUsecase interface {
	ShortenURLHandler(c *gin.Context)
	GetOriginalURLHandler(c *gin.Context)
	ListAllURLsHandler(c *gin.Context)
}

type urlUsecase struct {
	db *repository.MongoDB
}

func NewURLUsecase(mongoDB *repository.MongoDB) URLUsecase {
	return &urlUsecase{db: mongoDB}
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
}

func (u *urlUsecase) ShortenURLHandler(c *gin.Context) {
	var req ShortenURLRequest

	// Step 1: Parse and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	collection := u.db.Database.Collection("urls")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Step 2: Check if original URL already exists
	var existing URLMapping
	err := collection.FindOne(ctx, bson.M{"original_url": req.OriginalURL}).Decode(&existing)
	if err == nil {
		// URL already exists, return existing short URL
		shortURL := "http://localhost:8080/api/url/" + existing.ShortID // replace localhost later
		c.JSON(http.StatusOK, ShortenURLResponse{
			ShortID:  existing.ShortID,
			ShortURL: shortURL,
		})
		return
	} else if err != mongo.ErrNoDocuments {
		// Some other DB error
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store URL mapping"})
		return
	}

	// Step 5: Return the new short URL
	shortURL := "http://localhost:8080/api/url/" + shortID // replace localhost later
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
	// Fetch all URLs (for dev/testing)
	c.JSON(http.StatusOK, gin.H{"message": "list all urls endpoint"})
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
