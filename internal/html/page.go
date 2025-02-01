package html

import (
	"fmt"

	"github.com/mgnsk/calendar/internal"
	"github.com/mgnsk/calendar/internal/domain"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

const (
	currentEventsTitle = "Current Events"
	pastEventsTitle    = "Past Events"
)

// CurrentEventsPage displays current events page.
func CurrentEventsPage(events []*domain.Event) Node {
	return eventsPage(currentEventsTitle, events)
}

// PastEventsPage displays past events page.
func PastEventsPage(events []*domain.Event) Node {
	return eventsPage(pastEventsTitle, events)
}

func eventsPage(title string, events []*domain.Event) Node {
	return page(title,
		Div(Class("m-5 max-w-2xl mx-auto"),
			pageTitle(title),
			navLinks([]navLink{
				{
					Text:   currentEventsTitle,
					URL:    "/",
					Active: title == currentEventsTitle,
				},
				{
					Text:   pastEventsTitle,
					URL:    "/past",
					Active: title == pastEventsTitle,
				},
			}),
			EventList(events, true),
		),
	)
}

func pageTitle(title string) Node {
	return H1(Class("m-8 text-center text-4xl font-semibold"), Text(title))
}

type navLink struct {
	Text   string
	URL    string
	Active bool
}

func navLinks(links []navLink) Node {
	return Ul(Class("flex border-b"),
		Map(links, func(link navLink) Node {
			if link.Active {
				return activeLinkItem(link)
			}

			return inactiveLinkItem(link)
		}),
	)
}

func activeLinkItem(link navLink) Node {
	return Li(Class("-mb-px mr-1"),
		A(Class("bg-white inline-block border-l border-t border-r rounded-t py-2 px-4 text-amber-600 font-semibold"),
			Text(link.Text),
			Href(link.URL),
		),
	)
}

func inactiveLinkItem(link navLink) Node {
	return Li(Class("mr-1"),
		A(Class("bg-white inline-block py-2 px-4 text-gray-400 hover:text-amber-600 font-semibold"),
			Text(link.Text),
			Href(link.URL),
		),
	)
}

func page(title string, children ...Node) Node {
	return HTML5(HTML5Props{
		Title:    title,
		Language: "en",
		Head: []Node{
			Link(Rel("stylesheet"), Href(fmt.Sprintf("/dist/app.css?crc=%d", internal.Checksums["dist/app.css"]))),
		},
		Body: []Node{
			Group(children),
		},
	})
}
