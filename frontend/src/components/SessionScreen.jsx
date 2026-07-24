import { useState, useEffect, useCallback } from 'react'
import { api } from '../api'
import CardModal from './CardModal'

export default function SessionScreen({ onJoinSession }) {
  const [sessions, setSessions] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [newId, setNewId] = useState('')
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
      await api('/api/sessions', {
        method: 'POST',
        body: JSON.stringify({ id, name: id }),
      })
      setNewId('')
      onJoinSession(id)
    } catch (e) {
      alert(e.message)
    }
  }, [newId, onJoinSession])

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
          <h3>Сохранённые сессии</h3>
          {loading ? (
            <div className="loading">Загрузка сессий...</div>
          ) : error ? (
            <div className="loading">Ошибка: {error}</div>
          ) : sessions.length === 0 ? (
            <div className="loading">Нет сохранённых игр. Создайте новую.</div>
          ) : (
            sessions.map((s) => {
              const created = new Date(s.created_at).toLocaleString('ru')
              const icon = s.completed ? '🏆' : s.game_data ? '🎮' : '📄'
              const disabled = s.completed
              return (
                <div
                  key={s.id}
                  className="session-item"
                  onClick={() => !disabled && onJoinSession(s.id)}
                  style={disabled ? { opacity: 0.5, pointerEvents: 'none' } : undefined}
                >
                  <div>
                    <div className="sess-name">{icon} {s.name}</div>
                    <div className="sess-meta">{created}</div>
                  </div>
                  <div>
                    <button
                      className="btn-small"
                      onClick={(e) => { e.stopPropagation(); handleDelete(s.id) }}
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
            placeholder="имя новой сессии"
          />
          <button className="btn-primary" onClick={handleCreate}>Создать</button>
        </div>

        <div style={{ marginTop: 12 }}>
          <button className="btn-secondary" onClick={() => setShowCardModal(true)}>
            🃏 Создать карту
          </button>
        </div>
      </div>

      {showCardModal && <CardModal onClose={() => { setShowCardModal(false); loadSessions() }} />}
    </div>
  )
}
