import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

/**
 * Middleware pentru protecție rute
 * Redirecționează utilizatorii neautentificați către /login
 */
export function middleware(request: NextRequest) {
  const token = request.cookies.get('nv_access')?.value
  const { pathname } = request.nextUrl

  // Rute publice (nu necesită autentificare)
  const publicPaths = ['/login', '/register', '/auth']
  const isPublicPath = publicPaths.some(path => pathname.startsWith(path))

  // Dacă utilizatorul nu este autentificat și încearcă să acceseze o rută protejată
  if (!token && !isPublicPath && pathname !== '/') {
    const loginUrl = new URL('/login', request.url)
    loginUrl.searchParams.set('redirect', pathname)
    return NextResponse.redirect(loginUrl)
  }

  // Dacă utilizatorul este autentificat și încearcă să acceseze /login sau /register
  if (token && (pathname === '/login' || pathname === '/register')) {
    return NextResponse.redirect(new URL('/today', request.url))
  }

  return NextResponse.next()
}

export const config = {
  matcher: [
    /*
     * Match all request paths except:
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     * - public files (public folder)
     */
    '/((?!_next/static|_next/image|favicon.ico|.*\\..*|api).*)',
  ],
}