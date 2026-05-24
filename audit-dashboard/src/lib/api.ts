import { cookies } from 'next/headers';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// Оставляем X-API-Key только если он нужен как Server-to-Server защита на уровне балансировщика, 
// но для JWT авторизации пользователей он не используется.
const API_SECRET_KEY = process.env.API_SECRET_KEY;

interface FetchOptions extends RequestInit {
  params?: Record<string, string | number | boolean>;
}

export async function apiFetch<T>(
  endpoint: string,
  options: FetchOptions = {}
): Promise<T> {
  const url = new URL(`${API_URL}${endpoint}`);

  // O(N) формирование параметров, где N - количество query params
  if (options.params) {
    Object.entries(options.params).forEach(([key, value]) => {
      url.searchParams.append(key, String(value));
    });
  }

  // 1. Динамическое чтение куки (работает ТОЛЬКО в Server Components или Server Actions)
  const cookieStore = cookies();
  const authToken = (await cookieStore).get('auth_token')?.value;

  const headers = new Headers(options.headers);
  headers.set('Content-Type', 'application/json');

  // 2. Явно передаем JWT куку от клиента к Go-бэкенду
  if (authToken) {
    headers.set('Cookie', `auth_token=${authToken}`);
  }

  // 3. (Опционально) Передаем API ключ, если бэкенд все еще требует его через authMiddleware
  if (API_SECRET_KEY) {
    headers.set('X-API-Key', API_SECRET_KEY);
  }

  try {
    const response = await fetch(url.toString(), {
      ...options,
      headers,
      // В Next.js кэширование по умолчанию для fetch - 'force-cache'. 
      // Для динамичных данных дашборда лучше отключить:
      cache: options.cache || 'no-store', 
    });

    if (!response.ok) {
      // Логируем статус для дебага SRE
      console.error(`[SRE] Backend error on ${endpoint}: HTTP ${response.status}`);
      throw new Error(`API Error: ${response.status} ${response.statusText}`);
    }

    return await response.json();
  } catch (error) {
    console.error(`[SRE] Network/Parsing error on ${endpoint}:`, error);
    throw error;
  }
}