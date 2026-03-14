'use client'
import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import AppShell from '@/components/layout/AppShell'

interface RecapData {
  sprint_name: string; score: number; grade: string
  days_active: number; days_total: number; streak: number
  mrr_delta: number; next_sprint_name: string; goal_id: string
}

export default function RecapPage() {
  const router = useRouter()
  const [data, setData] = useState<RecapData|null>(null)
  const [q1, setQ1] = useState('')
  const [q2, setQ2] = useState('')
  const [energy, setEnergy] = useState(8)
  const [submitting, setSubmitting] = useState(false)
  const API = process.env.NEXT_PUBLIC_API_URL || 'https://api.nuviax.app'

  useEffect(() => {
    fetch(`${API}/v1/recap/current`, { credentials:'include' })
      .then(r => { if(!r.ok) throw new Error(); return r.json() })
      .then(setData)
      .catch(() => router.push('/dashboard'))
  }, [])

  async function startNext() {
    if (!data) return
    setSubmitting(true)
    await fetch(`${API}/v1/goals/${data.goal_id}/recap`, {
      method:'POST', credentials:'include',
      headers:{'Content-Type':'application/json'},
      body: JSON.stringify({ q1, q2, energy }),
    }).catch(()=>{})
    router.push('/dashboard')
  }

  if (!data) return (
    <AppShell>
      <div style={{display:'flex',alignItems:'center',justifyContent:'center',height:200}}>
        <div className="spinner" style={{width:24,height:24,borderTopColor:'var(--l0)'}}/>
      </div>
    </AppShell>
  )

  return (
    <AppShell>
      <div className="page">
        <div className="greet">
          <div>
            <div className="badge-done">
              <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><polyline points="20,6 9,17 4,12"/></svg>
              {data.sprint_name} — Finalizată
            </div>
            <div className="greet-title">Etapă realizată ✦</div>
            <div className="greet-sub">{data.days_active} din {data.days_total} zile menținute</div>
          </div>
          <div className="big-score">
            <div className="big-score-n">{data.score}%</div>
            <div className="big-score-g">{data.grade}</div>
          </div>
        </div>

        {/* 4 mini stats */}
        <div className="stats-4">
          <div className="mini"><div className="mini-v" style={{color:'var(--l2l)'}}>+{data.mrr_delta} RON</div><div className="mini-l">venit lunar</div></div>
          <div className="mini"><div className="mini-v" style={{color:'var(--ul)'}}>{data.days_active}</div><div className="mini-l">zile active</div></div>
          <div className="mini"><div className="mini-v" style={{color:'var(--l5l)'}}>{data.score}%</div><div className="mini-l">progres</div></div>
          <div className="mini"><div className="mini-v">{data.streak}</div><div className="mini-l">zile la rând</div></div>
        </div>

        {/* Questions */}
        <div className="sec-lbl">Câteva întrebări</div>
        <div className="ref-q" style={{cursor:'default'}}>
          <div className="ref-qt">Ce a mers cel mai bine în această etapă?</div>
          <textarea value={q1} onChange={e=>setQ1(e.target.value)}
            placeholder="Scrie gândul tău..."
            style={{width:'100%',marginTop:8,padding:'8px 0',background:'transparent',
              border:'none',outline:'none',color:'var(--ink)',fontFamily:'var(--ff-b)',
              fontSize:13,resize:'none',minHeight:60}}/>
        </div>
        <div className="ref-q" style={{cursor:'default'}}>
          <div className="ref-qt">Care a fost principalul obstacol?</div>
          <textarea value={q2} onChange={e=>setQ2(e.target.value)}
            placeholder="Scrie gândul tău..."
            style={{width:'100%',marginTop:8,padding:'8px 0',background:'transparent',
              border:'none',outline:'none',color:'var(--ink)',fontFamily:'var(--ff-b)',
              fontSize:13,resize:'none',minHeight:60}}/>
        </div>

        {/* Energy 1-10 */}
        <div className="card" style={{marginBottom:14}}>
          <div className="card-lbl" style={{marginBottom:10}}>Nivelul de energie în această etapă (1–10)</div>
          <div style={{display:'flex',gap:4}}>
            {Array.from({length:10},(_,i)=>i+1).map(n=>(
              <div key={n} onClick={()=>setEnergy(n)} style={{
                flex:1,height:22,borderRadius:4,cursor:'pointer',transition:'all .15s',
                background:n<=5?'var(--l2)':'var(--u)',
                opacity:n<=energy?1:.2,
                outline:n===energy?'2px solid rgba(217,119,6,.5)':'none',
              }}/>
            ))}
          </div>
          <div style={{display:'flex',justifyContent:'space-between',marginTop:5,fontFamily:'var(--ff-m)',fontSize:9,color:'var(--ink4)'}}>
            <span>1</span>
            <span style={{color:'var(--ul)',fontWeight:600}}>{energy} / 10 selectat</span>
            <span>10</span>
          </div>
        </div>

        <button className="submit-btn" onClick={startNext} disabled={submitting}>
          {submitting ? <span className="spinner" style={{margin:'0 auto'}}/> : `Pornește ${data.next_sprint_name} →`}
        </button>
      </div>
    </AppShell>
  )
}
