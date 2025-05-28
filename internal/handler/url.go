package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/william1nguyen/shortygo/internal/service"
)

type URLHandler struct {
	service *service.URLService
}

type ErrorResponse struct {
	Error     string `json:"error"`
	Timestamp int64  `json:"timestamp"`
}

func NewURLHandler(service *service.URLService) *URLHandler {
	return &URLHandler{service: service}
}

func (h *URLHandler) ShortenURL(c *gin.Context) {
	var req service.ShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:     "Invalid request body",
			Timestamp: time.Now().Unix(),
		})
		return
	}

	response, err := h.service.ShortenURL(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:     err.Error(),
			Timestamp: time.Now().Unix(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *URLHandler) RedirectURL(c *gin.Context) {
	shortID := c.Param("shortId")
	if shortID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:     "Short ID required",
			Timestamp: time.Now().Unix(),
		})
		return
	}

	originalURL, err := h.service.GetOriginalURL(c.Request.Context(), shortID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:     "URL not found",
			Timestamp: time.Now().Unix(),
		})
		return
	}

	c.JSON(http.StatusMovedPermanently, originalURL)
}

func (h *URLHandler) GetMetrics(c *gin.Context) {
	metrics := h.service.GetCacheMetrics()

	hitRatio := 0.0
	if metrics.Hits+metrics.Misses > 0 {
		hitRatio = float64(metrics.Hits) / float64(metrics.Hits+metrics.Misses)
	}

	c.JSON(http.StatusOK, gin.H{
		"cache_hits":     metrics.Hits,
		"cache_misses":   metrics.Misses,
		"cache_errors":   metrics.Errors,
		"total_requests": metrics.TotalRequests,
		"hit_ratio":      hitRatio,
		"timestamp":      time.Now().Unix(),
	})
}
