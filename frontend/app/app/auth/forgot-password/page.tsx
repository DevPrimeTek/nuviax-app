'use client'
import { useState, FormEvent } from 'react'
import Link from 'next/link'

export default function ForgotPasswordPage() {
  const [email, setEmail]     = useState('')
  const [sent, setSent]       = useState(false)
  const [loading, setLoading] = useState(false)

  async function submit(e: FormEvent) {
    e.preventDefault()
    setLoading(true)
    await fetch('/api/auth/forgot-password', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email }),
    }).catch(() => {})
    setSent(true)
    setLoading(false)
  }

  return (
    <div className="auth-page">
      <div className="auth-card">
        <div className="auth-logo">NuviaX</div>
        <h1 className="auth-title">Resetare parolă</h1>

        {sent ? (
          <div style={{ textAlign: 'center', padding: '16px 0' }}>
            <div style={{ fontSize: 40, marginBottom: 12 }}>📧</div>
            <p className="auth-sub" style={{ marginBottom: 24 }}>
              Dacă adresa există în sistem, vei primi un email cu instrucțiuni de resetare.
              Verifică și dosarul Spam.
            </p>
            <Link href="/auth/login" className="btn-primary" style={{ display: 'inline-block' }}>
              Înapoi la autentificare
            </Link>
          </div>
        ) : (
          <form onSubmit={submit}>
            <p className="auth-sub">
              Introdu adresa de email asociată contului tău.
            </p>
            <label className="field-label">Email</label>
            <input
              type="email"
              className="field-input"
              value={email}
              onChange={e => setEmail(e.target.value)}
              placeholder="email@exemplu.com"
              required
              autoComplete="email"
            />
            <button type="submit" className="btn-primary" disabled={loading} style={{ width: '100%', marginTop: 16 }}>
              {loading ? 'Se trimite…' : 'Trimite link de resetare'}
            </button>
            <p className="auth-sub" style={{ marginTop: 16, textAlign: 'center' }}>
              <Link href="/auth/login" style={{ color: 'var(--l0)' }}>
                Înapoi la autentificare
              </Link>
            </p>
          </form>
        )}
      </div>
    </div>
  )
}
