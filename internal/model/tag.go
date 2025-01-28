package model

import (
	"context"
	"errors"

	"github.com/mgnsk/calendar/internal/pkg/snowflake"
	"github.com/mgnsk/calendar/internal/pkg/sqlite"
	"github.com/mgnsk/calendar/internal/pkg/wreck"
	"github.com/uptrace/bun"
)

// Tag is a database model for a tag.
type Tag struct {
	ID   int64  `bun:"id,pk"`
	Name string `bun:"name"`

	bun.BaseModel `bun:"tags"`
}

// InsertTag inserts a tag into the database. If the tag exists, it is ignored.
func InsertTag(ctx context.Context, db *bun.DB, name string) error {
	if err := sqlite.WithErrorChecking(
		db.NewInsert().Model(&Tag{
			ID:   snowflake.Generate().Int64(),
			Name: name,
		}).Ignore().Exec(ctx),
	); err != nil {
		if e := new(wreck.PreconditionFailed); errors.As(err, &e) {
			return nil
		}

		return err
	}

	return nil
}

// ListTags lists tags.
func ListTags(ctx context.Context, db *bun.DB) ([]*Tag, error) {
	model := []*Tag{}

	q := db.NewSelect().Model(&model).
		Order("name ASC")

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return model, nil
}
