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
  const [error, setError] = useState(null)
  const [loading, setLoading] = useState(false)
  const gameStateRef = useRef(gameState)
  const errorTimerRef = useRef(null)

  useEffect(() => {
    if (error) {
      if (errorTimerRef.current) clearTimeout(errorTimerRef.current)
      errorTimerRef.current = setTimeout(() => setError(null), 5000)
    }
    return () => {
      if (errorTimerRef.current) clearTimeout(errorTimerRef.current)
    }
  }, [error])

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
    setLoading(true)
    try {
      const data = await api(`/api/game/${sessionId}/start`, { method: 'POST' })
      setGameState(data)
      setGameStarted(true)
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }, [sessionId])

  const logIdRef = useRef(0)

  const updateState = useCallback(async ({ data, appendLog } = {}) => {
    if (data) {
      setGameState(data)
      if (data.log && data.log.length > 0) {
        setEventLog(prev => [...prev, ...data.log.map(msg => ({
          id: ++logIdRef.current,
          text: msg,
          time: new Date().toLocaleTimeString('ru'),
        }))])
      }
    }
  }, [])

  const endTurnAndAI = useCallback(async (buyState) => {
    if (buyState.game_over) {
      await updateState({ data: buyState })
      return
    }

    try {
      const data = await api(`/api/game/${sessionId}/end-turn`, { method: 'POST' })
      await updateState({ data })
    } catch (e) {
      setError(e.message)
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
    setLoading(true)
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
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }, [sessionId, updateState])

  const handleReroll = useCallback(async (selectedDice) => {
    setLoading(true)
    try {
      const data = await api(`/api/game/${sessionId}/reroll`, {
        method: 'POST',
        body: JSON.stringify({ indices: selectedDice.length > 0 ? selectedDice : null }),
      })
      await updateState({ data })
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }, [sessionId, updateState])

  const handleContinue = useCallback(async () => {
    setLoading(true)
    try {
      const data = await api(`/api/game/${sessionId}/collect`, { method: 'POST' })
      await updateState({ data })
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }, [sessionId, updateState])

  const handleBuyMarket = useCallback(async (cardId) => {
    setLoading(true)
    try {
      const data = await api(`/api/game/${sessionId}/buy`, {
        method: 'POST',
        body: JSON.stringify({ type: 'market', card_id: cardId }),
      })
      await endTurnAndAI(data)
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }, [sessionId, endTurnAndAI])

  const handleBuyLandmark = useCallback(async (index) => {
    setLoading(true)
    try {
      const data = await api(`/api/game/${sessionId}/buy`, {
        method: 'POST',
        body: JSON.stringify({ type: 'landmark', index }),
      })
      await endTurnAndAI(data)
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }, [sessionId, endTurnAndAI])

  const handleSkip = useCallback(async () => {
    setLoading(true)
    try {
      const data = await api(`/api/game/${sessionId}/buy`, {
        method: 'POST',
        body: JSON.stringify({ type: 'skip' }),
      })
      await endTurnAndAI(data)
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }, [sessionId, endTurnAndAI])

  return (
    <div className="game-screen">
      {error && <div className="error-banner">{error}</div>}
      {loading && <div className="loading-overlay"><div className="spinner" /></div>}
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
        <WinnerOverlay winner={gameState.winner} onClose={onBack} />
      )}
    </div>
  )
}
