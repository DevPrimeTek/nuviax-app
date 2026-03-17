import { cookies } from 'next/headers'
import { redirect } from 'next/navigation'
import Link from 'next/link'
import AppShell from '@/components/layout/AppShell'
import DashboardClientLayer from '@/components/DashboardClientLayer'
import SRMWarning from '@/components/SRMWarning'
import { dashApi, ApiError } from '@/lib/api'
import type { Metadata } from 'next'

export const metadata: Metadata = { title: 'Acasă' }

const GOAL_COLORS = ['var(--l0)', 'var(--l2)', 'var(--l5)', 'var(--u)']

function formatDateLabel() {
  return new Date().toLocaleDateString('ro-RO', { weekday: 'long', day: 'numeric', month: 'long' })
}

export default async function DashboardPage() {
  const token = cookies().get('nv_access')?.value
  if (!token) redirect('/auth/login')

  let d: Awaited<ReturnType<typeof dashApi.get>>
  try {
    d = await dashApi.get(token)
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) redirect('/auth/login')
    // For non-auth errors, redirect to onboarding as fallback
    redirect('/onboarding')
  }

  // Redirecționează utilizatorii noi la onboarding
  const hasGoals = (d.active_goals?.length ?? 0) > 0 || (d.waiting_goals?.length ?? 0) > 0
  if (!hasGoals) redirect('/onboarding')

  const userName = d.user?.full_name || 'utilizator'
  const activeGoal = d.active_goals?.[0]
  const sprintDaysLeft = activeGoal?.days_left ?? 0
  const sprintPct = activeGoal ? Math.round((activeGoal.progress_score ?? 0) * 100) : 0
  const sprintName = activeGoal ? `Sprint ${activeGoal.sprint_number}` : '—'
  const todayCount = d.today_tasks_count ?? 0

  return (
    <AppShell userName={userName}>
      <div className="page">
        {/* Greeting */}
        <div className="greet">
          <div>
            <div className="greet-title">
              Bună, <em>{userName}.</em>
            </div>
            <div className="greet-sub">{formatDateLabel()}</div>
          </div>
          <div className="streak-chip">
            <svg width="15" height="15" viewBox="0 0 24 24" fill="var(--ul)" stroke="none">
              <path d="M12 2c0 0-7 6-7 11a7 7 0 0014 0C19 8 12 2 12 2z"/>
            </svg>
            <span className="streak-n">{todayCount}</span>
            <span className="streak-l">activități azi</span>
          </div>
        </div>

        {/* Sprint activ */}
        {activeGoal && (
          <div className="sprint-card">
            <div className="sprint-top">
              <span className="sprint-pulse"/>
              <span className="sprint-name">{sprintName} · {sprintDaysLeft} zile rămase</span>
              <span className="sprint-pct">{sprintPct}%</span>
            </div>
            <div className="sprint-track"><div style={{width:`${sprintPct}%`}}/></div>
          </div>
        )}

        {/* 3 stats */}
        <div className="stats-3">
          <div className="stat">
            <div className="stat-ico" style={{background:'var(--l2g)'}}>
              <svg viewBox="0 0 24 24" style={{stroke:'var(--l2l)'}}><polyline points="9,11 12,14 22,4"/><path d="M21 12v7a2 2 0 01-2 2H5a2 2 0 01-2-2V5a2 2 0 012-2h11"/></svg>
            </div>
            <div className="stat-v">0<span className="stat-of">/{todayCount}</span></div>
            <div className="stat-l">activități azi</div>
          </div>
          <div className="stat">
            <div className="stat-ico" style={{background:'var(--l0g)'}}>
              <svg viewBox="0 0 24 24" style={{stroke:'var(--l0l)'}}><line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/></svg>
            </div>
            <div className="stat-v" style={{color:'var(--l2l)'}}>{sprintPct}%</div>
            <div className="stat-l">progres sprint</div>
          </div>
          <div className="stat">
            <div className="stat-ico" style={{background:'rgba(167,139,250,0.12)'}}>
              <svg viewBox="0 0 24 24" style={{stroke:'var(--l5l)'}}><circle cx="12" cy="12" r="10"/><path d="M12 8v4l3 3"/></svg>
            </div>
            <div className="stat-v" style={{color:'var(--l5l)'}}>{sprintDaysLeft}</div>
            <div className="stat-l">zile rămase</div>
          </div>
        </div>

        {/* GO-uri active */}
        <div className="sec-lbl">GO-urile mele</div>
        {(d.active_goals ?? []).map((g, i) => {
          const pct = Math.round((g.progress_score ?? 0) * 100)
          const color = GOAL_COLORS[i % GOAL_COLORS.length]
          return (
            <div key={g.id}>
              <SRMWarning goalId={g.id} />
            <Link href={`/goals/${g.id}`} className="goal-card">
              <div className="goal-top">
                <span className="goal-dot" style={{background: color}}/>
                <div style={{flex:1}}>
                  <div className="goal-name">{g.name}</div>
                  <div className="goal-meta">
                    Sprint {g.sprint_number}/{g.total_sprints} · {g.days_left} zile rămase
                  </div>
                </div>
                <span className="goal-pct" style={{color}}>{pct}%</span>
              </div>
              <div className="goal-bar"><div style={{width:`${pct}%`, background: color}}/></div>
            </Link>
            </div>
          )
        })}

        {/* GO-uri în așteptare */}
        {(d.waiting_goals ?? []).length > 0 && (
          <>
            <div className="sec-lbl" style={{marginTop:12}}>În așteptare</div>
            {(d.waiting_goals ?? []).map((g, i) => (
              <Link key={g.id} href={`/goals/${g.id}`} className="goal-card" style={{opacity:0.6}}>
                <div className="goal-top">
                  <span className="goal-dot" style={{background:'var(--ink4)'}}/>
                  <div style={{flex:1}}>
                    <div className="goal-name">{g.name}</div>
                    <div className="goal-meta">În așteptare</div>
                  </div>
                </div>
              </Link>
            ))}
          </>
        )}

        {/* Link la Today */}
        <Link href="/today" className="sprint-card" style={{textDecoration:'none', marginTop:8, display:'block'}}>
          <div className="sprint-top">
            <span className="sprint-pulse"/>
            <span className="sprint-name">Vezi activitățile de azi →</span>
            <span className="sprint-pct" style={{color:'var(--l2l)'}}>{todayCount}</span>
          </div>
        </Link>
      </div>
      <DashboardClientLayer />
    </AppShell>
  )
}
