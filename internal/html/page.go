package html

import (
	"fmt"

	"github.com/mgnsk/calendar/internal"
	"github.com/mgnsk/calendar/internal/domain"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

// EventListPage renders event list page.
func EventListPage(events []*domain.Event) Node {
	return page("Event List",
		EventList(events),
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
