package html

import (
	"encoding/json"
	"fmt"
	"maps"

	"github.com/aybabtme/uniplot/histogram"
	"github.com/mgnsk/calendar/domain"
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

	hist, classes := calcHistogram(tags)

	getHistogramClasses := func(tag *domain.Tag) Classes {
		for i, bucket := range hist.Buckets {
			if tag.EventCount >= uint64(bucket.Min) && tag.EventCount <= uint64(bucket.Max) {
				return classes[i]
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
				maps.Copy(classes, getHistogramClasses(tag))

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

func calcHistogram(tags []*domain.Tag) (histogram.Histogram, []Classes) {
	classes := []Classes{
		{
			"text-sm":       true,
			"text-gray-400": true,
		},
		{
			"text-base":     true,
			"text-gray-500": true,
		},
		{
			"text-lg":       true,
			"text-gray-600": true,
		},
		{
			"text-xl":       true,
			"text-gray-700": true,
		},
		{
			"text-2xl":      true,
			"text-gray-800": true,
		},
		{
			"text-3xl":      true,
			"text-gray-900": true,
		},
		{
			"text-4xl":      true,
			"text-gray-950": true,
		},
		{
			"text-5xl":   true,
			"text-black": true,
		},
	}

	// When not too many tags or buckets, prefer the largest size classes.
	if len(tags) < len(classes) {
		classes = classes[len(classes)-len(tags):]
	}

	counts := lo.Map(tags, func(tag *domain.Tag, _ int) float64 {
		return float64(tag.EventCount)
	})

	hist := histogram.Hist(len(classes), counts)

	return hist, classes
}
