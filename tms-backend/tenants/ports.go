package tenants

import (
	"context"

	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
)

// RegionLookup is the slice of regions.RegionService that TenantService needs.
// Defining it here breaks the concrete import cycle risk between tenants and regions.
type RegionLookup interface {
	FindByRegionName(ctx context.Context, regionName string) (*models.Region, error)
}
