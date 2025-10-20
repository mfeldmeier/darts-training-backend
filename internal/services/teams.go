package services

import (
	"fmt"
	"darts-training-app/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TeamService struct {
	db *gorm.DB
}

func NewTeamService(db *gorm.DB) *TeamService {
	return &TeamService{
		db: db,
	}
}

func (s *TeamService) GetAllTeams() ([]models.Team, error) {
	var teams []models.Team
	err := s.db.Preload("Players").Find(&teams).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch teams: %w", err)
	}
	return teams, nil
}

func (s *TeamService) GetTeamByID(id uuid.UUID) (*models.Team, error) {
	var team models.Team
	err := s.db.Preload("Players").First(&team, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("team not found")
		}
		return nil, fmt.Errorf("failed to fetch team: %w", err)
	}
	return &team, nil
}

func (s *TeamService) CreateTeam(req *models.TeamCreateRequest) (*models.Team, error) {
	// Check if team name already exists
	var existingTeam models.Team
	err := s.db.Where("name = ?", req.Name).First(&existingTeam).Error
	if err == nil {
		return nil, fmt.Errorf("team with name '%s' already exists", req.Name)
	}
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing team: %w", err)
	}

	team := &models.Team{
		Name:    req.Name,
		LogoURL: req.LogoURL,
	}

	if err := s.db.Create(team).Error; err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	return team, nil
}

func (s *TeamService) UpdateTeam(id uuid.UUID, req *models.TeamUpdateRequest) (*models.Team, error) {
	var team models.Team
	if err := s.db.First(&team, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("team not found")
		}
		return nil, fmt.Errorf("failed to fetch team: %w", err)
	}

	// Check if team name already exists (if updating name)
	if req.Name != nil && *req.Name != team.Name {
		var existingTeam models.Team
		err := s.db.Where("name = ? AND id != ?", *req.Name, id).First(&existingTeam).Error
		if err == nil {
			return nil, fmt.Errorf("team with name '%s' already exists", *req.Name)
		}
		if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("failed to check existing team: %w", err)
		}
		team.Name = *req.Name
	}

	if req.LogoURL != nil {
		team.LogoURL = req.LogoURL
	}

	if err := s.db.Save(&team).Error; err != nil {
		return nil, fmt.Errorf("failed to update team: %w", err)
	}

	return &team, nil
}

func (s *TeamService) DeleteTeam(id uuid.UUID) error {
	var team models.Team
	if err := s.db.First(&team, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("team not found")
		}
		return fmt.Errorf("failed to fetch team: %w", err)
	}

	// Check if team has players
	playerCount := int64(0)
	s.db.Model(&models.Player{}).Where("team_id = ?", id).Count(&playerCount)
	if playerCount > 0 {
		return fmt.Errorf("cannot delete team with %d players. Please reassign or delete players first", playerCount)
	}

	if err := s.db.Delete(&team).Error; err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}

	return nil
}

func (s *TeamService) GetTeamPlayers(id uuid.UUID) ([]models.Player, error) {
	// Check if team exists
	var team models.Team
	if err := s.db.First(&team, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("team not found")
		}
		return nil, fmt.Errorf("failed to fetch team: %w", err)
	}

	var players []models.Player
	err := s.db.Where("team_id = ?", id).Find(&players).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch team players: %w", err)
	}

	return players, nil
}