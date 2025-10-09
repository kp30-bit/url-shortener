package interfaces

import (
	"github.com/gin-gonic/gin"
)

type URLUsecase interface {
	ShortenURLHandler(c *gin.Context)
	GetOriginalURLHandler(c *gin.Context)
	ListAllURLsHandler(c *gin.Context)
	DeleteURLHandler(c *gin.Context)
	GetAnalyticsHandler(c *gin.Context)
}
