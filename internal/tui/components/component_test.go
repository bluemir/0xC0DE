package components_test

import (
	"testing"

	component "github.com/bluemir/0xC0DE/internal/tui/components"
)

func TestSelectWraps(t *testing.T) {
	s := component.NewSelect([]string{"a", "b", "c"}, 0)
	if s.Prev().Index() != 2 {
		t.Fatalf("prev from 0 = %d, want 2", s.Prev().Index())
	}
	if s.Next().Next().Next().Index() != 0 {
		t.Fatalf("next x3 = %d, want 0 (wrap)", s.Next().Next().Next().Index())
	}
}

func TestNumberClamps(t *testing.T) {
	n := component.NewNumber(component.WithInitial(5), component.WithRange(5, 7))
	if n.Dec().Int() != 5 {
		t.Fatalf("dec below min = %d, want 5", n.Dec().Int())
	}
	hi := component.NewNumber(component.WithInitial(7), component.WithRange(5, 7))
	if hi.Inc().Int() != 7 {
		t.Fatalf("inc above max = %d, want 7", hi.Inc().Int())
	}
}
