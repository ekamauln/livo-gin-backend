package controllers

import (
	"livo-gin-backend/models"
	"livo-gin-backend/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ComplainController struct {
	DB *gorm.DB
}

// NewComplainController creates a new complain controller
func NewComplainController(db *gorm.DB) *ComplainController {
	return &ComplainController{DB: db}
}

// GetComplains godoc
// @Summary Get all complains
// @Description Get list of all complains (logged-in users only)
// @Tags complains
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search by complain code, tracking (partial match)"
// @Success 200 {object} utils.Response{data=ComplainsListResponse}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/complains [get]
func (cc *ComplainController) GetComplains(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Parse search parameter
	search := c.Query("search")

	var complains []models.Complain
	var total int64

	// Build query with optional search
	query := cc.DB.Model(&models.Complain{})

	if search != "" {
		// Search by complain code with partial match
		query = query.Where("code ILIKE ? OR tracking ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Get total count with search filter
	if err := query.Count(&total).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to count complains", err.Error())
		return
	}

	// ADDED: Preload relationships for complete data
	if err := query.
		Preload("ComplainProductDetails.Product").
		Preload("ComplainUserDetails.User").
		Preload("Channel").
		Preload("Store").
		Preload("User").
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&complains).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch complains", err.Error())
		return
	}

	// ADDED: Load order data for each complain
	for i := range complains {
		if complains[i].Tracking != "" {
			var order models.Order
			if err := cc.DB.Preload("OrderDetails").
				Preload("Picker.UserRoles.Role").
				Preload("Picker.UserRoles.Assigner").
				Where("tracking = ?", complains[i].Tracking).First(&order).Error; err == nil {
				complains[i].Order = &order
			}
		}
	}

	// Convert to response format
	complainResponse := make([]models.ComplainResponse, len(complains))
	for i, complain := range complains {
		complainResponse[i] = complain.ToComplainResponse()
	}

	response := ComplainsListResponse{
		Complains: complainResponse,
		Pagination: PaginationResponse{
			Page:  page,
			Limit: limit,
			Total: int(total),
		},
	}

	// Build success message
	message := "Complains retrieved successfully"
	if search != "" {
		message += " (filtered by search: " + search + ")"
	}

	utils.SuccessResponse(c, http.StatusOK, message, response)
}

// GetComplain godoc
// @Summary Get complain by ID
// @Description Get complain details by ID (logged-in users only)
// @Tags complains
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Complain ID"
// @Success 200 {object} utils.Response{data=models.ComplainResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/complains/{id} [get]
func (cc *ComplainController) GetComplain(c *gin.Context) {
	complainID := c.Param("id")

	var complain models.Complain
	if err := cc.DB.Preload("ComplainProductDetails.Product").
		Preload("ComplainUserDetails.User").
		Preload("Channel").
		Preload("Store").
		Preload("User").
		First(&complain, complainID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Complain not found", err.Error())
		return
	}

	// Load order data if tracking exists
	if complain.Tracking != "" {
		var order models.Order
		if err := cc.DB.Preload("OrderDetails").
			Preload("Picker.UserRoles.Role").
			Preload("Picker.UserRoles.Assigner").
			Where("tracking = ?", complain.Tracking).First(&order).Error; err == nil {
			complain.Order = &order
		}
	}

	utils.SuccessResponse(c, http.StatusOK, "Complain retrieved successfully", complain.ToComplainResponse())
}

// CreateComplain godoc
// @Summary Create a new complain
// @Description Create a new complain with automatic product and user details population (logged-in users only)
// @Tags complains
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param complain body CreateComplainRequest true "Create complain request"
// @Success 201 {object} utils.Response{data=models.ComplainResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/complains [post]
func (cc *ComplainController) CreateComplain(c *gin.Context) {
	// Get user ID from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "User not authenticated")
		return
	}

	// Get username from JWT token
	username, exists := c.Get("username")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "Username not found in token")
		return
	}

	var req CreateComplainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Check for duplicate tracking
	var existingComplain models.Complain
	if err := cc.DB.Where("tracking = ?", req.Tracking).First(&existingComplain).Error; err == nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Complain tracking already exists", "A complain with this tracking already exists")
		return
	}

	// Start database transaction
	tx := cc.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Generate complain code with username
	complainCode := utils.GenerateComplainCode(cc.DB, username.(string))

	complain := models.Complain{
		Code:        complainCode,
		Tracking:    req.Tracking,
		ChannelID:   req.ChannelID,
		StoreID:     req.StoreID,
		Description: req.Description,
		UserID:      userID.(uint),
	}

	// Create the complain
	if err := tx.Create(&complain).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create complain", err.Error())
		return
	}

	// Populate product details from order details
	var order models.Order
	if err := tx.Preload("OrderDetails").Where("tracking = ?", req.Tracking).First(&order).Error; err == nil {
		for _, orderDetail := range order.OrderDetails {
			// Find product by SKU
			var product models.Product
			if err := tx.Where("sku = ?", orderDetail.Sku).First(&product).Error; err == nil {
				productDetail := models.ComplainProductDetail{
					ComplainID: complain.ID,
					ProductID:  product.ID,
					Quantity:   orderDetail.Quantity,
				}

				if err := tx.Create(&productDetail).Error; err != nil {
					tx.Rollback()
					utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create product detail", err.Error())
					return
				}
			}
		}
	}

	// Populate user details from workflow tables
	userIDs := make(map[uint]bool) // To avoid duplicate users

	// 1. Check MB-Ribbon
	var mbRibbon models.MbRibbon
	if err := tx.Where("tracking = ?", req.Tracking).First(&mbRibbon).Error; err == nil {
		userIDs[mbRibbon.UserID] = true
	}

	// 2. Check QC-Ribbon
	var qcRibbon models.QcRibbon
	if err := tx.Where("tracking = ?", req.Tracking).First(&qcRibbon).Error; err == nil && qcRibbon.UserID != nil {
		userIDs[*qcRibbon.UserID] = true
	}

	// 3. Check MB-Online
	var mbOnline models.MbOnline
	if err := tx.Where("tracking = ?", req.Tracking).First(&mbOnline).Error; err == nil {
		userIDs[mbOnline.UserID] = true
	}

	// 4. Check QC-Online
	var qcOnline models.QcOnline
	if err := tx.Where("tracking = ?", req.Tracking).First(&qcOnline).Error; err == nil && qcOnline.UserID != nil {
		userIDs[*qcOnline.UserID] = true
	}

	// 5. Check PC-Online
	var pcOnline models.PcOnline
	if err := tx.Where("tracking = ?", req.Tracking).First(&pcOnline).Error; err == nil && pcOnline.UserID != nil {
		userIDs[*pcOnline.UserID] = true
	}

	// 6. Check Outbound
	var outbound models.Outbound
	if err := tx.Where("tracking = ?", req.Tracking).First(&outbound).Error; err == nil {
		userIDs[outbound.UserID] = true
	}

	// Create user details for each unique user found
	for userIDValue := range userIDs {
		userDetail := models.ComplainUserDetail{
			ComplainID: complain.ID,
			UserID:     userIDValue,
			FeeCharge:  0, // Default fee, can be updated later
		}

		if err := tx.Create(&userDetail).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create user detail", err.Error())
			return
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to commit transaction", err.Error())
		return
	}

	// Load the created complain with all relationships for complete response
	cc.DB.Preload("ComplainProductDetails.Product").
		Preload("ComplainUserDetails.User").
		Preload("Channel").
		Preload("Store").
		Preload("User").
		First(&complain, complain.ID)

	// Load order data if tracking exists
	if complain.Tracking != "" {
		var order models.Order
		if err := cc.DB.Preload("OrderDetails").
			Preload("Picker.UserRoles.Role").
			Preload("Picker.UserRoles.Assigner").
			Where("tracking = ?", complain.Tracking).First(&order).Error; err == nil {
			complain.Order = &order
		}
	}

	utils.SuccessResponse(c, http.StatusCreated, "Complain created successfully", complain.ToComplainResponse())
}

// UpdateSolutionComplain godoc
// @Summary Update complain solution and user details
// @Description Update complain solution, total fee, and manage user details (logged-in users only)
// @Tags complains
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Complain ID"
// @Param request body UpdateSolutionComplainRequest true "Update Solution Complain Request"
// @Success 200 {object} utils.Response{data=models.ComplainResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/complains/{id}/solution [put]
func (cc *ComplainController) UpdateSolutionComplain(c *gin.Context) {
	complainID := c.Param("id")

	var req UpdateSolutionComplainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	var complain models.Complain
	if err := cc.DB.First(&complain, complainID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Complain not found", err.Error())
		return
	}

	// Start database transaction
	tx := cc.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update complain solution and total fee
	complain.Solution = req.Solution
	complain.TotalFee = req.TotalFee

	if err := tx.Save(&complain).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update complain", err.Error())
		return
	}

	// Handle user details updates
	if len(req.UserDetails) > 0 {
		// Clear existing user details
		if err := tx.Where("complain_id = ?", complain.ID).Delete(&models.ComplainUserDetail{}).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to clear existing user details", err.Error())
			return
		}

		// Create new user details
		for _, userDetailReq := range req.UserDetails {
			// Validate user exists
			var user models.User
			if err := tx.First(&user, userDetailReq.UserID).Error; err != nil {
				tx.Rollback()
				utils.ErrorResponse(c, http.StatusBadRequest, "User not found", "User with ID "+strconv.Itoa(int(userDetailReq.UserID))+" not found")
				return
			}

			userDetail := models.ComplainUserDetail{
				ComplainID: complain.ID,
				UserID:     userDetailReq.UserID,
				FeeCharge:  userDetailReq.FeeCharge,
			}

			if err := tx.Create(&userDetail).Error; err != nil {
				tx.Rollback()
				utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create user detail", err.Error())
				return
			}
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to commit transaction", err.Error())
		return
	}

	// Load updated complain with all relationships
	cc.DB.Preload("ComplainProductDetails.Product").
		Preload("ComplainUserDetails.User").
		Preload("Channel").
		Preload("Store").
		Preload("User").
		First(&complain, complain.ID)

	// Load order data if tracking exists
	if complain.Tracking != "" {
		var order models.Order
		if err := cc.DB.Preload("OrderDetails").
			Preload("Picker.UserRoles.Role").
			Preload("Picker.UserRoles.Assigner").
			Where("tracking = ?", complain.Tracking).First(&order).Error; err == nil {
			complain.Order = &order
		}
	}

	utils.SuccessResponse(c, http.StatusOK, "Complain solution updated successfully", complain.ToComplainResponse())
}

// UpdateCheckComplain godoc
// @Summary Update complain check status
// @Description Update complain checked status (logged-in users only)
// @Tags complains
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Complain ID"
// @Param request body UpdateCheckComplainRequest true "Update Check Complain Request"
// @Success 200 {object} utils.Response{data=models.ComplainResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/complains/{id}/check [put]
func (cc *ComplainController) UpdateCheckComplain(c *gin.Context) {
	complainID := c.Param("id")

	var req UpdateCheckComplainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	var complain models.Complain
	if err := cc.DB.First(&complain, complainID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Complain not found", err.Error())
		return
	}

	// Update complain checked status
	complain.Checked = req.Checked

	if err := cc.DB.Save(&complain).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update complain check status", err.Error())
		return
	}

	// Load updated complain with all relationships
	cc.DB.Preload("ComplainProductDetails.Product").
		Preload("ComplainUserDetails.User").
		Preload("Channel").
		Preload("Store").
		Preload("User").
		First(&complain, complain.ID)

	// Load order data if tracking exists
	if complain.Tracking != "" {
		var order models.Order
		if err := cc.DB.Preload("OrderDetails").
			Preload("Picker.UserRoles.Role").
			Preload("Picker.UserRoles.Assigner").
			Where("tracking = ?", complain.Tracking).First(&order).Error; err == nil {
			complain.Order = &order
		}
	}

	utils.SuccessResponse(c, http.StatusOK, "Complain check status updated successfully", complain.ToComplainResponse())
}

// Request/Response structs
type ComplainsListResponse struct {
	Complains  []models.ComplainResponse `json:"complains"`
	Pagination PaginationResponse        `json:"pagination"`
}

type CreateComplainRequest struct {
	Tracking    string `json:"tracking" binding:"required"`
	ChannelID   uint   `json:"channel_id" binding:"required"`
	StoreID     uint   `json:"store_id" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type UpdateSolutionComplainRequest struct {
	Solution    string                      `json:"solution" binding:"required" example:"Replacement package sent"`
	TotalFee    uint                        `json:"total_fee" binding:"required" example:"50000"`
	UserDetails []ComplainUserDetailRequest `json:"user_details" binding:"required,dive,required"`
}

type ComplainUserDetailRequest struct {
	UserID    uint `json:"user_id" binding:"required" example:"1"`
	FeeCharge uint `json:"fee_charge" binding:"required" example:"10000"`
}

type UpdateCheckComplainRequest struct {
	Checked bool `json:"checked" binding:"required"`
}
