export default function LogPanel({ log }) {
  return (
    <aside className="panel-logs">
      <h3>События</h3>
      <div>
        {log.length === 0 ? (
          <div className="loading">Ожидание действий...</div>
        ) : (
          [...log].reverse().map(entry => (
            <div key={entry.id} className="log-entry">
              <span className="log-time">{entry.time}</span> {entry.text}
            </div>
          ))
        )}
      </div>
    </aside>
  )
}
