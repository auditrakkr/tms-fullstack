package roles

import (
	"net/http"
	"strconv"

	dto "github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/global"
	"github.com/auditrakkr/tms-fullstack/tms-backend/middleware"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/gin-gonic/gin"
)

// RoleController handles role CRUD over HTTP.
type RoleController struct {
	roleService *RoleService
}

func NewRoleController(roleService *RoleService) *RoleController {
	return &RoleController{roleService: roleService}
}

// CreateRole godoc
// @Summary      Create a new role
// @Tags         roles
// @Accept       json
// @Produce      json
// @Param        body body dto.CreateRoleDto true "Create role payload"
// @Success      201 {object} models.Role
// @Failure      400 {object} global.AppError
// @Security     BearerAuth
// @Router       /roles [post]
func (rc *RoleController) CreateRole(c *gin.Context) {
	var createRoleDto dto.CreateRoleDto
	if !middleware.BindAndValidate(c, &createRoleDto) {
		return
	}
	role, err := rc.roleService.CreateRole(c.Request.Context(), &createRoleDto)
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"role": role})
}

// UpdateRole godoc
// @Summary      Update a role
// @Tags         roles
// @Accept       json
// @Produce      json
// @Param        id   path int              true "Role ID"
// @Param        body body dto.UpdateRoleDto true "Update payload"
// @Success      200 {object} models.Role
// @Failure      400 {object} global.AppError
// @Security     BearerAuth
// @Router       /roles/{id} [patch]
func (rc *RoleController) UpdateRole(c *gin.Context) {
	roleId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid role ID"})
		return
	}
	var updateRoleDto dto.UpdateRoleDto
	if !middleware.BindAndValidate(c, &updateRoleDto) {
		return
	}
	role, err := rc.roleService.Update(c.Request.Context(), uint(roleId), &updateRoleDto)
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"role": role})
}

// SaveRole godoc
// @Summary      Create or update a role (upsert)
// @Tags         roles
// @Accept       json
// @Produce      json
// @Param        body body models.Role true "Role"
// @Success      200 {object} models.Role
// @Security     BearerAuth
// @Router       /roles [put]
func (rc *RoleController) SaveRole(c *gin.Context) {
	var role *models.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request payload"})
		return
	}
	saved, err := rc.roleService.Save(c.Request.Context(), role)
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"role": saved})
}

// GetAllRoles godoc
// @Summary      List all roles
// @Tags         roles
// @Produce      json
// @Success      200 {object} map[string][]models.Role
// @Security     BearerAuth
// @Router       /roles [get]
func (rc *RoleController) GetAllRoles(c *gin.Context) {
	roles, err := rc.roleService.GetAllRoles(c.Request.Context())
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"roles": roles})
}

// FindOne godoc
// @Summary      Get a role by ID
// @Tags         roles
// @Produce      json
// @Param        id  path int true "Role ID"
// @Success      200 {object} models.Role
// @Failure      400 {object} global.AppError
// @Security     BearerAuth
// @Router       /roles/{id} [get]
func (rc *RoleController) FindOne(c *gin.Context) {
	roleId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid role ID"})
		return
	}
	role, err := rc.roleService.FindOne(c.Request.Context(), uint(roleId))
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"role": role})
}

// DeleteRole godoc
// @Summary      Delete a role
// @Tags         roles
// @Param        id path int true "Role ID"
// @Success      200 {object} map[string]string
// @Security     BearerAuth
// @Router       /roles/{id} [delete]
func (rc *RoleController) DeleteRole(c *gin.Context) {
	roleId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid role ID"})
		return
	}
	if err := rc.roleService.Delete(c.Request.Context(), uint(roleId)); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Role deleted successfully"})
}
