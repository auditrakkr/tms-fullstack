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


const APP_NAME = "TMS"
const APP_VERSION = "1.0.0"
const APP_DESCRIPTION = "Tenant Management System"
const API_VERSION = "v1"
const API_BASE_URL = "/api/v1"
const USE_API_VERSION_IN_URL = true
const AUTO_SEND_CONFIRM_EMAIL = true
const AUTO_SEND_WELCOME_EMAIL = true