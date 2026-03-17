'use client'

export interface CeremonyData {
  id: string
  tier: 'PLATINUM' | 'GOLD' | 'SILVER' | 'BRONZE'
  message: string
  badge: string
  stats: {
    sprint_number: number
    score: number
    grade: string
    streak_days: number
    consistency: number
  }
  achievements_unlocked: string[]
  is_evolution: boolean
}

const TIER_COLORS = {
  PLATINUM: 'from-purple-500 to-pink-500',
  GOLD:     'from-yellow-400 to-orange-500',
  SILVER:   'from-gray-300 to-gray-500',
  BRONZE:   'from-amber-600 to-amber-800',
} as const

const TIER_EMOJI = {
  PLATINUM: '🏆',
  GOLD:     '✨',
  SILVER:   '👏',
  BRONZE:   '✓',
} as const

export default function CeremonyModal({
  ceremony,
  onClose,
}: {
  ceremony: CeremonyData | null
  onClose: () => void
}) {
  if (!ceremony) return null

  const markViewed = async () => {
    await fetch(`/api/proxy/ceremonies/${ceremony.id}/view`, { method: 'POST' })
    onClose()
  }

  return (
    <div
      className="fixed inset-0 flex items-center justify-center z-50 p-4"
      style={{ background: 'rgba(0,0,0,0.55)' }}
    >
      <div
        className="rounded-2xl max-w-sm w-full p-6 shadow-2xl"
        style={{ background: 'var(--bg2)', color: 'var(--ink1)' }}
      >
        {/* Header */}
        <div
          className={`bg-gradient-to-r ${TIER_COLORS[ceremony.tier]} rounded-xl p-6 text-white text-center mb-6`}
        >
          <div style={{ fontSize: 56, lineHeight: 1.1 }}>{TIER_EMOJI[ceremony.tier]}</div>
          <div style={{ fontSize: 11, opacity: 0.9, textTransform: 'uppercase', letterSpacing: 1, marginTop: 6 }}>
            {ceremony.tier}
          </div>
          {ceremony.is_evolution && (
            <div
              style={{
                fontSize: 12, marginTop: 8,
                background: 'rgba(255,255,255,0.2)',
                borderRadius: 99, padding: '2px 12px',
                display: 'inline-block',
              }}
            >
              🚀 Evolution Sprint
            </div>
          )}
        </div>

        {/* Message */}
        <p style={{ textAlign: 'center', fontSize: 15, marginBottom: 20 }}>
          {ceremony.message || `Sprint ${ceremony.stats.sprint_number} finalizat!`}
        </p>

        {/* Stats grid */}
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10, marginBottom: 20 }}>
          <div style={{ background: 'var(--bg3)', borderRadius: 10, padding: '10px 12px', textAlign: 'center' }}>
            <div style={{ fontSize: 22, fontWeight: 700, color: 'var(--l0l)' }}>{ceremony.stats.grade}</div>
            <div style={{ fontSize: 11, color: 'var(--ink3)', marginTop: 2 }}>Grade</div>
          </div>
          <div style={{ background: 'var(--bg3)', borderRadius: 10, padding: '10px 12px', textAlign: 'center' }}>
            <div style={{ fontSize: 22, fontWeight: 700, color: 'var(--l2l)' }}>{ceremony.stats.streak_days}</div>
            <div style={{ fontSize: 11, color: 'var(--ink3)', marginTop: 2 }}>Streak Days</div>
          </div>
        </div>

        {/* Achievements unlocked */}
        {ceremony.achievements_unlocked.length > 0 && (
          <div style={{ marginBottom: 20 }}>
            <div style={{ fontSize: 12, fontWeight: 600, marginBottom: 8 }}>🎉 Achievements Unlocked:</div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
              {ceremony.achievements_unlocked.map(ach => (
                <div
                  key={ach}
                  style={{
                    background: 'rgba(16,185,129,0.1)',
                    color: 'var(--l2l)',
                    borderRadius: 8,
                    padding: '6px 10px',
                    fontSize: 12,
                  }}
                >
                  {ach}
                </div>
              ))}
            </div>
          </div>
        )}

        {/* CTA */}
        <button
          onClick={markViewed}
          style={{
            width: '100%',
            padding: '12px',
            borderRadius: 10,
            background: 'var(--l0)',
            color: '#fff',
            fontWeight: 600,
            fontSize: 14,
            border: 'none',
            cursor: 'pointer',
          }}
        >
          Continuă
        </button>
      </div>
    </div>
  )
}
