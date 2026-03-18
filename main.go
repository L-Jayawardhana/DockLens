package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// --- Styling ---
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FFFF")). // Cyan
			MarginBottom(1)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00")). // Green
				Bold(true)

	unselectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")) // White
)

// --- Custom Messages ---
// Bubble Tea uses these to pass data around asynchronously.
type tickMsg time.Time
type containersMsg []types.Container
type errMsg struct{ err error }

// --- Model ---
type model struct {
	cli        *client.Client // Store the client in the model so we can reuse it
	containers []types.Container
	cursor     int
	err        error
}

// --- Commands ---

// doTick waits for 2 seconds, then sends a tickMsg
func doTick() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// fetchContainers reaches out to Docker and returns a containersMsg
func fetchContainers(cli *client.Client) tea.Cmd {
	return func() tea.Msg {
		containers, err := cli.ContainerList(context.Background(), container.ListOptions{})
		if err != nil {
			return errMsg{err}
		}
		return containersMsg(containers)
	}
}

// --- Init ---
func (m model) Init() tea.Cmd {
	// On startup, immediately fetch containers.
	return fetchContainers(m.cli)
}

// --- Update ---
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Handle keystrokes
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.containers)-1 {
				m.cursor++
			}
		}

	// Handle data arriving from Docker
	case containersMsg:
		m.containers = msg
		// Adjust cursor if a container disappeared and cursor is now out of bounds
		if m.cursor >= len(m.containers) && len(m.containers) > 0 {
			m.cursor = len(m.containers) - 1
		}
		// Data received, queue up the next tick
		return m, doTick()

	// Handle the tick (time to refresh)
	case tickMsg:
		// Trigger the fetch command again
		return m, fetchContainers(m.cli)

	// Handle errors
	case errMsg:
		m.err = msg.err
		return m, nil
	}

	return m, nil
}

func main() {
	// Connect to Docker
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Printf("Failed to create Docker client: %v\n", err)
		os.Exit(1)
	}
	defer cli.Close()

	// Initialize the model with the client (empty containers for now, Init() will fetch them)
	m := model{
		cli: cli,
	}

	// Start the Bubble Tea program
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
