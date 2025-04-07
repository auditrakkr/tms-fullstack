package users

import (
	"net/http"
	"strconv"

	dto "github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
	"github.com/gin-gonic/gin"
)


type UserController struct {
	userService *UserService
}

func NewUserController(userService *UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

/* CREATE */
func (uc *UserController) CreateUser(c *gin.Context) {
	var createUserDto dto.CreateUserDto
	if err := c.ShouldBindJSON(&createUserDto); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request payload"})
	}
	user, err := uc.userService.CreateUser(&createUserDto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusCreated, gin.H{"user": user})
}


/* FIND */
func (uc *UserController) GetAllUsers(c *gin.Context) {
	users, err := uc.userService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (uc *UserController) FindOne(c *gin.Context) {
	id, err := strconv.ParseUint(c.Params.ByName("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
	}
	user, err := uc.userService.FindOne(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

/* DELETE */
func (uc *UserController) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Params.ByName("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
	}
	err = uc.userService.DeleteUser(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}


/* UPDATE */
func (uc *UserController) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Params.ByName("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
	}
	var updateUserDto dto.UpdateUserDto
	if err := c.ShouldBindJSON(&updateUserDto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
	}
	user, err := uc.userService.Update(uint(id), &updateUserDto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}
