package tenants

import (
	"net/http"
	"strconv"

	dto "github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/global"
	"github.com/auditrakkr/tms-fullstack/tms-backend/middleware"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/gin-gonic/gin"
)

// TenantController handles tenant CRUD over HTTP.
type TenantController struct {
	tenantService *TenantService
}

func NewTenantController(tenantService *TenantService) *TenantController {
	return &TenantController{tenantService: tenantService}
}

/* ── CRUD ────────────────────────────────────────────────── */

// CreateTenant godoc
// @Summary      Create a new tenant
// @Tags         tenants
// @Accept       json
// @Produce      json
// @Param        createPrimaryContact query int false "1 = create new primary contact, 0 = use existing"
// @Param        body body dto.CreateTenantDto true "Create tenant payload"
// @Success      201 {object} models.Tenant
// @Failure      400 {object} global.AppError
// @Security     BearerAuth
// @Router       /tenants [post]
func (tc *TenantController) CreateTenant(c *gin.Context) {
	var createTenantDto dto.CreateTenantDto
	if !middleware.BindAndValidate(c, &createTenantDto) {
		return
	}
	createPrimaryContact, err := strconv.ParseUint(c.Query("createPrimaryContact"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid createPrimaryContact value"})
		return
	}
	tenant, err := tc.tenantService.CreateTenant(c.Request.Context(), &createTenantDto, uint(createPrimaryContact))
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"tenant": tenant})
}

func (tc *TenantController) GetAllTenants(c *gin.Context) {
	tenants, err := tc.tenantService.GetAllTenants(c.Request.Context())
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"tenants": tenants})
}

func (tc *TenantController) FindOne(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	tenant, err := tc.tenantService.FindOne(c.Request.Context(), uint(id))
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"tenant": tenant})
}

func (tc *TenantController) FindActiveTenantsByRegionName(c *gin.Context) {
	tenants, err := tc.tenantService.FindActiveTenantsByRegionName(c.Request.Context(), c.Param("regionName"))
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"tenants": tenants})
}

func (tc *TenantController) UpdateTenant(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	var updateTenantDto dto.UpdateTenantDto
	if !middleware.BindAndValidate(c, &updateTenantDto) {
		return
	}
	tenant, err := tc.tenantService.Update(c.Request.Context(), uint(id), &updateTenantDto)
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"tenant": tenant})
}

func (tc *TenantController) DeleteTenant(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	if err := tc.tenantService.Delete(c.Request.Context(), uint(id)); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tenant deleted successfully"})
}

/* ── BULK INSERT ─────────────────────────────────────────── */

func (tc *TenantController) InsertTenants(c *gin.Context) {
	var createDtos []dto.CreateTenantDto
	if err := c.ShouldBindJSON(&createDtos); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	tenants, err := tc.tenantService.InsertTenants(c.Request.Context(), createDtos)
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"tenants": tenants})
}

/* ── ADDITIONAL QUERIES ──────────────────────────────────── */

func (tc *TenantController) FindByActive(c *gin.Context) {
	active := c.Param("active") == "true"
	tenants, err := tc.tenantService.FindByActive(c.Request.Context(), active)
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"tenants": tenants})
}

func (tc *TenantController) FindByUniqueName(c *gin.Context) {
	subDomain := c.Query("subDomain")
	rootDomain := c.Query("rootDomain")
	if subDomain == "" || rootDomain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "subDomain and rootDomain are required"})
		return
	}
	tenant, err := tc.tenantService.FindByUniqueName(c.Request.Context(), subDomain, rootDomain)
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"tenant": tenant})
}

func (tc *TenantController) CheckDomainExists(c *gin.Context) {
	subDomain := c.Query("subDomain")
	rootDomain := c.Query("rootDomain")
	if subDomain == "" || rootDomain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "subDomain and rootDomain are required"})
		return
	}
	exists, err := tc.tenantService.CheckDomainExists(c.Request.Context(), subDomain, rootDomain)
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"exists": exists})
}

func (tc *TenantController) GetAccountOfficersByTenant(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	officers, err := tc.tenantService.GetAccountOfficersByTenant(c.Request.Context(), uint(id))
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"accountOfficers": officers})
}

func (tc *TenantController) GetTenantIdentities(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	identity, err := tc.tenantService.GetTenantIdentities(c.Request.Context(), uint(id))
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"identity": identity})
}

/* ── PRIMARY CONTACT ─────────────────────────────────────── */

func (tc *TenantController) CreateAndSetPrimaryContact(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	var createUserDto dto.CreateUserDto
	if !middleware.BindAndValidate(c, &createUserDto) {
		return
	}
	if err := tc.tenantService.CreateAndSetPrimaryContact(c.Request.Context(), uint(id), &createUserDto); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Primary contact created and set successfully"})
}

func (tc *TenantController) SetPrimaryContactByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}
	if err := tc.tenantService.SetPrimaryContactByID(c.Request.Context(), uint(id), uint(userID)); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Primary contact set successfully"})
}

func (tc *TenantController) UnsetPrimaryContactByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	if err := tc.tenantService.UnsetPrimaryContactByID(c.Request.Context(), uint(id)); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Primary contact unset successfully"})
}

/* ── TEAM MEMBERS ────────────────────────────────────────── */

func (tc *TenantController) CreateAndSetTeamMember(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	var teamDto dto.CreateTenantTeamDto
	if !middleware.BindAndValidate(c, &teamDto) {
		return
	}
	if err := tc.tenantService.CreateAndSetTeamMember(c.Request.Context(), uint(id), &teamDto); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Team member added successfully"})
}

func (tc *TenantController) SetTeamMemberByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}
	var roles TenantTeamRoles
	if !middleware.BindAndValidate(c, &roles) {
		return
	}
	if err := tc.tenantService.SetTeamMemberByID(c.Request.Context(), uint(id), uint(userID), &roles); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Team member set successfully"})
}

func (tc *TenantController) DeleteTeamMemberByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}
	if err := tc.tenantService.DeleteTeamMemberByID(c.Request.Context(), uint(id), uint(userID)); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Team member removed successfully"})
}

func (tc *TenantController) UpdateTeamMemberByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}
	var roles TenantTeamRoles
	if !middleware.BindAndValidate(c, &roles) {
		return
	}
	if err := tc.tenantService.UpdateTeamMemberByID(c.Request.Context(), uint(id), uint(userID), &roles); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Team member roles updated"})
}

/* ── ACCOUNT OFFICERS ────────────────────────────────────── */

func (tc *TenantController) CreateAndSetTenantAccountOfficer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	var officerDto dto.CreateTenantAccountOfficerDto
	if !middleware.BindAndValidate(c, &officerDto) {
		return
	}
	if err := tc.tenantService.CreateAndSetTenantAccountOfficer(c.Request.Context(), uint(id), &officerDto); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Account officer added successfully"})
}

func (tc *TenantController) SetTenantAccountOfficerByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}
	var roles TenantAccountOfficerRoles
	if !middleware.BindAndValidate(c, &roles) {
		return
	}
	if err := tc.tenantService.SetTenantAccountOfficerByID(c.Request.Context(), uint(id), uint(userID), &roles); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Account officer set successfully"})
}

func (tc *TenantController) DeleteTenantAccountOfficerByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}
	if err := tc.tenantService.DeleteTenantAccountOfficerByID(c.Request.Context(), uint(id), uint(userID)); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Account officer removed successfully"})
}

func (tc *TenantController) UpdateTenantAccountOfficerByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}
	var roles TenantAccountOfficerRoles
	if !middleware.BindAndValidate(c, &roles) {
		return
	}
	if err := tc.tenantService.UpdateTenantAccountOfficerByID(c.Request.Context(), uint(id), uint(userID), &roles); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Account officer roles updated"})
}

/* ── CUSTOM THEME ────────────────────────────────────────── */

func (tc *TenantController) CreateAndSetCustomTheme(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	var themeDto dto.CreateCustomThemeDto
	if !middleware.BindAndValidate(c, &themeDto) {
		return
	}
	if err := tc.tenantService.CreateAndSetCustomTheme(c.Request.Context(), uint(id), &themeDto); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Custom theme created and set"})
}

func (tc *TenantController) SetCustomThemeByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	themeID, err := strconv.ParseUint(c.Param("themeId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid theme ID"})
		return
	}
	if err := tc.tenantService.SetCustomThemeByID(c.Request.Context(), uint(id), uint(themeID)); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Custom theme set successfully"})
}

func (tc *TenantController) UnsetCustomThemeByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	if err := tc.tenantService.UnsetCustomThemeByID(c.Request.Context(), uint(id)); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Custom theme unset successfully"})
}

/* ── THEME ASSOCIATIONS ──────────────────────────────────── */

func (tc *TenantController) AddThemeByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	themeID, err := strconv.ParseUint(c.Param("themeId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid theme ID"})
		return
	}
	if err := tc.tenantService.AddThemeByID(c.Request.Context(), uint(id), uint(themeID)); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Theme added to tenant"})
}

func (tc *TenantController) RemoveThemeByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	themeID, err := strconv.ParseUint(c.Param("themeId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid theme ID"})
		return
	}
	if err := tc.tenantService.RemoveThemeByID(c.Request.Context(), uint(id), uint(themeID)); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Theme removed from tenant"})
}

/* ── BILLING ASSOCIATIONS ────────────────────────────────── */

func (tc *TenantController) CreateAndAddBilling(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	var billingDto dto.CreateBillingDto
	if !middleware.BindAndValidate(c, &billingDto) {
		return
	}
	if err := tc.tenantService.CreateAndAddBilling(c.Request.Context(), uint(id), &billingDto); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Billing created and added to tenant"})
}

func (tc *TenantController) AddBillingByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	billingID, err := strconv.ParseUint(c.Param("billingId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid billing ID"})
		return
	}
	if err := tc.tenantService.AddBillingByID(c.Request.Context(), uint(id), uint(billingID)); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Billing added to tenant"})
}

func (tc *TenantController) RemoveBillingByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	billingID, err := strconv.ParseUint(c.Param("billingId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid billing ID"})
		return
	}
	if err := tc.tenantService.RemoveBillingByID(c.Request.Context(), uint(id), uint(billingID)); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Billing removed from tenant"})
}

/* ── LOGO ────────────────────────────────────────────────── */

func (tc *TenantController) SetTenantLogo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	file, err := c.FormFile("logo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "logo file is required"})
		return
	}
	if err := tc.tenantService.SetTenantLogo(c.Request.Context(), uint(id), file); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Logo updated successfully"})
}

func (tc *TenantController) GetTenantLogo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	filePath, mimeType, err := tc.tenantService.GetTenantLogo(c.Request.Context(), uint(id))
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	if mimeType != "" {
		c.Header("Content-Type", mimeType)
	}
	c.File(filePath)
}

/* ── TENANT CONFIG DETAIL ────────────────────────────────── */

func (tc *TenantController) SetTenantConfigDetailByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	configDetailID, err := strconv.ParseUint(c.Param("configDetailId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid config detail ID"})
		return
	}
	if err := tc.tenantService.SetTenantConfigDetailByID(c.Request.Context(), uint(id), uint(configDetailID)); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tenant config detail set successfully"})
}

func (tc *TenantController) UnsetTenantConfigDetailByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	if err := tc.tenantService.UnsetTenantConfigDetailByID(c.Request.Context(), uint(id)); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tenant config detail unset successfully"})
}

/* ── CREATE-AND-SET TENANT CONFIG DETAIL (Phase 2) ──────── */

func (tc *TenantController) CreateAndSetTenantConfigDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	var createDto dto.CreateTenantConfigDetailDto
	if !middleware.BindAndValidate(c, &createDto) {
		return
	}
	if err := tc.tenantService.CreateAndSetTenantConfigDetail(c.Request.Context(), uint(id), &createDto); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Tenant config detail created, schema provisioned, and Redis updated"})
}

/* ── REDIS SYNC ADMIN (Phase 2) ──────────────────────────── */

func (tc *TenantController) SyncRegionRedis(c *gin.Context) {
	regionName := c.Param("name")
	if err := tc.tenantService.SetTenantPropertiesInRedisByRegionName(c.Request.Context(), regionName); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Region Redis sync complete"})
}

func (tc *TenantController) SyncTenantRedis(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	if err := tc.tenantService.SetTenantPropertiesInRedisByTenantID(c.Request.Context(), uint(id)); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tenant Redis sync complete"})
}

func (tc *TenantController) UnsyncTenantRedis(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	if err := tc.tenantService.UnsetTenantPropertiesInRedisByTenantID(c.Request.Context(), uint(id)); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tenant Redis keys removed"})
}

func (tc *TenantController) SetTenantStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant ID"})
		return
	}
	var body struct {
		Status global.TenantStatus `json:"status" validate:"required"`
	}
	if !middleware.BindAndValidate(c, &body) {
		return
	}
	if err := tc.tenantService.SetTenantStatusInRedisByTenantID(c.Request.Context(), uint(id), body.Status); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tenant status updated in Redis"})
}

// ensure compile-time check that models is used
var _ models.Tenant
