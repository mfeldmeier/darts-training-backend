package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TrainingSession struct {
	ID            uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name          string     `gorm:"not null" json:"name"`
	Description   *string    `json:"description"`
	TrainingDate  time.Time  `gorm:"not null" json:"training_date"`
	CostPerPlayer float64    `gorm:"default:5.00" json:"cost_per_player"`
	Status        string     `gorm:"default:'planned'" json:"status"` // planned, active, completed, cancelled
	CreatedBy     *uuid.UUID `json:"created_by"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Creator          *Player          `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	TrainingPlayers  []TrainingPlayer `gorm:"foreignKey:TrainingSessionID" json:"training_players,omitempty"`
	Games            []TrainingGame   `gorm:"foreignKey:TrainingSessionID" json:"games,omitempty"`
}

type TrainingPlayer struct {
	ID                 uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	TrainingSessionID  uuid.UUID `json:"training_session_id"`
	PlayerID           *uuid.UUID `json:"player_id"`
	GuestName          *string   `json:"guest_name"`
	IsGuest            bool      `gorm:"default:false" json:"is_guest"`
	Attended           bool      `gorm:"default:true" json:"attended"`
	CreatedAt          time.Time `json:"created_at"`

	// Relationships
	TrainingSession *TrainingSession `gorm:"foreignKey:TrainingSessionID" json:"training_session,omitempty"`
	Player          *Player          `gorm:"foreignKey:PlayerID" json:"player,omitempty"`
}

type GameMode struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name        string    `gorm:"not null" json:"name"`
	Description *string   `json:"description"`
	Rules       string    `gorm:"type:jsonb" json:"rules"` // JSONB for flexible rules
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Games []TrainingGame `gorm:"foreignKey:GameModeID" json:"games,omitempty"`
}

type TrainingGame struct {
	ID                 uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	TrainingSessionID  uuid.UUID  `json:"training_session_id"`
	GameModeID         uuid.UUID  `json:"game_mode_id"`
	Player1ID          *uuid.UUID `json:"player1_id"`
	Player2ID          *uuid.UUID `json:"player2_id"`
	Guest1Name         *string    `json:"guest1_name"`
	Guest2Name         *string    `json:"guest2_name"`
	Player1Score       int        `gorm:"default:0" json:"player1_score"`
	Player2Score       int        `gorm:"default:0" json:"player2_score"`
	Status             string     `gorm:"default:'pending'" json:"status"` // pending, playing, completed, cancelled
	Winner             *string    `json:"winner"` // 'player1', 'player2', 'draw'
	CompletedAt        *time.Time `json:"completed_at"`
	CreatedAt          time.Time  `json:"created_at"`

	// Relationships
	TrainingSession *TrainingSession `gorm:"foreignKey:TrainingSessionID" json:"training_session,omitempty"`
	GameMode        *GameMode        `gorm:"foreignKey:GameModeID" json:"game_mode,omitempty"`
	Player1         *Player          `gorm:"foreignKey:Player1ID" json:"player1,omitempty"`
	Player2         *Player          `gorm:"foreignKey:Player2ID" json:"player2,omitempty"`
}

// DTOs and Request/Response structures
type TrainingSessionCreateRequest struct {
	Name          string    `json:"name" binding:"required,min=1,max=200"`
	Description   *string   `json:"description"`
	TrainingDate  time.Time `json:"training_date" binding:"required"`
	CostPerPlayer *float64  `json:"cost_per_player"`
}

type TrainingSessionUpdateRequest struct {
	Name          *string    `json:"name"`
	Description   *string    `json:"description"`
	TrainingDate  *time.Time `json:"training_date"`
	CostPerPlayer *float64  `json:"cost_per_player"`
	Status        *string    `json:"status"`
}

type TrainingSessionResponse struct {
	ID               uuid.UUID               `json:"id"`
	Name             string                  `json:"name"`
	Description      *string                 `json:"description"`
	TrainingDate     time.Time               `json:"training_date"`
	CostPerPlayer    float64                 `json:"cost_per_player"`
	Status           string                  `json:"status"`
	CreatedBy        *uuid.UUID              `json:"created_by"`
	CreatedAt        time.Time               `json:"created_at"`
	UpdatedAt        time.Time               `json:"updated_at"`
	PlayerCount      int                     `json:"player_count"`
	GameCount        int                     `json:"game_count"`
	CreatorName      *string                 `json:"creator_name,omitempty"`
	TrainingPlayers  []TrainingPlayerResponse `json:"training_players,omitempty"`
	Games            []TrainingGameResponse   `json:"games,omitempty"`
}

type TrainingPlayerResponse struct {
	ID                uuid.UUID  `json:"id"`
	TrainingSessionID uuid.UUID  `json:"training_session_id"`
	PlayerID          *uuid.UUID `json:"player_id"`
	GuestName         *string    `json:"guest_name"`
	IsGuest           bool       `json:"is_guest"`
	Attended          bool       `json:"attended"`
	CreatedAt         time.Time  `json:"created_at"`
	PlayerName        *string    `json:"player_name,omitempty"`
}

type TrainingGameResponse struct {
	ID                uuid.UUID  `json:"id"`
	TrainingSessionID uuid.UUID  `json:"training_session_id"`
	GameModeID        uuid.UUID  `json:"game_mode_id"`
	Player1ID         *uuid.UUID `json:"player1_id"`
	Player2ID         *uuid.UUID `json:"player2_id"`
	Guest1Name        *string    `json:"guest1_name"`
	Guest2Name        *string    `json:"guest2_name"`
	Player1Score      int        `json:"player1_score"`
	Player2Score      int        `json:"player2_score"`
	Status            string     `json:"status"`
	Winner            *string    `json:"winner"`
	CompletedAt       *time.Time `json:"completed_at"`
	CreatedAt         time.Time  `json:"created_at"`
	GameModeName      *string    `json:"game_mode_name,omitempty"`
	Player1Name       *string    `json:"player1_name,omitempty"`
	Player2Name       *string    `json:"player2_name,omitempty"`
}

type GameModeResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Rules       string    `json:"rules"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PlayerCost struct {
	PlayerID   uuid.UUID  `json:"player_id"`
	PlayerName *string    `json:"player_name"`
	GuestName  *string    `json:"guest_name"`
	IsGuest    bool       `json:"is_guest"`
	TotalCost  float64    `json:"total_cost"`
	GamesPlayed int       `json:"games_played"`
}

type TrainingCostsResponse struct {
	TrainingSessionID uuid.UUID   `json:"training_session_id"`
	PlayerCosts       []PlayerCost `json:"player_costs"`
	TotalCollected    float64     `json:"total_collected"`
}

func (t *TrainingSession) ToResponse() TrainingSessionResponse {
	playerCount := len(t.TrainingPlayers)
	gameCount := len(t.Games)

	var creatorName *string
	if t.Creator != nil {
		creatorName = &t.Creator.Name
	}

	trainingPlayers := make([]TrainingPlayerResponse, len(t.TrainingPlayers))
	for i, tp := range t.TrainingPlayers {
		trainingPlayers[i] = tp.ToResponse()
	}

	games := make([]TrainingGameResponse, len(t.Games))
	for i, g := range t.Games {
		games[i] = g.ToResponse()
	}

	return TrainingSessionResponse{
		ID:               t.ID,
		Name:             t.Name,
		Description:      t.Description,
		TrainingDate:     t.TrainingDate,
		CostPerPlayer:    t.CostPerPlayer,
		Status:           t.Status,
		CreatedBy:        t.CreatedBy,
		CreatedAt:        t.CreatedAt,
		UpdatedAt:        t.UpdatedAt,
		PlayerCount:      playerCount,
		GameCount:        gameCount,
		CreatorName:      creatorName,
		TrainingPlayers:  trainingPlayers,
		Games:            games,
	}
}

func (tp *TrainingPlayer) ToResponse() TrainingPlayerResponse {
	var playerName *string
	if tp.Player != nil {
		playerName = &tp.Player.Name
	}

	return TrainingPlayerResponse{
		ID:                tp.ID,
		TrainingSessionID: tp.TrainingSessionID,
		PlayerID:          tp.PlayerID,
		GuestName:         tp.GuestName,
		IsGuest:           tp.IsGuest,
		Attended:          tp.Attended,
		CreatedAt:         tp.CreatedAt,
		PlayerName:        playerName,
	}
}

func (g *TrainingGame) ToResponse() TrainingGameResponse {
	var gameModeName, player1Name, player2Name *string

	if g.GameMode != nil {
		gameModeName = &g.GameMode.Name
	}
	if g.Player1 != nil {
		player1Name = &g.Player1.Name
	}
	if g.Player2 != nil {
		player2Name = &g.Player2.Name
	}

	return TrainingGameResponse{
		ID:                g.ID,
		TrainingSessionID: g.TrainingSessionID,
		GameModeID:        g.GameModeID,
		Player1ID:         g.Player1ID,
		Player2ID:         g.Player2ID,
		Guest1Name:        g.Guest1Name,
		Guest2Name:        g.Guest2Name,
		Player1Score:      g.Player1Score,
		Player2Score:      g.Player2Score,
		Status:            g.Status,
		Winner:            g.Winner,
		CompletedAt:       g.CompletedAt,
		CreatedAt:         g.CreatedAt,
		GameModeName:      gameModeName,
		Player1Name:       player1Name,
		Player2Name:       player2Name,
	}
}

func (gm *GameMode) ToResponse() GameModeResponse {
	return GameModeResponse{
		ID:          gm.ID,
		Name:        gm.Name,
		Description: gm.Description,
		Rules:       gm.Rules,
		IsActive:    gm.IsActive,
		CreatedAt:   gm.CreatedAt,
		UpdatedAt:   gm.UpdatedAt,
	}
}