package main

import (
	"testing"
)

func TestMediawiki(t *testing.T) {
	tests := []string{
		`Hello, world.`,
		`[[link]]`,
		`[[link1]][[link2]]`,
		`some stuff [[link1]] and more [[link with spaces]]`,
		`[[link|with title]]`,
		`here is a [[regular link]] and a ![[escaped one]]`,
	}

	text_results := []string{
		`Hello, world.`,
		`<a href="/link">link</a>`,
		`<a href="/link1">link1</a><a href="/link2">link2</a>`,
		`some stuff <a href="/link1">link1</a> and more <a href="/link with spaces">link with spaces</a>`,
		`<a href="/link">with title</a>`,
		`here is a <a href="/regular link">regular link</a> and a [[escaped one]]`,
	}
	link_results := [][]string{
		{},
		{"/link"},
		{"/link1", "/link2"},
		{"/link1", "/link with spaces"},
		{"/link"},
		{"/regular link"},
	}

	for i := 0; i < len(tests); i++ {
		test := tests[i]
		r, l := MediaWikiParse(test)
		if r != text_results[i] {
			t.Errorf("Expected >%s<, got >%s<\n", text_results[i], r)
		}
		for j, k := range l {
			if k != link_results[i][j] {
				t.Errorf("Expected >%s<, got >%s<\n", link_results[i][j], k)
			}
		}
	}

}
