import { NextRequest, NextResponse } from 'next/server'

const BACKEND = process.env.NEXT_PUBLIC_API_URL || 'https://api.nuviax.app'
const OPTS = { httpOnly: true, secure: process.env.NODE_ENV === 'production', sameSite: 'lax' as const, path: '/' }

export async function POST(req: NextRequest) {
  const body = await req.text()

  const res = await fetch(`${BACKEND}/api/v1/auth/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body,
  })

  const data = await res.json().catch(() => ({}))

  if (!res.ok) {
    return NextResponse.json(
      { error: data.error || 'Înregistrare eșuată' },
      { status: res.status },
    )
  }

  const resp = NextResponse.json({ ok: true })
  if (data.access_token) resp.cookies.set('nv_access', data.access_token, { ...OPTS, maxAge: 900 })
  if (data.refresh_token) resp.cookies.set('nv_refresh', data.refresh_token, { ...OPTS, maxAge: 2592000 })
  return resp
}
