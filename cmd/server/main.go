package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"darts-training-app/internal/config"
	"darts-training-app/internal/database"
	"darts-training-app/internal/handlers"
	"darts-training-app/internal/middleware"
	"darts-training-app/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := database.NewDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Seed default data
	if err := db.SeedDefaultData(); err != nil {
		log.Printf("Warning: Failed to seed default data: %v", err)
	}

	// Initialize services
	authManager := services.NewAuthManager(cfg)
	teamService := services.NewTeamService(db.DB)
	playerService := services.NewPlayerService(db.DB)
	trainingService := services.NewTrainingService(db.DB)
	gameService := services.NewGameService(db.DB)

	// Initialize handlers
	teamHandler := handlers.NewTeamHandler(teamService)
	playerHandler := handlers.NewPlayerHandler(playerService)
	trainingHandler := handlers.NewTrainingHandler(trainingService)
	gameHandler := handlers.NewGameHandler(gameService)

	// Setup Gin router
	if cfg.Port == "8080" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS configuration
	config := cors.DefaultConfig()
	config.AllowOrigins = strings.Split(cfg.FrontendURL, ",")
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	config.AllowCredentials = true
	router.Use(cors.New(config))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "darts-training-app",
			"version": "1.0.0",
		})
	})

	// API routes
	api := router.Group("/api")
	{
		// Protected routes (require authentication)
		protected := api.Group("")
		protected.Use(middleware.CheckAuth(authManager))
		{
			// Team routes
			teams := protected.Group("/teams")
			{
				teams.GET("", teamHandler.GetAllTeams)
				teams.POST("", teamHandler.CreateTeam)
				teams.GET("/:id", teamHandler.GetTeamByID)
				teams.PUT("/:id", teamHandler.UpdateTeam)
				teams.DELETE("/:id", teamHandler.DeleteTeam)
				teams.GET("/:id/players", teamHandler.GetTeamPlayers)
			}

			// Player routes
			players := protected.Group("/players")
			{
				players.GET("", playerHandler.GetAllPlayers)
				players.POST("", playerHandler.CreatePlayer)
				players.GET("/:id", playerHandler.GetPlayerByID)
				players.PUT("/:id", playerHandler.UpdatePlayer)
				players.DELETE("/:id", playerHandler.DeletePlayer)
				players.GET("/team/:teamId", playerHandler.GetPlayersByTeam)
				players.GET("/me", playerHandler.GetCurrentUser)
				players.POST("/me", playerHandler.CreateCurrentUser)
			}

			// Training session routes
			training := protected.Group("/training-sessions")
			{
				training.GET("", trainingHandler.GetAllTrainingSessions)
				training.POST("", trainingHandler.CreateTrainingSession)
				training.GET("/:id", trainingHandler.GetTrainingSessionByID)
				training.PUT("/:id", trainingHandler.UpdateTrainingSession)
				training.DELETE("/:id", trainingHandler.DeleteTrainingSession)
				training.POST("/:id/start", trainingHandler.StartTraining)
				training.POST("/:id/finish", trainingHandler.FinishTraining)
				training.GET("/:id/costs", trainingHandler.GetTrainingCosts)
				training.POST("/:id/players", trainingHandler.AddTrainingPlayer)
				training.DELETE("/players/:playerId", trainingHandler.RemoveTrainingPlayer)
			}

			// Game routes
			games := protected.Group("/games")
			{
				games.GET("/modes", gameHandler.GetAllGameModes)
				games.GET("/training/:sessionId", gameHandler.GetGamesByTrainingSession)
				games.POST("/training/:sessionId", gameHandler.CreateGame)
				games.POST("/training/:sessionId/generate", gameHandler.GenerateGames)
				games.PUT("/:id", gameHandler.UpdateGame)
				games.DELETE("/:id", gameHandler.DeleteGame)
			}
		}
	}

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Starting server on %s", addr)
	log.Printf("API Documentation available at http://localhost%s/health", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
