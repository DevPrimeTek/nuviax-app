import { cookies } from 'next/headers'
import { redirect } from 'next/navigation'
import Link from 'next/link'
import AppShell from '@/components/layout/AppShell'
import { goalsApi, ApiError } from '@/lib/api'
import type { Goal } from '@/lib/api'
import type { Metadata } from 'next'

export const metadata: Metadata = { title: 'Obiective' }

const GOAL_COLORS = ['var(--l0)', 'var(--l2)', 'var(--l5)', 'var(--u)']

export default async function GoalsPage() {
  const token = cookies().get('nv_access')?.value
  if (!token) redirect('/auth/login')

  let allGoals: Goal[] = []
  try {
    const data = await goalsApi.list(token)
    allGoals = [...(data.goals ?? []), ...(data.waiting ?? [])]
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) redirect('/auth/login')
  }

  const goals = Array.isArray(allGoals) ? allGoals : []
  const active = goals.filter(g => g.status === 'ACTIVE')
  const waiting = goals.filter(g => g.status === 'WAITING')
  const other = goals.filter(g => g.status !== 'ACTIVE' && g.status !== 'WAITING')

  return (
    <AppShell>
      <div className="page">
        <div className="greet" style={{marginBottom:18}}>
          <div>
            <div className="greet-title">Obiectivele mele</div>
            <div className="greet-sub">{active.length} active · {waiting.length} în așteptare</div>
          </div>
          <Link href="/onboarding" style={{
            display:'inline-flex',alignItems:'center',gap:6,padding:'8px 14px',
            borderRadius:10,border:'1.5px solid var(--line2)',background:'var(--bg3)',
            color:'var(--ink3)',textDecoration:'none',fontSize:13,fontWeight:500,
            transition:'all .18s',flexShrink:0
          }}>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            Obiectiv nou
          </Link>
        </div>

        {active.length === 0 && waiting.length === 0 ? (
          <div style={{textAlign:'center',padding:'48px 24px',color:'var(--ink3)',fontSize:14}}>
            Nu ai obiective active.{' '}
            <Link href="/onboarding" style={{color:'var(--l0l)'}}>Creează primul →</Link>
          </div>
        ) : active.map((g, i) => (
          <div key={g.id} className="goal-card">
            <div className="goal-top">
              <span className="goal-dot" style={{background: GOAL_COLORS[i % GOAL_COLORS.length]}}/>
              <div style={{flex:1}}>
                <div className="goal-name">{g.name}</div>
                <div className="goal-meta">{g.description || 'Fără descriere'}</div>
              </div>
              <span className="goal-pct" style={{color: GOAL_COLORS[i % GOAL_COLORS.length]}}>
                {g.status}
              </span>
            </div>
          </div>
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

        {other.length > 0 && <>
          <div className="sec-lbl">Altele</div>
          {other.map(g => (
            <div key={g.id} className="goal-card" style={{opacity:.5}}>
              <div className="goal-top">
                <span className="goal-dot" style={{background:'var(--ink4)'}}/>
                <div style={{flex:1}}>
                  <div className="goal-name">{g.name}</div>
                  <div className="goal-meta">{g.status}</div>
                </div>
              </div>
            </div>
          ))}
        </>}
      </div>
    </AppShell>
  )
}
