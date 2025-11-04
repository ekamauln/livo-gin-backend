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

type ReturnMobileController struct {
	DB *gorm.DB
}

// NewReturnMobileController creates a new return mobile controller
func NewReturnMobileController(db *gorm.DB) *ReturnMobileController {
	return &ReturnMobileController{DB: db}
}

// GetReturnMobiles godoc
// @Summary Get all return mobiles
// @Description Get list of all return mobiles (public access, no login required)
// @Tags return-mobiles
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search by return mobile tracking (partial match)"
// @Success 200 {object} utils.Response{data=ReturnMobilesListResponse}
// @Router /api/mobile/returns [get]
func (rmc *ReturnMobileController) GetReturnMobiles(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Parse search parameter
	search := c.Query("search")

	var returnMobiles []models.Return
	var total int64

	// Build query with optional search
	query := rmc.DB.Model(&models.Return{})

	if search != "" {
		// Search by return mobile tracking with partial match
		query = query.Where("tracking ILIKE ?", "%"+search+"%")
	}

	// Get total count with search filter
	if err := query.Count(&total).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to count return", err.Error())
		return
	}

	// Get return mobiles with pagination, search filter, and order by ID descending
	if err := query.Order("id DESC").Limit(limit).Offset(offset).Find(&returnMobiles).Error; err != nil {
		utils.ErrorResponse(c, 500, "Failed to fetch return mobiles", err.Error())
		return
	}

	// Convert to response format
	returnMobileResponses := make([]models.ReturnMobileResponse, len(returnMobiles))
	for i, returnMobile := range returnMobiles {
		returnMobileResponses[i] = returnMobile.ToReturnMobileResponse()
	}

	response := ReturnMobilesListResponse{
		ReturnMobiles: returnMobileResponses,
		Pagination: utils.PaginationResponse{
			Page:  page,
			Limit: limit,
			Total: int(total),
		},
	}

	// Build success message
	message := "Return mobiles retrieved successfully"
	if search != "" {
		message += " (filtered by tracking: " + search + ")"
	}

	utils.SuccessResponse(c, http.StatusOK, message, response)
}

// GetReturnMobile godoc
// @Summary Get a return mobile by ID
// @Description Get a return mobile by ID (public access, no login required)
// @Tags return-mobiles
// @Accept json
// @Produce json
// @Param id path int true "Return Mobile ID"
// @Success 200 {object} utils.Response{data=models.ReturnMobileResponse}
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/mobile/returns/{id} [get]
func (rmc *ReturnMobileController) GetReturnMobile(c *gin.Context) {
	returnMobileID := c.Param("id")

	var returnMobile models.Return
	if err := rmc.DB.First(&returnMobile, returnMobileID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Return not found", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Return mobile retrieved successfully", returnMobile.ToReturnMobileResponse())
}

// CreateReturnMobile godoc
// @Summary Create a new return mobile
// @Description Create a new return mobile (public access, no login required)
// @Tags return-mobiles
// @Accept json
// @Produce json
// @Param return_mobile body CreateReturnMobileRequest true "Create return mobile request"
// @Success 201 {object} utils.Response{data=models.ReturnMobileResponse}
// @Failure 400 {object} utils.Response
// @Router /api/mobile/returns [post]
func (rmc *ReturnMobileController) CreateReturnMobile(c *gin.Context) {
	var req CreateReturnMobileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Convert tracking to uppercase and trim spaces
	req.Tracking = strings.ToUpper(strings.TrimSpace(req.Tracking))

	returnMobile := models.Return{
		NewTracking: req.Tracking,
		ChannelID:   req.ChannelID,
		StoreID:     req.StoreID,
	}

	// Check for duplicate tracking
	var existingReturnMobile models.Return
	if err := rmc.DB.Where("new_tracking = ?", req.Tracking).First(&existingReturnMobile).Error; err == nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Return mobile tracking already exists", "A return mobile with this tracking already exists")
		return
	}

	// Create a new return mobile and return the response
	if err := rmc.DB.Create(&returnMobile).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create return mobile", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Return mobile created successfully", returnMobile.ToReturnMobileResponse())
}

// Request/Response structs
type ReturnMobilesListResponse struct {
	ReturnMobiles []models.ReturnMobileResponse `json:"return_mobiles"`
	Pagination    utils.PaginationResponse      `json:"pagination"`
}

type CreateReturnMobileRequest struct {
	Tracking  string `json:"tracking" binding:"required"`
	ChannelID uint   `json:"channel_id" binding:"required"`
	StoreID   uint   `json:"store_id" binding:"required"`
}
