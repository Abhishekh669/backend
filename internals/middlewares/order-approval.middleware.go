package middlewares

import (
	"context"
	"net/http"
	"time"

	"github.com/Abhishekh669/backend/internals/app"
	"github.com/Abhishekh669/backend/internals/lib"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

func CustomerMiddleware(app *app.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := lib.ExtractTokenFromCookie(c)
		if err != nil || tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "missing or invalid authorization header",
			})
			c.Abort()
			return
		}

		claims, err := lib.ValidateOrderApprovalToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "invalid or expired token",
			})
			c.Abort()
			return
		}

		validationUUID, err := uuid.FromString(claims.Id)
		if err != nil {

			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID", "success": false})
			c.Abort()
			return
		}

		table_data, err := app.OrderRepo.GetTableValidationByID(context.Background(), validationUUID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID", "success": false})
			c.Abort()
			return
		}

		if time.Since(table_data.CreatedAt) > 23*time.Hour {
			app.OrderRepo.DeleteTableApprovalByID(context.Background(), table_data.ID)
		}

	}
}
