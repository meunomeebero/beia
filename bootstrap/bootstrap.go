package bootstrap

import (
	"beia/bootstrap/app"
	"beia/handlers"
	"beia/middleware"
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Max-Age", "43200")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func BootStrapApp() {
	godotenv.Load()

	app, err := app.NewApplication()
	if err != nil {
		log.Fatalf("Failed to initialize the application settings: %v", err)
	}
	redisOpt, err := redis.ParseURL(app.RedisURL)

	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}

	redisOpt.DialTimeout = time.Duration(app.ClientOptions.DialTimeout)
	redisOpt.ReadTimeout = time.Duration(app.ClientOptions.ReadTimeout)
	redisOpt.WriteTimeout = time.Duration(app.ClientOptions.WriteTimeout)
	redisOpt.PoolTimeout = time.Duration(app.ClientOptions.PoolTimeout)
	redisOpt.MaxRetries = app.ClientOptions.MaxRetries

	redisClient := redis.NewClient(redisOpt)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Successfully connected to Redis")

	completionsHandler := handlers.NewCompletionsHandler(app.OpenAIKey)

	rateLimiter := middleware.NewRateLimiter(redisClient, app.RateLimits)

	router := gin.Default()

	router.Use(CORSMiddleware())

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

	port := app.Port.Env

	if port == "" {
		port = app.Port.Default
	}

	log.Printf("Starting server on port %s", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
