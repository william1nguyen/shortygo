package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/william1nguyen/shortygo/internal/cache"
	"github.com/william1nguyen/shortygo/internal/config"
	"github.com/william1nguyen/shortygo/internal/handler"
	"github.com/william1nguyen/shortygo/internal/service"
)

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}

func setupRedis(cfg *config.Config) *cache.RedisCache {
	redisCache, err := cache.NewRedisCache(cfg.Redis)
	if err != nil {
		log.Fatalf("failed to initialized redis: %v", err)
	}
	return redisCache
}

func setupRouter(urlHandler *handler.URLHandler) *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger(), gin.Recovery())
	router.Use(corsMiddleware())

	setupRoutes(router, urlHandler)
	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func setupRoutes(router *gin.Engine, urlHandler *handler.URLHandler) {

	router.GET("/health", handler.CheckHealth)

	api := router.Group("/api/v1")
	{
		api.POST("/shorten", urlHandler.ShortenURL)
		api.GET("/metrics", urlHandler.GetMetrics)
	}

	router.GET("/:shortId", urlHandler.RedirectURL)
}

func main() {
	loadEnv()
	cfg := config.Load()

	redisCache := setupRedis(cfg)
	defer redisCache.Close()

	urlService := service.NewURLService(redisCache)
	urlHandler := handler.NewURLHandler(urlService)

	router := setupRouter(urlHandler)
	router.Run()
}
