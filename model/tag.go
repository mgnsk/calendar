package model

import (
	"context"
	"errors"

	"github.com/mgnsk/calendar"
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/pkg/snowflake"
	"github.com/mgnsk/calendar/pkg/sqlite"
	"github.com/samber/lo"
	"github.com/uptrace/bun"
)

// Tag is the tag database model.
type Tag struct {
	ID         snowflake.ID `bun:"id,pk"`
	Name       string       `bun:"name"`
	EventCount uint64       `bun:"event_count,scanonly"`

	bun.BaseModel `bun:"tags"`
}

// InsertTags inserts tags into the database. If a tag exists, it is ignored.
func InsertTags(ctx context.Context, db bun.IDB, names ...string) error {
	model := lo.Map(names, func(name string, _ int) Tag {
		return Tag{
			ID:   snowflake.Generate(),
			Name: name,
		}
	})

	if err := sqlite.WithErrorChecking(db.NewInsert().Model(&model).Ignore().Exec(ctx)); err != nil {
		if errors.Is(err, calendar.PreconditionFailed) {
			return nil
		}
		return err
	}

	return nil
}

// ListTags lists most popular tags, excluding stopwords.
func ListTags(ctx context.Context, db bun.IDB, limit int) ([]*domain.Tag, error) {
	model := []*Tag{}

	if err := db.NewSelect().Model(&model).
		ColumnExpr("tag.id, tag.name, COUNT(et.event_id) AS event_count").
		Join("LEFT JOIN events_tags AS et ON et.tag_id = tag.id").
		Join("LEFT JOIN stopwords AS sw ON sw.word = tag.name COLLATE NOCASE").
		Where("sw.id IS NULL").
		Group("tag.id").
		Order("event_count DESC", "name ASC").
		Limit(limit).
		Scan(ctx); err != nil {
		return nil, sqlite.NormalizeError(err)
	}

	return lo.Map(model, func(tag *Tag, _ int) *domain.Tag {
		return &domain.Tag{
			Name:       tag.Name,
			EventCount: tag.EventCount,
		}
	}), nil
}

// CleanTags deletes tags which have no event relations.
func CleanTags(ctx context.Context, db bun.IDB) error {
	subQuery := db.NewSelect().TableExpr("tags AS tag").
		ColumnExpr("tag.id").
		Join("LEFT JOIN events_tags AS et ON et.tag_id = tag.id").
		Having("COUNT(et.event_id) = 0").
		Group("tag.id")

	if err := sqlite.WithErrorChecking(
		db.NewDelete().Model((*Tag)(nil)).
			Where("id IN (?)", subQuery).
			Exec(ctx),
	); err != nil {
		if errors.Is(err, calendar.PreconditionFailed) {
			return nil
		}

		return err
	}

	return nil
}

// DeleteTags deletes event's tags.
func DeleteTags(ctx context.Context, db bun.IDB, eventID snowflake.ID) error {
	// Delete old tag relations.
	if err := sqlite.WithErrorChecking(
		db.NewDelete().Model((*eventToTag)(nil)).
			Where("event_id = ?", eventID).
			Exec(ctx),
	); err != nil {
		if errors.Is(err, calendar.PreconditionFailed) {
			return nil
		}

		return err
	}

	return nil
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
