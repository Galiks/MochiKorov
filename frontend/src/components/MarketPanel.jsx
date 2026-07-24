import { ICONS, CARD_COLORS, EFFECT_NAMES } from '../constants'

export default function MarketPanel({ state, onBuyMarket, onBuyLandmark }) {
  if (!state.can_buy) {
    return (
      <div id="center-market">
        <div className="loading">🎲 Бросьте кубики, чтобы открыть рынок</div>
      </div>
    )
  }

  return (
    <div id="center-market">
      <h3>Рынок</h3>
      {(state.market || []).length === 0 ? (
        <div className="loading">Рынок пуст</div>
      ) : (
        state.market.map((item, i) => {
          if (item.count <= 0) return null
          const c = item.card
          const color = CARD_COLORS[c.color] || 'blue'
          const icon = ICONS[c.icon] || '❓'
          const effectName = EFFECT_NAMES[c.effect_type] || ''
          const effectDesc = c.effect_type
            ? `+${c.effect_value} ${effectName}`
            : `+${c.effect_value} монет`
          return (
            <div
              key={c.id}
              className="market-card"
              onClick={() => onBuyMarket(c.id)}
              title={`${c.name} | 🎲${(c.numbers || []).join(',')} | 💰${c.price} | ${effectDesc}`}
            >
              <div className="mc-info">
                <div className="mc-name">{icon} {c.name}</div>
                <div>
                  <span className="mc-price">💰 {c.price}</span>
                  <span className="mc-stock"> шт: {item.count}</span>
                  {c.numbers && (
                    <span className="mc-stock"> · 🎲{c.numbers.join(',')}</span>
                  )}
                  {effectName && (
                    <span className="mc-stock"> · {effectName}</span>
                  )}
                </div>
              </div>
              <span className={`mc-badge color-${color}`}>{c.color}</span>
            </div>
          )
        })
      )}

      <h3 style={{ marginTop: 12 }}>Достопримечательности</h3>
      {(state.available_landmarks || []).length === 0 ? (
        <div className="loading">Все достопримечательности куплены</div>
      ) : (
        state.available_landmarks.map((lm, i) => (
          <div
            key={i}
            className="market-card"
            onClick={() => onBuyLandmark(i)}
            title={`${lm.name} | 💰${lm.price} монет`}
          >
            <div className="mc-info">
              <div className="mc-name">🏛️ {lm.name}</div>
              <div className="mc-price">💰 {lm.price}</div>
            </div>
          </div>
        ))
      )}
    </div>
  )
}
