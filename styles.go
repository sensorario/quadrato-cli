package main

import "github.com/charmbracelet/lipgloss"

var (
	cyan    = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	green   = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	yellow  = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	red     = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
	dim     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	magenta = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	bold    = lipgloss.NewStyle().Bold(true)

	headerStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("6")).
			Padding(0, 1).
			Bold(true)
)

var statusLabel = map[int]string{0: "Da fare", 1: "In corso", 2: "In revisione", 3: "Fatto"}
var statusEmoji = map[int]string{0: "○", 1: "◑", 2: "◕", 3: "●"}
var statusStyle = map[int]func(string) string{
	0: func(s string) string { return s },
	1: func(s string) string { return yellow.Render(s) },
	2: func(s string) string { return cyan.Render(s) },
	3: func(s string) string { return green.Render(s) },
}
