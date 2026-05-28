package handler

import (
	"bufio"
	"net/http"
	"strings"

	"github.com/alexedwards/scs/v2"
	"github.com/ggicci/httpin"
	"github.com/mgnsk/calendar"
	"github.com/mgnsk/calendar/contract"
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/html"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/server"
	"github.com/uptrace/bun"
)

// StopWordsHandler handles stop word pages.
type StopWordsHandler struct {
	db *bun.DB
	sm *scs.SessionManager
}

// StopWords renders the stopwords form page.
func (h *StopWordsHandler) StopWords(w http.ResponseWriter, r *http.Request) {
	user := server.GetUser(r.Context())
	if user == nil {
		panic(calendar.Forbidden.New("Must be logged in"))
	}

	if user.Role != domain.Admin {
		panic(calendar.Forbidden.New("Only admins can view stopwords"))
	}

	switch r.Method {
	case http.MethodGet:
		words, err := model.ListStopWords(r.Context(), h.db)
		if err != nil {
			panic(err)
		}

		server.RenderPage(w, r, h.sm,
			html.StopWordsMain(words),
		)
		return

	case http.MethodPost:
		form := contract.EditStopWordsForm{}
		if err := httpin.DecodeTo(r, &form); err != nil {
			panic(err)
		}

		var words []string

		scanner := bufio.NewScanner(strings.NewReader(form.Words))
		for scanner.Scan() {
			words = append(words, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			panic(err)
		}

		if err := model.SetStopWords(r.Context(), h.db, domain.NewStopWordList(words...)); err != nil {
			panic(err)
		}

		http.Redirect(w, r, "/stopwords", http.StatusSeeOther)
		return

	default:
		panic(calendar.NotFound.New("Not found"))
	}
}

// Register the handler.
func (h *StopWordsHandler) Register(mux *http.ServeMux) {
	middlewares := []server.MiddlewareFunc{
		server.NewSessionMiddleware(h.sm),
		server.NewSettingsMiddleware(h.db),
		server.NewUserMiddleware(h.db, h.sm),
	}

	mux.Handle("GET /stopwords", server.WithMiddleware(http.HandlerFunc(h.StopWords), middlewares...))
	mux.Handle("POST /stopwords", server.WithMiddleware(http.HandlerFunc(h.StopWords), middlewares...))
}

// NewStopWordsHandler creates a new stop words handler.
func NewStopWordsHandler(db *bun.DB, sm *scs.SessionManager) *StopWordsHandler {
	return &StopWordsHandler{
		db: db,
		sm: sm,
	}
}
