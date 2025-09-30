package models

import (
	"time"

	"gorm.io/gorm"
)

type MbRibbon struct {
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

type MbRibbonResponse struct {
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

// ToMbRibbonResponse converts MbRibbon model to MbRibbonResponse
func (mbr *MbRibbon) ToMbRibbonResponse() MbRibbonResponse {
	response := MbRibbonResponse{
		ID:         mbr.ID,
		Tracking:   mbr.Tracking,
		UserID:     mbr.UserID,
		Complained: mbr.Complained,
		CreatedAt:  mbr.CreatedAt,
		UpdatedAt:  mbr.UpdatedAt,
	}

	// Include order data if loaded
	if mbr.Order != nil {
		orderResponse := mbr.Order.ToOrderResponse()
		response.Order = &orderResponse
	}

	// Include user data if loaded
	if mbr.User != nil {
		userResponse := mbr.User.ToUserResponse()
		response.User = &userResponse
	}

	return response
}
