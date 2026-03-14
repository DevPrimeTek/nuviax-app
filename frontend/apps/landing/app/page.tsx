import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'NUViaX — Crești deliberat, zi cu zi',
}

const APP = process.env.NEXT_PUBLIC_APP_URL || 'https://nuviax.app'

const FEATURES = [
  { color:'var(--l0l)', bg:'var(--l0g)', title:'Un obiectiv → activități zilnice',
    desc:'Scrii ce vrei să realizezi. Sistemul generează automat activitățile zilnice potrivite pentru nivelul tău actual.',
    icon:<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M12 2l2.4 7.4H22l-6.2 4.5 2.4 7.4L12 17l-6.2 4.3 2.4-7.4L2 9.4h7.6z"/></svg> },
  { color:'var(--l2l)', bg:'var(--l2g)', title:'9 etape, 365 de zile',
    desc:'Fiecare obiectiv este împărțit în 9 etape cu durată variabilă. Progresezi treptat, nu aleatoriu.',
    icon:<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/></svg> },
  { color:'var(--ul)', bg:'var(--ug)', title:'Ritm adaptat la viața ta',
    desc:'Dacă ai o perioadă grea, sistemul recalibrează. Zilele de pauză nu te aruncă din progres.',
    icon:<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12,6 12,12 16,14"/></svg> },
  { color:'var(--l5l)', bg:'var(--l5g)', title:'Metoda rămâne invizibilă',
    desc:'Nu vei vedea niciodată formule sau scoruri interne. Doar activitățile tale de azi.',
    icon:<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg> },
  { color:'var(--l1l)', bg:'var(--l1g)', title:'Română, Engleză, Rusă',
    desc:'Interfața completă în 3 limbi. Schimbi dintr-o apăsare, fără reload.',
    icon:<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10 15.3 15.3 0 014-10z"/></svg> },
  { color:'var(--l3l)', bg:'var(--l3g)', title:'Mobile-first, și pe desktop',
    desc:'Proiectat pentru telefon. Disponibil și pe web, cu aceeași experiență clară.',
    icon:<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><rect x="5" y="2" width="14" height="20" rx="2"/><line x1="12" y1="18" x2="12.01" y2="18"/></svg> },
]

const STEPS = [
  { n:'01', title:'Scrii un obiectiv',       desc:'O frază simplă: ce vrei să realizezi și în cât timp.' },
  { n:'02', title:'Sistemul generează totul', desc:'9 etape, checkpoints, activități zilnice — create automat.' },
  { n:'03', title:'Faci activitățile de azi', desc:'2–3 activități clare pe zi. Bifezi. Gata.' },
  { n:'04', title:'Recapitulezi la final',    desc:'Răspunzi la câteva întrebări. Sistemul recalibrează.' },
]

export default function LandingPage() {
  return (
    <div className="page">

      {/* ── Nav ── */}
      <nav className="nav">
        <div className="nav-inner">
          <a href="/" className="nav-logo">NUVia<span>X</span></a>
          <div className="nav-links">
            <a href="#features">Funcții</a>
            <a href="#how">Cum funcționează</a>
            <a href="#pricing">Prețuri</a>
          </div>
          <div className="nav-actions">
            <a href={`${APP}/login`} className="nav-login">Intră în cont</a>
            <a href={`${APP}/register`} className="nav-cta">Începe gratuit</a>
          </div>
        </div>
      </nav>

      {/* ── Hero ── */}
      <section className="hero">
        <div className="hero-bg">
          <div className="hero-glow1"/>
          <div className="hero-glow2"/>
          <div className="hero-grid"/>
        </div>
        <div className="hero-inner">
          <div>
            <div className="hero-badge">
              <span className="hero-badge-dot"/>
              Beta 1.0 · Disponibil acum
            </div>
            <h1 className="hero-title">
              Crești deliberat,<br/><em>zi cu zi.</em>
            </h1>
            <p className="hero-sub">
              Scrii un obiectiv. Primești activitățile zilnice potrivite, organizate
              în 9 etape pe 365 de zile. Fără teorie. Doar acțiune.
            </p>
            <div className="hero-ctas">
              <a href={`${APP}/register`} className="hero-primary">
                Încearcă gratuit
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round"><path d="M5 12h14M12 5l7 7-7 7"/></svg>
              </a>
              <a href={`${APP}/login`} className="hero-secondary">Am deja cont</a>
            </div>
            <p className="hero-note">Gratuit 14 zile · Fără card bancar</p>
          </div>

          {/* Phone mockup */}
          <div className="phone-wrap">
            <div className="phone-mock">
              <div className="phone-bar">
                <span>9:41</span>
                <svg width="32" height="10" viewBox="0 0 38 12" fill="none">
                  <rect x="0" y="4" width="3" height="8" rx="1" fill="currentColor" opacity=".6"/>
                  <rect x="5" y="2.5" width="3" height="9.5" rx="1" fill="currentColor" opacity=".7"/>
                  <rect x="10" y=".5" width="3" height="11.5" rx="1" fill="currentColor"/>
                  <rect x="17" y="1" width="18" height="10" rx="2.5" stroke="currentColor" strokeWidth="1.2"/>
                  <rect x="18.5" y="2.5" width="13" height="7" rx="1.5" fill="currentColor"/>
                </svg>
              </div>
              <div className="phone-content">
                <div className="pm-greeting">
                  <div className="pm-name">Bună, <span>Alexandru.</span></div>
                  <div className="pm-date">Joi · Ziua 14 din Etapa 3</div>
                </div>
                <div className="pm-sprint">
                  <div className="pm-sprint-top">
                    <div className="pm-sprint-dot"/>
                    <span>Etapa 3 · 16 zile rămase</span>
                    <span style={{marginLeft:'auto',color:'var(--l5l)',fontFamily:'var(--ff-m)',fontSize:'10px',fontWeight:700}}>65%</span>
                  </div>
                  <div className="pm-sprint-bar"><div style={{width:'65%'}}/></div>
                </div>
                <div className="pm-sec">Activitățile de azi</div>
                <div className="pm-task">
                  <div className="pm-chk done">
                    <svg viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><polyline points="20,6 9,17 4,12"/></svg>
                  </div>
                  <div>
                    <div className="pm-task-text" style={{textDecoration:'line-through',opacity:.45}}>Trimite oferta la ClientX</div>
                    <div className="pm-task-meta"><span className="pm-badge main">Principal</span><span style={{color:'var(--ink4)',fontSize:'9px',fontFamily:'var(--ff-m)'}}>~30 min</span></div>
                  </div>
                </div>
                <div className="pm-task">
                  <div className="pm-chk"/>
                  <div>
                    <div className="pm-task-text">Finalizează pachetul Premium</div>
                    <div className="pm-task-meta"><span className="pm-badge main">Principal</span><span style={{color:'var(--ink4)',fontSize:'9px',fontFamily:'var(--ff-m)'}}>~45 min</span></div>
                  </div>
                </div>
                <div className="pm-task" style={{opacity:.55}}>
                  <div className="pm-chk" style={{borderColor:'rgba(37,99,235,.4)'}}/>
                  <div>
                    <div className="pm-task-text">Citesc un articol despre servicii</div>
                    <div className="pm-task-meta"><span className="pm-badge pers">Personal</span><span style={{color:'var(--ink4)',fontSize:'9px',fontFamily:'var(--ff-m)'}}>~20 min</span></div>
                  </div>
                </div>
              </div>
              <div className="phone-nav">
                {['🏠','⏱️','📊','✦','⚙️'].map((ic,i)=>(
                  <div key={i} className={`phone-nav-tab${i===0?' active':''}`}>{ic}</div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ── Stats ── */}
      <section className="stats-strip">
        <div className="stats-inner">
          {[{n:'9',l:'etape per obiectiv'},{n:'365',l:'zile de activități'},{n:'3',l:'obiective simultane'},{n:'3',l:'limbi disponibile'}].map((s,i)=>(
            <div key={i} className="stat">
              <div className="stat-n">{s.n}</div>
              <div className="stat-l">{s.l}</div>
            </div>
          ))}
        </div>
      </section>

      {/* ── Features ── */}
      <section id="features">
        <div className="section-inner">
          <div className="section-label">Funcționalități</div>
          <h2 className="section-title">Tot ce îți trebuie<br/>pentru a progresa zilnic</h2>
          <div className="feat-grid">
            {FEATURES.map((f,i)=>(
              <div key={i} className="feat-card">
                <div className="feat-icon" style={{background:f.bg,color:f.color}}>{f.icon}</div>
                <h3 className="feat-title">{f.title}</h3>
                <p className="feat-desc">{f.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── How ── */}
      <section className="how" id="how">
        <div className="section-inner">
          <div className="section-label">Cum funcționează</div>
          <h2 className="section-title">De la obiectiv la acțiune<br/>în 4 pași</h2>
          <div className="steps">
            {STEPS.map((s,i)=>(
              <div key={i}>
                <div className="step-n">{s.n}</div>
                <div className="step-line"/>
                <h3 className="step-title">{s.title}</h3>
                <p className="step-desc">{s.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── Pricing ── */}
      <section id="pricing">
        <div className="section-inner">
          <div className="section-label">Prețuri</div>
          <h2 className="section-title">Simplu și transparent</h2>
          <div className="pricing-grid">
            <div className="pricing-card">
              <div className="pricing-name">Gratuit</div>
              <div className="pricing-price"><span className="pricing-amt">0</span><span className="pricing-cur">RON/lună</span></div>
              <div className="pricing-desc">14 zile trial complet, fără card</div>
              <ul className="pricing-list">
                <li>1 obiectiv activ</li>
                <li>Activități zilnice generate</li>
                <li>Istoricul etapelor</li>
                <li>3 limbi</li>
              </ul>
              <a href={`${APP}/register`} className="pricing-btn ghost">Începe gratuit</a>
            </div>
            <div className="pricing-card featured">
              <div className="pricing-badge">Recomandat</div>
              <div className="pricing-name">Pro</div>
              <div className="pricing-price"><span className="pricing-amt">49</span><span className="pricing-cur">RON/lună</span></div>
              <div className="pricing-desc">Tot ce ai nevoie pentru a progresa serios</div>
              <ul className="pricing-list">
                <li>3 obiective simultane</li>
                <li>Listă de așteptare nelimitată</li>
                <li>Recalibrare automată ritm</li>
                <li>Recap detaliat de etapă</li>
                <li>Export date</li>
                <li>Suport prioritar</li>
              </ul>
              <a href={`${APP}/register`} className="pricing-btn">Încearcă Pro — 14 zile gratuit</a>
            </div>
          </div>
        </div>
      </section>

      {/* ── CTA final ── */}
      <section className="cta-final">
        <div className="cta-final-bg"/>
        <div className="cta-final-inner">
          <h2 className="cta-final-title">Ești pregătit să crești<br/>deliberat?</h2>
          <p className="cta-final-sub">Primul obiectiv este gratuit. Fără card, fără complicații.</p>
          <a href={`${APP}/register`} className="cta-final-btn">
            Creează contul
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round"><path d="M5 12h14M12 5l7 7-7 7"/></svg>
          </a>
        </div>
      </section>

      {/* ── Footer ── */}
      <footer className="footer">
        <div className="footer-inner">
          <div className="footer-logo">NUVia<span>X</span></div>
          <div className="footer-links">
            <a href="/privacy">Confidențialitate</a>
            <a href="/terms">Termeni</a>
            <a href="mailto:hello@nuviax.app">Contact</a>
          </div>
          <div className="footer-copy">© 2026 NUViaX. Toate drepturile rezervate.</div>
        </div>
      </footer>

    </div>
  )
}
