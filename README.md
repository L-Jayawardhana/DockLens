# 🐳 DockLens

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)
![License](https://img.shields.io/badge/License-MIT-blue.svg)

**DockLens** is a blazing-fast, lightweight Terminal UI (TUI) for monitoring and managing Docker containers. 

Built entirely in Go, it interacts directly with the Docker daemon to provide real-time insights without the heavy resource footprint of traditional electron-based desktop applications. It's the perfect companion for terminal-heavy workflows, tiling window managers, and remote cloud VMs.

---

## ✨ Features

### Current
* **Instant Visibility:** View all currently running containers at a glance.
* **Vim-Style Navigation:** Seamlessly scroll through your containers using `j`/`k` or arrow keys.
* **Zero Dependencies:** Compiles to a single static binary. Drop it on any server and run it.
* **Beautiful UI:** Styled with Charmbracelet's Lip Gloss for a modern, colorful terminal experience.

### 🚀 Roadmap (Coming Soon)
- [ ] View real-time container logs.
- [ ] Start, stop, and restart containers directly from the UI.
- [ ] View local Docker images and their sizes.
- [ ] Monitor live container resource usage (CPU/Memory).
- [ ] Execute `/bin/sh` or `/bin/bash` inside a running container.

---

## 🛠️ Tech Stack

* **Language:** [Go (Golang)](https://go.dev/)
* **TUI Framework:** [Bubble Tea](https://github.com/charmbracelet/bubbletea)
* **Styling:** [Lip Gloss](https://github.com/charmbracelet/lipgloss)
* **Docker API:** [Docker Engine SDK for Go](https://docs.docker.com/engine/api/sdk/)

---

## 🚀 Getting Started

### Prerequisites
* Go 1.21 or higher installed on your system.
* A running Docker Daemon (`/var/run/docker.sock` accessible).

### Installation

1. **Clone the repository:**
   ```bash
   git clone [https://github.com/lakdinu/DockLens.git](https://github.com/lakdinu/DockLens.git)
   cd DockLens