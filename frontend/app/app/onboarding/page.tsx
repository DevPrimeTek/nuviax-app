'use client'
import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'

type Step = 'welcome' | 'input' | 'analyzing' | 'done'

const ANALYSIS_STEPS = [
  'Identificare pattern comportamental...',
  'Structurare obiective pe sprint-uri...',
  'Calibrare frecvență activități zilnice...',
  'Creare plan de execuție personalizat...',
  'Configurare sistem de urmărire progres...',
  'GO-urile tale sunt pregătite ✦',
]

const GOAL_COLORS = ['var(--l0)', 'var(--l2)', 'var(--l5)']

export default function OnboardingPage() {
  const router = useRouter()
  const [step, setStep] = useState<Step>('welcome')
  const [userName, setUserName] = useState('')
  const [goInputs, setGoInputs] = useState<string[]>([''])
  const [currentGoIndex, setCurrentGoIndex] = useState(0)
  const [analysisStep, setAnalysisStep] = useState(0)
  const [createdGoals, setCreatedGoals] = useState<{id:string;name:string;status:string}[]>([])

  useEffect(() => {
    fetch('/api/proxy/settings')
      .then(r => r.ok ? r.json() : null)
      .then(d => {
        if (d) {
          const name = d.user_name || d.full_name || ''
          if (name) setUserName(name)
        }
      })
      .catch(() => {})
  }, [])

  function handleAddGo() {
    const text = goInputs[currentGoIndex]?.trim()
    if (!text) return
    if (currentGoIndex < 2) {
      setGoInputs(prev => {
        const next = [...prev]
        if (!next[currentGoIndex + 1]) next[currentGoIndex + 1] = ''
        return next
      })
      setCurrentGoIndex(i => i + 1)
    }
  }

  function handleAnalyze() {
    const filled = goInputs.filter(g => g.trim())
    if (!filled.length) return
    setStep('analyzing')
    runAnalysis(filled)
  }

  async function runAnalysis(goals: string[]) {
    // Animatie progresiva
    for (let i = 0; i < ANALYSIS_STEPS.length; i++) {
      await new Promise(r => setTimeout(r, i === ANALYSIS_STEPS.length - 1 ? 600 : 700))
      setAnalysisStep(i)
    }

    // Creare goals in backend
    const today = new Date()
    const end = new Date(today)
    end.setDate(end.getDate() + 90)
    const startStr = today.toISOString().split('T')[0]
    const endStr = end.toISOString().split('T')[0]

    const created: {id:string;name:string;status:string}[] = []
    for (let i = 0; i < goals.length; i++) {
      const text = goals[i].trim()
      const name = text.length > 80 ? text.slice(0, 80) + '...' : text
      try {
        const res = await fetch('/api/proxy/goals', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            name,
            description: text,
            start_date: startStr,
            end_date: endStr,
            waiting_list: i > 0,
          }),
        })
        if (res.ok) {
          const goal = await res.json()
          created.push({ id: goal.id, name: goal.name, status: i === 0 ? 'ACTIVE' : 'WAITING' })
        }
      } catch {}
    }

    setCreatedGoals(created)
    setStep('done')
  }

  if (step === 'welcome') return (
    <div className="auth-page">
      <div className="auth-card" style={{maxWidth:480}}>
        <div className="auth-logo">NUVia<span>X</span></div>
        <div style={{fontSize:28,fontWeight:800,fontFamily:'var(--ff-h)',marginBottom:8,lineHeight:1.2}}>
          Bun venit{userName ? `, ${userName.split(' ')[0]}` : ''}! 👋
        </div>
        <p style={{color:'var(--ink3)',fontSize:15,lineHeight:1.6,marginBottom:24}}>
          NuviaX te ajută să-ți urmărești obiectivele mari — pe care le numim <strong style={{color:'var(--l0l)'}}>GO</strong>-uri.
          Poți avea maxim <strong>3 GO-uri active</strong> în același timp, ca să rămâi concentrat.
        </p>
        <div style={{display:'flex',flexDirection:'column',gap:10,marginBottom:28}}>
          {['Definești 1–3 GO-uri în cuvintele tale', 'NuviaX le analizează și creează un plan de sprint', 'Urmărești progresul zi cu zi'].map((t, i) => (
            <div key={i} style={{display:'flex',alignItems:'center',gap:12}}>
              <div style={{width:22,height:22,borderRadius:'50%',background:'var(--l0g)',display:'flex',alignItems:'center',justifyContent:'center',flexShrink:0}}>
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="var(--l0l)" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><polyline points="20,6 9,17 4,12"/></svg>
              </div>
              <span style={{fontSize:14,color:'var(--ink2)'}}>{t}</span>
            </div>
          ))}
        </div>
        <button className="auth-btn" onClick={() => setStep('input')}>
          Hai să începem →
        </button>
      </div>
    </div>
  )

  if (step === 'input') {
    const filledCount = goInputs.filter(g => g.trim()).length
    const canAnalyze = filledCount > 0
    const canAddMore = currentGoIndex < 2 && goInputs[currentGoIndex]?.trim()

    return (
      <div className="auth-page">
        <div className="auth-card" style={{maxWidth:520}}>
          <div className="auth-logo">NUVia<span>X</span></div>

          {/* Indicator GO x/3 */}
          <div style={{display:'flex',gap:6,marginBottom:20}}>
            {[0,1,2].map(i => (
              <div key={i} style={{
                flex:1, height:4, borderRadius:4,
                background: i < filledCount ? 'var(--l0)' : i === currentGoIndex ? 'var(--l0g)' : 'var(--line)',
                transition:'background .3s',
              }}/>
            ))}
          </div>

          <div style={{fontSize:13,color:'var(--ink4)',marginBottom:6,fontFamily:'var(--ff-m)'}}>
            GO {currentGoIndex + 1} din 3
          </div>
          <div style={{fontSize:20,fontWeight:700,fontFamily:'var(--ff-h)',marginBottom:16}}>
            {currentGoIndex === 0 ? 'Care este primul tău mare obiectiv?' : currentGoIndex === 1 ? 'Al doilea GO (opțional)' : 'Al treilea GO (opțional)'}
          </div>

          <textarea
            value={goInputs[currentGoIndex] || ''}
            onChange={e => setGoInputs(prev => { const n=[...prev]; n[currentGoIndex]=e.target.value; return n })}
            placeholder="Ex: Vreau să lansez un produs SaaS care să genereze 5.000 RON MRR până în septembrie..."
            rows={5}
            style={{
              width:'100%', padding:'12px 14px', borderRadius:10,
              border:'1.5px solid var(--line)', background:'var(--bg2)',
              color:'var(--ink)', fontFamily:'var(--ff-b)', fontSize:14,
              resize:'vertical', outline:'none', lineHeight:1.6,
              boxSizing:'border-box',
            }}
            onFocus={e => e.target.style.borderColor='var(--l0)'}
            onBlur={e => e.target.style.borderColor='var(--line)'}
          />

          <div style={{display:'flex',gap:10,marginTop:16}}>
            {currentGoIndex < 2 && (
              <button
                onClick={handleAddGo}
                disabled={!canAddMore}
                style={{
                  flex:1, padding:'11px 0', borderRadius:8,
                  border:'1.5px solid var(--line)', background:'transparent',
                  color: canAddMore ? 'var(--ink)' : 'var(--ink4)',
                  fontFamily:'var(--ff-b)', fontSize:14, cursor: canAddMore ? 'pointer' : 'default',
                }}
              >
                + Adaugă GO {currentGoIndex + 2}
              </button>
            )}
            <button
              onClick={handleAnalyze}
              disabled={!canAnalyze}
              className="auth-btn"
              style={{flex:2, opacity: canAnalyze ? 1 : 0.5}}
            >
              Analizează GO-urile →
            </button>
          </div>

          {filledCount > 0 && currentGoIndex > 0 && (
            <div style={{marginTop:16}}>
              <div style={{fontSize:12,color:'var(--ink4)',marginBottom:8}}>GO-uri introduse:</div>
              {goInputs.filter(g=>g.trim()).map((g,i) => (
                <div key={i} style={{fontSize:13,color:'var(--ink3)',padding:'6px 10px',background:'var(--bg2)',borderRadius:6,marginBottom:4}}>
                  <strong style={{color:'var(--l0l)'}}>GO {i+1}:</strong> {g.length > 60 ? g.slice(0,60)+'...' : g}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    )
  }

  if (step === 'analyzing') return (
    <div className="auth-page">
      <div className="auth-card" style={{maxWidth:440,textAlign:'center'}}>
        <div className="auth-logo">NUVia<span>X</span></div>
        <div style={{fontSize:22,fontWeight:800,fontFamily:'var(--ff-h)',marginBottom:8}}>
          NuviaX Framework
        </div>
        <div style={{fontSize:14,color:'var(--ink3)',marginBottom:32}}>
          analizează și structurează GO-urile tale...
        </div>

        {/* Animatie pași */}
        <div style={{display:'flex',flexDirection:'column',gap:12,marginBottom:32,textAlign:'left'}}>
          {ANALYSIS_STEPS.map((s, i) => (
            <div key={i} style={{
              display:'flex', alignItems:'center', gap:12,
              opacity: i <= analysisStep ? 1 : 0.25,
              transition:'opacity .4s',
            }}>
              <div style={{
                width:20, height:20, borderRadius:'50%', flexShrink:0,
                background: i < analysisStep ? 'var(--l0)' : i === analysisStep ? 'var(--l2g)' : 'var(--line)',
                display:'flex', alignItems:'center', justifyContent:'center',
                transition:'background .3s',
              }}>
                {i < analysisStep && (
                  <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round"><polyline points="20,6 9,17 4,12"/></svg>
                )}
                {i === analysisStep && (
                  <div style={{width:8,height:8,borderRadius:'50%',background:'var(--l2)',animation:'pulse 1s infinite'}}/>
                )}
              </div>
              <span style={{fontSize:13,fontFamily:'var(--ff-b)',color: i <= analysisStep ? 'var(--ink)' : 'var(--ink4)'}}>
                {s}
              </span>
            </div>
          ))}
        </div>

        <div style={{height:4,borderRadius:4,background:'var(--line)',overflow:'hidden'}}>
          <div style={{
            height:'100%', borderRadius:4, background:'var(--l0)',
            width:`${((analysisStep + 1) / ANALYSIS_STEPS.length) * 100}%`,
            transition:'width .7s ease',
          }}/>
        </div>

        <style jsx>{`
          @keyframes pulse {
            0%, 100% { opacity: 1; transform: scale(1); }
            50% { opacity: 0.5; transform: scale(0.8); }
          }
        `}</style>
      </div>
    </div>
  )

  // step === 'done'
  return (
    <div className="auth-page">
      <div className="auth-card" style={{maxWidth:480}}>
        <div className="auth-logo">NUVia<span>X</span></div>
        <div style={{display:'flex',alignItems:'center',gap:8,marginBottom:6}}>
          <div style={{width:28,height:28,borderRadius:'50%',background:'var(--l0g)',display:'flex',alignItems:'center',justifyContent:'center'}}>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="var(--l0l)" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><polyline points="20,6 9,17 4,12"/></svg>
          </div>
          <span style={{fontSize:12,color:'var(--l0l)',fontFamily:'var(--ff-m)',fontWeight:600}}>GO-urile sunt pregătite</span>
        </div>
        <div style={{fontSize:26,fontWeight:800,fontFamily:'var(--ff-h)',marginBottom:6}}>
          Totul e configurat ✦
        </div>
        <div style={{fontSize:14,color:'var(--ink3)',marginBottom:24}}>
          NuviaX a creat planul tău de execuție. Primul GO este activ acum.
        </div>

        {createdGoals.map((g, i) => (
          <div key={g.id} style={{
            padding:'14px 16px', borderRadius:10,
            border:`1.5px solid ${i === 0 ? 'var(--l0)' : 'var(--line)'}`,
            background: i === 0 ? 'var(--l0g)' : 'var(--bg2)',
            marginBottom:10,
          }}>
            <div style={{display:'flex',alignItems:'center',gap:8,marginBottom:4}}>
              <div style={{width:8,height:8,borderRadius:'50%',background:GOAL_COLORS[i],flexShrink:0}}/>
              <span style={{fontSize:11,fontFamily:'var(--ff-m)',color:i===0?'var(--l0l)':'var(--ink4)',fontWeight:600}}>
                GO {i+1} · {i === 0 ? 'ACTIV' : 'ÎN AȘTEPTARE'}
              </span>
            </div>
            <div style={{fontSize:14,fontWeight:600,color:'var(--ink)',lineHeight:1.4}}>
              {g.name.length > 80 ? g.name.slice(0,80)+'...' : g.name}
            </div>
            <div style={{fontSize:12,color:'var(--ink4)',marginTop:4}}>
              Sprint 1 · 90 zile
            </div>
          </div>
        ))}

        {createdGoals.length === 0 && (
          <div style={{padding:16,borderRadius:10,border:'1.5px solid var(--line)',background:'var(--bg2)',marginBottom:16,color:'var(--ink3)',fontSize:14}}>
            GO-urile vor apărea în dashboard după ce sunt procesate.
          </div>
        )}

        <button className="auth-btn" style={{marginTop:8}} onClick={() => { window.location.href = '/dashboard' }}>
          Intră în aplicație →
        </button>
      </div>
    </div>
  )
}
