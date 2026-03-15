package components

import (
	"strings"
	"testing"

	"github.com/fanghanjun/cctop/internal/tui/common"
)

func TestStatCardRender_Normal(t *testing.T) {
	card := RenderStatCard(StatCardProps{
		Title:   "Est. Cost",
		Value:   "$1,284.32",
		SubText: "today $42",
		DeltaUp: true,
		Width:   20,
		Height:  4,
		Theme:   common.DarkTheme,
	})

	if !strings.Contains(card, "$1,284.32") {
		t.Error("card should contain value")
	}
	if !strings.Contains(card, "Est. Cost") {
		t.Error("card should contain title")
	}
	if !strings.Contains(card, "▲") {
		t.Error("card should contain up arrow for DeltaUp")
	}
}

func TestStatCardRender_Zero(t *testing.T) {
	card := RenderStatCard(StatCardProps{
		Title:  "Cost",
		Value:  "$0.00",
		Width:  20,
		Height: 4,
		Theme:  common.DarkTheme,
	})

	if !strings.Contains(card, "$0.00") {
		t.Error("card should contain $0.00")
	}
}

func TestStatCardRender_TooSmall(t *testing.T) {
	card := RenderStatCard(StatCardProps{
		Title: "Cost",
		Value: "$100",
		Width: 3, // too small
		Theme: common.DarkTheme,
	})

	if card != "" {
		t.Error("too-small card should return empty string")
	}
}

func TestRenderStatCards_Row(t *testing.T) {
	cards := []StatCardProps{
		{Title: "A", Value: "1", Width: 20, Height: 4, Theme: common.DarkTheme},
		{Title: "B", Value: "2", Width: 20, Height: 4, Theme: common.DarkTheme},
	}
	result := RenderStatCards(cards)
	if result == "" {
		t.Error("should render non-empty cards row")
	}
}
