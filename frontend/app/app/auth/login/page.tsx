'use client'
import { useState, FormEvent } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'

export default function LoginPage() {
  const router = useRouter()
  const [email, setEmail]       = useState('')
  const [password, setPassword] = useState('')
  const [error, setError]       = useState('')
  const [loading, setLoading]   = useState(false)

  async function submit(e: FormEvent) {
    e.preventDefault(); setError(''); setLoading(true)
    try {
      const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL||'https://api.nuviax.app'}/api/v1/auth/login`, {
        method:'POST', headers:{'Content-Type':'application/json'},
        body: JSON.stringify({ email, password }),
      })
      if (!res.ok) { const j = await res.json().catch(()=>({})); throw new Error(j.error||'Date incorecte') }
      const { access_token, refresh_token } = await res.json()
      await fetch('/api/auth/set', { method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify({ access_token, refresh_token }) })
      window.location.href = '/dashboard'
    } catch(err: unknown) { setError(err instanceof Error ? err.message : 'Eroare') }
    finally { setLoading(false) }
  }

  return (
    <div className="auth-page">
      <div className="auth-card">
        <Link href="https://nuviaxapp.com" className="auth-logo">NUVia<span>X</span></Link>
        <h1 className="auth-title">Intră în cont</h1>
        <p className="auth-sub">Continuă unde ai rămas</p>
        {error && <div className="auth-err">{error}</div>}
        <form onSubmit={submit} className="auth-form">
          <div className="auth-field">
            <label className="auth-label">Email</label>
            <input type="email" className="auth-input" placeholder="tu@exemplu.com"
              value={email} onChange={e=>setEmail(e.target.value)} required autoComplete="email"/>
          </div>
          <div className="auth-field">
            <label className="auth-label">Parolă</label>
            <input type="password" className="auth-input" placeholder="••••••••"
              value={password} onChange={e=>setPassword(e.target.value)} required autoComplete="current-password"/>
          </div>
          <button type="submit" className="auth-btn" disabled={loading}>
            {loading ? <span className="spinner"/> : 'Intră în cont'}
          </button>
        </form>
        <p className="auth-footer">Nu ai cont? <Link href="/auth/register" className="auth-link">Creează unul gratuit</Link></p>
      </div>
    </div>
  )
}
