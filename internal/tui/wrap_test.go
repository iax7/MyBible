package tui

import (
	"testing"

	"github.com/iax7/mybible/internal/module"
)

// A link spans its words: the spaces inside it keep the link style, so the
// underline runs unbroken through the Reference.
func TestWrapKeepsTheStyleOfSpacesInsideALink(t *testing.T) {
	lines := wrapRuns([]module.Run{{Text: "2 Co. 4.6", Link: true}}, 80)
	if len(lines) != 1 {
		t.Fatalf("wrapRuns returned %d lines, want 1", len(lines))
	}
	for _, run := range lines[0] {
		if !run.Link {
			t.Errorf("the space or word %q lost the link style", run.Text)
		}
	}
}

// A bold phrase's band covers its spaces too, not just its words.
func TestWrapKeepsTheStyleOfSpacesInsideBold(t *testing.T) {
	lines := wrapRuns([]module.Run{{Text: "en los postreros días.", Bold: true}}, 80)
	if len(lines) != 1 {
		t.Fatalf("wrapRuns returned %d lines, want 1", len(lines))
	}
	for _, run := range lines[0] {
		if !run.Bold {
			t.Errorf("the space or word %q lost the bold style", run.Text)
		}
	}
}
