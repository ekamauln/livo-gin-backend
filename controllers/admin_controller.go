package controllers

import (
	"livo-gin-backend/models"
	"livo-gin-backend/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdminController struct {
	DB *gorm.DB
}

// NewAdminController creates a new admin controller
func NewAdminController(db *gorm.DB) *AdminController {
	return &AdminController{DB: db}
}

// GetUsers godoc
// @Summary Get all users
// @Description Get list of all users (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} utils.Response{data=UsersListResponse}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/admin/users [get]
func (ac *AdminController) GetUsers(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var users []models.User
	var total int64

	// Get total count
	ac.DB.Model(&models.User{}).Count(&total)

	// Get users with pagination
	if err := ac.DB.Preload("UserRoles.Role").Preload("UserRoles.Assigner").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve users", err.Error())
		return
	}

	// Convert to response format
	userResponses := make([]models.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToUserResponse()
	}

	response := UsersListResponse{
		Users: userResponses,
		Pagination: PaginationResponse{
			Page:  page,
			Limit: limit,
			Total: int(total),
		},
	}

	utils.SuccessResponse(c, http.StatusOK, "Users retrieved successfully", response)
}

// GetUser godoc
// @Summary Get user by ID
// @Description Get specific user information (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} utils.Response{data=models.UserResponse}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/admin/users/{id} [get]
func (ac *AdminController) GetUser(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	if err := ac.DB.Preload("UserRoles.Role").First(&user, userID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "User not found", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User retrieved successfully", user.ToUserResponse())
}

// UpdateUserStatus godoc
// @Summary Update user status
// @Description Activate or deactivate a user (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param request body UpdateUserStatusRequest true "Update status request"
// @Success 200 {object} utils.Response{data=models.UserResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/admin/users/{id}/status [put]
func (ac *AdminController) UpdateUserStatus(c *gin.Context) {
	userID := c.Param("id")

	var req UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	var user models.User
	if err := ac.DB.First(&user, userID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "User not found", err.Error())
		return
	}

	user.IsActive = req.IsActive
	if err := ac.DB.Save(&user).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update user status", err.Error())
		return
	}

	// Load user with roles
	ac.DB.Preload("UserRoles.Role").First(&user, user.ID)

	utils.SuccessResponse(c, http.StatusOK, "User status updated successfully", user.ToUserResponse())
}

// AssignRole godoc
// @Summary Assign role to user
// @Description Assign a role to a user (requires appropriate permissions)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param request body AssignRoleRequest true "Assign role request"
// @Success 200 {object} utils.Response{data=models.UserResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/admin/users/{id}/roles [post]
func (ac *AdminController) AssignRole(c *gin.Context) {
	userID := c.Param("id")

	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Find target user
	var user models.User
	if err := ac.DB.Preload("UserRoles.Role").First(&user, userID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "User not found", err.Error())
		return
	}

	// Find role
	var role models.Role
	if err := ac.DB.Where("name = ?", req.RoleName).First(&role).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Role not found", err.Error())
		return
	}

	// Check if user already has this role
	for _, userRole := range user.UserRoles {
		if userRole.RoleID == role.ID {
			utils.ErrorResponse(c, http.StatusConflict, "User already has this role", "role already assigned")
			return
		}
	}

	// Check permission hierarchy (get current user's roles)
	currentUserRoles, _ := c.Get("roles")
	currentRoles := currentUserRoles.([]string)

	// Get current user's highest role level
	hierarchy := models.GetRoleHierarchy()
	currentMaxLevel := 0
	for _, roleName := range currentRoles {
		if level, exists := hierarchy[roleName]; exists && level > currentMaxLevel {
			currentMaxLevel = level
		}
	}

	// Check if current user can assign this role
	targetRoleLevel, exists := hierarchy[req.RoleName]
	if !exists || currentMaxLevel <= targetRoleLevel {
		utils.ErrorResponse(c, http.StatusForbidden, "Insufficient permissions to assign this role", "permission denied")
		return
	}

	// Assign role
	userRole := models.UserRole{
		UserID: user.ID,
		RoleID: role.ID,
	}

	if err := ac.DB.Create(&userRole).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to assign role", err.Error())
		return
	}

	// Reload user with updated roles
	ac.DB.Preload("UserRoles.Role").First(&user, user.ID)

	utils.SuccessResponse(c, http.StatusOK, "Role assigned successfully", user.ToUserResponse())
}

// RemoveRole godoc
// @Summary Remove role from user
// @Description Remove a role from a user (requires appropriate permissions)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param request body RemoveRoleRequest true "Remove role request"
// @Success 200 {object} utils.Response{data=models.UserResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/admin/users/{id}/roles [delete]
func (ac *AdminController) RemoveRole(c *gin.Context) {
	userID := c.Param("id")

	var req RemoveRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Find role
	var role models.Role
	if err := ac.DB.Where("name = ?", req.RoleName).First(&role).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Role not found", err.Error())
		return
	}

	// Check permission hierarchy
	currentUserRoles, _ := c.Get("roles")
	currentRoles := currentUserRoles.([]string)

	hierarchy := models.GetRoleHierarchy()
	currentMaxLevel := 0
	for _, roleName := range currentRoles {
		if level, exists := hierarchy[roleName]; exists && level > currentMaxLevel {
			currentMaxLevel = level
		}
	}

	targetRoleLevel, exists := hierarchy[req.RoleName]
	if !exists || currentMaxLevel <= targetRoleLevel {
		utils.ErrorResponse(c, http.StatusForbidden, "Insufficient permissions to remove this role", "permission denied")
		return
	}

	// Remove role
	if err := ac.DB.Where("user_id = ? AND role_id = ?", userID, role.ID).Delete(&models.UserRole{}).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove role", err.Error())
		return
	}

	// Reload user with updated roles
	var user models.User
	ac.DB.Preload("UserRoles.Role").Preload("UserRoles.Assigner").First(&user, userID)

	utils.SuccessResponse(c, http.StatusOK, "Role removed successfully", user.ToUserResponse())
}

// GetRoles godoc
// @Summary Get all roles
// @Description Get list of all available roles
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=[]models.Role}
// @Failure 401 {object} utils.Response
// @Router /api/admin/roles [get]
func (ac *AdminController) GetRoles(c *gin.Context) {
	var roles []models.Role
	if err := ac.DB.Find(&roles).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve roles", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Roles retrieved successfully", roles)
}

// Request/Response structs
type UsersListResponse struct {
	Users      []models.UserResponse `json:"users"`
	Pagination PaginationResponse    `json:"pagination"`
}

type PaginationResponse struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

type UpdateUserStatusRequest struct {
	IsActive bool `json:"is_active" binding:"required" example:"true"`
}

type AssignRoleRequest struct {
	RoleName string `json:"role_name" binding:"required" example:"manager"`
}

type RemoveRoleRequest struct {
	RoleName string `json:"role_name" binding:"required" example:"manager"`
}
