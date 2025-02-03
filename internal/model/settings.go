package model

import (
	"context"
	"errors"
	"net/url"

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
	BaseURL       string `bun:"base_url"`
	SessionSecret []byte `bun:"session_secret"`

	bun.BaseModel `bun:"settings"`
}

// InsertOrIgnoreSettings inserts settings or ignores if settings table is already populated.
func InsertOrIgnoreSettings(ctx context.Context, db bun.IDB, s *domain.Settings) error {
	if err := sqlite.WithErrorChecking(db.NewInsert().Ignore().Model(&Settings{
		ID:            1,
		IsInitialized: s.IsInitialized,
		Title:         s.Title,
		Description:   s.Description,
		BaseURL:       s.BaseURL.String(),
		SessionSecret: s.SessionSecret,
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
		BaseURL:       s.BaseURL.String(),
		SessionSecret: s.SessionSecret,
		BaseModel:     bun.BaseModel{},
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

	u, err := url.Parse(model.BaseURL)
	if err != nil {
		return nil, err
	}

	return &domain.Settings{
		IsInitialized: model.IsInitialized,
		Title:         model.Title,
		Description:   model.Description,
		BaseURL:       u,
		SessionSecret: model.SessionSecret,
	}, nil
}
