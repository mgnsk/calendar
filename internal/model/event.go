package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/mgnsk/calendar/internal/domain"
	"github.com/mgnsk/calendar/internal/pkg/snowflake"
	"github.com/mgnsk/calendar/internal/pkg/sqlite"
	"github.com/mgnsk/calendar/internal/pkg/timestamp"
	"github.com/mgnsk/calendar/internal/pkg/wreck"
	"github.com/samber/lo"
	"github.com/uptrace/bun"
)

// Event is the event database model.
type Event struct {
	ID             snowflake.ID        `bun:"id,pk"`
	StartAtUnix    int64               `bun:"start_at_unix"`
	EndAtUnix      sql.NullInt64       `bun:"end_at_unix"`
	StartAtRFC3339 timestamp.Timestamp `bun:"start_at_rfc3339"`
	EndAtRFC3339   timestamp.Timestamp `bun:"end_at_rfc3339"`
	Title          string              `bun:"title"`
	Description    string              `bun:"description"`
	URL            string              `bun:"url"`
	Tags           []*Tag              `bun:"m2m:events_tags,join:Event=Tag"`

	bun.BaseModel `bun:"events"`
}

type eventFTS struct {
	ID          snowflake.ID `bun:"id"`
	Title       string       `bun:"title"`
	Description string       `bun:"description"`
	URL         string       `bun:"url"`
	Tags        string       `bun:"tags"`

	bun.BaseModel `bun:"events_fts"`
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
	return db.RunInTx(ctx, nil, func(ctx context.Context, db bun.Tx) error {
		if err := sqlite.WithErrorChecking(db.NewInsert().Model(&Event{
			ID:          ev.ID,
			StartAtUnix: ev.StartAt.Time().Unix(),
			EndAtUnix: sql.NullInt64{
				Int64: ev.EndAt.Time().Unix(),
				Valid: !ev.EndAt.Time().IsZero(),
			},
			StartAtRFC3339: ev.StartAt,
			EndAtRFC3339:   ev.EndAt,
			Title:          ev.Title,
			Description:    ev.Description,
			URL:            ev.URL,
			Tags:           nil,
		}).Exec(ctx)); err != nil {
			return err
		}

		if err := sqlite.WithErrorChecking(db.NewInsert().Model(&eventFTS{
			ID:          ev.ID,
			Title:       ev.Title,
			Description: ev.Description,
			URL:         ev.URL,
			Tags:        strings.Join(ev.Tags, " "),
		}).Exec(ctx)); err != nil {
			return err
		}

		if len(ev.Tags) == 0 {
			return nil
		}

		// Ensure tags exist.
		if err := InsertTags(ctx, db, ev.Tags...); err != nil {
			return err
		}

		// Create tag relations.
		relations := make([]eventToTag, 0, len(ev.Tags))
		tagIDs, err := getTagIDs(ctx, db, ev.Tags...)
		if err != nil {
			return err
		}

		for _, tagID := range tagIDs {
			relations = append(relations, eventToTag{
				TagID:   tagID,
				EventID: ev.ID,
			})
		}

		if err := sqlite.WithErrorChecking(db.NewInsert().Model(&relations).Exec(ctx)); err != nil {
			return err
		}

		return increaseEventCounts(ctx, db, ev.Tags...)
	})
}

// EventOrder is an event ordering type.
type EventOrder *[]string

// Order values.
var (
	OrderStartAtAsc    = &[]string{"event.start_at_unix ASC", "event.id ASC"}
	OrderStartAtDesc   = &[]string{"event.start_at_unix DESC", "event.id ASC"}
	OrderCreatedAtDesc = &[]string{"event.id DESC"}
)

// SearchEvents performs full text search on events.
// TODO: https://www.sqlite.org/fts5.html#the_highlight_function
// TODO: bm25 and columnsize
// TODO: to keep search result order and custom filtering, we fetch results one by one.
func SearchEvents(ctx context.Context, db bun.IDB, searchText string, startFrom, startUntil time.Time, filterTags ...string) ([]*domain.Event, error) {
	searchText = strip(searchText)

	if searchText == "" {
		return nil, &wreck.NotFound{Err: fmt.Errorf("no results found")}
	}

	model := []*eventFTS{}

	// if err := db.NewSelect().
	// 	TableExpr("events_fts(?)", searchText).
	// 	Order("rank").
	// 	Scan(ctx, &model); err != nil {
	// 	return nil, sqlite.NormalizeError(err)
	// }

	if err := db.NewSelect().Model(&model).
		Where("events_fts MATCH ?", searchText).
		Order("rank").
		Scan(ctx); err != nil {
		return nil, sqlite.NormalizeError(err)
	}

	if len(model) == 0 {
		return nil, &wreck.NotFound{Err: fmt.Errorf("no results found")}
	}

	var results []*domain.Event

	for _, ftsResult := range model {
		ev, err := getEvent(ctx, db, ftsResult.ID, startFrom, startUntil, filterTags...)
		if err != nil {
			if e := new(wreck.NotFound); errors.As(err, &e) {
				continue
			}

			return nil, err
		}

		results = append(results, ev)
	}

	if len(results) == 0 {
		return nil, &wreck.NotFound{Err: fmt.Errorf("no results found")}
	}

	return results, nil
}

// ListEvents lists events.
// TODO: verify if we need tag_id idx on events_tags table.
func ListEvents(ctx context.Context, db bun.IDB, startFrom, startUntil time.Time, order EventOrder, filterTags ...string) ([]*domain.Event, error) {
	model := []*Event{}

	q := db.NewSelect().Model(&model).
		Relation("Tags", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("tag.name ASC")
		})

	switch order {
	case OrderStartAtAsc:
		q = q.Order(*order...)
	case OrderStartAtDesc:
		q = q.Order(*order...)
	case OrderCreatedAtDesc:
		q = q.Order(*order...)
	default:
		panic("invalid order")
	}

	if !startFrom.IsZero() {
		q = q.Where("event.start_at_unix >= ?", startFrom.Unix())
	}

	if !startUntil.IsZero() {
		q = q.Where("event.start_at_unix <= ?", startUntil.Unix())
	}

	filterTags = slices.DeleteFunc(filterTags, func(tag string) bool {
		return tag == ""
	})

	if len(filterTags) > 0 {
		q = q.Join("LEFT JOIN events_tags ON event.id = events_tags.event_id").
			Join("LEFT JOIN tags ON tags.id = events_tags.tag_id").
			Where("tags.name IN (?)", bun.In(filterTags))
	}

	if err := q.Scan(ctx); err != nil {
		return nil, sqlite.NormalizeError(err)
	}

	return lo.Map(model, func(ev *Event, _ int) *domain.Event {
		return eventToDomain(ev)
	}), nil
}

func getEvent(ctx context.Context, db bun.IDB, id snowflake.ID, startFrom, startUntil time.Time, filterTags ...string) (*domain.Event, error) {
	model := &Event{}

	q := db.NewSelect().Model(model).
		Where("event.id = ?", id).
		Relation("Tags", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("tag.name ASC")
		})

	if !startFrom.IsZero() {
		q = q.Where("event.start_at_unix >= ?", startFrom.Unix())
	}

	if !startUntil.IsZero() {
		q = q.Where("event.start_at_unix <= ?", startUntil.Unix())
	}

	filterTags = slices.DeleteFunc(filterTags, func(tag string) bool {
		return tag == ""
	})

	if len(filterTags) > 0 {
		q = q.Join("LEFT JOIN events_tags ON event.id = events_tags.event_id").
			Join("LEFT JOIN tags ON tags.id = events_tags.tag_id").
			Where("tags.name IN (?)", bun.In(filterTags))
	}

	if err := q.Scan(ctx); err != nil {
		return nil, sqlite.NormalizeError(err)
	}

	return eventToDomain(model), nil
}

func eventToDomain(ev *Event) *domain.Event {
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
		TagRelations: lo.Map(ev.Tags, func(tag *Tag, _ int) *domain.Tag {
			return &domain.Tag{
				ID:         tag.ID,
				Name:       tag.Name,
				EventCount: tag.EventCount,
			}
		}),
	}
}

func strip(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) {
			return r
		}
		if unicode.IsNumber(r) {
			return r
		}
		if unicode.IsSpace(r) {
			return r
		}
		return -1
	}, str)
}
