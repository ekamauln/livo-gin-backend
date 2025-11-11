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

type PickOrderController struct {
	DB *gorm.DB
}

// NewPickOrderController creates a new PickOrderController
func NewPickOrderController(db *gorm.DB) *PickOrderController {
	return &PickOrderController{DB: db}
}

// GetPickOrders godoc
// @Summary Get all Pick Orders
// @Description Get a list of all pick orders with their details and search (logged in users only)
// @Tags Pick-Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param start_date query string false "Start date (YYYY-MM-DD format)"
// @Param end_date query string false "End date (YYYY-MM-DD format)"
// @Param search query string false "Search by Picker name (partial match)"
// @Success 200 {object} utils.Response{data=PickOrdersListResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/pick-orders [get]
func (poc *PickOrderController) GetPickOrders(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	offset := (page - 1) * limit

	// Parse date range parameters
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Parse search parameter
	search := c.Query("search")

	var pickOrders []models.PickOrder
	var total int64

	// Build query with optional search
	query := poc.DB.Model(&models.PickOrder{})

	// Apply date range filters if provided
	if startDate != "" {
		// Parse start date and set time to beginning of day
		parsedStartDate, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid start_date format", "start_date must be in YYYY-MM-DD format")
			return
		}
		startOfDay := parsedStartDate.Format("2006-01-02 00:00:00")
		query = query.Where("pick_orders.created_at >= ?", startOfDay)
	}

	if endDate != "" {
		// Parse end date and set time to end of day
		parsedEndDate, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid end_date format", "end_date must be in YYYY-MM-DD format")
			return
		}
		// Add 24 hours to get the start of next day, then use < instead of <=
		nextDay := parsedEndDate.AddDate(0, 0, 1).Format("2006-01-02 00:00:00")
		query = query.Where("pick_orders.created_at < ?", nextDay)
	}

	if search != "" {
		// Search by picker name with partial match
		query = query.Joins("JOIN users ON users.id = pick_orders.picker_id AND users.deleted_at IS NULL").
			Where("users.full_name ILIKE ?", "%"+search+"%")
	}

	// Get total count with search filter
	if err := query.Count(&total).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to count pick orders", err.Error())
		return
	}

	// Get pick orders with pagination, search filter, and order by ID desc
	if err := query.Preload("PickOrderDetails").
		Preload("Picker.UserRoles.Role").
		Preload("Picker.UserRoles.Assigner").
		Preload("Order.OrderDetails").
		Preload("Order.Picker.UserRoles.Role").
		Preload("Order.Picker.UserRoles.Assigner").
		Order("pick_orders.id DESC").
		Limit(limit).
		Offset(offset).
		Find(&pickOrders).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve pick orders", err.Error())
		return
	}

	// Load products for each pick order
	for i := range pickOrders {
		if err := pickOrders[i].LoadProducts(poc.DB); err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to load products for pick order details", err.Error())
			return
		}
	}

	// Convert to response format
	pickOrderResponses := make([]models.PickOrderResponse, len(pickOrders))
	for i, pickOrder := range pickOrders {
		pickOrderResponses[i] = pickOrder.ToPickOrderResponse()
	}

	response := PickOrdersListResponse{
		PickOrders: pickOrderResponses,
		Pagination: utils.PaginationResponse{
			Page:  page,
			Limit: limit,
			Total: int(total),
		},
	}

	// Build success message
	message := "Pick orders retrieved successfully"
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

// GetPickOrder godoc
// @Summary Get a pick order by ID
// @Description Get a pick order by ID (admin only)
// @Tags Pick-Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Pick order ID"
// @Success 200 {object} utils.Response{data=models.PickOrderResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/pick-orders/{id} [get]
func (poc *PickOrderController) GetPickOrder(c *gin.Context) {
	pickOrderId := c.Param("id")

	var pickOrder models.PickOrder
	if err := poc.DB.Preload("PickOrderDetails").
		Preload("Picker.UserRoles.Role").
		Preload("Picker.UserRoles.Assigner").
		Preload("Order.OrderDetails").
		Preload("Order.Picker.UserRoles.Role").
		Preload("Order.Picker.UserRoles.Assigner").
		First(&pickOrder, pickOrderId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Pick order not found", err.Error())
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve pick order", err.Error())
		return
	}

	// Load products for pick order details
	if err := pickOrder.LoadProducts(poc.DB); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to load products for pick order details", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Pick order retrieved successfully", pickOrder.ToPickOrderResponse())
}

// Request/Response structs
type PickOrdersListResponse struct {
	PickOrders []models.PickOrderResponse `json:"pick_orders"`
	Pagination utils.PaginationResponse   `json:"pagination"`
}
