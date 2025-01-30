package html

import (
	"fmt"
	"strings"
	"time"

	"github.com/mgnsk/calendar/internal/domain"
	"github.com/yuin/goldmark"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

// EventList returns a list of events.
func EventList(events []*domain.Event) Node {
	return Map(events, func(ev *domain.Event) Node {
		return EventCard(ev)
	})
}

// EventCard returns an event card.
func EventCard(ev *domain.Event) Node {
	return Div(
		Classes{
			// Less opacity for events that have already started.
			"opacity-50":      ev.StartAt.Time().Before(time.Now()),
			"max-w-md":        true,
			"mx-auto":         true,
			"bg-white":        true,
			"rounded-xl":      true,
			"shadow-md":       true,
			"overflow-hidden": true,
			"md:max-w-2xl":    true,
			"m-5":             true,
		},
		Div(Class("p-8 flex items-center"),
			Div(Class("pr-4"),
				EventDay(ev),
			),
			Div(
				EventTitle(ev),
				EventMonthYear(ev),
				EventTime(ev),
				If(len(ev.Tags) > 0, EventTags(ev)),
				EventDescription(ev),
			),
		),
	)
}

// EventTitle renders event title.
func EventTitle(ev *domain.Event) Node {
	return H1(Class("tracking-wide text-2xl font-semibold"),
		A(Class("hover:underline"), Href(ev.URL), Target("_blank"), Text(ev.Title)),
	)
}

// EventDescription renders event description.
func EventDescription(ev *domain.Event) Node {
	var buf strings.Builder
	if err := goldmark.Convert([]byte(ev.Description), &buf); err != nil {
		// TODO: event must be validated.
		panic(fmt.Errorf("error rendering markdown (event ID %d): %w", ev.ID.Int64(), err))
	}

	return P(Class("mt-2 text-gray-700"), Raw(buf.String()))
}

// EventDay renders event day.
func EventDay(ev *domain.Event) Node {
	day := ev.StartAt.Time().Day()

	return P(Class("text-4xl font-bold"),
		Textf("%d%s", day, getDaySuffix(day)),
	)
}

// EventMonthYear renders event month and year.
func EventMonthYear(ev *domain.Event) Node {
	return H2(Class("mt-2 uppercase tracking-wide text-sm text-amber-600 font-semibold"),
		Text(ev.StartAt.Time().Format("January, 2006")),
	)
}

// EventTime renders event start and end time.
func EventTime(ev *domain.Event) Node {
	return P(Class("mt-2 text-gray-500 text-sm"), Text(func() string {
		start := ev.StartAt.Time().Format("15:04")

		if !ev.EndAt.Time().IsZero() {
			end := ev.EndAt.Time().Format("15:04")
			return fmt.Sprintf("%s - %s", start, end)
		}

		return start
	}()))
}

// EventTags renders event tags.
func EventTags(ev *domain.Event) Node {
	return P(Class("mt-1 text-gray-500 text-sm"),
		MapIndexed(ev.Tags, func(i int, tag string) Node {
			link := A(Class("hover:underline"), Href("#"), Text(tag))

			if i > 0 {
				return Group{
					Text(", "),
					link,
				}
			}

			return link
		}),
	)
}

// MapIndexed maps a slice of anything to a [Group] (which is just a slice of [Node]-s).
func MapIndexed[T any](ts []T, cb func(int, T) Node) Group {
	var nodes []Node
	for i, t := range ts {
		nodes = append(nodes, cb(i, t))
	}
	return nodes
}

func getDaySuffix(n int) string {
	if n >= 11 && n <= 13 {
		return "th"
	}

	switch n % 10 {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	default:
		return "th"
	}
}
