package controllers

import (
	"fmt"
	"livo-gin-backend/models"
	"livo-gin-backend/utils"
	"net/http"
	"strconv"
	"strings"
	"time"

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
// @Param start_date query string false "Start date (YYYY-MM-DD or YYYY-M-D format)"
// @Param end_date query string false "End date (YYYY-MM-DD or YYYY-M-D format)"
// @Param search query string false "Search by tracking number"
// @Success 200 {object} utils.Response{data=RibbonFlowsListResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/ribbon-flow [get]
func (rfc *RibbonFlowController) GetRibbonFlows(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// ADDED: Parse date range parameters
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Parse search parameter (optional)
	search := c.Query("search")

	var trackingNumbers []string
	var total int64

	// Get tracking numbers primarily from mb_ribbons
	query := rfc.DB.Model(&models.MbRibbon{}).Select("DISTINCT tracking").Where("tracking IS NOT NULL AND tracking != ''")

	// Apply date range filters if provided
	if startDate != "" {
		// Parse start date and set time to beginning of day
		if parsedStartDate, err := time.Parse("2006-01-02", startDate); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid start_date format", "start_date must be in YYYY-MM-DD format")
			return
		} else {
			startOfDay := parsedStartDate.Format("2006-01-02 00:00:00")
			query = query.Where("created_at >= ?", startOfDay)
		}
	}

	if endDate != "" {
		// Parse end date and set time to end of day
		if parsedEndDate, err := time.Parse("2006-01-02", endDate); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid end_date format", "end_date must be in YYYY-MM-DD format")
			return
		} else {
			// Add 24 hours to get the start of next day, then use < instead of <=
			nextDay := parsedEndDate.AddDate(0, 0, 1).Format("2006-01-02 00:00:00")
			query = query.Where("created_at < ?", nextDay)
		}
	}

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
		Pagination: PaginationResponse{
			Page:  page,
			Limit: limit,
			Total: int(total),
		},
	}

	// Build success message with date filters
	message := "Ribbon flows retrieved successfully"
	var filters []string

	if startDate != "" || endDate != "" {
		var dateRange []string
		if startDate != "" {
			dateRange = append(dateRange, "from: "+startDate)
		}
		if endDate != "" {
			dateRange = append(dateRange, "to: "+endDate)
		}
		filters = append(filters, "date: "+strings.Join(dateRange, ", "))
	}

	if search != "" {
		filters = append(filters, "search: "+search)
	}

	if len(filters) > 0 {
		message += fmt.Sprintf(" (filtered by %s)", strings.Join(filters, " | "))
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
		var user *RibbonUserFlowInfo
		if mbRibbon.User != nil {
			user = &RibbonUserFlowInfo{
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
		var user *RibbonUserFlowInfo
		if qcRibbon.User != nil {
			user = &RibbonUserFlowInfo{
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
		var user *RibbonUserFlowInfo
		if outbound.User != nil {
			user = &RibbonUserFlowInfo{
				ID:       outbound.User.ID,
				Username: outbound.User.Username,
				FullName: outbound.User.FullName,
			}
		}

		response.Outbound = &RibbonOutboundFlowInfo{
			User:            user,
			Expedition:      outbound.Expedition,
			ExpeditionColor: outbound.ExpeditionColor, // ADDED
			CreatedAt:       outbound.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	// 4. Query Order (LAST)
	var order models.Order
	if err := rfc.DB.Where("tracking = ?", tracking).First(&order).Error; err == nil {
		response.Order = &RibbonOrderFlowInfo{
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
	RibbonFlows []RibbonFlowResponse `json:"ribbon_flows"`
	Pagination  PaginationResponse   `json:"pagination"`
}

// FIXED: Use unique pagination response to avoid conflicts
type RibbonFlowPaginationResponse struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

// REORDERED: mb-ribbon -> qc-ribbon -> outbound -> order
type RibbonFlowResponse struct {
	Tracking string                  `json:"tracking"`
	MbRibbon *MbRibbonFlowInfo       `json:"mb_ribbon,omitempty"`
	QcRibbon *QcRibbonFlowInfo       `json:"qc_ribbon,omitempty"`
	Outbound *RibbonOutboundFlowInfo `json:"outbound,omitempty"`
	Order    *RibbonOrderFlowInfo    `json:"order,omitempty"`
}

type MbRibbonFlowInfo struct {
	User      *RibbonUserFlowInfo `json:"user,omitempty"`
	CreatedAt string              `json:"created_at"`
}

type QcRibbonFlowInfo struct {
	User      *RibbonUserFlowInfo `json:"user,omitempty"`
	CreatedAt string              `json:"created_at"`
}

type RibbonOutboundFlowInfo struct {
	User            *RibbonUserFlowInfo `json:"user,omitempty"`
	Expedition      string              `json:"expedition"`
	ExpeditionColor string              `json:"expedition_color"`
	CreatedAt       string              `json:"created_at"`
}

type RibbonOrderFlowInfo struct {
	Tracking     string `json:"tracking"`
	OrderGineeID string `json:"order_ginee_id"`
	Complained   bool   `json:"complained"`
	CreatedAt    string `json:"created_at"`
}

type RibbonUserFlowInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
}
