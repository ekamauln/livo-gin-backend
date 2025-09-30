package models

import (
	"time"

	"gorm.io/gorm"
)

type MbOnline struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Tracking   string         `gorm:"unique;not null" json:"tracking" example:"SPXID056205885386"`
	UserID     uint           `gorm:"not null" json:"user_id" example:"1"`
	Complained bool           `gorm:"not null" json:"complained" example:"false" default:"false"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Order *Order `gorm:"-" json:"order,omitempty"`
	User  *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type MbOnlineResponse struct {
	ID         uint      `json:"id"`
	Tracking   string    `json:"tracking"`
	UserID     uint      `json:"user_id"`
	Complained bool      `json:"complained"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Related data
	Order *OrderResponse `json:"order,omitempty"`
	User  *UserResponse  `json:"user,omitempty"`
}

// ToMbOnlineResponse converts MbOnline model to MbOnlineResponse
func (mbo *MbOnline) ToMbOnlineResponse() MbOnlineResponse {
	response := MbOnlineResponse{
		ID:         mbo.ID,
		Tracking:   mbo.Tracking,
		UserID:     mbo.UserID,
		Complained: mbo.Complained,
		CreatedAt:  mbo.CreatedAt,
		UpdatedAt:  mbo.UpdatedAt,
	}

	// Include order data if loaded
	if mbo.Order != nil {
		orderResponse := mbo.Order.ToOrderResponse()
		response.Order = &orderResponse
	}

	// Include user data if loaded
	if mbo.User != nil {
		userResponse := mbo.User.ToUserResponse()
		response.User = &userResponse
	}

	return response
}
