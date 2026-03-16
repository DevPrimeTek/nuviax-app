'use client'
import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import AppShell from '@/components/layout/AppShell'

type Lang = 'ro'|'en'|'ru'

export default function ProfilePage() {
  const router = useRouter()
  const [name, setName]       = useState('')
  const [domain, setDomain]   = useState('')
  const [lang, setLang]       = useState<Lang>('ro')
  const [saving, setSaving]   = useState(false)
  const [saved, setSaved]     = useState(false)

  useEffect(() => {
    const storedName   = localStorage.getItem('nv_profile_name')   || ''
    const storedDomain = localStorage.getItem('nv_profile_domain') || ''
    const storedLang   = (localStorage.getItem('nv_lang') as Lang) || 'ro'
    setName(storedName)
    setDomain(storedDomain)
    setLang(storedLang)

    // Fetch locale din backend
    fetch('/api/proxy/settings')
      .then(r => r.ok ? r.json() : null)
      .then(d => { if (d?.locale) setLang(d.locale as Lang) })
      .catch(() => {})
  }, [])

  async function save() {
    setSaving(true)
    localStorage.setItem('nv_profile_name',   name)
    localStorage.setItem('nv_profile_domain', domain)
    localStorage.setItem('nv_lang', lang)

    await fetch('/api/proxy/settings', {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ locale: lang }),
    }).catch(() => {})

    setSaving(false)
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  const initials = name ? name.split(' ').map(w=>w[0]).join('').slice(0,2).toUpperCase() : 'U'

  return (
    <AppShell userName={name || 'Profil'}>
      <div className="page">
        <div className="greet-title" style={{marginBottom:4}}>Profilul meu</div>
        <div style={{fontSize:14,color:'var(--ink3)',marginBottom:24}}>Datele tale de bază</div>

        {/* Avatar */}
        <div style={{display:'flex',justifyContent:'center',marginBottom:28}}>
          <div style={{
            width:72, height:72, borderRadius:'50%',
            background:'linear-gradient(135deg, var(--l0), var(--l2))',
            display:'flex', alignItems:'center', justifyContent:'center',
            fontSize:26, fontWeight:800, color:'white', fontFamily:'var(--ff-h)',
          }}>
            {initials}
          </div>
        </div>

        {/* Câmpuri */}
        <div className="sg-group">
          <div className="sg-lbl">Date personale</div>
          <div className="sg-items">
            <div className="sg-item" style={{flexDirection:'column',alignItems:'stretch',gap:6}}>
              <label style={{fontSize:12,color:'var(--ink4)',fontFamily:'var(--ff-m)'}}>Nume complet</label>
              <input
                type="text"
                value={name}
                onChange={e => setName(e.target.value)}
                placeholder="ex. Alexandru Ionescu"
                style={{
                  background:'transparent', border:'none', outline:'none',
                  color:'var(--ink)', fontFamily:'var(--ff-b)', fontSize:14,
                  padding:'4px 0', borderBottom:'1px solid var(--line)',
                  width:'100%',
                }}
              />
            </div>
            <div className="sg-item" style={{flexDirection:'column',alignItems:'stretch',gap:6}}>
              <label style={{fontSize:12,color:'var(--ink4)',fontFamily:'var(--ff-m)'}}>Domeniu de activitate</label>
              <input
                type="text"
                value={domain}
                onChange={e => setDomain(e.target.value)}
                placeholder="ex. SaaS, E-commerce, Consultanță..."
                style={{
                  background:'transparent', border:'none', outline:'none',
                  color:'var(--ink)', fontFamily:'var(--ff-b)', fontSize:14,
                  padding:'4px 0', borderBottom:'1px solid var(--line)',
                  width:'100%',
                }}
              />
            </div>
          </div>
        </div>

        {/* Limbă */}
        <div className="sg-group">
          <div className="sg-lbl">Preferințe</div>
          <div className="sg-items">
            <div className="sg-item">
              <div className="sg-icon-wrap">
                <svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10 15.3 15.3 0 014-10z"/></svg>
              </div>
              <div style={{flex:1}}>
                <div className="sg-name">Limbă</div>
                <div className="sg-sub">{{ro:'Română',en:'English',ru:'Русский'}[lang]}</div>
              </div>
              <div className="sg-right" style={{gap:5}}>
                {(['ro','en','ru'] as Lang[]).map(l => (
                  <button
                    key={l}
                    className={`sg-lang-btn${lang===l?' on':''}`}
                    onClick={() => setLang(l)}
                  >
                    {l.toUpperCase()}
                  </button>
                ))}
              </div>
            </div>
          </div>
        </div>

        {/* Save */}
        <button
          className="auth-btn"
          style={{marginTop:8, opacity: saving ? 0.7 : 1}}
          onClick={save}
          disabled={saving}
        >
          {saving ? 'Se salvează...' : saved ? '✓ Salvat!' : 'Salvează profilul'}
        </button>

        <button
          onClick={() => router.back()}
          style={{
            marginTop:12, width:'100%', padding:'11px 0', borderRadius:8,
            border:'1px solid var(--line)', background:'transparent',
            color:'var(--ink3)', fontFamily:'var(--ff-b)', fontSize:14, cursor:'pointer',
          }}
        >
          Înapoi
        </button>
      </div>
    </AppShell>
  )
}
