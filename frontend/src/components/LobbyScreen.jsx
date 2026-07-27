import { useState, useEffect, useCallback } from 'react'
import { api } from '../api'

export default function LobbyScreen({ sessionId, onGameStart }) {
  const [lobby, setLobby] = useState(null)
  const [joinName, setJoinName] = useState('')
  const [starting, setStarting] = useState(false)
  const [error, setError] = useState(null)

  const token = localStorage.getItem(`mochi_token_${sessionId}`)
  const hasJoined = !!token

  const loadLobby = useCallback(async () => {
    try {
      const data = await api(`/api/sessions/${sessionId}/lobby`)
      setLobby(data)
    } catch (e) {
      setError(e.message)
    }
  }, [sessionId])

  useEffect(() => {
    loadLobby()
    const interval = setInterval(loadLobby, 2000)
    return () => clearInterval(interval)
  }, [loadLobby])

  const handleJoin = useCallback(async () => {
    const name = joinName.trim()
    if (!name) return
    try {
      const data = await api(`/api/sessions/${sessionId}/lobby/join`, {
        method: 'POST',
        body: JSON.stringify({ name }),
      })
      localStorage.setItem(`mochi_token_${sessionId}`, data.token)
      setJoinName('')
      await loadLobby()
    } catch (e) {
      alert(e.message)
    }
  }, [joinName, sessionId, loadLobby])

  const handleLeave = useCallback(async (idx) => {
    try {
      await api(`/api/sessions/${sessionId}/lobby/leave/${idx}`, { method: 'DELETE' })
      await loadLobby()
    } catch (e) {
      alert(e.message)
    }
  }, [sessionId, loadLobby])

  const handleStart = useCallback(async () => {
    setStarting(true)
    try {
      const data = await api(`/api/game/${sessionId}/start`, { method: 'POST' })
      onGameStart(data)
    } catch (e) {
      alert(e.message)
      setStarting(false)
    }
  }, [sessionId, onGameStart])

  if (!lobby) {
    return (
      <div className="session-screen">
        <div className="session-container">
          <h1>МАЧИ КОРО</h1>
          <div className="loading">Загрузка комнаты...</div>
        </div>
      </div>
    )
  }

  const players = lobby.players || []
  const playerCount = players.length
  const maxPlayers = lobby.max_players
  const isFull = playerCount >= maxPlayers
  const botCount = maxPlayers - playerCount
  const isCreator = lobby.your_index === 0

  return (
    <div className="session-screen">
      <div className="session-container lobby-container">
        <h1>МАЧИ КОРО</h1>
        <p className="subtitle">Комната: {sessionId}</p>

        <div className="lobby-info">
          Игроки: <strong>{playerCount} / {maxPlayers}</strong>
        </div>

        <div className="lobby-players">
          {players.map((p, i) => (
            <div key={i} className="lobby-player-row">
              <span className="lpp-icon">{i === 0 ? '👑' : '👤'}</span>
              <span className="lpp-name">{p.name}</span>
              {i === 0 && <span className="lpp-badge">Создатель</span>}
              {lobby.your_index === i && <span className="lpp-badge" style={{ background: 'var(--green)' }}>Это вы</span>}
              {isCreator && i > 0 && (
                <button
                  className="btn-small"
                  onClick={() => handleLeave(i)}
                  title="Удалить игрока"
                  style={{ marginLeft: 'auto', color: 'var(--red)' }}
                >
                  ✕
                </button>
              )}
            </div>
          ))}
        </div>

        {!hasJoined && !isFull && (
          <div className="lobby-join">
            <input
              value={joinName}
              onChange={(e) => setJoinName(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleJoin()}
              placeholder="Ваше имя"
            />
            <button className="btn-primary" onClick={handleJoin}>Войти</button>
          </div>
        )}
        {!hasJoined && isFull && (
          <div className="lobby-full">Комната заполнена</div>
        )}
        {hasJoined && (
          <div className="lobby-full" style={{ color: 'var(--text)' }}>Ожидание начала игры...</div>
        )}

        {botCount > 0 && (
          <div className="lobby-bots-info">
            При старте будет добавлено <strong>{botCount}</strong> {botCount === 1 ? 'бот' : 'бота(ов)'}
          </div>
        )}

        {error && <div className="error-banner">{error}</div>}

        <button
          className="btn-primary"
          onClick={handleStart}
          disabled={starting}
          style={{ marginTop: 16, width: '100%', padding: '12px 20px', fontSize: '1rem' }}
        >
          {starting ? '🎲 Запуск...' : '🎮 Начать игру'}
        </button>
      </div>
    </div>
  )
}
