package controllers

import (
	"livo-gin-backend/models"
	"livo-gin-backend/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ExpeditionController struct {
	DB *gorm.DB
}

// NewExpeditionController creates a new expedition controller
func NewExpeditionController(db *gorm.DB) *ExpeditionController {
	return &ExpeditionController{DB: db}
}

// GetExpeditions godoc
// @Summary Get all expeditions
// @Description Get list of all expeditions (logged-in users only)
// @Tags expeditions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search by Code or Name (partial match)"
// @Success 200 {object} utils.Response{data=ExpeditionsListResponse}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/expeditions [get]
func (ec *ExpeditionController) GetExpeditions(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Parse search parameter
	search := c.Query("search")

	var expeditions []models.Expedition
	var total int64

	// Build query with optional search
	query := ec.DB.Model(&models.Expedition{})

	if search != "" {
		// Search by Code or Name with partial match
		query = query.Where("code ILIKE ? OR name ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Get total count with search filter
	if err := query.Count(&total).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to count expeditions", err.Error())
		return
	}

	// Get expeditions with pagination, search filter, and order by ID ascending
	if err := query.Order("id ASC").Limit(limit).Offset(offset).Find(&expeditions).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve expeditions", err.Error())
		return
	}

	// Convert to response format
	expeditionResponses := make([]models.ExpeditionResponse, len(expeditions))
	for i, expedition := range expeditions {
		expeditionResponses[i] = expedition.ToExpeditionResponse()
	}

	response := ExpeditionsListResponse{
		Expeditions: expeditionResponses,
		Pagination: PaginationResponse{
			Page:  page,
			Limit: limit,
			Total: int(total),
		},
	}

	// Build success message
	message := "Expeditions retrieved successfully"
	if search != "" {
		message += " (filtered by code or name: " + search + ")"
	}

	utils.SuccessResponse(c, http.StatusOK, message, response)
}

// GetExpedition godoc
// @Summary Get expedition by ID
// @Description Get expedition details by ID (logged-in users only)
// @Tags expeditions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Expedition ID"
// @Success 200 {object} utils.Response{data=models.ExpeditionResponse}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/expeditions/{id} [get]
func (ec *ExpeditionController) GetExpedition(c *gin.Context) {
	expeditionID := c.Param("id")

	var expedition models.Expedition
	if err := ec.DB.First(&expedition, expeditionID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Expedition not found", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Expedition retrieved successfully", expedition.ToExpeditionResponse())
}

// UpdateExpedition godoc
// @Summary Update expedition
// @Description Update expedition (logged-in users only)
// @Tags expeditions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Expedition ID"
// @Param expedition body UpdateExpeditionRequest true "Expedition data"
// @Success 200 {object} utils.Response{data=models.ExpeditionResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/expeditions/{id} [put]
func (ec *ExpeditionController) UpdateExpedition(c *gin.Context) {
	expeditionID := c.Param("id")

	var req UpdateExpeditionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	var expedition models.Expedition
	if err := ec.DB.First(&expedition, expeditionID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Expedition not found", err.Error())
		return
	}

	// Check for duplicate code (excluding current expedition)
	var existingExpedition models.Expedition
	if err := ec.DB.Where("code = ? AND id != ?", req.Code, expedition.ID).First(&existingExpedition).Error; err == nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Expedition code already exists", "A expedition with this code already exists")
		return
	}

	// Update expedition fields
	expedition.Code = req.Code
	expedition.Name = req.Name
	expedition.Color = req.Color
	expedition.Slug = req.Slug

	if err := ec.DB.Save(&expedition).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update expedition", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Expedition updated successfully", expedition.ToExpeditionResponse())
}

// RemoveExpedition godoc
// @Summary Remove expedition
// @Description Soft delete expedition (logged-in users only)
// @Tags expeditions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Expedition ID"
// @Success 200 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/expeditions/{id} [delete]
func (ec *ExpeditionController) RemoveExpedition(c *gin.Context) {
	expeditionID := c.Param("id")

	var expedition models.Expedition
	if err := ec.DB.First(&expedition, expeditionID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Expedition not found", err.Error())
		return
	}

	if err := ec.DB.Delete(&expedition).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove expedition", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Expedition removed successfully", nil)
}

// CreateExpedition godoc
// @Summary Create new expedition
// @Description Create a new expedition (logged-in users only)
// @Tags expeditions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param expedition body CreateExpeditionRequest true "Create expedition request"
// @Success 201 {object} utils.Response{data=models.ExpeditionResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/expeditions [post]
func (ec *ExpeditionController) CreateExpedition(c *gin.Context) {
	var req CreateExpeditionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	expedition := models.Expedition{
		Code:  req.Code,
		Name:  req.Name,
		Slug:  req.Slug,
		Color: req.Color,
	}

	// Check for duplicate expedition code
	var existingExpedition models.Expedition
	if err := ec.DB.Where("code = ?", req.Code).First(&existingExpedition).Error; err == nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Expedition code already exists", "A expedition with this code already exists")
		return
	}

	// Create a new expedition and return the response
	if err := ec.DB.Create(&expedition).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create expedition", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Expedition created successfully", expedition.ToExpeditionResponse())
}

// Request/Response structs
type ExpeditionsListResponse struct {
	Expeditions []models.ExpeditionResponse `json:"expeditions"`
	Pagination  PaginationResponse          `json:"pagination"`
}

type UpdateExpeditionRequest struct {
	Code  string `json:"code" binding:"required"`
	Name  string `json:"name" binding:"required"`
	Color string `json:"color" binding:"required"`
	Slug  string `json:"slug" binding:"required"`
}

type CreateExpeditionRequest struct {
	Code  string `json:"code" binding:"required"`
	Name  string `json:"name" binding:"required"`
	Slug  string `json:"slug" binding:"required"`
	Color string `json:"color" binding:"required"`
}
