# Интеграция API с X-API-Key в Next.js

## 📋 Обзор

Это руководство по интеграции защищённого API бэкенда с фронтенд приложением Next.js. Все запросы к бэкенду используют заголовок `X-API-Key` для аутентификации.

## 🔒 Безопасность

### Важно: API ключ должен быть секретным!

- **API_SECRET_KEY** хранится в `.env.local` (серверная переменная)
- Он **никогда** не отправляется клиентской JavaScript bundle
- Все запросы с ключом происходят в **Server Components** или **Server Actions**

## 📁 Структура файлов

```
audit-dashboard/
├── .env.local                    # 🔒 Локальные переменные (не коммитить!)
├── .env.example                  # Пример конфигурации
└── src/
    ├── lib/
    │   └── api.ts               # ✨ Утилита для безопасных fetch запросов
    └── app/
        ├── page.tsx             # Server Component
        ├── audit/
        │   ├── page.tsx         # Server Component
        │   └── AuditClient.tsx   # Client Component (интерактивность)
        ├── appointments/
        │   └── page.tsx         # Server Component
        └── settings/
            ├── page.tsx         # Server Component
            ├── actions.ts       # Server Actions для операций
            └── SettingsClient.tsx # Client Component
```

## 🚀 Как это работает

### 1. Server Components (page.tsx)

**Пример: `src/app/page.tsx`**

```typescript
// ✅ Server Component - API ключ доступен
import { apiFetch } from '@/lib/api';

interface StatsData {
  total_audits: number;
  average_score: number;
  // ...
}

export default async function OverviewPage() {
  // Запрос с заголовком X-API-Key (ключ передан автоматически)
  let stats: StatsData | null = null;
  
  try {
    stats = await apiFetch<StatsData>('/api/stats');
  } catch (err) {
    // Обработка ошибок
  }

  return <div>{/* Рендеринг с данными */}</div>;
}
```

### 2. Server Actions (actions.ts)

**Пример: `src/app/settings/actions.ts`**

```typescript
'use server';

import { apiFetch } from '@/lib/api';

export async function toggleQuestionActive(question: SurveyQuestion) {
  // ✅ Server Action - API ключ доступен
  const updated = { ...question, is_active: !question.is_active };
  
  try {
    await apiFetch(`/api/questions`, {
      method: 'POST',
      body: JSON.stringify(updated)
    });
    return { success: true, updated };
  } catch (error) {
    return { success: false, error: error.message };
  }
}
```

### 3. Client Components с Server Actions

**Пример: `src/app/settings/SettingsClient.tsx`**

```typescript
'use client';

import { toggleQuestionActive } from './actions';

export default function SettingsClient({ initialQuestions }) {
  const handleToggle = async (question) => {
    // ✅ Вызывает Server Action (ключ недоступен на клиенте)
    const result = await toggleQuestionActive(question);
    // ...
  };

  return <button onClick={() => handleToggle(question)}>Активировать</button>;
}
```

## 🔧 Конфигурация

### 1. Установка переменных окружения

**`.env.local` (не коммитить!)**

```env
# Получи API ключ от своего бэкенда
API_SECRET_KEY=your-actual-api-key-here

# URL бэкенда (локальная разработка)
NEXT_PUBLIC_API_URL=http://localhost:8080

# Для production:
# NEXT_PUBLIC_API_URL=https://api.your-domain.com
```

### 2. Как получить API ключ?

Свяжись с командой бэкенда и получи:
- Значение `API_SECRET_KEY` для разработки
- Значение для production окружения

## 📝 Примеры использования

### Пример 1: GET запрос (страница Overview)

```typescript
// Server Component
const stats = await apiFetch<StatsData>('/api/stats');
```

**Что происходит:**
1. ✅ Запрос делается на сервере
2. ✅ Автоматически добавляется заголовок `X-API-Key: ${API_SECRET_KEY}`
3. ✅ Ключ никогда не видит клиент
4. ✅ Данные передаются компоненту как props

### Пример 2: POST запрос (Settings)

```typescript
// Server Action
export async function createQuestion(payload) {
  await apiFetch('/api/questions', {
    method: 'POST',
    body: JSON.stringify(payload)
  });
}

// Client Component вызывает Server Action
const result = await createQuestion(newQuestion);
```

### Пример 3: Export (Audit страница)

```typescript
// Client Component
const handleExport = () => {
  // ⚠️ Экспорт файла через окно браузера
  const exportUrl = new URL('/api/export', process.env.NEXT_PUBLIC_API_URL);
  window.open(exportUrl.toString(), '_blank');
};
```

## ⚙️ Утилита `apiFetch`

**Расположение:** `src/lib/api.ts`

```typescript
export async function apiFetch<T>(
  endpoint: string,
  options?: FetchOptions
): Promise<T>
```

**Возможности:**
- ✅ Автоматически добавляет `X-API-Key` заголовок
- ✅ Поддерживает query параметры через `options.params`
- ✅ Работает только на сервере (Server Components / Server Actions)
- ✅ Обработка ошибок и логирование

**Пример:**

```typescript
// GET с параметрами
const data = await apiFetch('/api/users', {
  params: { page: 1, limit: 10 }
});

// POST
const result = await apiFetch('/api/users', {
  method: 'POST',
  body: JSON.stringify({ name: 'John' })
});

// DELETE
await apiFetch('/api/users/123', {
  method: 'DELETE'
});
```

## 🔍 Проверка безопасности

### ✅ Безопасно:

```typescript
// Server Component - ключ недоступен клиенту
async function Page() {
  const data = await apiFetch('/api/data'); // ✅ Безопасно
  return <div>{data}</div>;
}
```

```typescript
// Server Action - ключ скрыт
'use server';
export async function updateData(item) {
  await apiFetch('/api/data', { method: 'POST', ... }); // ✅ Безопасно
}
```

### ⚠️ НЕ безопасно:

```typescript
// Client Component - никогда не используй API ключ!
'use client';
export default function Component() {
  // ❌ НЕПРАВИЛЬНО - ключ попадёт в bundle!
  const apiKey = process.env.API_SECRET_KEY;
  
  // ❌ НЕПРАВИЛЬНО - ключ отправляется в браузере!
  fetch('http://localhost:8080/api/data', {
    headers: { 'X-API-Key': apiKey }
  });
}
```

## 🐛 Troubleshooting

### Проблема: 401 Unauthorized

**Причины:**
- `API_SECRET_KEY` не установлен в `.env.local`
- Неправильное значение API ключа
- Бэкенд недоступен

**Решение:**
1. Проверь `.env.local`
2. Перезагрузи dev сервер (`npm run dev`)
3. Проверь, запущен ли бэкенд

### Проблема: API ключ видимый в браузере

**Причины:**
- Используется `NEXT_PUBLIC_` префикс для `API_SECRET_KEY` ❌
- Server Component используется как Client Component (`'use client'`)

**Решение:**
- Никогда не используй `NEXT_PUBLIC_API_SECRET_KEY`
- Используй Server Components или Server Actions
- Используй правильный паттерн разделения компонентов

## 📚 Дополнительные ресурсы

- [Next.js Server Components](https://nextjs.org/docs/app/building-your-application/rendering/server-components)
- [Next.js Server Actions](https://nextjs.org/docs/app/building-your-application/data-fetching/server-actions-and-mutations)
- [Environment Variables](https://nextjs.org/docs/app/building-your-application/configuring/environment-variables)
