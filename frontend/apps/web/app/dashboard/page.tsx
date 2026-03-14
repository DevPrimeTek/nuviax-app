import { cookies } from 'next/headers'
import { redirect } from 'next/navigation'
import Link from 'next/link'
import AppShell from '@/components/layout/AppShell'
import { dashApi } from '@/lib/api'
import type { Metadata } from 'next'

export const metadata: Metadata = { title: 'Acasă' }

export default async function DashboardPage() {
  const token = cookies().get('nv_access')?.value
  if (!token) redirect('/login')

  let d
  try { d = await dashApi.get(token) } catch { redirect('/login') }

  return (
    <AppShell>
      <div className="page">
        {/* Greeting + streak */}
        <div className="greet">
          <div>
            <div className="greet-title">
              Bună, <em>{d.greeting || 'Alexandru'}.</em>
            </div>
            <div className="greet-sub">{d.date_label}</div>
          </div>
          <div className="streak-chip">
            <svg width="15" height="15" viewBox="0 0 24 24" fill="var(--ul)" stroke="none">
              <path d="M12 2c0 0-7 6-7 11a7 7 0 0014 0C19 8 12 2 12 2z"/>
            </svg>
            <span className="streak-n">{d.streak}</span>
            <span className="streak-l">zile la rând</span>
          </div>
        </div>

        {/* Sprint */}
        <div className="sprint-card">
          <div className="sprint-top">
            <span className="sprint-pulse"/>
            <span className="sprint-name">{d.sprint_name} · {d.sprint_days_left} zile rămase</span>
            <span className="sprint-pct">{d.sprint_pct}%</span>
          </div>
          <div className="sprint-track"><div style={{width:`${d.sprint_pct}%`}}/></div>
        </div>

        {/* 3 stats */}
        <div className="stats-3">
          <div className="stat">
            <div className="stat-ico" style={{background:'var(--l2g)'}}>
              <svg viewBox="0 0 24 24" style={{stroke:'var(--l2l)'}}><polyline points="9,11 12,14 22,4"/><path d="M21 12v7a2 2 0 01-2 2H5a2 2 0 01-2-2V5a2 2 0 012-2h11"/></svg>
            </div>
            <div className="stat-v">{d.tasks_done}<span className="stat-of">/{d.tasks_total}</span></div>
            <div className="stat-l">activități azi</div>
          </div>
          <div className="stat">
            <div className="stat-ico" style={{background:'var(--l0g)'}}>
              <svg viewBox="0 0 24 24" style={{stroke:'var(--l0l)'}}><line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/></svg>
            </div>
            <div className="stat-v" style={{color:'var(--l2l)'}}>{d.milestone_pct}%</div>
            <div className="stat-l">obiectiv etapă</div>
          </div>
          <div className="stat">
            <div className="stat-ico" style={{background:'var(--ug)'}}>
              <svg viewBox="0 0 24 24" style={{stroke:'var(--ul)'}}><line x1="12" y1="1" x2="12" y2="23"/><path d="M17 5H9.5a3.5 3.5 0 000 7h5a3.5 3.5 0 010 7H6"/></svg>
            </div>
            <div className="stat-v" style={{color:'var(--ul)'}}>{d.mrr_current.toLocaleString('ro-RO')}</div>
            <div className="stat-l">RON MRR</div>
          </div>
        </div>

        {/* Goals */}
        <div className="sec-lbl">Obiectivele mele</div>
        {d.active_goals.map(g => (
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
          </Link>
        ))}

        {/* Today's tasks */}
        <div className="sec-lbl">Activitățile de azi</div>
        {d.today_tasks.slice(0,3).map(t => (
          <div key={t.id} className="task-row">
            <div className={`chk${t.done?' done':''}`}>
              {t.done && <svg viewBox="0 0 24 24"><polyline points="20,6 9,17 4,12"/></svg>}
            </div>
            <div style={{flex:1}}>
              <div className={`task-text${t.done?' task-done':''}`}>{t.text}</div>
              <div className="task-meta">
                <span className={`tag ${t.type==='personal'?'tbadge-pers':'tbadge-main'}`}>
                  {t.type==='personal'?'Personal':'Principal'}
                </span>
                <span className="task-time">~{t.estimated_min} min</span>
              </div>
            </div>
          </div>
        ))}
      </div>
    </AppShell>
  )
}
