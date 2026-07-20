# vk-fasting

A command-line intermittent fasting tracker written in Go. Log and manage your fasting sessions through an interactive terminal shell, with data persisted locally.

## Tech Stack

- [Go](https://go.dev/)
- Local file-based storage (no external database required)

## Project Structure

```
├── DATABASES/FASTING/   # Local data storage for fasting records
├── pkg/
│   ├── cmd/               # Command handling / interactive shell logic
│   ├── config/             # Configuration (e.g. local storage file path)
│   ├── db/                 # Fastings data model and load/save logic
│   └── util/                # Utility helpers (local storage init, etc.)
├── main.go                # Entry point
├── go.mod
└── go.sum
```

## How It Works

On startup, `vk-fasting`:

1. Initializes local storage (`util.InitLocalStorage`).
2. Loads existing fasting data from a local file (`db.Fastings.LoadFromFile`).
3. Starts an interactive command shell (`cmd.CommandLine`) where you can log and manage fasting sessions.

## Getting Started

### Prerequisites

- Go installed

### Installation

```bash
git clone https://github.com/VkHyperNova/vk-fasting.git
cd vk-fasting
go build -o vk-fasting .
```

### Usage

Run the built binary:

```bash
./vk-fasting
```

Or run directly with Go:

```bash
go run main.go
```

This launches an interactive shell where you can start/log fasting sessions and track your history, persisted locally between runs.

## License

No license specified — all rights reserved by the author unless stated otherwise.
