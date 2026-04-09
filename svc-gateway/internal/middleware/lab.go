package middleware

import (
	"fmt"
	"math"
	"net/http"
	"strconv"

	"gateway/pkg/utils"
	"github.com/gin-gonic/gin"
)

// ExtractLabID extracts the lab ID from the URL path parameter and stores it in the context.
// Does NOT query the database.
func ExtractLabID() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")

		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id == 0 || id > math.MaxUint {
			c.AbortWithStatusJSON(http.StatusBadRequest, utils.ErrorResponse(fmt.Errorf("invalid_lab_id")))
			return
		}

		c.Set("lab_id", uint(id))
		c.Next()
	}
}
