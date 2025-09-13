package models

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	OrderGineeID string         `gorm:"unique;not null" json:"order_id" example:"2509116GA36VM5"`
	Status       string         `json:"status" example:"pending"`
	Channel      string         `json:"channel" example:"Shopee"`
	Store        string         `json:"store" example:"SP deParcelRibbon"`
	Buyer        string         `json:"buyer" example:"John Doe"`
	Courier      string         `json:"courier" example:"JNE"`
	Tracking     string         `gorm:"index" json:"tracking" example:"JNE1234567890"`
	UserID       *uint          `gorm:"default:null" json:"user_id"`
	PickedAt     *time.Time     `gorm:"default:null" json:"picked_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	OrderDetails []OrderDetail  `gorm:"foreignKey:OrderID" json:"order_details"`
	User         *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type OrderDetail struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	OrderID     uint           `gorm:"not null"  json:"order_id"`
	Sku         string         `gorm:"not null"  json:"sku"`
	ProductName string         `json:"product_name"`
	Variant     string         `json:"variant"`
	Quantity    int            `gorm:"not null" json:"quantity" example:"2"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Order       Order          `gorm:"foreignKey:OrderID" json:"order"`
}

// OrderResponse represents order data for API responses
type OrderResponse struct {
	ID           uint                  `json:"id"`
	OrderGineeID string                `json:"order_id"`
	Status       string                `json:"status"`
	Channel      string                `json:"channel"`
	Store        string                `json:"store"`
	Buyer        string                `json:"buyer"`
	Courier      string                `json:"courier"`
	Tracking     string                `json:"tracking"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
	OrderDetails []OrderDetailResponse `json:"order_details"`
}

type OrderDetailResponse struct {
	ID          uint   `json:"id"`
	OrderID     uint   `json:"order_id"`
	Sku         string `json:"sku"`
	ProductName string `json:"product_name"`
	Variant     string `json:"variant"`
	Quantity    int    `json:"quantity"`
}

// ToOrderResponse converts Order model to OrderResponse
func (o *Order) ToOrderResponse() OrderResponse {
	orderDetails := make([]OrderDetailResponse, len(o.OrderDetails))
	for i, od := range o.OrderDetails {
		orderDetails[i] = OrderDetailResponse{
			ID:          od.ID,
			OrderID:     od.OrderID,
			Sku:         od.Sku,
			ProductName: od.ProductName,
			Variant:     od.Variant,
			Quantity:    od.Quantity,
		}
	}

	return OrderResponse{
		ID:           o.ID,
		OrderGineeID: o.OrderGineeID,
		Status:       o.Status,
		Channel:      o.Channel,
		Store:        o.Store,
		Buyer:        o.Buyer,
		Courier:      o.Courier,
		Tracking:     o.Tracking,
		CreatedAt:    o.CreatedAt,
		UpdatedAt:    o.UpdatedAt,
		OrderDetails: orderDetails,
	}
}
