'use client'

import { useEffect, useState } from 'react'

interface SRMStatus {
  goal_id: string
  srm_level: string
  message: string
}

const LEVEL_STYLE: Record<string, { bg: string; border: string; color: string }> = {
  L1: { bg: 'rgba(234,179,8,0.10)',  border: 'rgba(234,179,8,0.35)',  color: '#ca8a04' },
  L2: { bg: 'rgba(249,115,22,0.10)', border: 'rgba(249,115,22,0.35)', color: '#ea580c' },
  L3: { bg: 'rgba(239,68,68,0.10)',  border: 'rgba(239,68,68,0.35)',  color: '#dc2626' },
}

export default function SRMWarning({ goalId }: { goalId: string }) {
  const [srm, setSrm] = useState<SRMStatus | null>(null)

  useEffect(() => {
    fetch(`/api/proxy/srm/status/${goalId}`)
      .then(r => r.ok ? r.json() : null)
      .then(data => { if (data) setSrm(data) })
      .catch(() => {})
  }, [goalId])

  if (!srm || srm.srm_level === 'NONE') return null

  const style = LEVEL_STYLE[srm.srm_level] ?? LEVEL_STYLE.L1

  const handleConfirm = async () => {
    if (srm.srm_level !== 'L3') return
    await fetch(`/api/proxy/srm/confirm-l3/${goalId}`, { method: 'POST' })
    setSrm(null)
    window.location.reload()
  }

  return (
    <div
      style={{
        borderRadius: 10,
        padding: '12px 14px',
        border: `1.5px solid ${style.border}`,
        background: style.bg,
        marginBottom: 10,
      }}
    >
      <div style={{ display: 'flex', alignItems: 'flex-start', gap: 10 }}>
        <span style={{ fontSize: 18 }}>⚠️</span>
        <div style={{ flex: 1 }}>
          <div style={{ fontWeight: 600, fontSize: 13, color: style.color, marginBottom: 3 }}>
            Strategic Reset Mode {srm.srm_level}
          </div>
          <div style={{ fontSize: 12, color: 'var(--ink2)', lineHeight: 1.5 }}>
            {srm.message}
          </div>
          {srm.srm_level === 'L3' && (
            <button
              onClick={handleConfirm}
              style={{
                marginTop: 10,
                padding: '6px 14px',
                borderRadius: 8,
                background: '#dc2626',
                color: '#fff',
                fontWeight: 600,
                fontSize: 12,
                border: 'none',
                cursor: 'pointer',
              }}
            >
              Confirmă Resetare
            </button>
          )}
        </div>
      </div>
    </div>
  )
}
