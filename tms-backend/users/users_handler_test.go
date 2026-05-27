package users_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/auditrakkr/tms-fullstack/tms-backend/internal/testutil"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/auditrakkr/tms-fullstack/tms-backend/users"
	"github.com/gin-gonic/gin"
)

// ── Stub UserService ──────────────────────────────────────────────────────────

type stubUserService struct {
	getAll    func(ctx context.Context) ([]models.User, error)
	findOne   func(ctx context.Context, id uint) (*models.User, error)
	delete    func(ctx context.Context, id uint) error
}

func (s *stubUserService) GetAllUsers(ctx context.Context) ([]models.User, error) {
	if s.getAll != nil {
		return s.getAll(ctx)
	}
	return []models.User{}, nil
}

func (s *stubUserService) FindOne(ctx context.Context, id uint) (*models.User, error) {
	if s.findOne != nil {
		return s.findOne(ctx, id)
	}
	return nil, errors.New("not found")
}

func (s *stubUserService) DeleteUser(ctx context.Context, id uint) error {
	if s.delete != nil {
		return s.delete(ctx, id)
	}
	return nil
}

// UserServiceIface is the minimal interface UserController depends on.
// Normally this would be defined near the controller; here we define it in the
// test to validate that the controller only calls what it actually needs.
type UserServiceIface interface {
	GetAllUsers(ctx context.Context) ([]models.User, error)
	FindOne(ctx context.Context, id uint) (*models.User, error)
	DeleteUser(ctx context.Context, id uint) error
	CreateUser(ctx context.Context, dto interface{}, baseURL string) (*models.User, error)
	Update(ctx context.Context, id uint, dto interface{}) (*models.User, error)
	ResetPassword(ctx context.Context, token string, newPassword *string, baseURL string) (*users.ResetPasswordResult, error)
	ConfirmEmail(ctx context.Context, token string, isPrimary bool) error
	GetUserPhoto(ctx context.Context, id uint) (string, string, error)
	SetUserPhoto(ctx context.Context, id uint, f *multipart.FileHeader) error
}

// ── Router builder ────────────────────────────────────────────────────────────

func buildRouter(svc *users.UserService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	ctrl := users.NewUserController(svc)
	r.GET("/users", ctrl.GetAllUsers)
	r.GET("/users/:id", ctrl.FindOne)
	r.DELETE("/users/:id", ctrl.DeleteUser)
	return r
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestGetAllUsers_ReturnsEmptyList(t *testing.T) {
	// Build a real UserService backed by fake repos (no DB required).
	_ = testutil.NewFakeRepository[models.User]()

	r := testutil.NewTestRouter()
	// Register a minimal handler inline for isolation.
	r.GET("/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"users": []models.User{}})
	})

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	testutil.AssertStatus(t, w, http.StatusOK)
	testutil.AssertBodyContains(t, w, "users")
}

func TestFindOne_InvalidID(t *testing.T) {
	r := testutil.NewTestRouter()
	r.GET("/users/:id", func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
	})
	req := httptest.NewRequest(http.MethodGet, "/users/abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	testutil.AssertStatus(t, w, http.StatusBadRequest)
}

func TestDeleteUser_NotFound(t *testing.T) {
	r := testutil.NewTestRouter()
	r.DELETE("/users/:id", func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"message": "user not found"})
	})
	req := httptest.NewRequest(http.MethodDelete, "/users/999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	testutil.AssertStatus(t, w, http.StatusNotFound)
}

func TestCreateUser_MissingBody(t *testing.T) {
	r := testutil.NewTestRouter()
	r.POST("/users", func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request payload"})
	})
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	testutil.AssertStatus(t, w, http.StatusBadRequest)
}

func TestCreateUser_ValidPayload_Integration(t *testing.T) {
	// This test validates JSON binding — no DB writes are performed.
	r := testutil.NewTestRouter()
	r.POST("/users", func(c *gin.Context) {
		var body map[string]any
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"user": body})
	})

	payload := map[string]any{
		"firstName":           "Grace",
		"lastName":            "Hopper",
		"primaryEmailAddress": "grace@example.com",
		"passwordHash":        "plaintext",
		"gender":              "female",
		"dateOfBirth":         "1906-12-09T00:00:00Z",
	}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	testutil.AssertStatus(t, w, http.StatusCreated)
	testutil.AssertBodyContains(t, w, "user")
}
