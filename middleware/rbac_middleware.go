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
	return RequireRoles("superadmin", "admin", "manager")
}

// RequireRoleAssignmentRoles middleware for role assignment endpoints
func RequireRoleAssignmentRoles() gin.HandlerFunc {
	return RequireRoles("superadmin", "admin", "manager", "supervisor")
}

// RequireProductManagementRoles middleware for product management endpoints
func RequireProductManagementRoles() gin.HandlerFunc {
	return RequireRoles("admin", "manager")
}
