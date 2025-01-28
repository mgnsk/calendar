package model

import (
	"context"
	"time"

	"github.com/mgnsk/calendar/internal/pkg/snowflake"
	"github.com/mgnsk/calendar/internal/pkg/sqlite"
	"github.com/uptrace/bun"
)

// Event is a model for an event.
type Event struct {
	ID               int64  `bun:"id,pk"`
	UnixTimestamp    int64  `bun:"unix_timestamp"`
	RFC3339Timestamp string `bun:"rfc3339_timestamp"`
	Title            string `bun:"title"`
	Content          string `bun:"content"`
	Tags             []*Tag `bun:"m2m:events_tags,join:Event=Tag"`

	bun.BaseModel `bun:"events"`
}

type insertEventModel struct {
	ID               int64  `bun:"id"`
	UnixTimestamp    int64  `bun:"unix_timestamp"`
	RFC3339Timestamp string `bun:"rfc3339_timestamp"`
	Title            string `bun:"title"`
	Content          string `bun:"content"`

	bun.BaseModel `bun:"events"`
}

type eventToTag struct {
	TagID   int64  `bun:"tag_id"`
	Tag     *Tag   `bun:"rel:belongs-to,join:tag_id=id"`
	EventID int64  `bun:"event_id"`
	Event   *Event `bun:"rel:belongs-to,join:event_id=id"`

	bun.BaseModel `bun:"events_tags"`
}

// InsertEvent inserts an event to the database.
func InsertEvent(ctx context.Context, db *bun.DB, ts time.Time, title, content string, tags ...*Tag) error {
	eventID := snowflake.Generate()

	if err := sqlite.WithErrorChecking(db.NewInsert().Model(&insertEventModel{
		ID:               eventID.Int64(),
		UnixTimestamp:    ts.Unix(),
		RFC3339Timestamp: ts.Format(time.RFC3339),
		Title:            title,
		Content:          content,
	}).Exec(ctx)); err != nil {
		return err
	}

	// Insert relations.
	for _, tag := range tags {
		if err := sqlite.WithErrorChecking(db.NewInsert().Model(&eventToTag{
			TagID:   tag.ID,
			EventID: eventID.Int64(),
		}).Exec(ctx)); err != nil {
			return err
		}
	}

	return nil
}

// ListEvents lists events.
func ListEvents(ctx context.Context, db *bun.DB, filterTag string, from, to time.Time, limit uint) ([]*Event, error) {
	model := []*Event{}

	// [bun]  22:06:40.037   SELECT                  163µs  SELECT "event_to_tag"."even
	// t_id", "tag"."id", "tag"."name" FROM "tags" AS "tag" JOIN "events_tags" AS "even
	// t_to_tag" ON ("event_to_tag"."event_id") IN (416913968599042, 416913968599041, 4
	// 16913968599040) WHERE ("tag"."id" = "event_to_tag"."tag_id") AND (tag.id IS NOT
	// NULL) AND (tag.name = 'tag1') ORDER BY tag.name ASC
	q := db.NewSelect().Model(&model).
		Relation("Tags", func(q *bun.SelectQuery) *bun.SelectQuery {
			if filterTag != "" {
				q = q.Where("tag.name = ?", filterTag)
			}

			return q.OrderExpr("tag.name ASC")
		}).
		Order("unix_timestamp DESC").
		Limit(int(limit))

	if !from.IsZero() {
		q = q.Where("unix_timestamp > ?", from.Unix())
	}

	if !to.IsZero() {
		q = q.Where("unix_timestamp < ?", to.Unix())
	}

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return model, nil
}
