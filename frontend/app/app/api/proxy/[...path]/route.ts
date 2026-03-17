import { cookies } from 'next/headers'
import { NextRequest, NextResponse } from 'next/server'

const BACKEND = process.env.API_URL || process.env.NEXT_PUBLIC_API_URL || 'https://api.nuviax.app'
const OPTS = { httpOnly: true, secure: process.env.NODE_ENV === 'production', sameSite: 'lax' as const, path: '/' }

async function refreshAccessToken(refreshToken: string): Promise<string | null> {
  try {
    const res = await fetch(`${BACKEND}/api/v1/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    })
    if (!res.ok) return null
    const data = await res.json()
    return data.access_token || null
  } catch {
    return null
  }
}

async function handler(req: NextRequest, { params }: { params: { path: string[] } }) {
  const cookieStore = cookies()
  let token = cookieStore.get('nv_access')?.value
  if (!token) return NextResponse.json({ error: 'Neautentificat' }, { status: 401 })

  const path = params.path.join('/')
  const url = `${BACKEND}/api/v1/${path}${req.nextUrl.search}`

  const body = req.method !== 'GET' && req.method !== 'HEAD'
    ? await req.text()
    : undefined

  const makeRequest = (accessToken: string) =>
    fetch(url, {
      method: req.method,
      headers: { 'Authorization': `Bearer ${accessToken}`, 'Content-Type': 'application/json' },
      body,
    })

  let res = await makeRequest(token)

  // Dacă token-ul a expirat, încearcă refresh automat
  if (res.status === 401) {
    const refreshToken = cookieStore.get('nv_refresh')?.value
    if (refreshToken) {
      const newToken = await refreshAccessToken(refreshToken)
      if (newToken) {
        token = newToken
        res = await makeRequest(newToken)
        // Setează noul access token în cookie
        const responseData = await res.text()
        const response = new NextResponse(responseData, {
          status: res.status,
          headers: { 'Content-Type': 'application/json' },
        })
        response.cookies.set('nv_access', newToken, { ...OPTS, maxAge: 900 })
        return response
      }
    }
    // Refresh a eșuat — returnează 401
    return NextResponse.json({ error: 'Sesiune expirată' }, { status: 401 })
  }

  const data = await res.text()
  return new NextResponse(data, {
    status: res.status,
    headers: { 'Content-Type': 'application/json' },
  })
}

export { handler as GET, handler as POST, handler as PUT, handler as PATCH, handler as DELETE }
