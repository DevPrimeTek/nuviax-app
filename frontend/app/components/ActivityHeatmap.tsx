'use client'
import { useEffect, useState } from 'react'

type ActivityDay = {
  date: string          // "2026-01-15"
  score: number         // 0–1
  tasks_completed: number
}

type Props = {
  /** If not provided, fetches from /api/proxy/profile/activity */
  data?: ActivityDay[]
}

function getColor(score: number): string {
  if (score <= 0)   return 'var(--line)'
  if (score < 0.25) return 'rgba(13,148,136,.25)'
  if (score < 0.5)  return 'rgba(13,148,136,.45)'
  if (score < 0.75) return 'rgba(13,148,136,.65)'
  return 'var(--l5)'
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr + 'T00:00:00')
  return d.toLocaleDateString('ro-RO', { day: 'numeric', month: 'short', year: 'numeric' })
}

/** Build a 52-week grid (Mon–Sun columns) ending today */
function buildGrid(): { date: string; dayOfWeek: number }[][] {
  const today = new Date()
  today.setHours(0, 0, 0, 0)

  // Start from 364 days ago, adjusted to the nearest Monday
  const start = new Date(today)
  start.setDate(start.getDate() - 364)
  // Move to Monday
  const dow = start.getDay() // 0=Sun
  const daysBack = dow === 0 ? 6 : dow - 1
  start.setDate(start.getDate() - daysBack)

  const weeks: { date: string; dayOfWeek: number }[][] = []
  let current = new Date(start)

  while (current <= today) {
    const week: { date: string; dayOfWeek: number }[] = []
    for (let d = 0; d < 7; d++) {
      const iso = current.toISOString().split('T')[0]
      week.push({ date: iso, dayOfWeek: d })
      current = new Date(current)
      current.setDate(current.getDate() + 1)
    }
    weeks.push(week)
  }
  return weeks
}

export default function ActivityHeatmap({ data: dataProp }: Props) {
  const [activity, setActivity] = useState<Map<string, ActivityDay>>(new Map())
  const [loading, setLoading] = useState(!dataProp)
  const [tooltip, setTooltip] = useState<{ date: string; score: number; tasks: number } | null>(null)

  useEffect(() => {
    if (dataProp) {
      const map = new Map<string, ActivityDay>()
      dataProp.forEach(d => map.set(d.date, d))
      setActivity(map)
      return
    }
    fetch('/api/proxy/profile/activity')
      .then(r => r.ok ? r.json() : { activity: [] })
      .then(d => {
        const map = new Map<string, ActivityDay>()
        ;(d.activity || []).forEach((item: ActivityDay) => map.set(item.date, item))
        setActivity(map)
      })
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [dataProp])

  const weeks = buildGrid()

  const MONTH_LABELS = ['Ian', 'Feb', 'Mar', 'Apr', 'Mai', 'Iun', 'Iul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
  const DAY_LABELS = ['L', 'M', 'M', 'J', 'V', 'S', 'D']

  if (loading) {
    return (
      <div style={{padding:'16px 0',display:'flex',alignItems:'center',gap:10}}>
        <div className="spinner" style={{width:16,height:16,borderTopColor:'var(--l5)'}}/>
        <span style={{fontSize:13,color:'var(--ink3)'}}>Se încarcă activitatea...</span>
      </div>
    )
  }

  return (
    <div style={{position:'relative'}}>
      <div style={{display:'flex',gap:4,overflowX:'auto',paddingBottom:4}}>
        {/* Day labels column */}
        <div style={{display:'flex',flexDirection:'column',gap:3,paddingTop:18,flexShrink:0}}>
          {DAY_LABELS.map((d, i) => (
            <div key={i} style={{
              height:11, fontSize:9, color:'var(--ink4)',
              fontFamily:'var(--ff-m)', lineHeight:'11px',
              width:14, textAlign:'right',
            }}>
              {i % 2 === 0 ? d : ''}
            </div>
          ))}
        </div>

        {/* Grid */}
        <div style={{flex:1,minWidth:0}}>
          {/* Month labels */}
          <div style={{display:'flex',gap:3,marginBottom:4,height:14}}>
            {weeks.map((week, wi) => {
              // Show month label at the first week where month changes
              const firstDay = week[0].date
              const month = new Date(firstDay + 'T00:00:00').getMonth()
              const prevWeekMonth = wi > 0
                ? new Date(weeks[wi - 1][0].date + 'T00:00:00').getMonth()
                : -1
              const showLabel = wi === 0 || month !== prevWeekMonth
              return (
                <div key={wi} style={{
                  width:11, fontSize:9, color:'var(--ink3)',
                  fontFamily:'var(--ff-m)', flexShrink:0,
                  whiteSpace:'nowrap', overflow:'visible',
                }}>
                  {showLabel ? MONTH_LABELS[month] : ''}
                </div>
              )
            })}
          </div>

          {/* Cells */}
          <div style={{display:'flex',gap:3}}>
            {weeks.map((week, wi) => (
              <div key={wi} style={{display:'flex',flexDirection:'column',gap:3,flexShrink:0}}>
                {week.map(({ date }) => {
                  const day = activity.get(date)
                  const score = day?.score ?? 0
                  const tasks = day?.tasks_completed ?? 0
                  const isToday = date === new Date().toISOString().split('T')[0]
                  return (
                    <div
                      key={date}
                      onMouseEnter={() => setTooltip({ date, score, tasks })}
                      onMouseLeave={() => setTooltip(null)}
                      style={{
                        width: 11, height: 11,
                        borderRadius: 2,
                        background: getColor(score),
                        cursor: 'default',
                        outline: isToday ? '1.5px solid var(--l5l)' : undefined,
                        transition: 'background .2s',
                      }}
                    />
                  )
                })}
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Legend */}
      <div style={{display:'flex',alignItems:'center',gap:6,marginTop:10}}>
        <span style={{fontSize:11,color:'var(--ink4)',fontFamily:'var(--ff-m)'}}>Mai puțin</span>
        {[0, 0.2, 0.5, 0.75, 1].map(s => (
          <div key={s} style={{width:11,height:11,borderRadius:2,background:getColor(s)}}/>
        ))}
        <span style={{fontSize:11,color:'var(--ink4)',fontFamily:'var(--ff-m)'}}>Mai mult</span>
      </div>

      {/* Tooltip */}
      {tooltip && (
        <div style={{
          position:'fixed', bottom:20, left:'50%', transform:'translateX(-50%)',
          background:'var(--bg4)', border:'1px solid var(--line2)',
          borderRadius:8, padding:'7px 12px', fontSize:12,
          color:'var(--ink2)', fontFamily:'var(--ff-m)',
          pointerEvents:'none', zIndex:50, whiteSpace:'nowrap',
          boxShadow:'0 4px 16px var(--shadow)',
        }}>
          {formatDate(tooltip.date)} — {tooltip.tasks} activități, scor {Math.round(tooltip.score * 100)}%
        </div>
      )}
    </div>
  )
}
