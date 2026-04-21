import { NextRequest, NextResponse } from 'next/server'

const BACKEND = process.env.API_URL || process.env.NEXT_PUBLIC_API_URL || 'https://api.nuviax.app'
const OPTS = { httpOnly: true, secure: process.env.NODE_ENV === 'production', sameSite: 'lax' as const, path: '/' }

/**
 * Admin login:
 * 1. Apelează backend /auth/login cu email + parolă
 * 2. Verifică că user-ul e admin apelând /admin/stats (returnează 404 pentru non-admin)
 * 3. Doar dacă e admin, setează cookie-urile și răspunde cu succes
 *    (pentru utilizatori non-admin, NU setăm cookies — logica rămâne curată)
 */
export async function POST(req: NextRequest) {
  const body = await req.text()

  // 1. Login pe backend
  const loginRes = await fetch(`${BACKEND}/api/v1/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body,
  })

  const loginData = await loginRes.json().catch(() => ({}))

  if (!loginRes.ok) {
    return NextResponse.json(
      { error: loginData.error || 'Email sau parolă incorectă.' },
      { status: loginRes.status },
    )
  }

  // MFA nu este suportat în fluxul admin simplificat — dacă cineva are MFA pe cont,
  // backend-ul va răspunde cu mfa_required și nu vom avea access_token.
  if (loginData.mfa_required || !loginData.access_token) {
    return NextResponse.json(
      { error: 'Contul necesită MFA. Folosește fluxul normal de autentificare.' },
      { status: 401 },
    )
  }

  // 2. Verifică privilegii admin — /admin/stats returnează 404 pentru non-admin
  const statsRes = await fetch(`${BACKEND}/api/v1/admin/stats`, {
    method: 'GET',
    headers: { 'Authorization': `Bearer ${loginData.access_token}` },
  })

  if (!statsRes.ok) {
    // Nu este admin: nu setăm cookies; mesaj neutru pentru a nu divulga existența panelului
    return NextResponse.json(
      { error: 'Cont neautorizat pentru panoul de administrare.' },
      { status: 403 },
    )
  }

  // 3. Setează cookies doar după confirmare admin
  const resp = NextResponse.json({ ok: true })
  resp.cookies.set('nv_access', loginData.access_token, { ...OPTS, maxAge: 900 })
  if (loginData.refresh_token) {
    resp.cookies.set('nv_refresh', loginData.refresh_token, { ...OPTS, maxAge: 2592000 })
  }
  return resp
}
