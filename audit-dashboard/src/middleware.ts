// src/lib/middleware.ts
import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export function middleware(req: NextRequest) {
  const basicAuth = req.headers.get('authorization');
  const url = req.nextUrl;

  const expectedUser = process.env.ADMIN_USER || 'admin';
  const expectedPass = process.env.ADMIN_PASSWORD || 'default_secure_password';
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