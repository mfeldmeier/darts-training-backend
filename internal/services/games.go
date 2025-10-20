package services

import (
	"fmt"
	"time"

	"darts-training-app/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GameService struct {
	db *gorm.DB
}

func NewGameService(db *gorm.DB) *GameService {
	return &GameService{
		db: db,
	}
}

func (s *GameService) GetAllGameModes() ([]models.GameMode, error) {
	var gameModes []models.GameMode
	err := s.db.Where("is_active = ?", true).Find(&gameModes).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch game modes: %w", err)
	}
	return gameModes, nil
}

func (s *GameService) GetGamesByTrainingSession(trainingSessionID uuid.UUID) ([]models.TrainingGame, error) {
	var games []models.TrainingGame
	err := s.db.Preload("GameMode").
		Preload("Player1").
		Preload("Player2").
		Where("training_session_id = ?", trainingSessionID).
		Order("created_at").
		Find(&games).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch training games: %w", err)
	}
	return games, nil
}

func (s *GameService) CreateGame(trainingSessionID uuid.UUID, gameModeID uuid.UUID, player1ID, player2ID *uuid.UUID, guest1Name, guest2Name *string) (*models.TrainingGame, error) {
	// Validate training session
	var session models.TrainingSession
	if err := s.db.First(&session, "id = ?", trainingSessionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("training session not found")
		}
		return nil, fmt.Errorf("failed to fetch training session: %w", err)
	}

	if session.Status != "active" {
		return nil, fmt.Errorf("cannot create games for training session that is not active")
	}

	// Validate game mode
	var gameMode models.GameMode
	if err := s.db.First(&gameMode, "id = ?", gameModeID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("game mode not found")
		}
		return nil, fmt.Errorf("failed to fetch game mode: %w", err)
	}

	// Validate players
	if player1ID != nil {
		var player1 models.Player
		if err := s.db.First(&player1, "id = ?", *player1ID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, fmt.Errorf("player 1 not found")
			}
			return nil, fmt.Errorf("failed to fetch player 1: %w", err)
		}
	}

	if player2ID != nil {
		var player2 models.Player
		if err := s.db.First(&player2, "id = ?", *player2ID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, fmt.Errorf("player 2 not found")
			}
			return nil, fmt.Errorf("failed to fetch player 2: %w", err)
		}
	}

	// Validate that we have exactly two players (regular or guest)
	playerCount := 0
	if player1ID != nil || guest1Name != nil {
		playerCount++
	}
	if player2ID != nil || guest2Name != nil {
		playerCount++
	}

	if playerCount != 2 {
		return nil, fmt.Errorf("exactly two players are required for each game")
	}

	game := &models.TrainingGame{
		TrainingSessionID: trainingSessionID,
		GameModeID:        gameModeID,
		Player1ID:         player1ID,
		Player2ID:         player2ID,
		Guest1Name:        guest1Name,
		Guest2Name:        guest2Name,
		Status:            "pending",
	}

	if err := s.db.Create(game).Error; err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	// Load relationships for response
	var result models.TrainingGame
	err := s.db.Preload("GameMode").
		Preload("Player1").
		Preload("Player2").
		First(&result, "id = ?", game.ID).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created game: %w", err)
	}

	return &result, nil
}

func (s *GameService) UpdateGame(id uuid.UUID, player1Score, player2Score *int, status *string, winner *string) (*models.TrainingGame, error) {
	var game models.TrainingGame
	if err := s.db.First(&game, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("game not found")
		}
		return nil, fmt.Errorf("failed to fetch game: %w", err)
	}

	// Update fields if provided
	if player1Score != nil {
		game.Player1Score = *player1Score
	}
	if player2Score != nil {
		game.Player2Score = *player2Score
	}
	if status != nil {
		// Validate status
		validStatuses := []string{"pending", "playing", "completed", "cancelled"}
		statusValid := false
		for _, validStatus := range validStatuses {
			if *status == validStatus {
				statusValid = true
				break
			}
		}
		if !statusValid {
			return nil, fmt.Errorf("invalid status: %s", *status)
		}
		game.Status = *status
	}

	if winner != nil {
		// Validate winner
		validWinners := []string{"player1", "player2", "draw"}
		winnerValid := false
		for _, validWinner := range validWinners {
			if *winner == validWinner {
				winnerValid = true
				break
			}
		}
		if !winnerValid {
			return nil, fmt.Errorf("invalid winner: %s", *winner)
		}
		game.Winner = winner
	}

	// Set completion time if game is completed
	if status != nil && *status == "completed" && game.CompletedAt == nil {
		now := time.Now()
		game.CompletedAt = &now
	}

	// Auto-determine winner if scores are set and status is completed
	if status != nil && *status == "completed" && winner == nil {
		if game.Player1Score > game.Player2Score {
			game.Winner = stringPtr("player1")
		} else if game.Player2Score > game.Player1Score {
			game.Winner = stringPtr("player2")
		} else {
			game.Winner = stringPtr("draw")
		}
	}

	if err := s.db.Save(&game).Error; err != nil {
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	// Load relationships for response
	var result models.TrainingGame
	err := s.db.Preload("GameMode").
		Preload("Player1").
		Preload("Player2").
		First(&result, "id = ?", game.ID).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated game: %w", err)
	}

	return &result, nil
}

func (s *GameService) DeleteGame(id uuid.UUID) error {
	var game models.TrainingGame
	if err := s.db.First(&game, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("game not found")
		}
		return fmt.Errorf("failed to fetch game: %w", err)
	}

	// Check if game can be deleted
	if game.Status == "completed" || game.Status == "playing" {
		return fmt.Errorf("cannot delete game that is in progress or completed")
	}

	if err := s.db.Delete(&game).Error; err != nil {
		return fmt.Errorf("failed to delete game: %w", err)
	}

	return nil
}

func (s *GameService) GenerateGamesForTraining(trainingSessionID uuid.UUID, gameModeID uuid.UUID) ([]models.TrainingGame, error) {
	// Get training session with players
	var session models.TrainingSession
	err := s.db.Preload("TrainingPlayers.Player").First(&session, "id = ?", trainingSessionID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("training session not found")
		}
		return nil, fmt.Errorf("failed to fetch training session: %w", err)
	}

	if session.Status != "planned" {
		return nil, fmt.Errorf("can only generate games for planned training sessions")
	}

	// Filter attending players
	var attendingPlayers []models.TrainingPlayer
	for _, tp := range session.TrainingPlayers {
		if tp.Attended {
			attendingPlayers = append(attendingPlayers, tp)
		}
	}

	if len(attendingPlayers) < 2 {
		return nil, fmt.Errorf("need at least 2 attending players to generate games")
	}

	var games []models.TrainingGame

	// Simple round-robin pairing
	for i := 0; i < len(attendingPlayers); i++ {
		for j := i + 1; j < len(attendingPlayers); j++ {
			player1 := attendingPlayers[i]
			player2 := attendingPlayers[j]

			var player1ID, player2ID *uuid.UUID
			var guest1Name, guest2Name *string

			if player1.IsGuest {
				guest1Name = player1.GuestName
			} else {
				player1ID = player1.PlayerID
			}

			if player2.IsGuest {
				guest2Name = player2.GuestName
			} else {
				player2ID = player2.PlayerID
			}

			game := &models.TrainingGame{
				TrainingSessionID: trainingSessionID,
				GameModeID:        gameModeID,
				Player1ID:         player1ID,
				Player2ID:         player2ID,
				Guest1Name:        guest1Name,
				Guest2Name:        guest2Name,
				Status:            "pending",
			}

			if err := s.db.Create(game).Error; err != nil {
				return nil, fmt.Errorf("failed to create game: %w", err)
			}

			games = append(games, *game)
		}
	}

	// Load relationships for response
	for i := range games {
		err := s.db.Preload("GameMode").
			Preload("Player1").
			Preload("Player2").
			First(&games[i], "id = ?", games[i].ID).Error
		if err != nil {
			return nil, fmt.Errorf("failed to fetch created game: %w", err)
		}
	}

	return games, nil
}

// Helper function
func stringPtr(s string) *string {
	return &s
}