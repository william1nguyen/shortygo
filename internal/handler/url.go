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

// Shorten godoc
// @Summary 	Shorten URL
// @Description Shorten long URL
// @Tags 		URL
// @Security 	ApiKeyAuth
// @Accept 		json
// @Produce 	json
// @Param 		request body service.ShortenRequest true  "Request body for creating short URL"
// @Success      200      {object}  service.ShortenResponse
// @Failure      400      {object}  ErrorResponse
// @Router       /api/v1/shorten [post]
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

// RedirectURL godoc
// @Summary      Redirect URL
// @Description  Redirect to the original URL using short ID
// @Tags         URL
// @Security     ApiKeyAuth
// @Param        shortId  path      string  true  "Short URL ID"
// @Success      301      {string}  string  "Redirected to original URL"
// @Failure      400      {object}  ErrorResponse  "Bad request or not found"
// @Router       /{shortId} [get]
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

	c.Redirect(http.StatusMovedPermanently, originalURL)
}

// GetMetrics godoc
// @Summary      Get cache metrics
// @Description  Returns cache statistics including hit ratio and total requests
// @Tags         Metrics
// @Security     ApiKeyAuth
// @Produce      json
// @Router       /api/v1/metrics [get]
func (h *URLHandler) GetMetrics(c *gin.Context) {
	metrics := h.service.GetCacheMetrics()

	hitRatio := 0.0
	if metrics.Hits > 0 {
		hitRatio = float64(metrics.Hits) / float64(metrics.TotalRequests)
	}

	c.JSON(http.StatusOK, gin.H{
		"cache_hits":     metrics.Hits,
		"cache_errors":   metrics.Errors,
		"total_requests": metrics.TotalRequests,
		"hit_ratio":      hitRatio,
		"timestamp":      time.Now().Unix(),
	})
}
