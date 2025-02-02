package model

import (
	"context"
	"errors"

	"github.com/mgnsk/calendar/internal/domain"
	"github.com/mgnsk/calendar/internal/pkg/snowflake"
	"github.com/mgnsk/calendar/internal/pkg/sqlite"
	"github.com/mgnsk/calendar/internal/pkg/wreck"
	"github.com/samber/lo"
	"github.com/uptrace/bun"
)

// Tag is the tag database model.
type Tag struct {
	ID         snowflake.ID `bun:"id,pk"`
	Name       string       `bun:"name"`
	EventCount uint64       `bun:"event_count"`

	bun.BaseModel `bun:"tags"`
}

// InsertTags inserts tags into the database. If a tag exists, it is ignored.
func InsertTags(ctx context.Context, db bun.IDB, names ...string) error {
	model := lo.Map(names, func(name string, _ int) Tag {
		return Tag{
			ID:         snowflake.Generate(),
			Name:       name,
			EventCount: 0,
		}
	})

	if err := sqlite.WithErrorChecking(db.NewInsert().Model(&model).Ignore().Exec(ctx)); err != nil {
		if e := new(wreck.PreconditionFailed); errors.As(err, &e) {
			return nil
		}

		return err
	}

	return nil
}

// ListTags lists tags.
func ListTags(ctx context.Context, db bun.IDB) ([]*domain.Tag, error) {
	model := []*Tag{}

	if err := db.NewSelect().Model(&model).
		Order("tag.name ASC").
		Scan(ctx); err != nil {
		return nil, sqlite.NormalizeError(err)
	}

	return lo.Map(model, func(tag *Tag, _ int) *domain.Tag {
		return &domain.Tag{
			ID:         tag.ID,
			Name:       tag.Name,
			EventCount: tag.EventCount,
		}
	}), nil
}

// increaseEventCounts increases tags' event counts by one.
func increaseEventCounts(ctx context.Context, db bun.IDB, names ...string) error {
	return sqlite.WithErrorChecking(db.NewUpdate().Model((*Tag)(nil)).
		Set("event_count = event_count + 1").
		Where("name IN (?)", bun.In(names)).
		Exec(ctx))
}

// getTagID returns a tag IDs from database.
func getTagIDs(ctx context.Context, db bun.IDB, names ...string) ([]snowflake.ID, error) {
	model := []*Tag{}

	if err := db.NewSelect().Model(&model).
		Where("name IN (?)", bun.In(names)).
		Scan(ctx); err != nil {
		return nil, sqlite.NormalizeError(err)
	}

	return lo.Map(model, func(tag *Tag, _ int) snowflake.ID {
		return tag.ID
	}), nil
}
