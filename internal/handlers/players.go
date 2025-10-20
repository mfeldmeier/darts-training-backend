package handlers

import (
	"net/http"

	"darts-training-app/internal/models"
	"darts-training-app/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PlayerHandler struct {
	playerService *services.PlayerService
}

func NewPlayerHandler(playerService *services.PlayerService) *PlayerHandler {
	return &PlayerHandler{
		playerService: playerService,
	}
}

func (h *PlayerHandler) GetAllPlayers(c *gin.Context) {
	// Get query parameters for filtering
	teamIDParam := c.Query("team_id")

	var players []models.Player
	var err error

	if teamIDParam != "" {
		teamID, parseErr := uuid.Parse(teamIDParam)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID format"})
			return
		}
		players, err = h.playerService.GetPlayersByTeam(teamID)
	} else {
		players, err = h.playerService.GetAllPlayers()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch players"})
		return
	}

	// Convert to response format
	response := make([]models.PlayerResponse, len(players))
	for i, player := range players {
		response[i] = player.ToResponse()
	}

	c.JSON(http.StatusOK, response)
}

func (h *PlayerHandler) GetPlayerByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid player ID format"})
		return
	}

	player, err := h.playerService.GetPlayerByID(id)
	if err != nil {
		if err.Error() == "player not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Player not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch player"})
		return
	}

	c.JSON(http.StatusOK, player.ToResponseWithTeam())
}

func (h *PlayerHandler) GetPlayersByTeam(c *gin.Context) {
	idParam := c.Param("teamId")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID format"})
		return
	}

	players, err := h.playerService.GetPlayersByTeam(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch team players"})
		return
	}

	// Convert to response format
	response := make([]models.PlayerResponse, len(players))
	for i, player := range players {
		response[i] = player.ToResponse()
	}

	c.JSON(http.StatusOK, response)
}

func (h *PlayerHandler) CreatePlayer(c *gin.Context) {
	var req models.PlayerCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	player, err := h.playerService.CreatePlayer(&req)
	if err != nil {
		if err.Error() == "player with email '"+req.Email+"' already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "team not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create player"})
		return
	}

	c.JSON(http.StatusCreated, player.ToResponse())
}

func (h *PlayerHandler) UpdatePlayer(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid player ID format"})
		return
	}

	var req models.PlayerUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	player, err := h.playerService.UpdatePlayer(id, &req)
	if err != nil {
		if err.Error() == "player not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Player not found"})
			return
		}
		if err.Error() == "team not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "player with email '"+*req.Email+"' already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update player"})
		return
	}

	c.JSON(http.StatusOK, player.ToResponse())
}

func (h *PlayerHandler) DeletePlayer(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid player ID format"})
		return
	}

	err = h.playerService.DeletePlayer(id)
	if err != nil {
		if err.Error() == "player not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Player not found"})
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Player deleted successfully"})
}

func (h *PlayerHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	player, err := h.playerService.GetPlayerByAuth0ID(userID.(string))
	if err != nil {
		if err.Error() == "player not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Player profile not found. Please create a player profile first."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch player profile"})
		return
	}

	c.JSON(http.StatusOK, player.ToResponseWithTeam())
}

func (h *PlayerHandler) CreateCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userEmail, _ := c.Get("user_email")
	userName, _ := c.Get("user_name")
	userNickname, _ := c.Get("user_nickname")

	// Check if player already exists
	_, err := h.playerService.GetPlayerByAuth0ID(userID.(string))
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Player profile already exists"})
		return
	}

	if err.Error() != "player not found" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing player"})
		return
	}

	var req models.PlayerCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Override with authenticated user data
	req.Email = userEmail.(string)
	if req.Name == "" {
		req.Name = userName.(string)
	}
	if req.Nickname == nil {
		nickname := userNickname.(string)
		req.Nickname = &nickname
	}

	player, err := h.playerService.CreatePlayer(&req)
	if err != nil {
		if err.Error() == "player with email '"+req.Email+"' already exists" {
			// Player exists with email but not Auth0 ID, link them
			existingPlayer, fetchErr := h.playerService.GetPlayerByAuth0ID(userID.(string))
			if fetchErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link existing player profile"})
				return
			}
			c.JSON(http.StatusOK, existingPlayer.ToResponse())
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create player profile"})
		return
	}

	// Link with Auth0 ID
	if err := h.playerService.UpdatePlayerAuth0ID(player.ID, userID.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link player profile with Auth0"})
		return
	}

	c.JSON(http.StatusCreated, player.ToResponse())
}
