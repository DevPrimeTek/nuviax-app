import { cookies } from 'next/headers'
import { redirect, notFound } from 'next/navigation'
import AppShell from '@/components/layout/AppShell'
import GoalTabs from '@/components/GoalTabs'
import { goalsApi, ApiError } from '@/lib/api'
import type { Metadata } from 'next'

export const metadata: Metadata = { title: 'Obiectiv' }

export default async function GoalDetailPage({ params }: { params: { id: string } }) {
  const token = cookies().get('nv_access')?.value
  if (!token) redirect('/auth/login')

  let goal: Awaited<ReturnType<typeof goalsApi.get>>
  try {
    goal = await goalsApi.get(token, params.id)
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) redirect('/auth/login')
    if (err instanceof ApiError && err.status === 404) notFound()
    redirect('/app/goals')
  }

  const pct = Math.round(((goal as any).progress_score ?? 0) * 100)
  const statusLabel: Record<string, string> = {
    ACTIVE: 'Activ', WAITING: 'În așteptare', PAUSED: 'Pauză',
    COMPLETED: 'Finalizat', ARCHIVED: 'Arhivat',
  }

  const overviewContent = (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
      {goal.description && (
        <div style={{ fontSize: 14, color: 'var(--ink2)', lineHeight: 1.6 }}>
          {goal.description}
        </div>
      )}
      <div className="stats-3">
        <div className="stat">
          <div className="stat-v">{statusLabel[goal.status] ?? goal.status}</div>
          <div className="stat-l">status</div>
        </div>
        <div className="stat">
          <div className="stat-v" style={{ color: 'var(--l2l)' }}>{pct}%</div>
          <div className="stat-l">progres</div>
        </div>
        <div className="stat">
          <div className="stat-v">{(goal as any).days_left ?? '—'}</div>
          <div className="stat-l">zile rămase</div>
        </div>
      </div>
      <div style={{ fontSize: 12, color: 'var(--ink3)' }}>
        {new Date(goal.start_date).toLocaleDateString('ro-RO')} → {new Date(goal.end_date).toLocaleDateString('ro-RO')}
      </div>
    </div>
  )

  return (
    <AppShell>
      <div className="page">
        <div className="greet" style={{ marginBottom: 20 }}>
          <div>
            <div className="greet-title">{goal.name}</div>
            <div className="greet-sub">
              Sprint {(goal as any).sprint_number ?? 1}/{(goal as any).total_sprints ?? 1}
            </div>
          </div>
        </div>

        <GoalTabs goalId={params.id} overviewContent={overviewContent} />
      </div>
    </AppShell>
  )
}
