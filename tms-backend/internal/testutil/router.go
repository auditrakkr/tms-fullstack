package testutil

import (
	"github.com/gin-gonic/gin"
)

// NewTestRouter returns a Gin engine in test mode with no middleware so
// individual test cases can add exactly what they need.
func NewTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}
