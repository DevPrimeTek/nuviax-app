'use client'
import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import AppShell from '@/components/layout/AppShell'
import type { Task } from '@/lib/api'

interface TodayData {
  goal_name: string; sprint_label: string
  milestone: string; milestone_pct: number
  tasks: Task[]
}

export default function TodayPage() {
  const router = useRouter()
  const [data, setData] = useState<TodayData|null>(null)
  const [tasks, setTasks] = useState<Task[]>([])
  const [energy, setEnergy] = useState<'low'|'mid'|'hi'|null>(null)
  const API = process.env.NEXT_PUBLIC_API_URL || 'https://api.nuviax.app'

  useEffect(() => {
    fetch(`${API}/v1/tasks/today`, { credentials:'include' })
      .then(r => { if(!r.ok) throw new Error(); return r.json() })
      .then(d => { setData(d); setTasks(d.tasks||[]) })
      .catch(() => router.push('/login'))
  }, [])

  async function toggleTask(id: string) {
    setTasks(ts => ts.map(t => t.id===id ? {...t, done:!t.done} : t))
    await fetch(`${API}/v1/tasks/${id}/complete`, { method:'POST', credentials:'include' }).catch(()=>{})
  }

  const done = tasks.filter(t=>t.done).length
  const total = tasks.length
  const off = 125.7 - (total ? done/total : 0) * 125.7
  const main = tasks.filter(t=>t.type!=='personal')
  const pers = tasks.filter(t=>t.type==='personal')

  return (
    <AppShell>
      <div className="page">
        <div className="greet">
          <div>
            <div className="greet-title">Ce faci azi</div>
            <div className="greet-sub">↳ {data?.goal_name} · {data?.sprint_label}</div>
          </div>
          <div className="ring-box">
            <svg width="42" height="42" viewBox="0 0 52 52" style={{transform:'rotate(-90deg)'}}>
              <circle fill="none" stroke="var(--line)" strokeWidth="5" cx="26" cy="26" r="20"/>
              <circle fill="none" stroke="var(--l2)" strokeWidth="5" strokeLinecap="round"
                cx="26" cy="26" r="20" strokeDasharray="125.7" strokeDashoffset={off}
                style={{transition:'stroke-dashoffset .5s ease'}}/>
            </svg>
            <div>
              <div className="ring-v">{done}<span>/{total}</span></div>
              <div className="ring-l">activități</div>
            </div>
          </div>
        </div>

        {data && (
          <div className="card">
            <div className="card-row">
              <span className="card-lbl">Obiectiv etapă</span>
              <span className="card-val" style={{color:'var(--l2l)'}}>{data.milestone_pct}%</span>
            </div>
            <div className="card-bar"><div style={{width:`${data.milestone_pct}%`,background:'var(--l2)'}}/></div>
            <div className="card-sub">{data.milestone}</div>
          </div>
        )}

        <div className="sec-lbl">Activitățile principale</div>
        {main.map(t => (
          <div key={t.id} className="task-row" onClick={()=>toggleTask(t.id)}>
            <div className={`chk${t.done?' done':''}`}>
              {t.done && <svg viewBox="0 0 24 24"><polyline points="20,6 9,17 4,12"/></svg>}
            </div>
            <div style={{flex:1}}>
              <div className={`task-text${t.done?' task-done':''}`}>{t.text}</div>
              <div className="task-meta">
                <span className="tag tbadge-main">Principal</span>
                <span className="task-time">~{t.estimated_min} min</span>
              </div>
            </div>
          </div>
        ))}

        {pers.length>0 && <>
          <div className="sec-lbl">Activități personale</div>
          {pers.map(t => (
            <div key={t.id} className="task-row" style={{opacity:.7}} onClick={()=>toggleTask(t.id)}>
              <div className={`chk${t.done?' done':''}`} style={{borderColor:'rgba(37,99,235,.4)'}}>
                {t.done && <svg viewBox="0 0 24 24"><polyline points="20,6 9,17 4,12"/></svg>}
              </div>
              <div style={{flex:1}}>
                <div className={`task-text${t.done?' task-done':''}`}>{t.text}</div>
                <div className="task-meta">
                  <span className="tag tbadge-pers">Personal</span>
                  <span className="task-time">~{t.estimated_min} min</span>
                </div>
              </div>
            </div>
          ))}
        </>}

        {/* Energy */}
        <div className="card" style={{marginTop:4}}>
          <div className="card-lbl" style={{marginBottom:11}}>Cum te simți azi?</div>
          <div style={{display:'flex',gap:8}}>
            {[
              {k:'low' as const, l:'Obosit',  icon:<svg viewBox="0 0 24 24"><path d="M18 8h1a4 4 0 010 8h-1"/><path d="M2 8h16v9a4 4 0 01-4 4H6a4 4 0 01-4-4V8z"/><line x1="6" y1="1" x2="6" y2="4"/><line x1="10" y1="1" x2="10" y2="4"/><line x1="14" y1="1" x2="14" y2="4"/></svg>},
              {k:'mid' as const, l:'Normal',  icon:<svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><path d="M8 14s1.5 2 4 2 4-2 4-2"/><line x1="9" y1="9" x2="9.01" y2="9"/><line x1="15" y1="9" x2="15.01" y2="9"/></svg>},
              {k:'hi'  as const, l:'Energic', icon:<svg viewBox="0 0 24 24"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>},
            ].map(e=>(
              <button key={e.k} onClick={()=>setEnergy(e.k)}
                className={`e-btn${energy===e.k?' sel':''}`}>
                {e.icon}{e.l}
              </button>
            ))}
          </div>
        </div>
      </div>
    </AppShell>
  )
}
