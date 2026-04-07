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
      // ignore storage failures
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
  const token = session?.accessToken
  const user = session?.user ?? null

  const [view, setView] = useState(token ? 'dashboard' : 'home')
  const [authMode, setAuthMode] = useState('login')
  
  const [authForm, setAuthForm] = useState({ name: '', email: '', password: '' })
  const [taskForm, setTaskForm] = useState({ title: '', description: '' })
  const [tasks, setTasks] = useState([])
  const [users, setUsers] = useState([])
  const [statusFilter, setStatusFilter] = useState('')
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState('')

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
    if (token && view === 'dashboard') void loadTasks()
  }, [token, taskQuery, view])

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
        authMode === 'register'
          ? { name: authForm.name, email: authForm.email, password: authForm.password }
          : { email: authForm.email, password: authForm.password }

      const data = authMode === 'register' ? await api.register(payload) : await api.login(payload)
      save({
        accessToken: data.access_token,
        refreshToken: data.refresh_token,
        user: data.user
      })
      setAuthForm({ name: '', email: '', password: '' })
      setView('dashboard')
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
      setAuthMode('login')
      setAuthForm({ name: '', email: 'admin@example.com', password: 'Admin123!' })
      setView('dashboard')
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
      // ignore
    }
    save(null)
    setTasks([])
    setUsers([])
    setView('home')
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

  return (
    <div className="layout-shell fade-in">
      {/* Global Navigation */}
      <nav className="global-nav">
        <div className="nav-container">
          <div className="nav-brand" onClick={() => setView('home')}>
            <div className="brand-icon pulse"></div>
            <span>TaskManager Pro</span>
          </div>
          <div className="nav-links">
            <button className={`nav-item ${view === 'home' ? 'active' : ''}`} onClick={() => setView('home')}>Home</button>
            <button className={`nav-item ${view === 'about' ? 'active' : ''}`} onClick={() => setView('about')}>About</button>
          </div>
          <div className="nav-actions">
            {!token ? (
              <button className="primary nav-btn" onClick={() => { setView('auth'); setAuthMode('login'); }}>Sign in</button>
            ) : (
              <>
                <button className={`nav-item ${view === 'dashboard' ? 'active' : ''}`} onClick={() => setView('dashboard')}>Dashboard</button>
                <button className="ghost nav-btn" onClick={handleLogout}>Logout</button>
              </>
            )}
          </div>
        </div>
      </nav>

      {/* View Router Layer */}
      <main className="main-content">
        {view === 'home' && (
          <div className="page-animate">
            <section className="hero">
              <div className="glow-orb"></div>
              <p className="eyebrow fade-in-up" style={{animationDelay: '0.1s'}}>Organize better</p>
              <h1 className="fade-in-up" style={{animationDelay: '0.2s'}}>Manage your team’s tasks<br/>in one secure workspace.</h1>
              <p className="hero-subtext fade-in-up" style={{animationDelay: '0.3s'}}>
                Track progress, assign priorities, and ship work faster with a fully API-driven premium platform.
              </p>
              <div className="hero-actions fade-in-up" style={{animationDelay: '0.4s'}}>
                {!token ? (
                  <button className="primary large glow-btn" onClick={() => { setView('auth'); setAuthMode('register'); }}>Get Started Free</button>
                ) : (
                  <button className="primary large glow-btn" onClick={() => setView('dashboard')}>Go to Dashboard</button>
                )}
                <button className="ghost large" onClick={() => setView('about')}>Learn more</button>
              </div>
            </section>
            
            <section className="container">
              <div className="landing-head fade-in-up" style={{animationDelay: '0.5s'}}>
                <h3>What you can do on this platform</h3>
                <p className="subtext">All core features are elegantly animated and responsive.</p>
              </div>
              <div className="feature-grid">
                <article className="card float-hover fade-in-up" style={{animationDelay: '0.6s'}}>
                  <h4>Secure Authentication</h4>
                  <p>Register, login, refresh, and logout flows using JWT access and refresh tokens.</p>
                </article>
                <article className="card float-hover fade-in-up" style={{animationDelay: '0.7s'}}>
                  <h4>Task Board</h4>
                  <p>Create, update, filter, and delete tasks with real-time status changes and polished UI.</p>
                </article>
                <article className="card float-hover fade-in-up" style={{animationDelay: '0.8s'}}>
                  <h4>Admin Controls</h4>
                  <p>Admin users can load and review registered users and roles directly from the control panel.</p>
                </article>
              </div>
            </section>
          </div>
        )}

        {view === 'about' && (
          <div className="page-animate container about-page">
             <div className="card glass-card fade-in-up" style={{animationDelay: '0.1s'}}>
                <h2>About TaskManager Pro</h2>
                <p className="subtext">Task Manager Pro is an advanced digital workspace designed for high-performance teams to orchestrate their day-to-day operations seamlessly.</p>
                
                <h3 className="section-title">Our Architecture</h3>
                <p>This project utilizes a modern decoupled footprint:</p>
                <ul className="info-list">
                  <li><strong>Frontend:</strong> React + Vite with raw CSS to ensure lightning-fast browser rendering and beautiful micro-animations avoiding third-party CSS constraints.</li>
                  <li><strong>Backend:</strong> Golang RESTful API equipped with standard net/http and chi routing components for high throughput.</li>
                  <li><strong>Database:</strong> Scalable PostgreSQL persistence layer.</li>
                </ul>
                
                <h3 className="section-title">The Philosophy</h3>
                <p>We designed this environment focusing entirely on User Experience (UX), application performance, and premium glassmorphism visuals. Enjoy unparalleled speed coupled with a tailored dark aesthetic ideal for long developer sessions.</p>
             </div>
          </div>
        )}

        {view === 'auth' && !token && (
          <div className="page-animate container auth-layout fade-in-up">
            <section className="card auth-card glass-card">
              <p className="eyebrow">Task Manager Pro</p>
              <h2>{authMode === 'login' ? 'Welcome back' : 'Create your account'}</h2>
              <p className="subtext">{authMode === 'login' ? 'Sign in to continue to your workspace.' : 'Get started in a few seconds.'}</p>
              <div className="tabs">
                <button type="button" onClick={() => setAuthMode('login')} className={authMode === 'login' ? 'active' : ''}>Login</button>
                <button type="button" onClick={() => setAuthMode('register')} className={authMode === 'register' ? 'active' : ''}>Register</button>
              </div>
              <form onSubmit={handleAuthSubmit} className="form">
                {authMode === 'register' && (
                  <input
                    className="input-field"
                    placeholder="Full name"
                    value={authForm.name}
                    onChange={(e) => setAuthForm((p) => ({ ...p, name: e.target.value }))}
                    required
                  />
                )}
                <input
                  className="input-field"
                  type="email"
                  placeholder="Email address"
                  value={authForm.email}
                  onChange={(e) => setAuthForm((p) => ({ ...p, email: e.target.value }))}
                  required
                />
                <input
                  className="input-field"
                  type="password"
                  placeholder="Password"
                  value={authForm.password}
                  onChange={(e) => setAuthForm((p) => ({ ...p, password: e.target.value }))}
                  required
                />
                <button className="primary submit-btn glow-btn" disabled={loading}>
                  {loading ? <span className="spinner"></span> : authMode === 'login' ? 'Sign in' : 'Create account'}
                </button>
                <button type="button" className="ghost" onClick={loginAsAdmin} disabled={loading}>
                  Use admin demo login
                </button>
              </form>
              <p className="legal-note">By continuing, you agree to your workspace security policy.</p>
              {message && <p className="error shake">{message}</p>}
            </section>
          </div>
        )}

        {view === 'dashboard' && token && user && (
          <div className="page-animate container dashboard-layout">
            <header className="dashboard-header fade-in-up" style={{animationDelay: '0.1s'}}>
              <div>
                <p className="eyebrow">Interactive Space</p>
                <h1 className="title">Workspace Dashboard</h1>
                <p className="subtext">Manage tasks with clarity and control.</p>
              </div>
              <div className="identity-chip">
                <strong>{user.name}</strong>
                <span className="role-chip">{user.role}</span>
              </div>
            </header>

            <section className="metrics-grid fade-in-up" style={{animationDelay: '0.2s'}}>
              <article className="metric glass-card hover-glow"><span>Total Tasks</span><strong>{metrics.total}</strong></article>
              <article className="metric glass-card hover-glow"><span>Todo</span><strong>{metrics.todo}</strong></article>
              <article className="metric glass-card hover-glow"><span>In Progress</span><strong>{metrics.inProgress}</strong></article>
              <article className="metric glass-card hover-glow"><span>Done</span><strong>{metrics.done}</strong></article>
            </section>

            <div className="grid">
              <section className="card glass-card panel-left fade-in-up" style={{animationDelay: '0.3s'}}>
                <h2>Create Task</h2>
                <p className="subtext">Add a new item to your workflow.</p>
                <form onSubmit={createTask} className="form">
                  <input
                    className="input-field"
                    placeholder="Task title"
                    value={taskForm.title}
                    onChange={(e) => setTaskForm((p) => ({ ...p, title: e.target.value }))}
                    required
                  />
                  <textarea
                    className="input-field textarea-field"
                    placeholder="Description (optional)"
                    value={taskForm.description}
                    onChange={(e) => setTaskForm((p) => ({ ...p, description: e.target.value }))}
                  />
                  <button className="primary submit-btn glow-btn target-glow" disabled={loading}>
                     {loading ? <span className="spinner"></span> : 'Create task'}
                  </button>
                </form>
              </section>

              <section className="card glass-card tasks-card fade-in-up" style={{animationDelay: '0.4s'}}>
                <div className="tasks-head">
                  <div>
                    <h2>Task Board</h2>
                    <p className="subtext">Update status and keep momentum.</p>
                  </div>
                  <div className="row">
                    <select className="dropdown-select" value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)}>
                      <option value="">All statuses</option>
                      {TASK_STATUS.map((s) => (
                        <option key={s} value={s}>{formatStatus(s)}</option>
                      ))}
                    </select>
                    <button type="button" onClick={loadTasks}>Refresh</button>
                  </div>
                </div>

                <ul className="list">
                  {tasks.map((task, idx) => (
                    <li key={task.id} className="task-item scale-in" style={{animationDelay: `${0.05 * idx}s`}}>
                      <div className="task-content">
                        <strong>{task.title}</strong>
                        <p>{task.description || 'No description provided.'}</p>
                        <span className={`status-badge ${task.status}`}>{formatStatus(task.status)}</span>
                      </div>
                      <div className="row">
                        <select className="dropdown-select" value={task.status} onChange={(e) => updateStatus(task, e.target.value)}>
                          {TASK_STATUS.map((s) => (
                            <option key={s} value={s}>{formatStatus(s)}</option>
                          ))}
                        </select>
                        <button type="button" onClick={() => removeTask(task.id)} className="danger">Delete</button>
                      </div>
                    </li>
                  ))}
                  {!tasks.length && <li className="empty fade-in">No tasks yet. Create your first task.</li>}
                </ul>
              </section>

              {user?.role === 'admin' && (
                <section className="card glass-card admin-card fade-in-up" style={{animationDelay: '0.5s', gridColumn: '1 / -1'}}>
                  <div className="row between">
                    <div>
                      <h2>Admin · Users</h2>
                      <p className="subtext">Inspect registered users and roles.</p>
                    </div>
                    <button type="button" onClick={loadUsers}>Load users</button>
                  </div>
                  <ul className="list compact">
                    {users.map((u, idx) => (
                      <li key={u.id} className="user-item scale-in" style={{animationDelay: `${0.05 * idx}s`}}>
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
            {message && <p className="error shake" style={{marginTop: '24px'}}>{message}</p>}
          </div>
        )}
      </main>
    </div>
  )
}

function formatStatus(status) {
  if (status === 'in_progress') return 'In Progress'
  if (status === 'todo') return 'Todo'
  if (status === 'done') return 'Done'
  return status
}
