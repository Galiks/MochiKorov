export default function WinnerOverlay({ winner, onClose }) {
  return (
    <div className="winner-overlay">
      <div className="winner-box">
        <h2>🏆 ПОБЕДИТЕЛЬ!</h2>
        <p>{winner.name}</p>
        <button className="btn-primary" onClick={onClose || (() => {})}>
          Закрыть
        </button>
      </div>
    </div>
  )
}
