package handlers

import (
	"net/http"

	"darts-training-app/internal/models"
	"darts-training-app/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TeamHandler struct {
	teamService *services.TeamService
}

func NewTeamHandler(teamService *services.TeamService) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
	}
}

func (h *TeamHandler) GetAllTeams(c *gin.Context) {
	teams, err := h.teamService.GetAllTeams()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch teams"})
		return
	}

	// Convert to response format
	response := make([]models.TeamResponse, len(teams))
	for i, team := range teams {
		response[i] = team.ToResponse()
	}

	c.JSON(http.StatusOK, response)
}

func (h *TeamHandler) GetTeamByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID format"})
		return
	}

	team, err := h.teamService.GetTeamByID(id)
	if err != nil {
		if err.Error() == "team not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch team"})
		return
	}

	c.JSON(http.StatusOK, team.ToResponse())
}

func (h *TeamHandler) CreateTeam(c *gin.Context) {
	var req models.TeamCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	team, err := h.teamService.CreateTeam(&req)
	if err != nil {
		if err.Error() == "team with name '"+req.Name+"' already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create team"})
		return
	}

	c.JSON(http.StatusCreated, team.ToResponse())
}

func (h *TeamHandler) UpdateTeam(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID format"})
		return
	}

	var req models.TeamUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	team, err := h.teamService.UpdateTeam(id, &req)
	if err != nil {
		if err.Error() == "team not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
			return
		}
		if err.Error() == "team with name '"+*req.Name+"' already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update team"})
		return
	}

	c.JSON(http.StatusOK, team.ToResponse())
}

func (h *TeamHandler) DeleteTeam(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID format"})
		return
	}

	err = h.teamService.DeleteTeam(id)
	if err != nil {
		if err.Error() == "team not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Team deleted successfully"})
}

func (h *TeamHandler) GetTeamPlayers(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID format"})
		return
	}

	players, err := h.teamService.GetTeamPlayers(id)
	if err != nil {
		if err.Error() == "team not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
			return
		}
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