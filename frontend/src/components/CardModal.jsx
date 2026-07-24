import { useState } from 'react'
import { api } from '../api'
import { ICONS, CARD_COLORS, EFFECT_NAMES } from '../constants'

export default function CardModal({ onClose }) {
  const [name, setName] = useState('')
  const [color, setColor] = useState('Blue')
  const [dice, setDice] = useState('')
  const [price, setPrice] = useState('')
  const [reward, setReward] = useState('')
  const [stock, setStock] = useState('6')

  const handleCreate = async () => {
    const diceNum = parseInt(dice)
    const priceNum = parseInt(price)
    const rewardNum = parseInt(reward)
    const stockNum = parseInt(stock || '6')

    if (!diceNum || !priceNum || !rewardNum) {
      alert('Заполните все поля')
      return
    }

    if (!name.trim()) {
      alert('Введите название карты')
      return
    }

    const effectType = color === 'Red' ? 3 : 2
    const cardId = 'custom_' + Date.now()

    try {
      await api('/api/establishments', {
        method: 'POST',
        body: JSON.stringify({
          id: cardId, name: name.trim(), color, icon: '❓',
          numbers: [diceNum],
          price: priceNum, effect_type: effectType,
          effect_value: rewardNum,
          default_stock: stockNum,
        }),
      })
      alert('Карта создана!')
      onClose()
    } catch (e) {
      alert('Ошибка: ' + e.message)
    }
  }

  return (
    <div className="modal-overlay" onClick={(e) => e.target === e.currentTarget && onClose()}>
      <div className="modal-box">
        <h1>Создание карты</h1>
        <div className="cc-form">
          <input value={name} onChange={(e) => setName(e.target.value)} placeholder="Название карты" />
          <select value={color} onChange={(e) => setColor(e.target.value)}>
            <option value="Blue">🔵 Синяя</option>
            <option value="Green">🟢 Зелёная</option>
            <option value="Red">🔴 Красная</option>
            <option value="Purple">🟣 Фиолетовая</option>
          </select>
          <input type="number" min="1" max="12" value={dice} onChange={(e) => setDice(e.target.value)} placeholder="Число на кубике" />
          <input type="number" min="1" value={price} onChange={(e) => setPrice(e.target.value)} placeholder="Стоимость" />
          <input type="number" min="1" value={reward} onChange={(e) => setReward(e.target.value)} placeholder="Вознаграждение" />
          <input type="number" min="1" value={stock} onChange={(e) => setStock(e.target.value)} placeholder="Тираж" />
          <div className="cc-actions">
            <button className="btn-primary" onClick={handleCreate}>Создать</button>
            <button className="btn-small" onClick={onClose}>Закрыть</button>
          </div>
        </div>
      </div>
    </div>
  )
}
