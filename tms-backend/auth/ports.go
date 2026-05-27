package auth

import (
	"context"

	authdtos "github.com/auditrakkr/tms-fullstack/tms-backend/auth/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
)

// UserLookup is the minimal interface that AuthService needs from the user
// domain. Defined near the consumer (auth package) so AuthService does not
// import the concrete users package — the concrete *users.UserService satisfies
// this interface implicitly.
type UserLookup interface {
	FindByPrimaryEmailAddress(ctx context.Context, email string) (*models.User, error)
	SetRefreshTokenHash(ctx context.Context, userID uint, hash string) error
	FindByGoogleID(ctx context.Context, googleID string) (*models.User, error)
	FindByFacebookID(ctx context.Context, facebookID string) (*models.User, error)
	// OAuth upsert helpers — encapsulate find-or-create in the users package.
	FindOrCreateGoogleUser(ctx context.Context, profile *authdtos.GoogleProfileDto) (*models.User, error)
	FindOrCreateFacebookUser(ctx context.Context, profile *authdtos.FacebookProfileDto) (*models.User, error)
}
