import type { Metadata } from 'next'
import '../../packages/styles/globals.css'
import './landing.css'

export const metadata: Metadata = {
  title: 'NUViaX — Crești deliberat, zi cu zi',
  description: 'Scrii un obiectiv. Primești activitățile zilnice clare, organizate în 9 etape pe 365 de zile. Fără teorie. Doar acțiune.',
  keywords: ['productivitate','obiective','creștere personală','NUViaX'],
  openGraph: {
    title: 'NUViaX — Crești deliberat, zi cu zi',
    description: 'De la un obiectiv → activități zilnice clare, 365 de zile.',
    type: 'website',
    locale: 'ro_RO',
    url: 'https://nuviaxapp.com',
  },
  metadataBase: new URL('https://nuviaxapp.com'),
}

export default function LandingLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ro" data-theme="dark">
      <head>
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link href="https://fonts.googleapis.com/css2?family=Bricolage+Grotesque:opsz,wght@12..96,400;12..96,500;12..96,700;12..96,800&family=DM+Sans:ital,opsz,wght@0,9..40,300;0,9..40,400;0,9..40,500;0,9..40,600;0,9..40,700&family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet" />
      </head>
      <body>{children}</body>
    </html>
  )
}
