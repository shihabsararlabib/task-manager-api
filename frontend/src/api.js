const DEFAULT_HOST = typeof window !== 'undefined' ? window.location.hostname : 'localhost'
const envApiUrl = import.meta.env.VITE_API_URL

function resolveApiUrl() {
  if (!envApiUrl) {
    return `http://${DEFAULT_HOST}:8080`
  }

  if (typeof window !== 'undefined') {
    try {
      const parsed = new URL(envApiUrl)
      const isLocalhost = parsed.hostname === 'localhost' || parsed.hostname === '127.0.0.1'
      const browserIsRemoteHost = DEFAULT_HOST !== 'localhost' && DEFAULT_HOST !== '127.0.0.1'
      if (isLocalhost && browserIsRemoteHost) {
        parsed.hostname = DEFAULT_HOST
        return parsed.toString().replace(/\/$/, '')
      }
    } catch {
      // fall through to original configured value
    }
  }

  return envApiUrl
}

const API_URL = resolveApiUrl()

function parseJsonSafe(text) {
  try {
    return text ? JSON.parse(text) : null
  } catch {
    return null
  }
}

async function request(path, { method = 'GET', token, body } = {}) {
  const headers = { 'Content-Type': 'application/json' }
  if (token) headers.Authorization = `Bearer ${token}`

  let res
  try {
    res = await fetch(`${API_URL}${path}`, {
      method,
      headers,
      body: body ? JSON.stringify(body) : undefined
    })
  } catch {
    throw new Error(`Cannot reach API at ${API_URL}. Check backend/CORS configuration.`)
  }

  const text = await res.text()
  const data = parseJsonSafe(text)

  if (!res.ok) {
    const message = data?.error || `Request failed (${res.status})`
    throw new Error(message)
  }

  return data
}

export const api = {
  register: (payload) => request('/auth/register', { method: 'POST', body: payload }),
  login: (payload) => request('/auth/login', { method: 'POST', body: payload }),
  refresh: (refreshToken) => request('/auth/refresh', { method: 'POST', body: { refresh_token: refreshToken } }),
  logout: (refreshToken) => request('/auth/logout', { method: 'POST', body: { refresh_token: refreshToken } }),

  listTasks: (token, query = '') => request(`/tasks/${query}`, { token }),
  createTask: (token, payload) => request('/tasks/', { method: 'POST', token, body: payload }),
  updateTask: (token, id, payload) => request(`/tasks/${id}`, { method: 'PUT', token, body: payload }),
  deleteTask: (token, id) => request(`/tasks/${id}`, { method: 'DELETE', token }),

  listUsers: (token) => request('/admin/users', { token })
}
