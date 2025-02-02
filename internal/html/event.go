package html

import (
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/aybabtme/uniplot/histogram"
	"github.com/mgnsk/calendar/internal/domain"
	"github.com/samber/lo"
	"github.com/yuin/goldmark"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

// EventsPageParams is the params for events page.
type EventsPageParams struct {
	MainTitle             string
	SectionTitle          string
	Path                  string
	FilterTag             string
	User                  *domain.User
	PastEventsTransparent bool
	Events                []*domain.Event
}

// EventsPage display events page.
func EventsPage(p EventsPageParams) Node {
	sectionTitleSuffix := ""
	if p.FilterTag != "" {
		sectionTitleSuffix = fmt.Sprintf(" tagged %s (%d)", p.FilterTag, len(p.Events))
	}

	var navLinks []eventNavLink

	navLinks = append(navLinks,
		eventNavLink{
			Text: "Upcoming",
			URL: func() string {
				if p.FilterTag != "" {
					return fmt.Sprintf("/tag/%s", url.QueryEscape(p.FilterTag))
				}
				return "/"
			}(),
			Active: p.Path == "/" || p.Path == "/tag/:tagName",
		},
		eventNavLink{
			Text: "Past",
			URL: func() string {
				if p.FilterTag != "" {
					return fmt.Sprintf("/past/tag/%s", url.QueryEscape(p.FilterTag))
				}
				return "/past"
			}(),
			Active: p.Path == "/past" || p.Path == "/past/tag/:tagName",
		},
		eventNavLink{
			Text:   "Tags",
			URL:    "/tags",
			Active: false,
		},
	)

	// if p.Path == "/tag/:tagName" || p.Path == "/past/tag/:tagName" {
	// 	navLinks = append(navLinks, eventNavLink{
	// 		Text:   "All",
	// 		URL:    "/",
	// 		Active: false,
	// 	})
	// }

	return page(p.MainTitle, p.SectionTitle+sectionTitleSuffix, p.User,
		eventNav(navLinks),
		eventList(p.Events, p.Path, p.PastEventsTransparent),
	)
}

// TagsPageParams is the params for tags page.
type TagsPageParams struct {
	MainTitle    string
	SectionTitle string
	User         *domain.User
	Tags         []*domain.Tag
}

// TagsPage displays tags list page.
func TagsPage(p TagsPageParams) Node {
	return page(p.MainTitle, p.SectionTitle, p.User,
		eventNav([]eventNavLink{
			{
				Text:   "Upcoming",
				URL:    "/",
				Active: false,
			},
			{
				Text:   "Past",
				URL:    "/past",
				Active: false,
			},
			{
				Text:   "Tags",
				URL:    "/tags",
				Active: true,
			},
		}),
		tagsList(p.Tags),
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

func eventList(events []*domain.Event, path string, pastEventsTransparent bool) Node {
	return Map(events, func(ev *domain.Event) Node {
		return eventCard(ev, path, pastEventsTransparent)
	})
}

func calcHistogram(bins int, tags []*domain.Tag) (histogram.Histogram, []string, []string) {
	counts := lo.Map(tags, func(tag *domain.Tag, _ int) float64 {
		return float64(tag.EventCount)
	})

	sizes := []string{"text-sm", "text-base", "text-lg", "text-xl", "text-2xl", "text-3xl", "text-4xl", "text-5xl"}
	colors := []string{"text-gray-400", "text-gray-500", "text-gray-600", "text-gray-700", "text-gray-800", "text-gray-900", "text-gray-950", "text-black"}

	// When not too many tags or buckets, prefer the largest size classes.
	if len(tags) < len(sizes) {
		sizes = sizes[len(sizes)-len(tags):]
		colors = colors[len(colors)-len(tags):]
	} else if bins < len(sizes) {
		sizes = sizes[len(sizes)-bins:]
		colors = colors[len(colors)-bins:]
	}

	hist := histogram.Hist(len(sizes), counts)
	numBuckets := len(hist.Buckets)

	// Remove zero sized buckets.
	hist.Buckets = slices.DeleteFunc(hist.Buckets, func(b histogram.Bucket) bool {
		return b.Count == 0
	})

	if len(hist.Buckets) < numBuckets {
		// Recalculate with new bucket count.
		return calcHistogram(len(hist.Buckets), tags)
	}

	return hist, sizes, colors
}

func tagsList(tags []*domain.Tag) Node {
	hist, sizes, colors := calcHistogram(8, tags)

	getClassIndex := func(eventCount uint64) int {
		for i, bucket := range hist.Buckets {
			if eventCount >= uint64(bucket.Min) && eventCount <= uint64(bucket.Max) {
				return i
			}
		}
		panic("no bucket found")
	}

	return Ul(Class("my-5 flex justify-center flex-wrap align-center gap-2 leading-8"),
		Map(tags, func(tag *domain.Tag) Node {
			classes := Classes{"hover:underline": true}
			idx := getClassIndex(tag.EventCount)
			classes[sizes[idx]] = true
			classes[colors[idx]] = true

			return Li(
				A(classes,
					Href(fmt.Sprintf("/tag/%s", url.QueryEscape(tag.Name))), Text(tag.Name), Sup(Class("text-gray-400"), Textf("(%d)", tag.EventCount)),
				),
			)
		}),
	)
}

func eventCard(ev *domain.Event, path string, pastEventTransparent bool) Node {
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
				If(len(ev.Tags) > 0, eventTags(ev, path)),
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

	// TODO: syntax error here
	return P(Class("mt-2 text-gray-700"), Raw(buf.String()))
}

func eventDay(ev *domain.Event) Node {
	day := ev.StartAt.Time().Day()

	return P(Class("text-2xl md:text-4xl font-bold"),
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

func eventTags(ev *domain.Event, path string) Node {
	return P(Class("mt-1 text-gray-500 text-sm"),
		mapIndexed(ev.TagRelations, func(i int, tag *domain.Tag) Node {
			href := ""

			switch path {
			case "/", "/tag/:tagName":
				href = fmt.Sprintf("/tag/%s", url.QueryEscape(tag.Name))
			case "/past", "/past/tag/:tagName":
				href = fmt.Sprintf("/past/tag/%s", url.QueryEscape(tag.Name))
			}

			link := A(Class("hover:underline"), Href(href), Text(tag.Name), Sup(Class("text-gray-400"), Textf("(%d)", tag.EventCount)))

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
