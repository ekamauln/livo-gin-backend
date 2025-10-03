package controllers

import (
	"livo-gin-backend/models"
	"livo-gin-backend/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ReturnController struct {
	DB *gorm.DB
}

// NewReturnController creates a new return controller
func NewReturnController(db *gorm.DB) *ReturnController {
	return &ReturnController{DB: db}
}

// GetReturns godoc
// @Summary Get all returns
// @Description Get a list of all returns (logged in users only)
// @Tags returns
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(10)
// @Param search query string false "Search by return new tracking (partial match)"
// @Success 200 {object} utils.Response{data=ReturnsListResponse}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/returns [get]
func (rc *ReturnController) GetReturns(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	offset := (page - 1) * limit

	// Parse search parameter
	search := c.Query("search")

	var rets []models.Return
	var total int64

	// Build query with optional search
	query := rc.DB.Model(&models.Return{})

	if search != "" {
		// Seach by return new tracking with partial match
		query = query.Where("new_tracking ILIKE ?", "%"+search+"%")
	}

	// Get total count with search filter
	query.Count(&total)

	// Get returns with pagination, search filter, and order by ID desc
	if err := query.Order("id DESC").Limit(limit).Offset(offset).Find(&rets).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve returns", err.Error())
		return
	}

	// Convert to response format
	returnResponse := make([]models.ReturnResponse, len(rets))
	for i, ret := range rets {
		returnResponse[i] = ret.ToReturnResponse()
	}

	response := ReturnsListResponse{
		Returns: returnResponse,
		Pagination: PaginationResponse{
			Page:  page,
			Limit: limit,
			Total: int(total),
		},
	}

	// Build success message
	message := "Returns retrieved successfully"
	if search != "" {
		message += " (filtered by new tracking: " + search + ")"
	}

	utils.SuccessResponse(c, http.StatusOK, message, response)
}

// GetReturn godoc
// @Summary Get return by ID
// @Description Get return details by ID (logged in users only)
// @Tags returns
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Return ID"
// @Success 200 {object} utils.Response{data=models.ReturnResponse}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/returns/{id} [get]
func (rc *ReturnController) GetReturn(c *gin.Context) {
	returnID := c.Param("id")

	var ret models.Return
	if err := rc.DB.Preload("Order").Preload("Channel").Preload("Store").First(&ret, returnID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Return not found", err.Error())
	}

	utils.SuccessResponse(c, http.StatusOK, "Return retrieved successfully", ret.ToReturnResponse())
}

// CreateBaseReturn godoc
// @Summary Create a new base return
// @Description Create a new base return (logged in users only)
// @Tags returns
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateBaseReturnRequest true "Create Base Return Request"
// @Success 201 {object} utils.Response{data=models.ReturnResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/returns [post]
func (rc *ReturnController) CreateBaseReturn(c *gin.Context) {
	var req CreateBaseReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ret := models.Return{
		NewTracking: req.NewTracking,
		ChannelID:   req.ChannelID,
		StoreID:     req.StoreID,
	}

	// Check for duplicate new tracking
	var existingReturn models.Return
	if err := rc.DB.Where("new_tracking = ?", req.NewTracking).First(&existingReturn).Error; err == nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tracking already exists", "Return with this new tracking already exists")
		return
	}

	// Create a new base return and return the response
	if err := rc.DB.Create(&ret).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create return", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Return created successfully", ret.ToReturnResponse())
}

// UpdateDataReturn godoc
// @Summary Update return data
// @Description Update return data (logged in users only)
// @Tags returns
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Return ID"
// @Param request body UpdateDataReturnRequest true "Update Data Return Request"
// @Success 200 {object} utils.Response{data=models.ReturnResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/returns/{id}/data [put]
func (rc *ReturnController) UpdateDataReturn(c *gin.Context) {
	returnID := c.Param("id")

	var req UpdateDataReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	var ret models.Return
	if err := rc.DB.First(&ret, returnID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Return not found", err.Error())
		return
	}

	// Check for duplicate old tracking if changed
	var existingReturn models.Return
	if err := rc.DB.Where("old_tracking = ?", req.OldTracking, returnID).First(&existingReturn).Error; err == nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Old tracking already exists", "Return with this old tracking already exists")
		return
	}

	// Update return data fields
	ret.OldTracking = req.OldTracking
	ret.ReturnType = req.ReturnType
	ret.ReturnReason = req.ReturnReason

	if err := rc.DB.Save(&ret).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update return", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Return updated successfully", ret.ToReturnResponse())
}

// UpdateAdminReturn godoc
// @Summary Update return admin data
// @Description Update return admin (logged in users only)
// @Tags returns
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Return ID"
// @Param request body UpdateAdminReturnRequest true "Update Admin Return Request"
// @Success 200 {object} utils.Response{data=models.ReturnResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/returns/{id}/admin [put]
func (rc *ReturnController) UpdateAdminReturn(c *gin.Context) {
	returnID := c.Param("id")

	var req UpdateAdminReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	var ret models.Return
	if err := rc.DB.First(&ret, returnID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Return not found", err.Error())
		return
	}

	// Check for duplicate old tracking if changed
	var existingReturn models.Return
	if err := rc.DB.Where("old_tracking = ?", req.OldTracking, returnID).First(&existingReturn).Error; err == nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Old tracking already exists", "Return with this old tracking already exists")
		return
	}

	// Update return admin fields
	ret.OldTracking = req.OldTracking
	ret.ReturnType = req.ReturnType
	ret.ReturnReason = req.ReturnReason
	ret.ReturnNumber = req.ReturnNumber
	ret.ScrapNumber = req.ScrapNumber

	if err := rc.DB.Save(&ret).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update return", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Return updated successfully", ret.ToReturnResponse())
}

// Request/Response structs
type ReturnsListResponse struct {
	Returns    []models.ReturnResponse `json:"returns"`
	Pagination PaginationResponse      `json:"pagination"`
}

type CreateBaseReturnRequest struct {
	NewTracking string `json:"new_tracking" binding:"required"`
	ChannelID   uint   `json:"channel_id" binding:"required"`
	StoreID     uint   `json:"store_id" binding:"required"`
}

type UpdateDataReturnRequest struct {
	OldTracking  string `json:"old_tracking" binding:"required"`
	ReturnType   string `json:"return_type" binding:"required"`
	ReturnReason string `json:"return_reason" binding:"required"`
}

type UpdateAdminReturnRequest struct {
	OldTracking  string `json:"old_tracking" binding:"required"`
	ReturnType   string `json:"return_type" binding:"required"`
	ReturnReason string `json:"return_reason" binding:"required"`
	ReturnNumber string `json:"return_number" binding:"required"`
	ScrapNumber  string `json:"scrap_number" binding:"required"`
}
