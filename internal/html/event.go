package html

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mgnsk/calendar/internal/domain"
	"github.com/mgnsk/calendar/internal/pkg/timestamp"
	"github.com/yuin/goldmark"
	. "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

// EventListPartial renders the event list partial.
func EventListPartial(offset int64, events []*domain.Event, csrf, filterTag string) Node {
	if len(events) == 0 {
		return Div(Class("px-3 py-4 text-center"),
			P(Text("no events found")),
		)
	}

	return Group{
		If(filterTag != "", ScriptRaw(fmt.Sprintf(`setSearch("%s")`, filterTag))), // TODO: tag validation
		Map(events, func(ev *domain.Event) Node {
			return eventCard(ev)
		}),
		Div(ID("load-more"),
			hx.Post(""),
			hx.Include("[name='search']"), // CSS query to include data from inputs.
			hx.Vals(string(must(json.Marshal(map[string]string{
				"csrf":    csrf,
				"last_id": events[len(events)-1].ID.String(),
				"offset":  strconv.FormatInt(offset, 10),
			})))),
			hx.Trigger("intersect once"),
			hx.Target("#load-more"),
			hx.Swap("outerHTML"), // Swap the current element (#load-more) with new content.
			hx.Indicator("#loading-spinner"),
		),
	}
}

// EventsPageParams is the params for events page.
type EventsPageParams struct {
	MainTitle string
	Path      string
	User      *domain.User
	CSRF      string
}

// EventsPage display events page.
func EventsPage(p EventsPageParams) Node {
	var navLinks []eventNavLink

	navLinks = append(navLinks,
		eventNavLink{
			Text:   "Latest",
			URL:    "/",
			Active: p.Path == "/",
		},
		eventNavLink{
			Text:   "Upcoming",
			URL:    "/upcoming",
			Active: p.Path == "/upcoming",
		},
		eventNavLink{
			Text:   "Past",
			URL:    "/past",
			Active: p.Path == "/past",
		},
		eventNavLink{
			Text:   "Tags",
			URL:    "/tags",
			Active: false,
		},
	)

	return Page(p.MainTitle, "", p.User,
		eventNav(navLinks, p.CSRF),
		Main(
			Div(ID("event-list"),
				hx.Post(""),
				hx.Trigger("load"),
				hx.Swap("beforeend"),
				hx.Target("#event-list"),
				hx.Indicator("#loading-spinner"),
				hx.Vals(string(must(json.Marshal(map[string]string{
					"csrf": p.CSRF,
				})))),
			),
		),
		ScriptSync("dist/mark.min.js"),
		ScriptRaw(searchScript),
	)
}

func eventCard(ev *domain.Event) Node {
	inPast := ev.StartAt.Before(time.Now())

	return Div(
		Classes{
			"relative":  true,
			"max-w-3xl": true,
			"mx-auto":   true,

			// Less opacity for events that have already started.
			"opacity-60": inPast,
			// "grayscale":          inPast,
			"bg-white":           true,
			"rounded-xl":         true,
			"shadow-md":          true,
			"overflow-hidden":    true,
			"my-5":               true,
			"hover:bg-gray-300":  true,
			"hover:bg-opacity-5": true,
		},
		Div(Class("py-4 md:py-8 px-3 md:px-6 items-center grid grid-cols-7"),
			Div(Class("pr-4 col-span-1 hidden sm:inline-block"),
				eventDay(ev),
			),
			Div(Class("col-span-6 sm:col-span-5"),
				eventTitle(ev),
				eventDate(ev),
				eventDesc(ev),
			),
		),
		Div(Class("h-full flex flex-col justify-around absolute right-0 top-0 py-4 sm:py-8 px-6 sm:px-10"),
			A(Class("hover:text-amber-600"),
				Href("#"),
				Target("_blank"),
				iconShare,
			),
		),
	)
}

func eventTitle(ev *domain.Event) Node {
	return H1(Class("tracking-wide text-xl md:text-2xl font-semibold"),
		A(Class("hover:underline"), Href(ev.URL), Target("_blank"), Text(ev.Title)),
	)
}

func eventDesc(ev *domain.Event) Node {
	var buf strings.Builder
	if err := goldmark.Convert([]byte(ev.Description), &buf); err != nil {
		// TODO: event must be validated.
		panic(fmt.Errorf("error rendering markdown (event ID %d): %w", ev.ID.Int64(), err))
	}

	return Div(Class("text-justify"),
		// TODO: syntax error here
		Div(Class("mt-2 text-gray-700"), Raw(buf.String())),
	)
}

func eventDay(ev *domain.Event) Node {
	day := ev.StartAt.Day()

	return P(Class("text-2xl md:text-4xl font-bold text-center"),
		Text(timestamp.FormatDay(day)),
	)
}

func eventDate(ev *domain.Event) Node {
	return Group{
		H2(Class("block mt-2 uppercase tracking-wide text-sm text-amber-600 font-semibold"),
			Text(ev.GetDateString()),
		),
	}
}

func mapIndexed[T any](ts []T, cb func(int, T) Node) Group {
	var nodes []Node
	for i, t := range ts {
		nodes = append(nodes, cb(i, t))
	}
	return nodes
}

func must[V any](v V, err error) V {
	if err != nil {
		panic(err)
	}
	return v
}

var iconShare = Raw(`<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6">
  <path stroke-linecap="round" stroke-linejoin="round" d="M7.217 10.907a2.25 2.25 0 1 0 0 2.186m0-2.186c.18.324.283.696.283 1.093s-.103.77-.283 1.093m0-2.186 9.566-5.314m-9.566 7.5 9.566 5.314m0 0a2.25 2.25 0 1 0 3.935 2.186 2.25 2.25 0 0 0-3.935-2.186Zm0-12.814a2.25 2.25 0 1 0 3.933-2.185 2.25 2.25 0 0 0-3.933 2.185Z" />
</svg>`)
