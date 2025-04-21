package roles

import (
	"net/http"
	"strconv"

	dto "github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/gin-gonic/gin"
)

type RoleController struct {
	roleService *RoleService
}

func NewRoleController(roleService *RoleService) *RoleController {
	return &RoleController{
		roleService: roleService,
	}
}


/* CREATE */

func (rc *RoleController) CreateRole(c *gin.Context) {
	var createRoleDto dto.CreateUserDto
	if err := c.ShouldBindJSON(&createRoleDto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	role, err := rc.roleService.CreateRole(&createRoleDto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"role": role})
}

/* UPDATE */

func (rc *RoleController) UpdateRole(c *gin.Context) {
	roleId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	var updateRoleDto dto.UpdateUserDto
	if err := c.ShouldBindJSON(&updateRoleDto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	role, err := rc.roleService.Update(uint(roleId), &updateRoleDto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"role": role})
}

func (rc *RoleController) SaveRole(c *gin.Context) {
	var role *models.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	role, err := rc.roleService.Save(role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"role": role})	}


/* FIND */
func (rc *RoleController) GetAllRoles(c *gin.Context) {
	roles, err := rc.roleService.GetAllRoles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"roles": roles})
}


func (rc *RoleController) FindOne(c *gin.Context) {
	roleId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	role, err := rc.roleService.FindOne(uint(roleId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"role": role})
}

/* DELETE */
func (rc *RoleController) DeleteRole(c *gin.Context) {
	roleId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	err = rc.roleService.Delete(uint(roleId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Role deleted successfully"})
}