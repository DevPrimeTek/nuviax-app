'use client'
import { useState, FormEvent, useEffect, Suspense } from 'react'
import { useSearchParams } from 'next/navigation'
import Link from 'next/link'

function ResetPasswordForm() {
  const params           = useSearchParams()
  const token            = params.get('token') ?? ''
  const [pw, setPw]      = useState('')
  const [pw2, setPw2]    = useState('')
  const [error, setError]   = useState('')
  const [done, setDone]     = useState(false)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (!token) setError('Link invalid sau expirat. Solicită un link nou.')
  }, [token])

  async function submit(e: FormEvent) {
    e.preventDefault()
    setError('')
    if (pw.length < 8) { setError('Parola trebuie să aibă cel puțin 8 caractere.'); return }
    if (pw !== pw2)    { setError('Parolele nu coincid.'); return }
    setLoading(true)
    try {
      const res = await fetch('/api/auth/reset-password', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ token, new_password: pw }),
      })
      const data = await res.json().catch(() => ({}))
      if (!res.ok) throw new Error(data.error || 'Eroare la resetare.')
      setDone(true)
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Eroare')
    }
    setLoading(false)
  }

  return (
    <div className="auth-page">
      <div className="auth-card">
        <div className="auth-logo">NuviaX</div>
        <h1 className="auth-title">Parolă nouă</h1>

        {done ? (
          <div style={{ textAlign: 'center', padding: '16px 0' }}>
            <div style={{ fontSize: 40, marginBottom: 12 }}>✅</div>
            <p className="auth-sub" style={{ marginBottom: 24 }}>
              Parola a fost resetată cu succes. Te poți autentifica.
            </p>
            <Link href="/auth/login" className="btn-primary" style={{ display: 'inline-block' }}>
              Autentifică-te
            </Link>
          </div>
        ) : (
          <form onSubmit={submit}>
            {error && <div className="auth-error">{error}</div>}
            <label className="field-label">Parolă nouă</label>
            <input
              type="password"
              className="field-input"
              value={pw}
              onChange={e => setPw(e.target.value)}
              placeholder="Min. 8 caractere"
              required
              autoComplete="new-password"
              disabled={!token}
            />
            <label className="field-label" style={{ marginTop: 12 }}>Confirmă parola</label>
            <input
              type="password"
              className="field-input"
              value={pw2}
              onChange={e => setPw2(e.target.value)}
              placeholder="Repetă parola"
              required
              autoComplete="new-password"
              disabled={!token}
            />
            <button
              type="submit"
              className="btn-primary"
              disabled={loading || !token}
              style={{ width: '100%', marginTop: 16 }}
            >
              {loading ? 'Se salvează…' : 'Setează parola nouă'}
            </button>
            <p className="auth-sub" style={{ marginTop: 16, textAlign: 'center' }}>
              <Link href="/auth/forgot-password" style={{ color: 'var(--l0)' }}>
                Solicită un link nou
              </Link>
            </p>
          </form>
        )}
      </div>
    </div>
  )
}

export default function ResetPasswordPage() {
  return (
    <Suspense>
      <ResetPasswordForm />
    </Suspense>
  )
}
