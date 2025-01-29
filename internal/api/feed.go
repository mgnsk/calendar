package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/feeds"
	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/internal/model"
	"github.com/uptrace/bun"
)

// FeedType is a feed type.
type FeedType string

// Feed type constants.
const (
	RSS  FeedType = "rss"
	Atom FeedType = "atom"
)

// FeedConfig is feed configuration.
type FeedConfig struct {
	Title       string
	Description string
	Link        string
}

// FeedHandler handles feed output.
type FeedHandler struct {
	db     *bun.DB
	config FeedConfig
}

// CreateFeedHandler creates an echo HTTP handler.
func (h *FeedHandler) CreateFeedHandler(target FeedType) echo.HandlerFunc {
	return func(c echo.Context) error {
		feed, err := h.createFeed(c.Request().Context())
		if err != nil {
			return err
		}

		// TODO: consider adding an ETag so that proxies can cache the response.

		c.Response().Header().Set(echo.HeaderContentDisposition, `attachment; filename="feed.xml"`)

		switch target {
		case RSS:
			c.Response().Header().Set(echo.HeaderContentType, "application/rss+xml; charset=utf-8")
			c.Response().WriteHeader(http.StatusOK)
			return feed.WriteRss(c.Response())

		case Atom:
			c.Response().Header().Set(echo.HeaderContentType, "application/atom+xml; charset=utf-8")
			c.Response().WriteHeader(http.StatusOK)
			return feed.WriteAtom(c.Response())

		default:
			panic(fmt.Sprintf("invalid feed type %v", target))
		}
	}
}

func (h *FeedHandler) createFeed(ctx context.Context) (*feeds.Feed, error) {
	events, err := model.ListEvents(ctx, h.db, "asc")
	if err != nil {
		return nil, err
	}

	feed := &feeds.Feed{
		Title:       h.config.Title,
		Link:        &feeds.Link{Href: h.config.Link},
		Description: h.config.Description,
		Image:       nil,
	}

	for _, ev := range events {
		var buf strings.Builder
		buf.WriteString(ev.Description)
		buf.WriteString("\n\n")
		buf.WriteString(fmt.Sprintf("tags: %s", strings.Join(ev.Tags, ", ")))
		buf.WriteString("\n")
		buf.WriteString(fmt.Sprintf("starts at: %s", ev.StartAt.String()))
		if !ev.EndAt.Time().IsZero() {
			buf.WriteString("\n")
			buf.WriteString(fmt.Sprintf("ends at: %s", ev.EndAt.String()))
		}

		feed.Add(&feeds.Item{
			Title:       ev.Title,
			Link:        &feeds.Link{Href: ev.URL},
			Description: buf.String(),
			Id:          ev.ID.String(),
			Updated:     ev.StartAt.Time(),
			Created:     ev.StartAt.Time(),
		})
	}

	return feed, nil
}

// NewFeedHandler creates a new feed handler.
func NewFeedHandler(db *bun.DB, config FeedConfig) *FeedHandler {
	return &FeedHandler{
		db:     db,
		config: config,
	}
}
