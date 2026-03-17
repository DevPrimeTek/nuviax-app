'use client'
import { useEffect, useState } from 'react'
import Link from 'next/link'
import { usePathname, useRouter } from 'next/navigation'

type Lang = 'ro'|'en'|'ru'
const T: Record<Lang, Record<string,string>> = {
  ro: { home:'Acasă', today:'Azi', goals:'Obiective', recap:'Recap', settings:'Setări' },
  en: { home:'Home',  today:'Today', goals:'Goals', recap:'Recap', settings:'Settings' },
  ru: { home:'Главная', today:'Сегодня', goals:'Цели', recap:'Итоги', settings:'Настройки' },
}

const NAV = [
  { href:'/dashboard', key:'home',
    icon: <svg viewBox="0 0 24 24"><path d="M3 9l9-7 9 7v11a2 2 0 01-2 2H5a2 2 0 01-2-2z"/><polyline points="9,22 9,12 15,12 15,22"/></svg> },
  { href:'/today', key:'today', dot:true,
    icon: <svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><polyline points="12,6 12,12 16,14"/></svg> },
  { href:'/goals', key:'goals',
    icon: <svg viewBox="0 0 24 24"><line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/></svg> },
  { href:'/recap', key:'recap',
    icon: <svg viewBox="0 0 24 24"><polygon points="12,2 15.09,8.26 22,9.27 17,14.14 18.18,21.02 12,17.77 5.82,21.02 7,14.14 2,9.27 8.91,8.26 12,2"/></svg> },
  { href:'/settings', key:'settings',
    icon: <svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 010-4h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 012.83-2.83l.06.06A1.65 1.65 0 009 4.68a1.65 1.65 0 001-1.51V3a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 2.83l-.06.06A1.65 1.65 0 0019.4 9a1.65 1.65 0 001.51 1H21a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z"/></svg> },
]

export default function AppShell({ children, userName='Alexandru' }: { children: React.ReactNode; userName?: string }) {
  const pathname = usePathname()
  const router = useRouter()
  const [lang, setLang] = useState<Lang>('ro')
  const [theme, setTheme] = useState<'dark'|'light'>('dark')

  useEffect(() => {
    const savedLang = (localStorage.getItem('nv_lang') as Lang) || 'ro'
    const savedTheme = (localStorage.getItem('nv_theme') as 'dark'|'light') || 'dark'
    setLang(savedLang)
    setTheme(savedTheme)
    // Sincronizează cu backend pentru a asigura coerența
    fetch('/api/proxy/settings')
      .then(r => r.ok ? r.json() : null)
      .then(d => {
        if (d) {
          const serverLang = d.locale || d.Locale
          if (serverLang && serverLang !== savedLang) {
            setLang(serverLang as Lang)
            localStorage.setItem('nv_lang', serverLang)
          }
        }
      })
      .catch(() => {})
  }, [])

  function toggleTheme() {
    const t = theme === 'dark' ? 'light' : 'dark'
    setTheme(t); document.documentElement.dataset.theme = t
    localStorage.setItem('nv_theme', t)
  }
  async function logout() {
    await fetch('/api/auth/logout', { method:'POST' })
    window.location.href = '/auth/login'
  }

  const initials = userName.charAt(0).toUpperCase()
  const t = T[lang]

  return (
    <div className="shell">
      {/* Desktop top bar */}
      <div className="top-bar">
        <div className="top-logo">NUVia<span>X</span></div>
        <nav className="top-nav">
          {NAV.map(n => (
            <Link key={n.href} href={n.href}
              className={`top-btn${pathname.startsWith(n.href)?' on':''}`}>
              {n.icon}{t[n.key]}
              {n.dot && <span className="nav-dot"/>}
            </Link>
          ))}
        </nav>
        <div className="top-right">
          <button className="icon-btn" onClick={toggleTheme}>
            {theme==='dark'
              ? <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="5"/><line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/><line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/></svg>
              : <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M21 12.79A9 9 0 1111.21 3 7 7 0 0021 12.79z"/></svg>
            }
          </button>
          <Link href="/profile" className="user-chip">
            <div className="user-av">{initials}</div>
            <span className="user-name">{userName.split(' ')[0]}</span>
          </Link>
        </div>
      </div>

      {/* Scroll area */}
      <div className="app-scroll">
        <div className="app-col">{children}</div>
      </div>

      {/* Mobile bottom nav */}
      <nav className="mobile-nav">
        {NAV.map(n => (
          <Link key={n.href} href={n.href}
            className={`mob-tab${pathname.startsWith(n.href)?' on':''}`}>
            {n.icon}
            <span className="mob-lbl">{t[n.key]}</span>
            {n.dot && <span className="mob-dot"/>}
          </Link>
        ))}
      </nav>
    </div>
  )
}
