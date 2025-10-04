package controller

import (
	"github.com/kp30-bit/url-shortener/internal/usecase"

	"github.com/gin-gonic/gin"
)

func RegisterURLRoutes(r *gin.Engine, u usecase.URLUsecase) {
	urlGroup := r.Group("/api/url")
	{
		urlGroup.POST("/shorten", u.ShortenURLHandler)
		urlGroup.GET("/:shortID", u.GetOriginalURLHandler)
		urlGroup.GET("/list", u.ListAllURLsHandler) // optional: list all URLs
	}
}
