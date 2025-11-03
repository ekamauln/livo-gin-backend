package controllers

import (
	"livo-gin-backend/models"
	"livo-gin-backend/utils"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MbRibbonController struct {
	DB *gorm.DB
}

// NewMbRibbonController creates a new mb-ribbon controller
func NewMbRibbonController(db *gorm.DB) *MbRibbonController {
	return &MbRibbonController{DB: db}
}

// GetMbRibbons godoc
// @Summary Get all mb-ribbons for logged-in user
// @Description Get list of mb-ribbons for current user filtered by current date (logged-in users only)
// @Tags mb-ribbons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search by tracking number"
// @Success 200 {object} utils.Response{data=MbRibbonsListResponse}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/mb-ribbons [get]
func (mrc *MbRibbonController) GetMbRibbons(c *gin.Context) {
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

	var mbRibbons []models.MbRibbon
	var total int64

	// Build query with filters
	query := mrc.DB.Model(&models.MbRibbon{}).Where("user_id = ?", userID).Where("DATE(created_at) = CURRENT_DATE")

	if search != "" {
		// Search by tracking with partial match
		query = query.Where("tracking ILIKE ?", "%"+search+"%")
	}

	// Get total count with filters
	if err := query.Count(&total).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to count mb-ribbons", err.Error())
		return
	}

	// Get mbRibbons with pagination, search filter, and order by ID descending
	if err := query.Order("id DESC").Preload("User.UserRoles.Role").Preload("User.UserRoles.Assigner").Limit(limit).Offset(offset).Find(&mbRibbons).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve mb-ribbons", err.Error())
		return
	}

	// Manually load order data for each mb-ribbon
	for i := range mbRibbons {
		if mbRibbons[i].Tracking != "" {
			var order models.Order
			if err := mrc.DB.Where("tracking = ?", mbRibbons[i].Tracking).
				Preload("Picker.UserRoles.Role").
				Preload("Picker.UserRoles.Assigner").
				First(&order).Error; err == nil {
				mbRibbons[i].Order = &order
			}
		}
	}

	// Convert to response format
	mbRibbonResponses := make([]models.MbRibbonResponse, len(mbRibbons))
	for i, mbRibbon := range mbRibbons {
		mbRibbonResponses[i] = mbRibbon.ToMbRibbonResponse()
	}

	response := MbRibbonsListResponse{
		MbRibbons: mbRibbonResponses,
		Pagination: utils.PaginationResponse{
			Page:  page,
			Limit: limit,
			Total: int(total),
		},
	}

	// Build success message
	message := "Mb-ribbons retrieved successfully"
	if search != "" {
		message += " (filtered by tracking: " + search + ")"
	}

	utils.SuccessResponse(c, http.StatusOK, message, response)
}

// GetMbRibbon godoc
// @Summary Get a specific mb-ribbon by ID
// @Description Get a specific mb-ribbon by ID (logged-in users only)
// @Tags mb-ribbons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Mb-ribbon ID"
// @Success 200 {object} utils.Response{data=models.MbRibbonResponse}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/mb-ribbons/{id} [get]
func (mrc *MbRibbonController) GetMbRibbon(c *gin.Context) {
	mbRibbonID := c.Param("id")

	var mbRibbon models.MbRibbon
	if err := mrc.DB.Preload("User").First(&mbRibbon, mbRibbonID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Mb-ribbon not found", err.Error())
		return
	}

	// Manually load order if tracking exists
	if mbRibbon.Tracking != "" {
		var order models.Order
		if err := mrc.DB.Where("tracking = ?", mbRibbon.Tracking).
			Preload("OrderDetails").
			Preload("Picker.UserRoles.Role").
			Preload("Picker.UserRoles.Assigner").First(&order).Error; err == nil {
			mbRibbon.Order = &order
		}
	}

	utils.SuccessResponse(c, http.StatusOK, "Mb-ribbon retrieved successfully", mbRibbon.ToMbRibbonResponse())
}

// CreateMbRibbon godoc
// @Summary Create new mb-ribbon
// @Description Create a new mb-ribbon entry (logged-in users only)
// @Tags mb-ribbons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param mb_ribbon body CreateMbRibbonRequest true "Mb-ribbon data"
// @Success 201 {object} utils.Response{data=models.MbRibbonResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/mb-ribbons [post]
func (mrc *MbRibbonController) CreateMbRibbon(c *gin.Context) {
	// Get user ID from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "User not authenticated")
		return
	}

	var req CreateMbRibbonRequest
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
	if err := mrc.DB.Where("tracking = ?", req.Tracking).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Order not found", "No order found with the specified tracking number")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to check order", err.Error())
		return
	}

	// Check if tracking already exists in mb-online table
	var existingMbOnline models.MbOnline
	if err := mrc.DB.Where("tracking = ?", req.Tracking).First(&existingMbOnline).Error; err == nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tracking already processed in MB-Online", "This tracking number is already being processed in the online workflow")
		return
	}

	// Check for duplicate tracking
	var existingMbRibbon models.MbRibbon
	if err := mrc.DB.Where("tracking = ?", req.Tracking).First(&existingMbRibbon).Error; err == nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Mb-ribbon with this tracking already exists", "Duplicate tracking")
		return
	}

	mbRibbon := models.MbRibbon{
		Tracking: req.Tracking,
		UserID:   userIDUint,
	}

	// Create a new mb-ribbon and return the response
	if err := mrc.DB.Create(&mbRibbon).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create mb-ribbon", err.Error())
		return
	}

	// Load the created mb-ribbon with relationships for complete response
	mrc.DB.Preload("User").Preload("User.UserRoles.Role").Preload("User.UserRoles.Assigner").First(&mbRibbon, mbRibbon.ID)

	// Manually load order if tracking exists
	if mbRibbon.Tracking != "" {
		var order models.Order
		if err := mrc.DB.Where("tracking = ?", mbRibbon.Tracking).
			Preload("OrderDetails").
			Preload("Picker.UserRoles.Role").
			Preload("Picker.UserRoles.Assigner").
			First(&order).Error; err == nil {
			mbRibbon.Order = &order
		}
	}

	utils.SuccessResponse(c, http.StatusCreated, "Mb-ribbon created successfully", mbRibbon.ToMbRibbonResponse())
}

// GetChartMbRibbons godoc
// @Summary Get mb-ribbon counts per day for current month
// @Description Get daily count of mb-ribbons for current month (for chart data, logged-in users only)
// @Tags mb-ribbons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=MbRibbonsDailyCountResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/mb-ribbons/chart [get]
func (mrc *MbRibbonController) GetChartMbRibbons(c *gin.Context) {
	// Get current month start and end dates
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()

	// First day of current month at 00:00:00
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)

	// First day of next month at 00:00:00 (to use as upper bound)
	firstOfNextMonth := firstOfMonth.AddDate(0, 1, 0)

	// Query to get daily counts for current month
	var dailyCounts []MbRibbonDailyCount

	if err := mrc.DB.Model(&models.MbRibbon{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at >= ?", firstOfMonth).
		Where("created_at < ?", firstOfNextMonth).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&dailyCounts).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve mb-ribbon counts", err.Error())
		return
	}

	// Get total count for current month
	var totalCount int64
	if err := mrc.DB.Model(&models.MbRibbon{}).
		Where("created_at >= ?", firstOfMonth).
		Where("created_at < ?", firstOfNextMonth).
		Count(&totalCount).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to count mb-ribbons", err.Error())
		return
	}

	response := MbRibbonsDailyCountResponse{
		Month:       currentMonth.String(),
		Year:        currentYear,
		DailyCounts: dailyCounts,
		TotalCount:  int(totalCount),
	}

	message := "Mb-ribbon daily counts for " + currentMonth.String() + " " + strconv.Itoa(currentYear) + " retrieved successfully"

	utils.SuccessResponse(c, http.StatusOK, message, response)
}

// Request/Response structs
type MbRibbonsListResponse struct {
	MbRibbons  []models.MbRibbonResponse `json:"mb_ribbons"`
	Pagination utils.PaginationResponse  `json:"pagination"`
}

type CreateMbRibbonRequest struct {
	Tracking string `json:"tracking" binding:"required" example:"JNE1234567890"`
}

// MbRibbonsDailyCount represents the count of mb-ribbons for a specific date
type MbRibbonDailyCount struct {
	Date  string `json:"date"` // Format: YYYY-MM-DD
	Count int    `json:"count"`
}

// MbRibbonsDailyCountResponse represents the response for daily mb-ribbons counts
type MbRibbonsDailyCountResponse struct {
	Month       string               `json:"month"` // e.g., "November"
	Year        int                  `json:"year"`  // e.g., 2025
	DailyCounts []MbRibbonDailyCount `json:"daily_counts"`
	TotalCount  int                  `json:"total_count"` // Total for the month
}
