package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/pkg/snowflake"
	"github.com/mgnsk/calendar/pkg/sqlite"
	"github.com/mgnsk/calendar/pkg/textfilter"
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
	IsDraft        bool          `bun:"is_draft"`
	Tags           []*Tag        `bun:"m2m:events_tags,join:Event=Tag"`

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
			IsDraft:        ev.IsDraft,
			Tags:           nil,
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

		return increaseEventCounts(ctx, db, tagIDs...)
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
func (build EventsQueryBuilder) List(ctx context.Context, db *bun.DB, includeDrafts bool, searchText string) ([]*domain.Event, error) {
	q := db.NewSelect()

	build(q)

	model := []*Event{}

	q.Model(&model)

	if !includeDrafts {
		q.Where("event.is_draft = 0")
	}

	// q.Relation("Tags", func(q *bun.SelectQuery) *bun.SelectQuery {
	// 	return q.Order("tag.name ASC")
	// })

	// Note: for search, we do not order by `rank` column from fts table.
	// The original order of the events query is used.
	if searchText != "" {
		searchText = strings.TrimSpace(searchText)
		if len(searchText) < 3 {
			return []*domain.Event{}, nil
		}

		var (
			exact   string
			general []string
		)

		if !strings.Contains(searchText, `"`) {
			exact = textfilter.EnsureQuoted(searchText)
		}
		general = textfilter.PrepareFTSSearchStrings(searchText)

		var searchWord string
		if exact == "" {
			// Only general search.
			searchWord = strings.Join(general, " ")
		} else if len(general) == 0 || len(general) == 1 && exact == general[0] {
			// Only exact search.
			searchWord = exact
		} else {
			// Both exact and general.
			searchWord = fmt.Sprintf("(%s) OR (%s)", exact, strings.Join(general, " "))
		}

		ftsQuery := db.NewSelect().
			ColumnExpr("rowid").
			Table("events_fts_idx").
			Where("events_fts_idx MATCH ?", searchWord)

		q.Join("JOIN (?) AS exact_result ON exact_result.rowid = event.id", ftsQuery)
	}

	if err := q.Scan(ctx); err != nil {
		return nil, sqlite.NormalizeError(err)
	}

	return lo.Map(model, func(ev *Event, _ int) *domain.Event {
		return eventToDomain(ev)
	}), nil
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
		IsDraft:     ev.IsDraft,
	}
}
