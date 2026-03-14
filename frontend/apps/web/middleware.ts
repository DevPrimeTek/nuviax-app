import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

const PROTECTED = ['/dashboard', '/today', '/goals', '/recap', '/settings']
const AUTH_ONLY  = ['/login', '/register']

export function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl
  const access  = req.cookies.get('nv_access')?.value
  const refresh = req.cookies.get('nv_refresh')?.value
  const authed  = !!(access || refresh)

  if (AUTH_ONLY.some(p => pathname.startsWith(p)) && authed)
    return NextResponse.redirect(new URL('/dashboard', req.url))

  if (PROTECTED.some(p => pathname.startsWith(p)) && !authed) {
    const url = new URL('/login', req.url)
    url.searchParams.set('next', pathname)
    return NextResponse.redirect(url)
  }

  return NextResponse.next()
}

export const config = {
  matcher: ['/((?!_next/static|_next/image|favicon.ico|api/).*)'],
}
