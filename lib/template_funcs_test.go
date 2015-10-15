package gostatic

import (
	"testing"
)

// Tests that stripping HTML works as intended.
func TestStripHTML(t *testing.T) {
	expected := "cruft"
	var tests = []string{
		`<a href="link.html">cruft</a>`,
		`<b><span><div>cruf</div>t</span></b>`,
		`<a>cruf</a>t`,
		`<a ><span ><h1>cr</h1>u</span>f</a>t`,
		`<br/>c<br/>r<br />u<br />ft`,
	}

	for _, s := range tests {
		actual := StripHTML(s)
		if actual != expected {
			t.Errorf("expected '%s', got '%s'", expected, actual)
		}
	}
}

// Tests stripping of newlines.
func TestStripNewlines(t *testing.T) {
	expected := "cruft"
	var tests = []string{
		"c\r\nr\r\nu\r\nft",
		"\r\n\r\n\r\n\r\n\r\ncru\r\nft",
	}

	for _, s := range tests {
		actual := StripNewlines(s)
		if actual != expected {
			t.Errorf("expected '%s', got '%s'", expected, actual)
		}
	}

}

// Test making an excerpt out of a piece of text.
func TestExcerpt(t *testing.T) {
	inputText := "The quick'ned brown fox, jumps; over the lazy doo-dawg."
	var testTable = []struct {
		maxWords int
		expected string
	}{
		{0, ""},
		{1, "The [...]"},
		{3, "The quick'ned brown [...]"},
		{4, "The quick'ned brown fox, [...]"},
		{8, "The quick'ned brown fox, jumps; over the lazy [...]"},
		{99, "The quick'ned brown fox, jumps; over the lazy doo-dawg."},
	}

	for _, s := range testTable {
		out := Excerpt(inputText, s.maxWords)
		if out != s.expected {
			t.Errorf("Expected \"%s\", got \"%s\"", s.expected, out)
		}
	}
}
