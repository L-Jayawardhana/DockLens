package main

import "github.com/charmbracelet/lipgloss"

var (
	// Color palette
	colorCyan      = lipgloss.Color("#00f5f9ff")
	colorDimCyan   = lipgloss.Color("#006e73")
	colorGreen     = lipgloss.Color("#00e5a0")
	colorYellow    = lipgloss.Color("#f5c542")
	colorRed       = lipgloss.Color("#ff5f5f")
	colorGray      = lipgloss.Color("#4a4a5a")
	colorLightGray = lipgloss.Color("#8888aa")
	colorWhite     = lipgloss.Color("#e8e8f0")
	colorBg        = lipgloss.Color("#0d0d14")
	colorBgPanel   = lipgloss.Color("#12121c")
	colorBorder    = lipgloss.Color("#1e1e30")
	colorAccent    = lipgloss.Color("#7c3aed")

	// Base styles
	baseStyle = lipgloss.NewStyle().
			Background(colorBg).
			Foreground(colorWhite)

	// Logo style
	logoStyle = lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true)

	// Header bar
	headerStyle = lipgloss.NewStyle().
		//Background(colorBg).
		Foreground(colorLightGray).
		PaddingLeft(2).
		PaddingRight(2)

	// Tab bar
	tabBarStyle = lipgloss.NewStyle().
		//Background(colorBg).
		PaddingLeft(1)

	activeTabStyle = lipgloss.NewStyle().
			Foreground(colorCyan).
		//Background(colorBgPanel).
		Bold(true).
		PaddingLeft(2).
		PaddingRight(2).
		Border(lipgloss.Border{
			Top:         "─",
			Bottom:      "─",
			Left:        "│",
			Right:       "│",
			TopLeft:     "╭",
			TopRight:    "╮",
			BottomLeft:  "╰",
			BottomRight: "╯",
		}, true).
		BorderForeground(colorCyan)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(colorGray).
		//Background(colorBg).
		PaddingLeft(2).
		PaddingRight(2).
		Border(lipgloss.Border{
			Top:         "─",
			Bottom:      "─",
			Left:        "│",
			Right:       "│",
			TopLeft:     "╭",
			TopRight:    "╮",
			BottomLeft:  "╰",
			BottomRight: "╯",
		}, true).
		BorderForeground(colorBorder)

	// Panels
	leftPanelStyle = lipgloss.NewStyle().
		//Background(colorBgPanel).
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(colorBorder)

	rightPanelStyle = lipgloss.NewStyle().
		//Background(colorBgPanel).
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(colorBorder)

	// List item styles
	listItemStyle = lipgloss.NewStyle().
			Foreground(colorWhite).
			PaddingLeft(1).
			PaddingRight(1)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(colorCyan).
		//Background(lipgloss.Color("#111122")).
		Bold(true).
		PaddingLeft(1).
		PaddingRight(1)

	// Status badge styles
	statusRunningStyle = lipgloss.NewStyle().
				Foreground(colorGreen).
				Bold(true)

	statusStoppedStyle = lipgloss.NewStyle().
				Foreground(colorRed)

	statusPausedStyle = lipgloss.NewStyle().
				Foreground(colorYellow)

	// Detail panel text
	detailKeyStyle = lipgloss.NewStyle().
			Foreground(colorDimCyan).
			Width(18)

	detailValStyle = lipgloss.NewStyle().
			Foreground(colorWhite)

	detailSectionStyle = lipgloss.NewStyle().
				Foreground(colorCyan).
				Bold(true).
				MarginTop(1)

	// Footer
	footerStyle = lipgloss.NewStyle().
		//Background(colorBg).
		Foreground(colorGray).
		PaddingLeft(2).
		PaddingRight(2)

	keyStyle = lipgloss.NewStyle().
			Foreground(colorBgPanel).
			Background(colorDimCyan).
			Bold(true).
			PaddingLeft(1).
			PaddingRight(1)

	keyDescStyle = lipgloss.NewStyle().
			Foreground(colorGray)

	// Section title inside panel
	panelTitleStyle = lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true).
			PaddingBottom(1)

	// Divider
	dividerStyle = lipgloss.NewStyle().
			Foreground(colorBorder)

	// Metric styles
	metricLabelStyle = lipgloss.NewStyle().
				Foreground(colorLightGray).
				Width(14)

	metricValueStyle = lipgloss.NewStyle().
				Foreground(colorYellow).
				Bold(true)

	metricValueGreenStyle = lipgloss.NewStyle().
				Foreground(colorGreen).
				Bold(true)

	// Log line styles
	logTimestampStyle = lipgloss.NewStyle().
				Foreground(colorGray)

	logTextStyle = lipgloss.NewStyle().
			Foreground(colorWhite)

	// Context menu
	contextMenuStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#1a1a2e")).
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(colorAccent).
				PaddingLeft(1).
				PaddingRight(1)

	contextMenuItemStyle = lipgloss.NewStyle().
				Foreground(colorWhite).
				PaddingLeft(1)

	contextMenuSelectedStyle = lipgloss.NewStyle().
					Foreground(colorAccent).
					Bold(true).
					PaddingLeft(1)

	// System dashboard
	dashboardBoxStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#0f0f1e")).
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(colorAccent).
				Padding(0, 1)

	// Progress bar
	progressBarFilledStyle = lipgloss.NewStyle().
				Background(colorCyan)

	progressBarEmptyStyle = lipgloss.NewStyle().
				Background(colorBorder)

	// Help overlay
	helpOverlayStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#0a0a18")).
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(colorCyan).
				Padding(1, 3)
)

const logo = `
██████╗  ██████╗  ██████╗██╗  ██╗██╗     ███████╗███╗   ██╗███████╗
██╔══██╗██╔═══██╗██╔════╝██║ ██╔╝██║     ██╔════╝████╗  ██║██╔════╝
██║  ██║██║   ██║██║     █████╔╝ ██║     █████╗  ██╔██╗ ██║███████╗
██║  ██║██║   ██║██║     ██╔═██╗ ██║     ██╔══╝  ██║╚██╗██║╚════██║
██████╔╝╚██████╔╝╚██████╗██║  ██╗███████╗███████╗██║ ╚████║███████║
╚═════╝  ╚═════╝  ╚═════╝╚═╝  ╚═╝╚══════╝╚══════╝╚═╝  ╚═══╝╚══════╝`
