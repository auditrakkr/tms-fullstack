package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// BindAndValidate deserialises the JSON request body into out and runs
// struct-tag validation. Returns true if both succeed. On failure it writes
// the appropriate 400/422 JSON response and returns false — the caller should
// return immediately without sending another response.
func BindAndValidate[T any](c *gin.Context, out *T) bool {
	if err := c.ShouldBindJSON(out); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return false
	}
	if err := validate.Struct(out); err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			errs := make(map[string]string, len(validationErrs))
			for _, e := range validationErrs {
				errs[e.Field()] = e.Tag()
			}
			c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": errs})
		} else {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"message": err.Error()})
		}
		return false
	}
	return true
}
