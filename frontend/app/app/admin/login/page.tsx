'use client'
import { useState, FormEvent } from 'react'

export default function AdminLoginPage() {
  const [email, setEmail]       = useState('')
  const [password, setPassword] = useState('')
  const [error, setError]       = useState('')
  const [loading, setLoading]   = useState(false)

  async function submit(e: FormEvent) {
    e.preventDefault(); setError(''); setLoading(true)
    try {
      const res = await fetch('/api/admin/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      })
      const data = await res.json().catch(() => ({}))
      if (!res.ok) throw new Error(data.error || 'Date incorecte')
      window.location.href = '/admin'
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Eroare')
      setLoading(false)
    }
  }

  return (
    <div style={{
      minHeight: '100vh',
      display: 'flex', alignItems: 'center', justifyContent: 'center',
      padding: 24,
      background: '#0a0a0f',
      backgroundImage: 'radial-gradient(ellipse 60% 60% at 50% 0%, rgba(99,102,241,.12) 0%, transparent 70%)',
      fontFamily: 'system-ui, -apple-system, sans-serif',
      color: '#e8e8f0',
    }}>
      <div style={{
        width: '100%', maxWidth: 400,
        background: 'rgba(255,255,255,.03)',
        border: '1px solid rgba(255,255,255,.08)',
        borderRadius: 18,
        padding: '36px 32px 32px',
      }}>
        {/* Logo + ADMIN badge */}
        <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 28 }}>
          <div style={{
            width: 34, height: 34, borderRadius: 8,
            background: 'linear-gradient(135deg, #ff6b35 0%, #ff9a3c 100%)',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            fontSize: 16, fontWeight: 900, color: '#fff',
          }}>N</div>
          <span style={{ fontWeight: 800, fontSize: 18, color: '#fff', letterSpacing: '-0.02em' }}>
            NuviaX
          </span>
          <span style={{
            fontSize: 10, fontWeight: 700, padding: '3px 8px', borderRadius: 5,
            background: 'rgba(99,102,241,.2)', color: '#818cf8',
            letterSpacing: '.06em', textTransform: 'uppercase',
          }}>ADMIN</span>
        </div>

        <h1 style={{
          fontSize: 26, fontWeight: 800, color: '#fff',
          letterSpacing: '-0.03em', marginBottom: 6,
        }}>Panel Administrare</h1>
        <p style={{ fontSize: 14, color: 'rgba(255,255,255,.45)', marginBottom: 28 }}>
          Acces restricționat pentru administratori NuviaX
        </p>

        {error && (
          <div style={{
            background: 'rgba(239,68,68,.1)',
            border: '1px solid rgba(239,68,68,.25)',
            borderRadius: 10, padding: '10px 14px',
            fontSize: 13, color: '#fca5a5', marginBottom: 16,
          }}>{error}</div>
        )}

        <form onSubmit={submit} style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
            <label style={{
              fontSize: 10, fontWeight: 600, letterSpacing: '.1em',
              textTransform: 'uppercase', color: 'rgba(255,255,255,.5)',
            }}>Email</label>
            <input
              type="email"
              placeholder="admin@nuviax.app"
              value={email}
              onChange={e => setEmail(e.target.value)}
              required
              autoComplete="email"
              style={{
                padding: '11px 14px', borderRadius: 11,
                border: '1.5px solid rgba(255,255,255,.1)',
                background: 'rgba(255,255,255,.04)',
                color: '#e8e8f0', fontSize: 14, outline: 'none',
              }}
            />
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
            <label style={{
              fontSize: 10, fontWeight: 600, letterSpacing: '.1em',
              textTransform: 'uppercase', color: 'rgba(255,255,255,.5)',
            }}>Parolă</label>
            <input
              type="password"
              placeholder="••••••••"
              value={password}
              onChange={e => setPassword(e.target.value)}
              required
              autoComplete="current-password"
              style={{
                padding: '11px 14px', borderRadius: 11,
                border: '1.5px solid rgba(255,255,255,.1)',
                background: 'rgba(255,255,255,.04)',
                color: '#e8e8f0', fontSize: 14, outline: 'none',
              }}
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            style={{
              marginTop: 8, padding: 16, borderRadius: 12, border: 'none',
              background: 'linear-gradient(135deg, #ff6b35, #ff9a3c)',
              color: '#fff', fontSize: 15, fontWeight: 700,
              cursor: loading ? 'default' : 'pointer',
              opacity: loading ? 0.6 : 1,
              boxShadow: '0 4px 20px rgba(255,107,53,.25)',
              letterSpacing: '-0.01em',
            }}
          >
            {loading ? 'Se verifică...' : 'Intră în panel'}
          </button>
        </form>

        <div style={{
          marginTop: 24, paddingTop: 16,
          borderTop: '1px solid rgba(255,255,255,.06)',
          fontSize: 12, color: 'rgba(255,255,255,.35)', textAlign: 'center',
        }}>
          🔒 Acces exclusiv administratori. Conturile normale vor fi respinse.
        </div>
      </div>
    </div>
  )
}
