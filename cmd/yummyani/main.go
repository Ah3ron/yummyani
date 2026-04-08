// Package main is the entry point for the YummyAnime CLI player.
//
// It wires up all dependencies (API client, Kodik parser, mpv player)
// and launches the BubbleTea TUI program. No logic lives here — only
// composition and configuration.
package main

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"yummyani/internal/api"
	"yummyani/internal/config"
	"yummyani/internal/kodik"
	"yummyani/internal/player"
	"yummyani/internal/tui"
)

const version = "2.0.0"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-h", "--help":
			printUsage()
			return
		case "-v", "--version":
			fmt.Printf("yummyani %s\n", version)
			return
		}
	}

	cfg := config.Default()

	// Compose dependencies.
	animeClient := api.NewClient(cfg.APIBaseURL, cfg.APITimeout)

	kodikParser := kodik.NewParser(
		kodik.WithUserAgent(cfg.KodikUserAgent),
		kodik.WithMaxAttempts(cfg.KodikMaxAttempts),
	)

	mpvPlayer := player.NewMPV(cfg.PlayerCommand)

	// Create TUI model with injected dependencies.
	ctx := context.Background()
	m := tui.NewModel(
		ctx,
		animeClient,
		kodikParser,
		mpvPlayer,
		tui.WithSearchLimit(cfg.SearchLimit),
		tui.WithMaxVisibleLines(cfg.MaxVisibleLines),
	)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprint(os.Stderr, `
YummyAnime Player — TUI клиент для поиска и просмотра аниме

Использование:
  yummyani                     Запустить TUI интерфейс
  yummyani -h, --help          Показать справку
  yummyani -v, --version       Версия

Требования:
  mpv                          Плеер для воспроизведения

Источник: https://api.yani.tv
`)
}
