package controllers

import (
	"livo-gin-backend/models"
	"livo-gin-backend/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type QcRibbonController struct {
	DB *gorm.DB
}

// NewQcRibbonController creates a new qc-ribbon controller
func NewQcRibbonController(db *gorm.DB) *QcRibbonController {
	return &QcRibbonController{DB: db}
}

// GetQcRibbons godoc
// @Summary Get all qc-ribbons for logged-in user
// @Description Get list of qc-ribbons for current user filtered by current date (logged-in users only)
// @Tags qc-ribbons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search by tracking number"
// @Success 200 {object} utils.Response{data=QcRibbonsListResponse}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/qc-ribbons [get]
func (qrc *QcRibbonController) GetQcRibbons(c *gin.Context) {
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

	var qcRibbons []models.QcRibbon
	var total int64

	// Build query with filters
	query := qrc.DB.Model(&models.QcRibbon{}).Where("user_id = ?", userID).Where("DATE(created_at) = CURRENT_DATE")

	if search != "" {
		// Search by tracking with partial match
		query = query.Where("tracking ILIKE ?", "%"+search+"%")
	}

	// Get total count with filters
	if err := query.Count(&total).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to count qc-ribbons", err.Error())
		return
	}

	// Get qc-ribbons with pagination, filters, and preload relationships
	if err := query.Order("id DESC").
		Preload("Details.Box").
		Preload("User.UserRoles.Role").
		Preload("User.UserRoles.Assigner").
		Limit(limit).Offset(offset).
		Find(&qcRibbons).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve qc-ribbons", err.Error())
		return
	}

	// Manually Load all orders at once using tracking field
	if len(qcRibbons) > 0 {
		trackingNumbers := make([]string, 0, len(qcRibbons))
		for _, qcRibbon := range qcRibbons {
			if qcRibbon.Tracking != "" {
				trackingNumbers = append(trackingNumbers, qcRibbon.Tracking)
			}
		}

		var orders []models.Order
		orderMap := make(map[string]*models.Order)

		if len(trackingNumbers) > 0 {
			if err := qrc.DB.Where("tracking IN ?", trackingNumbers).
				Preload("OrderDetails").
				Preload("Picker.UserRoles.Role").
				Preload("Picker.UserRoles.Assigner").
				Find(&orders).Error; err == nil {

				for i := range orders {
					orderMap[orders[i].Tracking] = &orders[i]
				}
			}
		}

		for i := range qcRibbons {
			if order, exists := orderMap[qcRibbons[i].Tracking]; exists {
				qcRibbons[i].Order = order
			}
		}
	}

	// Convert to response format
	qcRibbonResponses := make([]models.QcRibbonResponse, len(qcRibbons))
	for i, qcRibbon := range qcRibbons {
		qcRibbonResponses[i] = qcRibbon.ToQcRibbonResponse()
	}

	response := QcRibbonsListResponse{
		QcRibbons: qcRibbonResponses,
		Pagination: PaginationResponse{
			Page:  page,
			Limit: limit,
			Total: int(total),
		},
	}

	// Build success message
	message := "Qc-ribbons retrieved successfully"
	if search != "" {
		message += " (filtered by tracking: " + search + ")"
	}

	utils.SuccessResponse(c, http.StatusOK, message, response)
}

// GetQcRibbon godoc
// @Summary Get a specific qc-ribbon by ID
// @Description Get a specific qc-ribbon by ID (logged-in users only)
// @Tags qc-ribbons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Qc-ribbon ID"
// @Success 200 {object} utils.Response{data=models.QcRibbonResponse}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/qc-ribbons/{id} [get]
func (qrc *QcRibbonController) GetQcRibbon(c *gin.Context) {
	qcRibbonID := c.Param("id")

	var qcRibbon models.QcRibbon

	if err := qrc.DB.Preload("Details.Box").
		Preload("User.UserRoles.Role").
		Preload("User.UserRoles.Assigner").
		First(&qcRibbon, qcRibbonID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Qc-ribbon not found", err.Error())
		return
	}

	// Load order call from the model's LoadOrder method
	if qcRibbon.Tracking != "" {
		qcRibbon.LoadOrder(qrc.DB)
	}

	utils.SuccessResponse(c, http.StatusOK, "Qc-ribbon retrieved successfully", qcRibbon.ToQcRibbonResponse())
}

// CreateQcRibbon godoc
// @Summary Create new qc-ribbon
// @Description Create a new qc-ribbon entry with multiple box details (logged-in users only)
// @Tags qc-ribbons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param qc_ribbon body CreateQcRibbonRequest true "Qc-ribbon data"
// @Success 201 {object} utils.Response{data=models.QcRibbonResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/qc-ribbons [post]
func (qrc *QcRibbonController) CreateQcRibbon(c *gin.Context) {
	// Get user ID from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "User not authenticated")
		return
	}

	var req CreateQcRibbonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Convert tracking to uppercase
	req.Tracking = strings.ToUpper(strings.TrimSpace(req.Tracking))

	// Convert userID to uint
	userIDUint, ok := userID.(uint)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user ID", "Failed to convert user ID")
		return
	}

	// Check if tracking exists in mb-ribbons table first
	var mbRibbon models.MbRibbon
	if err := qrc.DB.Where("tracking = ?", req.Tracking).First(&mbRibbon).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "MB Ribbon not found", "No MB ribbon found with the specified tracking number")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to validate tracking", err.Error())
		return
	}

	// Validate all boxes exist and no duplicates
	boxIDs := make(map[uint]bool)
	for _, detail := range req.Details {
		// Check for duplicate box IDs in the request
		if boxIDs[detail.BoxID] {
			utils.ErrorResponse(c, http.StatusBadRequest, "Duplicate box ID", "Each box can only be added once per QC ribbon")
			return
		}
		boxIDs[detail.BoxID] = true

		// Check if box exists
		var box models.Box
		if err := qrc.DB.First(&box, detail.BoxID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Box not found", "Invalid box ID: "+strconv.Itoa(int(detail.BoxID)))
			return
		}

		// Validate quantity
		if detail.Quantity <= 0 {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid quantity", "Quantity must be greater than 0")
			return
		}
	}

	// Check for duplicate tracking
	var existingQcRibbon models.QcRibbon
	if err := qrc.DB.Where("tracking = ?", req.Tracking).First(&existingQcRibbon).Error; err == nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Qc-ribbon with this tracking already exists", "Duplicate tracking")
		return
	}

	// Start database transaction
	tx := qrc.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create QC Ribbon
	qcRibbon := models.QcRibbon{
		Tracking: req.Tracking,
		UserID:   &userIDUint,
	}

	if err := tx.Create(&qcRibbon).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create qc-ribbon", err.Error())
		return
	}

	// Create QC Ribbon Detail for each box
	for _, detail := range req.Details {
		qcRibbonDetail := models.QcRibbonDetail{
			QcRibbonID: qcRibbon.ID,
			BoxID:      detail.BoxID,
			Quantity:   detail.Quantity,
		}

		if err := tx.Create(&qcRibbonDetail).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create qc-ribbon detail", err.Error())
			return
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to commit transaction", err.Error())
		return
	}

	// Load the created qc-ribbon with all relationships
	qrc.DB.Preload("Details.Box").
		Preload("User.UserRoles.Role").
		Preload("User.UserRoles.Assigner").
		First(&qcRibbon, qcRibbon.ID)

	// Load order call from the model's LoadOrder method
	if qcRibbon.Tracking != "" {
		qcRibbon.LoadOrder(qrc.DB)
	}

	utils.SuccessResponse(c, http.StatusCreated, "Qc-ribbon created successfully", qcRibbon.ToQcRibbonResponse())
}

// Request/Response structs
type QcRibbonsListResponse struct {
	QcRibbons  []models.QcRibbonResponse `json:"qc_ribbons"`
	Pagination PaginationResponse        `json:"pagination"`
}

type QcRibbonDetailRequest struct {
	BoxID    uint `json:"box_id" binding:"required" example:"1"`
	Quantity int  `json:"quantity" binding:"required,min=1" example:"5"`
}

type CreateQcRibbonRequest struct {
	Tracking string                  `json:"tracking" binding:"required" example:"250925AASB6BSDJUI3C"`
	Details  []QcRibbonDetailRequest `json:"details" binding:"required,dive,required"`
}
