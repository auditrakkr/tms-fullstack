package global

type TenantStatus string
const (
	Active    TenantStatus = "active"
	Suspended TenantStatus = "suspended"
	Owing     TenantStatus = "owing"
)


type TenantTeamRole string
const (
	A TenantTeamRole = "Admin"
	M TenantTeamRole = "Manager"
	E TenantTeamRole = "Employee"
)

type TenantAccountOfficerRole string
const (
	AOM TenantAccountOfficerRole = "Account Officer Manager"
	AOT TenantAccountOfficerRole = "Account Officer Tech-Support"
)

type LandlordRoles string
const (
	AdminLandlord	LandlordRoles = "admin_landlord"
	SuperAdminLandlord	LandlordRoles = "super_admin_landlord"
	UserLandlord	LandlordRoles = "user_landlord"
)

type TenantRoles string
const (
	Admin	TenantRoles = "admin"
	SuperAdmin	TenantRoles = "super_admin"
	User	TenantRoles = "user"
)