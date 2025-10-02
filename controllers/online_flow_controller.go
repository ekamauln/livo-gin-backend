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
// @Description Get all online flows with pagination and search (logged-in users only)
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
// @Router /api/online-flows [get]
func (ofc *OnlineFlowController) GetOnlineFlows(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Parse search parameter (optional)
	search := c.Query("search")

	// Get all unique tracking numbers from online flow tables
	var trackingNumbers []string
	var total int64

	// Build query to get all tracking numbers
	query := ofc.DB.Raw(`
        SELECT DISTINCT tracking FROM (
            SELECT tracking FROM orders WHERE tracking IS NOT NULL AND tracking != ''
            UNION
            SELECT tracking FROM mb_onlines WHERE tracking IS NOT NULL AND tracking != ''
            UNION
            SELECT tracking FROM qc_onlines WHERE tracking IS NOT NULL AND tracking != ''
            UNION
            SELECT tracking FROM pc_onlines WHERE tracking IS NOT NULL AND tracking != ''
            UNION
            SELECT tracking FROM outbounds WHERE tracking IS NOT NULL AND tracking != ''
        ) AS all_trackings
    `)

	// Add search filter if provided
	if search != "" {
		query = ofc.DB.Raw(`
            SELECT DISTINCT tracking FROM (
                SELECT tracking FROM orders WHERE tracking IS NOT NULL AND tracking != '' AND tracking ILIKE ?
                UNION
                SELECT tracking FROM mb_onlines WHERE tracking IS NOT NULL AND tracking != '' AND tracking ILIKE ?
                UNION
                SELECT tracking FROM qc_onlines WHERE tracking IS NOT NULL AND tracking != '' AND tracking ILIKE ?
                UNION
                SELECT tracking FROM pc_onlines WHERE tracking IS NOT NULL AND tracking != '' AND tracking ILIKE ?
                UNION
                SELECT tracking FROM outbounds WHERE tracking IS NOT NULL AND tracking != '' AND tracking ILIKE ?
            ) AS all_trackings
            ORDER BY tracking
        `, "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	} else {
		query = query.Order("tracking")
	}

	// Get total count
	var allTrackings []string
	if err := query.Scan(&allTrackings).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve tracking numbers", err.Error())
		return
	}
	total = int64(len(allTrackings))

	// Apply pagination
	if offset < len(allTrackings) {
		end := offset + limit
		if end > len(allTrackings) {
			end = len(allTrackings)
		}
		trackingNumbers = allTrackings[offset:end]
	}

	// Build online flows for each tracking
	var onlineFlows []OnlineFlowResponse
	for _, tracking := range trackingNumbers {
		flow := ofc.buildOnlineFlow(tracking)
		onlineFlows = append(onlineFlows, flow)
	}

	response := OnlineFlowsListResponse{
		OnlineFlows: onlineFlows,
		Pagination: PaginationResponse{
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
// @Description Get the complete flow tracking of an order through online process (order -> mb-online -> qc-online -> pc-online -> outbound) (logged-in users only)
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

	// Check if any data was found
	if flow.Order == nil && flow.MbOnline == nil && flow.QcOnline == nil && flow.PcOnline == nil && flow.Outbound == nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Tracking not found", "No records found for the specified tracking number")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Online flow retrieved successfully", flow)
}

// Helper function to build online flow for a tracking number
func (ofc *OnlineFlowController) buildOnlineFlow(tracking string) OnlineFlowResponse {
	var response OnlineFlowResponse
	response.Tracking = tracking

	// 1. Query Order
	var order models.Order
	if err := ofc.DB.Where("tracking = ?", tracking).First(&order).Error; err == nil {
		response.Order = &OnlineOrderFlowInfo{
			Tracking:     order.Tracking,
			OrderGineeID: order.OrderGineeID,
			Complained:   order.Complained,
			CreatedAt:    order.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	// 2. Query MB Online
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

	// 3. Query QC Online
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

	// 4. Query PC Online
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

	// 5. Query Outbound
	var outbound models.Outbound
	if err := ofc.DB.Preload("User").Where("tracking = ?", tracking).First(&outbound).Error; err == nil {
		var user *OnlineUserFlowInfo
		if outbound.User != nil {
			user = &OnlineUserFlowInfo{
				ID:       outbound.User.ID,
				Username: outbound.User.Username,
				FullName: outbound.User.FullName,
			}
		}

		response.Outbound = &OnlineOutboundFlowInfo{
			User:       user,
			Expedition: outbound.Expedition,
			CreatedAt:  outbound.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return response
}

// Request/Response structs
type OnlineFlowsListResponse struct {
	OnlineFlows []OnlineFlowResponse `json:"online_flows"`
	Pagination  PaginationResponse   `json:"pagination"`
}

type OnlineFlowResponse struct {
	Tracking string                  `json:"tracking"`
	Order    *OnlineOrderFlowInfo    `json:"order,omitempty"`
	MbOnline *MbOnlineFlowInfo       `json:"mb_online,omitempty"`
	QcOnline *QcOnlineFlowInfo       `json:"qc_online,omitempty"`
	PcOnline *PcOnlineFlowInfo       `json:"pc_online,omitempty"`
	Outbound *OnlineOutboundFlowInfo `json:"outbound,omitempty"`
}

type OnlineOrderFlowInfo struct {
	Tracking     string `json:"tracking"`
	OrderGineeID string `json:"order_ginee_id"`
	Complained   bool   `json:"complained"`
	CreatedAt    string `json:"created_at"`
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

type OnlineOutboundFlowInfo struct {
	User       *OnlineUserFlowInfo `json:"user,omitempty"`
	Expedition string              `json:"expedition"`
	CreatedAt  string              `json:"created_at"`
}

type OnlineUserFlowInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
}
