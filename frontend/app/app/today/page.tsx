'use client'
import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import AppShell from '@/components/layout/AppShell'
import type { TodayData, DailyTask } from '@/lib/api'
import { useTranslation } from '@/lib/i18n'

export default function TodayPage() {
  const router = useRouter()
  const { t } = useTranslation()
  const [data, setData] = useState<TodayData|null>(null)
  const [tasks, setTasks] = useState<DailyTask[]>([])
  const [energy, setEnergy] = useState<'low'|'mid'|'hi'|null>(null)
  // B-6: personal task add state
  const [newTask, setNewTask] = useState('')
  const [addingTask, setAddingTask] = useState(false)

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
    const res = await fetch(`/api/proxy/today/complete/${id}`, { method:'POST' }).catch(()=>null)
    if (res?.ok) {
      setTasks(ts => ts.map(tk => tk.id===id ? {...tk, completed:!tk.completed} : tk))
    }
  }

  // B-6: add personal task
  async function addPersonalTask() {
    const text = newTask.trim()
    if (!text || addingTask) return
    setAddingTask(true)
    try {
      const res = await fetch('/api/proxy/today/personal', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ text }),
      })
      if (res.ok) {
        const task = await res.json()
        setTasks(ts => [...ts, task])
        setNewTask('')
      }
    } catch { /* ignore */ }
    setAddingTask(false)
  }

  const done = tasks.filter(tk=>tk.completed).length
  const total = tasks.length
  const off = 125.7 - (total ? done/total : 0) * 125.7
  const main = tasks.filter(tk=>tk.type==='MAIN')
  const pers = tasks.filter(tk=>tk.type==='PERSONAL')

  return (
    <AppShell>
      <div className="page">
        <div className="greet">
          <div>
            <div className="greet-title">{t('today.title')}</div>
            <div className="greet-sub">
              {data ? `↳ ${data.goal_name} · ${t('today.day')} ${data.day_number}` : t('today.loading')}
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
              <div className="ring-l">{t('today.tasks_done_label')}</div>
            </div>
          </div>
        </div>

        {data?.checkpoint && (
          <div className="card">
            <div className="card-row">
              <span className="card-lbl">{t('today.checkpoint_label')}</span>
              <span className="card-val" style={{color:'var(--l2l)'}}>{data.checkpoint.progress_pct}%</span>
            </div>
            <div className="card-bar"><div style={{width:`${data.checkpoint.progress_pct}%`,background:'var(--l2)'}}/></div>
            <div className="card-sub">{data.checkpoint.name}</div>
          </div>
        )}

        {data?.streak_days ? (
          <div className="card" style={{padding:'10px 14px'}}>
            <div className="card-row">
              <span className="card-lbl">{t('today.streak_label')}</span>
              <span className="card-val" style={{color:'var(--ul)'}}>{data.streak_days} {t('today.streak_days')}</span>
            </div>
          </div>
        ) : null}

        {main.length > 0 && (
          <>
            <div className="sec-lbl">{t('today.main_tasks_section')}</div>
            {main.map(task => (
              <div key={task.id} className="task-row" onClick={()=>toggleTask(task.id)}>
                <div className={`chk${task.completed?' done':''}`}>
                  {task.completed && <svg viewBox="0 0 24 24"><polyline points="20,6 9,17 4,12"/></svg>}
                </div>
                <div style={{flex:1}}>
                  <div className={`task-text${task.completed?' task-done':''}`}>{task.text}</div>
                  <div className="task-meta">
                    <span className="tag tbadge-main">{t('today.tag_main')}</span>
                  </div>
                </div>
              </div>
            ))}
          </>
        )}

        {/* Personal tasks section — always shown so user can add tasks (B-6) */}
        <div className="sec-lbl">{t('today.personal_tasks_section')}</div>
        {pers.map(task => (
          <div key={task.id} className="task-row" style={{opacity:.7}} onClick={()=>toggleTask(task.id)}>
            <div className={`chk${task.completed?' done':''}`} style={{borderColor:'rgba(37,99,235,.4)'}}>
              {task.completed && <svg viewBox="0 0 24 24"><polyline points="20,6 9,17 4,12"/></svg>}
            </div>
            <div style={{flex:1}}>
              <div className={`task-text${task.completed?' task-done':''}`}>{task.text}</div>
              <div className="task-meta">
                <span className="tag tbadge-pers">{t('today.tag_personal')}</span>
              </div>
            </div>
          </div>
        ))}
        {/* B-6: add personal task input (max 2 per day, enforced server-side) */}
        {pers.length < 2 && (
          <div style={{display:'flex',gap:8,alignItems:'center',padding:'8px 0'}}>
            <input
              value={newTask}
              onChange={e=>setNewTask(e.target.value)}
              onKeyDown={e=>e.key==='Enter'&&addPersonalTask()}
              placeholder={t('today.add_personal_placeholder')}
              maxLength={120}
              style={{flex:1,background:'var(--bg3)',border:'1px solid var(--line)',borderRadius:10,
                padding:'9px 13px',color:'var(--ink)',fontSize:13,outline:'none',fontFamily:'var(--ff-b)'}}
            />
            <button
              onClick={addPersonalTask}
              disabled={!newTask.trim()||addingTask}
              style={{background:'var(--l2)',color:'white',border:'none',borderRadius:10,
                padding:'9px 14px',cursor:'pointer',fontSize:13,opacity:(!newTask.trim()||addingTask)?0.5:1}}>
              {addingTask ? t('today.adding_btn') : t('today.add_btn')}
            </button>
          </div>
        )}

        {!data && (
          <div style={{display:'flex',alignItems:'center',justifyContent:'center',height:120}}>
            <div className="spinner" style={{width:24,height:24,borderTopColor:'var(--l0)'}}/>
          </div>
        )}

        {data && total === 0 && (
          <div className="card" style={{textAlign:'center',padding:24}}>
            <div style={{color:'var(--ink3)',fontSize:14}}>{t('today.no_tasks')}</div>
          </div>
        )}

        {/* Energy */}
        <div className="card" style={{marginTop:4}}>
          <div className="card-lbl" style={{marginBottom:4}}>{t('today.energy_title')}</div>
          <div style={{fontSize:12,color:'var(--ink3)',marginBottom:11}}>{t('today.energy_subtitle')}</div>
          <div style={{display:'flex',gap:7}}>
            {([
              {k:'low' as const, cls:'sel-low', icon:<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" strokeWidth="1.9" strokeLinecap="round" strokeLinejoin="round"><path d="M18 8h1a4 4 0 010 8h-1"/><path d="M2 8h16v9a4 4 0 01-4 4H6a4 4 0 01-4-4V8z"/><line x1="6" y1="1" x2="6" y2="4"/><line x1="10" y1="1" x2="10" y2="4"/><line x1="14" y1="1" x2="14" y2="4"/></svg>},
              {k:'mid' as const, cls:'sel-mid', icon:<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" strokeWidth="1.9" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="10"/><path d="M8 14s1.5 2 4 2 4-2 4-2"/><line x1="9" y1="9" x2="9.01" y2="9"/><line x1="15" y1="9" x2="15.01" y2="9"/></svg>},
              {k:'hi'  as const, cls:'sel-hi', icon:<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" strokeWidth="1.9" strokeLinecap="round" strokeLinejoin="round"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>},
            ] as const).map(e=>(
              <button key={e.k} onClick={()=>{
                setEnergy(e.k)
                // B-5 fix: correct endpoint + level mapping (mid→normal, hi→high handled server-side)
                fetch('/api/proxy/context/energy',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({level:e.k})}).catch(()=>{})
              }}
                className={`e-btn${energy===e.k?' '+e.cls:''}`}>
                <div className="e-btn-icon">{e.icon}</div>
                {e.k === 'low' ? t('today.energy_low') : e.k === 'mid' ? t('today.energy_mid') : t('today.energy_hi')}
              </button>
            ))}
          </div>
          {energy && (
            <div style={{marginTop:11,padding:'12px 14px',borderRadius:12,background:'var(--bg3)',border:'1px solid var(--line)'}}>
              <div style={{display:'flex',alignItems:'center',gap:8,marginBottom:6}}>
                <div style={{width:20,height:20,borderRadius:6,background:'var(--bg4)',display:'flex',alignItems:'center',justifyContent:'center',flexShrink:0}}>
                  <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="var(--l5l)" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 01-2 2H5a2 2 0 01-2-2V5a2 2 0 012-2h11"/></svg>
                </div>
                <span style={{fontSize:12,color:'var(--ink3)'}}>
                  {energy==='low' ? t('today.energy_effect_low_intensity') : energy==='hi' ? t('today.energy_effect_hi_intensity') : t('today.energy_effect_mid_intensity')}
                </span>
              </div>
              <div style={{display:'flex',alignItems:'center',gap:8}}>
                <div style={{width:20,height:20,borderRadius:6,background:'var(--bg4)',display:'flex',alignItems:'center',justifyContent:'center',flexShrink:0}}>
                  <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="var(--l5l)" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/></svg>
                </div>
                <span style={{fontSize:12,color:'var(--ink3)'}}>
                  {energy==='low' ? t('today.energy_effect_low_progress') : energy==='hi' ? t('today.energy_effect_hi_progress') : t('today.energy_effect_mid_progress')}
                </span>
              </div>
            </div>
          )}
        </div>
      </div>
    </AppShell>
  )
}
