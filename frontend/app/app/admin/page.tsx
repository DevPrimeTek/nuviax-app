'use client'
import { useEffect, useState, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import AppShell from '@/components/layout/AppShell'

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

// ── Components ────────────────────────────────────────────────────────────────

function StatCard({ label, value, sub, color }: { label: string; value: string | number; sub?: string; color?: string }) {
  return (
    <div style={{
      background: 'var(--bg2)', border: '1.5px solid var(--line)', borderRadius: 14,
      padding: '16px 18px', display: 'flex', flexDirection: 'column', gap: 4,
    }}>
      <div style={{ fontSize: 12, color: 'var(--ink3)', fontWeight: 500 }}>{label}</div>
      <div style={{ fontSize: 26, fontWeight: 800, fontFamily: 'var(--ff-d)', color: color || 'var(--ink)', letterSpacing: '-0.03em' }}>
        {typeof value === 'number' ? fmt(value) : value}
      </div>
      {sub && <div style={{ fontSize: 11.5, color: 'var(--ink4)' }}>{sub}</div>}
    </div>
  )
}

function TabBtn({ id, current, label, onClick }: { id: Tab; current: Tab; label: string; onClick: (t: Tab) => void }) {
  const on = id === current
  return (
    <button onClick={() => onClick(id)} style={{
      padding: '7px 16px', borderRadius: 9, border: 'none', cursor: 'pointer',
      fontSize: 13, fontWeight: 600,
      background: on ? 'var(--l0g)' : 'transparent',
      color: on ? 'var(--l0l)' : 'var(--ink3)',
      transition: 'all .18s',
    }}>{label}</button>
  )
}

function Badge({ label, type }: { label: string; type: 'success' | 'warn' | 'error' | 'info' | 'neutral' }) {
  const colors: Record<string, [string, string]> = {
    success: ['rgba(16,185,129,.12)', 'var(--l2l)'],
    warn:    ['rgba(245,158,11,.12)', '#f59e0b'],
    error:   ['rgba(239,68,68,.12)',  '#ef4444'],
    info:    ['rgba(99,102,241,.12)', '#818cf8'],
    neutral: ['var(--bg3)',           'var(--ink3)'],
  }
  const [bg, fg] = colors[type]
  return (
    <span style={{
      display: 'inline-block', padding: '2px 8px', borderRadius: 6,
      fontSize: 11, fontWeight: 700, background: bg, color: fg,
    }}>{label}</span>
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

  useEffect(() => {
    // Check if user is admin by loading stats first
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
      setResetResult(`✓ Reset complet: ${d.deleted_users} utilizatori, ${d.deleted_goals} obiective, ${d.deleted_tasks} activități șterse.`)
      setResetConfirm(false)
      loadStats(); loadUsers(); loadAudit()
    } catch (e: unknown) {
      const err = e instanceof Error ? e.message : 'Eroare necunoscută'
      setResetResult(`✗ Eroare: ${err}`)
    } finally {
      setResetLoading(false)
    }
  }

  if (error) {
    return (
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100vh', background: 'var(--bg)' }}>
        <div style={{ textAlign: 'center', color: 'var(--ink3)' }}>
          <div style={{ fontSize: 48, marginBottom: 12 }}>🔒</div>
          <div style={{ fontSize: 18, fontWeight: 700, color: 'var(--ink)', marginBottom: 8 }}>{error}</div>
          <button onClick={() => router.push('/dashboard')} style={{
            padding: '10px 20px', borderRadius: 10, border: 'none', cursor: 'pointer',
            background: 'var(--l0g)', color: 'var(--l0l)', fontWeight: 600, fontSize: 14,
          }}>← Înapoi la dashboard</button>
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
    <AppShell userName={userName}>
      <div className="page" style={{ padding: '0 0 40px' }}>

        {/* Header */}
        <div style={{ marginBottom: 24 }}>
          <div style={{
            display: 'flex', alignItems: 'center', justifyContent: 'space-between',
            flexWrap: 'wrap', gap: 12, marginBottom: 8,
          }}>
            <div>
              <div style={{
                fontFamily: 'var(--ff-d)', fontSize: 22, fontWeight: 800,
                color: 'var(--ink)', letterSpacing: '-0.03em',
              }}>
                Panel Administrare
              </div>
              <div style={{ fontSize: 13, color: 'var(--ink3)', marginTop: 2 }}>
                {stats ? `Ultima actualizare: ${fmtDateTime(stats.computed_at)}` : 'Se încarcă...'}
              </div>
            </div>
            <button onClick={() => { loadStats(); loadUsers(); loadAudit(); loadHealth() }} style={{
              padding: '8px 16px', borderRadius: 10, border: '1.5px solid var(--line)',
              background: 'var(--bg2)', color: 'var(--ink2)', fontSize: 13, fontWeight: 600,
              cursor: 'pointer',
            }}>↺ Reîmprospătare</button>
          </div>

          {/* Tabs */}
          <div style={{ display: 'flex', gap: 4, flexWrap: 'wrap' }}>
            <TabBtn id="stats"  current={tab} label="📊 Statistici"   onClick={setTab} />
            <TabBtn id="users"  current={tab} label="👥 Utilizatori"  onClick={setTab} />
            <TabBtn id="audit"  current={tab} label="📋 Jurnal Audit" onClick={setTab} />
            <TabBtn id="health" current={tab} label="🔧 Sistem"       onClick={setTab} />
          </div>
        </div>

        {/* ── STATS TAB ── */}
        {tab === 'stats' && (
          <div>
            {loading ? (
              <div style={{ textAlign: 'center', color: 'var(--ink3)', padding: '40px 0' }}>Se încarcă statisticile...</div>
            ) : stats ? (
              <>
                <div style={{ fontSize: 11, fontWeight: 700, color: 'var(--ink4)', textTransform: 'uppercase', letterSpacing: '.06em', marginBottom: 12 }}>
                  Utilizatori
                </div>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(150px, 1fr))', gap: 10, marginBottom: 24 }}>
                  <StatCard label="Total utilizatori" value={stats.total_users} />
                  <StatCard label="Noi (7 zile)"     value={stats.new_users_7d}  color="var(--l0l)" />
                  <StatCard label="Noi (30 zile)"    value={stats.new_users_30d} color="var(--l0l)" />
                  <StatCard label="Admini"           value={stats.admin_users}   color="var(--l5l)" />
                </div>

                <div style={{ fontSize: 11, fontWeight: 700, color: 'var(--ink4)', textTransform: 'uppercase', letterSpacing: '.06em', marginBottom: 12 }}>
                  Obiective & Sprinturi
                </div>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(150px, 1fr))', gap: 10, marginBottom: 24 }}>
                  <StatCard label="Obiective active"    value={stats.active_goals}    color="var(--l2l)" />
                  <StatCard label="Obiective finalizate" value={stats.completed_goals} color="var(--l2l)" />
                  <StatCard label="Obiective în pauză"  value={stats.paused_goals} />
                  <StatCard label="Total obiective"     value={stats.total_goals} />
                  <StatCard label="Sprinturi active"    value={stats.active_sprints}    color="var(--l0l)" />
                  <StatCard label="Sprinturi finalizate" value={stats.completed_sprints} color="var(--l0l)" />
                </div>

                <div style={{ fontSize: 11, fontWeight: 700, color: 'var(--ink4)', textTransform: 'uppercase', letterSpacing: '.06em', marginBottom: 12 }}>
                  Activitate Zilnică
                </div>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(150px, 1fr))', gap: 10, marginBottom: 24 }}>
                  <StatCard label="Activități azi"     value={stats.tasks_today} />
                  <StatCard label="Completate azi"     value={stats.tasks_completed_today} color="var(--l2l)" />
                  <StatCard
                    label="Rata completare azi"
                    value={stats.tasks_today > 0 ? `${Math.round((stats.tasks_completed_today / stats.tasks_today) * 100)}%` : '—'}
                    color={stats.tasks_today > 0 && stats.tasks_completed_today / stats.tasks_today >= 0.7 ? 'var(--l2l)' : undefined}
                  />
                </div>

                <div style={{ fontSize: 11, fontWeight: 700, color: 'var(--ink4)', textTransform: 'uppercase', letterSpacing: '.06em', marginBottom: 12 }}>
                  Sănătatea Sistemului (ultimele 30 zile)
                </div>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(150px, 1fr))', gap: 10 }}>
                  <StatCard label="Evenimente SRM"   value={stats.srm_events_30d} />
                  <StatCard label="SRM Level 3"      value={stats.srm_l3_events_30d}      color={stats.srm_l3_events_30d > 0 ? '#f59e0b' : undefined} />
                  <StatCard label="Regresii detect."  value={stats.regression_events_30d}  color={stats.regression_events_30d > 0 ? '#ef4444' : undefined} />
                  <StatCard label="Ceremonii generate" value={stats.ceremonies_30d}        color="var(--l5l)" />
                  <StatCard label="Badge-uri acordate" value={stats.badges_awarded_30d}    color="var(--l5l)" />
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
                  flex: 1, minWidth: 200, padding: '9px 14px', borderRadius: 10,
                  border: '1.5px solid var(--line)', background: 'var(--bg2)',
                  color: 'var(--ink)', fontSize: 13, outline: 'none',
                }}
              />
              <span style={{ fontSize: 13, color: 'var(--ink3)' }}>{filteredUsers.length} utilizatori</span>
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
              {filteredUsers.map(u => (
                <div key={u.id} style={{
                  background: 'var(--bg2)', border: '1.5px solid var(--line)', borderRadius: 14,
                  padding: '14px 16px',
                  opacity: u.is_active ? 1 : 0.6,
                }}>
                  <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 12, flexWrap: 'wrap' }}>
                    <div style={{ flex: 1 }}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4, flexWrap: 'wrap' }}>
                        <span style={{ fontWeight: 700, color: 'var(--ink)', fontSize: 14 }}>
                          {u.full_name || '(fără nume)'}
                        </span>
                        {u.is_admin && <Badge label="ADMIN" type="info" />}
                        {!u.is_active && <Badge label="INACTIV" type="error" />}
                        {u.mfa_enabled && <Badge label="MFA" type="success" />}
                        <Badge label={u.locale.toUpperCase()} type="neutral" />
                      </div>
                      <div style={{ fontSize: 11, color: 'var(--ink4)', fontFamily: 'monospace', marginBottom: 8 }}>
                        {u.id}
                      </div>
                      <div style={{ display: 'flex', gap: 16, flexWrap: 'wrap' }}>
                        <span style={{ fontSize: 12, color: 'var(--ink3)' }}>
                          🎯 {u.active_goals} active / {u.total_goals} total
                        </span>
                        <span style={{ fontSize: 12, color: 'var(--ink3)' }}>
                          ⚡ {u.tasks_last_30d} activități/30z
                        </span>
                        <span style={{ fontSize: 12, color: 'var(--ink3)' }}>
                          🏃 {u.completed_sprints} sprinturi
                        </span>
                        <span style={{ fontSize: 12, color: 'var(--ink3)' }}>
                          🔗 {u.active_sessions} sesiuni
                        </span>
                        <span style={{ fontSize: 12, color: 'var(--ink3)' }}>
                          📅 Creat {fmtDate(u.created_at)}
                        </span>
                        <span style={{ fontSize: 12, color: 'var(--ink3)' }}>
                          ⏱ Activ {fmtDateTime(u.last_active_at)}
                        </span>
                      </div>
                    </div>

                    {!u.is_admin && (
                      <div style={{ display: 'flex', gap: 6, flexShrink: 0 }}>
                        {u.is_active ? (
                          <button onClick={() => handleUserAction(u.id, 'deactivate')} style={{
                            padding: '6px 12px', borderRadius: 8, border: '1.5px solid rgba(239,68,68,.3)',
                            background: 'rgba(239,68,68,.08)', color: '#ef4444',
                            fontSize: 12, fontWeight: 600, cursor: 'pointer',
                          }}>Dezactivează</button>
                        ) : (
                          <button onClick={() => handleUserAction(u.id, 'activate')} style={{
                            padding: '6px 12px', borderRadius: 8, border: '1.5px solid rgba(16,185,129,.3)',
                            background: 'rgba(16,185,129,.08)', color: 'var(--l2l)',
                            fontSize: 12, fontWeight: 600, cursor: 'pointer',
                          }}>Activează</button>
                        )}
                        <button onClick={() => { if (confirm('Promovezi acest utilizator la admin? Această acțiune nu poate fi anulată ușor.')) handleUserAction(u.id, 'promote') }} style={{
                          padding: '6px 12px', borderRadius: 8, border: '1.5px solid var(--line)',
                          background: 'var(--bg3)', color: 'var(--ink3)',
                          fontSize: 12, fontWeight: 600, cursor: 'pointer',
                        }}>Promovează Admin</button>
                      </div>
                    )}
                  </div>

                  {userMsg[u.id] && (
                    <div style={{ marginTop: 8, fontSize: 12, color: userMsg[u.id].startsWith('✓') ? 'var(--l2l)' : '#ef4444', fontWeight: 600 }}>
                      {userMsg[u.id]}
                    </div>
                  )}
                </div>
              ))}

              {filteredUsers.length === 0 && (
                <div style={{ textAlign: 'center', color: 'var(--ink3)', padding: '32px 0' }}>
                  {searchQ ? 'Niciun utilizator găsit.' : 'Nu există utilizatori înregistrați.'}
                </div>
              )}
            </div>
          </div>
        )}

        {/* ── AUDIT TAB ── */}
        {tab === 'audit' && (
          <div>
            <div style={{ marginBottom: 14, fontSize: 13, color: 'var(--ink3)' }}>
              Ultimele {audit.length} intrări din jurnalul de audit
            </div>
            <div style={{
              background: 'var(--bg2)', border: '1.5px solid var(--line)', borderRadius: 14,
              overflow: 'hidden',
            }}>
              {audit.length === 0 ? (
                <div style={{ padding: '24px', textAlign: 'center', color: 'var(--ink3)' }}>Jurnalul este gol.</div>
              ) : (
                <div style={{ overflowX: 'auto' }}>
                  <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 13 }}>
                    <thead>
                      <tr style={{ borderBottom: '1.5px solid var(--line)' }}>
                        {['Data', 'Utilizator', 'Acțiune'].map(h => (
                          <th key={h} style={{
                            padding: '10px 14px', textAlign: 'left',
                            fontSize: 11, fontWeight: 700, color: 'var(--ink4)',
                            textTransform: 'uppercase', letterSpacing: '.04em',
                          }}>{h}</th>
                        ))}
                      </tr>
                    </thead>
                    <tbody>
                      {audit.map((e, i) => (
                        <tr key={e.id} style={{
                          borderBottom: i < audit.length - 1 ? '1px solid var(--line)' : 'none',
                        }}>
                          <td style={{ padding: '9px 14px', color: 'var(--ink3)', whiteSpace: 'nowrap' }}>
                            {fmtDateTime(e.created_at)}
                          </td>
                          <td style={{ padding: '9px 14px', color: 'var(--ink2)' }}>
                            {e.full_name || (e.user_id ? e.user_id.substring(0, 8) + '…' : '—')}
                          </td>
                          <td style={{ padding: '9px 14px' }}>
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
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(150px, 1fr))', gap: 10, marginBottom: 24 }}>
                  <StatCard label="Stare sistem"     value={health.status === 'ok' ? '✓ OK' : '⚠ Degradat'} color={health.status === 'ok' ? 'var(--l2l)' : '#f59e0b'} />
                  <StatCard label="Mediu"            value={health.environment || 'production'} />
                  <StatCard label="Tabele DB"        value={health.database.table_count} />
                  <StatCard label="Joburi scheduler (24h)" value={health.scheduler.jobs_last_24h} />
                  <StatCard label="Conexiuni DB total"   value={health.database.pool_total_conns} />
                  <StatCard label="Conexiuni idle"       value={health.database.pool_idle_conns} />
                  <StatCard label="Conexiuni active"     value={health.database.pool_acquired_conns} />
                  <StatCard label="Conexiuni max pool"   value={health.database.pool_max_conns} />
                </div>

                <div style={{
                  background: 'var(--bg2)', border: '1.5px solid var(--line)', borderRadius: 14,
                  padding: '14px 16px', marginBottom: 24,
                }}>
                  <div style={{ fontSize: 11, fontWeight: 700, color: 'var(--ink4)', textTransform: 'uppercase', letterSpacing: '.06em', marginBottom: 8 }}>
                    Versiune PostgreSQL
                  </div>
                  <div style={{ fontSize: 12, color: 'var(--ink3)', fontFamily: 'monospace', wordBreak: 'break-all' }}>
                    {health.database.version}
                  </div>
                </div>

                {/* DEV RESET SECTION */}
                {health.environment === 'development' && (
                  <div style={{
                    background: 'rgba(239,68,68,.06)', border: '1.5px solid rgba(239,68,68,.25)',
                    borderRadius: 14, padding: '16px 18px',
                  }}>
                    <div style={{ fontWeight: 700, color: '#ef4444', marginBottom: 6, fontSize: 14 }}>
                      ⚠ Resetare Bază de Date (Development Only)
                    </div>
                    <div style={{ fontSize: 13, color: 'var(--ink3)', marginBottom: 14 }}>
                      Șterge TOȚI utilizatorii non-admin și datele lor (obiective, sprinturi, activități).
                      Schema și conturile admin sunt păstrate. Disponibil doar în mediu de development.
                    </div>

                    {resetResult && (
                      <div style={{
                        padding: '10px 14px', borderRadius: 9, marginBottom: 12,
                        background: resetResult.startsWith('✓') ? 'rgba(16,185,129,.1)' : 'rgba(239,68,68,.1)',
                        color: resetResult.startsWith('✓') ? 'var(--l2l)' : '#ef4444',
                        fontSize: 13, fontWeight: 600,
                      }}>{resetResult}</div>
                    )}

                    {!resetConfirm ? (
                      <button onClick={() => setResetConfirm(true)} style={{
                        padding: '9px 18px', borderRadius: 10, border: '1.5px solid rgba(239,68,68,.4)',
                        background: 'rgba(239,68,68,.12)', color: '#ef4444',
                        fontSize: 13, fontWeight: 700, cursor: 'pointer',
                      }}>Resetează Baza de Date</button>
                    ) : (
                      <div style={{ display: 'flex', gap: 10, alignItems: 'center', flexWrap: 'wrap' }}>
                        <div style={{ fontSize: 13, color: '#ef4444', fontWeight: 700 }}>
                          Ești sigur? Această acțiune nu poate fi anulată!
                        </div>
                        <button
                          onClick={handleDevReset}
                          disabled={resetLoading}
                          style={{
                            padding: '9px 18px', borderRadius: 10, border: 'none',
                            background: '#ef4444', color: 'white',
                            fontSize: 13, fontWeight: 700, cursor: resetLoading ? 'not-allowed' : 'pointer',
                            opacity: resetLoading ? 0.7 : 1,
                          }}
                        >{resetLoading ? 'Se resetează...' : 'Confirmă Resetarea'}</button>
                        <button onClick={() => setResetConfirm(false)} style={{
                          padding: '9px 18px', borderRadius: 10, border: '1.5px solid var(--line)',
                          background: 'transparent', color: 'var(--ink3)',
                          fontSize: 13, fontWeight: 600, cursor: 'pointer',
                        }}>Anulează</button>
                      </div>
                    )}
                  </div>
                )}

                {health.environment !== 'development' && (
                  <div style={{
                    background: 'var(--bg2)', border: '1.5px solid var(--line)', borderRadius: 14,
                    padding: '14px 16px', opacity: 0.6,
                  }}>
                    <div style={{ fontSize: 13, color: 'var(--ink3)' }}>
                      🔒 Resetarea bazei de date este disponibilă doar în mediul de <strong>development</strong>.
                      Mediu curent: <code>{health.environment || 'production'}</code>
                    </div>
                  </div>
                )}
              </>
            ) : (
              <div style={{ textAlign: 'center', color: 'var(--ink3)', padding: '40px 0' }}>Se încarcă datele de sistem...</div>
            )}
          </div>
        )}
      </div>
    </AppShell>
  )
}
