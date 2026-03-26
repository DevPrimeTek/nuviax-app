'use client'
import { useEffect, useState, useCallback } from 'react'
import { useRouter } from 'next/navigation'

// ── Types ──────────────────────────────────────────────────────────────────────

interface PlatformStats {
  total_users: number
  admin_users: number
  new_users_7d: number
  new_users_30d: number
  active_goals: number
  completed_goals: number
  paused_goals: number
  total_goals: number
  active_sprints: number
  completed_sprints: number
  tasks_today: number
  tasks_completed_today: number
  srm_events_30d: number
  srm_l3_events_30d: number
  regression_events_30d: number
  ceremonies_30d: number
  badges_awarded_30d: number
  computed_at: string
}

interface AdminUser {
  id: string
  full_name: string | null
  locale: string
  is_active: boolean
  is_admin: boolean
  mfa_enabled: boolean
  created_at: string
  active_goals: number
  completed_goals: number
  total_goals: number
  completed_sprints: number
  tasks_last_30d: number
  last_active_at: string | null
  active_sessions: number
}

interface AuditEntry {
  id: string
  user_id: string | null
  full_name: string | null
  action: string
  created_at: string
}

interface HealthData {
  status: string
  database: {
    version: string
    table_count: number
    pool_total_conns: number
    pool_idle_conns: number
    pool_acquired_conns: number
    pool_max_conns: number
  }
  scheduler: { jobs_last_24h: number }
  environment: string
}

type Tab = 'stats' | 'users' | 'audit' | 'health'

// ── Helpers ───────────────────────────────────────────────────────────────────

function api(path: string, init: RequestInit = {}) {
  return fetch(`/api/proxy/admin${path}`, {
    ...init,
    headers: { 'Content-Type': 'application/json', ...(init.headers as object) },
  }).then(async r => {
    if (r.status === 404) throw new Error('NOT_ADMIN')
    if (!r.ok) { const j = await r.json().catch(() => ({})); throw new Error(j.error || r.statusText) }
    return r.json()
  })
}

function fmt(n: number) { return n.toLocaleString('ro-RO') }
function fmtDate(s: string | null) {
  if (!s) return '—'
  return new Date(s).toLocaleDateString('ro-RO', { day: 'numeric', month: 'short', year: 'numeric' })
}
function fmtDateTime(s: string | null) {
  if (!s) return '—'
  return new Date(s).toLocaleString('ro-RO', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' })
}

// ── Standalone Admin Shell ─────────────────────────────────────────────────────

function AdminShell({ userName, onRefresh, children }: {
  userName: string
  onRefresh: () => void
  children: React.ReactNode
}) {
  const router = useRouter()

  async function handleLogout() {
    await fetch('/api/auth/logout', { method: 'POST' }).catch(() => {})
    router.push('/auth/login')
  }

  return (
    <div style={{ minHeight: '100vh', background: '#0a0a0f', color: '#e8e8f0', fontFamily: 'system-ui, -apple-system, sans-serif' }}>
      {/* Top bar */}
      <div style={{
        position: 'sticky', top: 0, zIndex: 100,
        background: 'rgba(10,10,15,.96)', backdropFilter: 'blur(12px)',
        borderBottom: '1px solid rgba(255,255,255,.08)',
        padding: '0 24px',
      }}>
        <div style={{
          maxWidth: 1100, margin: '0 auto',
          display: 'flex', alignItems: 'center', justifyContent: 'space-between',
          height: 52,
        }}>
          {/* Logo + label */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <div style={{
              width: 28, height: 28, borderRadius: 7,
              background: 'linear-gradient(135deg, #ff6b35 0%, #ff9a3c 100%)',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              fontSize: 14, fontWeight: 900, color: '#fff',
            }}>N</div>
            <span style={{ fontWeight: 800, fontSize: 15, color: '#fff', letterSpacing: '-0.02em' }}>
              NuviaX
            </span>
            <span style={{
              fontSize: 10, fontWeight: 700, padding: '2px 7px', borderRadius: 5,
              background: 'rgba(99,102,241,.2)', color: '#818cf8',
              letterSpacing: '.06em', textTransform: 'uppercase',
            }}>ADMIN</span>
          </div>

          {/* Right: user + actions */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <span style={{ fontSize: 13, color: 'rgba(255,255,255,.5)' }}>{userName}</span>
            <button onClick={onRefresh} style={{
              padding: '5px 12px', borderRadius: 7,
              border: '1px solid rgba(255,255,255,.1)',
              background: 'rgba(255,255,255,.05)', color: 'rgba(255,255,255,.6)',
              fontSize: 12, fontWeight: 600, cursor: 'pointer',
            }}>↺ Refresh</button>
            <button onClick={handleLogout} style={{
              padding: '5px 12px', borderRadius: 7,
              border: '1px solid rgba(239,68,68,.2)',
              background: 'rgba(239,68,68,.08)', color: '#ef4444',
              fontSize: 12, fontWeight: 600, cursor: 'pointer',
            }}>Deconectare</button>
          </div>
        </div>
      </div>

      {/* Content */}
      <div style={{ maxWidth: 1100, margin: '0 auto', padding: '28px 24px 60px' }}>
        {children}
      </div>
    </div>
  )
}

// ── Components ────────────────────────────────────────────────────────────────

function StatCard({ label, value, sub, color }: { label: string; value: string | number; sub?: string; color?: string }) {
  return (
    <div style={{
      background: 'rgba(255,255,255,.04)', border: '1px solid rgba(255,255,255,.08)', borderRadius: 12,
      padding: '14px 16px', display: 'flex', flexDirection: 'column', gap: 4,
    }}>
      <div style={{ fontSize: 11.5, color: 'rgba(255,255,255,.4)', fontWeight: 500 }}>{label}</div>
      <div style={{ fontSize: 26, fontWeight: 800, color: color || '#e8e8f0', letterSpacing: '-0.03em' }}>
        {typeof value === 'number' ? fmt(value) : value}
      </div>
      {sub && <div style={{ fontSize: 11.5, color: 'rgba(255,255,255,.3)' }}>{sub}</div>}
    </div>
  )
}

function TabBtn({ id, current, label, onClick }: { id: Tab; current: Tab; label: string; onClick: (t: Tab) => void }) {
  const on = id === current
  return (
    <button onClick={() => onClick(id)} style={{
      padding: '7px 16px', borderRadius: 8, border: 'none', cursor: 'pointer',
      fontSize: 13, fontWeight: 600,
      background: on ? 'rgba(255,107,53,.15)' : 'transparent',
      color: on ? '#ff9a3c' : 'rgba(255,255,255,.4)',
      transition: 'all .15s',
    }}>{label}</button>
  )
}

function Badge({ label, type }: { label: string; type: 'success' | 'warn' | 'error' | 'info' | 'neutral' }) {
  const colors: Record<string, [string, string]> = {
    success: ['rgba(16,185,129,.15)', '#10b981'],
    warn:    ['rgba(245,158,11,.15)', '#f59e0b'],
    error:   ['rgba(239,68,68,.15)',  '#ef4444'],
    info:    ['rgba(99,102,241,.15)', '#818cf8'],
    neutral: ['rgba(255,255,255,.07)', 'rgba(255,255,255,.4)'],
  }
  const [bg, fg] = colors[type]
  return (
    <span style={{
      display: 'inline-block', padding: '2px 8px', borderRadius: 5,
      fontSize: 11, fontWeight: 700, background: bg, color: fg,
    }}>{label}</span>
  )
}

function SectionLabel({ children }: { children: React.ReactNode }) {
  return (
    <div style={{
      fontSize: 11, fontWeight: 700, color: 'rgba(255,255,255,.3)',
      textTransform: 'uppercase', letterSpacing: '.07em', marginBottom: 12,
    }}>{children}</div>
  )
}

// ── Main Page ─────────────────────────────────────────────────────────────────

export default function AdminPage() {
  const router = useRouter()
  const [tab, setTab] = useState<Tab>('stats')
  const [stats, setStats] = useState<PlatformStats | null>(null)
  const [users, setUsers] = useState<AdminUser[]>([])
  const [audit, setAudit] = useState<AuditEntry[]>([])
  const [health, setHealth] = useState<HealthData | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [userName, setUserName] = useState('Admin')
  const [resetConfirm, setResetConfirm] = useState(false)
  const [resetLoading, setResetLoading] = useState(false)
  const [resetResult, setResetResult] = useState<string | null>(null)
  const [userMsg, setUserMsg] = useState<Record<string, string>>({})
  const [searchQ, setSearchQ] = useState('')

  const loadStats = useCallback(() => {
    setLoading(true)
    api('/stats').then(d => { setStats(d); setLoading(false) })
      .catch(e => { setError(e.message === 'NOT_ADMIN' ? 'Acces interzis.' : e.message); setLoading(false) })
  }, [])

  const loadUsers = useCallback(() => {
    api('/users').then(d => setUsers(d.users || [])).catch(() => {})
  }, [])

  const loadAudit = useCallback(() => {
    api('/audit?limit=200').then(d => setAudit(d.entries || [])).catch(() => {})
  }, [])

  const loadHealth = useCallback(() => {
    api('/health').then(d => setHealth(d)).catch(() => {})
  }, [])

  const handleRefresh = useCallback(() => {
    loadStats(); loadUsers(); loadAudit(); loadHealth()
  }, [loadStats, loadUsers, loadAudit, loadHealth])

  useEffect(() => {
    fetch('/api/proxy/settings').then(r => r.ok ? r.json() : null).then(d => {
      if (d?.full_name) setUserName(d.full_name)
    }).catch(() => {})
    loadStats()
    loadUsers()
    loadAudit()
    loadHealth()
  }, [loadStats, loadUsers, loadAudit, loadHealth])

  function handleUserAction(userId: string, action: 'deactivate' | 'activate' | 'promote') {
    const labels = { deactivate: 'dezactivat', activate: 'activat', promote: 'promovat admin' }
    api(`/users/${userId}/${action}`, { method: 'POST' })
      .then(() => {
        setUserMsg(prev => ({ ...prev, [userId]: `Utilizator ${labels[action]}.` }))
        loadUsers()
        setTimeout(() => setUserMsg(prev => { const n = { ...prev }; delete n[userId]; return n }), 3000)
      })
      .catch(e => setUserMsg(prev => ({ ...prev, [userId]: `Eroare: ${e.message}` })))
  }

  async function handleDevReset() {
    setResetLoading(true)
    try {
      const d = await api('/db/reset', {
        method: 'POST',
        body: JSON.stringify({ confirm_text: 'RESET_ALL_DATA' }),
      })
      setResetResult(`✓ Reset: ${d.deleted_users} utilizatori, ${d.deleted_goals} obiective, ${d.deleted_tasks} activități șterse.`)
      setResetConfirm(false)
      loadStats(); loadUsers(); loadAudit()
    } catch (e: unknown) {
      const err = e instanceof Error ? e.message : 'Eroare necunoscută'
      setResetResult(`✗ Eroare: ${err}`)
    } finally {
      setResetLoading(false)
    }
  }

  // ── Error state (not admin / not logged in) ────────────────────────────────

  if (error) {
    return (
      <div style={{
        minHeight: '100vh', background: '#0a0a0f',
        display: 'flex', alignItems: 'center', justifyContent: 'center',
      }}>
        <div style={{ textAlign: 'center' }}>
          <div style={{ fontSize: 52, marginBottom: 16 }}>🔒</div>
          <div style={{ fontSize: 18, fontWeight: 700, color: '#e8e8f0', marginBottom: 8 }}>
            {error === 'Acces interzis.' ? 'Acces restricționat' : error}
          </div>
          <div style={{ fontSize: 14, color: 'rgba(255,255,255,.35)', marginBottom: 24 }}>
            {error === 'Acces interzis.'
              ? 'Contul tău nu are drepturi de administrator.'
              : 'Trebuie să fii autentificat pentru a accesa panoul de administrare.'}
          </div>
          <div style={{ display: 'flex', gap: 10, justifyContent: 'center' }}>
            <button onClick={() => router.push('/dashboard')} style={{
              padding: '10px 20px', borderRadius: 9,
              border: '1px solid rgba(255,255,255,.1)', background: 'rgba(255,255,255,.06)',
              color: 'rgba(255,255,255,.7)', fontWeight: 600, fontSize: 14, cursor: 'pointer',
            }}>← Dashboard</button>
            <button onClick={() => router.push('/auth/login')} style={{
              padding: '10px 20px', borderRadius: 9, border: 'none',
              background: 'linear-gradient(135deg, #ff6b35, #ff9a3c)',
              color: '#fff', fontWeight: 700, fontSize: 14, cursor: 'pointer',
            }}>Login</button>
          </div>
        </div>
      </div>
    )
  }

  const filteredUsers = users.filter(u => {
    if (!searchQ) return true
    const q = searchQ.toLowerCase()
    return (u.full_name || '').toLowerCase().includes(q) || u.id.toLowerCase().includes(q)
  })

  return (
    <AdminShell userName={userName} onRefresh={handleRefresh}>

      {/* Page header */}
      <div style={{ marginBottom: 24 }}>
        <div style={{ fontWeight: 800, fontSize: 22, color: '#fff', letterSpacing: '-0.03em', marginBottom: 4 }}>
          Panel Administrare
        </div>
        <div style={{ fontSize: 13, color: 'rgba(255,255,255,.35)' }}>
          {stats ? `Ultima actualizare: ${fmtDateTime(stats.computed_at)}` : 'Se încarcă...'}
        </div>
      </div>

      {/* Tabs */}
      <div style={{
        display: 'flex', gap: 2, marginBottom: 24,
        background: 'rgba(255,255,255,.04)', borderRadius: 10, padding: 4,
        width: 'fit-content',
      }}>
        <TabBtn id="stats"  current={tab} label="📊 Statistici"   onClick={setTab} />
        <TabBtn id="users"  current={tab} label="👥 Utilizatori"  onClick={setTab} />
        <TabBtn id="audit"  current={tab} label="📋 Audit"        onClick={setTab} />
        <TabBtn id="health" current={tab} label="🔧 Sistem"       onClick={setTab} />
      </div>

      {/* ── STATS TAB ── */}
      {tab === 'stats' && (
        <div>
          {loading ? (
            <div style={{ textAlign: 'center', color: 'rgba(255,255,255,.3)', padding: '60px 0' }}>
              Se încarcă statisticile...
            </div>
          ) : stats ? (
            <>
              <SectionLabel>Utilizatori</SectionLabel>
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(150px, 1fr))', gap: 10, marginBottom: 28 }}>
                <StatCard label="Total utilizatori" value={stats.total_users} />
                <StatCard label="Noi (7 zile)"     value={stats.new_users_7d}  color="#ff9a3c" />
                <StatCard label="Noi (30 zile)"    value={stats.new_users_30d} color="#ff9a3c" />
                <StatCard label="Admini"           value={stats.admin_users}   color="#818cf8" />
              </div>

              <SectionLabel>Obiective &amp; Sprinturi</SectionLabel>
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(150px, 1fr))', gap: 10, marginBottom: 28 }}>
                <StatCard label="Obiective active"     value={stats.active_goals}    color="#10b981" />
                <StatCard label="Finalizate"           value={stats.completed_goals} color="#10b981" />
                <StatCard label="În pauză"             value={stats.paused_goals} />
                <StatCard label="Total obiective"      value={stats.total_goals} />
                <StatCard label="Sprinturi active"     value={stats.active_sprints}    color="#ff9a3c" />
                <StatCard label="Sprinturi finalizate" value={stats.completed_sprints} color="#ff9a3c" />
              </div>

              <SectionLabel>Activitate Zilnică</SectionLabel>
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(150px, 1fr))', gap: 10, marginBottom: 28 }}>
                <StatCard label="Activități azi"       value={stats.tasks_today} />
                <StatCard label="Completate azi"       value={stats.tasks_completed_today} color="#10b981" />
                <StatCard
                  label="Rată completare azi"
                  value={stats.tasks_today > 0 ? `${Math.round((stats.tasks_completed_today / stats.tasks_today) * 100)}%` : '—'}
                  color={stats.tasks_today > 0 && stats.tasks_completed_today / stats.tasks_today >= 0.7 ? '#10b981' : undefined}
                />
              </div>

              <SectionLabel>Sănătatea Sistemului (30 zile)</SectionLabel>
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(150px, 1fr))', gap: 10 }}>
                <StatCard label="Evenimente SRM"      value={stats.srm_events_30d} />
                <StatCard label="SRM Level 3"         value={stats.srm_l3_events_30d}     color={stats.srm_l3_events_30d > 0 ? '#f59e0b' : undefined} />
                <StatCard label="Regresii detectate"  value={stats.regression_events_30d} color={stats.regression_events_30d > 0 ? '#ef4444' : undefined} />
                <StatCard label="Ceremonii generate"  value={stats.ceremonies_30d}        color="#818cf8" />
                <StatCard label="Badge-uri acordate"  value={stats.badges_awarded_30d}    color="#818cf8" />
              </div>
            </>
          ) : null}
        </div>
      )}

      {/* ── USERS TAB ── */}
      {tab === 'users' && (
        <div>
          <div style={{ marginBottom: 14, display: 'flex', gap: 10, alignItems: 'center', flexWrap: 'wrap' }}>
            <input
              placeholder="Caută după nume sau ID..."
              value={searchQ}
              onChange={e => setSearchQ(e.target.value)}
              style={{
                flex: 1, minWidth: 220, padding: '9px 14px', borderRadius: 9,
                border: '1px solid rgba(255,255,255,.1)', background: 'rgba(255,255,255,.05)',
                color: '#e8e8f0', fontSize: 13, outline: 'none',
              }}
            />
            <span style={{ fontSize: 13, color: 'rgba(255,255,255,.35)' }}>{filteredUsers.length} utilizatori</span>
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            {filteredUsers.map(u => (
              <div key={u.id} style={{
                background: 'rgba(255,255,255,.04)', border: '1px solid rgba(255,255,255,.08)', borderRadius: 12,
                padding: '14px 16px',
                opacity: u.is_active ? 1 : 0.55,
              }}>
                <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 12, flexWrap: 'wrap' }}>
                  <div style={{ flex: 1 }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4, flexWrap: 'wrap' }}>
                      <span style={{ fontWeight: 700, color: '#e8e8f0', fontSize: 14 }}>
                        {u.full_name || '(fără nume)'}
                      </span>
                      {u.is_admin && <Badge label="ADMIN" type="info" />}
                      {!u.is_active && <Badge label="INACTIV" type="error" />}
                      {u.mfa_enabled && <Badge label="MFA" type="success" />}
                      <Badge label={u.locale.toUpperCase()} type="neutral" />
                    </div>
                    <div style={{ fontSize: 11, color: 'rgba(255,255,255,.25)', fontFamily: 'monospace', marginBottom: 8 }}>
                      {u.id}
                    </div>
                    <div style={{ display: 'flex', gap: 14, flexWrap: 'wrap' }}>
                      {[
                        `🎯 ${u.active_goals} active / ${u.total_goals} total`,
                        `⚡ ${u.tasks_last_30d} activități/30z`,
                        `🏃 ${u.completed_sprints} sprinturi`,
                        `🔗 ${u.active_sessions} sesiuni`,
                        `📅 Creat ${fmtDate(u.created_at)}`,
                        `⏱ Activ ${fmtDateTime(u.last_active_at)}`,
                      ].map(text => (
                        <span key={text} style={{ fontSize: 12, color: 'rgba(255,255,255,.35)' }}>{text}</span>
                      ))}
                    </div>
                  </div>

                  {!u.is_admin && (
                    <div style={{ display: 'flex', gap: 6, flexShrink: 0 }}>
                      {u.is_active ? (
                        <button onClick={() => handleUserAction(u.id, 'deactivate')} style={{
                          padding: '6px 12px', borderRadius: 7,
                          border: '1px solid rgba(239,68,68,.25)',
                          background: 'rgba(239,68,68,.1)', color: '#ef4444',
                          fontSize: 12, fontWeight: 600, cursor: 'pointer',
                        }}>Dezactivează</button>
                      ) : (
                        <button onClick={() => handleUserAction(u.id, 'activate')} style={{
                          padding: '6px 12px', borderRadius: 7,
                          border: '1px solid rgba(16,185,129,.25)',
                          background: 'rgba(16,185,129,.1)', color: '#10b981',
                          fontSize: 12, fontWeight: 600, cursor: 'pointer',
                        }}>Activează</button>
                      )}
                      <button onClick={() => { if (confirm('Promovezi acest utilizator la admin?')) handleUserAction(u.id, 'promote') }} style={{
                        padding: '6px 12px', borderRadius: 7,
                        border: '1px solid rgba(255,255,255,.1)',
                        background: 'rgba(255,255,255,.05)', color: 'rgba(255,255,255,.5)',
                        fontSize: 12, fontWeight: 600, cursor: 'pointer',
                      }}>→ Admin</button>
                    </div>
                  )}
                </div>

                {userMsg[u.id] && (
                  <div style={{
                    marginTop: 8, fontSize: 12, fontWeight: 600,
                    color: userMsg[u.id].startsWith('✓') ? '#10b981' : '#ef4444',
                  }}>
                    {userMsg[u.id]}
                  </div>
                )}
              </div>
            ))}

            {filteredUsers.length === 0 && (
              <div style={{ textAlign: 'center', color: 'rgba(255,255,255,.3)', padding: '40px 0' }}>
                {searchQ ? 'Niciun utilizator găsit.' : 'Nu există utilizatori înregistrați.'}
              </div>
            )}
          </div>
        </div>
      )}

      {/* ── AUDIT TAB ── */}
      {tab === 'audit' && (
        <div>
          <div style={{ marginBottom: 12, fontSize: 13, color: 'rgba(255,255,255,.35)' }}>
            Ultimele {audit.length} intrări din jurnalul de audit
          </div>
          <div style={{
            background: 'rgba(255,255,255,.03)', border: '1px solid rgba(255,255,255,.08)',
            borderRadius: 12, overflow: 'hidden',
          }}>
            {audit.length === 0 ? (
              <div style={{ padding: '28px', textAlign: 'center', color: 'rgba(255,255,255,.3)' }}>Jurnalul este gol.</div>
            ) : (
              <div style={{ overflowX: 'auto' }}>
                <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 13 }}>
                  <thead>
                    <tr style={{ borderBottom: '1px solid rgba(255,255,255,.08)' }}>
                      {['Data', 'Utilizator', 'Acțiune'].map(h => (
                        <th key={h} style={{
                          padding: '10px 16px', textAlign: 'left',
                          fontSize: 11, fontWeight: 700, color: 'rgba(255,255,255,.3)',
                          textTransform: 'uppercase', letterSpacing: '.05em',
                        }}>{h}</th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {audit.map((e, i) => (
                      <tr key={e.id} style={{
                        borderBottom: i < audit.length - 1 ? '1px solid rgba(255,255,255,.05)' : 'none',
                      }}>
                        <td style={{ padding: '9px 16px', color: 'rgba(255,255,255,.35)', whiteSpace: 'nowrap', fontSize: 12 }}>
                          {fmtDateTime(e.created_at)}
                        </td>
                        <td style={{ padding: '9px 16px', color: 'rgba(255,255,255,.6)', fontSize: 13 }}>
                          {e.full_name || (e.user_id ? e.user_id.substring(0, 8) + '…' : '—')}
                        </td>
                        <td style={{ padding: '9px 16px' }}>
                          <Badge
                            label={e.action}
                            type={
                              e.action.startsWith('ADMIN_') ? 'warn' :
                              e.action === 'REGISTER' ? 'success' :
                              e.action === 'LOGIN' ? 'info' :
                              e.action === 'LOGOUT' ? 'neutral' : 'neutral'
                            }
                          />
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </div>
      )}

      {/* ── HEALTH TAB ── */}
      {tab === 'health' && (
        <div>
          {health ? (
            <>
              <SectionLabel>Stare Sistem</SectionLabel>
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(150px, 1fr))', gap: 10, marginBottom: 24 }}>
                <StatCard label="Status"                  value={health.status === 'ok' ? '✓ OK' : '⚠ Degradat'} color={health.status === 'ok' ? '#10b981' : '#f59e0b'} />
                <StatCard label="Mediu"                   value={health.environment || 'production'} />
                <StatCard label="Tabele DB"               value={health.database.table_count} />
                <StatCard label="Joburi scheduler (24h)"  value={health.scheduler.jobs_last_24h} />
                <StatCard label="Conexiuni total"         value={health.database.pool_total_conns} />
                <StatCard label="Conexiuni idle"          value={health.database.pool_idle_conns} />
                <StatCard label="Conexiuni active"        value={health.database.pool_acquired_conns} />
                <StatCard label="Max pool"                value={health.database.pool_max_conns} />
              </div>

              <SectionLabel>Versiune PostgreSQL</SectionLabel>
              <div style={{
                background: 'rgba(255,255,255,.03)', border: '1px solid rgba(255,255,255,.08)',
                borderRadius: 10, padding: '12px 16px', marginBottom: 24,
              }}>
                <div style={{ fontSize: 12, color: 'rgba(255,255,255,.4)', fontFamily: 'monospace', wordBreak: 'break-all' }}>
                  {health.database.version}
                </div>
              </div>

              {/* DEV RESET */}
              {health.environment === 'development' && (
                <div style={{
                  background: 'rgba(239,68,68,.06)', border: '1px solid rgba(239,68,68,.2)',
                  borderRadius: 12, padding: '16px 18px',
                }}>
                  <div style={{ fontWeight: 700, color: '#ef4444', marginBottom: 6, fontSize: 14 }}>
                    ⚠ Resetare Bază de Date (Development Only)
                  </div>
                  <div style={{ fontSize: 13, color: 'rgba(255,255,255,.4)', marginBottom: 14 }}>
                    Șterge TOȚI utilizatorii non-admin și datele lor. Schema și conturile admin sunt păstrate.
                  </div>

                  {resetResult && (
                    <div style={{
                      padding: '10px 14px', borderRadius: 8, marginBottom: 12,
                      background: resetResult.startsWith('✓') ? 'rgba(16,185,129,.1)' : 'rgba(239,68,68,.1)',
                      color: resetResult.startsWith('✓') ? '#10b981' : '#ef4444',
                      fontSize: 13, fontWeight: 600,
                    }}>{resetResult}</div>
                  )}

                  {!resetConfirm ? (
                    <button onClick={() => setResetConfirm(true)} style={{
                      padding: '9px 18px', borderRadius: 9,
                      border: '1px solid rgba(239,68,68,.35)',
                      background: 'rgba(239,68,68,.12)', color: '#ef4444',
                      fontSize: 13, fontWeight: 700, cursor: 'pointer',
                    }}>Resetează Baza de Date</button>
                  ) : (
                    <div style={{ display: 'flex', gap: 10, alignItems: 'center', flexWrap: 'wrap' }}>
                      <span style={{ fontSize: 13, color: '#ef4444', fontWeight: 700 }}>
                        Ești sigur? Această acțiune nu poate fi anulată!
                      </span>
                      <button onClick={handleDevReset} disabled={resetLoading} style={{
                        padding: '9px 18px', borderRadius: 9, border: 'none',
                        background: '#ef4444', color: '#fff',
                        fontSize: 13, fontWeight: 700,
                        cursor: resetLoading ? 'not-allowed' : 'pointer',
                        opacity: resetLoading ? 0.7 : 1,
                      }}>{resetLoading ? 'Se resetează...' : 'Confirmă Resetarea'}</button>
                      <button onClick={() => setResetConfirm(false)} style={{
                        padding: '9px 18px', borderRadius: 9,
                        border: '1px solid rgba(255,255,255,.1)',
                        background: 'transparent', color: 'rgba(255,255,255,.4)',
                        fontSize: 13, fontWeight: 600, cursor: 'pointer',
                      }}>Anulează</button>
                    </div>
                  )}
                </div>
              )}

              {health.environment !== 'development' && (
                <div style={{
                  background: 'rgba(255,255,255,.03)', border: '1px solid rgba(255,255,255,.06)',
                  borderRadius: 10, padding: '12px 16px',
                }}>
                  <div style={{ fontSize: 13, color: 'rgba(255,255,255,.3)' }}>
                    🔒 Resetarea bazei de date este disponibilă doar în mediul <strong style={{ color: 'rgba(255,255,255,.5)' }}>development</strong>.
                    Mediu curent: <code style={{ color: '#ff9a3c' }}>{health.environment || 'production'}</code>
                  </div>
                </div>
              )}
            </>
          ) : (
            <div style={{ textAlign: 'center', color: 'rgba(255,255,255,.3)', padding: '60px 0' }}>
              Se încarcă datele de sistem...
            </div>
          )}
        </div>
      )}
    </AdminShell>
  )
}
