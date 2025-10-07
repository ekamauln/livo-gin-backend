package controllers

import (
	"livo-gin-backend/models"
	"livo-gin-backend/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StoreController struct {
	DB *gorm.DB
}

// NewStoreController creates a new store controller
func NewStoreController(db *gorm.DB) *StoreController {
	return &StoreController{DB: db}
}

// GetStores godoc
// @Summary Get all stores
// @Description Get list of all stores with optional search by Code or Name (logged-in users only)
// @Tags stores
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search by Code or Name (partial match)"
// @Success 200 {object} utils.Response{data=StoresListResponse}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/stores [get]
func (sc *StoreController) GetStores(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Parse search parameter
	search := c.Query("search")

	var stores []models.Store
	var total int64

	// Build query with optional search
	query := sc.DB.Model(&models.Store{})

	if search != "" {
		// Search by Code or Name with partial match
		query = query.Where("code ILIKE ? OR name ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Get total count with search filter
	if err := query.Count(&total).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to count stores", err.Error())
		return
	}

	// Get stores with pagination, search filter, and order by ID ascending
	if err := query.Order("id ASC").Limit(limit).Offset(offset).Find(&stores).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve stores", err.Error())
		return
	}

	// Convert to response format
	storeResponses := make([]models.StoreResponse, len(stores))
	for i, store := range stores {
		storeResponses[i] = store.ToStoreResponse()
	}

	response := StoresListResponse{
		Stores: storeResponses,
		Pagination: PaginationResponse{
			Page:  page,
			Limit: limit,
			Total: int(total),
		},
	}

	// Build success message
	message := "Stores retrieved successfully"
	if search != "" {
		message += " (filtered by code or name: " + search + ")"
	}

	utils.SuccessResponse(c, http.StatusOK, message, response)
}

// GetStore godoc
// @Summary Get store by ID
// @Description Get specific store by ID (logged-in users only)
// @Tags stores
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Store ID"
// @Success 200 {object} utils.Response{data=models.StoreResponse}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/stores/{id} [get]
func (sc *StoreController) GetStore(c *gin.Context) {
	storeID := c.Param("id")

	var store models.Store
	if err := sc.DB.First(&store, storeID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Store not found", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Store retrieved successfully", store.ToStoreResponse())
}

// UpdateStore godoc
// @Summary Update store
// @Description Update store information (logged-in users only)
// @Tags stores
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Store ID"
// @Param store body UpdateStoreRequest true "Update Store Request"
// @Success 200 {object} utils.Response{data=models.StoreResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/stores/{id} [put]
func (sc *StoreController) UpdateStore(c *gin.Context) {
	storeID := c.Param("id")

	var req UpdateStoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	var store models.Store
	if err := sc.DB.First(&store, storeID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Store not found", err.Error())
		return
	}

	// Check for duplicate store code (excluding current store)
	var existingStore models.Store
	if err := sc.DB.Where("code = ? AND id <> ?", req.Code, store.ID).First(&existingStore).Error; err == nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Store code already exists", "A store with this code already exists")
		return
	}

	// Update store fields
	store.Code = req.Code
	store.Name = req.Name

	if err := sc.DB.Save(&store).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update store", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Store updated successfully", store.ToStoreResponse())
}

// RemoveStore godoc
// @Summary Remove store
// @Description Soft delete a store (logged-in users only)
// @Tags stores
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Store ID"
// @Success 200 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/stores/{id} [delete]
func (sc *StoreController) RemoveStore(c *gin.Context) {
	storeID := c.Param("id")

	var store models.Store
	if err := sc.DB.First(&store, storeID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Store not found", err.Error())
		return
	}

	if err := sc.DB.Delete(&store).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove store", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Store removed successfully", nil)
}

// CreateStore godoc
// @Summary Create new store
// @Description Create a new store (logged-in users only)
// @Tags stores
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param store body CreateStoreRequest true "Create Store Request"
// @Success 201 {object} utils.Response{data=models.StoreResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/stores [post]
func (sc *StoreController) CreateStore(c *gin.Context) {
	var req CreateStoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	store := models.Store{
		Code: req.Code,
		Name: req.Name,
	}
	// Check for duplicate store code
	var existingStore models.Store
	if err := sc.DB.Where("code = ?", req.Code).First(&existingStore).Error; err == nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Store code already exists", "A store with this code already exists")
		return
	}

	// Create new store and return response
	if err := sc.DB.Create(&store).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create store", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Store created successfully", store.ToStoreResponse())
}

// Request/Response structs
type StoresListResponse struct {
	Stores     []models.StoreResponse `json:"stores"`
	Pagination PaginationResponse     `json:"pagination"`
}

type UpdateStoreRequest struct {
	Code string `json:"code" binding:"required"`
	Name string `json:"name" binding:"required"`
}

type CreateStoreRequest struct {
	Code string `json:"code" binding:"required"`
	Name string `json:"name" binding:"required"`
}
