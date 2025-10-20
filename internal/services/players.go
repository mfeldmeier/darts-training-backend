package services

import (
	"fmt"
	"darts-training-app/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PlayerService struct {
	db *gorm.DB
}

func NewPlayerService(db *gorm.DB) *PlayerService {
	return &PlayerService{
		db: db,
	}
}

func (s *PlayerService) GetAllPlayers() ([]models.Player, error) {
	var players []models.Player
	err := s.db.Preload("Team").Find(&players).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch players: %w", err)
	}
	return players, nil
}

func (s *PlayerService) GetPlayersByTeam(teamID uuid.UUID) ([]models.Player, error) {
	var players []models.Player
	err := s.db.Preload("Team").Where("team_id = ?", teamID).Find(&players).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch team players: %w", err)
	}
	return players, nil
}

func (s *PlayerService) GetPlayerByID(id uuid.UUID) (*models.Player, error) {
	var player models.Player
	err := s.db.Preload("Team").First(&player, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("player not found")
		}
		return nil, fmt.Errorf("failed to fetch player: %w", err)
	}
	return &player, nil
}

func (s *PlayerService) GetPlayerByAuth0ID(auth0UserID string) (*models.Player, error) {
	var player models.Player
	err := s.db.Preload("Team").Where("auth0_user_id = ?", auth0UserID).First(&player).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("player not found")
		}
		return nil, fmt.Errorf("failed to fetch player: %w", err)
	}
	return &player, nil
}

func (s *PlayerService) CreatePlayer(req *models.PlayerCreateRequest) (*models.Player, error) {
	// Check if email already exists
	var existingPlayer models.Player
	err := s.db.Where("email = ?", req.Email).First(&existingPlayer).Error
	if err == nil {
		return nil, fmt.Errorf("player with email '%s' already exists", req.Email)
	}
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing player: %w", err)
	}

	// Validate team ID if provided
	var teamID *uuid.UUID
	if req.TeamID != nil && *req.TeamID != "" {
		parsedTeamID, err := uuid.Parse(*req.TeamID)
		if err != nil {
			return nil, fmt.Errorf("invalid team ID format: %w", err)
		}

		// Check if team exists
		var team models.Team
		if err := s.db.First(&team, "id = ?", parsedTeamID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, fmt.Errorf("team not found")
			}
			return nil, fmt.Errorf("failed to validate team: %w", err)
		}
		teamID = &parsedTeamID
	}

	player := &models.Player{
		Name:      req.Name,
		Email:     req.Email,
		Nickname:  req.Nickname,
		IsCaptain: req.IsCaptain,
		IsActive:  req.IsActive,
		TeamID:    teamID,
	}

	if err := s.db.Create(player).Error; err != nil {
		return nil, fmt.Errorf("failed to create player: %w", err)
	}

	return player, nil
}

func (s *PlayerService) UpdatePlayer(id uuid.UUID, req *models.PlayerUpdateRequest) (*models.Player, error) {
	var player models.Player
	if err := s.db.First(&player, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("player not found")
		}
		return nil, fmt.Errorf("failed to fetch player: %w", err)
	}

	// Check if email already exists (if updating email)
	if req.Email != nil && *req.Email != player.Email {
		var existingPlayer models.Player
		err := s.db.Where("email = ? AND id != ?", *req.Email, id).First(&existingPlayer).Error
		if err == nil {
			return nil, fmt.Errorf("player with email '%s' already exists", *req.Email)
		}
		if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("failed to check existing player: %w", err)
		}
		player.Email = *req.Email
	}

	// Update fields if provided
	if req.Name != nil {
		player.Name = *req.Name
	}
	if req.Nickname != nil {
		player.Nickname = req.Nickname
	}
	if req.IsCaptain != nil {
		player.IsCaptain = *req.IsCaptain
	}
	if req.IsActive != nil {
		player.IsActive = *req.IsActive
	}

	// Handle team assignment
	if req.TeamID != nil {
		if *req.TeamID == "" {
			// Remove from team
			player.TeamID = nil
		} else {
			// Parse and validate new team ID
			parsedTeamID, err := uuid.Parse(*req.TeamID)
			if err != nil {
				return nil, fmt.Errorf("invalid team ID format: %w", err)
			}

			// Check if team exists
			var team models.Team
			if err := s.db.First(&team, "id = ?", parsedTeamID).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return nil, fmt.Errorf("team not found")
				}
				return nil, fmt.Errorf("failed to validate team: %w", err)
			}
			player.TeamID = &parsedTeamID
		}
	}

	if err := s.db.Save(&player).Error; err != nil {
		return nil, fmt.Errorf("failed to update player: %w", err)
	}

	return &player, nil
}

func (s *PlayerService) UpdatePlayerAuth0ID(id uuid.UUID, auth0UserID string) error {
	var player models.Player
	if err := s.db.First(&player, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("player not found")
		}
		return fmt.Errorf("failed to fetch player: %w", err)
	}

	// Check if Auth0 ID already exists for another player
	var existingPlayer models.Player
	err := s.db.Where("auth0_user_id = ? AND id != ?", auth0UserID, id).First(&existingPlayer).Error
	if err == nil {
		return fmt.Errorf("Auth0 user ID is already linked to another player")
	}
	if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check existing Auth0 ID: %w", err)
	}

	player.Auth0UserID = &auth0UserID
	if err := s.db.Save(&player).Error; err != nil {
		return fmt.Errorf("failed to update player Auth0 ID: %w", err)
	}

	return nil
}

func (s *PlayerService) DeletePlayer(id uuid.UUID) error {
	var player models.Player
	if err := s.db.First(&player, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("player not found")
		}
		return fmt.Errorf("failed to fetch player: %w", err)
	}

	// Check for related records that would prevent deletion
	// Check training sessions created by this player
	trainingCount := int64(0)
	s.db.Model(&models.TrainingSession{}).Where("created_by = ?", id).Count(&trainingCount)
	if trainingCount > 0 {
		return fmt.Errorf("cannot delete player who created %d training sessions", trainingCount)
	}

	// Check for training games where this player participated
	gameCount := int64(0)
	s.db.Model(&models.TrainingGame{}).
		Where("player1_id = ? OR player2_id = ?", id, id).
		Count(&gameCount)
	if gameCount > 0 {
		return fmt.Errorf("cannot delete player who participated in %d games", gameCount)
	}

	if err := s.db.Delete(&player).Error; err != nil {
		return fmt.Errorf("failed to delete player: %w", err)
	}

	return nil
}

func (s *PlayerService) FindOrCreatePlayerByAuth0(auth0User models.Auth0User) (*models.Player, error) {
	// First try to find existing player by Auth0 ID
	var player models.Player
	err := s.db.Preload("Team").Where("auth0_user_id = ?", auth0User.Sub).First(&player).Error
	if err == nil {
		return &player, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to search for player: %w", err)
	}

	// If not found by Auth0 ID, try to find by email
	err = s.db.Preload("Team").Where("email = ?", auth0User.Email).First(&player).Error
	if err == nil {
		// Player found by email, update with Auth0 ID
		if err := s.UpdatePlayerAuth0ID(player.ID, auth0User.Sub); err != nil {
			return nil, fmt.Errorf("failed to update player with Auth0 ID: %w", err)
		}
		// Refresh player data
		return s.GetPlayerByID(player.ID)
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to search for player by email: %w", err)
	}

	// Player doesn't exist, create new one
	createReq := &models.PlayerCreateRequest{
		Name:     auth0User.Name,
		Email:    auth0User.Email,
		Nickname: &auth0User.Nickname,
	}

	newPlayer, err := s.CreatePlayer(createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create new player: %w", err)
	}

	// Update with Auth0 ID
	if err := s.UpdatePlayerAuth0ID(newPlayer.ID, auth0User.Sub); err != nil {
		return nil, fmt.Errorf("failed to link new player to Auth0: %w", err)
	}

	return s.GetPlayerByID(newPlayer.ID)
}

// ActivatePlayer activates a player
func (s *PlayerService) ActivatePlayer(id uuid.UUID) error {
	var player models.Player
	if err := s.db.First(&player, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("player not found")
		}
		return fmt.Errorf("failed to fetch player: %w", err)
	}

	player.IsActive = true
	if err := s.db.Save(&player).Error; err != nil {
		return fmt.Errorf("failed to activate player: %w", err)
	}

	return nil
}

// DeactivatePlayer deactivates a player
func (s *PlayerService) DeactivatePlayer(id uuid.UUID) error {
	var player models.Player
	if err := s.db.First(&player, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("player not found")
		}
		return fmt.Errorf("failed to fetch player: %w", err)
	}

	player.IsActive = false
	if err := s.db.Save(&player).Error; err != nil {
		return fmt.Errorf("failed to deactivate player: %w", err)
	}

	return nil
}

// GetActivePlayers returns only active players
func (s *PlayerService) GetActivePlayers() ([]models.Player, error) {
	var players []models.Player
	err := s.db.Preload("Team").Where("is_active = ?", true).Find(&players).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch active players: %w", err)
	}
	return players, nil
}

// GetActivePlayersByTeam returns only active players from a specific team
func (s *PlayerService) GetActivePlayersByTeam(teamID uuid.UUID) ([]models.Player, error) {
	var players []models.Player
	err := s.db.Preload("Team").Where("team_id = ? AND is_active = ?", teamID, true).Find(&players).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch active team players: %w", err)
	}
	return players, nil
}