package common

import "testing"

func TestDarkThemeComplete(t *testing.T) {
	theme := DarkTheme
	if theme.Bg == "" {
		t.Error("DarkTheme.Bg is empty")
	}
	if theme.Fg == "" {
		t.Error("DarkTheme.Fg is empty")
	}
	if theme.Primary == "" {
		t.Error("DarkTheme.Primary is empty")
	}
	if theme.Secondary == "" {
		t.Error("DarkTheme.Secondary is empty")
	}
	if theme.Green == "" {
		t.Error("DarkTheme.Green is empty")
	}
	if theme.Red == "" {
		t.Error("DarkTheme.Red is empty")
	}
}

func TestLightThemeComplete(t *testing.T) {
	theme := LightTheme
	if theme.Bg == "" {
		t.Error("LightTheme.Bg is empty")
	}
	if theme.Fg == "" {
		t.Error("LightTheme.Fg is empty")
	}
	if theme.Primary == "" {
		t.Error("LightTheme.Primary is empty")
	}
}

func TestModelColor(t *testing.T) {
	theme := DarkTheme
	if theme.ModelColor("opus-4-6") != theme.Primary {
		t.Error("opus should be Primary color")
	}
	if theme.ModelColor("sonnet-4-6") != theme.Secondary {
		t.Error("sonnet should be Secondary color")
	}
	if theme.ModelColor("haiku-4-5") != theme.Orange {
		t.Error("haiku should be Orange color")
	}
	if theme.ModelColor("unknown") != theme.Cyan {
		t.Error("unknown should be Cyan color")
	}
}

func TestCostColor(t *testing.T) {
	theme := DarkTheme
	if theme.CostColor(10.0) != theme.Red {
		t.Error("high cost should be Red")
	}
	if theme.CostColor(2.0) != theme.Yellow {
		t.Error("medium cost should be Yellow")
	}
	if theme.CostColor(0.5) != theme.Green {
		t.Error("low cost should be Green")
	}
}
