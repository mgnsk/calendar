package model

import (
	"context"

	"github.com/mgnsk/calendar/internal/domain"
	"github.com/mgnsk/calendar/internal/pkg/sqlite"
	"github.com/uptrace/bun"
)

// Settings is the settings database model.
type Settings struct {
	ID          int64  `bun:"id"`
	Title       string `bun:"title"`
	Description string `bun:"description"`

	bun.BaseModel `bun:"settings"`
}

// InsertSettings inserts settings.
func InsertSettings(ctx context.Context, db bun.IDB, s *domain.Settings) error {
	return sqlite.WithErrorChecking(db.NewInsert().Model(&Settings{
		ID:          1,
		Title:       s.Title,
		Description: s.Description,
	}).Exec(ctx))
}

// UpdateSettings updates settings.
func UpdateSettings(ctx context.Context, db bun.IDB, s *domain.Settings) error {
	return sqlite.WithErrorChecking(db.NewUpdate().Model(&Settings{
		ID:          1,
		Title:       s.Title,
		Description: s.Description,
	}).Where("id = 1").Exec(ctx))
}

// GetSettings returns settings.
func GetSettings(ctx context.Context, db bun.IDB) (*domain.Settings, error) {
	model := &Settings{}

	if err := db.NewSelect().Model(model).
		Where("id = 1").
		Scan(ctx); err != nil {
		return nil, sqlite.NormalizeError(err)
	}

	return &domain.Settings{
		Title:       model.Title,
		Description: model.Description,
	}, nil
}
