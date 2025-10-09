package domain

import "time"

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
