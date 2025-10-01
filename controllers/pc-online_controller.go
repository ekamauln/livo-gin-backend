package controllers

import (
	"livo-gin-backend/models"
	"livo-gin-backend/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PcOnlineController struct {
	DB *gorm.DB
}

// NewPcOnlineController creates a new pc-online controller
func NewPcOnlineController(db *gorm.DB) *PcOnlineController {
	return &PcOnlineController{DB: db}
}

// GetPcOnlines godoc
// @Summary Get all pc-onlines for logged-in user
// @Description Get list of pc-onlines for current user filtered by current date (logged-in users only)
// @Tags pc-onlines
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search by tracking number"
// @Success 200 {object} utils.Response{data=PcOnlinesListResponse}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/pc-onlines [get]
func (pc *PcOnlineController) GetPcOnlines(c *gin.Context) {
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

	var pcOnlines []models.PcOnline
	var total int64

	// Build query with filters
	query := pc.DB.Model(&models.PcOnline{}).Where("user_id = ?", userID).Where("DATE(created_at) = CURRENT_DATE")

	if search != "" {
		// Search by tracking with partial match
		query = query.Where("tracking ILIKE ?", "%"+search+"%")
	}

	// Get total count with filters
	query.Count(&total)

	// Get pc-onlines with pagination, filters, and preload relationships
	if err := query.Order("id DESC").
		Preload("Details.Box").
		Preload("User.UserRoles.Role").
		Preload("Order.OrderDetails").
		Limit(limit).Offset(offset).Find(&pcOnlines).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve pc-onlines", err.Error())
		return
	}

	// Manually Load all orders at once using tracking field
	if len(pcOnlines) > 0 {
		trackingNumbers := make([]string, 0, len(pcOnlines))
		for _, pcOnline := range pcOnlines {
			if pcOnline.Tracking != "" {
				trackingNumbers = append(trackingNumbers, pcOnline.Tracking)
			}
		}

		var orders []models.Order
		orderMap := make(map[string]*models.Order)

		if len(trackingNumbers) > 0 {
			if err := pc.DB.Where("tracking IN ?", trackingNumbers).
				Preload("OrderDetails").
				Preload("Picker.UserRoles.Role").
				Preload("Picker.UserRoles.Assigner").
				Find(&orders).Error; err == nil {

				for i := range orders {
					orderMap[orders[i].Tracking] = &orders[i]
				}
			}
		}

		for i := range pcOnlines {
			if order, exists := orderMap[pcOnlines[i].Tracking]; exists {
				pcOnlines[i].Order = order
			}
		}
	}

	// Convert to response format
	pcOnlineResponses := make([]models.PcOnlineResponse, len(pcOnlines))
	for i, pcOnline := range pcOnlines {
		pcOnlineResponses[i] = pcOnline.ToPcOnlineResponse()
	}

	response := PcOnlinesListResponse{
		PcOnlines: pcOnlineResponses,
		Pagination: PaginationResponse{
			Page:  page,
			Limit: limit,
			Total: int(total),
		},
	}

	// Build success message
	message := "Pc-onlines retrieved successfully"
	if search != "" {
		message += " (filtered by tracking: " + search + ")"
	}

	utils.SuccessResponse(c, http.StatusOK, message, response)
}

// Request/Response structs
type PcOnlinesListResponse struct {
	PcOnlines  []models.PcOnlineResponse `json:"pc_onlines"`
	Pagination PaginationResponse        `json:"pagination"`
}

type PcOnlineDetailRequest struct {
	BoxID    uint `json:"box_id" binding:"required" example:"1"`
	Quantity int  `json:"quantity" binding:"required,min=1" example:"5"`
}

type CreatePcOnlineRequest struct {
	Tracking string                  `json:"tracking" binding:"required" example:"TRK123456"`
	Details  []PcOnlineDetailRequest `json:"details" binding:"required,dive,required"`
}
