package models

import (
	"time"

	"gorm.io/gorm"
)

type PcOnline struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Tracking   string         `gorm:"unique;not null" json:"tracking" example:"PC1234567890"`
	UserID     *uint          `gorm:"default:null" json:"user_id"`
	Complained bool           `gorm:"default:false" json:"complained"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// One-to-many relationship with PcOnlineDetail
	Details []PcOnlineDetail `gorm:"foreignKey:PcOnlineID" json:"details"`

	// Relations
	Order *Order `gorm:"-" json:"order,omitempty"`
	User  *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type PcOnlineDetail struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	PcOnlineID uint           `gorm:"not null" json:"pc_online_id"`
	BoxID      uint           `gorm:"not null" json:"box_id"`
	Quantity   int            `gorm:"not null" json:"quantity"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	PcOnline PcOnline `gorm:"foreignKey:PcOnlineID" json:"-"`
	Box      *Box     `gorm:"foreignKey:BoxID" json:"box,omitempty"`
}

// Response structures
type PcOnlineDetailResponse struct {
	ID         uint        `json:"id"`
	PcOnlineID uint        `json:"pc_online_id"`
	BoxID      uint        `json:"box_id"`
	Quantity   int         `json:"quantity"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
	Box        BoxResponse `json:"box"`
}

type PcOnlineResponse struct {
	ID         uint                     `json:"id"`
	Tracking   string                   `json:"tracking"`
	UserID     *uint                    `json:"user_id"`
	Complained bool                     `json:"complained"`
	CreatedAt  time.Time                `json:"created_at"`
	UpdatedAt  time.Time                `json:"updated_at"`
	Details    []PcOnlineDetailResponse `json:"details"`

	// Related data
	Order *OrderResponse `json:"order,omitempty"`
	User  *UserResponse  `json:"user,omitempty"`
}

// ToPcOnlineResponse converts PcOnline model to PcOnlineResponse
func (pco *PcOnline) ToPcOnlineResponse() PcOnlineResponse {
	// Convert details to response format
	detailResponses := make([]PcOnlineDetailResponse, len(pco.Details))
	for i, detail := range pco.Details {
		detailResponse := PcOnlineDetailResponse{
			ID:         detail.ID,
			PcOnlineID: detail.PcOnlineID,
			BoxID:      detail.BoxID,
			Quantity:   detail.Quantity,
			CreatedAt:  detail.CreatedAt,
			UpdatedAt:  detail.UpdatedAt,
		}

		// Include box data if loaded
		if detail.Box != nil && detail.Box.ID != 0 {
			detailResponse.Box = detail.Box.ToBoxResponse()
		}

		detailResponses[i] = detailResponse
	}

	response := PcOnlineResponse{
		ID:         pco.ID,
		Tracking:   pco.Tracking,
		UserID:     pco.UserID,
		Complained: pco.Complained,
		CreatedAt:  pco.CreatedAt,
		UpdatedAt:  pco.UpdatedAt,
		Details:    detailResponses,
	}

	// Include order data if loaded
	if pco.Order != nil {
		orderResponse := pco.Order.ToOrderResponse()
		response.Order = &orderResponse
	}

	// Include user data if loaded
	if pco.User != nil {
		userResponse := pco.User.ToUserResponse()
		response.User = &userResponse
	}

	return response
}

// LoadOrder manually loads the related order by tracking number
func (pco *PcOnline) LoadOrder(db *gorm.DB) error {
	if pco.Tracking == "" {
		return nil
	}

	var order Order
	if err := db.Where("tracking = ?", pco.Tracking).
		Preload("OrderDetails").
		Preload("Picker.UserRoles.Role").
		Preload("Picker.UserRoles.Assigner").
		First(&order).Error; err != nil {
		return err
	}

	pco.Order = &order
	return nil
}

// Helper method to convert multiple PcOnline to responses
func ToPcOnlineResponses(pcOnlines []PcOnline) []PcOnlineResponse {
	responses := make([]PcOnlineResponse, len(pcOnlines))
	for i, pco := range pcOnlines {
		responses[i] = pco.ToPcOnlineResponse()
	}

	return responses
}
