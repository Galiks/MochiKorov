import { useState, useCallback } from 'react'
import SessionScreen from './components/SessionScreen'
import GameScreen from './components/GameScreen'

export default function App() {
  const [sessionId, setSessionId] = useState(null)

  const handleJoinSession = useCallback((id) => {
    setSessionId(id)
  }, [])

  const handleBack = useCallback(() => {
    setSessionId(null)
  }, [])

  if (!sessionId) {
    return <SessionScreen onJoinSession={handleJoinSession} />
  }

  return <GameScreen sessionId={sessionId} onBack={handleBack} />
}
