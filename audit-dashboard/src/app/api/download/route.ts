import { NextResponse } from 'next/server';
// src/app/api/download/route.ts
export async function GET() {
  const backendUrl = `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/export`;
  const apiKey = process.env.API_SECRET_KEY;

  try {
    const response = await fetch(backendUrl, {
      headers: { 'X-API-Key': apiKey || '' }
    });

    if (!response.ok) throw new Error('Ошибка бэкенда');

    const blob = await response.blob();
    const headers = new Headers();
    headers.set('Content-Type', 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet');
    headers.set('Content-Disposition', 'attachment; filename="Audit_Report.xlsx"');

    return new NextResponse(blob, { status: 200, headers });
  } catch (error) {
    return new NextResponse('Ошибка экспорта', { status: 500 });
  }
}