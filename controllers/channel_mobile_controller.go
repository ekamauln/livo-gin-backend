package controllers

import (
	"livo-gin-backend/models"
	"livo-gin-backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ChannelMobileController struct {
	DB *gorm.DB
}

// NewChannelMobileController creates a new channel mobile controller
func NewChannelMobileController(db *gorm.DB) *ChannelMobileController {
	return &ChannelMobileController{DB: db}
}

// GetChannelMobiles godoc
// @Summary Get all channel mobiles
// @Description Get list of all channel mobiles (public access, no login required)
// @Tags channel-mobiles
// @Accept json
// @Produce json
// @Param search query string false "Search by channel mobile code or name (partial match)"
// @Success 200 {object} utils.Response{data=ChannelsMobileListResponse}
// @Router /api/mobile/channels [get]
func (cmc *ChannelMobileController) GetChannelMobiles(c *gin.Context) {
	// Parse search parameter
	search := c.Query("search")

	var channels []models.Channel
	var total int64

	// Build query with optional search
	query := cmc.DB.Model(&models.Channel{})

	if search != "" {
		// Search by channel code or name with partial match
		query = query.Where("code ILIKE ? OR name ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Get total count with search filter
	if err := query.Count(&total).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to count channels", err.Error())
		return
	}

	// Execute query to get all channels (no pagination)
	if err := query.Order("id ASC").Find(&channels).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve channels", err.Error())
		return
	}

	// Convert to response format
	channelResponses := make([]models.ChannelResponse, len(channels))
	for i, channel := range channels {
		channelResponses[i] = channel.ToChannelResponse()
	}

	response := ChannelsMobileListResponse{
		Channels: channelResponses,
		Total:    int(total),
	}

	// Build success message
	message := "Channels retrieved successfully"
	if search != "" {
		message += " (filtered by code or name: " + search + ")"
	}

	utils.SuccessResponse(c, http.StatusOK, message, response)
}

// Request/Response structs
type ChannelsMobileListResponse struct {
	Channels []models.ChannelResponse `json:"channels"`
	Total    int                      `json:"total"`
}
