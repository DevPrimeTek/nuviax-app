'use client'
import { useEffect } from 'react'

export default function HomePage() {
  useEffect(() => {
    window.location.href = '/dashboard'
  }, [])

  return (
    <div style={{
      display: 'flex', alignItems: 'center', justifyContent: 'center',
      height: '100vh', background: 'var(--bg)', color: 'var(--ink3)',
    }}>
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '16px' }}>
        <div style={{
          width: '32px', height: '32px', border: '3px solid var(--line)',
          borderTopColor: 'var(--l0)', borderRadius: '50%',
          animation: 'spin 0.8s linear infinite',
        }} />
        <div style={{ fontFamily: 'var(--ff-b)', fontSize: '14px', fontWeight: 500 }}>
          Se încarcă...
        </div>
      </div>
      <style jsx>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
    </div>
  )
}
