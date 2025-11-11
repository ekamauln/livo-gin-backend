package middleware

import (
	"livo-gin-backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireRoles middleware checks if user has any of the required roles
func RequireRoles(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := c.Get("roles")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "No roles found in token", "missing roles")
			c.Abort()
			return
		}

		userRoles, ok := roles.([]string)
		if !ok {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid roles format", "roles format error")
			c.Abort()
			return
		}

		// Check if user has any of the required roles
		hasRole := false
		for _, requiredRole := range requiredRoles {
			for _, userRole := range userRoles {
				if userRole == requiredRole {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}

		if !hasRole {
			utils.ErrorResponse(c, http.StatusForbidden, "Insufficient permissions", "access denied")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireManagementRoles middleware for user management endpoints
func RequireManagementRoles() gin.HandlerFunc {
	return RequireRoles("superadmin", "coordinator")
}

// RequireRoleAssignmentRoles middleware for role assignment endpoints
func RequireRoleAssignmentRoles() gin.HandlerFunc {
	return RequireRoles("superadmin", "coordinator")
}

// RequireProductManagementRoles middleware for product management endpoints
func RequireProductManagementRoles() gin.HandlerFunc {
	return RequireRoles("superadmin", "admin", "coordinator", "finance")
}

// RequireOrderManagementRoles middleware for order management endpoints
func RequireOrderManagementRoles() gin.HandlerFunc {
	return RequireRoles("superadmin", "admin", "coordinator", "picker")
}

// RequiredSuperadminRole middleware for superadmin-only endpoints
func RequiredSuperadminRole() gin.HandlerFunc {
	return RequireRoles("superadmin")
}

// RequireCoordinatorRole middleware for coordinator-only endpoints
func RequireCoordinatorRole() gin.HandlerFunc {
	return RequireRoles("coordinator")
}

// RequireAdminRole middleware for admin-only endpoints
func RequireAdminRole() gin.HandlerFunc {
	return RequireRoles("superadmin", "coordinator", "admin")
}

// RequireFinanceRole middleware for finance-only endpoints
func RequireFinanceRole() gin.HandlerFunc {
	return RequireRoles("superadmin", "coordinator", "finance")
}

// RequirePickerRole middleware for picker-only endpoints
func RequirePickerRole() gin.HandlerFunc {
	return RequireRoles("superadmin", "coordinator", "picker")
}

// RequireOutboundRole middleware for outbound-only endpoints
func RequireOutboundRole() gin.HandlerFunc {
	return RequireRoles("superadmin", "coordinator", "outbound")
}

// RequireQCRibbonRole middleware for quality control for ribbon-only endpoints
func RequireQCRibbonRole() gin.HandlerFunc {
	return RequireRoles("superadmin", "coordinator", "qc-ribbon")
}

// RequireQCOnlineRole middleware for quality control for online-only endpoints
func RequireQCOnlineRole() gin.HandlerFunc {
	return RequireRoles("superadmin", "coordinator", "qc-online")
}

// RequireMBRibbonRole middleware for product checker for ribbon-only endpoints
func RequireMBRibbonRole() gin.HandlerFunc {
	return RequireRoles("superadmin", "coordinator", "mb-ribbon")
}

// RequireMBOnlineRole middleware for product checker for online-only endpoints
func RequireMBOnlineRole() gin.HandlerFunc {
	return RequireRoles("superadmin", "coordinator", "mb-online")
}

// RequirePackingRole middleware for packing-only endpoints
func RequirePackingRole() gin.HandlerFunc {
	return RequireRoles("superadmin", "coordinator", "packing")
}

// RequireGuestRole middleware for guest-only endpoints
func RequireGuestRole() gin.HandlerFunc {
	return RequireRoles("guest")
}
