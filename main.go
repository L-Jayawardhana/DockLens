package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// --- Styling ---
// This is where Lip Gloss shines. We define our colors and styles here.
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FFFF")). // Cyan
			MarginBottom(1)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00")). // Green text
				Bold(true)

	unselectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")) // White text
)

// --- Model ---
// This holds the state of our application.
type model struct {
	containers []types.Container
	cursor     int
	err        error
}

// --- Init ---
// Runs when the program starts. We return nil because we don't have initial I/O commands yet.
func (m model) Init() tea.Cmd {
	return nil
}

// --- Update ---
// The brain of the app. It handles keystrokes and updates the model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit // Exit the app
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.containers)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

// --- View ---
// Renders the UI based on the current state of the model.
func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error connecting to Docker: %v\n\nPress 'q' to quit.", m.err)
	}

	if len(m.containers) == 0 {
		return "No running containers found.\n\nPress 'q' to quit."
	}

	s := titleStyle.Render("🐳 Running Docker Containers") + "\n"

	// Iterate over our containers and render them
	for i, c := range m.containers {
		// Clean up the container name (Docker prepends a '/')
		name := strings.TrimPrefix(c.Names[0], "/")

		// Create the row text
		row := fmt.Sprintf("%s | %s | %s", c.ID[:12], c.Image, name)

		// Apply highlight if the cursor is pointing to this row
		if m.cursor == i {
			s += selectedItemStyle.Render("❯ "+row) + "\n"
		} else {
			s += unselectedItemStyle.Render("  "+row) + "\n"
		}
	}

	s += "\nPress j/k or up/down to move • Press q to quit\n"
	return s
}

func main() {
	// 1. Connect to the Docker Daemon
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Printf("Failed to create Docker client: %v\n", err)
		os.Exit(1)
	}
	defer cli.Close()

	// 2. Fetch running containers
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{})

	// 3. Initialize our model with the fetched data
	initialModel := model{
		containers: containers,
		err:        err,
	}

	// 4. Start the Bubble Tea program
	p := tea.NewProgram(initialModel)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
