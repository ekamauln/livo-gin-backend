package models

import (
	"time"

	"gorm.io/gorm"
)

type PickOrder struct {
	ID               uint              `gorm:"primaryKey" json:"id"`
	OrderID          uint              `gorm:"not null;index" json:"order_id"`
	PickerID         uint              `gorm:"not null;index" json:"picker_id"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
	DeletedAt        gorm.DeletedAt    `gorm:"index" json:"-"`
	Order            *Order            `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Picker           *User             `gorm:"foreignKey:PickerID" json:"picker,omitempty"`
	PickOrderDetails []PickOrderDetail `gorm:"foreignKey:PickOrderID" json:"pick_order_details"`
}

type PickOrderDetail struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	PickOrderID uint      `gorm:"not null;index" json:"pick_order_id"`
	Sku         string    `gorm:"not null;index" json:"sku"`
	ProductName string    `gorm:"not null" json:"product_name"`
	Variant     string    `json:"variant"`
	Quantity    int       `gorm:"not null" json:"quantity"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Product     *Product  `json:"product,omitempty" gorm:"-"`
}

type PickOrderResponse struct {
	ID               uint                      `json:"id"`
	OrderID          uint                      `json:"order_id"`
	PickerID         uint                      `json:"picker_id"`
	CreatedAt        time.Time                 `json:"created_at"`
	UpdatedAt        time.Time                 `json:"updated_at"`
	Order            *OrderResponse            `json:"order,omitempty"`
	Picker           *UserResponse             `json:"picker,omitempty"`
	PickOrderDetails []PickOrderDetailResponse `json:"pick_order_details"`
}

type PickOrderDetailResponse struct {
	ID          uint             `json:"id"`
	PickOrderID uint             `json:"pick_order_id"`
	Sku         string           `json:"sku"`
	ProductName string           `json:"product_name"`
	Variant     string           `json:"variant"`
	Quantity    int              `json:"quantity"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	Product     *ProductResponse `json:"product,omitempty"`
}

// ToPickOrderResponse converts PickOrder model to PickOrderResponse
func (po *PickOrder) ToPickOrderResponse() PickOrderResponse {
	details := make([]PickOrderDetailResponse, len(po.PickOrderDetails))
	for i, detail := range po.PickOrderDetails {
		detailResp := PickOrderDetailResponse{
			ID:          detail.ID,
			PickOrderID: detail.PickOrderID,
			Sku:         detail.Sku,
			ProductName: detail.ProductName,
			Variant:     detail.Variant,
			Quantity:    detail.Quantity,
			CreatedAt:   detail.CreatedAt,
			UpdatedAt:   detail.UpdatedAt,
		}

		// Include product data if exists
		if detail.Product != nil {
			productResp := detail.Product.ToProductResponse()
			detailResp.Product = &productResp
		}

		details[i] = detailResp
	}

	response := PickOrderResponse{
		ID:               po.ID,
		OrderID:          po.OrderID,
		PickerID:         po.PickerID,
		CreatedAt:        po.CreatedAt,
		UpdatedAt:        po.UpdatedAt,
		PickOrderDetails: details,
	}

	// Include order data if exists
	if po.Order != nil {
		orderResp := po.Order.ToOrderResponse()
		response.Order = &orderResp
	}

	// Include picker data if exists
	if po.Picker != nil {
		pickerResp := po.Picker.ToUserResponse()
		response.Picker = &pickerResp
	}

	return response
}

// LoadProducts manually loads products for all pick order details by SKU
func (po *PickOrder) LoadProducts(db *gorm.DB) error {
	for i := range po.PickOrderDetails {
		var product Product
		if err := db.Where("sku = ?", po.PickOrderDetails[i].Sku).First(&product).Error; err == nil {
			po.PickOrderDetails[i].Product = &product
		}
		// Silently skip if product not found
	}
	return nil
}
