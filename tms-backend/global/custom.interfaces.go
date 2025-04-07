package global

import (
	"net/http"

	//"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/gin-gonic/gin"
)


type Reply struct {
    Context *gin.Context
}

// View renders a template with optional data
func (r *Reply) View(page string, data interface{}) {
    r.Context.HTML(http.StatusOK, page, data)
}

/* type Request struct {
    Context *gin.Context
    User    *models.User // Add the User field to the request
} */


type Gender string
const (
	Male   Gender = "male"
	Female Gender = "female"
)

type ThemeType string
const (
	Standard ThemeType = "standard"
	Auditrakkr ThemeType = "auditrakkr"
)



