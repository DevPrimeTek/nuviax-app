'use client'

import { useEffect, useState } from 'react'
import {
  LineChart, Line,
  BarChart, Bar,
  XAxis, YAxis,
  CartesianGrid, Tooltip,
  ResponsiveContainer,
} from 'recharts'

interface TrajectoryPoint {
  date: string
  actual_pct: number
  expected_pct: number
  delta: number
  trend: string
}

interface ProgressVizRaw {
  goal_id: string
  trajectory: TrajectoryPoint[]
}

// Normalised data shapes for the charts
interface TimelinePoint { label: string; actual: number; expected: number }
interface VelocityPoint { label: string; delta: number }

function toChartData(raw: ProgressVizRaw) {
  const timeline: TimelinePoint[] = (raw.trajectory ?? []).map((p, i) => ({
    label: `#${i + 1}`,
    actual: Math.round(p.actual_pct * 100),
    expected: Math.round(p.expected_pct * 100),
  }))
  const velocity: VelocityPoint[] = (raw.trajectory ?? []).map((p, i) => ({
    label: `#${i + 1}`,
    delta: Math.round(p.delta * 100),
  }))
  return { timeline, velocity }
}

export default function ProgressCharts({ goalId }: { goalId: string }) {
  const [raw, setRaw] = useState<ProgressVizRaw | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  useEffect(() => {
    fetch(`/api/proxy/goals/${goalId}/visualize`)
      .then(r => { if (!r.ok) throw new Error(); return r.json() })
      .then(data => { setRaw(data); setLoading(false) })
      .catch(() => { setError(true); setLoading(false) })
  }, [goalId])

  if (loading) return (
    <div style={{ padding: 32, textAlign: 'center', color: 'var(--ink3)', fontSize: 13 }}>
      Se generează vizualizarea...
    </div>
  )
  if (error || !raw) return (
    <div style={{ padding: 32, textAlign: 'center', color: 'var(--ink3)', fontSize: 13 }}>
      Nu există date de vizualizare încă.
    </div>
  )

  const { timeline, velocity } = toChartData(raw)

  // Latest trajectory point for current status
  const latest = raw.trajectory?.[raw.trajectory.length - 1]
  const trend = latest?.trend ?? 'ON_TRACK'
  const trendLabel: Record<string, string> = {
    AHEAD: '↑ Înaintea planului', ON_TRACK: '→ Pe plan',
    SLIGHTLY_BEHIND: '↓ Ușor în urmă', BEHIND: '↓ În urmă', AT_RISK: '⚠ Risc',
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 28, padding: '16px 0' }}>
      {/* Current trend chip */}
      <div
        style={{
          display: 'inline-flex', alignItems: 'center', gap: 6,
          padding: '6px 14px', borderRadius: 99, alignSelf: 'flex-start',
          background: trend === 'AHEAD' || trend === 'ON_TRACK'
            ? 'var(--l2g)' : 'rgba(249,115,22,0.12)',
          color: trend === 'AHEAD' || trend === 'ON_TRACK'
            ? 'var(--l2l)' : 'var(--ul)',
          fontSize: 12, fontWeight: 600,
        }}
      >
        {trendLabel[trend] ?? trend}
      </div>

      {/* Progress vs Expected */}
      <div>
        <div style={{ fontSize: 13, fontWeight: 600, marginBottom: 12, color: 'var(--ink2)' }}>
          Progres real vs. așteptat (%)
        </div>
        {timeline.length === 0 ? (
          <div style={{ fontSize: 12, color: 'var(--ink3)' }}>Insuficiente date.</div>
        ) : (
          <ResponsiveContainer width="100%" height={180}>
            <LineChart data={timeline} margin={{ top: 4, right: 4, left: -20, bottom: 0 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--line2)" />
              <XAxis dataKey="label" tick={{ fontSize: 11, fill: 'var(--ink3)' }} />
              <YAxis tick={{ fontSize: 11, fill: 'var(--ink3)' }} />
              <Tooltip
                contentStyle={{ background: 'var(--bg2)', border: '1px solid var(--line2)', borderRadius: 8 }}
                labelStyle={{ color: 'var(--ink2)', fontSize: 11 }}
              />
              <Line type="monotone" dataKey="actual"   stroke="var(--l0l)"  strokeWidth={2} dot={false} name="Real" />
              <Line type="monotone" dataKey="expected" stroke="var(--ink4)" strokeWidth={1.5} strokeDasharray="4 2" dot={false} name="Așteptat" />
            </LineChart>
          </ResponsiveContainer>
        )}
      </div>

      {/* Delta per snapshot */}
      <div>
        <div style={{ fontSize: 13, fontWeight: 600, marginBottom: 12, color: 'var(--ink2)' }}>
          Diferență față de plan (pp)
        </div>
        {velocity.length === 0 ? (
          <div style={{ fontSize: 12, color: 'var(--ink3)' }}>Insuficiente date.</div>
        ) : (
          <ResponsiveContainer width="100%" height={160}>
            <BarChart data={velocity} margin={{ top: 4, right: 4, left: -20, bottom: 0 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--line2)" />
              <XAxis dataKey="label" tick={{ fontSize: 11, fill: 'var(--ink3)' }} />
              <YAxis tick={{ fontSize: 11, fill: 'var(--ink3)' }} />
              <Tooltip
                contentStyle={{ background: 'var(--bg2)', border: '1px solid var(--line2)', borderRadius: 8 }}
                labelStyle={{ color: 'var(--ink2)', fontSize: 11 }}
              />
              <Bar
                dataKey="delta"
                fill="var(--l2l)"
                radius={[4, 4, 0, 0]}
                name="Delta"
              />
            </BarChart>
          </ResponsiveContainer>
        )}
      </div>

      {/* Trajectory table (compact) */}
      {raw.trajectory.length > 0 && (
        <div>
          <div style={{ fontSize: 13, fontWeight: 600, marginBottom: 8, color: 'var(--ink2)' }}>
            Snapshoturi traiectorie
          </div>
          <div style={{ overflowX: 'auto' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 11 }}>
              <thead>
                <tr style={{ color: 'var(--ink3)' }}>
                  <th style={{ textAlign: 'left', padding: '4px 6px' }}>Data</th>
                  <th style={{ textAlign: 'right', padding: '4px 6px' }}>Real %</th>
                  <th style={{ textAlign: 'right', padding: '4px 6px' }}>Așteptat %</th>
                  <th style={{ textAlign: 'right', padding: '4px 6px' }}>Trend</th>
                </tr>
              </thead>
              <tbody>
                {raw.trajectory.map((p, i) => (
                  <tr key={i} style={{ borderTop: '1px solid var(--line2)' }}>
                    <td style={{ padding: '5px 6px', color: 'var(--ink2)' }}>
                      {new Date(p.date).toLocaleDateString('ro-RO', { day: 'numeric', month: 'short' })}
                    </td>
                    <td style={{ textAlign: 'right', padding: '5px 6px', color: 'var(--l0l)', fontWeight: 600 }}>
                      {Math.round(p.actual_pct * 100)}%
                    </td>
                    <td style={{ textAlign: 'right', padding: '5px 6px', color: 'var(--ink3)' }}>
                      {Math.round(p.expected_pct * 100)}%
                    </td>
                    <td style={{ textAlign: 'right', padding: '5px 6px', color: 'var(--ink3)' }}>
                      {trendLabel[p.trend] ?? p.trend}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  )
}
