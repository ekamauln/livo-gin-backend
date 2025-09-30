package models

import (
	"time"

	"gorm.io/gorm"
)

type QcRibbon struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Tracking   string         `gorm:"unique;not null" json:"tracking" example:"QC1234567890"`
	UserID     *uint          `gorm:"default:null" json:"user_id"`
	Complained bool           `gorm:"default:false" json:"complained"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// One-to-many relationship with QcRibbonDetail
	Details []QcRibbonDetail `gorm:"foreignKey:QcRibbonID" json:"details"`

	// Relations
	Order *Order `gorm:"-" json:"order,omitempty"`
	User  *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type QcRibbonDetail struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	QcRibbonID uint           `gorm:"not null" json:"qc_ribbon_id"`
	BoxID      uint           `gorm:"not null" json:"box_id"`
	Quantity   int            `json:"quantity"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	QcRibbon QcRibbon `gorm:"foreignKey:QcRibbonID" json:"-"`
	Box      *Box     `gorm:"foreignKey:BoxID" json:"box,omitempty"`
}

// Response structures
type QcRibbonDetailResponse struct {
	ID         uint        `json:"id"`
	QcRibbonID uint        `json:"qc_ribbon_id"`
	BoxID      uint        `json:"box_id"`
	Quantity   int         `json:"quantity"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
	Box        BoxResponse `json:"box"`
}

type QcRibbonResponse struct {
	ID         uint                     `json:"id"`
	Tracking   string                   `json:"tracking"`
	UserID     *uint                    `json:"user_id"`
	Complained bool                     `json:"complained"`
	CreatedAt  time.Time                `json:"created_at"`
	UpdatedAt  time.Time                `json:"updated_at"`
	Details    []QcRibbonDetailResponse `json:"details"`

	// Related data
	Order *OrderResponse `json:"order,omitempty"`
	User  *UserResponse  `json:"user,omitempty"`
}

// ToQcRibbonResponse converts QcRibbon to QcRibbonResponse
func (qcr *QcRibbon) ToQcRibbonResponse() QcRibbonResponse {
	// Convert details to response format
	detailResponses := make([]QcRibbonDetailResponse, len(qcr.Details))
	for i, detail := range qcr.Details {
		detailResponse := QcRibbonDetailResponse{
			ID:         detail.ID,
			QcRibbonID: detail.QcRibbonID,
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

	response := QcRibbonResponse{
		ID:         qcr.ID,
		Tracking:   qcr.Tracking,
		UserID:     qcr.UserID,
		Complained: qcr.Complained,
		CreatedAt:  qcr.CreatedAt,
		UpdatedAt:  qcr.UpdatedAt,
		Details:    detailResponses,
	}

	// Include order data if loaded
	if qcr.Order != nil {
		orderResponse := qcr.Order.ToOrderResponse()
		response.Order = &orderResponse
	}

	// Include user data if loaded
	if qcr.User != nil {
		userResponse := qcr.User.ToUserResponse()
		response.User = &userResponse
	}

	return response
}

// LoadOrder manually loads the related order by tracking number
func (qcr *QcRibbon) LoadOrder(db *gorm.DB) error {
	if qcr.Tracking == "" {
		return nil
	}

	var order Order
	if err := db.Preload("OrderDetails").Where("order_ginee_id = ?", qcr.Tracking).First(&order).Error; err != nil {
		return err
	}

	qcr.Order = &order
	return nil
}

// Helper method to convert multiple QcRibbon to responses
func ToQcRibbonResponses(qcRibbons []QcRibbon) []QcRibbonResponse {
	responses := make([]QcRibbonResponse, len(qcRibbons))
	for i, qcr := range qcRibbons {
		responses[i] = qcr.ToQcRibbonResponse()
	}

	return responses
}
