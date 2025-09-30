package controllers

import (
	"livo-gin-backend/models"
	"livo-gin-backend/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ChannelController struct {
	DB *gorm.DB
}

// NewChannelController creates a new channel controller
func NewChannelController(db *gorm.DB) *ChannelController {
	return &ChannelController{DB: db}
}

// GetChannels godoc
// @Summary Get all channels
// @Description Get list of all channels (logged-in users only)
// @Tags channels
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search by Code or Name (partial match)"
// @Success 200 {object} utils.Response{data=ChannelsListResponse}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/channels [get]
func (cc *ChannelController) GetChannels(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Parse search parameter
	search := c.Query("search")

	var channels []models.Channel
	var total int64

	// Build query with optional search
	query := cc.DB.Model(&models.Channel{})

	if search != "" {
		// Search by Code or Name with partial match
		query = query.Where("code ILIKE ? OR name ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Get total count with search filter
	query.Count(&total)

	// Get channels with pagination, search filter, and order by ID ascending
	if err := query.Order("id ASC").Limit(limit).Offset(offset).Find(&channels).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve channels", err.Error())
		return
	}

	// Convert to response format
	channelResponses := make([]models.ChannelResponse, len(channels))
	for i, channel := range channels {
		channelResponses[i] = channel.ToChannelResponse()
	}

	response := ChannelsListResponse{
		Channels: channelResponses,
		Pagination: PaginationResponse{
			Page:  page,
			Limit: limit,
			Total: int(total),
		},
	}

	// Build success message
	message := "Channels retrieved successfully"
	if search != "" {
		message += " (filtered by code or name: " + search + ")"
	}

	utils.SuccessResponse(c, http.StatusOK, message, response)
}

// GetChannel godoc
// @Summary Get channel by ID
// @Description Get specific channel by ID (logged-in users only)
// @Tags channels
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Channel ID"
// @Success 200 {object} utils.Response{data=models.ChannelResponse}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/channels/{id} [get]
func (cc *ChannelController) GetChannel(c *gin.Context) {
	channelID := c.Param("id")

	var channel models.Channel
	if err := cc.DB.First(&channel, channelID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Channel not found", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Channel retrieved successfully", channel.ToChannelResponse())
}

// UpdateChannel godoc
// @Summary Update channel
// @Description Update specific channel information (logged-in users only)
// @Tags channels
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Channel ID"
// @Param channel body UpdateChannelRequest true "Update channel request"
// @Success 200 {object} utils.Response{data=models.ChannelResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/channels/{id} [put]
func (cc *ChannelController) UpdateChannel(c *gin.Context) {
	channelID := c.Param("id")

	var req UpdateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	var channel models.Channel
	if err := cc.DB.First(&channel, channelID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Channel not found", err.Error())
		return
	}

	// Check for duplicate channel code (excluding current channel)
	var existingChannel models.Channel
	if err := cc.DB.Where("code = ? AND id <> ?", req.Code, channelID).First(&existingChannel).Error; err == nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Channel code already exists", "A channel with this code already exists")
		return
	}

	// Update channel fields
	channel.Code = req.Code
	channel.Name = req.Name

	if err := cc.DB.Save(&channel).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update channel", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Channel updated successfully", channel.ToChannelResponse())
}

// RemoveChannel godoc
// @Summary Remove channel
// @Description Soft delete a channel (logged-in users only)
// @Tags channels
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Channel ID"
// @Success 200 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/channels/{id} [delete]
func (cc *ChannelController) RemoveChannel(c *gin.Context) {
	channelID := c.Param("id")

	var channel models.Channel
	if err := cc.DB.First(&channel, channelID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Channel not found", err.Error())
		return
	}

	if err := cc.DB.Delete(&channel).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove channel", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Channel removed successfully", nil)
}

// CreateChannel godoc
// @Summary Create new channel
// @Description Create a new channel (logged-in users only)
// @Tags channels
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param channel body CreateChannelRequest true "Create channel request"
// @Success 201 {object} utils.Response{data=models.ChannelResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/channels [post]
func (cc *ChannelController) CreateChannel(c *gin.Context) {
	var req CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	channel := models.Channel{
		Code: req.Code,
		Name: req.Name,
	}

	// Check for duplicate channel code
	var existingChannel models.Channel
	if err := cc.DB.Where("code = ?", req.Code).First(&existingChannel).Error; err == nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Channel code already exists", "A channel with this code already exists")
		return
	}

	// Create a new channel and return the response
	if err := cc.DB.Create(&channel).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create channel", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Channel created successfully", channel.ToChannelResponse())
}

// Request/Response structs
type ChannelsListResponse struct {
	Channels   []models.ChannelResponse `json:"channels"`
	Pagination PaginationResponse       `json:"pagination"`
}

type UpdateChannelRequest struct {
	Code string `json:"code" binding:"required"`
	Name string `json:"name" binding:"required"`
}

type CreateChannelRequest struct {
	Code string `json:"code" binding:"required"`
	Name string `json:"name" binding:"required"`
}
