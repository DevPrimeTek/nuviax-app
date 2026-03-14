import { NextRequest, NextResponse } from 'next/server'
const OPTS = { httpOnly:true, secure:process.env.NODE_ENV==='production', sameSite:'lax' as const, path:'/' }
export async function POST(req: NextRequest) {
  const { access_token, refresh_token } = await req.json()
  const res = NextResponse.json({ ok:true })
  if (access_token)  res.cookies.set('nv_access',  access_token,  { ...OPTS, maxAge:900 })
  if (refresh_token) res.cookies.set('nv_refresh', refresh_token, { ...OPTS, maxAge:2592000 })
  return res
}
