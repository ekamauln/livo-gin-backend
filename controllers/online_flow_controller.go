package controllers

import (
	"livo-gin-backend/models"
	"livo-gin-backend/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OnlineFlowController struct {
	DB *gorm.DB
}

// NewOnlineFlowController creates a new online flow controller
func NewOnlineFlowController(db *gorm.DB) *OnlineFlowController {
	return &OnlineFlowController{DB: db}
}

// GetOnlineFlows godoc
// @Summary Get all online flows
// @Description Get all online flows with pagination and search, primary tracking from mb-online (logged-in users only)
// @Tags online-flow
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search by tracking number"
// @Success 200 {object} utils.Response{data=OnlineFlowsListResponse}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/online-flow [get]
func (ofc *OnlineFlowController) GetOnlineFlows(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Parse search parameter (optional)
	search := c.Query("search")

	var trackingNumbers []string
	var total int64

	// CHANGED: Get tracking numbers primarily from mb_onlines
	query := ofc.DB.Model(&models.MbOnline{}).Select("DISTINCT tracking").Where("tracking IS NOT NULL AND tracking != ''")

	// Add search filter if provided
	if search != "" {
		query = query.Where("tracking ILIKE ?", "%"+search+"%")
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to count tracking numbers", err.Error())
		return
	}

	// Get paginated tracking numbers
	if err := query.Order("tracking").Limit(limit).Offset(offset).Pluck("tracking", &trackingNumbers).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve tracking numbers", err.Error())
		return
	}

	// Build online flows for each tracking
	var onlineFlows []OnlineFlowResponse
	for _, tracking := range trackingNumbers {
		flow := ofc.buildOnlineFlow(tracking)
		onlineFlows = append(onlineFlows, flow)
	}

	response := OnlineFlowsListResponse{
		OnlineFlows: onlineFlows,
		Pagination: OnlineFlowPaginationResponse{ // FIXED: Use unique pagination
			Page:  page,
			Limit: limit,
			Total: int(total),
		},
	}

	// Build success message
	message := "Online flows retrieved successfully"
	if search != "" {
		message += " (filtered by tracking: " + search + ")"
	}

	utils.SuccessResponse(c, http.StatusOK, message, response)
}

// GetOnlineFlow godoc
// @Summary Get online flow tracking
// @Description Get the complete flow tracking through online process (mb-online -> qc-online -> pc-online -> order) (logged-in users only)
// @Tags online-flow
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tracking path string true "Tracking number"
// @Success 200 {object} utils.Response{data=OnlineFlowResponse}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/online-flow/{tracking} [get]
func (ofc *OnlineFlowController) GetOnlineFlow(c *gin.Context) {
	tracking := c.Param("tracking")

	if tracking == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid tracking", "Tracking number is required")
		return
	}

	flow := ofc.buildOnlineFlow(tracking)

	// CHANGED: Check if mb-online exists (since it's the primary source)
	if flow.MbOnline == nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Tracking not found", "No mb-online record found for the specified tracking number")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Online flow retrieved successfully", flow)
}

// Helper function to build online flow for a tracking number
func (ofc *OnlineFlowController) buildOnlineFlow(tracking string) OnlineFlowResponse {
	var response OnlineFlowResponse
	response.Tracking = tracking

	// 1. Query MB Online (PRIMARY SOURCE)
	var mbOnline models.MbOnline
	if err := ofc.DB.Preload("User").Where("tracking = ?", tracking).First(&mbOnline).Error; err == nil {
		var user *OnlineUserFlowInfo
		if mbOnline.User != nil {
			user = &OnlineUserFlowInfo{
				ID:       mbOnline.User.ID,
				Username: mbOnline.User.Username,
				FullName: mbOnline.User.FullName,
			}
		}

		response.MbOnline = &MbOnlineFlowInfo{
			User:      user,
			CreatedAt: mbOnline.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	// 2. Query QC Online
	var qcOnline models.QcOnline
	if err := ofc.DB.Preload("User").Where("tracking = ?", tracking).First(&qcOnline).Error; err == nil {
		var user *OnlineUserFlowInfo
		if qcOnline.User != nil {
			user = &OnlineUserFlowInfo{
				ID:       qcOnline.User.ID,
				Username: qcOnline.User.Username,
				FullName: qcOnline.User.FullName,
			}
		}

		response.QcOnline = &QcOnlineFlowInfo{
			User:      user,
			CreatedAt: qcOnline.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	// 3. Query PC Online
	var pcOnline models.PcOnline
	if err := ofc.DB.Preload("User").Where("tracking = ?", tracking).First(&pcOnline).Error; err == nil {
		var user *OnlineUserFlowInfo
		if pcOnline.User != nil {
			user = &OnlineUserFlowInfo{
				ID:       pcOnline.User.ID,
				Username: pcOnline.User.Username,
				FullName: pcOnline.User.FullName,
			}
		}

		response.PcOnline = &PcOnlineFlowInfo{
			User:      user,
			CreatedAt: pcOnline.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	// 4. Query Order (LAST)
	var order models.Order
	if err := ofc.DB.Where("tracking = ?", tracking).First(&order).Error; err == nil {
		response.Order = &OnlineOrderFlowInfo{
			Tracking:     order.Tracking,
			OrderGineeID: order.OrderGineeID,
			Complained:   order.Complained,
			CreatedAt:    order.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return response
}

// Request/Response structs - REORDERED to match flow
type OnlineFlowsListResponse struct {
	OnlineFlows []OnlineFlowResponse         `json:"online_flows"`
	Pagination  OnlineFlowPaginationResponse `json:"pagination"`
}

// FIXED: Use unique pagination response name
type OnlineFlowPaginationResponse struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

// REORDERED: mb-online -> qc-online -> pc-online -> order
type OnlineFlowResponse struct {
	Tracking string               `json:"tracking"`
	MbOnline *MbOnlineFlowInfo    `json:"mb_online,omitempty"`
	QcOnline *QcOnlineFlowInfo    `json:"qc_online,omitempty"`
	PcOnline *PcOnlineFlowInfo    `json:"pc_online,omitempty"`
	Order    *OnlineOrderFlowInfo `json:"order,omitempty"`
}

type MbOnlineFlowInfo struct {
	User      *OnlineUserFlowInfo `json:"user,omitempty"`
	CreatedAt string              `json:"created_at"`
}

type QcOnlineFlowInfo struct {
	User      *OnlineUserFlowInfo `json:"user,omitempty"`
	CreatedAt string              `json:"created_at"`
}

type PcOnlineFlowInfo struct {
	User      *OnlineUserFlowInfo `json:"user,omitempty"`
	CreatedAt string              `json:"created_at"`
}

type OnlineOrderFlowInfo struct {
	Tracking     string `json:"tracking"`
	OrderGineeID string `json:"order_ginee_id"`
	Complained   bool   `json:"complained"`
	CreatedAt    string `json:"created_at"`
}

type OnlineUserFlowInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
}
