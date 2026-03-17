import { cookies } from 'next/headers'
import { redirect } from 'next/navigation'
import Link from 'next/link'
import AppShell from '@/components/layout/AppShell'
import { goalsApi, ApiError } from '@/lib/api'
import type { Goal } from '@/lib/api'
import type { Metadata } from 'next'

export const metadata: Metadata = { title: 'Obiective' }

const GOAL_COLORS = ['var(--l0)', 'var(--l2)', 'var(--l5)', 'var(--l3)']
const GOAL_COLORS_G = ['var(--l0g)', 'var(--l2g)', 'var(--l5g)', 'var(--l3g)']

function GoalRing({ pct, color }: { pct: number; color: string }) {
  const r = 22
  const circ = 2 * Math.PI * r
  const offset = circ - (pct / 100) * circ
  return (
    <div style={{width:54,height:54,position:'relative',flexShrink:0}}>
      <svg width="54" height="54" viewBox="0 0 54 54" style={{transform:'rotate(-90deg)'}}>
        <circle fill="none" stroke="var(--line)" strokeWidth="4.5" cx="27" cy="27" r={r}/>
        <circle fill="none" stroke={color} strokeWidth="4.5" strokeLinecap="round"
          cx="27" cy="27" r={r} strokeDasharray={circ.toFixed(1)} strokeDashoffset={offset.toFixed(1)}/>
      </svg>
      <div style={{position:'absolute',inset:0,display:'flex',alignItems:'center',justifyContent:'center',flexDirection:'column'}}>
        <span style={{fontFamily:'var(--ff-d)',fontSize:12,fontWeight:800,lineHeight:1,color}}>{pct}%</span>
        <span style={{fontFamily:'var(--ff-m)',fontSize:7,color:'var(--ink3)',marginTop:1}}>progres</span>
      </div>
    </div>
  )
}

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
            display:'inline-flex',alignItems:'center',gap:6,padding:'9px 15px',
            borderRadius:12,background:'var(--l0g)',border:'1.5px solid var(--l0b)',
            color:'var(--l0l)',textDecoration:'none',fontSize:13,fontWeight:600,
            transition:'all .18s',flexShrink:0,
          }}>
            <div style={{width:18,height:18,borderRadius:5,background:'rgba(124,58,237,.25)',display:'flex',alignItems:'center',justifyContent:'center'}}>
              <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="var(--l0l)" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            </div>
            Adaugă Obiectiv
          </Link>
        </div>

        {active.length === 0 && waiting.length === 0 ? (
          <div style={{textAlign:'center',padding:'48px 24px',color:'var(--ink3)',fontSize:14}}>
            Nu ai obiective active.{' '}
            <Link href="/onboarding" style={{color:'var(--l0l)'}}>Creează primul →</Link>
          </div>
        ) : active.map((g, i) => {
          const color = GOAL_COLORS[i % GOAL_COLORS.length]
          const colorG = GOAL_COLORS_G[i % GOAL_COLORS_G.length]
          const pct = Math.round((g as any).progress_pct ?? 0)
          const daysLeft = g.end_date ? Math.max(0, Math.round((new Date(g.end_date).getTime() - Date.now()) / 86400000)) : 0
          return (
            <div key={g.id} style={{
              background:'var(--bg2)',border:'1.5px solid var(--line)',borderRadius:20,
              padding:'18px 20px',position:'relative',overflow:'hidden',
              cursor:'pointer',transition:'border-color .2s,transform .15s',marginBottom:11,
            }}>
              {/* Color bar at top */}
              <div style={{position:'absolute',top:0,left:0,right:0,height:3,borderRadius:'20px 20px 0 0',background:color}}/>
              <div style={{display:'flex',alignItems:'flex-start',justifyContent:'space-between',gap:12,marginBottom:12}}>
                <div style={{flex:1}}>
                  <span style={{fontFamily:'var(--ff-m)',fontSize:9.5,fontWeight:600,letterSpacing:'.05em',
                    textTransform:'uppercase',padding:'3px 8px',borderRadius:7,
                    color,background:colorG,border:`1px solid ${color}25`,display:'inline-block',marginBottom:8}}>
                    Activ
                  </span>
                  <div style={{fontFamily:'var(--ff-d)',fontSize:15,fontWeight:700,color:'var(--ink)',lineHeight:1.3}}>
                    {g.name}
                  </div>
                </div>
                <GoalRing pct={pct} color={color}/>
              </div>
              <div style={{width:'100%',height:6,borderRadius:4,background:'var(--line)',overflow:'hidden',margin:'0 0 8px'}}>
                <div style={{height:'100%',borderRadius:4,background:color,width:`${pct}%`,transition:'width 1.2s cubic-bezier(.4,0,.2,1)'}}/>
              </div>
              <div style={{display:'flex',alignItems:'center',justifyContent:'space-between'}}>
                <span style={{fontFamily:'var(--ff-m)',fontSize:10,color:'var(--ink3)',fontWeight:500}}>
                  <span style={{color:'var(--ink2)'}}>Sprint 1</span> · {daysLeft} zile rămase
                </span>
              </div>
            </div>
          )
        })}

        {waiting.length > 0 && (
          <>
            <div style={{padding:'20px 0 9px',fontFamily:'var(--ff-m)',fontSize:10,color:'var(--ink3)',
              letterSpacing:'.15em',textTransform:'uppercase',fontWeight:600}}>
              Lista de așteptare · {waiting.length} {waiting.length === 1 ? 'obiectiv' : 'obiective'}
            </div>
            {waiting.map(g => (
              <div key={g.id} style={{
                display:'flex',alignItems:'center',gap:13,padding:'14px 16px',
                background:'var(--bg3)',border:'1.5px solid var(--line)',borderRadius:16,
                cursor:'pointer',transition:'border-color .2s',marginBottom:8,opacity:.75,
              }}>
                <div style={{width:38,height:38,borderRadius:11,background:'rgba(217,119,6,.12)',
                  display:'flex',alignItems:'center',justifyContent:'center',flexShrink:0}}>
                  <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="var(--ul)" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
                    <circle cx="12" cy="12" r="10"/><polyline points="12,6 12,12 16,14"/>
                  </svg>
                </div>
                <div style={{flex:1}}>
                  <div style={{fontFamily:'var(--ff-d)',fontSize:14,fontWeight:600,color:'var(--ink)',lineHeight:1.3,marginBottom:3}}>{g.name}</div>
                  <div style={{fontFamily:'var(--ff-m)',fontSize:10,color:'var(--ink4)'}}>Așteaptă un slot activ</div>
                </div>
                <span style={{fontFamily:'var(--ff-m)',fontSize:9.5,fontWeight:600,letterSpacing:'.05em',
                  textTransform:'uppercase',padding:'3px 8px',borderRadius:7,
                  color:'var(--ul)',background:'rgba(217,119,6,.1)',border:'1px solid rgba(217,119,6,.2)'}}>
                  Așteptare
                </span>
              </div>
            ))}
          </>
        )}

        {other.length > 0 && (
          <>
            <div style={{padding:'20px 0 9px',fontFamily:'var(--ff-m)',fontSize:10,color:'var(--ink3)',
              letterSpacing:'.15em',textTransform:'uppercase',fontWeight:600}}>Altele</div>
            {other.map(g => (
              <div key={g.id} style={{
                background:'var(--bg3)',border:'1.5px solid var(--line)',borderRadius:16,
                padding:'14px 16px',marginBottom:8,opacity:.5,
              }}>
                <div style={{fontFamily:'var(--ff-d)',fontSize:14,fontWeight:600,color:'var(--ink)'}}>{g.name}</div>
                <div style={{fontFamily:'var(--ff-m)',fontSize:10,color:'var(--ink4)',marginTop:3}}>{g.status}</div>
              </div>
            ))}
          </>
        )}
      </div>
    </AppShell>
  )
}
