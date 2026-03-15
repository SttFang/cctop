package common

import tea "github.com/charmbracelet/bubbletea"

func IsQuit(msg tea.KeyMsg) bool {
	return msg.String() == "q" || msg.String() == "ctrl+c"
}

func TabFromKey(msg tea.KeyMsg) int {
	switch msg.String() {
	case "1":
		return 0
	case "2":
		return 1
	case "3":
		return 2
	case "4":
		return 3
	default:
		return -1
	}
}

func IsTab(msg tea.KeyMsg) bool       { return msg.Type == tea.KeyTab }
func IsShiftTab(msg tea.KeyMsg) bool   { return msg.Type == tea.KeyShiftTab }
func IsRefresh(msg tea.KeyMsg) bool    { return msg.String() == "r" }
func IsThemeToggle(msg tea.KeyMsg) bool { return msg.String() == "t" }
func IsHelp(msg tea.KeyMsg) bool       { return msg.String() == "?" }
func IsSearch(msg tea.KeyMsg) bool     { return msg.String() == "/" }
func IsSort(msg tea.KeyMsg) bool       { return msg.String() == "s" }
func IsEnter(msg tea.KeyMsg) bool      { return msg.Type == tea.KeyEnter }
func IsEscape(msg tea.KeyMsg) bool     { return msg.Type == tea.KeyEscape }
