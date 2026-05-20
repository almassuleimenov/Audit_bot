// src/lib/middleware.ts
import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export function middleware(req: NextRequest) {
  const basicAuth = req.headers.get('authorization');
  const url = req.nextUrl;

  const expectedUser = process.env.ADMIN_USER;
  const expectedPass = process.env.ADMIN_PASSWORD;

  // Fail-fast: Если секреты не заданы в окружении, жестко блокируем вход
  if (!expectedUser || !expectedPass) {
    return new NextResponse('Internal Server Error: Admin credentials are not configured.', {
      status: 500,
    });
  }

  const expectedAuthValue = Buffer.from(`${expectedUser}:${expectedPass}`).toString('base64');

  if (basicAuth) {
    const authValue = basicAuth.split(' ')[1];
    if (authValue === expectedAuthValue) {
      return NextResponse.next();
    }
  }

  url.pathname = '/api/auth';
  return new NextResponse('Auth required', {
    status: 401,
    headers: {
      'WWW-Authenticate': 'Basic realm="Secure Area"',
    },
  });
}

export const config = {
  matcher: ['/((?!api|_next/static|_next/image|favicon.ico).*)'],
};