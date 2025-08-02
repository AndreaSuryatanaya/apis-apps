package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
	Name     string    `json:"name" gorm:"not null"`
	Username string    `json:"username" gorm:"unique;not null"`
	Password string    `json:"password,omitempty" gorm:"column:password;not null"`
}

type Task struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	Todo      string    `json:"todo" gorm:"not null"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	User      User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
}
type Position struct {
	ID   uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
	Name string    `json:"name" gorm:"unique;not null"`
}

type UserPosition struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
	UserID     uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	PositionID uuid.UUID `json:"position_id" gorm:"type:uuid;not null"`
	User       User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Position   Position  `json:"position,omitempty" gorm:"foreignKey:PositionID"`
}

// BeforeCreate hook to generate UUID if not set
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

func (p *Position) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

func (up *UserPosition) BeforeCreate(tx *gorm.DB) error {
	if up.ID == uuid.Nil {
		up.ID = uuid.New()
	}
	return nil
}
