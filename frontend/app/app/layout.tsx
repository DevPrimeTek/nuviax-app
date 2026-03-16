import type { Metadata } from 'next'
import '../styles/globals.css'
import './app.css'                            // app-specific styles

export const metadata: Metadata = {
  title: { default: 'NUViaX', template: '%s · NUViaX' },
  description: 'Crești deliberat, zi cu zi.',
  icons: { icon: '/favicon.ico' },
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ro" data-theme="dark" suppressHydrationWarning>
      <head>
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link href="https://fonts.googleapis.com/css2?family=Bricolage+Grotesque:opsz,wght@12..96,400;12..96,500;12..96,700;12..96,800&family=DM+Sans:ital,opsz,wght@0,9..40,300;0,9..40,400;0,9..40,500;0,9..40,600;0,9..40,700&family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet" />
      </head>
      <body>
        {children}
        <script dangerouslySetInnerHTML={{ __html:
          `try{document.documentElement.dataset.theme=localStorage.getItem('nv_theme')||'dark'}catch(e){}`
        }}/>
      </body>
    </html>
  )
}
