package create

import (
	"context"

	"github.com/Skorpsrgvch/reviewer-service/internal/domain"
)

// TeamSaver сохраняет команду (включая пользователей и связи).
type TeamSaver interface {
	SaveTeam(ctx context.Context, team *domain.Team) error
}
