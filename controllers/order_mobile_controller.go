package controllers

import (
	"livo-gin-backend/models"
	"livo-gin-backend/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrderMobileController struct {
	DB *gorm.DB
}

// NewOrderMobileController creates a new order mobile controller
func NewOrderMobileController(db *gorm.DB) *OrderMobileController {
	return &OrderMobileController{DB: db}
}

// GetOrders godoc
// @Summary Get all orders for pickers
// @Description Get list of all orders with "ready to pick" status (mobile - picker only)
// @Tags mobile-orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} utils.Response{data=OrdersListResponse}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/mobile/orders [get]
func (omc *OrderMobileController) GetOrders(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var orders []models.Order
	var total int64

	// Get total count of "ready to pick" orders
	omc.DB.Model(&models.Order{}).Where("status = ?", "ready to pick").Count(&total)

	// Get orders with "ready to pick" status only
	if err := omc.DB.Where("status = ?", "ready to pick").
		Limit(limit).Offset(offset).
		Preload("OrderDetails").
		Find(&orders).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve orders", err.Error())
		return
	}

	// Convert to response format
	orderResponses := make([]models.OrderResponse, len(orders))
	for i, order := range orders {
		orderResponses[i] = order.ToOrderResponse()
	}

	response := OrdersListResponse{
		Orders: orderResponses,
		Pagination: PaginationResponse{
			Page:  page,
			Limit: limit,
			Total: int(total),
		},
	}

	utils.SuccessResponse(c, http.StatusOK, "Orders retrieved successfully", response)
}

// PickingOrder godoc
// @Summary Pick an order for processing
// @Description Change order status from "ready to pick" to "picking process" and assign to current picker
// @Tags mobile-orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Success 200 {object} utils.Response{data=models.OrderResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/mobile/orders/{id}/pick [put]
func (omc *OrderMobileController) PickingOrder(c *gin.Context) {
	// Get order ID from URL parameter
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid order ID", err.Error())
		return
	}

	// Get current user ID from context (set by auth middleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not found", "user ID not found in context")
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user ID", "user ID has invalid type")
		return
	}

	var order models.Order
	// Find order and check if it's available to pick
	if err := omc.DB.Where("id = ? AND status = ?", orderID, "ready to pick").First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Order not found or not available for picking", "order not found or already picked")
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to find order", err.Error())
		}
		return
	}

	// Update order status and assign picker
	now := time.Now()
	order.Status = "picking process"
	order.PickerID = &userID
	order.PickedAt = &now

	// Save the changes
	if err := omc.DB.Save(&order).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update order", err.Error())
		return
	}

	// Load order with details and picker for response
	omc.DB.Preload("OrderDetails").Preload("Picker").First(&order, order.ID)

	utils.SuccessResponse(c, http.StatusOK, "Order picked successfully", order.ToOrderResponse())
}

// GetOrder godoc
// @Summary Get order details with product information
// @Description Get specific order details with product location and barcode joined by SKU
// @Tags mobile-orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Success 200 {object} utils.Response{data=MobileOrderDetailResponse}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/mobile/orders/{id} [get]
func (omc *OrderMobileController) GetOrder(c *gin.Context) {
	// Get order ID from URL parameter
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid order ID", err.Error())
		return
	}

	// Get current user ID from context
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not found", "user ID not found in context")
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user ID", "user ID has invalid type")
		return
	}

	var order models.Order
	// Find order assigned to current picker
	if err := omc.DB.Where("id = ? AND picker_id = ?", orderID, userID).
		Preload("OrderDetails").
		Preload("Picker").
		First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Order not found", "order not found or not assigned to you")
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to find order", err.Error())
		}
		return
	}

	// Get product details for each order detail
	var orderDetailsWithProduct []MobileOrderDetailWithProduct
	for _, detail := range order.OrderDetails {
		var product models.Product

		// Find product by SKU
		if err := omc.DB.Where("sku = ?", detail.Sku).First(&product).Error; err != nil {
			// If product not found, use empty location and barcode
			orderDetailsWithProduct = append(orderDetailsWithProduct, MobileOrderDetailWithProduct{
				OrderDetailResponse: models.OrderDetailResponse{
					ID:          detail.ID,
					OrderID:     detail.OrderID,
					Sku:         detail.Sku,
					ProductName: detail.ProductName,
					Variant:     detail.Variant,
					Quantity:    detail.Quantity,
				},
				Location: "Location not found",
				Barcode:  "Barcode not found",
			})
		} else {
			// Product found, include location and barcode
			orderDetailsWithProduct = append(orderDetailsWithProduct, MobileOrderDetailWithProduct{
				OrderDetailResponse: models.OrderDetailResponse{
					ID:          detail.ID,
					OrderID:     detail.OrderID,
					Sku:         detail.Sku,
					ProductName: detail.ProductName,
					Variant:     detail.Variant,
					Quantity:    detail.Quantity,
				},
				Location: product.Location,
				Barcode:  product.Barcode,
			})
		}
	}

	// Handle picked_at field
	var pickedAtStr string
	if order.PickedAt != nil {
		pickedAtStr = order.PickedAt.Format("2006-01-02 15:04:05")
	} else {
		pickedAtStr = "Not picked yet"
	}

	// Handle picked_by field
	var pickedByStr string
	if order.Picker != nil {
		pickedByStr = order.Picker.FullName + " (" + order.Picker.Username + ")"
	} else {
		pickedByStr = "Not picked yet"
	}

	response := MobileOrderDetailResponse{
		ID:           order.ID,
		OrderGineeID: order.OrderGineeID,
		Status:       order.Status,
		Channel:      order.Channel,
		Store:        order.Store,
		Buyer:        order.Buyer,
		Courier:      order.Courier,
		Tracking:     order.Tracking,
		PickedAt:     pickedAtStr,
		CreatedAt:    order.CreatedAt,
		UpdatedAt:    order.UpdatedAt,
		PickedBy:     pickedByStr,
		OrderDetails: orderDetailsWithProduct,
	}

	utils.SuccessResponse(c, http.StatusOK, "Order details retrieved successfully", response)
}

// CompletePickingOrder godoc
// @Summary Complete picking process
// @Description Change order status from "picking process" to "picking complete"
// @Tags mobile-orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Success 200 {object} utils.Response{data=models.OrderResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/mobile/orders/{id}/complete [put]
func (omc *OrderMobileController) CompletePickingOrder(c *gin.Context) {
	// Get order ID from URL parameter
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid order ID", err.Error())
		return
	}

	// Get current user ID from context
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not found", "user ID not found in context")
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user ID", "user ID has invalid type")
		return
	}

	var order models.Order
	// Find order assigned to current picker with "picking process" status
	if err := omc.DB.Where("id = ? AND picker_id = ? AND status = ?", orderID, userID, "picking process").First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Order not found or not in picking process", "order not found or not in picking process")
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to find order", err.Error())
		}
		return
	}

	// Update order status to complete
	order.Status = "picking complete"

	// Save the changes
	if err := omc.DB.Save(&order).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to complete order", err.Error())
		return
	}

	// Load order with details and picker for response
	omc.DB.Preload("OrderDetails").Preload("Picker").First(&order, order.ID)

	utils.SuccessResponse(c, http.StatusOK, "Order picking completed successfully", order.ToOrderResponse())
}

// Response structs for mobile endpoints - using shared types from other controllers
type MobileOrderDetailResponse struct {
	ID           uint                            `json:"id"`
	OrderGineeID string                          `json:"order_id"`
	Status       string                          `json:"status"`
	Channel      string                          `json:"channel"`
	Store        string                          `json:"store"`
	Buyer        string                          `json:"buyer"`
	Courier      string                          `json:"courier"`
	Tracking     string                          `json:"tracking"`
	PickedAt     string                          `json:"picked_at"`
	CreatedAt    time.Time                       `json:"created_at"`
	UpdatedAt    time.Time                       `json:"updated_at"`
	PickedBy     string                          `json:"picked_by"`
	OrderDetails []MobileOrderDetailWithProduct `json:"order_details"`
}

type MobileOrderDetailWithProduct struct {
	models.OrderDetailResponse
	Location string `json:"location"`
	Barcode  string `json:"barcode"`
}
