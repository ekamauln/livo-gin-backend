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
		Preload("User.UserRoles.Assigner").
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

// GetPcOnline godoc
// @Summary Get a specific pc-online by ID
// @Description Get a specific pc-online by ID (logged-in users only)
// @Tags pc-onlines
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Pc-online ID"
// @Success 200 {object} utils.Response{data=models.PcOnlineResponse}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/pc-onlines/{id} [get]
func (pc *PcOnlineController) GetPcOnline(c *gin.Context) {
	pcOnlineID := c.Param("id")

	var pcOnline models.PcOnline

	if err := pc.DB.Preload("Details.Box").
		Preload("User.UserRoles.Role").
		Preload("User.UserRoles.Assigner").
		First(&pcOnline, pcOnlineID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Pc-online not found", err.Error())
		return
	}

	// Load order call from the model's LoadOrder method
	if pcOnline.Tracking != "" {
		pcOnline.LoadOrder(pc.DB)
	}

	utils.SuccessResponse(c, http.StatusOK, "Pc-online retrieved successfully", pcOnline.ToPcOnlineResponse())
}

// CreatePcOnline godoc
// @Summary Create a new pc-online
// @Description Create a new pc-online entry (logged-in users only)
// @Tags pc-onlines
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreatePcOnlineRequest true "Create pc-online request"
// @Success 201 {object} utils.Response{data=models.PcOnlineResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/pc-onlines [post]
func (pc *PcOnlineController) CreatePcOnline(c *gin.Context) {
	// Get user ID from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "User not authenticated")
		return
	}

	var req CreatePcOnlineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Convert userID to uint
	userIDUint, ok := userID.(uint)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user ID", "User ID is not a valid uint")
		return
	}

	// Check if tracking exists in qc-online table with boxID = 1
	var qcOnline models.QcOnline
	if err := pc.DB.Preload("Details").Where("tracking = ?", req.Tracking).First(&qcOnline).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "QC Online not found", "No QC online found with the specified tracking number")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to validate tracking", err.Error())
		return
	}

	// Check if QC Online has detail with boxID = 1
	hasBoxID1 := false
	for _, detail := range qcOnline.Details {
		if detail.BoxID == 1 {
			hasBoxID1 = true
			break
		}
	}

	if !hasBoxID1 {
		utils.ErrorResponse(c, http.StatusBadRequest, "QC Online validation failed", "QC online must have a detail with box ID = 1")
		return
	}

	// Validate all boxes exist and no duplicates
	boxIDs := make(map[uint]bool)
	for _, detail := range req.Details {
		// Check for duplicate box IDs in request
		if boxIDs[detail.BoxID] {
			utils.ErrorResponse(c, http.StatusBadRequest, "Duplicate box ID", "Each box can only be added once per PC online")
			return
		}
		boxIDs[detail.BoxID] = true

		// Check if box exists
		var box models.Box
		if err := pc.DB.First(&box, detail.BoxID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Box not found", "No box found with ID "+strconv.Itoa(int(detail.BoxID)))
			return
		}

		// Validate quantity
		if detail.Quantity <= 0 {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid quantity", "Quantity must be greater than 0")
			return
		}
	}

	// Check for duplicate tracking in pc-online
	var existingPcOnline models.PcOnline
	if err := pc.DB.Where("tracking = ?", req.Tracking).First(&existingPcOnline).Error; err == nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Pc-online with this tracking already exists", "Duplicate tracking")
		return
	}

	// Start database transaction
	tx := pc.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create PC Online first
	pcOnline := models.PcOnline{
		Tracking: req.Tracking,
		UserID:   &userIDUint,
	}

	if err := tx.Create(&pcOnline).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create pc-online", err.Error())
		return
	}

	// Create PC Online Details from request
	for _, detail := range req.Details {
		pcOnlineDetail := models.PcOnlineDetail{
			PcOnlineID: pcOnline.ID,
			BoxID:      detail.BoxID,
			Quantity:   detail.Quantity,
		}

		if err := tx.Create(&pcOnlineDetail).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create pc-online detail", err.Error())
			return
		}
	}

	// FIXED: Delete QC Online details for this tracking
	if err := tx.Where("qc_online_id = ?", qcOnline.ID).Delete(&models.QcOnlineDetail{}).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to clear QC online details", err.Error())
		return
	}

	// ADDED: Replace QC Online details with PC Online details data
	for _, detail := range req.Details {
		newQcOnlineDetail := models.QcOnlineDetail{
			QcOnlineID: qcOnline.ID,
			BoxID:      detail.BoxID,
			Quantity:   detail.Quantity,
		}

		if err := tx.Create(&newQcOnlineDetail).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create new QC online detail", err.Error())
			return
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to commit transaction", err.Error())
		return
	}

	// Load the created pc-online with relationships
	pc.DB.Preload("Details.Box").
		Preload("User.UserRoles.Role").
		Preload("User.UserRoles.Assigner").
		First(&pcOnline, pcOnline.ID)

	// Load order call from the model's LoadOrder method
	if pcOnline.Tracking != "" {
		pcOnline.LoadOrder(pc.DB)
	}

	utils.SuccessResponse(c, http.StatusCreated, "Pc-online created successfully", pcOnline.ToPcOnlineResponse())
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
