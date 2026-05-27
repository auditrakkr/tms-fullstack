package users

import (
	"net/http"
	"strconv"

	dto "github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/global"
	"github.com/auditrakkr/tms-fullstack/tms-backend/middleware"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/gin-gonic/gin"
)

// UserController exposes user-domain endpoints over HTTP.
type UserController struct {
	userService *UserService
}

func NewUserController(userService *UserService) *UserController {
	return &UserController{userService: userService}
}

// CreateUser godoc
// @Summary      Create a new user
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body body dto.CreateUserDto true "Create user payload"
// @Success      201  {object}  models.User
// @Failure      400  {object}  global.AppError
// @Failure      422  {object}  global.AppError
// @Security     BearerAuth
// @Router       /users [post]
func (uc *UserController) CreateUser(c *gin.Context) {
	var createUserDto dto.CreateUserDto
	if !middleware.BindAndValidate(c, &createUserDto) {
		return
	}
	baseURL := global.BuildBaseURL(c.Request)
	user, err := uc.userService.CreateUser(c.Request.Context(), &createUserDto, baseURL)
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"user": user})
}

// GetAllUsers godoc
// @Summary      List all users
// @Tags         users
// @Produce      json
// @Success      200  {object}  map[string][]models.User
// @Failure      500  {object}  global.AppError
// @Security     BearerAuth
// @Router       /users [get]
func (uc *UserController) GetAllUsers(c *gin.Context) {
	users, err := uc.userService.GetAllUsers(c.Request.Context())
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

// FindOne godoc
// @Summary      Get a user by ID
// @Tags         users
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  models.User
// @Failure      400  {object}  global.AppError
// @Failure      404  {object}  global.AppError
// @Security     BearerAuth
// @Router       /users/{id} [get]
func (uc *UserController) FindOne(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}
	user, err := uc.userService.FindOne(c.Request.Context(), uint(id))
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// UpdateUser godoc
// @Summary      Update a user
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int              true  "User ID"
// @Param        body body      dto.UpdateUserDto true "Update payload"
// @Success      200  {object}  models.User
// @Failure      400  {object}  global.AppError
// @Failure      422  {object}  global.AppError
// @Security     BearerAuth
// @Router       /users/{id} [patch]
func (uc *UserController) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}
	var updateUserDto dto.UpdateUserDto
	if !middleware.BindAndValidate(c, &updateUserDto) {
		return
	}
	user, err := uc.userService.Update(c.Request.Context(), uint(id), &updateUserDto)
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// DeleteUser godoc
// @Summary      Delete a user
// @Tags         users
// @Param        id   path  int  true  "User ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  global.AppError
// @Failure      500  {object}  global.AppError
// @Security     BearerAuth
// @Router       /users/{id} [delete]
func (uc *UserController) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}
	if err := uc.userService.DeleteUser(c.Request.Context(), uint(id)); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

/* PHOTO */

func (uc *UserController) GetUserPhoto(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}
	filePath, mimeType, err := uc.userService.GetUserPhoto(c.Request.Context(), uint(id))
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	if mimeType != "" {
		c.Header("Content-Type", mimeType)
	}
	c.File(filePath)
}

func (uc *UserController) SetUserPhoto(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}
	file, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "photo file is required"})
		return
	}
	if err := uc.userService.SetUserPhoto(c.Request.Context(), uint(id), file); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Photo updated successfully"})
}

/* PASSWORD RESET (SSR) */

func (uc *UserController) ShowResetPasswordForm(c *gin.Context) {
	token := c.Param("token")
	result, err := uc.userService.ResetPassword(c.Request.Context(), token, nil, global.BuildBaseURL(c.Request))
	if err != nil {
		c.HTML(http.StatusOK, "users/reset-password.html", gin.H{
			"title": global.APP_NAME + " - Reset Password", "sendForm": false,
			"notificationClass": "is-danger", "notificationMessage": "Invalid or expired reset link.",
			"notificationVisibility": "",
		})
		return
	}
	if result.TokenExpired {
		c.HTML(http.StatusOK, "users/reset-password.html", gin.H{
			"title": global.APP_NAME + " - Reset Password", "sendForm": false,
			"notificationClass": "is-danger", "notificationMessage": "Invalid token: expired",
			"notificationVisibility": "",
		})
		return
	}
	c.HTML(http.StatusOK, "users/reset-password.html", gin.H{
		"title": global.APP_NAME + " - Reset Password", "sendForm": true,
		"returnUrl": result.FormReturnURL, "notificationVisibility": "is-hidden",
	})
}

func (uc *UserController) ProcessResetPassword(c *gin.Context) {
	token := c.Param("token")
	newPassword := c.PostForm("password")
	result, err := uc.userService.ResetPassword(c.Request.Context(), token, &newPassword, global.BuildBaseURL(c.Request))
	if err != nil {
		c.HTML(http.StatusOK, "users/reset-password.html", gin.H{
			"title": global.APP_NAME + " - Reset Password", "sendForm": false,
			"notificationClass": "is-danger", "notificationMessage": "Password reset failed.",
			"notificationVisibility": "",
		})
		return
	}
	if result.NewPasswordSaved {
		c.HTML(http.StatusOK, "users/reset-password.html", gin.H{
			"title": global.APP_NAME + " - Reset Password", "sendForm": false,
			"notificationClass": "is-success", "notificationMessage": "New password successfully saved",
			"notificationVisibility": "",
		})
	}
}

func (uc *UserController) ConfirmPrimaryEmail(c *gin.Context) {
	result, err := uc.userService.ConfirmEmail(c.Request.Context(), c.Param("token"), true)
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	notifClass := "is-success"
	if !result.Success {
		notifClass = "is-danger"
	}
	c.HTML(http.StatusOK, "users/confirm-email-feedback.html", gin.H{
		"title":               global.APP_NAME + " - Confirm Email",
		"notificationClass":   notifClass,
		"notificationMessage": result.Message,
	})
}

func (uc *UserController) ConfirmBackupEmail(c *gin.Context) {
	result, err := uc.userService.ConfirmEmail(c.Request.Context(), c.Param("token"), false)
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	notifClass := "is-success"
	if !result.Success {
		notifClass = "is-danger"
	}
	c.HTML(http.StatusOK, "users/confirm-email-feedback.html", gin.H{
		"title":               global.APP_NAME + " - Confirm Email",
		"notificationClass":   notifClass,
		"notificationMessage": result.Message,
	})
}

/* ── BULK INSERT ─────────────────────────────────────────── */

func (uc *UserController) InsertUsers(c *gin.Context) {
	var createUserDtos []dto.CreateUserDto
	if err := c.ShouldBindJSON(&createUserDtos); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	baseURL := global.BuildBaseURL(c.Request)
	users, err := uc.userService.InsertUsers(c.Request.Context(), createUserDtos, baseURL)
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"users": users})
}

/* ── USER-SIDE TENANT RELATIONS ──────────────────────────── */

func (uc *UserController) CreateAndSetTenantForPrimaryContact(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}
	var createTenantDto dto.CreateTenantDto
	if !middleware.BindAndValidate(c, &createTenantDto) {
		return
	}
	if err := uc.userService.CreateAndSetTenantForPrimaryContact(c.Request.Context(), uint(id), &createTenantDto); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Tenant created and set for primary contact"})
}

func (uc *UserController) SetAsPrimaryContactByTenantID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}
	tenantID, err := strconv.ParseUint(c.Param("tenantId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	tenants, err := uc.userService.SetAsPrimaryContactByTenantID(c.Request.Context(), uint(id), uint(tenantID))
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"tenants": tenants})
}

func (uc *UserController) RemoveAsPrimaryContactByTenantID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}
	tenantID, err := strconv.ParseUint(c.Param("tenantId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	tenants, err := uc.userService.RemoveAsPrimaryContactByTenantID(c.Request.Context(), uint(id), uint(tenantID))
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"tenants": tenants})
}

/* ── USER-SIDE TEAM MEMBERSHIPS ──────────────────────────── */

func (uc *UserController) CreateTenantAndSetTeamMembership(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}
	var teamDto dto.CreateTenantTeamDto
	if !middleware.BindAndValidate(c, &teamDto) {
		return
	}
	if err := uc.userService.CreateTenantAndSetTeamMembership(c.Request.Context(), uint(id), &teamDto); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Tenant created and team membership set"})
}

func (uc *UserController) SetTeamMembershipByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}
	tenantID, err := strconv.ParseUint(c.Param("tenantId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	var rolesDto dto.CreateTenantTeamRolesDto
	if !middleware.BindAndValidate(c, &rolesDto) {
		return
	}
	if err := uc.userService.SetTeamMembershipByID(c.Request.Context(), uint(id), uint(tenantID), rolesDto.Roles); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Team membership set"})
}

func (uc *UserController) DeleteTeamMembershipByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}
	tenantID, err := strconv.ParseUint(c.Param("tenantId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	memberships, err := uc.userService.DeleteTeamMembershipByID(c.Request.Context(), uint(id), uint(tenantID))
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"memberships": memberships})
}

func (uc *UserController) UpdateTeamMembershipRolesByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}
	tenantID, err := strconv.ParseUint(c.Param("tenantId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	var rolesDto dto.UpdateTenantTeamRolesDto
	if !middleware.BindAndValidate(c, &rolesDto) {
		return
	}
	memberships, err := uc.userService.UpdateTeamMembershipRolesByID(c.Request.Context(), uint(id), uint(tenantID), rolesDto.Roles)
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"memberships": memberships})
}

/* ── USER-SIDE ACCOUNT OFFICER ASSIGNMENTS ───────────────── */

func (uc *UserController) CreateAndSetAccountOfficerForWhichTenant(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}
	var officerDto dto.CreateTenantAccountOfficerDto
	if !middleware.BindAndValidate(c, &officerDto) {
		return
	}
	if err := uc.userService.CreateAndSetAccountOfficerForWhichTenant(c.Request.Context(), uint(id), &officerDto); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Account officer assignment created"})
}

func (uc *UserController) SetTenantAccountOfficerForWhichTenantByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}
	tenantID, err := strconv.ParseUint(c.Param("tenantId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	var officerDto dto.CreateTenantAccountOfficerDto
	if !middleware.BindAndValidate(c, &officerDto) {
		return
	}
	if err := uc.userService.SetTenantAccountOfficerForWhichTenantByID(c.Request.Context(), uint(id), uint(tenantID), &officerDto); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Account officer assignment set"})
}

func (uc *UserController) DeleteTenantAccountOfficerForWhichTenantByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}
	tenantID, err := strconv.ParseUint(c.Param("tenantId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	if err := uc.userService.DeleteTenantAccountOfficerForWhichTenantByID(c.Request.Context(), uint(id), uint(tenantID)); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Account officer assignment removed"})
}

func (uc *UserController) UpdateTenantAccountOfficerForWhichTenantByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}
	tenantID, err := strconv.ParseUint(c.Param("tenantId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	var officerDto dto.UpdateTenantAccountOfficerDto
	if !middleware.BindAndValidate(c, &officerDto) {
		return
	}
	if err := uc.userService.UpdateTenantAccountOfficerForWhichTenantByID(c.Request.Context(), uint(id), uint(tenantID), &officerDto); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Account officer roles updated"})
}

// ensure models is used (referenced by swagger annotations)
var _ models.User
