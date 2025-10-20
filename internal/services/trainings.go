package services

import (
	"fmt"

	"darts-training-app/internal/models"
	"darts-training-app/internal/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TrainingService struct {
	db *gorm.DB
}

func NewTrainingService(db *gorm.DB) *TrainingService {
	return &TrainingService{
		db: db,
	}
}

func (s *TrainingService) GetAllTrainingSessions() ([]models.TrainingSession, error) {
	var sessions []models.TrainingSession
	err := s.db.Preload("Creator").
		Preload("TrainingPlayers.Player").
		Preload("Games.GameMode").
		Preload("Games.Player1").
		Preload("Games.Player2").
		Order("training_date DESC").
		Find(&sessions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch training sessions: %w", err)
	}
	return sessions, nil
}

func (s *TrainingService) GetTrainingSessionByID(id uuid.UUID) (*models.TrainingSession, error) {
	var session models.TrainingSession
	err := s.db.Preload("Creator").
		Preload("TrainingPlayers.Player").
		Preload("Games.GameMode").
		Preload("Games.Player1").
		Preload("Games.Player2").
		First(&session, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("training session not found")
		}
		return nil, fmt.Errorf("failed to fetch training session: %w", err)
	}
	return &session, nil
}

func (s *TrainingService) CreateTrainingSession(req *models.TrainingSessionCreateRequest, creatorID uuid.UUID) (*models.TrainingSession, error) {
	session := &models.TrainingSession{
		Name:          req.Name,
		Description:   req.Description,
		TrainingDate:  req.TrainingDate,
		CostPerPlayer: 5.00, // Default cost
		Status:        "planned",
		CreatedBy:     &creatorID,
	}

	if req.CostPerPlayer != nil {
		session.CostPerPlayer = *req.CostPerPlayer
	}

	// Start transaction
	tx := s.db.Begin()

	if err := tx.Create(session).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create training session: %w", err)
	}

	// Auto-assign all players to training
	var players []models.Player
	if err := tx.Find(&players).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to fetch players for auto-assignment: %w", err)
	}

	// Create training players for all existing players
	for _, player := range players {
		trainingPlayer := &models.TrainingPlayer{
			TrainingSessionID: session.ID,
			PlayerID:          &player.ID,
			IsGuest:           false,
			Attended:          true, // Default to attended
		}
		if err := tx.Create(trainingPlayer).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to assign player to training: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return s.GetTrainingSessionByID(session.ID)
}

func (s *TrainingService) UpdateTrainingSession(id uuid.UUID, req *models.TrainingSessionUpdateRequest) (*models.TrainingSession, error) {
	var session models.TrainingSession
	if err := s.db.First(&session, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("training session not found")
		}
		return nil, fmt.Errorf("failed to fetch training session: %w", err)
	}

	// Update fields if provided
	if req.Name != nil {
		session.Name = *req.Name
	}
	if req.Description != nil {
		session.Description = req.Description
	}
	if req.TrainingDate != nil {
		session.TrainingDate = *req.TrainingDate
	}
	if req.CostPerPlayer != nil {
		session.CostPerPlayer = *req.CostPerPlayer
	}
	if req.Status != nil {
		// Validate status
		validStatuses := []string{"planned", "active", "completed", "cancelled"}
		statusValid := false
		for _, status := range validStatuses {
			if *req.Status == status {
				statusValid = true
				break
			}
		}
		if !statusValid {
			return nil, fmt.Errorf("invalid status: %s", *req.Status)
		}
		session.Status = *req.Status
	}

	if err := s.db.Save(&session).Error; err != nil {
		return nil, fmt.Errorf("failed to update training session: %w", err)
	}

	return s.GetTrainingSessionByID(session.ID)
}

func (s *TrainingService) DeleteTrainingSession(id uuid.UUID) error {
	var session models.TrainingSession
	if err := s.db.First(&session, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("training session not found")
		}
		return fmt.Errorf("failed to fetch training session: %w", err)
	}

	// Check if training can be deleted (only if not started)
	if session.Status == "active" || session.Status == "completed" {
		return fmt.Errorf("cannot delete training session that is active or completed")
	}

	// Start transaction for cascading deletes
	tx := s.db.Begin()

	// Delete training games
	if err := tx.Where("training_session_id = ?", id).Delete(&models.TrainingGame{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete training games: %w", err)
	}

	// Delete training players
	if err := tx.Where("training_session_id = ?", id).Delete(&models.TrainingPlayer{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete training players: %w", err)
	}

	// Delete training session
	if err := tx.Delete(&session).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete training session: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *TrainingService) StartTraining(id uuid.UUID) (*models.TrainingSession, error) {
	session, err := s.GetTrainingSessionByID(id)
	if err != nil {
		return nil, err
	}

	if session.Status != "planned" {
		return nil, fmt.Errorf("training session is not in planned status")
	}

	return s.UpdateTrainingSession(id, &models.TrainingSessionUpdateRequest{
		Status: utils.StringPtr("active"),
	})
}

func (s *TrainingService) FinishTraining(id uuid.UUID) (*models.TrainingSession, error) {
	session, err := s.GetTrainingSessionByID(id)
	if err != nil {
		return nil, err
	}

	if session.Status != "active" {
		return nil, fmt.Errorf("training session is not active")
	}

	return s.UpdateTrainingSession(id, &models.TrainingSessionUpdateRequest{
		Status: utils.StringPtr("completed"),
	})
}


func (s *TrainingService) AddTrainingPlayer(trainingID uuid.UUID, guestName *string) (*models.TrainingPlayer, error) {
	// Check if training exists
	var session models.TrainingSession
	if err := s.db.First(&session, "id = ?", trainingID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("training session not found")
		}
		return nil, fmt.Errorf("failed to fetch training session: %w", err)
	}

	if session.Status != "planned" && session.Status != "active" {
		return nil, fmt.Errorf("cannot add players to completed or cancelled training")
	}

	// Determine if this is a guest player
	isGuest := guestName != nil && *guestName != ""

	var trainingPlayer *models.TrainingPlayer

	if isGuest {
		// Add guest player
		trainingPlayer = &models.TrainingPlayer{
			TrainingSessionID: trainingID,
			GuestName:         guestName,
			IsGuest:           true,
			Attended:          true,
		}
	} else {
		return nil, fmt.Errorf("automatic player assignment is handled during training creation")
	}

	if err := s.db.Create(trainingPlayer).Error; err != nil {
		return nil, fmt.Errorf("failed to add training player: %w", err)
	}

	// Load relationships for response
	var result models.TrainingPlayer
	if err := s.db.Preload("TrainingSession").Preload("Player").First(&result, "id = ?", trainingPlayer.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch created training player: %w", err)
	}

	return &result, nil
}

func (s *TrainingService) RemoveTrainingPlayer(trainingPlayerID uuid.UUID) error {
	var trainingPlayer models.TrainingPlayer
	if err := s.db.First(&trainingPlayer, "id = ?", trainingPlayerID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("training player not found")
		}
		return fmt.Errorf("failed to fetch training player: %w", err)
	}

	// Check if training session allows player removal
	var session models.TrainingSession
	if err := s.db.First(&session, "id = ?", trainingPlayer.TrainingSessionID).Error; err != nil {
		return fmt.Errorf("failed to fetch training session: %w", err)
	}

	if session.Status != "planned" {
		return fmt.Errorf("cannot remove players from active, completed, or cancelled training")
	}

	// Only allow removal of guest players
	if !trainingPlayer.IsGuest {
		return fmt.Errorf("cannot remove regular players, only guests can be removed")
	}

	if err := s.db.Delete(&trainingPlayer).Error; err != nil {
		return fmt.Errorf("failed to remove training player: %w", err)
	}

	return nil
}

func (s *TrainingService) GetTrainingCosts(trainingID uuid.UUID) (*models.TrainingCostsResponse, error) {
	session, err := s.GetTrainingSessionByID(trainingID)
	if err != nil {
		return nil, err
	}

	var playerCosts []models.PlayerCost
	totalCollected := 0.0

	for _, tp := range session.TrainingPlayers {
		if !tp.Attended {
			continue // Skip non-attending players
		}

		cost := session.CostPerPlayer
		gamesPlayed := 0

		// Count games for this player
		for _, game := range session.Games {
			if game.Status == "completed" || game.Status == "playing" {
				// Check if player participated in this game
				if (tp.PlayerID != nil && (*tp.PlayerID == *game.Player1ID || *tp.PlayerID == *game.Player2ID)) ||
					(tp.IsGuest && tp.GuestName != nil &&
						(*tp.GuestName == *game.Guest1Name || *tp.GuestName == *game.Guest2Name)) {
					gamesPlayed++
				}
			}
		}

		// If player didn't participate in any games, cost is 0
		if gamesPlayed == 0 {
			cost = 0
		}

		playerCost := models.PlayerCost{
			TotalCost:   cost,
			GamesPlayed: gamesPlayed,
			IsGuest:     tp.IsGuest,
		}

		if tp.IsGuest {
			playerCost.GuestName = tp.GuestName
		} else {
			playerCost.PlayerID = *tp.PlayerID
			if tp.Player != nil {
				playerName := tp.Player.Name
				playerCost.PlayerName = &playerName
			}
		}

		playerCosts = append(playerCosts, playerCost)
		totalCollected += cost
	}

	return &models.TrainingCostsResponse{
		TrainingSessionID: trainingID,
		PlayerCosts:       playerCosts,
		TotalCollected:    totalCollected,
	}, nil
}

