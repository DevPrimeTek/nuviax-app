'use client'
import { useState, FormEvent } from 'react'
import Link from 'next/link'

export default function RegisterPage() {
  const [name, setName]         = useState('')
  const [email, setEmail]       = useState('')
  const [password, setPassword] = useState('')
  const [error, setError]       = useState('')
  const [loading, setLoading]   = useState(false)

  async function submit(e: FormEvent) {
    e.preventDefault(); setError('')
    if (password.length < 8) { setError('Parola trebuie să aibă minim 8 caractere'); return }
    setLoading(true)
    try {
      const res = await fetch('/api/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, email, password }),
      })
      const data = await res.json().catch(() => ({}))
      if (!res.ok) throw new Error(data.error || 'Înregistrare eșuată')
      window.location.href = '/onboarding'
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Eroare')
      setLoading(false)
    }
  }

  return (
    <div className="auth-page">
      <div className="auth-card">
        <Link href="https://nuviaxapp.com" className="auth-logo">NUVia<span>X</span></Link>
        <h1 className="auth-title">Creează cont</h1>
        <p className="auth-sub">Gratuit 14 zile · Fără card bancar</p>
        {error && <div className="auth-err">{error}</div>}
        <form onSubmit={submit} className="auth-form">
          <div className="auth-field">
            <label className="auth-label">Numele tău</label>
            <input type="text" className="auth-input" placeholder="Alexandru"
              value={name} onChange={e=>setName(e.target.value)} required autoComplete="name"/>
          </div>
          <div className="auth-field">
            <label className="auth-label">Email</label>
            <input type="email" className="auth-input" placeholder="tu@exemplu.com"
              value={email} onChange={e=>setEmail(e.target.value)} required autoComplete="email"/>
          </div>
          <div className="auth-field">
            <label className="auth-label">Parolă</label>
            <input type="password" className="auth-input" placeholder="minim 8 caractere"
              value={password} onChange={e=>setPassword(e.target.value)} required autoComplete="new-password"/>
          </div>
          <button type="submit" className="auth-btn" disabled={loading}>
            {loading ? <span className="spinner"/> : 'Creează contul gratuit'}
          </button>
        </form>
        <p className="auth-footer">Ai deja cont? <Link href="/auth/login" className="auth-link">Intră</Link></p>
      </div>
    </div>
  )
}
