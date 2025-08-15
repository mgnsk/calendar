package handler

import (
	"bufio"
	"net/http"
	"strings"

	"github.com/alexedwards/scs/v2"
	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar"
	"github.com/mgnsk/calendar/contract"
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/html"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/server"
	"github.com/samber/lo"
	"github.com/uptrace/bun"
)

// StopWordsHandler handles stop word pages.
type StopWordsHandler struct {
	db *bun.DB
	sm *scs.SessionManager
}

// StopWords renders the stopwords form page.
func (h *StopWordsHandler) StopWords(c *server.Context) error {
	if c.User == nil {
		return calendar.Forbidden.New("Must be logged in")
	}

	if c.User.Role != domain.Admin {
		return calendar.Forbidden.New("Only admins can view stopwords")
	}

	switch c.Request().Method {
	case http.MethodGet:
		words, err := model.ListStopWords(c.Request().Context(), h.db)
		if err != nil {
			return err
		}

		return server.RenderPage(c, h.sm,
			html.StopWordsMain(lo.Map(words, func(word *domain.StopWord, _ int) string {
				return word.Word
			}), c.CSRF),
		)

	case http.MethodPost:
		form := contract.EditStopWordsForm{}
		if err := c.Bind(&form); err != nil {
			return err
		}

		var words []string

		scanner := bufio.NewScanner(strings.NewReader(form.Words))
		for scanner.Scan() {
			words = append(words, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			return err
		}

		if err := model.SetStopWords(c.Request().Context(), h.db, domain.NewStopWordList(words...)); err != nil {
			return err
		}

		return c.Redirect(http.StatusSeeOther, "/stopwords")

	default:
		return calendar.NotFound.New("Not found")
	}
}

// Register the handler.
func (h *StopWordsHandler) Register(g *echo.Group) {
	g.GET("/stopwords", server.Wrap(h.db, h.sm, h.StopWords))
	g.POST("/stopwords", server.Wrap(h.db, h.sm, h.StopWords))
}

// NewStopWordsHandler creates a new stop words handler.
func NewStopWordsHandler(db *bun.DB, sm *scs.SessionManager) *StopWordsHandler {
	return &StopWordsHandler{
		db: db,
		sm: sm,
	}
}
