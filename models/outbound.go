package models

import (
	"time"

	"gorm.io/gorm"
)

type Outbound struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Tracking     string         `gorm:"unique;not null" json:"tracking" example:"SPXID056205885386"`
	UserID       uint           `gorm:"not null" json:"user_id" example:"1"`
	ExpeditionID uint           `gorm:"not null" json:"expedition_id" example:"1"`
	Complained   bool           `gorm:"default:false" json:"complained" example:"false"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Order      *Order      `gorm:"-" json:"order,omitempty"`
	User       *User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Expedition *Expedition `gorm:"foreignKey:ExpeditionID" json:"expedition,omitempty"`
}

type OutboundResponse struct {
	ID           uint      `json:"id"`
	Tracking     string    `json:"tracking"`
	UserID       uint      `json:"user_id"`
	ExpeditionID uint      `json:"expedition_id"`
	Complained   bool      `json:"complained"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Related data
	Order      *OrderResponse      `json:"order,omitempty"`
	User       *UserResponse       `json:"user,omitempty"`
	Expedition *ExpeditionResponse `json:"expedition,omitempty"`
}

// ToOutboundResponse converts Outbound model to OutboundResponse
func (ob *Outbound) ToOutboundResponse() OutboundResponse {
	response := OutboundResponse{
		ID:           ob.ID,
		Tracking:     ob.Tracking,
		UserID:       ob.UserID,
		ExpeditionID: ob.ExpeditionID,
		Complained:   ob.Complained,
		CreatedAt:    ob.CreatedAt,
		UpdatedAt:    ob.UpdatedAt,
	}

	// Include order data if loaded
	if ob.Order != nil {
		orderResponse := ob.Order.ToOrderResponse()
		response.Order = &orderResponse
	}

	// Include user data if loaded
	if ob.User != nil {
		userResponse := ob.User.ToUserResponse()
		response.User = &userResponse
	}

	// Include expedition data if loaded
	if ob.Expedition != nil {
		expeditionResponse := ob.Expedition.ToExpeditionResponse()
		response.Expedition = &expeditionResponse
	}

	return response
}
