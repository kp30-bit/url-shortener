package usecase

import (
	"net/http"

	repository "github.com/kp30-bit/url-shortener/internal/repository/mongo"

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

func (u *urlUsecase) ShortenURLHandler(c *gin.Context) {
	// Parse JSON request, generate short ID, save to Mongo
	c.JSON(http.StatusOK, gin.H{"message": "shorten url endpoint"})
}

func (u *urlUsecase) GetOriginalURLHandler(c *gin.Context) {
	// Read shortID from URL, fetch original URL from Mongo
	c.JSON(http.StatusOK, gin.H{"message": "get original url endpoint"})
}

func (u *urlUsecase) ListAllURLsHandler(c *gin.Context) {
	// Fetch all URLs (for dev/testing)
	c.JSON(http.StatusOK, gin.H{"message": "list all urls endpoint"})
}
