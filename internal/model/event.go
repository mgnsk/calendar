package model

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/mgnsk/calendar/internal/domain"
	"github.com/mgnsk/calendar/internal/pkg/snowflake"
	"github.com/mgnsk/calendar/internal/pkg/sqlite"
	"github.com/mgnsk/calendar/internal/pkg/textfilter"
	"github.com/mgnsk/calendar/internal/pkg/wreck"
	"github.com/samber/lo"
	"github.com/uptrace/bun"
)

// Event is the event database model.
type Event struct {
	ID             snowflake.ID  `bun:"id,pk"`
	StartAtUnix    int64         `bun:"start_at_unix"`
	EndAtUnix      sql.NullInt64 `bun:"end_at_unix"`
	TimezoneOffset int           `bun:"tz_offset"`
	Title          string        `bun:"title"`
	Description    string        `bun:"description"`
	URL            string        `bun:"url"`
	Tags           []*Tag        `bun:"m2m:events_tags,join:Event=Tag"`
	FTSData        string        `bun:"fts_data"`

	bun.BaseModel `bun:"events"`
}

type eventToTag struct {
	TagID   snowflake.ID `bun:"tag_id"`
	Tag     *Tag         `bun:"rel:belongs-to,join:tag_id=id"`
	EventID snowflake.ID `bun:"event_id"`
	Event   *Event       `bun:"rel:belongs-to,join:event_id=id"`

	bun.BaseModel `bun:"events_tags"`
}

type eventFTS struct {
	ID   snowflake.ID `bun:"rowid"`
	Data string       `bun:"fts_data"`

	bun.BaseModel `bun:"events_fts_idx"`
}

// InsertEvent inserts an event to the database.
func InsertEvent(ctx context.Context, db *bun.DB, ev *domain.Event) error {
	_, offset := ev.StartAt.Zone()

	return db.RunInTx(ctx, nil, func(ctx context.Context, db bun.Tx) error {
		if err := sqlite.WithErrorChecking(db.NewInsert().Model(&Event{
			ID:          ev.ID,
			StartAtUnix: ev.StartAt.Unix(),
			EndAtUnix: sql.NullInt64{
				Int64: ev.EndAt.Unix(),
				Valid: !ev.EndAt.IsZero(),
			},
			TimezoneOffset: offset,
			Title:          ev.Title,
			Description:    ev.Description,
			URL:            ev.URL,
			Tags:           nil,
			FTSData:        ev.GetFTSData(),
		}).Exec(ctx)); err != nil {
			return err
		}

		tags := ev.GetTags()
		if len(tags) == 0 {
			return nil
		}

		// Ensure tags exist.
		if err := InsertTags(ctx, db, tags...); err != nil {
			return err
		}

		// Create tag relations.
		relations := make([]eventToTag, 0, len(tags))
		tagIDs, err := getTagIDs(ctx, db, tags...)
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

		return increaseEventCounts(ctx, db, tags...)
	})
}

// EventOrder is an event ordering type.
type EventOrder *[]string

// Order values.
var (
	OrderStartAtAsc    = &[]string{"event.start_at_unix ASC", "event.id ASC"}
	OrderStartAtDesc   = &[]string{"event.start_at_unix DESC", "event.id ASC"}
	OrderCreatedAtAsc  = &[]string{"event.id ASC"}
	OrderCreatedAtDesc = &[]string{"event.id DESC"}
)

// EventsQueryBuilder builds an event list query.
type EventsQueryBuilder func(*bun.SelectQuery)

// NewEventsQuery creates a new events list query.
// Note: cursor is ID cursor when sorting by created at
// and offset when sorting by start time.
func NewEventsQuery() EventsQueryBuilder {
	return func(*bun.SelectQuery) {}
}

// WithLimit configures results limit.
func (build EventsQueryBuilder) WithLimit(limit int) EventsQueryBuilder {
	return func(q *bun.SelectQuery) {
		build(q)

		q.Limit(limit)
	}
}

// WithOrder configures results order.
func (build EventsQueryBuilder) WithOrder(cursor int64, orders EventOrder) EventsQueryBuilder {
	return func(q *bun.SelectQuery) {
		build(q)

		switch orders {
		case OrderStartAtAsc:
			q.Order(*orders...)
			if cursor > 0 {
				q.Offset(int(cursor))
			}

		case OrderStartAtDesc:
			q.Order(*orders...)
			if cursor > 0 {
				q.Offset(int(cursor))
			}

		case OrderCreatedAtAsc:
			q.Order(*orders...)
			if cursor > 0 {
				q.Where("event.id > ?", cursor)
			}

		case OrderCreatedAtDesc:
			q.Order(*orders...)
			if cursor > 0 {
				q.Where("event.id < ?", cursor)
			}

		default:
			panic("invalid order")
		}
	}
}

// WithStartAtFrom configures minimum start at time.
func (build EventsQueryBuilder) WithStartAtFrom(from time.Time) EventsQueryBuilder {
	return func(q *bun.SelectQuery) {
		build(q)

		q.Where("event.start_at_unix >= ?", from.Unix())
	}
}

// WithStartAtUntil configures maximum start at time.
func (build EventsQueryBuilder) WithStartAtUntil(until time.Time) EventsQueryBuilder {
	return func(q *bun.SelectQuery) {
		build(q)

		q.Where("event.start_at_unix <= ?", until.Unix())
	}
}

// List executes the query.
func (build EventsQueryBuilder) List(ctx context.Context, db *bun.DB, searchText string) ([]*domain.Event, error) {
	q := db.NewSelect()

	build(q)

	model := []*Event{}

	q.Model(&model).Relation("Tags", func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Order("tag.name ASC")
	})

	var searchResults []*eventFTS

	if searchText != "" {
		searchText = textfilter.Clean(searchText)
		if len(searchText) < 3 {
			return nil, wreck.NotFound.New("No search results were found")
		}

		// Try exact match first and only when non-quoted input.
		if !strings.Contains(searchText, `"`) {
			quoted := textfilter.EnsureQuoted(searchText)
			results, err := searchEvents(ctx, q.DB(), quoted)
			if err != nil {
				return nil, err
			}
			searchResults = results
		}

		// Search again more generally.
		if len(searchResults) == 0 {
			searchText = textfilter.PrepareFTSSearchString(searchText)
			results, err := searchEvents(ctx, q.DB(), searchText)
			if err != nil {
				return nil, err
			}
			searchResults = results
		}

		if len(searchResults) == 0 {
			return nil, wreck.NotFound.New("No search results were found")
		}
	}

	if len(searchResults) > 0 {
		q.Where("event.id IN (?)", bun.In(lo.Map(searchResults, func(r *eventFTS, _ int) snowflake.ID {
			return r.ID
		})))
	}

	if err := q.Scan(ctx); err != nil {
		return nil, sqlite.NormalizeError(err)
	}

	return lo.Map(model, func(ev *Event, _ int) *domain.Event {
		return eventToDomain(ev)
	}), nil
}

func searchEvents(ctx context.Context, db bun.IDB, text string) ([]*eventFTS, error) {
	model := []*eventFTS{}

	if err := sqlite.NormalizeError(db.NewSelect().Model(&model).
		Column("rowid").
		Where("events_fts_idx MATCH ?", text).
		Order("rank").
		Scan(ctx)); err != nil {
		if errors.Is(err, wreck.NotFound) {
			return nil, nil
		}
		return nil, err
	}

	return model, nil
}

func eventToDomain(ev *Event) *domain.Event {
	zone := time.FixedZone("", ev.TimezoneOffset)

	return &domain.Event{
		ID:      snowflake.ID(ev.ID),
		StartAt: time.Unix(ev.StartAtUnix, 0).In(zone),
		EndAt: func() time.Time {
			if ev.EndAtUnix.Valid {
				return time.Unix(ev.EndAtUnix.Int64, 0).In(zone)
			}
			return time.Time{}
		}(),
		Title:       ev.Title,
		Description: ev.Description,
		URL:         ev.URL,
	}
}
