package handlers

import (
	"net/http"

	"darts-training-app/internal/models"
	"darts-training-app/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GameHandler struct {
	gameService *services.GameService
}

func NewGameHandler(gameService *services.GameService) *GameHandler {
	return &GameHandler{
		gameService: gameService,
	}
}

func (h *GameHandler) GetAllGameModes(c *gin.Context) {
	gameModes, err := h.gameService.GetAllGameModes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch game modes"})
		return
	}

	// Convert to response format
	response := make([]models.GameModeResponse, len(gameModes))
	for i, gameMode := range gameModes {
		response[i] = gameMode.ToResponse()
	}

	c.JSON(http.StatusOK, response)
}

func (h *GameHandler) GetGamesByTrainingSession(c *gin.Context) {
	sessionIDParam := c.Param("sessionId")
	sessionID, err := uuid.Parse(sessionIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid training session ID format"})
		return
	}

	games, err := h.gameService.GetGamesByTrainingSession(sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch training games"})
		return
	}

	// Convert to response format
	response := make([]models.TrainingGameResponse, len(games))
	for i, game := range games {
		response[i] = game.ToResponse()
	}

	c.JSON(http.StatusOK, response)
}

func (h *GameHandler) CreateGame(c *gin.Context) {
	sessionIDParam := c.Param("sessionId")
	trainingSessionID, err := uuid.Parse(sessionIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid training session ID format"})
		return
	}

	var req struct {
		GameModeID uuid.UUID  `json:"game_mode_id" binding:"required"`
		Player1ID  *uuid.UUID `json:"player1_id"`
		Player2ID  *uuid.UUID `json:"player2_id"`
		Guest1Name *string    `json:"guest1_name"`
		Guest2Name *string    `json:"guest2_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	game, err := h.gameService.CreateGame(
		trainingSessionID,
		req.GameModeID,
		req.Player1ID,
		req.Player2ID,
		req.Guest1Name,
		req.Guest2Name,
	)
	if err != nil {
		if err.Error() == "training session not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Training session not found"})
			return
		}
		if err.Error() == "game mode not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Game mode not found"})
			return
		}
		if err.Error() == "player 1 not found" || err.Error() == "player 2 not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "cannot create games for training session that is not active" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "exactly two players are required for each game" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create game"})
		return
	}

	c.JSON(http.StatusCreated, game.ToResponse())
}

func (h *GameHandler) UpdateGame(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID format"})
		return
	}

	var req struct {
		Player1Score *int    `json:"player1_score"`
		Player2Score *int    `json:"player2_score"`
		Status       *string `json:"status"`
		Winner       *string `json:"winner"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	game, err := h.gameService.UpdateGame(id, req.Player1Score, req.Player2Score, req.Status, req.Winner)
	if err != nil {
		if err.Error() == "game not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
			return
		}
		if err.Error() == "invalid status" || err.Error() == "invalid winner" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update game"})
		return
	}

	c.JSON(http.StatusOK, game.ToResponse())
}

func (h *GameHandler) DeleteGame(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID format"})
		return
	}

	err = h.gameService.DeleteGame(id)
	if err != nil {
		if err.Error() == "game not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
			return
		}
		if err.Error() == "cannot delete game that is in progress or completed" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete game"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Game deleted successfully"})
}

func (h *GameHandler) GenerateGames(c *gin.Context) {
	sessionIDParam := c.Param("sessionId")
	trainingSessionID, err := uuid.Parse(sessionIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid training session ID format"})
		return
	}

	var req struct {
		GameModeID uuid.UUID `json:"game_mode_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	games, err := h.gameService.GenerateGamesForTraining(trainingSessionID, req.GameModeID)
	if err != nil {
		if err.Error() == "training session not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Training session not found"})
			return
		}
		if err.Error() == "can only generate games for planned training sessions" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "need at least 2 attending players to generate games" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate games"})
		return
	}

	// Convert to response format
	response := make([]models.TrainingGameResponse, len(games))
	for i, game := range games {
		response[i] = game.ToResponse()
	}

	c.JSON(http.StatusCreated, response)
}