package dtos

// GoogleProfileDto represents a Google OIDC profile
type GoogleProfileDto struct {
	GoogleID      *string `json:"googleId,omitempty"`  // This is equivalent of sub
	GivenName     *string `json:"given_name,omitempty"`
	FamilyName    *string `json:"family_name,omitempty"`
	Name          *string `json:"name,omitempty"`
	Email         *string `json:"email,omitempty"`
	EmailVerified *bool   `json:"email_verified,omitempty"`
	Gender        *string `json:"gender,omitempty"`
	Birthdate     *struct {
		Month int `json:"month"`
		Day   int `json:"day"`
		Year  *int `json:"year"`
	} `json:"birthdate,omitempty"`
	Picture      *string `json:"picture,omitempty"`
	Profile      *string `json:"profile,omitempty"`
	AccessToken  *string `json:"access_token,omitempty"`
	RefreshToken *string `json:"refresh_token,omitempty"`
	Exp          *int64  `json:"exp,omitempty"`
	HD           *string `json:"hd,omitempty"`
}