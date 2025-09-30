package models

import (
	"time"

	"gorm.io/gorm"
)

type QcOnline struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Tracking   string         `gorm:"unique;not null" json:"tracking" example:"QC1234567890"`
	UserID     *uint          `gorm:"default:null" json:"user_id"`
	Complained bool           `gorm:"default:false" json:"complained"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// One-to-many relationship with QcOnlineDetail
	Details []QcOnlineDetail `gorm:"foreignKey:QcOnlineID" json:"details"`

	// Relations
	Order *Order `gorm:"-" json:"order,omitempty"`
	User  *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type QcOnlineDetail struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	QcOnlineID uint           `gorm:"not null" json:"qc_online_id"`
	BoxID      uint           `gorm:"not null" json:"box_id"`
	Quantity   int            `gorm:"not null" json:"quantity"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	QcOnline QcOnline `gorm:"foreignKey:QcOnlineID" json:"-"`
	Box      *Box     `gorm:"foreignKey:BoxID" json:"box,omitempty"`
}

// Response structures
type QcOnlineDetailResponse struct {
	ID         uint        `json:"id"`
	QcOnlineID uint        `json:"qc_online_id"`
	BoxID      uint        `json:"box_id"`
	Quantity   int         `json:"quantity"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
	Box        BoxResponse `json:"box"`
}

type QcOnlineResponse struct {
	ID         uint                     `json:"id"`
	Tracking   string                   `json:"tracking"`
	UserID     *uint                    `json:"user_id"`
	Complained bool                     `json:"complained"`
	CreatedAt  time.Time                `json:"created_at"`
	UpdatedAt  time.Time                `json:"updated_at"`
	Details    []QcOnlineDetailResponse `json:"details"`

	// Related data
	Order *OrderResponse `json:"order,omitempty"`
	User  *UserResponse  `json:"user,omitempty"`
}

// ToQcOnlineResponse converts QcOnline to QcOnlineResponse
func (qco *QcOnline) ToQcOnlineResponse() QcOnlineResponse {
	// Convert details to response format
	detailResponses := make([]QcOnlineDetailResponse, len(qco.Details))
	for i, detail := range qco.Details {
		detailResponse := QcOnlineDetailResponse{
			ID:         detail.ID,
			QcOnlineID: detail.QcOnlineID,
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

	response := QcOnlineResponse{
		ID:         qco.ID,
		Tracking:   qco.Tracking,
		UserID:     qco.UserID,
		Complained: qco.Complained,
		CreatedAt:  qco.CreatedAt,
		UpdatedAt:  qco.UpdatedAt,
		Details:    detailResponses,
	}

	// Include order data if loaded
	if qco.Order != nil {
		orderResponse := qco.Order.ToOrderResponse()
		response.Order = &orderResponse
	}

	// Include user data if loaded
	if qco.User != nil {
		userResponse := qco.User.ToUserResponse()
		response.User = &userResponse
	}

	return response
}

// LoadOrder manually loads the related order by tracking number
func (qco *QcOnline) LoadOrder(db *gorm.DB) error {
	if qco.Tracking == "" {
		return nil
	}

	var order Order
	if err := db.Where("tracking = ?", qco.Tracking).
		Preload("OrderDetails").
		Preload("Picker").
		First(&order).Error; err != nil {
		return err
	}

	qco.Order = &order
	return nil
}

// Helper method to convert multiple QcOnline to responses
func ToQcOnlineResponses(qcOnlines []QcOnline) []QcOnlineResponse {
	responses := make([]QcOnlineResponse, len(qcOnlines))
	for i, qco := range qcOnlines {
		responses[i] = qco.ToQcOnlineResponse()
	}

	return responses
}
