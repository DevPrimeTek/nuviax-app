'use client'
import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import AppShell from '@/components/layout/AppShell'
import type { TodayData, DailyTask } from '@/lib/api'

export default function TodayPage() {
  const router = useRouter()
  const [data, setData] = useState<TodayData|null>(null)
  const [tasks, setTasks] = useState<DailyTask[]>([])
  const [energy, setEnergy] = useState<'low'|'mid'|'hi'|null>(null)

  useEffect(() => {
    fetch('/api/proxy/today')
      .then(r => { if(!r.ok) throw new Error(r.status.toString()); return r.json() })
      .then((d: TodayData) => {
        setData(d)
        setTasks([...(d.main_tasks||[]), ...(d.personal_tasks||[])])
      })
      .catch((err) => {
        if (err.message === '401') window.location.href = '/auth/login'
      })
  }, [])

  async function toggleTask(id: string) {
    setTasks(ts => ts.map(t => t.id===id ? {...t, completed:!t.completed} : t))
    await fetch(`/api/proxy/today/complete/${id}`, { method:'POST' }).catch(()=>{})
  }

  const done = tasks.filter(t=>t.completed).length
  const total = tasks.length
  const off = 125.7 - (total ? done/total : 0) * 125.7
  const main = tasks.filter(t=>t.type==='MAIN')
  const pers = tasks.filter(t=>t.type==='PERSONAL')

  return (
    <AppShell>
      <div className="page">
        <div className="greet">
          <div>
            <div className="greet-title">Ce faci azi</div>
            <div className="greet-sub">
              {data ? `↳ ${data.goal_name} · Ziua ${data.day_number}` : 'Se încarcă...'}
            </div>
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

        {data?.checkpoint && (
          <div className="card">
            <div className="card-row">
              <span className="card-lbl">Obiectiv etapă</span>
              <span className="card-val" style={{color:'var(--l2l)'}}>{data.checkpoint.progress_pct}%</span>
            </div>
            <div className="card-bar"><div style={{width:`${data.checkpoint.progress_pct}%`,background:'var(--l2)'}}/></div>
            <div className="card-sub">{data.checkpoint.name}</div>
          </div>
        )}

        {data?.streak_days ? (
          <div className="card" style={{padding:'10px 14px'}}>
            <div className="card-row">
              <span className="card-lbl">Streak</span>
              <span className="card-val" style={{color:'var(--ul)'}}>{data.streak_days} zile la rând</span>
            </div>
          </div>
        ) : null}

        {main.length > 0 && (
          <>
            <div className="sec-lbl">Activitățile principale</div>
            {main.map(t => (
              <div key={t.id} className="task-row" onClick={()=>toggleTask(t.id)}>
                <div className={`chk${t.completed?' done':''}`}>
                  {t.completed && <svg viewBox="0 0 24 24"><polyline points="20,6 9,17 4,12"/></svg>}
                </div>
                <div style={{flex:1}}>
                  <div className={`task-text${t.completed?' task-done':''}`}>{t.text}</div>
                  <div className="task-meta">
                    <span className="tag tbadge-main">Principal</span>
                  </div>
                </div>
              </div>
            ))}
          </>
        )}

        {pers.length > 0 && (
          <>
            <div className="sec-lbl">Activități personale</div>
            {pers.map(t => (
              <div key={t.id} className="task-row" style={{opacity:.7}} onClick={()=>toggleTask(t.id)}>
                <div className={`chk${t.completed?' done':''}`} style={{borderColor:'rgba(37,99,235,.4)'}}>
                  {t.completed && <svg viewBox="0 0 24 24"><polyline points="20,6 9,17 4,12"/></svg>}
                </div>
                <div style={{flex:1}}>
                  <div className={`task-text${t.completed?' task-done':''}`}>{t.text}</div>
                  <div className="task-meta">
                    <span className="tag tbadge-pers">Personal</span>
                  </div>
                </div>
              </div>
            ))}
          </>
        )}

        {!data && (
          <div style={{display:'flex',alignItems:'center',justifyContent:'center',height:120}}>
            <div className="spinner" style={{width:24,height:24,borderTopColor:'var(--l0)'}}/>
          </div>
        )}

        {data && total === 0 && (
          <div className="card" style={{textAlign:'center',padding:24}}>
            <div style={{color:'var(--ink3)',fontSize:14}}>Nu ai activități programate pentru azi.</div>
          </div>
        )}

        {/* Energy */}
        <div className="card" style={{marginTop:4}}>
          <div className="card-lbl" style={{marginBottom:11}}>Cum te simți azi?</div>
          <div style={{display:'flex',gap:8}}>
            {([
              {k:'low' as const, l:'Obosit'},
              {k:'mid' as const, l:'Normal'},
              {k:'hi'  as const, l:'Energic'},
            ] as const).map(e=>(
              <button key={e.k} onClick={()=>setEnergy(e.k)}
                className={`e-btn${energy===e.k?' sel':''}`}>
                {e.l}
              </button>
            ))}
          </div>
        </div>
      </div>
    </AppShell>
  )
}
