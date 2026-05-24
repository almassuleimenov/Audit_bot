import { cookies } from 'next/headers';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
const API_SECRET_KEY = process.env.API_SECRET_KEY;

interface FetchOptions extends RequestInit {
  params?: Record<string, string | number | boolean>;
}

// ЭКСПОРТ ОБЯЗАТЕЛЕН
export async function apiFetch<T>(
  endpoint: string,
  options: FetchOptions = {}
): Promise<T> {
  const url = new URL(`${API_URL}${endpoint}`);

  if (options.params) {
    Object.entries(options.params).forEach(([key, value]) => {
      url.searchParams.append(key, String(value));
    });
  }

  const cookieStore = await cookies();
  const authToken = cookieStore.get('auth_token')?.value;

  const headers = new Headers(options.headers);
  headers.set('Content-Type', 'application/json');

  if (authToken) {
    headers.set('Cookie', `auth_token=${authToken}`);
  }

  if (API_SECRET_KEY) {
    headers.set('X-API-Key', API_SECRET_KEY);
  }

  try {
    const response = await fetch(url.toString(), {
      ...options,
      headers,
      cache: options.cache || 'no-store',
    });

    if (!response.ok) {
      console.error(`[SRE] Backend error on ${endpoint}: HTTP ${response.status}`);
      throw new Error(`API Error: ${response.status} ${response.statusText}`);
    }

    return await response.json();
  } catch (error) {
    console.error(`[SRE] Network/Parsing error on ${endpoint}:`, error);
    throw error;
  }
}