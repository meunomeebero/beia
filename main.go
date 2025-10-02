package main

import (
	"context"
	"log"
	"os"
	"time"

	"beia/handlers"
	"beia/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	godotenv.Load()

	redisURL := os.Getenv("REDIS_URL")

	redisOpt, err := redis.ParseURL(redisURL)

	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}

	redisOpt.DialTimeout = 10 * time.Second
	redisOpt.ReadTimeout = 5 * time.Second
	redisOpt.WriteTimeout = 5 * time.Second
	redisOpt.PoolTimeout = 10 * time.Second
	redisOpt.MaxRetries = 3

	redisClient := redis.NewClient(redisOpt)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Successfully connected to Redis")

	openaiAPIKey := os.Getenv("OPENAI_KEY")

	completionsHandler := handlers.NewCompletionsHandler(openaiAPIKey)

	rateLimiter := middleware.NewRateLimiter(redisClient)

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := redisClient.Ping(ctx).Err(); err != nil {
			c.JSON(500, gin.H{
				"status": "unhealthy",
				"redis":  "disconnected",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "healthy",
			"redis":  "connected",
		})
	})

	router.POST("/completions", rateLimiter.Middleware(), completionsHandler.HandleCompletion)

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
