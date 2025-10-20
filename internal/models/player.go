package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Player struct {
	ID           uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name         string     `gorm:"not null" json:"name"`
	Email        string     `gorm:"uniqueIndex;not null" json:"email"`
	Nickname     *string    `json:"nickname"`
	IsCaptain    bool       `gorm:"default:false" json:"is_captain"`
	Auth0UserID  *string    `gorm:"uniqueIndex" json:"auth0_user_id"`
	TeamID       *uuid.UUID `json:"team_id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	Team *Team `gorm:"foreignKey:TeamID" json:"team,omitempty"`

	// Relationships
	CreatedTrainings []TrainingSession `gorm:"foreignKey:CreatedBy" json:"-"`
	TrainingPlayers  []TrainingPlayer `gorm:"foreignKey:PlayerID" json:"-"`
	Player1Games     []TrainingGame   `gorm:"foreignKey:Player1ID" json:"-"`
	Player2Games     []TrainingGame   `gorm:"foreignKey:Player2ID" json:"-"`
}

type PlayerCreateRequest struct {
	Name      string  `json:"name" binding:"required,min=1,max=100"`
	Email     string  `json:"email" binding:"required,email"`
	Nickname  *string `json:"nickname"`
	IsCaptain bool    `json:"is_captain"`
	TeamID    *string `json:"team_id"`
}

type PlayerUpdateRequest struct {
	Name      *string `json:"name"`
	Email     *string `json:"email"`
	Nickname  *string `json:"nickname"`
	IsCaptain *bool   `json:"is_captain"`
	TeamID    *string `json:"team_id"`
}

type PlayerResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Nickname    *string   `json:"nickname"`
	IsCaptain   bool      `json:"is_captain"`
	TeamID      *uuid.UUID `json:"team_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	TeamName    *string   `json:"team_name,omitempty"`
}

type PlayerWithTeamResponse struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	Nickname    *string    `json:"nickname"`
	IsCaptain   bool       `json:"is_captain"`
	Team        *Team      `json:"team,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (p *Player) ToResponse() PlayerResponse {
	var teamName *string
	if p.Team != nil {
		teamName = &p.Team.Name
	}

	return PlayerResponse{
		ID:        p.ID,
		Name:      p.Name,
		Email:     p.Email,
		Nickname:  p.Nickname,
		IsCaptain: p.IsCaptain,
		TeamID:    p.TeamID,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
		TeamName:  teamName,
	}
}

func (p *Player) ToResponseWithTeam() PlayerWithTeamResponse {
	return PlayerWithTeamResponse{
		ID:        p.ID,
		Name:      p.Name,
		Email:     p.Email,
		Nickname:  p.Nickname,
		IsCaptain: p.IsCaptain,
		Team:      p.Team,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}