package model

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"strconv"
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
	FTSData        string              `bun:"fts_data"`

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
			FTSData:        ev.GetFTSData(),
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
	OrderCreatedAtAsc  = &[]string{"event.id ASC"}
	OrderCreatedAtDesc = &[]string{"event.id DESC"}
)

// ListEvents lists events.
// TODO: verify if we need tag_id idx on events_tags table.
// Note: cursor is ID cursor when sorting by created at
// and offset when sorting by start time.
func ListEvents(
	ctx context.Context,
	db bun.IDB,
	startFrom,
	startUntil time.Time,
	searchText string,
	order EventOrder,
	cursor int64,
	limit int,
	filterTags ...string,
) ([]*domain.Event, error) {
	model := []*Event{}

	q := db.NewSelect().Model(&model).
		Relation("Tags", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("tag.name ASC")
		})

	switch order {
	case OrderStartAtAsc:
		q = q.Order(*order...)
		q = q.Limit(limit)
		if cursor > 0 {
			q = q.Offset(int(cursor))
		}

	case OrderStartAtDesc:
		q = q.Order(*order...)
		q = q.Limit(limit)
		if cursor > 0 {
			q = q.Offset(int(cursor))
		}

	case OrderCreatedAtAsc:
		q = q.Order(*order...)
		q = q.Limit(limit)
		if cursor > 0 {
			q = q.Where("id > ?", cursor)
		}

	case OrderCreatedAtDesc:
		q = q.Order(*order...)
		q = q.Limit(limit)
		if cursor > 0 {
			q = q.Where("id < ?", cursor)
		}

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

	var searchResults []*eventFTS

	if searchText != "" {
		searchText = cleanString(searchText)
		if len(searchText) < 3 {
			return nil, wreck.NotFound.New("No search results were found")
		}

		// Try exact match first and only when non-quoted input.
		if !strings.Contains(searchText, `"`) {
			quoted := ensureQuoted(searchText)
			results, err := searchEvents(ctx, db, quoted)
			if err != nil {
				return nil, err
			}
			searchResults = results
		}

		// Search again more generally.
		if len(searchResults) == 0 {
			searchText = prepareGeneralSearchString(searchText)
			results, err := searchEvents(ctx, db, searchText)
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
		q = q.Where("event.id IN (?)", bun.In(lo.Map(searchResults, func(r *eventFTS, _ int) snowflake.ID {
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

func ensureQuoted(s string) string {
	if unquoted, err := strconv.Unquote(s); err == nil {
		return strconv.Quote(unquoted)
	}
	return strconv.Quote(s)
}

func prepareGeneralSearchString(s string) string {
	fields := splitString(s)
	quoted := make([]string, 0, len(fields))

	for _, field := range fields {
		quoted = append(quoted, ensureQuoted(field))
	}

	s = strings.Join(quoted, " ")

	return s
}

func cleanString(s string) string {
	return strings.Map(func(r rune) rune {
		if r == unicode.ReplacementChar {
			return -1
		}
		if !unicode.IsPrint(r) {
			return -1
		}
		return r
	}, s)
}

// splitString splits a string by whitespace while
// attempting to keep the most common bases of quote usage.
func splitString(s string) []string {
	quoted := false
	return strings.FieldsFunc(s, func(r rune) bool {
		if unicode.In(r, unicode.Quotation_Mark) {
			quoted = !quoted
		}
		return !quoted && unicode.IsSpace(r)
	})
}
