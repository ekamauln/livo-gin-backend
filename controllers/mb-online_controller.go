package controllers

import (
	"livo-gin-backend/models"
	"livo-gin-backend/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MbOnlineController struct {
	DB *gorm.DB
}

// NewMbOnlineController creates a new mb-online controller
func NewMbOnlineController(db *gorm.DB) *MbOnlineController {
	return &MbOnlineController{DB: db}
}

// GetMbOnlines godoc
// @Summary Get all mb-onlines for logged-in user
// @Description Get list of mb-onlines for current user filtered by current date (logged-in users only)
// @Tags mb-onlines
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search by tracking number"
// @Success 200 {object} utils.Response{data=MbOnlinesListResponse}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/mb-onlines [get]
func (moc *MbOnlineController) GetMbOnlines(c *gin.Context) {
	// Get user ID from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "User not authenticated")
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Parse search parameter (optional)
	search := c.Query("search")

	var mbOnlines []models.MbOnline
	var total int64

	// Build query with filters
	query := moc.DB.Model(&models.MbOnline{}).Where("user_id = ?", userID).Where("DATE(created_at) = CURRENT_DATE")

	if search != "" {
		// Search by tracking with partial match
		query = query.Where("tracking ILIKE ?", "%"+search+"%")
	}

	// Get total count with filters
	if err := query.Count(&total).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to count mb-onlines", err.Error())
		return
	}

	// Get mbOnlines with pagination, search filter, and order by ID descending
	if err := query.Order("id DESC").Preload("User.UserRoles.Role").Preload("User.UserRoles.Assigner").Limit(limit).Offset(offset).Find(&mbOnlines).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve mb-onlines", err.Error())
		return
	}

	// Manually load order data for each mb-online
	for i := range mbOnlines {
		if mbOnlines[i].Tracking != "" {
			var order models.Order
			if err := moc.DB.Where("tracking = ?", mbOnlines[i].Tracking).
				Preload("OrderDetails").
				Preload("Picker.UserRoles.Role").
				First(&order).Error; err == nil {
				mbOnlines[i].Order = &order
			}
		}
	}

	// Convert to response format
	mbOnlineResponses := make([]models.MbOnlineResponse, len(mbOnlines))
	for i, mbOnline := range mbOnlines {
		mbOnlineResponses[i] = mbOnline.ToMbOnlineResponse()
	}

	response := MbOnlinesListResponse{
		MbOnlines: mbOnlineResponses,
		Pagination: utils.PaginationResponse{
			Page:  page,
			Limit: limit,
			Total: int(total),
		},
	}

	// Build success message
	message := "Mb-onlines retrieved successfully"
	if search != "" {
		message += " (filtered by tracking: " + search + ")"
	}

	utils.SuccessResponse(c, http.StatusOK, message, response)
}

// GetMbOnline godoc
// @Summary Get a specific mb-online by ID
// @Description Get a specific mb-online by ID (logged-in users only)
// @Tags mb-onlines
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Mb-online ID"
// @Success 200 {object} utils.Response{data=models.MbOnlineResponse}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/mb-onlines/{id} [get]
func (moc *MbOnlineController) GetMbOnline(c *gin.Context) {
	mbOnlineID := c.Param("id")

	var mbOnline models.MbOnline
	if err := moc.DB.Preload("User").First(&mbOnline, mbOnlineID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Mb-online not found", err.Error())
		return
	}

	// Manually load order if tracking exists
	if mbOnline.Tracking != "" {
		var order models.Order
		if err := moc.DB.Where("tracking = ?", mbOnline.Tracking).
			Preload("OrderDetails").
			Preload("Picker.UserRoles.Role").
			Preload("Picker.UserRoles.Assigner").
			First(&order).Error; err == nil {
			mbOnline.Order = &order
		}
	}

	utils.SuccessResponse(c, http.StatusOK, "Mb-online retrieved successfully", mbOnline.ToMbOnlineResponse())
}

// CreateMbOnline godoc
// @Summary Create new mb-online
// @Description Create a new mb-online entry (logged-in users only)
// @Tags mb-onlines
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param mb_online body CreateMbOnlineRequest true "Mb-online data"
// @Success 201 {object} utils.Response{data=models.MbOnlineResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/mb-onlines [post]
func (moc *MbOnlineController) CreateMbOnline(c *gin.Context) {
	// Get user ID from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "User not authenticated")
		return
	}

	var req CreateMbOnlineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Convert tracking to uppercase
	req.Tracking = strings.ToUpper(strings.TrimSpace(req.Tracking))

	// Convert userID to uint
	userIDUint, ok := userID.(uint)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user ID", "Failed to convert user ID")
		return
	}

	// Check if tracking exists in orders table first
	var order models.Order
	if err := moc.DB.Where("tracking = ?", req.Tracking).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Order not found", "No order found with the specified tracking number")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to validate tracking", err.Error())
		return
	}

	// Check if tracking already exists in mb-ribbon table
	var existingMbRibbon models.MbRibbon
	if err := moc.DB.Where("tracking = ?", req.Tracking).First(&existingMbRibbon).Error; err == nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tracking already processed in MB-Ribbon", "This tracking number is already being processed in the ribbon workflow")
		return
	}

	// Check for duplicate tracking
	var existingMbOnline models.MbOnline
	if err := moc.DB.Where("tracking = ?", req.Tracking).First(&existingMbOnline).Error; err == nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Mb-online with this tracking already exists", "Duplicate tracking")
		return
	}

	mbOnline := models.MbOnline{
		Tracking: req.Tracking,
		UserID:   userIDUint,
	}

	// Create a new mb-online and return the response
	if err := moc.DB.Create(&mbOnline).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create mb-online", err.Error())
		return
	}

	// Load the created mb-online with relationships for complete response
	moc.DB.Preload("User").Preload("User.UserRoles.Role").
		Preload("User.UserRoles.Assigner").First(&mbOnline, mbOnline.ID)

	// Manually load order if tracking exists
	if mbOnline.Tracking != "" {
		var order models.Order
		if err := moc.DB.Where("tracking = ?", mbOnline.Tracking).
			Preload("OrderDetails").
			Preload("Picker.UserRoles.Role").
			Preload("Picker.UserRoles.Assigner").
			First(&order).Error; err == nil {
			mbOnline.Order = &order
		}
	}

	utils.SuccessResponse(c, http.StatusCreated, "Mb-online created successfully", mbOnline.ToMbOnlineResponse())
}

// Request/Response structs
type MbOnlinesListResponse struct {
	MbOnlines  []models.MbOnlineResponse `json:"mb_onlines"`
	Pagination utils.PaginationResponse  `json:"pagination"`
}

type CreateMbOnlineRequest struct {
	Tracking string `json:"tracking" binding:"required" example:"JNE1234567890"`
}
