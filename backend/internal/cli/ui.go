package cli

import (
	"bufio"
	"fmt"
	"mochi_korov/internal/game"
	"strconv"
	"strings"
)

func StartCardChoice(g *game.Game, reader *bufio.Reader) {
	p := g.Current()
	fmt.Printf("Добро пожаловать, %s! У вас %d монет.\n", p.Name, p.Money)
	fmt.Println("Выберите стартовую карту (бесплатно):")
	fmt.Println("  1 — Пшеничное поле (1 монета, на 1)")
	fmt.Println("  2 — Пекарня (1 монета, на 2)")
	fmt.Print("Ваш выбор (1-2): ")

	choice := readInt(reader)
	cardID := "wheat_field"
	if choice == 2 {
		cardID = "bakery"
	}

	for _, item := range g.AvailableMarket() {
		if item.Card.ID == cardID {
			for mi := range g.Market {
				if g.Market[mi].Card.ID == cardID && g.Market[mi].Count > 0 {
					g.Market[mi].Count--
					p.BuyCard(g.Market[mi].Card)
					p.Money += int(item.Card.Price)
					fmt.Printf("Получено: %s\n", item.Card.Name)
					break
				}
			}
			break
		}
	}
	fmt.Println()
}

func HumanTurn(g *game.Game, reader *bufio.Reader) {
	p := g.Current()

	fmt.Println("  Нажмите Enter, чтобы бросить кубики...")
	reader.ReadString('\n')

	result, err := g.Roll()
	if err != nil {
		fmt.Println("  Ошибка:", err)
		return
	}
	fmt.Printf("  Выпало: %s\n", FormatDice(result))

	if p.CanReroll() {
		fmt.Print("  Перебросить? (д/Н): ")
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "д" || line == "Д" || line == "y" || line == "Y" {
			result, err = g.Reroll()
			if err == nil {
				fmt.Printf("  Переброс: %s\n", FormatDice(result))
			}
		}
	}

	actLog := g.ActivateCards()
	for _, l := range actLog {
		fmt.Println("  " + l)
	}
	fmt.Printf("  Денег после дохода: %d\n", p.Money)

	for {
		fmt.Println()
		fmt.Println("  Что купить?")
		fmt.Println("  0 — ничего, пропустить")

		market := g.AvailableMarket()
		for i, item := range market {
			fmt.Printf("  %d — карта: %s (%d монет) [шт: %d]\n", i+1, item.Card.Name, item.Card.Price, item.Count)
		}

		landmarks := g.AvailableLandmarks()
		for i, lm := range landmarks {
			fmt.Printf("  %d — достопримечательность: %s (%d монет)\n", len(market)+i+1, lm.Name, lm.Price)
		}

		total := len(market) + len(landmarks)
		fmt.Printf("  Ваш выбор (0-%d): ", total)
		choice := readInt(reader)

		if choice <= 0 {
			if choice == 0 {
				fmt.Println("  Ход пропущен")
			} else {
				fmt.Println("  Неверный ввод. Ход пропущен.")
			}
			g.EndTurn()
			break
		}

		if choice >= 1 && choice <= len(market) {
			for mi := range g.Market {
				if g.Market[mi].Card.ID == market[choice-1].Card.ID && g.Market[mi].Count > 0 {
					err := g.BuyCardFromMarket(mi)
					if err != nil {
						fmt.Println("  Ошибка:", err)
					} else {
						fmt.Printf("  Куплено: %s\n", market[choice-1].Card.Name)
						g.EndTurn()
						return
					}
					break
				}
			}
			continue
		}

		lmIdx := choice - len(market) - 1
		if lmIdx >= 0 && lmIdx < len(landmarks) {
			err := g.BuyLandmark(landmarks[lmIdx].ID)
			if err != nil {
				fmt.Println("  Ошибка:", err)
				continue
			}
			fmt.Printf("  Куплено: %s\n", landmarks[lmIdx].Name)
			g.EndTurn()
			return
		}

		fmt.Println("  Неверный ввод")
	}
}

func PrintStatus(g *game.Game) {
	fmt.Println()
	fmt.Println("── СТАТУС ИГРОКОВ ──")
	for _, p := range g.Players {
		count := 0
		for _, lm := range p.Landmarks {
			if lm.Price > 0 {
				count++
			}
		}
		fmt.Printf("  %s: %d монет, карт: %d, достопримечательностей: %d/%d\n",
			p.Name, p.Money, len(p.Cards), count, len(game.AllLandmarks()))
	}
	fmt.Println()
}

func FormatDice(result game.DiceResult) string {
	s := ""
	for i, n := range result.Numbers {
		if i > 0 {
			s += ", "
		}
		s += fmt.Sprintf("%d", n)
	}
	return s + fmt.Sprintf(" (сумма: %d)", result.Sum)
}

func readInt(reader *bufio.Reader) int {
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	n, err := strconv.Atoi(line)
	if err != nil {
		return -1
	}
	return n
}
