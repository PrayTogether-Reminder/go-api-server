package http

import (
	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/error"
	"github.com/gin-gonic/gin"
)

// RespondError sends an error response with logging
//
// Usage:
//
//	if err := service.DoSomething(); err != nil {
//	    http.RespondError(c, err, sharedError.InternalServerError)
//	    return
//	}
func RespondError(c *gin.Context, err error, errResp sharedError.ErrorResponse) {
	// Add error to context for middleware logging
	_ = c.Error(err)

	// Send error response
	c.JSON(errResp.Status, errResp)
}
