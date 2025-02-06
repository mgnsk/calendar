package model

import (
	"context"
	"errors"

	"github.com/mgnsk/calendar/internal/domain"
	"github.com/mgnsk/calendar/internal/pkg/sqlite"
	"github.com/mgnsk/calendar/internal/pkg/wreck"
	"github.com/uptrace/bun"
)

// Settings is the settings database model.
type Settings struct {
	ID            int64  `bun:"id"`
	IsInitialized bool   `bun:"is_initialized"`
	Title         string `bun:"title"`
	Description   string `bun:"description"`

	bun.BaseModel `bun:"settings"`
}

// InsertOrIgnoreSettings inserts settings or ignores if settings table is already populated.
func InsertOrIgnoreSettings(ctx context.Context, db bun.IDB, s *domain.Settings) error {
	if err := sqlite.WithErrorChecking(db.NewInsert().Ignore().Model(&Settings{
		ID:            1,
		IsInitialized: s.IsInitialized,
		Title:         s.Title,
		Description:   s.Description,
	}).Exec(ctx)); err != nil {
		if errors.Is(err, wreck.PreconditionFailed) {
			return nil
		}
		return err
	}

	return nil
}

// UpdateSettings updates settings.
func UpdateSettings(ctx context.Context, db bun.IDB, s *domain.Settings) error {
	return sqlite.WithErrorChecking(db.NewUpdate().Model(&Settings{
		ID:            1,
		IsInitialized: s.IsInitialized,
		Title:         s.Title,
		Description:   s.Description,
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
		IsInitialized: model.IsInitialized,
		Title:         model.Title,
		Description:   model.Description,
	}, nil
}
