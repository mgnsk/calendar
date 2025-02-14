package html

import (
	"encoding/json"
	"fmt"

	"github.com/aybabtme/uniplot/histogram"
	"github.com/mgnsk/calendar/internal/domain"
	"github.com/samber/lo"
	. "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

// TagsMain renders the tags page main content.
func TagsMain(csrf string) Node {
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

// TagListPartial renders the tag list partial.
func TagListPartial(tags []*domain.Tag, csrf string) Node {
	if len(tags) == 0 {
		return Div(Class("px-3 py-4 text-center"),
			P(Text("no tags found")),
		)
	}

	hist, sizes, colors := calcHistogram(8, tags)

	getClassIndex := func(eventCount uint64) int {
		for i, bucket := range hist.Buckets {
			if eventCount >= uint64(bucket.Min) && eventCount <= uint64(bucket.Max) {
				return i
			}
		}
		panic("no bucket found")
	}

	return Div(Class("max-w-3xl mx-auto my-5"),
		Ul(Class("flex justify-center flex-wrap align-center gap-2 leading-8"),
			Map(tags, func(tag *domain.Tag) Node {
				classes := Classes{
					"hover:underline":      true,
					"hover:cursor-pointer": true,
				}
				idx := getClassIndex(tag.EventCount)
				classes[sizes[idx]] = true
				classes[colors[idx]] = true

				return Li(
					A(classes,
						Text(tag.Name),
						Sup(Class("text-gray-400"),
							Textf("(%d)", tag.EventCount),
						),
						// Show latest tagged events on click.
						hx.Post("/"),
						hx.Trigger("click"),
						Attr("onclick", fmt.Sprintf(`changeTab(document.querySelectorAll(".nav-link")[0]); setSearch("%s")`, tag.Name)),

						hx.Target("#event-list"),
						hx.Swap("innerHTML"),
						hx.PushURL("true"),
						hx.Indicator("#loading-spinner"),
						hx.Vals(string(must(json.Marshal(map[string]string{
							"csrf":   csrf,
							"search": tag.Name,
						})))),
					),
				)
			}),
		),
	)
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

	return hist, sizes, colors
}
