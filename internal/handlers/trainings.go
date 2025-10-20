package handlers

import (
	"net/http"

	"darts-training-app/internal/models"
	"darts-training-app/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TrainingHandler struct {
	trainingService *services.TrainingService
}

func NewTrainingHandler(trainingService *services.TrainingService) *TrainingHandler {
	return &TrainingHandler{
		trainingService: trainingService,
	}
}

func (h *TrainingHandler) GetAllTrainingSessions(c *gin.Context) {
	sessions, err := h.trainingService.GetAllTrainingSessions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch training sessions"})
		return
	}

	// Convert to response format
	response := make([]models.TrainingSessionResponse, len(sessions))
	for i, session := range sessions {
		response[i] = session.ToResponse()
	}

	c.JSON(http.StatusOK, response)
}

func (h *TrainingHandler) GetTrainingSessionByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid training session ID format"})
		return
	}

	session, err := h.trainingService.GetTrainingSessionByID(id)
	if err != nil {
		if err.Error() == "training session not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Training session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch training session"})
		return
	}

	c.JSON(http.StatusOK, session.ToResponse())
}

func (h *TrainingHandler) CreateTrainingSession(c *gin.Context) {
	var req models.TrainingSessionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get creator ID from context (for now, we'll use a placeholder)
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// For now, we'll use a UUID placeholder until Auth0 integration is complete
	creatorID, err := uuid.Parse("00000000-0000-0000-0000-000000000000")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse creator ID"})
		return
	}

	session, err := h.trainingService.CreateTrainingSession(&req, creatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create training session"})
		return
	}

	c.JSON(http.StatusCreated, session.ToResponse())
}

func (h *TrainingHandler) UpdateTrainingSession(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid training session ID format"})
		return
	}

	var req models.TrainingSessionUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, err := h.trainingService.UpdateTrainingSession(id, &req)
	if err != nil {
		if err.Error() == "training session not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Training session not found"})
			return
		}
		if err.Error() == "invalid status" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update training session"})
		return
	}

	c.JSON(http.StatusOK, session.ToResponse())
}

func (h *TrainingHandler) DeleteTrainingSession(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid training session ID format"})
		return
	}

	err = h.trainingService.DeleteTrainingSession(id)
	if err != nil {
		if err.Error() == "training session not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Training session not found"})
			return
		}
		if err.Error() == "cannot delete training session that is active or completed" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete training session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Training session deleted successfully"})
}

func (h *TrainingHandler) StartTraining(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid training session ID format"})
		return
	}

	session, err := h.trainingService.StartTraining(id)
	if err != nil {
		if err.Error() == "training session not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Training session not found"})
			return
		}
		if err.Error() == "training session is not in planned status" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start training"})
		return
	}

	c.JSON(http.StatusOK, session.ToResponse())
}

func (h *TrainingHandler) FinishTraining(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid training session ID format"})
		return
	}

	session, err := h.trainingService.FinishTraining(id)
	if err != nil {
		if err.Error() == "training session not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Training session not found"})
			return
		}
		if err.Error() == "training session is not active" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to finish training"})
		return
	}

	c.JSON(http.StatusOK, session.ToResponse())
}

func (h *TrainingHandler) GetTrainingCosts(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid training session ID format"})
		return
	}

	costs, err := h.trainingService.GetTrainingCosts(id)
	if err != nil {
		if err.Error() == "training session not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Training session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate training costs"})
		return
	}

	c.JSON(http.StatusOK, costs)
}

func (h *TrainingHandler) AddTrainingPlayer(c *gin.Context) {
	idParam := c.Param("id")
	trainingID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid training session ID format"})
		return
	}

	var req struct {
		GuestName *string `json:"guest_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trainingPlayer, err := h.trainingService.AddTrainingPlayer(trainingID, req.GuestName)
	if err != nil {
		if err.Error() == "training session not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Training session not found"})
			return
		}
		if err.Error() == "cannot add players to completed or cancelled training" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "automatic player assignment is handled during training creation" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Regular players are automatically assigned during training creation"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add training player"})
		return
	}

	c.JSON(http.StatusCreated, trainingPlayer.ToResponse())
}

func (h *TrainingHandler) RemoveTrainingPlayer(c *gin.Context) {
	playerIDParam := c.Param("playerId")
	playerID, err := uuid.Parse(playerIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid player ID format"})
		return
	}

	err = h.trainingService.RemoveTrainingPlayer(playerID)
	if err != nil {
		if err.Error() == "training player not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Training player not found"})
			return
		}
		if err.Error() == "cannot remove players from active, completed, or cancelled training" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "cannot remove regular players, only guests can be removed" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove training player"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Training player removed successfully"})
}