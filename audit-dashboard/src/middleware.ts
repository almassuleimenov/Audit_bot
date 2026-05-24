import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';
// D:\Project\backend_projects\audit_bot\audit-dashboard\src\middleware.ts

export function middleware(request: NextRequest) {
  const token = request.cookies.get('auth_token')?.value;

  // Если токена нет, и мы не на странице логина — перехватываем и редиректим
  if (!token && !request.nextUrl.pathname.startsWith('/login')) {
    return NextResponse.redirect(new URL('/login', request.url));
  }

  return NextResponse.next();
}

// Указываем пути, для которых должен срабатывать middleware
export const config = {
  matcher: ['/((?!api|_next/static|_next/image|favicon.ico|login).*)'],
};