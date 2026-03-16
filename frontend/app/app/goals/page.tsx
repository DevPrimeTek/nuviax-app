import { cookies } from 'next/headers'
import { redirect } from 'next/navigation'
import Link from 'next/link'
import AppShell from '@/components/layout/AppShell'
import { goalsApi } from '@/lib/api'
import type { Metadata } from 'next'

export const metadata: Metadata = { title: 'Obiective' }

interface Goal {
  id: string
  name: string
  status: string
  color?: string
  progress_pct: number
  current_sprint: number
  total_sprints: number
  sprint_days_left: number
  overall_score: number
}

interface GoalsData {
  goals: Goal[]
  waiting: Goal[]
}

export default async function GoalsPage() {
  const token = cookies().get('nv_access')?.value
  if (!token) redirect('/auth/login')
  
  let data: GoalsData = { goals: [], waiting: [] }
  try { 
    data = await goalsApi.list(token)
  } catch { 
    redirect('/auth/login') 
  }

  const { goals, waiting } = data

  return (
    <AppShell>
      <div className="page">
        <div className="greet" style={{marginBottom:18}}>
          <div>
            <div className="greet-title">Obiectivele mele</div>
            <div className="greet-sub">{goals.length} active · {waiting.length} în așteptare</div>
          </div>
          <Link href="/goals/new" style={{
            display:'inline-flex',alignItems:'center',gap:6,padding:'8px 14px',
            borderRadius:10,border:'1.5px solid var(--line2)',background:'var(--bg3)',
            color:'var(--ink3)',textDecoration:'none',fontSize:13,fontWeight:500,
            transition:'all .18s',flexShrink:0
          }}>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            Obiectiv nou
          </Link>
        </div>

        {goals.length === 0 ? (
          <div style={{textAlign:'center',padding:'48px 24px',color:'var(--ink3)',fontSize:14}}>
            Nu ai obiective active.{' '}
            <Link href="/goals/new" style={{color:'var(--l0l)'}}>Creează primul →</Link>
          </div>
        ) : goals.map(g => (
          <Link key={g.id} href={`/goals/${g.id}`} className="goal-card">
            <div className="goal-top">
              <span className="goal-dot" style={{background:g.color||'var(--l5)'}}/>
              <div style={{flex:1}}>
                <div className="goal-name">{g.name}</div>
                <div className="goal-meta">Etapa {g.current_sprint}/{g.total_sprints} · {g.sprint_days_left} zile rămase</div>
              </div>
              <span className="goal-pct" style={{color:g.color||'var(--l5l)'}}>{g.progress_pct}%</span>
            </div>
            <div className="goal-bar"><div style={{width:`${g.progress_pct}%`,background:g.color||'var(--l5)'}}/></div>
            <div style={{display:'flex',justifyContent:'space-between',marginTop:8}}>
              <span className="tag" style={{color:g.color||'var(--l5l)',background:`rgba(13,148,136,.12)`,border:`1px solid rgba(13,148,136,.22)`}}>
                {g.status==='active'?'Activ':'Pauză'}
              </span>
              <span style={{fontFamily:'var(--ff-m)',fontSize:10,color:'var(--ink4)'}}>
                Scor: {g.overall_score}%
              </span>
            </div>
          </Link>
        ))}

        {waiting.length > 0 && <>
          <div className="sec-lbl">Listă de așteptare</div>
          {waiting.map(g => (
            <div key={g.id} className="goal-card" style={{opacity:.6}}>
              <div className="goal-top">
                <span className="goal-dot" style={{background:'var(--ink4)'}}/>
                <div style={{flex:1}}>
                  <div className="goal-name">{g.name}</div>
                  <div className="goal-meta">Așteaptă un slot activ</div>
                </div>
              </div>
            </div>
          ))}
        </>}
      </div>
    </AppShell>
  )
}