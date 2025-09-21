package main

import (
	"fmt"
	"log"
	"os"

	"github.com/CTSDM/motbwa-tui/internal/api"
	"github.com/CTSDM/motbwa-tui/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
		return
	}

	PORT_NUMBER := os.Getenv("PORT")

	// create state to hold the api information
	state := api.State{
		Server: api.ServerInfo{
			BaseURL:      fmt.Sprintf("http://localhost:%s/", PORT_NUMBER),
			WebsocketURL: fmt.Sprintf("ws://localhost:%s/ws", PORT_NUMBER),
			Login:        "api/login",
			Users:        "api/users",
		},
	}

	// start bubbletea
	p := tea.NewProgram(ui.InitialModel(state))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
