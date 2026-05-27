package global

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AppError carries an HTTP status code alongside the error message so that
// controllers can map service errors to the correct HTTP response without
// inspecting error strings.
type AppError struct {
	HTTPStatus int    `json:"-"`
	Message    string `json:"message"`
	Err        error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Err }

// Constructor helpers

func NotFound(msg string) *AppError {
	return &AppError{HTTPStatus: http.StatusNotFound, Message: msg}
}

func BadRequest(msg string) *AppError {
	return &AppError{HTTPStatus: http.StatusBadRequest, Message: msg}
}

func Unauthorized(msg string) *AppError {
	if msg == "" {
		msg = "unauthorized"
	}
	return &AppError{HTTPStatus: http.StatusUnauthorized, Message: msg}
}

func Forbidden(msg string) *AppError {
	return &AppError{HTTPStatus: http.StatusForbidden, Message: msg}
}

func Internal(err error) *AppError {
	return &AppError{HTTPStatus: http.StatusInternalServerError, Message: "internal server error", Err: err}
}

func Conflict(msg string) *AppError {
	return &AppError{HTTPStatus: http.StatusConflict, Message: msg}
}

// RespondWithError writes the appropriate JSON error response to c.
// If err wraps an *AppError the status code is taken from it; otherwise
// a generic 500 is returned.
func RespondWithError(c *gin.Context, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.HTTPStatus, gin.H{"message": appErr.Message})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
}
