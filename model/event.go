package model

import (
	"context"
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
	ID             snowflake.ID `bun:"id,pk"`
	StartAtUnix    int64        `bun:"start_at_unix"`
	TimezoneOffset int          `bun:"tz_offset"`
	Title          string       `bun:"title"`
	Description    string       `bun:"description"`
	URL            string       `bun:"url"`
	Location       string       `bun:"location"`
	Latitude       float64      `bun:"latitude"`
	Longitude      float64      `bun:"longitude"`

	IsDraft bool         `bun:"is_draft"`
	UserID  snowflake.ID `bun:"user_id"`

	bun.BaseModel `bun:"events"`
}

type eventToTag struct {
	TagID   snowflake.ID `bun:"tag_id"`
	EventID snowflake.ID `bun:"event_id"`

	bun.BaseModel `bun:"events_tags"`
}

// GetEvent retrieves a single event.
func GetEvent(ctx context.Context, db *bun.DB, id snowflake.ID) (*domain.Event, error) {
	model := &Event{}

	if err := db.NewSelect().Model(model).
		Where("id = ?", id).
		Scan(ctx); err != nil {
		return nil, sqlite.NormalizeError(err)
	}

	return eventToDomain(model), nil
}

// InsertEvent inserts an event to the database.
func InsertEvent(ctx context.Context, db *bun.DB, ev *domain.Event) error {
	_, offset := ev.StartAt.Zone()

	return db.RunInTx(ctx, nil, func(ctx context.Context, db bun.Tx) error {
		if err := sqlite.WithErrorChecking(db.NewInsert().Model(&Event{
			ID:             ev.ID,
			StartAtUnix:    ev.StartAt.Unix(),
			TimezoneOffset: offset,
			Title:          ev.Title,
			Description:    ev.Description,
			URL:            ev.URL,
			Location:       ev.Location,
			Latitude:       ev.Latitude,
			Longitude:      ev.Longitude,
			IsDraft:        ev.IsDraft,
			UserID:         ev.UserID,
		}).Exec(ctx)); err != nil {
			return err
		}

		return createEventTagRelations(ctx, db, ev)
	})
}

// UpdateEvent updates an event.
func UpdateEvent(ctx context.Context, db *bun.DB, ev *domain.Event) error {
	_, offset := ev.StartAt.Zone()

	return db.RunInTx(ctx, nil, func(ctx context.Context, db bun.Tx) error {
		if err := sqlite.WithErrorChecking(
			db.NewUpdate().Model(&Event{
				StartAtUnix:    ev.StartAt.Unix(),
				TimezoneOffset: offset,
				Title:          ev.Title,
				Description:    ev.Description,
				URL:            ev.URL,
				Location:       ev.Location,
				Latitude:       ev.Latitude,
				Longitude:      ev.Longitude,
				IsDraft:        ev.IsDraft,
			}).
				Column(
					"start_at_unix",
					"tz_offset",
					"title",
					"description",
					"url",
					"location",
					"latitude",
					"longitude",
					"is_draft",
				).
				Where("id = ?", ev.ID).
				Exec(ctx),
		); err != nil {
			return err
		}

		// Delete old tag relations.
		if err := sqlite.WithErrorChecking(
			db.NewDelete().Model((*eventToTag)(nil)).
				Where("event_id = ?", ev.ID).
				Exec(ctx),
		); err != nil {
			return err
		}

		// Clean up orphaned tags.
		if err := CleanTags(ctx, db); err != nil {
			return err
		}

		// Recreate tag relations.
		return createEventTagRelations(ctx, db, ev)
	})
}

// DeleteEvent deletes an event..
func DeleteEvent(ctx context.Context, db *bun.DB, ev *domain.Event) error {
	return db.RunInTx(ctx, nil, func(ctx context.Context, db bun.Tx) error {
		if err := sqlite.WithErrorChecking(
			db.NewDelete().Model((*Event)(nil)).
				Where("id = ?", ev.ID).
				Exec(ctx),
		); err != nil {
			return err
		}

		// Delete tag relations.
		if err := sqlite.WithErrorChecking(
			db.NewDelete().Model((*eventToTag)(nil)).
				Where("event_id = ?", ev.ID).
				Exec(ctx),
		); err != nil {
			return err
		}

		// Clean up orphaned tags.
		return CleanTags(ctx, db)
	})
}

func createEventTagRelations(ctx context.Context, db bun.IDB, ev *domain.Event) error {
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

	return sqlite.WithErrorChecking(db.NewInsert().Model(&relations).Exec(ctx))
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

// WithUserID filters the event list by user ID.
func (build EventsQueryBuilder) WithUserID(userID snowflake.ID) EventsQueryBuilder {
	return func(q *bun.SelectQuery) {
		build(q)

		q.Where("event.user_id = ?", userID)
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
		ID:          snowflake.ID(ev.ID),
		StartAt:     time.Unix(ev.StartAtUnix, 0).In(zone),
		Title:       ev.Title,
		Description: ev.Description,
		URL:         ev.URL,
		Location:    ev.Location,
		Latitude:    ev.Latitude,
		Longitude:   ev.Longitude,
		IsDraft:     ev.IsDraft,
		UserID:      ev.UserID,
	}
}
