import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';
// D:\Project\backend_projects\audit_bot\audit-dashboard\src\lib\middleware.ts
export function middleware(req: NextRequest) {
  const basicAuth = req.headers.get('authorization');
  const url = req.nextUrl;

  // Ожидаемые логин и пароль: admin / mamba2026
  // Закодировано в Base64: YWRtaW46bWFtYmEyMDI2
  if (basicAuth) {
    const authValue = basicAuth.split(' ')[1];
    if (authValue === 'YWRtaW46bWFtYmEyMDI2') {
      return NextResponse.next();
    }
  }

  // Если пароль неверный или его нет — запрашиваем окно браузера
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