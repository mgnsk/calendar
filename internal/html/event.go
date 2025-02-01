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

// CurrentEventsPage displays current events page.
func CurrentEventsPage(mainTitle string, user *domain.User, events []*domain.Event) Node {
	return eventsPage(mainTitle, currentEventsTitle, user, true, events)
}

// PastEventsPage displays past events page.
func PastEventsPage(mainTitle string, user *domain.User, events []*domain.Event) Node {
	return eventsPage(mainTitle, pastEventsTitle, user, false, events)
}

const (
	currentEventsTitle = "Current Events"
	pastEventsTitle    = "Past Events"
)

func eventsPage(mainTitle, sectionTitle string, user *domain.User, pastEventsTransparent bool, events []*domain.Event) Node {
	return page(mainTitle, sectionTitle, user,
		eventNav([]eventNavLink{
			{
				Text:   currentEventsTitle,
				URL:    "/",
				Active: sectionTitle == currentEventsTitle,
			},
			{
				Text:   pastEventsTitle,
				URL:    "/past",
				Active: sectionTitle == pastEventsTitle,
			},
		}),
		eventList(events, pastEventsTransparent),
	)
}

type eventNavLink struct {
	Text   string
	URL    string
	Active bool
}

func eventNav(links []eventNavLink) Node {
	return Ul(Class("flex border-b"),
		Map(links, func(link eventNavLink) Node {
			if link.Active {
				return Li(Class("-mb-px mr-1"),
					A(Aria("current", "page"), Class("bg-white inline-block border-l border-t border-r rounded-t py-2 px-4 text-amber-600 font-semibold"),
						Text(link.Text),
						Href(link.URL),
					),
				)
			}

			return Li(Class("mr-1"),
				A(Class("bg-white inline-block py-2 px-4 text-gray-400 hover:text-amber-600 font-semibold"),
					Text(link.Text),
					Href(link.URL),
				),
			)
		}),
	)
}

// eventList returns a list of events.
func eventList(events []*domain.Event, pastEventsTransparent bool) Node {
	return Map(events, func(ev *domain.Event) Node {
		return eventCard(ev, pastEventsTransparent)
	})
}

func eventCard(ev *domain.Event, pastEventTransparent bool) Node {
	return Div(
		Classes{
			// Less opacity for events that have already started.
			"opacity-50":         pastEventTransparent && ev.StartAt.Time().Before(time.Now()),
			"bg-white":           true,
			"rounded-xl":         true,
			"shadow-md":          true,
			"overflow-hidden":    true,
			"my-5":               true,
			"hover:bg-amber-600": true,
			"hover:bg-opacity-5": true,
		},
		Div(Class("p-8 flex items-center"),
			Div(Class("pr-4"),
				eventDay(ev),
			),
			Div(
				eventTitle(ev),
				eventMonthYear(ev),
				eventTime(ev),
				If(len(ev.Tags) > 0, eventTags(ev)),
				eventDesc(ev),
			),
		),
	)
}

func eventTitle(ev *domain.Event) Node {
	return H1(Class("tracking-wide text-2xl font-semibold"),
		A(Class("hover:underline"), Href(ev.URL), Target("_blank"), Text(ev.Title)),
	)
}

func eventDesc(ev *domain.Event) Node {
	var buf strings.Builder
	if err := goldmark.Convert([]byte(ev.Description), &buf); err != nil {
		// TODO: event must be validated.
		panic(fmt.Errorf("error rendering markdown (event ID %d): %w", ev.ID.Int64(), err))
	}

	// TODO: syntax error here
	return P(Class("mt-2 text-gray-700"), Raw(buf.String()))
}

func eventDay(ev *domain.Event) Node {
	day := ev.StartAt.Time().Day()

	return P(Class("text-4xl font-bold"),
		Textf("%d%s", day, getDaySuffix(day)),
	)
}

func eventMonthYear(ev *domain.Event) Node {
	return H2(Class("mt-2 uppercase tracking-wide text-sm text-amber-600 font-semibold"),
		Text(ev.StartAt.Time().Format("January, 2006")),
	)
}

func eventTime(ev *domain.Event) Node {
	return P(Class("mt-2 text-gray-500 text-sm"), Text(func() string {
		start := ev.StartAt.Time().Format("15:04")

		if !ev.EndAt.Time().IsZero() {
			end := ev.EndAt.Time().Format("15:04")
			return fmt.Sprintf("%s - %s", start, end)
		}

		return start
	}()))
}

func eventTags(ev *domain.Event) Node {
	return P(Class("mt-1 text-gray-500 text-sm"),
		mapIndexed(ev.Tags, func(i int, tag string) Node {
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

func mapIndexed[T any](ts []T, cb func(int, T) Node) Group {
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
