package api

import (
	"fmt"
	"net/http"

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

// FeedHandler handles feed output.
type FeedHandler struct {
	db     *bun.DB
	config FeedConfig
}

// CreateFeedHandler creates an echo HTTP handler.
func (h *FeedHandler) CreateFeedHandler(target FeedType) echo.HandlerFunc {
	return func(c echo.Context) error {
		events, err := model.ListEvents(c.Request().Context(), h.db, "asc")
		if err != nil {
			return err
		}

		feed := &feeds.Feed{
			Title: h.config.Title,
			Link:  &feeds.Link{Href: h.config.Link},
			Image: nil,
		}

		for _, ev := range events {
			feed.Add(&feeds.Item{
				Title:       ev.Title,
				Link:        &feeds.Link{Href: ev.URL},
				Description: ev.GetDescription(),
				Id:          ev.ID.String(),
				Updated:     ev.StartAt.Time(),
				Created:     ev.StartAt.Time(),
			})
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

// NewFeedHandler creates a new feed handler.
func NewFeedHandler(db *bun.DB, config FeedConfig) *FeedHandler {
	return &FeedHandler{
		db:     db,
		config: config,
	}
}
