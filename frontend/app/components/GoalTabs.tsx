'use client'

import { useState } from 'react'
import ProgressCharts from '@/components/ProgressCharts'
import SRMWarning from '@/components/SRMWarning'

interface GoalTabsProps {
  goalId: string
  overviewContent: React.ReactNode
}

export default function GoalTabs({ goalId, overviewContent }: GoalTabsProps) {
  const [tab, setTab] = useState<'overview' | 'progress'>('overview')

  return (
    <>
      <div style={{ display: 'flex', gap: 4, marginBottom: 20, background: 'var(--bg3)', borderRadius: 10, padding: 4 }}>
        {(['overview', 'progress'] as const).map(t => (
          <button
            key={t}
            onClick={() => setTab(t)}
            style={{
              flex: 1,
              padding: '7px 0',
              borderRadius: 8,
              border: 'none',
              cursor: 'pointer',
              fontSize: 13,
              fontWeight: 600,
              background: tab === t ? 'var(--bg2)' : 'transparent',
              color: tab === t ? 'var(--ink1)' : 'var(--ink3)',
              transition: 'all .15s',
            }}
          >
            {t === 'overview' ? 'Prezentare' : 'Progres'}
          </button>
        ))}
      </div>

      {tab === 'overview' && (
        <>
          <SRMWarning goalId={goalId} />
          {overviewContent}
        </>
      )}
      {tab === 'progress' && <ProgressCharts goalId={goalId} />}
    </>
  )
}
