package middleware

import (
	"fmt"
	"math"
	"net/http"
	"strconv"

	"gateway/pkg/utils"
	"github.com/gin-gonic/gin"
)

// ExtractDocID parses the :doc_id URL param and stores it in the gin context
// under key "doc_id". Handlers then read it via c.GetUint("doc_id").
// Does NOT query the database.
func ExtractDocID() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("doc_id")

		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id == 0 || id > math.MaxUint {
			c.AbortWithStatusJSON(http.StatusBadRequest, utils.ErrorResponse(fmt.Errorf("invalid_doc_id")))
			return
		}

		c.Set("doc_id", uint(id))
		c.Next()
	}
}
