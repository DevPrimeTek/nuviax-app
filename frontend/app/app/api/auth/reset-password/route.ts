import { NextRequest, NextResponse } from 'next/server'

const BACKEND = process.env.API_URL || process.env.NEXT_PUBLIC_API_URL || 'https://api.nuviax.app'

export async function POST(req: NextRequest) {
  const body = await req.text()

  const res = await fetch(`${BACKEND}/api/v1/auth/reset-password`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body,
  }).catch(() => null)

  if (!res || !res.ok) {
    const data = await res?.json().catch(() => ({})) ?? {}
    return NextResponse.json(
      { error: data.error || 'Token invalid sau expirat.' },
      { status: res?.status ?? 500 },
    )
  }

  const data = await res.json().catch(() => ({}))
  return NextResponse.json(data)
}
