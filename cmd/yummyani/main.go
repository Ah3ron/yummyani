// Package main is the entry point for the YummyAnime CLI player.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"yummyani/internal/tui"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-h", "--help":
			printUsage()
			return
		case "-v", "--version":
			fmt.Println("yummyani 1.0.0")
			return
		}
	}

	p := tea.NewProgram(tui.NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `
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
