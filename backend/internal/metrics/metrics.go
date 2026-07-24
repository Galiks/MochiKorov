package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	GamesStarted = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "mochikorov_games_started_total",
		Help: "Total number of games started.",
	})

	GamesCompleted = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "mochikorov_games_completed_total",
		Help: "Total number of games completed.",
	})

	TurnsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "mochikorov_turns_total",
		Help: "Total number of turns played.",
	})

	RollsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "mochikorov_rolls_total",
		Help: "Total number of dice rolls.",
	})

	RerollsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "mochikorov_rerolls_total",
		Help: "Total number of dice rerolls.",
	})

	CardsBoughtTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mochikorov_cards_bought_total",
			Help: "Cards bought by card ID.",
		},
		[]string{"card_id", "card_name"},
	)

	LandmarksBoughtTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mochikorov_landmarks_bought_total",
			Help: "Landmarks bought by landmark ID.",
		},
		[]string{"landmark_id", "landmark_name"},
	)

	PlayerMoney = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mochikorov_player_money",
			Help: "Current money per player per session.",
		},
		[]string{"session_id", "player_id", "player_name"},
	)

	ActiveGames = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "mochikorov_active_games",
		Help: "Number of currently active games.",
	})
)

func Register() {
	prometheus.MustRegister(
		GamesStarted, GamesCompleted, TurnsTotal,
		RollsTotal, RerollsTotal,
		CardsBoughtTotal, LandmarksBoughtTotal,
		PlayerMoney, ActiveGames,
	)
}
