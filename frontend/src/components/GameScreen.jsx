import { useState, useEffect, useRef, useCallback } from 'react'
import { api } from '../api'
import DicePanel from './DicePanel'
import MarketPanel from './MarketPanel'
import PlayerList from './PlayerList'
import LogPanel from './LogPanel'
import WinnerOverlay from './WinnerOverlay'

export default function GameScreen({ sessionId, onBack }) {
  const [gameState, setGameState] = useState(null)
  const [gameStarted, setGameStarted] = useState(false)
  const [eventLog, setEventLog] = useState([])
  const gameStateRef = useRef(gameState)

  useEffect(() => {
    gameStateRef.current = gameState
  }, [gameState])

  const refreshGame = useCallback(async () => {
    try {
      const data = await api(`/api/game/${sessionId}/state`)
      setGameState(data)
      setGameStarted(true)
    } catch (e) {
      if (e.message === 'no game data in session ' + sessionId) {
        setGameStarted(false)
        return
      }
      console.error(e)
    }
  }, [sessionId])

  useEffect(() => {
    refreshGame()
  }, [refreshGame])

  const startGame = useCallback(async () => {
    try {
      const data = await api(`/api/game/${sessionId}/start?cards=base`, { method: 'POST' })
      setGameState(data)
      setGameStarted(true)
    } catch (e) {
      alert(e.message)
    }
  }, [sessionId])

  const updateState = useCallback(async ({ data, appendLog } = {}) => {
    if (data) {
      setGameState(data)
      if (data.log && data.log.length > 0) {
        setEventLog(prev => [...prev, ...data.log.map(msg => ({
          text: msg,
          time: new Date().toLocaleTimeString('ru'),
        }))])
      }
    }
  }, [])

  const endTurnAndAI = useCallback(async (buyState) => {
    await updateState({ data: buyState })
    if (buyState.game_over) return

    try {
      const data = await api(`/api/game/${sessionId}/end-turn`, { method: 'POST' })
      await updateState({ data })
    } catch (e) {
      alert(e.message)
    }
  }, [sessionId, updateState])

  const handleBack = useCallback(async () => {
    await api(`/api/sessions/${sessionId}/game`, {
      method: 'PUT',
      body: JSON.stringify(gameStateRef.current),
    }).catch(() => {})
    onBack()
  }, [sessionId, onBack])

  const handleRoll = useCallback(async (diceCount) => {
    try {
      let data = await api(`/api/game/${sessionId}/roll`, {
        method: 'POST',
        body: JSON.stringify({ dice_count: diceCount }),
      })
      await updateState({ data })
      if (!data.can_reroll) {
        const cd = await api(`/api/game/${sessionId}/collect`, { method: 'POST' })
        await updateState({ data: cd })
      }
    } catch (e) {
      alert(e.message)
    }
  }, [sessionId, updateState])

  const handleReroll = useCallback(async (selectedDice) => {
    try {
      const data = await api(`/api/game/${sessionId}/reroll`, {
        method: 'POST',
        body: JSON.stringify({ indices: selectedDice.length > 0 ? selectedDice : null }),
      })
      await updateState({ data })
    } catch (e) {
      alert(e.message)
    }
  }, [sessionId, updateState])

  const handleContinue = useCallback(async () => {
    try {
      const data = await api(`/api/game/${sessionId}/collect`, { method: 'POST' })
      await updateState({ data })
    } catch (e) {
      alert(e.message)
    }
  }, [sessionId, updateState])

  const handleBuyMarket = useCallback(async (cardId) => {
    try {
      const data = await api(`/api/game/${sessionId}/buy`, {
        method: 'POST',
        body: JSON.stringify({ type: 'market', card_id: cardId }),
      })
      await endTurnAndAI(data)
    } catch (e) {
      alert(e.message)
    }
  }, [sessionId, endTurnAndAI])

  const handleBuyLandmark = useCallback(async (index) => {
    try {
      const data = await api(`/api/game/${sessionId}/buy`, {
        method: 'POST',
        body: JSON.stringify({ type: 'landmark', index }),
      })
      await endTurnAndAI(data)
    } catch (e) {
      alert(e.message)
    }
  }, [sessionId, endTurnAndAI])

  const handleSkip = useCallback(async () => {
    try {
      const data = await api(`/api/game/${sessionId}/buy`, {
        method: 'POST',
        body: JSON.stringify({ type: 'skip' }),
      })
      await endTurnAndAI(data)
    } catch (e) {
      alert(e.message)
    }
  }, [sessionId, endTurnAndAI])

  return (
    <div className="game-screen">
      <header>
        <div className="header-left">
          <span className="game-title">МАЧИ КОРО</span>
          <span className="session-badge">{sessionId}</span>
        </div>
        <div className="header-right">
          {gameState && (
            <span className="turn-display">Ход {gameState.turn + 1}</span>
          )}
          <button className="btn-small" onClick={handleBack}>← К сессиям</button>
        </div>
      </header>

      {!gameStarted ? (
        <div className="start-prompt">
          <h2>Новая игра</h2>
          <button className="btn-primary" onClick={startGame}>🎮 Начать игру</button>
        </div>
      ) : gameState ? (
        <div className="game-layout">
          <PlayerList
            players={gameState.players}
            state={gameState}
          />

          <main className="panel-center">
            <MarketPanel
              state={gameState}
              onBuyMarket={handleBuyMarket}
              onBuyLandmark={handleBuyLandmark}
            />
            <DicePanel
              state={gameState}
              onRoll={handleRoll}
              onReroll={handleReroll}
              onContinue={handleContinue}
              onSkip={handleSkip}
            />
          </main>

          <LogPanel log={eventLog} />
        </div>
      ) : null}

      {gameState?.winner && (
        <WinnerOverlay winner={gameState.winner} />
      )}
    </div>
  )
}
