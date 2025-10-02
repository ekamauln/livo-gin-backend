package controllers

import (
	"livo-gin-backend/models"
	"livo-gin-backend/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RibbonFlowController struct {
	DB *gorm.DB
}

// NewRibbonFlowController creates a new ribbon flow controller
func NewRibbonFlowController(db *gorm.DB) *RibbonFlowController {
	return &RibbonFlowController{DB: db}
}

// GetRibbonFlows godoc
// @Summary Get all ribbon flows
// @Description Get all ribbon flows with pagination and search, primary tracking from mb-ribbon (logged-in users only)
// @Tags ribbon-flow
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search by tracking number"
// @Success 200 {object} utils.Response{data=RibbonFlowsListResponse}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/ribbon-flow [get]
func (rfc *RibbonFlowController) GetRibbonFlows(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Parse search parameter (optional)
	search := c.Query("search")

	var trackingNumbers []string
	var total int64

	// CHANGED: Get tracking numbers primarily from mb_ribbons
	query := rfc.DB.Model(&models.MbRibbon{}).Select("DISTINCT tracking").Where("tracking IS NOT NULL AND tracking != ''")

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

	// Build ribbon flows for each tracking
	var ribbonFlows []RibbonFlowResponse
	for _, tracking := range trackingNumbers {
		flow := rfc.buildRibbonFlow(tracking)
		ribbonFlows = append(ribbonFlows, flow)
	}

	response := RibbonFlowsListResponse{
		RibbonFlows: ribbonFlows,
		Pagination: RibbonFlowPaginationResponse{
			Page:  page,
			Limit: limit,
			Total: int(total),
		},
	}

	// Build success message
	message := "Ribbon flows retrieved successfully"
	if search != "" {
		message += " (filtered by tracking: " + search + ")"
	}

	utils.SuccessResponse(c, http.StatusOK, message, response)
}

// GetRibbonFlow godoc
// @Summary Get ribbon flow tracking
// @Description Get the complete flow tracking through ribbon process (mb-ribbon -> qc-ribbon -> outbound -> order) (logged-in users only)
// @Tags ribbon-flow
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tracking path string true "Tracking number"
// @Success 200 {object} utils.Response{data=RibbonFlowResponse}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/ribbon-flow/{tracking} [get]
func (rfc *RibbonFlowController) GetRibbonFlow(c *gin.Context) {
	tracking := c.Param("tracking")

	if tracking == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid tracking", "Tracking number is required")
		return
	}

	flow := rfc.buildRibbonFlow(tracking)

	// CHANGED: Check if mb-ribbon exists (since it's the primary source)
	if flow.MbRibbon == nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Tracking not found", "No mb-ribbon record found for the specified tracking number")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Ribbon flow retrieved successfully", flow)
}

// Helper function to build ribbon flow for a tracking number
func (rfc *RibbonFlowController) buildRibbonFlow(tracking string) RibbonFlowResponse {
	var response RibbonFlowResponse
	response.Tracking = tracking

	// 1. Query MB Ribbon (PRIMARY SOURCE)
	var mbRibbon models.MbRibbon
	if err := rfc.DB.Preload("User").Where("tracking = ?", tracking).First(&mbRibbon).Error; err == nil {
		var user *UserFlowInfo
		if mbRibbon.User != nil {
			user = &UserFlowInfo{
				ID:       mbRibbon.User.ID,
				Username: mbRibbon.User.Username,
				FullName: mbRibbon.User.FullName,
			}
		}

		response.MbRibbon = &MbRibbonFlowInfo{
			User:      user,
			CreatedAt: mbRibbon.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	// 2. Query QC Ribbon
	var qcRibbon models.QcRibbon
	if err := rfc.DB.Preload("User").Where("tracking = ?", tracking).First(&qcRibbon).Error; err == nil {
		var user *UserFlowInfo
		if qcRibbon.User != nil {
			user = &UserFlowInfo{
				ID:       qcRibbon.User.ID,
				Username: qcRibbon.User.Username,
				FullName: qcRibbon.User.FullName,
			}
		}

		response.QcRibbon = &QcRibbonFlowInfo{
			User:      user,
			CreatedAt: qcRibbon.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	// 3. Query Outbound
	var outbound models.Outbound
	if err := rfc.DB.Preload("User").Where("tracking = ?", tracking).First(&outbound).Error; err == nil {
		var user *UserFlowInfo
		if outbound.User != nil {
			user = &UserFlowInfo{
				ID:       outbound.User.ID,
				Username: outbound.User.Username,
				FullName: outbound.User.FullName,
			}
		}

		response.Outbound = &OutboundFlowInfo{
			User:       user,
			Expedition: outbound.Expedition,
			CreatedAt:  outbound.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	// 4. Query Order (LAST)
	var order models.Order
	if err := rfc.DB.Where("tracking = ?", tracking).First(&order).Error; err == nil {
		response.Order = &OrderFlowInfo{
			Tracking:     order.Tracking,
			OrderGineeID: order.OrderGineeID,
			Complained:   order.Complained,
			CreatedAt:    order.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return response
}

// Request/Response structs - REORDERED to match flow
type RibbonFlowsListResponse struct {
	RibbonFlows []RibbonFlowResponse         `json:"ribbon_flows"`
	Pagination  RibbonFlowPaginationResponse `json:"pagination"`
}

type RibbonFlowPaginationResponse struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

// REORDERED: mb-ribbon -> qc-ribbon -> outbound -> order
type RibbonFlowResponse struct {
	Tracking string            `json:"tracking"`
	MbRibbon *MbRibbonFlowInfo `json:"mb_ribbon,omitempty"`
	QcRibbon *QcRibbonFlowInfo `json:"qc_ribbon,omitempty"`
	Outbound *OutboundFlowInfo `json:"outbound,omitempty"`
	Order    *OrderFlowInfo    `json:"order,omitempty"`
}

type MbRibbonFlowInfo struct {
	User      *UserFlowInfo `json:"user,omitempty"`
	CreatedAt string        `json:"created_at"`
}

type QcRibbonFlowInfo struct {
	User      *UserFlowInfo `json:"user,omitempty"`
	CreatedAt string        `json:"created_at"`
}

type OutboundFlowInfo struct {
	User       *UserFlowInfo `json:"user,omitempty"`
	Expedition string        `json:"expedition"`
	CreatedAt  string        `json:"created_at"`
}

type OrderFlowInfo struct {
	Tracking     string `json:"tracking"`
	OrderGineeID string `json:"order_ginee_id"`
	Complained   bool   `json:"complained"`
	CreatedAt    string `json:"created_at"`
}

type UserFlowInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
}
