import { cookies } from 'next/headers'
import { NextRequest, NextResponse } from 'next/server'

const BACKEND = process.env.API_URL || process.env.NEXT_PUBLIC_API_URL || 'https://api.nuviax.app'

async function handler(req: NextRequest, { params }: { params: { path: string[] } }) {
  const token = cookies().get('nv_access')?.value
  if (!token) return NextResponse.json({ error: 'Neautentificat' }, { status: 401 })

  const path = params.path.join('/')
  const url = `${BACKEND}/api/v1/${path}${req.nextUrl.search}`

  const headers: Record<string, string> = {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
  }

  const body = req.method !== 'GET' && req.method !== 'HEAD'
    ? await req.text()
    : undefined

  const res = await fetch(url, { method: req.method, headers, body })
  const data = await res.text()

  return new NextResponse(data, {
    status: res.status,
    headers: { 'Content-Type': 'application/json' },
  })
}

export { handler as GET, handler as POST, handler as PUT, handler as PATCH, handler as DELETE }
