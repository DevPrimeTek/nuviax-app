'use client'
import { useState, FormEvent } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'

export default function RegisterPage() {
  const router = useRouter()
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
      const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL||'https://api.nuviax.app'}/api/v1/auth/register`, {
        method:'POST', headers:{'Content-Type':'application/json'},
        body: JSON.stringify({ name, email, password }),
      })
      if (!res.ok) { const j = await res.json().catch(()=>({})); throw new Error(j.error||'Înregistrare eșuată') }
      const { access_token, refresh_token } = await res.json()
      await fetch('/api/auth/set', { method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify({ access_token, refresh_token }) })
      router.push('/dashboard')
    } catch(err: unknown) { setError(err instanceof Error ? err.message : 'Eroare') }
    finally { setLoading(false) }
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
