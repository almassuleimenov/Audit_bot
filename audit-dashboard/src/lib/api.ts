/**
 * Утилита для server-side API запросов с аутентификацией
 * Используется только в Server Components для безопасной передачи API ключа
 */
// src/lib/api.ts
const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
const API_SECRET_KEY = process.env.API_SECRET_KEY;

interface FetchOptions extends RequestInit {
  params?: Record<string, string | number | boolean>;
}

export async function apiFetch<T>(
  endpoint: string,
  options: FetchOptions = {}
): Promise<T> {
  if (!API_SECRET_KEY) {
    throw new Error('API_SECRET_KEY не определен в переменных окружения');
  }

  const url = new URL(`${API_URL}${endpoint}`);
  
  // Добавляем query параметры если были переданы
  if (options.params) {
    Object.entries(options.params).forEach(([key, value]) => {
      url.searchParams.append(key, String(value));
    });
  }

  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    'X-API-Key': API_SECRET_KEY,
    ...options.headers,
  };

  try {
    const response = await fetch(url.toString(), {
      ...options,
      headers,
    });

    if (!response.ok) {
      throw new Error(`API ошибка: ${response.status} ${response.statusText}`);
    }

    return await response.json();
  } catch (error) {
    console.error(`Ошибка при запросе ${endpoint}:`, error);
    throw error;
  }
}
