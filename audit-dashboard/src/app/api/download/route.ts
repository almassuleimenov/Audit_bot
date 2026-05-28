import { NextResponse } from 'next/server';
import { cookies } from 'next/headers';

export async function GET() {
  const backendUrl = `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/export`;
  const apiKey = process.env.API_SECRET_KEY;

  try {
    const cookieStore = await cookies();
    const authToken = cookieStore.get('auth_token')?.value;

    const headers = new Headers();
    if (apiKey) {
      headers.set('X-API-Key', apiKey);
    }
    if (authToken) {
      headers.set('Cookie', `auth_token=${authToken}`);
    }

    const response = await fetch(backendUrl, { headers });

    if (!response.ok) {
      const errorText = await response.text();
      console.error(`Ошибка бэкенда при экспорте: ${response.status} - ${errorText}`);
      throw new Error(`Ошибка бэкенда: HTTP ${response.status}`);
    }

    const blob = await response.blob();
    const resHeaders = new Headers();
    resHeaders.set('Content-Type', 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet');
    resHeaders.set('Content-Disposition', 'attachment; filename="Audit_Report.xlsx"');

    return new NextResponse(blob, { status: 200, headers: resHeaders });
  } catch (error) {
    console.error("Download route error:", error);
    return new NextResponse('Ошибка экспорта', { status: 500 });
  }
}