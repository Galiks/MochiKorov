import { useState, useEffect, useCallback } from 'react'
import { api } from '../api'
import CardModal from './CardModal'

export default function SessionScreen({ onJoinSession }) {
  const [sessions, setSessions] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [newId, setNewId] = useState('')
  const [maxPlayers, setMaxPlayers] = useState(2)
  const [joinName, setJoinName] = useState('')
  const [showCardModal, setShowCardModal] = useState(false)

  const loadSessions = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const data = await api('/api/sessions')
      setSessions(data.sessions || [])
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    loadSessions()
  }, [loadSessions])

  const handleCreate = useCallback(async () => {
    const id = newId.trim()
    if (!id) return
    if (!/^[a-zA-Z0-9_-]+$/.test(id)) {
      alert('Имя сессии может содержать только буквы, цифры, - и _')
      return
    }
    try {
      const data = await api('/api/sessions', {
        method: 'POST',
        body: JSON.stringify({ id, name: id, max_players: maxPlayers }),
      })
      localStorage.setItem(`mochi_token_${id}`, data.token)
      setNewId('')
      onJoinSession(id)
    } catch (e) {
      alert(e.message)
    }
  }, [newId, maxPlayers, onJoinSession])

  const handleJoin = useCallback(async (sessionId) => {
    const name = prompt('Ваше имя для этой игры:')
    if (!name || !name.trim()) return
    try {
      const data = await api(`/api/sessions/${sessionId}/lobby/join`, {
        method: 'POST',
        body: JSON.stringify({ name: name.trim() }),
      })
      localStorage.setItem(`mochi_token_${sessionId}`, data.token)
      onJoinSession(sessionId)
    } catch (e) {
      alert(e.message)
    }
  }, [onJoinSession])

  const handleDelete = useCallback(async (id) => {
    if (!confirm(`Удалить сессию "${id}"?`)) return
    try {
      await api(`/api/sessions/${id}`, { method: 'DELETE' })
      loadSessions()
    } catch (e) {
      alert(e.message)
    }
  }, [loadSessions])

  const handleKeyDown = useCallback((e) => {
    if (e.key === 'Enter') handleCreate()
  }, [handleCreate])

  return (
    <div className="session-screen">
      <div className="session-container">
        <h1>МАЧИ КОРО</h1>
        <p className="subtitle">Градостроительная карточная игра</p>

        <div className="session-list">
          <h3>Комнаты</h3>
          {loading ? (
            <div className="loading">Загрузка комнат...</div>
          ) : error ? (
            <div className="loading">Ошибка: {error}</div>
          ) : sessions.length === 0 ? (
            <div className="loading">Нет активных комнат. Создайте новую.</div>
          ) : (
            sessions.map((s) => {
              const created = new Date(s.created_at).toLocaleString('ru')
              const playerCount = (s.lobby_players || []).length
              const isFull = playerCount >= s.max_players
              const inLobby = !s.game_data && !s.completed

              let icon = '📄'
              let action = null
              if (s.completed) {
                icon = '🏆'
              } else if (s.game_data) {
                icon = '🎮'
              } else if (isFull) {
                icon = '🚫'
              }

              return (
                <div
                  key={s.id}
                  className="session-item"
                  onClick={() => !s.completed && onJoinSession(s.id)}
                  style={s.completed ? { opacity: 0.5, cursor: 'default' } : undefined}
                >
                  <div>
                    <div className="sess-name">{icon} {s.name}</div>
                    <div className="sess-meta">
                      {s.completed
                        ? 'Завершена'
                        : s.game_data
                          ? 'В игре'
                          : `${playerCount} / ${s.max_players} игроков`}
                      {' · '}{created}
                    </div>
                  </div>
                  <div style={{ display: 'flex', gap: 6, alignItems: 'center' }}>
                    {inLobby && !isFull && (
                      <button
                        className="btn-small"
                        onClick={(e) => { e.stopPropagation(); handleJoin(s.id) }}
                        style={{ borderColor: 'var(--green)', color: 'var(--green)' }}
                      >
                        Войти
                      </button>
                    )}
                    {inLobby && isFull && (
                      <span style={{ fontSize: '0.75rem', color: 'var(--text2)' }}>Полная</span>
                    )}
                    <button
                      className="btn-small"
                      onClick={(e) => { e.stopPropagation(); handleDelete(s.id) }}
                      title="Удалить"
                    >
                      ✕
                    </button>
                  </div>
                </div>
              )
            })
          )}
        </div>

        <div className="session-new">
          <input
            value={newId}
            onChange={(e) => setNewId(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="название комнаты"
          />
          <select
            value={maxPlayers}
            onChange={(e) => setMaxPlayers(Number(e.target.value))}
            style={{
              padding: '10px',
              borderRadius: 'var(--radius)',
              background: 'var(--bg2)',
              color: 'var(--text)',
              border: '1px solid var(--wood)',
              fontSize: '0.95rem',
              cursor: 'pointer',
            }}
          >
            {[2, 3, 4, 5].map(n => (
              <option key={n} value={n}>{n} игрока</option>
            ))}
          </select>
          <button className="btn-primary" onClick={handleCreate}>Создать</button>
        </div>

        <div style={{ marginTop: 12, display: 'flex', gap: 8, justifyContent: 'center' }}>
          <button className="btn-secondary" onClick={() => setShowCardModal(true)}>
            🃏 Создать карту
          </button>
          <button className="btn-secondary" onClick={loadSessions}>
            🔄 Обновить
          </button>
        </div>
      </div>

      {showCardModal && <CardModal onClose={() => { setShowCardModal(false); loadSessions() }} />}
    </div>
  )
}
