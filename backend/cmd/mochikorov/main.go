package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"mochi_korov/internal/api"
	"mochi_korov/internal/cli"
	"mochi_korov/internal/game"
	"mochi_korov/internal/store"
)

func main() {
	godotenv.Load()

	dsnDefault := os.Getenv("DATABASE_URL")
	if dsnDefault == "" {
		dsnDefault = "postgres://mochikorov:mochikorov@localhost:5432/mochikorov"
	}

	mode := flag.String("mode", "cli", "run mode: cli or server")
	addr := flag.String("addr", ":8080", "server address (used with -mode server)")
	sessionName := flag.String("session", "", "session name for CLI mode")
	cardSetName := flag.String("cards", "base", "card set name")
	dsn := flag.String("dsn", dsnDefault, "PostgreSQL DSN")
	flag.Parse()

	st, err := store.NewPostgresStore(context.Background(), *dsn)
	if err != nil {
		log.Fatalf("Failed to connect to storage: %v", err)
	}
	defer st.Close()

	if err := st.SeedDefaults(); err != nil {
		log.Printf("Warning: seed defaults: %v", err)
	}

	loadCardSet(st, *cardSetName)

	switch *mode {
	case "server":
		svr := api.NewServer(*addr, st)
		if err := svr.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	fmt.Println("=== МАЧИ КОРО ===")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	var g *game.Game
	sessID := *sessionName

	if sessID == "" {
		sessions, err := st.ListSessions()
		if err == nil && len(sessions) > 0 {
			fmt.Println("Доступные сессии:")
			for i, s := range sessions {
				fmt.Printf("  %d — %s (%s)\n", i+1, s.Name, s.ID)
			}
			fmt.Print("Введите номер сессии или имя новой: ")
			line, _ := reader.ReadString('\n')
			line = line[:len(line)-1]
			if line != "" {
				sessID = line
			}
		}
		if sessID == "" {
			sessID = "default"
		}

		existing, err := st.GetSession(sessID)
		if err == nil && existing != nil {
			data, err := st.LoadGameData(sessID)
			if err == nil && data != nil {
				g = game.GameFromSaveData(data)
				fmt.Printf("Сессия %q загружена.\n", sessID)
				fmt.Println()
			}
		}
	}

	if g == nil {
		_, err := st.CreateSession(sessID, sessID)
		if err != nil {
			log.Printf("Warning: create session: %v", err)
		}
		g = game.NewGame([]string{"Игрок", "Бот 1", "Бот 2", "Бот 3"})
		cli.StartCardChoice(g, reader)
	}

	for {
		current := g.Current()
		fmt.Println("──────────────────────────────")
		fmt.Printf("Ход %d | Ходит: %s | Деньги: %d | Сессия: %s\n", g.Turn+1, current.Name, current.Money, sessID)
		fmt.Println()

		if current.ID == 0 {
			cli.HumanTurn(g, reader)
		} else {
			log := g.DoAITurn()
			for _, l := range log {
				fmt.Println("  " + l)
			}
		}

		if err := st.SaveGameData(sessID, g.ToSaveData()); err != nil {
			log.Printf("Warning: save game: %v", err)
		}

		if winner := g.CheckWin(); winner != nil {
			fmt.Println()
			fmt.Printf("ПОБЕДИТЕЛЬ: %s!\n", winner.Name)
			st.DeleteSession(sessID)
			break
		}

		cli.PrintStatus(g)
	}
}

func loadCardSet(st store.Store, setName string) {
	cards, err := st.LoadCardSetCards(setName)
	if err != nil {
		log.Printf("Warning: card set %q not found, using defaults: %v", setName, err)
		return
	}

	game.RemoveAllCards()

	for _, c := range cards {
		if c.IsLandmark() {
			game.RegisterLandmark(c)
		} else {
			game.RegisterEstablishment(c)
		}
	}
}
