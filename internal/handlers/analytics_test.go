package handlers

import (
	"testing"

	"github.com/Jeanpigi/blog/db"
)

func TestFillDays(t *testing.T) {
	got := fillDays([]db.DailyPoint{{Date: "2026-10-10", Views: 5, Visitors: 3}}, "2026-10-12", 7)
	if len(got) != 7 {
		t.Fatalf("len = %d, want 7", len(got))
	}
	if got[0].Date != "2026-10-06" || got[6].Date != "2026-10-12" {
		t.Errorf("rango = %s..%s", got[0].Date, got[6].Date)
	}
	if got[4].Views != 5 || got[5].Views != 0 {
		t.Errorf("valores mal ubicados: %+v", got)
	}
}
