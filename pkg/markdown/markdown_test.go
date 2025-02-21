package markdown_test

import (
	"strings"

	"github.com/mgnsk/calendar/pkg/markdown"
	. "github.com/mgnsk/calendar/pkg/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("rendering markdown", func() {
	DescribeTable("allowed markdown elements",
		func(source, expectedHTML string) {
			html := Must(markdown.Convert(source))
			html = strings.TrimSpace(html)
			Expect(html).To(Equal(expectedHTML))
		},

		Entry("bold",
			`**Text**`,
			`<p><strong>Text</strong></p>`,
		),

		Entry("italic",
			`*Text*`,
			`<p><em>Text</em></p>`,
		),

		Entry("strikethrough",
			`~~Text~~`,
			`<p><del>Text</del></p>`,
		),

		Entry("link",
			`[Link](https://calendar.testing)`,
			`<p><a href="https://calendar.testing" target="_blank" rel="noopener">Link</a></p>`,
		),

		Entry("autolink",
			`https://calendar.testing`,
			`<p><a href="https://calendar.testing" target="_blank" rel="noopener">https://calendar.testing</a></p>`,
		),

		// Any elements not specifically allowed are forbidden.
		Entry("forbidden element is ignored",
			"```Text```",
			`<p></p>`,
		),
	)

	DescribeTable("XSS mitigation",
		func(source, expectedHTML string) {
			html := Must(markdown.Convert(source))
			html = strings.TrimSpace(html)
			Expect(html).To(Equal(expectedHTML))
		},

		Entry("bold tag",
			`<b>Text</b>`,
			`<p><!-- raw HTML omitted -->Text<!-- raw HTML omitted --></p>`,
		),

		Entry("script tag",
			`<script>alert('Borked!')</script>`,
			`<!-- raw HTML omitted -->`,
		),

		Entry("javascript in link",
			`[Link](javascript:alert('xss'))`,
			`<p><a href="" target="_blank" rel="noopener">Link</a></p>`,
		),

		Entry("multiline",
			`hello <a name="n"
href="javascript:alert('xss')">you</a>`,
			`<p>hello <!-- raw HTML omitted -->you<!-- raw HTML omitted --></p>`,
		),

		Entry("attribute",
			`[Link](" onclick="alert('xss')")`,
			`<p>[Link](&quot; onclick=&quot;alert('xss')&quot;)</p>`,
		),
	)

	DescribeTable("linebreaks",
		func(source, expectedHTML string) {
			html := Must(markdown.Convert(source))
			html = strings.TrimSpace(html)
			Expect(html).To(Equal(expectedHTML))
		},

		Entry("single line break",
			"one\ntwo",
			"<p>one<br>\ntwo</p>",
		),

		Entry("two line breaks",
			"one\n\ntwo",
			"<p>one</p>\n<p>two</p>",
		),

		Entry("three line breaks",
			"one\n\n\ntwo",
			"<p>one</p>\n<p>two</p>",
		),
	)
})
