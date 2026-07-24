import { useState } from 'react'
import { ICONS, CARD_COLORS, EFFECT_NAMES } from '../constants'

function groupCards(cards) {
  const grouped = {}
  for (const c of cards) {
    if (!grouped[c.id]) grouped[c.id] = { card: c, count: 0 }
    grouped[c.id].count++
  }
  return Object.values(grouped)
}

export default function PlayerList({ players, state }) {
  const [handVisible, setHandVisible] = useState(false)
  const human = players.find(p => p.id === 0)

  return (
    <aside className="panel-players">
      <h3>Игроки</h3>
      <div>
        {players.map((p) => {
          const cards = p.cards || []
          const grouped = groupCards(cards)
          return (
            <div key={p.id} className={`player-card ${p.is_current ? 'active' : ''}`}>
              <div className="p-name">{p.is_current ? '▶ ' : ''}{p.name}</div>
              <div className="p-money">💰 {p.money} монет</div>
              <div className="p-stats">
                🏛️ {p.landmark_count}/{state?.total_landmarks || 7} · 🃏 {cards.length} карт
                {p.can_roll_two_dice ? ' · 🎲2' : ''}
                {p.can_reroll ? ' · 🔄' : ''}
                {p.shopping_mall ? ' · 🏪' : ''}
              </div>
              <div className="p-cards">
                {grouped.map((g) => {
                  const c = g.card
                  const color = CARD_COLORS[c.color] || 'blue'
                  return (
                    <span
                      key={c.id}
                      className={`mini-card color-${color}`}
                      title={`${c.name} ×${g.count} | 🎲${(c.numbers || []).join(',')} | 💰${c.price}`}
                    >
                      {g.count}{ICONS[c.icon] || '?'}
                    </span>
                  )
                })}
              </div>
            </div>
          )
        })}
      </div>

      {human && (
        <div className="player-hand">
          <h3 onClick={() => setHandVisible(!handVisible)}>
            🃏 Мои карты {handVisible ? '▴' : '▾'}
          </h3>
          {handVisible && (
            <div>
              {groupCards(human.cards || []).map((g) => {
                const c = g.card
                const tip = `${c.name} | 🎲${(c.numbers || []).join(',')} | 💰${c.price} | ${EFFECT_NAMES[c.effect_type] || 'доход'} +${c.effect_value}`
                const col = CARD_COLORS[c.color] || 'blue'
                return (
                  <div key={c.id} className="hand-card" title={tip}>
                    <span className={`mini-color ${col}`}></span>
                    {' '}{ICONS[c.icon] || '?'} {c.name}
                    <span className="hc-count">×{g.count}</span>
                  </div>
                )
              })}
              {human.landmarks && human.landmarks.map((lm) =>
                lm.price > 0 ? (
                  <div key={lm.id} className="hand-card purchased">
                    🏛️ {lm.name} ✔
                  </div>
                ) : null
              )}
            </div>
          )}
        </div>
      )}
    </aside>
  )
}
