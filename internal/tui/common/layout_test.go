package common

import "testing"

func TestComputeLayout_Standard(t *testing.T) {
	l := ComputeLayout(100, 30, 0)
	if l.Mode != LayoutStandard {
		t.Errorf("Mode = %d, want LayoutStandard(%d)", l.Mode, LayoutStandard)
	}
	if l.ContentW != 100 {
		t.Errorf("ContentW = %d, want 100", l.ContentW)
	}
	if l.ContentH != 28 { // 30 - 1 header - 1 statusbar
		t.Errorf("ContentH = %d, want 28", l.ContentH)
	}
	if l.StatCardW != 25 { // 100 / 4
		t.Errorf("StatCardW = %d, want 25", l.StatCardW)
	}
	if l.StatCardH != 4 {
		t.Errorf("StatCardH = %d, want 4", l.StatCardH)
	}
}

func TestComputeLayout_Compact(t *testing.T) {
	l := ComputeLayout(70, 20, 0)
	if l.Mode != LayoutCompact {
		t.Errorf("Mode = %d, want LayoutCompact(%d)", l.Mode, LayoutCompact)
	}
	if l.StatCardW != 35 { // 70 / 2
		t.Errorf("StatCardW = %d, want 35", l.StatCardW)
	}
	if l.ChartRowH != 0 {
		t.Errorf("ChartRowH = %d, want 0 in compact", l.ChartRowH)
	}
}

func TestComputeLayout_Comfort(t *testing.T) {
	l := ComputeLayout(140, 50, 0)
	if l.Mode != LayoutComfort {
		t.Errorf("Mode = %d, want LayoutComfort(%d)", l.Mode, LayoutComfort)
	}
	if l.ChartLeftW == 0 {
		t.Error("ChartLeftW should not be 0 in comfort mode")
	}
	if l.ChartRightW == 0 {
		t.Error("ChartRightW should not be 0 in comfort mode")
	}
	// 60/40 split
	if l.ChartLeftW != 84 { // 140 * 60 / 100
		t.Errorf("ChartLeftW = %d, want 84", l.ChartLeftW)
	}
}

func TestComputeLayout_Wide(t *testing.T) {
	l := ComputeLayout(200, 60, 0)
	if l.Mode != LayoutWide {
		t.Errorf("Mode = %d, want LayoutWide(%d)", l.Mode, LayoutWide)
	}
}

func TestComputeLayout_Sessions(t *testing.T) {
	l := ComputeLayout(100, 30, 1) // tab 1 = sessions
	if l.FilterBarH != 2 {
		t.Errorf("FilterBarH = %d, want 2", l.FilterBarH)
	}
	if l.PreviewH != 3 {
		t.Errorf("PreviewH = %d, want 3", l.PreviewH)
	}
	expected := 28 - 2 - 3 // contentH - filterBar - preview
	if l.TableH != expected {
		t.Errorf("TableH = %d, want %d", l.TableH, expected)
	}
}

func TestComputeLayout_Analytics(t *testing.T) {
	l := ComputeLayout(100, 30, 2) // tab 2 = analytics
	if l.TopChartH == 0 {
		t.Error("TopChartH should not be 0")
	}
	total := l.TopChartH + l.MidChartH + l.BotChartH
	if total != l.ContentH {
		t.Errorf("chart heights sum = %d, want ContentH = %d", total, l.ContentH)
	}
}
