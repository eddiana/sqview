package main

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors — all white for now, will customize later
	colorPrimary    = lipgloss.Color("#FFFFFF")
	colorSecondary  = lipgloss.Color("#FFFFFF")
	colorBg         = lipgloss.Color("#FFFFFF")
	colorSurface    = lipgloss.Color("#FFFFFF")
	colorText       = lipgloss.Color("#FFFFFF")
	colorDim        = lipgloss.Color("#FFFFFF")
	colorAccent     = lipgloss.Color("#FFFFFF")
	colorError      = lipgloss.Color("#FFFFFF")
	colorSuccess    = lipgloss.Color("#FFFFFF")
	colorBorder     = lipgloss.Color("#FFFFFF")
	colorHeaderBg   = lipgloss.Color("#FFFFFF")
	colorSelectedBg = lipgloss.Color("#FFFFFF")
	colorFooterBg   = lipgloss.Color("#FFFFFF")

	// Panel styles
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	panelFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.DoubleBorder()).
				BorderForeground(colorBorder).
				Padding(0, 1)

	panelTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorSecondary).
			Padding(0, 1)

	// Table panel styles
	tableListStyle = lipgloss.NewStyle().
			Foreground(colorText)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true).
				Width(20)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(colorText).
			Width(20)

	// Content panel styles
	headerCellStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	cellStyle = lipgloss.NewStyle().
			Foreground(colorText)

	// SQL panel styles
	sqlInputStyle = lipgloss.NewStyle().
			Foreground(colorText).
			BorderForeground(colorBorder)

	sqlErrorStyle = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true).
			Padding(0, 1)

	sqlHintStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	// Footer styles
	footerStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			Padding(0, 1)

	exitButtonStyle = lipgloss.NewStyle().
			Foreground(colorError).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorError).
			Padding(0, 2).
			Bold(true)

	// Status message
	statusStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			Padding(0, 1)
)
