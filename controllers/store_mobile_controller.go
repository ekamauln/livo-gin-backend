package controllers

import (
	"livo-gin-backend/models"
	"livo-gin-backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StoreMobileController struct {
	DB *gorm.DB
}

// NewStoreMobileController creates a new store mobile controller
func NewStoreMobileController(db *gorm.DB) *StoreMobileController {
	return &StoreMobileController{DB: db}
}

// GetStoreMobiles godoc
// @Summary Get all store mobiles
// @Description Get list of all store mobiles (public access, no login required)
// @Tags store-mobiles
// @Accept json
// @Produce json
// @Param search query string false "Search by store mobile tracking (partial match)"
// @Success 200 {object} utils.Response{data=StoreMobilesListResponse}
// @Router /api/mobile/stores [get]
func (smc *StoreMobileController) GetStoreMobiles(c *gin.Context) {
	// Parse search parameter
	search := c.Query("search")

	var stores []models.Store
	var total int64

	// Build query with optional search
	query := smc.DB.Model(&models.Store{})

	if search != "" {
		// Search by store mobile tracking with partial match
		query = query.Where("code ILIKE ? OR name ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Get total count with search filter
	if err := query.Count(&total).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to count store mobiles", err.Error())
		return
	}

	// Execute query to get store mobiles
	if err := query.Order("id ASC").Find(&stores).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve stores", err.Error())
		return
	}

	// Convert to response format
	storeResponses := make([]models.StoreResponse, len(stores))
	for i, store := range stores {
		storeResponses[i] = store.ToStoreResponse()
	}

	response := StoreMobilesListResponse{
		Stores: storeResponses,
		Total:  int(total),
	}

	// Build success message
	message := "Store mobiles retrieved successfully"
	if search != "" {
		message += " (filtered by code or name: " + search + ")"
	}

	utils.SuccessResponse(c, http.StatusOK, message, response)
}

// Request/Response structs
type StoreMobilesListResponse struct {
	Stores []models.StoreResponse `json:"stores"`
	Total  int                    `json:"total"`
}
