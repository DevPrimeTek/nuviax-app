'use client'

import { useEffect, useState } from 'react'
import AppShell from '@/components/layout/AppShell'

interface AchievementBadge {
  id: string
  badge_type: string
  goal_id?: string
  sprint_id?: string
  awarded_at: string
}

const BADGE_META: Record<string, { label: string; description: string; icon: string; category: string }> = {
  STARTER:          { label: 'Starter',          description: 'Ai completat prima activitate',           icon: '🌱', category: 'MILESTONE' },
  CONSISTENT_WEEK:  { label: 'O săptămână',       description: '7 zile consecutive active',              icon: '🗓️', category: 'STREAK' },
  CONSISTENT_MONTH: { label: 'O lună',            description: '30 de zile consecutive active',          icon: '📅', category: 'STREAK' },
  GRADE_HUNTER:     { label: 'Grade Hunter',      description: 'Primul sprint cu grad A',                icon: '🎯', category: 'EXCELLENCE' },
  PERFECTIONIST:    { label: 'Perfectionist',     description: 'Sprint completat 100%',                  icon: '💎', category: 'EXCELLENCE' },
  GOAL_SLAYER:      { label: 'Goal Slayer',        description: 'Ai finalizat primul obiectiv',           icon: '🏁', category: 'MILESTONE' },
  MULTI_TASKER:     { label: 'Multi Tasker',       description: '3 obiective active simultan',            icon: '⚡', category: 'MILESTONE' },
  COMEBACK_KID:     { label: 'Comeback Kid',       description: 'Ai revenit după o pauză',               icon: '💪', category: 'RESILIENCE' },
  EARLY_BIRD:       { label: 'Early Bird',         description: 'Activități completate înainte de 9:00', icon: '🌅', category: 'STREAK' },
  MARATHON_RUNNER:  { label: 'Marathon Runner',    description: 'Obiectiv de 90+ zile finalizat',        icon: '🏃', category: 'MILESTONE' },
}

const CATEGORIES = ['MILESTONE', 'STREAK', 'EXCELLENCE', 'RESILIENCE']

export default function AchievementsPage() {
  const [badges, setBadges] = useState<AchievementBadge[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetch('/api/proxy/achievements')
      .then(r => r.json())
      .then(data => {
        setBadges(data.achievements || [])
        setLoading(false)
      })
      .catch(() => setLoading(false))
  }, [])

  const unlockedTypes = new Set(badges.map(b => b.badge_type))

  return (
    <AppShell>
      <div className="page">
        <div className="greet" style={{ marginBottom: 8 }}>
          <div>
            <div className="greet-title">Achievements</div>
            <div className="greet-sub">{badges.length} deblocate</div>
          </div>
        </div>

        {loading ? (
          <div style={{ padding: 32, textAlign: 'center', color: 'var(--ink3)', fontSize: 14 }}>
            Se încarcă...
          </div>
        ) : (
          CATEGORIES.map(cat => {
            const catBadges = Object.entries(BADGE_META).filter(([, m]) => m.category === cat)
            return (
              <div key={cat} style={{ marginBottom: 28 }}>
                <div className="sec-lbl">{cat}</div>
                <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                  {catBadges.map(([type, meta]) => {
                    const unlocked = unlockedTypes.has(type)
                    const badge = badges.find(b => b.badge_type === type)
                    return (
                      <div
                        key={type}
                        className="goal-card"
                        style={{ opacity: unlocked ? 1 : 0.5 }}
                      >
                        <div className="goal-top">
                          <div style={{ fontSize: 28, marginRight: 12, lineHeight: 1 }}>{meta.icon}</div>
                          <div style={{ flex: 1 }}>
                            <div className="goal-name">{meta.label}</div>
                            <div className="goal-meta">{meta.description}</div>
                            {unlocked && badge && (
                              <div style={{ fontSize: 11, color: 'var(--l2l)', marginTop: 3 }}>
                                ✓ Deblocat {new Date(badge.awarded_at).toLocaleDateString('ro-RO')}
                              </div>
                            )}
                          </div>
                          {unlocked && (
                            <div
                              style={{
                                width: 24, height: 24, borderRadius: 99,
                                background: 'var(--l2g)', display: 'flex',
                                alignItems: 'center', justifyContent: 'center',
                              }}
                            >
                              <svg width="13" height="13" viewBox="0 0 24 24" fill="none"
                                stroke="var(--l2l)" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                                <polyline points="20,6 9,17 4,12"/>
                              </svg>
                            </div>
                          )}
                        </div>
                      </div>
                    )
                  })}
                </div>
              </div>
            )
          })
        )}
      </div>
    </AppShell>
  )
}
