package tenants

import (
	"fmt"
	"net/http"

	"github.com/auditrakkr/tms-fullstack/tms-backend/database"
	"github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/auditrakkr/tms-fullstack/tms-backend/regions"
	"github.com/auditrakkr/tms-fullstack/tms-backend/repositories"
	"github.com/gin-gonic/gin"
)

//initialize the repositories for Tenant
/* tenantRepository := repositories.Repository[models.Tenant]{DB: database.DB}
tenantTeamRepository := repositories.Repository[models.TenantTeam]{DB: database.DB}
tenantAccountOfficerRepository := repositories.Repository[models.TenantAccountOfficer]{DB: database.DB}
tenantAccountOfficerRoleRepository := repositories.Repository[models.TenantAccountOfficerRole]{DB: database.DB}
userRepository := repositories.Repository[models.User]{DB: database.DB}
customThemeRepository := repositories.Repository[models.CustomTheme]{DB: database.DB}
themeRepository := repositories.Repository[models.Theme]{DB: database.DB}
billingRepository := repositories.Repository[models.Billing]{DB: database.DB}
tenantConfigDetailRepository := repositories.Repository[models.TenantConfigDetail]{DB: database.DB}
*/

type Request struct {
    Context *gin.Context
    User    *models.User // Add the User field to the request
	}

// TenantService struct
type TenantService struct {
	tenantRepo     repositories.Repository[models.Tenant]
	regionRepo     repositories.Repository[models.Region]
	regionService  *regions.RegionService
	tenantTeamRepo repositories.Repository[models.TenantTeam]
	userRepo       repositories.Repository[models.User]
}

// NewTenantService creates a new instance of TenantService
func NewTenantService() *TenantService {
	return &TenantService{
		tenantRepo:    repositories.Repository[models.Tenant]{DB: database.DB},
		regionRepo:    repositories.Repository[models.Region]{DB: database.DB},
		regionService: regions.NewRegionService(),
		userRepo: repositories.Repository[models.User]{DB: database.DB},
	}
}

/* func (s *TenantService)createAndSetTenantConfigDetail(tenantId uint, createTenantConfigDetailDto *dto.CreateTenantConfigDetailDto) error {
	tenantConfigDetail := &models.TenantConfigDetail{
		TenantID: tenantId,
		// Set other fields from createTenantConfigDetailDto
	}

	err := s.tenantRepo.Save(tenantConfigDetail)
	if err != nil {
		return fmt.Errorf("failed to create tenant config detail: %w", err)
	}
	return nil
} */

// CreateTenant creates a new tenant
func (s *TenantService) CreateTenant(createTenantDto *dto.CreateTenantDto, createPrimaryContact uint, req Request) (*models.Tenant, error) {
	//Step 1: Seperate tenantConfigDetail
	//tenantConfigDetail := createTenantDto.TenantConfigDetail

	// Step 2: Map the remaining fields to the Tenant model
	newTenant := &models.Tenant{}
	err := newTenant.MapFromCreateTenantDto(createTenantDto)
	if err != nil {
		http.Error(req.Context.Writer, "failed to map tenant DTO", http.StatusInternalServerError)
		return nil, fmt.Errorf("failed to map tenant DTO: %w", err)
	}

	// Fetch regionRootDomainName and set it for the new tenant
	region, err := s.regionService.FindByRegionName(createTenantDto.RegionName)
	if err != nil {
		return nil, fmt.Errorf("failed to find region: %w", err)
	}
	newTenant.RegionRootDomain = region.RootDomainName

	if createPrimaryContact != 1 {
		primaryContact, err := s.userRepo.FindOne(map[string]any{"primary_email_address": createTenantDto.PrimaryContact.PrimaryEmailAddress})
		if err != nil {
			http.Error(req.Context.Writer, "failed to find primary contact", http.StatusInternalServerError)
			return nil, fmt.Errorf("failed to find primary contact: %w", err)
		}
		if primaryContact == nil {
			http.Error(req.Context.Writer, "primary contact not found", http.StatusNotFound)
			return nil, fmt.Errorf("primary contact not found")
		}
		newTenant.PrimaryContactID = primaryContact.ID
	} else {
		//TODO: Create a new primary contact user and hash the password
		fmt.Println("Creating new primary contact user")
	}

	tenant, err := s.tenantRepo.Save(newTenant)
	if err != nil {
		http.Error(req.Context.Writer, "failed to create tenant", http.StatusInternalServerError)
		return nil, fmt.Errorf("failed to create tenant")
	}

	if createPrimaryContact == 1 {
		user, err := s.userRepo.FindOne(map[string]any{"primary_email_address": tenant.PrimaryContact.PrimaryEmailAddress})
		if err != nil || user == nil {
			http.Error(req.Context.Writer, "failed to find primary contact", http.StatusInternalServerError)
			return nil, fmt.Errorf("failed to find primary contact: %w", err)
		}
		if err == nil && !tenant.PrimaryContact.IsPrimaryEmailVerified{
			// TODO: Send email to primary contact for verification
		}
	}

	// Create tenantConfigDetail if primary email address is verified
    /* if createdTenant.PrimaryContact != nil && createdTenant.PrimaryContact.IsPrimaryEmailAddressVerified {
        err = s.createAndSetTenantConfigDetail(createdTenant.ID, tenantConfigDetail)
        if err != nil {
            return nil, fmt.Errorf("failed to create tenant config detail: %w", err)
        }
    } */

	return tenant, nil
}

/* Update Section */

func (s *TenantService) Update(tenantId uint, updateTenantDto *dto.UpdateTenantDto) (*models.Tenant, error) {
	// Find the tenant by ID
	tenant, err := s.tenantRepo.FindByID(tenantId)
	if err != nil {
		return nil, fmt.Errorf("failed to find tenant: %w", err)
	}

	// Update the tenant fields
	tenant.Name = *updateTenantDto.Name
	tenant.Address = *updateTenantDto.Address
	tenant.MoreInfo = *updateTenantDto.MoreInfo
	tenant.Logo = *updateTenantDto.Logo
	tenant.LogoMimeType = *updateTenantDto.LogoMimeType
	tenant.Status = *updateTenantDto.Status

	err = s.tenantRepo.Update(tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	return tenant, nil
}


/* Delete Section */

func (s *TenantService) Delete(tenantId uint) error {
	err := s.tenantRepo.Delete(tenantId)
	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}
	return nil
}

// Remove the tenant specified. Returns tenant removed

func (s *TenantService) Remove(tenant *models.Tenant) (*models.Tenant, error) {
	removedTenant, err := s.tenantRepo.Remove(tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to remove tenant: %w", err)
	}
	return &removedTenant, nil

}


/* Read Section */

func (s *TenantService) GetAllTenants() ([]models.Tenant, error) {
	tenants, err := s.tenantRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get all tenants: %w", err)
	}
	return tenants, nil
}

func (s *TenantService) FindAllWithOptions(findOptions map[string]any) ([]models.Tenant, int64, error) {
	tenants, totalCount, err := s.tenantRepo.FindAndCount(findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get all tenants with options: %w", err)
	}
	return tenants, totalCount, nil
}


func (s *TenantService) FindOne(tenantId uint) (*models.Tenant, error) {
	tenant, err := s.tenantRepo.FindByID(tenantId)
	if err != nil {
		return nil, fmt.Errorf("failed to find tenant: %w", err)
	}
	if tenant == nil {
		return nil, fmt.Errorf("tenant not found")
	}
	return tenant, nil
}

func (s *TenantService) FindActiveTenantsByRegionName(regionName string) ([]models.Tenant, error) {
	var tenants []models.Tenant
	err := s.tenantRepo.CreateQueryBuilder().Where("region_name = ? AND status = ?", regionName, "active").Find(&tenants).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find active tenants by region name: %w", err)
	}
	return tenants, nil
}

func (s *TenantService) FindTenantsByRegionName(regionName string) (*[]models.Tenant, error) {
	var tenants []models.Tenant

	err := s.tenantRepo.CreateQueryBuilder().Where("region_name = ?", regionName).Find(&tenants).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find tenants by region name: %w", err)
	}
	return &tenants, nil
}


