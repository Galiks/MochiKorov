import { ICONS, CARD_COLORS, EFFECT_NAMES } from '../constants'

export default function MarketPanel({ state, onBuyMarket, onBuyLandmark }) {
  if (!state.can_buy) {
    return (
      <div id="center-market">
        <div className="loading">🎲 Бросьте кубики, чтобы открыть рынок</div>
      </div>
    )
  }

  const items = (state.market || []).filter(item => item.count > 0)

  return (
    <div id="center-market">
      <h3>Рынок</h3>
      {items.length === 0 && (state.available_landmarks || []).length === 0 ? (
        <div className="loading">Рынок пуст</div>
      ) : (
        <>
          <div className="market-grid">
            {items.map((item, i) => {
              const c = item.card
              const color = CARD_COLORS[c.color] || 'blue'
              const icon = ICONS[c.icon] || '❓'
              const effectName = EFFECT_NAMES[c.effect_type] || ''
              const effectDesc = c.effect_type
                ? `+${c.effect_value} ${effectName}`
                : `+${c.effect_value} монет`
              const rot = (i % 2 === 0 ? -0.5 : 0.5) + (Math.floor(i / 2) % 2 === 0 ? 0.3 : -0.3)
              return (
                <div
                  key={c.id}
                  className="market-card"
                  style={{ transform: `rotate(${rot}deg)` }}
                  onClick={() => onBuyMarket(c.id)}
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
