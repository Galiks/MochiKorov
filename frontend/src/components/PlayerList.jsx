import { useState } from 'react'
import { ICONS, CARD_COLORS, EFFECT_NAMES } from '../constants'

const COLOR_ORDER = ['Blue', 'Green', 'Red', 'Purple']

function groupCards(cards) {
  const grouped = {}
  for (const c of cards) {
    if (!grouped[c.id]) grouped[c.id] = { card: c, count: 0 }
    grouped[c.id].count++
  }
  return Object.values(grouped)
}

function groupCardsByColor(cards) {
  const byColor = {}
  for (const c of cards) {
    if (!byColor[c.color]) byColor[c.color] = {}
    if (!byColor[c.color][c.id]) byColor[c.color][c.id] = { card: c, count: 0 }
    byColor[c.color][c.id].count++
  }
  return byColor
}

export default function PlayerList({ players, state, yourID }) {
  const [handVisible, setHandVisible] = useState(false)
  const [collapsedColors, setCollapsedColors] = useState(new Set())
  const myPlayer = players.find(p => p.id === yourID)

  const toggleColor = (color) => {
    setCollapsedColors(prev => {
      const next = new Set(prev)
      if (next.has(color)) next.delete(color)
      else next.add(color)
      return next
    })
  }

  const renderPlayerSummary = (p) => {
    const cards = p.cards || []
    const grouped = groupCards(cards)
    return (
      <div key={p.id} className={`player-card ${p.is_current ? 'active' : ''}`}>
        <div className="p-name">{p.is_current ? '▶ ' : ''}{p.name}{p.id === yourID ? ' (это вы)' : ''}</div>
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
  }

  return (
    <aside className="panel-players">
      <h3>Игроки</h3>
      <div>
        {players.map(renderPlayerSummary)}
      </div>

      {myPlayer && (
        <div className="player-hand">
          <h3 onClick={() => setHandVisible(!handVisible)}>
            🃏 Мои карты {handVisible ? '▴' : '▾'}
          </h3>
          {handVisible && (
            <div>
              {(() => {
                const byColor = groupCardsByColor(myPlayer.cards || [])
                return COLOR_ORDER.map(colorName => {
                  const colorGroups = byColor[colorName]
                  if (!colorGroups || Object.keys(colorGroups).length === 0) return null
                  const col = CARD_COLORS[colorName] || 'blue'
                  const isCollapsed = collapsedColors.has(colorName)
                  const totalCount = Object.values(colorGroups).reduce((s, g) => s + g.count, 0)
                  return (
                    <div key={colorName}>
                      <div
                        className={`color-group-header${isCollapsed ? ' collapsed' : ''}`}
                        onClick={() => toggleColor(colorName)}
                      >
                        <span className={`mini-color ${col}`}></span>
                        <span>{colorName}</span>
                        <span className="cg-count">{totalCount}</span>
                        <span className="cg-arrow">{isCollapsed ? '▶' : '▼'}</span>
                      </div>
                      {!isCollapsed && (
                        <div className="hand-cards-grid">
                          {Object.values(colorGroups).map((g) => {
                            const c = g.card
                            const isDeactivated = c.min_landmark > 0 && myPlayer.landmark_count < c.min_landmark
                            const tip = `${c.name} | 🎲${(c.numbers || []).join(',')} | 💰${c.price} | ${EFFECT_NAMES[c.effect_type] || 'доход'} +${c.effect_value}`
                            return (
                              <div
                                key={c.id}
                                className={`hand-mini-card color-${col}${isDeactivated ? ' card-deactivated' : ''}`}
                                title={tip}
                              >
                                <div className="hmc-icon">{ICONS[c.icon] || '?'}</div>
                                <div className="hmc-name">{c.name}</div>
                                <div className="hmc-count">×{g.count}</div>
                              </div>
                            )
                          })}
                        </div>
                      )}
                    </div>
                  )
                })
              })()}
              {myPlayer.landmarks && myPlayer.landmarks.map((lm) =>
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
