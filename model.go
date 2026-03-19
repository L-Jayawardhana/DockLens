package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/docker/docker/client"
)

// ─── Tabs ─────────────────────────────────────────────────────────────────────

type Tab int

const (
	TabContainers Tab = iota
	TabImages
	TabVolumes
	TabNetworks
	TabSystem
	tabCount
)

var tabNames = []string{"  Containers", "  Images", "  Volumes", "  Networks", "  System"}
var tabIcons = []string{"󰡨 ", "󰙔 ", "󱂵 ", "󰇉 ", "󰒋 "}

// ─── Model ────────────────────────────────────────────────────────────────────

type Model struct {
	// Layout
	width  int
	height int

	// Navigation
	activeTab     Tab
	selectedIndex int

	// Data
	containers []Container
	images     []Image
	volumes    []Volume
	networks   []Network
	systemInfo SystemInfo

	// Docker Client
	cli *client.Client

	// UI state
	showHelp        bool
	showLogo        bool
	showContextMenu bool
	contextMenuIdx  int

	// Scroll state
	detailScrollOffset int
	listScrollOffset   int
}

func NewModel() Model {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		// Fallback to empty model if client fails
		return Model{
			activeTab: TabContainers,
			showLogo:  true,
		}
	}

	return Model{
		activeTab: TabContainers,
		cli:       cli,
		showLogo:  true,
	}
}

// ─── Messages ─────────────────────────────────────────────────────────────────

type tickMsg struct{}
type dataMsg struct {
	containers []Container
	images     []Image
	volumes    []Volume
	networks   []Network
	system     SystemInfo
}
type errorMsg struct{ err error }
type hideSplashMsg struct{}

// ─── Init ─────────────────────────────────────────────────────────────────────

func tick() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tick(), m.fetchData(nil))
}

func (m Model) fetchData(prev []Container) tea.Cmd {
	return func() tea.Msg {
		if m.cli == nil {
			return errorMsg{err: fmt.Errorf("Docker client not initialized")}
		}

		c, _ := getContainers(m.cli, prev)
		i, _ := getImages(m.cli)
		v, _ := getVolumes(m.cli)
		n, _ := getNetworks(m.cli)
		s, _ := getSystemInfo(m.cli)

		return dataMsg{
			containers: c,
			images:     i,
			volumes:    v,
			networks:   n,
			system:     s,
		}
	}
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.showLogo && m.width > 0 {
			m.showLogo = false
		}

	case tea.KeyMsg:
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}
		if m.showContextMenu {
			return m.handleContextMenuKey(msg)
		}
		return m.handleKey(msg)

	case dataMsg:
		m.containers = msg.containers
		m.images = msg.images
		m.volumes = msg.volumes
		m.networks = msg.networks
		m.systemInfo = msg.system

	case errorMsg:
		// Logging could go here

	case tickMsg:
		// Refresh data from real docker client
		return m, tea.Batch(tick(), m.fetchData(m.containers))
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	listLen := m.currentListLen()

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "?":
		m.showHelp = !m.showHelp

	// Tab switching
	case "tab", "l", "right":
		m.activeTab = (m.activeTab + 1) % tabCount
		m.selectedIndex = 0
		m.listScrollOffset = 0
		m.detailScrollOffset = 0

	case "shift+tab", "h", "left":
		m.activeTab = (m.activeTab - 1 + tabCount) % tabCount
		m.selectedIndex = 0
		m.listScrollOffset = 0
		m.detailScrollOffset = 0

	case "1":
		m.activeTab = TabContainers
		m.selectedIndex = 0
	case "2":
		m.activeTab = TabImages
		m.selectedIndex = 0
	case "3":
		m.activeTab = TabVolumes
		m.selectedIndex = 0
	case "4":
		m.activeTab = TabNetworks
		m.selectedIndex = 0
	case "5":
		m.activeTab = TabSystem
		m.selectedIndex = 0

	// List navigation
	case "j", "down":
		if m.selectedIndex < listLen-1 {
			m.selectedIndex++
			m.detailScrollOffset = 0
		}

	case "k", "up":
		if m.selectedIndex > 0 {
			m.selectedIndex--
			m.detailScrollOffset = 0
		}

	case "g":
		m.selectedIndex = 0
		m.listScrollOffset = 0

	case "G":
		m.selectedIndex = listLen - 1

	// Detail panel scroll
	case "J":
		m.detailScrollOffset++

	case "K":
		if m.detailScrollOffset > 0 {
			m.detailScrollOffset--
		}

	// Context menu
	case "enter", " ":
		if m.activeTab != TabSystem && listLen > 0 {
			m.showContextMenu = true
			m.contextMenuIdx = 0
		}
	}

	// Adjust list scroll
	visibleRows := m.listVisibleRows()
	if m.selectedIndex >= m.listScrollOffset+visibleRows {
		m.listScrollOffset = m.selectedIndex - visibleRows + 1
	}
	if m.selectedIndex < m.listScrollOffset {
		m.listScrollOffset = m.selectedIndex
	}

	return m, nil
}

func (m Model) handleContextMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	actions := m.contextMenuActions()
	switch msg.String() {
	case "j", "down":
		if m.contextMenuIdx < len(actions)-1 {
			m.contextMenuIdx++
		}
	case "k", "up":
		if m.contextMenuIdx > 0 {
			m.contextMenuIdx--
		}
	case "esc", "q":
		m.showContextMenu = false
	case "enter":
		m.showContextMenu = false
		// Action handling would trigger Docker API calls here
	}
	return m, nil
}

func (m Model) currentListLen() int {
	switch m.activeTab {
	case TabContainers:
		return len(m.containers)
	case TabImages:
		return len(m.images)
	case TabVolumes:
		return len(m.volumes)
	case TabNetworks:
		return len(m.networks)
	}
	return 0
}

func (m Model) listVisibleRows() int {
	// Approximate: height - header - tabs - footer - borders
	return m.height - 10
}

func (m Model) contextMenuActions() []string {
	switch m.activeTab {
	case TabContainers:
		return []string{"  View Logs", "  Start", "  Stop", "  Restart", "  Kill", "  Stats", "  Inspect", "  Remove"}
	case TabImages:
		return []string{"  Run Container", "  Remove Image", "  View History", "  Inspect"}
	case TabVolumes:
		return []string{"  Inspect", "  Remove Volume"}
	case TabNetworks:
		return []string{"  Inspect", "  Remove Network"}
	}
	return nil
}

// ─── View ─────────────────────────────────────────────────────────────────────

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	if m.showHelp {
		return m.renderHelp()
	}

	header := m.renderHeader()
	tabBar := m.renderTabBar()
	body := m.renderBody()
	footer := m.renderFooter()

	view := lipgloss.JoinVertical(lipgloss.Left,
		header,
		tabBar,
		body,
		footer,
	)

	if m.showContextMenu {
		// Overlay context menu (simple append for now)
		_ = view
		return m.renderWithContextMenu(header, tabBar, body, footer)
	}

	return view
}

func (m Model) renderHeader() string {
	title := logoStyle.Render("  DOCKERLENS")
	version := lipgloss.NewStyle().Foreground(colorGray).Render("v1.0.0")
	hostInfo := lipgloss.NewStyle().Foreground(colorDimCyan).Render("⬡ docker.sock  local")

	left := lipgloss.JoinHorizontal(lipgloss.Center, title, "  ", version)
	right := hostInfo

	padding := m.width - lipgloss.Width(left) - lipgloss.Width(right) - 4
	if padding < 0 {
		padding = 0
	}

	return headerStyle.Width(m.width).Render(
		left + strings.Repeat(" ", padding) + right,
	)
}

func (m Model) renderTabBar() string {
	var tabs []string

	for i := Tab(0); i < tabCount; i++ {
		label := tabNames[i]
		if i == m.activeTab {
			tabs = append(tabs, activeTabStyle.Render(label))
		} else {
			tabs = append(tabs, inactiveTabStyle.Render(label))
		}
	}

	bar := lipgloss.JoinHorizontal(lipgloss.Bottom, tabs...)
	return tabBarStyle.Width(m.width).Render(bar)
}

func (m Model) renderBody() string {
	bodyHeight := m.height - 7 // header + tabbar + footer + some padding

	leftWidth := m.width * 30 / 100
	rightWidth := m.width - leftWidth - 3

	leftContent := m.renderLeftPanel(leftWidth-4, bodyHeight-2)
	rightContent := m.renderRightPanel(rightWidth-4, bodyHeight-2)

	left := leftPanelStyle.
		Width(leftWidth).
		Height(bodyHeight).
		Render(leftContent)

	right := rightPanelStyle.
		Width(rightWidth).
		Height(bodyHeight).
		Render(rightContent)

	return lipgloss.JoinHorizontal(lipgloss.Top, left, " ", right)
}

func (m Model) renderLeftPanel(w, h int) string {
	var lines []string

	title := panelTitleStyle.Render(tabNames[m.activeTab])
	lines = append(lines, title)
	lines = append(lines, dividerStyle.Render(strings.Repeat("─", w)))

	switch m.activeTab {
	case TabContainers:
		lines = append(lines, m.renderContainerList(w, h-3)...)
	case TabImages:
		lines = append(lines, m.renderImageList(w, h-3)...)
	case TabVolumes:
		lines = append(lines, m.renderVolumeList(w, h-3)...)
	case TabNetworks:
		lines = append(lines, m.renderNetworkList(w, h-3)...)
	case TabSystem:
		lines = append(lines, m.renderSystemList(w, h-3)...)
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderRightPanel(w, h int) string {
	switch m.activeTab {
	case TabContainers:
		return m.renderContainerDetail(w, h)
	case TabImages:
		return m.renderImageDetail(w, h)
	case TabVolumes:
		return m.renderVolumeDetail(w, h)
	case TabNetworks:
		return m.renderNetworkDetail(w, h)
	case TabSystem:
		return m.renderSystemDashboard(w, h)
	}
	return ""
}

// ─── Container List ───────────────────────────────────────────────────────────

func (m Model) renderContainerList(w, h int) []string {
	var lines []string
	visibleItems := h

	for i, c := range m.containers {
		if i < m.listScrollOffset {
			continue
		}
		if len(lines) >= visibleItems {
			break
		}

		statusDot := statusDot(c.Status)
		name := truncate(c.Name, w-12)
		line := fmt.Sprintf("%s %-*s", statusDot, w-10, name)

		if i == m.selectedIndex {
			lines = append(lines, selectedItemStyle.Width(w).Render("▶ "+line))
		} else {
			lines = append(lines, listItemStyle.Width(w).Render("  "+line))
		}
	}
	return lines
}

// ─── Image List ───────────────────────────────────────────────────────────────

func (m Model) renderImageList(w, h int) []string {
	var lines []string
	for i, img := range m.images {
		if i < m.listScrollOffset {
			continue
		}
		if len(lines) >= h {
			break
		}
		name := truncate(fmt.Sprintf("%s:%s", img.Name, img.Tag), w-10)
		size := formatBytes(img.Size)
		sizeStr := lipgloss.NewStyle().Foreground(colorGray).Render(size)
		line := fmt.Sprintf("%-*s %s", w-len(size)-4, name, sizeStr)
		if i == m.selectedIndex {
			lines = append(lines, selectedItemStyle.Width(w).Render("▶ "+line))
		} else {
			lines = append(lines, listItemStyle.Width(w).Render("  "+line))
		}
	}
	return lines
}

// ─── Volume List ──────────────────────────────────────────────────────────────

func (m Model) renderVolumeList(w, h int) []string {
	var lines []string
	for i, v := range m.volumes {
		if i < m.listScrollOffset {
			continue
		}
		if len(lines) >= h {
			break
		}
		name := truncate(v.Name, w-10)
		size := lipgloss.NewStyle().Foreground(colorGray).Render(v.Size)
		isOrphaned := !strings.HasPrefix(v.Name, "orphaned")
		_ = isOrphaned
		line := fmt.Sprintf("%-*s %s", w-len(v.Size)-4, name, size)
		if strings.HasPrefix(v.Name, "orphaned") {
			line = lipgloss.NewStyle().Foreground(colorYellow).Render("⚠ ") + line
		} else {
			line = "  " + line
		}
		if i == m.selectedIndex {
			lines = append(lines, selectedItemStyle.Width(w).Render("▶ "+strings.TrimPrefix(line, "  ")))
		} else {
			lines = append(lines, listItemStyle.Width(w).Render(line))
		}
	}
	return lines
}

// ─── Network List ─────────────────────────────────────────────────────────────

func (m Model) renderNetworkList(w, h int) []string {
	var lines []string
	for i, n := range m.networks {
		if i < m.listScrollOffset {
			continue
		}
		if len(lines) >= h {
			break
		}
		name := truncate(n.Name, w-10)
		driver := lipgloss.NewStyle().Foreground(colorGray).Render(n.Driver)
		line := fmt.Sprintf("%-*s %s", w-len(n.Driver)-4, name, driver)
		if i == m.selectedIndex {
			lines = append(lines, selectedItemStyle.Width(w).Render("▶ "+line))
		} else {
			lines = append(lines, listItemStyle.Width(w).Render("  "+line))
		}
	}
	return lines
}

// ─── System List (mini nav) ───────────────────────────────────────────────────

func (m Model) renderSystemList(w, h int) []string {
	items := []string{"  Docker Daemon Info", "  Disk Usage (df)", "  System Prune"}
	var lines []string
	for i, item := range items {
		if i == m.selectedIndex {
			lines = append(lines, selectedItemStyle.Width(w).Render("▶ "+strings.TrimSpace(item)))
		} else {
			lines = append(lines, listItemStyle.Width(w).Render(item))
		}
	}
	_ = h
	return lines
}

// ─── Container Detail ─────────────────────────────────────────────────────────

func (m Model) renderContainerDetail(w, h int) string {
	if len(m.containers) == 0 {
		return ""
	}

	var sb strings.Builder

	// ─── Summary Table of All Containers ─────────────────────────────────────
	sb.WriteString(detailSectionStyle.Render("  ALL RUNNING CONTAINERS") + "\n")

	tableHeader := fmt.Sprintf("  %-25s %-12s %-10s %-10s", "NAME", "STATUS", "CPU", "MEM")
	sb.WriteString(lipgloss.NewStyle().Foreground(colorGray).Bold(true).Render(tableHeader) + "\n")
	sb.WriteString(dividerStyle.Render(strings.Repeat("─", w)) + "\n")

	for i, c := range m.containers {
		status := string(c.Status)
		if c.Status == StatusRunning {
			status = lipgloss.NewStyle().Foreground(colorGreen).Render("running")
		}

		memStr := fmt.Sprintf("%.1f MiB", c.Memory)
		cpuStr := fmt.Sprintf("%.1f%%", c.CPU)

		line := fmt.Sprintf("%-25s %-12s %-10s %-10s",
			truncate(c.Name, 24), status, cpuStr, memStr)

		if i == m.selectedIndex {
			sb.WriteString(lipgloss.NewStyle().Background(colorDimCyan).Foreground(colorBg).Render("  "+line) + "\n")
		} else {
			sb.WriteString("  " + line + "\n")
		}
	}
	sb.WriteString("\n")

	// ─── Selected Container Details ──────────────────────────────────────────
	c := m.containers[m.selectedIndex]
	sb.WriteString(detailSectionStyle.Render("  DETAILED INFO: "+c.Name) + "  " + renderStatusBadge(c.Status) + "\n")
	sb.WriteString(dividerStyle.Render(strings.Repeat("─", w)) + "\n")

	// Stats row
	cpuBar := renderProgressBar(c.CPU, 100, 20)
	memPct := 0.0
	if c.MemMax > 0 {
		memPct = c.Memory / c.MemMax * 100
	}
	memBar := renderProgressBar(memPct, 100, 20)

	sb.WriteString(fmt.Sprintf("  %s CPU   %s  %.1f%%\n",
		metricLabelStyle.Render(""), cpuBar, c.CPU))
	sb.WriteString(fmt.Sprintf("  %s MEM   %s  %.1f / %.0f MiB\n\n",
		metricLabelStyle.Render(""), memBar, c.Memory, c.MemMax))

	// Info table
	rows := [][]string{
		{"ID", c.ID},
		{"Image", c.Image},
		{"Ports", c.Ports},
		{"Network", c.Network},
		{"Created", timeAgo(c.Created)},
	}
	for _, row := range rows {
		key := detailKeyStyle.Render(row[0])
		val := detailValStyle.Render(row[1])
		sb.WriteString(fmt.Sprintf("  %s  %s\n", key, val))
	}

	// Logs
	sb.WriteString("\n" + detailSectionStyle.Render("  RECENT LOGS") + "\n")
	sb.WriteString(dividerStyle.Render(strings.Repeat("─", w)) + "\n")

	logStart := m.detailScrollOffset
	if logStart > len(c.Logs) {
		logStart = len(c.Logs)
	}

	totalVisibleLogs := h - 18 // Estimate remaining space
	if totalVisibleLogs < 0 {
		totalVisibleLogs = 0
	}

	logsToShow := c.Logs[logStart:]
	if len(logsToShow) > totalVisibleLogs && totalVisibleLogs > 0 {
		logsToShow = logsToShow[:totalVisibleLogs]
	}

	for _, log := range logsToShow {
		ts := logTimestampStyle.Render(log.Timestamp)
		text := logTextStyle.Render(truncate(log.Text, w-22))
		sb.WriteString(fmt.Sprintf("  %s  %s\n", ts, text))
	}

	// Key hints
	sb.WriteString("\n" + dividerStyle.Render(strings.Repeat("─", w)) + "\n")
	sb.WriteString(keyDescStyle.Render("  [enter] actions  [J/K] scroll logs"))

	return sb.String()
}

// ─── Image Detail ─────────────────────────────────────────────────────────────

func (m Model) renderImageDetail(w, h int) string {
	if len(m.images) == 0 {
		return ""
	}
	img := m.images[m.selectedIndex]

	var sb strings.Builder
	sb.WriteString(panelTitleStyle.Render(fmt.Sprintf("%s:%s", img.Name, img.Tag)) + "\n")
	sb.WriteString(dividerStyle.Render(strings.Repeat("─", w)) + "\n\n")

	sb.WriteString(detailSectionStyle.Render("  IMAGE INFO") + "\n")
	rows := [][]string{
		{"ID", img.ID},
		{"Name", img.Name},
		{"Tag", img.Tag},
		{"Size", formatBytes(img.Size)},
		{"Created", timeAgo(img.Created)},
	}
	for _, row := range rows {
		key := detailKeyStyle.Render(row[0])
		val := detailValStyle.Render(row[1])
		sb.WriteString(fmt.Sprintf("  %s  %s\n", key, val))
	}

	sb.WriteString("\n" + detailSectionStyle.Render("  LAYERS (HISTORY)") + "\n")
	sb.WriteString(dividerStyle.Render(strings.Repeat("─", w)) + "\n")
	// Simulated layer history
	layers := []string{
		"ADD file:  # base layer",
		"RUN /bin/sh -c set -eux; ...  (12 MB)",
		"RUN /bin/sh -c addgroup ...   (1.2 KB)",
		"COPY --chown=...              (8.4 MB)",
		"ENV LANG=C.UTF-8",
		"EXPOSE 80/tcp",
		"ENTRYPOINT [\"/docker-entrypoint.sh\"]",
		"CMD [\"nginx\", \"-g\", \"daemon off;\"]",
	}
	for _, l := range layers {
		sb.WriteString(logTextStyle.Render("  "+truncate(l, w-4)) + "\n")
	}

	sb.WriteString("\n" + keyDescStyle.Render("  [enter] actions"))
	_ = h
	return sb.String()
}

// ─── Volume Detail ────────────────────────────────────────────────────────────

func (m Model) renderVolumeDetail(w, h int) string {
	if len(m.volumes) == 0 {
		return ""
	}
	v := m.volumes[m.selectedIndex]

	var sb strings.Builder
	sb.WriteString(panelTitleStyle.Render(v.Name) + "\n")
	sb.WriteString(dividerStyle.Render(strings.Repeat("─", w)) + "\n\n")

	if strings.HasPrefix(v.Name, "orphaned") {
		warning := lipgloss.NewStyle().
			Foreground(colorYellow).
			Bold(true).
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(colorYellow).
			Padding(0, 1).
			Render("⚠  ORPHANED VOLUME — No container is using this volume.\n     Consider removing it to free disk space.")
		sb.WriteString(warning + "\n\n")
	}

	sb.WriteString(detailSectionStyle.Render("  VOLUME INFO") + "\n")
	rows := [][]string{
		{"Driver", v.Driver},
		{"Size", v.Size},
		{"Created", timeAgo(v.Created)},
		{"Mount Point", truncate(v.MountPoint, w-22)},
	}
	for _, row := range rows {
		key := detailKeyStyle.Render(row[0])
		val := detailValStyle.Render(row[1])
		sb.WriteString(fmt.Sprintf("  %s  %s\n", key, val))
	}

	sb.WriteString("\n" + keyDescStyle.Render("  [enter] actions"))
	_ = h
	return sb.String()
}

// ─── Network Detail ───────────────────────────────────────────────────────────

func (m Model) renderNetworkDetail(w, h int) string {
	if len(m.networks) == 0 {
		return ""
	}
	n := m.networks[m.selectedIndex]

	var sb strings.Builder
	sb.WriteString(panelTitleStyle.Render(n.Name) + "\n")
	sb.WriteString(dividerStyle.Render(strings.Repeat("─", w)) + "\n\n")

	sb.WriteString(detailSectionStyle.Render("  NETWORK INFO") + "\n")
	rows := [][]string{
		{"ID", n.ID},
		{"Driver", n.Driver},
		{"Scope", n.Scope},
		{"Subnet", n.Subnet},
	}
	for _, row := range rows {
		key := detailKeyStyle.Render(row[0])
		val := detailValStyle.Render(row[1])
		sb.WriteString(fmt.Sprintf("  %s  %s\n", key, val))
	}

	sb.WriteString("\n" + detailSectionStyle.Render("  CONNECTED CONTAINERS") + "\n")
	sb.WriteString(dividerStyle.Render(strings.Repeat("─", w)) + "\n")
	if len(n.Containers) == 0 {
		sb.WriteString(lipgloss.NewStyle().Foreground(colorGray).Render("  (none)") + "\n")
	} else {
		for _, cn := range n.Containers {
			sb.WriteString(lipgloss.NewStyle().Foreground(colorGreen).Render("  ● ") +
				detailValStyle.Render(cn) + "\n")
		}
	}

	sb.WriteString("\n" + keyDescStyle.Render("  [enter] actions"))
	_ = h
	return sb.String()
}

// ─── System Dashboard ─────────────────────────────────────────────────────────

func (m Model) renderSystemDashboard(w, h int) string {
	s := m.systemInfo
	var sb strings.Builder

	sb.WriteString(panelTitleStyle.Render("  System Overview") + "\n")
	sb.WriteString(dividerStyle.Render(strings.Repeat("─", w)) + "\n\n")

	// Stat cards row
	cards := []struct{ label, value string }{
		{"CONTAINERS", fmt.Sprintf("%d / %d up", s.ContainersUp, s.Containers)},
		{"IMAGES", fmt.Sprintf("%d total", s.Images)},
		{"DISK USED", s.DiskUsed},
		{"RECLAIMABLE", s.ReclaimableSize},
	}

	var cardParts []string
	cardW := (w - 3) / 4
	for _, c := range cards {
		box := dashboardBoxStyle.Width(cardW).Render(
			lipgloss.NewStyle().Foreground(colorGray).Render(c.label) + "\n" +
				metricValueGreenStyle.Render(c.value),
		)
		cardParts = append(cardParts, box)
	}
	sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, cardParts...) + "\n\n")

	// Daemon info
	sb.WriteString(detailSectionStyle.Render("  DAEMON INFO") + "\n")
	rows := [][]string{
		{"Docker Version", s.DockerVersion},
		{"API Version", s.APIVersion},
		{"OS / Arch", fmt.Sprintf("%s / %s", s.OS, s.Arch)},
		{"Kernel", s.Kernel},
		{"CPUs", fmt.Sprintf("%d", s.CPUs)},
		{"Total Memory", s.TotalMemory},
		{"Storage Driver", s.StorageDriver},
	}
	for _, row := range rows {
		key := detailKeyStyle.Render(row[0])
		val := detailValStyle.Render(row[1])
		sb.WriteString(fmt.Sprintf("  %s  %s\n", key, val))
	}

	// Disk usage
	sb.WriteString("\n" + detailSectionStyle.Render("  DISK USAGE") + "\n")
	diskRows := [][]string{
		{"Images", s.ImagesSize},
		{"Volumes", s.VolumesSize},
		{"Build Cache", s.BuildCacheSize},
		{"Total Used", s.DiskUsed},
	}
	for _, row := range diskRows {
		key := detailKeyStyle.Render(row[0])
		val := metricValueStyle.Render(row[1])
		sb.WriteString(fmt.Sprintf("  %s  %s\n", key, val))
	}

	// Prune warning
	sb.WriteString("\n")
	pruneBtn := lipgloss.NewStyle().
		Foreground(colorRed).
		Bold(true).
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(colorRed).
		Padding(0, 2).
		Render("  System Prune  [p]  — removes all unused data")
	sb.WriteString(pruneBtn)

	_ = h
	return sb.String()
}

// ─── Context Menu ─────────────────────────────────────────────────────────────

func (m Model) renderWithContextMenu(header, tabBar, body, footer string) string {
	actions := m.contextMenuActions()

	var menuLines []string
	for i, a := range actions {
		if i == m.contextMenuIdx {
			menuLines = append(menuLines, contextMenuSelectedStyle.Render("▶ "+a))
		} else {
			menuLines = append(menuLines, contextMenuItemStyle.Render("  "+a))
		}
	}

	menuContent := panelTitleStyle.Render("Actions") + "\n" +
		dividerStyle.Render(strings.Repeat("─", 24)) + "\n" +
		strings.Join(menuLines, "\n") + "\n\n" +
		keyDescStyle.Render("[esc] close")

	menu := contextMenuStyle.Render(menuContent)

	// Place menu overlaid on body
	bodyWithMenu := lipgloss.JoinHorizontal(lipgloss.Top,
		body[:len(body)/2],
		menu,
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		tabBar,
		bodyWithMenu,
		footer,
	)
}

// ─── Footer ───────────────────────────────────────────────────────────────────

func (m Model) renderFooter() string {
	keys := []struct{ key, desc string }{
		{"tab", "switch tab"},
		{"j/k", "navigate"},
		{"enter", "actions"},
		{"?", "help"},
		{"q", "quit"},
	}

	var parts []string
	for _, k := range keys {
		parts = append(parts, keyStyle.Render(" "+k.key+" ")+keyDescStyle.Render(" "+k.desc))
	}

	content := strings.Join(parts, "  ")
	right := lipgloss.NewStyle().Foreground(colorGray).Render(fmt.Sprintf("[%d/%d]",
		m.selectedIndex+1, m.currentListLen()))

	pad := m.width - lipgloss.Width(content) - lipgloss.Width(right) - 4
	if pad < 0 {
		pad = 0
	}

	return footerStyle.Width(m.width).Render(
		content + strings.Repeat(" ", pad) + right,
	)
}

// ─── Help Overlay ─────────────────────────────────────────────────────────────

func (m Model) renderHelp() string {
	sections := []struct {
		title string
		keys  [][]string
	}{
		{"Navigation", [][]string{
			{"tab / shift+tab", "Switch tabs forward/back"},
			{"1-5", "Jump to tab directly"},
			{"j / k", "Move up/down in list"},
			{"g / G", "Jump to first/last item"},
		}},
		{"Actions", [][]string{
			{"enter / space", "Open context menu"},
			{"J / K", "Scroll detail panel"},
			{"esc", "Close menu/overlay"},
		}},
		{"App", [][]string{
			{"?", "Toggle this help"},
			{"q / ctrl+c", "Quit"},
		}},
	}

	var sb strings.Builder
	sb.WriteString(logoStyle.Render("  DOCKERLENS") + "  " +
		lipgloss.NewStyle().Foreground(colorGray).Render("Keyboard Reference") + "\n\n")

	for _, section := range sections {
		sb.WriteString(detailSectionStyle.Render("  "+section.title) + "\n")
		for _, kv := range section.keys {
			k := lipgloss.NewStyle().Foreground(colorCyan).Width(24).Render(kv[0])
			d := lipgloss.NewStyle().Foreground(colorWhite).Render(kv[1])
			sb.WriteString(fmt.Sprintf("  %s  %s\n", k, d))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(keyDescStyle.Render("  Press any key to close"))

	content := helpOverlayStyle.Render(sb.String())

	// Center in terminal
	vPad := (m.height - lipgloss.Height(content)) / 2
	hPad := (m.width - lipgloss.Width(content)) / 2

	return strings.Repeat("\n", vPad) +
		strings.Repeat(" ", hPad) + content
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func statusDot(s ContainerStatus) string {
	switch s {
	case StatusRunning:
		return statusRunningStyle.Render("●")
	case StatusStopped:
		return statusStoppedStyle.Render("●")
	case StatusPaused:
		return statusPausedStyle.Render("●")
	default:
		return lipgloss.NewStyle().Foreground(colorGray).Render("●")
	}
}

func renderStatusBadge(s ContainerStatus) string {
	switch s {
	case StatusRunning:
		return lipgloss.NewStyle().
			Foreground(colorBg).Background(colorGreen).
			Bold(true).Padding(0, 1).Render("  RUNNING")
	case StatusStopped:
		return lipgloss.NewStyle().
			Foreground(colorBg).Background(colorRed).
			Bold(true).Padding(0, 1).Render("  STOPPED")
	case StatusPaused:
		return lipgloss.NewStyle().
			Foreground(colorBg).Background(colorYellow).
			Bold(true).Padding(0, 1).Render("  PAUSED")
	default:
		return lipgloss.NewStyle().
			Foreground(colorBg).Background(colorGray).
			Bold(true).Padding(0, 1).Render(string(s))
	}
}

func renderProgressBar(val, max float64, width int) string {
	if max == 0 {
		max = 1
	}
	filled := int(val / max * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	color := colorGreen
	if val/max > 0.7 {
		color = colorYellow
	}
	if val/max > 0.9 {
		color = colorRed
	}

	bar := lipgloss.NewStyle().Background(color).Render(strings.Repeat(" ", filled))
	bar += lipgloss.NewStyle().Background(colorBorder).Render(strings.Repeat(" ", empty))
	return "[" + bar + "]"
}

func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
