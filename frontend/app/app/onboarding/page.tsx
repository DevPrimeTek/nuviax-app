'use client'
import { useState, useEffect } from 'react'

type Step = 'welcome' | 'input' | 'parsing' | 'suggestions' | 'analyzing' | 'done'

type Category = 'HEALTH' | 'CAREER' | 'FINANCE' | 'RELATIONSHIPS' | 'LEARNING' | 'CREATIVITY' | 'OTHER'
type BehaviorModel = 'CREATE' | 'INCREASE' | 'REDUCE' | 'MAINTAIN' | 'EVOLVE'

const CATEGORIES: { key: Category; label: string; emoji: string }[] = [
  { key: 'HEALTH',        label: 'Sănătate',     emoji: '🏃' },
  { key: 'CAREER',        label: 'Carieră',       emoji: '💼' },
  { key: 'FINANCE',       label: 'Finanțe',       emoji: '💰' },
  { key: 'RELATIONSHIPS', label: 'Relații',        emoji: '🤝' },
  { key: 'LEARNING',      label: 'Educație',       emoji: '📚' },
  { key: 'CREATIVITY',    label: 'Creativitate',   emoji: '🎨' },
  { key: 'OTHER',         label: 'Altele',         emoji: '✨' },
]

const ANALYSIS_STEPS = [
  'Identificare pattern comportamental...',
  'Structurare obiective pe sprint-uri...',
  'Calibrare frecvență activități zilnice...',
  'Creare plan de execuție personalizat...',
  'Configurare sistem de urmărire progres...',
  'GO-urile tale sunt pregătite ✦',
]

const GOAL_COLORS = ['var(--l0)', 'var(--l2)', 'var(--l5)']

type GoalSuggestion = {
  text: string
  category: Category
  behavior_model: BehaviorModel
  confidence: number
}

type ParsedGoal = {
  rawText: string
  suggestions: GoalSuggestion[]
}

export default function OnboardingPage() {
  const [step, setStep] = useState<Step>('welcome')
  const [userName, setUserName] = useState('')
  const [goInputs, setGoInputs] = useState<string[]>([''])
  const [currentGoIndex, setCurrentGoIndex] = useState(0)

  // Results from POST /goals/parse
  const [parsedGoals, setParsedGoals] = useState<ParsedGoal[]>([])
  const [suggestionGoIndex, setSuggestionGoIndex] = useState(0)
  const [chosenGoals, setChosenGoals] = useState<GoalSuggestion[]>([])

  const [analysisStep, setAnalysisStep] = useState(0)
  const [createdGoals, setCreatedGoals] = useState<{ id: string; name: string; status: string }[]>([])
  const [createError, setCreateError] = useState<string | null>(null)

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

  async function handleAnalyze() {
    const filled = goInputs.filter(g => g.trim())
    if (!filled.length) return

    setStep('parsing')
    const results: ParsedGoal[] = []

    for (const text of filled) {
      try {
        const res = await fetch('/api/proxy/goals/parse', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ text }),
        })
        if (res.ok) {
          const data = await res.json()
          const raw: GoalSuggestion[] = Array.isArray(data.suggestions) ? data.suggestions : []
          const suggestions = raw.filter(s => s && typeof s.text === 'string' && s.text.trim())
          results.push({
            rawText: text,
            suggestions: suggestions.length > 0 ? suggestions : [
              { text, category: 'OTHER', behavior_model: 'INCREASE', confidence: 0.5 },
            ],
          })
        } else {
          results.push({
            rawText: text,
            suggestions: [{ text, category: 'OTHER', behavior_model: 'INCREASE', confidence: 0.5 }],
          })
        }
      } catch {
        results.push({
          rawText: text,
          suggestions: [{ text, category: 'OTHER', behavior_model: 'INCREASE', confidence: 0.5 }],
        })
      }
    }

    setParsedGoals(results)
    setSuggestionGoIndex(0)
    setChosenGoals([])
    setStep('suggestions')
  }

  function handleSuggestionChosen(suggestion: GoalSuggestion) {
    const newChosen = [...chosenGoals, suggestion]
    setChosenGoals(newChosen)

    const next = suggestionGoIndex + 1
    if (next < parsedGoals.length) {
      setSuggestionGoIndex(next)
    } else {
      setStep('analyzing')
      runAnalysis(newChosen)
    }
  }

  async function runAnalysis(goals: GoalSuggestion[]) {
    for (let i = 0; i < ANALYSIS_STEPS.length; i++) {
      await new Promise(r => setTimeout(r, i === ANALYSIS_STEPS.length - 1 ? 600 : 700))
      setAnalysisStep(i)
    }

    const today = new Date()
    const end = new Date(today)
    end.setDate(end.getDate() + 90)
    const startStr = today.toISOString().split('T')[0]
    const endStr = end.toISOString().split('T')[0]

    const created: { id: string; name: string; status: string }[] = []
    const errors: string[] = []

    for (let i = 0; i < goals.length; i++) {
      const g = goals[i]
      const name = g.text.length > 80 ? g.text.slice(0, 80) + '...' : g.text
      try {
        const res = await fetch('/api/proxy/goals', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            name,
            description: parsedGoals[i]?.rawText || g.text,
            start_date: startStr,
            end_date: endStr,
            dominant_behavior_model: g.behavior_model,
            domain: g.category,
          }),
        })
        if (res.ok) {
          const goal = await res.json()
          created.push({ id: goal.id, name: goal.name, status: goal.status || 'ACTIVE' })
        } else {
          const body = await res.text().catch(() => '')
          errors.push(`GO ${i + 1}: ${res.status} ${body.slice(0, 120)}`)
        }
      } catch (e: unknown) {
        errors.push(`GO ${i + 1}: ${e instanceof Error ? e.message : 'network error'}`)
      }
    }

    setCreatedGoals(created)
    if (created.length === 0 && errors.length > 0) {
      setCreateError(errors[0])
    }
    setStep('done')
  }

  if (step === 'welcome') return (
    <div className="auth-page">
      <div className="auth-card" style={{ maxWidth: 480 }}>
        <div className="auth-logo">NUVia<span>X</span></div>
        <div style={{ fontSize: 28, fontWeight: 800, fontFamily: 'var(--ff-h)', marginBottom: 8, lineHeight: 1.2 }}>
          Bun venit{userName ? `, ${userName.split(' ')[0]}` : ''}! 👋
        </div>
        <p style={{ color: 'var(--ink3)', fontSize: 15, lineHeight: 1.6, marginBottom: 24 }}>
          NuviaX te ajută să-ți urmărești obiectivele mari — pe care le numim <strong style={{ color: 'var(--l0l)' }}>GO</strong>-uri.
          Poți avea maxim <strong>3 GO-uri active</strong> în același timp, ca să rămâi concentrat.
        </p>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 10, marginBottom: 28 }}>
          {[
            'Descrie obiectivele tale în cuvintele tale',
            'AI NuviaX generează variante SMART concrete',
            'Tu alegi varianta care ți se potrivește cel mai bine',
          ].map((t, i) => (
            <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
              <div style={{ width: 22, height: 22, borderRadius: '50%', background: 'var(--l0g)', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="var(--l0l)" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><polyline points="20,6 9,17 4,12" /></svg>
              </div>
              <span style={{ fontSize: 14, color: 'var(--ink2)' }}>{t}</span>
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
    const canAddMore = currentGoIndex < 2 && !!goInputs[currentGoIndex]?.trim()

    return (
      <div className="auth-page">
        <div className="auth-card" style={{ maxWidth: 520 }}>
          <div className="auth-logo">NUVia<span>X</span></div>

          <div style={{ display: 'flex', gap: 6, marginBottom: 20 }}>
            {[0, 1, 2].map(i => (
              <div key={i} style={{
                flex: 1, height: 4, borderRadius: 4,
                background: i < filledCount ? 'var(--l0)' : i === currentGoIndex ? 'var(--l0g)' : 'var(--line)',
                transition: 'background .3s',
              }} />
            ))}
          </div>

          <div style={{ fontSize: 13, color: 'var(--ink4)', marginBottom: 6, fontFamily: 'var(--ff-m)' }}>
            GO {currentGoIndex + 1} din 3
          </div>
          <div style={{ fontSize: 20, fontWeight: 700, fontFamily: 'var(--ff-h)', marginBottom: 8 }}>
            {currentGoIndex === 0
              ? 'Care este primul tău mare obiectiv?'
              : currentGoIndex === 1
              ? 'Al doilea GO (opțional)'
              : 'Al treilea GO (opțional)'}
          </div>
          <p style={{ fontSize: 13, color: 'var(--ink4)', marginBottom: 14, lineHeight: 1.5 }}>
            Scrie în cuvintele tale — nu trebuie să fie perfect. NuviaX va genera variante SMART din care vei alege.
          </p>

          <textarea
            value={goInputs[currentGoIndex] || ''}
            onChange={e => setGoInputs(prev => { const n = [...prev]; n[currentGoIndex] = e.target.value; return n })}
            placeholder="Ex: Vreau să slăbesc, să câștig mai mult, să învăț programare..."
            rows={4}
            style={{
              width: '100%', padding: '12px 14px', borderRadius: 10,
              border: '1.5px solid var(--line)', background: 'var(--bg2)',
              color: 'var(--ink)', fontFamily: 'var(--ff-b)', fontSize: 14,
              resize: 'vertical', outline: 'none', lineHeight: 1.6,
              boxSizing: 'border-box',
            }}
            onFocus={e => e.target.style.borderColor = 'var(--l0)'}
            onBlur={e => e.target.style.borderColor = 'var(--line)'}
          />

          <div style={{ display: 'flex', gap: 10, marginTop: 16 }}>
            {currentGoIndex < 2 && (
              <button
                onClick={handleAddGo}
                disabled={!canAddMore}
                style={{
                  flex: 1, padding: '14px 12px', borderRadius: 14,
                  border: 'none',
                  background: canAddMore ? 'var(--l0)' : 'var(--bg3)',
                  color: canAddMore ? 'white' : 'var(--ink4)',
                  fontFamily: 'var(--ff-d)', fontSize: 14, fontWeight: 700,
                  cursor: canAddMore ? 'pointer' : 'default',
                  display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8,
                  minHeight: 50,
                  boxShadow: canAddMore ? '0 4px 16px rgba(124,58,237,.3)' : 'none',
                  transition: 'all .2s',
                }}
              >
                <div style={{ width: 20, height: 20, borderRadius: '50%', border: `1.5px solid ${canAddMore ? 'rgba(255,255,255,.5)' : 'var(--line2)'}`, display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                  <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round"><line x1="12" y1="5" x2="12" y2="19" /><line x1="5" y1="12" x2="19" y2="12" /></svg>
                </div>
                Adaugă Obiectiv
              </button>
            )}
            <button
              onClick={handleAnalyze}
              disabled={!canAnalyze}
              className="auth-btn"
              style={{ flex: 2, opacity: canAnalyze ? 1 : 0.5 }}
            >
              Analizează Obiectivele →
            </button>
          </div>

          {filledCount > 0 && currentGoIndex > 0 && (
            <div style={{ marginTop: 16 }}>
              <div style={{ fontSize: 12, color: 'var(--ink4)', marginBottom: 8 }}>GO-uri introduse:</div>
              {goInputs.filter(g => g.trim()).map((g, i) => (
                <div key={i} style={{ fontSize: 13, color: 'var(--ink3)', padding: '6px 10px', background: 'var(--bg2)', borderRadius: 6, marginBottom: 4 }}>
                  <strong style={{ color: 'var(--l0l)' }}>GO {i + 1}:</strong> {g.length > 60 ? g.slice(0, 60) + '...' : g}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    )
  }

  if (step === 'parsing') return (
    <div className="auth-page">
      <div className="auth-card" style={{ maxWidth: 440, textAlign: 'center' }}>
        <div className="auth-logo">NUVia<span>X</span></div>
        <div style={{ fontSize: 22, fontWeight: 800, fontFamily: 'var(--ff-h)', marginBottom: 8 }}>
          Analizez obiectivele tale
        </div>
        <p style={{ color: 'var(--ink3)', fontSize: 14, lineHeight: 1.6, marginBottom: 32 }}>
          AI NuviaX generează variante SMART concrete din ceea ce ai descris...
        </p>
        <div className="spinner" style={{ width: 32, height: 32, margin: '0 auto', borderTopColor: 'var(--l0)', borderWidth: 3 }} />
      </div>
    </div>
  )

  if (step === 'suggestions') {
    const current = parsedGoals[suggestionGoIndex]
    if (!current) return null
    const total = parsedGoals.length

    return (
      <div className="auth-page">
        <div className="auth-card" style={{ maxWidth: 540 }}>
          <div className="auth-logo">NUVia<span>X</span></div>

          <div style={{ display: 'flex', gap: 6, marginBottom: 20 }}>
            {parsedGoals.map((_, i) => (
              <div key={i} style={{
                flex: 1, height: 4, borderRadius: 4,
                background: i < suggestionGoIndex ? 'var(--l0)' : i === suggestionGoIndex ? 'var(--l0g)' : 'var(--line)',
                transition: 'background .3s',
              }} />
            ))}
          </div>

          <div style={{ fontSize: 12, color: 'var(--l0l)', fontFamily: 'var(--ff-m)', fontWeight: 600, marginBottom: 8, letterSpacing: '0.05em' }}>
            OBIECTIV {suggestionGoIndex + 1} DIN {total} · Formulare SMART
          </div>
          <div style={{ fontSize: 20, fontWeight: 700, fontFamily: 'var(--ff-h)', marginBottom: 8 }}>
            Alege varianta care ți se potrivește
          </div>
          <div style={{
            fontSize: 13, color: 'var(--ink4)', marginBottom: 20,
            padding: '10px 12px', background: 'var(--bg2)', borderRadius: 8,
            borderLeft: '3px solid var(--line)',
          }}>
            Ai scris: <em style={{ color: 'var(--ink3)' }}>
              {current.rawText.length > 100 ? current.rawText.slice(0, 100) + '...' : current.rawText}
            </em>
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: 10, marginBottom: 16 }}>
            {current.suggestions.map((s, i) => {
              const cat = CATEGORIES.find(c => c.key === s.category)
              return (
                <button
                  key={i}
                  onClick={() => handleSuggestionChosen(s)}
                  style={{
                    textAlign: 'left', padding: '16px', borderRadius: 12,
                    border: '1.5px solid var(--line)', background: 'var(--bg2)',
                    cursor: 'pointer', transition: 'all .15s', lineHeight: 1.5,
                    width: '100%',
                  }}
                  onMouseEnter={e => { e.currentTarget.style.borderColor = 'var(--l0)'; e.currentTarget.style.background = 'var(--l0g)' }}
                  onMouseLeave={e => { e.currentTarget.style.borderColor = 'var(--line)'; e.currentTarget.style.background = 'var(--bg2)' }}
                >
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
                    <span style={{ fontSize: 16 }}>{cat?.emoji || '✨'}</span>
                    <span style={{ fontSize: 11, color: 'var(--ink4)', fontFamily: 'var(--ff-m)', fontWeight: 600 }}>
                      {cat?.label || 'Altele'}
                    </span>
                    <span style={{
                      fontSize: 10, padding: '2px 8px', borderRadius: 20,
                      background: 'var(--bg3)', color: 'var(--l0l)',
                      fontFamily: 'var(--ff-m)', fontWeight: 600, marginLeft: 'auto',
                    }}>
                      {s.behavior_model}
                    </span>
                  </div>
                  <div style={{ fontSize: 14, fontWeight: 600, color: 'var(--ink)', lineHeight: 1.5 }}>
                    {s.text}
                  </div>
                </button>
              )
            })}
          </div>

          <button
            onClick={() => handleSuggestionChosen({
              text: current.rawText,
              category: 'OTHER',
              behavior_model: 'INCREASE',
              confidence: 0.5,
            })}
            style={{
              width: '100%', padding: '11px 0', borderRadius: 8,
              border: '1.5px solid var(--line)', background: 'transparent',
              color: 'var(--ink4)', fontFamily: 'var(--ff-b)', fontSize: 13, cursor: 'pointer',
            }}
          >
            Păstrează formularea mea originală
          </button>
        </div>
      </div>
    )
  }

  if (step === 'analyzing') return (
    <div className="auth-page">
      <div className="auth-card" style={{ maxWidth: 440, textAlign: 'center' }}>
        <div className="auth-logo">NUVia<span>X</span></div>
        <div style={{ fontSize: 22, fontWeight: 800, fontFamily: 'var(--ff-h)', marginBottom: 8 }}>
          NuviaX Framework
        </div>
        <div style={{ fontSize: 14, color: 'var(--ink3)', marginBottom: 32 }}>
          analizează și structurează GO-urile tale...
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: 12, marginBottom: 32, textAlign: 'left' }}>
          {ANALYSIS_STEPS.map((s, i) => (
            <div key={i} style={{
              display: 'flex', alignItems: 'center', gap: 12,
              opacity: i <= analysisStep ? 1 : 0.25,
              transition: 'opacity .4s',
            }}>
              <div style={{
                width: 20, height: 20, borderRadius: '50%', flexShrink: 0,
                background: i < analysisStep ? 'var(--l0)' : i === analysisStep ? 'var(--l2g)' : 'var(--line)',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                transition: 'background .3s',
              }}>
                {i < analysisStep && (
                  <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round"><polyline points="20,6 9,17 4,12" /></svg>
                )}
                {i === analysisStep && (
                  <div style={{ width: 8, height: 8, borderRadius: '50%', background: 'var(--l2)', animation: 'pulse 1s infinite' }} />
                )}
              </div>
              <span style={{ fontSize: 13, fontFamily: 'var(--ff-b)', color: i <= analysisStep ? 'var(--ink)' : 'var(--ink4)' }}>
                {s}
              </span>
            </div>
          ))}
        </div>

        <div style={{ height: 4, borderRadius: 4, background: 'var(--line)', overflow: 'hidden' }}>
          <div style={{
            height: '100%', borderRadius: 4, background: 'var(--l0)',
            width: `${((analysisStep + 1) / ANALYSIS_STEPS.length) * 100}%`,
            transition: 'width .7s ease',
          }} />
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
      <div className="auth-card" style={{ maxWidth: 480 }}>
        <div className="auth-logo">NUVia<span>X</span></div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 6 }}>
          <div style={{ width: 28, height: 28, borderRadius: '50%', background: 'var(--l0g)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="var(--l0l)" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><polyline points="20,6 9,17 4,12" /></svg>
          </div>
          <span style={{ fontSize: 12, color: 'var(--l0l)', fontFamily: 'var(--ff-m)', fontWeight: 600 }}>GO-urile sunt pregătite</span>
        </div>
        <div style={{ fontSize: 26, fontWeight: 800, fontFamily: 'var(--ff-h)', marginBottom: 6 }}>
          Totul e configurat ✦
        </div>
        <div style={{ fontSize: 14, color: 'var(--ink3)', marginBottom: 24 }}>
          NuviaX a creat planul tău de execuție. Primul GO este activ acum.
        </div>

        {createdGoals.map((g, i) => (
          <div key={g.id} style={{
            padding: '14px 16px', borderRadius: 10,
            border: `1.5px solid ${i === 0 ? 'var(--l0)' : 'var(--line)'}`,
            background: i === 0 ? 'var(--l0g)' : 'var(--bg2)',
            marginBottom: 10,
          }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
              <div style={{ width: 8, height: 8, borderRadius: '50%', background: GOAL_COLORS[i], flexShrink: 0 }} />
              <span style={{ fontSize: 11, fontFamily: 'var(--ff-m)', color: 'var(--l0l)', fontWeight: 600 }}>
                GO {i + 1} · {g.status === 'WAITING' ? 'FUTURE VAULT' : 'ACTIV'}
              </span>
            </div>
            <div style={{ fontSize: 14, fontWeight: 600, color: 'var(--ink)', lineHeight: 1.4 }}>
              {g.name.length > 80 ? g.name.slice(0, 80) + '...' : g.name}
            </div>
            <div style={{ fontSize: 12, color: 'var(--ink4)', marginTop: 4 }}>
              {g.status === 'WAITING' ? 'Future Vault — activat când un GO se încheie' : 'Sprint 1 · 30 zile'}
            </div>
          </div>
        ))}

        {createdGoals.length === 0 && (
          <div style={{ padding: 16, borderRadius: 10, border: '1.5px solid #7f1d1d', background: 'rgba(127,29,29,.15)', marginBottom: 16, color: '#fecaca', fontSize: 13, lineHeight: 1.5 }}>
            <div style={{ fontWeight: 600, marginBottom: 4, color: '#fca5a5' }}>Niciun GO nu a fost creat.</div>
            {createError
              ? <div style={{ fontFamily: 'var(--ff-m)', fontSize: 12, opacity: .9, wordBreak: 'break-word' }}>{createError}</div>
              : <div>Verifică conexiunea și încearcă din nou.</div>}
          </div>
        )}

        <button className="auth-btn" style={{ marginTop: 8 }} onClick={() => { window.location.href = '/dashboard' }}>
          Intră în aplicație →
        </button>
      </div>
    </div>
  )
}
