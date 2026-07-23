let currentSession = null;
let currentState = null;
let eventLog = [];
let diceCount = 1;
let selectedDice = [];

const ICONS = {
  "Пшеница": "🌾",
  "Ранчо": "🐄",
  "Пекарня": "🍞",
  "Кафе": "☕",
  "Магазин": "🏪",
  "Лес": "🌲",
  "Ресторан": "🍣",
  "Фабрика": "🏭",
  "Фрукты": "🍎",
  "Шахта": "⛏️",
  "Крупное предприятие": "🏢",
  "Цветник": "🌸",
  "Яблоня": "🍏",
  "Налог": "💰",
  "Мэрия": "🏛️",
  "Бухта": "⚓"
};

const EFFECT_NAMES = {
  0: "",
  1: "с каждого",
  2: "из банка",
  3: "с активного",
  4: "кража",
  5: "×ранчо",
  6: "×лес",
  7: "×пшеница",
  8: "×фиол.",
  9: "половина",
  10: "фикс."
};

const CARD_COLORS = {
  "Blue": "blue",
  "Green": "green",
  "Red": "red",
  "Purple": "purple"
};

async function api(path, opts = {}) {
  const res = await fetch(path, {
    headers: { "Content-Type": "application/json" },
    ...opts
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || res.statusText);
  return data;
}

function $(id) { return document.getElementById(id); }

function showScreen(id) {
  document.querySelectorAll(".screen").forEach(s => s.classList.remove("active"));
  $(id).classList.add("active");
}

function showCardModal() {
  $("card-creator-modal").style.display = "flex";
}
function closeCardModal() {
  $("card-creator-modal").style.display = "none";
}

/* ── Session Screen ── */
async function loadSessions() {
  try {
    const data = await api("/api/sessions");
    renderSessionList(data.sessions || []);
  } catch (e) {
    $("session-list").innerHTML = `<div class="loading">Ошибка: ${e.message}</div>`;
  }
}

function renderSessionList(sessions) {
  const el = $("session-list");
  if (sessions.length === 0) {
    el.innerHTML = `
      <h3>Сохранённые сессии</h3>
      <div class="loading">Нет сохранённых игр. Создайте новую.</div>
    `;
    return;
  }
  let html = `<h3>Выберите сессию (${sessions.length})</h3>`;
  for (const s of sessions) {
    const created = new Date(s.created_at).toLocaleString("ru");
    const icon = s.completed ? "🏆" : (s.game_data ? "🎮" : "📄");
    const disabled = s.completed ? 'style="opacity:0.5;pointer-events:none"' : '';
    html += `
      <div class="session-item" onclick="joinSession('${s.id}')" ${disabled}>
        <div>
          <div class="sess-name">${icon} ${s.name}</div>
          <div class="sess-meta">${created}</div>
        </div>
        <div>
          <button class="btn-small" onclick="event.stopPropagation(); deleteSession('${s.id}')">✕</button>
        </div>
      </div>
    `;
  }
  el.innerHTML = html;
}

async function joinSession(id) {
  currentSession = id;
  try {
    const data = await api(`/api/sessions/${id}`);
    showScreen("game-screen");
    $("session-badge").textContent = id;
    await refreshGame();
  } catch (e) {
    alert(e.message);
  }
}

async function deleteSession(id) {
  if (!confirm(`Удалить сессию "${id}"?`)) return;
  try {
    await api(`/api/sessions/${id}`, { method: "DELETE" });
    await loadSessions();
  } catch (e) {
    alert(e.message);
  }
}

$("btn-new-session").addEventListener("click", async () => {
  const id = $("new-session-id").value.trim();
  if (!id) return;
  if (!/^[a-zA-Z0-9_-]+$/.test(id)) {
    alert("Имя сессии может содержать только буквы, цифры, - и _");
    return;
  }
  try {
    await api("/api/sessions", {
      method: "POST",
      body: JSON.stringify({ id, name: id })
    });
    $("new-session-id").value = "";
    await joinSession(id);
  } catch (e) {
    alert(e.message);
  }
});

$("new-session-id").addEventListener("keydown", e => {
  if (e.key === "Enter") $("btn-new-session").click();
});

/* ── Game Screen ── */
async function startGame() {
  if (!currentSession) return;
  $("start-prompt").style.display = "none";
  $("game-content").style.display = "";
  try {
    const data = await api(`/api/game/${currentSession}/start?cards=base`, { method: "POST" });
    currentState = data;
    renderGame(data);
  } catch (e) {
    alert(e.message);
  }
}

async function refreshGame() {
  if (!currentSession) return;
  try {
    const data = await api(`/api/game/${currentSession}/state`);
    currentState = data;
    $("start-prompt").style.display = "none";
    $("game-content").style.display = "";
    renderGame(data);
  } catch (e) {
    if (e.message === "no game data in session " + currentSession) {
      $("game-content").style.display = "none";
      $("start-prompt").style.display = "flex";
      return;
    }
    console.error(e);
  }
}

function renderGame(state) {
  if (!state) return;

  $("turn-display").textContent = `Ход ${state.turn + 1}`;

  if (state.game_over && state.winner) {
    showWinner(state.winner);
  }

  renderPlayers(state);
  renderCurrentPlayer(state);
  renderDice(state);
  renderActions(state);
  renderMarket(state);
  renderLandmarks(state);
  renderLog(state);
  renderHand(state);
}

function groupCards(cards) {
  const grouped = {};
  for (const c of cards) {
    if (!grouped[c.id]) grouped[c.id] = { card: c, count: 0 };
    grouped[c.id].count++;
  }
  return Object.values(grouped);
}

function toggleHand() {
  const h = $("hand-content");
  h.style.display = h.style.display === "none" ? "block" : "none";
}

function renderHand(state) {
  const human = state.players.find(p => p.id === 0);
  if (!human) return;
  const grouped = groupCards(human.cards || []);
  let html = '';
  for (const g of grouped) {
    const c = g.card;
    const tip = `${c.name} | 🎲${(c.numbers||[]).join(',')} | 💰${c.price} | ${EFFECT_NAMES[c.effect_type]||'доход'} +${c.effect_value}`;
    const col = CARD_COLORS[c.color] || "blue";
    html += `<div class="hand-card" title="${tip}"><span class="mini-color ${col}"></span> ${ICONS[c.icon]||'?'} ${c.name} <span class="hc-count">×${g.count}</span></div>`;
  }
  for (const lm of human.landmarks) {
    if (lm.price > 0) {
      html += `<div class="hand-card purchased">🏛️ ${lm.name} ✔</div>`;
    }
  }
  $("hand-content").innerHTML = html;
}

async function createCard() {
  const color = $("cc-color").value;
  const dice = parseInt($("cc-dice").value);
  const price = parseInt($("cc-price").value);
  const reward = parseInt($("cc-reward").value);
  const stock = parseInt($("cc-stock").value || "6");

  if (!dice || !price || !reward) {
    alert("Заполните все поля");
    return;
  }

  const name = $("cc-name").value.trim();
  if (!name) {
    alert("Введите название карты");
    return;
  }

  const effectType = (color === "Red") ? 3 : 2;
  const id = "custom_" + Date.now();

  try {
    await api("/api/establishments", {
      method: "POST",
      body: JSON.stringify({
        id, name, color, icon: "❓",
        numbers: [dice],
        price, effect_type: effectType,
        effect_value: reward,
        default_stock: stock
      })
    });
    alert("Карта создана!");
    closeCardModal();
    loadSessions();
  } catch (e) {
    alert("Ошибка: " + e.message);
  }
}

function renderPlayers(state) {
  const el = $("players-list");
  let html = "";
  for (const p of state.players) {
    const cards = p.cards || [];
    let cardsHtml = "";
    const grouped = groupCards(cards);
    for (const g of grouped) {
      const c = g.card;
      const color = CARD_COLORS[c.color] || "blue";
      const tip = `${c.name} ×${g.count} | 🎲${(c.numbers||[]).join(',')} | 💰${c.price} | ${EFFECT_NAMES[c.effect_type]||'доход'} +${c.effect_value}`;
      cardsHtml += `<span class="mini-card color-${color}" title="${tip}">${g.count}${ICONS[c.icon] || "?"}</span>`;
    }
    html += `
      <div class="player-card ${p.is_current ? "active" : ""}">
        <div class="p-name">${p.is_current ? "▶ " : ""}${p.name}</div>
        <div class="p-money">💰 ${p.money} монет</div>
        <div class="p-stats">
          🏛️ ${p.landmark_count}/${currentState?.total_landmarks || 7} · 🃏 ${cards.length} карт
          ${p.can_roll_two_dice ? " · 🎲2" : ""}
          ${p.can_reroll ? " · 🔄" : ""}
          ${p.shopping_mall ? " · 🏪" : ""}
        </div>
        <div class="p-cards">${cardsHtml}</div>
      </div>
    `;
  }
  el.innerHTML = html;
}

function renderCurrentPlayer(state) {
  const p = state.current_player;
  $("dice-display").innerHTML = "";
  if (state.phase === "roll") {
    $("dice-display").innerHTML = `<div class="loading">Нажмите "Бросить кубики"</div>`;
  }
}

function renderDice(state) {
  if (!state.dice || !state.dice.numbers || state.dice.numbers.length === 0) return;
  const dice = state.dice;
  let html = "";
  for (const n of dice.numbers) {
    const isSel = selectedDice.includes(i);
    const cls = `die${isSel ? ' selected' : ''}`;
    const onclick = state.can_reroll ? ` onclick="toggleDie(${i})"` : '';
    html += `<div class="${cls}"${onclick}>${n}</div>`;
  }
  if (dice.numbers.length > 1) {
    html += `<div class="die-sum">= ${dice.sum}</div>`;
  } else {
    html += `<div class="die-sum">сумма: ${dice.sum}</div>`;
  }
  $("dice-display").innerHTML = html;
}

function renderActions(state) {
  const rollBtn = $("btn-roll");
  const rerollBtn = $("btn-reroll");
  const continueBtn = $("btn-continue");
  const skipBtn = $("btn-skip");
  const diceSwitch = $("dice-switch");

  rollBtn.style.display = state.can_roll ? "inline-block" : "none";
  rerollBtn.style.display = state.can_reroll ? "inline-block" : "none";
  continueBtn.style.display = state.phase === "income" ? "inline-block" : "none";
  skipBtn.style.display = state.can_buy ? "inline-block" : "none";
  diceSwitch.style.display = state.can_roll ? "flex" : "none";

  const canTwoDice = state.current_player?.can_roll_two_dice;
  const twoBtn = $("btn-2dice");
  twoBtn.disabled = !canTwoDice;
  twoBtn.title = canTwoDice ? "" : "Нужен Порт или ЖД Вокзал";
  if (!canTwoDice && diceCount === 2) {
    diceCount = 1;
    updateDiceButtons();
  }

  const p = state.current_player;
  $("current-name").textContent = `🎲 Ход: ${p.name} (💰 ${p.money})`;
}

function renderMarket(state) {
  const el = $("market-items");
  if (!state.can_buy) {
    el.innerHTML = '<div class="loading">🎲 Бросьте кубики, чтобы открыть рынок</div>';
    return;
  }
  const market = state.market || [];
  if (market.length === 0) {
    el.innerHTML = '<div class="loading">Рынок пуст</div>';
    return;
  }
  let html = "";
  for (let i = 0; i < market.length; i++) {
    const item = market[i];
    const c = item.card;
    if (item.count <= 0) continue;
    const color = CARD_COLORS[c.color] || "blue";
    const icon = ICONS[c.icon] || "❓";
    const effectName = EFFECT_NAMES[c.effect_type] || "";
    const effectDesc = c.effect_type ? `+${c.effect_value} ${effectName}` : `+${c.effect_value} монет`;
    const tip = `${c.name} | 🎲${(c.numbers||[]).join(',')} | 💰${c.price} | ${effectDesc} ${c.condition ? ' | 🏛️ ' + c.condition : ''}`;
    html += `
      <div class="market-card" onclick="buyMarketItem('${c.id}')" title="${tip}">
        <div class="mc-info">
          <div class="mc-name">${icon} ${c.name}</div>
          <div>
            <span class="mc-price">💰 ${c.price}</span>
            <span class="mc-stock">шт: ${item.count}</span>
            ${c.numbers ? `<span class="mc-stock"> · 🎲${c.numbers.join(",")}</span>` : ""}
            ${effectName ? `<span class="mc-stock"> · ${effectName}</span>` : ""}
          </div>
        </div>
        <span class="mc-badge color-${color}">${c.color}</span>
      </div>
    `;
  }
  el.innerHTML = html;
}

function renderLandmarks(state) {
  const el = $("landmark-items");
  if (!state.can_buy) {
    el.innerHTML = '<div class="loading">🎲 Бросьте кубики, чтобы открыть рынок</div>';
    return;
  }
  const landmarks = state.available_landmarks || [];
  if (landmarks.length === 0) {
    el.innerHTML = '<div class="loading">Все достопримечательности куплены</div>';
    return;
  }
  let html = "";
  for (let i = 0; i < landmarks.length; i++) {
    const lm = landmarks[i];
    const lmTip = `${lm.name} | 💰${lm.price} монет`;
    html += `
      <div class="market-card" onclick="buyLandmarkItem(${i})" title="${lmTip}">
        <div class="mc-info">
          <div class="mc-name">🏛️ ${lm.name}</div>
          <div class="mc-price">💰 ${lm.price}</div>
        </div>
      </div>
    `;
  }
  el.innerHTML = html;
}

function renderLog(state) {
  const el = $("log-list");
  if (state.log && state.log.length > 0) {
    for (const msg of state.log) {
      eventLog.push({ text: msg, time: new Date().toLocaleTimeString("ru") });
    }
  }
  let html = "";
  for (let i = eventLog.length - 1; i >= 0; i--) {
    const e = eventLog[i];
    html += `<div class="log-entry"><span class="log-time">${e.time}</span> ${e.text}</div>`;
  }
  if (!html) {
    html = '<div class="loading">Ожидание действий...</div>';
  }
  el.innerHTML = html;
}

function showWinner(winner) {
  const existing = document.querySelector(".winner-overlay");
  if (existing) return;
  const overlay = document.createElement("div");
  overlay.className = "winner-overlay";
  overlay.innerHTML = `
    <div class="winner-box">
      <h2>🏆 ПОБЕДИТЕЛЬ!</h2>
      <p>${winner.name}</p>
      <button class="btn-primary" onclick="this.closest('.winner-overlay').remove()">Закрыть</button>
    </div>
  `;
  document.body.appendChild(overlay);
}

/* ── Game Actions ── */
const b1 = $("btn-1dice");
const b2 = $("btn-2dice");
if (b1 && b2) {
  b1.addEventListener("click", () => { diceCount = 1; updateDiceButtons(); });
  b2.addEventListener("click", () => { diceCount = 2; updateDiceButtons(); });
}

function updateDiceButtons() {
  $("btn-1dice").className = diceCount === 1 ? "btn-dice active" : "btn-dice";
  $("btn-2dice").className = diceCount === 2 ? "btn-dice active" : "btn-dice";
}

$("btn-roll").addEventListener("click", async () => {
  if (!currentSession) return;
  try {
    const data = await api(`/api/game/${currentSession}/roll`, {
      method: "POST",
      body: JSON.stringify({ dice_count: diceCount })
    });
    currentState = data;
    renderGame(data);
    if (!data.can_reroll) {
      const cd = await api(`/api/game/${currentSession}/collect`, { method: "POST" });
      currentState = cd;
      renderGame(cd);
    }
  } catch (e) {
    alert(e.message);
  }
});

$("btn-reroll").addEventListener("click", async () => {
  if (!currentSession) return;
  const indices = selectedDice.length > 0 ? selectedDice : null;
  try {
    const data = await api(`/api/game/${currentSession}/reroll`, {
      method: "POST",
      body: JSON.stringify({ indices })
    });
    currentState = data;
    selectedDice = [];
    renderGame(data);
  } catch (e) {
    alert(e.message);
  }
});

function toggleDie(idx) {
  const pos = selectedDice.indexOf(idx);
  if (pos >= 0) selectedDice.splice(pos, 1);
  else selectedDice.push(idx);
  renderDice(currentState);
}

$("btn-continue").addEventListener("click", async () => {
  if (!currentSession) return;
  selectedDice = [];
  try {
    const data = await api(`/api/game/${currentSession}/collect`, { method: "POST" });
    currentState = data;
    renderGame(data);
  } catch (e) {
    alert(e.message);
  }
});

$("btn-skip").addEventListener("click", async () => {
  if (!currentSession) return;
  try {
    const data = await api(`/api/game/${currentSession}/buy`, {
      method: "POST",
      body: JSON.stringify({ type: "skip" })
    });
    await endTurnAndAI(data);
  } catch (e) {
    alert(e.message);
  }
});

async function buyMarketItem(cardID) {
  if (!currentSession) return;
  try {
    const data = await api(`/api/game/${currentSession}/buy`, {
      method: "POST",
      body: JSON.stringify({ type: "market", card_id: cardID })
    });
    await endTurnAndAI(data);
  } catch (e) {
    alert(e.message);
  }
}

async function buyLandmarkItem(index) {
  if (!currentSession) return;
  try {
    const data = await api(`/api/game/${currentSession}/buy`, {
      method: "POST",
      body: JSON.stringify({ type: "landmark", index })
    });
    await endTurnAndAI(data);
  } catch (e) {
    alert(e.message);
  }
}

async function endTurnAndAI(buyState) {
  currentState = buyState;
  renderGame(buyState);
  if (buyState.game_over) return;

  try {
    const data = await api(`/api/game/${currentSession}/end-turn`, { method: "POST" });
    currentState = data;
    renderGame(data);
  } catch (e) {
    alert(e.message);
  }
}

$("btn-back").addEventListener("click", () => {
  currentSession = null;
  currentState = null;
  showScreen("session-screen");
  loadSessions();
});

/* ── Init ── */
loadSessions();
