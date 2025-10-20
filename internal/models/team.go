package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Team struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name      string    `gorm:"uniqueIndex;not null" json:"name"`
	LogoURL   *string   `json:"logo_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Players []Player `gorm:"foreignKey:TeamID" json:"players,omitempty"`
}

type TeamCreateRequest struct {
	Name    string  `json:"name" binding:"required,min=1,max=100"`
	LogoURL *string `json:"logo_url"`
}

type TeamUpdateRequest struct {
	Name    *string `json:"name"`
	LogoURL *string `json:"logo_url"`
}

type TeamResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	LogoURL   *string   `json:"logo_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	PlayerCount int    `json:"player_count"`
}

func (t *Team) ToResponse() TeamResponse {
	return TeamResponse{
		ID:        t.ID,
		Name:      t.Name,
		LogoURL:   t.LogoURL,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
		PlayerCount: len(t.Players),
	}
}