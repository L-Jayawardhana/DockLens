# 🐳 DockLens

![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)
![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)
![License](https://img.shields.io/badge/License-MIT-blue.svg)

**DockLens** is a blazing-fast, lightweight Terminal UI (TUI) for monitoring and managing Docker containers.

Built entirely in Go, it interacts directly with the Docker daemon to provide real-time insights without the heavy resource footprint of traditional Electron-based desktop applications. It fits terminal-heavy workflows, tiling window managers, and remote cloud VMs.

---

## ✨ Features

### Current
* **Instant Visibility:** View all currently running containers at a glance.
* **Vim-Style Navigation:** Use `j`/`k` or arrow keys to move.
* **Zero Dependencies:** Single static binary once built.
* **Beautiful UI:** Styled with Charmbracelet's Lip Gloss.

### 🚀 Roadmap (Coming Soon)
- [ ] View real-time container logs.
- [ ] Start, stop, and restart containers from the UI.
- [ ] View local Docker images and their sizes.
- [ ] Monitor live container resource usage (CPU/Memory).
- [ ] Exec into a running container (`/bin/sh` or `/bin/bash`).

---

## 🛠️ Tech Stack
* **Language:** Go
* **TUI Framework:** Bubble Tea
* **Styling:** Lip Gloss
* **Docker API:** Docker Engine SDK for Go

---

## 🚀 Getting Started

### Prerequisites
* Go 1.25 or higher installed.
* Docker daemon running and your user allowed to access it (usually via `/var/run/docker.sock`).

### Installation
1. Clone the repository:
   ```bash
   git clone https://github.com/lakdinu/DockLens.git
   cd DockLens
   ```

2. Install Go (if needed):
   * Arch Linux: `sudo pacman -S go`
   * Debian/Ubuntu: `sudo apt-get install -y golang`
   * macOS (Homebrew): `brew install go`

3. Install dependencies:
   ```bash
   go mod tidy
   ```

4. Run DockLens:
   ```bash
   go run main.go
   ```

5. (Optional) Build a binary:
   ```bash
   go build -o docklens
   ./docklens
   ```

### Usage
- Navigate with `j`/`k` or arrow keys.
- Quit with `q` or `Ctrl+C`.

### Troubleshooting
- Ensure Docker is running and your user can access the Docker socket.
- If module downloads fail (e.g., EOF), re-run `go mod tidy`; transient network issues are common.
- If permissions are denied on the Docker socket, add your user to the `docker` group or run via `sudo` (as a last resort).

### Contributing
Issues and PRs are welcome. Please run `go fmt ./...` before opening a pull request.

### License
MIT
