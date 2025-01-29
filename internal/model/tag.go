package model

import (
	"context"
	"errors"
	"fmt"

	"github.com/mgnsk/calendar/internal/pkg/snowflake"
	"github.com/mgnsk/calendar/internal/pkg/sqlite"
	"github.com/mgnsk/calendar/internal/pkg/wreck"
	"github.com/uptrace/bun"
)

// Tag is the tag database model.
type Tag struct {
	ID   snowflake.ID `bun:"id,pk"`
	Name string       `bun:"name"`

	bun.BaseModel `bun:"tags"`
}

// InsertTag inserts a tag into the database. If the tag exists, it is ignored.
func InsertTag(ctx context.Context, db bun.IDB, name string) error {
	if err := sqlite.WithErrorChecking(
		db.NewInsert().Model(&Tag{
			ID:   snowflake.Generate(),
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

// GetTag returns a tag from database.
func GetTag(ctx context.Context, db bun.IDB, name string) (*Tag, error) {
	model := &Tag{}

	if err := db.NewSelect().Model(model).
		Where("name = ?", name).
		Scan(ctx); err != nil {
		return nil, err
	}

	return model, nil
}

// ListTags lists tags.
func ListTags(ctx context.Context, db bun.IDB, filterName string) ([]*Tag, error) {
	model := []*Tag{}

	q := db.NewSelect().Model(&model).
		Order("name ASC")

	if filterName != "" {
		q = q.Where("name LIKE ?", fmt.Sprintf("%%%s%%", filterName))
	}

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return model, nil
}
