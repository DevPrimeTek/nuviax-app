'use client'
import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import AppShell from '@/components/layout/AppShell'

type Lang = 'ro'|'en'|'ru'

export default function SettingsPage() {
  const router = useRouter()
  const [name, setName]   = useState('')
  const [email, setEmail] = useState('')
  const [lang, setLangS]  = useState<Lang>('ro')
  const [theme, setThemeS]= useState<'dark'|'light'>('dark')
  const [notif, setNotif] = useState(true)
  const [review, setReview] = useState(true)
  useEffect(() => {
    setLangS((localStorage.getItem('nv_lang') as Lang)||'ro')
    setThemeS((localStorage.getItem('nv_theme') as 'dark'|'light')||'dark')
    setName(localStorage.getItem('nv_profile_name') || '')
    fetch('/api/proxy/settings')
      .then(r => { if(!r.ok) throw new Error(r.status.toString()); return r.json() })
      .then(d => { const l=d.locale||d.Locale; if(l) setLangS(l) })
      .catch((err) => { if(err.message==='401') router.push('/auth/login') })
  }, [])

  function setLang(l: Lang) {
    setLangS(l); localStorage.setItem('nv_lang', l)
    fetch('/api/proxy/settings', { method:'PATCH',
      headers:{'Content-Type':'application/json'}, body:JSON.stringify({locale:l}) }).catch(()=>{})
  }
  function setTheme(t: 'dark'|'light') {
    setThemeS(t); document.documentElement.dataset.theme = t
    localStorage.setItem('nv_theme', t)
  }
  async function logout() {
    await fetch('/api/auth/logout', { method:'POST' })
    router.push('/auth/login')
  }

  const themeSubLabel = theme==='dark' ? 'Întunecat activ' : 'Deschis activ'
  const langLabel = {ro:'Română',en:'English',ru:'Русский'}[lang]

  return (
    <AppShell userName={name||'A'}>
      <div className="page">
        <div className="greet-title" style={{marginBottom:20}}>Setări</div>

        {/* Profile */}
        <div className="profile-card">
          <div className="profile-av">{name.charAt(0)||'A'}</div>
          <div style={{flex:1}}>
            <div className="profile-name">{name}</div>
            <div className="profile-email">{email}</div>
          </div>
          <div className="sg-arr"><svg viewBox="0 0 24 24"><polyline points="9,18 15,12 9,6"/></svg></div>
        </div>

        {/* Sistem */}
        <div className="sg-group">
          <div className="sg-lbl">Sistem</div>
          <div className="sg-items">
            {/* Notificări */}
            <div className="sg-item">
              <div className="sg-icon-wrap"><svg viewBox="0 0 24 24"><path d="M18 8A6 6 0 006 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 01-3.46 0"/></svg></div>
              <div style={{flex:1}}><div className="sg-name">Notificări</div><div className="sg-sub">Activitățile zilnice la 8:00</div></div>
              <div className="sg-right">
                <div className={`toggle${notif?' on':''}`} onClick={()=>setNotif(!notif)}/>
              </div>
            </div>
            {/* Temă */}
            <div className="sg-item">
              <div className="sg-icon-wrap"><svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="5"/><line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/><line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/></svg></div>
              <div style={{flex:1}}><div className="sg-name">Temă</div><div className="sg-sub">{themeSubLabel}</div></div>
              <div className="sg-right" style={{gap:5}}>
                <button className={`sg-lang-btn${theme==='dark'?' on':''}`} onClick={()=>setTheme('dark')}>Dark</button>
                <button className={`sg-lang-btn${theme==='light'?' on':''}`} onClick={()=>setTheme('light')}>Light</button>
              </div>
            </div>
            {/* Limbă */}
            <div className="sg-item">
              <div className="sg-icon-wrap"><svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10 15.3 15.3 0 014-10z"/></svg></div>
              <div style={{flex:1}}><div className="sg-name">Limbă</div><div className="sg-sub">{langLabel}</div></div>
              <div className="sg-right" style={{gap:5}}>
                <button className={`sg-lang-btn${lang==='ro'?' on':''}`} onClick={()=>setLang('ro')}>RO</button>
                <button className={`sg-lang-btn${lang==='en'?' on':''}`} onClick={()=>setLang('en')}>EN</button>
                <button className={`sg-lang-btn${lang==='ru'?' on':''}`} onClick={()=>setLang('ru')}>RU</button>
              </div>
            </div>
          </div>
        </div>

        {/* Securitate */}
        <div className="sg-group">
          <div className="sg-lbl">Cont și securitate</div>
          <div className="sg-items">
            <div className="sg-item">
              <div className="sg-icon-wrap"><svg viewBox="0 0 24 24"><rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0110 0v4"/></svg></div>
              <div style={{flex:1}}><div className="sg-name">Verificare în doi pași</div><div className="sg-sub">Autentificare cu telefon</div></div>
              <div className="sg-right"><span className="sg-badge">Activ</span></div>
            </div>
            <div className="sg-item">
              <div className="sg-icon-wrap"><svg viewBox="0 0 24 24"><path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 11-7.778 7.778 5.5 5.5 0 017.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4"/></svg></div>
              <div style={{flex:1}}><div className="sg-name">Schimbă parola</div></div>
              <div className="sg-right"><div className="sg-arr"><svg viewBox="0 0 24 24"><polyline points="9,18 15,12 9,6"/></svg></div></div>
            </div>
          </div>
        </div>

        {/* Preferințe */}
        <div className="sg-group">
          <div className="sg-lbl">Preferințe</div>
          <div className="sg-items">
            <div className="sg-item">
              <div className="sg-icon-wrap"><svg viewBox="0 0 24 24"><polyline points="23,4 23,10 17,10"/><polyline points="1,20 1,14 7,14"/><path d="M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15"/></svg></div>
              <div style={{flex:1}}><div className="sg-name">Recapitulare de etapă</div><div className="sg-sub">Câteva întrebări la finalul etapei</div></div>
              <div className="sg-right"><div className={`toggle${review?' on':''}`} onClick={()=>setReview(!review)}/></div>
            </div>
            <div className="sg-item">
              <div className="sg-icon-wrap"><svg viewBox="0 0 24 24"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7,10 12,15 17,10"/><line x1="12" y1="15" x2="12" y2="3"/></svg></div>
              <div style={{flex:1}}><div className="sg-name">Descarcă datele tale</div></div>
              <div className="sg-right"><div className="sg-arr"><svg viewBox="0 0 24 24"><polyline points="9,18 15,12 9,6"/></svg></div></div>
            </div>
          </div>
        </div>

        {/* Logout */}
        <div className="sg-group">
          <div className="sg-items">
            <div className="sg-item" onClick={logout} style={{cursor:'pointer'}}>
              <div className="sg-icon-wrap" style={{background:'rgba(219,39,119,.1)'}}>
                <svg viewBox="0 0 24 24" style={{stroke:'var(--l4l)'}}><path d="M9 21H5a2 2 0 01-2-2V5a2 2 0 012-2h4"/><polyline points="16,17 21,12 16,7"/><line x1="21" y1="12" x2="9" y2="12"/></svg>
              </div>
              <div style={{flex:1}}><div className="sg-name" style={{color:'var(--l4l)'}}>Deconectează-te</div></div>
            </div>
          </div>
        </div>

        <div style={{height:32}}/>
      </div>
    </AppShell>
  )
}
