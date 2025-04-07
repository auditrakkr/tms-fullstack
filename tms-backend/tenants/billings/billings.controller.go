package billings

import (
	dto "github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
	"github.com/gin-gonic/gin"
)


type BillingController struct {
	billingService *BillingService
}

func NewBillingController(billingService *BillingService) *BillingController {
	return &BillingController{
		billingService: billingService,
	}
}

func (bc *BillingController) CreateBilling(c *gin.Context) {
	var createBillingDto dto.CreateBillingDto
	if err := c.ShouldBindJSON(&createBillingDto); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request payload"})
		return
	}

	billing, err := bc.billingService.CreateBilling(&createBillingDto)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, gin.H{"billing": billing})
}

func (bc *BillingController) FindAll(c *gin.Context) {
	c.String(200, "This is billings")
}
