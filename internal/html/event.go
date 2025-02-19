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

// EventsMain renders the events page main content.
func EventsMain(csrf string) Node {
	return Main(
		Div(ID("event-list"),
			hx.Post(""),
			hx.Trigger("load"),
			hx.Swap("beforeend"),
			hx.Target("#event-list"),
			hx.Indicator("#loading-spinner"),
			hx.Vals(string(must(json.Marshal(map[string]string{
				"csrf": csrf,
			})))),
		),
	)
}

// EventListPartial renders the event list partial.
func EventListPartial(offset int64, events []*domain.Event, csrf string) Node {
	if len(events) == 0 {
		return Div(Class("px-3 py-4 text-center"),
			P(Text("no events found")),
		)
	}

	return Group{
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

func eventCard(ev *domain.Event) Node {
	inPast := ev.StartAt.Before(time.Now())

	return Div(
		Classes{
			"event-card": true,
			"relative":   true,
			"max-w-3xl":  true,
			"mx-auto":    true,

			// Less opacity for events that have already started.
			"opacity-60":         inPast,
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
			Div(Class("col-span-7 sm:col-span-6"),
				eventTitle(ev),
				eventDate(ev),
				eventDesc(ev),
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
