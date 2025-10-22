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
	Tracking     string         `gorm:"unique;not null" json:"tracking" example:"JNE1234567890"`
	PickerID     *uint          `gorm:"default:null" json:"picker_id"`
	PickedAt     *time.Time     `gorm:"default:null" json:"picked_at"`
	Complained   bool           `gorm:"default:false" json:"complained" example:"false"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	OrderDetails []OrderDetail  `gorm:"foreignKey:OrderID" json:"order_details"`
	Picker       *User          `gorm:"foreignKey:PickerID" json:"picker,omitempty"`
}

type OrderDetail struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	OrderID     uint      `json:"order_id"`
	Sku         string    `json:"sku" gorm:"index"`
	ProductName string    `json:"product_name"`
	Variant     string    `json:"variant"`
	Quantity    int       `json:"quantity"`
	Product     *Product  `json:"product,omitempty" gorm:"-"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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
	Complained   bool                  `json:"complained"`
	PickedBy     string                `json:"picked_by"`
	PickedAt     string                `json:"picked_at"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
	OrderDetails []OrderDetailResponse `json:"order_details"`
}

type OrderDetailResponse struct {
	ID          uint             `json:"id"`
	Sku         string           `json:"sku"`
	ProductName string           `json:"product_name"`
	Variant     string           `json:"variant"`
	Quantity    int              `json:"quantity"`
	Product     *ProductResponse `json:"product,omitempty"`
}

// ToOrderResponse converts Order model to OrderResponse
func (o *Order) ToOrderResponse() OrderResponse {
	details := make([]OrderDetailResponse, len(o.OrderDetails))
	for i, detail := range o.OrderDetails {
		detailResp := OrderDetailResponse{
			ID:          detail.ID,
			Sku:         detail.Sku,
			ProductName: detail.ProductName,
			Variant:     detail.Variant,
			Quantity:    detail.Quantity,
		}

		// Include product data if exists
		if detail.Product != nil {
			detailResp.Product = &ProductResponse{
				ID:    detail.Product.ID,
				Sku:   detail.Product.Sku,
				Name:  detail.Product.Name,
				Image: detail.Product.Image,
			}
		}

		details[i] = detailResp
	}

	// Handle picked_at field
	var pickedAtStr string
	if o.PickedAt != nil {
		pickedAtStr = o.PickedAt.Format("2006-01-02 15:04:05")
	} else {
		pickedAtStr = "Not picked yet"
	}

	// Handle picked_by field
	var pickedByStr string
	if o.Picker != nil {
		pickedByStr = o.Picker.FullName + " (" + o.Picker.Username + ")"
	} else {
		pickedByStr = "Not picked yet"
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
		Complained:   o.Complained,
		CreatedAt:    o.CreatedAt,
		UpdatedAt:    o.UpdatedAt,
		PickedBy:     pickedByStr,
		PickedAt:     pickedAtStr,
		OrderDetails: details,
	}
}
