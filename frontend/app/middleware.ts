import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

/**
 * Middleware pentru protecție rute
 * - /admin/* = spațiu autentificare separat (login la /admin/login)
 * - restul aplicației = login la /auth/login
 */
export function middleware(request: NextRequest) {
  const token = request.cookies.get('nv_access')?.value
  const { pathname } = request.nextUrl

  // ── Admin space (izolat de aplicație) ──────────────────────────
  if (pathname.startsWith('/admin')) {
    // /admin/login este public
    if (pathname === '/admin/login') {
      // dacă are deja token, lasă pagina să facă verificarea (poate nu e admin);
      // nu facem redirect automat aici pentru a evita bucle de login
      return NextResponse.next()
    }
    // oricare altă rută /admin/* necesită token
    if (!token) {
      const loginUrl = new URL('/admin/login', request.url)
      return NextResponse.redirect(loginUrl)
    }
    return NextResponse.next()
  }

  // ── Restul aplicației ──────────────────────────────────────────
  const publicPaths = ['/auth']
  const isPublicPath = publicPaths.some(path => pathname.startsWith(path))

  if (!token && !isPublicPath && pathname !== '/') {
    const loginUrl = new URL('/auth/login', request.url)
    loginUrl.searchParams.set('redirect', pathname)
    return NextResponse.redirect(loginUrl)
  }

  if (token && (pathname === '/auth/login' || pathname === '/auth/register')) {
    return NextResponse.redirect(new URL('/dashboard', request.url))
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
