package billings

import (
	"net/http"

	dto "github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/global"
	"github.com/auditrakkr/tms-fullstack/tms-backend/middleware"
	"github.com/gin-gonic/gin"
)

type BillingController struct {
	billingService *BillingService
}

func NewBillingController(billingService *BillingService) *BillingController {
	return &BillingController{billingService: billingService}
}

func (bc *BillingController) CreateBilling(c *gin.Context) {
	var createBillingDto dto.CreateBillingDto
	if !middleware.BindAndValidate(c, &createBillingDto) {
		return
	}
	billing, err := bc.billingService.CreateBilling(c.Request.Context(), &createBillingDto)
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"billing": billing})
}

func (bc *BillingController) FindAll(c *gin.Context) {
	billings, err := bc.billingService.FindAll(c.Request.Context())
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"billings": billings})
}
