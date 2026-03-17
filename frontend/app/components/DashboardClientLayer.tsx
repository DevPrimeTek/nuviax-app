'use client'

import { useEffect, useState } from 'react'
import CeremonyModal, { type CeremonyData } from '@/components/CeremonyModal'

export default function DashboardClientLayer() {
  const [ceremony, setCeremony] = useState<CeremonyData | null>(null)

  useEffect(() => {
    fetch('/api/proxy/ceremonies/unviewed')
      .then(r => r.ok ? r.json() : null)
      .then(data => {
        const first = data?.ceremonies?.[0]
        if (first) setCeremony(first as CeremonyData)
      })
      .catch((err) => { console.error('[DashboardClientLayer] ceremonies fetch failed:', err) })
  }, [])

  return (
    <CeremonyModal
      ceremony={ceremony}
      onClose={() => setCeremony(null)}
    />
  )
}
