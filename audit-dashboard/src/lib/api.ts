import { cookies } from 'next/headers';

// В продакшене это будет внутри контейнера, например http://backend:8080
const API_URL = process.env.NEXT_API_INTERNAL_URL || 'http://localhost:8080'; 

interface FetchOptions extends RequestInit {
  params?: Record<string, string | number | boolean>;
}

// ВАЖНО: Эту функцию можно вызывать ТОЛЬКО в Server Components (page.tsx, layout.tsx) 
// или в Server Actions. Для Client Components используй стандартный браузерный fetch 
// с параметром { credentials: 'include' }.
export async function serverApiFetch<T>(
  endpoint: string,
  options: FetchOptions = {}
): Promise<T> {
  const url = new URL(`${API_URL}${endpoint}`);

  if (options.params) {
    Object.entries(options.params).forEach(([key, value]) => {
      url.searchParams.append(key, String(value));
    });
  }

  // Динамическое чтение куки (Server Side)
  const cookieStore = cookies();
  const authToken = (await cookieStore).get('auth_token')?.value;

  const headers = new Headers(options.headers);
  headers.set('Content-Type', 'application/json');

  if (authToken) {
    headers.set('Cookie', `auth_token=${authToken}`);
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