package model

import (
	"context"
	"fmt"

	"github.com/mgnsk/calendar/internal/domain"
	"github.com/mgnsk/calendar/internal/pkg/snowflake"
	"github.com/mgnsk/calendar/internal/pkg/sqlite"
	"github.com/mgnsk/calendar/internal/pkg/timestamp"
	"github.com/samber/lo"
	"github.com/uptrace/bun"
)

// Event is the event database model.
type Event struct {
	ID             snowflake.ID        `bun:"id,pk"`
	StartAtUnix    int64               `bun:"start_at_unix"`
	EndAtUnix      int64               `bun:"end_at_unix"`
	StartAtRFC3339 timestamp.Timestamp `bun:"start_at_rfc3339"`
	EndAtRFC3339   timestamp.Timestamp `bun:"end_at_rfc3339"`
	Title          string              `bun:"title"`
	Description    string              `bun:"description"`
	URL            string              `bun:"url"`
	Tags           []*Tag              `bun:"m2m:events_tags,join:Event=Tag"`

	bun.BaseModel `bun:"events"`
}

type eventToTag struct {
	TagID   snowflake.ID `bun:"tag_id"`
	Tag     *Tag         `bun:"rel:belongs-to,join:tag_id=id"`
	EventID snowflake.ID `bun:"event_id"`
	Event   *Event       `bun:"rel:belongs-to,join:event_id=id"`

	bun.BaseModel `bun:"events_tags"`
}

// InsertEvent inserts an event to the database.
func InsertEvent(ctx context.Context, db *bun.DB, ev *domain.Event) error {
	return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if err := sqlite.WithErrorChecking(tx.NewInsert().Model(&Event{
			ID:             ev.ID,
			StartAtUnix:    ev.StartAt.Time().Unix(),
			EndAtUnix:      ev.EndAt.Time().Unix(),
			StartAtRFC3339: ev.StartAt,
			EndAtRFC3339:   ev.EndAt,
			Title:          ev.Title,
			Description:    ev.Description,
			URL:            ev.URL,
			Tags:           nil,
		}).Exec(ctx)); err != nil {
			return err
		}

		if len(ev.Tags) == 0 {
			return nil
		}

		// Ensure tags exist.
		for _, name := range ev.Tags {
			if err := InsertTag(ctx, tx, name); err != nil {
				return err
			}
		}

		// Get tags.
		tags := make([]*Tag, 0, len(ev.Tags))
		for _, name := range ev.Tags {
			tag, err := GetTag(ctx, tx, name)
			if err != nil {
				return err
			}
			tags = append(tags, tag)
		}

		// Insert relations.
		for _, tag := range tags {
			if err := sqlite.WithErrorChecking(tx.NewInsert().Model(&eventToTag{
				TagID:   tag.ID,
				EventID: ev.ID,
			}).Exec(ctx)); err != nil {
				return err
			}
		}

		return nil
	})
}

// ListEvents lists events.
func ListEvents(ctx context.Context, db *bun.DB, order string, filterTags ...string) ([]*domain.Event, error) {
	model := []*Event{}

	q := db.NewSelect().Model(&model).
		Relation("Tags", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("tag.name ASC")
		})

	switch order {
	case "asc":
		q = q.Order("start_at_unix ASC")
	case "desc":
		q = q.Order("start_at_unix DESC")
	default:
		panic(fmt.Sprintf("invalid order %s, expected asc or desc", order))
	}

	if len(filterTags) > 0 {
		q = q.Join("LEFT JOIN events_tags ON event.id = events_tags.event_id").
			Join("LEFT JOIN tags ON tags.id = events_tags.tag_id").
			Where("tags.name IN (?)", bun.In(filterTags))
	}

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return lo.Map(model, func(ev *Event, _ int) *domain.Event {
		return &domain.Event{
			ID:          snowflake.ID(ev.ID),
			StartAt:     ev.StartAtRFC3339,
			EndAt:       ev.EndAtRFC3339,
			Title:       ev.Title,
			Description: ev.Description,
			URL:         ev.URL,
			Tags: lo.Map(ev.Tags, func(tag *Tag, _ int) string {
				return tag.Name
			}),
		}
	}), nil
}
