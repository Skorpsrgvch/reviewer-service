package setActive

import (
	"context"

	"github.com/Skorpsrgvch/reviewer-service/internal/domain"
)

type UserUpdater interface {
	UpdateUser(ctx context.Context, user *domain.User) error
}

type UserFinder interface {
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	GetTeamByUser(ctx context.Context, userID string) (string, error)
}
