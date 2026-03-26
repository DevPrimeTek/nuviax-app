import { NextRequest, NextResponse } from 'next/server'

const BACKEND = process.env.API_URL || process.env.NEXT_PUBLIC_API_URL || 'https://api.nuviax.app'

export async function POST(req: NextRequest) {
  const body = await req.text()

  // Always return 200 — backend is timing-safe, never reveals if email exists
  await fetch(`${BACKEND}/api/v1/auth/forgot-password`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body,
  }).catch(() => {})

  return NextResponse.json({ message: 'Dacă adresa există, vei primi un email.' })
}
