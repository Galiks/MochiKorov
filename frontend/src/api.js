export async function api(path, opts = {}) {
  const segments = path.split('/')
  const sessionId = segments[3]
  const tokenKey = sessionId ? `mochi_token_${sessionId}` : null
  const token = tokenKey ? localStorage.getItem(tokenKey) : null

  const headers = { 'Content-Type': 'application/json' }
  if (token) headers['X-Player-Token'] = token

  const res = await fetch(path, {
    headers,
    ...opts,
  })
  const data = await res.json()
  if (!res.ok) throw new Error(data.error || res.statusText)
  return data
}
