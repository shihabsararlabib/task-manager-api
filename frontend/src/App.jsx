import React, { useEffect, useMemo, useState } from 'react'
import { api } from './api'

const TASK_STATUS = ['todo', 'in_progress', 'done']

function useSession() {
  const readStorage = () => {
    try {
      return localStorage.getItem('tm_session')
    } catch {
      return null
    }
  }

  const writeStorage = (value) => {
    try {
      if (value) localStorage.setItem('tm_session', JSON.stringify(value))
      else localStorage.removeItem('tm_session')
    } catch {
      // ignore storage failures (e.g. embedded browser restrictions)
    }
  }

  const [session, setSession] = useState(() => {
    const raw = readStorage()
    if (!raw) return null
    try {
      const parsed = JSON.parse(raw)
      if (!parsed || typeof parsed !== 'object') return null
      const hasAccessToken = typeof parsed.accessToken === 'string' && parsed.accessToken.length > 0
      const hasUser = parsed.user && typeof parsed.user === 'object'
      return hasAccessToken && hasUser ? parsed : null
    } catch {
      writeStorage(null)
      return null
    }
  })

  const save = (value) => {
    setSession(value)
    writeStorage(value)
  }

  return { session, save }
}

export default function App() {
  const { session, save } = useSession()
  const [mode, setMode] = useState('login')
  const [authForm, setAuthForm] = useState({ name: '', email: '', password: '' })
  const [taskForm, setTaskForm] = useState({ title: '', description: '' })
  const [tasks, setTasks] = useState([])
  const [users, setUsers] = useState([])
  const [statusFilter, setStatusFilter] = useState('')
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState('')

  const token = session?.accessToken
  const user = session?.user ?? null

  const taskQuery = useMemo(() => {
    const params = new URLSearchParams()
    if (statusFilter) params.set('status', statusFilter)
    const q = params.toString()
    return q ? `?${q}` : ''
  }, [statusFilter])

  const metrics = useMemo(() => {
    const total = tasks.length
    const todo = tasks.filter((t) => t.status === 'todo').length
    const inProgress = tasks.filter((t) => t.status === 'in_progress').length
    const done = tasks.filter((t) => t.status === 'done').length
    return { total, todo, inProgress, done }
  }, [tasks])

  useEffect(() => {
    if (token) void loadTasks()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token, taskQuery])

  async function loadTasks() {
    if (!token) return
    try {
      setLoading(true)
      const data = await api.listTasks(token, taskQuery)
      setTasks(Array.isArray(data) ? data : [])
    } catch (err) {
      setMessage(err.message)
    } finally {
      setLoading(false)
    }
  }

  async function loadUsers() {
    if (!token) return
    try {
      setLoading(true)
      const data = await api.listUsers(token)
      setUsers(Array.isArray(data) ? data : [])
    } catch (err) {
      setMessage(err.message)
    } finally {
      setLoading(false)
    }
  }

  async function handleAuthSubmit(e) {
    e.preventDefault()
    setMessage('')
    try {
      setLoading(true)
      const payload =
        mode === 'register'
          ? { name: authForm.name, email: authForm.email, password: authForm.password }
          : { email: authForm.email, password: authForm.password }

      const data = mode === 'register' ? await api.register(payload) : await api.login(payload)
      save({
        accessToken: data.access_token,
        refreshToken: data.refresh_token,
        user: data.user
      })
      setAuthForm({ name: '', email: '', password: '' })
    } catch (err) {
      setMessage(err.message)
    } finally {
      setLoading(false)
    }
  }

  async function loginAsAdmin() {
    setMessage('')
    try {
      setLoading(true)
      const data = await api.login({ email: 'admin@example.com', password: 'Admin123!' })
      save({
        accessToken: data.access_token,
        refreshToken: data.refresh_token,
        user: data.user
      })
      setMode('login')
      setAuthForm({ name: '', email: 'admin@example.com', password: 'Admin123!' })
    } catch (err) {
      setMessage(err.message)
    } finally {
      setLoading(false)
    }
  }

  async function handleLogout() {
    try {
      if (session?.refreshToken) {
        await api.logout(session.refreshToken)
      }
    } catch {
      // ignore logout failure
    }
    save(null)
    setTasks([])
    setUsers([])
  }

  async function createTask(e) {
    e.preventDefault()
    setMessage('')
    try {
      setLoading(true)
      await api.createTask(token, taskForm)
      setTaskForm({ title: '', description: '' })
      await loadTasks()
    } catch (err) {
      setMessage(err.message)
    } finally {
      setLoading(false)
    }
  }

  async function updateStatus(task, status) {
    try {
      setLoading(true)
      await api.updateTask(token, task.id, {
        title: task.title,
        description: task.description,
        status
      })
      await loadTasks()
    } catch (err) {
      setMessage(err.message)
    } finally {
      setLoading(false)
    }
  }

  async function removeTask(id) {
    try {
      setLoading(true)
      await api.deleteTask(token, id)
      await loadTasks()
    } catch (err) {
      setMessage(err.message)
    } finally {
      setLoading(false)
    }
  }

  if (!session || !token || !user) {
    return (
      <div className="auth-shell">
        <div className="container auth-layout">
          <section className="card auth-card auth-panel">
            <p className="eyebrow">Task Manager Pro</p>
            <h2>{mode === 'login' ? 'Welcome back' : 'Create your account'}</h2>
            <p className="subtext">{mode === 'login' ? 'Sign in to continue to your workspace.' : 'Get started in a few seconds.'}</p>
            <div className="tabs">
              <button type="button" onClick={() => setMode('login')} className={mode === 'login' ? 'active' : ''}>Login</button>
              <button type="button" onClick={() => setMode('register')} className={mode === 'register' ? 'active' : ''}>Register</button>
            </div>
            <form onSubmit={handleAuthSubmit} className="form">
              {mode === 'register' && (
                <input
                  placeholder="Full name"
                  value={authForm.name}
                  onChange={(e) => setAuthForm((p) => ({ ...p, name: e.target.value }))}
                  required
                />
              )}
              <input
                type="email"
                placeholder="Email address"
                value={authForm.email}
                onChange={(e) => setAuthForm((p) => ({ ...p, email: e.target.value }))}
                required
              />
              <input
                type="password"
                placeholder="Password"
                value={authForm.password}
                onChange={(e) => setAuthForm((p) => ({ ...p, password: e.target.value }))}
                required
              />
              <button className="primary" disabled={loading}>
                {loading ? 'Please wait...' : mode === 'login' ? 'Sign in' : 'Create account'}
              </button>
              <button type="button" className="ghost" onClick={loginAsAdmin} disabled={loading}>
                Use admin demo login
              </button>
            </form>
            <p className="legal-note">By continuing, you agree to your workspace security policy.</p>
            {message && <p className="error">{message}</p>}
          </section>

          <section className="hero card marketing-panel">
            <p className="eyebrow">Organize better</p>
            <h1>Manage your team’s tasks in one secure workspace.</h1>
            <p>Track progress, assign priorities, and ship work faster with an API-driven task platform.</p>
            <ul className="feature-list">
              <li><span>✓</span>JWT authentication</li>
              <li><span>✓</span>Admin user management</li>
              <li><span>✓</span>Status-based task pipeline</li>
            </ul>
          </section>
        </div>

        <section className="landing-features card">
          <div className="landing-head">
            <h3>What you can do on this platform</h3>
            <p className="subtext">All core features are already connected to the backend API.</p>
          </div>
          <div className="feature-grid">
            <article>
              <h4>Secure Authentication</h4>
              <p>Register, login, refresh, and logout flows using JWT access and refresh tokens.</p>
            </article>
            <article>
              <h4>Task Board</h4>
              <p>Create, update, filter, and delete tasks with real-time status changes.</p>
            </article>
            <article>
              <h4>Admin Controls</h4>
              <p>Admin users can load and review registered users and roles.</p>
            </article>
            <article>
              <h4>Production-ready API</h4>
              <p>Role-based access, PostgreSQL persistence, and migration-backed schema.</p>
            </article>
          </div>
          <div className="quickstart">
            <strong>Quick start:</strong>
            <span>Login with admin@example.com / Admin123!</span>
          </div>
        </section>
      </div>
    )
  }

  return (
    <div className="app-shell">
      <div className="container">
        <header className="header card">
          <div>
            <p className="eyebrow">Task Manager Pro</p>
            <h1 className="title">Workspace Dashboard</h1>
            <p className="subtext">Manage tasks with clarity and control.</p>
          </div>

          <div className="header-actions">
            <div className="identity-chip">
              <strong>{user?.name ?? 'User'}</strong>
              <span>{user?.role ?? 'user'}</span>
            </div>
            <button type="button" onClick={handleLogout}>Logout</button>
          </div>
        </header>

        <section className="metrics-grid">
          <article className="metric card"><span>Total Tasks</span><strong>{metrics.total}</strong></article>
          <article className="metric card"><span>Todo</span><strong>{metrics.todo}</strong></article>
          <article className="metric card"><span>In Progress</span><strong>{metrics.inProgress}</strong></article>
          <article className="metric card"><span>Done</span><strong>{metrics.done}</strong></article>
        </section>

        <div className="grid">
          <section className="card">
            <h2>Create Task</h2>
            <p className="subtext">Add a new item to your workflow.</p>
            <form onSubmit={createTask} className="form">
              <input
                placeholder="Task title"
                value={taskForm.title}
                onChange={(e) => setTaskForm((p) => ({ ...p, title: e.target.value }))}
                required
              />
              <textarea
                placeholder="Description (optional)"
                value={taskForm.description}
                onChange={(e) => setTaskForm((p) => ({ ...p, description: e.target.value }))}
              />
              <button className="primary" disabled={loading}>Create task</button>
            </form>
          </section>

          <section className="card tasks-card">
            <div className="tasks-head">
              <div>
                <h2>Task Board</h2>
                <p className="subtext">Update status and keep momentum.</p>
              </div>
              <div className="row">
                <select value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)}>
                  <option value="">All statuses</option>
                  {TASK_STATUS.map((s) => (
                    <option key={s} value={s}>{formatStatus(s)}</option>
                  ))}
                </select>
                <button type="button" onClick={loadTasks}>Refresh</button>
              </div>
            </div>

            <ul className="list">
              {tasks.map((task) => (
                <li key={task.id}>
                  <div className="task-content">
                    <strong>{task.title}</strong>
                    <p>{task.description || 'No description provided.'}</p>
                    <span className={`status-badge ${task.status}`}>{formatStatus(task.status)}</span>
                  </div>
                  <div className="row">
                    <select value={task.status} onChange={(e) => updateStatus(task, e.target.value)}>
                      {TASK_STATUS.map((s) => (
                        <option key={s} value={s}>{formatStatus(s)}</option>
                      ))}
                    </select>
                    <button type="button" onClick={() => removeTask(task.id)} className="danger">Delete</button>
                  </div>
                </li>
              ))}
              {!tasks.length && <li className="empty">No tasks yet. Create your first task.</li>}
            </ul>
          </section>

          {user?.role === 'admin' && (
            <section className="card admin-card">
              <div className="row between">
                <div>
                  <h2>Admin · Users</h2>
                  <p className="subtext">Inspect registered users and roles.</p>
                </div>
                <button type="button" onClick={loadUsers}>Load users</button>
              </div>
              <ul className="list compact">
                {users.map((u) => (
                  <li key={u.id}>
                    <div>
                      <strong>{u.name}</strong>
                      <p>{u.email}</p>
                    </div>
                    <span className="status-badge done">{u.role}</span>
                  </li>
                ))}
                {!users.length && <li className="empty">No users loaded.</li>}
              </ul>
            </section>
          )}
        </div>

        {message && <p className="error">{message}</p>}
      </div>
    </div>
  )
}

function formatStatus(status) {
  if (status === 'in_progress') return 'In Progress'
  if (status === 'todo') return 'Todo'
  if (status === 'done') return 'Done'
  return status
}
