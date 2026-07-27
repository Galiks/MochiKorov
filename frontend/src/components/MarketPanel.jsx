import { useState } from 'react'
import { ICONS, CARD_COLORS, EFFECT_NAMES } from '../constants'

const COLOR_ORDER = ['Blue', 'Green', 'Red', 'Purple']

export default function MarketPanel({ state, onBuyMarket, onBuyLandmark }) {
  const [hiddenColors, setHiddenColors] = useState(new Set())

  if (!state.can_buy) {
    return (
      <div id="center-market">
        <div className="loading">🎲 Бросьте кубики, чтобы открыть рынок</div>
      </div>
    )
  }

  const items = (state.market || []).filter(item => item.count > 0)

  const groupedByColor = {}
  for (const item of items) {
    const color = item.card.color
    if (!groupedByColor[color]) groupedByColor[color] = []
    groupedByColor[color].push(item)
  }

  const toggleColor = (color) => {
    setHiddenColors(prev => {
      const next = new Set(prev)
      if (next.has(color)) next.delete(color)
      else next.add(color)
      return next
    })
  }

  const currentLandmarkCount = state.current_player?.landmark_count || 0

  const renderMarketCard = (item, color, index) => {
    const c = item.card
    const icon = ICONS[c.icon] || '❓'
    const effectName = EFFECT_NAMES[c.effect_type] || ''
    const effectDesc = c.effect_type
      ? `+${c.effect_value} ${effectName}`
      : `+${c.effect_value} монет`
    const isDeactivated = c.min_landmark > 0 && currentLandmarkCount < c.min_landmark
    const rot = (index % 2 === 0 ? -0.5 : 0.5) + (Math.floor(index / 2) % 2 === 0 ? 0.3 : -0.3)
    return (
      <div
        key={c.id + index}
        className={`market-card${isDeactivated ? ' card-deactivated' : ''}`}
        style={isDeactivated ? {} : { transform: `rotate(${rot}deg)` }}
        onClick={() => !isDeactivated && onBuyMarket(c.id)}
        title={`${c.name} | 🎲${(c.numbers || []).join(',')} | 💰${c.price} | ${effectDesc}`}
      >
        <span className={`mc-badge color-${color}`}>{c.color}</span>
        <div className="mc-icon">{icon}</div>
        <div className="mc-name">{c.name}</div>
        <div className="mc-meta">
          <span className="mc-price">💰 {c.price}</span>
          <span className="mc-stock">×{item.count}</span>
        </div>
        {c.numbers && (
          <div className="mc-dice">🎲{c.numbers.join(',')}</div>
        )}
        {effectName && (
          <div className="mc-effect">{effectDesc}</div>
        )}
      </div>
    )
  }

  return (
    <div id="center-market">
      <h3>Рынок</h3>
      {items.length === 0 && (state.available_landmarks || []).length === 0 ? (
        <div className="loading">Рынок пуст</div>
      ) : (
        <>
          <div className="market-grid">
            {COLOR_ORDER.map(colorName => {
              const colorItems = groupedByColor[colorName]
              if (!colorItems || colorItems.length === 0) return null
              const cssColor = CARD_COLORS[colorName] || 'blue'
              const isHidden = hiddenColors.has(colorName)
              const totalCount = colorItems.reduce((s, item) => s + item.count, 0)
              return (
                <span key={colorName} style={{ display: 'contents' }}>
                  <div
                    className={`color-section-header color-${cssColor}${isHidden ? ' hidden' : ''}`}
                    onClick={() => toggleColor(colorName)}
                  >
                    <span>{colorName}</span>
                    <span className="cs-count">{totalCount} шт.</span>
                    <span className="cs-arrow">{isHidden ? '▶' : '▼'}</span>
                  </div>
                  {!isHidden && colorItems.map((item, idx) =>
                    renderMarketCard(item, cssColor, idx)
                  )}
                </span>
              )
            })}
          </div>

          {(state.available_landmarks || []).length > 0 && (
            <>
              <h3 style={{ marginTop: 16 }}>Достопримечательности</h3>
              <div className="market-grid">
                {state.available_landmarks.map((lm, i) => (
                  <div
                    key={i}
                    className="market-card landmark-card"
                    style={{ transform: `rotate(${i % 2 === 0 ? -0.3 : 0.3}deg)` }}
                    onClick={() => onBuyLandmark(i)}
                    title={`${lm.name} | 💰${lm.price} монет`}
                  >
                    <span className="mc-badge color-purple">Landmark</span>
                    <div className="mc-icon">🏛️</div>
                    <div className="mc-name">{lm.name}</div>
                    <div className="mc-meta">
                      <span className="mc-price">💰 {lm.price}</span>
                    </div>
                  </div>
                ))}
              </div>
            </>
          )}
        </>
      )}
    </div>
  )
}
