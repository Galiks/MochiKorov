import { useState } from 'react'

const pipPositions = {
  1: [[1, 1]],
  2: [[0, 2], [2, 0]],
  3: [[0, 2], [1, 1], [2, 0]],
  4: [[0, 0], [0, 2], [2, 0], [2, 2]],
  5: [[0, 0], [0, 2], [1, 1], [2, 0], [2, 2]],
  6: [[0, 0], [0, 2], [1, 0], [1, 2], [2, 0], [2, 2]],
}

function DieFace({ value, selected, clickable, onClick }) {
  const pips = pipPositions[value] || []
  return (
    <div
      className={`die${selected ? ' selected' : ''}${clickable ? ' clickable' : ''}`}
      onClick={onClick}
    >
      <div className="die-face">
        {pips.map((pos, i) => (
          <div
            key={i}
            className="die-pip"
            style={{ '--px': pos[0], '--py': pos[1] }}
          />
        ))}
      </div>
    </div>
  )
}

export default function DicePanel({ state, onRoll, onReroll, onContinue, onSkip }) {
  const [diceCount, setDiceCount] = useState(1)
  const [selectedDice, setSelectedDice] = useState([])

  const toggleDie = (idx) => {
    setSelectedDice((prev) =>
      prev.includes(idx) ? prev.filter((i) => i !== idx) : [...prev, idx]
    )
  }

  const handleRoll = () => {
    onRoll(diceCount)
  }

  const handleReroll = () => {
    onReroll(selectedDice)
    setSelectedDice([])
  }

  const p = state.current_player
  const canTwoDice = p?.can_roll_two_dice
  const effectiveDiceCount = (state.can_roll && state.phase === 'roll') ? diceCount : (state.dice?.numbers?.length || 1)
  const isRerollPhase = state.can_reroll && state.phase === 'income'

  return (
    <div className="current-player-info">
      <h2 id="current-name">🎲 Ход: {p?.name} (💰 {p?.money})</h2>

      <div className="dice-area">
        <div className="dice-display">
          {state.phase === 'roll' && !state.dice?.numbers?.length ? (
            <div className="loading">Нажмите "Бросить кубики"</div>
          ) : state.dice?.numbers?.length > 0 ? (
            <>
              {state.dice.numbers.map((n, i) => {
                const isSel = selectedDice.includes(i)
                return (
                  <DieFace
                    key={i}
                    value={n}
                    selected={isSel}
                    clickable={isRerollPhase}
                    onClick={() => isRerollPhase && toggleDie(i)}
                  />
                )
              })}
              <div className="die-sum">
                {state.dice.numbers.length > 1 ? '= ' : 'сумма: '}{state.dice.sum}
              </div>
            </>
          ) : null}
        </div>

        <div className="actions">
          {state.can_roll && (
            <button className="btn-primary" onClick={handleRoll}>🎲 Бросить кубики</button>
          )}
          {state.can_reroll && (
            <button className="btn-secondary" onClick={handleReroll}>🔄 Перебросить</button>
          )}
          {state.phase === 'income' && !state.can_reroll && (
            <button className="btn-primary" onClick={onContinue}>▶ Продолжить</button>
          )}
          {state.can_buy && (
            <button className="btn-secondary" onClick={onSkip}>⏭ Пропустить ход</button>
          )}
        </div>

        {state.can_roll && (
          <div className="dice-switch">
            <button
              className={`btn-dice${diceCount === 1 ? ' active' : ''}`}
              onClick={() => setDiceCount(1)}
            >
              1 кубик
            </button>
            <button
              className={`btn-dice${diceCount === 2 ? ' active' : ''}`}
              onClick={() => setDiceCount(2)}
              disabled={!canTwoDice}
              title={canTwoDice ? '' : 'Нужен Порт или ЖД Вокзал'}
            >
              2 кубика
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
