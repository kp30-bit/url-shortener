package controller

import (
	"github.com/kp30-bit/url-shortener/internal/interfaces"

	"github.com/gin-gonic/gin"
)

func RegisterURLRoutes(r *gin.Engine, u interfaces.URLUsecase) {

	r.POST("/shorten", u.ShortenURLHandler)
	r.GET("/:shortID", u.GetOriginalURLHandler)
	r.DELETE("/:shortID", u.DeleteURLHandler)
	r.GET("/list", u.ListAllURLsHandler) // dashboard / observability API
	r.GET("/analytics", u.GetAnalyticsHandler)
}
