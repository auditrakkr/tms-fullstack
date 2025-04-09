package auth


// AuthTokenPayload represents the JWT payload
type AuthTokenPayload struct {
	Username string `json:"username"`
	Sub struct {
		ID uint `json:"id"`
		FirstName string `json:"first_name"`
		LastName string `json:"last_name"`
		Landlord bool `json:"landlord"`
		Roles []string `json:"roles"`
	} `json:"sub"`
}

// JwtConstants holds configuration for JWT
type JwtConstants struct {
	Secret                   string
	SecretKeyExpiration      int
	RefreshSecret            string
	RefreshSecretKeyExpiration string
	SecretPrivateKey         string
	SecretPrivateKeyPassphrase string
	SecretPublicKey          string
	SignAlgorithm            string
}

// GoogleOidcConstants holds configuration for Google OIDC
type GoogleOidcConstants struct {
	GoogleOauth2ClientOidcIssuer string
	GoogleApiKey                 string
	GoogleOauth2ClientID         string
	GoogleOauth2ClientSecret     string
	GoogleOauth2RedirectURI      string
	GoogleOauth2Scope            string
	CreateUserIfNotExists        bool
}

// FacebookConstants holds configuration for Facebook OAuth
type FacebookConstants struct {
	AppID                  string
	AppSecret              string
	CallbackURL            string
	Scope                  string
	ProfileFields          []string
	CreateUserIfNotExists  bool
}