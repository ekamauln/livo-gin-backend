package controllers

import (
	"fmt"
	"livo-gin-backend/models"
	"livo-gin-backend/utils"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrderController struct {
	DB *gorm.DB
}

// NewOrderController creates a new order controller
func NewOrderController(db *gorm.DB) *OrderController {
	return &OrderController{DB: db}
}

// GetOrders godoc
// @Summary Get all orders
// @Description Get list of all orders with optional date range filtering and search (logged-in users only)
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param start_date query string false "Start date (YYYY-MM-DD format)"
// @Param end_date query string false "End date (YYYY-MM-DD format)"
// @Param search query string false "Search by Order ID or Tracking number"
// @Success 200 {object} utils.Response{data=OrdersListResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/orders [get]
func (oc *OrderController) GetOrders(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Parse date range parameters
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Parse search parameter
	search := c.Query("search")

	var orders []models.Order
	var total int64

	// Build the query
	query := oc.DB.Model(&models.Order{})

	// Apply date range filters if provided
	if startDate != "" {
		// Parse start date and set time to beginning of day
		if parsedStartDate, err := time.Parse("2006-01-02", startDate); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid start_date format", "start_date must be in YYYY-MM-DD format")
			return
		} else {
			startOfDay := parsedStartDate.Format("2006-01-02 00:00:00")
			query = query.Where("created_at >= ?", startOfDay)
		}
	}

	if endDate != "" {
		// Parse end date and set time to end of day
		if parsedEndDate, err := time.Parse("2006-01-02", endDate); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid end_date format", "end_date must be in YYYY-MM-DD format")
			return
		} else {
			// Add 24 hours to get the start of next day, then use < instead of <=
			nextDay := parsedEndDate.AddDate(0, 0, 1).Format("2006-01-02 00:00:00")
			query = query.Where("created_at < ?", nextDay)
		}
	}

	// Apply search filter if provided
	if search != "" {
		// Search in both order_ginee_id and tracking fields
		query = query.Where("order_ginee_id ILIKE ? OR tracking ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Get total count with all filters
	query.Count(&total)

	// Get orders with pagination, filters, sorted by ID ascending
	if err := query.Order("id ASC").Limit(limit).Offset(offset).Preload("Picker.UserRoles.Role").Preload("Picker.UserRoles.Assigner").Preload("OrderDetails").Find(&orders).Error; err != nil {
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

	// Build success message
	message := "Orders retrieved successfully"
	var filters []string
	
	if startDate != "" || endDate != "" {
		var dateRange []string
		if startDate != "" {
			dateRange = append(dateRange, "from: "+startDate)
		}
		if endDate != "" {
			dateRange = append(dateRange, "to: "+endDate)
		}
		filters = append(filters, "date: "+strings.Join(dateRange, ", "))
	}
	
	if search != "" {
		filters = append(filters, "search: "+search)
	}
	
	if len(filters) > 0 {
		message += fmt.Sprintf(" (filtered by %s)", strings.Join(filters, " | "))
	}

	utils.SuccessResponse(c, http.StatusOK, message, response)
}

// CreateOrder godoc
// @Summary Create a new order
// @Description Create a new order with order details (logged-in users only)
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateOrderRequest true "Create order request"
// @Success 201 {object} utils.Response{data=models.OrderResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/orders [post]
func (oc *OrderController) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Check if order with same OrderGineeID already exists
	var existingOrder models.Order
	if err := oc.DB.Where("order_ginee_id = ?", req.OrderGineeID).First(&existingOrder).Error; err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "Order already exists", "order with this ID already exists")
		return
	}

	// Create order
	order := models.Order{
		OrderGineeID: req.OrderGineeID,
		Status:       "ready to pick", // Always set to "ready to pick"
		Channel:      req.Channel,
		Store:        req.Store,
		Buyer:        req.Buyer,
		Courier:      req.Courier,
		Tracking:     req.Tracking,
	}

	// Create order details
	for _, detailReq := range req.OrderDetails {
		orderDetail := models.OrderDetail{
			Sku:         detailReq.Sku,
			ProductName: detailReq.ProductName,
			Variant:     detailReq.Variant,
			Quantity:    detailReq.Quantity,
		}
		order.OrderDetails = append(order.OrderDetails, orderDetail)
	}

	// Create order with details in a transaction
	if err := oc.DB.Create(&order).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create order", err.Error())
		return
	}

	// Load order with details for response
	oc.DB.Preload("OrderDetails").Preload("Picker").First(&order, order.ID)

	utils.SuccessResponse(c, http.StatusCreated, "Order created successfully", order.ToOrderResponse())
}

// BulkCreateOrders godoc
// @Summary Bulk create orders
// @Description Create multiple orders at once, skipping duplicates (logged-in users only)
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body BulkCreateOrderRequest true "Bulk create order request"
// @Success 201 {object} utils.Response{data=BulkCreateOrderResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/orders/bulk [post]
func (oc *OrderController) BulkCreateOrders(c *gin.Context) {
	var req BulkCreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	var createdOrders []models.Order
	var skippedOrders []SkippedOrder
	var failedOrders []FailedOrder

	for i, orderReq := range req.Orders {
		// Check if order with same OrderGineeID already exists
		var existingOrder models.Order
		if err := oc.DB.Where("order_ginee_id = ?", orderReq.OrderGineeID).First(&existingOrder).Error; err == nil {
			// Order exists, skip it
			skippedOrders = append(skippedOrders, SkippedOrder{
				Index:        i,
				OrderGineeID: orderReq.OrderGineeID,
				Reason:       "Order already exists",
			})
			continue
		}

		// Create order
		order := models.Order{
			OrderGineeID: orderReq.OrderGineeID,
			Status:       "ready to pick", // Always set to "ready to pick"
			Channel:      orderReq.Channel,
			Store:        orderReq.Store,
			Buyer:        orderReq.Buyer,
			Courier:      orderReq.Courier,
			Tracking:     orderReq.Tracking,
		}

		// Create order details
		for _, detailReq := range orderReq.OrderDetails {
			orderDetail := models.OrderDetail{
				Sku:         detailReq.Sku,
				ProductName: detailReq.ProductName,
				Variant:     detailReq.Variant,
				Quantity:    detailReq.Quantity,
			}
			order.OrderDetails = append(order.OrderDetails, orderDetail)
		}

		// Try to create the order
		if err := oc.DB.Create(&order).Error; err != nil {
			// Failed to create order
			failedOrders = append(failedOrders, FailedOrder{
				Index:        i,
				OrderGineeID: orderReq.OrderGineeID,
				Error:        err.Error(),
			})
			continue
		}

		// Load order with details for response
		oc.DB.Preload("OrderDetails").Preload("Picker").First(&order, order.ID)
		createdOrders = append(createdOrders, order)
	}

	// Convert created orders to response format
	createdOrderResponses := make([]models.OrderResponse, len(createdOrders))
	for i, order := range createdOrders {
		createdOrderResponses[i] = order.ToOrderResponse()
	}

	response := BulkCreateOrderResponse{
		Summary: BulkCreateSummary{
			Total:   len(req.Orders),
			Created: len(createdOrders),
			Skipped: len(skippedOrders),
			Failed:  len(failedOrders),
		},
		CreatedOrders: createdOrderResponses,
		SkippedOrders: skippedOrders,
		FailedOrders:  failedOrders,
	}

	// Determine response status
	statusCode := http.StatusCreated
	message := "Bulk order creation completed"

	if len(createdOrders) == 0 {
		if len(skippedOrders) > 0 {
			statusCode = http.StatusOK
			message = "All orders were skipped (already exist)"
		} else {
			statusCode = http.StatusBadRequest
			message = "No orders could be created"
		}
	} else if len(failedOrders) > 0 || len(skippedOrders) > 0 {
		message = "Bulk order creation completed with some issues"
	}

	utils.SuccessResponse(c, statusCode, message, response)
}

type OrdersListResponse struct {
	Orders     []models.OrderResponse `json:"orders"`
	Pagination PaginationResponse     `json:"pagination"`
}

type CreateOrderRequest struct {
	OrderGineeID string                     `json:"order_id" binding:"required" example:"2509116GA36VM5"`
	Status       string                     `json:"status" example:"ready to pick"`
	Channel      string                     `json:"channel" binding:"required" example:"Shopee"`
	Store        string                     `json:"store" binding:"required" example:"SP deParcelRibbon"`
	Buyer        string                     `json:"buyer" binding:"required" example:"John Doe"`
	Courier      string                     `json:"courier" example:"JNE"`
	Tracking     string                     `json:"tracking" example:"JNE1234567890"`
	OrderDetails []CreateOrderDetailRequest `json:"order_details" binding:"required,min=1"`
}

type CreateOrderDetailRequest struct {
	Sku         string `json:"sku" binding:"required" example:"PROD001"`
	ProductName string `json:"product_name" binding:"required" example:"Sample Product"`
	Variant     string `json:"variant" example:"Red - Size M"`
	Quantity    int    `json:"quantity" binding:"required,min=1" example:"2"`
}

type BulkCreateOrderRequest struct {
	Orders []CreateOrderRequest `json:"orders" binding:"required,min=1"`
}

type BulkCreateOrderResponse struct {
	Summary       BulkCreateSummary      `json:"summary"`
	CreatedOrders []models.OrderResponse `json:"created_orders"`
	SkippedOrders []SkippedOrder         `json:"skipped_orders"`
	FailedOrders  []FailedOrder          `json:"failed_orders"`
}

type BulkCreateSummary struct {
	Total   int `json:"total"`
	Created int `json:"created"`
	Skipped int `json:"skipped"`
	Failed  int `json:"failed"`
}

type SkippedOrder struct {
	Index        int    `json:"index"`
	OrderGineeID string `json:"order_id"`
	Reason       string `json:"reason"`
}

type FailedOrder struct {
	Index        int    `json:"index"`
	OrderGineeID string `json:"order_id"`
	Error        string `json:"error"`
}
